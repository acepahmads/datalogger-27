package repository

import (
	"context"
	"strings"
	"time"

	"datalogger/internal/model"
	"datalogger/pkg/tenantctx"

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
	RecordDestinationLastDelivery(ctx context.Context, destinationID uint, deliveredAt time.Time) error
	SyncDestinationStatsFromQueue(ctx context.Context, destinationID *uint) error

	// Destination Payload Configs
	GetPayloadConfigByDestinationID(ctx context.Context, destinationID uint) (*model.DestinationPayloadConfig, error)
	UpsertPayloadConfig(ctx context.Context, cfg *model.DestinationPayloadConfig) error

	// Aggregation Results Query for Outbound Batching
	GetCustomerAggregationResults(ctx context.Context, deviceID uint, bucketEnd time.Time, intervalSeconds int) ([]model.AggregationResult, error)
	GetCustomerAggregationResultsForWindow(ctx context.Context, bucketEnd time.Time, intervalSeconds int, deviceIDs []uint) ([]model.AggregationResult, error)
	GetCustomerAggregationResultsForRange(ctx context.Context, windowStart time.Time, windowEnd time.Time, intervalSeconds int, deviceIDs []uint) ([]model.AggregationResult, error)
	GetLatestAggregationResults(ctx context.Context, deviceID uint) ([]model.AggregationResult, error)
	GetLatestAggregationResultsMultiDevice(ctx context.Context, deviceIDs []uint) ([]model.AggregationResult, error)
	GetRecentAggregationResultsMultiDevice(ctx context.Context, intervalSeconds int, windowDuration time.Duration, deviceIDs []uint) ([]model.AggregationResult, error)

	// Delivery Queue
	EnqueueDelivery(ctx context.Context, item *model.OutboundDeliveryQueue) error
	EnqueueOrRefreshDelivery(ctx context.Context, item *model.OutboundDeliveryQueue) (string, model.OutboundDeliveryStatus, error)
	GetQueueItemByID(ctx context.Context, id uint64) (*model.OutboundDeliveryQueue, error)
	GetQueueItems(ctx context.Context, filter QueueFilter) ([]model.OutboundDeliveryQueue, int64, error)
	GetQueueSummary(ctx context.Context) (*model.DeliveryQueueSummary, error)
	FetchPendingQueueItems(ctx context.Context, limit int) ([]model.OutboundDeliveryQueue, error)
	GetUndeliveredQueueItemsByDestination(ctx context.Context, destinationID uint) ([]model.OutboundDeliveryQueue, error)
	GetUndeliveredQueueItemsOverlappingWindow(ctx context.Context, startTime, endTime time.Time) ([]model.OutboundDeliveryQueue, error)
	UpdateQueueItemPayloadAndRequeue(ctx context.Context, item *model.OutboundDeliveryQueue, requeue bool) error
	ClaimQueueItemForDelivery(ctx context.Context, id uint64) (bool, error)
	UpdateDeliveryResult(ctx context.Context, item *model.OutboundDeliveryQueue) error
	CancelQueueItem(ctx context.Context, id uint64) error
	ResetQueueItemForRetry(ctx context.Context, id uint64) error
	DeleteQueueItem(ctx context.Context, id uint64) error
	PurgeQueueItems(ctx context.Context, filter QueuePurgeFilter) (int64, int64, error)
	CountQueuePurgeItems(ctx context.Context, filter QueuePurgeFilter) (int64, int64, error)
	PurgeQueueBatch(ctx context.Context, filter QueuePurgeFilter, batchSize int) (int64, int64, bool, error)

	// Delivery Logs
	CreateDeliveryLog(ctx context.Context, log *model.DeliveryLog) error
}

// QueuePurgeFilter defines criteria for manual bulk removal of queue items and delivery logs
type QueuePurgeFilter struct {
	TenantID      uint       `json:"tenant_id,omitempty"`
	SiteID        uint       `json:"site_id,omitempty"`
	DestinationID *uint      `json:"destination_id,omitempty"`
	Status        string     `json:"status,omitempty"`
	OlderThan     *time.Time `json:"older_than,omitempty"`
	PurgeLogs     bool       `json:"purge_logs"`
}

// QueueFilter provides filtering and pagination options for queue listing
type QueueFilter struct {
	TenantID        uint
	SiteID          uint
	AllowedSiteIDs  []uint
	DestinationID   *uint
	DeviceID        *uint
	Status          string
	UndeliveredOnly bool
	Search          string
	StartDate       *time.Time
	EndDate         *time.Time
	WindowStart     *time.Time
	WindowEnd       *time.Time
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
	scope := tenantctx.FromContext(ctx)
	var destinations []model.OutputDestination
	query := r.db.WithContext(ctx).Preload("PayloadConfig").Preload("Device").Preload("Site")
	if scope.TenantID > 0 {
		query = query.Where("tenant_id = ?", scope.TenantID)
	}
	if scope.SiteID > 0 {
		query = query.Where("(site_id IS NULL OR site_id = ?)", scope.SiteID)
	} else if len(scope.AllowedSiteIDs) > 0 {
		query = query.Where("(site_id IS NULL OR site_id IN ?)", scope.AllowedSiteIDs)
	}
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
		Preload("Site").
		Where("id = ?", id).
		First(&dest).Error
	if err != nil {
		return nil, err
	}
	return &dest, nil
}

