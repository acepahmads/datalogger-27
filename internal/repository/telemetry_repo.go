package repository

import (
	"context"
	"time"

	"datalogger/internal/model"

	"gorm.io/gorm"
)

// TelemetryFilterParams defines criteria for querying historical or raw telemetry
type TelemetryFilterParams struct {
	DeviceID    uint
	ParameterID uint
	Quality     string
	StartTime   *time.Time
	EndTime     *time.Time
	Page        int
	PageSize    int
}

// TelemetryRepository handles telemetry persistence and fast retrieval
type TelemetryRepository struct {
	db *gorm.DB
}

// NewTelemetryRepository creates a new telemetry repository
func NewTelemetryRepository(db *gorm.DB) *TelemetryRepository {
	return &TelemetryRepository{db: db}
}

// SaveBatch persists a batch of raw telemetry records and updates parameter latest values
func (r *TelemetryRepository) SaveBatch(ctx context.Context, batch []*model.RawData) error {
	if len(batch) == 0 {
		return nil
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. Bulk insert raw telemetry records
		if err := tx.Create(&batch).Error; err != nil {
			return err
		}

		// 2. Track latest values per parameter in this batch
		latestPerParam := make(map[uint]*model.RawData)
		var maxReceivedAt time.Time
		var deviceID uint

		for _, item := range batch {
			if item == nil {
				continue
			}
			deviceID = item.DeviceID
			if item.ReceivedAt.After(maxReceivedAt) {
				maxReceivedAt = item.ReceivedAt
			}

			existing, ok := latestPerParam[item.ParameterID]
			if !ok || item.ReceivedAt.After(existing.ReceivedAt) {
				latestPerParam[item.ParameterID] = item
			}
		}

		// 3. Update parameter current/latest values
		for paramID, item := range latestPerParam {
			updates := map[string]interface{}{
				"current_value":            &item.Value,
				"current_value_numeric":    item.ValueNumeric,
				"current_value_text":       item.ValueText,
				"current_value_bool":       item.ValueBool,
				"current_quality":          item.Quality,
				"current_raw_hex":          item.RawHex,
				"current_received_at":      &item.ReceivedAt,
				"current_device_timestamp": item.DeviceTimestamp,
				"last_updated":             &item.ReceivedAt,
			}
			if err := tx.Model(&model.Parameter{}).Where("id = ?", paramID).Updates(updates).Error; err != nil {
				return err
			}
		}

		// 4. Update device last_data_at
		if deviceID > 0 && !maxReceivedAt.IsZero() {
			_ = tx.Model(&model.Device{}).Where("id = ?", deviceID).Update("last_data_at", maxReceivedAt).Error
		}

		return nil
	})
}

// GetLatestForDevice returns all enabled parameters for a device with current values
func (r *TelemetryRepository) GetLatestForDevice(ctx context.Context, deviceID uint) ([]model.Parameter, error) {
	var params []model.Parameter
	err := r.db.WithContext(ctx).
		Where("device_id = ? AND deleted_at IS NULL", deviceID).
		Order("id ASC").
		Find(&params).Error
	return params, err
}

// GetLatestForParameter returns a specific parameter with current value
func (r *TelemetryRepository) GetLatestForParameter(ctx context.Context, deviceID, paramID uint) (*model.Parameter, error) {
	var param model.Parameter
	err := r.db.WithContext(ctx).
		Where("id = ? AND device_id = ? AND deleted_at IS NULL", paramID, deviceID).
		First(&param).Error
	if err != nil {
		return nil, err
	}
	return &param, nil
}

// GetHistorical queries raw data records with filtering, pagination, and sorting
func (r *TelemetryRepository) GetHistorical(ctx context.Context, filter TelemetryFilterParams) ([]model.RawData, int64, error) {
	var records []model.RawData
	var total int64

	query := r.db.WithContext(ctx).Model(&model.RawData{})

	if filter.DeviceID > 0 {
		query = query.Where("device_id = ?", filter.DeviceID)
	}
	if filter.ParameterID > 0 {
		query = query.Where("parameter_id = ?", filter.ParameterID)
	}
	if filter.Quality != "" {
		query = query.Where("quality = ?", filter.Quality)
	}
	if filter.StartTime != nil && !filter.StartTime.IsZero() {
		query = query.Where("received_at >= ?", filter.StartTime)
	}
	if filter.EndTime != nil && !filter.EndTime.IsZero() {
		query = query.Where("received_at <= ?", filter.EndTime)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	page := filter.Page
	if page < 1 {
		page = 1
	}
	pageSize := filter.PageSize
	if pageSize < 1 {
		pageSize = 50
	}
	if pageSize > 1000 {
		pageSize = 1000
	}
	offset := (page - 1) * pageSize

	err := query.Order("received_at DESC, id DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&records).Error

	return records, total, err
}

// DeleteOlderThan is the foundation for future data retention and housekeeping (Section 11)
func (r *TelemetryRepository) DeleteOlderThan(ctx context.Context, cutoff time.Time) (int64, error) {
	res := r.db.WithContext(ctx).Where("received_at < ?", cutoff).Delete(&model.RawData{})
	return res.RowsAffected, res.Error
}
