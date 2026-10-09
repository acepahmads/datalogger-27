package queue

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"hash/crc32"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"datalogger/internal/logger"
	"datalogger/internal/model"
)

var (
	// WAL magic bytes: "TLM1" (Telemetry Log Version 1)
	walMagic = [4]byte{'T', 'L', 'M', '1'}

	ErrQueueDiskFull    = errors.New("persistent queue disk limit exceeded")
	ErrQueueCorrupt     = errors.New("persistent queue record checksum mismatch or corruption")
	ErrInvalidBatchToken = errors.New("invalid batch acknowledgement token")
)

const (
	headerSize        = 28 // 4 (magic) + 4 (len) + 4 (crc32) + 8 (seq) + 8 (timestamp_nano)
	defaultMaxSegment = 10 * 1024 * 1024 // 10 MB per WAL segment
)

// WALConfig holds configuration options for the persistent Write-Ahead Log queue
type WALConfig struct {
	Dir               string
	MaxSizeBytes      int64
	MaxSegmentSize    int64
	SyncMode          string // "always", "batch", "none"
	DiskWarnPercent   float64
}

// DefaultWALConfig returns sensible defaults for edge devices
func DefaultWALConfig(dir string) WALConfig {
	if dir == "" {
		dir = filepath.Join("data", "queue")
	}
	return WALConfig{
		Dir:             dir,
		MaxSizeBytes:    100 * 1024 * 1024, // 100 MB
		MaxSegmentSize:  defaultMaxSegment,
		SyncMode:        "batch",
		DiskWarnPercent: 80.0,
	}
}

// Checkpoint represents the persistent committed state of the WAL
type Checkpoint struct {
	SegmentIndex uint32    `json:"segment_index"`
	Offset       int64     `json:"offset"`
	TotalAcked   uint64    `json:"total_acked"`
	LastAckedAt  time.Time `json:"last_acked_at"`
}

// BatchToken represents an uncommitted batch position returned to the caller
type BatchToken struct {
	SegmentIndex uint32
	StartOffset  int64
	EndOffset    int64
	Count        int
	MaxTimestamp time.Time
}

// WALQueue implements an append-only, checksummed, segmented persistent spool
type WALQueue struct {
	cfg WALConfig

	mu           sync.Mutex // Protects write path and segment rotation
	activeFile   *os.File
	activeIndex  uint32
	activeOffset int64

	readMu       sync.Mutex // Protects read path and checkpoints
	checkpoint   Checkpoint

	// Statistics & Health
	totalAppended     uint64
	totalAcked        uint64
	totalReplayed     uint64
	checksumFailures  uint64
	corruptedSegments uint64
	currentSpoolSize  int64
	oldestPendingTime time.Time

	closed int32
}

// OpenWALQueue opens an existing WAL queue directory or initializes a fresh one
func OpenWALQueue(cfg WALConfig) (*WALQueue, error) {
	if cfg.MaxSegmentSize <= 0 {
		cfg.MaxSegmentSize = defaultMaxSegment
	}
	if cfg.MaxSizeBytes <= 0 {
		cfg.MaxSizeBytes = 100 * 1024 * 1024
	}
	if cfg.SyncMode == "" {
		cfg.SyncMode = "batch"
	}

	if err := os.MkdirAll(cfg.Dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create WAL queue directory: %w", err)
	}
	_ = os.MkdirAll(filepath.Join(cfg.Dir, "quarantine"), 0755)

	q := &WALQueue{
		cfg: cfg,
	}

	// 1. Load checkpoint
	if err := q.loadCheckpoint(); err != nil {
		logger.Warn("Failed to load WAL checkpoint, starting from beginning: %v", err)
		q.checkpoint = Checkpoint{
			SegmentIndex: 1,
			Offset:       0,
			LastAckedAt:  time.Now().UTC(),
		}
	}

	// 2. Scan existing segments
	segments, err := q.listSegments()
	if err != nil {
		return nil, fmt.Errorf("failed listing WAL segments: %w", err)
	}

	if len(segments) == 0 {
		// Fresh queue
		q.activeIndex = 1
		if q.checkpoint.SegmentIndex == 0 {
			q.checkpoint.SegmentIndex = 1
		}
	} else {
		q.activeIndex = segments[len(segments)-1]
		if q.checkpoint.SegmentIndex == 0 {
			q.checkpoint.SegmentIndex = segments[0]
		}
	}

	// 3. Open active segment for appending
	if err := q.openActiveSegmentForAppend(); err != nil {
		return nil, fmt.Errorf("failed opening active WAL segment: %w", err)
	}

	// 4. Update initial spool size and pending metrics
	q.updateMetrics()

	logger.Info("Persistent WAL Queue opened at %s (Active Segment: %d, Acked Segment: %d at offset %d, Spool Size: %d bytes)",
		q.cfg.Dir, q.activeIndex, q.checkpoint.SegmentIndex, q.checkpoint.Offset, q.currentSpoolSize)

	return q, nil
}

