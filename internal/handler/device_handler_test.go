package handler_test

import (
	"bytes"
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

type TestEnv struct {
	Router     *gin.Engine
	DB         *gorm.DB
	AdminToken string
	EngToken   string
	OpToken    string
}

func setupDeviceTestRouter(t *testing.T) *TestEnv {
	gin.SetMode(gin.TestMode)

	dsn := fmt.Sprintf("file:mem_dev_api_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	_ = database.RunMigrations(db)

	// Seed roles & test users
	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	adminRole := model.Role{Name: "Administrator", Description: "Full access"}
	engRole := model.Role{Name: "Engineer", Description: "Engineering access"}
	opRole := model.Role{Name: "Operator", Description: "Read access"}
	db.Create(&adminRole)
	db.Create(&engRole)
	db.Create(&opRole)

	adminUser := model.User{Username: "test_admin", Email: "admin@test.local", Password: string(hash), RoleID: adminRole.ID, IsActive: true}
	engUser := model.User{Username: "test_eng", Email: "eng@test.local", Password: string(hash), RoleID: engRole.ID, IsActive: true}
	opUser := model.User{Username: "test_op", Email: "op@test.local", Password: string(hash), RoleID: opRole.ID, IsActive: true}
	db.Create(&adminUser)
	db.Create(&engUser)
	db.Create(&opUser)

	cfg := config.DefaultConfig()
	systemRepo := repository.NewSystemRepository(db)
	deviceRepo := repository.NewDeviceRepository(db)

	authService := service.NewAuthService(systemRepo, cfg)
	deviceService := service.NewDeviceService(deviceRepo, systemRepo)

	deviceHandler := handler.NewDeviceHandler(deviceService)
	systemHandler := handler.NewSystemHandler(service.NewSystemService(systemRepo, repository.NewPhaseRepository(db), cfg))

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.CORS())

	api := r.Group("/api")
	{
		devicesGroup := api.Group("/devices")
		devicesGroup.Use(middleware.JWTAuth(authService))
		{
			devicesGroup.GET("", middleware.RequirePermission("device.view"), deviceHandler.ListDevices)
			devicesGroup.GET("/:id", middleware.RequirePermission("device.view"), deviceHandler.GetDeviceByID)
			devicesGroup.GET("/:id/status", middleware.RequirePermission("device.view"), systemHandler.GetDeviceStatus)
			devicesGroup.GET("/:id/activity", middleware.RequirePermission("device.view"), deviceHandler.GetDeviceActivity)

			devicesGroup.POST("", middleware.RequirePermission("device.create"), deviceHandler.CreateDevice)

			devicesGroup.PUT("/:id", middleware.RequirePermission("device.update"), deviceHandler.UpdateDevice)
			devicesGroup.PUT("/:id/enable", middleware.RequirePermission("device.update"), deviceHandler.ToggleDeviceEnabled)
			devicesGroup.PUT("/:id/status", middleware.RequirePermission("device.update"), deviceHandler.UpdateDeviceStatus)
			devicesGroup.PUT("/:id/connection", middleware.RequirePermission("device.update"), deviceHandler.UpdateConnection)

			devicesGroup.DELETE("/:id", middleware.RequirePermission("device.delete"), deviceHandler.DeleteDevice)

			devicesGroup.GET("/:id/parameters", middleware.RequirePermission("device.view"), deviceHandler.ListParameters)
			devicesGroup.GET("/:id/parameters/:paramId", middleware.RequirePermission("device.view"), deviceHandler.GetParameter)
			devicesGroup.POST("/:id/parameters", middleware.RequirePermission("device.manage"), deviceHandler.CreateParameter)
			devicesGroup.PUT("/:id/parameters/:paramId", middleware.RequirePermission("device.manage"), deviceHandler.UpdateParameter)
			devicesGroup.DELETE("/:id/parameters/:paramId", middleware.RequirePermission("device.manage"), deviceHandler.DeleteParameter)
			devicesGroup.PUT("/:id/parameters/:paramId/enable", middleware.RequirePermission("device.manage"), deviceHandler.ToggleParameterEnabled)
			devicesGroup.POST("/parameters/validate-formula", middleware.RequirePermission("device.view"), deviceHandler.ValidateFormula)
			devicesGroup.GET("/:id/parameters/:paramId/quality", middleware.RequirePermission("device.view"), deviceHandler.GetParameterQuality)
			devicesGroup.PUT("/:id/parameters/:paramId/quality", middleware.RequirePermission("device.manage"), deviceHandler.UpdateParameterQuality)
		}
	}

	adminToken, _, _ := authService.Login("test_admin", "password123", "127.0.0.1", "")
	engToken, _, _ := authService.Login("test_eng", "password123", "127.0.0.1", "")
	opToken, _, _ := authService.Login("test_op", "password123", "127.0.0.1", "")

	return &TestEnv{
		Router:     r,
		DB:         db,
		AdminToken: adminToken,
		EngToken:   engToken,
		OpToken:    opToken,
	}
}

func TestDeviceSecurityAuthAndPermissions(t *testing.T) {
	env := setupDeviceTestRouter(t)

	// 1. Unauthenticated request -> 401
	reqUnauth, _ := http.NewRequest("GET", "/api/devices", nil)
	wUnauth := httptest.NewRecorder()
	env.Router.ServeHTTP(wUnauth, reqUnauth)
	if wUnauth.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401 Unauthorized for request without token, got %d", wUnauth.Code)
	}

	// 2. Operator token can GET /api/devices (device.view) -> 200
	reqOpGet, _ := http.NewRequest("GET", "/api/devices", nil)
	reqOpGet.Header.Set("Authorization", "Bearer "+env.OpToken)
	wOpGet := httptest.NewRecorder()
	env.Router.ServeHTTP(wOpGet, reqOpGet)
	if wOpGet.Code != http.StatusOK {
		t.Errorf("Expected 200 OK for operator viewing devices, got %d", wOpGet.Code)
	}

	// 3. Operator token trying to POST /api/devices -> 403 Forbidden
	payload, _ := json.Marshal(map[string]interface{}{
		"device_code": "SEC-DEV-01",
		"device_name": "Security Test Device",
		"device_type": "MODBUS_TCP",
	})
	reqOpPost, _ := http.NewRequest("POST", "/api/devices", bytes.NewBuffer(payload))
	reqOpPost.Header.Set("Authorization", "Bearer "+env.OpToken)
	reqOpPost.Header.Set("Content-Type", "application/json")
	wOpPost := httptest.NewRecorder()
	env.Router.ServeHTTP(wOpPost, reqOpPost)
	if wOpPost.Code != http.StatusForbidden {
		t.Errorf("Expected 403 Forbidden for operator creating device, got %d", wOpPost.Code)
	}

	// 4. Administrator token trying to POST /api/devices -> 201 Created
	reqAdminPost, _ := http.NewRequest("POST", "/api/devices", bytes.NewBuffer(payload))
	reqAdminPost.Header.Set("Authorization", "Bearer "+env.AdminToken)
	reqAdminPost.Header.Set("Content-Type", "application/json")
	wAdminPost := httptest.NewRecorder()
	env.Router.ServeHTTP(wAdminPost, reqAdminPost)
	if wAdminPost.Code != http.StatusCreated {
		t.Fatalf("Expected 201 Created for admin creating device, got %d (body: %s)", wAdminPost.Code, wAdminPost.Body.String())
	}
}

