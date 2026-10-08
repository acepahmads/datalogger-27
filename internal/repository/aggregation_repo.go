package repository

import (
	"context"
	"fmt"
	"time"

	"datalogger/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// AggregationResultFilter encapsulates query criteria for aggregation results
type AggregationResultFilter struct {
	DefinitionID uint
	DeviceID     uint
	ParameterID  uint
	SourceType   model.AggregationSourceType
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
	query := r.db.WithContext(ctx).Preload("Device").Preload("Parameter").Order("id asc")
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

// UpdateDefinitionLastCalculated updates the last calculation timestamp for a definition
func (r *AggregationRepository) UpdateDefinitionLastCalculated(ctx context.Context, defID uint, t time.Time) error {
	return r.db.WithContext(ctx).Model(&model.AggregationDefinition{}).Where("id = ?", defID).Update("last_calculated_at", t).Error
}

// SaveResult executes an idempotent UPSERT for an aggregation result
func (r *AggregationRepository) SaveResult(ctx context.Context, result *model.AggregationResult) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "aggregation_definition_id"},
			{Name: "period_start"},
		},
		DoUpdates: clause.AssignmentColumns([]string{
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
	query := r.db.WithContext(ctx).Where("identifier = ?", identifier)
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
	query := r.db.WithContext(ctx).Model(&model.AggregationResult{})

	if filter.DefinitionID > 0 {
		query = query.Where("aggregation_definition_id = ?", filter.DefinitionID)
	}
	if filter.DeviceID > 0 {
		query = query.Where("device_id = ?", filter.DeviceID)
	}
	if filter.ParameterID > 0 {
		query = query.Where("parameter_id = ?", filter.ParameterID)
	}
	if filter.SourceType != "" {
		query = query.Where("source_type = ?", filter.SourceType)
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
	err := r.db.WithContext(ctx).
		Where("device_id = ? AND parameter_id = ? AND received_at >= ? AND received_at < ?", deviceID, paramID, start, end).
		Order("received_at asc").
		Find(&samples).Error
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
