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

func setupPackageBTestEnv(t *testing.T) (*gin.Engine, *gorm.DB, string, string, *service.SystemService, *service.NotificationEngine) {
	gin.SetMode(gin.TestMode)

	dsn := fmt.Sprintf("file:mem_pkg_b_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		t.Fatalf("Failed to open SQLite: %v", err)
	}

	_ = database.RunMigrations(db)
	_ = database.Seed(db)

	cfg := config.DefaultConfig()

	// Repositories
	systemRepo := repository.NewSystemRepository(db)
	alarmRepo := repository.NewAlarmRepository(db)
	notifRepo := repository.NewNotificationRepository(db)

	// Services
	alarmService := service.NewAlarmService(alarmRepo, systemRepo)
	systemService := service.NewSystemService(systemRepo, nil, cfg)
	systemService.SetAlarmService(alarmService)
	notifEngine := service.NewNotificationEngine(notifRepo, alarmRepo, systemRepo, service.DefaultNotificationEngineConfig())
	authService := service.NewAuthService(systemRepo, cfg)

	// Handlers
	systemHandler := handler.NewSystemHandler(systemService)
	notifHandler := handler.NewNotificationHandler(notifEngine, notifRepo)

	router := gin.New()
	api := router.Group("/api")
	{
		// System logs, audit trails, and communication logs
		api.GET("/logs", systemHandler.GetLogs)
		api.GET("/audit-trails", systemHandler.GetAuditTrails)
		api.GET("/communication-logs", systemHandler.GetCommunicationLogs)

		// Alarms with JWT auth
		alarmsGroup := api.Group("/alarms")
		alarmsGroup.Use(middleware.JWTAuth(authService))
		{
			alarmsGroup.PUT("/:id/acknowledge", middleware.RequirePermission("alarm.acknowledge"), systemHandler.AcknowledgeAlarm)
			alarmsGroup.PUT("/:id/clear", systemHandler.ClearAlarm)
		}

		// Notifications with JWT auth
		notifsGroup := api.Group("/notifications")
		notifsGroup.Use(middleware.JWTAuth(authService))
		{
			notifsGroup.GET("", middleware.RequirePermission("notification.view"), notifHandler.ListNotifications)
			notifsGroup.GET("/metrics", middleware.RequirePermission("notification.view"), notifHandler.GetMetrics)
			notifsGroup.GET("/channels", middleware.RequirePermission("notification.view"), notifHandler.ListChannels)
			notifsGroup.POST("/:id/retry", middleware.RequirePermission("notification.manage"), notifHandler.RetryNotification)
			notifsGroup.POST("/:id/cancel", middleware.RequirePermission("notification.manage"), notifHandler.CancelNotification)
		}
	}

	// Create operator user
	var opRole model.Role
	db.Where("name = ?", "Operator").First(&opRole)
	opHash, _ := bcrypt.GenerateFromPassword([]byte("operator123"), bcrypt.DefaultCost)
	db.Create(&model.User{
		Username: "operator_test",
		Email:    "operator@test.local",
		Password: string(opHash),
		FullName: "Test Operator",
		RoleID:   opRole.ID,
		IsActive: true,
	})

	// Login as admin and operator
	adminToken, _, errA := authService.Login("admin", "admin123", "127.0.0.1", "TestClient")
	if errA != nil {
		t.Fatalf("failed login as admin: %v", errA)
	}

	operatorToken, _, errO := authService.Login("operator_test", "operator123", "127.0.0.1", "TestClient")
	if errO != nil {
		t.Fatalf("failed login as operator: %v", errO)
	}

	return router, db, adminToken, operatorToken, systemService, notifEngine
}