func TestDeviceCRUDAPI(t *testing.T) {
	env := setupDeviceTestRouter(t)

	// 1. Create Device
	devicePayload := map[string]interface{}{
		"device_code":      "CRUD-PM-01",
		"device_name":      "Industrial Power Meter",
		"device_type":      "MODBUS_TCP",
		"manufacturer":     "Schneider Electric",
		"model":            "PM8000",
		"serial_number":    "SN-88219",
		"location":         "Substation 3",
		"status":           "ACTIVE",
		"enabled":          true,
		"connection": map[string]interface{}{
			"protocol":         "MODBUS_TCP",
			"connection_type":  "ETHERNET",
			"host":             "10.0.0.55",
			"port":             502,
			"timeout":          1500,
			"retry_count":      3,
			"polling_interval": 1000,
		},
	}
	body, _ := json.Marshal(devicePayload)
	req, _ := http.NewRequest("POST", "/api/devices", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+env.AdminToken)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	env.Router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("Expected 201, got %d (body: %s)", w.Code, w.Body.String())
	}

	var res struct {
		Success bool `json:"success"`
		Data    struct {
			ID         uint   `json:"id"`
			DeviceCode string `json:"device_code"`
			DeviceName string `json:"device_name"`
			Status     string `json:"status"`
			Connection *struct {
				Protocol string `json:"protocol"`
				Host     string `json:"host"`
				Port     int    `json:"port"`
			} `json:"connection"`
		} `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &res)
	deviceID := res.Data.ID
	if deviceID == 0 {
		t.Fatalf("Expected positive device ID")
	}
	if res.Data.Connection == nil || res.Data.Connection.Host != "10.0.0.55" {
		t.Errorf("Connection config not saved properly: %+v", res.Data.Connection)
	}

	// 2. Reject duplicate device code
	dupReq, _ := http.NewRequest("POST", "/api/devices", bytes.NewBuffer(body))
	dupReq.Header.Set("Authorization", "Bearer "+env.AdminToken)
	dupReq.Header.Set("Content-Type", "application/json")
	dupW := httptest.NewRecorder()
	env.Router.ServeHTTP(dupW, dupReq)
	if dupW.Code != http.StatusBadRequest {
		t.Errorf("Expected 400 Bad Request on duplicate device_code, got %d", dupW.Code)
	}

	// 3. Get Device By ID
	getReq, _ := http.NewRequest("GET", fmt.Sprintf("/api/devices/%d", deviceID), nil)
	getReq.Header.Set("Authorization", "Bearer "+env.AdminToken)
	getW := httptest.NewRecorder()
	env.Router.ServeHTTP(getW, getReq)
	if getW.Code != http.StatusOK {
		t.Fatalf("Expected 200, got %d", getW.Code)
	}

	// 4. Update Device
	updatePayload, _ := json.Marshal(map[string]interface{}{
		"device_name": "Updated Substation Meter",
		"status":      "MAINTENANCE",
	})
	putReq, _ := http.NewRequest("PUT", fmt.Sprintf("/api/devices/%d", deviceID), bytes.NewBuffer(updatePayload))
	putReq.Header.Set("Authorization", "Bearer "+env.AdminToken)
	putReq.Header.Set("Content-Type", "application/json")
	putW := httptest.NewRecorder()
	env.Router.ServeHTTP(putW, putReq)
	if putW.Code != http.StatusOK {
		t.Fatalf("Expected 200 on update, got %d", putW.Code)
	}

	// 5. Toggle Device Enabled
	enablePayload, _ := json.Marshal(map[string]bool{"enabled": false})
	enReq, _ := http.NewRequest("PUT", fmt.Sprintf("/api/devices/%d/enable", deviceID), bytes.NewBuffer(enablePayload))
	enReq.Header.Set("Authorization", "Bearer "+env.AdminToken)
	enReq.Header.Set("Content-Type", "application/json")
	enW := httptest.NewRecorder()
	env.Router.ServeHTTP(enW, enReq)
	if enW.Code != http.StatusOK {
		t.Fatalf("Expected 200 on toggle enable, got %d", enW.Code)
	}

	// 6. Delete Device (Soft Delete)
	delReq, _ := http.NewRequest("DELETE", fmt.Sprintf("/api/devices/%d", deviceID), nil)
	delReq.Header.Set("Authorization", "Bearer "+env.AdminToken)
	delW := httptest.NewRecorder()
	env.Router.ServeHTTP(delW, delReq)
	if delW.Code != http.StatusOK {
		t.Fatalf("Expected 200 on delete, got %d", delW.Code)
	}

	// Verification that device is now soft-deleted
	checkReq, _ := http.NewRequest("GET", fmt.Sprintf("/api/devices/%d", deviceID), nil)
	checkReq.Header.Set("Authorization", "Bearer "+env.AdminToken)
	checkW := httptest.NewRecorder()
	env.Router.ServeHTTP(checkW, checkReq)
	if checkW.Code != http.StatusNotFound {
		t.Errorf("Expected 404 for soft deleted device, got %d", checkW.Code)
	}
}

func TestDeviceParameterAPI(t *testing.T) {
	env := setupDeviceTestRouter(t)

	// Create parent device
	devBody, _ := json.Marshal(map[string]interface{}{
		"device_code": "DEV-PARENT",
		"device_name": "Parent Sensor Node",
		"device_type": "MODBUS_RTU",
	})
	req, _ := http.NewRequest("POST", "/api/devices", bytes.NewBuffer(devBody))
	req.Header.Set("Authorization", "Bearer "+env.AdminToken)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	env.Router.ServeHTTP(w, req)

	var devRes struct {
		Data struct {
			ID uint `json:"id"`
		} `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &devRes)
	devID := devRes.Data.ID

	// 1. Create Parameter
	paramBody, _ := json.Marshal(map[string]interface{}{
		"parameter_code":   "HUMIDITY",
		"parameter_name":   "Relative Humidity",
		"data_type":        "FLOAT32",
		"unit":             "%RH",
		"scale":            0.1,
		"offset":           0.0,
		"register_address": 30002,
		"register_type":    "INPUT_REGISTER",
	})
	pReq, _ := http.NewRequest("POST", fmt.Sprintf("/api/devices/%d/parameters", devID), bytes.NewBuffer(paramBody))
	pReq.Header.Set("Authorization", "Bearer "+env.AdminToken)
	pReq.Header.Set("Content-Type", "application/json")
	pW := httptest.NewRecorder()
	env.Router.ServeHTTP(pW, pReq)

	if pW.Code != http.StatusCreated {
		t.Fatalf("Expected 201 on parameter create, got %d (body: %s)", pW.Code, pW.Body.String())
	}

	var pRes struct {
		Data struct {
			ID            uint   `json:"id"`
			ParameterCode string `json:"parameter_code"`
		} `json:"data"`
	}
	_ = json.Unmarshal(pW.Body.Bytes(), &pRes)
	paramID := pRes.Data.ID

	// 2. Reject duplicate parameter_code for this device
	dupPReq, _ := http.NewRequest("POST", fmt.Sprintf("/api/devices/%d/parameters", devID), bytes.NewBuffer(paramBody))
	dupPReq.Header.Set("Authorization", "Bearer "+env.AdminToken)
	dupPReq.Header.Set("Content-Type", "application/json")
	dupPW := httptest.NewRecorder()
	env.Router.ServeHTTP(dupPW, dupPReq)
	if dupPW.Code != http.StatusBadRequest {
		t.Errorf("Expected 400 Bad Request on duplicate parameter_code, got %d", dupPW.Code)
	}

	// 3. List parameters
	listReq, _ := http.NewRequest("GET", fmt.Sprintf("/api/devices/%d/parameters", devID), nil)
	listReq.Header.Set("Authorization", "Bearer "+env.AdminToken)
	listW := httptest.NewRecorder()
	env.Router.ServeHTTP(listW, listReq)
	if listW.Code != http.StatusOK {
		t.Fatalf("Expected 200 on list parameters, got %d", listW.Code)
	}

	// 4. Update parameter
	updateBody, _ := json.Marshal(map[string]interface{}{
		"parameter_name": "Ambient Humidity Calibrated",
		"scale":          0.12,
	})
	putReq, _ := http.NewRequest("PUT", fmt.Sprintf("/api/devices/%d/parameters/%d", devID, paramID), bytes.NewBuffer(updateBody))
	putReq.Header.Set("Authorization", "Bearer "+env.AdminToken)
	putReq.Header.Set("Content-Type", "application/json")
	putW := httptest.NewRecorder()
	env.Router.ServeHTTP(putW, putReq)
	if putW.Code != http.StatusOK {
		t.Fatalf("Expected 200 on update parameter, got %d", putW.Code)
	}

	// 5. Delete parameter
	delReq, _ := http.NewRequest("DELETE", fmt.Sprintf("/api/devices/%d/parameters/%d", devID, paramID), nil)
	delReq.Header.Set("Authorization", "Bearer "+env.AdminToken)
	delW := httptest.NewRecorder()
	env.Router.ServeHTTP(delW, delReq)
	if delW.Code != http.StatusOK {
		t.Fatalf("Expected 200 on delete parameter, got %d", delW.Code)
	}
}