// Close flushes and safely closes open file handles
func (q *WALQueue) Close() error {
	if !atomic.CompareAndSwapInt32(&q.closed, 0, 1) {
		return nil
	}

	q.mu.Lock()
	defer q.mu.Unlock()

	var err error
	if q.activeFile != nil {
		_ = q.activeFile.Sync()
		err = q.activeFile.Close()
		q.activeFile = nil
	}
	return err
}

// Append persists a telemetry record to the active WAL segment with checksum
func (q *WALQueue) Append(record *model.RawData) error {
	if atomic.LoadInt32(&q.closed) == 1 {
		return errors.New("WAL queue is closed")
	}
	if record == nil {
		return errors.New("cannot append nil telemetry record")
	}

	payload, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("failed serializing telemetry record: %w", err)
	}

	q.mu.Lock()
	defer q.mu.Unlock()

	// Check disk limit
	if q.currentSpoolSize+int64(len(payload)+headerSize) > q.cfg.MaxSizeBytes {
		return ErrQueueDiskFull
	}

	// Rotate segment if active file exceeds MaxSegmentSize
	if q.activeOffset+int64(len(payload)+headerSize) > q.cfg.MaxSegmentSize {
		if err := q.rotateSegmentLocked(); err != nil {
			return fmt.Errorf("failed rotating WAL segment: %w", err)
		}
	}

	// Frame header
	var hdr [headerSize]byte
	copy(hdr[0:4], walMagic[:])
	binary.BigEndian.PutUint32(hdr[4:8], uint32(len(payload)))
	crc := crc32.ChecksumIEEE(payload)
	binary.BigEndian.PutUint32(hdr[8:12], crc)
	binary.BigEndian.PutUint64(hdr[12:20], record.Sequence)
	binary.BigEndian.PutUint64(hdr[20:28], uint64(record.ReceivedAt.UnixNano()))

	// Write header + payload
	if _, err := q.activeFile.Write(hdr[:]); err != nil {
		return fmt.Errorf("failed writing WAL header: %w", err)
	}
	if _, err := q.activeFile.Write(payload); err != nil {
		return fmt.Errorf("failed writing WAL payload: %w", err)
	}

	bytesWritten := int64(headerSize + len(payload))
	q.activeOffset += bytesWritten
	q.currentSpoolSize += bytesWritten
	atomic.AddUint64(&q.totalAppended, 1)

	if q.cfg.SyncMode == "always" {
		_ = q.activeFile.Sync()
	}

	return nil
}

// AppendBatch writes multiple records in a single synchronized transaction
func (q *WALQueue) AppendBatch(records []*model.RawData) error {
	if atomic.LoadInt32(&q.closed) == 1 {
		return errors.New("WAL queue is closed")
	}
	if len(records) == 0 {
		return nil
	}

	q.mu.Lock()
	defer q.mu.Unlock()

	var buf bytes.Buffer
	for _, record := range records {
		if record == nil {
			continue
		}
		payload, err := json.Marshal(record)
		if err != nil {
			return fmt.Errorf("failed serializing telemetry record: %w", err)
		}

		var hdr [headerSize]byte
		copy(hdr[0:4], walMagic[:])
		binary.BigEndian.PutUint32(hdr[4:8], uint32(len(payload)))
		crc := crc32.ChecksumIEEE(payload)
		binary.BigEndian.PutUint32(hdr[8:12], crc)
		binary.BigEndian.PutUint64(hdr[12:20], record.Sequence)
		binary.BigEndian.PutUint64(hdr[20:28], uint64(record.ReceivedAt.UnixNano()))

		buf.Write(hdr[:])
		buf.Write(payload)
	}

	totalLen := int64(buf.Len())
	if q.currentSpoolSize+totalLen > q.cfg.MaxSizeBytes {
		return ErrQueueDiskFull
	}

	if q.activeOffset+totalLen > q.cfg.MaxSegmentSize {
		if err := q.rotateSegmentLocked(); err != nil {
			return fmt.Errorf("failed rotating WAL segment: %w", err)
		}
	}

	if _, err := q.activeFile.Write(buf.Bytes()); err != nil {
		return fmt.Errorf("failed writing batch to WAL: %w", err)
	}

	q.activeOffset += totalLen
	q.currentSpoolSize += totalLen
	atomic.AddUint64(&q.totalAppended, uint64(len(records)))

	if q.cfg.SyncMode == "always" || q.cfg.SyncMode == "batch" {
		_ = q.activeFile.Sync()
	}

	return nil
}

