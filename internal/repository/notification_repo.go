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
	ErrChannelNotFound      = errors.New("notification channel not found")
	ErrNotificationNotFound = errors.New("notification not found")
	ErrDuplicateNotification = errors.New("notification already enqueued for this event")
)

// NotificationFilterParams defines filtering, search, and pagination criteria for notifications
type NotificationFilterParams struct {
	Status    string
	Channel   string
	AlarmID   uint
	Search    string
	StartTime *time.Time
	EndTime   *time.Time
	Page      int
	PageSize  int
}

// NotificationRepository provides database storage operations for notification channels and delivery queue
type NotificationRepository struct {
	db *gorm.DB
}

// NewNotificationRepository creates a new instance of NotificationRepository
func NewNotificationRepository(db *gorm.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

// ListChannels returns all registered notification channels
func (r *NotificationRepository) ListChannels(ctx context.Context) ([]model.NotificationChannel, error) {
	var channels []model.NotificationChannel
	err := r.db.WithContext(ctx).Order("id ASC").Find(&channels).Error
	return channels, err
}

// GetChannelByID retrieves a channel by primary key
func (r *NotificationRepository) GetChannelByID(ctx context.Context, id uint) (*model.NotificationChannel, error) {
	var ch model.NotificationChannel
	err := r.db.WithContext(ctx).First(&ch, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrChannelNotFound
		}
		return nil, err
	}
	return &ch, nil
}

// GetChannelByType retrieves a channel by its type enum
func (r *NotificationRepository) GetChannelByType(ctx context.Context, chType model.NotificationChannelType) (*model.NotificationChannel, error) {
	var ch model.NotificationChannel
	err := r.db.WithContext(ctx).Where("type = ?", chType).First(&ch).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrChannelNotFound
		}
		return nil, err
	}
	return &ch, nil
}

// CreateChannel persists a new channel configuration
func (r *NotificationRepository) CreateChannel(ctx context.Context, ch *model.NotificationChannel) error {
	if ch == nil {
		return errors.New("channel cannot be nil")
	}
	if ch.RateLimitPerMin <= 0 {
		ch.RateLimitPerMin = 10
	}
	if ch.MaxRetries <= 0 {
		ch.MaxRetries = 3
	}
	if ch.RetryIntervalSec <= 0 {
		ch.RetryIntervalSec = 30
	}
	if ch.Timezone == "" {
		ch.Timezone = "UTC"
	}
	return r.db.WithContext(ctx).Create(ch).Error
}

// UpdateChannel updates configuration and settings of an existing channel
func (r *NotificationRepository) UpdateChannel(ctx context.Context, ch *model.NotificationChannel) error {
	if ch == nil {
		return errors.New("channel cannot be nil")
	}
	return r.db.WithContext(ctx).Save(ch).Error
}

