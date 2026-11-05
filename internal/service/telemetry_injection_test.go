package service_test

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"datalogger/internal/model"
	"datalogger/internal/repository"
	"datalogger/internal/service"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func setupTelemetryInjectionTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := fmt.Sprintf("file:meminj_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		t.Fatalf("failed opening sqlite: %v", err)
	}

	if err := db.AutoMigrate(
		&model.DeviceType{},
		&model.Device{},
		&model.DeviceConnection{},
		&model.ParameterCategory{},
		&model.Parameter{},
		&model.RawData{},
		&model.AggregationDefinition{},
		&model.AggregationResult{},
		&model.OutputDestination{},
		&model.DestinationPayloadConfig{},
		&model.OutboundDeliveryQueue{},
		&model.DeliveryLog{},
		&model.TelemetryInjectionBatch{},
		&model.TelemetryInjectionRecord{},
		&model.SystemLog{},
		&model.AuditTrail{},
	); err != nil {
		t.Fatalf("failed migrating test db: %v", err)
	}
	return db
}

func TestTelemetryInjection_FillMissingSparingWindow_ZeroLeak_AndRollback(t *testing.T) {
	db := setupTelemetryInjectionTestDB(t)
	ctx := context.Background()

	devRepo := repository.NewDeviceRepository(db)
	aggRepo := repository.NewAggregationRepository(db)
	sysRepo := repository.NewSystemRepository(db)
	outRepo := repository.NewOutboundRepository(db)
	injRepo := repository.NewTelemetryInjectionRepository(db)

	aggSvc := service.NewAggregationService(aggRepo, devRepo, sysRepo, nil)
	outSvc := service.NewOutboundService(outRepo, devRepo, service.NewConfigurablePayloadFormatter())
	injSvc := service.NewTelemetryInjectionService(injRepo, devRepo, aggRepo, aggSvc, outSvc)

	// 1. Create active WQMS/SPARING device + 5 parameters (pH, cod, tss, nh3n, debit)
	dev := &model.Device{
		ID:         10,
		DeviceCode: "SPARING-01",
		DeviceName: "WWTP Outlet",
		DeviceType: "MODBUS_RTU",
		Timezone:   "Asia/Jakarta",
		Status:     model.DeviceStatusActive,
		Enabled:    true,
	}
	if err := db.Create(dev).Error; err != nil {
		t.Fatalf("create device: %v", err)
	}

	paramSpecs := []struct {
		id    uint
		code  string
		name  string
		unit  string
		scale float64
		dt    model.ParameterDataType
	}{
		{1, "pH", "pH", "", 0.01, model.DataTypeInt16},
		{2, "cod", "COD", "mg/L", 1.0, model.DataTypeFloat32},
		{3, "tss", "TSS", "mg/L", 1.0, model.DataTypeFloat32},
		{4, "nh3n", "NH3-N", "mg/L", 1.0, model.DataTypeFloat32},
		{5, "debit", "Debit", "m3/s", 1.0, model.DataTypeFloat32},
	}
	for _, ps := range paramSpecs {
		p := &model.Parameter{
			ID:            ps.id,
			DeviceID:      dev.ID,
			ParameterCode: ps.code,
			ParameterName: ps.name,
			Unit:          ps.unit,
			DataType:      ps.dt,
			Scale:         ps.scale,
			Precision:     2,
			ByteOrder:     "ABCD",
			Enabled:       true,
		}
		if err := db.Create(p).Error; err != nil {
			t.Fatalf("create param %s: %v", ps.code, err)
		}

		// Create both CUSTOMER_PROCESSED and INTERNAL_RAW 2-minute definitions
		for _, srcType := range []model.AggregationSourceType{model.SourceCustomerProcessed, model.SourceInternalRaw} {
			def := &model.AggregationDefinition{
				Name:            fmt.Sprintf("%s 2m %s", ps.code, srcType),
				Code:            fmt.Sprintf("%s_2m_%s", ps.code, srcType),
				SourceType:      srcType,
				DeviceID:        dev.ID,
				ParameterID:     ps.id,
				Function:        model.FunctionAvg,
				IntervalSeconds: 120,
				Timezone:        "UTC",
				Enabled:         true,
			}
			if err := db.Create(def).Error; err != nil {
				t.Fatalf("create agg def: %v", err)
			}
		}
	}

	// 2. Create SPARING 1-hour bundled destination (disabled during test so background worker doesn't dial HTTP, while queue payload is built)
	devID := dev.ID
	dest := &model.OutputDestination{
		Name:                    "KLHK SPARING",
		Type:                    model.DestREST,
		DeviceID:                &devID,
		Endpoint:                "https://sparing.example.test/post",
		AuthType:                "SPARING_JWT_STATIC",
		AuthToken:               "test-secret",
		IntervalSeconds:         120,
		DispatchIntervalSeconds: 3600,
		Enabled:                 true,
		RetryLimit:              3,
	}
	if err := outSvc.CreateDestination(ctx, dest); err != nil {
		t.Fatalf("create sparing destination: %v", err)
	}
	cfg, err := outSvc.GetPayloadConfig(ctx, dest.ID)
	if err != nil {
		t.Fatalf("get payload config: %v", err)
	}
	cfg.StationIDValue = "65c05cac405f557259fed7cc"
	cfg.Structure = "SPARING_DATA_ARRAY"
	cfg.TemplateType = "KLHK_SPARING_HOURLY"
	if err := outSvc.UpdatePayloadConfig(ctx, cfg); err != nil {
		t.Fatalf("update payload config: %v", err)
	}

	// Choose a completed 1-hour window aligned to 3600s two hours ago
	nowAligned := time.Unix((time.Now().UTC().Unix()/3600)*3600, 0).UTC()
	winStart := nowAligned.Add(-2 * time.Hour)
	winEnd := nowAligned.Add(-1 * time.Hour)

	// 3. Seed partial real data matching user's exact scenario:
	// - Buckets 0..11 (first 24m): only debit BAD (comm error), pH/cod/tss/nh3n missing
	// - Buckets 12..14 (24..30m): valid pH/cod/tss/nh3n, debit BAD (null)
	// - Buckets 15..29 (30..60m): all 5 valid
	insertSample := func(paramID uint, ts time.Time, val float64, qual model.TelemetryQuality) {
		v := val
		r := &model.RawData{
			DeviceID:       dev.ID,
			ParameterID:    paramID,
			Value:          val,
			ProcessedValue: val,
			ValueNumeric:   &v,
			RawValue:       val,
			RawHex:         "40 DD CC CD",
			Quality:        qual,
			ReceivedAt:     ts,
			StoredAt:       ts,
			Timestamp:      ts,
		}
		if err := db.Create(r).Error; err != nil {
			t.Fatalf("seed raw: %v", err)
		}
	}

	// 3. Seed partial real data matching user's exact production scenario:
	// - Buckets 0..11 (first 24m): debit had empty aggregation buckets (Value=nil -> "debit":null), pH/cod/tss/nh3n missing
	// - Buckets 12..13 (24..28m): valid pH/cod/tss/nh3n, debit empty bucket (Value=nil -> "debit":null)
	// - Bucket 14 (28..30m): valid pH/cod/tss/nh3n, debit has QualityBad comm-error sample
	// - Buckets 15..29 (30..60m): all 5 valid
	var debitCustDef model.AggregationDefinition
	if err := db.Where("device_id = ? AND parameter_id = ? AND source_type = ?", dev.ID, 5, model.SourceCustomerProcessed).First(&debitCustDef).Error; err != nil {
		t.Fatalf("find debit def: %v", err)
	}

	for b := 0; b < 30; b++ {
		bStart := winStart.Add(time.Duration(b*120) * time.Second)
		bEnd := bStart.Add(120 * time.Second)
		ts := bStart.Add(15 * time.Second) // Real sensor offset (+15s, not exact :00!)
		if b < 12 {
			// Empty bucket persisted by real-time worker (Value=nil -> "debit": null)
			if err := aggRepo.SaveResult(ctx, &model.AggregationResult{
				AggregationDefinitionID: debitCustDef.ID,
				DeviceID:                dev.ID,
				ParameterID:             5,
				SourceType:              model.SourceCustomerProcessed,
				PeriodStart:             bStart,
				PeriodEnd:               bEnd,
				Identifier:              bEnd.Format("20060102150405"),
				Function:                model.FunctionAvg,
				IntervalSeconds:         120,
				Value:                   nil,
				Quality:                 model.QualityBad,
				QualityReason:           model.ReasonNullValue,
			}); err != nil {
				t.Fatalf("save empty debit bucket: %v", err)
			}
		} else if b < 14 {
			insertSample(1, ts, 6.90, model.QualityGood)
			insertSample(2, ts, 68.05, model.QualityGood)
			insertSample(3, ts, 0.19, model.QualityGood)
			insertSample(4, ts, 84.05, model.QualityGood)
			_ = aggRepo.SaveResult(ctx, &model.AggregationResult{
				AggregationDefinitionID: debitCustDef.ID,
				DeviceID:                dev.ID,
				ParameterID:             5,
				SourceType:              model.SourceCustomerProcessed,
				PeriodStart:             bStart,
				PeriodEnd:               bEnd,
				Identifier:              bEnd.Format("20060102150405"),
				Function:                model.FunctionAvg,
				IntervalSeconds:         120,
				Value:                   nil,
				Quality:                 model.QualityBad,
				QualityReason:           model.ReasonNullValue,
			})
		} else if b == 14 {
			insertSample(1, ts, 6.90, model.QualityGood)
			insertSample(2, ts, 68.05, model.QualityGood)
			insertSample(3, ts, 0.19, model.QualityGood)
			insertSample(4, ts, 84.05, model.QualityGood)
			insertSample(5, ts, 0, model.QualityBad) // Comm error sample -> sanitized once GOOD samples are injected!
		} else {
			insertSample(1, ts, 6.90, model.QualityGood)
			insertSample(2, ts, 109.05, model.QualityGood)
			insertSample(3, ts, 28.50, model.QualityGood)
			insertSample(4, ts, 6.89, model.QualityGood)
			insertSample(5, ts, 0.85, model.QualityGood)
		}
	}

	// Calculate initial aggregations and enqueue the incomplete 30-bucket SPARING queue item
	if _, _, err := aggSvc.RecalculateRangeForDevice(ctx, dev.ID, nil, winStart, winEnd, false); err != nil {
		t.Fatalf("initial recalc: %v", err)
	}
	if err := outSvc.HandleBucketCompleted(ctx, dev.ID, winEnd); err != nil {
		t.Fatalf("handle bucket completed: %v", err)
	}

	qItemsBefore, totalQ, err := outSvc.GetQueueItems(ctx, repository.QueueFilter{DestinationID: &dest.ID, Page: 1, PageSize: 10})
	if err != nil || totalQ != 1 {
		t.Fatalf("expected 1 queued SPARING batch, got %d (err=%v)", totalQ, err)
	}
	if !strings.Contains(qItemsBefore[0].PayloadSnapshot, `"debit":null`) {
		t.Fatalf("expected initial SPARING payload to contain null debit, got: %s", qItemsBefore[0].PayloadSnapshot)
	}

	// 4. Execute Superadmin Telemetry Injection in FILL_MISSING_AND_NULL mode across full 1h window
	execRes, err := injSvc.ExecuteInjection(ctx, service.TelemetryInjectionRequest{
		DeviceID:          dev.ID,
		StartTime:         winStart,
		EndTime:           winEnd,
		SampleStepSeconds: 30,
		ConflictMode:      service.ConflictFillMissingAndNull,
		RandomMode:        service.RandomModeSmoothWalk,
		ValueTarget:       service.ValueTargetProcessed,
		Seed:              42,
		Note:              "Isi gap sensor awal jam",
		Parameters: []service.TelemetryInjectionParamInput{
			{ParameterID: 1, MinValue: 6.86, MaxValue: 6.94},
			{ParameterID: 2, MinValue: 67.50, MaxValue: 68.50},
			{ParameterID: 3, MinValue: 0.18, MaxValue: 0.22},
			{ParameterID: 4, MinValue: 83.50, MaxValue: 84.50},
			{ParameterID: 5, MinValue: 0.83, MaxValue: 0.87},
		},
	}, "superadmin")
	if err != nil {
		t.Fatalf("ExecuteInjection failed: %v", err)
	}

	// pH/cod/tss/nh3n have 12 missing 2m buckets * 4 samples/bucket = 48 each (192)
	// debit has 15 missing/BAD 2m buckets * 4 samples/bucket = 60
	// Total expected injected samples = 192 + 60 = 252!
	if execRes.Batch.InjectedRawCount != 252 {
		t.Fatalf("expected exactly 252 injected raw samples (preserving valid buckets), got %d", execRes.Batch.InjectedRawCount)
	}
	if execRes.Batch.RefreshedOutbound != 1 {
		t.Fatalf("expected 1 refreshed outbound queue item, got %d", execRes.Batch.RefreshedOutbound)
	}

	// 5. Verify existing SPARING outbound payload is now 100% complete with 30 rows and ZERO nulls, and ZERO injection metadata leaked!
	qItemsAfter, _, _ := outSvc.GetQueueItems(ctx, repository.QueueFilter{DestinationID: &dest.ID, Page: 1, PageSize: 10})
	updatedSnap := qItemsAfter[0].PayloadSnapshot
	if strings.Contains(updatedSnap, "null") {
		t.Fatalf("expected zero nulls in re-hydrated SPARING payload, got: %s", updatedSnap)
	}
	if strings.Contains(updatedSnap, "is_injected") || strings.Contains(updatedSnap, "has_injected") {
		t.Fatalf("LEAK DETECTED in outbound payload: %s", updatedSnap)
	}

	var sparingDoc struct {
		UID  string                   `json:"uid"`
		Data []map[string]interface{} `json:"data"`
	}
	if err := json.Unmarshal([]byte(updatedSnap), &sparingDoc); err != nil {
		t.Fatalf("invalid SPARING JSON: %v", err)
	}
	if len(sparingDoc.Data) != 30 {
		t.Fatalf("expected 30 buckets in SPARING data[], got %d", len(sparingDoc.Data))
	}
	// Bucket index 12 (24m..26m) must preserve untouched real cod=68.05 while filling debit in [0.83..0.87]
	if codB12, ok := sparingDoc.Data[12]["cod"].(float64); !ok || codB12 != 68.05 {
		t.Fatalf("expected untouched real cod=68.05 in bucket 12, got %v", sparingDoc.Data[12]["cod"])
	}
	if debitB12, ok := sparingDoc.Data[12]["debit"].(float64); !ok || debitB12 < 0.80 || debitB12 > 0.90 {
		t.Fatalf("expected injected debit in [0.80, 0.90] in bucket 12, got %v", sparingDoc.Data[12]["debit"])
	}
	// Bucket index 20 (40m..42m) must preserve untouched real cod=109.05 and debit=0.85
	if codB20, ok := sparingDoc.Data[20]["cod"].(float64); !ok || codB20 != 109.05 {
		t.Fatalf("expected untouched real cod=109.05 in bucket 20, got %v", sparingDoc.Data[20]["cod"])
	}

	// 6. Verify Zero-Leak in AuditTrail/SystemLog & Non-Superadmin DB Queries vs Superadmin Annotation
	var auditCount, sysLogCount int64
	db.Model(&model.AuditTrail{}).Count(&auditCount)
	db.Model(&model.SystemLog{}).Count(&sysLogCount)
	if auditCount != 0 || sysLogCount != 0 {
		t.Fatalf("expected zero audit_trails/system_logs entries for injection, got audit=%d syslog=%d", auditCount, sysLogCount)
	}

	var debitAggs []model.AggregationResult
	db.Where("device_id = ? AND parameter_id = ? AND source_type = ?", dev.ID, 5, model.SourceCustomerProcessed).
		Order("period_start ASC").
		Find(&debitAggs)
	if len(debitAggs) != 30 {
		t.Fatalf("expected 30 debit aggregation buckets, got %d", len(debitAggs))
	}
	// Without Superadmin annotation (e.g. ADMIN / ENGINEER / OPERATOR), HasInjectedSamples MUST be false
	for _, r := range debitAggs {
		if r.HasInjectedSamples {
			t.Fatalf("expected HasInjectedSamples=false for non-Superadmin query")
		}
	}
	// With Superadmin annotation, buckets 0..14 MUST be true and buckets 15..29 MUST be false!
	injSvc.AnnotateAggregationsForSuperadmin(ctx, debitAggs)
	for idx, r := range debitAggs {
		if idx < 15 && !r.HasInjectedSamples {
			t.Fatalf("expected debit bucket %d to be marked HasInjectedSamples=true for Superadmin", idx)
		}
		if idx >= 15 && r.HasInjectedSamples {
			t.Fatalf("expected untouched debit bucket %d to have HasInjectedSamples=false", idx)
		}
	}

	// Check INT16 scaling & RawHex on pH (scale=0.01 -> raw ~ 690 -> 0x02B2 = "02 B2")
	var phRaw []model.RawData
	db.Where("device_id = ? AND parameter_id = ? AND received_at < ?", dev.ID, 1, winStart.Add(2*time.Minute)).Find(&phRaw)
	if len(phRaw) != 4 {
		t.Fatalf("expected 4 injected pH raw samples in first bucket, got %d", len(phRaw))
	}
	if phRaw[0].RawValue < 680 || phRaw[0].RawValue > 700 || len(phRaw[0].RawHex) != 5 {
		t.Fatalf("expected realistic INT16 scaled RawValue (~690) and 2-byte RawHex on pH, got raw=%v hex=%q", phRaw[0].RawValue, phRaw[0].RawHex)
	}
	injSvc.AnnotateRawDataForSuperadmin(ctx, phRaw)
	if !phRaw[0].IsInjected || phRaw[0].InjectionBatchID != execRes.Batch.BatchCode {
		t.Fatalf("expected IsInjected=true and BatchCode=%s for Superadmin raw query", execRes.Batch.BatchCode)
	}

	// 7. Verify 1-Click Rollback restores exact pre-injection state!
	rbRes, err := injSvc.RollbackBatch(ctx, execRes.Batch.ID, "superadmin")
	if err != nil {
		t.Fatalf("RollbackBatch failed: %v", err)
	}
	if rbRes.DeletedRawCount != 252 {
		t.Fatalf("expected 252 deleted raw rows on rollback, got %d", rbRes.DeletedRawCount)
	}
	if rbRes.Batch.Status != "ROLLED_BACK" {
		t.Fatalf("expected batch status ROLLED_BACK, got %s", rbRes.Batch.Status)
	}

	// Synthetic-only pH buckets 0..11 must be deleted, leaving 18 real buckets (12..29)
	var phAggCountAfterRB int64
	db.Model(&model.AggregationResult{}).
		Where("device_id = ? AND parameter_id = ? AND source_type = ?", dev.ID, 1, model.SourceCustomerProcessed).
		Count(&phAggCountAfterRB)
	if phAggCountAfterRB != 18 {
		t.Fatalf("expected 18 remaining real pH buckets after rollback, got %d", phAggCountAfterRB)
	}

	// And SPARING outbound payload must reflect rolled-back state again (synthetic buckets 0..11 removed, leaving 18 real buckets)
	qItemsRB, _, _ := outSvc.GetQueueItems(ctx, repository.QueueFilter{DestinationID: &dest.ID, Page: 1, PageSize: 10})
	var sparingDocRB struct {
		Data []map[string]interface{} `json:"data"`
	}
	if err := json.Unmarshal([]byte(qItemsRB[0].PayloadSnapshot), &sparingDocRB); err != nil {
		t.Fatalf("invalid SPARING JSON after rollback: %v", err)
	}
	if len(sparingDocRB.Data) != 18 {
		t.Fatalf("expected 18 remaining real buckets in SPARING payload after rollback, got %d", len(sparingDocRB.Data))
	}
	if _, hasDebitInB12 := sparingDocRB.Data[0]["debit"]; hasDebitInB12 {
		t.Fatalf("expected injected debit removed from bucket 12 after rollback, got %v", sparingDocRB.Data[0])
	}
}

