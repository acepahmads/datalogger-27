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
		api.GET("/alarms/history", systemHandler.GetAlarmHistory)
		api.GET("/alarms/:id", systemHandler.GetAlarmByID)
		api.POST("/alarms", systemHandler.CreateAlarm)
		api.PUT("/alarms/:id/acknowledge", systemHandler.AcknowledgeAlarm)
		api.PUT("/alarms/:id/clear", systemHandler.ClearAlarm)
	}

	return r, db, alarmService
}

func setupAuthenticatedAlarmRouter(t *testing.T) (*gin.Engine, *gorm.DB, string, string, string, *service.AlarmService, *repository.SystemRepository) {
	gin.SetMode(gin.TestMode)

	dsn := fmt.Sprintf("file:mem_alarm_auth_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		t.Fatalf("failed opening test sqlite db: %v", err)
	}

	_ = database.RunMigrations(db)
	_ = database.Seed(db)

	cfg := config.DefaultConfig()
	systemRepo := repository.NewSystemRepository(db)
	phaseRepo := repository.NewPhaseRepository(db)
	systemService := service.NewSystemService(systemRepo, phaseRepo, cfg)

	alarmRepo := repository.NewAlarmRepository(db)
	alarmService := service.NewAlarmService(alarmRepo, systemRepo)
	systemService.SetAlarmService(alarmService)

	authService := service.NewAuthService(systemRepo, cfg)
	systemHandler := handler.NewSystemHandler(systemService)

	r := gin.New()
	api := r.Group("/api")
	{
		alarmsGroup := api.Group("/alarms")
		alarmsGroup.Use(middleware.JWTAuth(authService))
		{
			alarmsGroup.GET("", middleware.RequirePermission("alarm.view"), systemHandler.GetAlarms)
			alarmsGroup.GET("/active", middleware.RequirePermission("alarm.view"), systemHandler.GetActiveAlarms)
			alarmsGroup.GET("/history", middleware.RequirePermission("alarm.view"), systemHandler.GetAlarmHistory)
			alarmsGroup.DELETE("/history", middleware.RequireRole("SUPERADMIN", "ADMINISTRATOR", "ADMIN"), systemHandler.PurgeAlarmHistory)
			alarmsGroup.POST("/history/purge", middleware.RequireRole("SUPERADMIN", "ADMINISTRATOR", "ADMIN"), systemHandler.PurgeAlarmHistory)
			alarmsGroup.GET("/:id", middleware.RequirePermission("alarm.view"), systemHandler.GetAlarmByID)
			alarmsGroup.DELETE("/:id", middleware.RequireRole("SUPERADMIN", "ADMINISTRATOR", "ADMIN"), systemHandler.DeleteAlarm)
			alarmsGroup.POST("", middleware.RequirePermission("alarm.manage"), systemHandler.CreateAlarm)
			alarmsGroup.PUT("/:id/acknowledge", middleware.RequirePermission("alarm.acknowledge"), systemHandler.AcknowledgeAlarm)
			alarmsGroup.PUT("/:id/clear", middleware.RequirePermission("alarm.manage"), systemHandler.ClearAlarm)
		}
	}

	// Create test users
	var opRole, engRole model.Role
	db.Where("name = ? OR name = ?", "Operator", "OPERATOR").First(&opRole)
	db.Where("name = ? OR name = ?", "Engineer", "ENGINEER").First(&engRole)

	opHash, _ := bcrypt.GenerateFromPassword([]byte("operator123"), bcrypt.DefaultCost)
	db.Create(&model.User{
		Username: "operator_test",
		Email:    "operator@test.local",
		Password: string(opHash),
		FullName: "Test Operator",
		RoleID:   opRole.ID,
		IsActive: true,
	})

	engHash, _ := bcrypt.GenerateFromPassword([]byte("engineer123"), bcrypt.DefaultCost)
	db.Create(&model.User{
		Username: "engineer_test",
		Email:    "engineer@test.local",
		Password: string(engHash),
		FullName: "Test Engineer",
		RoleID:   engRole.ID,
		IsActive: true,
	})

	adminToken, _, errA := authService.Login("admin", "admin123", "127.0.0.1", "TestClient")
	if errA != nil {
		t.Fatalf("failed login as admin: %v", errA)
	}
	operatorToken, _, errO := authService.Login("operator_test", "operator123", "127.0.0.1", "TestClient")
	if errO != nil {
		t.Fatalf("failed login as operator: %v", errO)
	}
	engineerToken, _, errE := authService.Login("engineer_test", "engineer123", "127.0.0.1", "TestClient")
	if errE != nil {
		t.Fatalf("failed login as engineer: %v", errE)
	}

	return r, db, adminToken, engineerToken, operatorToken, alarmService, systemRepo
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

	// 1. Acknowledge alarm PUT /api/alarms/:id/acknowledge with optional notes
	ackBody := bytes.NewBufferString(`{"notes":"Bearing inspected, lubrication added"}`)
	reqAck, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("/api/alarms/%d/acknowledge", alarm.ID), ackBody)
	reqAck.Header.Set("Content-Type", "application/json")
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
			Notes          string `json:"notes"`
		} `json:"data"`
	}
	_ = json.Unmarshal(wAck.Body.Bytes(), &ackResp)
	if !ackResp.Data.Acknowledged || ackResp.Data.AcknowledgedBy != "admin" {
		t.Errorf("expected acknowledged true by admin, got %v by %s", ackResp.Data.Acknowledged, ackResp.Data.AcknowledgedBy)
	}
	if ackResp.Data.Notes != "Bearing inspected, lubrication added" {
		t.Errorf("expected notes 'Bearing inspected, lubrication added', got '%s'", ackResp.Data.Notes)
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

func TestAlarmAPI_Wave4_AlarmAcknowledgement(t *testing.T) {
	r, db, adminToken, _, operatorToken, svc, systemRepo := setupAuthenticatedAlarmRouter(t)
	ctx := context.Background()

	// 1. Create an active alarm
	alarm := &model.Alarm{
		AlarmCode: "MOTOR_VIB_HIGH",
		Severity:  model.SeverityWarning,
		Message:   "Motor #2 vibration above baseline",
	}
	if err := svc.CreateAlarm(ctx, alarm); err != nil {
		t.Fatalf("failed creating test alarm: %v", err)
	}

	// 2. Unauthenticated request returns 401 Unauthorized
	reqUnauth, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("/api/alarms/%d/acknowledge", alarm.ID), bytes.NewBufferString(`{}`))
	wUnauth := httptest.NewRecorder()
	r.ServeHTTP(wUnauth, reqUnauth)
	if wUnauth.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 Unauthorized without token, got %d", wUnauth.Code)
	}

	// 3. Operator acknowledgement with server-derived identity and optional notes
	// Notice: Client attempts to spoof user_id and username in payload; server MUST ignore them!
	payload := map[string]interface{}{
		"notes":    "Inspected motor mounting bolts, tightened loose bracket",
		"user_id":  99999,
		"username": "spoofed_hacker",
	}
	body, _ := json.Marshal(payload)
	reqAck, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("/api/alarms/%d/acknowledge", alarm.ID), bytes.NewBuffer(body))
	reqAck.Header.Set("Authorization", "Bearer "+operatorToken)
	reqAck.Header.Set("Content-Type", "application/json")
	wAck := httptest.NewRecorder()
	r.ServeHTTP(wAck, reqAck)

	if wAck.Code != http.StatusOK {
		t.Fatalf("expected 200 OK on operator acknowledgement, got %d: %s", wAck.Code, wAck.Body.String())
	}

	var resp struct {
		Success bool `json:"success"`
		Data    struct {
			Acknowledged         bool   `json:"acknowledged"`
			ID                   uint   `json:"id"`
			AlarmCode            string `json:"alarm_code"`
			Status               string `json:"status"`
			AcknowledgedBy       string `json:"acknowledged_by"`
			AcknowledgedByUserID *uint  `json:"acknowledged_by_user_id"`
			AcknowledgedByName   string `json:"acknowledged_by_name"`
			Notes                string `json:"notes"`
		} `json:"data"`
	}
	if err := json.Unmarshal(wAck.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed unmarshaling response: %v", err)
	}

	if !resp.Data.Acknowledged || resp.Data.Status != "ACKNOWLEDGED" {
		t.Errorf("expected acknowledged true with status ACKNOWLEDGED, got %v / %s", resp.Data.Acknowledged, resp.Data.Status)
	}
	if resp.Data.AcknowledgedBy != "operator_test" {
		t.Errorf("expected server-derived acknowledged_by 'operator_test', got '%s'", resp.Data.AcknowledgedBy)
	}
	if resp.Data.AcknowledgedByName != "Test Operator" {
		t.Errorf("expected acknowledged_by_name 'Test Operator', got '%s'", resp.Data.AcknowledgedByName)
	}
	if resp.Data.AcknowledgedByUserID == nil || *resp.Data.AcknowledgedByUserID == 99999 {
		t.Errorf("spoofed user_id was accepted instead of server-derived ID: %v", resp.Data.AcknowledgedByUserID)
	}
	if resp.Data.Notes != "Inspected motor mounting bolts, tightened loose bracket" {
		t.Errorf("expected operator notes to match payload, got '%s'", resp.Data.Notes)
	}

	// 4. Verify audit trail record was created
	trails, err := systemRepo.GetAuditTrails(50)
	if err != nil {
		t.Fatalf("failed getting audit trails: %v", err)
	}
	var ackTrail *model.AuditTrail
	for i := range trails {
		if trails[i].Action == "ALARM_ACKNOWLEDGE" {
			ackTrail = &trails[i]
			break
		}
	}
	if ackTrail == nil {
		t.Fatalf("audit trail record not found for ALARM_ACKNOWLEDGE")
	}
	if ackTrail.Username != "operator_test" {
		t.Errorf("expected audit trail username 'operator_test', got '%s'", ackTrail.Username)
	}

	// 5. Duplicate acknowledgement on already acknowledged alarm returns 400 Bad Request
	reqDup, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("/api/alarms/%d/acknowledge", alarm.ID), bytes.NewBuffer(body))
	reqDup.Header.Set("Authorization", "Bearer "+operatorToken)
	wDup := httptest.NewRecorder()
	r.ServeHTTP(wDup, reqDup)
	if wDup.Code != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request on duplicate acknowledgement, got %d: %s", wDup.Code, wDup.Body.String())
	}

	// 6. Nonexistent alarm ID returns 404 Not Found
	req404, _ := http.NewRequest(http.MethodPut, "/api/alarms/999999/acknowledge", bytes.NewBuffer(body))
	req404.Header.Set("Authorization", "Bearer "+operatorToken)
	w404 := httptest.NewRecorder()
	r.ServeHTTP(w404, req404)
	if w404.Code != http.StatusNotFound {
		t.Errorf("expected 404 Not Found for nonexistent alarm, got %d", w404.Code)
	}

	// 7. Clear alarm using Admin token
	reqClear, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("/api/alarms/%d/clear", alarm.ID), nil)
	reqClear.Header.Set("Authorization", "Bearer "+adminToken)
	wClear := httptest.NewRecorder()
	r.ServeHTTP(wClear, reqClear)
	if wClear.Code != http.StatusOK {
		t.Fatalf("expected 200 OK on clear, got %d", wClear.Code)
	}

	// 8. Attempting to acknowledge a CLEARED alarm returns 400 Bad Request
	reqAckCleared, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("/api/alarms/%d/acknowledge", alarm.ID), bytes.NewBuffer(body))
	reqAckCleared.Header.Set("Authorization", "Bearer "+operatorToken)
	wAckCleared := httptest.NewRecorder()
	r.ServeHTTP(wAckCleared, reqAckCleared)
	if wAckCleared.Code != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request on acknowledging cleared alarm, got %d", wAckCleared.Code)
	}

	_ = db
}

