package repository

import (
	"time"

	"datalogger/internal/model"

	"gorm.io/gorm"
)

type PhaseRepository struct {
	db *gorm.DB
}

func NewPhaseRepository(db *gorm.DB) *PhaseRepository {
	return &PhaseRepository{db: db}
}

func (r *PhaseRepository) GetAllPhases() ([]model.DevelopmentPhase, error) {
	var phases []model.DevelopmentPhase
	err := r.db.Preload("Subphases", func(db *gorm.DB) *gorm.DB {
		return db.Order("order_index ASC")
	}).Preload("Tasks").Order("phase_number ASC").Find(&phases).Error
	return phases, err
}

func (r *PhaseRepository) GetPhaseByID(id uint) (*model.DevelopmentPhase, error) {
	var phase model.DevelopmentPhase
	err := r.db.Preload("Tasks.Logs").Preload("Tasks.Evidences").Preload("Subphases.Tasks").Preload("Subphases", func(db *gorm.DB) *gorm.DB {
		return db.Order("order_index ASC")
	}).First(&phase, id).Error
	if err != nil {
		return nil, err
	}
	return &phase, nil
}

func (r *PhaseRepository) UpdatePhase(phase *model.DevelopmentPhase) error {
	return r.db.Save(phase).Error
}

func (r *PhaseRepository) GetAllTasks(phaseID uint, status string) ([]model.DevelopmentTask, error) {
	var tasks []model.DevelopmentTask
	query := r.db.Preload("Logs").Preload("Evidences").Order("phase_id ASC, order_index ASC")
	if phaseID > 0 {
		query = query.Where("phase_id = ?", phaseID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	err := query.Find(&tasks).Error
	return tasks, err
}

func (r *PhaseRepository) GetTaskByID(id uint) (*model.DevelopmentTask, error) {
	var task model.DevelopmentTask
	err := r.db.Preload("Logs").Preload("Evidences").Preload("Dependencies").First(&task, id).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

func (r *PhaseRepository) CreateTask(task *model.DevelopmentTask) error {
	return r.db.Create(task).Error
}

func (r *PhaseRepository) UpdateTask(task *model.DevelopmentTask) error {
	return r.db.Save(task).Error
}

func (r *PhaseRepository) DeleteTask(id uint) error {
	return r.db.Delete(&model.DevelopmentTask{}, id).Error
}

func (r *PhaseRepository) AddTaskLog(log *model.DevelopmentTaskLog) error {
	return r.db.Create(log).Error
}

func (r *PhaseRepository) AddEvidence(evidence *model.DevelopmentEvidence) error {
	return r.db.Create(evidence).Error
}

func (r *PhaseRepository) GetRecentTaskLogs(limit int) ([]model.DevelopmentTaskLog, error) {
	var logs []model.DevelopmentTaskLog
	err := r.db.Order("timestamp DESC").Limit(limit).Find(&logs).Error
	return logs, err
}

// RecalculateProgress recalculates phase and subphase progress and updates status
func (r *PhaseRepository) RecalculateProgress(phaseID uint) (float64, error) {
	now := time.Now()

	// 1. Recalculate child subphases if any exist
	var subphases []model.DevelopmentSubphase
	allSubphasesDone := true
	hasSubphases := false

	if err := r.db.Where("phase_id = ?", phaseID).Find(&subphases).Error; err == nil && len(subphases) > 0 {
		hasSubphases = true
		for _, sp := range subphases {
			var spTasks []model.DevelopmentTask
			if err := r.db.Where("subphase_id = ?", sp.ID).Find(&spTasks).Error; err == nil && len(spTasks) > 0 {
				var sum float64
				activeCount := 0
				doneCount := 0
				hasBlocker := false
				hasUnfinished := false

				for _, t := range spTasks {
					// PLANNED and SUPERSEDED tasks must not prevent completion or corrupt active denominator
					if t.Status == model.StatusSuperseded || t.Status == model.StatusPlanned {
						continue
					}
					activeCount++
					sum += t.Progress
					if t.Status == model.StatusDone {
						doneCount++
					} else if t.Status == model.StatusBlocked || t.Status == model.StatusFailed || t.Status == model.StatusWaitingApproval {
						hasBlocker = true
					} else { // WORKING, TESTING, PENDING
						hasUnfinished = true
					}
				}

				spProgress := 0.0
				if activeCount > 0 {
					spProgress = sum / float64(activeCount)
				}

				spStatus := "PENDING"
				if activeCount > 0 && doneCount == activeCount && spProgress >= 100.0 && !hasBlocker && !hasUnfinished {
					spStatus = "DONE"
				} else if hasBlocker {
					spStatus = "WORKING"
					allSubphasesDone = false
				} else if spProgress > 0 || doneCount > 0 || hasUnfinished {
					spStatus = "WORKING"
					allSubphasesDone = false
				} else if sp.Status == "PLANNED" && spProgress == 0 {
					spStatus = "PLANNED"
				} else {
					allSubphasesDone = false
				}

				if spStatus != "DONE" && spStatus != "PLANNED" {
					allSubphasesDone = false
				}

				_ = r.db.Model(&model.DevelopmentSubphase{}).Where("id = ?", sp.ID).Updates(map[string]interface{}{
					"progress": spProgress,
					"status":   spStatus,
				}).Error
			} else {
				if sp.Status != "DONE" && sp.Status != "PLANNED" {
					allSubphasesDone = false
				}
			}
		}
	}

	// 2. Recalculate Phase progress from valid active deliverables
	var tasks []model.DevelopmentTask
	if err := r.db.Where("phase_id = ?", phaseID).Find(&tasks).Error; err != nil {
		return 0, err
	}

	var sum float64
	activeTaskCount := 0
	doneCount := 0
	hasBlocker := false
	hasUnfinished := false

	for _, t := range tasks {
		// Superseded and Planned tasks do not corrupt active denominator
		if t.Status == model.StatusSuperseded || t.Status == model.StatusPlanned {
			continue
		}
		activeTaskCount++
		sum += t.Progress
		if t.Status == model.StatusDone {
			doneCount++
		} else if t.Status == model.StatusBlocked || t.Status == model.StatusFailed || t.Status == model.StatusWaitingApproval {
			hasBlocker = true
		} else { // WORKING, TESTING, PENDING
			hasUnfinished = true
		}
	}

	var currentPhase model.DevelopmentPhase
	_ = r.db.Where("id = ?", phaseID).First(&currentPhase).Error

	avgProgress := 0.0
	if activeTaskCount > 0 {
		avgProgress = sum / float64(activeTaskCount)
	} else if currentPhase.Progress > 0 {
		avgProgress = currentPhase.Progress
	}

	phaseStatus := model.PhasePending
	var completedDate *time.Time

	// Phase is COMPLETED if:
	// 1. All active tasks are DONE and avgProgress is 100%
	// 2. No BLOCKED, FAILED, WAITING_APPROVAL, or unfinished required tasks
	// 3. If subphases exist, all active subphases are DONE
	isComplete := (activeTaskCount > 0 && doneCount == activeTaskCount && avgProgress >= 100.0) &&
		!hasBlocker && !hasUnfinished &&
		(!hasSubphases || allSubphasesDone)

	if isComplete {
		phaseStatus = model.PhaseCompleted
		if currentPhase.CompletedDate != nil {
			completedDate = currentPhase.CompletedDate
		} else {
			completedDate = &now
		}
	} else if hasBlocker || hasUnfinished || avgProgress > 0 || doneCount > 0 {
		phaseStatus = model.PhaseWorking
		completedDate = nil
	}

	updates := map[string]interface{}{
		"progress": avgProgress,
		"status":   phaseStatus,
	}
	if phaseStatus == model.PhaseCompleted {
		updates["completed_date"] = completedDate
	} else {
		updates["completed_date"] = nil
	}

	err := r.db.Model(&model.DevelopmentPhase{}).Where("id = ?", phaseID).Updates(updates).Error
	return avgProgress, err
}

// ReconcileAllPhases dynamically recalculates all phases and subphases across the database
func (r *PhaseRepository) ReconcileAllPhases() error {
	var phases []model.DevelopmentPhase
	if err := r.db.Order("order_index ASC").Find(&phases).Error; err != nil {
		return err
	}
	for _, p := range phases {
		_, _ = r.RecalculateProgress(p.ID)
	}
	return nil
}
