package scheduler

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"datalogger/internal/logger"
	"datalogger/internal/model"
	"datalogger/internal/repository"
	"datalogger/internal/service"
)

type ActuatorScheduler struct {
	actuatorRepo     *repository.ActuatorRepository
	actuationService *service.ActuationService
	checkInterval    time.Duration

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
	mu     sync.Mutex
}

func NewActuatorScheduler(
	actuatorRepo *repository.ActuatorRepository,
	actuationService *service.ActuationService,
) *ActuatorScheduler {
	return &ActuatorScheduler{
		actuatorRepo:     actuatorRepo,
		actuationService: actuationService,
		checkInterval:    5 * time.Second,
	}
}

func (s *ActuatorScheduler) SetCheckInterval(d time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if d > 0 {
		s.checkInterval = d
	}
}

func (s *ActuatorScheduler) Start() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.cancel != nil {
		return // already running
	}

	s.ctx, s.cancel = context.WithCancel(context.Background())
	s.wg.Add(1)

	go func() {
		defer s.wg.Done()
		logger.Info("Actuator and watchdog scheduler worker started (interval: %v)", s.checkInterval)

		// Immediate check on startup
		s.tick()

		ticker := time.NewTicker(s.checkInterval)
		defer ticker.Stop()

		for {
			select {
			case <-s.ctx.Done():
				logger.Info("Actuator scheduler worker stopped")
				return
			case <-ticker.C:
				s.tick()
			}
		}
	}()
}

func (s *ActuatorScheduler) Stop() {
	s.mu.Lock()
	if s.cancel != nil {
		s.cancel()
		s.cancel = nil
	}
	s.mu.Unlock()
	s.wg.Wait()
}

func (s *ActuatorScheduler) tick() {
	now := time.Now().UTC()

	// 1. Process Auto-Off Watchdog Timers
	s.processAutoOffWatchdogs(now)

	// 2. Process Due Actuation Schedules
	s.processDueSchedules(now)
}

func (s *ActuatorScheduler) processAutoOffWatchdogs(now time.Time) {
	actuators, err := s.actuatorRepo.ListActuatorsDueForAutoOff(now)
	if err != nil {
		logger.Error("Failed to query actuators due for auto-off watchdog: %v", err)
		return
	}

	for _, act := range actuators {
		logger.Info("Auto-off watchdog expired for actuator %s (%s). Executing automatic OFF command.",
			act.ActuatorName, act.ActuatorCode)

		ctx, cancel := context.WithTimeout(s.ctx, 5*time.Second)
		_, err := s.actuationService.Actuate(ctx, act.ID, model.ActionTurnOff, "WATCHDOG: Auto-Off Expired", nil)
		cancel()

		if err != nil {
			logger.Error("Auto-off watchdog failed to turn off actuator %s: %v", act.ActuatorName, err)
		}
	}
}

func (s *ActuatorScheduler) processDueSchedules(now time.Time) {
	schedules, err := s.actuatorRepo.ListDueSchedules(now)
	if err != nil {
		logger.Error("Failed to query due actuation schedules: %v", err)
		return
	}

	for _, sched := range schedules {
		if !s.isDayEnabled(sched.DaysOfWeek, now) {
			// Not scheduled for today, advance next run
			nextRun := service.CalculateNextScheduleRun(&sched, now)
			_ = s.actuatorRepo.UpdateScheduleRun(sched.ID, now, nextRun, "SKIPPED_DAY")
			continue
		}

		if sched.Actuator == nil || !sched.Actuator.Enabled {
			// Actuator disabled, advance schedule without firing
			nextRun := service.CalculateNextScheduleRun(&sched, now)
			_ = s.actuatorRepo.UpdateScheduleRun(sched.ID, now, nextRun, "ACTUATOR_DISABLED")
			continue
		}

		logger.Info("Executing scheduled actuation '%s' for actuator %s", sched.ScheduleName, sched.Actuator.ActuatorName)

		ctx, cancel := context.WithTimeout(s.ctx, 5*time.Second)
		triggeredBy := fmt.Sprintf("SCHEDULER: %s", sched.ScheduleName)
		_, actErr := s.actuationService.Actuate(ctx, sched.ActuatorID, model.ActionTurnOn, triggeredBy, &sched.ID)
		cancel()

		status := "SUCCESS"
		if actErr != nil {
			status = "FAILED"
			logger.Error("Scheduled actuation '%s' failed: %v", sched.ScheduleName, actErr)
		}

		// Calculate Next Run time based on recurrence rules
		nextRun := service.CalculateNextScheduleRun(&sched, now)
		_ = s.actuatorRepo.UpdateScheduleRun(sched.ID, now, nextRun, status)

		// Set auto-off duration timer if specified
		if actErr == nil && sched.DurationSeconds > 0 {
			due := now.Add(time.Duration(sched.DurationSeconds) * time.Second)
			_ = s.actuatorRepo.UpdateActuatorState(sched.ActuatorID, model.ActuatorStateOn, &due, triggeredBy, status)
		}
	}
}

func (s *ActuatorScheduler) isDayEnabled(daysOfWeek string, now time.Time) bool {
	if strings.TrimSpace(daysOfWeek) == "" || daysOfWeek == "*" {
		return true
	}

	// Sunday in Go is 0; map to ISO 7
	isoWeekday := int(now.Weekday())
	if isoWeekday == 0 {
		isoWeekday = 7
	}

	dayStr := strconv.Itoa(isoWeekday)
	tokens := strings.Split(daysOfWeek, ",")
	for _, t := range tokens {
		if strings.TrimSpace(t) == dayStr {
			return true
		}
	}
	return false
}