// CreateDestination persists a new outbound destination
func (r *outboundRepository) CreateDestination(ctx context.Context, dest *model.OutputDestination) error {
	if dest.TenantID == 0 {
		dest.TenantID = tenantctx.FromContext(ctx).WriteTenantID(0)
	}
	return r.db.WithContext(ctx).Omit("PayloadConfig", "Device", "Site").Create(dest).Error
}

// UpdateDestination updates mutable fields on an existing outbound destination without touching associations
func (r *outboundRepository) UpdateDestination(ctx context.Context, dest *model.OutputDestination) error {
	return r.db.WithContext(ctx).Omit("PayloadConfig", "Device", "Site").Save(dest).Error
}

// DeleteDestination removes a destination setup and its associated payload config,
// leaving outbound delivery queues and delivery logs intact for manual admin inspection/cleanup.
func (r *outboundRepository) DeleteDestination(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. Delete associated payload config
		if err := tx.Where("destination_id = ?", id).Delete(&model.DestinationPayloadConfig{}).Error; err != nil {
			return err
		}
		// 2. Soft-delete the destination setup (antrean & log riwayat tetap tersimpan utuh di database)
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
	desiredIncludeMetadata := cfg.IncludeMetadata
	if err := r.db.WithContext(ctx).Create(cfg).Error; err != nil {
		return err
	}
	// Preserve explicit IncludeMetadata=false on initial INSERT (GORM skips zero-valued bools when column has default:true)
	if !desiredIncludeMetadata && cfg.IncludeMetadata {
		if err := r.db.WithContext(ctx).Model(cfg).Update("include_metadata", false).Error; err != nil {
			return err
		}
		cfg.IncludeMetadata = false
	}
	return nil
}

// applyActiveDeviceParamScope restricts aggregation_results queries strictly to active, non-deleted devices and active, non-deleted parameters,
// while honoring any multi-tenant/site scope attached to ctx.
func applyActiveDeviceParamScope(ctx context.Context, db *gorm.DB) *gorm.DB {
	scope := tenantctx.FromContext(ctx)
	q := db.WithContext(ctx).
		Joins("INNER JOIN devices ON devices.id = aggregation_results.device_id AND devices.deleted_at IS NULL AND devices.enabled = ? AND UPPER(COALESCE(devices.status, 'ACTIVE')) NOT IN ('INACTIVE', 'DISABLED')", true).
		Joins("INNER JOIN parameters ON parameters.id = aggregation_results.parameter_id AND parameters.deleted_at IS NULL AND parameters.enabled = ?", true)
	if scope.TenantID > 0 {
		q = q.Where("devices.tenant_id = ?", scope.TenantID)
	}
	if scope.SiteID > 0 {
		q = q.Where("devices.site_id = ?", scope.SiteID)
	} else if len(scope.AllowedSiteIDs) > 0 {
		q = q.Where("devices.site_id IN ?", scope.AllowedSiteIDs)
	}
	return q
}

// GetCustomerAggregationResults queries strictly persisted CUSTOMER_PROCESSED aggregation records for a single device
func (r *outboundRepository) GetCustomerAggregationResults(ctx context.Context, deviceID uint, bucketEnd time.Time, intervalSeconds int) ([]model.AggregationResult, error) {
	var results []model.AggregationResult
	query := applyActiveDeviceParamScope(ctx, r.db).
		Preload("Parameter").
		Preload("Device").
		Where("aggregation_results.device_id = ?", deviceID).
		Where("aggregation_results.source_type = ?", model.SourceCustomerProcessed).
		Where("aggregation_results.period_end = ?", bucketEnd)

	if intervalSeconds > 0 {
		query = query.Where("aggregation_results.interval_seconds = ?", intervalSeconds)
	}

	err := query.Order("aggregation_results.parameter_id ASC").Find(&results).Error
	return results, err
}

// GetCustomerAggregationResultsForWindow queries strictly persisted CUSTOMER_PROCESSED aggregation records across all relevant devices for a window
func (r *outboundRepository) GetCustomerAggregationResultsForWindow(ctx context.Context, bucketEnd time.Time, intervalSeconds int, deviceIDs []uint) ([]model.AggregationResult, error) {
	var results []model.AggregationResult
	query := applyActiveDeviceParamScope(ctx, r.db).
		Preload("Parameter").
		Preload("Device").
		Where("aggregation_results.source_type = ?", model.SourceCustomerProcessed).
		Where("aggregation_results.period_end = ?", bucketEnd)

	if intervalSeconds > 0 {
		query = query.Where("aggregation_results.interval_seconds = ?", intervalSeconds)
	}
	if len(deviceIDs) > 0 {
		query = query.Where("aggregation_results.device_id IN (?)", deviceIDs)
	}

	err := query.Order("aggregation_results.device_id ASC, aggregation_results.parameter_id ASC").Find(&results).Error
	return results, err
}

