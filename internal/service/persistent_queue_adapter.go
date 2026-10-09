package service

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"datalogger/internal/logger"
	"datalogger/internal/model"
	"datalogger/internal/queue"
	"datalogger/internal/repository"
)

// PersistentQueueAdapter implements TelemetryPersistenceBoundary with an append-only WAL disk spool
type PersistentQueueAdapter struct {
	repo       *repository.TelemetryRepository
	wal        *queue.WALQueue
	batchSize  int
	maxRetries int

	mu                    sync.RWMutex
	persistenceStatus     string // "HEALTHY", "DEGRADED", "FAILING"
	consecutiveDBFailures int
	lastDBError           string
	lastDBErrorAt         *time.Time
	lastCommitAt          time.Time
	lastSpoolAt           time.Time

	replayerCtx    context.Context
	replayerCancel context.CancelFunc
	replayerWg     sync.WaitGroup
	isReplaying    int32
}

// NewPersistentQueueAdapter constructs the durable WAL persistence adapter and starts the recovery replayer
func NewPersistentQueueAdapter(
	repo *repository.TelemetryRepository,
	wal *queue.WALQueue,
	batchSize int,
	maxRetries int,
) *PersistentQueueAdapter {
	if batchSize <= 0 {
		batchSize = 50
	}
	if maxRetries <= 0 {
		maxRetries = 3
	}

	ctx, cancel := context.WithCancel(context.Background())
	a := &PersistentQueueAdapter{
		repo:              repo,
		wal:               wal,
		batchSize:         batchSize,
		maxRetries:        maxRetries,
		persistenceStatus: "HEALTHY",
		replayerCtx:       ctx,
		replayerCancel:    cancel,
	}

	a.replayerWg.Add(1)
	go a.backgroundReplayWorker()

	return a
}

// Close stops the replayer worker and closes the underlying WAL queue
func (a *PersistentQueueAdapter) Close() error {
	a.replayerCancel()
	a.replayerWg.Wait()

	if a.wal != nil {
		return a.wal.Close()
	}
	return nil
}

// SaveBatch durably spools incoming records to the disk WAL first, then attempts MariaDB batch commit
func (a *PersistentQueueAdapter) SaveBatch(ctx context.Context, batch []*model.RawData) error {
	if len(batch) == 0 {
		return nil
	}

	now := time.Now().UTC()

	// 1. Durably spool records to disk WAL
	if a.wal != nil {
		if err := a.wal.AppendBatch(batch); err != nil {
			logger.Error("Critical: failed appending telemetry batch to WAL spool: %v", err)
			a.mu.Lock()
			a.persistenceStatus = "FAILING"
			a.lastDBError = fmt.Sprintf("WAL append failure: %v", err)
			a.lastDBErrorAt = &now
			a.mu.Unlock()
			return err
		}
		a.mu.Lock()
		a.lastSpoolAt = now
		a.mu.Unlock()
	}

	// If no database repository is configured, data is safely spooled
	if a.repo == nil {
		return nil
	}

	// 2. Attempt MariaDB persistence
	var dbErr error
	for attempt := 1; attempt <= a.maxRetries; attempt++ {
		dbErr = a.repo.SaveBatch(ctx, batch)
		if dbErr == nil {
			break
		}
		if attempt < a.maxRetries {
			// Exponential backoff with jitter
			delay := time.Duration(attempt*100)*time.Millisecond + time.Duration(rand.Intn(50))*time.Millisecond
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
			}
		}
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	if dbErr == nil {
		a.consecutiveDBFailures = 0
		a.persistenceStatus = "HEALTHY"
		a.lastCommitAt = now

		// Since direct commit succeeded, acknowledge batch in WAL if single batch without historical backlog
		if a.wal != nil && a.wal.PendingCount() <= int64(len(batch)) {
			_, token, err := a.wal.ReadPendingBatch(len(batch))
			if err == nil && token != nil && token.Count > 0 {
				_ = a.wal.Acknowledge(token)
			}
		} else {
			// Historical backlog exists, trigger async replayer
			a.triggerReplay()
		}
		return nil
	}

	// MariaDB write failed -> records are safely retained in WAL disk spool
	a.consecutiveDBFailures++
	a.lastDBError = dbErr.Error()
	a.lastDBErrorAt = &now
	if a.consecutiveDBFailures >= 5 {
		a.persistenceStatus = "FAILING"
	} else {
		a.persistenceStatus = "DEGRADED"
	}

	logger.Warn("MariaDB unavailable: batch (%d records) safely retained in persistent disk spool (consecutive failures: %d): %v",
		len(batch), a.consecutiveDBFailures, dbErr)

	// Return nil because records are durably safely spooled to disk WAL without data loss
	return nil
}

