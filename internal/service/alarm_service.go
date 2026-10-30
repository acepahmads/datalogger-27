package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"datalogger/internal/logger"
	"datalogger/internal/model"
	"datalogger/internal/repository"
	"datalogger/internal/websocket"
)

// AlarmService manages alarm persistence and lifecycle state transitions
type AlarmService struct {
	alarmRepo          *repository.AlarmRepository
	systemRepo         *repository.SystemRepository
	notificationEngine *NotificationEngine
	alarmEvaluator     *AlarmEvaluator
	hub                *websocket.Hub
}

// NewAlarmService creates a new instance of AlarmService
func NewAlarmService(alarmRepo *repository.AlarmRepository, systemRepo *repository.SystemRepository) *AlarmService {
	return &AlarmService{
		alarmRepo:  alarmRepo,
		systemRepo: systemRepo,
	}
}

// SetHub injects the WebSocket hub for real-time alarm broadcasting
func (s *AlarmService) SetHub(hub *websocket.Hub) {
	s.hub = hub
}

// SetNotificationEngine injects the Phase 6 NotificationEngine
func (s *AlarmService) SetNotificationEngine(ne *NotificationEngine) {
	s.notificationEngine = ne
}

// SetAlarmEvaluator injects the in-memory AlarmEvaluator so active caches stay synced upon alarm deletion/purge
func (s *AlarmService) SetAlarmEvaluator(ae *AlarmEvaluator) {
	s.alarmEvaluator = ae
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

	// Broadcast alarm.created over WebSocket for real-time global banners and alerts
	if s.hub != nil {
		s.hub.Broadcast(websocket.WSMessage{
			Type:      "alarm.created",
			Timestamp: time.Now().UTC(),
			Data:      alarm,
		})
	}

	// Phase 6 Wave 3: Asynchronously dispatch to Notification Engine without blocking alarm evaluation
	if s.notificationEngine != nil {
		go s.notificationEngine.DispatchAlarm(context.Background(), alarm)
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

// GetActiveOrAckedAlarms returns all alarms currently in ACTIVE or ACKNOWLEDGED state
func (s *AlarmService) GetActiveOrAckedAlarms(ctx context.Context) ([]model.Alarm, error) {
	return s.alarmRepo.GetActiveOrAcked(ctx)
}

// AcknowledgeAlarm handles operator acknowledgement of an active alarm with optional operator notes
func (s *AlarmService) AcknowledgeAlarm(ctx context.Context, id uint, username string, notes ...string) (*model.Alarm, error) {
	return s.AcknowledgeAlarmWithUser(ctx, id, nil, username, username, notes...)
}

// AcknowledgeAlarmWithUser handles operator acknowledgement with user ID, username, display name, and optional notes
func (s *AlarmService) AcknowledgeAlarmWithUser(ctx context.Context, id uint, userID *uint, username, displayName string, notes ...string) (*model.Alarm, error) {
	if strings.TrimSpace(username) == "" {
		username = "admin"
	}
	if strings.TrimSpace(displayName) == "" {
		displayName = username
	}
	ackAt := time.Now().UTC()

	alarm, err := s.alarmRepo.AcknowledgeWithUser(ctx, id, userID, username, displayName, ackAt, notes...)
	if err != nil {
		return nil, err
	}

	noteStr := ""
	if len(notes) > 0 && strings.TrimSpace(notes[0]) != "" {
		noteStr = fmt.Sprintf(" Notes: '%s'", strings.TrimSpace(notes[0]))
	}

	logger.Info("Alarm %d (%s) acknowledged by operator '%s' (%s)%s", id, alarm.AlarmCode, username, displayName, noteStr)

	if s.systemRepo != nil {
		details := fmt.Sprintf("Alarm %d (%s) acknowledged by %s", id, alarm.AlarmCode, username)
		if displayName != "" && displayName != username {
			details = fmt.Sprintf("Alarm %d (%s) acknowledged by %s (%s)", id, alarm.AlarmCode, displayName, username)
		}
		if noteStr != "" {
			details += noteStr
		}
		_ = s.systemRepo.AddAuditTrail(&model.AuditTrail{
			UserID:    userID,
			Username:  username,
			Action:    "ALARM_ACKNOWLEDGE",
			Resource:  "ALARM",
			Details:   details,
			CreatedAt: ackAt,
		})
	}

	if s.hub != nil {
		s.hub.Broadcast(websocket.WSMessage{
			Type:      "alarm.acknowledged",
			Timestamp: ackAt,
			Data:      alarm,
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

	if s.hub != nil {
		s.hub.Broadcast(websocket.WSMessage{
			Type:      "alarm.cleared",
			Timestamp: clearedAt,
			Data:      alarm,
		})
	}

	return alarm, nil
}

// ClearDeviceAlarms clears all active and acknowledged alarms of a specific code for a device and broadcasts resolution
func (s *AlarmService) ClearDeviceAlarms(ctx context.Context, deviceID uint, alarmCode string) error {
	clearedAt := time.Now().UTC()
	clearedList, err := s.alarmRepo.ClearActiveByDeviceAndCode(ctx, deviceID, alarmCode, clearedAt)
	if err != nil {
		return err
	}

	for i := range clearedList {
		alarmCopy := clearedList[i]
		alarmCopy.Status = model.AlarmCleared
		alarmCopy.ClearedAt = &clearedAt

		logger.Info("Device alarm %d (%s) for device %d cleared upon communication recovery", alarmCopy.ID, alarmCode, deviceID)

		if s.systemRepo != nil {
			_ = s.systemRepo.AddAuditTrail(&model.AuditTrail{
				Username:  "system",
				Action:    "ALARM_CLEAR",
				Resource:  "ALARM",
				Details:   fmt.Sprintf("Alarm %d (%s) auto-cleared upon recovery", alarmCopy.ID, alarmCopy.AlarmCode),
				CreatedAt: clearedAt,
			})
		}

		if s.hub != nil {
			s.hub.Broadcast(websocket.WSMessage{
				Type:      "alarm.cleared",
				Timestamp: clearedAt,
				Data:      &alarmCopy,
			})
		}
	}
	return nil
}

// ClearAllDeviceAlarms clears all active and acknowledged alarms for a device (e.g. when disabled or deleted)
func (s *AlarmService) ClearAllDeviceAlarms(ctx context.Context, deviceID uint) error {
	clearedAt := time.Now().UTC()
	clearedList, err := s.alarmRepo.ClearActiveByDevice(ctx, deviceID, clearedAt)
	if err != nil {
		return err
	}

	for i := range clearedList {
		alarmCopy := clearedList[i]
		alarmCopy.Status = model.AlarmCleared
		alarmCopy.ClearedAt = &clearedAt

		logger.Info("Device alarm %d (%s) for device %d auto-cleared upon device disablement", alarmCopy.ID, alarmCopy.AlarmCode, deviceID)

		if s.systemRepo != nil {
			_ = s.systemRepo.AddAuditTrail(&model.AuditTrail{
				Username:  "system",
				Action:    "ALARM_CLEAR",
				Resource:  "ALARM",
				Details:   fmt.Sprintf("Alarm %d (%s) auto-cleared upon device disablement", alarmCopy.ID, alarmCopy.AlarmCode),
				CreatedAt: clearedAt,
			})
		}

		if s.hub != nil {
			s.hub.Broadcast(websocket.WSMessage{
				Type:      "alarm.cleared",
				Timestamp: clearedAt,
				Data:      &alarmCopy,
			})
		}
	}
	return nil
}

// ClearAllParameterAlarms clears all active and acknowledged alarms for a parameter (e.g. when disabled or deleted)
func (s *AlarmService) ClearAllParameterAlarms(ctx context.Context, paramID uint) error {
	clearedAt := time.Now().UTC()
	clearedList, err := s.alarmRepo.ClearActiveByParameter(ctx, paramID, clearedAt)
	if err != nil {
		return err
	}

	for i := range clearedList {
		alarmCopy := clearedList[i]
		alarmCopy.Status = model.AlarmCleared
		alarmCopy.ClearedAt = &clearedAt

		logger.Info("Parameter alarm %d (%s) for param %d auto-cleared upon disablement", alarmCopy.ID, alarmCopy.AlarmCode, paramID)

		if s.systemRepo != nil {
			_ = s.systemRepo.AddAuditTrail(&model.AuditTrail{
				Username:  "system",
				Action:    "ALARM_CLEAR",
				Resource:  "ALARM",
				Details:   fmt.Sprintf("Alarm %d (%s) auto-cleared upon parameter disablement", alarmCopy.ID, alarmCopy.AlarmCode),
				CreatedAt: clearedAt,
			})
		}

		if s.hub != nil {
			s.hub.Broadcast(websocket.WSMessage{
				Type:      "alarm.cleared",
				Timestamp: clearedAt,
				Data:      &alarmCopy,
			})
		}
	}
	return nil
}

// DeleteAlarm permanently removes a single alarm record (and dependent notification entries),
// syncing in-memory active alarm caches and audit logs.
func (s *AlarmService) DeleteAlarm(ctx context.Context, id uint, userID *uint, username string) (*model.Alarm, error) {
	if strings.TrimSpace(username) == "" {
		username = "admin"
	}

	deleted, err := s.alarmRepo.DeleteAlarm(ctx, id)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	logger.Info("Alarm %d (%s, status=%s) deleted permanently by '%s'", deleted.ID, deleted.AlarmCode, deleted.Status, username)

	if s.systemRepo != nil {
		_ = s.systemRepo.AddAuditTrail(&model.AuditTrail{
			UserID:    userID,
			Username:  username,
			Action:    "ALARM_DELETE",
			Resource:  "ALARM",
			Details:   fmt.Sprintf("Alarm %d (%s, status=%s) deleted permanently by %s", deleted.ID, deleted.AlarmCode, deleted.Status, username),
			CreatedAt: now,
		})
	}

	if deleted.Status == model.AlarmActive || deleted.Status == model.AlarmAcked {
		if s.alarmEvaluator != nil {
			_ = s.alarmEvaluator.ReloadActiveAlarms(ctx)
		}
		if s.hub != nil {
			s.hub.Broadcast(websocket.WSMessage{
				Type:      "alarm.cleared",
				Timestamp: now,
				Data:      map[string]interface{}{"id": deleted.ID},
			})
		}
	}

	return deleted, nil
}

// PurgeAlarms permanently deletes alarm records matching filter (or all alarms when unfiltered),
// syncing in-memory active alarm caches and audit logs.
func (s *AlarmService) PurgeAlarms(ctx context.Context, filter repository.AlarmFilterParams, userID *uint, username string) (int64, error) {
	if strings.TrimSpace(username) == "" {
		username = "admin"
	}

	deletedCount, activeDeletedIDs, err := s.alarmRepo.PurgeAlarms(ctx, filter)
	if err != nil {
		return 0, err
	}

	now := time.Now().UTC()
	logger.Info("Purged %d alarm record(s) (active_affected=%d) by '%s'", deletedCount, len(activeDeletedIDs), username)

	if s.systemRepo != nil {
		_ = s.systemRepo.AddAuditTrail(&model.AuditTrail{
			UserID:    userID,
			Username:  username,
			Action:    "ALARM_PURGE",
			Resource:  "ALARM",
			Details:   fmt.Sprintf("Purged %d alarm history record(s) by %s", deletedCount, username),
			CreatedAt: now,
		})
	}

	if len(activeDeletedIDs) > 0 {
		if s.alarmEvaluator != nil {
			_ = s.alarmEvaluator.ReloadActiveAlarms(ctx)
		}
		if s.hub != nil {
			for _, activeID := range activeDeletedIDs {
				s.hub.Broadcast(websocket.WSMessage{
					Type:      "alarm.cleared",
					Timestamp: now,
					Data:      map[string]interface{}{"id": activeID},
				})
			}
		}
	}

	return deletedCount, nil
}
