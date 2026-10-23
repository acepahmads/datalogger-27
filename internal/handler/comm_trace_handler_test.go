package handler_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"datalogger/internal/communication"
	"datalogger/internal/communication/modbus"
	"datalogger/internal/config"
	"datalogger/internal/database"
	"datalogger/internal/handler"
	"datalogger/internal/middleware"
	"datalogger/internal/model"
	"datalogger/internal/repository"
	"datalogger/internal/service"
)

func TestCommTraceEndpoints_RBACAndLifecycle(t *testing.T) {
	gin.SetMode(gin.TestMode)

	dsn := fmt.Sprintf("file:mem_trace_api_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	_ = database.RunMigrations(db)

	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	var saRole, opRole model.Role
	db.Where("name = ?", "SUPERADMIN").First(&saRole)
	db.Where("name = ?", "OPERATOR").First(&opRole)

	adminUser := model.User{Username: "superadmin_user", Email: "sa@test.local", Password: string(hash), RoleID: saRole.ID, IsActive: true}
	opUser := model.User{Username: "op_user", Email: "op@test.local", Password: string(hash), RoleID: opRole.ID, IsActive: true}
	db.Create(&adminUser)
	db.Create(&opUser)

	cfg := config.DefaultConfig()
	systemRepo := repository.NewSystemRepository(db)
	deviceRepo := repository.NewDeviceRepository(db)
	authService := service.NewAuthService(systemRepo, cfg)
	deviceService := service.NewDeviceService(deviceRepo, systemRepo)

	tracer := communication.NewCommTracer(nil)
	traceHandler := handler.NewCommTraceHandler(tracer, deviceService)

	r := gin.New()
	r.Use(gin.Recovery())

	api := r.Group("/api")
	alarmsGroup := api.Group("/alarms")
	alarmsGroup.Use(middleware.JWTAuth(authService))
	{
		alarmsGroup.GET("/live-comm/status", middleware.RequirePermission("alarm.view"), traceHandler.GetStatus)
		alarmsGroup.GET("/live-comm/frames", middleware.RequirePermission("alarm.view"), traceHandler.GetFrames)
		alarmsGroup.POST("/live-comm/toggle", middleware.RequireRole("SUPERADMIN"), traceHandler.ToggleTrace)
		alarmsGroup.DELETE("/live-comm/frames", middleware.RequireRole("SUPERADMIN"), traceHandler.ClearFrames)
	}

	saToken, _, _ := authService.Login("superadmin_user", "password123", "127.0.0.1", "")
	opToken, _, _ := authService.Login("op_user", "password123", "127.0.0.1", "")

	// 1. Initial status query as Operator
	statusReq, _ := http.NewRequest("GET", "/api/alarms/live-comm/status", nil)
	statusReq.Header.Set("Authorization", "Bearer "+opToken)
	statusW := httptest.NewRecorder()
	r.ServeHTTP(statusW, statusReq)

	if statusW.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK for status, got %d: %s", statusW.Code, statusW.Body.String())
	}
	var statusRes struct {
		Data struct {
			Active    bool `json:"active"`
			CanToggle bool `json:"can_toggle"`
		} `json:"data"`
	}
	_ = json.Unmarshal(statusW.Body.Bytes(), &statusRes)
	if statusRes.Data.Active {
		t.Errorf("Expected active=false initially")
	}
	if statusRes.Data.CanToggle {
		t.Errorf("Operator should NOT have can_toggle=true")
	}

	// 2. Operator attempts to toggle live trace (Must return 403 Forbidden)
	togglePayload := map[string]interface{}{
		"active":           true,
		"duration_minutes": 10,
	}
	tBody, _ := json.Marshal(togglePayload)
	opToggleReq, _ := http.NewRequest("POST", "/api/alarms/live-comm/toggle", bytes.NewBuffer(tBody))
	opToggleReq.Header.Set("Authorization", "Bearer "+opToken)
	opToggleReq.Header.Set("Content-Type", "application/json")
	opToggleW := httptest.NewRecorder()
	r.ServeHTTP(opToggleW, opToggleReq)

	if opToggleW.Code != http.StatusForbidden {
		t.Fatalf("Expected 403 Forbidden when Operator tries to toggle live-comm, got %d", opToggleW.Code)
	}

	// 3. Superadmin activates live trace (Must return 200 OK)
	saToggleReq, _ := http.NewRequest("POST", "/api/alarms/live-comm/toggle", bytes.NewBuffer(tBody))
	saToggleReq.Header.Set("Authorization", "Bearer "+saToken)
	saToggleReq.Header.Set("Content-Type", "application/json")
	saToggleW := httptest.NewRecorder()
	r.ServeHTTP(saToggleW, saToggleReq)

	if saToggleW.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK when Superadmin toggles live-comm, got %d: %s", saToggleW.Code, saToggleW.Body.String())
	}

	if !tracer.IsActive() {
		t.Fatalf("Tracer should be active after Superadmin activation")
	}

	// 4. Record a mock transaction frame
	dev := &model.Device{
		ID:         5,
		DeviceCode: "AQMS-02",
		DeviceName: "Station Ambient Air",
		Connection: &model.DeviceConnection{
			Protocol:   model.ProtocolModbusRTU,
			SerialPort: "/dev/ttyUSB0",
		},
	}
	req := modbus.ModbusReadRequest{SlaveID: 1, FunctionCode: 3, StartingAddress: 20, Quantity: 4}
	resp := &modbus.ModbusReadResponse{
		SlaveID:      1,
		FunctionCode: 3,
		ByteCount:    8,
		RequestADU:   []byte{0x01, 0x03, 0x00, 0x14, 0x00, 0x04, 0x04, 0x0B},
		ResponseADU:  []byte{0x01, 0x03, 0x08, 0x41, 0xC8, 0x00, 0x00, 0x41, 0x90, 0x00, 0x00, 0x5E, 0xD2},
	}
	tracer.Record(dev, req, resp, nil, 80)

	// 5. Fetch frames via GET /api/alarms/live-comm/frames
	framesReq, _ := http.NewRequest("GET", "/api/alarms/live-comm/frames?limit=50", nil)
	framesReq.Header.Set("Authorization", "Bearer "+opToken)
	framesW := httptest.NewRecorder()
	r.ServeHTTP(framesW, framesReq)

	if framesW.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK on get frames, got %d", framesW.Code)
	}

	var framesRes struct {
		Data []communication.CommTraceFrame `json:"data"`
	}
	_ = json.Unmarshal(framesW.Body.Bytes(), &framesRes)
	if len(framesRes.Data) != 1 {
		t.Fatalf("Expected 1 frame in response, got %d", len(framesRes.Data))
	}
	if framesRes.Data[0].TXHex != "01 03 00 14 00 04 04 0B" {
		t.Errorf("Unexpected TX hex: %s", framesRes.Data[0].TXHex)
	}

	// 6. Superadmin stops live trace
	stopPayload := map[string]interface{}{"active": false}
	sBody, _ := json.Marshal(stopPayload)
	stopReq, _ := http.NewRequest("POST", "/api/alarms/live-comm/toggle", bytes.NewBuffer(sBody))
	stopReq.Header.Set("Authorization", "Bearer "+saToken)
	stopReq.Header.Set("Content-Type", "application/json")
	stopW := httptest.NewRecorder()
	r.ServeHTTP(stopW, stopReq)

	if stopW.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK on stop, got %d", stopW.Code)
	}
	if tracer.IsActive() {
		t.Errorf("Tracer should be inactive after Stop")
	}
}
