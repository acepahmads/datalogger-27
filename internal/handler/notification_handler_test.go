package handler_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"datalogger/internal/config"
	"datalogger/internal/database"
	"datalogger/internal/handler"
	"datalogger/internal/middleware"
	"datalogger/internal/model"
	"datalogger/internal/repository"
	"datalogger/internal/service"
	"datalogger/internal/service/notification"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func setupTestNotificationRouter(t *testing.T) (*gin.Engine, *gorm.DB, string, string, string, *repository.NotificationRepository, *service.NotificationEngine) {
	gin.SetMode(gin.TestMode)

	dsn := fmt.Sprintf("file:mem_notif_handler_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		t.Fatalf("failed opening test DB: %v", err)
	}

	_ = database.RunMigrations(db)
	_ = database.Seed(db)

	cfg := config.DefaultConfig()
	systemRepo := repository.NewSystemRepository(db)
	alarmRepo := repository.NewAlarmRepository(db)
	notifRepo := repository.NewNotificationRepository(db)

	notifEngine := service.NewNotificationEngine(notifRepo, alarmRepo, systemRepo, service.DefaultNotificationEngineConfig())
	notifHandler := handler.NewNotificationHandler(notifEngine, notifRepo)
	authService := service.NewAuthService(systemRepo, cfg)

	r := gin.New()
	r.Use(gin.Recovery())

	api := r.Group("/api")
	{
		notifGroup := api.Group("/notifications")
		notifGroup.Use(middleware.JWTAuth(authService))
		{
			notifGroup.GET("", middleware.RequirePermission("notification.view"), notifHandler.ListNotifications)
			notifGroup.DELETE("", middleware.RequirePermission("notification.manage"), notifHandler.PurgeNotifications)
			notifGroup.POST("/purge", middleware.RequirePermission("notification.manage"), notifHandler.PurgeNotifications)
			notifGroup.GET("/metrics", middleware.RequirePermission("notification.view"), notifHandler.GetMetrics)
			notifGroup.GET("/:id", middleware.RequirePermission("notification.view"), notifHandler.GetNotification)
			notifGroup.DELETE("/:id", middleware.RequirePermission("notification.manage"), notifHandler.DeleteNotification)
			notifGroup.POST("/:id/retry", middleware.RequirePermission("notification.manage"), notifHandler.RetryNotification)
			notifGroup.POST("/:id/cancel", middleware.RequirePermission("notification.manage"), notifHandler.CancelNotification)
			notifGroup.GET("/channels", middleware.RequirePermission("notification.view"), notifHandler.ListChannels)
			notifGroup.GET("/channels/:id", middleware.RequirePermission("notification.view"), notifHandler.GetChannel)
			notifGroup.PUT("/channels/:id", middleware.RequirePermission("notification.manage"), notifHandler.UpdateChannel)
			notifGroup.POST("/channels/:id/test", middleware.RequirePermission("notification.manage"), notifHandler.TestSend)
		}
	}

	// Create operator and engineer test users
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

	return r, db, adminToken, engineerToken, operatorToken, notifRepo, notifEngine
}

func TestNotificationHandler_RBACAndEndpoints(t *testing.T) {
	r, db, adminToken, engineerToken, operatorToken, _, notifEngine := setupTestNotificationRouter(t)
	defer notifEngine.Stop()

	// 1. GET /api/notifications unauthenticated -> 401
	req, _ := http.NewRequest(http.MethodGet, "/api/notifications", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 Unauthorized without token, got %d", w.Code)
	}

	// 2. GET /api/notifications with Operator Token -> 200 OK (has notification.view)
	req, _ = http.NewRequest(http.MethodGet, "/api/notifications", nil)
	req.Header.Set("Authorization", "Bearer "+operatorToken)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 OK for operator viewing notifications, got %d", w.Code)
	}

	// 3. Seed a notification and test Operator vs Engineer on POST /api/notifications/:id/retry
	n := &model.Notification{
		AlarmID:        1,
		Channel:        string(model.ChannelWeb),
		Recipient:      "Web Client",
		Subject:        "RBAC Test Alert",
		Message:        "Checking permission enforcement",
		Status:         model.NotificationFailed,
		IdempotencyKey: "test:rbac:1",
	}
	_ = db.Create(n).Error

	// Operator cannot retry (missing notification.manage) -> 403 Forbidden
	req, _ = http.NewRequest(http.MethodPost, fmt.Sprintf("/api/notifications/%d/retry", n.ID), nil)
	req.Header.Set("Authorization", "Bearer "+operatorToken)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403 Forbidden for operator retrying notification, got %d", w.Code)
	}

	// Engineer can retry -> 200 OK
	req, _ = http.NewRequest(http.MethodPost, fmt.Sprintf("/api/notifications/%d/retry", n.ID), nil)
	req.Header.Set("Authorization", "Bearer "+engineerToken)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 OK for engineer retrying notification, got %d", w.Code)
	}

	// Admin can cancel -> 200 OK
	req, _ = http.NewRequest(http.MethodPost, fmt.Sprintf("/api/notifications/%d/cancel", n.ID), nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 OK for admin cancelling notification, got %d", w.Code)
	}
}