// GetLatestAggregationResults fetches the most recent customer aggregation values for live payload preview
func (r *outboundRepository) GetLatestAggregationResults(ctx context.Context, deviceID uint) ([]model.AggregationResult, error) {
	var results []model.AggregationResult
	// Find latest period_end for this device
	var latest model.AggregationResult
	err := applyActiveDeviceParamScope(ctx, r.db).
		Where("aggregation_results.device_id = ? AND aggregation_results.source_type = ?", deviceID, model.SourceCustomerProcessed).
		Order("aggregation_results.period_end DESC").
		First(&latest).Error
	if err != nil {
		return nil, err
	}

	err = applyActiveDeviceParamScope(ctx, r.db).
		Preload("Parameter").
		Preload("Device").
		Where("aggregation_results.device_id = ? AND aggregation_results.source_type = ? AND aggregation_results.period_end = ?", deviceID, model.SourceCustomerProcessed, latest.PeriodEnd).
		Order("aggregation_results.parameter_id ASC").
		Find(&results).Error
	return results, err
}

// GetLatestAggregationResultsMultiDevice fetches latest customer aggregation values across all relevant devices for live payload preview
func (r *outboundRepository) GetLatestAggregationResultsMultiDevice(ctx context.Context, deviceIDs []uint) ([]model.AggregationResult, error) {
	var results []model.AggregationResult
	// Find latest period_end overall across active devices/parameters
	var latest model.AggregationResult
	q := applyActiveDeviceParamScope(ctx, r.db).
		Where("aggregation_results.source_type = ?", model.SourceCustomerProcessed)
	if len(deviceIDs) > 0 {
		q = q.Where("aggregation_results.device_id IN (?)", deviceIDs)
	}
	err := q.Order("aggregation_results.period_end DESC").First(&latest).Error
	if err != nil {
		return nil, err
	}

	query := applyActiveDeviceParamScope(ctx, r.db).
		Preload("Parameter").
		Preload("Device").
		Where("aggregation_results.source_type = ?", model.SourceCustomerProcessed).
		Where("aggregation_results.period_end = ?", latest.PeriodEnd)

	if len(deviceIDs) > 0 {
		query = query.Where("aggregation_results.device_id IN (?)", deviceIDs)
	}

	err = query.Order("aggregation_results.device_id ASC, aggregation_results.parameter_id ASC").Find(&results).Error
	return results, err
}

// GetCustomerAggregationResultsForRange queries strictly persisted CUSTOMER_PROCESSED aggregation records across a multi-bucket time range [windowStart, windowEnd]
// ordered chronologically by period_start ASC (e.g. 30 buckets of 2-minute SPARING data within 1 hour).
func (r *outboundRepository) GetCustomerAggregationResultsForRange(ctx context.Context, windowStart time.Time, windowEnd time.Time, intervalSeconds int, deviceIDs []uint) ([]model.AggregationResult, error) {
	var results []model.AggregationResult
	query := applyActiveDeviceParamScope(ctx, r.db).
		Preload("Parameter").
		Preload("Device").
		Where("aggregation_results.source_type = ?", model.SourceCustomerProcessed).
		Where("aggregation_results.period_start >= ? AND aggregation_results.period_end <= ?", windowStart, windowEnd)

	if intervalSeconds > 0 {
		query = query.Where("aggregation_results.interval_seconds = ?", intervalSeconds)
	}
	if len(deviceIDs) > 0 {
		query = query.Where("aggregation_results.device_id IN (?)", deviceIDs)
	}

	err := query.Order("aggregation_results.period_start ASC, aggregation_results.device_id ASC, aggregation_results.parameter_id ASC").Find(&results).Error
	return results, err
}

// GetRecentAggregationResultsMultiDevice fetches chronological multi-bucket customer aggregation values ending at the latest completed bucket
// for live preview and test delivery of bundled schemas (e.g. KLHK SPARING 1-hour / 30x 2-minute data array).
func (r *outboundRepository) GetRecentAggregationResultsMultiDevice(ctx context.Context, intervalSeconds int, windowDuration time.Duration, deviceIDs []uint) ([]model.AggregationResult, error) {
	if windowDuration <= 0 {
		windowDuration = time.Hour
	}

	buildBase := func() *gorm.DB {
		q := applyActiveDeviceParamScope(ctx, r.db).
			Where("aggregation_results.source_type = ?", model.SourceCustomerProcessed)
		if intervalSeconds > 0 {
			q = q.Where("aggregation_results.interval_seconds = ?", intervalSeconds)
		}
		if len(deviceIDs) > 0 {
			q = q.Where("aggregation_results.device_id IN (?)", deviceIDs)
		}
		return q
	}

	var latest model.AggregationResult
	nowUTC := time.Now().UTC()
	err := buildBase().Where("aggregation_results.period_end <= ?", nowUTC).Order("aggregation_results.period_end DESC").First(&latest).Error
	if err != nil {
		err = buildBase().Order("aggregation_results.period_end DESC").First(&latest).Error
		if err != nil {
			return nil, err
		}
	}

	windowStart := latest.PeriodEnd.Add(-windowDuration)
	var results []model.AggregationResult
	err = buildBase().
		Preload("Parameter").
		Preload("Device").
		Where("aggregation_results.period_start >= ? AND aggregation_results.period_end <= ?", windowStart, latest.PeriodEnd).
		Order("aggregation_results.period_start ASC, aggregation_results.device_id ASC, aggregation_results.parameter_id ASC").
		Find(&results).Error
	return results, err
}

