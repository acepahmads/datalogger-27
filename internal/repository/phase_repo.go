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
	err := r.db.Preload("Tasks").Order("phase_number ASC").Find(&phases).Error
	return phases, err
}

func (r *PhaseRepository) GetPhaseByID(id uint) (*model.DevelopmentPhase, error) {
	var phase model.DevelopmentPhase
	err := r.db.Preload("Tasks.Logs").Preload("Tasks.Evidences").Preload("Subphases").First(&phase, id).Error
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

// RecalculateProgress recalculates phase progress and updates status
func (r *PhaseRepository) RecalculateProgress(phaseID uint) (float64, error) {
	var tasks []model.DevelopmentTask
	if err := r.db.Where("phase_id = ?", phaseID).Find(&tasks).Error; err != nil {
		return 0, err
	}
	if len(tasks) == 0 {
		return 0, nil
	}

	var totalProgress float64
	doneCount := 0
	for _, t := range tasks {
		totalProgress += t.Progress
		if t.Status == model.StatusDone {
			doneCount++
		}
	}
	avgProgress := totalProgress / float64(len(tasks))

	phaseStatus := model.PhasePending
	now := time.Now()
	var completedDate *time.Time
	if avgProgress >= 100 || doneCount == len(tasks) {
		phaseStatus = model.PhaseCompleted
		completedDate = &now
	} else if avgProgress > 0 || doneCount > 0 {
		phaseStatus = model.PhaseWorking
	}

	err := r.db.Model(&model.DevelopmentPhase{}).Where("id = ?", phaseID).Updates(map[string]interface{}{
		"progress":       avgProgress,
		"status":         phaseStatus,
		"completed_date": completedDate,
	}).Error

	return avgProgress, err
}
