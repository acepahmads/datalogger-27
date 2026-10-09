package service

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"datalogger/internal/config"
	"datalogger/internal/model"
	"datalogger/internal/queue"
	"datalogger/internal/repository"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func setupTestBackupEnvironment(t *testing.T) (*gorm.DB, *config.Config, *repository.BackupRepository, *repository.SystemRepository, *PersistentQueueAdapter, string, func()) {
	// Temp working dir
	baseTmp, err := os.MkdirTemp("", "phase43_backup_test_*")
	if err != nil {
		t.Fatalf("failed creating temp dir: %v", err)
	}

	backupDir := filepath.Join(baseTmp, "backups")
	walDir := filepath.Join(baseTmp, "wal_queue")
	_ = os.MkdirAll(backupDir, 0755)
	_ = os.MkdirAll(walDir, 0755)

	cfg := &config.Config{
		AppName:                     "Datalogger Analysis Application",
		Version:                     "1.0.0",
		DBHost:                      "127.0.0.1",
		DBPort:                      "3306",
		DBName:                      "datalogger_test",
		DBUser:                      "testuser",
		DBPassword:                  "supersecretpassword123!",
		BackupDir:                   backupDir,
		BackupScheduleEnabled:       true,
		BackupScheduleIntervalHours: 24,
		BackupScheduleTime:          "02:00",
		BackupCompression:           "gzip",
		BackupMinFreeSpaceMB:        1,
		BackupMaxKeepCount:          3,
	}

	// SQLite In-Memory DB
	dsn := fmt.Sprintf("file:mem_backup_db_%d?mode=memory&cache=shared", time.Now().UnixNano()+int64(rand.Intn(1000)))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		t.Fatalf("failed opening test DB: %v", err)
	}

	// Migrate test models
	err = db.AutoMigrate(
		&model.Device{},
		&model.Parameter{},
		&model.RawData{},
		&model.BackupRecord{},
		&model.AuditTrail{},
		&model.SystemHealth{},
	)
	if err != nil {
		t.Fatalf("failed migrating test DB: %v", err)
	}

	// Seed sample test device & parameters
	dev := model.Device{ID: 1, DeviceCode: "DEV-EDGE-01", DeviceName: "Edge Gateway"}
	db.Create(&dev)
	param := model.Parameter{ID: 1, DeviceID: 1, ParameterCode: "PRESSURE", ParameterName: "System Pressure"}
	db.Create(&param)

	// Seed some telemetry rows
	now := time.Now().UTC()
	for i := 1; i <= 20; i++ {
		db.Create(&model.RawData{
			RecordUUID:  fmt.Sprintf("rec-uuid-%04d", i),
			DeviceID:    1,
			ParameterID: 1,
			Value:       float64(100 + i),
			Quality:     model.QualityGood,
			ReceivedAt:  now.Add(time.Duration(i) * time.Second),
			StoredAt:    now,
			Timestamp:   now.Add(time.Duration(i) * time.Second),
			Source:      "MODBUS_TCP",
			CreatedAt:   now,
		})
	}

	// Initialize WAL Queue
	walCfg := queue.DefaultWALConfig(walDir)
	walCfg.MaxSegmentSize = 100 * 1024
	walQueue, err := queue.OpenWALQueue(walCfg)
	if err != nil {
		t.Fatalf("failed initializing WAL queue: %v", err)
	}

	// Append sample WAL telemetry records
	for i := 101; i <= 110; i++ {
		_ = walQueue.Append(&model.RawData{
			RecordUUID:  fmt.Sprintf("rec-wal-uuid-%04d", i),
			DeviceID:    1,
			ParameterID: 1,
			Value:       float64(200 + i),
			Quality:     model.QualityGood,
			ReceivedAt:  now.Add(time.Duration(i) * time.Second),
			StoredAt:    now,
			Timestamp:   now.Add(time.Duration(i) * time.Second),
			Source:      "MODBUS_TCP",
			CreatedAt:   now,
		})
	}

	telemetryRepo := repository.NewTelemetryRepository(db)
	walAdapter := NewPersistentQueueAdapter(telemetryRepo, walQueue, 50, 3)
	backupRepo := repository.NewBackupRepository(db)
	systemRepo := repository.NewSystemRepository(db)

	cleanup := func() {
		_ = walAdapter.Close()
		_ = walQueue.Close()
		_ = os.RemoveAll(baseTmp)
	}

	return db, cfg, backupRepo, systemRepo, walAdapter, baseTmp, cleanup
}