func TestNotificationHandler_ChannelsAndRedaction(t *testing.T) {
	r, db, adminToken, _, _, _, notifEngine := setupTestNotificationRouter(t)
	defer notifEngine.Stop()

	// Seed an email channel with secret password
	emailChannel := &model.NotificationChannel{
		Name:            "Corporate SMTP",
		Type:            model.ChannelEmail,
		Enabled:         true,
		ConfigJSON:      `{"smtp_host":"mail.corp.com","smtp_port":587,"username":"alert@corp.com","password":"super-confidential-password","from_email":"alerts@corp.com"}`,
		MinSeverity:     model.SeverityWarning,
		RateLimitPerMin: 15,
		MaxRetries:      3,
	}
	_ = db.Create(emailChannel).Error

	// 1. GET /api/notifications/channels -> verify password is masked as "********"
	req, _ := http.NewRequest(http.MethodGet, "/api/notifications/channels", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK listing channels, got %d", w.Code)
	}

	bodyStr := w.Body.String()
	if strings.Contains(bodyStr, "super-confidential-password") {
		t.Fatalf("CRITICAL SECURITY VULNERABILITY: Stored password exposed in channel list response: %s", bodyStr)
	}
	if !strings.Contains(bodyStr, notification.MaskedSecret) {
		t.Fatalf("expected password to be masked with '********', response was: %s", bodyStr)
	}

	// 2. PUT /api/notifications/channels/:id updating name but sending "********" for password
	updatePayload := map[string]interface{}{
		"name": "Updated Corporate SMTP",
		"config": map[string]interface{}{
			"smtp_host":  "mail.corp.com",
			"smtp_port":  587,
			"username":   "alert@corp.com",
			"password":   "********", // Retained secret
			"from_email": "alerts@corp.com",
		},
		"rate_limit_per_min": 25,
	}
	payloadBytes, _ := json.Marshal(updatePayload)

	req, _ = http.NewRequest(http.MethodPut, fmt.Sprintf("/api/notifications/channels/%d", emailChannel.ID), bytes.NewBuffer(payloadBytes))
	req.Header.Set("Authorization", "Bearer "+adminToken)
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK updating channel, got %d: %s", w.Code, w.Body.String())
	}

	// Verify in database that original password was NOT overwritten with "********"
	var updatedChannel model.NotificationChannel
	db.First(&updatedChannel, emailChannel.ID)
	if !strings.Contains(updatedChannel.ConfigJSON, "super-confidential-password") {
		t.Fatalf("UpdateChannel wiped original password when receiving masked placeholder! Got: %s", updatedChannel.ConfigJSON)
	}
	if updatedChannel.Name != "Updated Corporate SMTP" {
		t.Errorf("channel name was not updated: %s", updatedChannel.Name)
	}
	if updatedChannel.RateLimitPerMin != 25 {
		t.Errorf("rate limit was not updated: %d", updatedChannel.RateLimitPerMin)
	}

	// 3. PUT /api/notifications/channels/:id with invalid configuration (empty smtp_host) -> 400 Bad Request
	badPayload := map[string]interface{}{
		"config": map[string]interface{}{
			"smtp_host": "", // Invalid!
		},
	}
	badBytes, _ := json.Marshal(badPayload)

	req, _ = http.NewRequest(http.MethodPut, fmt.Sprintf("/api/notifications/channels/%d", emailChannel.ID), bytes.NewBuffer(badBytes))
	req.Header.Set("Authorization", "Bearer "+adminToken)
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request on invalid configuration, got %d", w.Code)
	}

	// 4. POST /api/notifications/channels/:id/test -> runs test send
	req, _ = http.NewRequest(http.MethodPost, fmt.Sprintf("/api/notifications/channels/%d/test", emailChannel.ID), bytes.NewBufferString(`{"recipient":"test@corp.com"}`))
	req.Header.Set("Authorization", "Bearer "+adminToken)
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// In test environment with unroutable mail.corp.com, this should return 422 Unprocessable Entity with diagnostic details
	if w.Code != http.StatusUnprocessableEntity && w.Code != http.StatusOK {
		t.Errorf("expected 422 or 200 for test-send diagnostic, got %d: %s", w.Code, w.Body.String())
	}
}

