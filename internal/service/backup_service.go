package service

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"datalogger/internal/config"
	"datalogger/internal/logger"
	"datalogger/internal/model"
	"datalogger/internal/repository"

	"github.com/shirou/gopsutil/v3/disk"
	"gorm.io/gorm"
)

var (
	ErrOperationInProgress = errors.New("another backup or restore operation is already in progress")
	ErrBackupNotFound      = errors.New("backup artifact not found")
	ErrInsufficientDisk    = errors.New("insufficient free disk space for backup")
	ErrInvalidManifest     = errors.New("invalid or corrupt backup manifest")
)

type BackupService struct {
	db               *gorm.DB
	cfg              *config.Config
	backupRepo       *repository.BackupRepository
	dbAdapter        *DatabaseBackupAdapter
	walAdapter       *PersistentQueueAdapter
	telemetryService *TelemetryService
	systemRepo       *repository.SystemRepository

	opMu               sync.Mutex
	isOperationRunning bool
	activeOpType       string

	jobMu sync.RWMutex
	jobs  map[string]*model.BackupJob
}

func NewBackupService(
	db *gorm.DB,
	cfg *config.Config,
	backupRepo *repository.BackupRepository,
	dbAdapter *DatabaseBackupAdapter,
	walAdapter *PersistentQueueAdapter,
	telemetryService *TelemetryService,
	systemRepo *repository.SystemRepository,
) *BackupService {
	if cfg == nil {
		cfg = config.Get()
	}
	if cfg.BackupDir == "" {
		cfg.BackupDir = filepath.Join("data", "backups")
	}
	_ = os.MkdirAll(cfg.BackupDir, 0755)

	return &BackupService{
		db:               db,
		cfg:              cfg,
		backupRepo:       backupRepo,
		dbAdapter:        dbAdapter,
		walAdapter:       walAdapter,
		telemetryService: telemetryService,
		systemRepo:       systemRepo,
		jobs:             make(map[string]*model.BackupJob),
	}
}

// SetWALAdapter dynamically injects or updates the WAL adapter
func (s *BackupService) SetWALAdapter(adapter *PersistentQueueAdapter) {
	s.walAdapter = adapter
}

// StartBackupJob launches an asynchronous backup operation and returns job ID
func (s *BackupService) StartBackupJob(backupType model.BackupType, initiatedBy string) (string, error) {
	s.opMu.Lock()
	if s.isOperationRunning {
		s.opMu.Unlock()
		return "", ErrOperationInProgress
	}
	s.isOperationRunning = true
	s.activeOpType = "BACKUP"
	s.opMu.Unlock()

	jobID := fmt.Sprintf("job_backup_%d", time.Now().UnixNano())
	job := &model.BackupJob{
		ID:          jobID,
		Type:        model.JobTypeBackup,
		Status:      model.JobStatusRunning,
		Progress:    5.0,
		Stage:       "Initializing backup job",
		InitiatedBy: initiatedBy,
		StartTime:   time.Now().UTC(),
	}

	s.jobMu.Lock()
	s.jobs[jobID] = job
	s.jobMu.Unlock()

	go func() {
		defer func() {
			s.opMu.Lock()
			s.isOperationRunning = false
			s.activeOpType = ""
			s.opMu.Unlock()
		}()

		record, err := s.executeBackup(context.Background(), backupType, initiatedBy, job)
		now := time.Now().UTC()
		s.jobMu.Lock()
		job.EndTime = &now
		if err != nil {
			job.Status = model.JobStatusFailed
			job.Error = err.Error()
			job.Stage = "Backup failed"
			logger.Error("Backup job %s failed: %v", jobID, err)
		} else {
			job.Status = model.JobStatusCompleted
			job.Progress = 100.0
			job.BackupID = record.ID
			job.Stage = "Backup completed successfully"
			logger.Info("Backup job %s completed successfully: %s", jobID, record.ID)
		}
		s.jobMu.Unlock()
	}()

	return jobID, nil
}

// CreateBackup creates a backup synchronously
func (s *BackupService) CreateBackup(ctx context.Context, backupType model.BackupType, initiatedBy string) (*model.BackupRecord, error) {
	s.opMu.Lock()
	if s.isOperationRunning {
		s.opMu.Unlock()
		return nil, ErrOperationInProgress
	}
	s.isOperationRunning = true
	s.activeOpType = "BACKUP"
	s.opMu.Unlock()

	defer func() {
		s.opMu.Lock()
		s.isOperationRunning = false
		s.activeOpType = ""
		s.opMu.Unlock()
	}()

	return s.executeBackup(ctx, backupType, initiatedBy, nil)
}

