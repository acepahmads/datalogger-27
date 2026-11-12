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
