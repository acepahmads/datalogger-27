package queue

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"datalogger/internal/model"
)

func createTestRecord(id uint, seq uint64, val float64) *model.RawData {
	now := time.Now().UTC()
	return &model.RawData{
		RecordUUID:     fmt.Sprintf("TEST-REC-%d-%d", id, seq),
		DeviceID:       10,
		ParameterID:    id,
		Value:          val,
		ProcessedValue: val,
		Quality:        model.QualityGood,
		QualityReason:  model.ReasonNone,
		Sequence:       seq,
		ReceivedAt:     now,
		StoredAt:       now,
		Timestamp:      now,
	}
}

func TestWALQueue_AppendAndRead(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "wal_test_append_*")
	if err != nil {
		t.Fatalf("failed creating temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := DefaultWALConfig(tmpDir)
	q, err := OpenWALQueue(cfg)
	if err != nil {
		t.Fatalf("failed opening WAL: %v", err)
	}
	defer q.Close()

	rec1 := createTestRecord(1, 101, 23.5)
	rec2 := createTestRecord(2, 102, 45.8)

	if err := q.Append(rec1); err != nil {
		t.Fatalf("failed appending rec1: %v", err)
	}
	if err := q.Append(rec2); err != nil {
		t.Fatalf("failed appending rec2: %v", err)
	}

	batch, token, err := q.ReadPendingBatch(10)
	if err != nil {
		t.Fatalf("failed reading batch: %v", err)
	}
	if len(batch) != 2 {
		t.Fatalf("expected 2 records, got %d", len(batch))
	}
	if batch[0].RecordUUID != rec1.RecordUUID || batch[1].RecordUUID != rec2.RecordUUID {
		t.Fatalf("mismatch in record UUIDs")
	}

	// Acknowledge batch
	if err := q.Acknowledge(token); err != nil {
		t.Fatalf("failed acknowledging batch: %v", err)
	}

	// Subsequent read should return empty batch
	batch2, _, err := q.ReadPendingBatch(10)
	if err != nil {
		t.Fatalf("failed reading batch2: %v", err)
	}
	if len(batch2) != 0 {
		t.Fatalf("expected 0 pending records after ack, got %d", len(batch2))
	}
}

func TestWALQueue_BatchAppendAndPurge(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "wal_test_batch_*")
	if err != nil {
		t.Fatalf("failed creating temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := DefaultWALConfig(tmpDir)
	cfg.MaxSegmentSize = 1024 // Small 1KB segment size to trigger rotation
	q, err := OpenWALQueue(cfg)
	if err != nil {
		t.Fatalf("failed opening WAL: %v", err)
	}
	defer q.Close()

	var records []*model.RawData
	for i := 1; i <= 20; i++ {
		records = append(records, createTestRecord(uint(i), uint64(i), float64(i)*1.5))
	}

	if err := q.AppendBatch(records); err != nil {
		t.Fatalf("failed batch append: %v", err)
	}

	// Should have rotated across multiple segments
	segments, err := q.listSegments()
	if err != nil {
		t.Fatalf("failed listing segments: %v", err)
	}
	if len(segments) <= 1 {
		t.Fatalf("expected multiple segments rotated, got %d", len(segments))
	}

	// Read in two batches
	batch1, token1, err := q.ReadPendingBatch(10)
	if err != nil || len(batch1) != 10 {
		t.Fatalf("expected 10 items in batch1, got %d (err: %v)", len(batch1), err)
	}
	if err := q.Acknowledge(token1); err != nil {
		t.Fatalf("failed ack batch1: %v", err)
	}

	batch2, token2, err := q.ReadPendingBatch(15)
	if err != nil || len(batch2) != 10 {
		t.Fatalf("expected 10 items in batch2, got %d (err: %v)", len(batch2), err)
	}
	if err := q.Acknowledge(token2); err != nil {
		t.Fatalf("failed ack batch2: %v", err)
	}
}

func TestWALQueue_RestartRecovery(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "wal_test_restart_*")
	if err != nil {
		t.Fatalf("failed creating temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := DefaultWALConfig(tmpDir)
	q1, err := OpenWALQueue(cfg)
	if err != nil {
		t.Fatalf("failed opening WAL: %v", err)
	}

	rec1 := createTestRecord(1, 1, 10.0)
	rec2 := createTestRecord(2, 2, 20.0)
	_ = q1.Append(rec1)
	_ = q1.Append(rec2)
	_ = q1.Close() // Simulate application restart/shutdown

	// Re-open
	q2, err := OpenWALQueue(cfg)
	if err != nil {
		t.Fatalf("failed re-opening WAL: %v", err)
	}
	defer q2.Close()

	batch, token, err := q2.ReadPendingBatch(10)
	if err != nil {
		t.Fatalf("failed reading recovered batch: %v", err)
	}
	if len(batch) != 2 {
		t.Fatalf("expected 2 recovered records, got %d", len(batch))
	}
	if batch[0].Value != 10.0 || batch[1].Value != 20.0 {
		t.Fatalf("recovered values mismatch")
	}

	_ = q2.Acknowledge(token)
}

func TestWALQueue_IncompleteTailRecovery(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "wal_test_tail_*")
	if err != nil {
		t.Fatalf("failed creating temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := DefaultWALConfig(tmpDir)
	q1, err := OpenWALQueue(cfg)
	if err != nil {
		t.Fatalf("failed opening WAL: %v", err)
	}

	rec1 := createTestRecord(1, 1, 10.0)
	_ = q1.Append(rec1)
	_ = q1.Close()

	// Simulate sudden power loss by appending partial/corrupted 10 bytes to the active segment file
	segPath := filepath.Join(tmpDir, "segment_000001.wal")
	f, err := os.OpenFile(segPath, os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		t.Fatalf("failed opening segment for tail injection: %v", err)
	}
	_, _ = f.Write([]byte("TRUNCATED_"))
	_ = f.Close()

	// Re-open: WAL should gracefully recover rec1 and ignore/halt at truncated tail
	q2, err := OpenWALQueue(cfg)
	if err != nil {
		t.Fatalf("failed re-opening WAL after crash: %v", err)
	}
	defer q2.Close()

	batch, _, err := q2.ReadPendingBatch(10)
	if err != nil {
		t.Fatalf("read pending failed: %v", err)
	}
	if len(batch) != 1 {
		t.Fatalf("expected 1 valid record recovered before truncated tail, got %d", len(batch))
	}
}

func TestWALQueue_DiskLimitEnforcement(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "wal_test_limit_*")
	if err != nil {
		t.Fatalf("failed creating temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := DefaultWALConfig(tmpDir)
	cfg.MaxSizeBytes = 500 // Extremely small 500 bytes limit
	q, err := OpenWALQueue(cfg)
	if err != nil {
		t.Fatalf("failed opening WAL: %v", err)
	}
	defer q.Close()

	rec1 := createTestRecord(1, 1, 10.0)
	rec2 := createTestRecord(2, 2, 20.0)

	_ = q.Append(rec1)
	err = q.Append(rec2)
	if err != ErrQueueDiskFull {
		t.Fatalf("expected ErrQueueDiskFull, got: %v", err)
	}
}
