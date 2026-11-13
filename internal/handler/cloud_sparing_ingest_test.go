package handler_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"datalogger/internal/config"
	"datalogger/internal/model"
	"datalogger/internal/service"
)

func TestCloudMode_Sparing30x2HourlyIngestFormatsAndAggregationRecalc(t *testing.T) {
	env := setupMultiTenantTestEnv(t)

	// Activate Cloud mode
	if g := config.Get(); g != nil {
		g.CloudMode = true
		g.MonitoringDomain = "CLOUD"
		g.SensorPollingEnabled = false
	}

	// 1. Create Tenant + SPARING Station + Station-Locked Bearer Token
	wTenant := doJSONReq(t, env.Router, "POST", "/api/tenants", env.SuperAdminToken, nil, map[string]interface{}{
		"name":   "PT Industri Sparing Cloud",
		"code":   "pt-sparing-cloud",
		"status": "ACTIVE",
	})
	if wTenant.Code != http.StatusCreated && wTenant.Code != http.StatusOK {
		t.Fatalf("create tenant failed (%d): %s", wTenant.Code, wTenant.Body.String())
	}
	var createdTenant struct {
		Data struct {
			Tenant model.Tenant `json:"tenant"`
		} `json:"data"`
	}
	_ = json.Unmarshal(wTenant.Body.Bytes(), &createdTenant)
	tenantID := createdTenant.Data.Tenant.ID

	wSite := doJSONReq(t, env.Router, "POST", "/api/sites", env.SuperAdminToken, nil, map[string]interface{}{
		"tenant_id":         tenantID,
		"name":              "WWTP Outlet SPARING 01",
		"code":              "wwtp-sparing-01",
		"monitoring_domain": "SPARING",
		"timezone":          "Asia/Jakarta",
		"enabled":           true,
	})
	if wSite.Code != http.StatusCreated && wSite.Code != http.StatusOK {
		t.Fatalf("create SPARING site failed (%d): %s", wSite.Code, wSite.Body.String())
	}
	var createdSite struct {
		Data model.Site `json:"data"`
	}
	_ = json.Unmarshal(wSite.Body.Bytes(), &createdSite)
	siteID := createdSite.Data.ID

	wKey := doJSONReq(t, env.Router, "POST", "/api/ingest-keys", env.SuperAdminToken, nil, map[string]interface{}{
		"tenant_id": tenantID,
		"site_id":   siteID,
		"name":      "SPARING Hourly 30x2 Token",
	})
	if wKey.Code != http.StatusCreated && wKey.Code != http.StatusOK {
		t.Fatalf("create ingest key failed (%d): %s", wKey.Code, wKey.Body.String())
	}
	var keyResp struct {
		Data struct {
			PlainKey string `json:"plain_key"`
		} `json:"data"`
	}
	_ = json.Unmarshal(wKey.Body.Bytes(), &keyResp)
	plainKey := keyResp.Data.PlainKey

	// 2. Manually create passive EDGE_PUSH SPARING device + parameters
	wDev := doJSONReq(t, env.Router, "POST", "/api/devices", env.SuperAdminToken, map[string]string{
		"X-Tenant-ID": fmt.Sprintf("%d", tenantID),
	}, map[string]interface{}{
		"site_id":     siteID,
		"device_code": "SPARING-CBI-01",
		"device_name": "Datalogger SPARING Outlet",
		"device_type": "EDGE_PUSH",
		"timezone":    "Asia/Jakarta",
		"status":      "ACTIVE",
		"enabled":     true,
		"connection": map[string]interface{}{
			"protocol":        "EDGE_PUSH",
			"connection_type": "HTTPS_PUSH",
			"timeout":         3900, // 65 minutes for hourly 30x2 push
			"enabled":         true,
		},
	})
	if wDev.Code != http.StatusCreated && wDev.Code != http.StatusOK {
		t.Fatalf("create SPARING device failed (%d): %s", wDev.Code, wDev.Body.String())
	}
	var devResp struct {
		Data model.Device `json:"data"`
	}
	_ = json.Unmarshal(wDev.Body.Bytes(), &devResp)
	devID := devResp.Data.ID

	paramSpecs := []map[string]interface{}{
		{"parameter_code": "ph", "parameter_name": "pH Air Limbah", "unit": "-", "data_type": "FLOAT64", "precision": 2},
		{"parameter_code": "debit", "parameter_name": "Debit Limbah", "unit": "m3/jam", "data_type": "FLOAT64", "precision": 2},
		{"parameter_code": "tss", "parameter_name": "TSS", "unit": "mg/L", "data_type": "FLOAT64", "precision": 2},
		{"parameter_code": "cod", "parameter_name": "COD", "unit": "mg/L", "data_type": "FLOAT64", "precision": 2},
		{"parameter_code": "temp_panel", "parameter_name": "Suhu Panel", "unit": "C", "data_type": "FLOAT64", "precision": 2},
		{"parameter_code": "humidity", "parameter_name": "Kelembaban Panel", "unit": "%", "data_type": "FLOAT64", "precision": 2},
	}
	var createdParams []model.Parameter
	for _, spec := range paramSpecs {
		wP := doJSONReq(t, env.Router, "POST", fmt.Sprintf("/api/devices/%d/parameters", devID), env.SuperAdminToken, map[string]string{
			"X-Tenant-ID": fmt.Sprintf("%d", tenantID),
		}, spec)
		if wP.Code != http.StatusCreated && wP.Code != http.StatusOK {
			t.Fatalf("create param %v failed (%d): %s", spec["parameter_code"], wP.Code, wP.Body.String())
		}
		var pr struct {
			Data model.Parameter `json:"data"`
		}
		_ = json.Unmarshal(wP.Body.Bytes(), &pr)
		createdParams = append(createdParams, pr.Data)
	}

	// Provision 2-minute (120s) and 1-hour (3600s) AggregationDefinitions with LastCalculatedAt ALREADY advanced to now,
	// simulating Cloud server's background worker having already stepped past empty 2-minute slots before the hourly push arrives!
	wib := time.FixedZone("WIB", 7*3600)
	nowUTC := time.Now().UTC()
	hourAnchorLocal := time.Now().In(wib).Add(-2 * time.Hour).Truncate(time.Hour)

	for _, p := range createdParams {
		for _, iv := range []int{120, 3600} {
			lastCalc := nowUTC
			def := model.AggregationDefinition{
				TenantID:         tenantID,
				SiteID:           siteID,
				Name:             fmt.Sprintf("%s_%ds_AVG", p.ParameterCode, iv),
				Code:             fmt.Sprintf("%s_%ds_avg", p.ParameterCode, iv),
				DeviceID:         devID,
				ParameterID:      p.ID,
				IntervalSeconds:  iv,
				Function:         model.FunctionAvg,
				SourceType:       model.SourceCustomerProcessed,
				Timezone:         "Asia/Jakarta",
				Enabled:          true,
				LastCalculatedAt: &lastCalc,
			}
			if err := env.DB.Create(&def).Error; err != nil {
				t.Fatalf("create agg def failed: %v", err)
			}
		}
	}

	// =========================================================================
	// CASE A: Cloud CBI Multi 30x2-Minute ("module_id", "station_id", "hourly_data")
	// =========================================================================
	hourlyRows := make([]map[string]interface{}, 0, 30)
	for i := 0; i < 30; i++ {
		slotTS := hourAnchorLocal.Add(time.Duration(i*2) * time.Minute)
		tsStr := slotTS.Format("2006-01-02 15:04:05 -0700")
		hourlyRows = append(hourlyRows, map[string]interface{}{
			"ph":         7.00 + float64(i)*0.01,
			"debit":      0.23,
			"tss":        19.0 + float64(i)*0.5,
			"cod":        85.0 + float64(i)*0.2,
			"temp_panel": 28.1,
			"humidity":   0,
			"created_at": tsStr,
			"updated_at": tsStr,
		})
	}
	wCBI := doJSONReq(t, env.Router, "POST", "/api/v1/ingest/telemetry", plainKey, nil, map[string]interface{}{
		"module_id":   "4",
		"station_id":  "3",
		"hourly_data": hourlyRows,
	})
	if wCBI.Code != http.StatusOK {
		t.Fatalf("CBI hourly_data 30x2 HTTP push failed (%d): %s", wCBI.Code, wCBI.Body.String())
	}
	var cbiResp struct {
		Data service.EdgeIngestResponse `json:"data"`
	}
	_ = json.Unmarshal(wCBI.Body.Bytes(), &cbiResp)
	ingestResp := cbiResp.Data
	if ingestResp.RecordsProcessed != 30 {
		t.Fatalf("expected 30 records_processed for CBI hourly_data, got %d", ingestResp.RecordsProcessed)
	}
	if ingestResp.ParametersIngested != 180 {
		t.Fatalf("expected 180 parameters_ingested (30*6), got %d", ingestResp.ParametersIngested)
	}
	if ingestResp.DeviceCode != "SPARING-CBI-01" {
		t.Fatalf("expected resolved device SPARING-CBI-01, got %s", ingestResp.DeviceCode)
	}

	// Verify latest live parameter reading reflects the 30th slot (:58 -> ph = 7.29)
	var phParam model.Parameter
	if err := env.DB.Where("device_id = ? AND parameter_code = ?", devID, "ph").First(&phParam).Error; err != nil {
		t.Fatalf("fetch ph parameter failed: %v", err)
	}
	if phParam.CurrentValue == nil || *phParam.CurrentValue != 7.29 {
		t.Fatalf("expected ph CurrentValue=7.29 (slot :58), got %v", phParam.CurrentValue)
	}

	// Verify all 30 x 2-minute buckets AND 1 x 1-hour bucket were automatically recalculated for pH!
	var ph2mCount int64
	env.DB.Model(&model.AggregationResult{}).
		Where("device_id = ? AND parameter_id = ? AND interval_seconds = 120 AND period_start >= ? AND period_end <= ?",
			devID, phParam.ID, hourAnchorLocal.UTC(), hourAnchorLocal.Add(time.Hour).UTC()).
		Count(&ph2mCount)
	if ph2mCount != 30 {
		t.Fatalf("expected 30 recalculated 2-minute buckets for ph, got %d", ph2mCount)
	}

	var ph1hCount int64
	env.DB.Model(&model.AggregationResult{}).
		Where("device_id = ? AND parameter_id = ? AND interval_seconds = 3600 AND period_start = ?",
			devID, phParam.ID, hourAnchorLocal.UTC()).
		Count(&ph1hCount)
	if ph1hCount != 1 {
		t.Fatalf("expected 1 recalculated 1-hour bucket for ph, got %d", ph1hCount)
	}

	// =========================================================================
	// CASE B: KLHK SPARING Hourly 30x2-Minute ("uid" + "data": [{datetime, pH, cod, tss, nh3n, debit}])
	// via HTTP Router POST /api/v1/ingest/telemetry
	// =========================================================================
	nextHourLocal := hourAnchorLocal.Add(time.Hour)
	klhkRows := make([]map[string]interface{}, 0, 30)
	for i := 0; i < 30; i++ {
		slotTS := nextHourLocal.Add(time.Duration(i*2) * time.Minute)
		klhkRows = append(klhkRows, map[string]interface{}{
			"datetime": slotTS.Unix(),
			"pH":       7.10 + float64(i)*0.01, // uppercase H matches manual 'ph' case-insensitively
			"cod":      80.0 + float64(i)*0.1,
			"tss":      20.0 + float64(i)*0.2,
			"nh3n":     0.35, // unregistered on SPARING-CBI-01 -> skipped cleanly without auto-creating!
			"debit":    0.25,
		})
	}

	wKLHK := doJSONReq(t, env.Router, "POST", "/api/v1/ingest/telemetry", plainKey, nil, map[string]interface{}{
		"uid":  "79456d9c51128700b78b6fd0",
		"data": klhkRows,
	})
	if wKLHK.Code != http.StatusOK {
		t.Fatalf("KLHK SPARING 30x2 HTTP push failed (%d): %s", wKLHK.Code, wKLHK.Body.String())
	}
	var klhkResp struct {
		Data service.EdgeIngestResponse `json:"data"`
	}
	_ = json.Unmarshal(wKLHK.Body.Bytes(), &klhkResp)
	if klhkResp.Data.RecordsProcessed != 30 {
		t.Fatalf("expected 30 records_processed for KLHK SPARING data[], got %d", klhkResp.Data.RecordsProcessed)
	}
	if klhkResp.Data.ParametersIngested != 120 { // 30 * 4 registered (ph, cod, tss, debit)
		t.Fatalf("expected 120 parameters_ingested (30*4), got %d", klhkResp.Data.ParametersIngested)
	}
	if klhkResp.Data.ParametersSkipped != 30 { // 30 * 1 unregistered (nh3n)
		t.Fatalf("expected 30 parameters_skipped (nh3n), got %d", klhkResp.Data.ParametersSkipped)
	}

	// =========================================================================
	// CASE C: Datalogger-to-Datalogger Bundled 30x2-Minute ARRAY_BY_DEVICE Sync (INGEST_KEY preset)
	// =========================================================================
	formatter := service.NewConfigurablePayloadFormatter()
	edgeSyncCfg := &model.DestinationPayloadConfig{
		TemplateType:       "INGEST_KEY",
		Structure:          "ARRAY_BY_DEVICE",
		StationIDField:     "site_code",
		StationIDValue:     "wwtp-sparing-01",
		TimestampField:     "timestamp",
		TimestampFormat:    "ISO8601",
		TimestampTimezone:  "Asia/Jakarta",
		TimestampSource:    "PERIOD_START",
		NestedDataKey:      "devices",
		ConflictResolution: "PRIORITY_PRIMARY",
	}
	var batchVals []model.SourceAggregationValueItem
	thirdHourStart := hourAnchorLocal.Add(-time.Hour).UTC()
	for i := 0; i < 30; i++ {
		ps := thirdHourStart.Add(time.Duration(i*2) * time.Minute)
		pe := ps.Add(2 * time.Minute)
		phV := 7.05
		codV := 75.4
		batchVals = append(batchVals,
			model.SourceAggregationValueItem{
				DeviceID: devID, DeviceCode: "SPARING-CBI-01", DeviceName: "Datalogger SPARING Outlet",
				ParameterID: createdParams[0].ID, ParameterCode: "ph", Value: &phV, Quality: "GOOD",
				PeriodStart: ps, PeriodEnd: pe,
			},
			model.SourceAggregationValueItem{
				DeviceID: devID, DeviceCode: "SPARING-CBI-01", DeviceName: "Datalogger SPARING Outlet",
				ParameterID: createdParams[3].ID, ParameterCode: "cod", Value: &codV, Quality: "GOOD",
				PeriodStart: ps, PeriodEnd: pe,
			},
		)
	}
	bundledJSON, err := formatter.Format(&service.OutboundBatchData{
		DeviceID:        devID,
		DeviceCode:      "SPARING-CBI-01",
		DeviceName:      "Datalogger SPARING Outlet",
		PeriodStart:     thirdHourStart,
		PeriodEnd:       thirdHourStart.Add(time.Hour),
		IntervalSeconds: 120,
		Values:          batchVals,
	}, edgeSyncCfg)
	if err != nil {
		t.Fatalf("format bundled ARRAY_BY_DEVICE failed: %v", err)
	}

	var bundledRaw map[string]interface{}
	_ = json.Unmarshal(bundledJSON, &bundledRaw)
	wBundled := doJSONReq(t, env.Router, "POST", "/api/v1/ingest/telemetry", plainKey, nil, bundledRaw)
	if wBundled.Code != http.StatusOK {
		t.Fatalf("Bundled ARRAY_BY_DEVICE 30x2 HTTP push failed (%d): %s", wBundled.Code, wBundled.Body.String())
	}
	var bundledResp struct {
		Data service.EdgeIngestResponse `json:"data"`
	}
	_ = json.Unmarshal(wBundled.Body.Bytes(), &bundledResp)
	if bundledResp.Data.RecordsProcessed != 30 || bundledResp.Data.ParametersIngested != 60 {
		t.Fatalf("expected 30 records & 60 params for bundled ARRAY_BY_DEVICE, got %+v", bundledResp.Data)
	}

	// =========================================================================
	// CASE D: Payload Template Endpoint Returns SPARING 30x2 Templates
	// =========================================================================
	wTpl := doJSONReq(t, env.Router, "GET",
		fmt.Sprintf("/api/ingest-keys/payload-template?logger_type=SPARING&site_id=%d&device_id=%d", siteID, devID),
		env.SuperAdminToken,
		map[string]string{"X-Tenant-ID": fmt.Sprintf("%d", tenantID)},
		nil,
	)
	if wTpl.Code != http.StatusOK {
		t.Fatalf("GET payload-template failed (%d): %s", wTpl.Code, wTpl.Body.String())
	}
	var tplResp struct {
		Data service.PayloadTemplateResponse `json:"data"`
	}
	_ = json.Unmarshal(wTpl.Body.Bytes(), &tplResp)
	stdRecs, _ := tplResp.Data.BatchPayload["records"].([]interface{})
	cbiTplRows, _ := tplResp.Data.CBIHourlyPayload["hourly_data"].([]interface{})
	klhkTplRows, _ := tplResp.Data.SparingHourlyPayload["data"].([]interface{})
	if len(stdRecs) != 30 || len(cbiTplRows) != 30 || len(klhkTplRows) != 30 {
		t.Fatalf("expected 30 slots in batch_payload (%d), cbi_hourly_payload (%d), sparing_hourly_payload (%d)",
			len(stdRecs), len(cbiTplRows), len(klhkTplRows))
	}
}

