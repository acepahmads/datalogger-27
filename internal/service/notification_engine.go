package service

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"datalogger/internal/logger"
	"datalogger/internal/model"
	"datalogger/internal/repository"
	"datalogger/internal/service/notification"
)

// ChannelRateLimiter tracks per-channel dispatch timestamps within a rolling window
type ChannelRateLimiter struct {
	mu         sync.Mutex
	timestamps map[uint][]time.Time
}

func newChannelRateLimiter() *ChannelRateLimiter {
	return &ChannelRateLimiter{
		timestamps: make(map[uint][]time.Time),
	}
}

func (rl *ChannelRateLimiter) Allow(channelID uint, limitPerMin int) bool {
	if limitPerMin <= 0 {
		return true
	}

	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now().UTC()
	windowStart := now.Add(-60 * time.Second)

	// Filter timestamps within window
	valid := make([]time.Time, 0, len(rl.timestamps[channelID]))
	for _, t := range rl.timestamps[channelID] {
		if t.After(windowStart) {
			valid = append(valid, t)
		}
	}
	rl.timestamps[channelID] = valid

	if len(valid) >= limitPerMin {
		return false
	}

	rl.timestamps[channelID] = append(rl.timestamps[channelID], now)
	return true
}

// NotificationEngineConfig provides runtime tuning parameters for the dispatcher
type NotificationEngineConfig struct {
	PollInterval       time.Duration
	BatchSize          int
	DefaultMaxRetries  int
	DefaultBackoffSec  int
}

// DefaultNotificationEngineConfig returns standard production settings
func DefaultNotificationEngineConfig() NotificationEngineConfig {
	return NotificationEngineConfig{
		PollInterval:      1 * time.Second,
		BatchSize:         20,
		DefaultMaxRetries: 3,
		DefaultBackoffSec: 15,
	}
}

// NotificationEngine coordinates asynchronous, persistent notification delivery with retries, rate limiting, and redaction
type NotificationEngine struct {
	repo        *repository.NotificationRepository
	alarmRepo   *repository.AlarmRepository
	systemRepo  *repository.SystemRepository
	cfg         NotificationEngineConfig
	providers   map[model.NotificationChannelType]notification.ChannelProvider
	rateLimiter *ChannelRateLimiter
	notifyCh    chan struct{}
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
	mu          sync.RWMutex

	// Operational metrics
	dispatchedCount uint64
	deliveredCount  uint64
	failedCount     uint64
	retriedCount    uint64
	cancelledCount  uint64
}

// NewNotificationEngine constructs and initializes the notification engine
func NewNotificationEngine(
	repo *repository.NotificationRepository,
	alarmRepo *repository.AlarmRepository,
	systemRepo *repository.SystemRepository,
	cfg NotificationEngineConfig,
) *NotificationEngine {
	if cfg.PollInterval <= 0 {
		cfg.PollInterval = 1 * time.Second
	}
	if cfg.BatchSize <= 0 {
		cfg.BatchSize = 20
	}
	if cfg.DefaultMaxRetries <= 0 {
		cfg.DefaultMaxRetries = 3
	}
	if cfg.DefaultBackoffSec <= 0 {
		cfg.DefaultBackoffSec = 15
	}

	ctx, cancel := context.WithCancel(context.Background())

	engine := &NotificationEngine{
		repo:        repo,
		alarmRepo:   alarmRepo,
		systemRepo:  systemRepo,
		cfg:         cfg,
		providers:   make(map[model.NotificationChannelType]notification.ChannelProvider),
		rateLimiter: newChannelRateLimiter(),
		notifyCh:    make(chan struct{}, 100),
		ctx:         ctx,
		cancel:      cancel,
	}

	// Register standard channel providers
	engine.RegisterProvider(notification.NewWebProvider(nil))
	engine.RegisterProvider(notification.NewEmailProvider())
	engine.RegisterProvider(notification.NewTelegramProvider())
	engine.RegisterProvider(notification.NewWebhookProvider())

	return engine
}

// RegisterProvider registers or replaces a channel provider adapter
func (e *NotificationEngine) RegisterProvider(p notification.ChannelProvider) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.providers[p.ChannelType()] = p
}

// GetProvider retrieves a registered provider by channel type
func (e *NotificationEngine) GetProvider(chType model.NotificationChannelType) notification.ChannelProvider {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.providers[chType]
}

