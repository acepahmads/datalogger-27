package scheduler

import (
	"context"
	"sync"
	"time"

	"datalogger/internal/logger"
	"datalogger/internal/model"
	"datalogger/internal/repository"
	"datalogger/pkg/sysinfo"
)

type JobFunc func()

type Job struct {
	Name     string
	Interval time.Duration
	Fn       JobFunc
}

type Scheduler struct {
	jobs       []Job
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
	systemRepo *repository.SystemRepository
}

func NewScheduler(systemRepo *repository.SystemRepository) *Scheduler {
	ctx, cancel := context.WithCancel(context.Background())
	return &Scheduler{
		ctx:        ctx,
		cancel:     cancel,
		systemRepo: systemRepo,
	}
}

func (s *Scheduler) RegisterJob(name string, interval time.Duration, fn JobFunc) {
	s.jobs = append(s.jobs, Job{
		Name:     name,
		Interval: interval,
		Fn:       fn,
	})
}

func (s *Scheduler) Start() {
	logger.Info("Starting lightweight edge scheduler with %d registered jobs", len(s.jobs)+1)

	// Built-in system health recording job (every 60s)
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		ticker := time.NewTicker(60 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-s.ctx.Done():
				return
			case <-ticker.C:
				info := sysinfo.GetInfo()
				_ = s.systemRepo.RecordSystemHealth(&model.SystemHealth{
					Timestamp:      time.Now(),
					CPUPercent:     info.CPUPercent,
					RAMUsedBytes:   info.RAMUsedBytes,
					RAMTotalBytes:  info.RAMTotalBytes,
					RAMPercent:     info.RAMPercent,
					DiskUsedBytes:  info.DiskUsedBytes,
					DiskTotalBytes: info.DiskTotalBytes,
					DiskPercent:    info.DiskPercent,
					UptimeSeconds:  info.UptimeSeconds,
					GoroutineCount: info.Goroutines,
					DBStatus:       "OK",
					ServiceStatus:  "RUNNING",
					WorkerStatus:   "IDLE",
					QueueStatus:    "OPTIMAL",
				})
			}
		}
	}()

	// Run user-registered jobs
	for _, j := range s.jobs {
		job := j
		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			ticker := time.NewTicker(job.Interval)
			defer ticker.Stop()

			for {
				select {
				case <-s.ctx.Done():
					return
				case <-ticker.C:
					job.Fn()
				}
			}
		}()
	}
}

func (s *Scheduler) Stop() {
	logger.Info("Stopping scheduler workers...")
	s.cancel()
	s.wg.Wait()
	logger.Info("Scheduler workers stopped successfully")
}