// ReadPendingBatch reads up to maxBatch unacknowledged records in FIFO order
func (q *WALQueue) ReadPendingBatch(maxBatch int) ([]*model.RawData, *BatchToken, error) {
	if atomic.LoadInt32(&q.closed) == 1 {
		return nil, nil, errors.New("WAL queue is closed")
	}
	if maxBatch <= 0 {
		maxBatch = 50
	}

	q.readMu.Lock()
	defer q.readMu.Unlock()

	// Flush active file buffer to disk so reader can observe freshly appended bytes
	q.mu.Lock()
	if q.activeFile != nil {
		_ = q.activeFile.Sync()
	}
	q.mu.Unlock()

	segments, err := q.listSegments()
	if err != nil {
		return nil, nil, err
	}

	var batch []*model.RawData
	token := &BatchToken{
		SegmentIndex: q.checkpoint.SegmentIndex,
		StartOffset:  q.checkpoint.Offset,
		EndOffset:    q.checkpoint.Offset,
	}

	currSegIdx := q.checkpoint.SegmentIndex
	currOffset := q.checkpoint.Offset

	for _, segIdx := range segments {
		if segIdx < currSegIdx {
			continue // Already fully acknowledged
		}

		segPath := q.segmentPath(segIdx)
		f, err := os.Open(segPath)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, nil, fmt.Errorf("failed opening WAL segment %d for read: %w", segIdx, err)
		}

		readOffset := int64(0)
		if segIdx == currSegIdx {
			readOffset = currOffset
		}

		if _, err := f.Seek(readOffset, io.SeekStart); err != nil {
			_ = f.Close()
			return nil, nil, fmt.Errorf("failed seeking segment %d to offset %d: %w", segIdx, readOffset, err)
		}

		for len(batch) < maxBatch {
			recordStartOffset := readOffset
			var hdr [headerSize]byte
			n, err := io.ReadFull(f, hdr[:])
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				// Clean EOF or incomplete tail record on crash
				if n > 0 && n < headerSize {
					logger.Warn("Truncated header (%d bytes) detected at tail of segment %d, halting read", n, segIdx)
				}
				break
			}
			if err != nil {
				_ = f.Close()
				return nil, nil, fmt.Errorf("error reading header in segment %d: %w", segIdx, err)
			}

			// Validate magic
			if !bytes.Equal(hdr[0:4], walMagic[:]) {
				_ = f.Close()
				atomic.AddUint64(&q.checksumFailures, 1)
				logger.Error("Corrupted magic bytes in segment %d at offset %d, quarantining", segIdx, recordStartOffset)
				_ = q.quarantineSegment(segIdx)
				break
			}

			payloadLen := binary.BigEndian.Uint32(hdr[4:8])
			expectedCRC := binary.BigEndian.Uint32(hdr[8:12])
			recordNano := int64(binary.BigEndian.Uint64(hdr[20:28]))

			if payloadLen > 10*1024*1024 { // Sanity check: 10MB record ceiling
				_ = f.Close()
				atomic.AddUint64(&q.checksumFailures, 1)
				logger.Error("Excessive payload length (%d bytes) in segment %d, quarantining", payloadLen, segIdx)
				_ = q.quarantineSegment(segIdx)
				break
			}

			payload := make([]byte, payloadLen)
			if _, err := io.ReadFull(f, payload); err != nil {
				// Interrupted write at end of segment
				logger.Warn("Incomplete payload (%d bytes expected) at tail of segment %d, stopping read", payloadLen, segIdx)
				break
			}

			// Checksum validation
			actualCRC := crc32.ChecksumIEEE(payload)
			if actualCRC != expectedCRC {
				_ = f.Close()
				atomic.AddUint64(&q.checksumFailures, 1)
				logger.Error("Checksum mismatch in segment %d at offset %d (expected %08x, got %08x)",
					segIdx, recordStartOffset, expectedCRC, actualCRC)
				_ = q.quarantineSegment(segIdx)
				break
			}

			var rec model.RawData
			if err := json.Unmarshal(payload, &rec); err != nil {
				_ = f.Close()
				atomic.AddUint64(&q.checksumFailures, 1)
				logger.Error("JSON unmarshal failure in segment %d at offset %d: %v", segIdx, recordStartOffset, err)
				_ = q.quarantineSegment(segIdx)
				break
			}

			readOffset += int64(headerSize + payloadLen)
			batch = append(batch, &rec)

			recTime := time.Unix(0, recordNano).UTC()
			if recTime.After(token.MaxTimestamp) {
				token.MaxTimestamp = recTime
			}

			token.SegmentIndex = segIdx
			token.EndOffset = readOffset
			token.Count = len(batch)
		}

		_ = f.Close()

		if len(batch) >= maxBatch {
			break
		}
	}

	return batch, token, nil
}

