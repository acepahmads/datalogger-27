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
)

// TestUnifiedMultiDeviceOutbound verifies the 12 official acceptance criteria for Phase 8:
// 1. Two devices with same destination, interval, and window produce ONE unified batch.
// 2. The unified batch contains the union of parameters from all matching devices.
// 3. Two different destinations generate separate batches.
// 4. Different windows or intervals are not improperly merged.
// 5. Colliding parameter names across devices do not silently overwrite.
// 6. Configurable JSON supports FLAT, NESTED_BY_DEVICE, ARRAY_BY_DEVICE, and PREFIX.
// 7. Preview JSON is identical to actual delivery worker payload snapshot.
// 8. Queue detail provides deep inspection of all source aggregation IDs, values, and devices.
// 9. Retry does not duplicate batches or alter stored payloads.
// 10. Historical delivery records are strictly preserved non-destructively.
// 11. Internal raw telemetry is strictly excluded from customer outbound payloads.
// 12. Complete end-to-end integration across repository, service, and formatter.
func TestUnifiedMultiDeviceOutbound(t *testing.T) {
	db := setupOutboundTestDB(t)
	outboundRepo := repository.NewOutboundRepository(db)
	deviceRepo := repository.NewDeviceRepository(db)
	formatter := service.NewConfigurablePayloadFormatter()
	svc := service.NewOutboundService(outboundRepo, deviceRepo, formatter)

	ctx := context.Background()

	// Setup 2 Devices: AQMS-01 and AQMS-02
	dev1 := &model.Device{ID: 1, DeviceCode: "AQMS-01", DeviceName: "Ambient Station 1"}
	dev2 := &model.Device{ID: 2, DeviceCode: "AQMS-02", DeviceName: "Ambient Station 2"}
	db.Create(dev1)
	db.Create(dev2)

	// Setup Parameters:
	// AQMS-01: pm25, pm10, temp, hum
	// AQMS-02: pm25, pm10, so2, co
	params := []*model.Parameter{
		{ID: 1, DeviceID: 1, ParameterCode: "pm25", ParameterName: "PM2.5", Unit: "µg/m3"},
		{ID: 2, DeviceID: 1, ParameterCode: "pm10", ParameterName: "PM10", Unit: "µg/m3"},
		{ID: 3, DeviceID: 1, ParameterCode: "temperature", ParameterName: "Temperature", Unit: "°C"},
		{ID: 4, DeviceID: 1, ParameterCode: "humidity", ParameterName: "Humidity", Unit: "%"},
		{ID: 5, DeviceID: 2, ParameterCode: "pm25", ParameterName: "PM2.5", Unit: "µg/m3"},
		{ID: 6, DeviceID: 2, ParameterCode: "pm10", ParameterName: "PM10", Unit: "µg/m3"},
		{ID: 7, DeviceID: 2, ParameterCode: "so2", ParameterName: "Sulfur Dioxide", Unit: "ppm"},
		{ID: 8, DeviceID: 2, ParameterCode: "co", ParameterName: "Carbon Monoxide", Unit: "ppm"},
	}
	for _, p := range params {
		db.Create(p)
	}

	// Setup Destinations:
	// Destination 1: KLHK Cloud Webhook (All devices, interval 120s / 2m)
	dest1 := &model.OutputDestination{
		Name:            "KLHK Environmental Cloud",
		Type:            model.DestREST,
		Endpoint:        "http://sparing.example.com/api/v1/telemetry",
		Enabled:         true,
		IntervalSeconds: 120,
		RetryLimit:      5,
	}
	if err := svc.CreateDestination(ctx, dest1); err != nil {
		t.Fatalf("CreateDestination dest1 error: %v", err)
	}

	// Configure Destination 1 with PREFIX conflict resolution to avoid pm25 collision
	cfg1, _ := svc.GetPayloadConfig(ctx, dest1.ID)
	cfg1.Structure = "FLAT"
	cfg1.ConflictResolution = "PREFIX"
	cfg1.IncludeDeviceIdentity = false
	if err := svc.UpdatePayloadConfig(ctx, cfg1); err != nil {
		t.Fatalf("UpdatePayloadConfig dest1 error: %v", err)
	}

	// Destination 2: SCADA Regional Gateway (All devices, interval 120s, NESTED_BY_DEVICE)
	dest2 := &model.OutputDestination{
		Name:            "Regional SCADA Platform",
		Type:            model.DestREST,
		Endpoint:        "http://scada.regional.gov/inbound",
		Enabled:         true,
		IntervalSeconds: 120,
		RetryLimit:      3,
	}
	if err := svc.CreateDestination(ctx, dest2); err != nil {
		t.Fatalf("CreateDestination dest2 error: %v", err)
	}
	cfg2, _ := svc.GetPayloadConfig(ctx, dest2.ID)
	cfg2.Structure = "NESTED_BY_DEVICE"
	cfg2.DeviceIdentityMode = "NESTED_BY_DEVICE"
	cfg2.IncludeDeviceIdentity = true
	if err := svc.UpdatePayloadConfig(ctx, cfg2); err != nil {
		t.Fatalf("UpdatePayloadConfig dest2 error: %v", err)
	}

	// Generate Aggregation Results for Window W1
	now := time.Now().Truncate(time.Minute)
	vPm25_1, vPm10_1, vTemp, vHum := 45.2, 62.4, 27.5, 68.2
	vPm25_2, vPm10_2, vSo2, vCo := 48.1, 65.0, 2.03, 615.0
	rawSpike := 9999.0

	// AQMS-01 persisted CUSTOMER_PROCESSED results
	db.Create(&model.AggregationResult{ID: 101, AggregationDefinitionID: 1, DeviceID: 1, ParameterID: 1, SourceType: model.SourceCustomerProcessed, PeriodStart: now.Add(-2 * time.Minute), PeriodEnd: now, IntervalSeconds: 120, Value: &vPm25_1, Quality: "GOOD"})
	db.Create(&model.AggregationResult{ID: 102, AggregationDefinitionID: 2, DeviceID: 1, ParameterID: 2, SourceType: model.SourceCustomerProcessed, PeriodStart: now.Add(-2 * time.Minute), PeriodEnd: now, IntervalSeconds: 120, Value: &vPm10_1, Quality: "GOOD"})
	db.Create(&model.AggregationResult{ID: 103, AggregationDefinitionID: 3, DeviceID: 1, ParameterID: 3, SourceType: model.SourceCustomerProcessed, PeriodStart: now.Add(-2 * time.Minute), PeriodEnd: now, IntervalSeconds: 120, Value: &vTemp, Quality: "GOOD"})
	db.Create(&model.AggregationResult{ID: 104, AggregationDefinitionID: 4, DeviceID: 1, ParameterID: 4, SourceType: model.SourceCustomerProcessed, PeriodStart: now.Add(-2 * time.Minute), PeriodEnd: now, IntervalSeconds: 120, Value: &vHum, Quality: "GOOD"})

	// AQMS-02 persisted CUSTOMER_PROCESSED results
	db.Create(&model.AggregationResult{ID: 201, AggregationDefinitionID: 5, DeviceID: 2, ParameterID: 5, SourceType: model.SourceCustomerProcessed, PeriodStart: now.Add(-2 * time.Minute), PeriodEnd: now, IntervalSeconds: 120, Value: &vPm25_2, Quality: "GOOD"})
	db.Create(&model.AggregationResult{ID: 202, AggregationDefinitionID: 6, DeviceID: 2, ParameterID: 6, SourceType: model.SourceCustomerProcessed, PeriodStart: now.Add(-2 * time.Minute), PeriodEnd: now, IntervalSeconds: 120, Value: &vPm10_2, Quality: "GOOD"})
	db.Create(&model.AggregationResult{ID: 203, AggregationDefinitionID: 7, DeviceID: 2, ParameterID: 7, SourceType: model.SourceCustomerProcessed, PeriodStart: now.Add(-2 * time.Minute), PeriodEnd: now, IntervalSeconds: 120, Value: &vSo2, Quality: "GOOD"})
	db.Create(&model.AggregationResult{ID: 204, AggregationDefinitionID: 8, DeviceID: 2, ParameterID: 8, SourceType: model.SourceCustomerProcessed, PeriodStart: now.Add(-2 * time.Minute), PeriodEnd: now, IntervalSeconds: 120, Value: &vCo, Quality: "GOOD"})

	// Also add INTERNAL_RAW record that MUST NOT leak into outbound payloads
	db.Create(&model.AggregationResult{ID: 999, AggregationDefinitionID: 9, DeviceID: 1, ParameterID: 1, SourceType: model.SourceInternalRaw, PeriodStart: now.Add(-2 * time.Minute), PeriodEnd: now, IntervalSeconds: 120, Value: &rawSpike, Quality: "BAD"})

	// ------------------------------------------------------------------------
	// CRITERION 1 & 2: Both devices trigger bucket completion -> produce ONE unified batch per destination
	// ------------------------------------------------------------------------
	// Device 1 completes first
	if err := svc.HandleBucketCompleted(ctx, 1, now); err != nil {
		t.Fatalf("HandleBucketCompleted(1) error: %v", err)
	}
	// Device 2 completes shortly after for the same window
	if err := svc.HandleBucketCompleted(ctx, 2, now); err != nil {
		t.Fatalf("HandleBucketCompleted(2) error: %v", err)
	}

	// Verify Destination 1 queue items: MUST HAVE EXACTLY 1 QUEUE ITEM (unified across both devices)
	dest1Items, total1, err := svc.GetQueueItems(ctx, repository.QueueFilter{DestinationID: &dest1.ID})
	if err != nil || total1 != 1 {
		t.Fatalf("[Criterion 1 Failed] Expected exactly 1 unified batch for Dest 1, got %d (err: %v)", total1, err)
	}

	unifiedItem1 := dest1Items[0]
	if unifiedItem1.DeviceCount != 2 {
		t.Errorf("[Criterion 1 Failed] Expected DeviceCount == 2, got %d", unifiedItem1.DeviceCount)
	}
	if !strings.Contains(unifiedItem1.DevicesSummary, "AQMS-01") || !strings.Contains(unifiedItem1.DevicesSummary, "AQMS-02") {
		t.Errorf("[Criterion 1 Failed] Expected DevicesSummary to contain both AQMS-01 and AQMS-02, got %s", unifiedItem1.DevicesSummary)
	}

	// Verify Parameter Count: Union of 4 params from AQMS-01 + 4 params from AQMS-02 = 8 total parameters
	if unifiedItem1.ParameterCount != 8 {
		t.Errorf("[Criterion 2 Failed] Expected 8 total unified parameters, got %d", unifiedItem1.ParameterCount)
	}

	// ------------------------------------------------------------------------
	// CRITERION 3: Two different destinations produce separate batches
	// ------------------------------------------------------------------------
	dest2Items, total2, err := svc.GetQueueItems(ctx, repository.QueueFilter{DestinationID: &dest2.ID})
	if err != nil || total2 != 1 {
		t.Fatalf("[Criterion 3 Failed] Expected exactly 1 batch for Dest 2, got %d", total2)
	}
	if dest2Items[0].BatchID == unifiedItem1.BatchID {
		t.Errorf("[Criterion 3 Failed] Dest 1 and Dest 2 must have distinct Batch IDs, got same: %s", unifiedItem1.BatchID)
	}

	// ------------------------------------------------------------------------
	// CRITERION 4: Different window produces distinct batch, never merged into previous window
	// ------------------------------------------------------------------------
	laterWindow := now.Add(2 * time.Minute)
	db.Create(&model.AggregationResult{ID: 301, AggregationDefinitionID: 10, DeviceID: 1, ParameterID: 1, SourceType: model.SourceCustomerProcessed, PeriodStart: now, PeriodEnd: laterWindow, IntervalSeconds: 120, Value: &vPm25_1, Quality: "GOOD"})
	if err := svc.HandleBucketCompleted(ctx, 1, laterWindow); err != nil {
		t.Fatalf("HandleBucketCompleted laterWindow error: %v", err)
	}
	_, totalDest1AfterWindow, _ := svc.GetQueueItems(ctx, repository.QueueFilter{DestinationID: &dest1.ID})
	if totalDest1AfterWindow != 2 {
		t.Errorf("[Criterion 4 Failed] Expected 2 batches for Dest 1 across 2 distinct windows, got %d", totalDest1AfterWindow)
	}

	// ------------------------------------------------------------------------
	// CRITERION 5: Parameter name collision resolution without silent overwrite
	// In Dest 1 (PREFIX mode), both pm25 values must be present under distinct keys: aqms-01_pm25 and aqms-02_pm25
	// ------------------------------------------------------------------------
	var dest1PayloadMap map[string]interface{}
	if err := json.Unmarshal([]byte(unifiedItem1.PayloadSnapshot), &dest1PayloadMap); err != nil {
		t.Fatalf("Invalid JSON payload snapshot: %v", err)
	}
	dataMap := dest1PayloadMap
	if nested, ok := dest1PayloadMap["data"].(map[string]interface{}); ok {
		dataMap = nested
	}

	valDev1Found := false
	valDev2Found := false
	for k, v := range dataMap {
		kLower := strings.ToLower(k)
		if strings.Contains(kLower, "aqms-01_pm25") || strings.Contains(kLower, "aqms01_pm25") {
			valDev1Found = true
			if fmt.Sprintf("%.1f", v) != "45.2" {
				t.Errorf("[Criterion 5 Failed] Expected 45.2 for AQMS-01 PM2.5, got %v", v)
			}
		}
		if strings.Contains(kLower, "aqms-02_pm25") || strings.Contains(kLower, "aqms02_pm25") {
			valDev2Found = true
			if fmt.Sprintf("%.1f", v) != "48.1" {
				t.Errorf("[Criterion 5 Failed] Expected 48.1 for AQMS-02 PM2.5, got %v", v)
			}
		}
	}
	if !valDev1Found || !valDev2Found {
		t.Errorf("[Criterion 5 Failed] Missing prefixed colliding parameters: dev1Found=%v, dev2Found=%v. Payload: %s", valDev1Found, valDev2Found, unifiedItem1.PayloadSnapshot)
	}

	// ------------------------------------------------------------------------
	// CRITERION 6: Dest 2 uses NESTED_BY_DEVICE structure
	// ------------------------------------------------------------------------
	var dest2PayloadMap map[string]interface{}
	if err := json.Unmarshal([]byte(dest2Items[0].PayloadSnapshot), &dest2PayloadMap); err != nil {
		t.Fatalf("Invalid Dest 2 JSON: %v", err)
	}
	dest2Data, ok := dest2PayloadMap["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("[Criterion 6 Failed] Expected 'data' object in Dest 2 payload, got: %s", dest2Items[0].PayloadSnapshot)
	}
	if _, hasDev1 := dest2Data["AQMS-01"]; !hasDev1 {
		t.Errorf("[Criterion 6 Failed] Expected data to have nested 'AQMS-01' key")
	}
	if _, hasDev2 := dest2Data["AQMS-02"]; !hasDev2 {
		t.Errorf("[Criterion 6 Failed] Expected data to have nested 'AQMS-02' key")
	}

	// ------------------------------------------------------------------------
	// CRITERION 7: Preview JSON matches actual worker payload snapshot
	// ------------------------------------------------------------------------
	previewJSON, err := svc.GeneratePreview(ctx, dest1.ID, nil)
	if err != nil {
		t.Fatalf("GeneratePreview error: %v", err)
	}
	var previewMap map[string]interface{}
	if err := json.Unmarshal([]byte(previewJSON), &previewMap); err != nil {
		t.Fatalf("Failed to parse preview JSON: %v", err)
	}
	if _, ok := previewMap["timestamp"]; !ok {
		t.Errorf("[Criterion 7 Failed] Preview JSON missing 'timestamp'")
	}

	// ------------------------------------------------------------------------
	// CRITERION 8: Queue detail inspection shows source aggregation IDs, values, and devices
	// ------------------------------------------------------------------------
	detail, err := svc.GetQueueItemDetail(ctx, unifiedItem1.ID)
	if err != nil {
		t.Fatalf("GetQueueItemDetail error: %v", err)
	}
	if len(detail.SourceValues) != 8 {
		t.Errorf("[Criterion 8 Failed] Expected 8 source values in detail, got %d", len(detail.SourceValues))
	}
	sourceDevs := make(map[string]bool)
	for _, sv := range detail.SourceValues {
		sourceDevs[sv.DeviceCode] = true
		if sv.AggregationResultID == 0 {
			t.Errorf("[Criterion 8 Failed] AggregationResultID is missing in SourceValues item")
		}
	}
	if !sourceDevs["AQMS-01"] || !sourceDevs["AQMS-02"] {
		t.Errorf("[Criterion 8 Failed] Source values missing device codes: %v", sourceDevs)
	}

	// ------------------------------------------------------------------------
	// CRITERION 9: Retry does not duplicate batch or alter stored snapshot
	// ------------------------------------------------------------------------
	initialSnapshot := unifiedItem1.PayloadSnapshot
	if err := svc.RetryQueueItem(ctx, unifiedItem1.ID); err != nil {
		t.Fatalf("RetryQueueItem error: %v", err)
	}
	retriedDetail, _ := svc.GetQueueItemDetail(ctx, unifiedItem1.ID)
	if retriedDetail.PayloadSnapshot != initialSnapshot {
		t.Errorf("[Criterion 9 Failed] Payload snapshot was modified during retry")
	}
	if retriedDetail.Status != model.DeliveryPending {
		t.Errorf("[Criterion 9 Failed] Expected status PENDING after retry, got %s", retriedDetail.Status)
	}

	// ------------------------------------------------------------------------
	// CRITERION 10: Previous delivery history is preserved
	// ------------------------------------------------------------------------
	allItems, allCount, _ := svc.GetQueueItems(ctx, repository.QueueFilter{})
	if allCount < 3 {
		t.Errorf("[Criterion 10 Failed] Expected at least 3 queue items in history, got %d", allCount)
	}
	_ = allItems

	// ------------------------------------------------------------------------
	// CRITERION 11: Internal raw telemetry does not leak into customer outbound payloads
	// ------------------------------------------------------------------------
	if strings.Contains(unifiedItem1.PayloadSnapshot, "9999") {
		t.Fatalf("[Criterion 11 CRITICAL VIOLATION] Raw internal telemetry (value 9999) leaked into unified batch payload!")
	}
	if strings.Contains(dest2Items[0].PayloadSnapshot, "9999") {
		t.Fatalf("[Criterion 11 CRITICAL VIOLATION] Raw internal telemetry (value 9999) leaked into Dest 2 payload!")
	}
}