// 1. TASK #74 — Alarm Dashboard: Acknowledge & Server-Side Audit Attribution
func TestPackageB_AlarmAcknowledgementAndAuditTrail(t *testing.T) {
	router, db, adminToken, _, _, _ := setupPackageBTestEnv(t)

	// Seed active alarm
	now := time.Now().UTC()
	val := 55.4
	alarm := model.Alarm{
		AlarmCode:    "AQMS_CO_HIGH",
		Severity:     model.SeverityCritical,
		Status:       model.AlarmActive,
		Message:      "Carbon Monoxide exceeds 50.0 ppm limit",
		TriggerValue: &val,
		TriggeredAt:  now,
	}
	db.Create(&alarm)

	// 1. Acknowledge alarm via PUT /api/alarms/:id/acknowledge with operator notes
	body, _ := json.Marshal(map[string]string{
		"notes": "Field sensor recalibration initiated by operator",
	})
	req, _ := http.NewRequest("PUT", fmt.Sprintf("/api/alarms/%d/acknowledge", alarm.ID), bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+adminToken)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK on alarm acknowledge, got %d (body: %s)", w.Code, w.Body.String())
	}

	// Verify DB state
	var updatedAlarm model.Alarm
	db.First(&updatedAlarm, alarm.ID)
	if updatedAlarm.Status != model.AlarmAcked {
		t.Errorf("Expected alarm status ACKNOWLEDGED, got %s", updatedAlarm.Status)
	}
	if updatedAlarm.AcknowledgedBy != "admin" {
		t.Errorf("Expected acknowledged_by admin, got %s", updatedAlarm.AcknowledgedBy)
	}

	// Verify Audit Trail entry generated automatically
	var audit model.AuditTrail
	err := db.Where("action = ? AND resource = ?", "ALARM_ACKNOWLEDGE", "ALARM").Order("id DESC").First(&audit).Error
	if err != nil {
		t.Fatalf("Expected audit trail entry for ALARM_ACKNOWLEDGE: %v", err)
	}
	if audit.Username != "admin" {
		t.Errorf("Expected audit trail username admin, got %s", audit.Username)
	}
	if !bytes.Contains([]byte(audit.Details), []byte("recalibration")) {
		t.Errorf("Expected operator notes in audit trail details, got: %s", audit.Details)
	}
}

