package service_test

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"datalogger/internal/model"
	"datalogger/internal/repository"
	"datalogger/internal/service"
	"datalogger/internal/service/notification"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// MockProvider is a test double for notification delivery
type MockProvider struct {
	mu           sync.Mutex
	channelType  model.NotificationChannelType
	sendFunc     func(ctx context.Context, msg notification.NotificationMessage, configJSON string) notification.DeliveryResult
	validateFunc func(configJSON string) error
	sendCount    int32
}

func newMockProvider(chType model.NotificationChannelType) *MockProvider {
	return &MockProvider{
		channelType: chType,
		sendFunc: func(ctx context.Context, msg notification.NotificationMessage, configJSON string) notification.DeliveryResult {
			return notification.DeliveryResult{Success: true, ProviderResponse: "OK: 200"}
		},
		validateFunc: func(configJSON string) error { return nil },
	}
}

func (m *MockProvider) ChannelType() model.NotificationChannelType {
	return m.channelType
}

func (m *MockProvider) ValidateConfig(configJSON string) error {
	if m.validateFunc != nil {
		return m.validateFunc(configJSON)
	}
	return nil
}

func (m *MockProvider) Send(ctx context.Context, msg notification.NotificationMessage, configJSON string) notification.DeliveryResult {
	atomic.AddInt32(&m.sendCount, 1)
	if m.sendFunc != nil {
		return m.sendFunc(ctx, msg, configJSON)
	}
	return notification.DeliveryResult{Success: true, ProviderResponse: "mock delivered"}
}

func setupTestNotificationEnv(t *testing.T) (*gorm.DB, *repository.NotificationRepository, *repository.AlarmRepository, *repository.SystemRepository, *service.NotificationEngine) {
	dsn := fmt.Sprintf("file:mem_notif_db_%d?mode=memory&cache=shared", time.Now().UnixNano()+int64(rand.Intn(100000)))
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
		&model.NotificationChannel{},
		&model.Notification{},
		&model.AuditTrail{},
		&model.SystemLog{},
	)
	if err != nil {
		t.Fatalf("failed migrating test models: %v", err)
	}

	notifRepo := repository.NewNotificationRepository(db)
	alarmRepo := repository.NewAlarmRepository(db)
	systemRepo := repository.NewSystemRepository(db)

	cfg := service.NotificationEngineConfig{
		PollInterval:      50 * time.Millisecond,
		BatchSize:         10,
		DefaultMaxRetries: 3,
		DefaultBackoffSec: 1,
	}

	engine := service.NewNotificationEngine(notifRepo, alarmRepo, systemRepo, cfg)
	return db, notifRepo, alarmRepo, systemRepo, engine
}

func TestNotificationEngine_DispatchAlarmAndIdempotency(t *testing.T) {
	ctx := context.Background()
	db, notifRepo, alarmRepo, systemRepo, engine := setupTestNotificationEnv(t)
	defer engine.Stop()

	// Seed an active channel
	ch := &model.NotificationChannel{
		Name:            "Telegram Ops",
		Type:            model.ChannelTelegram,
		Enabled:         true,
		ConfigJSON:      `{"bot_token":"123","chat_id":"999"}`,
		MinSeverity:     model.SeverityWarning,
		RateLimitPerMin: 60,
		MaxRetries:      3,
	}
	if err := db.Create(ch).Error; err != nil {
		t.Fatalf("failed creating channel: %v", err)
	}

	mockTelegram := newMockProvider(model.ChannelTelegram)
	engine.RegisterProvider(mockTelegram)

	alarmService := service.NewAlarmService(alarmRepo, systemRepo)
	alarmService.SetNotificationEngine(engine)

	alarm := &model.Alarm{
		AlarmCode: "ALM_TEMP_HIGH",
		Severity:  model.SeverityCritical,
		Status:    model.AlarmActive,
		Message:   "Temperature 92C exceeds threshold",
	}
	if err := alarmService.CreateAlarm(ctx, alarm); err != nil {
		t.Fatalf("CreateAlarm failed: %v", err)
	}

	// Wait for async dispatch
	time.Sleep(100 * time.Millisecond)

	// Check delivery queue
	items, total, err := notifRepo.List(ctx, repository.NotificationFilterParams{AlarmID: alarm.ID})
	if err != nil {
		t.Fatalf("List notifications failed: %v", err)
	}
	if total != 1 || len(items) != 1 {
		t.Fatalf("expected 1 notification enqueued, got total=%d items=%d", total, len(items))
	}

	// Test Duplicate Dispatch: Dispatch same alarm again
	engine.DispatchAlarm(ctx, alarm)
	time.Sleep(50 * time.Millisecond)

	items2, total2, _ := notifRepo.List(ctx, repository.NotificationFilterParams{AlarmID: alarm.ID})
	if total2 != 1 || len(items2) != 1 {
		t.Fatalf("idempotency check failed: expected duplicate alarm dispatch to be suppressed, got total=%d", total2)
	}
}

