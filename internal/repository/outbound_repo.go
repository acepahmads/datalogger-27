package repository

import (
	"context"
	"strings"
	"time"

	"datalogger/internal/model"

	"gorm.io/gorm"
)

// OutboundRepository defines database operations for destinations, payload configs, and outbound delivery queue
type OutboundRepository interface {
	// Destinations
	GetAllDestinations(ctx context.Context, enabledOnly bool) ([]model.OutputDestination, error)
	GetDestinationByID(ctx context.Context, id uint) (*model.OutputDestination, error)
	CreateDestination(ctx context.Context, dest *model.OutputDestination) error
	UpdateDestination(ctx context.Context, dest *model.OutputDestination) error
	DeleteDestination(ctx context.Context, id uint) error

	// Destination Payload Configs
	GetPayloadConfigByDestinationID(ctx context.Context, destinationID uint) (*model.DestinationPayloadConfig, error)
	UpsertPayloadConfig(ctx context.Context, cfg *model.DestinationPayloadConfig) error

	// Aggregation Results Query for Outbound Batching
	GetCustomerAggregationResults(ctx context.Context, deviceID uint, bucketEnd time.Time, intervalSeconds int) ([]model.AggregationResult, error)
	GetCustomerAggregationResultsForWindow(ctx context.Context, bucketEnd time.Time, intervalSeconds int, deviceIDs []uint) ([]model.AggregationResult, error)
	GetLatestAggregationResults(ctx context.Context, deviceID uint) ([]model.AggregationResult, error)
	GetLatestAggregationResultsMultiDevice(ctx context.Context, deviceIDs []uint) ([]model.AggregationResult, error)

	// Delivery Queue
	EnqueueDelivery(ctx context.Context, item *model.OutboundDeliveryQueue) error
	GetQueueItemByID(ctx context.Context, id uint64) (*model.OutboundDeliveryQueue, error)
	GetQueueItems(ctx context.Context, filter QueueFilter) ([]model.OutboundDeliveryQueue, int64, error)
	GetQueueSummary(ctx context.Context) (*model.DeliveryQueueSummary, error)
	FetchPendingQueueItems(ctx context.Context, limit int) ([]model.OutboundDeliveryQueue, error)
	ClaimQueueItemForDelivery(ctx context.Context, id uint64) (bool, error)
	UpdateDeliveryResult(ctx context.Context, item *model.OutboundDeliveryQueue) error
	CancelQueueItem(ctx context.Context, id uint64) error
	ResetQueueItemForRetry(ctx context.Context, id uint64) error

	// Delivery Logs
	CreateDeliveryLog(ctx context.Context, log *model.DeliveryLog) error
}

// QueueFilter provides filtering and pagination options for queue listing
type QueueFilter struct {
	DestinationID   *uint
	DeviceID        *uint
	Status          string
	Search          string
	StartDate       *time.Time
	EndDate         *time.Time
	IntervalSeconds *int
	Page            int
	PageSize        int
}

type outboundRepository struct {
	db *gorm.DB
}

// NewOutboundRepository constructs a new GORM-backed outbound repository
func NewOutboundRepository(db *gorm.DB) OutboundRepository {
	return &outboundRepository{db: db}
}

// GetAllDestinations retrieves all outbound destinations with their payload configs and device profiles
func (r *outboundRepository) GetAllDestinations(ctx context.Context, enabledOnly bool) ([]model.OutputDestination, error) {
	var destinations []model.OutputDestination
	query := r.db.WithContext(ctx).Preload("PayloadConfig").Preload("Device")
	if enabledOnly {
		query = query.Where("enabled = ?", true)
	}
	err := query.Order("id ASC").Find(&destinations).Error
	return destinations, err
}

// GetDestinationByID retrieves a single destination by its primary key
func (r *outboundRepository) GetDestinationByID(ctx context.Context, id uint) (*model.OutputDestination, error) {
	var dest model.OutputDestination
	err := r.db.WithContext(ctx).
		Preload("PayloadConfig").
		Preload("Device").
		Where("id = ?", id).
		First(&dest).Error
	if err != nil {
		return nil, err
	}
	return &dest, nil
}

// CreateDestination persists a new outbound destination
func (r *outboundRepository) CreateDestination(ctx context.Context, dest *model.OutputDestination) error {
	return r.db.WithContext(ctx).Create(dest).Error
}

