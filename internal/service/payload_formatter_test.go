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