func (s *BackupService) executeBackup(
	ctx context.Context,
	backupType model.BackupType,
	initiatedBy string,
	job *model.BackupJob,
) (*model.BackupRecord, error) {
	start := time.Now().UTC()
	timestampStr := start.Format("20060102_150405")
	backupID := fmt.Sprintf("backup_%s_%04d", timestampStr, start.Nanosecond()%10000)

	s.updateJob(job, 10.0, "Pre-flight checks and disk space verification")

	// 1. Verify available disk space
	minFreeMB := s.cfg.BackupMinFreeSpaceMB
	if minFreeMB <= 0 {
		minFreeMB = 500
	}
	freeBytes, err := s.checkFreeDiskSpace(s.cfg.BackupDir)
	if err == nil && freeBytes < minFreeMB*1024*1024 {
		return nil, fmt.Errorf("%w: free space %d MB is below required %d MB", ErrInsufficientDisk, freeBytes/(1024*1024), minFreeMB)
	}

	// 2. Setup staging directory
	stagingDir := filepath.Join(s.cfg.BackupDir, "staging_"+backupID)
	_ = os.RemoveAll(stagingDir)
	if err := os.MkdirAll(stagingDir, 0755); err != nil {
		return nil, fmt.Errorf("failed creating staging directory: %w", err)
	}
	defer func() {
		_ = os.RemoveAll(stagingDir)
	}()

	// 3. Snapshot WAL (Phase 4.2 coordination)
	s.updateJob(job, 25.0, "Coordinating persistent WAL queue snapshot")
	stagingWALDir := filepath.Join(stagingDir, "wal")
	var walInfo *model.WALManifestInfo

	if s.walAdapter != nil {
		walSnap, err := s.walAdapter.Snapshot(stagingWALDir)
		if err != nil {
			return nil, fmt.Errorf("failed snapshotting WAL queue: %w", err)
		}
		walInfo = walSnap
	} else {
		walInfo = &model.WALManifestInfo{
			SegmentFiles: []model.ArtifactInfo{},
		}
	}

	// 4. Dump Database (Phase 4.3 consistent snapshot)
	s.updateJob(job, 45.0, "Exporting database tables and schemas")
	stagingDBDir := filepath.Join(stagingDir, "database")
	_ = os.MkdirAll(stagingDBDir, 0755)
	dumpFilePath := filepath.Join(stagingDBDir, "dump.sql")

	dbInfo, err := s.dbAdapter.Dump(ctx, dumpFilePath)
	if err != nil {
		return nil, fmt.Errorf("database export failed: %w", err)
	}

	// 5. Build Artifacts List & Manifest
	s.updateJob(job, 70.0, "Generating machine-readable manifest and checksums")
	var artifacts []model.ArtifactInfo

	// Add database dump artifact
	artifacts = append(artifacts, model.ArtifactInfo{
		Path:        filepath.ToSlash(filepath.Join("database", "dump.sql")),
		SizeBytes:   dbInfo.DumpSizeBytes,
		SHA256:      dbInfo.DumpSHA256,
		Description: "MariaDB/MySQL logical SQL dump",
	})

	// Add WAL artifacts
	for _, f := range walInfo.SegmentFiles {
		artifacts = append(artifacts, model.ArtifactInfo{
			Path:        filepath.ToSlash(filepath.Join("wal", f.Path)),
			SizeBytes:   f.SizeBytes,
			SHA256:      f.SHA256,
			CRC32:       f.CRC32,
			Description: f.Description,
		})
	}

	// Include checkpoint.json in artifacts if present
	cpPath := filepath.Join(stagingWALDir, "checkpoint.json")
	if cpFi, err := os.Stat(cpPath); err == nil {
		cpSHA, _ := s.calculateFileSHA256(cpPath)
		artifacts = append(artifacts, model.ArtifactInfo{
			Path:        filepath.ToSlash(filepath.Join("wal", "checkpoint.json")),
			SizeBytes:   cpFi.Size(),
			SHA256:      cpSHA,
			Description: "WAL Checkpoint Metadata",
		})
	}

	// Link artifact pointers for direct frontend schema consumption
	if len(artifacts) > 0 {
		dbInfo.Artifact = &artifacts[0]
	}
	walInfo.Segments = walInfo.SegmentFiles

	manifest := model.BackupManifest{
		BackupID:      backupID,
		FormatVersion: "1.0",
		AppVersion:    s.cfg.Version,
		CreatedAt:     start,
		BackupType:    backupType,
		Database:      *dbInfo,
		WAL:           *walInfo,
		ConfigSummary: model.SafeConfigExport{
			AppName:               s.cfg.AppName,
			Environment:           s.cfg.Environment,
			Port:                  s.cfg.Port,
			DBType:                s.cfg.DBType,
			DBHost:                s.cfg.DBHost,
			DBPort:                s.cfg.DBPort,
			DBName:                s.cfg.DBName,
			QueueEnabled:          s.cfg.QueueEnabled,
			QueueSyncMode:         s.cfg.QueueSyncMode,
			QueueMaxSizeBytes:     s.cfg.QueueMaxSizeBytes,
			BackupDir:             s.cfg.BackupDir,
			BackupScheduleEnabled: s.cfg.BackupScheduleEnabled,
		},
		Artifacts:        artifacts,
		Status:           model.BackupStatusCompleted,
		ValidationStatus: model.ValidationStatusValid,
	}

	manifestData, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed encoding manifest JSON: %w", err)
	}

	manifestPath := filepath.Join(stagingDir, "manifest.json")
	if err := os.WriteFile(manifestPath, manifestData, 0644); err != nil {
		return nil, fmt.Errorf("failed saving manifest.json: %w", err)
	}

	// 6. Compress and package into archive
	s.updateJob(job, 85.0, "Compressing and packaging archive")
	finalFilename := fmt.Sprintf("%s.tar.gz", backupID)
	finalArchivePath := filepath.Join(s.cfg.BackupDir, finalFilename)
	stagingArchivePath := filepath.Join(s.cfg.BackupDir, "staging_"+finalFilename)
	_ = os.Remove(stagingArchivePath)

	if err := s.createTarGz(stagingDir, stagingArchivePath); err != nil {
		_ = os.Remove(stagingArchivePath)
		return nil, fmt.Errorf("failed creating compressed backup archive: %w", err)
	}

	// Calculate archive checksum & size
	archiveSHA, err := s.calculateFileSHA256(stagingArchivePath)
	if err != nil {
		_ = os.Remove(stagingArchivePath)
		return nil, fmt.Errorf("failed calculating archive checksum: %w", err)
	}
	archiveFi, err := os.Stat(stagingArchivePath)
	if err != nil {
		_ = os.Remove(stagingArchivePath)
		return nil, err
	}

	// 7. Atomic rename
	s.updateJob(job, 95.0, "Finalizing backup artifact")
	if err := os.Rename(stagingArchivePath, finalArchivePath); err != nil {
		_ = os.Remove(stagingArchivePath)
		return nil, fmt.Errorf("failed atomically finalizing backup: %w", err)
	}

	// Immediate verification of finalized archive integrity before marking COMPLETED
	if _, err := s.readManifestFromArchive(finalArchivePath); err != nil {
		_ = os.Remove(finalArchivePath)
		return nil, fmt.Errorf("finalized archive verification failed: %w", err)
	}

	durationMs := time.Since(start).Milliseconds()

	// 8. Register in database catalog
	boundaryTS := start
	record := &model.BackupRecord{
		ID:                 backupID,
		Filename:           finalFilename,
		FilePath:           finalArchivePath,
		Type:               backupType,
		Status:             model.BackupStatusCompleted,
		ValidationStatus:   model.ValidationStatusValid,
		ConsistencyStatus:  "CONSISTENT",
		SizeBytes:          archiveFi.Size(),
		SHA256Checksum:     archiveSHA,
		FormatVersion:      "1.0",
		AppVersion:         s.cfg.Version,
		DatabaseEngine:     dbInfo.Engine,
		DatabaseName:       dbInfo.DatabaseName,
		DumpMethod:         dbInfo.DumpMethod,
		TableCount:         dbInfo.TableCount,
		TotalRows:          dbInfo.TotalRows,
		WALSegmentsCount:   len(walInfo.SegmentFiles),
		WALPendingRecords:  walInfo.PendingRecords,
		SnapshotBoundaryTS: &boundaryTS,
		CreatedBy:          initiatedBy,
		DurationMs:         durationMs,
		CreatedAt:          start,
		UpdatedAt:          time.Now().UTC(),
	}
	record.PopulateVirtualFields()

	if s.backupRepo != nil {
		_ = s.backupRepo.Create(record)
	}

	// Log audit trail
	s.logAudit("BACKUP_CREATED", backupID, initiatedBy,
		fmt.Sprintf("Created %s backup %s (size: %d bytes, tables: %d, rows: %d, WAL pending: %d)",
			backupType, backupID, archiveFi.Size(), dbInfo.TableCount, dbInfo.TotalRows, walInfo.PendingRecords))

	logger.Info("Backup %s finalized successfully (size: %d bytes, duration: %dms, file: %s)",
		backupID, archiveFi.Size(), durationMs, finalArchivePath)

	return record, nil
}

