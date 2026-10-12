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
			notifGroup.GET("/metrics", middleware.RequirePermission("notification.view"), notifHandler.GetMetrics)
			notifGroup.GET("/:id", middleware.RequirePermission("notification.view"), notifHandler.GetNotification)
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
	db.Where("name = ?", "Operator").First(&opRole)
	db.Where("name = ?", "Engineer").First(&engRole)

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
