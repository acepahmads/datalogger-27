package handler_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"datalogger/internal/config"
	"datalogger/internal/database"
	"datalogger/internal/handler"
	"datalogger/internal/middleware"
	"datalogger/internal/model"
	"datalogger/internal/queue"
	"datalogger/internal/repository"
	"datalogger/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func setupTestBackupRouter(t *testing.T) (*gin.Engine, *gorm.DB, string, string, func()) {
	gin.SetMode(gin.TestMode)

	baseTmp, err := os.MkdirTemp("", "api_backup_test_*")
	if err != nil {
		t.Fatalf("failed creating temp dir: %v", err)
	}

	backupDir := filepath.Join(baseTmp, "backups")
	walDir := filepath.Join(baseTmp, "wal_queue")
	_ = os.MkdirAll(backupDir, 0755)
	_ = os.MkdirAll(walDir, 0755)

	dsn := fmt.Sprintf("file:mem_api_backup_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		t.Fatalf("failed opening test DB: %v", err)
	}

	// Auto migrate
	_ = database.RunMigrations(db)
	_ = database.Seed(db)

	cfg := config.DefaultConfig()
	cfg.BackupDir = backupDir
	cfg.BackupMinFreeSpaceMB = 1

	walCfg := queue.DefaultWALConfig(walDir)
	walQueue, _ := queue.OpenWALQueue(walCfg)
	telemetryRepo := repository.NewTelemetryRepository(db)
	walAdapter := service.NewPersistentQueueAdapter(telemetryRepo, walQueue, 50, 3)

	backupRepo := repository.NewBackupRepository(db)
	dbAdapter := service.NewDatabaseBackupAdapter(db, cfg)
	systemRepo := repository.NewSystemRepository(db)
	backupService := service.NewBackupService(db, cfg, backupRepo, dbAdapter, walAdapter, nil, systemRepo)
	backupScheduler := service.NewBackupScheduler(backupService, cfg)

	authService := service.NewAuthService(systemRepo, cfg)
	backupHandler := handler.NewBackupHandler(backupService, backupScheduler)

	r := gin.New()
	r.Use(gin.Recovery())

	api := r.Group("/api")
	{
		backupGroup := api.Group("/backups")
		backupGroup.Use(middleware.JWTAuth(authService))
		{
			backupGroup.GET("", middleware.RequirePermission("backup.view"), backupHandler.ListBackups)
			backupGroup.GET("/status", middleware.RequirePermission("backup.view"), backupHandler.GetStatus)
			backupGroup.GET("/schedule", middleware.RequirePermission("backup.view"), backupHandler.GetSchedule)
			backupGroup.PUT("/schedule", middleware.RequirePermission("backup.manage"), backupHandler.UpdateSchedule)
			backupGroup.GET("/jobs/:jobId", middleware.RequirePermission("backup.view"), backupHandler.GetJob)
			backupGroup.GET("/:id", middleware.RequirePermission("backup.view"), backupHandler.GetBackup)
			backupGroup.GET("/:id/restore-preview", middleware.RequirePermission("backup.view"), backupHandler.GetRestorePreview)
			backupGroup.GET("/:id/download", middleware.RequirePermission("backup.view"), backupHandler.DownloadBackup)
			backupGroup.POST("", middleware.RequirePermission("backup.create"), backupHandler.CreateBackup)
			backupGroup.POST("/:id/validate", middleware.RequirePermission("backup.validate"), backupHandler.ValidateBackup)
			backupGroup.POST("/:id/restore", middleware.RequirePermission("backup.restore"), backupHandler.RestoreBackup)
			backupGroup.DELETE("/:id", middleware.RequirePermission("backup.manage"), backupHandler.DeleteBackup)
		}
	}

	// Create operator user with Operator role (which does not have backup.create or backup.restore)
	var opRole model.Role
	db.Where("name = ?", "Operator").First(&opRole)
	opHash, _ := bcrypt.GenerateFromPassword([]byte("operator123"), bcrypt.DefaultCost)
	db.Create(&model.User{
		Username: "operator",
		Email:    "operator@datalogger.local",
		Password: string(opHash),
		FullName: "Standard Operator",
		RoleID:   opRole.ID,
		IsActive: true,
	})

	// Login as admin and operator to obtain JWT tokens
	adminToken, _, err := authService.Login("admin", "admin123", "127.0.0.1", "test-agent")
	if err != nil {
		t.Fatalf("failed login as admin: %v", err)
	}

	operatorToken, _, err := authService.Login("operator", "operator123", "127.0.0.1", "test-agent")
	if err != nil {
		t.Fatalf("failed login as operator: %v", err)
	}

	cleanup := func() {
		walAdapter.Close()
		_ = walQueue.Close()
		_ = os.RemoveAll(baseTmp)
	}

	return r, db, adminToken, operatorToken, cleanup
}

func TestPhase43_API_AuthenticationRequired(t *testing.T) {
	r, _, _, _, cleanup := setupTestBackupRouter(t)
	defer cleanup()

	// 1. Unauthenticated request to /api/backups must return 401
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/backups", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401 Unauthorized, got %d", w.Code)
	}
}

func TestPhase43_API_RBACEnforcement(t *testing.T) {
	r, _, adminToken, operatorToken, cleanup := setupTestBackupRouter(t)
	defer cleanup()

	// 1. Operator attempts to trigger backup (requires backup.create) -> should return 403 Forbidden
	body, _ := json.Marshal(map[string]interface{}{"type": "MANUAL"})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/backups", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+operatorToken)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected status 403 Forbidden for operator without backup.create, got %d", w.Code)
	}

	// 2. Administrator triggers backup (has backup.create) -> should succeed with 200 OK
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("POST", "/api/backups", bytes.NewBuffer(body))
	req2.Header.Set("Authorization", "Bearer "+adminToken)
	req2.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Errorf("expected status 200 OK for admin, got %d (body: %s)", w2.Code, w2.Body.String())
	}
}

func TestPhase43_API_CatalogAndStatus(t *testing.T) {
	r, _, adminToken, _, cleanup := setupTestBackupRouter(t)
	defer cleanup()

	// 1. GET /api/backups
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/backups?page=1&page_size=10", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200 OK, got %d", w.Code)
	}

	// 2. GET /api/backups/status
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/api/backups/status", nil)
	req2.Header.Set("Authorization", "Bearer "+adminToken)
	r.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Errorf("expected status 200 OK, got %d", w2.Code)
	}

	// 3. GET /api/backups/schedule
	w3 := httptest.NewRecorder()
	req3, _ := http.NewRequest("GET", "/api/backups/schedule", nil)
	req3.Header.Set("Authorization", "Bearer "+adminToken)
	r.ServeHTTP(w3, req3)

	if w3.Code != http.StatusOK {
		t.Errorf("expected status 200 OK, got %d", w3.Code)
	}
}

func TestPhase43_API_RestoreRequiresConfirmation(t *testing.T) {
	r, _, adminToken, _, cleanup := setupTestBackupRouter(t)
	defer cleanup()

	// Attempting restore without confirm: true must return 400 Bad Request
	body, _ := json.Marshal(map[string]interface{}{"confirm": false})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/backups/fake_backup_id/restore", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+adminToken)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 Bad Request without confirmation, got %d", w.Code)
	}
}