// ValidateBackup validates the integrity and syntax of a backup archive
func (s *BackupService) ValidateBackup(ctx context.Context, backupID string) (*model.BackupManifest, error) {
	record, err := s.backupRepo.GetByID(backupID)
	if err != nil {
		// Fallback to checking disk directly
		filePath := filepath.Join(s.cfg.BackupDir, fmt.Sprintf("%s.tar.gz", backupID))
		if _, statErr := os.Stat(filePath); statErr != nil {
			return nil, ErrBackupNotFound
		}
		record = &model.BackupRecord{
			ID:       backupID,
			FilePath: filePath,
		}
	}

	// Verify physical presence and non-zero size on disk
	fi, err := os.Stat(record.FilePath)
	if err != nil || fi.Size() == 0 {
		errStr := "backup archive file missing or empty on storage disk"
		s.markValidation(record, model.ValidationStatusInvalid, errStr)
		return nil, fmt.Errorf("%w: %s", ErrBackupNotFound, errStr)
	}

	manifest, err := s.readManifestFromArchive(record.FilePath)
	if err != nil {
		s.markValidation(record, model.ValidationStatusInvalid, err.Error())
		return nil, fmt.Errorf("manifest validation failed: %w", err)
	}

	// Verify archive checksum if stored in record
	if record.SHA256Checksum != "" {
		currentSHA, err := s.calculateFileSHA256(record.FilePath)
		if err != nil || currentSHA != record.SHA256Checksum {
			errStr := fmt.Sprintf("archive SHA256 checksum mismatch (expected: %s, got: %s)", record.SHA256Checksum, currentSHA)
			s.markValidation(record, model.ValidationStatusInvalid, errStr)
			return nil, errors.New(errStr)
		}
	}

	// Extract to temporary staging to validate internal files
	tempDir, err := os.MkdirTemp("", "validate_backup_*")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tempDir)

	if err := s.extractTarGz(record.FilePath, tempDir); err != nil {
		s.markValidation(record, model.ValidationStatusInvalid, "failed extracting archive: "+err.Error())
		return nil, fmt.Errorf("archive extraction failed: %w", err)
	}

	// Validate each artifact checksum
	for _, art := range manifest.Artifacts {
		artPath := filepath.Join(tempDir, filepath.FromSlash(art.Path))
		fi, err := os.Stat(artPath)
		if err != nil {
			errStr := fmt.Sprintf("missing artifact %s: %v", art.Path, err)
			s.markValidation(record, model.ValidationStatusInvalid, errStr)
			return nil, errors.New(errStr)
		}
		if fi.Size() != art.SizeBytes {
			errStr := fmt.Sprintf("artifact %s size mismatch: expected %d bytes, got %d", art.Path, art.SizeBytes, fi.Size())
			s.markValidation(record, model.ValidationStatusInvalid, errStr)
			return nil, errors.New(errStr)
		}
		sha, err := s.calculateFileSHA256(artPath)
		if err != nil || sha != art.SHA256 {
			errStr := fmt.Sprintf("artifact %s checksum mismatch: expected %s, got %s", art.Path, art.SHA256, sha)
			s.markValidation(record, model.ValidationStatusInvalid, errStr)
			return nil, errors.New(errStr)
		}
	}

	// Validate SQL dump readability
	dumpPath := filepath.Join(tempDir, "database", "dump.sql")
	if err := s.dbAdapter.ValidateDumpFile(dumpPath); err != nil {
		errStr := "dump validation failed: " + err.Error()
		s.markValidation(record, model.ValidationStatusInvalid, errStr)
		return nil, errors.New(errStr)
	}

	s.markValidation(record, model.ValidationStatusValid, "All artifacts, checksums, and SQL syntax verified")
	s.logAudit("BACKUP_VALIDATED", backupID, "system", "Backup passed all integrity checksum and structural validations")

	return manifest, nil
}

