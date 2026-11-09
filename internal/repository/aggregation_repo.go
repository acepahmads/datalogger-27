package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"datalogger/internal/model"
	"datalogger/pkg/tenantctx"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// AggregationResultFilter encapsulates query criteria for aggregation results
type AggregationResultFilter struct {
	TenantID       uint
	SiteID         uint
	AllowedSiteIDs []uint
	DefinitionID   uint
	DeviceID     uint
	DeviceIDs    []uint
	ParameterID  uint
	ParameterIDs []uint
	SourceType   model.AggregationSourceType
	Function     model.AggregationFunction
	Identifier   string
	StartTime    *time.Time
	EndTime      *time.Time
	Quality      string
	Page         int
	PageSize     int
}

// AggregationRepository handles database operations for aggregation definitions and results
type AggregationRepository struct {
	db *gorm.DB
}

// NewAggregationRepository creates a new instance of AggregationRepository
func NewAggregationRepository(db *gorm.DB) *AggregationRepository {
	return &AggregationRepository{db: db}
}

// CreateDefinition persists a new aggregation definition
func (r *AggregationRepository) CreateDefinition(ctx context.Context, def *model.AggregationDefinition) error {
	return r.db.WithContext(ctx).Create(def).Error
}

// UpdateDefinition updates an existing aggregation definition
func (r *AggregationRepository) UpdateDefinition(ctx context.Context, def *model.AggregationDefinition) error {
	return r.db.WithContext(ctx).Save(def).Error
}

// DeleteDefinition soft-deletes an aggregation definition
func (r *AggregationRepository) DeleteDefinition(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.AggregationDefinition{}, id).Error
}

// GetDefinitionByID retrieves an aggregation definition with preloaded relations
func (r *AggregationRepository) GetDefinitionByID(ctx context.Context, id uint) (*model.AggregationDefinition, error) {
	var def model.AggregationDefinition
	if err := r.db.WithContext(ctx).Preload("Device").Preload("Parameter").First(&def, id).Error; err != nil {
		return nil, err
	}
	return &def, nil
}

// ListDefinitions retrieves aggregation definitions with optional device and enabled filtering
func (r *AggregationRepository) ListDefinitions(ctx context.Context, deviceID uint, enabledOnly bool) ([]model.AggregationDefinition, error) {
	scope := tenantctx.FromContext(ctx)
	query := r.db.WithContext(ctx).Preload("Device").Preload("Parameter").Order("id asc")
	if scope.TenantID > 0 {
		query = query.Where("tenant_id = ?", scope.TenantID)
	}
	if len(scope.AllowedSiteIDs) > 0 {
		query = query.Where("site_id IN ?", scope.AllowedSiteIDs)
	}
	if scope.SiteID > 0 {
		query = query.Where("site_id = ?", scope.SiteID)
	}
	if deviceID > 0 {
		query = query.Where("device_id = ?", deviceID)
	}
	if enabledOnly {
		query = query.Where("enabled = ?", true)
	}
	var defs []model.AggregationDefinition
	if err := query.Find(&defs).Error; err != nil {
		return nil, err
	}
	return defs, nil
}

// isTransientAggregationDBError detects transient MariaDB/MySQL/SQLite lock contention,
// deadlocks (Error 1213), lock wait timeouts (Error 1205), or pooled idle connection resets.
func isTransientAggregationDBError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "deadlock") ||
		strings.Contains(msg, "1213") ||
		strings.Contains(msg, "lock wait timeout") ||
		strings.Contains(msg, "1205") ||
		strings.Contains(msg, "try restarting transaction") ||
		strings.Contains(msg, "database is locked") ||
		strings.Contains(msg, "database table is locked") ||
		strings.Contains(msg, "sqlite_busy") ||
		strings.Contains(msg, "bad connection") ||
		strings.Contains(msg, "invalid connection")
}

func withAggregationDBRetry(ctx context.Context, op func() error) error {
	backoffs := []time.Duration{15 * time.Millisecond, 40 * time.Millisecond, 100 * time.Millisecond, 250 * time.Millisecond}
	var lastErr error
	for attempt := 0; attempt <= len(backoffs); attempt++ {
		if ctx != nil && ctx.Err() != nil {
			return ctx.Err()
		}
		lastErr = op()
		if lastErr == nil || !isTransientAggregationDBError(lastErr) || attempt == len(backoffs) {
			return lastErr
		}
		wait := backoffs[attempt]
		if ctx != nil {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(wait):
			}
		} else {
			time.Sleep(wait)
		}
	}
	return lastErr
}

// UpdateDefinitionLastCalculated updates the last calculation timestamp for a definition
func (r *AggregationRepository) UpdateDefinitionLastCalculated(ctx context.Context, defID uint, t time.Time) error {
	return withAggregationDBRetry(ctx, func() error {
		return r.db.WithContext(ctx).Model(&model.AggregationDefinition{}).Where("id = ?", defID).Update("last_calculated_at", t).Error
	})
}

// SaveResult executes an idempotent UPSERT for an aggregation result with automatic transient deadlock/lock retry
func (r *AggregationRepository) SaveResult(ctx context.Context, result *model.AggregationResult) error {
	if result == nil {
		return nil
	}
	origID := result.ID
	return withAggregationDBRetry(ctx, func() error {
		result.ID = origID
		return r.db.WithContext(ctx).Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "aggregation_definition_id"},
				{Name: "period_start"},
			},
			DoUpdates: clause.AssignmentColumns([]string{
				"tenant_id",
				"site_id",
				"identifier",
				"period_end",
				"value",
				"min_value",
				"max_value",
				"avg_value",
				"sum_value",
				"first_value",
				"last_value",
				"sample_count",
				"valid_count",
				"good_count",
				"uncertain_count",
				"bad_count",
				"stale_count",
				"quality",
				"quality_reason",
				"updated_at",
			}),
		}).Create(result).Error
	})
}