func TestNotificationHandler_ToggleActiveInactiveUnconfiguredChannel(t *testing.T) {
	r, db, adminToken, _, _, _, notifEngine := setupTestNotificationRouter(t)
	defer notifEngine.Stop()

	// Seed an unconfigured Telegram channel that is initially Enabled=true (like legacy DB rows)
	tgChannel := &model.NotificationChannel{
		Name:            "Engineering Telegram Bot",
		Type:            model.ChannelTelegram,
		Enabled:         true,
		ConfigJSON:      `{"bot_token":"","chat_id":""}`,
		MinSeverity:     model.SeverityWarning,
		RateLimitPerMin: 10,
		MaxRetries:      3,
	}
	if err := db.Create(tgChannel).Error; err != nil {
		t.Fatalf("failed seeding tgChannel: %v", err)
	}

	// Seed aRETRYING notification queued for this channel
	chID := tgChannel.ID
	queuedNotif := &model.Notification{
		AlarmID:   1,
		ChannelID: &chID,
		Channel:   string(model.ChannelTelegram),
		Recipient: "Engineering Telegram Bot",
		Message:   "Test alarm",
		Status:    model.NotificationRetrying,
	}
	_ = db.Create(queuedNotif).Error

	// 1. Disabling (enabled=false) an unconfigured channel MUST succeed (200 OK) and cancel pending/retrying items
	req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("/api/notifications/channels/%d", tgChannel.ID), bytes.NewBufferString(`{"enabled":false}`))
	req.Header.Set("Authorization", "Bearer "+adminToken)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK when disabling unconfigured channel, got %d: %s", w.Code, w.Body.String())
	}

	var dbChan model.NotificationChannel
	db.First(&dbChan, tgChannel.ID)
	if dbChan.Enabled {
		t.Fatalf("expected channel Enabled=false in DB after disabling, got Enabled=true")
	}

	var dbNotif model.Notification
	db.First(&dbNotif, queuedNotif.ID)
	if dbNotif.Status != model.NotificationCancelled {
		t.Errorf("expected queued retrying notification to be CANCELLED upon disabling channel, got %s", dbNotif.Status)
	}

	// 2. Trying to re-enable (enabled=true) while still unconfigured MUST fail with 400 Bad Request
	req, _ = http.NewRequest(http.MethodPut, fmt.Sprintf("/api/notifications/channels/%d", tgChannel.ID), bytes.NewBufferString(`{"enabled":true}`))
	req.Header.Set("Authorization", "Bearer "+adminToken)
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request when enabling unconfigured channel without bot_token, got %d", w.Code)
	}

	// 3. Configuring bot_token & chat_id + enabled=true MUST succeed (200 OK)
	req, _ = http.NewRequest(
		http.MethodPut,
		fmt.Sprintf("/api/notifications/channels/%d", tgChannel.ID),
		bytes.NewBufferString(`{"enabled":true,"config":{"bot_token":"123456:VALID_TOKEN","chat_id":"-100999"}}`),
	)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK when configuring and enabling channel, got %d: %s", w.Code, w.Body.String())
	}

	db.First(&dbChan, tgChannel.ID)
	if !dbChan.Enabled {
		t.Errorf("expected channel Enabled=true in DB after saving valid config")
	}
}