// UpdateDestination updates mutable fields on an existing outbound destination
func (r *outboundRepository) UpdateDestination(ctx context.Context, dest *model.OutputDestination) error {
	return r.db.WithContext(ctx).Save(dest).Error
}

// DeleteDestination removes a destination and its associated payload config, delivery logs, and delivery queue items
func (r *outboundRepository) DeleteDestination(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. Delete associated payload config
		if err := tx.Where("destination_id = ?", id).Delete(&model.DestinationPayloadConfig{}).Error; err != nil {
			return err
		}
		// 2. Delete associated delivery logs
		if err := tx.Where("destination_id = ?", id).Delete(&model.DeliveryLog{}).Error; err != nil {
			return err
		}
		// 3. Delete associated delivery queue items
		if err := tx.Where("destination_id = ?", id).Delete(&model.OutboundDeliveryQueue{}).Error; err != nil {
			return err
		}
		// 4. Finally delete the destination itself
		return tx.Where("id = ?", id).Delete(&model.OutputDestination{}).Error
	})
}

// GetPayloadConfigByDestinationID fetches the payload format rules for a specific destination
func (r *outboundRepository) GetPayloadConfigByDestinationID(ctx context.Context, destinationID uint) (*model.DestinationPayloadConfig, error) {
	var cfg model.DestinationPayloadConfig
	err := r.db.WithContext(ctx).Where("destination_id = ?", destinationID).First(&cfg).Error
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

// UpsertPayloadConfig creates or updates destination payload config with version increment
func (r *outboundRepository) UpsertPayloadConfig(ctx context.Context, cfg *model.DestinationPayloadConfig) error {
	var existing model.DestinationPayloadConfig
	err := r.db.WithContext(ctx).Where("destination_id = ?", cfg.DestinationID).First(&existing).Error
	if err == nil {
		cfg.ID = existing.ID
		cfg.Version = existing.Version + 1
		return r.db.WithContext(ctx).Save(cfg).Error
	}
	if cfg.Version <= 0 {
		cfg.Version = 1
	}
	return r.db.WithContext(ctx).Create(cfg).Error
}

// GetCustomerAggregationResults queries strictly persisted CUSTOMER_PROCESSED aggregation records for a single device
func (r *outboundRepository) GetCustomerAggregationResults(ctx context.Context, deviceID uint, bucketEnd time.Time, intervalSeconds int) ([]model.AggregationResult, error) {
	var results []model.AggregationResult
	query := r.db.WithContext(ctx).
		Preload("Parameter").
		Preload("Device").
		Where("device_id = ?", deviceID).
		Where("source_type = ?", model.SourceCustomerProcessed).
		Where("period_end = ?", bucketEnd)

	if intervalSeconds > 0 {
		query = query.Where("interval_seconds = ?", intervalSeconds)
	}

	err := query.Order("parameter_id ASC").Find(&results).Error
	return results, err
}

// GetCustomerAggregationResultsForWindow queries strictly persisted CUSTOMER_PROCESSED aggregation records across all relevant devices for a window
func (r *outboundRepository) GetCustomerAggregationResultsForWindow(ctx context.Context, bucketEnd time.Time, intervalSeconds int, deviceIDs []uint) ([]model.AggregationResult, error) {
	var results []model.AggregationResult
	query := r.db.WithContext(ctx).
		Preload("Parameter").
		Preload("Device").
		Where("source_type = ?", model.SourceCustomerProcessed).
		Where("period_end = ?", bucketEnd)

	if intervalSeconds > 0 {
		query = query.Where("interval_seconds = ?", intervalSeconds)
	}
	if len(deviceIDs) > 0 {
		query = query.Where("device_id IN (?)", deviceIDs)
	}

	err := query.Order("device_id ASC, parameter_id ASC").Find(&results).Error
	return results, err
}

// GetLatestAggregationResults fetches the most recent customer aggregation values for live payload preview
func (r *outboundRepository) GetLatestAggregationResults(ctx context.Context, deviceID uint) ([]model.AggregationResult, error) {
	var results []model.AggregationResult
	// Find latest period_end for this device
	var latest model.AggregationResult
	err := r.db.WithContext(ctx).
		Where("device_id = ? AND source_type = ?", deviceID, model.SourceCustomerProcessed).
		Order("period_end DESC").
		First(&latest).Error
	if err != nil {
		return nil, err
	}

	err = r.db.WithContext(ctx).
		Preload("Parameter").
		Preload("Device").
		Where("device_id = ? AND source_type = ? AND period_end = ?", deviceID, model.SourceCustomerProcessed, latest.PeriodEnd).
		Order("parameter_id ASC").
		Find(&results).Error
	return results, err
}

// GetLatestAggregationResultsMultiDevice fetches latest customer aggregation values across all relevant devices for live payload preview
func (r *outboundRepository) GetLatestAggregationResultsMultiDevice(ctx context.Context, deviceIDs []uint) ([]model.AggregationResult, error) {
	var results []model.AggregationResult
	// Find latest period_end overall
	var latest model.AggregationResult
	q := r.db.WithContext(ctx).Where("source_type = ?", model.SourceCustomerProcessed)
	if len(deviceIDs) > 0 {
		q = q.Where("device_id IN (?)", deviceIDs)
	}
	err := q.Order("period_end DESC").First(&latest).Error
	if err != nil {
		return nil, err
	}

	query := r.db.WithContext(ctx).
		Preload("Parameter").
		Preload("Device").
		Where("source_type = ?", model.SourceCustomerProcessed).
		Where("period_end = ?", latest.PeriodEnd)

	if len(deviceIDs) > 0 {
		query = query.Where("device_id IN (?)", deviceIDs)
	}

	err = query.Order("device_id ASC, parameter_id ASC").Find(&results).Error
	return results, err
}

// EnqueueDelivery persists a new delivery queue entry with idempotency check and pending batch merge
func (r *outboundRepository) EnqueueDelivery(ctx context.Context, item *model.OutboundDeliveryQueue) error {
	var existing model.OutboundDeliveryQueue
	err := r.db.WithContext(ctx).Where("idempotency_key = ?", item.IdempotencyKey).First(&existing).Error
	if err == nil {
		// Existing item with this idempotency key
		if item.IsReplay {
			return r.db.WithContext(ctx).Create(item).Error
		}
		// If still pending, allow updating with additional unified devices / parameters
		if existing.Status == model.DeliveryPending {
			existing.DeviceCount = item.DeviceCount
			existing.DevicesSummary = item.DevicesSummary
			existing.ParameterCount = item.ParameterCount
			existing.SourceAggregationIDs = item.SourceAggregationIDs
			existing.SourceValuesJSON = item.SourceValuesJSON
			existing.PayloadSnapshot = item.PayloadSnapshot
			existing.PayloadVersion = item.PayloadVersion
			existing.UpdatedAt = time.Now()
			return r.db.WithContext(ctx).Save(&existing).Error
		}
		// If already claimed (SENDING) or DELIVERED, do not duplicate
		return nil
	}
	return r.db.WithContext(ctx).Create(item).Error
}

// GetQueueItemByID retrieves a delivery queue item by its primary key
func (r *outboundRepository) GetQueueItemByID(ctx context.Context, id uint64) (*model.OutboundDeliveryQueue, error) {
	var item model.OutboundDeliveryQueue
	err := r.db.WithContext(ctx).
		Preload("Destination").
		Preload("Device").
		Where("id = ?", id).
		First(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

// GetQueueItems queries filtered, paginated queue entries with total count
func (r *outboundRepository) GetQueueItems(ctx context.Context, filter QueueFilter) ([]model.OutboundDeliveryQueue, int64, error) {
	var items []model.OutboundDeliveryQueue
	var total int64

	query := r.db.WithContext(ctx).Model(&model.OutboundDeliveryQueue{})

	if filter.DestinationID != nil && *filter.DestinationID > 0 {
		query = query.Where("destination_id = ?", *filter.DestinationID)
	}
	if filter.DeviceID != nil && *filter.DeviceID > 0 {
		query = query.Where("device_id = ?", *filter.DeviceID)
	}
	if filter.Status != "" && filter.Status != "ALL" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.IntervalSeconds != nil && *filter.IntervalSeconds > 0 {
		query = query.Where("interval_seconds = ?", *filter.IntervalSeconds)
	}
	if filter.StartDate != nil {
		query = query.Where("created_at >= ?", *filter.StartDate)
	}
	if filter.EndDate != nil {
		query = query.Where("created_at <= ?", *filter.EndDate)
	}
	if strings.TrimSpace(filter.Search) != "" {
		s := "%" + strings.TrimSpace(filter.Search) + "%"
		query = query.Where("batch_id LIKE ? OR idempotency_key LIKE ?", s, s)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	page := filter.Page
	if page <= 0 {
		page = 1
	}
	pageSize := filter.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	offset := (page - 1) * pageSize

	err := query.
		Preload("Destination").
		Preload("Device").
		Order("id DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&items).Error

	return items, total, err
}

// GetQueueSummary returns fast aggregate counts across delivery statuses
func (r *outboundRepository) GetQueueSummary(ctx context.Context) (*model.DeliveryQueueSummary, error) {
	summary := &model.DeliveryQueueSummary{}

	type statusCount struct {
		Status string
		Count  int64
	}
	var counts []statusCount

	err := r.db.WithContext(ctx).
		Model(&model.OutboundDeliveryQueue{}).
		Select("status, COUNT(*) as count").
		Group("status").
		Scan(&counts).Error
	if err != nil {
		return nil, err
	}

	for _, sc := range counts {
		summary.Total += sc.Count
		switch model.OutboundDeliveryStatus(sc.Status) {
		case model.DeliveryPending:
			summary.Pending = sc.Count
		case model.DeliverySending:
			summary.Sending = sc.Count
		case model.DeliveryDelivered:
			summary.Delivered = sc.Count
		case model.DeliveryRetrying:
			summary.Retrying = sc.Count
		case model.DeliveryFailed:
			summary.Failed = sc.Count
		case model.DeliveryAcknowledged:
			summary.Acknowledged = sc.Count
		}
	}

	return summary, nil
}

// FetchPendingQueueItems fetches pending or retryable queue items that are ready for transmission
func (r *outboundRepository) FetchPendingQueueItems(ctx context.Context, limit int) ([]model.OutboundDeliveryQueue, error) {
	var items []model.OutboundDeliveryQueue
	now := time.Now()

	err := r.db.WithContext(ctx).
		Preload("Destination").
		Preload("Destination.PayloadConfig").
		Preload("Device").
		Where("status IN ?", []string{string(model.DeliveryPending), string(model.DeliveryRetrying)}).
		Where("next_retry_at IS NULL OR next_retry_at <= ?", now).
		Order("id ASC").
		Limit(limit).
		Find(&items).Error

	return items, err
}

// ClaimQueueItemForDelivery atomically claims a queue item from PENDING/RETRYING to SENDING
func (r *outboundRepository) ClaimQueueItemForDelivery(ctx context.Context, id uint64) (bool, error) {
	now := time.Now()
	res := r.db.WithContext(ctx).
		Model(&model.OutboundDeliveryQueue{}).
		Where("id = ? AND status IN ?", id, []string{string(model.DeliveryPending), string(model.DeliveryRetrying)}).
		Updates(map[string]interface{}{
			"status":          model.DeliverySending,
			"last_attempt_at": &now,
			"attempts":        gorm.Expr("attempts + 1"),
		})

	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected > 0, nil
}

// UpdateDeliveryResult updates the outcome of an attempted transmission
func (r *outboundRepository) UpdateDeliveryResult(ctx context.Context, item *model.OutboundDeliveryQueue) error {
	return r.db.WithContext(ctx).Model(&model.OutboundDeliveryQueue{}).
		Where("id = ?", item.ID).
		Updates(map[string]interface{}{
			"status":          item.Status,
			"next_retry_at":   item.NextRetryAt,
			"last_attempt_at": item.LastAttemptAt,
			"delivered_at":    item.DeliveredAt,
			"last_http_code":  item.LastHTTPCode,
			"last_response":   item.LastResponse,
			"last_error":      item.LastError,
		}).Error
}

// CancelQueueItem marks a pending or retrying queue item as CANCELLED
func (r *outboundRepository) CancelQueueItem(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Model(&model.OutboundDeliveryQueue{}).
		Where("id = ? AND status IN ?", id, []string{string(model.DeliveryPending), string(model.DeliveryRetrying)}).
		Update("status", model.DeliveryCancelled).Error
}

// ResetQueueItemForRetry resets an item to PENDING for immediate retry
func (r *outboundRepository) ResetQueueItemForRetry(ctx context.Context, id uint64) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&model.OutboundDeliveryQueue{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":        model.DeliveryPending,
			"next_retry_at": &now,
		}).Error
}

// CreateDeliveryLog persists a transmission log entry
func (r *outboundRepository) CreateDeliveryLog(ctx context.Context, log *model.DeliveryLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}
