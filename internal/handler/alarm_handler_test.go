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
	"datalogger/internal/handler"
	"datalogger/internal/model"
	"datalogger/internal/repository"
	"datalogger/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func setupTestAlarmRouter(t *testing.T) (*gin.Engine, *gorm.DB, *service.AlarmService) {
	gin.SetMode(gin.TestMode)

	dsn := fmt.Sprintf("file:mem_alarm_api_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		t.Fatalf("failed opening test sqlite db: %v", err)
	}

	err = db.AutoMigrate(
		&model.DeviceType{},
		&model.Device{},
		&model.DeviceConnection{},
		&model.Parameter{},
		&model.Alarm{},
		&model.AuditTrail{},
		&model.SystemLog{},
	)
	if err != nil {
		t.Fatalf("failed migrating test models: %v", err)
	}

	cfg := &config.Config{AppName: "Alarm Test"}
	systemRepo := repository.NewSystemRepository(db)
	phaseRepo := repository.NewPhaseRepository(db)
	systemService := service.NewSystemService(systemRepo, phaseRepo, cfg)

	alarmRepo := repository.NewAlarmRepository(db)
	alarmService := service.NewAlarmService(alarmRepo, systemRepo)
	systemService.SetAlarmService(alarmService)

	systemHandler := handler.NewSystemHandler(systemService)

	r := gin.New()
	api := r.Group("/api")
	{
		api.GET("/alarms", systemHandler.GetAlarms)
		api.GET("/alarms/active", systemHandler.GetActiveAlarms)
		api.GET("/alarms/:id", systemHandler.GetAlarmByID)
		api.POST("/alarms", systemHandler.CreateAlarm)
		api.PUT("/alarms/:id/acknowledge", systemHandler.AcknowledgeAlarm)
		api.PUT("/alarms/:id/clear", systemHandler.ClearAlarm)
	}

	return r, db, alarmService
}