func TestCloudMode_LegacyCBICreateMultiTickAndFormatSwitch(t *testing.T) {
	env := setupMultiTenantTestEnv(t)

	if g := config.Get(); g != nil {
		g.CloudMode = true
		g.MonitoringDomain = "CLOUD"
		g.SensorPollingEnabled = false
	}

	// 1. Create Tenant + two AQMS Stations (to verify dynamic station_id: 202 routing on an unlocked token!)
	wTenant := doJSONReq(t, env.Router, "POST", "/api/tenants", env.SuperAdminToken, nil, map[string]interface{}{
		"name":   "PT Udara Bersih Legacy CBI",
		"code":   "pt-udara-cbi",
		"status": "ACTIVE",
	})
	if wTenant.Code != http.StatusCreated && wTenant.Code != http.StatusOK {
		t.Fatalf("create tenant failed (%d): %s", wTenant.Code, wTenant.Body.String())
	}
	var createdTenant struct {
		Data struct {
			Tenant model.Tenant `json:"tenant"`
		} `json:"data"`
	}
	_ = json.Unmarshal(wTenant.Body.Bytes(), &createdTenant)
	tenantID := createdTenant.Data.Tenant.ID

	// Create AQMS station mapped to legacy_station_id = "202"
	wSite := doJSONReq(t, env.Router, "POST", "/api/sites", env.SuperAdminToken, nil, map[string]interface{}{
		"tenant_id":         tenantID,
		"name":              "AQMS Stasiun 202",
		"code":              "aqms-stasiun-202",
		"monitoring_domain": "AQMS",
		"ingest_format":     "AUTO",
		"legacy_station_id": "202",
		"timezone":          "Asia/Jakarta",
		"enabled":           true,
	})
	if wSite.Code != http.StatusCreated && wSite.Code != http.StatusOK {
		t.Fatalf("create AQMS site failed (%d): %s", wSite.Code, wSite.Body.String())
	}
	var createdSite struct {
		Data model.Site `json:"data"`
	}
	_ = json.Unmarshal(wSite.Body.Bytes(), &createdSite)
	siteID := createdSite.Data.ID

	// 2. Create two EDGE_PUSH devices in site 202 so module_id: 199 must route specifically to the second device!
	_ = doJSONReq(t, env.Router, "POST", "/api/devices", env.SuperAdminToken, map[string]string{
		"X-Tenant-ID": fmt.Sprintf("%d", tenantID),
	}, map[string]interface{}{
		"site_id":          siteID,
		"device_code":      "AQMS-OTHER-MOD",
		"device_name":      "Other Module",
		"legacy_module_id": "188",
		"device_type":      "EDGE_PUSH",
		"enabled":          true,
		"connection": map[string]interface{}{
			"protocol":        "EDGE_PUSH",
			"connection_type": "HTTPS_PUSH",
			"timeout":         300,
			"enabled":         true,
		},
	})

	wDev := doJSONReq(t, env.Router, "POST", "/api/devices", env.SuperAdminToken, map[string]string{
		"X-Tenant-ID": fmt.Sprintf("%d", tenantID),
	}, map[string]interface{}{
		"site_id":          siteID,
		"device_code":      "AQMS-MOD-199",
		"device_name":      "AQMS Main Logger 199",
		"legacy_module_id": "199",
		"device_type":      "EDGE_PUSH",
		"enabled":          true,
		"connection": map[string]interface{}{
			"protocol":        "EDGE_PUSH",
			"connection_type": "HTTPS_PUSH",
			"timeout":         300,
			"enabled":         true,
		},
	})
	if wDev.Code != http.StatusCreated && wDev.Code != http.StatusOK {
		t.Fatalf("create AQMS device 199 failed (%d): %s", wDev.Code, wDev.Body.String())
	}
	var devResp struct {
		Data model.Device `json:"data"`
	}
	_ = json.Unmarshal(wDev.Body.Bytes(), &devResp)
	devID := devResp.Data.ID

	// Register the 14 manual AQMS parameter codes (matching production setup)
	aqmsParamCodes := []string{
		"air-pressure", "co", "hc", "humidity", "no2", "o3", "pm10", "pm2.5",
		"rain-fall", "so2", "temperature", "solar-radiation", "w-speed", "wind-angle",
	}
	for _, code := range aqmsParamCodes {
		wP := doJSONReq(t, env.Router, "POST", fmt.Sprintf("/api/devices/%d/parameters", devID), env.SuperAdminToken, map[string]string{
			"X-Tenant-ID": fmt.Sprintf("%d", tenantID),
		}, map[string]interface{}{
			"parameter_code": code,
			"parameter_name": code,
			"unit":           "ug/m3",
			"data_type":      "FLOAT32",
			"precision":      3,
		})
		if wP.Code != http.StatusCreated && wP.Code != http.StatusOK {
			t.Fatalf("create param %s failed (%d): %s", code, wP.Code, wP.Body.String())
		}
	}

	// 3. Create Bearer Token with Custom Legacy Token + IngestFormat = CBI_LEGACY (Tenant-wide, unlocked site_id!)
	customBearer := "legacy_cbi_cloud_token_station202_mod199"
	wKey := doJSONReq(t, env.Router, "POST", "/api/ingest-keys", env.SuperAdminToken, nil, map[string]interface{}{
		"tenant_id":     tenantID,
		"name":          "Token Logger Lama 202",
		"ingest_format": "CBI_LEGACY",
		"custom_token":  customBearer,
	})
	if wKey.Code != http.StatusCreated && wKey.Code != http.StatusOK {
		t.Fatalf("create custom legacy ingest key failed (%d): %s", wKey.Code, wKey.Body.String())
	}
	var keyResp struct {
		Data struct {
			Key      model.TenantAPIKey `json:"key"`
			PlainKey string             `json:"plain_key"`
		} `json:"data"`
	}
	_ = json.Unmarshal(wKey.Body.Bytes(), &keyResp)
	keyID := keyResp.Data.Key.ID
	if keyResp.Data.PlainKey != customBearer {
		t.Fatalf("expected custom plain_key %q, got %q", customBearer, keyResp.Data.PlainKey)
	}
	if keyResp.Data.Key.IngestFormat != model.IngestFormatCBILegacy {
		t.Fatalf("expected key ingest_format=CBI_LEGACY, got %q", keyResp.Data.Key.IngestFormat)
	}

	// 4. Send exact 15-sensor legacy CBI multi[tick] payload to POST /api/v1/ticks/create_multi with Authorization: Bearer ...
	tsStr := time.Now().UTC().Add(-1 * time.Minute).Format(time.RFC3339)
	legacyPayload := map[string]interface{}{
		"multi": []map[string]interface{}{
			{"tick": map[string]interface{}{"module_id": 199, "station_id": 202, "sensor_name": "no2", "value": 14.236, "created_at": tsStr, "updated_at": tsStr}},
			{"tick": map[string]interface{}{"module_id": 199, "station_id": 202, "sensor_name": "so2", "value": 5.673, "created_at": tsStr, "updated_at": tsStr}},
			{"tick": map[string]interface{}{"module_id": 199, "station_id": 202, "sensor_name": "o3", "value": 70.782, "created_at": tsStr, "updated_at": tsStr}},
			{"tick": map[string]interface{}{"module_id": 199, "station_id": 202, "sensor_name": "co", "value": 586.191, "created_at": tsStr, "updated_at": tsStr}},
			{"tick": map[string]interface{}{"module_id": 199, "station_id": 202, "sensor_name": "pm25", "value": 34.811, "created_at": tsStr, "updated_at": tsStr}},
			{"tick": map[string]interface{}{"module_id": 199, "station_id": 202, "sensor_name": "pm10", "value": 66.176, "created_at": tsStr, "updated_at": tsStr}},
			{"tick": map[string]interface{}{"module_id": 199, "station_id": 202, "sensor_name": "temp", "value": 32.941, "created_at": tsStr, "updated_at": tsStr}},
			{"tick": map[string]interface{}{"module_id": 199, "station_id": 202, "sensor_name": "humidity", "value": 63.874, "created_at": tsStr, "updated_at": tsStr}},
			{"tick": map[string]interface{}{"module_id": 199, "station_id": 202, "sensor_name": "pressure", "value": 1007.5, "created_at": tsStr, "updated_at": tsStr}},
			{"tick": map[string]interface{}{"module_id": 199, "station_id": 202, "sensor_name": "wspeed", "value": 3.543, "created_at": tsStr, "updated_at": tsStr}},
			{"tick": map[string]interface{}{"module_id": 199, "station_id": 202, "sensor_name": "intensity", "value": 813.654, "created_at": tsStr, "updated_at": tsStr}},
			{"tick": map[string]interface{}{"module_id": 199, "station_id": 202, "sensor_name": "tprecipitation", "value": 0.4, "created_at": tsStr, "updated_at": tsStr}},
			{"tick": map[string]interface{}{"module_id": 199, "station_id": 202, "sensor_name": "hc", "value": 4.686, "created_at": tsStr, "updated_at": tsStr}},
			{"tick": map[string]interface{}{"module_id": 199, "station_id": 202, "sensor_name": "wind_angle", "value": 237.4, "created_at": tsStr, "updated_at": tsStr}},
			{"tick": map[string]interface{}{"module_id": 199, "station_id": 202, "sensor_name": "rain_intensity", "value": 0, "created_at": tsStr, "updated_at": tsStr}},
		},
	}

	wLegacy := doJSONReq(t, env.Router, "POST", "/api/v1/ticks/create_multi", customBearer, nil, legacyPayload)
	if wLegacy.Code != http.StatusOK {
		t.Fatalf("POST /api/v1/ticks/create_multi failed (%d): %s", wLegacy.Code, wLegacy.Body.String())
	}
	var legacyResp struct {
		Data service.EdgeIngestResponse `json:"data"`
	}
	_ = json.Unmarshal(wLegacy.Body.Bytes(), &legacyResp)
	if legacyResp.Data.SiteCode != "aqms-stasiun-202" || legacyResp.Data.DeviceCode != "AQMS-MOD-199" {
		t.Fatalf("expected routing to aqms-stasiun-202 / AQMS-MOD-199 via station_id:202 & module_id:199, got %s / %s",
			legacyResp.Data.SiteCode, legacyResp.Data.DeviceCode)
	}
	if legacyResp.Data.LoggerType != "AQMS" {
		t.Fatalf("expected LoggerType=AQMS (inherited from site domain), got %s", legacyResp.Data.LoggerType)
	}
	if legacyResp.Data.DetectedFormat != model.IngestFormatCBILegacy {
		t.Fatalf("expected detected_format=CBI_LEGACY, got %s", legacyResp.Data.DetectedFormat)
	}
	if legacyResp.Data.ParametersIngested != 14 || legacyResp.Data.ParametersSkipped != 1 {
		t.Fatalf("expected 14 parameters_ingested and 1 skipped (duplicate rain_intensity), got ingested=%d skipped=%d",
			legacyResp.Data.ParametersIngested, legacyResp.Data.ParametersSkipped)
	}

	// Verify aliased values landed on exact registered parameter codes
	expectedVals := map[string]float64{
		"pm2.5":           34.811,
		"temperature":     32.941,
		"air-pressure":    1007.5,
		"w-speed":         3.543,
		"solar-radiation": 813.654,
		"rain-fall":       0.4,
		"wind-angle":      237.4,
	}
	for code, want := range expectedVals {
		var p model.Parameter
		if err := env.DB.Where("device_id = ? AND parameter_code = ?", devID, code).First(&p).Error; err != nil {
			t.Fatalf("fetch param %s failed: %v", code, err)
		}
		if p.CurrentValue == nil || *p.CurrentValue != want {
			t.Fatalf("param %s expected CurrentValue=%v, got %v", code, want, p.CurrentValue)
		}
	}

	// 5. Verify Format Switch Enforcement:
	// 5a. Token is currently CBI_LEGACY -> STANDARD_V1 payload MUST be rejected (HTTP 400)
	stdPayload := map[string]interface{}{
		"logger_type": "AQMS",
		"site_code":   "aqms-stasiun-202",
		"device_code": "AQMS-MOD-199",
		"timestamp":   tsStr,
		"values": map[string]interface{}{
			"no2":   15.1,
			"pm2.5": 31.2,
		},
	}
	wRejectStd := doJSONReq(t, env.Router, "POST", "/api/v1/ingest/telemetry", customBearer, nil, stdPayload)
	if wRejectStd.Code != http.StatusBadRequest {
		t.Fatalf("expected HTTP 400 when pushing STANDARD_V1 to CBI_LEGACY token, got %d: %s", wRejectStd.Code, wRejectStd.Body.String())
	}

	// 5b. Live-switch Token to STANDARD_V1 via PUT /api/ingest-keys/:id
	wSwitchStd := doJSONReq(t, env.Router, "PUT", fmt.Sprintf("/api/ingest-keys/%d", keyID), env.SuperAdminToken, nil, map[string]interface{}{
		"ingest_format": "STANDARD_V1",
	})
	if wSwitchStd.Code != http.StatusOK {
		t.Fatalf("switch token ingest_format to STANDARD_V1 failed (%d): %s", wSwitchStd.Code, wSwitchStd.Body.String())
	}

	// Legacy multi[tick] MUST now be rejected (HTTP 400), while STANDARD_V1 MUST succeed (HTTP 200)!
	wRejectLegacy := doJSONReq(t, env.Router, "POST", "/api/v1/ticks/create_multi", customBearer, nil, legacyPayload)
	if wRejectLegacy.Code != http.StatusBadRequest {
		t.Fatalf("expected HTTP 400 when pushing CBI_LEGACY to STANDARD_V1 token, got %d: %s", wRejectLegacy.Code, wRejectLegacy.Body.String())
	}
	wAcceptStd := doJSONReq(t, env.Router, "POST", "/api/v1/ingest/telemetry", customBearer, nil, stdPayload)
	if wAcceptStd.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200 when pushing STANDARD_V1 to STANDARD_V1 token, got %d: %s", wAcceptStd.Code, wAcceptStd.Body.String())
	}

	// 5c. Live-switch Token to AUTO -> BOTH formats succeed!
	wSwitchAuto := doJSONReq(t, env.Router, "PUT", fmt.Sprintf("/api/ingest-keys/%d", keyID), env.SuperAdminToken, nil, map[string]interface{}{
		"ingest_format": "AUTO",
	})
	if wSwitchAuto.Code != http.StatusOK {
		t.Fatalf("switch token ingest_format to AUTO failed (%d): %s", wSwitchAuto.Code, wSwitchAuto.Body.String())
	}
	if wBoth1 := doJSONReq(t, env.Router, "POST", "/api/ticks/create_multi", customBearer, nil, legacyPayload); wBoth1.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200 on /api/ticks/create_multi in AUTO mode, got %d: %s", wBoth1.Code, wBoth1.Body.String())
	}
	if wBoth2 := doJSONReq(t, env.Router, "POST", "/api/v1/ingest/telemetry", customBearer, nil, stdPayload); wBoth2.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200 on /api/v1/ingest/telemetry in AUTO mode, got %d: %s", wBoth2.Code, wBoth2.Body.String())
	}

	// 5d. Station-level switch: set Site ingest_format to CBI_LEGACY -> blocks STANDARD_V1 even when token is AUTO
	wSiteSwitch := doJSONReq(t, env.Router, "PUT", fmt.Sprintf("/api/sites/%d", siteID), env.SuperAdminToken, nil, map[string]interface{}{
		"ingest_format": "CBI_LEGACY",
	})
	if wSiteSwitch.Code != http.StatusOK {
		t.Fatalf("switch site ingest_format failed (%d): %s", wSiteSwitch.Code, wSiteSwitch.Body.String())
	}
	if wSiteBlock := doJSONReq(t, env.Router, "POST", "/api/v1/ingest/telemetry", customBearer, nil, stdPayload); wSiteBlock.Code != http.StatusBadRequest {
		t.Fatalf("expected HTTP 400 when station ingest_format=CBI_LEGACY receives STANDARD_V1, got %d: %s", wSiteBlock.Code, wSiteBlock.Body.String())
	}
}