func TestValidateFormulaEndpoint(t *testing.T) {
	env := setupDeviceTestRouter(t)

	// 1. Valid formula evaluation
	body, _ := json.Marshal(map[string]interface{}{
		"formula":      "x * 1.8 + 32",
		"sample_value": 25.0,
	})
	req, _ := http.NewRequest("POST", "/api/devices/parameters/validate-formula", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+env.AdminToken)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	env.Router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200 on valid formula, got %d: %s", w.Code, w.Body.String())
	}

	var res struct {
		Success bool `json:"success"`
		Data    struct {
			Valid   bool    `json:"valid"`
			Formula string  `json:"formula"`
			Result  float64 `json:"result"`
		} `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &res)
	if !res.Data.Valid || res.Data.Result != 77.0 {
		t.Errorf("Expected result 77.0, got %v", res.Data.Result)
	}

	// 2. Invalid syntax formula
	badBody, _ := json.Marshal(map[string]interface{}{
		"formula": "x +* 2",
	})
	badReq, _ := http.NewRequest("POST", "/api/devices/parameters/validate-formula", bytes.NewBuffer(badBody))
	badReq.Header.Set("Authorization", "Bearer "+env.AdminToken)
	badReq.Header.Set("Content-Type", "application/json")
	badW := httptest.NewRecorder()
	env.Router.ServeHTTP(badW, badReq)

	if badW.Code != http.StatusBadRequest {
		t.Fatalf("Expected 400 on bad formula, got %d", badW.Code)
	}
}

func TestParameterQualityEndpoints(t *testing.T) {
	env := setupDeviceTestRouter(t)

	// Create test device & parameter
	dev := model.Device{
		DeviceCode: "DEV-QUAL-01",
		DeviceName: "Quality Test Device",
		DeviceType: "MODBUS_TCP",
		Status:     model.DeviceStatusActive,
		Enabled:    true,
	}
	env.DB.Create(&dev)

	param := model.Parameter{
		DeviceID:                 dev.ID,
		ParameterCode:            "TEMP_TEST",
		ParameterName:            "Test Temperature",
		DataType:                 model.DataTypeFloat32,
		Unit:                     "°C",
		QualityValidationEnabled: true,
		ProcessingEnabled:        true,
		StaleTimeoutSeconds:      120,
	}
	env.DB.Create(&param)

	// 1. GET quality config (Authorized)
	getReq, _ := http.NewRequest("GET", fmt.Sprintf("/api/devices/%d/parameters/%d/quality", dev.ID, param.ID), nil)
	getReq.Header.Set("Authorization", "Bearer "+env.AdminToken)
	getW := httptest.NewRecorder()
	env.Router.ServeHTTP(getW, getReq)

	if getW.Code != http.StatusOK {
		t.Fatalf("Expected 200 on get quality config, got %d: %s", getW.Code, getW.Body.String())
	}
	var getResp struct {
		Success bool                        `json:"success"`
		Data    service.ParameterQualityDTO `json:"data"`
	}
	if err := json.Unmarshal(getW.Body.Bytes(), &getResp); err != nil {
		t.Fatalf("Failed to parse quality config response: %v", err)
	}
	if getResp.Data.StaleTimeoutSeconds != 120 {
		t.Errorf("Expected stale timeout 120, got %d", getResp.Data.StaleTimeoutSeconds)
	}

	// 2. PUT quality config (Authorized Admin)
	minVal := 0.0
	maxVal := 100.0
	warnLow := 10.0
	warnHigh := 85.0
	staleSec := 60
	spikeEn := true
	spikeTh := 15.5
	spikeWin := 4

	putBody, _ := json.Marshal(service.UpdateParameterQualityRequest{
		MinValue:              &minVal,
		MaxValue:              &maxVal,
		WarningLow:            &warnLow,
		WarningHigh:           &warnHigh,
		StaleTimeoutSeconds:   &staleSec,
		SpikeDetectionEnabled: &spikeEn,
		SpikeThreshold:        &spikeTh,
		SpikeWindowSize:       &spikeWin,
	})
	putReq, _ := http.NewRequest("PUT", fmt.Sprintf("/api/devices/%d/parameters/%d/quality", dev.ID, param.ID), bytes.NewBuffer(putBody))
	putReq.Header.Set("Authorization", "Bearer "+env.AdminToken)
	putReq.Header.Set("Content-Type", "application/json")
	putW := httptest.NewRecorder()
	env.Router.ServeHTTP(putW, putReq)

	if putW.Code != http.StatusOK {
		t.Fatalf("Expected 200 on update quality config, got %d: %s", putW.Code, putW.Body.String())
	}

	// Verify persistence in DB
	var reloaded model.Parameter
	env.DB.First(&reloaded, param.ID)
	if reloaded.StaleTimeoutSeconds != 60 || !reloaded.SpikeDetectionEnabled || reloaded.SpikeThreshold != 15.5 {
		t.Errorf("Database did not persist quality config: %+v", reloaded)
	}

	// 3. Reject invalid limits (warning_low > warning_high)
	invalidLow := 95.0
	invalidHigh := 50.0
	badPutBody, _ := json.Marshal(service.UpdateParameterQualityRequest{
		WarningLow:  &invalidLow,
		WarningHigh: &invalidHigh,
	})
	badPutReq, _ := http.NewRequest("PUT", fmt.Sprintf("/api/devices/%d/parameters/%d/quality", dev.ID, param.ID), bytes.NewBuffer(badPutBody))
	badPutReq.Header.Set("Authorization", "Bearer "+env.AdminToken)
	badPutReq.Header.Set("Content-Type", "application/json")
	badPutW := httptest.NewRecorder()
	env.Router.ServeHTTP(badPutW, badPutReq)

	if badPutW.Code != http.StatusBadRequest {
		t.Errorf("Expected 400 Bad Request on invalid warning limits, got %d", badPutW.Code)
	}

	// 4. RBAC Permission check: Operator token cannot modify quality configuration (requires device.manage)
	opPutReq, _ := http.NewRequest("PUT", fmt.Sprintf("/api/devices/%d/parameters/%d/quality", dev.ID, param.ID), bytes.NewBuffer(putBody))
	opPutReq.Header.Set("Authorization", "Bearer "+env.OpToken)
	opPutReq.Header.Set("Content-Type", "application/json")
	opPutW := httptest.NewRecorder()
	env.Router.ServeHTTP(opPutW, opPutReq)

	if opPutW.Code != http.StatusForbidden {
		t.Errorf("Expected 403 Forbidden for operator modifying quality, got %d", opPutW.Code)
	}
}

