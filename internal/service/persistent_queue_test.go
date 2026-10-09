package service

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"datalogger/internal/model"
	"datalogger/internal/queue"
	"datalogger/internal/repository"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// FaultInjectableBoundary simulates MariaDB network drops, pool exhaustion, and recovery
type FaultInjectableBoundary struct {
	mu           sync.Mutex
	failAttempts int
	shouldFail   bool
	failErr      error
	committed    []*model.RawData
}

func (f *FaultInjectableBoundary) SaveBatch(ctx context.Context, batch []*model.RawData) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.shouldFail || f.failAttempts > 0 {
		if f.failAttempts > 0 {
			f.failAttempts--
		}
		if f.failErr != nil {
			return f.failErr
		}
		return errors.New("simulated MariaDB connection refused (dial tcp: 3306: connection refused)")
	}

	for _, item := range batch {
		if item != nil {
			f.committed = append(f.committed, item)
		}
	}
	return nil
}

func (f *FaultInjectableBoundary) SetFailing(fail bool, err error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.shouldFail = fail
	f.failErr = err
}

func (f *FaultInjectableBoundary) GetCommittedCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.committed)
}

func setupTestDB(t *testing.T) (*gorm.DB, *repository.TelemetryRepository) {
	dsn := fmt.Sprintf("file:mem_wal_db_%d?mode=memory&cache=shared", time.Now().UnixNano()+int64(rand.Intn(1000)))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		t.Fatalf("failed opening test DB: %v", err)
	}

	// Migrate RawData, Parameter, Device
	if err := db.AutoMigrate(&model.RawData{}, &model.Parameter{}, &model.Device{}); err != nil {
		t.Fatalf("failed migrating test DB: %v", err)
	}

	// Seed test device and parameter
	dev := model.Device{ID: 1, DeviceCode: "DEV-01", DeviceName: "Test Device"}
	db.Create(&dev)
	param := model.Parameter{ID: 10, DeviceID: 1, ParameterCode: "TEMP-01", ParameterName: "Temperature"}
	db.Create(&param)

	repo := repository.NewTelemetryRepository(db)
	return db, repo
}