// Acknowledge commits the batch token to the checkpoint file and purges fully acknowledged segments
func (q *WALQueue) Acknowledge(token *BatchToken) error {
	if token == nil || token.Count == 0 {
		return nil
	}

	q.readMu.Lock()
	defer q.readMu.Unlock()

	q.checkpoint.SegmentIndex = token.SegmentIndex
	q.checkpoint.Offset = token.EndOffset
	q.checkpoint.TotalAcked += uint64(token.Count)
	q.checkpoint.LastAckedAt = time.Now().UTC()

	atomic.AddUint64(&q.totalAcked, uint64(token.Count))
	atomic.AddUint64(&q.totalReplayed, uint64(token.Count))

	if err := q.saveCheckpoint(); err != nil {
		return fmt.Errorf("failed saving checkpoint: %w", err)
	}

	// Purge fully acknowledged old segments
	q.purgeAcknowledgedSegments(token.SegmentIndex)
	q.updateMetrics()

	return nil
}

// PendingCount calculates the count of unacknowledged records approximately
func (q *WALQueue) PendingCount() int64 {
	appended := atomic.LoadUint64(&q.totalAppended)
	acked := atomic.LoadUint64(&q.totalAcked)
	if appended >= acked {
		return int64(appended - acked)
	}
	return 0
}

// Stats returns comprehensive diagnostic metrics
func (q *WALQueue) Stats() map[string]interface{} {
	q.updateMetrics()

	return map[string]interface{}{
		"enabled":             true,
		"dir":                 q.cfg.Dir,
		"pending_records":     q.PendingCount(),
		"total_appended":      atomic.LoadUint64(&q.totalAppended),
		"total_acked":         atomic.LoadUint64(&q.totalAcked),
		"total_replayed":      atomic.LoadUint64(&q.totalReplayed),
		"checksum_failures":   atomic.LoadUint64(&q.checksumFailures),
		"corrupted_segments":  atomic.LoadUint64(&q.corruptedSegments),
		"spool_size_bytes":    q.currentSpoolSize,
		"max_size_bytes":      q.cfg.MaxSizeBytes,
		"spool_utilization":   float64(q.currentSpoolSize) / float64(q.cfg.MaxSizeBytes) * 100.0,
		"acked_segment_index": q.checkpoint.SegmentIndex,
		"acked_offset":        q.checkpoint.Offset,
		"active_segment_index": q.activeIndex,
		"last_acked_at":       q.checkpoint.LastAckedAt,
		"sync_mode":           q.cfg.SyncMode,
	}
}

// ----------------- Internal Helper Methods -----------------

func (q *WALQueue) segmentPath(index uint32) string {
	return filepath.Join(q.cfg.Dir, fmt.Sprintf("segment_%06d.wal", index))
}

func (q *WALQueue) checkpointPath() string {
	return filepath.Join(q.cfg.Dir, "checkpoint.json")
}

func (q *WALQueue) openActiveSegmentForAppend() error {
	segPath := q.segmentPath(q.activeIndex)
	f, err := os.OpenFile(segPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}

	info, err := f.Stat()
	if err != nil {
		_ = f.Close()
		return err
	}

	q.activeFile = f
	q.activeOffset = info.Size()
	return nil
}