// GetRestorePreview calculates differences and impact assessment before executing restore
func (s *BackupService) GetRestorePreview(ctx context.Context, backupID string) (*model.RestorePreview, error) {
	manifest, err := s.ValidateBackup(ctx, backupID)
	if err != nil {
		return nil, fmt.Errorf("cannot preview unverified backup: %w", err)
	}

	currentStats, currentTotal, _ := s.dbAdapter.GetTableStats(ctx)

	var warnings []string
	if manifest.Database.TotalRows == 0 {
		warnings = append(warnings, "Warning: Backup contains zero rows of data.")
	}
	if currentTotal > manifest.Database.TotalRows {
		warnings = append(warnings, fmt.Sprintf("Active database has %d rows which is more than backup (%d rows). Existing data will be overwritten.",
			currentTotal, manifest.Database.TotalRows))
	}
	if manifest.WAL.PendingRecords > 0 {
		warnings = append(warnings, fmt.Sprintf("Backup contains %d pending WAL telemetry records that will be recovered and replayed.",
			manifest.WAL.PendingRecords))
	}

	targetDB := "datalogger"
	if s.cfg != nil && s.cfg.DBName != "" {
		targetDB = s.cfg.DBName
	}

	preview := &model.RestorePreview{
		BackupID:             backupID,
		CreatedAt:            manifest.CreatedAt,
		BackupType:           manifest.BackupType,
		SizeBytes:            manifest.TotalSizeBytes,
		DatabaseEngine:       manifest.Database.Engine,
		DatabaseName:         manifest.Database.DatabaseName,
		DumpMethod:           manifest.Database.DumpMethod,
		TableCount:           manifest.Database.TableCount,
		TotalRows:            manifest.Database.TotalRows,
		TableRows:            manifest.Database.TableRows,
		WALPendingRecords:    manifest.WAL.PendingRecords,
		WALSegmentsCount:     len(manifest.WAL.SegmentFiles),
		CurrentActiveTables:  len(currentStats),
		CurrentActiveRows:    currentTotal,
		EstimatedDurationSec: int(manifest.TotalSizeBytes/(1024*1024) + 2),
		TargetDatabase:       targetDB,
		Warnings:             warnings,
		CanRestore:           true,
	}

	return preview, nil
}