// EnqueueDelivery persists a new delivery queue entry with idempotency check and pending batch merge
func (r *outboundRepository) EnqueueDelivery(ctx context.Context, item *model.OutboundDeliveryQueue) error {
	_, _, err := r.EnqueueOrRefreshDelivery(ctx, item)
	return err
}

// EnqueueOrRefreshDelivery persists a delivery queue entry or refreshes an undelivered/cancelled entry with the same idempotency key.
// Returns action ("CREATED", "UPDATED", or "SKIPPED_DELIVERED") and the prior status when an item already existed.
func (r *outboundRepository) EnqueueOrRefreshDelivery(ctx context.Context, item *model.OutboundDeliveryQueue) (string, model.OutboundDeliveryStatus, error) {
	if item.TenantID == 0 {
		if item.Destination != nil && item.Destination.TenantID > 0 {
			item.TenantID = item.Destination.TenantID
			if item.SiteID == nil && item.Destination.SiteID != nil {
				item.SiteID = item.Destination.SiteID
			}
		} else if item.DestinationID > 0 {
			var dest model.OutputDestination
			if err := r.db.WithContext(ctx).Select("id", "tenant_id", "site_id").First(&dest, item.DestinationID).Error; err == nil && dest.TenantID > 0 {
				item.TenantID = dest.TenantID
				if item.SiteID == nil && dest.SiteID != nil {
					item.SiteID = dest.SiteID
				}
			}
		}
		if item.TenantID == 0 {
			item.TenantID = tenantctx.FromContext(ctx).WriteTenantID(0)
		}
	}
	var existing model.OutboundDeliveryQueue
	err := r.db.WithContext(ctx).Where("idempotency_key = ?", item.IdempotencyKey).First(&existing).Error
	if err == nil {
		if item.IsReplay {
			if cErr := r.db.WithContext(ctx).Create(item).Error; cErr != nil {
				return "", existing.Status, cErr
			}
			return "CREATED", existing.Status, nil
		}
		prevStatus := existing.Status
		if existing.Status == model.DeliveryPending || existing.Status == model.DeliveryCancelled {
			if strings.TrimSpace(item.BatchID) != "" {
				existing.BatchID = item.BatchID
			}
			existing.DeviceCount = item.DeviceCount
			existing.DevicesSummary = item.DevicesSummary
			existing.ParameterCount = item.ParameterCount
			existing.SourceAggregationIDs = item.SourceAggregationIDs
			existing.SourceValuesJSON = item.SourceValuesJSON
			existing.PayloadSnapshot = item.PayloadSnapshot
			existing.PayloadVersion = item.PayloadVersion
			if existing.Status == model.DeliveryCancelled {
				existing.Status = model.DeliveryPending
				existing.Attempts = 0
				existing.LastError = ""
				existing.NextRetryAt = nil
			}
			existing.UpdatedAt = time.Now()
			if sErr := r.db.WithContext(ctx).Save(&existing).Error; sErr != nil {
				return "", prevStatus, sErr
			}
			return "UPDATED", prevStatus, nil
		}
		// If already claimed (SENDING) or DELIVERED/ACKNOWLEDGED, preserve historical delivery record
		return "SKIPPED_DELIVERED", prevStatus, nil
	}
	if cErr := r.db.WithContext(ctx).Create(item).Error; cErr != nil {
		return "", "", cErr
	}
	return "CREATED", model.DeliveryPending, nil
}

