package repository

import (
	"time"

	"datalogger/internal/model"

	"gorm.io/gorm"
)

type ActuatorRepository struct {
	db *gorm.DB
}

func NewActuatorRepository(db *gorm.DB) *ActuatorRepository {
	return &ActuatorRepository{db: db}
}

// -------------------------------------------------------------
// Actuator Operations
// -------------------------------------------------------------

func (r *ActuatorRepository) CreateActuator(actuator *model.DeviceActuator) error {
	wantWatchdog := actuator.SafetyWatchdogActive
	wantEnabled := actuator.Enabled

	if err := r.db.Omit("Device", "FeedbackParameter", "Schedules").Create(actuator).Error; err != nil {
		return err
	}

	// Persist explicit false booleans that GORM default:true tags would otherwise override on INSERT
	if !wantWatchdog || !wantEnabled {
		updates := map[string]interface{}{
			"safety_watchdog_active": wantWatchdog,
			"enabled":                wantEnabled,
		}
		if err := r.db.Model(&model.DeviceActuator{}).Where("id = ?", actuator.ID).Updates(updates).Error; err != nil {
			return err
		}
		actuator.SafetyWatchdogActive = wantWatchdog
		actuator.Enabled = wantEnabled
	}
	return nil
}

func (r *ActuatorRepository) UpdateActuator(actuator *model.DeviceActuator) error {
	return r.db.Omit("Device", "FeedbackParameter", "Schedules").Save(actuator).Error
}

func (r *ActuatorRepository) DeleteActuator(id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("actuator_id = ?", id).Delete(&model.ActuationSchedule{}).Error; err != nil {
			return err
		}
		if err := tx.Where("actuator_id = ?", id).Delete(&model.ActuationLog{}).Error; err != nil {
			return err
		}
		return tx.Delete(&model.DeviceActuator{}, id).Error
	})
}

func (r *ActuatorRepository) GetActuatorByID(id uint) (*model.DeviceActuator, error) {
	var act model.DeviceActuator
	err := r.db.Preload("Device").Preload("Device.Connection").Preload("FeedbackParameter").Preload("Schedules").First(&act, id).Error
	if err != nil {
		return nil, err
	}
	return &act, nil
}

func (r *ActuatorRepository) GetActuatorByCode(code string) (*model.DeviceActuator, error) {
	var act model.DeviceActuator
	err := r.db.Preload("Device").Preload("Device.Connection").Preload("FeedbackParameter").Preload("Schedules").Where("actuator_code = ?", code).First(&act).Error
	if err != nil {
		return nil, err
	}
	return &act, nil
}

func (r *ActuatorRepository) ListActuatorsByDevice(deviceID uint) ([]model.DeviceActuator, error) {
	var list []model.DeviceActuator
	err := r.db.Preload("FeedbackParameter").Preload("Schedules").Where("device_id = ?", deviceID).Order("id ASC").Find(&list).Error
	return list, err
}

func (r *ActuatorRepository) ListAllActiveActuators() ([]model.DeviceActuator, error) {
	var list []model.DeviceActuator
	err := r.db.Preload("Device").Preload("Device.Connection").Preload("FeedbackParameter").Preload("Schedules").Where("enabled = ?", true).Order("device_id ASC, id ASC").Find(&list).Error
	return list, err
}

func (r *ActuatorRepository) UpdateActuatorState(id uint, state model.ActuatorState, autoOffDueAt *time.Time, actuatedBy, lastStatus string) error {
	now := time.Now().UTC()
	updates := map[string]interface{}{
		"current_state":          state,
		"state_changed_at":       &now,
		"auto_off_due_at":        autoOffDueAt,
		"last_actuated_by":       actuatedBy,
		"last_actuation_status":  lastStatus,
		"updated_at":             now,
	}
	return r.db.Model(&model.DeviceActuator{}).Where("id = ?", id).Updates(updates).Error
}

func (r *ActuatorRepository) ListActuatorsDueForAutoOff(now time.Time) ([]model.DeviceActuator, error) {
	var list []model.DeviceActuator
	err := r.db.Preload("Device").Preload("Device.Connection").
		Where("enabled = ? AND current_state = ? AND auto_off_due_at IS NOT NULL AND auto_off_due_at <= ?", true, model.ActuatorStateOn, now).
		Find(&list).Error
	return list, err
}

// -------------------------------------------------------------
// Schedule Operations
// -------------------------------------------------------------

func (r *ActuatorRepository) CreateSchedule(schedule *model.ActuationSchedule) error {
	wantEnabled := schedule.Enabled
	if err := r.db.Omit("Actuator").Create(schedule).Error; err != nil {
		return err
	}
	if !wantEnabled {
		if err := r.db.Model(&model.ActuationSchedule{}).Where("id = ?", schedule.ID).Update("enabled", false).Error; err != nil {
			return err
		}
		schedule.Enabled = false
	}
	return nil
}

func (r *ActuatorRepository) UpdateSchedule(schedule *model.ActuationSchedule) error {
	return r.db.Omit("Actuator").Save(schedule).Error
}

func (r *ActuatorRepository) DeleteSchedule(id uint) error {
	return r.db.Delete(&model.ActuationSchedule{}, id).Error
}

func (r *ActuatorRepository) GetScheduleByID(id uint) (*model.ActuationSchedule, error) {
	var sched model.ActuationSchedule
	err := r.db.Preload("Actuator").Preload("Actuator.Device").Preload("Actuator.Device.Connection").First(&sched, id).Error
	if err != nil {
		return nil, err
	}
	return &sched, nil
}

func (r *ActuatorRepository) ListSchedulesByActuator(actuatorID uint) ([]model.ActuationSchedule, error) {
	var list []model.ActuationSchedule
	err := r.db.Where("actuator_id = ?", actuatorID).Order("id ASC").Find(&list).Error
	return list, err
}

func (r *ActuatorRepository) ListDueSchedules(now time.Time) ([]model.ActuationSchedule, error) {
	var list []model.ActuationSchedule
	err := r.db.Preload("Actuator").Preload("Actuator.Device").Preload("Actuator.Device.Connection").
		Where("enabled = ? AND (next_run_at IS NULL OR next_run_at <= ?)", true, now).
		Find(&list).Error
	return list, err
}

func (r *ActuatorRepository) UpdateScheduleRun(id uint, lastRunAt, nextRunAt time.Time, status string) error {
	now := time.Now().UTC()
	updates := map[string]interface{}{
		"last_run_at": &lastRunAt,
		"next_run_at": &nextRunAt,
		"last_status": status,
		"updated_at":  now,
	}
	return r.db.Model(&model.ActuationSchedule{}).Where("id = ?", id).Updates(updates).Error
}

// -------------------------------------------------------------
// Actuation Log Operations
// -------------------------------------------------------------

func (r *ActuatorRepository) CreateLog(log *model.ActuationLog) error {
	if log.ExecutedAt.IsZero() {
		log.ExecutedAt = time.Now().UTC()
	}
	return r.db.Create(log).Error
}

func (r *ActuatorRepository) ListLogs(actuatorID *uint, deviceID *uint, limit int) ([]model.ActuationLog, error) {
	if limit <= 0 || limit > 500 {
		limit = 50
	}
	query := r.db.Order("executed_at DESC").Limit(limit)
	if actuatorID != nil && *actuatorID > 0 {
		query = query.Where("actuator_id = ?", *actuatorID)
	}
	if deviceID != nil && *deviceID > 0 {
		query = query.Where("device_id = ?", *deviceID)
	}
	var logs []model.ActuationLog
	err := query.Find(&logs).Error
	return logs, err
}

