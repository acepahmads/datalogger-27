package service

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"datalogger/internal/model"
)

func TestPayloadFormatter_FlatAndNested(t *testing.T) {
	formatter := NewConfigurablePayloadFormatter()

	val1 := 42.5
	val2 := 18.2
	now := time.Date(2026, 10, 20, 10, 0, 0, 0, time.UTC)

	batch := &OutboundBatchData{
		DeviceID:        1,
		DeviceCode:      "AQMS-001",
		DeviceName:      "Jakarta Central AQMS",
		IntervalSeconds: 300,
		PeriodStart:     now.Add(-5 * time.Minute),
		PeriodEnd:       now,
		Values: []model.SourceAggregationValueItem{
			{
				ParameterID:   1,
				ParameterCode: "pm25",
				ParameterName: "PM2.5",
				Value:         &val1,
				Unit:          "µg/m3",
				Quality:       "GOOD",
			},
			{
				ParameterID:   2,
				ParameterCode: "so2",
				ParameterName: "Sulfur Dioxide",
				Value:         &val2,
				Unit:          "ppm",
				Quality:       "GOOD",
			},
			{
				ParameterID:   3,
				ParameterCode: "co",
				ParameterName: "Carbon Monoxide",
				Value:         nil, // missing telemetry!
				Unit:          "ppm",
				Quality:       "BAD",
			},
		},
	}

	// 1. Test FLAT formatting
	flatCfg := &model.DestinationPayloadConfig{
		StationIDField:  "station_code",
		TimestampField:  "datetime",
		TimestampFormat: "ISO8601",
		IntervalField:   "interval_min",
		Structure:       "FLAT",
		IncludeUnits:    false,
		IncludeQuality:  false,
		ParameterMappingsJSON: `[
			{"parameter_code":"pm25","external_name":"PM25_PARTICULATE","enabled":true},
			{"parameter_code":"so2","external_name":"SO2_GAS","enabled":true},
			{"parameter_code":"co","external_name":"CO_GAS","enabled":true}
		]`,
	}

	bytesFlat, err := formatter.Format(batch, flatCfg)
	if err != nil {
		t.Fatalf("unexpected error formatting flat payload: %v", err)
	}

	var flatMap map[string]interface{}
	if err := json.Unmarshal(bytesFlat, &flatMap); err != nil {
		t.Fatalf("failed to parse generated flat json: %v", err)
	}

	if flatMap["station_code"] != "AQMS-001" {
		t.Errorf("expected station_code AQMS-001, got %v", flatMap["station_code"])
	}
	if flatMap["interval_min"] != float64(5) {
		t.Errorf("expected interval_min 5, got %v", flatMap["interval_min"])
	}
	if flatMap["PM25_PARTICULATE"] != 42.5 {
		t.Errorf("expected PM25_PARTICULATE 42.5, got %v", flatMap["PM25_PARTICULATE"])
	}
	if flatMap["CO_GAS"] != nil {
		t.Errorf("expected CO_GAS to be nil (not 0), got %v", flatMap["CO_GAS"])
	}

	// 2. Test NESTED formatting with units and quality
	nestedCfg := &model.DestinationPayloadConfig{
		StationIDField:  "station_id",
		TimestampField:  "timestamp",
		TimestampFormat: "UNIX_MS",
		IntervalField:   "interval_minutes",
		Structure:       "NESTED",
		NestedDataKey:   "measurements",
		IncludeUnits:    true,
		IncludeQuality:  true,
		ParameterMappingsJSON: `[
			{"parameter_code":"pm25","external_name":"PM25","enabled":true},
			{"parameter_code":"so2","external_name":"SO2","enabled":false}
		]`,
	}

	bytesNested, err := formatter.Format(batch, nestedCfg)
	if err != nil {
		t.Fatalf("unexpected error formatting nested payload: %v", err)
	}

	var nestedMap map[string]interface{}
	if err := json.Unmarshal(bytesNested, &nestedMap); err != nil {
		t.Fatalf("failed to parse generated nested json: %v", err)
	}

	meas, ok := nestedMap["measurements"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected measurements key in nested map, got %v", nestedMap)
	}

	// PM25 should exist with value, unit, quality
	pm25Obj, ok := meas["PM25"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected PM25 object in measurements, got %v", meas["PM25"])
	}
	if pm25Obj["value"] != 42.5 {
		t.Errorf("expected PM25 value 42.5, got %v", pm25Obj["value"])
	}
	if pm25Obj["unit"] != "µg/m3" {
		t.Errorf("expected unit µg/m3, got %v", pm25Obj["unit"])
	}
	if pm25Obj["quality"] != "GOOD" {
		t.Errorf("expected quality GOOD, got %v", pm25Obj["quality"])
	}

	// SO2 was disabled in mapping, should NOT exist
	if _, exists := meas["SO2"]; exists {
		t.Errorf("expected SO2 to be excluded because enabled=false")
	}

	// 3. Test Required Fields Validation
	reqCfg := &model.DestinationPayloadConfig{
		Structure:          "FLAT",
		RequiredFieldsJSON: `["station_id","non_existent_field"]`,
	}
	_, reqErr := formatter.Format(batch, reqCfg)
	if reqErr == nil || !strings.Contains(reqErr.Error(), "required field") {
		t.Errorf("expected required field validation error, got: %v", reqErr)
	}
}