func (q *WALQueue) rotateSegmentLocked() error {
	if q.activeFile != nil {
		_ = q.activeFile.Sync()
		_ = q.activeFile.Close()
		q.activeFile = nil
	}

	q.activeIndex++
	return q.openActiveSegmentForAppend()
}

func (q *WALQueue) listSegments() ([]uint32, error) {
	entries, err := os.ReadDir(q.cfg.Dir)
	if err != nil {
		return nil, err
	}

	var segments []uint32
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasPrefix(name, "segment_") && strings.HasSuffix(name, ".wal") {
			numPart := strings.TrimPrefix(name, "segment_")
			numPart = strings.TrimSuffix(numPart, ".wal")
			if idx, err := strconv.ParseUint(numPart, 10, 32); err == nil {
				segments = append(segments, uint32(idx))
			}
		}
	}

	sort.Slice(segments, func(i, j int) bool {
		return segments[i] < segments[j]
	})

	return segments, nil
}

func (q *WALQueue) loadCheckpoint() error {
	data, err := os.ReadFile(q.checkpointPath())
	if err != nil {
		return err
	}

	var cp Checkpoint
	if err := json.Unmarshal(data, &cp); err != nil {
		return err
	}

	q.checkpoint = cp
	atomic.StoreUint64(&q.totalAcked, cp.TotalAcked)
	return nil
}

func (q *WALQueue) saveCheckpoint() error {
	data, err := json.MarshalIndent(q.checkpoint, "", "  ")
	if err != nil {
		return err
	}

	tmpPath := q.checkpointPath() + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return err
	}

	return os.Rename(tmpPath, q.checkpointPath())
}

func (q *WALQueue) purgeAcknowledgedSegments(ackedSegment uint32) {
	segments, err := q.listSegments()
	if err != nil {
		return
	}

	for _, segIdx := range segments {
		// Only delete segments strictly older than ackedSegment (and never the active segment)
		if segIdx < ackedSegment && segIdx != q.activeIndex {
			path := q.segmentPath(segIdx)
			if err := os.Remove(path); err != nil {
				logger.Warn("Failed purging acknowledged segment %d: %v", segIdx, err)
			} else {
				logger.Debug("Purged acknowledged WAL segment %d", segIdx)
			}
		}
	}
}

func (q *WALQueue) quarantineSegment(segIdx uint32) error {
	atomic.AddUint64(&q.corruptedSegments, 1)

	q.mu.Lock()
	isActive := (q.activeIndex == segIdx)
	if isActive && q.activeFile != nil {
		_ = q.activeFile.Close()
		q.activeFile = nil
	}
	q.mu.Unlock()

	srcPath := q.segmentPath(segIdx)
	dstPath := filepath.Join(q.cfg.Dir, "quarantine", filepath.Base(srcPath))

	logger.Error("Quarantining damaged WAL segment %s to %s", srcPath, dstPath)
	err := os.Rename(srcPath, dstPath)

	if isActive {
		q.mu.Lock()
		q.activeIndex++
		_ = q.openActiveSegmentForAppend()
		q.mu.Unlock()
	}

	return err
}

func (q *WALQueue) updateMetrics() {
	segments, err := q.listSegments()
	if err != nil {
		return
	}

	var totalSize int64
	for _, segIdx := range segments {
		if fi, err := os.Stat(q.segmentPath(segIdx)); err == nil {
			totalSize += fi.Size()
		}
	}

	q.currentSpoolSize = totalSize
}

