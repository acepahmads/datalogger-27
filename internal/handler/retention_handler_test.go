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
	"datalogger/internal/repository"
	"datalogger/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func setupTestRetentionRouter(t *testing.T) (*gin.Engine, *gorm.DB, string, string, string, *service.RetentionService, func()) {
	gin.SetMode(gin.TestMode)

	baseTmp, err := os.MkdirTemp("", "api_retention_test_*")
	if err != nil {
		t.Fatalf("failed creating temp dir: %v", err)
	}

	backupDir := filepath.Join(baseTmp, "backups")
	dataDir := filepath.Join(baseTmp, "data")
	_ = os.MkdirAll(backupDir, 0755)
	_ = os.MkdirAll(dataDir, 0755)

	dsn := fmt.Sprintf("file:mem_api_ret_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		t.Fatalf("failed opening test DB: %v", err)
	}

	// Auto migrate & seed database
	_ = database.RunMigrations(db)
	_ = database.Seed(db)

	cfg := config.DefaultConfig()
	cfg.DataDir = dataDir
	cfg.BackupDir = backupDir

	retentionRepo := repository.NewRetentionRepository(db)
	backupRepo := repository.NewBackupRepository(db)
	systemRepo := repository.NewSystemRepository(db)
	retentionService := service.NewRetentionService(db, cfg, retentionRepo, backupRepo, systemRepo, nil, nil)
	retentionScheduler := service.NewRetentionScheduler(retentionService, cfg)
	retentionHandler := handler.NewRetentionHandler(retentionService, retentionScheduler)

	authService := service.NewAuthService(systemRepo, cfg)

	r := gin.New()
	r.Use(gin.Recovery())

	api := r.Group("/api")
	{
		retentionGroup := api.Group("/retention")
		retentionGroup.Use(middleware.JWTAuth(authService))
		{
			retentionGroup.GET("/policies", middleware.RequirePermission("retention.view"), retentionHandler.ListPolicies)
			retentionGroup.GET("/policies/:id", middleware.RequirePermission("retention.view"), retentionHandler.GetPolicy)
			retentionGroup.PUT("/policies/:id", middleware.RequirePermission("retention.manage"), retentionHandler.UpdatePolicy)
			retentionGroup.PUT("/policies/:id/toggle", middleware.RequirePermission("retention.manage"), retentionHandler.TogglePolicy)
			retentionGroup.POST("/policies/:id/dry-run", middleware.RequirePermission("retention.view"), retentionHandler.DryRun)
			retentionGroup.POST("/policies/:id/execute", middleware.RequirePermission("retention.execute"), retentionHandler.ExecutePolicy)
			retentionGroup.POST("/housekeeping/trigger", middleware.RequirePermission("retention.execute"), retentionHandler.TriggerHousekeeping)
			retentionGroup.POST("/raw-telemetry/purge", middleware.RequireRole("SUPERADMIN"), retentionHandler.PurgeAllRawTelemetry)
			retentionGroup.DELETE("/raw-telemetry", middleware.RequireRole("SUPERADMIN"), retentionHandler.PurgeAllRawTelemetry)
			retentionGroup.GET("/storage", middleware.RequirePermission("retention.view"), retentionHandler.GetStorageOverview)
			retentionGroup.GET("/history", middleware.RequirePermission("retention.view"), retentionHandler.GetHistory)
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

	cleanup := func() {
		_ = os.RemoveAll(baseTmp)
	}

	return r, db, adminToken, engineerToken, operatorToken, retentionService, cleanup
}

func TestRetentionHandlerEndpoints(t *testing.T) {
	r, db, adminToken, engineerToken, operatorToken, svc, cleanup := setupTestRetentionRouter(t)
	defer cleanup()

	// 1. GET /api/retention/policies without auth -> 401 Unauthorized
	req, _ := http.NewRequest(http.MethodGet, "/api/retention/policies", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 Unauthorized without token, got %d", w.Code)
	}

	// 2. GET /api/retention/policies with Engineer token (retention.view) -> 200 OK
	req, _ = http.NewRequest(http.MethodGet, "/api/retention/policies", nil)
	req.Header.Set("Authorization", "Bearer "+engineerToken)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 OK for engineer viewing policies, got %d", w.Code)
	}
	var polResp struct {
		Success bool                    `json:"success"`
		Data    []model.RetentionPolicy `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &polResp); err != nil {
		t.Fatalf("failed to parse policies response: %v", err)
	}
	if !polResp.Success {
		t.Errorf("expected policies response success=true")
	}
	if len(polResp.Data) != 8 {
		t.Errorf("expected 8 default policies returned, got %d", len(polResp.Data))
	}

	// 3. PUT /api/retention/policies/:id with Operator token (lacks retention.manage) -> 403 Forbidden
	updatePayload := map[string]interface{}{
		"name":           "Updated Raw Telemetry",
		"retention_days": 45,
	}
	body, _ := json.Marshal(updatePayload)
	req, _ = http.NewRequest(http.MethodPut, "/api/retention/policies/pol_raw_telemetry", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+operatorToken)
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403 Forbidden for operator updating policy, got %d", w.Code)
	}

	// 4. PUT /api/retention/policies/:id with Administrator token (has retention.manage) -> 200 OK
	req, _ = http.NewRequest(http.MethodPut, "/api/retention/policies/pol_raw_telemetry", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+adminToken)
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 OK for admin updating policy, got %d (body: %s)", w.Code, w.Body.String())
	}

	// 5. POST /api/retention/policies/:id/dry-run with Engineer token -> 200 OK
	req, _ = http.NewRequest(http.MethodPost, "/api/retention/policies/pol_raw_telemetry/dry-run", nil)
	req.Header.Set("Authorization", "Bearer "+engineerToken)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 OK for dry run preview, got %d (body: %s)", w.Code, w.Body.String())
	}

	// 6. POST /api/retention/policies/:id/execute without confirm: true -> 400 Bad Request
	execReqEmpty := map[string]interface{}{
		"confirm": false,
	}
	bodyEmpty, _ := json.Marshal(execReqEmpty)
	req, _ = http.NewRequest(http.MethodPost, "/api/retention/policies/pol_raw_telemetry/execute", bytes.NewBuffer(bodyEmpty))
	req.Header.Set("Authorization", "Bearer "+adminToken)
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request without explicit confirmation, got %d", w.Code)
	}

	// 7. POST /api/retention/policies/:id/execute with Operator token (lacks retention.execute) -> 403 Forbidden
	execReqValid := map[string]interface{}{
		"confirm": true,
	}
	bodyValid, _ := json.Marshal(execReqValid)
	req, _ = http.NewRequest(http.MethodPost, "/api/retention/policies/pol_raw_telemetry/execute", bytes.NewBuffer(bodyValid))
	req.Header.Set("Authorization", "Bearer "+operatorToken)
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403 Forbidden for operator executing retention cleanup, got %d", w.Code)
	}

	// 8. POST /api/retention/policies/:id/execute on DISABLED policy -> returns disabled error
	req, _ = http.NewRequest(http.MethodPost, "/api/retention/policies/pol_raw_telemetry/execute", bytes.NewBuffer(bodyValid))
	req.Header.Set("Authorization", "Bearer "+adminToken)
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 response with success: false for disabled policy, got %d", w.Code)
	}
	var res map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &res)
	if success, ok := res["success"].(bool); !ok || success {
		t.Errorf("expected success: false for disabled policy, got: %v", res)
	}

	// 9. GET /api/retention/storage with Engineer token -> 200 OK
	req, _ = http.NewRequest(http.MethodGet, "/api/retention/storage", nil)
	req.Header.Set("Authorization", "Bearer "+engineerToken)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 OK for storage overview, got %d", w.Code)
	}
	var storResp struct {
		Success bool                    `json:"success"`
		Data    *model.StorageOverview  `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &storResp); err != nil {
		t.Fatalf("failed to parse storage overview response: %v", err)
	}
	if !storResp.Success || storResp.Data == nil {
		t.Fatalf("expected storage overview success=true and non-nil data")
	}
	if storResp.Data.CapacityStatus == "" {
		t.Errorf("expected non-empty CapacityStatus")
	}
	if storResp.Data.DatabaseType == "" {
		t.Errorf("expected non-empty DatabaseType")
	}

	// 10. GET /api/retention/history with Engineer token -> 200 OK
	req, _ = http.NewRequest(http.MethodGet, "/api/retention/history", nil)
	req.Header.Set("Authorization", "Bearer "+engineerToken)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 OK for history, got %d", w.Code)
	}

	// 11. Verify Audit trail records were logged
	var auditCount int64
	db.Model(&model.AuditTrail{}).Where("user_agent = ?", "RetentionService").Count(&auditCount)
	if auditCount < 1 {
		t.Errorf("expected audit trails logged for policy operations, got %d", auditCount)
	}
	_ = svc
}