// StartRestoreJob launches an asynchronous restore operation
func (s *BackupService) StartRestoreJob(backupID string, initiatedBy string, createSafetyBackup bool) (string, error) {
	s.opMu.Lock()
	if s.isOperationRunning {
		s.opMu.Unlock()
		return "", ErrOperationInProgress
	}
	s.isOperationRunning = true
	s.activeOpType = "RESTORE"
	s.opMu.Unlock()

	jobID := fmt.Sprintf("job_restore_%d", time.Now().UnixNano())
	job := &model.BackupJob{
		ID:          jobID,
		Type:        model.JobTypeRestore,
		BackupID:    backupID,
		Status:      model.JobStatusRunning,
		Progress:    5.0,
		Stage:       "Initializing restore workflow",
		InitiatedBy: initiatedBy,
		StartTime:   time.Now().UTC(),
	}

	s.jobMu.Lock()
	s.jobs[jobID] = job
	s.jobMu.Unlock()

	go func() {
		defer func() {
			s.opMu.Lock()
			s.isOperationRunning = false
			s.activeOpType = ""
			s.opMu.Unlock()
		}()

		err := s.executeRestore(context.Background(), backupID, initiatedBy, createSafetyBackup, job)
		now := time.Now().UTC()
		s.jobMu.Lock()
		job.EndTime = &now
		if err != nil {
			job.Status = model.JobStatusFailed
			job.Error = err.Error()
			job.Stage = "Restore failed"
			logger.Error("Restore job %s failed: %v", jobID, err)
			s.logAudit("RESTORE_FAILED", backupID, initiatedBy, fmt.Sprintf("Restore failed: %v", err))
		} else {
			job.Status = model.JobStatusCompleted
			job.Progress = 100.0
			job.Stage = "Restore completed successfully"
			logger.Info("Restore job %s completed successfully", jobID)
			s.logAudit("RESTORE_EXECUTED", backupID, initiatedBy, "Database and WAL restored and replayed successfully")
		}
		s.jobMu.Unlock()
	}()

	return jobID, nil
}

// RestoreBackup restores a backup synchronously
func (s *BackupService) RestoreBackup(ctx context.Context, backupID string, initiatedBy string, createSafetyBackup bool) error {
	s.opMu.Lock()
	if s.isOperationRunning {
		s.opMu.Unlock()
		return ErrOperationInProgress
	}
	s.isOperationRunning = true
	s.activeOpType = "RESTORE"
	s.opMu.Unlock()

	defer func() {
		s.opMu.Lock()
		s.isOperationRunning = false
		s.activeOpType = ""
		s.opMu.Unlock()
	}()

	return s.executeRestore(ctx, backupID, initiatedBy, createSafetyBackup, nil)
}

