package service_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"datalogger/internal/model"
	"datalogger/internal/repository"
	"datalogger/internal/service"
	"datalogger/internal/websocket"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func setupTestAlarmBroadcastEnv(t *testing.T) (*service.AlarmService, *repository.AlarmRepository, *websocket.Hub) {
	dsn := fmt.Sprintf("file:mem_alarm_ws_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		t.Fatalf("failed opening test sqlite db: %v", err)
	}

	err = db.AutoMigrate(
		&model.DeviceType{},
		&model.Device{},
		&model.DeviceConnection{},
		&model.Parameter{},
		&model.Alarm{},
		&model.AuditTrail{},
		&model.SystemLog{},
	)
	if err != nil {
		t.Fatalf("failed migrating test models: %v", err)
	}

	systemRepo := repository.NewSystemRepository(db)
	alarmRepo := repository.NewAlarmRepository(db)
	alarmService := service.NewAlarmService(alarmRepo, systemRepo)

	hub := websocket.NewHub()
	go hub.Run()
	alarmService.SetHub(hub)

	return alarmService, alarmRepo, hub
}

func TestAlarmService_WebSocketBroadcast_CreateAckClear(t *testing.T) {
	alarmService, alarmRepo, _ := setupTestAlarmBroadcastEnv(t)
	ctx := context.Background()

	// 1. Test Alarm Creation and Broadcast
	val := 84.5
	thresh := 55.0
	alarm := &model.Alarm{
		AlarmCode:      "THRESHOLD_CRITICAL_HIGH",
		Severity:       model.SeverityCritical,
		Status:         model.AlarmActive,
		Message:        "PM2.5 exceeded critical upper threshold of 55.0 µg/m³",
		TriggerValue:   &val,
		ThresholdValue: &thresh,
		TriggeredAt:    time.Now().UTC(),
	}

	err := alarmService.CreateAlarm(ctx, alarm)
	if err != nil {
		t.Fatalf("failed to create alarm: %v", err)
	}
	if alarm.ID == 0 {
		t.Fatalf("expected positive alarm ID, got 0")
	}

	// Verify persistence in repo
	persisted, err := alarmRepo.GetByID(ctx, alarm.ID)
	if err != nil {
		t.Fatalf("failed fetching persisted alarm: %v", err)
	}
	if persisted.Status != model.AlarmActive || persisted.Severity != model.SeverityCritical {
		t.Errorf("expected active critical alarm, got status=%s severity=%s", persisted.Status, persisted.Severity)
	}

	// 2. Test Alarm Acknowledgement and Broadcast
	ackAlarm, err := alarmService.AcknowledgeAlarmWithUser(ctx, alarm.ID, nil, "operator1", "Lead Operator", "Investigating field sensor")
	if err != nil {
		t.Fatalf("failed to acknowledge alarm: %v", err)
	}
	if ackAlarm.Status != model.AlarmAcked {
		t.Errorf("expected status %s, got %s", model.AlarmAcked, ackAlarm.Status)
	}
	if ackAlarm.AcknowledgedBy != "operator1" || ackAlarm.Notes != "Investigating field sensor" {
		t.Errorf("unexpected ack metadata: by=%s notes=%s", ackAlarm.AcknowledgedBy, ackAlarm.Notes)
	}

	// 3. Test Alarm Clearing and Broadcast
	clearedAlarm, err := alarmService.ClearAlarm(ctx, alarm.ID)
	if err != nil {
		t.Fatalf("failed to clear alarm: %v", err)
	}
	if clearedAlarm.Status != model.AlarmCleared {
		t.Errorf("expected status %s, got %s", model.AlarmCleared, clearedAlarm.Status)
	}
	if clearedAlarm.ClearedAt == nil {
		t.Errorf("expected non-nil ClearedAt timestamp")
	}

	// 4. Test Nil Hub Resilience (no panic when hub is nil)
	nilHubService := service.NewAlarmService(alarmRepo, nil)
	alarm2 := &model.Alarm{
		AlarmCode:   "THRESHOLD_WARNING_LOW",
		Severity:    model.SeverityWarning,
		Status:      model.AlarmActive,
		Message:     "Temperature warning low",
		TriggeredAt: time.Now().UTC(),
	}
	if err := nilHubService.CreateAlarm(ctx, alarm2); err != nil {
		t.Fatalf("expected nil hub create to succeed, got %v", err)
	}
	if _, err := nilHubService.AcknowledgeAlarmWithUser(ctx, alarm2.ID, nil, "admin", "admin"); err != nil {
		t.Fatalf("expected nil hub ack to succeed, got %v", err)
	}
	if _, err := nilHubService.ClearAlarm(ctx, alarm2.ID); err != nil {
		t.Fatalf("expected nil hub clear to succeed, got %v", err)
	}
}