func TestAlarmAPI_Wave4_RBAC(t *testing.T) {
	r, _, adminToken, engineerToken, operatorToken, svc, _ := setupAuthenticatedAlarmRouter(t)
	ctx := context.Background()

	alarm := &model.Alarm{
		AlarmCode: "RBAC_TEST",
		Severity:  model.SeverityInfo,
		Message:   "RBAC test message",
	}
	_ = svc.CreateAlarm(ctx, alarm)

	// 1. Operator role CAN view alarms (GET /api/alarms, GET /api/alarms/active, GET /api/alarms/history)
	for _, path := range []string{"/api/alarms", "/api/alarms/active", "/api/alarms/history"} {
		req, _ := http.NewRequest(http.MethodGet, path, nil)
		req.Header.Set("Authorization", "Bearer "+operatorToken)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("expected 200 OK for operator on %s, got %d", path, w.Code)
		}
	}

	// 2. Operator role CAN acknowledge alarms
	reqAck, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("/api/alarms/%d/acknowledge", alarm.ID), bytes.NewBufferString(`{"notes":"op ack"}`))
	reqAck.Header.Set("Authorization", "Bearer "+operatorToken)
	wAck := httptest.NewRecorder()
	r.ServeHTTP(wAck, reqAck)
	if wAck.Code != http.StatusOK {
		t.Errorf("expected 200 OK for operator acknowledge, got %d", wAck.Code)
	}

	// 3. Operator role CANNOT create alarms (POST /api/alarms requires alarm.manage)
	newAlarmBody, _ := json.Marshal(map[string]interface{}{
		"alarm_code": "OP_FORBIDDEN",
		"severity":   "WARNING",
		"message":    "Forbidden create",
	})
	reqCreate, _ := http.NewRequest(http.MethodPost, "/api/alarms", bytes.NewBuffer(newAlarmBody))
	reqCreate.Header.Set("Authorization", "Bearer "+operatorToken)
	wCreate := httptest.NewRecorder()
	r.ServeHTTP(wCreate, reqCreate)
	if wCreate.Code != http.StatusForbidden {
		t.Errorf("expected 403 Forbidden for operator creating alarm, got %d", wCreate.Code)
	}

	// 4. Operator role CANNOT clear alarms (PUT /api/alarms/:id/clear requires alarm.manage)
	reqClear, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("/api/alarms/%d/clear", alarm.ID), nil)
	reqClear.Header.Set("Authorization", "Bearer "+operatorToken)
	wClear := httptest.NewRecorder()
	r.ServeHTTP(wClear, reqClear)
	if wClear.Code != http.StatusForbidden {
		t.Errorf("expected 403 Forbidden for operator clearing alarm, got %d", wClear.Code)
	}

	// 5. Engineer role CAN clear alarms
	reqEngClear, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("/api/alarms/%d/clear", alarm.ID), nil)
	reqEngClear.Header.Set("Authorization", "Bearer "+engineerToken)
	wEngClear := httptest.NewRecorder()
	r.ServeHTTP(wEngClear, reqEngClear)
	if wEngClear.Code != http.StatusOK {
		t.Errorf("expected 200 OK for engineer clearing alarm, got %d", wEngClear.Code)
	}

	// 6. Admin role CAN create alarms
	reqAdminCreate, _ := http.NewRequest(http.MethodPost, "/api/alarms", bytes.NewBuffer(newAlarmBody))
	reqAdminCreate.Header.Set("Authorization", "Bearer "+adminToken)
	wAdminCreate := httptest.NewRecorder()
	r.ServeHTTP(wAdminCreate, reqAdminCreate)
	if wAdminCreate.Code != http.StatusCreated {
		t.Errorf("expected 201 Created for admin creating alarm, got %d", wAdminCreate.Code)
	}
}

