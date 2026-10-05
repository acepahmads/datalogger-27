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
	"datalogger/internal/websocket"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func setupTestRouter(t *testing.T) (*gin.Engine, *gorm.DB) {
	gin.SetMode(gin.TestMode)

	dsn := fmt.Sprintf("file:memapi_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	_ = db.AutoMigrate(
		&model.User{},
		&model.Role{},
		&model.Permission{},
		&model.DevelopmentPhase{},
		&model.DevelopmentSubphase{},
		&model.DevelopmentTask{},
		&model.DevelopmentTaskLog{},
		&model.DevelopmentDependency{},
		&model.DevelopmentEvidence{},
		&model.DeviceType{},
		&model.Device{},
		&model.DeviceConnection{},
		&model.Parameter{},
		&model.Sensor{},
		&model.RawData{},
		&model.ProcessedData{},
		&model.AggregatedData{},
		&model.Alarm{},
		&model.NotificationChannel{},
		&model.Notification{},
		&model.AuditTrail{},
		&model.SystemLog{},
		&model.CommunicationLog{},
		&model.OutputDestination{},
		&model.DeliveryLog{},
		&model.SystemHealth{},
	)

	// Seed test data
	_ = database.Seed(db)

	cfg := config.DefaultConfig()
	phaseRepo := repository.NewPhaseRepository(db)
	systemRepo := repository.NewSystemRepository(db)

	authService := service.NewAuthService(systemRepo, cfg)
	phaseService := service.NewPhaseService(phaseRepo, systemRepo)
	systemService := service.NewSystemService(systemRepo, phaseRepo, cfg)
	hub := websocket.NewHub()

	authHandler := handler.NewAuthHandler(authService)
	devHandler := handler.NewDevHandler(phaseService, hub)
	systemHandler := handler.NewSystemHandler(systemService)

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.CORS())

	api := r.Group("/api")
	{
		api.POST("/auth/login", authHandler.Login)
		api.GET("/phases", devHandler.GetPhases)
		api.GET("/phases/:id", devHandler.GetPhaseByID)
		api.GET("/progress", devHandler.GetProgress)
		api.GET("/tasks", devHandler.GetTasks)
		api.POST("/tasks", devHandler.CreateTask)
		api.GET("/tasks/:id", devHandler.GetTaskByID)
		api.PUT("/tasks/:id", devHandler.UpdateTask)
		api.DELETE("/tasks/:id", devHandler.DeleteTask)
		api.GET("/activity", devHandler.GetActivity)
		api.GET("/system/status", systemHandler.GetStatus)
		api.GET("/system/health", systemHandler.GetHealth)
		api.GET("/system/resources", systemHandler.GetResources)
		api.GET("/devices", systemHandler.GetDevices)
		api.GET("/alarms", systemHandler.GetAlarms)
		api.GET("/logs", systemHandler.GetLogs)
	}

	return r, db
}

func TestAPIGetProgress(t *testing.T) {
	r, _ := setupTestRouter(t)

	req, _ := http.NewRequest("GET", "/api/progress", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", w.Code)
	}

	var res struct {
		Success bool `json:"success"`
		Data    struct {
			OverallPercentage float64 `json:"overall_percentage"`
			CurrentPhase      string  `json:"current_phase"`
			TotalTasksCount   int     `json:"total_tasks_count"`
		} `json:"data"`
	}

	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if !res.Success {
		t.Errorf("Expected success: true")
	}
	if res.Data.TotalTasksCount == 0 {
		t.Errorf("Expected total tasks > 0, got %d", res.Data.TotalTasksCount)
	}
}