func TestNotificationEngine_QueueSurvivesRestart(t *testing.T) {
	ctx := context.Background()
	db, notifRepo, _, _, engine1 := setupTestNotificationEnv(t)

	ch := &model.NotificationChannel{
		Name:            "Web Alerts",
		Type:            model.ChannelWeb,
		Enabled:         true,
		ConfigJSON:      `{}`,
		MinSeverity:     model.SeverityInfo,
		RateLimitPerMin: 60,
	}
	if err := db.Create(ch).Error; err != nil {
		t.Fatalf("failed creating channel: %v", err)
	}

	// Manually enqueue a PENDING notification into DB before engine starts
	chID := ch.ID
	notif := &model.Notification{
		AlarmID:        999,
		ChannelID:      &chID,
		Channel:        string(model.ChannelWeb),
		Recipient:      "Web Client",
		Subject:        "Restart Recovery Test",
		Message:        "Testing persistent queue survival across restart",
		Status:         model.NotificationPending,
		MaxAttempts:    3,
		IdempotencyKey: "test:restart:999",
	}
	if err := notifRepo.Enqueue(ctx, notif); err != nil {
		t.Fatalf("failed enqueuing initial notification: %v", err)
	}

	// Stop Engine 1
	engine1.Stop()

	// Spin up Engine 2 on the same database
	cfg := service.NotificationEngineConfig{
		PollInterval:      50 * time.Millisecond,
		BatchSize:         10,
		DefaultMaxRetries: 3,
		DefaultBackoffSec: 1,
	}
	engine2 := service.NewNotificationEngine(notifRepo, nil, nil, cfg)

	mockWeb := newMockProvider(model.ChannelWeb)
	engine2.RegisterProvider(mockWeb)

	engine2.Start()
	defer engine2.Stop()

	// Allow engine2 to poll and process recovered queue
	time.Sleep(200 * time.Millisecond)

	saved, err := notifRepo.GetByID(ctx, notif.ID)
	if err != nil {
		t.Fatalf("failed retrieving notification: %v", err)
	}
	if saved.Status != model.NotificationSent {
		t.Fatalf("expected recovered notification to be SENT, got %s (failure: %s)", saved.Status, saved.ErrorMsg)
	}
	if atomic.LoadInt32(&mockWeb.sendCount) < 1 {
		t.Errorf("expected provider to be called for recovered item")
	}
}

func TestNotificationEngine_RetryBackoffAndMaxAttempts(t *testing.T) {
	ctx := context.Background()
	db, notifRepo, _, _, engine := setupTestNotificationEnv(t)
	defer engine.Stop()

	ch := &model.NotificationChannel{
		Name:             "SMTP Retry Test",
		Type:             model.ChannelEmail,
		Enabled:          true,
		ConfigJSON:       `{"smtp_host":"smtp.example.com"}`,
		MinSeverity:      model.SeverityWarning,
		RateLimitPerMin:  60,
		MaxRetries:       2,
		RetryIntervalSec: 1, // 1 sec base
	}
	if err := db.Create(ch).Error; err != nil {
		t.Fatalf("failed creating channel: %v", err)
	}

	mockEmail := newMockProvider(model.ChannelEmail)
	// Always return retryable error
	mockEmail.sendFunc = func(ctx context.Context, msg notification.NotificationMessage, configJSON string) notification.DeliveryResult {
		return notification.DeliveryResult{
			Success:          false,
			ErrorMsg:         "Connection timeout to SMTP gateway",
			Retryable:        true,
			ProviderResponse: "504 Gateway Timeout",
		}
	}
	engine.RegisterProvider(mockEmail)

	chID := ch.ID
	notif := &model.Notification{
		AlarmID:        101,
		ChannelID:      &chID,
		Channel:        string(model.ChannelEmail),
		Recipient:      "ops@example.com",
		Subject:        "Retry Test Alert",
		Message:        "Checking backoff progression",
		Status:         model.NotificationPending,
		MaxAttempts:    2,
		IdempotencyKey: "test:retry:101",
	}
	if err := notifRepo.Enqueue(ctx, notif); err != nil {
		t.Fatalf("Enqueue failed: %v", err)
	}

	engine.Start()

	// 1st attempt: fails, transitions to RETRYING
	time.Sleep(150 * time.Millisecond)

	n1, _ := notifRepo.GetByID(ctx, notif.ID)
	if n1.Status != model.NotificationRetrying {
		t.Fatalf("expected status RETRYING after attempt 1, got %s", n1.Status)
	}
	if n1.Attempts != 1 {
		t.Errorf("expected 1 attempt, got %d", n1.Attempts)
	}
	if n1.NextRetryAt == nil {
		t.Fatalf("expected NextRetryAt to be set for backoff")
	}

	// Manually fast-forward NextRetryAt in DB to simulate passage of backoff time
	past := time.Now().UTC().Add(-1 * time.Minute)
	_ = db.Model(&model.Notification{}).Where("id = ?", notif.ID).Update("next_retry_at", past).Error

	// 2nd attempt: reaches MaxAttempts (2), should transition to terminal FAILED
	time.Sleep(200 * time.Millisecond)

	n2, _ := notifRepo.GetByID(ctx, notif.ID)
	if n2.Status != model.NotificationFailed {
		t.Fatalf("expected status FAILED after reaching max attempts, got %s", n2.Status)
	}
	if n2.Attempts != 2 {
		t.Errorf("expected 2 attempts, got %d", n2.Attempts)
	}
}