// GetQueueItemByID retrieves a delivery queue item by its primary key
func (r *outboundRepository) GetQueueItemByID(ctx context.Context, id uint64) (*model.OutboundDeliveryQueue, error) {
	var item model.OutboundDeliveryQueue
	err := r.db.WithContext(ctx).
		Preload("Destination", func(db *gorm.DB) *gorm.DB { return db.Unscoped() }).
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

	scope := tenantctx.FromContext(ctx)
	if filter.TenantID == 0 && scope.TenantID > 0 {
		filter.TenantID = scope.TenantID
	}
	if filter.SiteID == 0 && scope.SiteID > 0 {
		filter.SiteID = scope.SiteID
	}
	if len(filter.AllowedSiteIDs) == 0 && len(scope.AllowedSiteIDs) > 0 {
		filter.AllowedSiteIDs = scope.AllowedSiteIDs
	}

	query := r.db.WithContext(ctx).Model(&model.OutboundDeliveryQueue{})

	if filter.TenantID > 0 {
		query = query.Where("tenant_id = ?", filter.TenantID)
	}
	if filter.SiteID > 0 {
		query = query.Where("(site_id IS NULL OR site_id = ?)", filter.SiteID)
	} else if len(filter.AllowedSiteIDs) > 0 {
		query = query.Where("(site_id IS NULL OR site_id IN ?)", filter.AllowedSiteIDs)
	}
	if filter.DestinationID != nil && *filter.DestinationID > 0 {
		query = query.Where("destination_id = ?", *filter.DestinationID)
	}
	if filter.DeviceID != nil && *filter.DeviceID > 0 {
		query = query.Where("device_id = ?", *filter.DeviceID)
	}
	if filter.Status != "" && filter.Status != "ALL" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.UndeliveredOnly {
		query = query.Where("status NOT IN ?", []string{
			string(model.DeliveryDelivered),
			string(model.DeliveryAcknowledged),
			string(model.DeliveryCancelled),
		})
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
	if filter.WindowStart != nil {
		query = query.Where("period_end > ?", *filter.WindowStart)
	}
	if filter.WindowEnd != nil {
		query = query.Where("period_start < ?", *filter.WindowEnd)
	}
	if strings.TrimSpace(filter.Search) != "" {
		s := "%" + strings.TrimSpace(filter.Search) + "%"
		query = query.Where("batch_id LIKE ? OR idempotency_key LIKE ? OR devices_summary LIKE ?", s, s, s)
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
		Preload("Destination", func(db *gorm.DB) *gorm.DB { return db.Unscoped() }).
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

	scope := tenantctx.FromContext(ctx)
	q := r.db.WithContext(ctx).Model(&model.OutboundDeliveryQueue{})
	if scope.TenantID > 0 {
		q = q.Where("tenant_id = ?", scope.TenantID)
	}
	if scope.SiteID > 0 {
		q = q.Where("(site_id IS NULL OR site_id = ?)", scope.SiteID)
	} else if len(scope.AllowedSiteIDs) > 0 {
		q = q.Where("(site_id IS NULL OR site_id IN ?)", scope.AllowedSiteIDs)
	}

	err := q.
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

// RecordDestinationLastDelivery updates only the last_delivery timestamp on a destination without clobbering concurrent config edits
func (r *outboundRepository) RecordDestinationLastDelivery(ctx context.Context, destinationID uint, deliveredAt time.Time) error {
	return r.db.WithContext(ctx).Model(&model.OutputDestination{}).
		Where("id = ?", destinationID).
		Update("last_delivery", &deliveredAt).Error
}

// SyncDestinationStatsFromQueue reconciles success_deliveries, failed_deliveries, and status on output_destinations
// directly from active outbound_delivery_queues rows so Network & Delivery Monitor and Destination Cards never retain stale/purged errors.
func (r *outboundRepository) SyncDestinationStatsFromQueue(ctx context.Context, destinationID *uint) error {
	var destinations []model.OutputDestination
	dq := r.db.WithContext(ctx).Model(&model.OutputDestination{})
	if destinationID != nil && *destinationID > 0 {
		dq = dq.Where("id = ?", *destinationID)
	}
	if err := dq.Find(&destinations).Error; err != nil {
		return err
	}
	if len(destinations) == 0 {
		return nil
	}

	type destStatusCount struct {
		DestinationID uint
		Status        string
		Count         int64
	}
	var counts []destStatusCount
	qq := r.db.WithContext(ctx).Model(&model.OutboundDeliveryQueue{}).
		Select("destination_id, status, COUNT(*) as count").
		Group("destination_id, status")
	if destinationID != nil && *destinationID > 0 {
		qq = qq.Where("destination_id = ?", *destinationID)
	}
	if err := qq.Scan(&counts).Error; err != nil {
		return err
	}

	type agg struct {
		delivered int64
		failed    int64
		retrying  int64
	}
	aggByDest := make(map[uint]agg, len(destinations))
	for _, c := range counts {
		a := aggByDest[c.DestinationID]
		switch model.OutboundDeliveryStatus(c.Status) {
		case model.DeliveryDelivered, model.DeliveryAcknowledged:
			a.delivered += c.Count
		case model.DeliveryFailed:
			a.failed += c.Count
		case model.DeliveryRetrying:
			a.retrying += c.Count
		}
		aggByDest[c.DestinationID] = a
	}

	for _, d := range destinations {
		a := aggByDest[d.ID]
		nextStatus := "IDLE"
		if a.failed > 0 || a.retrying > 0 {
			nextStatus = "ERROR"
		} else if a.delivered > 0 {
			nextStatus = "ACTIVE"
		}
		if err := r.db.WithContext(ctx).Model(&model.OutputDestination{}).
			Where("id = ?", d.ID).
			Updates(map[string]interface{}{
				"success_deliveries": a.delivered,
				"failed_deliveries":  a.failed,
				"status":             nextStatus,
			}).Error; err != nil {
			return err
		}
	}

	return nil
}

// CancelQueueItem marks a pending or retrying queue item as CANCELLED
func (r *outboundRepository) CancelQueueItem(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Model(&model.OutboundDeliveryQueue{}).
		Where("id = ? AND status IN ?", id, []string{string(model.DeliveryPending), string(model.DeliveryRetrying)}).
		Update("status", model.DeliveryCancelled).Error
}

// GetUndeliveredQueueItemsByDestination retrieves all PENDING, RETRYING, and FAILED queue items for a destination
func (r *outboundRepository) GetUndeliveredQueueItemsByDestination(ctx context.Context, destinationID uint) ([]model.OutboundDeliveryQueue, error) {
	var items []model.OutboundDeliveryQueue
	err := r.db.WithContext(ctx).
		Where("destination_id = ? AND status IN ?", destinationID, []string{
			string(model.DeliveryPending),
			string(model.DeliveryRetrying),
			string(model.DeliveryFailed),
		}).
		Order("id ASC").
		Find(&items).Error
	return items, err
}

// GetUndeliveredQueueItemsOverlappingWindow retrieves all PENDING, RETRYING, and FAILED queue items whose [period_start, period_end] overlaps [startTime, endTime]
func (r *outboundRepository) GetUndeliveredQueueItemsOverlappingWindow(ctx context.Context, startTime, endTime time.Time) ([]model.OutboundDeliveryQueue, error) {
	scope := tenantctx.FromContext(ctx)
	var items []model.OutboundDeliveryQueue
	q := r.db.WithContext(ctx).
		Preload("Destination").
		Preload("Destination.PayloadConfig").
		Where("status IN ?", []string{
			string(model.DeliveryPending),
			string(model.DeliveryRetrying),
			string(model.DeliveryFailed),
		})
	if scope.TenantID > 0 {
		q = q.Where("tenant_id = ?", scope.TenantID)
	}
	if startTime.Equal(endTime) {
		q = q.Where("period_end >= ? AND period_start <= ?", startTime, endTime)
	} else {
		q = q.Where("period_end > ? AND period_start < ?", startTime, endTime)
	}
	err := q.Order("id ASC").Find(&items).Error
	return items, err
}

// UpdateQueueItemPayloadAndRequeue persists refreshed payload snapshot metadata and optionally resets the item to PENDING (attempts=0)
func (r *outboundRepository) UpdateQueueItemPayloadAndRequeue(ctx context.Context, item *model.OutboundDeliveryQueue, requeue bool) error {
	if item == nil || item.ID == 0 {
		return nil
	}

	updates := map[string]interface{}{
		"payload_snapshot": item.PayloadSnapshot,
		"payload_version":  item.PayloadVersion,
		"parameter_count":  item.ParameterCount,
		"device_count":     item.DeviceCount,
		"devices_summary":  item.DevicesSummary,
	}
	if strings.TrimSpace(item.BatchID) != "" {
		updates["batch_id"] = item.BatchID
	}
	if item.SourceValuesJSON != "" {
		updates["source_values_json"] = item.SourceValuesJSON
	}
	if item.SourceAggregationIDs != "" {
		updates["source_aggregation_ids"] = item.SourceAggregationIDs
	}

	query := r.db.WithContext(ctx).Model(&model.OutboundDeliveryQueue{}).Where("id = ?", item.ID)
	if requeue {
		now := time.Now()
		updates["status"] = model.DeliveryPending
		updates["attempts"] = 0
		updates["next_retry_at"] = &now
		updates["last_error"] = ""
		query = query.Where("status IN ?", []string{
			string(model.DeliveryPending),
			string(model.DeliveryRetrying),
			string(model.DeliveryFailed),
		})
	}

	return query.Updates(updates).Error
}

// ResetQueueItemForRetry resets an item to PENDING for immediate retry and syncs destination stats
func (r *outboundRepository) ResetQueueItemForRetry(ctx context.Context, id uint64) error {
	var item model.OutboundDeliveryQueue
	_ = r.db.WithContext(ctx).Select("id", "destination_id").Where("id = ?", id).First(&item).Error

	now := time.Now()
	if err := r.db.WithContext(ctx).Model(&model.OutboundDeliveryQueue{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":        model.DeliveryPending,
			"attempts":      0,
			"next_retry_at": &now,
			"last_error":    "",
		}).Error; err != nil {
		return err
	}
	if item.DestinationID > 0 {
		_ = r.SyncDestinationStatsFromQueue(ctx, &item.DestinationID)
	}
	return nil
}

// CreateDeliveryLog persists a transmission log entry
func (r *outboundRepository) CreateDeliveryLog(ctx context.Context, log *model.DeliveryLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

// DeleteQueueItem permanently removes a single queue item from the database and syncs destination stats
func (r *outboundRepository) DeleteQueueItem(ctx context.Context, id uint64) error {
	var item model.OutboundDeliveryQueue
	_ = r.db.WithContext(ctx).Select("id", "destination_id").Where("id = ?", id).First(&item).Error

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		_ = tx.Unscoped().Where("queue_item_id = ?", id).Delete(&model.DeliveryLog{}).Error
		return tx.Unscoped().Where("id = ?", id).Delete(&model.OutboundDeliveryQueue{}).Error
	})
	if err == nil && item.DestinationID > 0 {
		_ = r.SyncDestinationStatsFromQueue(ctx, &item.DestinationID)
	}
	return err
}

// PurgeQueueItems manually purges queue items and optional delivery logs matching criteria
func (r *outboundRepository) PurgeQueueItems(ctx context.Context, filter QueuePurgeFilter) (int64, int64, error) {
	scope := tenantctx.FromContext(ctx)
	if filter.TenantID == 0 && scope.TenantID > 0 {
		filter.TenantID = scope.TenantID
	}
	var deletedQueue int64
	var deletedLogs int64

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. Purge delivery logs if requested (delete before queues to avoid foreign key / dangling references)
		if filter.PurgeLogs {
			logQuery := "DELETE FROM delivery_logs WHERE 1 = 1"
			var logArgs []interface{}

			if filter.TenantID > 0 {
				logQuery += " AND destination_id IN (SELECT id FROM output_destinations WHERE tenant_id = ?)"
				logArgs = append(logArgs, filter.TenantID)
			}
			if filter.DestinationID != nil && *filter.DestinationID > 0 {
				logQuery += " AND destination_id = ?"
				logArgs = append(logArgs, *filter.DestinationID)
			}
			if filter.OlderThan != nil {
				logQuery += " AND delivered_at <= ?"
				logArgs = append(logArgs, *filter.OlderThan)
			}
			if filter.Status != "" && filter.Status != "ALL" {
				if filter.Status == string(model.DeliveryDelivered) {
					logQuery += " AND status IN ('DELIVERED', 'ACKNOWLEDGED', 'SUCCESS')"
				} else {
					logQuery += " AND status = ?"
					logArgs = append(logArgs, filter.Status)
				}
			}

			lres := tx.Exec(logQuery, logArgs...)
			if lres.Error != nil {
				return lres.Error
			}
			deletedLogs = lres.RowsAffected
		}

		// 2. Purge outbound delivery queue items
		queueQuery := "DELETE FROM outbound_delivery_queues WHERE 1 = 1"
		var queueArgs []interface{}

		if filter.TenantID > 0 {
			queueQuery += " AND tenant_id = ?"
			queueArgs = append(queueArgs, filter.TenantID)
		}
		if filter.DestinationID != nil && *filter.DestinationID > 0 {
			queueQuery += " AND destination_id = ?"
			queueArgs = append(queueArgs, *filter.DestinationID)
		}
		if filter.Status != "" && filter.Status != "ALL" {
			if filter.Status == string(model.DeliveryDelivered) {
				queueQuery += " AND status IN ('DELIVERED', 'ACKNOWLEDGED')"
			} else {
				queueQuery += " AND status = ?"
				queueArgs = append(queueArgs, filter.Status)
			}
		}
		if filter.OlderThan != nil {
			queueQuery += " AND created_at <= ?"
			queueArgs = append(queueArgs, *filter.OlderThan)
		}

		res := tx.Exec(queueQuery, queueArgs...)
		if res.Error != nil {
			return res.Error
		}
		deletedQueue = res.RowsAffected

		return nil
	})

	if err == nil {
		_ = r.SyncDestinationStatsFromQueue(ctx, filter.DestinationID)
	}

	return deletedQueue, deletedLogs, err
}