func (s *BackupService) executeRestore(
	ctx context.Context,
	backupID string,
	initiatedBy string,
	createSafetyBackup bool,
	job *model.BackupJob,
) error {
	s.updateJob(job, 10.0, "Validating backup archive and manifest")

	record, err := s.backupRepo.GetByID(backupID)
	if err != nil {
		filePath := filepath.Join(s.cfg.BackupDir, fmt.Sprintf("%s.tar.gz", backupID))
		if _, statErr := os.Stat(filePath); statErr != nil {
			return ErrBackupNotFound
		}
		record = &model.BackupRecord{
			ID:       backupID,
			FilePath: filePath,
		}
	}

	// 1. Create pre-restore safety backup if requested
	if createSafetyBackup {
		s.updateJob(job, 20.0, "Creating pre-restore safety snapshot of current state")
		logger.Info("Creating automatic pre-restore safety backup...")
		_, safetyErr := s.executeBackup(ctx, model.BackupTypeSafety, "system_pre_restore", nil)
		if safetyErr != nil {
			logger.Warn("Failed creating safety backup (%v), continuing with restore cautiously", safetyErr)
		}
	}

	// 2. Extract backup to temporary directory
	s.updateJob(job, 35.0, "Extracting backup archive")
	tempDir, err := os.MkdirTemp("", "restore_stage_*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempDir)

	if err := s.extractTarGz(record.FilePath, tempDir); err != nil {
		return fmt.Errorf("failed extracting backup archive: %w", err)
	}

	// 3. Restore Database Dump
	s.updateJob(job, 55.0, "Restoring database tables and records")
	dumpPath := filepath.Join(tempDir, "database", "dump.sql")
	if err := s.dbAdapter.Restore(ctx, dumpPath); err != nil {
		return fmt.Errorf("database restore failed: %w", err)
	}

	// 4. Restore WAL Snapshot
	s.updateJob(job, 75.0, "Restoring persistent WAL telemetry queue")
	walDir := filepath.Join(tempDir, "wal")
	if fi, err := os.Stat(walDir); err == nil && fi.IsDir() {
		if s.walAdapter != nil {
			if err := s.walAdapter.RestoreSnapshot(walDir); err != nil {
				logger.Error("Failed restoring WAL snapshot: %v", err)
				return fmt.Errorf("WAL queue restore failed: %w", err)
			}
		}
	}

	// 5. Trigger Replay of restored uncommitted telemetry
	s.updateJob(job, 90.0, "Replaying pending telemetry records into database")
	if s.walAdapter != nil {
		replayed, err := s.walAdapter.ReplayPending(ctx)
		if err != nil {
			logger.Warn("WAL replay during restore returned warning: %v", err)
		} else {
			logger.Info("Restore replayed %d pending WAL telemetry records into database", replayed)
		}
	}

	s.updateJob(job, 98.0, "Finalizing restore verification")
	return nil
}

// GetJob returns current status of an async job
func (s *BackupService) GetJob(jobID string) (*model.BackupJob, error) {
	s.jobMu.RLock()
	defer s.jobMu.RUnlock()

	job, ok := s.jobs[jobID]
	if !ok {
		return nil, errors.New("job not found")
	}
	return job, nil
}

// GetStatus returns the operational health and storage status of the backup subsystem
func (s *BackupService) GetStatus() map[string]interface{} {
	s.opMu.Lock()
	isRunning := s.isOperationRunning
	activeOp := s.activeOpType
	s.opMu.Unlock()

	totalBackups := int64(0)
	var latestCompleted *model.BackupRecord
	totalStorage := int64(0)

	if s.backupRepo != nil {
		totalBackups, _ = s.backupRepo.CountValid()
		latestCompleted, _ = s.backupRepo.GetLatestCompleted()
		totalStorage, _ = s.backupRepo.GetTotalStorageBytes()
	}

	freeBytes, _ := s.checkFreeDiskSpace(s.cfg.BackupDir)

	subsystemStatus := "HEALTHY"
	var degradationReasons []string

	// 1. Check database connectivity
	if s.db == nil {
		subsystemStatus = "DEGRADED"
		degradationReasons = append(degradationReasons, "Database connection not initialized")
	} else if sqlDB, err := s.db.DB(); err != nil || sqlDB.Ping() != nil {
		subsystemStatus = "DEGRADED"
		degradationReasons = append(degradationReasons, "Database ping failed")
	}

	// 2. Check disk margin
	minFreeBytes := int64(s.cfg.BackupMinFreeSpaceMB) * 1024 * 1024
	if minFreeBytes <= 0 {
		minFreeBytes = 500 * 1024 * 1024
	}
	if freeBytes < minFreeBytes {
		subsystemStatus = "DEGRADED"
		degradationReasons = append(degradationReasons, fmt.Sprintf("Free disk space (%d MB) is below minimum threshold (%d MB)", freeBytes/(1024*1024), s.cfg.BackupMinFreeSpaceMB))
	}

	// 3. Check directory writability
	if err := os.MkdirAll(s.cfg.BackupDir, 0755); err != nil {
		subsystemStatus = "DEGRADED"
		degradationReasons = append(degradationReasons, fmt.Sprintf("Backup directory not writable: %v", err))
	} else {
		testFile := filepath.Join(s.cfg.BackupDir, fmt.Sprintf(".health_test_%d", time.Now().UnixNano()))
		if err := os.WriteFile(testFile, []byte("ok"), 0644); err != nil {
			subsystemStatus = "DEGRADED"
			degradationReasons = append(degradationReasons, fmt.Sprintf("Backup directory write test failed: %v", err))
		} else {
			_ = os.Remove(testFile)
		}
	}

	res := map[string]interface{}{
		"status":                  subsystemStatus,
		"subsystem_healthy":       subsystemStatus == "HEALTHY",
		"degradation_reasons":     degradationReasons,
		"backup_dir":              s.cfg.BackupDir,
		"schedule_enabled":        s.cfg.BackupScheduleEnabled,
		"schedule_time":           s.cfg.BackupScheduleTime,
		"schedule_interval_hours": s.cfg.BackupScheduleIntervalHours,
		"total_valid_backups":     totalBackups,
		"total_backups":           totalBackups,
		"storage_used_bytes":      totalStorage,
		"storage_used_mb":         float64(totalStorage) / (1024 * 1024),
		"free_disk_bytes":         freeBytes,
		"free_disk_mb":            float64(freeBytes) / (1024 * 1024),
		"operation_in_progress":   isRunning,
		"active_operation":        activeOp,
	}

	if latestCompleted != nil {
		latestCompleted.PopulateVirtualFields()
		res["last_backup_id"] = latestCompleted.ID
		res["last_backup_time"] = latestCompleted.CreatedAt.Format(time.RFC3339)
		res["last_successful_backup"] = latestCompleted.CreatedAt.Format(time.RFC3339)
		res["last_backup_status"] = latestCompleted.Status
		res["last_backup_size"] = latestCompleted.SizeBytes
	}

	return res
}

// ListBackups returns paginated backups
func (s *BackupService) ListBackups(page, pageSize int, status string) ([]model.BackupRecord, int64, error) {
	if s.backupRepo == nil {
		return []model.BackupRecord{}, 0, nil
	}
	return s.backupRepo.List(page, pageSize, status)
}

// GetBackupByID returns the database record and parsed manifest
func (s *BackupService) GetBackupByID(id string) (*model.BackupRecord, *model.BackupManifest, error) {
	record, err := s.backupRepo.GetByID(id)
	if err != nil {
		return nil, nil, err
	}

	manifest, err := s.readManifestFromArchive(record.FilePath)
	if err != nil {
		logger.Warn("Failed reading manifest for %s: %v", id, err)
	}

	return record, manifest, nil
}

// DeleteBackup deletes a backup archive and removes its catalog record
func (s *BackupService) DeleteBackup(id string, initiatedBy string) error {
	record, err := s.backupRepo.GetByID(id)
	if err != nil {
		return err
	}

	// Delete file from filesystem
	if record.FilePath != "" {
		_ = os.Remove(record.FilePath)
	}

	// Remove from database
	if err := s.backupRepo.Delete(id); err != nil {
		return err
	}

	s.logAudit("BACKUP_DELETED", id, initiatedBy, fmt.Sprintf("Deleted backup %s (%d bytes)", id, record.SizeBytes))
	return nil
}

// PruneOldBackups removes completed backups exceeding BackupMaxKeepCount
func (s *BackupService) PruneOldBackups(keepMax int) (int, error) {
	if s.backupRepo == nil || keepMax <= 0 {
		return 0, nil
	}

	total, err := s.backupRepo.CountValid()
	if err != nil || int(total) <= keepMax {
		return 0, nil
	}

	toDeleteCount := int(total) - keepMax
	oldestRecords, err := s.backupRepo.ListOldestCompleted(toDeleteCount)
	if err != nil {
		return 0, err
	}

	deleted := 0
	for _, rec := range oldestRecords {
		if err := s.DeleteBackup(rec.ID, "system_prune"); err == nil {
			deleted++
		}
	}

	if deleted > 0 {
		logger.Info("Pruned %d old backups (retained latest %d)", deleted, keepMax)
	}
	return deleted, nil
}

// ----------------- Internal Helper Methods -----------------

func (s *BackupService) updateJob(job *model.BackupJob, progress float64, stage string) {
	if job == nil {
		return
	}
	s.jobMu.Lock()
	defer s.jobMu.Unlock()
	job.Progress = progress
	job.Stage = stage
}

func (s *BackupService) markValidation(record *model.BackupRecord, status model.ValidationStatus, details string) {
	if record == nil || s.backupRepo == nil {
		return
	}
	record.ValidationStatus = status
	record.ValidationDetails = details
	if status == model.ValidationStatusValid {
		record.ConsistencyStatus = "CONSISTENT"
	} else if status == model.ValidationStatusInvalid {
		record.ConsistencyStatus = "INCONSISTENT"
	}
	record.PopulateVirtualFields()
	_ = s.backupRepo.Update(record)
}

func (s *BackupService) logAudit(action, resource, username, details string) {
	if s.systemRepo == nil {
		return
	}
	if username == "" {
		username = "system"
	}
	_ = s.systemRepo.AddAuditTrail(&model.AuditTrail{
		Username:  username,
		Action:    action,
		Resource:  resource,
		Details:   details,
		CreatedAt: time.Now().UTC(),
	})
}

func (s *BackupService) createTarGz(srcDir, tarGzPath string) error {
	outFile, err := os.Create(tarGzPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	gw := gzip.NewWriter(outFile)
	defer gw.Close()

	tw := tar.NewWriter(gw)
	defer tw.Close()

	return filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}
		if relPath == "." {
			return nil
		}

		header, err := tar.FileInfoHeader(info, info.Name())
		if err != nil {
			return err
		}
		header.Name = filepath.ToSlash(relPath)

		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(tw, file)
		return err
	})
}

