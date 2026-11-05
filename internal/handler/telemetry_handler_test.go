package handler_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"datalogger/internal/config"
	"datalogger/internal/database"
	"datalogger/internal/handler"
	"datalogger/internal/middleware"
	"datalogger/internal/model"
	"datalogger/internal/repository"
	"datalogger/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

type TelemetryTestEnv struct {
	Router           *gin.Engine
	DB               *gorm.DB
	TelemetryService *service.TelemetryService
	AdminToken       string
	OperatorToken    string
}

func setupTelemetryTestRouter(t *testing.T) *TelemetryTestEnv {
	gin.SetMode(gin.TestMode)

	dsn := fmt.Sprintf("file:mem_telem_api_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	_ = database.RunMigrations(db)

	cfg := config.DefaultConfig()

	sysRepo := repository.NewSystemRepository(db)
	authService := service.NewAuthService(sysRepo, cfg)
	telemetryRepo := repository.NewTelemetryRepository(db)
	telemetryService := service.NewTelemetryService(telemetryRepo, nil, service.DefaultTelemetryConfig())

	telemetryHandler := handler.NewTelemetryHandler(telemetryService)

	// Seed roles & users
	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	adminRole := model.Role{Name: "Administrator", Description: "Full access"}
	opRole := model.Role{Name: "OPERATOR", Description: "Read access"}
	db.Create(&adminRole)
	db.Create(&opRole)

	adminUser := model.User{
		Username: "admin_telem",
		Email:    "admin_telem@test.local",
		Password: string(hash),
		RoleID:   adminRole.ID,
		IsActive: true,
	}
	db.Create(&adminUser)

	opUser := model.User{
		Username: "op_telem",
		Email:    "op_telem@test.local",
		Password: string(hash),
		RoleID:   opRole.ID,
		IsActive: true,
	}
	db.Create(&opUser)

	adminToken, _, _ := authService.Login("admin_telem", "password123", "127.0.0.1", "")
	opToken, _, _ := authService.Login("op_telem", "password123", "127.0.0.1", "")

	r := gin.New()
	r.Use(gin.Recovery())

	api := r.Group("/api")
	{
		devicesGroup := api.Group("/devices")
		devicesGroup.Use(middleware.JWTAuth(authService))
		{
			devicesGroup.GET("/:id/telemetry/latest", middleware.RequirePermission("device.view"), telemetryHandler.GetLatestForDevice)
			devicesGroup.GET("/:id/parameters/:paramId/telemetry/latest", middleware.RequirePermission("device.view"), telemetryHandler.GetLatestForParameter)
			devicesGroup.GET("/:id/telemetry/history", middleware.RequirePermission("device.view"), telemetryHandler.GetHistorical)
			devicesGroup.GET("/:id/telemetry/raw", middleware.RequirePermission("device.view"), telemetryHandler.GetRawTelemetry)
			devicesGroup.GET("/:id/telemetry/quality-summary", middleware.RequirePermission("device.view"), telemetryHandler.GetDeviceQualitySummary)
		}
		telemetryGroup := api.Group("/telemetry")
		telemetryGroup.Use(middleware.JWTAuth(authService))
		{
			telemetryGroup.GET("/history", middleware.RequirePermission("device.view"), telemetryHandler.GetAllHistorical)
			telemetryGroup.GET("/raw", middleware.RequirePermission("device.view"), telemetryHandler.GetAllRawTelemetry)
			telemetryGroup.GET("/metrics", middleware.RequirePermission("device.view"), telemetryHandler.GetMetrics)
			telemetryGroup.GET("/quality-summary", middleware.RequirePermission("device.view"), telemetryHandler.GetQualitySummary)
		}
	}

	return &TelemetryTestEnv{
		Router:           r,
		DB:               db,
		TelemetryService: telemetryService,
		AdminToken:       adminToken,
		OperatorToken:    opToken,
	}
}

func TestTelemetryAPIEndpoints(t *testing.T) {
	env := setupTelemetryTestRouter(t)
	defer env.TelemetryService.Stop()

	// 1. Seed Device & Parameter
	dev := model.Device{ID: 1, DeviceCode: "DEV-01", DeviceName: "Solar Inverter 01"}
	if err := env.DB.Create(&dev).Error; err != nil {
		t.Fatalf("Failed to create device: %v", err)
	}

	curVal := 48.75
	recAt := time.Now().UTC()
	param := model.Parameter{
		ID:                  101,
		DeviceID:            1,
		ParameterCode:       "TEMP_HEATSINK",
		ParameterName:       "Inverter Heatsink Temp",
		Unit:                "°C",
		DataType:            model.DataTypeFloat32,
		CurrentValue:        &curVal,
		CurrentValueNumeric: &curVal,
		CurrentValueText:    "48.75",
		CurrentQuality:      model.QualityGood,
		CurrentRawHex:       "42 43 00 00",
		CurrentReceivedAt:   &recAt,
		LastUpdated:         &recAt,
	}
	if err := env.DB.Create(&param).Error; err != nil {
		t.Fatalf("Failed to create parameter: %v", err)
	}

	// 2. Test GET /api/devices/:id/telemetry/latest (Authorized)
	req, _ := http.NewRequest("GET", "/api/devices/1/telemetry/latest", nil)
	req.Header.Set("Authorization", "Bearer "+env.AdminToken)
	w := httptest.NewRecorder()
	env.Router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK, got %d: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Success bool                      `json:"success"`
		Data    []service.LatestTelemetry `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal latest resp: %v", err)
	}
	if !resp.Success || len(resp.Data) == 0 {
		t.Fatalf("Expected success with latest data")
	}
	if resp.Data[0].Value != 48.75 {
		t.Errorf("Expected latest value 48.75, got %f", resp.Data[0].Value)
	}
	if resp.Data[0].Unit != "°C" {
		t.Errorf("Expected unit °C, got %s", resp.Data[0].Unit)
	}

	// 3. Test GET /api/devices/:id/parameters/:paramId/telemetry/latest
	reqSingle, _ := http.NewRequest("GET", "/api/devices/1/parameters/101/telemetry/latest", nil)
	reqSingle.Header.Set("Authorization", "Bearer "+env.AdminToken)
	wSingle := httptest.NewRecorder()
	env.Router.ServeHTTP(wSingle, reqSingle)

	if wSingle.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK for single param latest, got %d: %s", wSingle.Code, wSingle.Body.String())
	}

	// 4. Test GET /api/devices/:id/telemetry/history
	// Seed historical points
	for i := 1; i <= 10; i++ {
		_ = env.DB.Create(&model.RawData{
			DeviceID:    1,
			ParameterID: 101,
			Value:       float64(40 + i),
			Quality:     model.QualityGood,
			ReceivedAt:  time.Now().UTC().Add(time.Duration(-i) * time.Minute),
			StoredAt:    time.Now().UTC(),
		}).Error
	}

	reqHist, _ := http.NewRequest("GET", "/api/devices/1/telemetry/history?parameter_id=101&page=1&page_size=5", nil)
	reqHist.Header.Set("Authorization", "Bearer "+env.AdminToken)
	wHist := httptest.NewRecorder()
	env.Router.ServeHTTP(wHist, reqHist)

	if wHist.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK for history, got %d: %s", wHist.Code, wHist.Body.String())
	}

	var histResp struct {
		Success bool `json:"success"`
		Data    struct {
			Total    int64            `json:"total"`
			Page     int              `json:"page"`
			PageSize int              `json:"page_size"`
			Items    []model.RawData  `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(wHist.Body.Bytes(), &histResp); err != nil {
		t.Fatalf("Failed to parse history response: %v", err)
	}
	if histResp.Data.Total != 10 {
		t.Errorf("Expected total 10 records, got %d", histResp.Data.Total)
	}
	if len(histResp.Data.Items) != 5 {
		t.Errorf("Expected 5 paginated items, got %d", len(histResp.Data.Items))
	}

	// 5. Test GET /api/devices/:id/telemetry/raw
	reqRaw, _ := http.NewRequest("GET", "/api/devices/1/telemetry/raw?page=1&limit=5", nil)
	reqRaw.Header.Set("Authorization", "Bearer "+env.AdminToken)
	wRaw := httptest.NewRecorder()
	env.Router.ServeHTTP(wRaw, reqRaw)
	if wRaw.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK for raw telemetry, got %d: %s", wRaw.Code, wRaw.Body.String())
	}

	// 6. Test GET /api/telemetry/metrics
	reqMetrics, _ := http.NewRequest("GET", "/api/telemetry/metrics", nil)
	reqMetrics.Header.Set("Authorization", "Bearer "+env.AdminToken)
	wMetrics := httptest.NewRecorder()
	env.Router.ServeHTTP(wMetrics, reqMetrics)
	if wMetrics.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK for metrics, got %d: %s", wMetrics.Code, wMetrics.Body.String())
	}

	// 7. Test GET /api/telemetry/quality-summary (Phase 3.2)
	reqQS, _ := http.NewRequest("GET", "/api/telemetry/quality-summary", nil)
	reqQS.Header.Set("Authorization", "Bearer "+env.AdminToken)
	wQS := httptest.NewRecorder()
	env.Router.ServeHTTP(wQS, reqQS)
	if wQS.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK for quality summary, got %d: %s", wQS.Code, wQS.Body.String())
	}
	var qsResp struct {
		Data service.QualitySummaryDTO `json:"data"`
	}
	if err := json.Unmarshal(wQS.Body.Bytes(), &qsResp); err != nil {
		t.Fatalf("Failed to parse quality summary response: %v", err)
	}
	if qsResp.Data.TotalCount == 0 {
		t.Errorf("Expected active parameters in quality summary, got %d", qsResp.Data.TotalCount)
	}

	// 8. Test GET /api/devices/:id/telemetry/quality-summary (Phase 3.2)
	reqDevQS, _ := http.NewRequest("GET", "/api/devices/1/telemetry/quality-summary", nil)
	reqDevQS.Header.Set("Authorization", "Bearer "+env.AdminToken)
	wDevQS := httptest.NewRecorder()
	env.Router.ServeHTTP(wDevQS, reqDevQS)
	if wDevQS.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK for device quality summary, got %d: %s", wDevQS.Code, wDevQS.Body.String())
	}

	// 9. Test GET /api/telemetry/history across all devices (Unified Analysis)
	reqAllHist, _ := http.NewRequest("GET", "/api/telemetry/history?page=1&page_size=50", nil)
	reqAllHist.Header.Set("Authorization", "Bearer "+env.AdminToken)
	wAllHist := httptest.NewRecorder()
	env.Router.ServeHTTP(wAllHist, reqAllHist)
	if wAllHist.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK for /api/telemetry/history, got %d: %s", wAllHist.Code, wAllHist.Body.String())
	}

	// 10. Test GET /api/telemetry/history with device_ids list
	reqMultiHist, _ := http.NewRequest("GET", "/api/telemetry/history?device_ids=1,2&quality=GOOD", nil)
	reqMultiHist.Header.Set("Authorization", "Bearer "+env.AdminToken)
	wMultiHist := httptest.NewRecorder()
	env.Router.ServeHTTP(wMultiHist, reqMultiHist)
	if wMultiHist.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK for /api/telemetry/history with device_ids, got %d", wMultiHist.Code)
	}

	// 11. Test GET /api/telemetry/raw across all devices (Unified Analysis)
	reqAllRaw, _ := http.NewRequest("GET", "/api/telemetry/raw?page=1&page_size=50", nil)
	reqAllRaw.Header.Set("Authorization", "Bearer "+env.AdminToken)
	wAllRaw := httptest.NewRecorder()
	env.Router.ServeHTTP(wAllRaw, reqAllRaw)
	if wAllRaw.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK for /api/telemetry/raw, got %d: %s", wAllRaw.Code, wAllRaw.Body.String())
	}

	// 12. Test Unauthorized Rejection (no token)
	reqUnauth, _ := http.NewRequest("GET", "/api/devices/1/telemetry/latest", nil)
	wUnauth := httptest.NewRecorder()
	env.Router.ServeHTTP(wUnauth, reqUnauth)
	if wUnauth.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401 Unauthorized, got %d", wUnauth.Code)
	}
}

