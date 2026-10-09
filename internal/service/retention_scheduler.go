package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"datalogger/internal/config"
	"datalogger/internal/logger"
	"datalogger/internal/model"
)

type RetentionScheduler struct {
	retentionService *RetentionService
	cfg              *config.Config

	mu         sync.RWMutex
	enabled    bool
	interval   time.Duration
	timeOfDay  string
	lastRun    *time.Time
	nextRun    *time.Time
	lastStatus string

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

func NewRetentionScheduler(retentionService *RetentionService, cfg *config.Config) *RetentionScheduler {
	if cfg == nil {
		cfg = config.Get()
	}

	return &RetentionScheduler{
		retentionService: retentionService,
		cfg:              cfg,
		enabled:          true, // Scheduler ticker is enabled, but individual policies control whether deletion happens!
		interval:         1 * time.Hour, // Check hourly for due policies
		timeOfDay:        "03:30",
		lastStatus:       "INITIALIZED",
	}
}

// Start initiates the housekeeping scheduler worker
func (s *RetentionScheduler) Start() {
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
	go s.workerLoop(ctx)

	logger.Info("RetentionScheduler started (Check Interval: %v, Housekeeping Time: %s)", s.interval, s.timeOfDay)
}

// Stop gracefully halts the housekeeping scheduler
func (s *RetentionScheduler) Stop() {
	s.mu.Lock()
	if s.cancel == nil {
		s.mu.Unlock()
		return
	}
	s.cancel()
	s.cancel = nil
	s.mu.Unlock()

	s.wg.Wait()
	logger.Info("RetentionScheduler stopped cleanly")
}

func (s *RetentionScheduler) calculateNextRunLocked() {
	now := time.Now().UTC()
	var hr, min int
	_, _ = fmt.Sscanf(s.timeOfDay, "%02d:%02d", &hr, &min)

	target := time.Date(now.Year(), now.Month(), now.Day(), hr, min, 0, 0, time.UTC)
	if !target.After(now) {
		target = target.Add(24 * time.Hour)
	}
	s.nextRun = &target
}

func (s *RetentionScheduler) workerLoop(ctx context.Context) {
	defer s.wg.Done()

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.evaluateDuePolicies(ctx)
		}
	}
}

// RunHousekeepingCycle manually triggers a full sweep across all enabled policies
func (s *RetentionScheduler) RunHousekeepingCycle(ctx context.Context, initiatedBy string) []model.RetentionExecutionLog {
	if s.retentionService == nil {
		return nil
	}

	policies, err := s.retentionService.GetAllPolicies()
	if err != nil {
		logger.Warn("Housekeeping cycle failed to load policies: %v", err)
		return nil
	}

	results := make([]model.RetentionExecutionLog, 0)

	for _, p := range policies {
		if !p.Enabled {
			continue // Critical safety rule: skip disabled policies
		}

		logger.Info("Housekeeping running policy %s (%s)...", p.ID, p.Category)
		log, execErr := s.retentionService.ExecutePolicy(ctx, p.ID, initiatedBy, false)
		if log != nil {
			results = append(results, *log)
		}
		if execErr != nil {
			logger.Warn("Housekeeping policy %s execution warning: %v", p.ID, execErr)
		}
	}

	now := time.Now().UTC()
	s.mu.Lock()
	s.lastRun = &now
	s.lastStatus = fmt.Sprintf("COMPLETED (%d policies evaluated)", len(results))
	s.calculateNextRunLocked()
	s.mu.Unlock()

	return results
}

func (s *RetentionScheduler) evaluateDuePolicies(ctx context.Context) {
	s.mu.RLock()
	if !s.enabled {
		s.mu.RUnlock()
		return
	}
	next := s.nextRun
	s.mu.RUnlock()

	now := time.Now().UTC()
	if next != nil && now.After(*next) {
		logger.Info("Scheduled daily housekeeping threshold reached. Executing enabled retention policies...")
		s.RunHousekeepingCycle(ctx, "scheduler")
	}
}