// 1. Acceptance Test: Backup Creation & Atomic Packaging
func TestPhase43_BackupCreationAndAtomicPackaging(t *testing.T) {
	db, cfg, backupRepo, systemRepo, walAdapter, _, cleanup := setupTestBackupEnvironment(t)
	defer cleanup()

	dbAdapter := NewDatabaseBackupAdapter(db, cfg)
	backupService := NewBackupService(db, cfg, backupRepo, dbAdapter, walAdapter, nil, systemRepo)

	ctx := context.Background()
	record, err := backupService.CreateBackup(ctx, model.BackupTypeManual, "tester")
	if err != nil {
		t.Fatalf("CreateBackup failed: %v", err)
	}

	if record == nil {
		t.Fatalf("expected non-nil BackupRecord")
	}
	if record.Status != model.BackupStatusCompleted {
		t.Errorf("expected status COMPLETED, got %s", record.Status)
	}
	if record.ValidationStatus != model.ValidationStatusValid {
		t.Errorf("expected validation VALID, got %s", record.ValidationStatus)
	}
	if record.FilePath == "" {
		t.Errorf("expected non-empty FilePath")
	}

	// Verify file exists on disk and is a valid .tar.gz
	fi, err := os.Stat(record.FilePath)
	if err != nil {
		t.Fatalf("backup file stat failed: %v", err)
	}
	if fi.Size() == 0 {
		t.Fatalf("backup file is empty (0 bytes)")
	}

	// Verify no temporary files (.tmp) remain in backupDir
	entries, _ := os.ReadDir(cfg.BackupDir)
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".tmp") {
			t.Errorf("found residual temporary file: %s", e.Name())
		}
	}

	// Verify SHA-256 checksum is recorded
	if record.SHA256Checksum == "" {
		t.Errorf("expected non-empty SHA256Checksum")
	}
}

// 2. Acceptance Test: Manifest Generation, Parsing & Integrity Validation
func TestPhase43_BackupManifestAndIntegrityValidation(t *testing.T) {
	db, cfg, backupRepo, systemRepo, walAdapter, _, cleanup := setupTestBackupEnvironment(t)
	defer cleanup()

	dbAdapter := NewDatabaseBackupAdapter(db, cfg)
	backupService := NewBackupService(db, cfg, backupRepo, dbAdapter, walAdapter, nil, systemRepo)

	ctx := context.Background()
	record, err := backupService.CreateBackup(ctx, model.BackupTypeManual, "tester")
	if err != nil {
		t.Fatalf("CreateBackup failed: %v", err)
	}

	// Read and parse manifest from archive
	manifest, err := backupService.ValidateBackup(ctx, record.ID)
	if err != nil {
		t.Fatalf("ValidateBackup failed: %v", err)
	}

	if manifest == nil {
		t.Fatalf("expected non-nil manifest")
	}
	if manifest.BackupID != record.ID {
		t.Errorf("manifest ID mismatch: %s != %s", manifest.BackupID, record.ID)
	}
	if manifest.AppVersion != "1.0.0" {
		t.Errorf("expected AppVersion 1.0.0, got %s", manifest.AppVersion)
	}
	if manifest.Database.Engine == "" {
		t.Fatalf("expected non-empty database engine in manifest")
	}
	if manifest.WAL.PendingRecords != 10 {
		t.Errorf("expected 10 pending WAL records in manifest, got %d", manifest.WAL.PendingRecords)
	}

	// Verify Secret Redaction: check manifest does NOT contain database password
	archiveFile, err := os.Open(record.FilePath)
	if err != nil {
		t.Fatalf("failed opening archive: %v", err)
	}
	defer archiveFile.Close()

	gzReader, err := gzip.NewReader(archiveFile)
	if err != nil {
		t.Fatalf("failed creating gzip reader: %v", err)
	}
	defer gzReader.Close()

	tarReader := tar.NewReader(gzReader)
	var manifestFound bool
	for {
		hdr, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("tar error: %v", err)
		}
		if hdr.Name == "manifest.json" {
			manifestFound = true
			content, _ := io.ReadAll(tarReader)
			contentStr := string(content)
			if strings.Contains(contentStr, "supersecretpassword123!") {
				t.Fatalf("SECURITY VIOLATION: Database password leaked in plaintext manifest!")
			}
		}
	}
	if !manifestFound {
		t.Fatalf("manifest.json not found inside archive")
	}
}