// ReplayPending flushes uncommitted records from the WAL queue into MariaDB
func (a *PersistentQueueAdapter) ReplayPending(ctx context.Context) (int, error) {
	if a.wal == nil || a.repo == nil {
		return 0, nil
	}

	totalReplayed := 0

	for {
		select {
		case <-ctx.Done():
			return totalReplayed, ctx.Err()
		default:
		}

		batch, token, err := a.wal.ReadPendingBatch(a.batchSize)
		if err != nil {
			return totalReplayed, fmt.Errorf("WAL read failure during replay: %w", err)
		}
		if len(batch) == 0 {
			break // All pending records committed and acknowledged
		}

		// Commit to MariaDB
		dbErr := a.repo.SaveBatch(ctx, batch)
		if dbErr != nil {
			a.mu.Lock()
			a.consecutiveDBFailures++
			a.lastDBError = dbErr.Error()
			now := time.Now().UTC()
			a.lastDBErrorAt = &now
			if a.consecutiveDBFailures >= 5 {
				a.persistenceStatus = "FAILING"
			} else {
				a.persistenceStatus = "DEGRADED"
			}
			a.mu.Unlock()
			return totalReplayed, fmt.Errorf("database commit failed during replay: %w", dbErr)
		}

		// Acknowledge committed batch
		if err := a.wal.Acknowledge(token); err != nil {
			logger.Error("Failed acknowledging replayed WAL batch: %v", err)
		}

		totalReplayed += len(batch)
		a.mu.Lock()
		a.consecutiveDBFailures = 0
		a.persistenceStatus = "HEALTHY"
		a.lastCommitAt = time.Now().UTC()
		a.mu.Unlock()
	}

	if totalReplayed > 0 {
		logger.Info("WAL Replay successfully committed %d pending records to MariaDB", totalReplayed)
	}

	return totalReplayed, nil
}

// backgroundReplayWorker periodically checks for and replays pending WAL records
func (a *PersistentQueueAdapter) backgroundReplayWorker() {
	defer a.replayerWg.Done()

	// Initial startup replay attempt after 500ms
	select {
	case <-a.replayerCtx.Done():
		return
	case <-time.After(500 * time.Millisecond):
		_, _ = a.ReplayPending(a.replayerCtx)
	}

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-a.replayerCtx.Done():
			return
		case <-ticker.C:
			if a.wal != nil && a.wal.PendingCount() > 0 {
				a.triggerReplay()
			}
		}
	}
}

func (a *PersistentQueueAdapter) triggerReplay() {
	if !atomic.CompareAndSwapInt32(&a.isReplaying, 0, 1) {
		return // Replay already active
	}
	go func() {
		defer atomic.StoreInt32(&a.isReplaying, 0)
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		_, _ = a.ReplayPending(ctx)
	}()
}

// GetStats returns comprehensive diagnostic metrics for system health monitoring
func (a *PersistentQueueAdapter) GetStats() map[string]interface{} {
	a.mu.RLock()
	defer a.mu.RUnlock()

	stats := map[string]interface{}{
		"status":                 a.persistenceStatus,
		"consecutive_db_errors": a.consecutiveDBFailures,
		"last_db_error":         a.lastDBError,
		"last_db_error_at":      a.lastDBErrorAt,
		"last_commit_at":        a.lastCommitAt,
		"last_spool_at":         a.lastSpoolAt,
	}

	if a.wal != nil {
		for k, v := range a.wal.Stats() {
			stats["wal_"+k] = v
		}
		stats["pending_records"] = a.wal.PendingCount()
	} else {
		stats["pending_records"] = int64(0)
	}

	return stats
}

// GetHealth returns persistence health and consecutive error counts
func (a *PersistentQueueAdapter) GetHealth() (string, int) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.persistenceStatus, a.consecutiveDBFailures
}

// Snapshot delegates WAL snapshot creation to underlying WAL queue
func (a *PersistentQueueAdapter) Snapshot(destDir string) (*model.WALManifestInfo, error) {
	if a.wal == nil {
		return &model.WALManifestInfo{}, nil
	}
	return a.wal.Snapshot(destDir)
}

// RestoreSnapshot restores WAL queue files from a snapshot directory and resets replayer
func (a *PersistentQueueAdapter) RestoreSnapshot(snapshotDir string) error {
	if a.wal == nil {
		return nil
	}
	return a.wal.RestoreSnapshot(snapshotDir)
}

// GetWAL returns the underlying WAL queue instance
func (a *PersistentQueueAdapter) GetWAL() *queue.WALQueue {
	return a.wal
}