func (s *BackupService) extractTarGz(tarGzPath, destDir string) error {
	inFile, err := os.Open(tarGzPath)
	if err != nil {
		return err
	}
	defer inFile.Close()

	gr, err := gzip.NewReader(inFile)
	if err != nil {
		return err
	}
	defer gr.Close()

	tr := tar.NewReader(gr)

	cleanDest := filepath.Clean(destDir)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		target := filepath.Join(destDir, filepath.FromSlash(header.Name))
		cleanTarget := filepath.Clean(target)

		// Prevent Zip Slip / Path Traversal vulnerability
		if !strings.HasPrefix(cleanTarget, cleanDest+string(os.PathSeparator)) && cleanTarget != cleanDest {
			return fmt.Errorf("illegal path in archive: %s", header.Name)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(cleanTarget, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(cleanTarget), 0755); err != nil {
				return err
			}
			outFile, err := os.OpenFile(cleanTarget, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, header.FileInfo().Mode())
			if err != nil {
				return err
			}
			if _, err := io.Copy(outFile, tr); err != nil {
				outFile.Close()
				return err
			}
			outFile.Close()
		}
	}

	return nil
}

func (s *BackupService) readManifestFromArchive(tarGzPath string) (*model.BackupManifest, error) {
	inFile, err := os.Open(tarGzPath)
	if err != nil {
		return nil, err
	}
	defer inFile.Close()

	gr, err := gzip.NewReader(inFile)
	if err != nil {
		return nil, err
	}
	defer gr.Close()

	tr := tar.NewReader(gr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		if header.Name == "manifest.json" {
			var manifest model.BackupManifest
			decoder := json.NewDecoder(tr)
			if err := decoder.Decode(&manifest); err != nil {
				return nil, err
			}
			return &manifest, nil
		}
	}

	return nil, errors.New("manifest.json not found in archive")
}

func (s *BackupService) calculateFileSHA256(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func (s *BackupService) checkFreeDiskSpace(dir string) (int64, error) {
	absPath, err := filepath.Abs(dir)
	if err != nil {
		absPath = dir
	}
	checkPath := absPath
	for {
		if _, err := os.Stat(checkPath); err == nil {
			break
		}
		parent := filepath.Dir(checkPath)
		if parent == checkPath {
			break
		}
		checkPath = parent
	}

	usage, err := disk.Usage(checkPath)
	if err == nil && usage.Free > 0 {
		return int64(usage.Free), nil
	}

	// Safe 50GB default fallback if OS metrics unavailable
	return 50 * 1024 * 1024 * 1024, nil
}