func TestTelemetryInjection_DisabledDestination_CrossIntervalRollup_AndResync(t *testing.T) {
	db := setupTelemetryInjectionTestDB(t)
	ctx := context.Background()

	devRepo := repository.NewDeviceRepository(db)
	aggRepo := repository.NewAggregationRepository(db)
	sysRepo := repository.NewSystemRepository(db)
	outRepo := repository.NewOutboundRepository(db)
	injRepo := repository.NewTelemetryInjectionRepository(db)

	aggSvc := service.NewAggregationService(aggRepo, devRepo, sysRepo, nil)
	outSvc := service.NewOutboundService(outRepo, devRepo, service.NewConfigurablePayloadFormatter())
	injSvc := service.NewTelemetryInjectionService(injRepo, devRepo, aggRepo, aggSvc, outSvc)

	// 1. Create com1 device with 1 old disabled AQMS param (1h agg) + 5 active SPARING params (2m CUSTOMER_PROCESSED agg only = 150 buckets/hour)
	dev := &model.Device{
		ID:         45,
		DeviceCode: "com1",
		DeviceName: "COM 1",
		DeviceType: "MODBUS_RTU",
		Timezone:   "Asia/Jakarta",
		Status:     model.DeviceStatusActive,
		Enabled:    true,
	}
	if err := db.Create(dev).Error; err != nil {
		t.Fatalf("create device: %v", err)
	}

	oldPM10 := &model.Parameter{
		ID:            100,
		DeviceID:      dev.ID,
		ParameterCode: "pm10",
		ParameterName: "PM10",
		Unit:          "ug/m3",
		DataType:      model.DataTypeFloat32,
		Scale:         1.0,
		Precision:     2,
		Enabled:       false,
	}
	if err := db.Create(oldPM10).Error; err != nil {
		t.Fatalf("create disabled pm10: %v", err)
	}
	// Disabled old parameter must explicitly stay disabled in GORM (boolean false)
	db.Model(oldPM10).Update("enabled", false)
	_ = db.Create(&model.AggregationDefinition{
		Name:            "PM10 1h",
		Code:            "pm10_1h",
		SourceType:      model.SourceCustomerProcessed,
		DeviceID:        dev.ID,
		ParameterID:     oldPM10.ID,
		Function:        model.FunctionAvg,
		IntervalSeconds: 3600,
		Timezone:        "UTC",
		Enabled:         true,
	}).Error

	sparingParams := []struct {
		id   uint
		code string
		name string
		unit string
		minV float64
		maxV float64
	}{
		{101, "ammonia", "Ammonia (NH3-N)", "mg/L", 6.7, 7.0},
		{102, "cod", "COD", "mg/L", 106.0, 112.0},
		{103, "debit", "Debit", "m3/s", 0.82, 0.88},
		{104, "ph", "PH", "", 6.8, 7.1},
		{105, "tss", "TSS", "mg/L", 28.0, 29.5},
	}
	for _, sp := range sparingParams {
		p := &model.Parameter{
			ID:            sp.id,
			DeviceID:      dev.ID,
			ParameterCode: sp.code,
			ParameterName: sp.name,
			Unit:          sp.unit,
			DataType:      model.DataTypeFloat32,
			Scale:         1.0,
			Precision:     3,
			ByteOrder:     "ABCD",
			Enabled:       true,
		}
		if err := db.Create(p).Error; err != nil {
			t.Fatalf("create param %s: %v", sp.code, err)
		}
		if err := db.Create(&model.AggregationDefinition{
			Name:            sp.code + " 2m",
			Code:            sp.code + "_2m",
			SourceType:      model.SourceCustomerProcessed,
			DeviceID:        dev.ID,
			ParameterID:     sp.id,
			Function:        model.FunctionAvg,
			IntervalSeconds: 120,
			Timezone:        "UTC",
			Enabled:         true,
		}).Error; err != nil {
			t.Fatalf("create 2m agg def: %v", err)
		}
	}

	// 2. Destination #1: CBI (30m = 1800s, ENABLED, stale ParameterMappingsJSON referencing only old disabled pm10)
	destCBI := &model.OutputDestination{
		Name:            "CBI",
		Type:            model.DestREST,
		Endpoint:        "http://127.0.0.1:1/cbi-noop",
		IntervalSeconds: 1800,
		Enabled:         false, // keep false so test never dials network, then flip via DB update to simulate active without background worker
		RetryLimit:      5,
	}
	if err := outSvc.CreateDestination(ctx, destCBI); err != nil {
		t.Fatalf("create CBI dest: %v", err)
	}
	db.Model(destCBI).Update("enabled", true)
	cbiCfg, _ := outSvc.GetPayloadConfig(ctx, destCBI.ID)
	staleCBIMappings, _ := json.Marshal([]model.ParameterMappingItem{
		{
			DeviceID:      dev.ID,
			DeviceCode:    "com1",
			ParameterID:   oldPM10.ID,
			ParameterCode: "pm10",
			ParameterName: "PM10",
			ExternalName:  "pm10",
			Enabled:       true,
		},
	})
	cbiCfg.TemplateType = "MULTI_TICK_SENSOR"
	cbiCfg.Structure = "MULTI_TICK_ARRAY"
	cbiCfg.ParameterMappingsJSON = string(staleCBIMappings)
	if err := outRepo.UpsertPayloadConfig(ctx, cbiCfg); err != nil {
		t.Fatalf("upsert CBI payload cfg: %v", err)
	}

	// 3. Destination #2: KLHK SPARING (DISABLED = false, 2m -> 1h bundle, stale ParameterMappingsJSON with only debit)
	destSparing := &model.OutputDestination{
		Name:                    "KLHK SPARING",
		Type:                    model.DestREST,
		Endpoint:                "https://sparing.example.test/send",
		AuthType:                "SPARING_JWT_STATIC",
		AuthToken:               "secret-key",
		IntervalSeconds:         120,
		DispatchIntervalSeconds: 3600,
		Enabled:                 false,
		RetryLimit:              5,
	}
	if err := outSvc.CreateDestination(ctx, destSparing); err != nil {
		t.Fatalf("create SPARING dest: %v", err)
	}
	// Ensure Enabled is false in DB
	db.Model(destSparing).Update("enabled", false)
	sparingCfg, _ := outSvc.GetPayloadConfig(ctx, destSparing.ID)
	staleSparingMappings, _ := json.Marshal([]model.ParameterMappingItem{
		{
			DeviceID:      dev.ID,
			DeviceCode:    "com1",
			ParameterID:   103,
			ParameterCode: "debit",
			ParameterName: "Debit",
			ExternalName:  "debit",
			Enabled:       true,
		},
	})
	sparingCfg.StationIDField = "uid"
	sparingCfg.StationIDValue = "65c05cac405f557259fed7cc"
	sparingCfg.TemplateType = "KLHK_SPARING_HOURLY"
	sparingCfg.Structure = "SPARING_DATA_ARRAY"
	sparingCfg.TimestampField = "datetime"
	sparingCfg.TimestampFormat = "UNIX_SEC"
	sparingCfg.TimestampSource = "PERIOD_END"
	sparingCfg.ParameterMappingsJSON = string(staleSparingMappings)
	if err := outRepo.UpsertPayloadConfig(ctx, sparingCfg); err != nil {
		t.Fatalf("upsert SPARING payload cfg: %v", err)
	}

	// 4. Verify GetDeviceParameterHints detects DISABLED SPARING (recommended=true, 5/5 matched) & CBI (recommended=false, rollup=true)
	hints, err := injSvc.GetDeviceParameterHints(ctx, dev.ID)
	if err != nil {
		t.Fatalf("GetDeviceParameterHints: %v", err)
	}
	if len(hints.ActiveAggIntervals) != 1 || hints.ActiveAggIntervals[0] != 120 {
		t.Fatalf("expected only active parameter agg interval [120], got %v", hints.ActiveAggIntervals)
	}
	if len(hints.OutboundDestinations) != 2 {
		t.Fatalf("expected 2 outbound destination hints (including DISABLED), got %d", len(hints.OutboundDestinations))
	}
	var cbiHint, sparingHint *service.OutboundDestinationInjectionHint
	for i := range hints.OutboundDestinations {
		h := &hints.OutboundDestinations[i]
		if h.DestinationID == destCBI.ID {
			cbiHint = h
		} else if h.DestinationID == destSparing.ID {
			sparingHint = h
		}
	}
	if sparingHint == nil || !sparingHint.RecommendedSelected || sparingHint.MatchedParamCount != 5 {
		t.Fatalf("expected DISABLED SPARING destination to be recommended with 5/5 matched params, got %+v", sparingHint)
	}
	if cbiHint == nil || cbiHint.RecommendedSelected || !cbiHint.NeedsIntervalRollup {
		t.Fatalf("expected CBI destination to have RecommendedSelected=false and NeedsIntervalRollup=true, got %+v", cbiHint)
	}

	// 5. Inject 1 August 2026 00:00:00 .. 01:00:00 Asia/Jakarta (1800 samples, 10s step) targeting recommended SPARING destination
	wib := time.FixedZone("WIB", 7*3600)
	aug1Start := time.Date(2026, 8, 1, 0, 0, 0, 0, wib)
	aug1End := time.Date(2026, 8, 1, 1, 0, 0, 0, wib)
	var inputs []service.TelemetryInjectionParamInput
	for _, sp := range sparingParams {
		inputs = append(inputs, service.TelemetryInjectionParamInput{
			ParameterID: sp.id,
			MinValue:    sp.minV,
			MaxValue:    sp.maxV,
		})
	}

	refreshFalse := false
	execRes, err := injSvc.ExecuteInjection(ctx, service.TelemetryInjectionRequest{
		DeviceID:               dev.ID,
		StartTime:              aug1Start,
		EndTime:                aug1End,
		SampleStepSeconds:      10,
		ConflictMode:           service.ConflictFillMissingAndNull,
		RandomMode:             service.RandomModeSmoothWalk,
		ValueTarget:            service.ValueTargetProcessed,
		RefreshOutbound:        &refreshFalse, // even if refresh_outbound=false, enqueue_missing_outbound=true MUST still enqueue!
		EnqueueMissingOutbound: true,
		OutboundDestinationIDs: []uint{destSparing.ID},
		Seed:                   77,
		Parameters:             inputs,
	}, "admin")
	if err != nil {
		t.Fatalf("ExecuteInjection failed: %v", err)
	}
	if execRes.Batch.InjectedRawCount != 1800 || execRes.Batch.AggregatedBuckets != 150 {
		t.Fatalf("expected 1800 raw & 150 agg buckets, got raw=%d agg=%d", execRes.Batch.InjectedRawCount, execRes.Batch.AggregatedBuckets)
	}
	if execRes.Batch.RefreshedOutbound != 1 {
		t.Fatalf("expected 1 enqueued SPARING batch even though destination is DISABLED, got %d (%+v)", execRes.Batch.RefreshedOutbound, execRes.OutboundResult)
	}

	// Worker must NOT dispatch PENDING item while destination is DISABLED
	outSvc.ProcessPendingQueueOnce()

	dayStartUTC := time.Date(2026, 8, 1, 0, 0, 0, 0, wib).UTC()
	dayEndUTC := time.Date(2026, 8, 2, 0, 0, 0, 0, wib).UTC()
	augItems, augTotal, err := outSvc.GetQueueItems(ctx, repository.QueueFilter{
		WindowStart: &dayStartUTC,
		WindowEnd:   &dayEndUTC,
		Page:        1,
		PageSize:    20,
	})
	if err != nil || augTotal != 1 || len(augItems) != 1 {
		t.Fatalf("expected 1 queue item on 2026-08-01 via WindowStart/WindowEnd filter, got total=%d err=%v", augTotal, err)
	}
	if augItems[0].Status != model.DeliveryPending || augItems[0].Attempts != 0 {
		t.Fatalf("expected DISABLED destination queue item to remain PENDING (attempts=0), got status=%s attempts=%d", augItems[0].Status, augItems[0].Attempts)
	}
	expectedLocalBatchID := fmt.Sprintf("BATCH-DEST%d-60M-2M-20260801010000", destSparing.ID)
	if augItems[0].BatchID != expectedLocalBatchID {
		t.Fatalf("expected BatchID in Asia/Jakarta local time %q (not UTC 20260731180000), got %q", expectedLocalBatchID, augItems[0].BatchID)
	}
	expectedIdempKey := fmt.Sprintf("DEST%d_DISP3600_SRC120_%d", destSparing.ID, aug1End.Unix())
	if augItems[0].IdempotencyKey != expectedIdempKey {
		t.Fatalf("expected UTC epoch IdempotencyKey %q, got %q", expectedIdempKey, augItems[0].IdempotencyKey)
	}

	var sparingPayload struct {
		UID  string                   `json:"uid"`
		Data []map[string]interface{} `json:"data"`
	}
	if err := json.Unmarshal([]byte(augItems[0].PayloadSnapshot), &sparingPayload); err != nil {
		t.Fatalf("unmarshal SPARING payload: %v", err)
	}
	if len(sparingPayload.Data) != 30 {
		t.Fatalf("expected 30 SPARING 2-minute slots on Aug 1, got %d", len(sparingPayload.Data))
	}
	for _, key := range []string{"pH", "cod", "tss", "nh3n", "debit"} {
		if sparingPayload.Data[0][key] == nil {
			t.Fatalf("expected auto-mapped unseen SPARING key %q to be non-nil in slot 0, got %v", key, sparingPayload.Data[0])
		}
	}

	// Simulate legacy UTC-formatted BatchID on existing PENDING row and verify SyncBatchOutbound repairs it in-place
	legacyUTCBatchID := fmt.Sprintf("BATCH-DEST%d-60M-2M-20260731180000", destSparing.ID)
	db.Model(&model.OutboundDeliveryQueue{}).Where("id = ?", augItems[0].ID).Update("batch_id", legacyUTCBatchID)

	// 6. Now test 1-Click SyncBatchOutbound explicitly including 30m CBI destination (verifies 2m -> 30m cross-interval rollup!)
	// Keep CBI disabled before syncing so no HTTP attempt alters status during assertion
	db.Model(destCBI).Update("enabled", false)
	syncRes, err := injSvc.SyncBatchOutbound(ctx, execRes.Batch.ID, []uint{destCBI.ID, destSparing.ID})
	if err != nil {
		t.Fatalf("SyncBatchOutbound failed: %v", err)
	}
	// Expect 2 new 30m CBI batches (00:00->00:30 and 00:30->01:00) + 1 updated SPARING batch = 3 total touched!
	if syncRes.OutboundResult.EnqueuedBatches != 2 || syncRes.OutboundResult.UpdatedBatches != 1 {
		t.Fatalf("expected 2 new CBI 30m rolled-up batches + 1 updated SPARING batch, got %+v", syncRes.OutboundResult)
	}

	repairedSparing, err := outRepo.GetQueueItemByID(ctx, augItems[0].ID)
	if err != nil || repairedSparing.BatchID != expectedLocalBatchID {
		t.Fatalf("expected SyncBatchOutbound to repair legacy UTC BatchID to %q, got %+v (err=%v)", expectedLocalBatchID, repairedSparing, err)
	}

	cbiItems, cbiTotal, _ := outSvc.GetQueueItems(ctx, repository.QueueFilter{
		DestinationID: &destCBI.ID,
		WindowStart:   &dayStartUTC,
		WindowEnd:     &dayEndUTC,
		Page:          1,
		PageSize:      20,
	})
	if cbiTotal != 2 || len(cbiItems) != 2 {
		t.Fatalf("expected 2 rolled-up 30m CBI queue batches on 2026-08-01, got %d", cbiTotal)
	}
	if cbiItems[0].ParameterCount != 5 || !strings.Contains(cbiItems[0].PayloadSnapshot, "ammonia") {
		t.Fatalf("expected 5 rolled-up parameters in CBI 30m batch, got count=%d payload=%s", cbiItems[0].ParameterCount, cbiItems[0].PayloadSnapshot)
	}
}

