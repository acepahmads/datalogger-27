package handler_test

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
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

func setupTestBackupRouter(t *testing.T) (*gin.Engine, *gorm.DB, string, string, string, *service.BackupService, func()) {
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
			backupGroup.POST("/import", middleware.RequirePermission("backup.create"), backupHandler.ImportBackup)
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

	return r, db, adminToken, operatorToken, backupDir, backupService, cleanup
}

func TestPhase43_API_AuthenticationRequired(t *testing.T) {
	r, _, _, _, _, _, cleanup := setupTestBackupRouter(t)
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
	r, _, adminToken, operatorToken, _, _, cleanup := setupTestBackupRouter(t)
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
	r, _, adminToken, _, _, _, cleanup := setupTestBackupRouter(t)
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
	r, _, adminToken, _, _, _, cleanup := setupTestBackupRouter(t)
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

func TestPhase43_API_Download_Workflow(t *testing.T) {
	r, db, adminToken, operatorToken, backupDir, _, cleanup := setupTestBackupRouter(t)
	defer cleanup()

	backupRepo := repository.NewBackupRepository(db)

	// Create a real mock backup archive file in backupDir
	validBackupID := "backup_VALID_20261009_120000"
	validFilename := validBackupID + ".tar.gz"
	validFilePath := filepath.Join(backupDir, validFilename)
	archivePayload := []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xffmock_binary_gzip_archive_stream_data_here")
	if err := os.WriteFile(validFilePath, archivePayload, 0644); err != nil {
		t.Fatalf("failed creating test backup archive file: %v", err)
	}

	record := &model.BackupRecord{
		ID:               validBackupID,
		Filename:         validFilename,
		FilePath:         validFilePath,
		Type:             model.BackupTypeManual,
		Status:           model.BackupStatusCompleted,
		ValidationStatus: model.ValidationStatusValid,
		SizeBytes:        int64(len(archivePayload)),
		CreatedBy:        "admin",
	}
	if err := backupRepo.Create(record); err != nil {
		t.Fatalf("failed creating backup record: %v", err)
	}

	// 1. Unauthenticated download must be rejected with 401 Unauthorized
	{
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", fmt.Sprintf("/api/backups/%s/download", validBackupID), nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected 401 Unauthorized for missing token, got %d", w.Code)
		}
		var errResp map[string]interface{}
		_ = json.Unmarshal(w.Body.Bytes(), &errResp)
		if errResp["success"] != false || errResp["error"] != "Authorization header required" {
			t.Errorf("unexpected error payload: %v", errResp)
		}
	}

	// 2. User without backup.view permission (e.g., Operator) must be rejected with 403 Forbidden
	{
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", fmt.Sprintf("/api/backups/%s/download", validBackupID), nil)
		req.Header.Set("Authorization", "Bearer "+operatorToken)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("expected 403 Forbidden for operator without backup.view, got %d", w.Code)
		}
		var errResp map[string]interface{}
		_ = json.Unmarshal(w.Body.Bytes(), &errResp)
		if errResp["success"] != false {
			t.Errorf("unexpected error payload: %v", errResp)
		}
	}

	// 3. Nonexistent backup ID must return 404 Not Found
	{
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/backups/nonexistent_backup_id/download", nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected 404 Not Found for missing backup, got %d", w.Code)
		}
		var errResp map[string]interface{}
		_ = json.Unmarshal(w.Body.Bytes(), &errResp)
		if errResp["success"] != false {
			t.Errorf("unexpected error payload: %v", errResp)
		}
	}

	// 4. Missing physical file on disk must return 404 Not Found
	{
		ghostID := "backup_ghost_file"
		ghostRecord := &model.BackupRecord{
			ID:               ghostID,
			Filename:         ghostID + ".tar.gz",
			FilePath:         filepath.Join(backupDir, ghostID+".tar.gz"),
			Type:             model.BackupTypeManual,
			Status:           model.BackupStatusCompleted,
			ValidationStatus: model.ValidationStatusValid,
			SizeBytes:        12345,
			CreatedBy:        "admin",
		}
		_ = backupRepo.Create(ghostRecord)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", fmt.Sprintf("/api/backups/%s/download", ghostID), nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected 404 Not Found for missing file on disk, got %d", w.Code)
		}
	}

	// 5. Path traversal attempts must be rejected (400 or 403)
	{
		traversalRecord := &model.BackupRecord{
			ID:               "backup_traversal_exploit",
			Filename:         "system_file.tar.gz",
			FilePath:         filepath.Join(filepath.Dir(backupDir), "outside_secret.txt"),
			Type:             model.BackupTypeManual,
			Status:           model.BackupStatusCompleted,
			ValidationStatus: model.ValidationStatusValid,
			SizeBytes:        100,
			CreatedBy:        "admin",
		}
		_ = os.WriteFile(traversalRecord.FilePath, []byte("forbidden data"), 0644)
		_ = backupRepo.Create(traversalRecord)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", fmt.Sprintf("/api/backups/%s/download", traversalRecord.ID), nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusForbidden && w.Code != http.StatusBadRequest {
			t.Errorf("expected 403 Forbidden or 400 Bad Request for path traversal attempt, got %d", w.Code)
		}
	}

	// 6. Direct ID traversal attempts (e.g. ..%2f..) must be rejected with 400 Bad Request
	{
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/backups/..%2f..%2fmalicious/download", nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest && w.Code != http.StatusNotFound {
			t.Errorf("expected 400 Bad Request or 404 for invalid ID traversal, got %d", w.Code)
		}
	}

	// 7. Authenticated user with backup.view downloads valid archive: 200 OK with correct stream & headers
	{
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", fmt.Sprintf("/api/backups/%s/download", validBackupID), nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 OK for authenticated download, got %d (body: %s)", w.Code, w.Body.String())
		}

		// Verify Content-Type
		contentType := w.Header().Get("Content-Type")
		if !strings.Contains(contentType, "application/gzip") {
			t.Errorf("expected Content-Type containing application/gzip, got %s", contentType)
		}

		// Verify Content-Disposition
		contentDisposition := w.Header().Get("Content-Disposition")
		expectedDisposition := fmt.Sprintf("attachment; filename=\"%s\"", validFilename)
		if contentDisposition != expectedDisposition {
			t.Errorf("expected Content-Disposition '%s', got '%s'", expectedDisposition, contentDisposition)
		}

		// Verify Content-Length
		contentLength := w.Header().Get("Content-Length")
		expectedLength := fmt.Sprintf("%d", len(archivePayload))
		if contentLength != expectedLength {
			t.Errorf("expected Content-Length '%s', got '%s'", expectedLength, contentLength)
		}

		// Verify Access-Control-Expose-Headers
		exposedHeaders := w.Header().Get("Access-Control-Expose-Headers")
		if !strings.Contains(exposedHeaders, "Content-Disposition") {
			t.Errorf("expected Access-Control-Expose-Headers to contain Content-Disposition, got '%s'", exposedHeaders)
		}

		// Verify exact binary contents
		bodyBytes := w.Body.Bytes()
		if !bytes.Equal(bodyBytes, archivePayload) {
			t.Errorf("downloaded content mismatch: expected %d bytes, got %d bytes", len(archivePayload), len(bodyBytes))
		}
	}
}

