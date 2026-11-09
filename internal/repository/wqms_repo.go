package repository

import (
	"context"
	"time"

	"datalogger/internal/model"
	"datalogger/pkg/tenantctx"

	"gorm.io/gorm"
)

// WQMSHistoryFilter encapsulates query criteria for historical WQMS IP calculations.
type WQMSHistoryFilter struct {
	DeviceID          uint
	CalculationMode   string // "REGULATORY", "OPERATIONAL", or "ALL"
	WaterClass        string
	StartTime         *time.Time
	EndTime           *time.Time
	Category          string
	DominantParameter string
	Status            string
	Page              int
	PageSize          int
}

// WQMSRepository handles MariaDB/SQLite persistence for WQMS IP & IKA configs, mappings, and results.
type WQMSRepository struct {
	db *gorm.DB
}

// NewWQMSRepository creates a new instance of WQMSRepository.
func NewWQMSRepository(db *gorm.DB) *WQMSRepository {
	return &WQMSRepository{db: db}
}

func applyWQMSScope(ctx context.Context, q *gorm.DB) *gorm.DB {
	scope := tenantctx.FromContext(ctx)
	if scope.TenantID > 0 {
		q = q.Where("tenant_id = ?", scope.TenantID)
	}
	if scope.SiteID > 0 {
		q = q.Where("(site_id = ? OR site_id = 0)", scope.SiteID)
	} else if len(scope.AllowedSiteIDs) > 0 {
		q = q.Where("(site_id IN ? OR site_id = 0)", scope.AllowedSiteIDs)
	}
	return q
}

// GetOrCreateStationConfig retrieves or creates the station WQMS configuration (deviceID=0 for unified station).
func (r *WQMSRepository) GetOrCreateStationConfig(ctx context.Context, deviceID uint) (*model.WQMSStationConfig, error) {
	scope := tenantctx.FromContext(ctx)
	tID := scope.WriteTenantID(0)
	sID := scope.WriteSiteID(0)

	var cfg model.WQMSStationConfig
	err := r.db.WithContext(ctx).Where("tenant_id = ? AND site_id = ? AND device_id = ?", tID, sID, deviceID).First(&cfg).Error
	if err == gorm.ErrRecordNotFound {
		cfg = model.WQMSStationConfig{
			TenantID:        tID,
			SiteID:          sID,
			DeviceID:        deviceID,
			WaterClass:      model.WaterClass2,
			IKAWindowDays:   30,
			DOSaturationMgL: 8.0,
			MinCompleteness: 75.0,
		}
		if createErr := r.db.WithContext(ctx).Create(&cfg).Error; createErr != nil {
			// Handle concurrent insert race cleanly
			_ = r.db.WithContext(ctx).Where("tenant_id = ? AND site_id = ? AND device_id = ?", tID, sID, deviceID).First(&cfg).Error
		}
		return &cfg, nil
	}
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

// SaveStationConfig updates or creates the station WQMS configuration.
func (r *WQMSRepository) SaveStationConfig(ctx context.Context, cfg *model.WQMSStationConfig) error {
	scope := tenantctx.FromContext(ctx)
	if cfg.TenantID == 0 {
		cfg.TenantID = scope.WriteTenantID(0)
	}
	if cfg.SiteID == 0 {
		cfg.SiteID = scope.WriteSiteID(0)
	}
	var existing model.WQMSStationConfig
	err := r.db.WithContext(ctx).Where("tenant_id = ? AND site_id = ? AND device_id = ?", cfg.TenantID, cfg.SiteID, cfg.DeviceID).First(&existing).Error
	if err == nil {
		cfg.ID = existing.ID
		cfg.CreatedAt = existing.CreatedAt
		return r.db.WithContext(ctx).Save(cfg).Error
	}
	return r.db.WithContext(ctx).Create(cfg).Error
}

// UpsertCalculationResult performs deterministic, idempotent upsert of a WQMS IP calculation and its parameter breakdown.
func (r *WQMSRepository) UpsertCalculationResult(ctx context.Context, res *model.WQMSCalculationResult) error {
	if res.CalculationMode == "" {
		res.CalculationMode = model.CalculationModeRegulatory
	}
	scope := tenantctx.FromContext(ctx)
	if res.TenantID == 0 {
		res.TenantID = scope.WriteTenantID(0)
	}
	if res.SiteID == 0 {
		res.SiteID = scope.WriteSiteID(0)
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing model.WQMSCalculationResult
		err := tx.Where("tenant_id = ? AND site_id = ? AND device_id = ? AND period_start = ? AND period_end = ? AND calculation_mode = ?",
			res.TenantID, res.SiteID, res.DeviceID, res.PeriodStart, res.PeriodEnd, res.CalculationMode).First(&existing).Error

		if err == nil {
			res.ID = existing.ID
			res.CreatedAt = existing.CreatedAt

			if err := tx.Where("calculation_result_id = ?", existing.ID).Delete(&model.WQMSParameterResult{}).Error; err != nil {
				return err
			}

			paramsCopy := make([]model.WQMSParameterResult, len(res.ParameterResults))
			copy(paramsCopy, res.ParameterResults)
			res.ParameterResults = nil

			if err := tx.Save(res).Error; err != nil {
				return err
			}

			for i := range paramsCopy {
				paramsCopy[i].ID = 0
				paramsCopy[i].CalculationResultID = res.ID
				if err := tx.Create(&paramsCopy[i]).Error; err != nil {
					return err
				}
			}
			res.ParameterResults = paramsCopy
			return nil
		} else if err == gorm.ErrRecordNotFound {
			return tx.Create(res).Error
		}
		return err
	})
}