func TestTelemetryInjection_ConcurrentWorkerContention_Guaranteed150Of150_And149Recovery(t *testing.T) {
	db := setupTelemetryInjectionTestDB(t)
	ctx := context.Background()

	devRepo := repository.NewDeviceRepository(db)
	aggRepo := repository.NewAggregationRepository(db)
	sysRepo := repository.NewSystemRepository(db)
	outRepo := repository.NewOutboundRepository(db)
	injRepo := repository.NewTelemetryInjectionRepository(db)

	aggSvc := service.NewAggregationService(aggRepo, devRepo, sysRepo, nil)
	outSvc := service.NewOutboundService(outRepo, devRepo, service.NewConfigurablePayloadFormatter())
	injSvc := service.NewTelemetryInjectionService(injRepo, devRepo, aggRepo, aggSvc, outSvc)

	dev := &model.Device{
		ID:         45,
		DeviceCode: "com1",
		DeviceName: "COM 1",
		DeviceType: "MODBUS_RTU",
		Timezone:   "Asia/Jakarta",
		Status:     model.DeviceStatusActive,
		Enabled:    true,
	}
	if err := db.Create(dev).Error; err != nil {
		t.Fatalf("create device: %v", err)
	}

	sparingParams := []struct {
		id   uint
		code string
		name string
		unit string
		minV float64
		maxV float64
	}{
		{101, "ammonia", "Ammonia (NH3-N)", "mg/L", 6.75, 7.02},
		{102, "cod", "COD", "mg/L", 107.0, 112.0},
		{103, "debit", "Debit", "-", 0.83, 0.87},
		{104, "ph", "PH", "-", 6.76, 7.02},
		{105, "tss", "TSS", "mg/L", 28.2, 29.0},
	}
	for _, sp := range sparingParams {
		if err := db.Create(&model.Parameter{
			ID:            sp.id,
			DeviceID:      dev.ID,
			ParameterCode: sp.code,
			ParameterName: sp.name,
			Unit:          sp.unit,
			DataType:      model.DataTypeFloat32,
			Scale:         1.0,
			Precision:     2,
			ByteOrder:     "ABCD",
			Enabled:       true,
		}).Error; err != nil {
			t.Fatalf("create param %s: %v", sp.code, err)
		}
		if err := db.Create(&model.AggregationDefinition{
			Name:            sp.code + " 2m",
			Code:            sp.code + "_2m",
			SourceType:      model.SourceCustomerProcessed,
			DeviceID:        dev.ID,
			ParameterID:     sp.id,
			Function:        model.FunctionAvg,
			IntervalSeconds: 120,
			Timezone:        "Asia/Jakarta",
			Enabled:         true,
		}).Error; err != nil {
			t.Fatalf("create 2m agg def: %v", err)
		}
	}

	// Seed recent real-time telemetry so the background worker & GetResults actively upsert into aggregation_results concurrently!
	recentNow := time.Now().UTC().Truncate(2 * time.Minute)
	for _, sp := range sparingParams {
		v := (sp.minV + sp.maxV) / 2
		_ = db.Create(&model.RawData{
			DeviceID:       dev.ID,
			ParameterID:    sp.id,
			Value:          v,
			ProcessedValue: v,
			ValueNumeric:   &v,
			RawValue:       v,
			Quality:        model.QualityGood,
			ReceivedAt:     recentNow.Add(-90 * time.Second),
			StoredAt:       recentNow.Add(-90 * time.Second),
			Timestamp:      recentNow.Add(-90 * time.Second),
		}).Error
	}

	// Register an onAggregationCompleted hook that queries the DB (simulating ISPU/WQMS listener load)
	aggSvc.SetOnAggregationCompleted(func(cbCtx context.Context, ev service.AggregationCompletedEvent) {
		var cnt int64
		db.WithContext(cbCtx).Model(&model.AggregationResult{}).Where("device_id = ?", ev.DeviceID).Count(&cnt)
	})

	destSparing := &model.OutputDestination{
		Name:                    "KLH Sparing",
		Type:                    model.DestREST,
		Endpoint:                "https://sparing.example.test/send",
		AuthType:                "SPARING_JWT_STATIC",
		AuthToken:               "secret-key",
		IntervalSeconds:         120,
		DispatchIntervalSeconds: 3600,
		Enabled:                 false,
		RetryLimit:              5,
	}
	if err := outSvc.CreateDestination(ctx, destSparing); err != nil {
		t.Fatalf("create SPARING dest: %v", err)
	}
	db.Model(destSparing).Update("enabled", false)
	sparingCfg, _ := outSvc.GetPayloadConfig(ctx, destSparing.ID)
	sparingCfg.StationIDField = "uid"
	sparingCfg.StationIDValue = "65c05cac405f557259fed7cc"
	sparingCfg.TemplateType = "KLHK_SPARING_HOURLY"
	sparingCfg.Structure = "SPARING_DATA_ARRAY"
	sparingCfg.TimestampField = "datetime"
	sparingCfg.TimestampFormat = "UNIX_SEC"
	sparingCfg.TimestampSource = "PERIOD_START"
	if err := outRepo.UpsertPayloadConfig(ctx, sparingCfg); err != nil {
		t.Fatalf("upsert SPARING payload cfg: %v", err)
	}

	// Start background worker + concurrent GetResults hammer goroutines to provoke maximum DB contention
	aggSvc.StartWorker()
	defer aggSvc.StopWorker()

	stopHammer := make(chan struct{})
	hammerDone := make(chan struct{}, 2)
	for g := 0; g < 2; g++ {
		go func() {
			defer func() { hammerDone <- struct{}{} }()
			for {
				select {
				case <-stopHammer:
					return
				default:
					_, _, _ = aggSvc.GetResults(ctx, repository.AggregationResultFilter{
						DeviceID: dev.ID,
						Page:     1,
						PageSize: 20,
					})
					time.Sleep(5 * time.Millisecond)
				}
			}
		}()
	}

	wib := time.FixedZone("WIB", 7*3600)
	winStart := time.Date(2026, 8, 1, 2, 0, 0, 0, wib)
	winEnd := time.Date(2026, 8, 1, 3, 0, 0, 0, wib)
	var inputs []service.TelemetryInjectionParamInput
	for _, sp := range sparingParams {
		inputs = append(inputs, service.TelemetryInjectionParamInput{
			ParameterID: sp.id,
			MinValue:    sp.minV,
			MaxValue:    sp.maxV,
		})
	}

	execRes, err := injSvc.ExecuteInjection(ctx, service.TelemetryInjectionRequest{
		DeviceID:               dev.ID,
		StartTime:              winStart,
		EndTime:                winEnd,
		SampleStepSeconds:      10,
		ConflictMode:           service.ConflictFillMissingAndNull,
		RandomMode:             service.RandomModeSmoothWalk,
		ValueTarget:            service.ValueTargetProcessed,
		EnqueueMissingOutbound: true,
		OutboundDestinationIDs: []uint{destSparing.ID},
		Seed:                   101,
		Parameters:             inputs,
	}, "admin")
	close(stopHammer)
	<-hammerDone
	<-hammerDone

	if err != nil {
		t.Fatalf("ExecuteInjection under concurrent worker contention failed: %v", err)
	}
	if execRes.Batch.InjectedRawCount != 1800 || execRes.Batch.AggregatedBuckets != 150 {
		t.Fatalf("expected 1800 raw & 150/150 agg buckets, got raw=%d agg=%d", execRes.Batch.InjectedRawCount, execRes.Batch.AggregatedBuckets)
	}

	qItems, qTotal, err := outSvc.GetQueueItems(ctx, repository.QueueFilter{
		DestinationID: &destSparing.ID,
		Page:          1,
		PageSize:      10,
	})
	if err != nil || qTotal != 1 {
		t.Fatalf("expected 1 SPARING queue item, got total=%d err=%v", qTotal, err)
	}
	if qItems[0].ParameterCount != 150 {
		t.Fatalf("expected full 150/150 persisted source aggregation records in queue item, got %d", qItems[0].ParameterCount)
	}

	var doc struct {
		UID  string                   `json:"uid"`
		Data []map[string]interface{} `json:"data"`
	}
	if err := json.Unmarshal([]byte(qItems[0].PayloadSnapshot), &doc); err != nil {
		t.Fatalf("unmarshal SPARING payload: %v", err)
	}
	if len(doc.Data) != 30 {
		t.Fatalf("expected 30 SPARING slots, got %d", len(doc.Data))
	}
	for idx, slot := range doc.Data {
		for _, k := range []string{"pH", "cod", "tss", "nh3n", "debit"} {
			if slot[k] == nil {
				t.Fatalf("slot #%d (datetime=%v) missing key %q: %v", idx, slot["datetime"], k, slot)
			}
		}
	}

	// Simulate pre-fix 149-record state (missing ammonia bucket #22 at 02:42:00 WIB = unix 1785526920)
	bucket22Start := winStart.Add(21 * 2 * time.Minute).UTC()
	if err := db.Where("device_id = ? AND parameter_id = ? AND period_start = ?", dev.ID, 101, bucket22Start).
		Delete(&model.AggregationResult{}).Error; err != nil {
		t.Fatalf("delete bucket 22 ammonia: %v", err)
	}
	if _, err := outSvc.RefreshQueueForWindowDetailed(ctx, service.OutboundWindowRefreshOptions{
		DeviceID:                    dev.ID,
		StartTime:                   winStart.UTC(),
		EndTime:                     winEnd.UTC(),
		RefreshExisting:             true,
		IncludeDisabledDestinations: true,
		DestinationIDs:              []uint{destSparing.ID},
	}); err != nil {
		t.Fatalf("refresh degraded 149 queue: %v", err)
	}
	degradedItem, _ := outRepo.GetQueueItemByID(ctx, qItems[0].ID)
	if degradedItem.ParameterCount != 149 {
		t.Fatalf("expected simulated pre-fix degraded queue item to have 149 records, got %d", degradedItem.ParameterCount)
	}

	// Verify SyncBatchOutbound recalculates missing bucket #22 from persisted raw_data and restores 149 -> 150 in-place!
	syncRes, err := injSvc.SyncBatchOutbound(ctx, execRes.Batch.ID, []uint{destSparing.ID})
	if err != nil {
		t.Fatalf("SyncBatchOutbound recovery failed: %v", err)
	}
	if syncRes.RecalculatedBuckets != 150 {
		t.Fatalf("expected SyncBatchOutbound to recalculate all 150 buckets, got %d", syncRes.RecalculatedBuckets)
	}
	recoveredItem, _ := outRepo.GetQueueItemByID(ctx, qItems[0].ID)
	if recoveredItem.ParameterCount != 150 {
		t.Fatalf("expected SyncBatchOutbound to restore queue item from 149 -> 150 records, got %d", recoveredItem.ParameterCount)
	}
	var recoveredDoc struct {
		Data []map[string]interface{} `json:"data"`
	}
	_ = json.Unmarshal([]byte(recoveredItem.PayloadSnapshot), &recoveredDoc)
	if recoveredDoc.Data[21]["nh3n"] == nil {
		t.Fatalf("expected bucket #22 (datetime=%v) nh3n restored in SPARING payload, got %v", recoveredDoc.Data[21]["datetime"], recoveredDoc.Data[21])
	}
}