func TestPhase43_DashboardVerification_TwoArchivesDownloadAndIntegrity(t *testing.T) {
	r, _, adminToken, _, _, backupService, cleanup := setupTestBackupRouter(t)
	defer cleanup()

	ctx := context.Background()

	// Create Archive 1 (Manual)
	archive1, err := backupService.CreateBackup(ctx, model.BackupTypeManual, "operator_admin")
	if err != nil {
		t.Fatalf("failed creating archive 1: %v", err)
	}
	if archive1.SizeBytes <= 0 {
		t.Fatalf("expected archive 1 size > 0, got %d", archive1.SizeBytes)
	}

	// Create Archive 2 (Scheduled)
	time.Sleep(10 * time.Millisecond)
	archive2, err := backupService.CreateBackup(ctx, model.BackupTypeScheduled, "scheduler_cron")
	if err != nil {
		t.Fatalf("failed creating archive 2: %v", err)
	}
	if archive2.SizeBytes <= 0 {
		t.Fatalf("expected archive 2 size > 0, got %d", archive2.SizeBytes)
	}

	archivesToTest := []*model.BackupRecord{archive1, archive2}

	for idx, targetArchive := range archivesToTest {
		t.Run(fmt.Sprintf("Archive_%d_%s", idx+1, targetArchive.ID), func(t *testing.T) {
			// 1. Send authenticated GET request to download endpoint
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", fmt.Sprintf("/api/backups/%s/download", targetArchive.ID), nil)
			req.Header.Set("Authorization", "Bearer "+adminToken)
			r.ServeHTTP(w, req)

			// Confirm HTTP 200 OK
			if w.Code != http.StatusOK {
				t.Fatalf("expected HTTP 200 OK, got %d (body: %s)", w.Code, w.Body.String())
			}

			// Confirm Content-Type is gzip
			contentType := w.Header().Get("Content-Type")
			if !strings.Contains(contentType, "application/gzip") {
				t.Errorf("expected Content-Type containing application/gzip, got %s", contentType)
			}

			// Confirm Content-Disposition contains expected attachment filename
			contentDisposition := w.Header().Get("Content-Disposition")
			expectedFilename := targetArchive.Filename
			if !strings.Contains(contentDisposition, expectedFilename) {
				t.Errorf("expected Content-Disposition to contain filename '%s', got '%s'", expectedFilename, contentDisposition)
			}

			// Confirm downloaded byte size matches catalog size and is non-empty
			downloadedBytes := w.Body.Bytes()
			if len(downloadedBytes) == 0 {
				t.Fatalf("downloaded archive body is unexpectedly empty")
			}
			if int64(len(downloadedBytes)) != targetArchive.SizeBytes {
				t.Errorf("downloaded archive size mismatch: catalog=%d bytes, downloaded=%d bytes", targetArchive.SizeBytes, len(downloadedBytes))
			}

			// Confirm the file can be opened and decompressed as a valid gzip archive
			gzr, err := gzip.NewReader(bytes.NewReader(downloadedBytes))
			if err != nil {
				t.Fatalf("failed decompressing downloaded gzip payload: %v", err)
			}
			defer gzr.Close()

			// Confirm TAR archive contents: manifest.json and database/WAL artifacts are present
			tr := tar.NewReader(gzr)
			var foundManifest, foundDatabase bool
			var manifestBytes []byte

			for {
				header, err := tr.Next()
				if err == io.EOF {
					break
				}
				if err != nil {
					t.Fatalf("corrupted tar archive stream: %v", err)
				}

				if header.Name == "manifest.json" {
					foundManifest = true
					manifestBytes, _ = io.ReadAll(tr)
				}
				if strings.HasSuffix(header.Name, ".sql") || strings.Contains(header.Name, "database") {
					foundDatabase = true
				}
			}

			if !foundManifest {
				t.Errorf("archive %s is missing manifest.json", targetArchive.ID)
			}
			if !foundDatabase {
				t.Errorf("archive %s is missing database artifact", targetArchive.ID)
			}

			// Verify manifest contents
			if len(manifestBytes) > 0 {
				var manifest model.BackupManifest
				if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
					t.Errorf("failed unmarshaling manifest.json: %v", err)
				} else {
					if manifest.BackupID != targetArchive.ID {
						t.Errorf("manifest backup_id mismatch: expected %s, got %s", targetArchive.ID, manifest.BackupID)
					}
					if manifest.FormatVersion == "" {
						t.Errorf("manifest format_version is empty")
					}
				}
			}

			// Confirm the downloaded archive passes the existing backup validation process
			validatedManifest, valErr := backupService.ValidateBackup(ctx, targetArchive.ID)
			if valErr != nil {
				t.Fatalf("archive %s failed validation process: %v", targetArchive.ID, valErr)
			}
			if validatedManifest == nil {
				t.Fatalf("validation returned nil manifest")
			}
		})
	}
}