// DeleteChannel removes a channel by ID
func (r *NotificationRepository) DeleteChannel(ctx context.Context, id uint) error {
	res := r.db.WithContext(ctx).Delete(&model.NotificationChannel{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrChannelNotFound
	}
	return nil
}

// Enqueue adds a notification record to the delivery queue with idempotency protection
func (r *NotificationRepository) Enqueue(ctx context.Context, n *model.Notification) error {
	if n == nil {
		return errors.New("notification cannot be nil")
	}
	if n.Status == "" {
		n.Status = model.NotificationPending
	}
	if n.MaxAttempts <= 0 {
		n.MaxAttempts = 3
	}
	now := time.Now().UTC()
	if n.CreatedAt.IsZero() {
		n.CreatedAt = now
	}
	n.UpdatedAt = now

	// Check idempotency if key provided
	if n.IdempotencyKey != "" {
		var existing model.Notification
		err := r.db.WithContext(ctx).Where("idempotency_key = ?", n.IdempotencyKey).First(&existing).Error
		if err == nil {
			return ErrDuplicateNotification
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
	}

	return r.db.WithContext(ctx).Create(n).Error
}

// GetByID retrieves a single notification by primary key with relations preloaded
func (r *NotificationRepository) GetByID(ctx context.Context, id uint) (*model.Notification, error) {
	var n model.Notification
	err := r.db.WithContext(ctx).
		Preload("Alarm").
		Preload("Alarm.Device").
		Preload("Alarm.Parameter").
		First(&n, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotificationNotFound
		}
		return nil, err
	}
	return &n, nil
}

// FindByIdempotencyKey retrieves an existing notification by its unique deduplication key
func (r *NotificationRepository) FindByIdempotencyKey(ctx context.Context, key string) (*model.Notification, error) {
	var n model.Notification
	err := r.db.WithContext(ctx).Where("idempotency_key = ?", key).First(&n).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &n, nil
}

// List returns notifications based on filtering, search, and pagination
func (r *NotificationRepository) List(ctx context.Context, params NotificationFilterParams) ([]model.Notification, int64, error) {
	var notifications []model.Notification
	var total int64

	query := r.db.WithContext(ctx).Model(&model.Notification{}).
		Preload("Alarm").
		Preload("Alarm.Device").
		Preload("Alarm.Parameter")

	if params.Status != "" {
		query = query.Where("status = ?", strings.ToUpper(strings.TrimSpace(params.Status)))
	}
	if params.Channel != "" {
		query = query.Where("channel = ?", strings.ToUpper(strings.TrimSpace(params.Channel)))
	}
	if params.AlarmID > 0 {
		query = query.Where("alarm_id = ?", params.AlarmID)
	}
	if params.Search != "" {
		s := "%" + strings.TrimSpace(params.Search) + "%"
		query = query.Where("recipient LIKE ? OR subject LIKE ? OR message LIKE ? OR error_msg LIKE ?", s, s, s, s)
	}
	if params.StartTime != nil && !params.StartTime.IsZero() {
		query = query.Where("created_at >= ?", params.StartTime)
	}
	if params.EndTime != nil && !params.EndTime.IsZero() {
		query = query.Where("created_at <= ?", params.EndTime)
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

	err := query.Order("created_at DESC, id DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&notifications).Error
	if err != nil {
		return nil, 0, err
	}

	return notifications, total, nil
}

// FindPending returns pending or retrying notifications that are ready to be dispatched
func (r *NotificationRepository) FindPending(ctx context.Context, limit int) ([]model.Notification, error) {
	if limit <= 0 {
		limit = 25
	}
	now := time.Now().UTC()

	var notifications []model.Notification
	err := r.db.WithContext(ctx).
		Preload("Alarm").
		Preload("Alarm.Device").
		Preload("Alarm.Parameter").
		Where("status IN ? AND (next_retry_at IS NULL OR next_retry_at <= ?)",
			[]model.NotificationStatus{model.NotificationPending, model.NotificationRetrying}, now).
		Order("id ASC").
		Limit(limit).
		Find(&notifications).Error

	return notifications, err
}

// UpdateDeliveryAttempt updates delivery status, attempts counter, provider response, and retry schedule
func (r *NotificationRepository) UpdateDeliveryAttempt(
	ctx context.Context,
	id uint,
	status model.NotificationStatus,
	attempts int,
	nextRetryAt *time.Time,
	errorMsg, providerResp string,
	sentAt *time.Time,
) error {
	updates := map[string]interface{}{
		"status":            status,
		"attempts":          attempts,
		"next_retry_at":     nextRetryAt,
		"error_msg":         errorMsg,
		"provider_response": providerResp,
		"updated_at":        time.Now().UTC(),
	}
	if sentAt != nil {
		updates["sent_at"] = sentAt
	}

	return r.db.WithContext(ctx).Model(&model.Notification{}).
		Where("id = ?", id).
		Updates(updates).Error
}

// Cancel sets a pending or retrying notification to CANCELLED state
func (r *NotificationRepository) Cancel(ctx context.Context, id uint) (*model.Notification, error) {
	now := time.Now().UTC()
	res := r.db.WithContext(ctx).Model(&model.Notification{}).
		Where("id = ? AND status IN ?", id, []model.NotificationStatus{model.NotificationPending, model.NotificationRetrying, model.NotificationSending}).
		Updates(map[string]interface{}{
			"status":     model.NotificationCancelled,
			"error_msg":  "Cancelled by operator",
			"updated_at": now,
		})

	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected == 0 {
		return r.GetByID(ctx, id)
	}

	return r.GetByID(ctx, id)
}

// Retry resets a failed or cancelled notification back to PENDING state for immediate redelivery
func (r *NotificationRepository) Retry(ctx context.Context, id uint) (*model.Notification, error) {
	now := time.Now().UTC()
	res := r.db.WithContext(ctx).Model(&model.Notification{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":        model.NotificationPending,
			"attempts":      0,
			"next_retry_at": nil,
			"error_msg":     "",
			"updated_at":    now,
		})

	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected == 0 {
		return nil, ErrNotificationNotFound
	}

	return r.GetByID(ctx, id)
}

// CancelPendingByChannelID cancels any pending or retrying notifications belonging to a disabled channel
func (r *NotificationRepository) CancelPendingByChannelID(ctx context.Context, channelID uint) (int64, error) {
	if channelID == 0 {
		return 0, nil
	}
	now := time.Now().UTC()
	res := r.db.WithContext(ctx).Model(&model.Notification{}).
		Where("channel_id = ? AND status IN ?", channelID, []model.NotificationStatus{
			model.NotificationPending,
			model.NotificationRetrying,
		}).
		Updates(map[string]interface{}{
			"status":        model.NotificationCancelled,
			"next_retry_at": nil,
			"error_msg":     "Channel disabled by administrator",
			"updated_at":    now,
		})
	return res.RowsAffected, res.Error
}

// CountRecentByChannel returns the number of notifications sent on a channel since a given timestamp (for rate limiting)
func (r *NotificationRepository) CountRecentByChannel(ctx context.Context, channel string, since time.Time) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Notification{}).
		Where("channel = ? AND created_at >= ?", channel, since).
		Count(&count).Error
	return count, err
}