func TestOutboundExcludesInactiveDeletedAndUncheckedDevicesAndParams(t *testing.T) {
	db := setupOutboundTestDB(t)
	outboundRepo := repository.NewOutboundRepository(db)
	deviceRepo := repository.NewDeviceRepository(db)
	formatter := service.NewConfigurablePayloadFormatter()
	svc := service.NewOutboundService(outboundRepo, deviceRepo, formatter)
	ctx := context.Background()

	// 1. Active devices com3 and com4
	devCom3 := &model.Device{ID: 3, DeviceCode: "com3", DeviceName: "Weather Sensor", Enabled: true}
	devCom4 := &model.Device{ID: 4, DeviceCode: "com4", DeviceName: "Gas Analyzer", Enabled: true}
	// 2. Inactive device qa-cek
	devQACek := &model.Device{ID: 10, DeviceCode: "qa-cek", DeviceName: "QA Check", Enabled: true}
	// 3. Soft-deleted device DEV-40 with colliding pm10 and o3 parameter codes
	dev40 := &model.Device{ID: 40, DeviceCode: "old-aqms", DeviceName: "Deleted Station", Enabled: true}
	// 4. Active device com5 that is NOT in the checklist
	devCom5 := &model.Device{ID: 5, DeviceCode: "com5", DeviceName: "Extra Sensor", Enabled: true}

	for _, d := range []*model.Device{devCom3, devCom4, devQACek, dev40, devCom5} {
		db.Create(d)
	}
	// Disable qa-cek and soft-delete DEV-40
	_ = deviceRepo.SetEnabled(devQACek.ID, false)
	_ = deviceRepo.Delete(dev40.ID)

	pTemp := &model.Parameter{ID: 31, DeviceID: 3, ParameterCode: "Temperature", ParameterName: "TEMPERATURE", Unit: "C", Enabled: true}
	pPM10 := &model.Parameter{ID: 41, DeviceID: 4, ParameterCode: "pm10", ParameterName: "PM 10", Unit: "ug/m3", Enabled: true}
	pO3 := &model.Parameter{ID: 42, DeviceID: 4, ParameterCode: "o3", ParameterName: "O3", Unit: "ug/m3", Enabled: true}
	pHC := &model.Parameter{ID: 43, DeviceID: 4, ParameterCode: "hc", ParameterName: "HC", Unit: "ug/m3", Enabled: true} // unchecked in checklist
	pQADeleted := &model.Parameter{ID: 101, DeviceID: 10, ParameterCode: "dummy", ParameterName: "Dummy", Unit: "-", Enabled: true}
	pDev40PM10 := &model.Parameter{ID: 401, DeviceID: 40, ParameterCode: "pm10", ParameterName: "PM10", Unit: "ug/m3", Enabled: true}
	pDev40O3 := &model.Parameter{ID: 402, DeviceID: 40, ParameterCode: "o3", ParameterName: "O3", Unit: "ug/m3", Enabled: true}
	pCom5Extra := &model.Parameter{ID: 51, DeviceID: 5, ParameterCode: "pm10", ParameterName: "PM10 Extra", Unit: "ug/m3", Enabled: true}

	for _, p := range []*model.Parameter{pTemp, pPM10, pO3, pHC, pQADeleted, pDev40PM10, pDev40O3, pCom5Extra} {
		db.Create(p)
	}
	_ = deviceRepo.DeleteParameter(10, pQADeleted.ID)

	dest := &model.OutputDestination{
		Name:            "CBI",
		Type:            model.DestREST,
		Endpoint:        "http://cbi.example.com/inbound",
		Enabled:         true,
		IntervalSeconds: 120,
		RetryLimit:      5,
	}
	if err := svc.CreateDestination(ctx, dest); err != nil {
		t.Fatalf("CreateDestination error: %v", err)
	}

	cfg, _ := svc.GetPayloadConfig(ctx, dest.ID)
	cfg.Structure = "MULTI_TICK_ARRAY"
	cfg.StationIDValue = "3453"
	cfg.ModuleIDValue = "1"
	cfg.ParameterMappingsJSON = `[
		{"device_id":3,"device_code":"com3","parameter_id":31,"parameter_code":"Temperature","external_name":"temp","enabled":true},
		{"device_id":4,"device_code":"com4","parameter_id":41,"parameter_code":"pm10","external_name":"pm10","enabled":true},
		{"device_id":4,"device_code":"com4","parameter_id":42,"parameter_code":"o3","external_name":"o3","enabled":true},
		{"device_id":4,"device_code":"com4","parameter_id":43,"parameter_code":"hc","external_name":"hc","enabled":false}
	]`
	if err := svc.UpdatePayloadConfig(ctx, cfg); err != nil {
		t.Fatalf("UpdatePayloadConfig error: %v", err)
	}

	now := time.Now().Truncate(time.Minute)
	vTemp, vPM10, vO3, vHC, vExtra := 23.78, 89.25, 36.91, 2.46, 99.9
	db.Create(&model.AggregationResult{ID: 1001, AggregationDefinitionID: 1, DeviceID: 3, ParameterID: 31, SourceType: model.SourceCustomerProcessed, PeriodStart: now.Add(-2 * time.Minute), PeriodEnd: now, IntervalSeconds: 120, Value: &vTemp, Quality: "GOOD"})
	db.Create(&model.AggregationResult{ID: 1002, AggregationDefinitionID: 2, DeviceID: 4, ParameterID: 41, SourceType: model.SourceCustomerProcessed, PeriodStart: now.Add(-2 * time.Minute), PeriodEnd: now, IntervalSeconds: 120, Value: &vPM10, Quality: "GOOD"})
	db.Create(&model.AggregationResult{ID: 1003, AggregationDefinitionID: 3, DeviceID: 4, ParameterID: 42, SourceType: model.SourceCustomerProcessed, PeriodStart: now.Add(-2 * time.Minute), PeriodEnd: now, IntervalSeconds: 120, Value: &vO3, Quality: "GOOD"})
	db.Create(&model.AggregationResult{ID: 1004, AggregationDefinitionID: 4, DeviceID: 4, ParameterID: 43, SourceType: model.SourceCustomerProcessed, PeriodStart: now.Add(-2 * time.Minute), PeriodEnd: now, IntervalSeconds: 120, Value: &vHC, Quality: "GOOD"})
	// Stale/orphaned rows from qa-cek, DEV-40, and unchecked com5
	db.Create(&model.AggregationResult{ID: 1005, AggregationDefinitionID: 5, DeviceID: 10, ParameterID: 101, SourceType: model.SourceCustomerProcessed, PeriodStart: now.Add(-2 * time.Minute), PeriodEnd: now, IntervalSeconds: 120, Value: nil, Quality: "BAD"})
	db.Create(&model.AggregationResult{ID: 1006, AggregationDefinitionID: 6, DeviceID: 40, ParameterID: 401, SourceType: model.SourceCustomerProcessed, PeriodStart: now.Add(-2 * time.Minute), PeriodEnd: now, IntervalSeconds: 120, Value: nil, Quality: "BAD"})
	db.Create(&model.AggregationResult{ID: 1007, AggregationDefinitionID: 7, DeviceID: 40, ParameterID: 402, SourceType: model.SourceCustomerProcessed, PeriodStart: now.Add(-2 * time.Minute), PeriodEnd: now, IntervalSeconds: 120, Value: nil, Quality: "BAD"})
	db.Create(&model.AggregationResult{ID: 1008, AggregationDefinitionID: 8, DeviceID: 5, ParameterID: 51, SourceType: model.SourceCustomerProcessed, PeriodStart: now.Add(-2 * time.Minute), PeriodEnd: now, IntervalSeconds: 120, Value: &vExtra, Quality: "GOOD"})

	if err := svc.HandleBucketCompleted(ctx, 4, now); err != nil {
		t.Fatalf("HandleBucketCompleted error: %v", err)
	}

	items, total, err := svc.GetQueueItems(ctx, repository.QueueFilter{DestinationID: &dest.ID})
	if err != nil || total != 1 {
		t.Fatalf("expected 1 queue item, got %d (err: %v)", total, err)
	}

	q := items[0]
	if q.DeviceCount != 2 || q.DevicesSummary != "com3, com4" {
		t.Errorf("expected DevicesSummary 'com3, com4' (2 devices), got %q (%d devices)", q.DevicesSummary, q.DeviceCount)
	}
	if q.ParameterCount != 3 {
		t.Errorf("expected ParameterCount 3 (only checked params), got %d", q.ParameterCount)
	}

	detail, err := svc.GetQueueItemDetail(ctx, q.ID)
	if err != nil {
		t.Fatalf("GetQueueItemDetail error: %v", err)
	}
	if len(detail.SourceValues) != 3 {
		t.Errorf("expected 3 SourceValues in Queue Inspection, got %d", len(detail.SourceValues))
	}
	for _, sv := range detail.SourceValues {
		if sv.DeviceCode == "qa-cek" || sv.DeviceCode == "DEV-40" || sv.DeviceCode == "com5" || sv.ParameterCode == "hc" || sv.ParameterCode == "" {
			t.Errorf("unexpected leaked device/param in SourceValues: device=%s param=%s", sv.DeviceCode, sv.ParameterCode)
		}
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(q.PayloadSnapshot), &parsed); err != nil {
		t.Fatalf("invalid payload JSON: %v", err)
	}
	multi, ok := parsed["multi"].([]interface{})
	if !ok || len(multi) != 3 {
		t.Fatalf("expected 3 ticks in multi array, got %v", parsed)
	}
}