// Start launches the background persistent queue worker
func (e *NotificationEngine) Start() {
	e.wg.Add(1)
	go e.runDispatcher()
	logger.Info("NotificationEngine background dispatcher started (poll interval: %v)", e.cfg.PollInterval)
}

// Stop initiates graceful shutdown of the background dispatcher
func (e *NotificationEngine) Stop() {
	e.cancel()
	e.wg.Wait()
	logger.Info("NotificationEngine background dispatcher stopped gracefully")
}

// DispatchAlarm inspects an active alarm, resolves eligible channels, and enqueues notifications asynchronously
func (e *NotificationEngine) DispatchAlarm(ctx context.Context, alarm *model.Alarm) {
	if alarm == nil || e.repo == nil {
		return
	}

	channels, err := e.repo.ListChannels(ctx)
	if err != nil {
		logger.Error("NotificationEngine failed querying active channels: %v", err)
		return
	}

	enqueuedAny := false
	for _, ch := range channels {
		if !ch.Enabled {
			continue
		}

		// Severity filtering
		if !isSeverityEligible(ch.MinSeverity, alarm.Severity) {
			continue
		}

		// Idempotency key per alarm and channel: prevents duplicate deliveries for the same event
		idempotencyKey := fmt.Sprintf("alarm:%d:channel:%d", alarm.ID, ch.ID)

		// Check if already enqueued
		existing, _ := e.repo.FindByIdempotencyKey(ctx, idempotencyKey)
		if existing != nil {
			continue
		}

		subject := fmt.Sprintf("[%s] %s Alarm: %s", alarm.Severity, ch.Name, alarm.AlarmCode)
		chID := ch.ID
		maxRetries := ch.MaxRetries
		if maxRetries <= 0 {
			maxRetries = e.cfg.DefaultMaxRetries
		}

		recipient := ""
		// Extract default recipient from channel config if applicable
		if ch.Type == model.ChannelEmail || ch.Type == model.ChannelTelegram || ch.Type == model.ChannelWebhook {
			recipient = ch.Name
		}

		notif := &model.Notification{
			AlarmID:        alarm.ID,
			ChannelID:      &chID,
			Channel:        string(ch.Type),
			Recipient:      recipient,
			Subject:        subject,
			Message:        alarm.Message,
			Status:         model.NotificationPending,
			MaxAttempts:    maxRetries,
			IdempotencyKey: idempotencyKey,
			CreatedAt:      time.Now().UTC(),
		}

		if err := e.repo.Enqueue(ctx, notif); err != nil {
			if err != repository.ErrDuplicateNotification {
				logger.Error("NotificationEngine failed enqueuing notification for alarm %d to channel %s: %v", alarm.ID, ch.Name, err)
			}
			continue
		}

		atomic.AddUint64(&e.dispatchedCount, 1)
		enqueuedAny = true
	}

	if enqueuedAny {
		// Signal worker to immediately process queue without waiting for next poll tick
		select {
		case e.notifyCh <- struct{}{}:
		default:
		}
	}
}

// runDispatcher is the central queue consumer loop
func (e *NotificationEngine) runDispatcher() {
	defer e.wg.Done()
	ticker := time.NewTicker(e.cfg.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-e.ctx.Done():
			return
		case <-ticker.C:
			e.processQueue()
		case <-e.notifyCh:
			e.processQueue()
		}
	}
}

// processQueue fetches pending/retrying notifications and delivers them through providers
func (e *NotificationEngine) processQueue() {
	if e.repo == nil {
		return
	}

	pending, err := e.repo.FindPending(e.ctx, e.cfg.BatchSize)
	if err != nil {
		logger.Error("NotificationEngine error fetching pending notifications: %v", err)
		return
	}

	for i := range pending {
		item := &pending[i]
		e.deliverNotification(item)
	}
}

