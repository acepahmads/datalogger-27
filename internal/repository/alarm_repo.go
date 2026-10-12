package repository

import (
	"context"
	"errors"
	"strings"
	"time"

	"datalogger/internal/model"

	"gorm.io/gorm"
)

var (
	ErrAlarmNotFound            = errors.New("alarm not found")
	ErrCannotAcknowledgeCleared  = errors.New("cannot acknowledge an alarm that is already cleared")
	ErrAlarmAlreadyAcknowledged  = errors.New("alarm is already acknowledged")
	ErrAlarmAlreadyCleared      = errors.New("alarm is already cleared")
	ErrInvalidAlarmSeverity     = errors.New("invalid alarm severity: must be INFO, WARNING, or CRITICAL")
	ErrMissingAlarmCode         = errors.New("alarm_code is required")
	ErrMissingAlarmMessage      = errors.New("message is required")
)

// AlarmFilterParams defines filtering, search, and pagination criteria for alarms
type AlarmFilterParams struct {
	Status         string
	Severity       string
	DeviceID       uint
	ParameterID    uint
	AlarmCode      string
	AcknowledgedBy string
	Search         string
	StartTime      *time.Time
	EndTime        *time.Time
	Page           int
	PageSize       int
}

// AlarmRepository provides database storage operations for alarm records
type AlarmRepository struct {
	db *gorm.DB
}

// NewAlarmRepository creates a new instance of AlarmRepository
func NewAlarmRepository(db *gorm.DB) *AlarmRepository {
	return &AlarmRepository{db: db}
}

// Create persists a new alarm record to the database
func (r *AlarmRepository) Create(ctx context.Context, alarm *model.Alarm) error {
	if alarm == nil {
		return errors.New("alarm cannot be nil")
	}
	if strings.TrimSpace(alarm.AlarmCode) == "" {
		return ErrMissingAlarmCode
	}
	if strings.TrimSpace(alarm.Message) == "" {
		return ErrMissingAlarmMessage
	}

	if alarm.Severity == "" {
		alarm.Severity = model.SeverityWarning
	}
	switch alarm.Severity {
	case model.SeverityInfo, model.SeverityWarning, model.SeverityCritical:
		// Valid
	default:
		return ErrInvalidAlarmSeverity
	}

	if alarm.Status == "" {
		alarm.Status = model.AlarmActive
	}
	if alarm.TriggeredAt.IsZero() {
		alarm.TriggeredAt = time.Now().UTC()
	}

	return r.db.WithContext(ctx).Create(alarm).Error
}

// GetByID retrieves a single alarm by primary key with relations preloaded
func (r *AlarmRepository) GetByID(ctx context.Context, id uint) (*model.Alarm, error) {
	var alarm model.Alarm
	err := r.db.WithContext(ctx).
		Preload("Device").
		Preload("Parameter").
		First(&alarm, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrAlarmNotFound
		}
		return nil, err
	}
	return &alarm, nil
}