func TestAlarmAPI_Wave4_AlarmHistory(t *testing.T) {
	r, db, adminToken, _, _, svc, _ := setupAuthenticatedAlarmRouter(t)
	ctx := context.Background()

	now := time.Now().UTC()
	t1 := now.Add(-3 * time.Hour)
	t2 := now.Add(-2 * time.Hour)
	t3 := now.Add(-1 * time.Hour)

	triggerVal1 := 98.6
	threshVal1 := 90.0
	clearedTime1 := now.Add(-30 * time.Minute)

	// Seed 3 alarms:
	// 1. Cleared alarm with trigger/threshold context
	alm1 := &model.Alarm{
		AlarmCode:      "HIGH_TEMP",
		Severity:       model.SeverityCritical,
		Status:         model.AlarmCleared,
		Message:        "Chiller temperature critically high",
		TriggerValue:   &triggerVal1,
		ThresholdValue: &threshVal1,
		TriggeredAt:    t1,
		AcknowledgedAt: &t2,
		AcknowledgedBy: "operator_test",
		Notes:          "Coolant flow valve reopened",
		ClearedAt:      &clearedTime1,
	}
	_ = svc.CreateAlarm(ctx, alm1)

	// 2. Active alarm
	triggerVal2 := 370.0
	threshVal2 := 380.0
	alm2 := &model.Alarm{
		AlarmCode:      "LOW_VOLTAGE",
		Severity:       model.SeverityWarning,
		Status:         model.AlarmActive,
		Message:        "Line voltage below 380V threshold",
		TriggerValue:   &triggerVal2,
		ThresholdValue: &threshVal2,
		TriggeredAt:    t2,
	}
	_ = svc.CreateAlarm(ctx, alm2)

	// 3. Acknowledged alarm without numeric values (e.g. comm error - nulls preserved)
	alm3 := &model.Alarm{
		AlarmCode:      "COMMUNICATION_ERROR_TIMEOUT",
		Severity:       model.SeverityCritical,
		Status:         model.AlarmAcked,
		Message:        "Modbus device #1 timeout after 3 retries",
		TriggeredAt:    t3,
		AcknowledgedAt: &now,
		AcknowledgedBy: "admin",
		Notes:          "Checking RS-485 bus wiring",
	}
	_ = svc.CreateAlarm(ctx, alm3)

	// 1. Paginated Query (page 1, page_size 2)
	reqPage, _ := http.NewRequest(http.MethodGet, "/api/alarms/history?page=1&page_size=2", nil)
	reqPage.Header.Set("Authorization", "Bearer "+adminToken)
	wPage := httptest.NewRecorder()
	r.ServeHTTP(wPage, reqPage)

	if wPage.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for alarm history, got %d: %s", wPage.Code, wPage.Body.String())
	}

	var pageResp struct {
		Success bool `json:"success"`
		Data    struct {
			Items []struct {
				ID                uint     `json:"id"`
				AlarmCode         string   `json:"alarm_code"`
				Severity          string   `json:"severity"`
				Status            string   `json:"status"`
				DurationSeconds   int64    `json:"duration_seconds"`
				DurationFormatted string   `json:"duration_formatted"`
				IsOngoing         bool     `json:"is_ongoing"`
				TriggerValue      *float64 `json:"trigger_value"`
				ThresholdValue    *float64 `json:"threshold_value"`
				RecoveryReason    string   `json:"recovery_reason"`
			} `json:"items"`
			Total      int64 `json:"total"`
			Page       int   `json:"page"`
			PageSize   int   `json:"page_size"`
			TotalPages int   `json:"total_pages"`
			HasNext    bool  `json:"has_next"`
			HasPrev    bool  `json:"has_prev"`
		} `json:"data"`
	}
	if err := json.Unmarshal(wPage.Body.Bytes(), &pageResp); err != nil {
		t.Fatalf("failed unmarshaling page response: %v", err)
	}

	if pageResp.Data.Total < 3 {
		t.Errorf("expected total >= 3, got %d", pageResp.Data.Total)
	}
	if len(pageResp.Data.Items) != 2 {
		t.Errorf("expected 2 items on page 1, got %d", len(pageResp.Data.Items))
	}
	if !pageResp.Data.HasNext {
		t.Errorf("expected has_next true on page 1 of 2")
	}

	// 2. Filter by status=CLEARED
	reqCleared, _ := http.NewRequest(http.MethodGet, "/api/alarms/history?status=CLEARED", nil)
	reqCleared.Header.Set("Authorization", "Bearer "+adminToken)
	wCleared := httptest.NewRecorder()
	r.ServeHTTP(wCleared, reqCleared)

	var clearedResp struct {
		Data struct {
			Items []struct {
				AlarmCode         string   `json:"alarm_code"`
				Status            string   `json:"status"`
				IsOngoing         bool     `json:"is_ongoing"`
				DurationSeconds   int64    `json:"duration_seconds"`
				DurationFormatted string   `json:"duration_formatted"`
				TriggerValue      *float64 `json:"trigger_value"`
				ThresholdValue    *float64 `json:"threshold_value"`
				RecoveryReason    string   `json:"recovery_reason"`
			} `json:"items"`
		} `json:"data"`
	}
	_ = json.Unmarshal(wCleared.Body.Bytes(), &clearedResp)
	if len(clearedResp.Data.Items) == 0 {
		t.Fatalf("expected at least 1 cleared alarm")
	}
	item0 := clearedResp.Data.Items[0]
	if item0.Status != "CLEARED" {
		t.Errorf("expected status CLEARED, got %s", item0.Status)
	}
	if item0.IsOngoing {
		t.Errorf("expected is_ongoing false for cleared alarm")
	}
	if item0.DurationSeconds <= 0 || item0.DurationFormatted == "" {
		t.Errorf("expected valid duration for cleared alarm, got %d / '%s'", item0.DurationSeconds, item0.DurationFormatted)
	}
	if item0.TriggerValue == nil || *item0.TriggerValue != 98.6 {
		t.Errorf("expected trigger_value 98.6, got %v", item0.TriggerValue)
	}

	// 3. Filter by severity=CRITICAL
	reqCrit, _ := http.NewRequest(http.MethodGet, "/api/alarms/history?severity=CRITICAL", nil)
	reqCrit.Header.Set("Authorization", "Bearer "+adminToken)
	wCrit := httptest.NewRecorder()
	r.ServeHTTP(wCrit, reqCrit)
	var critResp struct {
		Data struct {
			Total int64 `json:"total"`
		} `json:"data"`
	}
	_ = json.Unmarshal(wCrit.Body.Bytes(), &critResp)
	if critResp.Data.Total < 2 {
		t.Errorf("expected at least 2 CRITICAL alarms, got %d", critResp.Data.Total)
	}

	// 4. Filter by alarm_code
	reqCode, _ := http.NewRequest(http.MethodGet, "/api/alarms/history?alarm_code=LOW_VOLTAGE", nil)
	reqCode.Header.Set("Authorization", "Bearer "+adminToken)
	wCode := httptest.NewRecorder()
	r.ServeHTTP(wCode, reqCode)
	var codeResp struct {
		Data struct {
			Items []struct {
				AlarmCode string `json:"alarm_code"`
				Status    string `json:"status"`
				IsOngoing bool   `json:"is_ongoing"`
			} `json:"items"`
		} `json:"data"`
	}
	_ = json.Unmarshal(wCode.Body.Bytes(), &codeResp)
	if len(codeResp.Data.Items) != 1 || codeResp.Data.Items[0].AlarmCode != "LOW_VOLTAGE" {
		t.Errorf("expected exactly 1 LOW_VOLTAGE alarm, got %v", codeResp.Data.Items)
	}
	if !codeResp.Data.Items[0].IsOngoing {
		t.Errorf("expected is_ongoing true for active alarm")
	}

	// 5. Un-fabricated legacy null values: alarm 3 has null TriggerValue and ThresholdValue
	reqNull, _ := http.NewRequest(http.MethodGet, "/api/alarms/history?alarm_code=COMMUNICATION_ERROR_TIMEOUT", nil)
	reqNull.Header.Set("Authorization", "Bearer "+adminToken)
	wNull := httptest.NewRecorder()
	r.ServeHTTP(wNull, reqNull)
	var nullResp struct {
		Data struct {
			Items []struct {
				TriggerValue   *float64 `json:"trigger_value"`
				ThresholdValue *float64 `json:"threshold_value"`
			} `json:"items"`
		} `json:"data"`
	}
	_ = json.Unmarshal(wNull.Body.Bytes(), &nullResp)
	if len(nullResp.Data.Items) > 0 {
		if nullResp.Data.Items[0].TriggerValue != nil || nullResp.Data.Items[0].ThresholdValue != nil {
			t.Errorf("expected trigger_value and threshold_value to remain null for communication alarm, got %v / %v",
				nullResp.Data.Items[0].TriggerValue, nullResp.Data.Items[0].ThresholdValue)
		}
	}

	_ = db
}

