package repository

import (
	"context"
	"time"

	"datalogger/internal/model"

	"gorm.io/gorm"
)

// ISPUHistoryFilter encapsulates query criteria for querying historical ISPU calculation records.
type ISPUHistoryFilter struct {
	DeviceID          uint
	StartTime         *time.Time
	EndTime           *time.Time
	Category          string
	DominantPollutant string
	Status            string
	Page              int
	PageSize          int
}

// ISPURepository handles database operations for ISPU calculations, breakdown details, and parameter mappings.
type ISPURepository struct {
	db *gorm.DB
}

// NewISPURepository creates a new instance of ISPURepository.
func NewISPURepository(db *gorm.DB) *ISPURepository {
	return &ISPURepository{db: db}
}

// UpsertCalculationResult performs deterministic, idempotent upsert of an ISPU calculation and its parameter results.
func (r *ISPURepository) UpsertCalculationResult(ctx context.Context, res *model.ISPUCalculationResult) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing model.ISPUCalculationResult
		err := tx.Where("device_id = ? AND period_start = ? AND period_end = ?",
			res.DeviceID, res.PeriodStart, res.PeriodEnd).First(&existing).Error

		if err == nil {
			// Update existing record
			res.ID = existing.ID
			res.CreatedAt = existing.CreatedAt

			// Delete old parameter results
			if err := tx.Where("calculation_result_id = ?", existing.ID).Delete(&model.ISPUParameterResult{}).Error; err != nil {
				return err
			}

			// Update parent calculation record
			if err := tx.Save(res).Error; err != nil {
				return err
			}

			// Insert new parameter results
			for i := range res.ParameterResults {
				res.ParameterResults[i].ID = 0
				res.ParameterResults[i].CalculationResultID = res.ID
				if err := tx.Create(&res.ParameterResults[i]).Error; err != nil {
					return err
				}
			}
			return nil
		} else if err == gorm.ErrRecordNotFound {
			// Create brand new calculation record with cascade
			return tx.Create(res).Error
		}
		return err
	})
}

// GetLatestByDeviceID retrieves the latest calculation result for a device, preloading relations.
func (r *ISPURepository) GetLatestByDeviceID(ctx context.Context, deviceID uint) (*model.ISPUCalculationResult, error) {
	var res model.ISPUCalculationResult
	err := r.db.WithContext(ctx).
		Preload("Device").
		Preload("ParameterResults").
		Preload("ParameterResults.Parameter").
		Where("device_id = ?", deviceID).
		Order("period_end DESC, calculated_at DESC").
		First(&res).Error
	if err != nil {
		return nil, err
	}
	return &res, nil
}

// GetByID retrieves a calculation result by ID, preloading all relations.
func (r *ISPURepository) GetByID(ctx context.Context, id uint) (*model.ISPUCalculationResult, error) {
	var res model.ISPUCalculationResult
	err := r.db.WithContext(ctx).
		Preload("Device").
		Preload("ParameterResults").
		Preload("ParameterResults.Parameter").
		First(&res, id).Error
	if err != nil {
		return nil, err
	}
	return &res, nil
}

// GetHistory retrieves paginated ISPU calculation records with filters.
func (r *ISPURepository) GetHistory(ctx context.Context, filter ISPUHistoryFilter) ([]model.ISPUCalculationResult, int64, error) {
	var results []model.ISPUCalculationResult
	var total int64

	query := r.db.WithContext(ctx).Model(&model.ISPUCalculationResult{})

	if filter.DeviceID > 0 {
		query = query.Where("device_id = ?", filter.DeviceID)
	}
	if filter.StartTime != nil {
		query = query.Where("period_end >= ?", *filter.StartTime)
	}
	if filter.EndTime != nil {
		query = query.Where("period_start <= ?", *filter.EndTime)
	}
	if filter.Category != "" && filter.Category != "ALL" {
		query = query.Where("overall_category = ?", filter.Category)
	}
	if filter.DominantPollutant != "" && filter.DominantPollutant != "ALL" {
		query = query.Where("dominant_pollutant = ? OR id IN (SELECT calculation_result_id FROM ispu_parameter_results WHERE pollutant_code = ?)",
			filter.DominantPollutant, filter.DominantPollutant)
	}
	if filter.Status != "" && filter.Status != "ALL" {
		query = query.Where("calculation_status = ?", filter.Status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	page := filter.Page
	if page < 1 {
		page = 1
	}
	pageSize := filter.PageSize
	if pageSize < 1 || pageSize > 200 {
		pageSize = 25
	}
	offset := (page - 1) * pageSize

	err := query.
		Preload("Device").
		Preload("ParameterResults").
		Preload("ParameterResults.Parameter").
		Order("period_end DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&results).Error

	if err != nil {
		return nil, 0, err
	}
	return results, total, nil
}

// GetTrend retrieves historical calculation series for trend analysis charts.
func (r *ISPURepository) GetTrend(ctx context.Context, deviceID uint, startTime, endTime time.Time) ([]model.ISPUCalculationResult, error) {
	var results []model.ISPUCalculationResult
	query := r.db.WithContext(ctx).
		Where("device_id = ? AND period_end >= ? AND period_end <= ?", deviceID, startTime.UTC(), endTime.UTC()).
		Preload("ParameterResults").
		Order("period_end ASC")

	if err := query.Find(&results).Error; err != nil {
		return nil, err
	}
	return results, nil
}

// GetMappingsByDeviceID retrieves configured ISPU parameter mappings for a device.
func (r *ISPURepository) GetMappingsByDeviceID(ctx context.Context, deviceID uint) ([]model.ISPUParameterMapping, error) {
	var mappings []model.ISPUParameterMapping
	err := r.db.WithContext(ctx).
		Preload("Parameter").
		Where("device_id = ?", deviceID).
		Find(&mappings).Error
	return mappings, err
}

// GetMappingByID retrieves a single mapping.
func (r *ISPURepository) GetMappingByID(ctx context.Context, id uint) (*model.ISPUParameterMapping, error) {
	var mapping model.ISPUParameterMapping
	err := r.db.WithContext(ctx).Preload("Parameter").First(&mapping, id).Error
	if err != nil {
		return nil, err
	}
	return &mapping, nil
}

// SaveMapping creates or updates an ISPU parameter mapping.
func (r *ISPURepository) SaveMapping(ctx context.Context, mapping *model.ISPUParameterMapping) error {
	if mapping.ID > 0 {
		return r.db.WithContext(ctx).Save(mapping).Error
	}
	return r.db.WithContext(ctx).Create(mapping).Error
}

// DeleteMapping removes an ISPU parameter mapping.
func (r *ISPURepository) DeleteMapping(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.ISPUParameterMapping{}, id).Error
}

// GetAllActiveMappings retrieves all active mappings across all devices for background calculation.
func (r *ISPURepository) GetAllActiveMappings(ctx context.Context) ([]model.ISPUParameterMapping, error) {
	var mappings []model.ISPUParameterMapping
	err := r.db.WithContext(ctx).
		Preload("Parameter").
		Preload("Device").
		Where("enabled = true").
		Find(&mappings).Error
	return mappings, err
}