func TestNotificationEngine_RateLimiting(t *testing.T) {
	ctx := context.Background()
	db, notifRepo, _, _, engine := setupTestNotificationEnv(t)
	defer engine.Stop()

	// Channel rate limited to 2 per minute
	ch := &model.NotificationChannel{
		Name:            "Rate Limited Webhook",
		Type:            model.ChannelWebhook,
		Enabled:         true,
		ConfigJSON:      `{"webhook_url":"https://example.com/hook"}`,
		MinSeverity:     model.SeverityInfo,
		RateLimitPerMin: 2,
		MaxRetries:      3,
	}
	if err := db.Create(ch).Error; err != nil {
		t.Fatalf("failed creating channel: %v", err)
	}

	mockWebhook := newMockProvider(model.ChannelWebhook)
	engine.RegisterProvider(mockWebhook)

	chID := ch.ID
	// Enqueue 3 notifications
	for i := 1; i <= 3; i++ {
		n := &model.Notification{
			AlarmID:        uint(200 + i),
			ChannelID:      &chID,
			Channel:        string(model.ChannelWebhook),
			Recipient:      "Webhook Endpoint",
			Subject:        fmt.Sprintf("Rate Limit Alert %d", i),
			Message:        "Testing sliding window limiter",
			Status:         model.NotificationPending,
			MaxAttempts:    3,
			IdempotencyKey: fmt.Sprintf("test:rate:%d", i),
		}
		if err := notifRepo.Enqueue(ctx, n); err != nil {
			t.Fatalf("failed enqueuing item %d: %v", i, err)
		}
	}

	engine.Start()
	time.Sleep(250 * time.Millisecond)

	// Since limit is 2/min, exactly 2 should be sent, 1 should be delayed to RETRYING
	sentCount := 0
	retryingCount := 0
	items, _, _ := notifRepo.List(ctx, repository.NotificationFilterParams{})
	for _, it := range items {
		if it.Status == model.NotificationSent {
			sentCount++
		} else if it.Status == model.NotificationRetrying {
			retryingCount++
		}
	}

	if sentCount != 2 {
		t.Errorf("expected exactly 2 SENT items within rate limit, got %d", sentCount)
	}
	if retryingCount != 1 {
		t.Errorf("expected 1 RETRYING item due to rate limit, got %d", retryingCount)
	}
}

func TestNotificationEngine_UnregisteredProviderAndDisabledChannel(t *testing.T) {
	ctx := context.Background()
	db, notifRepo, _, _, engine := setupTestNotificationEnv(t)
	defer engine.Stop()

	// 1. Channel without registered provider
	chUnregistered := &model.NotificationChannel{
		Name:        "Unknown Provider Channel",
		Type:        model.NotificationChannelType("UNKNOWN_TYPE"),
		Enabled:     true,
		ConfigJSON:  `{}`,
		MinSeverity: model.SeverityInfo,
	}
	_ = db.Create(chUnregistered).Error

	chID1 := chUnregistered.ID
	n1 := &model.Notification{
		AlarmID:        301,
		ChannelID:      &chID1,
		Channel:        "UNKNOWN_TYPE",
		Status:         model.NotificationPending,
		IdempotencyKey: "test:unknown:1",
	}
	_ = notifRepo.Enqueue(ctx, n1)

	// 2. Disabled Channel
	chDisabled := &model.NotificationChannel{
		Name:        "Disabled Telegram",
		Type:        model.ChannelTelegram,
		Enabled:     true,
		ConfigJSON:  `{"bot_token":"123"}`,
		MinSeverity: model.SeverityInfo,
	}
	_ = db.Create(chDisabled).Error
	_ = db.Model(chDisabled).Update("enabled", false).Error

	chID2 := chDisabled.ID
	n2 := &model.Notification{
		AlarmID:        302,
		ChannelID:      &chID2,
		Channel:        string(model.ChannelTelegram),
		Status:         model.NotificationPending,
		IdempotencyKey: "test:disabled:2",
	}
	_ = notifRepo.Enqueue(ctx, n2)

	engine.Start()
	time.Sleep(200 * time.Millisecond)

	res1, _ := notifRepo.GetByID(ctx, n1.ID)
	if res1.Status != model.NotificationFailed {
		t.Errorf("expected unhandled channel to be FAILED, got %s", res1.Status)
	}

	res2, _ := notifRepo.GetByID(ctx, n2.ID)
	if res2.Status != model.NotificationCancelled {
		t.Errorf("expected disabled channel notification to be CANCELLED, got %s", res2.Status)
	}
}