// GetLatestByDeviceID retrieves the latest WQMS IP calculation result for a station/device.
func (r *WQMSRepository) GetLatestByDeviceID(ctx context.Context, deviceID uint, mode string) (*model.WQMSCalculationResult, error) {
	var res model.WQMSCalculationResult
	query := applyWQMSScope(ctx, r.db.WithContext(ctx)).
		Preload("Device").
		Preload("ParameterResults").
		Preload("ParameterResults.Parameter")
	if deviceID > 0 {
		query = query.Where("device_id = ?", deviceID)
	}
	if mode != "" && mode != "ALL" {
		query = query.Where("calculation_mode = ?", mode)
	}
	if err := query.Order("period_end DESC, calculated_at DESC").First(&res).Error; err != nil {
		return nil, err
	}
	return &res, nil
}

// GetByID retrieves a single WQMS calculation result by ID.
func (r *WQMSRepository) GetByID(ctx context.Context, id uint) (*model.WQMSCalculationResult, error) {
	var res model.WQMSCalculationResult
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

// GetTrend retrieves chronological WQMS IP calculation records since a start time.
func (r *WQMSRepository) GetTrend(ctx context.Context, deviceID uint, since time.Time, mode string, limit int) ([]model.WQMSCalculationResult, error) {
	if limit <= 0 || limit > 2000 {
		limit = 720
	}
	var list []model.WQMSCalculationResult
	q := applyWQMSScope(ctx, r.db.WithContext(ctx)).
		Preload("ParameterResults").
		Where("period_end >= ?", since.UTC())
	if deviceID > 0 {
		q = q.Where("device_id = ?", deviceID)
	}
	if mode != "" && mode != "ALL" {
		q = q.Where("calculation_mode = ?", mode)
	}
	err := q.Order("period_end ASC").Limit(limit).Find(&list).Error
	return list, err
}

// GetHistory queries paginated WQMS IP calculation records.
func (r *WQMSRepository) GetHistory(ctx context.Context, filter WQMSHistoryFilter) ([]model.WQMSCalculationResult, int64, error) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 || filter.PageSize > 2000 {
		filter.PageSize = 25
	}

	q := applyWQMSScope(ctx, r.db.WithContext(ctx).Model(&model.WQMSCalculationResult{}))
	if filter.DeviceID > 0 {
		q = q.Where("device_id = ?", filter.DeviceID)
	}
	if filter.CalculationMode != "" && filter.CalculationMode != "ALL" {
		q = q.Where("calculation_mode = ?", filter.CalculationMode)
	}
	if filter.WaterClass != "" && filter.WaterClass != "ALL" {
		q = q.Where("water_class = ?", filter.WaterClass)
	}
	if filter.Category != "" && filter.Category != "ALL" {
		q = q.Where("overall_category = ?", filter.Category)
	}
	if filter.DominantParameter != "" && filter.DominantParameter != "ALL" {
		q = q.Where("dominant_parameter = ?", filter.DominantParameter)
	}
	if filter.Status != "" && filter.Status != "ALL" {
		q = q.Where("calculation_status = ?", filter.Status)
	}
	if filter.StartTime != nil {
		q = q.Where("period_end >= ?", filter.StartTime.UTC())
	}
	if filter.EndTime != nil {
		q = q.Where("period_end <= ?", filter.EndTime.UTC())
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var records []model.WQMSCalculationResult
	offset := (filter.Page - 1) * filter.PageSize
	err := q.Preload("Device").
		Preload("ParameterResults").
		Preload("ParameterResults.Parameter").
		Order("period_end DESC, calculated_at DESC").
		Offset(offset).
		Limit(filter.PageSize).
		Find(&records).Error

	return records, total, err
}

// GetResultsForIKAWindow retrieves recent valid IP results within the lookback window to compute IKA.
func (r *WQMSRepository) GetResultsForIKAWindow(ctx context.Context, deviceID uint, since time.Time, mode model.ISPUCalculationMode) ([]model.WQMSCalculationResult, error) {
	var records []model.WQMSCalculationResult
	q := applyWQMSScope(ctx, r.db.WithContext(ctx)).
		Where("period_end >= ? AND calculation_mode = ? AND overall_category != ?",
			since.UTC(), mode, model.WQMSCategoryUnavailable)
	if deviceID > 0 {
		q = q.Where("device_id = ?", deviceID)
	}
	err := q.Order("period_end ASC").Find(&records).Error
	return records, err
}

// GetMappingsByDeviceID retrieves WQMS parameter mappings (deviceID=0 returns all station mappings).
func (r *WQMSRepository) GetMappingsByDeviceID(ctx context.Context, deviceID uint) ([]model.WQMSParameterMapping, error) {
	var mappings []model.WQMSParameterMapping
	q := applyWQMSScope(ctx, r.db.WithContext(ctx)).Preload("Device").Preload("Parameter")
	if deviceID > 0 {
		q = q.Where("device_id = ? OR source_type = 'LAB'", deviceID)
	}
	err := q.Order("id ASC").Find(&mappings).Error
	return mappings, err
}

// GetMappingByID retrieves a single WQMS parameter mapping by ID.
func (r *WQMSRepository) GetMappingByID(ctx context.Context, id uint) (*model.WQMSParameterMapping, error) {
	var m model.WQMSParameterMapping
	if err := r.db.WithContext(ctx).Preload("Device").Preload("Parameter").First(&m, id).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

// SaveMapping creates or updates a WQMS parameter mapping.
func (r *WQMSRepository) SaveMapping(ctx context.Context, m *model.WQMSParameterMapping) error {
	scope := tenantctx.FromContext(ctx)
	if m.TenantID == 0 {
		m.TenantID = scope.WriteTenantID(0)
	}
	if m.SiteID == 0 {
		m.SiteID = scope.WriteSiteID(0)
	}
	if m.ID > 0 {
		return r.db.WithContext(ctx).Save(m).Error
	}
	return r.db.WithContext(ctx).Create(m).Error
}

// DeleteMapping deletes a WQMS parameter mapping by ID.
func (r *WQMSRepository) DeleteMapping(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.WQMSParameterMapping{}, id).Error
}

// GetAllActiveMappings retrieves all enabled WQMS parameter mappings.
func (r *WQMSRepository) GetAllActiveMappings(ctx context.Context) ([]model.WQMSParameterMapping, error) {
	var mappings []model.WQMSParameterMapping
	err := r.db.WithContext(ctx).
		Preload("Device").
		Preload("Parameter").
		Where("enabled = ?", true).
		Find(&mappings).Error
	return mappings, err
}

