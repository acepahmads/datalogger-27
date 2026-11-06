package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"datalogger/internal/aggregation"
	"datalogger/internal/calibration"
	"datalogger/internal/handler"
	"datalogger/internal/middleware"
	"datalogger/internal/model"
	"datalogger/internal/repository"
	"datalogger/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func setupCalibrationTestSuite(t *testing.T) (*gin.Engine, *gorm.DB, *service.TelemetryService, *aggregation.Engine) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	dsn := fmt.Sprintf("file:memcal_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		t.Fatalf("failed to open sqlite db: %v", err)
	}

	if err := db.AutoMigrate(
		&model.DeviceType{},
		&model.Device{},
		&model.DeviceConnection{},
		&model.ParameterCategory{},
		&model.Parameter{},
		&model.SensorCalibrationLog{},
		&model.SystemBranding{},
		&model.AuditTrail{},
		&model.SystemLog{},
		&model.RawData{},
		&model.AggregationDefinition{},
		&model.AggregationResult{},
		&model.Alarm{},
	); err != nil {
		t.Fatalf("failed to migrate models: %v", err)
	}

	deviceRepo := repository.NewDeviceRepository(db)
	systemRepo := repository.NewSystemRepository(db)
	calRepo := repository.NewCalibrationRepository(db)
	telRepo := repository.NewTelemetryRepository(db)
	alarmRepo := repository.NewAlarmRepository(db)

	aggEngine := aggregation.NewEngine()
	alarmService := service.NewAlarmService(alarmRepo, systemRepo)
	alarmEval := service.NewAlarmEvaluator(
		context.Background(),
		alarmService,
		service.DefaultAlarmEvaluatorConfig(),
		func(ctx context.Context, paramID uint) (*model.Parameter, error) {
			return deviceRepo.FindParameterByID(paramID)
		},
	)

	telService := service.NewTelemetryService(telRepo, nil, service.TelemetryConfig{
		BufferSize:    100,
		BatchSize:     1,
		FlushInterval: 20 * time.Millisecond,
		QueueEnabled:  false,
	})
	t.Cleanup(func() {
		telService.Stop()
	})
	telService.SetDeviceLookup(deviceRepo.GetByID)
	telService.SetParamLookup(deviceRepo.FindParameterByID)
	telService.SetAlarmEvaluator(alarmEval)

	calService := service.NewCalibrationService(calRepo, deviceRepo, systemRepo, telService)
	calHandler := handler.NewCalibrationHandler(calService)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		role := c.GetHeader("X-Test-Role")
		user := c.GetHeader("X-Test-User")
		if role != "" {
			c.Set("role", role)
		}
		if user != "" {
			c.Set("username", user)
		}
		c.Next()
	})

	calRoutes := r.Group("/api/calibrations")
	{
		calRoutes.GET("/parameters", middleware.RequirePermission("calibration.view"), calHandler.ListParameters)
		calRoutes.GET("/parameters/:paramId/live", middleware.RequirePermission("calibration.view"), calHandler.GetLiveReading)
		calRoutes.POST("/calculate", middleware.RequirePermission("calibration.view"), calHandler.CalculatePreview)
		calRoutes.GET("/logs", middleware.RequirePermission("calibration.view"), calHandler.ListLogs)
		calRoutes.PUT("/parameters/:paramId/config", middleware.RequirePermission("calibration.manage"), calHandler.UpdateConfig)
		calRoutes.PUT("/parameters/:paramId/session", middleware.RequirePermission("calibration.manage"), calHandler.ToggleSession)
		calRoutes.POST("/parameters/:paramId/apply", middleware.RequirePermission("calibration.manage"), calHandler.ApplyCalibration)
		calRoutes.POST("/parameters/:paramId/reset", middleware.RequirePermission("calibration.manage"), calHandler.ResetCalibration)
		calRoutes.POST("/logs/:logId/rollback", middleware.RequirePermission("calibration.manage"), calHandler.RollbackLog)
	}

	return r, db, telService, aggEngine
}

