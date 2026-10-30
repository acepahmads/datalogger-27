package service_test

import (
	"context"
	"testing"
	"time"

	"datalogger/internal/model"
	"datalogger/internal/quality"
	"datalogger/internal/repository"
	"datalogger/internal/service"
)

func TestDiscreteStatusParameter_SuppressesStuckSpikeAndZeroAlarms(t *testing.T) {
	db, _, alarmService, _ := setupTestEvaluatorEnvironment(t)
	_ = db.AutoMigrate(&model.ParameterCategory{})
	ctx := context.Background()

	statusCat := model.ParameterCategory{
		Code:    "status",
		Name:    "Status",
		Domain:  model.DomainAll,
		IconKey: "power",
		Enabled: true,
	}
	db.Create(&statusCat)

	dev := model.Device{
		DeviceCode: "com2",
		DeviceName: "Pump Controller",
		Status:     model.DeviceStatusActive,
		Enabled:    true,
	}
	db.Create(&dev)

	warnLow := 1.0 // Even if cloned with a legacy WarningLow=1.0, 0 (MATI) must NOT alarm
	pumpParam := model.Parameter{
		DeviceID:                 dev.ID,
		ParameterCode:            "pompa",
		ParameterName:            "POMPA",
		DataType:                 model.DataTypeFloat32,
		CategoryID:               &statusCat.ID,
		Enabled:                  true,
		QualityValidationEnabled: true,
		SpikeDetectionEnabled:    true,
		SpikeThreshold:           0.5,
		WarningLow:               &warnLow,
		ZeroValueAlarmPolicy:     model.ZeroValuePolicyInvalidReading, // Even if misconfigured, Status overrides 0 as valid OFF
	}
	db.Create(&pumpParam)

	devRepo := repository.NewDeviceRepository(db)

	// Seed a legacy false STUCK_VALUE alarm in DB (simulating existing production state)
	zeroVal := 0.0
	legacyStuck := &model.Alarm{
		DeviceID:     &dev.ID,
		ParameterID:  &pumpParam.ID,
		AlarmCode:    service.AlarmCodeAbnormalityStuck,
		Severity:     model.SeverityWarning,
		Status:       model.AlarmActive,
		Message:      "Parameter POMPA on device com2 has STUCK value: 0.00 (32 identical readings over 30s)",
		TriggerValue: &zeroVal,
		TriggeredAt:  time.Now().UTC().Add(-5 * time.Minute),
	}
	db.Create(legacyStuck)

	evaluator := service.NewAlarmEvaluator(
		ctx,
		alarmService,
		service.DefaultAlarmEvaluatorConfig(),
		func(c context.Context, paramID uint) (*model.Parameter, error) {
			return devRepo.FindParameterByID(paramID)
		},
	)

	// Allow async ClearAlarm in ReloadActiveAlarms to finish
	time.Sleep(50 * time.Millisecond)
	var reloaded model.Alarm
	db.First(&reloaded, legacyStuck.ID)
	if reloaded.Status != model.AlarmCleared {
		t.Fatalf("Expected legacy STUCK_VALUE alarm on POMPA to auto-clear on ReloadActiveAlarms, got %s", reloaded.Status)
	}

	qp := quality.NewQualityProcessor()
	baseTime := time.Now().UTC().Add(-2 * time.Minute)

	// 1. Feed 35 consecutive 0.00 readings over 70 seconds (pump MATI)
	// Pass param WITHOUT Category preloaded to also verify lazy Category hydration via paramLookup!
	unpreloadedCopy := pumpParam
	unpreloadedCopy.Category = nil

	for i := 0; i < 35; i++ {
		ts := baseTime.Add(time.Duration(i*2) * time.Second)
		vres := qp.Validate(&pumpParam, &quality.ValidationInput{
			DeviceID:    dev.ID,
			ParameterID: pumpParam.ID,
			RawValue:    0,
			Value:       0,
			DataType:    model.DataTypeFloat32,
			ReceivedAt:  ts,
		})
		// Reload Category onto pumpParam for quality test check too
		loadedParam, _ := devRepo.FindParameterByID(pumpParam.ID)
		vres = qp.Validate(loadedParam, &quality.ValidationInput{
			DeviceID:    dev.ID,
			ParameterID: pumpParam.ID,
			RawValue:    0,
			Value:       0,
			DataType:    model.DataTypeFloat32,
			ReceivedAt:  ts,
		})
		if vres.Quality != model.QualityGood {
			t.Fatalf("Expected POMPA value 0.0 to validate as GOOD, got %s (%s)", vres.Quality, vres.QualityReason)
		}

		evaluator.EvaluateTelemetry(ctx, &service.LatestTelemetry{
			DeviceID:       dev.ID,
			DeviceCode:     dev.DeviceCode,
			ParameterID:    pumpParam.ID,
			ParameterCode:  pumpParam.ParameterCode,
			ProcessedValue: 0.0,
			Quality:        vres.Quality,
			QualityFlags:   vres.QualityFlags,
			QualityReason:  vres.QualityReason,
			ReceivedAt:     ts,
		}, &unpreloadedCopy)
	}

	// 2. Instant transition 0 -> 1 followed by 35 consecutive 1.00 readings (pump NYALA)
	loadedParam, _ := devRepo.FindParameterByID(pumpParam.ID)
	for i := 0; i < 35; i++ {
		ts := baseTime.Add(time.Duration(70+i*2) * time.Second)
		vres := qp.Validate(loadedParam, &quality.ValidationInput{
			DeviceID:    dev.ID,
			ParameterID: pumpParam.ID,
			RawValue:    1.0,
			Value:       1.0,
			DataType:    model.DataTypeFloat32,
			ReceivedAt:  ts,
		})
		if vres.IsSpike || vres.Quality != model.QualityGood {
			t.Fatalf("Expected 0->1 transition on POMPA NOT to flag spike, got IsSpike=%v Quality=%s", vres.IsSpike, vres.Quality)
		}

		evaluator.EvaluateTelemetry(ctx, &service.LatestTelemetry{
			DeviceID:       dev.ID,
			DeviceCode:     dev.DeviceCode,
			ParameterID:    pumpParam.ID,
			ParameterCode:  pumpParam.ParameterCode,
			ProcessedValue: 1.0,
			Quality:        vres.Quality,
			QualityFlags:   vres.QualityFlags,
			QualityReason:  vres.QualityReason,
			ReceivedAt:     ts,
		}, loadedParam)
	}

	activeAlarms, err := alarmService.GetActiveOrAckedAlarms(ctx)
	if err != nil {
		t.Fatalf("Failed querying active alarms: %v", err)
	}
	if len(activeAlarms) != 0 {
		t.Fatalf("Expected 0 active alarms for POMPA status across steady 0, 0->1 transition, and steady 1, got %d: %+v", len(activeAlarms), activeAlarms)
	}

	// 3. Genuine communication / bad quality failure MUST still trigger DATA_ABNORMALITY_BAD_QUALITY!
	for i := 0; i < 6; i++ {
		ts := baseTime.Add(time.Duration(150+i*2) * time.Second)
		evaluator.EvaluateTelemetry(ctx, &service.LatestTelemetry{
			DeviceID:       dev.ID,
			DeviceCode:     dev.DeviceCode,
			ParameterID:    pumpParam.ID,
			ParameterCode:  pumpParam.ParameterCode,
			ProcessedValue: 0,
			Quality:        model.QualityBad,
			QualityReason:  model.ReasonCommunicationError,
			ReceivedAt:     ts,
		}, loadedParam)
	}
	activeAfterCommFail, _ := alarmService.GetActiveOrAckedAlarms(ctx)
	if len(activeAfterCommFail) != 1 || activeAfterCommFail[0].AlarmCode != service.AlarmCodeAbnormalityBad {
		t.Fatalf("Expected genuine communication failure on POMPA to trigger DATA_ABNORMALITY_BAD_QUALITY, got: %+v", activeAfterCommFail)
	}
}
