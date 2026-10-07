package handler_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"datalogger/internal/communication"
	"datalogger/internal/communication/modbus"
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

type CommTestEnv struct {
	Router        *gin.Engine
	DB            *gorm.DB
	DeviceService *service.DeviceService
	ConnManager   *communication.ConnectionManager
	PollingEngine *communication.PollingEngine
	Server        *modbus.MockModbusServer
	Host          string
	Port          int
	AdminToken    string
	OpToken       string
}

func setupCommTestEnv(t *testing.T) *CommTestEnv {
	gin.SetMode(gin.TestMode)

	// 1. Mock Server
	server := modbus.NewMockModbusServer()
	server.SetHoldingRegister(0, 0x1234)
	server.SetHoldingRegister(1, 0x5678)
	addr, err := server.StartTCP()
	if err != nil {
		t.Fatalf("failed starting TCP mock server: %v", err)
	}

	host, portStr, _ := net.SplitHostPort(addr)
	port, _ := strconv.Atoi(portStr)

	// 2. In-memory DB
	dsn := fmt.Sprintf("file:mem_comm_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		t.Fatalf("Failed opening test db: %v", err)
	}

	_ = database.RunMigrations(db)
	database.Seed(db)

	// Fetch seeded roles & permissions
	hash, _ := bcrypt.GenerateFromPassword([]byte("pass"), bcrypt.DefaultCost)
	var adminRole, opRole model.Role
	db.Where("name = ?", "Administrator").First(&adminRole)
	db.Where("name = ?", "Operator").First(&opRole)

	var allPerms []model.Permission
	db.Find(&allPerms)
	_ = db.Model(&adminRole).Association("Permissions").Replace(allPerms)

	var opPerms []model.Permission
	db.Where("code IN ?", []string{"device.view", "device.communication.view"}).Find(&opPerms)
	_ = db.Model(&opRole).Association("Permissions").Replace(opPerms)

	adminUser := model.User{Username: "comm_admin", Email: "admin@comm.local", Password: string(hash), RoleID: adminRole.ID, IsActive: true}
	opUser := model.User{Username: "comm_op", Email: "op@comm.local", Password: string(hash), RoleID: opRole.ID, IsActive: true}
	db.Create(&adminUser)
	db.Create(&opUser)

	cfg := config.DefaultConfig()
	systemRepo := repository.NewSystemRepository(db)
	deviceRepo := repository.NewDeviceRepository(db)

	authService := service.NewAuthService(systemRepo, cfg)
	adminToken, _, _ := authService.Login("comm_admin", "pass", "127.0.0.1", "")
	opToken, _, _ := authService.Login("comm_op", "pass", "127.0.0.1", "")

	deviceService := service.NewDeviceService(deviceRepo, systemRepo)

	connManager := communication.NewConnectionManager(deviceService, communication.DefaultAdapterFactory)
	pollingEngine := communication.NewPollingEngine(connManager, deviceService, nil)

	commHandler := handler.NewCommunicationHandler(connManager, pollingEngine, deviceService)

	r := gin.New()
	api := r.Group("/api")
	devicesGroup := api.Group("/devices")
	devicesGroup.Use(middleware.JWTAuth(authService))
	{
		devicesGroup.POST("/:id/communication/connect", middleware.RequirePermission("device.communication.manage"), commHandler.Connect)
		devicesGroup.POST("/:id/communication/disconnect", middleware.RequirePermission("device.communication.manage"), commHandler.Disconnect)
		devicesGroup.POST("/:id/communication/reconnect", middleware.RequirePermission("device.communication.manage"), commHandler.Reconnect)
		devicesGroup.POST("/:id/communication/test", middleware.RequirePermission("device.communication.test"), commHandler.TestConnection)
		devicesGroup.GET("/:id/communication/status", middleware.RequirePermission("device.communication.view"), commHandler.GetStatus)
		devicesGroup.POST("/:id/parameters/:paramId/test-read", middleware.RequirePermission("device.communication.test"), commHandler.TestReadParameter)
	}

	return &CommTestEnv{
		Router:        r,
		DB:            db,
		DeviceService: deviceService,
		ConnManager:   connManager,
		PollingEngine: pollingEngine,
		Server:        server,
		Host:          host,
		Port:          port,
		AdminToken:    adminToken,
		OpToken:       opToken,
	}
}