func TestAlarmAPI_CreateAndGetByID(t *testing.T) {
	r, _, _ := setupTestAlarmRouter(t)

	// 1. Create Alarm via POST /api/alarms
	payload := map[string]interface{}{
		"alarm_code":      "OVERHEAT",
		"severity":        "CRITICAL",
		"message":         "Main bearing temperature exceeded 95C",
		"trigger_value":   98.2,
		"threshold_value": 95.0,
	}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest(http.MethodPost, "/api/alarms", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201 Created, got %d: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Success bool        `json:"success"`
		Data    model.Alarm `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed unmarshaling response: %v", err)
	}
	if resp.Data.ID == 0 {
		t.Fatalf("expected non-zero ID, got 0")
	}
	if resp.Data.Status != model.AlarmActive {
		t.Errorf("expected status ACTIVE, got %s", resp.Data.Status)
	}

	// 2. Retrieve Alarm via GET /api/alarms/:id
	reqGet, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/api/alarms/%d", resp.Data.ID), nil)
	wGet := httptest.NewRecorder()
	r.ServeHTTP(wGet, reqGet)

	if wGet.Code != http.StatusOK {
		t.Fatalf("expected status 200 OK, got %d: %s", wGet.Code, wGet.Body.String())
	}

	var respGet struct {
		Success bool        `json:"success"`
		Data    model.Alarm `json:"data"`
	}
	_ = json.Unmarshal(wGet.Body.Bytes(), &respGet)
	if respGet.Data.AlarmCode != "OVERHEAT" {
		t.Errorf("expected alarm_code OVERHEAT, got %s", respGet.Data.AlarmCode)
	}

	// 3. Non-existent ID returns 404
	req404, _ := http.NewRequest(http.MethodGet, "/api/alarms/99999", nil)
	w404 := httptest.NewRecorder()
	r.ServeHTTP(w404, req404)
	if w404.Code != http.StatusNotFound {
		t.Errorf("expected status 404 for non-existent alarm, got %d", w404.Code)
	}
}

func TestAlarmAPI_ListBackwardCompatibilityAndPagination(t *testing.T) {
	r, db, _ := setupTestAlarmRouter(t)

	// Seed 3 alarms
	db.Create(&model.Alarm{AlarmCode: "A1", Severity: model.SeverityWarning, Message: "M1", Status: model.AlarmActive, TriggeredAt: time.Now()})
	db.Create(&model.Alarm{AlarmCode: "A2", Severity: model.SeverityCritical, Message: "M2", Status: model.AlarmActive, TriggeredAt: time.Now()})
	db.Create(&model.Alarm{AlarmCode: "A3", Severity: model.SeverityInfo, Message: "M3", Status: model.AlarmCleared, TriggeredAt: time.Now()})

	// 1. Backward-compatible unpaginated GET /api/alarms returns direct array in `data`
	reqList, _ := http.NewRequest(http.MethodGet, "/api/alarms", nil)
	wList := httptest.NewRecorder()
	r.ServeHTTP(wList, reqList)

	if wList.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", wList.Code)
	}

	var rawResp struct {
		Success bool          `json:"success"`
		Data    []model.Alarm `json:"data"`
	}
	if err := json.Unmarshal(wList.Body.Bytes(), &rawResp); err != nil {
		t.Fatalf("failed decoding backward-compatible array response: %v", err)
	}
	if len(rawResp.Data) != 3 {
		t.Errorf("expected 3 items in unpaginated list, got %d", len(rawResp.Data))
	}

	// 2. Active Alarms endpoint GET /api/alarms/active
	reqActive, _ := http.NewRequest(http.MethodGet, "/api/alarms/active", nil)
	wActive := httptest.NewRecorder()
	r.ServeHTTP(wActive, reqActive)

	var activeResp struct {
		Success bool          `json:"success"`
		Data    []model.Alarm `json:"data"`
	}
	_ = json.Unmarshal(wActive.Body.Bytes(), &activeResp)
	if len(activeResp.Data) != 2 {
		t.Errorf("expected 2 active alarms, got %d", len(activeResp.Data))
	}

	// 3. Paginated GET /api/alarms?page=1&page_size=2
	reqPage, _ := http.NewRequest(http.MethodGet, "/api/alarms?page=1&page_size=2", nil)
	wPage := httptest.NewRecorder()
	r.ServeHTTP(wPage, reqPage)

	var pageResp struct {
		Success bool `json:"success"`
		Data    struct {
			Items    []model.Alarm `json:"items"`
			Total    int64         `json:"total"`
			Page     int           `json:"page"`
			PageSize int           `json:"page_size"`
		} `json:"data"`
	}
	if err := json.Unmarshal(wPage.Body.Bytes(), &pageResp); err != nil {
		t.Fatalf("failed decoding paginated response: %v", err)
	}
	if pageResp.Data.Total != 3 || len(pageResp.Data.Items) != 2 {
		t.Errorf("expected paginated total 3, page items 2; got total %d, items %d", pageResp.Data.Total, len(pageResp.Data.Items))
	}
}

func TestAlarmAPI_LifecycleAcknowledgeAndClear(t *testing.T) {
	r, db, _ := setupTestAlarmRouter(t)

	alarm := model.Alarm{
		AlarmCode:   "VALVE_STUCK",
		Severity:    model.SeverityCritical,
		Message:     "Emergency shutoff valve failed to close",
		Status:      model.AlarmActive,
		TriggeredAt: time.Now(),
	}
	db.Create(&alarm)

	// 1. Acknowledge alarm PUT /api/alarms/:id/acknowledge
	reqAck, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("/api/alarms/%d/acknowledge", alarm.ID), nil)
	wAck := httptest.NewRecorder()
	r.ServeHTTP(wAck, reqAck)

	if wAck.Code != http.StatusOK {
		t.Fatalf("expected 200 OK on acknowledge, got %d: %s", wAck.Code, wAck.Body.String())
	}

	var ackResp struct {
		Success bool `json:"success"`
		Data    struct {
			Acknowledged   bool   `json:"acknowledged"`
			AcknowledgedBy string `json:"acknowledged_by"`
		} `json:"data"`
	}
	_ = json.Unmarshal(wAck.Body.Bytes(), &ackResp)
	if !ackResp.Data.Acknowledged || ackResp.Data.AcknowledgedBy != "admin" {
		t.Errorf("expected acknowledged true by admin, got %v by %s", ackResp.Data.Acknowledged, ackResp.Data.AcknowledgedBy)
	}

	// 2. Clear alarm PUT /api/alarms/:id/clear
	reqClear, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("/api/alarms/%d/clear", alarm.ID), nil)
	wClear := httptest.NewRecorder()
	r.ServeHTTP(wClear, reqClear)

	if wClear.Code != http.StatusOK {
		t.Fatalf("expected 200 OK on clear, got %d: %s", wClear.Code, wClear.Body.String())
	}

	// 3. Attempting to clear already CLEARED alarm returns 400 Bad Request
	reqClearAgain, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("/api/alarms/%d/clear", alarm.ID), nil)
	wClearAgain := httptest.NewRecorder()
	r.ServeHTTP(wClearAgain, reqClearAgain)

	if wClearAgain.Code != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request on re-clearing, got %d", wClearAgain.Code)
	}

	// 4. Attempting to acknowledge a CLEARED alarm returns 400 Bad Request
	reqAckCleared, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("/api/alarms/%d/acknowledge", alarm.ID), nil)
	wAckCleared := httptest.NewRecorder()
	r.ServeHTTP(wAckCleared, reqAckCleared)

	if wAckCleared.Code != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request on acknowledging cleared alarm, got %d", wAckCleared.Code)
	}
}
