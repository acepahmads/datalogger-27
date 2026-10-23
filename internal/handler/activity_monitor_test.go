package handler_test

import (
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
	"datalogger/internal/config"
	"datalogger/internal/database"
	"datalogger/internal/handler"
	"datalogger/internal/middleware"
	"datalogger/internal/model"
	"datalogger/internal/repository"
	"datalogger/internal/service"
)

func TestActivityMonitorHandler_RBACAndEndpoints(t *testing.T) {
	gin.SetMode(gin.TestMode)

	dsn := fmt.Sprintf("file:mem_act_api_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	_ = database.RunMigrations(db)

	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	var saRole, adminRole, engRole, opRole model.Role
	db.Where("name = ?", "SUPERADMIN").First(&saRole)
	db.Where("name = ?", "ADMIN").First(&adminRole)
	db.Where("name = ?", "ENGINEER").First(&engRole)
	db.Where("name = ?", "OPERATOR").First(&opRole)

	saUser := model.User{Username: "act_sa_user", Email: "sa@act.local", Password: string(hash), RoleID: saRole.ID, IsActive: true}
	adminUser := model.User{Username: "act_admin_user", Email: "admin@act.local", Password: string(hash), RoleID: adminRole.ID, IsActive: true}
	engUser := model.User{Username: "act_eng_user", Email: "eng@act.local", Password: string(hash), RoleID: engRole.ID, IsActive: true}
	opUser := model.User{Username: "act_op_user", Email: "op@act.local", Password: string(hash), RoleID: opRole.ID, IsActive: true}
	db.Create(&saUser)
	db.Create(&adminUser)
	db.Create(&engUser)
	db.Create(&opUser)

	cfg := config.DefaultConfig()
	systemRepo := repository.NewSystemRepository(db)
	deviceRepo := repository.NewDeviceRepository(db)
	authService := service.NewAuthService(systemRepo, cfg)
	deviceService := service.NewDeviceService(deviceRepo, systemRepo)

	tracer := communication.NewCommTracer(nil)
	monitor := communication.NewActivityMonitor(nil, systemRepo, tracer)
	activityHandler := handler.NewActivityMonitorHandler(monitor, deviceService, tracer)

	// Populate some mock events
	monitor.RecordEvent(communication.ActivityEvent{
		Severity:    communication.SeverityInfo,
		Category:    communication.CategoryPolling,
		Component:   "PollingEngine",
		EventName:   "POLLING_STARTED",
		Description: "Routine sensor check initiated",
	})
	monitor.RecordEvent(communication.ActivityEvent{
		Severity:    communication.SeverityWarning,
		Category:    communication.CategorySerial,
		Component:   "ConnectionManager",
		EventName:   "SERIAL_PORT_DISCONNECTED",
		Description: "Port COM3 unplugged",
	})

	r := gin.New()
	r.Use(gin.Recovery())

	api := r.Group("/api")
	activityGroup := api.Group("/system/activity-monitor")
	activityGroup.Use(middleware.JWTAuth(authService))
	activityGroup.Use(middleware.RequireRole("SUPERADMIN", "ADMINISTRATOR"))
	{
		activityGroup.GET("/overview", activityHandler.GetOverview)
		activityGroup.GET("/events", activityHandler.GetEvents)
		activityGroup.GET("/polling", activityHandler.GetPollingSnapshots)
		activityGroup.GET("/device/:id", activityHandler.GetDeviceDetail)
		activityGroup.GET("/background-jobs", activityHandler.GetBackgroundJobs)
		activityGroup.POST("/events/clear", activityHandler.ClearEvents)
		activityGroup.GET("/export", activityHandler.ExportEvents)
	}

	saToken, _, _ := authService.Login("act_sa_user", "password123", "127.0.0.1", "")
	adminToken, _, _ := authService.Login("act_admin_user", "password123", "127.0.0.1", "")
	engToken, _, _ := authService.Login("act_eng_user", "password123", "127.0.0.1", "")
	opToken, _, _ := authService.Login("act_op_user", "password123", "127.0.0.1", "")

	// 1. Unauthenticated Request -> 401 Unauthorized
	unauthReq, _ := http.NewRequest("GET", "/api/system/activity-monitor/overview", nil)
	unauthW := httptest.NewRecorder()
	r.ServeHTTP(unauthW, unauthReq)
	if unauthW.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 Unauthorized for unauthenticated request, got %d", unauthW.Code)
	}

	// 2. Operator Request -> 403 Forbidden
	opReq, _ := http.NewRequest("GET", "/api/system/activity-monitor/overview", nil)
	opReq.Header.Set("Authorization", "Bearer "+opToken)
	opW := httptest.NewRecorder()
	r.ServeHTTP(opW, opReq)
	if opW.Code != http.StatusForbidden {
		t.Fatalf("expected 403 Forbidden for OPERATOR, got %d", opW.Code)
	}

	// 3. Engineer Request -> 403 Forbidden
	engReq, _ := http.NewRequest("GET", "/api/system/activity-monitor/events", nil)
	engReq.Header.Set("Authorization", "Bearer "+engToken)
	engW := httptest.NewRecorder()
	r.ServeHTTP(engW, engReq)
	if engW.Code != http.StatusForbidden {
		t.Fatalf("expected 403 Forbidden for ENGINEER, got %d", engW.Code)
	}

	// 4. Admin Request -> 403 Forbidden (SUPERADMIN only)
	adminReq, _ := http.NewRequest("GET", "/api/system/activity-monitor/overview", nil)
	adminReq.Header.Set("Authorization", "Bearer "+adminToken)
	adminW := httptest.NewRecorder()
	r.ServeHTTP(adminW, adminReq)
	if adminW.Code != http.StatusForbidden {
		t.Fatalf("expected 403 Forbidden for ADMIN, got %d", adminW.Code)
	}

	// 5. Superadmin Request -> 200 OK (Overview)
	saOverviewReq, _ := http.NewRequest("GET", "/api/system/activity-monitor/overview", nil)
	saOverviewReq.Header.Set("Authorization", "Bearer "+saToken)
	saOverviewW := httptest.NewRecorder()
	r.ServeHTTP(saOverviewW, saOverviewReq)
	if saOverviewW.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for SUPERADMIN on overview, got %d: %s", saOverviewW.Code, saOverviewW.Body.String())
	}
	var overviewRes struct {
		Success bool `json:"success"`
		Data    struct {
			OverallStatus        string `json:"overall_status"`
			BackendServiceStatus string `json:"backend_service_status"`
		} `json:"data"`
	}
	if err := json.Unmarshal(saOverviewW.Body.Bytes(), &overviewRes); err != nil || !overviewRes.Success {
		t.Errorf("failed parsing overview response: %v", err)
	}

	// 6. Superadmin Request -> 200 OK (Events)
	saEventsReq, _ := http.NewRequest("GET", "/api/system/activity-monitor/events", nil)
	saEventsReq.Header.Set("Authorization", "Bearer "+saToken)
	saEventsW := httptest.NewRecorder()
	r.ServeHTTP(saEventsW, saEventsReq)
	if saEventsW.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for SUPERADMIN on events, got %d", saEventsW.Code)
	}
	var eventsRes struct {
		Data struct {
			Events []communication.ActivityEvent `json:"events"`
			Total  int                           `json:"total"`
		} `json:"data"`
	}
	_ = json.Unmarshal(saEventsW.Body.Bytes(), &eventsRes)
	if eventsRes.Data.Total < 2 {
		t.Errorf("expected at least 2 events, got %d", eventsRes.Data.Total)
	}

	// 7. Superadmin Request -> 200 OK (Background Jobs)
	saJobsReq, _ := http.NewRequest("GET", "/api/system/activity-monitor/background-jobs", nil)
	saJobsReq.Header.Set("Authorization", "Bearer "+saToken)
	saJobsW := httptest.NewRecorder()
	r.ServeHTTP(saJobsW, saJobsReq)
	if saJobsW.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for SUPERADMIN on background-jobs, got %d", saJobsW.Code)
	}

	// 8. Superadmin Export -> 200 OK (CSV)
	saExportReq, _ := http.NewRequest("GET", "/api/system/activity-monitor/export?format=csv", nil)
	saExportReq.Header.Set("Authorization", "Bearer "+saToken)
	saExportW := httptest.NewRecorder()
	r.ServeHTTP(saExportW, saExportReq)
	if saExportW.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for CSV export, got %d", saExportW.Code)
	}
	if saExportW.Header().Get("Content-Type") != "text/csv" {
		t.Errorf("expected text/csv Content-Type, got %s", saExportW.Header().Get("Content-Type"))
	}
}