// DeleteResultForBucket deletes an aggregation result row for a specific definition and bucket start (used during injection rollback)
func (r *AggregationRepository) DeleteResultForBucket(ctx context.Context, defID uint, periodStart time.Time) error {
	return withAggregationDBRetry(ctx, func() error {
		return r.db.WithContext(ctx).
			Where("aggregation_definition_id = ? AND period_start = ?", defID, periodStart).
			Delete(&model.AggregationResult{}).Error
	})
}

// GetResultByID retrieves an aggregation result by primary key ID
func (r *AggregationRepository) GetResultByID(ctx context.Context, id uint) (*model.AggregationResult, error) {
	var result model.AggregationResult
	err := r.db.WithContext(ctx).First(&result, id).Error
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GetResultByIdentifier retrieves results by deterministic identifier (e.g. 20261008140000)
func (r *AggregationRepository) GetResultByIdentifier(ctx context.Context, identifier string, sourceType model.AggregationSourceType) ([]model.AggregationResult, error) {
	scope := tenantctx.FromContext(ctx)
	query := r.db.WithContext(ctx).Where("identifier = ?", identifier)
	if scope.TenantID > 0 {
		query = query.Where("tenant_id = ?", scope.TenantID)
	}
	if len(scope.AllowedSiteIDs) > 0 {
		query = query.Where("site_id IN ?", scope.AllowedSiteIDs)
	}
	if scope.SiteID > 0 {
		query = query.Where("site_id = ?", scope.SiteID)
	}
	if sourceType != "" {
		query = query.Where("source_type = ?", sourceType)
	}
	var results []model.AggregationResult
	if err := query.Order("parameter_id asc").Find(&results).Error; err != nil {
		return nil, err
	}
	return results, nil
}

// GetResultsByFilter retrieves paginated aggregation results based on filter criteria
func (r *AggregationRepository) GetResultsByFilter(ctx context.Context, filter AggregationResultFilter) ([]model.AggregationResult, int64, error) {
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

	query := r.db.WithContext(ctx).Model(&model.AggregationResult{}).Preload("Device").Preload("Parameter")
	if filter.TenantID > 0 {
		query = query.Where("tenant_id = ?", filter.TenantID)
	}
	if len(filter.AllowedSiteIDs) > 0 {
		query = query.Where("site_id IN ?", filter.AllowedSiteIDs)
	}
	if filter.SiteID > 0 {
		query = query.Where("site_id = ?", filter.SiteID)
	}

	if filter.DefinitionID > 0 {
		query = query.Where("aggregation_definition_id = ?", filter.DefinitionID)
	}
	if len(filter.DeviceIDs) > 0 {
		query = query.Where("device_id IN (?)", filter.DeviceIDs)
	} else if filter.DeviceID > 0 {
		query = query.Where("device_id = ?", filter.DeviceID)
	}

	if len(filter.ParameterIDs) > 0 {
		query = query.Where("parameter_id IN (?)", filter.ParameterIDs)
	} else if filter.ParameterID > 0 {
		query = query.Where("parameter_id = ?", filter.ParameterID)
	}
	if filter.SourceType != "" {
		query = query.Where("source_type = ?", filter.SourceType)
	}
	if filter.Function != "" {
		query = query.Where("function = ?", filter.Function)
	}
	if filter.Identifier != "" {
		query = query.Where("identifier LIKE ?", "%"+filter.Identifier+"%")
	}
	if filter.Quality != "" {
		query = query.Where("quality = ?", filter.Quality)
	}
	if filter.StartTime != nil {
		query = query.Where("period_start >= ?", *filter.StartTime)
	}
	if filter.EndTime != nil {
		query = query.Where("period_end <= ?", *filter.EndTime)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 50
	}
	offset := (filter.Page - 1) * filter.PageSize

	var results []model.AggregationResult
	if err := query.Order("period_start desc, parameter_id asc").Limit(filter.PageSize).Offset(offset).Find(&results).Error; err != nil {
		return nil, 0, err
	}

	return results, total, nil
}

// GetRawTelemetryForBucket retrieves telemetry samples from raw_data for a specific device, parameter, and time window
func (r *AggregationRepository) GetRawTelemetryForBucket(ctx context.Context, deviceID, paramID uint, start, end time.Time) ([]model.RawData, error) {
	var samples []model.RawData
	err := withAggregationDBRetry(ctx, func() error {
		samples = nil
		return r.db.WithContext(ctx).
			Where("device_id = ? AND parameter_id = ? AND received_at >= ? AND received_at < ?", deviceID, paramID, start, end).
			Order("received_at asc").
			Find(&samples).Error
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch raw telemetry for bucket [%v, %v): %w", start, end, err)
	}
	return samples, nil
}

// GetLatestTelemetryTime returns the timestamp of the newest telemetry sample for a parameter
func (r *AggregationRepository) GetLatestTelemetryTime(ctx context.Context, deviceID, paramID uint) (*time.Time, error) {
	var latestRec model.RawData
	err := r.db.WithContext(ctx).
		Where("device_id = ? AND parameter_id = ?", deviceID, paramID).
		Order("received_at desc").
		First(&latestRec).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &latestRec.ReceivedAt, nil
}