func TestCloudMode_LegacyTokenPreservation_ReplaceAndGlobalMultiTenant(t *testing.T) {
	env := setupMultiTenantTestEnv(t)

	if g := config.Get(); g != nil {
		g.CloudMode = true
		g.MonitoringDomain = "CLOUD"
		g.SensorPollingEnabled = false
	}

	// Create Tenant A & Tenant B
	createTenantAndStation := func(tenantName, tenantCode, siteCode, legacyStaID, devCode, legacyModID string) (uint, uint, uint) {
		wT := doJSONReq(t, env.Router, "POST", "/api/tenants", env.SuperAdminToken, nil, map[string]interface{}{
			"name":   tenantName,
			"code":   tenantCode,
			"status": "ACTIVE",
		})
		if wT.Code != http.StatusCreated && wT.Code != http.StatusOK {
			t.Fatalf("create tenant %s failed (%d): %s", tenantCode, wT.Code, wT.Body.String())
		}
		var tr struct {
			Data model.Tenant `json:"data"`
		}
		_ = json.Unmarshal(wT.Body.Bytes(), &tr)
		tid := tr.Data.ID
		if tid <= 1 {
			t.Fatalf("expected distinct new tenant ID > 1, got %d (%s)", tid, wT.Body.String())
		}

		wS := doJSONReq(t, env.Router, "POST", "/api/sites", env.SuperAdminToken, nil, map[string]interface{}{
			"tenant_id":         tid,
			"name":              "Stasiun " + legacyStaID,
			"code":              siteCode,
			"monitoring_domain": "AQMS",
			"ingest_format":     "AUTO",
			"legacy_station_id": legacyStaID,
			"timezone":          "Asia/Jakarta",
			"enabled":           true,
		})
		if wS.Code != http.StatusCreated && wS.Code != http.StatusOK {
			t.Fatalf("create site %s failed (%d): %s", siteCode, wS.Code, wS.Body.String())
		}
		var sr struct {
			Data model.Site `json:"data"`
		}
		_ = json.Unmarshal(wS.Body.Bytes(), &sr)
		sid := sr.Data.ID

		tHeaders := map[string]string{
			"X-Tenant-ID": fmt.Sprintf("%d", tid),
			"X-Site-ID":   fmt.Sprintf("%d", sid),
		}
		wD := doJSONReq(t, env.Router, "POST", "/api/devices", env.SuperAdminToken, tHeaders, map[string]interface{}{
			"tenant_id":        tid,
			"site_id":          sid,
			"device_code":      devCode,
			"device_name":      "Logger " + devCode,
			"device_type":      "EDGE_PUSH",
			"legacy_module_id": legacyModID,
			"status":           "ACTIVE",
			"enabled":          true,
		})
		if wD.Code != http.StatusCreated && wD.Code != http.StatusOK {
			t.Fatalf("create device %s failed (%d): %s", devCode, wD.Code, wD.Body.String())
		}
		var dr struct {
			Data model.Device `json:"data"`
		}
		_ = json.Unmarshal(wD.Body.Bytes(), &dr)
		did := dr.Data.ID

		wP := doJSONReq(t, env.Router, "POST", fmt.Sprintf("/api/devices/%d/parameters", did), env.SuperAdminToken, tHeaders, map[string]interface{}{
			"parameter_code": "pm2.5",
			"parameter_name": "PM2.5",
			"unit":           "ug/m3",
			"category":       "TELEMETRY",
			"data_type":      "FLOAT32",
			"scale":          1,
			"precision":      2,
			"enabled":        true,
		})
		if wP.Code != http.StatusCreated && wP.Code != http.StatusOK {
			t.Fatalf("create param pm2.5 on %s failed (%d): %s", devCode, wP.Code, wP.Body.String())
		}
		return tid, sid, did
	}

	tenantAID, siteAID, devAID := createTenantAndStation("DLH Kota A", "dlh-kota-a", "sta-a-202", "202", "LOGGER-A-199", "199")
	tenantBID, _, devBID := createTenantAndStation("DLH Kota B", "dlh-kota-b", "sta-b-305", "305", "LOGGER-B-288", "288")

	// Scenario 1: Key initially created as generated token, then Admin pastes old field logger token via PUT /api/ingest-keys/:id
	wGenKey := doJSONReq(t, env.Router, "POST", "/api/ingest-keys", env.SuperAdminToken, nil, map[string]interface{}{
		"tenant_id":     tenantAID,
		"site_id":       siteAID,
		"name":          "Logger A Key",
		"ingest_format": "CBI_LEGACY",
	})
	if wGenKey.Code != http.StatusCreated && wGenKey.Code != http.StatusOK {
		t.Fatalf("create initial generated key failed (%d): %s", wGenKey.Code, wGenKey.Body.String())
	}
	var genKeyResp struct {
		Data struct {
			Key model.TenantAPIKey `json:"key"`
		} `json:"data"`
	}
	_ = json.Unmarshal(wGenKey.Body.Bytes(), &genKeyResp)

	oldFieldTokenA := "old_field_tok_a_9911"
	wReplace := doJSONReq(t, env.Router, "PUT", fmt.Sprintf("/api/ingest-keys/%d", genKeyResp.Data.Key.ID), env.SuperAdminToken, nil, map[string]interface{}{
		"custom_token": "Bearer " + oldFieldTokenA,
	})
	if wReplace.Code != http.StatusOK {
		t.Fatalf("replace key with old field token failed (%d): %s", wReplace.Code, wReplace.Body.String())
	}

	tsStr := time.Now().UTC().Add(-30 * time.Second).Format(time.RFC3339)
	payloadA := map[string]interface{}{
		"multi": []map[string]interface{}{
			{"tick": map[string]interface{}{"module_id": 199, "station_id": 202, "sensor_name": "pm25", "value": 41.25, "created_at": tsStr, "updated_at": tsStr}},
		},
	}

	// Verify Rails-style `Authorization: Token token="<old_token>"` works without touching the field datalogger!
	wPushRailsHdr := doJSONReq(t, env.Router, "POST", "/api/v1/ticks/create_multi", "", map[string]string{
		"Authorization": fmt.Sprintf(`Token token="%s"`, oldFieldTokenA),
	}, payloadA)
	if wPushRailsHdr.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200 with Rails Authorization: Token token=\"...\" header, got %d: %s", wPushRailsHdr.Code, wPushRailsHdr.Body.String())
	}

	// Scenario 2: Global Shared Master Legacy Token (`is_global: true`) used by multiple dataloggers across Tenant A & Tenant B
	sharedMasterLegacyToken := "cbi_master_shared_2019"
	wGlobalKey := doJSONReq(t, env.Router, "POST", "/api/ingest-keys", env.SuperAdminToken, nil, map[string]interface{}{
		"is_global":     true,
		"name":          "Master Legacy CBI Cloud Token",
		"ingest_format": "CBI_LEGACY",
		"custom_token":  sharedMasterLegacyToken,
	})
	if wGlobalKey.Code != http.StatusCreated && wGlobalKey.Code != http.StatusOK {
		t.Fatalf("create global master legacy token failed (%d): %s", wGlobalKey.Code, wGlobalKey.Body.String())
	}

	payloadGlobalA := map[string]interface{}{
		"multi": []map[string]interface{}{
			{"tick": map[string]interface{}{"module_id": 199, "station_id": 202, "sensor_name": "pm25", "value": 52.5, "created_at": tsStr, "updated_at": tsStr}},
		},
	}
	payloadGlobalB := map[string]interface{}{
		"multi": []map[string]interface{}{
			{"tick": map[string]interface{}{"module_id": 288, "station_id": 305, "sensor_name": "pm25", "value": 19.75, "created_at": tsStr, "updated_at": tsStr}},
		},
	}

	wGA := doJSONReq(t, env.Router, "POST", "/api/v1/ticks/create_multi", sharedMasterLegacyToken, nil, payloadGlobalA)
	if wGA.Code != http.StatusOK {
		t.Fatalf("global legacy token push for Tenant A (station_id=202) failed (%d): %s", wGA.Code, wGA.Body.String())
	}
	wGB := doJSONReq(t, env.Router, "POST", "/api/v1/ticks/create_multi", sharedMasterLegacyToken, nil, payloadGlobalB)
	if wGB.Code != http.StatusOK {
		t.Fatalf("global legacy token push for Tenant B (station_id=305) failed (%d): %s", wGB.Code, wGB.Body.String())
	}

	var paramA, paramB model.Parameter
	_ = env.DB.Where("device_id = ? AND parameter_code = ?", devAID, "pm2.5").First(&paramA).Error
	_ = env.DB.Where("device_id = ? AND parameter_code = ?", devBID, "pm2.5").First(&paramB).Error
	if paramA.CurrentValue == nil || *paramA.CurrentValue != 52.5 {
		t.Fatalf("Tenant A pm2.5 expected 52.5, got %v", paramA.CurrentValue)
	}
	if paramB.CurrentValue == nil || *paramB.CurrentValue != 19.75 {
		t.Fatalf("Tenant B pm2.5 expected 19.75, got %v", paramB.CurrentValue)
	}

	// Verify suspended tenant is still strictly blocked even when pushing via Global Master Legacy Token
	wSuspendB := doJSONReq(t, env.Router, "PUT", fmt.Sprintf("/api/tenants/%d", tenantBID), env.SuperAdminToken, nil, map[string]interface{}{
		"status": "SUSPENDED",
	})
	if wSuspendB.Code != http.StatusOK {
		t.Fatalf("suspend Tenant B failed (%d): %s", wSuspendB.Code, wSuspendB.Body.String())
	}
	wBlockedB := doJSONReq(t, env.Router, "POST", "/api/v1/ticks/create_multi", sharedMasterLegacyToken, nil, payloadGlobalB)
	if wBlockedB.Code != http.StatusForbidden {
		t.Fatalf("expected HTTP 403 when suspended Tenant B pushes via global token, got %d: %s", wBlockedB.Code, wBlockedB.Body.String())
	}
}