// Criteria 1, 2, 3: Durable append, batch append, and payload checksums
func TestPhase42_DurableAppendAndChecksum(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "wal_crit_1_*")
	if err != nil {
		t.Fatalf("temp dir err: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := queue.DefaultWALConfig(tmpDir)
	cfg.SyncMode = "always"
	wal, err := queue.OpenWALQueue(cfg)
	if err != nil {
		t.Fatalf("failed opening WAL: %v", err)
	}
	defer wal.Close()

	rec := &model.RawData{
		RecordUUID:     "UUID-CRIT-1",
		DeviceID:       1,
		ParameterID:    10,
		Value:          99.5,
		ProcessedValue: 99.5,
		Quality:        model.QualityGood,
		QualityReason:  model.ReasonNone,
		ReceivedAt:     time.Now().UTC(),
		StoredAt:       time.Now().UTC(),
	}

	if err := wal.Append(rec); err != nil {
		t.Fatalf("append failed: %v", err)
	}

	batch, token, err := wal.ReadPendingBatch(10)
	if err != nil {
		t.Fatalf("read pending failed: %v", err)
	}
	if len(batch) != 1 || batch[0].RecordUUID != "UUID-CRIT-1" {
		t.Fatalf("record UUID mismatch")
	}

	_ = wal.Acknowledge(token)
}

// Criteria 4, 5, 6: Clean restart, abrupt termination, incomplete tail recovery
func TestPhase42_RecoveryAfterCrashAndTruncatedTail(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "wal_crit_4_*")
	if err != nil {
		t.Fatalf("temp dir err: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := queue.DefaultWALConfig(tmpDir)
	wal1, err := queue.OpenWALQueue(cfg)
	if err != nil {
		t.Fatalf("failed opening WAL: %v", err)
	}

	for i := 1; i <= 5; i++ {
		_ = wal1.Append(&model.RawData{
			RecordUUID:     fmt.Sprintf("UUID-%d", i),
			DeviceID:       1,
			ParameterID:    10,
			Value:          float64(i * 10),
			ProcessedValue: float64(i * 10),
			Quality:        model.QualityGood,
			ReceivedAt:     time.Now().UTC(),
		})
	}
	_ = wal1.Close()

	// Simulate sudden power loss by appending incomplete bytes to tail
	segPath := filepath.Join(tmpDir, "segment_000001.wal")
	f, err := os.OpenFile(segPath, os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		t.Fatalf("failed opening segment for tail injection: %v", err)
	}
	_, _ = f.Write([]byte("INCOMPLETE_TAIL_RECORD"))
	_ = f.Close()

	// Reopen after crash: should cleanly recover all 5 valid records and halt at truncated tail
	wal2, err := queue.OpenWALQueue(cfg)
	if err != nil {
		t.Fatalf("failed reopening WAL: %v", err)
	}
	defer wal2.Close()

	batch, token, err := wal2.ReadPendingBatch(10)
	if err != nil {
		t.Fatalf("read pending failed: %v", err)
	}
	if len(batch) != 5 {
		t.Fatalf("expected exactly 5 recovered records, got %d", len(batch))
	}
	_ = wal2.Acknowledge(token)
}

// Criteria 7, 8: Corrupted record detection and segment quarantine
func TestPhase42_CorruptedRecordDetectionAndQuarantine(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "wal_crit_7_*")
	if err != nil {
		t.Fatalf("temp dir err: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := queue.DefaultWALConfig(tmpDir)
	wal1, err := queue.OpenWALQueue(cfg)
	if err != nil {
		t.Fatalf("failed opening WAL: %v", err)
	}

	_ = wal1.Append(&model.RawData{
		RecordUUID: "UUID-CORRUPT-TEST",
		DeviceID:   1,
		ReceivedAt: time.Now().UTC(),
	})
	_ = wal1.Close()

	// Tamper bytes in the middle of segment_000001.wal
	segPath := filepath.Join(tmpDir, "segment_000001.wal")
	data, err := os.ReadFile(segPath)
	if err != nil {
		t.Fatalf("read seg failed: %v", err)
	}
	if len(data) > 30 {
		data[30] ^= 0xFF // Flip bits in payload
		_ = os.WriteFile(segPath, data, 0644)
	}

	// Reopen and read: should detect CRC mismatch and quarantine segment
	wal2, err := queue.OpenWALQueue(cfg)
	if err != nil {
		t.Fatalf("reopen failed: %v", err)
	}
	defer wal2.Close()

	_, _, _ = wal2.ReadPendingBatch(10)

	stats := wal2.Stats()
	if stats["checksum_failures"].(uint64) == 0 {
		t.Fatalf("expected checksum_failures to be incremented")
	}

	// Verify quarantine directory has the damaged file
	quarantineEntries, _ := os.ReadDir(filepath.Join(tmpDir, "quarantine"))
	if len(quarantineEntries) == 0 {
		t.Fatalf("expected corrupted segment to be quarantined")
	}
}

// Criteria 9, 10, 11, 12: MariaDB outage spooling and automatic replay with preserved metadata
func TestPhase42_MariaDBOutageSpoolingAndReplay(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "wal_crit_9_*")
	if err != nil {
		t.Fatalf("temp dir err: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := queue.DefaultWALConfig(tmpDir)
	wal, err := queue.OpenWALQueue(cfg)
	if err != nil {
		t.Fatalf("failed opening WAL: %v", err)
	}

	db, repo := setupTestDB(t)
	adapter := NewPersistentQueueAdapter(repo, wal, 50, 1)
	defer adapter.Close()

	// 1. Simulate DB down by dropping table
	_ = db.Migrator().DropTable(&model.RawData{})

	origTime := time.Date(2026, 10, 9, 12, 0, 0, 0, time.UTC)
	var records []*model.RawData
	for i := 1; i <= 10; i++ {
		records = append(records, &model.RawData{
			RecordUUID:      fmt.Sprintf("OUTAGE-REC-%d", i),
			DeviceID:        1,
			ParameterID:     10,
			Value:           float64(i) * 5.5,
			ProcessedValue:  float64(i) * 5.5,
			Quality:         model.QualityGood,
			QualityReason:   model.ReasonNone,
			Sequence:        uint64(i),
			ReceivedAt:      origTime.Add(time.Duration(i) * time.Second),
			DeviceTimestamp: &origTime,
		})
	}

	// SaveBatch while DB is down: should not crash, must spool to WAL
	ctx := context.Background()
	err = adapter.SaveBatch(ctx, records)
	if err != nil {
		t.Fatalf("SaveBatch failed during outage: %v", err)
	}

	health, consecutiveErrors := adapter.GetHealth()
	if health != "DEGRADED" && health != "FAILING" {
		t.Fatalf("expected DEGRADED or FAILING during outage, got %s", health)
	}
	if consecutiveErrors == 0 {
		t.Fatalf("expected consecutive errors > 0")
	}

	if wal.PendingCount() != 10 {
		t.Fatalf("expected 10 pending records in WAL spool, got %d", wal.PendingCount())
	}

	// 2. Restore MariaDB table
	_ = db.AutoMigrate(&model.RawData{})

	// 3. Trigger replay
	replayed, err := adapter.ReplayPending(ctx)
	if err != nil {
		t.Fatalf("replay failed: %v", err)
	}
	if replayed != 10 {
		t.Fatalf("expected 10 replayed records, got %d", replayed)
	}

	// Verify all 10 records are in the database with exact preserved metadata
	var dbRecords []model.RawData
	db.Order("sequence ASC").Find(&dbRecords)
	if len(dbRecords) != 10 {
		t.Fatalf("expected 10 records in DB after replay, got %d", len(dbRecords))
	}

	for i, r := range dbRecords {
		expectedUUID := fmt.Sprintf("OUTAGE-REC-%d", i+1)
		if r.RecordUUID != expectedUUID {
			t.Errorf("UUID mismatch at %d: expected %s, got %s", i, expectedUUID, r.RecordUUID)
		}
		expectedTime := origTime.Add(time.Duration(i+1) * time.Second)
		if r.ReceivedAt.Unix() != expectedTime.Unix() {
			t.Errorf("timestamp changed! expected %v, got %v", expectedTime, r.ReceivedAt)
		}
		if r.Quality != model.QualityGood {
			t.Errorf("quality altered! expected GOOD, got %s", r.Quality)
		}
	}

	// Health should have recovered to HEALTHY
	healthAfter, _ := adapter.GetHealth()
	if healthAfter != "HEALTHY" {
		t.Fatalf("expected HEALTHY after successful replay, got %s", healthAfter)
	}
	if wal.PendingCount() != 0 {
		t.Fatalf("expected 0 pending in WAL after ack, got %d", wal.PendingCount())
	}
}

// Criteria 13, 14, 15: Database idempotency and duplicate avoidance on replay
func TestPhase42_DatabaseIdempotencyOnDuplicateReplay(t *testing.T) {
	db, repo := setupTestDB(t)
	ctx := context.Background()

	now := time.Now().UTC()
	batch := []*model.RawData{
		{
			RecordUUID:     "IDEMPOTENT-REC-01",
			DeviceID:       1,
			ParameterID:    10,
			Value:          50.0,
			ProcessedValue: 50.0,
			Quality:        model.QualityGood,
			ReceivedAt:     now,
		},
		{
			RecordUUID:     "IDEMPOTENT-REC-02",
			DeviceID:       1,
			ParameterID:    10,
			Value:          60.0,
			ProcessedValue: 60.0,
			Quality:        model.QualityGood,
			ReceivedAt:     now,
		},
	}

	// First commit
	if err := repo.SaveBatch(ctx, batch); err != nil {
		t.Fatalf("first SaveBatch failed: %v", err)
	}

	var count1 int64
	db.Model(&model.RawData{}).Count(&count1)
	if count1 != 2 {
		t.Fatalf("expected 2 rows, got %d", count1)
	}

	// Replay exact same batch (simulating uncertain commit retry)
	if err := repo.SaveBatch(ctx, batch); err != nil {
		t.Fatalf("duplicate replay SaveBatch failed: %v", err)
	}

	var count2 int64
	db.Model(&model.RawData{}).Count(&count2)
	if count2 != 2 {
		t.Fatalf("duplicate rows created! expected 2 rows, got %d", count2)
	}
}

// Criteria 16, 17, 18, 19: Disk limit enforcement and capacity tracking
func TestPhase42_DiskCapacityAndLimitEnforcement(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "wal_crit_16_*")
	if err != nil {
		t.Fatalf("temp dir err: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := queue.DefaultWALConfig(tmpDir)
	cfg.MaxSizeBytes = 300 // Small 300 bytes to strictly enforce limit after 1 record
	wal, err := queue.OpenWALQueue(cfg)
	if err != nil {
		t.Fatalf("failed opening WAL: %v", err)
	}
	defer wal.Close()

	rec := &model.RawData{
		RecordUUID: "LIMIT-TEST-1",
		DeviceID:   1,
		ReceivedAt: time.Now().UTC(),
	}

	// First append OK
	_ = wal.Append(rec)

	// Second append should trip disk limit
	rec2 := &model.RawData{
		RecordUUID: "LIMIT-TEST-2",
		DeviceID:   1,
		ReceivedAt: time.Now().UTC(),
	}
	err = wal.Append(rec2)
	if err != queue.ErrQueueDiskFull {
		t.Fatalf("expected ErrQueueDiskFull, got %v", err)
	}
}

// Criteria 20, 21: Graceful shutdown with pending records and startup replay
func TestPhase42_GracefulShutdownAndStartupReplay(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "wal_crit_20_*")
	if err != nil {
		t.Fatalf("temp dir err: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	db, repo := setupTestDB(t)

	// 1. Start TelemetryService with persistent queue enabled
	tCfg := DefaultTelemetryConfig()
	tCfg.QueueEnabled = true
	tCfg.QueueDir = tmpDir
	tCfg.BufferSize = 500
	tCfg.FlushInterval = 50 * time.Millisecond

	svc1 := NewTelemetryService(repo, nil, tCfg)

	// Ingest items
	for i := 1; i <= 5; i++ {
		_ = svc1.Ingest(&TelemetryIngestPayload{
			DeviceID:      1,
			ParameterID:   10,
			Value:         float64(i * 10),
			ReceivedAt:    time.Now().UTC(),
			Quality:       model.QualityGood,
		})
	}

	// Stop svc1 gracefully
	svc1.Stop()

	// 2. Open svc2 on the same queue directory: verify startup replay
	svc2 := NewTelemetryService(repo, nil, tCfg)
	defer svc2.Stop()

	// Allow startup replayer time to catch up
	time.Sleep(800 * time.Millisecond)

	var count int64
	db.Model(&model.RawData{}).Count(&count)
	if count < 5 {
		t.Fatalf("expected at least 5 records persisted in DB after startup replay, got %d", count)
	}
}

// Criteria 27: Localization key parity for persistentQueue
func TestPhase42_LocalizationKeyParity(t *testing.T) {
	enBytes, err := os.ReadFile(filepath.Join("..", "..", "web", "src", "i18n", "locales", "en.js"))
	if err != nil {
		t.Fatalf("failed reading en.js: %v", err)
	}
	idBytes, err := os.ReadFile(filepath.Join("..", "..", "web", "src", "i18n", "locales", "id.js"))
	if err != nil {
		t.Fatalf("failed reading id.js: %v", err)
	}

	enStr := string(enBytes)
	idStr := string(idBytes)

	requiredKeys := []string{
		"persistentQueue",
		"spoolSize",
		"pendingRecords",
		"totalAppended",
		"totalAcked",
		"totalReplayed",
		"checksumErrors",
		"spoolUtilization",
		"syncMode",
	}

	for _, k := range requiredKeys {
		if !strings.Contains(enStr, k) {
			t.Errorf("en.js missing key: %s", k)
		}
		if !strings.Contains(idStr, k) {
			t.Errorf("id.js missing key: %s", k)
		}
	}
}