// CountQueuePurgeItems counts the number of queue entries and logs that match the purge criteria
func (r *outboundRepository) CountQueuePurgeItems(ctx context.Context, filter QueuePurgeFilter) (int64, int64, error) {
	scope := tenantctx.FromContext(ctx)
	if filter.TenantID == 0 && scope.TenantID > 0 {
		filter.TenantID = scope.TenantID
	}
	var queueCount int64
	var logsCount int64

	q := r.db.WithContext(ctx).Model(&model.OutboundDeliveryQueue{})
	if filter.TenantID > 0 {
		q = q.Where("tenant_id = ?", filter.TenantID)
	}
	if filter.DestinationID != nil && *filter.DestinationID > 0 {
		q = q.Where("destination_id = ?", *filter.DestinationID)
	}
	if filter.Status != "" && filter.Status != "ALL" {
		if filter.Status == string(model.DeliveryDelivered) {
			q = q.Where("status IN ?", []string{string(model.DeliveryDelivered), string(model.DeliveryAcknowledged)})
		} else {
			q = q.Where("status = ?", filter.Status)
		}
	}
	if filter.OlderThan != nil {
		q = q.Where("created_at <= ?", *filter.OlderThan)
	}
	if err := q.Count(&queueCount).Error; err != nil {
		return 0, 0, err
	}

	if filter.PurgeLogs {
		lq := r.db.WithContext(ctx).Model(&model.DeliveryLog{})
		if filter.TenantID > 0 {
			lq = lq.Where("destination_id IN (?)", r.db.Model(&model.OutputDestination{}).Select("id").Where("tenant_id = ?", filter.TenantID))
		}
		if filter.DestinationID != nil && *filter.DestinationID > 0 {
			lq = lq.Where("destination_id = ?", *filter.DestinationID)
		}
		if filter.OlderThan != nil {
			lq = lq.Where("delivered_at <= ?", *filter.OlderThan)
		}
		if filter.Status != "" && filter.Status != "ALL" {
			if filter.Status == string(model.DeliveryDelivered) {
				lq = lq.Where("status IN ?", []string{string(model.DeliveryDelivered), string(model.DeliveryAcknowledged), "SUCCESS"})
			} else {
				lq = lq.Where("status = ?", filter.Status)
			}
		}
		if err := lq.Count(&logsCount).Error; err != nil {
			return 0, 0, err
		}
	}

	return queueCount, logsCount, nil
}