// 2. TASK #75 — Audit Trail Viewer: Pagination, Filtering & Secret Redaction
func TestPackageB_AuditTrailFilteringPaginationAndRedaction(t *testing.T) {
	router, db, _, _, _, _ := setupPackageBTestEnv(t)

	// Seed diverse audit trails
	baseTime := time.Date(2026, 10, 12, 10, 0, 0, 0, time.UTC)
	for i := 1; i <= 15; i++ {
		actor := "operator_lead"
		if i%2 == 0 {
			actor = "admin_user"
		}
		action := "DEVICE_UPDATE"
		if i%3 == 0 {
			action = "CONFIG_CHANGE"
		}
		res := "DEVICE"
		if i%3 == 0 {
			res = "SYSTEM"
		}
		details := fmt.Sprintf("Updated device parameter %d", i)
		if i == 5 {
			details = "Configured SMTP password='SuperSecretPassword123' and api_key: 'top_secret_token_abc'"
		}

		db.Create(&model.AuditTrail{
			Username:  actor,
			Action:    action,
			Resource:  res,
			Details:   details,
			CreatedAt: baseTime.Add(time.Duration(i) * time.Minute),
		})
	}

	// Test 1: Paginated query GET /api/audit-trails?page=1&page_size=5
	req, _ := http.NewRequest("GET", "/api/audit-trails?page=1&page_size=5", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK, got %d: %s", w.Code, w.Body.String())
	}

	var pageResp struct {
		Code int `json:"code"`
		Data struct {
			Items      []model.AuditTrail `json:"items"`
			Total      int64              `json:"total"`
			Page       int                `json:"page"`
			PageSize   int                `json:"page_size"`
			TotalPages int                `json:"total_pages"`
		} `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &pageResp)

	if pageResp.Data.Total < 15 {
		t.Errorf("Expected at least 15 total items, got %d", pageResp.Data.Total)
	}
	if len(pageResp.Data.Items) != 5 {
		t.Errorf("Expected 5 items on page 1, got %d", len(pageResp.Data.Items))
	}

	// Test 2: Actor filter GET /api/audit-trails?username=admin_user
	reqFilter, _ := http.NewRequest("GET", "/api/audit-trails?username=admin_user&page=1&page_size=10", nil)
	wFilter := httptest.NewRecorder()
	router.ServeHTTP(wFilter, reqFilter)

	var filterResp struct {
		Data struct {
			Items []model.AuditTrail `json:"items"`
			Total int64              `json:"total"`
		} `json:"data"`
	}
	_ = json.Unmarshal(wFilter.Body.Bytes(), &filterResp)
	if filterResp.Data.Total != 7 {
		t.Errorf("Expected 7 records for admin_user, got %d", filterResp.Data.Total)
	}

	// Test 3: Secret redaction in audit trail details
	reqRedact, _ := http.NewRequest("GET", "/api/audit-trails?search=SuperSecretPassword123", nil)
	wRedact := httptest.NewRecorder()
	router.ServeHTTP(wRedact, reqRedact)

	bodyStr := wRedact.Body.String()
	if bytes.Contains([]byte(bodyStr), []byte("SuperSecretPassword123")) {
		t.Errorf("Sensitive password leaked in audit trail output! Response: %s", bodyStr)
	}
	if bytes.Contains([]byte(bodyStr), []byte("top_secret_token_abc")) {
		t.Errorf("Sensitive token leaked in audit trail output! Response: %s", bodyStr)
	}
	if !bytes.Contains([]byte(bodyStr), []byte("[REDACTED]")) {
		t.Errorf("Expected [REDACTED] token in audit trail details, got: %s", bodyStr)
	}
}

// 3. TASK #76 — Log Viewer: Level Filtering, Search, Redaction & Communication Logs
func TestPackageB_SystemLogFilteringSearchAndRedaction(t *testing.T) {
	router, db, _, _, _, _ := setupPackageBTestEnv(t)

	// Seed system logs
	db.Create(&model.SystemLog{
		Level:     "ERROR",
		Component: "database",
		Message:   "Connection lost to MariaDB broker with password='db_secret_pass'",
		Details:   "Reconnecting in 5s; auth_token='secret_session_token'",
		CreatedAt: time.Now().Add(-10 * time.Minute),
	})
	db.Create(&model.SystemLog{
		Level:     "WARN",
		Component: "modbus",
		Message:   "CRC timeout on slave 2",
		Details:   "Packet retry #2",
		CreatedAt: time.Now().Add(-5 * time.Minute),
	})
	db.Create(&model.SystemLog{
		Level:     "INFO",
		Component: "telemetry",
		Message:   "Telemetry ingestion active, 120 points/sec",
		Details:   "Healthy throughput",
		CreatedAt: time.Now(),
	})

	// Test 1: Filter by level=ERROR
	req, _ := http.NewRequest("GET", "/api/logs?level=ERROR", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var logResp struct {
		Data struct {
			Items []model.SystemLog `json:"items"`
			Total int64             `json:"total"`
		} `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &logResp)

	if logResp.Data.Total != 1 {
		t.Errorf("Expected 1 ERROR log, got %d", logResp.Data.Total)
	}
	if len(logResp.Data.Items) > 0 && logResp.Data.Items[0].Level != "ERROR" {
		t.Errorf("Expected level ERROR, got %s", logResp.Data.Items[0].Level)
	}

	// Verify Redaction in Message and Details
	logBody := w.Body.String()
	if bytes.Contains([]byte(logBody), []byte("db_secret_pass")) || bytes.Contains([]byte(logBody), []byte("secret_session_token")) {
		t.Errorf("Credential leaked in system logs! Body: %s", logBody)
	}
	if !bytes.Contains([]byte(logBody), []byte("[REDACTED]")) {
		t.Errorf("Expected [REDACTED] in sanitized log output, got: %s", logBody)
	}

	// Seed communication logs
	db.Create(&model.CommunicationLog{
		DeviceID:   1,
		Protocol:   "Modbus-RTU",
		Direction:  "TX",
		PayloadHex: "01030000000A45CD",
		Status:     "SUCCESS",
		LatencyMs:  18,
		CreatedAt:  time.Now().Add(-2 * time.Minute),
	})
	db.Create(&model.CommunicationLog{
		DeviceID:     2,
		Protocol:     "Modbus-TCP",
		Direction:    "TX",
		PayloadHex:   "000100000006010300000002",
		Status:       "TIMEOUT",
		ErrorMessage: "i/o timeout after 2000ms",
		LatencyMs:    2005,
		CreatedAt:    time.Now().Add(-1 * time.Minute),
	})

	// Test 2: GET /api/communication-logs?status=TIMEOUT
	reqComm, _ := http.NewRequest("GET", "/api/communication-logs?status=TIMEOUT", nil)
	wComm := httptest.NewRecorder()
	router.ServeHTTP(wComm, reqComm)

	var commResp struct {
		Data struct {
			Items []model.CommunicationLog `json:"items"`
			Total int64                    `json:"total"`
		} `json:"data"`
	}
	_ = json.Unmarshal(wComm.Body.Bytes(), &commResp)

	if commResp.Data.Total != 1 {
		t.Errorf("Expected 1 TIMEOUT comm log, got %d", commResp.Data.Total)
	}
	if len(commResp.Data.Items) > 0 && commResp.Data.Items[0].Status != "TIMEOUT" {
		t.Errorf("Expected status TIMEOUT, got %s", commResp.Data.Items[0].Status)
	}
}