func TestAlarmAPI_HistoryDeleteAndDeleteAllRBAC(t *testing.T) {
	r, db, superadminToken, engineerToken, operatorToken, svc, systemRepo := setupAuthenticatedAlarmRouter(t)
	ctx := context.Background()

	// Create explicit ADMIN role user (seed 'admin' user is Superadmin)
	var adminRole model.Role
	if err := db.Where("UPPER(name) = ?", "ADMIN").First(&adminRole).Error; err != nil || adminRole.ID == 0 {
		adminRole = model.Role{Name: "Admin", Description: "Admin role"}
		db.Create(&adminRole)
	}
	adminHash, _ := bcrypt.GenerateFromPassword([]byte("adminonly123"), bcrypt.DefaultCost)
	db.Create(&model.User{
		Username: "admin_history_test",
		Email:    "admin_history@test.local",
		Password: string(adminHash),
		FullName: "Admin History User",
		RoleID:   adminRole.ID,
		IsActive: true,
	})
	authSvc := service.NewAuthService(systemRepo, config.DefaultConfig())
	adminRoleToken, _, errLogin := authSvc.Login("admin_history_test", "adminonly123", "127.0.0.1", "TestClient")
	if errLogin != nil {
		t.Fatalf("failed login as admin_history_test: %v", errLogin)
	}

	// Create 3 alarms (2 CLEARED, 1 ACTIVE) + dependent Notification rows
	now := time.Now().UTC()
	a1 := &model.Alarm{
		AlarmCode:   "HIST_DEL_1",
		Severity:    model.SeverityWarning,
		Message:     "Cleared history 1",
		Status:      model.AlarmCleared,
		TriggeredAt: now.Add(-10 * time.Minute),
		ClearedAt:   &now,
	}
	a2 := &model.Alarm{
		AlarmCode:   "HIST_DEL_2",
		Severity:    model.SeverityWarning,
		Message:     "Cleared history 2",
		Status:      model.AlarmCleared,
		TriggeredAt: now.Add(-5 * time.Minute),
		ClearedAt:   &now,
	}
	a3 := &model.Alarm{
		AlarmCode:   "HIST_DEL_3",
		Severity:    model.SeverityCritical,
		Message:     "Active incident 3",
		Status:      model.AlarmActive,
		TriggeredAt: now.Add(-1 * time.Minute),
	}
	_ = svc.CreateAlarm(ctx, a1)
	_ = svc.CreateAlarm(ctx, a2)
	_ = svc.CreateAlarm(ctx, a3)

	// Attach dependent notifications to verify FK-safe deletion
	db.Create(&model.Notification{
		AlarmID:   a1.ID,
		Channel:   "EMAIL",
		Recipient: "ops@plant.local",
		Message:   "Notif for a1",
		Status:    model.NotificationFailed,
	})
	db.Create(&model.Notification{
		AlarmID:   a2.ID,
		Channel:   "TELEGRAM",
		Recipient: "-100111",
		Message:   "Notif for a2",
		Status:    model.NotificationSent,
	})

	// 1. OPERATOR and ENGINEER MUST get 403 Forbidden on both single Delete and Delete All
	for _, tc := range []struct {
		roleName string
		token    string
	}{
		{roleName: "OPERATOR", token: operatorToken},
		{roleName: "ENGINEER", token: engineerToken},
	} {
		reqDel, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/api/alarms/%d", a1.ID), nil)
		reqDel.Header.Set("Authorization", "Bearer "+tc.token)
		wDel := httptest.NewRecorder()
		r.ServeHTTP(wDel, reqDel)
		if wDel.Code != http.StatusForbidden {
			t.Fatalf("expected 403 Forbidden for %s single alarm delete, got %d: %s", tc.roleName, wDel.Code, wDel.Body.String())
		}

		reqPurge, _ := http.NewRequest(http.MethodPost, "/api/alarms/history/purge", bytes.NewBufferString(`{}`))
		reqPurge.Header.Set("Authorization", "Bearer "+tc.token)
		reqPurge.Header.Set("Content-Type", "application/json")
		wPurge := httptest.NewRecorder()
		r.ServeHTTP(wPurge, reqPurge)
		if wPurge.Code != http.StatusForbidden {
			t.Fatalf("expected 403 Forbidden for %s alarm history purge, got %d: %s", tc.roleName, wPurge.Code, wPurge.Body.String())
		}
	}

	// 2. ADMIN role can delete a single alarm (and its dependent notification)
	reqAdminDel, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/api/alarms/%d", a1.ID), nil)
	reqAdminDel.Header.Set("Authorization", "Bearer "+adminRoleToken)
	wAdminDel := httptest.NewRecorder()
	r.ServeHTTP(wAdminDel, reqAdminDel)
	if wAdminDel.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for ADMIN single alarm delete, got %d: %s", wAdminDel.Code, wAdminDel.Body.String())
	}

	// Repeat delete -> 404 Not Found
	reqAgain, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/api/alarms/%d", a1.ID), nil)
	reqAgain.Header.Set("Authorization", "Bearer "+adminRoleToken)
	wAgain := httptest.NewRecorder()
	r.ServeHTTP(wAgain, reqAgain)
	if wAgain.Code != http.StatusNotFound {
		t.Errorf("expected 404 Not Found deleting already-deleted alarm, got %d", wAgain.Code)
	}

	// 3. SUPERADMIN can Delete All alarm history via POST /api/alarms/history/purge
	reqAll, _ := http.NewRequest(http.MethodPost, "/api/alarms/history/purge", bytes.NewBufferString(`{}`))
	reqAll.Header.Set("Authorization", "Bearer "+superadminToken)
	reqAll.Header.Set("Content-Type", "application/json")
	wAll := httptest.NewRecorder()
	r.ServeHTTP(wAll, reqAll)
	if wAll.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for SUPERADMIN Delete All alarm history, got %d: %s", wAll.Code, wAll.Body.String())
	}

	var purgeResp struct {
		Success bool `json:"success"`
		Data    struct {
			DeletedCount int64 `json:"deleted_count"`
		} `json:"data"`
	}
	_ = json.Unmarshal(wAll.Body.Bytes(), &purgeResp)
	if !purgeResp.Success || purgeResp.Data.DeletedCount != 2 {
		t.Errorf("expected deleted_count=2 on Delete All, got %+v", purgeResp)
	}

	var remainingAlarms, remainingNotifs int64
	db.Model(&model.Alarm{}).Count(&remainingAlarms)
	db.Model(&model.Notification{}).Count(&remainingNotifs)
	if remainingAlarms != 0 {
		t.Errorf("expected 0 alarms remaining after Delete All, got %d", remainingAlarms)
	}
	if remainingNotifs != 0 {
		t.Errorf("expected 0 dependent notifications remaining after Delete All, got %d", remainingNotifs)
	}
}