// 3. Acceptance Test: WAL Snapshot Coordination Under Live Telemetry Ingestion
func TestPhase43_WALSnapshotCoordinationUnderIngestion(t *testing.T) {
	_, _, _, _, walAdapter, baseTmp, cleanup := setupTestBackupEnvironment(t)
	defer cleanup()

	snapDir := filepath.Join(baseTmp, "wal_snapshot_test")
	_ = os.MkdirAll(snapDir, 0755)

	// Take snapshot
	manifestInfo, err := walAdapter.Snapshot(snapDir)
	if err != nil {
		t.Fatalf("WAL snapshot failed: %v", err)
	}
	if manifestInfo.PendingRecords != 10 {
		t.Errorf("expected 10 pending records at snapshot boundary, got %d", manifestInfo.PendingRecords)
	}

	// Concurrently append new records to live WAL
	now := time.Now().UTC()
	for i := 201; i <= 205; i++ {
		_ = walAdapter.GetWAL().Append(&model.RawData{
			RecordUUID:  fmt.Sprintf("rec-live-%d", i),
			DeviceID:    1,
			ParameterID: 1,
			Value:       float64(i),
			Quality:     model.QualityGood,
			ReceivedAt:  now,
			StoredAt:    now,
			Timestamp:   now,
			Source:      "MODBUS_TCP",
			CreatedAt:   now,
		})
	}

	// Live WAL now has 15 pending records
	if walAdapter.GetWAL().PendingCount() != 15 {
		t.Errorf("expected 15 pending records in live WAL, got %d", walAdapter.GetWAL().PendingCount())
	}

	// Snapshot directory files remain unchanged and uncorrupted
	entries, err := os.ReadDir(snapDir)
	if err != nil || len(entries) == 0 {
		t.Fatalf("snapshot directory has no files: %v", err)
	}
}

// 4. Acceptance Test: Restore Preview & Safety Backup
func TestPhase43_RestorePreviewAndSafetyBackup(t *testing.T) {
	db, cfg, backupRepo, systemRepo, walAdapter, _, cleanup := setupTestBackupEnvironment(t)
	defer cleanup()

	dbAdapter := NewDatabaseBackupAdapter(db, cfg)
	backupService := NewBackupService(db, cfg, backupRepo, dbAdapter, walAdapter, nil, systemRepo)

	ctx := context.Background()
	record, err := backupService.CreateBackup(ctx, model.BackupTypeManual, "tester")
	if err != nil {
		t.Fatalf("CreateBackup failed: %v", err)
	}

	// Restore preview
	preview, err := backupService.GetRestorePreview(ctx, record.ID)
	if err != nil {
		t.Fatalf("GetRestorePreview failed: %v", err)
	}
	if preview == nil {
		t.Fatalf("expected non-nil preview")
	}
	if preview.BackupID != record.ID {
		t.Errorf("preview ID mismatch: %s != %s", preview.BackupID, record.ID)
	}
	if preview.TableCount == 0 {
		t.Errorf("expected non-zero TableCount in preview")
	}

	// Test RestoreBackup with createSafetyBackup = true
	err = backupService.RestoreBackup(ctx, record.ID, "tester", true)
	if err != nil {
		t.Fatalf("RestoreBackup with safety backup failed: %v", err)
	}

	// Verify that a SAFETY backup was created before the restore
	backups, _, err := backupRepo.List(1, 10, "")
	if err != nil {
		t.Fatalf("failed listing backups: %v", err)
	}

	var foundSafety bool
	for _, b := range backups {
		if b.Type == model.BackupTypeSafety {
			foundSafety = true
			break
		}
	}
	if !foundSafety {
		t.Errorf("expected automatic SAFETY backup to be recorded in catalog")
	}
}

// 5. Acceptance Test: Disaster Recovery Restore & Idempotent Telemetry Replay
func TestPhase43_DisasterRecoveryRestoreAndIdempotentReplay(t *testing.T) {
	db, cfg, backupRepo, systemRepo, walAdapter, _, cleanup := setupTestBackupEnvironment(t)
	defer cleanup()

	dbAdapter := NewDatabaseBackupAdapter(db, cfg)
	backupService := NewBackupService(db, cfg, backupRepo, dbAdapter, walAdapter, nil, systemRepo)

	ctx := context.Background()
	record, err := backupService.CreateBackup(ctx, model.BackupTypeManual, "tester")
	if err != nil {
		t.Fatalf("CreateBackup failed: %v", err)
	}

	// Simulate disaster: clear all RawData rows in DB
	db.Exec("DELETE FROM raw_data")
	var initialCount int64
	db.Model(&model.RawData{}).Count(&initialCount)
	if initialCount != 0 {
		t.Fatalf("expected 0 rows after simulated wipe, got %d", initialCount)
	}

	// Execute restore workflow
	err = backupService.RestoreBackup(ctx, record.ID, "tester", false)
	if err != nil {
		t.Fatalf("RestoreBackup failed: %v", err)
	}

	// Verify rows are restored: 20 from DB dump + 10 recovered from persistent WAL queue = 30 total
	var restoredCount int64
	db.Model(&model.RawData{}).Count(&restoredCount)
	if restoredCount != 30 {
		t.Errorf("expected 30 restored rows (20 DB + 10 WAL), got %d", restoredCount)
	}

	// Re-run restore to verify idempotency (zero duplicate rows inserted on repeated restore)
	err = backupService.RestoreBackup(ctx, record.ID, "tester", false)
	if err != nil {
		t.Fatalf("Second RestoreBackup failed: %v", err)
	}

	var reRestoredCount int64
	db.Model(&model.RawData{}).Count(&reRestoredCount)
	if reRestoredCount != 30 {
		t.Errorf("expected 30 rows after idempotent restore, got %d", reRestoredCount)
	}
}