func TestNotificationHandler_DeleteAndDeleteAllRBAC(t *testing.T) {
	r, db, superadminToken, engineerToken, operatorToken, _, notifEngine := setupTestNotificationRouter(t)
	defer notifEngine.Stop()

	// Create an explicit ADMIN role user as well (admin seed user is Superadmin)
	var adminRole model.Role
	if err := db.Where("UPPER(name) = ?", "ADMIN").First(&adminRole).Error; err != nil || adminRole.ID == 0 {
		adminRole = model.Role{Name: "Admin", Description: "Admin role"}
		db.Create(&adminRole)
	}
	adminHash, _ := bcrypt.GenerateFromPassword([]byte("adminrole123"), bcrypt.DefaultCost)
	db.Create(&model.User{
		Username: "admin_only_test",
		Email:    "admin_only@test.local",
		Password: string(adminHash),
		FullName: "Admin Role User",
		RoleID:   adminRole.ID,
		IsActive: true,
	})
	authSvc := service.NewAuthService(repository.NewSystemRepository(db), config.DefaultConfig())
	adminRoleToken, _, errLogin := authSvc.Login("admin_only_test", "adminrole123", "127.0.0.1", "TestClient")
	if errLogin != nil {
		t.Fatalf("failed login as admin_only_test: %v", errLogin)
	}

	// Seed 3 notifications across statuses
	alarm := model.Alarm{
		AlarmCode:   "ALM-DEL-001",
		Severity:    "CRITICAL",
		Message:     "Queue delete test alarm",
		Status:      "ACTIVE",
		TriggeredAt: time.Now().UTC(),
	}
	db.Create(&alarm)

	now := time.Now().UTC()
	n1 := model.Notification{
		AlarmID:     alarm.ID,
		Channel:     "EMAIL",
		Recipient:   "ops@plant.local",
		Subject:     "CRITICAL Alarm",
		Message:     "Queue item 1",
		Status:      model.NotificationFailed,
		Attempts:    1,
		MaxAttempts: 3,
		ErrorMsg:    "smtp error",
	}
	n2 := model.Notification{
		AlarmID:     alarm.ID,
		Channel:     "TELEGRAM",
		Recipient:   "-100123",
		Subject:     "CRITICAL Alarm",
		Message:     "Queue item 2",
		Status:      model.NotificationFailed,
		Attempts:    1,
		MaxAttempts: 3,
		ErrorMsg:    "missing bot token",
	}
	n3 := model.Notification{
		AlarmID:     alarm.ID,
		Channel:     "WEB",
		Recipient:   "Web Broadcast",
		Subject:     "CRITICAL Alarm",
		Message:     "Queue item 3",
		Status:      model.NotificationSent,
		Attempts:    1,
		MaxAttempts: 3,
		SentAt:      &now,
	}
	db.Create(&n1)
	db.Create(&n2)
	db.Create(&n3)

	// 1. OPERATOR must get 403 Forbidden on both single Delete and Delete All
	for _, tc := range []struct {
		name  string
		token string
	}{
		{name: "OPERATOR", token: operatorToken},
		{name: "ENGINEER", token: engineerToken},
	} {
		req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/api/notifications/%d", n1.ID), nil)
		req.Header.Set("Authorization", "Bearer "+tc.token)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusForbidden {
			t.Fatalf("expected 403 Forbidden for %s single delete, got %d: %s", tc.name, w.Code, w.Body.String())
		}

		reqPurge, _ := http.NewRequest(http.MethodPost, "/api/notifications/purge", bytes.NewBufferString(`{}`))
		reqPurge.Header.Set("Authorization", "Bearer "+tc.token)
		reqPurge.Header.Set("Content-Type", "application/json")
		wPurge := httptest.NewRecorder()
		r.ServeHTTP(wPurge, reqPurge)
		if wPurge.Code != http.StatusForbidden {
			t.Fatalf("expected 403 Forbidden for %s delete-all purge, got %d: %s", tc.name, wPurge.Code, wPurge.Body.String())
		}
	}

	// 2. ADMIN role can delete a single notification (200 OK)
	reqDel, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/api/notifications/%d", n1.ID), nil)
	reqDel.Header.Set("Authorization", "Bearer "+adminRoleToken)
	wDel := httptest.NewRecorder()
	r.ServeHTTP(wDel, reqDel)
	if wDel.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for ADMIN single delete, got %d: %s", wDel.Code, wDel.Body.String())
	}

	// Deleting same ID again returns 404 Not Found
	reqDelAgain, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/api/notifications/%d", n1.ID), nil)
	reqDelAgain.Header.Set("Authorization", "Bearer "+adminRoleToken)
	wDelAgain := httptest.NewRecorder()
	r.ServeHTTP(wDelAgain, reqDelAgain)
	if wDelAgain.Code != http.StatusNotFound {
		t.Errorf("expected 404 Not Found deleting already-deleted notification, got %d", wDelAgain.Code)
	}

	// Check live metrics reflect 2 remaining rows (1 SENT, 1 FAILED)
	reqMet, _ := http.NewRequest(http.MethodGet, "/api/notifications/metrics", nil)
	reqMet.Header.Set("Authorization", "Bearer "+superadminToken)
	wMet := httptest.NewRecorder()
	r.ServeHTTP(wMet, reqMet)
	if wMet.Code != http.StatusOK {
		t.Fatalf("expected 200 OK on metrics, got %d", wMet.Code)
	}
	var metResp struct {
		Data map[string]int64 `json:"data"`
	}
	_ = json.Unmarshal(wMet.Body.Bytes(), &metResp)
	if metResp.Data["dispatched_count"] != 2 || metResp.Data["delivered_count"] != 1 || metResp.Data["failed_count"] != 1 {
		t.Errorf("unexpected live metrics after single delete: %+v", metResp.Data)
	}

	// 3. SUPERADMIN can Delete All via POST /api/notifications/purge
	reqAll, _ := http.NewRequest(http.MethodPost, "/api/notifications/purge", bytes.NewBufferString(`{}`))
	reqAll.Header.Set("Authorization", "Bearer "+superadminToken)
	reqAll.Header.Set("Content-Type", "application/json")
	wAll := httptest.NewRecorder()
	r.ServeHTTP(wAll, reqAll)
	if wAll.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for SUPERADMIN delete-all purge, got %d: %s", wAll.Code, wAll.Body.String())
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

	var remaining int64
	db.Model(&model.Notification{}).Count(&remaining)
	if remaining != 0 {
		t.Errorf("expected 0 notifications remaining after Delete All, got %d", remaining)
	}
}