func TestSuperadminPurgeAllRawTelemetry(t *testing.T) {
	r, db, superadminToken, engineerToken, operatorToken, _, cleanup := setupTestRetentionRouter(t)
	defer cleanup()

	// Create a regular ADMIN (non-SUPERADMIN) user to verify strict SUPERADMIN-only access
	var regularAdminRole model.Role
	if err := db.Where("name = ?", "ADMIN").First(&regularAdminRole).Error; err != nil || regularAdminRole.ID == 0 {
		regularAdminRole = model.Role{Name: "ADMIN", Description: "Operational Admin"}
		db.Create(&regularAdminRole)
	}
	regAdminHash, _ := bcrypt.GenerateFromPassword([]byte("regadmin123"), bcrypt.DefaultCost)
	db.Create(&model.User{
		Username: "regular_admin_test",
		Email:    "regadmin@test.local",
		Password: string(regAdminHash),
		FullName: "Regular Admin",
		RoleID:   regularAdminRole.ID,
		IsActive: true,
	})
	systemRepo := repository.NewSystemRepository(db)
	authSvc := service.NewAuthService(systemRepo, config.DefaultConfig())
	regAdminToken, _, errRA := authSvc.Login("regular_admin_test", "regadmin123", "127.0.0.1", "TestClient")
	if errRA != nil {
		t.Fatalf("failed login as regular admin: %v", errRA)
	}

	// Seed fresh live raw_data rows (including 1 second ago — younger than normal 1-day retention!)
	now := time.Now().UTC()
	for i := 0; i < 15; i++ {
		db.Create(&model.RawData{
			DeviceID:       1,
			ParameterID:    1,
			Value:          float64(10 + i),
			ProcessedValue: float64(10 + i),
			RawValue:       float64(100 + i),
			Quality:        model.QualityGood,
			ReceivedAt:     now.Add(-time.Duration(i) * time.Minute),
			StoredAt:       now,
		})
	}

	var beforeCount int64
	db.Model(&model.RawData{}).Count(&beforeCount)
	if beforeCount < 15 {
		t.Fatalf("expected at least 15 raw_data rows before purge, got %d", beforeCount)
	}

	validPayload, _ := json.Marshal(map[string]interface{}{
		"confirm":      true,
		"confirm_text": "PURGE",
	})

	// 1. Operator, Engineer, and Regular ADMIN must be strictly rejected with 403 Forbidden
	for roleName, tok := range map[string]string{
		"OPERATOR":      operatorToken,
		"ENGINEER":      engineerToken,
		"REGULAR_ADMIN": regAdminToken,
	} {
		req, _ := http.NewRequest(http.MethodPost, "/api/retention/raw-telemetry/purge", bytes.NewBuffer(validPayload))
		req.Header.Set("Authorization", "Bearer "+tok)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusForbidden {
			t.Errorf("expected 403 Forbidden for %s purging all raw telemetry, got %d (%s)", roleName, w.Code, w.Body.String())
		}
	}

	// 2. Superadmin without confirmation keyword must be rejected with 400 Bad Request
	badPayload, _ := json.Marshal(map[string]interface{}{
		"confirm":      true,
		"confirm_text": "WRONG",
	})
	reqBad, _ := http.NewRequest(http.MethodPost, "/api/retention/raw-telemetry/purge", bytes.NewBuffer(badPayload))
	reqBad.Header.Set("Authorization", "Bearer "+superadminToken)
	reqBad.Header.Set("Content-Type", "application/json")
	wBad := httptest.NewRecorder()
	r.ServeHTTP(wBad, reqBad)
	if wBad.Code != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request when confirm_text is invalid, got %d", wBad.Code)
	}

	// Ensure rows are still untouched after rejected attempts
	var midCount int64
	db.Model(&model.RawData{}).Count(&midCount)
	if midCount != beforeCount {
		t.Fatalf("expected raw_data count unchanged (%d) after rejected attempts, got %d", beforeCount, midCount)
	}

	// 3. Superadmin with valid confirmation ("HAPUS" or "PURGE") wipes 100% of raw_data
	idPayload, _ := json.Marshal(map[string]interface{}{
		"confirm":      true,
		"confirm_text": "HAPUS",
	})
	reqOK, _ := http.NewRequest(http.MethodPost, "/api/retention/raw-telemetry/purge", bytes.NewBuffer(idPayload))
	reqOK.Header.Set("Authorization", "Bearer "+superadminToken)
	reqOK.Header.Set("Content-Type", "application/json")
	wOK := httptest.NewRecorder()
	r.ServeHTTP(wOK, reqOK)
	if wOK.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for superadmin purge, got %d (%s)", wOK.Code, wOK.Body.String())
	}

	var purgeResp struct {
		Success bool                         `json:"success"`
		Data    *model.RetentionExecutionLog `json:"data"`
	}
	if err := json.Unmarshal(wOK.Body.Bytes(), &purgeResp); err != nil {
		t.Fatalf("failed unmarshaling purge response: %v", err)
	}
	if !purgeResp.Success || purgeResp.Data == nil {
		t.Fatalf("expected success=true and execution log data, got %s", wOK.Body.String())
	}
	if purgeResp.Data.DeletedCount != beforeCount {
		t.Errorf("expected DeletedCount=%d, got %d", beforeCount, purgeResp.Data.DeletedCount)
	}
	if purgeResp.Data.ExecutionType != model.ExecutionPurgeAll {
		t.Errorf("expected ExecutionType=PURGE_ALL, got %s", purgeResp.Data.ExecutionType)
	}

	// Verify 0 rows remain in raw_data
	var afterCount int64
	db.Model(&model.RawData{}).Count(&afterCount)
	if afterCount != 0 {
		t.Errorf("expected 0 rows remaining in raw_data after purge, got %d", afterCount)
	}

	// Verify AuditTrail recorded RAW_TELEMETRY_PURGED
	var purgeAudits int64
	db.Model(&model.AuditTrail{}).Where("action = ?", "RAW_TELEMETRY_PURGED").Count(&purgeAudits)
	if purgeAudits != 1 {
		t.Errorf("expected 1 RAW_TELEMETRY_PURGED audit trail entry, got %d", purgeAudits)
	}
}
