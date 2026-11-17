package service_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"datalogger/internal/model"
	"datalogger/internal/repository"
	"datalogger/internal/service"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type countingPersistenceBoundary struct {
	mu           sync.Mutex
	savedRecords []*model.RawData
	batchCalls   int
}

func (c *countingPersistenceBoundary) SaveBatch(ctx context.Context, batch []*model.RawData) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.batchCalls++
	for _, r := range batch {
		cp := *r
		c.savedRecords = append(c.savedRecords, &cp)
	}
	return nil
}

func (c *countingPersistenceBoundary) CountByQuality(q model.TelemetryQuality) int {
	c.mu.Lock()
	defer c.mu.Unlock()
	n := 0
	for _, r := range c.savedRecords {
		if r.Quality == q {
			n++
		}
	}
	return n
}

func TestOfflineCPUOptimization_NoPingPongAndDeduplicatedBadIngest(t *testing.T) {
	dsn := fmt.Sprintf("file:offline_cpu_opt_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("failed opening in-memory sqlite: %v", err)
	}

	if err := db.AutoMigrate(
		&model.Tenant{},
		&model.Site{},
		&model.Device{},
		&model.DeviceConnection{},
		&model.ParameterCategory{},
		&model.Parameter{},
		&model.RawData{},
		&model.Alarm{},
		&model.AuditTrail{},
		&model.SystemLog{},
	); err != nil {
		t.Fatalf("failed migrating schema: %v", err)
	}

	deviceRepo := repository.NewDeviceRepository(db)
	alarmRepo := repository.NewAlarmRepository(db)
	systemRepo := repository.NewSystemRepository(db)
	alarmService := service.NewAlarmService(alarmRepo, systemRepo)

	dev := &model.Device{
		TenantID:         1,
		SiteID:           1,
		DeviceCode:       "AQMS-GAS-OFFLINE-TEST",
		DeviceName:       "AQMS Gas Analyzer",
		DeviceType:       "MODBUS_RTU",
		Status:           model.DeviceStatusActive,
		Enabled:          true,
		ConnectionStatus: model.DeviceConnOnline,
		Connection: &model.DeviceConnection{
			Protocol:        model.ProtocolModbusRTU,
			SerialPort:      "COM3",
			BaudRate:        9600,
			SlaveID:         1,
			Timeout:         1000,
			PollingInterval: 1000,
			Enabled:         true,
		},
	}
	if err := deviceRepo.Create(dev); err != nil {
		t.Fatalf("failed creating device: %v", err)
	}

	const paramCount = 20
	params := make([]*model.Parameter, 0, paramCount)
	for i := 1; i <= paramCount; i++ {
		stuckOn := true
		p := &model.Parameter{
			DeviceID:                 dev.ID,
			ParameterCode:            fmt.Sprintf("GAS_%02d", i),
			ParameterName:            fmt.Sprintf("Gas Channel %02d", i),
			DataType:                 model.DataTypeFloat32,
			Unit:                     "ppb",
			Scale:                    1.0,
			Precision:                2,
			QualityValidationEnabled: true,
			ProcessingEnabled:        true,
			StaleTimeoutSeconds:      120,
			StuckDetectionEnabled:    &stuckOn,
			Enabled:                  true,
		}
		if err := deviceRepo.CreateParameter(p); err != nil {
			t.Fatalf("failed creating parameter %d: %v", i, err)
		}
		params = append(params, p)
	}

	var onlineStatusCalls int32
	var offlineStatusCalls int32

	evalCfg := service.DefaultAlarmEvaluatorConfig()
	evalCfg.BadQualityCountThreshold = 5
	evalCfg.CommErrorConsecutiveLimit = 3
	evalCfg.DeviceOfflineTimeout = 60 * time.Second

	evaluator := service.NewAlarmEvaluator(
		context.Background(),
		alarmService,
		evalCfg,
		func(ctx context.Context, paramID uint) (*model.Parameter, error) {
			return deviceRepo.FindCachedParameterForPipeline(paramID)
		},
	)
	evaluator.SetDeviceLookup(func(ctx context.Context, deviceID uint) (*model.Device, error) {
		return deviceRepo.GetCachedDeviceForPipeline(deviceID)
	})
	evaluator.SetDeviceStatusUpdater(func(deviceID uint, status model.DeviceConnectionStatus) error {
		if status == model.DeviceConnOnline {
			atomic.AddInt32(&onlineStatusCalls, 1)
		} else if status == model.DeviceConnOffline {
			atomic.AddInt32(&offlineStatusCalls, 1)
		}
		return deviceRepo.SetConnectionStatus(deviceID, status)
	})
	alarmService.SetAlarmEvaluator(evaluator)

	telCfg := service.DefaultTelemetryConfig()
	telCfg.QueueEnabled = false
	telCfg.FlushInterval = 20 * time.Millisecond
	telCfg.BatchSize = 50

	telSvc := service.NewTelemetryService(nil, nil, telCfg)
	defer telSvc.Stop()

	boundary := &countingPersistenceBoundary{}
	telSvc.SetPersistenceBoundary(boundary)
	telSvc.SetDeviceLookup(deviceRepo.GetCachedDeviceForPipeline)
	telSvc.SetParamLookup(deviceRepo.FindCachedParameterForPipeline)
	telSvc.SetAlarmEvaluator(evaluator)

	// Phase 1: Ingest 1 healthy poll cycle across all 20 parameters
	baseNow := time.Now().UTC().Add(-2 * time.Hour)
	for idx, p := range params {
		val := float64(10 + idx)
		if err := telSvc.Ingest(&service.TelemetryIngestPayload{
			TenantID:      dev.TenantID,
			SiteID:        dev.SiteID,
			DeviceID:      dev.ID,
			DeviceCode:    dev.DeviceCode,
			DeviceName:    dev.DeviceName,
			ParameterID:   p.ID,
			ParameterCode: p.ParameterCode,
			ParameterName: p.ParameterName,
			Unit:          p.Unit,
			DataType:      p.DataType,
			RawValue:      val,
			Value:         val,
			Quality:       model.QualityGood,
			ReceivedAt:    baseNow,
			Parameter:     p,
		}); err != nil {
			t.Fatalf("healthy ingest failed: %v", err)
		}
	}
	_ = telSvc.FlushNow(context.Background())

	if got := atomic.LoadInt32(&onlineStatusCalls); got != 1 {
		t.Fatalf("expected steady-state healthy sync to update ONLINE status once (not 20x), got %d", got)
	}

	// Reset onlineStatusCalls counter before simulating prolonged offline outage
	atomic.StoreInt32(&onlineStatusCalls, 0)

	// Mark device offline in evaluator
	pastSuccess := baseNow
	devSnap, err := deviceRepo.GetCachedDeviceForPipeline(dev.ID)
	if err != nil {
		t.Fatalf("cached device lookup failed: %v", err)
	}
	devSnap.LastDataAt = &pastSuccess
	readErr := errors.New("serial read timeout")
	evaluator.EvaluateDeviceCommFailure(context.Background(), devSnap, &service.DeviceHealthInput{
		DeviceID:            dev.ID,
		State:               "ERROR",
		ConsecutiveFailures: 5,
		LastCommSuccess:     &pastSuccess,
	}, readErr)
	evaluator.EvaluateDeviceOffline(context.Background(), devSnap, &service.DeviceHealthInput{
		DeviceID:            dev.ID,
		State:               "OFFLINE",
		ConsecutiveFailures: 5,
		LastCommSuccess:     &pastSuccess,
	})

	if gotOff := atomic.LoadInt32(&offlineStatusCalls); gotOff != 1 {
		t.Fatalf("expected offline status update once, got %d", gotOff)
	}

	// Phase 2: Simulate 30 failed offline poll cycles across all 20 parameters (600 BAD ingests!)
	for cycle := 1; cycle <= 30; cycle++ {
		errTs := time.Now().UTC()
		for idx, p := range params {
			lastVal := float64(10 + idx)
			if err := telSvc.Ingest(&service.TelemetryIngestPayload{
				TenantID:      dev.TenantID,
				SiteID:        dev.SiteID,
				DeviceID:      dev.ID,
				DeviceCode:    dev.DeviceCode,
				DeviceName:    dev.DeviceName,
				ParameterID:   p.ID,
				ParameterCode: p.ParameterCode,
				ParameterName: p.ParameterName,
				Unit:          p.Unit,
				DataType:      p.DataType,
				RawValue:      lastVal,
				Value:         lastVal,
				Quality:       model.QualityBad,
				QualityReason: model.ReasonCommunicationError,
				QualityFlags:  model.FlagCommError,
				ReceivedAt:    errTs,
				Parameter:     p,
				ReadError:     readErr,
			}); err != nil {
				t.Fatalf("offline ingest cycle %d failed: %v", cycle, err)
			}
		}
	}
	_ = telSvc.FlushNow(context.Background())

	// 1. Must NEVER falsely flip device back to ONLINE during failed reads!
	if gotOnline := atomic.LoadInt32(&onlineStatusCalls); gotOnline != 0 {
		t.Fatalf("CRITICAL: expected 0 ONLINE status flips during 600 failed offline reads, got %d", gotOnline)
	}

	// 2. Only the initial transition to BAD (20 records, 1 per parameter) should be persisted, NOT 600!
	if badSaved := boundary.CountByQuality(model.QualityBad); badSaved != paramCount {
		t.Fatalf("expected exactly %d transition BAD records persisted (1 per param), got %d", paramCount, badSaved)
	}

	// 3. All 20 parameters must still trigger DATA_ABNORMALITY_BAD_QUALITY alarms on 5th consecutive failure!
	activeAlarms, err := alarmRepo.GetActive(context.Background())
	if err != nil {
		t.Fatalf("failed listing active alarms: %v", err)
	}
	badQualityAlarms := 0
	hasOfflineAlarm := false
	hasCommAlarm := false
	for _, a := range activeAlarms {
		switch a.AlarmCode {
		case service.AlarmCodeAbnormalityBad:
			badQualityAlarms++
		case service.AlarmCodeDeviceOffline:
			hasOfflineAlarm = true
		case service.AlarmCodeCommError:
			hasCommAlarm = true
		}
	}
	if badQualityAlarms != paramCount {
		t.Fatalf("expected %d DATA_ABNORMALITY_BAD_QUALITY alarms, got %d", paramCount, badQualityAlarms)
	}
	if !hasOfflineAlarm || !hasCommAlarm {
		t.Fatalf("expected DEVICE_OFFLINE (%v) and COMMUNICATION_ERROR (%v) to remain active without ping-pong clearing", hasOfflineAlarm, hasCommAlarm)
	}

	// 4. Verify pipeline cache invalidation immediately reflects parameter disablement
	if err := deviceRepo.SetParameterEnabled(dev.ID, params[0].ID, false); err != nil {
		t.Fatalf("failed disabling parameter 0: %v", err)
	}
	cachedP0, err := deviceRepo.FindCachedParameterForPipeline(params[0].ID)
	if err != nil || cachedP0 == nil || cachedP0.Enabled {
		t.Fatalf("expected FindCachedParameterForPipeline to immediately reflect Enabled=false, got %+v (err=%v)", cachedP0, err)
	}
	if err := deviceRepo.SetParameterEnabled(dev.ID, params[0].ID, true); err != nil {
		t.Fatalf("failed re-enabling parameter 0: %v", err)
	}

	// Phase 3: Device recovers and sends valid GOOD readings -> all alarms auto-clear and status flips ONLINE once
	recovNow := time.Now().UTC()
	for idx, p := range params {
		val := float64(25 + idx)
		if err := telSvc.Ingest(&service.TelemetryIngestPayload{
			TenantID:      dev.TenantID,
			SiteID:        dev.SiteID,
			DeviceID:      dev.ID,
			DeviceCode:    dev.DeviceCode,
			DeviceName:    dev.DeviceName,
			ParameterID:   p.ID,
			ParameterCode: p.ParameterCode,
			ParameterName: p.ParameterName,
			Unit:          p.Unit,
			DataType:      p.DataType,
			RawValue:      val,
			Value:         val,
			Quality:       model.QualityGood,
			ReceivedAt:    recovNow,
			Parameter:     p,
		}); err != nil {
			t.Fatalf("recovery ingest failed: %v", err)
		}
	}
	_ = telSvc.FlushNow(context.Background())

	if gotOnline := atomic.LoadInt32(&onlineStatusCalls); gotOnline != 1 {
		t.Fatalf("expected exactly 1 ONLINE status sync upon recovery across 20 parameters, got %d", gotOnline)
	}
	if goodSaved := boundary.CountByQuality(model.QualityGood); goodSaved != paramCount*2 {
		t.Fatalf("expected %d GOOD records persisted (initial 20 + recovered 20), got %d", paramCount*2, goodSaved)
	}
	remainingAlarms, _ := alarmRepo.GetActive(context.Background())
	if len(remainingAlarms) != 0 {
		t.Fatalf("expected all offline/comm/bad-quality alarms to auto-clear on recovery, got %d remaining", len(remainingAlarms))
	}
}