func TestRawTelemetryOperatorSanitizationAndMaxRangeEnforcement(t *testing.T) {
	env := setupTelemetryTestRouter(t)
	defer env.TelemetryService.Stop()

	cfg := config.Get()
	origMax := cfg.RawTelemetryMaxRangeHours
	cfg.RawTelemetryMaxRangeHours = 48 // 2 days max window for test
	defer func() {
		cfg.RawTelemetryMaxRangeHours = origMax
	}()

	dev := model.Device{ID: 10, DeviceCode: "DEV-OP-RAW", DeviceName: "WQMS Station"}
	_ = env.DB.Create(&dev).Error
	param := model.Parameter{
		ID:            201,
		DeviceID:      10,
		ParameterCode: "PH",
		ParameterName: "Acidity pH",
		Unit:          "pH",
		DataType:      model.DataTypeFloat32,
		Enabled:       true,
	}
	_ = env.DB.Create(&param).Error

	// Seed a record 10 days ago (to verify historical window inspection works when span <= 48h)
	histTime := time.Now().UTC().Add(-10 * 24 * time.Hour).Truncate(time.Second)
	_ = env.DB.Create(&model.RawData{
		DeviceID:       10,
		ParameterID:    201,
		Value:          6.92,
		ProcessedValue: 6.92,
		RawValue:       692,
		RawHex:         "02 B4",
		Quality:        model.QualityGood,
		Timestamp:      histTime,
		ReceivedAt:     histTime,
		StoredAt:       histTime,
	}).Error

	// 1. Operator CANNOT hit /telemetry/raw directly (403 Forbidden)
	reqRawForbidden, _ := http.NewRequest("GET", "/api/devices/10/telemetry/raw?page=1&page_size=20", nil)
	reqRawForbidden.Header.Set("Authorization", "Bearer "+env.OperatorToken)
	wRawForbidden := httptest.NewRecorder()
	env.Router.ServeHTTP(wRawForbidden, reqRawForbidden)
	if wRawForbidden.Code != http.StatusForbidden {
		t.Fatalf("Expected 403 Forbidden for Operator on /telemetry/raw, got %d", wRawForbidden.Code)
	}

	// 2. Operator CAN query /telemetry/history?raw_window=1 within max range (24h span <= 48h max, 10 days ago)
	//    and receives Value while RawHex and RawValue are sanitized
	startValid := histTime.Add(-1 * time.Hour).Format(time.RFC3339)
	endValid := histTime.Add(23 * time.Hour).Format(time.RFC3339)
	urlValid := fmt.Sprintf("/api/devices/10/telemetry/history?raw_window=1&start_time=%s&end_time=%s", startValid, endValid)
	reqOpValid, _ := http.NewRequest("GET", urlValid, nil)
	reqOpValid.Header.Set("Authorization", "Bearer "+env.OperatorToken)
	wOpValid := httptest.NewRecorder()
	env.Router.ServeHTTP(wOpValid, reqOpValid)
	if wOpValid.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK for Operator valid 24h window, got %d: %s", wOpValid.Code, wOpValid.Body.String())
	}

	var opResp struct {
		Data struct {
			Total int             `json:"total"`
			Items []model.RawData `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(wOpValid.Body.Bytes(), &opResp); err != nil {
		t.Fatalf("Failed to parse operator response: %v", err)
	}
	if len(opResp.Data.Items) != 1 {
		t.Fatalf("Expected 1 record in historical 24h window, got %d", len(opResp.Data.Items))
	}
	if opResp.Data.Items[0].Value != 6.92 {
		t.Errorf("Expected scaled Value 6.92 for Operator, got %v", opResp.Data.Items[0].Value)
	}
	if opResp.Data.Items[0].RawHex != "" || opResp.Data.Items[0].RawValue != 0 {
		t.Errorf("Expected RawHex and RawValue to be sanitized for Operator, got hex=%q raw=%v", opResp.Data.Items[0].RawHex, opResp.Data.Items[0].RawValue)
	}

	// 3. Span exceeding configured max range (72h > 48h max) is rejected with 400 Bad Request for both Operator and Admin
	endTooWide := histTime.Add(72 * time.Hour).Format(time.RFC3339)
	urlTooWideOp := fmt.Sprintf("/api/devices/10/telemetry/history?raw_window=1&start_time=%s&end_time=%s", startValid, endTooWide)
	reqOpTooWide, _ := http.NewRequest("GET", urlTooWideOp, nil)
	reqOpTooWide.Header.Set("Authorization", "Bearer "+env.OperatorToken)
	wOpTooWide := httptest.NewRecorder()
	env.Router.ServeHTTP(wOpTooWide, reqOpTooWide)
	if wOpTooWide.Code != http.StatusBadRequest {
		t.Fatalf("Expected 400 BadRequest when Operator exceeds max range (72h > 48h), got %d", wOpTooWide.Code)
	}

	urlTooWideAdmin := fmt.Sprintf("/api/devices/10/telemetry/raw?start_time=%s&end_time=%s", startValid, endTooWide)
	reqAdminTooWide, _ := http.NewRequest("GET", urlTooWideAdmin, nil)
	reqAdminTooWide.Header.Set("Authorization", "Bearer "+env.AdminToken)
	wAdminTooWide := httptest.NewRecorder()
	env.Router.ServeHTTP(wAdminTooWide, reqAdminTooWide)
	if wAdminTooWide.Code != http.StatusBadRequest {
		t.Fatalf("Expected 400 BadRequest when Admin exceeds max range (72h > 48h), got %d", wAdminTooWide.Code)
	}
}

func TestRawAndHistoricalTelemetryMultiPagePaginationPreservesTotal(t *testing.T) {
	env := setupTelemetryTestRouter(t)
	defer env.TelemetryService.Stop()

	dev := model.Device{ID: 45, DeviceCode: "com1", DeviceName: "COM 1"}
	_ = env.DB.Create(&dev).Error
	param := model.Parameter{
		ID:            4501,
		DeviceID:      45,
		ParameterCode: "debit",
		ParameterName: "Debit",
		Unit:          "m3/s",
		DataType:      model.DataTypeFloat32,
		Enabled:       true,
	}
	_ = env.DB.Create(&param).Error

	baseNow := time.Now().UTC().Add(-30 * time.Minute).Truncate(time.Second)
	for i := 0; i < 45; i++ {
		ts := baseNow.Add(time.Duration(i*10) * time.Second)
		_ = env.DB.Create(&model.RawData{
			DeviceID:       45,
			ParameterID:    4501,
			Value:          0.85,
			ProcessedValue: 0.85,
			RawValue:       0.85,
			RawHex:         "3F 59 99 9A",
			Quality:        model.QualityGood,
			Timestamp:      ts,
			ReceivedAt:     ts,
			StoredAt:       ts,
		}).Error
	}

	startISO := baseNow.Add(-1 * time.Minute).Format(time.RFC3339)
	endISO := baseNow.Add(15 * time.Minute).Format(time.RFC3339)

	for _, endpoint := range []string{"/api/devices/45/telemetry/raw", "/api/devices/45/telemetry/history"} {
		expectedCounts := map[int]int{1: 20, 2: 20, 3: 5}
		for page := 1; page <= 3; page++ {
			u := fmt.Sprintf("%s?page=%d&page_size=20&raw_window=1&start_time=%s&end_time=%s", endpoint, page, startISO, endISO)
			req, _ := http.NewRequest("GET", u, nil)
			req.Header.Set("Authorization", "Bearer "+env.AdminToken)
			w := httptest.NewRecorder()
			env.Router.ServeHTTP(w, req)
			if w.Code != http.StatusOK {
				t.Fatalf("Expected 200 OK for %s page %d, got %d: %s", endpoint, page, w.Code, w.Body.String())
			}

			var resp struct {
				Data struct {
					Total int             `json:"total"`
					Page  int             `json:"page"`
					Items []model.RawData `json:"items"`
				} `json:"data"`
			}
			if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
				t.Fatalf("Failed to decode %s page %d: %v", endpoint, page, err)
			}
			if resp.Data.Total != 45 {
				t.Fatalf("%s page %d: expected total=45 (must not collapse to 0 on page > 1), got %d", endpoint, page, resp.Data.Total)
			}
			if len(resp.Data.Items) != expectedCounts[page] {
				t.Fatalf("%s page %d: expected %d items, got %d", endpoint, page, expectedCounts[page], len(resp.Data.Items))
			}
		}
	}

	// Bulk export loop optimization check: explicit skip_count=1 skips COUNT(*) on page > 1
	uSkip := fmt.Sprintf("/api/devices/45/telemetry/history?page=2&page_size=20&skip_count=1&start_time=%s&end_time=%s", startISO, endISO)
	reqSkip, _ := http.NewRequest("GET", uSkip, nil)
	reqSkip.Header.Set("Authorization", "Bearer "+env.AdminToken)
	wSkip := httptest.NewRecorder()
	env.Router.ServeHTTP(wSkip, reqSkip)
	var skipResp struct {
		Data struct {
			Total int             `json:"total"`
			Items []model.RawData `json:"items"`
		} `json:"data"`
	}
	_ = json.Unmarshal(wSkip.Body.Bytes(), &skipResp)
	if skipResp.Data.Total != 0 || len(skipResp.Data.Items) != 20 {
		t.Fatalf("expected skip_count=1 on page 2 to return total=0 and 20 items, got total=%d items=%d", skipResp.Data.Total, len(skipResp.Data.Items))
	}
}