func TestCommunicationHandlerEndpoints(t *testing.T) {
	env := setupCommTestEnv(t)
	defer env.Server.Stop()

	// Seed a test device
	dev, err := env.DeviceService.CreateDevice(&service.CreateDeviceRequest{
		DeviceCode: "DEV-COMM-01",
		DeviceName: "Communication Test Meter",
		DeviceType: "MODBUS_TCP",
		Connection: &service.ConnectionConfigDTO{
			Protocol:        model.ProtocolModbusTCP,
			Host:            env.Host,
			Port:            env.Port,
			Timeout:         1000,
			RetryCount:      1,
			PollingInterval: 1000,
			SlaveID:         1,
			ByteOrder:       "ABCD",
			Enabled:         true,
		},
	}, "test_admin", "127.0.0.1", "test-agent")
	if err != nil {
		t.Fatalf("Failed creating device: %v", err)
	}

	// Add parameter
	param, err := env.DeviceService.CreateParameter(dev.ID, &service.CreateParameterRequest{
		ParameterCode:   "VOLTAGE",
		ParameterName:   "Phase Voltage",
		RegisterType:    "HOLDING_REGISTER",
		RegisterAddress: 40001,
		DataType:        model.DataTypeUInt32,
		Scale:           1.0,
		Offset:          0.0,
		ByteOrder:       "ABCD",
	}, "test_admin", "127.0.0.1", "test-agent")
	if err != nil {
		t.Fatalf("Failed creating parameter: %v", err)
	}

	// 1. RBAC Check: Operator cannot connect (Forbidden)
	reqConnOp, _ := http.NewRequestWithContext(context.Background(), "POST", fmt.Sprintf("/api/devices/%d/communication/connect", dev.ID), nil)
	reqConnOp.Header.Set("Authorization", "Bearer "+env.OpToken)
	wConnOp := httptest.NewRecorder()
	env.Router.ServeHTTP(wConnOp, reqConnOp)
	if wConnOp.Code != http.StatusForbidden {
		t.Fatalf("expected 403 Forbidden for operator on connect, got %d", wConnOp.Code)
	}

	// 2. Connect Device as Admin (200 OK)
	reqConn, _ := http.NewRequestWithContext(context.Background(), "POST", fmt.Sprintf("/api/devices/%d/communication/connect", dev.ID), nil)
	reqConn.Header.Set("Authorization", "Bearer "+env.AdminToken)
	wConn := httptest.NewRecorder()
	env.Router.ServeHTTP(wConn, reqConn)
	if wConn.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for connect, got %d: %s", wConn.Code, wConn.Body.String())
	}

	// 3. Test Connection
	reqTest, _ := http.NewRequestWithContext(context.Background(), "POST", fmt.Sprintf("/api/devices/%d/communication/test", dev.ID), nil)
	reqTest.Header.Set("Authorization", "Bearer "+env.AdminToken)
	wTest := httptest.NewRecorder()
	env.Router.ServeHTTP(wTest, reqTest)
	if wTest.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for test connection, got %d", wTest.Code)
	}

	// 4. Get Communication Status (Accessible by Operator)
	reqStatus, _ := http.NewRequestWithContext(context.Background(), "GET", fmt.Sprintf("/api/devices/%d/communication/status", dev.ID), nil)
	reqStatus.Header.Set("Authorization", "Bearer "+env.OpToken)
	wStatus := httptest.NewRecorder()
	env.Router.ServeHTTP(wStatus, reqStatus)
	if wStatus.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for status, got %d", wStatus.Code)
	}

	// 5. Test Read Parameter
	reqRead, _ := http.NewRequestWithContext(context.Background(), "POST", fmt.Sprintf("/api/devices/%d/parameters/%d/test-read", dev.ID, param.ID), nil)
	reqRead.Header.Set("Authorization", "Bearer "+env.AdminToken)
	wRead := httptest.NewRecorder()
	env.Router.ServeHTTP(wRead, reqRead)
	if wRead.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for test-read, got %d: %s", wRead.Code, wRead.Body.String())
	}

	var readRes map[string]interface{}
	_ = json.Unmarshal(wRead.Body.Bytes(), &readRes)
	dataMap, ok := readRes["data"].(map[string]interface{})
	if !ok || dataMap["success"] != true {
		t.Fatalf("expected success in test-read response: %v", readRes)
	}
	if dataMap["raw_value"] != float64(0x12345678) {
		t.Fatalf("expected raw_value 0x12345678 (%d), got %v", 0x12345678, dataMap["raw_value"])
	}

	// 6. Disconnect Device
	reqDis, _ := http.NewRequestWithContext(context.Background(), "POST", fmt.Sprintf("/api/devices/%d/communication/disconnect", dev.ID), nil)
	reqDis.Header.Set("Authorization", "Bearer "+env.AdminToken)
	wDis := httptest.NewRecorder()
	env.Router.ServeHTTP(wDis, reqDis)
	if wDis.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for disconnect, got %d", wDis.Code)
	}
}
