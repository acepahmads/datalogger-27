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
