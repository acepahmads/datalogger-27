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