// deliverNotification delivers an individual notification item
func (e *NotificationEngine) deliverNotification(n *model.Notification) {
	ctx, cancel := context.WithTimeout(e.ctx, 15*time.Second)
	defer cancel()

	// 1. Resolve Channel
	var channel *model.NotificationChannel
	if n.ChannelID != nil {
		ch, err := e.repo.GetChannelByID(ctx, *n.ChannelID)
		if err == nil {
			channel = ch
		}
	}
	if channel == nil {
		ch, err := e.repo.GetChannelByType(ctx, model.NotificationChannelType(n.Channel))
		if err == nil {
			channel = ch
		}
	}

	if channel == nil {
		_ = e.repo.UpdateDeliveryAttempt(ctx, n.ID, model.NotificationFailed, n.Attempts+1, nil, "Notification channel does not exist", "", nil)
		atomic.AddUint64(&e.failedCount, 1)
		return
	}

	if !channel.Enabled {
		_ = e.repo.UpdateDeliveryAttempt(ctx, n.ID, model.NotificationCancelled, n.Attempts, nil, "Channel disabled by administrator", "", nil)
		atomic.AddUint64(&e.cancelledCount, 1)
		return
	}

	// 2. Quiet Hours Check
	if e.isQuietHours(channel) {
		// Non-critical alarms are held during quiet hours
		if n.Alarm == nil || n.Alarm.Severity != model.SeverityCritical {
			nextTime := time.Now().UTC().Add(30 * time.Minute)
			_ = e.repo.UpdateDeliveryAttempt(ctx, n.ID, model.NotificationRetrying, n.Attempts, &nextTime, "Held during configured quiet hours", "", nil)
			return
		}
	}

	// 3. Rate Limit Check
	limitPerMin := channel.RateLimitPerMin
	if limitPerMin <= 0 {
		limitPerMin = 10
	}
	if !e.rateLimiter.Allow(channel.ID, limitPerMin) {
		nextTime := time.Now().UTC().Add(10 * time.Second)
		_ = e.repo.UpdateDeliveryAttempt(ctx, n.ID, model.NotificationRetrying, n.Attempts, &nextTime, fmt.Sprintf("Rate limit exceeded (%d msg/min); delaying delivery", limitPerMin), "", nil)
		atomic.AddUint64(&e.retriedCount, 1)
		return
	}

	// 4. Resolve Provider
	e.mu.RLock()
	provider, exists := e.providers[channel.Type]
	e.mu.RUnlock()

	if !exists {
		_ = e.repo.UpdateDeliveryAttempt(ctx, n.ID, model.NotificationFailed, n.Attempts+1, nil, fmt.Sprintf("No provider registered for channel type %s", channel.Type), "", nil)
		atomic.AddUint64(&e.failedCount, 1)
		return
	}

	// Mark SENDING
	_ = e.repo.UpdateDeliveryAttempt(ctx, n.ID, model.NotificationSending, n.Attempts, nil, "", "", nil)

	// 5. Build Message
	devCode := ""
	paramCode := ""
	sev := model.SeverityWarning
	alarmCode := ""
	if n.Alarm != nil {
		sev = n.Alarm.Severity
		alarmCode = n.Alarm.AlarmCode
		if n.Alarm.Device != nil {
			devCode = n.Alarm.Device.DeviceCode
		}
		if n.Alarm.Parameter != nil {
			paramCode = n.Alarm.Parameter.ParameterCode
		}
	}

	msg := notification.NotificationMessage{
		ID:        n.ID,
		AlarmID:   n.AlarmID,
		Severity:  sev,
		AlarmCode: alarmCode,
		Device:    devCode,
		Parameter: paramCode,
		Subject:   n.Subject,
		Body:      n.Message,
		Recipient: n.Recipient,
	}

	// 6. Execute Send
	res := provider.Send(ctx, msg, channel.ConfigJSON)
	now := time.Now().UTC()

	if res.Success {
		_ = e.repo.UpdateDeliveryAttempt(ctx, n.ID, model.NotificationSent, n.Attempts+1, nil, "", res.ProviderResponse, &now)
		atomic.AddUint64(&e.deliveredCount, 1)
		logger.Info("Notification %d successfully delivered via %s to %s", n.ID, channel.Type, n.Recipient)
	} else {
		attempts := n.Attempts + 1
		maxAttempts := n.MaxAttempts
		if maxAttempts <= 0 {
			maxAttempts = channel.MaxRetries
		}
		if maxAttempts <= 0 {
			maxAttempts = e.cfg.DefaultMaxRetries
		}

		if res.Retryable && attempts < maxAttempts {
			// Exponential backoff: base * 2^(attempts-1)
			backoffSec := channel.RetryIntervalSec
			if backoffSec <= 0 {
				backoffSec = e.cfg.DefaultBackoffSec
			}
			delay := time.Duration(1<<(attempts-1)) * time.Duration(backoffSec) * time.Second
			nextRetry := now.Add(delay)

			_ = e.repo.UpdateDeliveryAttempt(ctx, n.ID, model.NotificationRetrying, attempts, &nextRetry, res.ErrorMsg, res.ProviderResponse, nil)
			atomic.AddUint64(&e.retriedCount, 1)
			logger.Warn("Notification %d delivery failed (attempt %d/%d): %s. Retrying in %v", n.ID, attempts, maxAttempts, res.ErrorMsg, delay)
		} else {
			// Terminal Failure
			_ = e.repo.UpdateDeliveryAttempt(ctx, n.ID, model.NotificationFailed, attempts, nil, res.ErrorMsg, res.ProviderResponse, nil)
			atomic.AddUint64(&e.failedCount, 1)
			logger.Error("Notification %d delivery failed permanently: %s", n.ID, res.ErrorMsg)
		}
	}
}

