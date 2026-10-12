package service

import (
	"context"
	"fmt"
	"time"

	"datalogger/internal/logger"
	"datalogger/internal/model"
	"datalogger/internal/repository"
)

// AlarmService manages alarm persistence and lifecycle state transitions
type AlarmService struct {
	alarmRepo  *repository.AlarmRepository
	systemRepo *repository.SystemRepository
}

// NewAlarmService creates a new instance of AlarmService
func NewAlarmService(alarmRepo *repository.AlarmRepository, systemRepo *repository.SystemRepository) *AlarmService {
	return &AlarmService{
		alarmRepo:  alarmRepo,
		systemRepo: systemRepo,
	}
}

// CreateAlarm validates and persists a new alarm record
func (s *AlarmService) CreateAlarm(ctx context.Context, alarm *model.Alarm) error {
	if err := s.alarmRepo.Create(ctx, alarm); err != nil {
		return err
	}

	logger.Info("Alarm persisted: [%s] Code=%s DeviceID=%v ParamID=%v Severity=%s Message=%s",
		alarm.Status, alarm.AlarmCode, alarm.DeviceID, alarm.ParameterID, alarm.Severity, alarm.Message)

	if s.systemRepo != nil {
		_ = s.systemRepo.AddAuditTrail(&model.AuditTrail{
			Username:  "system",
			Action:    "ALARM_CREATE",
			Resource:  "ALARM",
			Details:   fmt.Sprintf("Alarm triggered: Code=%s, Severity=%s, Message=%s", alarm.AlarmCode, alarm.Severity, alarm.Message),
			CreatedAt: time.Now().UTC(),
		})
	}

	return nil
}

// GetAlarmByID retrieves an alarm by its unique primary key
func (s *AlarmService) GetAlarmByID(ctx context.Context, id uint) (*model.Alarm, error) {
	return s.alarmRepo.GetByID(ctx, id)
}

// ListAlarms retrieves filtered, paginated alarms with preloaded relations
func (s *AlarmService) ListAlarms(ctx context.Context, filter repository.AlarmFilterParams) ([]model.Alarm, int64, error) {
	return s.alarmRepo.List(ctx, filter)
}

// GetActiveAlarms returns all alarms currently in ACTIVE state
func (s *AlarmService) GetActiveAlarms(ctx context.Context) ([]model.Alarm, error) {
	return s.alarmRepo.GetActive(ctx)
}

// GetAlarmsByStatus returns alarms matching a specific status (e.g. ACTIVE, ACKNOWLEDGED, CLEARED)
func (s *AlarmService) GetAlarmsByStatus(ctx context.Context, status string) ([]model.Alarm, error) {
	filter := repository.AlarmFilterParams{
		Status:   status,
		Page:     1,
		PageSize: 100,
	}
	alarms, _, err := s.alarmRepo.List(ctx, filter)
	return alarms, err
}

// AcknowledgeAlarm handles operator acknowledgement of an active alarm
func (s *AlarmService) AcknowledgeAlarm(ctx context.Context, id uint, username string) (*model.Alarm, error) {
	if username == "" {
		username = "admin"
	}
	ackAt := time.Now().UTC()

	alarm, err := s.alarmRepo.Acknowledge(ctx, id, username, ackAt)
	if err != nil {
		return nil, err
	}

	logger.Info("Alarm %d (%s) acknowledged by operator '%s'", id, alarm.AlarmCode, username)

	if s.systemRepo != nil {
		_ = s.systemRepo.AddAuditTrail(&model.AuditTrail{
			Username:  username,
			Action:    "ALARM_ACKNOWLEDGE",
			Resource:  "ALARM",
			Details:   fmt.Sprintf("Alarm %d (%s) acknowledged by %s", id, alarm.AlarmCode, username),
			CreatedAt: ackAt,
		})
	}

	return alarm, nil
}

// ClearAlarm transitions an active or acknowledged alarm to CLEARED state
func (s *AlarmService) ClearAlarm(ctx context.Context, id uint) (*model.Alarm, error) {
	clearedAt := time.Now().UTC()

	alarm, err := s.alarmRepo.Clear(ctx, id, clearedAt)
	if err != nil {
		return nil, err
	}

	logger.Info("Alarm %d (%s) cleared at %s", id, alarm.AlarmCode, clearedAt.Format(time.RFC3339))

	if s.systemRepo != nil {
		_ = s.systemRepo.AddAuditTrail(&model.AuditTrail{
			Username:  "system",
			Action:    "ALARM_CLEAR",
			Resource:  "ALARM",
			Details:   fmt.Sprintf("Alarm %d (%s) cleared", id, alarm.AlarmCode),
			CreatedAt: clearedAt,
		})
	}

	return alarm, nil
}
