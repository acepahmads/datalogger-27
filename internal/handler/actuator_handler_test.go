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

	"datalogger/internal/communication/modbus"
	"datalogger/internal/handler"
	"datalogger/internal/middleware"
	"datalogger/internal/model"
	"datalogger/internal/repository"
	"datalogger/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

type stubModbusWriter struct{}

func (s *stubModbusWriter) Connect(ctx context.Context) error { return nil }
func (s *stubModbusWriter) IsConnected() bool                 { return true }
func (s *stubModbusWriter) WriteRegisters(ctx context.Context, req modbus.ModbusWriteRequest) (*modbus.ModbusWriteResponse, error) {
	return &modbus.ModbusWriteResponse{
		SlaveID:      req.SlaveID,
		FunctionCode: req.FunctionCode,
		Address:      req.Address,
		Value:        req.Value,
		ResponseTime: 8 * time.Millisecond,
		RequestADU:   []byte{req.SlaveID, req.FunctionCode, 0x00, byte(req.Address), byte(req.Value >> 8), byte(req.Value), 0x11, 0x22},
		ResponseADU:  []byte{req.SlaveID, req.FunctionCode, 0x00, byte(req.Address), byte(req.Value >> 8), byte(req.Value), 0x11, 0x22},
		CRCPassed:    true,
	}, nil
}

func setupActuatorRouter(t *testing.T, role string) (*gin.Engine, *gorm.DB) {
	gin.SetMode(gin.TestMode)
	dsn := fmt.Sprintf("file:actuator_handler_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite db: %v", err)
	}

	if err := db.AutoMigrate(
		&model.DeviceType{},
		&model.Device{},
		&model.DeviceConnection{},
		&model.ParameterCategory{},
		&model.Parameter{},
		&model.DeviceActuator{},
		&model.ActuationSchedule{},
		&model.ActuationLog{},
	); err != nil {
		t.Fatalf("failed to migrate models: %v", err)
	}

	actuatorRepo := repository.NewActuatorRepository(db)
	deviceRepo := repository.NewDeviceRepository(db)
	actuationSvc := service.NewActuationService(actuatorRepo, deviceRepo, func(d *model.Device) (service.ModbusWriter, error) {
		return &stubModbusWriter{}, nil
	}, nil)
	actuatorHandler := handler.NewActuatorHandler(actuationSvc)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("userID", uint(1))
		c.Set("username", "admin")
		c.Set("role", role)
		c.Next()
	})

	api := r.Group("/api")
	{
		devicesGroup := api.Group("/devices")
		{
			devicesGroup.GET("/:id/actuators", middleware.RequirePermission("device.view"), actuatorHandler.ListByDevice)
			devicesGroup.POST("/:id/actuators", middleware.RequirePermission("device.manage"), actuatorHandler.Create)
		}
		actuatorsGroup := api.Group("/actuators")
		{
			actuatorsGroup.GET("", actuatorHandler.ListAllActive)
			actuatorsGroup.PUT("/:id", middleware.RequirePermission("device.manage"), actuatorHandler.Update)
			actuatorsGroup.DELETE("/:id", middleware.RequirePermission("device.manage"), actuatorHandler.Delete)
			actuatorsGroup.POST("/:id/actuate", actuatorHandler.Actuate)
			actuatorsGroup.POST("/:id/schedules", middleware.RequirePermission("device.manage"), actuatorHandler.CreateSchedule)
		}
	}

	return r, db
}

func TestActuatorHandler_CreateEditAndCompositeUniquePerDevice(t *testing.T) {
	r, db := setupActuatorRouter(t, "ADMIN")

	dev5 := &model.Device{ID: 5, DeviceCode: "com4", DeviceName: "COM4 Station", Enabled: true}
	dev6 := &model.Device{ID: 6, DeviceCode: "com5", DeviceName: "COM5 Station", Enabled: true}
	if err := db.Create(dev5).Error; err != nil {
		t.Fatalf("seed dev5: %v", err)
	}
	if err := db.Create(dev6).Error; err != nil {
		t.Fatalf("seed dev6: %v", err)
	}

	param := &model.Parameter{
		DeviceID:      5,
		ParameterCode: "pump on",
		ParameterName: "PUMP ON",
		Enabled:       true,
	}
	if err := db.Create(param).Error; err != nil {
		t.Fatalf("seed param: %v", err)
	}

	// 1. Exact payload matching Add Actuator modal screenshot on /api/devices/5/actuators
	payload := map[string]interface{}{
		"actuator_code":          "PUMP_ON",
		"actuator_name":          "PUMP ON",
		"description":            "",
		"icon":                   "motor",
		"control_type":           "COIL_FC05",
		"slave_id":               6,
		"register_address":       50,
		"on_payload":             65280,
		"off_payload":            0,
		"safety_timeout_seconds": 900,
		"safety_watchdog_active": true,
		"feedback_parameter_id":  param.ID,
		"enabled":                true,
	}
	bodyBytes, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/api/devices/5/actuators", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created, got %d: %s", w.Code, w.Body.String())
	}

	// 2. Same actuator_code on Device 6 must succeed (per-device composite unique index)
	req2 := httptest.NewRequest(http.MethodPost, "/api/devices/6/actuators", bytes.NewReader(bodyBytes))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)

	if w2.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created for same actuator_code on different device, got %d: %s", w2.Code, w2.Body.String())
	}

	// 3. Explicit false booleans on creation must persist as false (not overridden by GORM default:true)
	payloadFalse := map[string]interface{}{
		"actuator_code":          "VALVE_PURGE",
		"actuator_name":          "Purge Valve",
		"icon":                   "valve",
		"control_type":           "HOLDING_REGISTER_FC06",
		"slave_id":               6,
		"register_address":       51,
		"on_payload":             1,
		"off_payload":            0,
		"safety_timeout_seconds": 300,
		"safety_watchdog_active": false,
		"feedback_parameter_id":  nil,
		"enabled":                false,
	}
	bFalse, _ := json.Marshal(payloadFalse)
	req3 := httptest.NewRequest(http.MethodPost, "/api/devices/5/actuators", bytes.NewReader(bFalse))
	req3.Header.Set("Content-Type", "application/json")
	w3 := httptest.NewRecorder()
	r.ServeHTTP(w3, req3)
	if w3.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created for VALVE_PURGE, got %d: %s", w3.Code, w3.Body.String())
	}

	// Verify GET /api/devices/5/actuators returns both with accurate boolean values
	reqList := httptest.NewRequest(http.MethodGet, "/api/devices/5/actuators", nil)
	wList := httptest.NewRecorder()
	r.ServeHTTP(wList, reqList)
	if wList.Code != http.StatusOK {
		t.Fatalf("expected 200 OK on list, got %d: %s", wList.Code, wList.Body.String())
	}

	var listResp struct {
		Success bool                   `json:"success"`
		Data    []model.DeviceActuator `json:"data"`
	}
	if err := json.Unmarshal(wList.Body.Bytes(), &listResp); err != nil {
		t.Fatalf("unmarshal list: %v", err)
	}
	if len(listResp.Data) != 2 {
		t.Fatalf("expected 2 actuators on device 5, got %d", len(listResp.Data))
	}
	if listResp.Data[1].SafetyWatchdogActive || listResp.Data[1].Enabled {
		t.Fatalf("expected explicit false booleans to persist, got watchdog=%v enabled=%v",
			listResp.Data[1].SafetyWatchdogActive, listResp.Data[1].Enabled)
	}
}