func TestAPICreateAndUpdateTask(t *testing.T) {
	r, _ := setupTestRouter(t)

	// Create task
	newTask := map[string]interface{}{
		"phase_id":    1,
		"task_name":   "New Integration Test Task",
		"description": "Integration test task description",
		"status":      "WORKING",
		"progress":    50,
		"priority":    "HIGH",
		"owner":       "QA Engineer",
	}
	payload, _ := json.Marshal(newTask)

	req, _ := http.NewRequest("POST", "/api/tasks", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("Expected status 201, got %d (body: %s)", w.Code, w.Body.String())
	}

	var createRes struct {
		Success bool `json:"success"`
		Data    struct {
			ID       uint   `json:"id"`
			TaskName string `json:"task_name"`
			Status   string `json:"status"`
		} `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &createRes)

	if createRes.Data.ID == 0 {
		t.Fatalf("Expected valid task ID, got 0")
	}

	// Update task to DONE
	updatePayload, _ := json.Marshal(map[string]interface{}{
		"status":      "DONE",
		"progress":    100,
		"test_result": "All integration checks passed",
	})
	reqUpdate, _ := http.NewRequest("PUT", fmt.Sprintf("/api/tasks/%d", createRes.Data.ID), bytes.NewBuffer(updatePayload))
	reqUpdate.Header.Set("Content-Type", "application/json")
	wUpdate := httptest.NewRecorder()
	r.ServeHTTP(wUpdate, reqUpdate)

	if wUpdate.Code != http.StatusOK {
		t.Fatalf("Expected status 200 on update, got %d", wUpdate.Code)
	}
}

func TestAPISystemStatus(t *testing.T) {
	r, _ := setupTestRouter(t)

	req, _ := http.NewRequest("GET", "/api/system/status", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", w.Code)
	}

	var res struct {
		Success bool `json:"success"`
		Data    struct {
			ServiceStatus  string `json:"service_status"`
			DatabaseStatus string `json:"database_status"`
			TotalDevices   int    `json:"total_devices"`
		} `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &res)

	if res.Data.ServiceStatus != "RUNNING" {
		t.Errorf("Expected ServiceStatus RUNNING, got %s", res.Data.ServiceStatus)
	}
	if res.Data.TotalDevices != 3 {
		t.Errorf("Expected 3 seeded devices, got %d", res.Data.TotalDevices)
	}
}

func TestAPIAuthLogin(t *testing.T) {
	r, _ := setupTestRouter(t)

	// 1. Missing credentials -> 400
	reqMissing, _ := http.NewRequest("POST", "/api/auth/login", bytes.NewBuffer([]byte("{}")))
	reqMissing.Header.Set("Content-Type", "application/json")
	wMissing := httptest.NewRecorder()
	r.ServeHTTP(wMissing, reqMissing)
	if wMissing.Code != http.StatusBadRequest {
		t.Errorf("Expected 400 for missing credentials, got %d", wMissing.Code)
	}

	// 2. Invalid credentials -> 401
	invalidCreds, _ := json.Marshal(map[string]string{"username": "admin", "password": "wrongpassword"})
	reqInvalid, _ := http.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(invalidCreds))
	reqInvalid.Header.Set("Content-Type", "application/json")
	wInvalid := httptest.NewRecorder()
	r.ServeHTTP(wInvalid, reqInvalid)
	if wInvalid.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401 for wrong credentials, got %d", wInvalid.Code)
	}

	// 3. Valid credentials -> 200 with JWT
	validCreds, _ := json.Marshal(map[string]string{"username": "admin", "password": "admin123"})
	reqValid, _ := http.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(validCreds))
	reqValid.Header.Set("Content-Type", "application/json")
	wValid := httptest.NewRecorder()
	r.ServeHTTP(wValid, reqValid)
	if wValid.Code != http.StatusOK {
		t.Fatalf("Expected 200 for valid login, got %d (body: %s)", wValid.Code, wValid.Body.String())
	}

	var loginRes struct {
		Success bool `json:"success"`
		Data    struct {
			Token string `json:"token"`
			User  struct {
				Username string `json:"username"`
				Role     string `json:"role"`
			} `json:"user"`
		} `json:"data"`
	}
	_ = json.Unmarshal(wValid.Body.Bytes(), &loginRes)
	if loginRes.Data.Token == "" {
		t.Fatalf("Expected non-empty JWT token")
	}
	if loginRes.Data.User.Username != "admin" {
		t.Errorf("Expected username admin, got %s", loginRes.Data.User.Username)
	}
}