// List retrieves alarms based on filtering, search, and pagination
func (r *AlarmRepository) List(ctx context.Context, params AlarmFilterParams) ([]model.Alarm, int64, error) {
	var alarms []model.Alarm
	var total int64

	query := r.db.WithContext(ctx).Model(&model.Alarm{}).
		Preload("Device").
		Preload("Parameter")

	if params.Status != "" {
		query = query.Where("status = ?", strings.ToUpper(strings.TrimSpace(params.Status)))
	}
	if params.Severity != "" {
		query = query.Where("severity = ?", strings.ToUpper(strings.TrimSpace(params.Severity)))
	}
	if params.DeviceID > 0 {
		query = query.Where("device_id = ?", params.DeviceID)
	}
	if params.ParameterID > 0 {
		query = query.Where("parameter_id = ?", params.ParameterID)
	}
	if params.AlarmCode != "" {
		query = query.Where("alarm_code = ?", strings.TrimSpace(params.AlarmCode))
	}
	if params.AcknowledgedBy != "" {
		ackUser := strings.TrimSpace(params.AcknowledgedBy)
		query = query.Where("acknowledged_by = ? OR acknowledged_by_name LIKE ?", ackUser, "%"+ackUser+"%")
	}
	if params.Search != "" {
		s := "%" + strings.TrimSpace(params.Search) + "%"
		query = query.Where("alarm_code LIKE ? OR message LIKE ? OR notes LIKE ? OR acknowledged_by LIKE ? OR acknowledged_by_name LIKE ?", s, s, s, s, s)
	}
	if params.StartTime != nil && !params.StartTime.IsZero() {
		query = query.Where("triggered_at >= ?", params.StartTime)
	}
	if params.EndTime != nil && !params.EndTime.IsZero() {
		query = query.Where("triggered_at <= ?", params.EndTime)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	page := params.Page
	if page < 1 {
		page = 1
	}
	pageSize := params.PageSize
	if pageSize < 1 {
		pageSize = 50
	}
	if pageSize > 1000 {
		pageSize = 1000
	}
	offset := (page - 1) * pageSize

	err := query.Order("triggered_at DESC, id DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&alarms).Error
	if err != nil {
		return nil, 0, err
	}

	return alarms, total, nil
}

// GetActive retrieves all active alarms ordered by triggered_at descending
func (r *AlarmRepository) GetActive(ctx context.Context) ([]model.Alarm, error) {
	var alarms []model.Alarm
	err := r.db.WithContext(ctx).
		Preload("Device").
		Preload("Parameter").
		Where("status = ?", model.AlarmActive).
		Order("triggered_at DESC").
		Find(&alarms).Error
	return alarms, err
}

// Acknowledge transitions an ACTIVE alarm to ACKNOWLEDGED state
func (r *AlarmRepository) Acknowledge(ctx context.Context, id uint, ackBy string, ackAt time.Time, notes ...string) (*model.Alarm, error) {
	return r.AcknowledgeWithUser(ctx, id, nil, ackBy, ackBy, ackAt, notes...)
}

// AcknowledgeWithUser transitions an ACTIVE alarm to ACKNOWLEDGED state with concurrency control and user tracking
func (r *AlarmRepository) AcknowledgeWithUser(ctx context.Context, id uint, userID *uint, username, displayName string, ackAt time.Time, notes ...string) (*model.Alarm, error) {
	if ackAt.IsZero() {
		ackAt = time.Now().UTC()
	}
	if strings.TrimSpace(username) == "" {
		username = "admin"
	}
	if strings.TrimSpace(displayName) == "" {
		displayName = username
	}

	updates := map[string]interface{}{
		"status":                  model.AlarmAcked,
		"acknowledged_at":         &ackAt,
		"acknowledged_by":         username,
		"acknowledged_by_user_id": userID,
		"acknowledged_by_name":    displayName,
	}
	if len(notes) > 0 && strings.TrimSpace(notes[0]) != "" {
		updates["notes"] = strings.TrimSpace(notes[0])
	}

	// Atomic update: only acknowledge if currently ACTIVE
	res := r.db.WithContext(ctx).Model(&model.Alarm{}).
		Where("id = ? AND status = ?", id, model.AlarmActive).
		Updates(updates)

	if res.Error != nil {
		return nil, res.Error
	}

	if res.RowsAffected == 0 {
		// Inspect state to distinguish between not found, already acked, or cleared
		var current model.Alarm
		if err := r.db.WithContext(ctx).First(&current, id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, ErrAlarmNotFound
			}
			return nil, err
		}
		if current.Status == model.AlarmCleared {
			return nil, ErrCannotAcknowledgeCleared
		}
		if current.Status == model.AlarmAcked {
			return nil, ErrAlarmAlreadyAcknowledged
		}
		return nil, errors.New("alarm is not in an acknowledgeable state")
	}

	return r.GetByID(ctx, id)
}

// GetActiveOrAcked retrieves all alarms currently in ACTIVE or ACKNOWLEDGED state
func (r *AlarmRepository) GetActiveOrAcked(ctx context.Context) ([]model.Alarm, error) {
	var alarms []model.Alarm
	err := r.db.WithContext(ctx).
		Preload("Device").
		Preload("Parameter").
		Where("status IN ?", []model.AlarmStatus{model.AlarmActive, model.AlarmAcked}).
		Order("triggered_at DESC").
		Find(&alarms).Error
	return alarms, err
}

// Clear transitions an ACTIVE or ACKNOWLEDGED alarm to CLEARED state
func (r *AlarmRepository) Clear(ctx context.Context, id uint, clearedAt time.Time) (*model.Alarm, error) {
	if clearedAt.IsZero() {
		clearedAt = time.Now().UTC()
	}

	// Atomic update: only clear if not already CLEARED
	res := r.db.WithContext(ctx).Model(&model.Alarm{}).
		Where("id = ? AND status != ?", id, model.AlarmCleared).
		Updates(map[string]interface{}{
			"status":     model.AlarmCleared,
			"cleared_at": &clearedAt,
		})

	if res.Error != nil {
		return nil, res.Error
	}

	if res.RowsAffected == 0 {
		var current model.Alarm
		if err := r.db.WithContext(ctx).First(&current, id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, ErrAlarmNotFound
			}
			return nil, err
		}
		if current.Status == model.AlarmCleared {
			return nil, ErrAlarmAlreadyCleared
		}
	}

	return r.GetByID(ctx, id)
}