// PurgeQueueBatch deletes a single batch of matching items by ID, returning deleted counts and done indicator
func (r *outboundRepository) PurgeQueueBatch(ctx context.Context, filter QueuePurgeFilter, batchSize int) (int64, int64, bool, error) {
	scope := tenantctx.FromContext(ctx)
	if filter.TenantID == 0 && scope.TenantID > 0 {
		filter.TenantID = scope.TenantID
	}
	if batchSize <= 0 {
		batchSize = 1000
	}
	var deletedLogs int64
	var deletedQueue int64
	var moreLogs bool
	var moreQueue bool

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. Drain matching delivery logs in batches first to respect foreign key ordering
		if filter.PurgeLogs {
			var logIDs []uint
			lq := tx.Model(&model.DeliveryLog{}).Select("id")
			if filter.TenantID > 0 {
				lq = lq.Where("destination_id IN (?)", tx.Model(&model.OutputDestination{}).Select("id").Where("tenant_id = ?", filter.TenantID))
			}
			if filter.DestinationID != nil && *filter.DestinationID > 0 {
				lq = lq.Where("destination_id = ?", *filter.DestinationID)
			}
			if filter.OlderThan != nil {
				lq = lq.Where("delivered_at <= ?", *filter.OlderThan)
			}
			if filter.Status != "" && filter.Status != "ALL" {
				if filter.Status == string(model.DeliveryDelivered) {
					lq = lq.Where("status IN ?", []string{string(model.DeliveryDelivered), string(model.DeliveryAcknowledged), "SUCCESS"})
				} else {
					lq = lq.Where("status = ?", filter.Status)
				}
			}
			if err := lq.Limit(batchSize).Pluck("id", &logIDs).Error; err != nil {
				return err
			}
			if len(logIDs) > 0 {
				res := tx.Unscoped().Where("id IN ?", logIDs).Delete(&model.DeliveryLog{})
				if res.Error != nil {
					return res.Error
				}
				deletedLogs += res.RowsAffected
				moreLogs = len(logIDs) == batchSize
				if moreLogs {
					return nil
				}
			}
		}

		// 2. Delete queue items batch by primary key IDs
		var queueIDs []uint64
		q := tx.Model(&model.OutboundDeliveryQueue{}).Select("id")
		if filter.TenantID > 0 {
			q = q.Where("tenant_id = ?", filter.TenantID)
		}
		if filter.DestinationID != nil && *filter.DestinationID > 0 {
			q = q.Where("destination_id = ?", *filter.DestinationID)
		}
		if filter.Status != "" && filter.Status != "ALL" {
			if filter.Status == string(model.DeliveryDelivered) {
				q = q.Where("status IN ?", []string{string(model.DeliveryDelivered), string(model.DeliveryAcknowledged)})
			} else {
				q = q.Where("status = ?", filter.Status)
			}
		}
		if filter.OlderThan != nil {
			q = q.Where("created_at <= ?", *filter.OlderThan)
		}
		if err := q.Limit(batchSize).Pluck("id", &queueIDs).Error; err != nil {
			return err
		}
		if len(queueIDs) > 0 {
			if filter.PurgeLogs {
				// Clean up any remaining delivery logs referencing these queue items (e.g. earlier RETRYING logs on a FAILED item)
				lres := tx.Unscoped().Where("queue_item_id IN ?", queueIDs).Delete(&model.DeliveryLog{})
				if lres.Error != nil {
					return lres.Error
				}
				deletedLogs += lres.RowsAffected
			} else {
				// Unlink queue_item_id so FK constraints cannot block queue deletion while keeping log history
				if err := tx.Model(&model.DeliveryLog{}).Where("queue_item_id IN ?", queueIDs).Update("queue_item_id", nil).Error; err != nil {
					return err
				}
			}

			res := tx.Unscoped().Where("id IN ?", queueIDs).Delete(&model.OutboundDeliveryQueue{})
			if res.Error != nil {
				return res.Error
			}
			deletedQueue = res.RowsAffected
			moreQueue = len(queueIDs) == batchSize
		}

		return nil
	})

	done := !moreLogs && !moreQueue
	if err == nil && done {
		_ = r.SyncDestinationStatsFromQueue(ctx, filter.DestinationID)
	}
	return deletedQueue, deletedLogs, done, err
}
