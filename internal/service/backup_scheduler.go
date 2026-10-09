package service

import (
	"context"
	"sync"
	"time"

	"datalogger/internal/config"
	"datalogger/internal/logger"
	"datalogger/internal/model"
)

type BackupScheduler struct {
	backupService *BackupService
	cfg           *config.Config

	mu          sync.RWMutex
	enabled     bool
	interval    time.Duration
	timeOfDay   string
	maxKeep     int
	lastRun     *time.Time
	nextRun     *time.Time
	lastStatus  string

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

func NewBackupScheduler(backupService *BackupService, cfg *config.Config) *BackupScheduler {
	if cfg == nil {
		cfg = config.Get()
	}

	intervalHours := cfg.BackupScheduleIntervalHours
	if intervalHours <= 0 {
		intervalHours = 24
	}

	maxKeep := cfg.BackupMaxKeepCount
	if maxKeep <= 0 {
		maxKeep = 10
	}

	timeOfDay := cfg.BackupScheduleTime
	if timeOfDay == "" {
		timeOfDay = "02:00"
	}

	return &BackupScheduler{
		backupService: backupService,
		cfg:           cfg,
		enabled:       cfg.BackupScheduleEnabled,
		interval:      time.Duration(intervalHours) * time.Hour,
		timeOfDay:     timeOfDay,
		maxKeep:       maxKeep,
		lastStatus:    "INITIALIZED",
	}
}

// Start begins the background backup schedule worker
func (s *BackupScheduler) Start() {
	s.mu.Lock()
	if s.cancel != nil {
		s.mu.Unlock()
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	s.ctx = ctx
	s.cancel = cancel
	s.calculateNextRunLocked()
	s.mu.Unlock()

	s.wg.Add(1)
	go s.scheduleWorker(ctx)

	logger.Info("BackupScheduler started (Enabled: %v, Interval: %v, Daily Time: %s)",
		s.enabled, s.interval, s.timeOfDay)
}

// Stop gracefully terminates the backup scheduler
func (s *BackupScheduler) Stop() {
	s.mu.Lock()
	if s.cancel == nil {
		s.mu.Unlock()
		return
	}
	s.cancel()
	s.cancel = nil
	s.mu.Unlock()

	s.wg.Wait()
	logger.Info("BackupScheduler stopped cleanly")
}

// GetScheduleConfig returns current schedule configuration
func (s *BackupScheduler) GetScheduleConfig() model.BackupScheduleConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return model.BackupScheduleConfig{
		Enabled:       s.enabled,
		IntervalHours: int(s.interval.Hours()),
		TimeOfDay:     s.timeOfDay,
		Compression:   s.cfg.BackupCompression,
		KeepMaxCount:  s.maxKeep,
		NextRunTime:   s.nextRun,
		LastRunTime:   s.lastRun,
		LastStatus:    s.lastStatus,
	}
}

// UpdateScheduleConfig updates the scheduler runtime options
func (s *BackupScheduler) UpdateScheduleConfig(enabled bool, intervalHours int, timeOfDay string, keepMax int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.enabled = enabled
	if intervalHours > 0 {
		s.interval = time.Duration(intervalHours) * time.Hour
	}
	if timeOfDay != "" {
		s.timeOfDay = timeOfDay
	}
	if keepMax > 0 {
		s.maxKeep = keepMax
	}

	s.calculateNextRunLocked()

	// Update global config
	s.cfg.BackupScheduleEnabled = s.enabled
	s.cfg.BackupScheduleIntervalHours = int(s.interval.Hours())
	s.cfg.BackupScheduleTime = s.timeOfDay
	s.cfg.BackupMaxKeepCount = s.maxKeep

	logger.Info("BackupScheduler configuration updated (Enabled: %v, Interval: %v, Daily Time: %s, KeepMax: %d)",
		s.enabled, s.interval, s.timeOfDay, s.maxKeep)
}

func (s *BackupScheduler) scheduleWorker(ctx context.Context) {
	defer s.wg.Done()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case now := <-ticker.C:
			s.mu.RLock()
			enabled := s.enabled
			next := s.nextRun
			s.mu.RUnlock()

			if !enabled || next == nil {
				continue
			}

			if now.After(*next) {
				s.runScheduledBackup(now)
			}
		}
	}
}

func (s *BackupScheduler) runScheduledBackup(triggerTime time.Time) {
	s.mu.Lock()
	s.lastRun = &triggerTime
	s.calculateNextRunLocked()
	s.mu.Unlock()

	logger.Info("Executing scheduled automatic backup...")
	jobID, err := s.backupService.StartBackupJob(model.BackupTypeScheduled, "scheduler")
	if err != nil {
		logger.Warn("Scheduled backup skipped or failed to start: %v", err)
		s.mu.Lock()
		s.lastStatus = "SKIPPED: " + err.Error()
		s.mu.Unlock()
		return
	}

	s.mu.Lock()
	s.lastStatus = "RUNNING (" + jobID + ")"
	s.mu.Unlock()

	// Wait asynchronously for job to finish to trigger retention pruning
	go func() {
		for i := 0; i < 60; i++ {
			time.Sleep(2 * time.Second)
			job, err := s.backupService.GetJob(jobID)
			if err == nil && job != nil {
				if job.Status == model.JobStatusCompleted {
					s.mu.Lock()
					s.lastStatus = "SUCCESS"
					s.mu.Unlock()

					// Automatic pruning of old backups
					_, _ = s.backupService.PruneOldBackups(s.maxKeep)
					break
				} else if job.Status == model.JobStatusFailed {
					s.mu.Lock()
					s.lastStatus = "FAILED: " + job.Error
					s.mu.Unlock()
					break
				}
			}
		}
	}()
}

func (s *BackupScheduler) calculateNextRunLocked() {
	if !s.enabled {
		s.nextRun = nil
		return
	}

	now := time.Now()
	next := now.Add(s.interval)
	s.nextRun = &next
}