// 4. TASK #77 — Delivery Status: Real Engine Data, Channel Config State & RBAC
func TestPackageB_NotificationDeliveryQueueAndRBAC(t *testing.T) {
	router, db, adminToken, operatorToken, _, _ := setupPackageBTestEnv(t)

	// Seed channels
	db.Create(&model.NotificationChannel{
		Type:            model.ChannelEmail,
		Name:            "Primary SMTP Dispatcher",
		Enabled:         true,
		MinSeverity:     model.SeverityCritical,
		RateLimitPerMin: 10,
		MaxRetries:      3,
		ConfigJSON:      "{}", // Empty credentials -> NOT CONFIGURED
	})

	// Seed queued notifications
	now := time.Now().UTC()
	notifFailed := model.Notification{
		AlarmID:     1,
		Channel:     "EMAIL",
		Recipient:   "admin@plant.local",
		Subject:     "Critical Alarm: Temperature Exceeded",
		Message:     "High temp alarm on Reactor 1",
		Status:      model.NotificationFailed,
		Attempts:    3,
		MaxAttempts: 3,
		ErrorMsg:    "SMTP server not configured",
		CreatedAt:   now.Add(-10 * time.Minute),
	}
	db.Create(&notifFailed)

	// 1. Admin can list notifications
	reqList, _ := http.NewRequest("GET", "/api/notifications?status=FAILED", nil)
	reqList.Header.Set("Authorization", "Bearer "+adminToken)
	wList := httptest.NewRecorder()
	router.ServeHTTP(wList, reqList)

	if wList.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK for list notifications, got %d: %s", wList.Code, wList.Body.String())
	}

	var listData struct {
		Data struct {
			Items []model.Notification `json:"items"`
			Total int64                `json:"total"`
		} `json:"data"`
	}
	_ = json.Unmarshal(wList.Body.Bytes(), &listData)
	if listData.Data.Total != 1 {
		t.Errorf("Expected 1 failed notification, got %d", listData.Data.Total)
	}

	// 2. Real engine metrics check
	reqMetrics, _ := http.NewRequest("GET", "/api/notifications/metrics", nil)
	reqMetrics.Header.Set("Authorization", "Bearer "+adminToken)
	wMetrics := httptest.NewRecorder()
	router.ServeHTTP(wMetrics, reqMetrics)

	if wMetrics.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK for metrics, got %d", wMetrics.Code)
	}

	// 3. RBAC Enforcement: operator cannot retry notification (requires notification.manage which operator does not have)
	reqRetryOperator, _ := http.NewRequest("POST", fmt.Sprintf("/api/notifications/%d/retry", notifFailed.ID), nil)
	reqRetryOperator.Header.Set("Authorization", "Bearer "+operatorToken)
	wRetryOperator := httptest.NewRecorder()
	router.ServeHTTP(wRetryOperator, reqRetryOperator)

	if wRetryOperator.Code != http.StatusForbidden {
		t.Errorf("Expected 403 Forbidden for operator retrying notification without notification.manage, got %d", wRetryOperator.Code)
	}

	// 4. Admin CAN retry notification
	reqRetryAdmin, _ := http.NewRequest("POST", fmt.Sprintf("/api/notifications/%d/retry", notifFailed.ID), nil)
	reqRetryAdmin.Header.Set("Authorization", "Bearer "+adminToken)
	wRetryAdmin := httptest.NewRecorder()
	router.ServeHTTP(wRetryAdmin, reqRetryAdmin)

	if wRetryAdmin.Code != http.StatusOK {
		t.Errorf("Expected 200 OK for admin retrying notification, got %d (body: %s)", wRetryAdmin.Code, wRetryAdmin.Body.String())
	}

	// Check that notification reset to PENDING
	var refreshedNotif model.Notification
	db.First(&refreshedNotif, notifFailed.ID)
	if refreshedNotif.Status != model.NotificationPending {
		t.Errorf("Expected status PENDING after manual retry, got %s", refreshedNotif.Status)
	}
	if refreshedNotif.Attempts != 0 {
		t.Errorf("Expected attempts reset to 0, got %d", refreshedNotif.Attempts)
	}

	// 5. Channel status inspection: Email with empty config must be marked not configured
	reqChannels, _ := http.NewRequest("GET", "/api/notifications/channels", nil)
	reqChannels.Header.Set("Authorization", "Bearer "+adminToken)
	wChannels := httptest.NewRecorder()
	router.ServeHTTP(wChannels, reqChannels)

	var chanResp struct {
		Data []struct {
			Type         string `json:"type"`
			IsConfigured bool   `json:"is_configured"`
		} `json:"data"`
	}
	_ = json.Unmarshal(wChannels.Body.Bytes(), &chanResp)
	if len(chanResp.Data) > 0 {
		var emailConfigured bool
		for _, ch := range chanResp.Data {
			if ch.Type == "EMAIL" {
				emailConfigured = ch.IsConfigured
				break
			}
		}
		if emailConfigured {
			t.Errorf("Expected unconfigured SMTP channel to have is_configured=false")
		}
	}
}