func performJSONReq(r *gin.Engine, method, path, role, user string, body interface{}) *httptest.ResponseRecorder {
	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	if role != "" {
		req.Header.Set("X-Test-Role", role)
	}
	if user != "" {
		req.Header.Set("X-Test-User", user)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestCalibrationHandler_RBAC_Domains_PerParamR2_ApplyResetRollback_SessionGuard(t *testing.T) {
	router, db, telService, aggEngine := setupCalibrationTestSuite(t)

	// Seed device + 3 parameters representing AQMS, SPARING, and WQMS ONLIMO
	dev := model.Device{
		DeviceCode: "DEV-MULTI-01",
		DeviceName: "Analyzer Rack",
		Status:     model.DeviceStatusActive,
		Enabled:    true,
	}
	if err := db.Create(&dev).Error; err != nil {
		t.Fatalf("create device failed: %v", err)
	}

	uncalSO2 := 245.5
	critHigh := 400.0
	so2Param := model.Parameter{
		DeviceID:                 dev.ID,
		ParameterCode:            "SO2",
		ParameterName:            "Sulfur Dioxide",
		DataType:                 model.DataTypeFloat32,
		Unit:                     "µg/m³",
		Scale:                    0.1,  // Base Modbus register scale (MUST NOT be touched by calibration reset)
		Offset:                   -5.0, // Base Modbus register offset
		Precision:                2,
		CriticalHigh:             &critHigh,
		Enabled:                  true,
		CalMinR2:                 0.995,
		CurrentValue:             &uncalSO2,
		CurrentUncalibratedValue: &uncalSO2,
	}
	codParam := model.Parameter{
		DeviceID:          dev.ID,
		ParameterCode:     "COD",
		ParameterName:     "Chemical Oxygen Demand",
		DataType:          model.DataTypeFloat32,
		Unit:              "mg/L",
		Scale:             1.0,
		Enabled:           true,
		CalibrationDomain: model.CalibrationDomainSparing,
		CalMinR2:          0.995,
	}
	doParam := model.Parameter{
		DeviceID:      dev.ID,
		ParameterCode: "DO",
		ParameterName: "Dissolved Oxygen",
		DataType:      model.DataTypeFloat32,
		Unit:          "mg/L",
		Scale:         0.01,
		Enabled:       true,
		CalMinR2:      0.995,
	}
	if err := db.Create(&so2Param).Error; err != nil {
		t.Fatalf("create SO2 param: %v", err)
	}
	if err := db.Create(&codParam).Error; err != nil {
		t.Fatalf("create COD param: %v", err)
	}
	if err := db.Create(&doParam).Error; err != nil {
		t.Fatalf("create DO param: %v", err)
	}

	// Seed a disabled device + disabled parameter to ensure they are excluded from /api/calibrations/parameters
	disabledDev := model.Device{
		DeviceCode: "DEV-DISABLED-99",
		DeviceName: "QA CEK Disabled",
		Status:     model.DeviceStatusDisabled,
		Enabled:    false,
	}
	_ = db.Create(&disabledDev).Error
	_ = db.Model(&disabledDev).Update("enabled", false).Error
	_ = db.Create(&model.Parameter{
		DeviceID:      disabledDev.ID,
		ParameterCode: "HIDDEN_DEV_PARAM",
		ParameterName: "Hidden Disabled Device Param",
		DataType:      model.DataTypeFloat32,
		Enabled:       true,
	}).Error

	disabledParam := model.Parameter{
		DeviceID:      dev.ID,
		ParameterCode: "HIDDEN_PARAM",
		ParameterName: "Hidden Disabled Param",
		DataType:      model.DataTypeFloat32,
		Enabled:       false,
	}
	_ = db.Create(&disabledParam).Error
	_ = db.Model(&disabledParam).Update("enabled", false).Error

	// Verify disabled device parameters and disabled parameters never appear in ListParameters
	wActiveOnly := performJSONReq(router, http.MethodGet, "/api/calibrations/parameters", "ENGINEER", "eng_budi", nil)
	if wActiveOnly.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", wActiveOnly.Code)
	}
	var activeOnlyResp struct {
		Data struct {
			Items []service.CalibratableParameterDTO `json:"items"`
		} `json:"data"`
	}
	_ = json.Unmarshal(wActiveOnly.Body.Bytes(), &activeOnlyResp)
	for _, it := range activeOnlyResp.Data.Items {
		if it.ParameterCode == "HIDDEN_DEV_PARAM" || it.ParameterCode == "HIDDEN_PARAM" {
			t.Fatalf("disabled device/parameter leaked into calibration parameters list: %+v", it)
		}
	}

	// 1. RBAC Verification: OPERATOR is strictly forbidden (403), while ENGINEER, ADMIN, SUPERADMIN, and custom non-Operator role succeed (200)
	wOp := performJSONReq(router, http.MethodGet, "/api/calibrations/parameters", "OPERATOR", "op_shift1", nil)
	if wOp.Code != http.StatusForbidden {
		t.Fatalf("expected OPERATOR to get 403 Forbidden, got %d", wOp.Code)
	}

	for _, allowedRole := range []string{"ENGINEER", "ADMIN", "ADMINISTRATOR", "SUPERADMIN", "QC_SPECIALIST"} {
		w := performJSONReq(router, http.MethodGet, "/api/calibrations/parameters", allowedRole, "user_"+allowedRole, nil)
		if w.Code != http.StatusOK {
			t.Fatalf("expected role %s to get 200 OK, got %d: %s", allowedRole, w.Code, w.Body.String())
		}
	}

	// 2. Domain Filtering Verification: AQMS, SPARING, WQMS_ONLIMO
	checkDomainCount := func(domain string, expectedCode string) {
		w := performJSONReq(router, http.MethodGet, "/api/calibrations/parameters?domain="+domain, "ENGINEER", "eng_budi", nil)
		if w.Code != http.StatusOK {
			t.Fatalf("domain %s failed: %d", domain, w.Code)
		}
		var resp struct {
			Data struct {
				Items []service.CalibratableParameterDTO `json:"items"`
			} `json:"data"`
		}
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		found := false
		for _, it := range resp.Data.Items {
			if it.ParameterCode == expectedCode {
				found = true
			}
		}
		if !found {
			t.Fatalf("expected parameter %s to appear under domain filter %s, got %+v", expectedCode, domain, resp.Data.Items)
		}
	}
	checkDomainCount("AQMS", "SO2")
	checkDomainCount("SPARING", "COD")
	checkDomainCount("WQMS_ONLIMO", "DO")

	// 3. Per-Parameter Minimum R2 Configuration (PUT /api/calibrations/parameters/:id/config)
	customMinR2 := 0.9985
	wCfg := performJSONReq(
		router,
		http.MethodPut,
		fmt.Sprintf("/api/calibrations/parameters/%d/config", so2Param.ID),
		"ENGINEER",
		"eng_budi",
		map[string]interface{}{
			"cal_min_r2": customMinR2,
			"domain":     "AQMS",
		},
	)
	if wCfg.Code != http.StatusOK {
		t.Fatalf("expected 200 OK updating per-parameter cal_min_r2, got %d: %s", wCfg.Code, wCfg.Body.String())
	}
	var updatedSO2 model.Parameter
	_ = db.First(&updatedSO2, so2Param.ID).Error
	if updatedSO2.CalMinR2 != customMinR2 {
		t.Fatalf("expected DB CalMinR2=%f, got %f", customMinR2, updatedSO2.CalMinR2)
	}

	// 4. Calibration Session Guard: suppresses threshold alarms & protects AQMS/SPARING/ONLIMO CUSTOMER_PROCESSED aggregation
	bucketBase := time.Now().UTC().Add(-2 * time.Minute)
	normalVal := 42.0
	if err := telService.Ingest(&service.TelemetryIngestPayload{
		DeviceID:          dev.ID,
		DeviceCode:        dev.DeviceCode,
		ParameterID:       so2Param.ID,
		ParameterCode:     "SO2",
		DataType:          model.DataTypeFloat32,
		Value:             normalVal,
		UncalibratedValue: &normalVal,
		RawValue:          470,
		Quality:           model.QualityGood,
		ReceivedAt:        bucketBase.Add(15 * time.Second),
	}); err != nil {
		t.Fatalf("ingest normal sample failed: %v", err)
	}

	// Activate Mode Sesi Kalibrasi
	wSessOn := performJSONReq(
		router,
		http.MethodPut,
		fmt.Sprintf("/api/calibrations/parameters/%d/session", so2Param.ID),
		"ENGINEER",
		"eng_budi",
		map[string]interface{}{"active": true},
	)
	if wSessOn.Code != http.StatusOK {
		t.Fatalf("toggle session ON failed: %d %s", wSessOn.Code, wSessOn.Body.String())
	}

	// Inject high span gas = 800.0 µg/m³ (above CriticalHigh = 400.0!) during active calibration session
	spanGasVal := 800.0
	if err := telService.Ingest(&service.TelemetryIngestPayload{
		DeviceID:          dev.ID,
		DeviceCode:        dev.DeviceCode,
		ParameterID:       so2Param.ID,
		ParameterCode:     "SO2",
		DataType:          model.DataTypeFloat32,
		Value:             spanGasVal,
		UncalibratedValue: &spanGasVal,
		RawValue:          8050,
		Quality:           model.QualityGood,
		ReceivedAt:        bucketBase.Add(45 * time.Second),
	}); err != nil {
		t.Fatalf("ingest calibration span gas failed: %v", err)
	}

	// Wait briefly for async persistence flush
	time.Sleep(80 * time.Millisecond)

	// Verify NO false critical-high alarm was triggered by the 800 µg/m³ span gas
	var activeAlarmCount int64
	db.Model(&model.Alarm{}).Where("parameter_id = ? AND status = ?", so2Param.ID, model.AlarmActive).Count(&activeAlarmCount)
	if activeAlarmCount != 0 {
		t.Fatalf("expected 0 active alarms during calibration session, got %d", activeAlarmCount)
	}

	// Fetch persisted RawData samples and verify CUSTOMER_PROCESSED compliance bucket excludes calibration session gas
	var persistedSamples []model.RawData
	db.Where("parameter_id = ?", so2Param.ID).Order("received_at ASC").Find(&persistedSamples)
	if len(persistedSamples) != 2 {
		t.Fatalf("expected 2 persisted RawData samples, got %d", len(persistedSamples))
	}
	if persistedSamples[1].QualityReason != model.ReasonCalibrationMode {
		t.Fatalf("expected second sample QualityReason=%s, got %s", model.ReasonCalibrationMode, persistedSamples[1].QualityReason)
	}

	aggDef := &model.AggregationDefinition{
		ID:              1,
		DeviceID:        dev.ID,
		ParameterID:     so2Param.ID,
		SourceType:      model.SourceCustomerProcessed,
		Function:        model.FunctionAvg,
		IntervalSeconds: 300,
		Timezone:        "UTC",
	}
	aggRes := aggEngine.Calculate(aggDef, bucketBase, bucketBase.Add(5*time.Minute), persistedSamples)
	if aggRes.Value == nil || *aggRes.Value != 42.0 {
		t.Fatalf("expected compliance CUSTOMER_PROCESSED bucket average to remain 42.0 (excluding 800 span gas), got %v", aggRes.Value)
	}

	// 5. Apply 5-Point Linear Calibration & End Session
	applyBody := service.ApplyCalibrationRequest{
		Model:        calibration.ModelLinear,
		Domain:       model.CalibrationDomainAQMS,
		StandardNote: "Gas Standar SO2 Cylinder #CRM-2026-88",
		EndSession:   true,
		Points: []calibration.CalibrationPoint{
			{Index: 1, Active: true, Label: "Zero Air", RawX: 0.4, RefY: 0.0},
			{Index: 2, Active: true, Label: "Span 100", RawX: 98.2, RefY: 100.0},
			{Index: 3, Active: true, Label: "Span 250", RawX: 245.5, RefY: 250.0},
			{Index: 4, Active: true, Label: "Span 500", RawX: 490.8, RefY: 500.0},
			{Index: 5, Active: true, Label: "Span 800", RawX: 785.0, RefY: 800.0},
		},
	}
	wApply := performJSONReq(
		router,
		http.MethodPost,
		fmt.Sprintf("/api/calibrations/parameters/%d/apply", so2Param.ID),
		"ENGINEER",
		"eng_budi",
		applyBody,
	)
	if wApply.Code != http.StatusOK {
		t.Fatalf("apply calibration failed: %d %s", wApply.Code, wApply.Body.String())
	}

	var appliedResp struct {
		Data struct {
			Parameter service.CalibratableParameterDTO `json:"parameter"`
			Log       model.SensorCalibrationLog       `json:"log"`
		} `json:"data"`
	}
	_ = json.Unmarshal(wApply.Body.Bytes(), &appliedResp)
	if !appliedResp.Data.Parameter.CalibrationEnabled || appliedResp.Data.Parameter.CalibrationSessionActive {
		t.Fatalf("expected calibration_enabled=true and session ended=false, got %+v", appliedResp.Data.Parameter)
	}
	applyLogID := appliedResp.Data.Log.ID

	// 6. Reset to Default (y = x) MUST preserve base Modbus Scale (0.1) & Offset (-5.0)
	wReset := performJSONReq(
		router,
		http.MethodPost,
		fmt.Sprintf("/api/calibrations/parameters/%d/reset", so2Param.ID),
		"ADMIN",
		"admin_sari",
		map[string]string{"reason_note": "Ganti modul optik baru"},
	)
	if wReset.Code != http.StatusOK {
		t.Fatalf("reset calibration failed: %d %s", wReset.Code, wReset.Body.String())
	}

	var afterReset model.Parameter
	_ = db.First(&afterReset, so2Param.ID).Error
	if afterReset.CalibrationEnabled {
		t.Fatalf("expected CalibrationEnabled=false after reset")
	}
	if afterReset.Scale != 0.1 || afterReset.Offset != -5.0 {
		t.Fatalf("CRITICAL: reset altered base Modbus Scale/Offset! got scale=%f offset=%f", afterReset.Scale, afterReset.Offset)
	}

	// 7. Rollback to the 5-Point Linear Calibration Log
	wRb := performJSONReq(
		router,
		http.MethodPost,
		fmt.Sprintf("/api/calibrations/logs/%d/rollback", applyLogID),
		"SUPERADMIN",
		"super_andi",
		nil,
	)
	if wRb.Code != http.StatusOK {
		t.Fatalf("rollback calibration failed: %d %s", wRb.Code, wRb.Body.String())
	}

	var afterRollback model.Parameter
	_ = db.First(&afterRollback, so2Param.ID).Error
	if !afterRollback.CalibrationEnabled || afterRollback.CalPointCount != 5 {
		t.Fatalf("expected CalibrationEnabled=true & 5 points restored after rollback, got %+v", afterRollback)
	}
}

func TestCalibration_ClearsCustomFormula_PreventsDoubleFormula_AndPersistsUncalibratedRawData(t *testing.T) {
	router, db, telService, _ := setupCalibrationTestSuite(t)

	dev := model.Device{
		DeviceCode: "DEV-FORMULA-01",
		DeviceName: "Water Analyzer",
		Status:     model.DeviceStatusActive,
		Enabled:    true,
	}
	if err := db.Create(&dev).Error; err != nil {
		t.Fatalf("create device failed: %v", err)
	}

	fVal := 8.25
	phParam := model.Parameter{
		DeviceID:                 dev.ID,
		ParameterCode:            "PH",
		ParameterName:            "pH Air",
		DataType:                 model.DataTypeFloat32,
		Unit:                     "-",
		Scale:                    1.0,
		Enabled:                  true,
		Formula:                  "x * 1.15 + 0.5",
		CurrentFormulaValue:      &fVal,
		CalMinR2:                 0.995,
		CalibrationDomain:        model.CalibrationDomainSparing,
	}
	if err := db.Create(&phParam).Error; err != nil {
		t.Fatalf("create pH param: %v", err)
	}

	// 1. Starting calibration session MUST automatically clear pre-existing param.Formula
	wSess := performJSONReq(
		router,
		http.MethodPut,
		fmt.Sprintf("/api/calibrations/parameters/%d/session", phParam.ID),
		"ENGINEER",
		"eng_rina",
		map[string]bool{"active": true},
	)
	if wSess.Code != http.StatusOK {
		t.Fatalf("toggle session failed: %d %s", wSess.Code, wSess.Body.String())
	}

	var afterSess model.Parameter
	_ = db.First(&afterSess, phParam.ID).Error
	if afterSess.Formula != "" || afterSess.CurrentFormulaValue != nil {
		t.Fatalf("expected Formula & CurrentFormulaValue to be cleared when calibration session starts, got formula=%q", afterSess.Formula)
	}

	// 2. Apply calibration also guarantees Formula is cleared even if set right before apply
	_ = db.Model(&model.Parameter{}).Where("id = ?", phParam.ID).Update("formula", "x + 0.25").Error
	applyBody := service.ApplyCalibrationRequest{
		Model: calibration.ModelLinear,
		Points: []calibration.CalibrationPoint{
			{Index: 1, Active: true, Label: "Buffer 4", RawX: 3.90, RefY: 4.01},
			{Index: 2, Active: true, Label: "Buffer 7", RawX: 6.88, RefY: 7.00},
			{Index: 3, Active: true, Label: "Buffer 10", RawX: 9.85, RefY: 10.01},
		},
		EndSession: true,
	}
	wApply := performJSONReq(
		router,
		http.MethodPost,
		fmt.Sprintf("/api/calibrations/parameters/%d/apply", phParam.ID),
		"ENGINEER",
		"eng_rina",
		applyBody,
	)
	if wApply.Code != http.StatusOK {
		t.Fatalf("apply calibration failed: %d %s", wApply.Code, wApply.Body.String())
	}

	var afterApply model.Parameter
	_ = db.First(&afterApply, phParam.ID).Error
	if !afterApply.CalibrationEnabled || afterApply.Formula != "" {
		t.Fatalf("expected CalibrationEnabled=true and Formula=\"\", got enabled=%v formula=%q", afterApply.CalibrationEnabled, afterApply.Formula)
	}

	// 3. DeviceService.UpdateParameter MUST reject non-empty custom Formula while CalibrationEnabled=true
	devRepo := repository.NewDeviceRepository(db)
	sysRepo := repository.NewSystemRepository(db)
	devSvc := service.NewDeviceService(devRepo, sysRepo)
	customFormula := "x * 2"
	_, updErr := devSvc.UpdateParameter(dev.ID, phParam.ID, &service.UpdateParameterRequest{
		Formula: &customFormula,
	}, "admin", "127.0.0.1", "test")
	if updErr == nil {
		t.Fatalf("expected UpdateParameter to reject non-empty Formula when CalibrationEnabled is true")
	}

	// 4. TelemetryService.Ingest MUST persist UncalibratedValue alongside calibrated Value in raw_data
	uncalReading := 6.88
	calReading := afterApply.ApplyCalibration(uncalReading)
	if err := telService.Ingest(&service.TelemetryIngestPayload{
		DeviceID:          dev.ID,
		DeviceCode:        dev.DeviceCode,
		DeviceName:        dev.DeviceName,
		ParameterID:       afterApply.ID,
		ParameterCode:     afterApply.ParameterCode,
		ParameterName:     afterApply.ParameterName,
		Unit:              afterApply.Unit,
		DataType:          afterApply.DataType,
		RawValue:          688,
		UncalibratedValue: &uncalReading,
		Value:             calReading,
		Quality:           model.QualityGood,
		Source:            "MODBUS_RTU",
		ReceivedAt:        time.Now().UTC(),
		Parameter:         &afterApply,
	}); err != nil {
		t.Fatalf("ingest failed: %v", err)
	}

	time.Sleep(80 * time.Millisecond)

	var rawRec model.RawData
	if err := db.Where("parameter_id = ?", afterApply.ID).Order("id desc").First(&rawRec).Error; err != nil {
		t.Fatalf("expected persisted raw_data row: %v", err)
	}
	if rawRec.UncalibratedValue == nil || *rawRec.UncalibratedValue != uncalReading {
		t.Fatalf("expected raw_data.uncalibrated_value=%.4f, got %v", uncalReading, rawRec.UncalibratedValue)
	}
	if rawRec.Value == uncalReading {
		t.Fatalf("expected raw_data.value (%.4f) to differ from uncalibrated_value (%.4f)", rawRec.Value, uncalReading)
	}
}