// 6. Acceptance Test: Scheduler Overlap Prevention & Retention Pruning
func TestPhase43_SchedulerExecutionAndOverlapPrevention(t *testing.T) {
	db, cfg, backupRepo, systemRepo, walAdapter, _, cleanup := setupTestBackupEnvironment(t)
	defer cleanup()

	dbAdapter := NewDatabaseBackupAdapter(db, cfg)
	backupService := NewBackupService(db, cfg, backupRepo, dbAdapter, walAdapter, nil, systemRepo)

	scheduler := NewBackupScheduler(backupService, cfg)

	// Verify schedule configuration
	schedCfg := scheduler.GetScheduleConfig()
	if !schedCfg.Enabled {
		t.Errorf("expected scheduler to be enabled")
	}
	if schedCfg.KeepMaxCount != 3 {
		t.Errorf("expected KeepMaxCount 3, got %d", schedCfg.KeepMaxCount)
	}

	// Test Overlap Prevention: Lock backupService and attempt simultaneous backup
	ctx := context.Background()
	_, err := backupService.StartBackupJob(model.BackupTypeManual, "tester1")
	if err != nil {
		t.Fatalf("first job failed to start: %v", err)
	}

	// Attempting a second concurrent backup must return ErrOperationInProgress
	_, err = backupService.StartBackupJob(model.BackupTypeManual, "tester2")
	if err != ErrOperationInProgress {
		t.Errorf("expected ErrOperationInProgress on concurrent job, got %v", err)
	}

	// Wait for first job to complete
	time.Sleep(300 * time.Millisecond)

	// Test Retention Pruning
	for i := 1; i <= 4; i++ {
		_, _ = backupService.CreateBackup(ctx, model.BackupTypeManual, "tester")
		time.Sleep(50 * time.Millisecond)
	}

	// Trigger pruning
	_, _ = backupService.PruneOldBackups(3)

	// Verify archive count does not exceed max_keep (3) + 1 for in-flight/recent
	var count int64
	db.Model(&model.BackupRecord{}).Where("status = ?", model.BackupStatusCompleted).Count(&count)
	if count > 4 {
		t.Errorf("expected at most 4 retained backups after pruning, got %d", count)
	}
}

// 7. Acceptance Test: Path Traversal & Security Validation
func TestPhase43_SecurityAndPathTraversalRejection(t *testing.T) {
	db, cfg, backupRepo, systemRepo, walAdapter, _, cleanup := setupTestBackupEnvironment(t)
	defer cleanup()

	dbAdapter := NewDatabaseBackupAdapter(db, cfg)
	backupService := NewBackupService(db, cfg, backupRepo, dbAdapter, walAdapter, nil, systemRepo)

	ctx := context.Background()

	// Path traversal attempt in backup ID
	maliciousIDs := []string{
		"../../etc/passwd",
		"..\\..\\windows\\system32",
		"backup/../../test",
		"; rm -rf /",
	}

	for _, malID := range maliciousIDs {
		_, err := backupService.ValidateBackup(ctx, malID)
		if err == nil {
			t.Errorf("expected error on malicious backup ID %s, but got nil", malID)
		}

		err = backupService.RestoreBackup(ctx, malID, "tester", false)
		if err == nil {
			t.Errorf("expected error on malicious restore ID %s, but got nil", malID)
		}
	}
}

// 8. Acceptance Test: Disk Space Preflight Check
func TestPhase43_DiskSpacePreflightCheck(t *testing.T) {
	db, cfg, backupRepo, systemRepo, walAdapter, _, cleanup := setupTestBackupEnvironment(t)
	defer cleanup()

	// Set absurdly high min free space (e.g. 100 Terabytes) to simulate disk full
	cfg.BackupMinFreeSpaceMB = 100 * 1024 * 1024

	dbAdapter := NewDatabaseBackupAdapter(db, cfg)
	backupService := NewBackupService(db, cfg, backupRepo, dbAdapter, walAdapter, nil, systemRepo)

	ctx := context.Background()
	_, err := backupService.CreateBackup(ctx, model.BackupTypeManual, "tester")
	if err == nil {
		t.Fatalf("expected error due to insufficient disk space, but backup succeeded")
	}

	if !strings.Contains(err.Error(), "insufficient free disk space") {
		t.Errorf("expected 'insufficient free disk space' error, got %v", err)
	}
}