// isQuietHours checks whether current time in channel's timezone falls within configured quiet hours
func (e *NotificationEngine) isQuietHours(channel *model.NotificationChannel) bool {
	if channel.QuietHoursStart == "" || channel.QuietHoursEnd == "" {
		return false
	}

	loc, err := time.LoadLocation(channel.Timezone)
	if err != nil {
		loc = time.UTC
	}
	now := time.Now().In(loc)
	currentTimeStr := now.Format("15:04")

	start := channel.QuietHoursStart
	end := channel.QuietHoursEnd

	if start < end {
		return currentTimeStr >= start && currentTimeStr < end
	}
	// Wraps across midnight (e.g. 22:00 to 06:00)
	return currentTimeStr >= start || currentTimeStr < end
}

// TestSend validates provider connection and sends a sample message without creating an alarm
func (e *NotificationEngine) TestSend(ctx context.Context, channelID uint, recipient string) notification.DeliveryResult {
	ch, err := e.repo.GetChannelByID(ctx, channelID)
	if err != nil {
		return notification.DeliveryResult{
			Success:  false,
			ErrorMsg: "Channel not found",
		}
	}

	e.mu.RLock()
	provider, exists := e.providers[ch.Type]
	e.mu.RUnlock()

	if !exists {
		return notification.DeliveryResult{
			Success:  false,
			ErrorMsg: fmt.Sprintf("No provider registered for %s", ch.Type),
		}
	}

	if recipient == "" {
		recipient = ch.Name
	}

	msg := notification.NotificationMessage{
		ID:        0,
		AlarmID:   0,
		Severity:  model.SeverityInfo,
		AlarmCode: "TEST_NOTIFICATION",
		Device:    "EDGE-DIAGNOSTIC",
		Parameter: "SYSTEM_CHECK",
		Subject:   fmt.Sprintf("[TEST] Datalogger Diagnostic Alert via %s", ch.Name),
		Body:      fmt.Sprintf("This is an authorized test notification from Datalogger Edge System at %s", time.Now().UTC().Format(time.RFC3339)),
		Recipient: recipient,
	}

	return provider.Send(ctx, msg, ch.ConfigJSON)
}

// RetryNotification manually resets a notification to PENDING and triggers immediate queue evaluation
func (e *NotificationEngine) RetryNotification(ctx context.Context, id uint) (*model.Notification, error) {
	n, err := e.repo.Retry(ctx, id)
	if err != nil {
		return nil, err
	}
	select {
	case e.notifyCh <- struct{}{}:
	default:
	}
	return n, nil
}

// CancelNotification cancels a pending or retrying notification
func (e *NotificationEngine) CancelNotification(ctx context.Context, id uint) (*model.Notification, error) {
	return e.repo.Cancel(ctx, id)
}

// GetMetrics returns real-time engine dispatch statistics
func (e *NotificationEngine) GetMetrics() (dispatched, delivered, failed, retried, cancelled uint64) {
	return atomic.LoadUint64(&e.dispatchedCount),
		atomic.LoadUint64(&e.deliveredCount),
		atomic.LoadUint64(&e.failedCount),
		atomic.LoadUint64(&e.retriedCount),
		atomic.LoadUint64(&e.cancelledCount)
}

// Helper: Check severity hierarchy eligibility
func isSeverityEligible(minSeverity, alarmSeverity model.AlarmSeverity) bool {
	rank := func(s model.AlarmSeverity) int {
		switch s {
		case model.SeverityInfo:
			return 1
		case model.SeverityWarning:
			return 2
		case model.SeverityCritical:
			return 3
		default:
			return 1
		}
	}
	return rank(alarmSeverity) >= rank(minSeverity)
}