// Snapshot creates a consistent, validated copy of uncommitted WAL segments and checkpoint in destDir
func (q *WALQueue) Snapshot(destDir string) (*model.WALManifestInfo, error) {
	if atomic.LoadInt32(&q.closed) == 1 {
		return nil, errors.New("WAL queue is closed")
	}

	q.mu.Lock()
	defer q.mu.Unlock()
	q.readMu.Lock()
	defer q.readMu.Unlock()

	// 1. Flush active file buffer to disk
	if q.activeFile != nil {
		_ = q.activeFile.Sync()
	}

	// 2. Ensure destDir exists
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return nil, fmt.Errorf("failed creating WAL snapshot destination: %w", err)
	}

	// 3. Copy checkpoint.json
	cpSrc := q.checkpointPath()
	cpDst := filepath.Join(destDir, "checkpoint.json")
	if err := copyFile(cpSrc, cpDst); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed copying checkpoint file: %w", err)
	}

	// 4. Identify segments from checkpoint.SegmentIndex up to activeIndex
	segments, err := q.listSegments()
	if err != nil {
		return nil, fmt.Errorf("failed listing segments for snapshot: %w", err)
	}

	var artifacts []model.ArtifactInfo
	var totalSize int64

	for _, segIdx := range segments {
		if segIdx < q.checkpoint.SegmentIndex && segIdx != q.activeIndex {
			continue
		}

		srcPath := q.segmentPath(segIdx)
		dstPath := filepath.Join(destDir, filepath.Base(srcPath))

		fi, err := os.Stat(srcPath)
		if err != nil {
			continue
		}

		sha256Hex, crc, err := copyAndChecksum(srcPath, dstPath)
		if err != nil {
			return nil, fmt.Errorf("failed copying segment %d: %w", segIdx, err)
		}

		artifacts = append(artifacts, model.ArtifactInfo{
			Path:        filepath.Base(srcPath),
			SizeBytes:   fi.Size(),
			SHA256:      sha256Hex,
			CRC32:       crc,
			Description: fmt.Sprintf("WAL Segment %06d", segIdx),
		})
		totalSize += fi.Size()
	}

	pendingCount := q.PendingCount()

	info := &model.WALManifestInfo{
		PendingRecords:    pendingCount,
		SpoolSizeBytes:    totalSize,
		CheckpointSegment: q.checkpoint.SegmentIndex,
		CheckpointOffset:  q.checkpoint.Offset,
		ActiveSegment:     q.activeIndex,
		ActiveOffset:      q.activeOffset,
		SegmentFiles:      artifacts,
	}

	return info, nil
}

// RestoreSnapshot restores the WAL queue directory from a verified snapshot directory
func (q *WALQueue) RestoreSnapshot(snapshotDir string) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.readMu.Lock()
	defer q.readMu.Unlock()

	// 1. Close current active file
	if q.activeFile != nil {
		_ = q.activeFile.Close()
		q.activeFile = nil
	}

	// 2. Remove all existing segment files and checkpoint in q.cfg.Dir
	entries, err := os.ReadDir(q.cfg.Dir)
	if err == nil {
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			if strings.HasSuffix(e.Name(), ".wal") || e.Name() == "checkpoint.json" {
				_ = os.Remove(filepath.Join(q.cfg.Dir, e.Name()))
			}
		}
	}

	// 3. Copy all files from snapshotDir into q.cfg.Dir
	snapEntries, err := os.ReadDir(snapshotDir)
	if err != nil {
		return fmt.Errorf("failed reading snapshot directory: %w", err)
	}

	for _, e := range snapEntries {
		if e.IsDir() {
			continue
		}
		src := filepath.Join(snapshotDir, e.Name())
		dst := filepath.Join(q.cfg.Dir, e.Name())
		if err := copyFile(src, dst); err != nil {
			return fmt.Errorf("failed restoring WAL file %s: %w", e.Name(), err)
		}
	}

	// 4. Re-open checkpoint and active segment
	if err := q.loadCheckpoint(); err != nil {
		q.checkpoint = Checkpoint{
			SegmentIndex: 1,
			Offset:       0,
			LastAckedAt:  time.Now().UTC(),
		}
	}
	segments, err := q.listSegments()
	if err == nil && len(segments) > 0 {
		q.activeIndex = segments[len(segments)-1]
	} else {
		q.activeIndex = 1
	}

	if err := q.openActiveSegmentForAppend(); err != nil {
		return fmt.Errorf("failed opening active segment after restore: %w", err)
	}

	q.updateMetrics()
	return nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Sync()
}

func copyAndChecksum(src, dst string) (string, uint32, error) {
	in, err := os.Open(src)
	if err != nil {
		return "", 0, err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return "", 0, err
	}
	defer out.Close()

	shaWriter := sha256.New()
	crcWriter := crc32.NewIEEE()
	multiWriter := io.MultiWriter(out, shaWriter, crcWriter)

	if _, err := io.Copy(multiWriter, in); err != nil {
		return "", 0, err
	}
	if err := out.Sync(); err != nil {
		return "", 0, err
	}

	shaHex := hex.EncodeToString(shaWriter.Sum(nil))
	crcVal := crcWriter.Sum32()
	return shaHex, crcVal, nil
}