func TestNotificationEngine_AlarmDecouplingAndLifecycleIntegrity(t *testing.T) {
	ctx := context.Background()
	db, notifRepo, alarmRepo, systemRepo, engine := setupTestNotificationEnv(t)
	defer engine.Stop()

	ch := &model.NotificationChannel{
		Name:            "Failing Email",
		Type:            model.ChannelEmail,
		Enabled:         true,
		ConfigJSON:      `{"smtp_host":"invalid.smtp"}`,
		MinSeverity:     model.SeverityWarning,
		RateLimitPerMin: 60,
	}
	_ = db.Create(ch).Error

	// Provider fails permanently
	mockEmail := newMockProvider(model.ChannelEmail)
	mockEmail.sendFunc = func(ctx context.Context, msg notification.NotificationMessage, configJSON string) notification.DeliveryResult {
		return notification.DeliveryResult{
			Success:   false,
			ErrorMsg:  "SMTP Authentication Failed: Invalid credentials",
			Retryable: false,
		}
	}
	engine.RegisterProvider(mockEmail)
	engine.Start()

	alarmService := service.NewAlarmService(alarmRepo, systemRepo)
	alarmService.SetNotificationEngine(engine)

	alarm := &model.Alarm{
		AlarmCode: "ALM_CRITICAL_LEAK",
		Severity:  model.SeverityCritical,
		Status:    model.AlarmActive,
		Message:   "Freon leak detected on manifold 3",
	}

	// Create alarm: should succeed without error
	if err := alarmService.CreateAlarm(ctx, alarm); err != nil {
		t.Fatalf("CreateAlarm failed: %v", err)
	}

	// Allow notification dispatch to fail in background
	time.Sleep(200 * time.Millisecond)

	// Verify notification failed
	notifs, _, _ := notifRepo.List(ctx, repository.NotificationFilterParams{AlarmID: alarm.ID})
	if len(notifs) != 1 || notifs[0].Status != model.NotificationFailed {
		t.Fatalf("expected notification to be FAILED, got: %+v", notifs)
	}

	// CRITICAL TEST: Verify the original alarm in DB was NOT deleted, cleared, or altered
	savedAlarm, err := alarmService.GetAlarmByID(ctx, alarm.ID)
	if err != nil {
		t.Fatalf("alarm was lost or deleted: %v", err)
	}
	if savedAlarm.Status != model.AlarmActive {
		t.Errorf("alarm status was altered by notification failure! Expected ACTIVE, got %s", savedAlarm.Status)
	}
	if savedAlarm.AlarmCode != "ALM_CRITICAL_LEAK" {
		t.Errorf("alarm data was corrupted: %s", savedAlarm.AlarmCode)
	}
}

func TestNotificationEngine_ManualRetryAndCancel(t *testing.T) {
	ctx := context.Background()
	_, notifRepo, _, _, engine := setupTestNotificationEnv(t)
	defer engine.Stop()

	n := &model.Notification{
		AlarmID:        500,
		Channel:        string(model.ChannelWeb),
		Recipient:      "Web User",
		Subject:        "Manual Action Test",
		Message:        "Testing manual retry and cancel API",
		Status:         model.NotificationPending,
		IdempotencyKey: "test:manual:500",
	}
	_ = notifRepo.Enqueue(ctx, n)

	// Cancel notification
	cancelled, err := engine.CancelNotification(ctx, n.ID)
	if err != nil {
		t.Fatalf("CancelNotification failed: %v", err)
	}
	if cancelled.Status != model.NotificationCancelled {
		t.Errorf("expected CANCELLED status, got %s", cancelled.Status)
	}

	// Retry notification
	retried, err := engine.RetryNotification(ctx, n.ID)
	if err != nil {
		t.Fatalf("RetryNotification failed: %v", err)
	}
	if retried.Status != model.NotificationPending {
		t.Errorf("expected PENDING status after retry, got %s", retried.Status)
	}
	if retried.Attempts != 0 {
		t.Errorf("expected attempts reset to 0, got %d", retried.Attempts)
	}
}