func buildMultipartUploadRequest(t *testing.T, url, fieldName, filename string, payload []byte, bearerToken string) *http.Request {
	t.Helper()
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	part, err := mw.CreateFormFile(fieldName, filename)
	if err != nil {
		t.Fatalf("failed creating form file: %v", err)
	}
	if _, err := part.Write(payload); err != nil {
		t.Fatalf("failed writing upload payload: %v", err)
	}
	if err := mw.Close(); err != nil {
		t.Fatalf("failed closing multipart writer: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, url, &body)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	if bearerToken != "" {
		req.Header.Set("Authorization", "Bearer "+bearerToken)
	}
	return req
}

func TestBackupImportAndRestoreEndToEnd(t *testing.T) {
	// 1. Setup Server A, seed custom device, create backup, and download .tar.gz bytes
	routerA, dbA, adminTokenA, _, _, backupSvcA, cleanupA := setupTestBackupRouter(t)
	defer cleanupA()

	srcDevice := model.Device{
		DeviceCode: "IMPORT-EDGE-99",
		DeviceName: "Imported External Gateway",
	}
	if err := dbA.Create(&srcDevice).Error; err != nil {
		t.Fatalf("failed creating test device on Server A: %v", err)
	}
	srcParam := model.Parameter{
		DeviceID:      srcDevice.ID,
		ParameterCode: "PM10_IMP",
		ParameterName: "Imported PM10",
	}
	if err := dbA.Create(&srcParam).Error; err != nil {
		t.Fatalf("failed creating test parameter on Server A: %v", err)
	}

	// Seed 600 telemetry rows (>128 KB SQL dump) so streaming multi-64KB chunk head/tail validation is exercised
	nowUTC := time.Now().UTC()
	batchRows := make([]model.RawData, 0, 600)
	for i := 0; i < 600; i++ {
		batchRows = append(batchRows, model.RawData{
			RecordUUID:  fmt.Sprintf("imp-stream-uuid-%04d", i),
			DeviceID:    srcDevice.ID,
			ParameterID: srcParam.ID,
			Value:       float64(42 + (i % 50)),
			Quality:     model.QualityGood,
			ReceivedAt:  nowUTC.Add(time.Duration(i) * time.Second),
			StoredAt:    nowUTC,
			Timestamp:   nowUTC.Add(time.Duration(i) * time.Second),
			Source:      "MODBUS_TCP",
			CreatedAt:   nowUTC,
		})
	}
	if err := dbA.CreateInBatches(batchRows, 200).Error; err != nil {
		t.Fatalf("failed seeding telemetry rows on Server A: %v", err)
	}

	ctx := context.Background()
	exportedRec, err := backupSvcA.CreateBackup(ctx, model.BackupTypeManual, "admin")
	if err != nil {
		t.Fatalf("failed creating backup on Server A: %v", err)
	}

	dlReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/backups/%s/download", exportedRec.ID), nil)
	dlReq.Header.Set("Authorization", "Bearer "+adminTokenA)
	dlW := httptest.NewRecorder()
	routerA.ServeHTTP(dlW, dlReq)
	if dlW.Code != http.StatusOK {
		t.Fatalf("expected 200 OK downloading Server A backup, got %d", dlW.Code)
	}
	archiveBytes := dlW.Body.Bytes()

	// 2. Setup fresh Server B (simulating a newly installed mini-PC / edge gateway)
	routerB, dbB, adminTokenB, operatorTokenB, _, backupSvcB, cleanupB := setupTestBackupRouter(t)
	defer cleanupB()

	// Verify Server B does NOT yet have IMPORT-EDGE-99
	var preCount int64
	dbB.Model(&model.Device{}).Where("device_code = ?", "IMPORT-EDGE-99").Count(&preCount)
	if preCount != 0 {
		t.Fatalf("expected 0 IMPORT-EDGE-99 devices on fresh Server B, got %d", preCount)
	}

	// 2a. RBAC check: Operator role without backup.create must be rejected with 403
	opReq := buildMultipartUploadRequest(t, "/api/backups/import", "file", exportedRec.Filename, archiveBytes, operatorTokenB)
	opW := httptest.NewRecorder()
	routerB.ServeHTTP(opW, opReq)
	if opW.Code != http.StatusForbidden {
		t.Fatalf("expected 403 Forbidden for operator import, got %d (%s)", opW.Code, opW.Body.String())
	}

	// 2b. Validation check: non-.tar.gz extension must be rejected with 400
	badExtReq := buildMultipartUploadRequest(t, "/api/backups/import", "file", "notes.txt", []byte("hello"), adminTokenB)
	badExtW := httptest.NewRecorder()
	routerB.ServeHTTP(badExtW, badExtReq)
	if badExtW.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 Bad Request for .txt upload, got %d", badExtW.Code)
	}

	// 2c. Validation check: corrupted .tar.gz payload must be rejected with 400
	corruptReq := buildMultipartUploadRequest(t, "/api/backups/import", "file", "corrupted.tar.gz", []byte("not-a-valid-gzip-stream"), adminTokenB)
	corruptW := httptest.NewRecorder()
	routerB.ServeHTTP(corruptW, corruptReq)
	if corruptW.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 Bad Request for corrupted .tar.gz, got %d (%s)", corruptW.Code, corruptW.Body.String())
	}

	// 3. Import valid external .tar.gz onto Server B
	impReq := buildMultipartUploadRequest(t, "/api/backups/import", "file", exportedRec.Filename, archiveBytes, adminTokenB)
	impW := httptest.NewRecorder()
	routerB.ServeHTTP(impW, impReq)
	if impW.Code != http.StatusOK {
		t.Fatalf("expected 200 OK importing valid .tar.gz, got %d (%s)", impW.Code, impW.Body.String())
	}

	var impResp struct {
		Success bool `json:"success"`
		Data    struct {
			Backup model.BackupRecord `json:"backup"`
		} `json:"data"`
	}
	if err := json.Unmarshal(impW.Body.Bytes(), &impResp); err != nil {
		t.Fatalf("failed decoding import response: %v", err)
	}
	if !impResp.Success {
		t.Fatalf("expected success=true on import response")
	}
	imported := impResp.Data.Backup
	if imported.Type != model.BackupTypeImported {
		t.Errorf("expected imported backup type IMPORTED, got %s", imported.Type)
	}
	if imported.ValidationStatus != model.ValidationStatusValid {
		t.Errorf("expected VALID status, got %s", imported.ValidationStatus)
	}
	if !imported.IsCompatible {
		t.Errorf("expected IsCompatible=true on imported backup")
	}

	// 4. Re-import same archive to verify non-destructive collision handling (*_imp_*)
	dupReq := buildMultipartUploadRequest(t, "/api/backups/import", "file", exportedRec.Filename, archiveBytes, adminTokenB)
	dupW := httptest.NewRecorder()
	routerB.ServeHTTP(dupW, dupReq)
	if dupW.Code != http.StatusOK {
		t.Fatalf("expected 200 OK on duplicate import, got %d (%s)", dupW.Code, dupW.Body.String())
	}
	var dupResp struct {
		Data struct {
			Backup model.BackupRecord `json:"backup"`
		} `json:"data"`
	}
	_ = json.Unmarshal(dupW.Body.Bytes(), &dupResp)
	if dupResp.Data.Backup.ID == imported.ID {
		t.Errorf("expected unique collision-safe ID on duplicate import, got identical ID %s", dupResp.Data.Backup.ID)
	}
	if !strings.Contains(dupResp.Data.Backup.ID, "_imp_") {
		t.Errorf("expected collision-safe ID to contain _imp_, got %s", dupResp.Data.Backup.ID)
	}

	// 5. Verify Restore Preview works on imported archive
	prevReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/backups/%s/restore-preview", imported.ID), nil)
	prevReq.Header.Set("Authorization", "Bearer "+adminTokenB)
	prevW := httptest.NewRecorder()
	routerB.ServeHTTP(prevW, prevReq)
	if prevW.Code != http.StatusOK {
		t.Fatalf("expected 200 OK on restore-preview for imported archive, got %d (%s)", prevW.Code, prevW.Body.String())
	}

	// 6. Execute synchronous Restore from imported archive onto Server B
	if err := backupSvcB.RestoreBackup(ctx, imported.ID, "admin", true); err != nil {
		t.Fatalf("RestoreBackup from imported archive failed: %v", err)
	}

	// Confirm IMPORT-EDGE-99 from Server A now exists on Server B!
	var postCount int64
	dbB.Model(&model.Device{}).Where("device_code = ?", "IMPORT-EDGE-99").Count(&postCount)
	if postCount != 1 {
		t.Fatalf("expected IMPORT-EDGE-99 device to be restored on Server B, got count=%d", postCount)
	}

	// Confirm imported backup record is still intact & valid in Server B's catalog after restore
	afterRec, _, err := backupSvcB.GetBackupByID(imported.ID)
	if err != nil || afterRec == nil {
		t.Fatalf("expected imported backup %s to remain in Server B catalog after restore, got err=%v", imported.ID, err)
	}
}