func TestPayloadFormatter_MultiTickArray(t *testing.T) {
	formatter := NewConfigurablePayloadFormatter()

	valNO2 := 15.4
	valSO2 := 2.05
	now := time.Date(2026, 10, 29, 14, 0, 0, 0, time.UTC)

	batch := &OutboundBatchData{
		DeviceID:        1,
		DeviceCode:      "AQMS-001",
		IntervalSeconds: 120,
		PeriodStart:     now.Add(-2 * time.Minute),
		PeriodEnd:       now,
		Values: []model.SourceAggregationValueItem{
			{
				DeviceID:      1,
				DeviceCode:    "AQMS-001",
				ParameterID:   1,
				ParameterCode: "no2",
				ParameterName: "NO2",
				Value:         &valNO2,
			},
			{
				DeviceID:      1,
				DeviceCode:    "AQMS-001",
				ParameterID:   2,
				ParameterCode: "so2",
				ParameterName: "SO2",
				Value:         &valSO2,
			},
		},
	}

	cfg := &model.DestinationPayloadConfig{
		StationIDValue:  "3453",
		ModuleIDValue:   "3453",
		TimestampFormat: "CUSTOM",
		Structure:       "MULTI_TICK_ARRAY",
		ParameterMappingsJSON: `[
			{"parameter_code":"no2","external_name":"no2","enabled":true},
			{"parameter_code":"so2","external_name":"so2","enabled":true}
		]`,
	}

	bytes, err := formatter.Format(batch, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var root map[string]interface{}
	if err := json.Unmarshal(bytes, &root); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	multi, ok := root["multi"].([]interface{})
	if !ok {
		t.Fatalf("expected 'multi' array at root, got %v", root)
	}
	if len(multi) != 2 {
		t.Fatalf("expected 2 tick items, got %d", len(multi))
	}

	firstTickWrapper, ok := multi[0].(map[string]interface{})
	if !ok {
		t.Fatalf("expected object wrapper for tick, got %v", multi[0])
	}
	tick, ok := firstTickWrapper["tick"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected 'tick' inner object, got %v", firstTickWrapper)
	}

	if tick["sensor_name"] != "no2" {
		t.Errorf("expected sensor_name 'no2', got %v", tick["sensor_name"])
	}
	if tick["value"] != 15.4 {
		t.Errorf("expected value 15.4, got %v", tick["value"])
	}
	if tick["station_id"] != float64(3453) {
		t.Errorf("expected station_id 3453, got %v", tick["station_id"])
	}
	if tick["module_id"] != float64(3453) {
		t.Errorf("expected module_id 3453, got %v", tick["module_id"])
	}
	if tick["created_at"] != "2026-10-29 13:58:00" {
		t.Errorf("expected created_at '2026-10-29 13:58:00', got %v", tick["created_at"])
	}
	if tick["updated_at"] != "2026-10-29 14:00:00" {
		t.Errorf("expected updated_at '2026-10-29 14:00:00', got %v", tick["updated_at"])
	}
}

func TestPayloadFormatter_TwoDecimalRounding(t *testing.T) {
	formatter := NewConfigurablePayloadFormatter()

	valTemp := 23.326666666666664
	valPM10 := 88.5
	now := time.Date(2026, 10, 29, 16, 24, 0, 0, time.UTC)

	batch := &OutboundBatchData{
		DeviceID:        3,
		DeviceCode:      "com3",
		IntervalSeconds: 120,
		PeriodStart:     now.Add(-2 * time.Minute),
		PeriodEnd:       now,
		Values: []model.SourceAggregationValueItem{
			{DeviceID: 3, DeviceCode: "com3", ParameterID: 1, ParameterCode: "Temperature", Value: &valTemp},
			{DeviceID: 4, DeviceCode: "com4", ParameterID: 2, ParameterCode: "pm10", Value: &valPM10},
		},
	}

	cfg := &model.DestinationPayloadConfig{
		StationIDValue:  "3453",
		ModuleIDValue:   "3453",
		TimestampFormat: "CUSTOM",
		Structure:       "MULTI_TICK_ARRAY",
		ParameterMappingsJSON: `[
			{"device_id":3,"device_code":"com3","parameter_id":1,"parameter_code":"Temperature","external_name":"temp","enabled":true},
			{"device_id":4,"device_code":"com4","parameter_id":2,"parameter_code":"pm10","external_name":"pm10","enabled":true}
		]`,
	}

	raw, err := formatter.Format(batch, cfg)
	if err != nil {
		t.Fatalf("unexpected Format error: %v", err)
	}
	rawStr := string(raw)
	if !strings.Contains(rawStr, `"value":23.33`) {
		t.Errorf("expected 23.33 in raw JSON, got %s", rawStr)
	}
	if !strings.Contains(rawStr, `"value":88.50`) {
		t.Errorf("expected 88.50 in raw JSON, got %s", rawStr)
	}

	preview, err := formatter.GeneratePreview(batch, cfg)
	if err != nil {
		t.Fatalf("unexpected GeneratePreview error: %v", err)
	}
	if !strings.Contains(preview, `"value": 23.33`) || !strings.Contains(preview, `"value": 88.50`) {
		t.Errorf("expected 2-decimal formatted values in preview JSON, got:\n%s", preview)
	}
}

func TestPayloadFormatter_KLHKISPUAQMS(t *testing.T) {
	formatter := NewConfigurablePayloadFormatter()

	pm10 := 89.25
	pm25 := 139.83333333
	so2 := 2.05669
	co := 630.4166
	o3 := 36.9166
	no2 := 3.7890
	ws := 2.45
	wd := 185.0
	hum := 51.1758
	temp := 23.7816
	press := 935.0891
	sr := 412.5
	rain := 0.0

	// 10:00:00 UTC -> 17:00:00 WIB (Asia/Jakarta)
	periodEndUTC := time.Date(2026, 10, 29, 10, 0, 0, 0, time.UTC)

	batch := &OutboundBatchData{
		DeviceCount:     2,
		DevicesSummary:  "com3, com4",
		IntervalSeconds: 1800,
		PeriodStart:     periodEndUTC.Add(-30 * time.Minute),
		PeriodEnd:       periodEndUTC,
		Values: []model.SourceAggregationValueItem{
			{DeviceID: 4, DeviceCode: "com4", ParameterID: 1, ParameterCode: "pm10", Value: &pm10, Quality: "GOOD"},
			{DeviceID: 4, DeviceCode: "com4", ParameterID: 2, ParameterCode: "pm2.5", Value: &pm25, Quality: "GOOD"},
			{DeviceID: 4, DeviceCode: "com4", ParameterID: 3, ParameterCode: "so2", Value: &so2, Quality: "GOOD"},
			{DeviceID: 4, DeviceCode: "com4", ParameterID: 4, ParameterCode: "co", Value: &co, Quality: "GOOD"},
			{DeviceID: 4, DeviceCode: "com4", ParameterID: 5, ParameterCode: "o3", Value: &o3, Quality: "GOOD"},
			{DeviceID: 4, DeviceCode: "com4", ParameterID: 6, ParameterCode: "no2", Value: &no2, Quality: "GOOD"},
			{DeviceID: 4, DeviceCode: "com4", ParameterID: 7, ParameterCode: "hc", Value: nil, Quality: "BAD"},
			{DeviceID: 3, DeviceCode: "com3", ParameterID: 8, ParameterCode: "w-speed", Value: &ws, Quality: "GOOD"},
			{DeviceID: 3, DeviceCode: "com3", ParameterID: 9, ParameterCode: "wind-angle", Value: &wd, Quality: "GOOD"},
			{DeviceID: 3, DeviceCode: "com3", ParameterID: 10, ParameterCode: "humidity", Value: &hum, Quality: "GOOD"},
			{DeviceID: 3, DeviceCode: "com3", ParameterID: 11, ParameterCode: "Temperature", Value: &temp, Quality: "GOOD"},
			{DeviceID: 3, DeviceCode: "com3", ParameterID: 12, ParameterCode: "air-pressure", Value: &press, Quality: "GOOD"},
			{DeviceID: 3, DeviceCode: "com3", ParameterID: 13, ParameterCode: "solar-radiation", Value: &sr, Quality: "GOOD"},
			{DeviceID: 3, DeviceCode: "com3", ParameterID: 14, ParameterCode: "rain-fall", Value: &rain, Quality: "GOOD"},
		},
	}

	cfg := &model.DestinationPayloadConfig{
		TemplateType:           "KLHK_ISPU_AQMS",
		StationIDField:         "id_stasiun",
		StationIDValue:         "KABUPATEN_JOMBANG",
		TimestampField:         "waktu",
		TimestampFormat:        "CUSTOM",
		TimestampTimezone:      "Asia/Jakarta",
		OmitIntervalField:      true,
		Structure:              "FLAT",
		ConflictResolution:     "PRIORITY_PRIMARY",
		IncludeMetadata:        false,
		IncludeSensorStatFlags: true,
		SensorStatMode:         "STATIC_1",
	}

	raw, err := formatter.Format(batch, cfg)
	if err != nil {
		t.Fatalf("unexpected Format error: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	if parsed["id_stasiun"] != "KABUPATEN_JOMBANG" {
		t.Errorf("expected id_stasiun KABUPATEN_JOMBANG, got %v", parsed["id_stasiun"])
	}
	if parsed["waktu"] != "2026-10-29 17:00:00" {
		t.Errorf("expected WIB waktu '2026-10-29 17:00:00', got %v", parsed["waktu"])
	}
	if _, hasInterval := parsed["interval_minutes"]; hasInterval {
		t.Errorf("expected interval_minutes to be omitted when OmitIntervalField=true")
	}
	if parsed["pm25"] != 139.83 {
		t.Errorf("expected pm25=139.83, got %v", parsed["pm25"])
	}
	if parsed["ws"] != 2.45 {
		t.Errorf("expected ws=2.45, got %v", parsed["ws"])
	}
	if parsed["wd"] != 185.0 {
		t.Errorf("expected wd=185.0, got %v", parsed["wd"])
	}
	if parsed["temperature"] != 23.78 {
		t.Errorf("expected temperature=23.78, got %v", parsed["temperature"])
	}
	if parsed["pressure"] != 935.09 {
		t.Errorf("expected pressure=935.09, got %v", parsed["pressure"])
	}
	if parsed["sr"] != 412.5 {
		t.Errorf("expected sr=412.5, got %v", parsed["sr"])
	}
	if parsed["rain_intensity"] != 0.0 {
		t.Errorf("expected rain_intensity=0.0, got %v", parsed["rain_intensity"])
	}

	// Verify STATIC_1 status flags
	for _, code := range []string{"pm10", "pm25", "so2", "co", "o3", "no2", "hc"} {
		if parsed["stat_"+code] != float64(1) {
			t.Errorf("expected stat_%s=1 in STATIC_1 mode, got %v", code, parsed["stat_"+code])
		}
	}

	// Verify AUTO_QUALITY status flags (hc is BAD/nil -> stat_hc should be 0, pm25 GOOD -> stat_pm25 should be 1)
	cfg.SensorStatMode = "AUTO_QUALITY"
	rawAuto, err := formatter.Format(batch, cfg)
	if err != nil {
		t.Fatalf("unexpected Format error in AUTO_QUALITY: %v", err)
	}
	var parsedAuto map[string]interface{}
	_ = json.Unmarshal(rawAuto, &parsedAuto)
	if parsedAuto["stat_pm25"] != float64(1) {
		t.Errorf("expected stat_pm25=1 in AUTO_QUALITY, got %v", parsedAuto["stat_pm25"])
	}
	if parsedAuto["stat_hc"] != float64(0) {
		t.Errorf("expected stat_hc=0 for BAD/null hc in AUTO_QUALITY, got %v", parsedAuto["stat_hc"])
	}
}


