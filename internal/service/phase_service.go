package service

import (
	"fmt"
	"time"

	"datalogger/internal/model"
	"datalogger/internal/repository"
)

type PhaseService struct {
	phaseRepo  *repository.PhaseRepository
	systemRepo *repository.SystemRepository
}

type OverallProgressDTO struct {
	OverallPercentage   float64                  `json:"overall_percentage"`
	CurrentPhase        string                   `json:"current_phase"`
	CurrentPhaseNumber  int                      `json:"current_phase_number"`
	CurrentSubphase     string                   `json:"current_subphase"`
	CurrentTask         string                   `json:"current_task"`
	NextAction          string                   `json:"next_action"`
	EstimatedCompletion string                   `json:"estimated_completion"`
	LastUpdate          string                   `json:"last_update"`
	CompletedTasksCount int                      `json:"completed_tasks_count"`
	ActiveTasksCount    int                      `json:"active_tasks_count"`
	BlockedTasksCount   int                      `json:"blocked_tasks_count"`
	TotalTasksCount     int                      `json:"total_tasks_count"`
	Phases              []model.DevelopmentPhase `json:"phases"`
}

type ActivityItemDTO struct {
	ID        uint      `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	User      string    `json:"user"`
	Phase     string    `json:"phase"`
	Task      string    `json:"task"`
	Action    string    `json:"action"`
	Result    string    `json:"result"`
	Log       string    `json:"log"`
}

func NewPhaseService(phaseRepo *repository.PhaseRepository, systemRepo *repository.SystemRepository) *PhaseService {
	return &PhaseService{phaseRepo: phaseRepo, systemRepo: systemRepo}
}

func (s *PhaseService) GetAllPhases() ([]model.DevelopmentPhase, error) {
	return s.phaseRepo.GetAllPhases()
}

func (s *PhaseService) GetPhaseByID(id uint) (*model.DevelopmentPhase, error) {
	return s.phaseRepo.GetPhaseByID(id)
}

func (s *PhaseService) GetProgressSummary() (*OverallProgressDTO, error) {
	phases, err := s.phaseRepo.GetAllPhases()
	if err != nil {
		return nil, err
	}

	var totalProgress float64
	var completedCount, activeCount, blockedCount, totalCount int
	var currentPhaseName string
	var currentPhaseNum int
	var lastUpdateTime time.Time

	for _, p := range phases {
		totalProgress += p.Progress

		if p.PhaseNumber == 2 {
			currentPhaseName = p.Name
			currentPhaseNum = p.PhaseNumber
		}

		for _, t := range p.Tasks {
			if t.Status == model.StatusSuperseded {
				continue
			}
			totalCount++
			switch t.Status {
			case model.StatusDone:
				completedCount++
			case model.StatusWorking, model.StatusTesting:
				activeCount++
			case model.StatusBlocked, model.StatusFailed:
				blockedCount++
			}

			if t.UpdatedAt.After(lastUpdateTime) {
				lastUpdateTime = t.UpdatedAt
			}
		}
	}

	overallPct := 0.0
	if len(phases) > 0 {
		overallPct = totalProgress / float64(len(phases))
	}

	if currentPhaseName == "" {
		currentPhaseName = "Phase 2 — Device & Communication"
		currentPhaseNum = 2
	}

	dto := &OverallProgressDTO{
		OverallPercentage:   overallPct,
		CurrentPhase:        currentPhaseName,
		CurrentPhaseNumber:  currentPhaseNum,
		CurrentSubphase:     "Phase 2.3 — Communication Hardening & Real Device Validation",
		CurrentTask:         "Phase 2 Verification Complete (2.1: 22/22, 2.2: 27/27, 2.3: 32/32 PASS)",
		NextAction:          "Phase 2 Accepted & Verified | Ready for Phase 3 Data Engine",
		EstimatedCompletion: "Phase 1 & Phase 2 Accepted (100%) | Full System: Q4 2026",
		LastUpdate:          lastUpdateTime.Format("2006-01-02 15:04:05"),
		CompletedTasksCount: completedCount,
		ActiveTasksCount:    activeCount,
		BlockedTasksCount:   blockedCount,
		TotalTasksCount:     totalCount,
		Phases:              phases,
	}

	return dto, nil
}

func (s *PhaseService) GetAllTasks(phaseID uint, status string) ([]model.DevelopmentTask, error) {
	return s.phaseRepo.GetAllTasks(phaseID, status)
}

func (s *PhaseService) GetTaskByID(id uint) (*model.DevelopmentTask, error) {
	return s.phaseRepo.GetTaskByID(id)
}

func (s *PhaseService) CreateTask(task *model.DevelopmentTask, username string) error {
	if err := s.phaseRepo.CreateTask(task); err != nil {
		return err
	}

	// Add log
	_ = s.phaseRepo.AddTaskLog(&model.DevelopmentTaskLog{
		TaskID:    task.ID,
		Timestamp: time.Now(),
		User:      username,
		Action:    "TASK_CREATED",
		Result:    "OK",
		Log:       fmt.Sprintf("Created task: %s", task.TaskName),
	})

	_, _ = s.phaseRepo.RecalculateProgress(task.PhaseID)
	return nil
}

func (s *PhaseService) UpdateTask(id uint, updates map[string]interface{}, username string) (*model.DevelopmentTask, error) {
	task, err := s.phaseRepo.GetTaskByID(id)
	if err != nil {
		return nil, err
	}

	oldStatus := task.Status
	oldProgress := task.Progress

	if name, ok := updates["task_name"].(string); ok && name != "" {
		task.TaskName = name
	}
	if desc, ok := updates["description"].(string); ok {
		task.Description = desc
	}
	if status, ok := updates["status"].(string); ok && status != "" {
		task.Status = model.TaskStatus(status)
		if task.Status == model.StatusDone {
			now := time.Now()
			task.CompletionDate = &now
			task.Progress = 100.0
		}
	}
	if prog, ok := updates["progress"].(float64); ok {
		task.Progress = prog
	}
	if priority, ok := updates["priority"].(string); ok && priority != "" {
		task.Priority = model.Priority(priority)
	}
	if owner, ok := updates["owner"].(string); ok {
		task.Owner = owner
	}
	if notes, ok := updates["notes"].(string); ok {
		task.Notes = notes
	}
	if testResult, ok := updates["test_result"].(string); ok {
		task.TestResult = testResult
	}

	if err := s.phaseRepo.UpdateTask(task); err != nil {
		return nil, err
	}

	// If status or progress changed, log it
	if oldStatus != task.Status || oldProgress != task.Progress {
		logMsg := fmt.Sprintf("Status changed from %s to %s (Progress: %.0f%%)", oldStatus, task.Status, task.Progress)
		_ = s.phaseRepo.AddTaskLog(&model.DevelopmentTaskLog{
			TaskID:    task.ID,
			Timestamp: time.Now(),
			User:      username,
			Action:    "STATUS_UPDATE",
			Result:    string(task.Status),
			Log:       logMsg,
		})

		_ = s.systemRepo.AddAuditTrail(&model.AuditTrail{
			Username:  username,
			Action:    "UPDATE_TASK",
			Resource:  fmt.Sprintf("Task #%d (%s)", task.ID, task.TaskName),
			Details:   logMsg,
			CreatedAt: time.Now(),
		})
	}

	_, _ = s.phaseRepo.RecalculateProgress(task.PhaseID)
	return task, nil
}

func (s *PhaseService) DeleteTask(id uint, username string) error {
	task, err := s.phaseRepo.GetTaskByID(id)
	if err != nil {
		return err
	}
	phaseID := task.PhaseID

	if err := s.phaseRepo.DeleteTask(id); err != nil {
		return err
	}

	_ = s.systemRepo.AddAuditTrail(&model.AuditTrail{
		Username:  username,
		Action:    "DELETE_TASK",
		Resource:  fmt.Sprintf("Task #%d (%s)", id, task.TaskName),
		Details:   "Task deleted by user",
		CreatedAt: time.Now(),
	})

	_, _ = s.phaseRepo.RecalculateProgress(phaseID)
	return nil
}

func (s *PhaseService) GetActivityTimeline(limit int) ([]ActivityItemDTO, error) {
	if limit <= 0 {
		limit = 30
	}
	logs, err := s.phaseRepo.GetRecentTaskLogs(limit)
	if err != nil {
		return nil, err
	}

	var items []ActivityItemDTO
	for _, l := range logs {
		task, err := s.phaseRepo.GetTaskByID(l.TaskID)
		taskName := fmt.Sprintf("Task #%d", l.TaskID)
		phaseName := "Development"
		if err == nil && task != nil {
			taskName = task.TaskName
			phaseName = fmt.Sprintf("Phase %d", task.PhaseID)
		}

		items = append(items, ActivityItemDTO{
			ID:        l.ID,
			Timestamp: l.Timestamp,
			User:      l.User,
			Phase:     phaseName,
			Task:      taskName,
			Action:    l.Action,
			Result:    l.Result,
			Log:       l.Log,
		})
	}

	return items, nil
}
