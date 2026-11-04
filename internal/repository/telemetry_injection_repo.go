package repository

import (
	"context"
	"fmt"
	"sort"
	"time"

	"datalogger/internal/model"

	"gorm.io/gorm"
)

// OverwriteWindowSpec defines an optional [StartTime, EndTime) range to clear prior to OVERWRITE_ALL injection
type OverwriteWindowSpec struct {
	DeviceID     uint
	ParameterIDs []uint
	StartTime    time.Time
	EndTime      time.Time
	InclusiveEnd bool
}

// ExistingRawSummary holds lightweight metadata for conflict detection during injection planning
type ExistingRawSummary struct {
	ID             uint64
	DeviceID       uint
	ParameterID    uint
	ReceivedAt     time.Time
	Quality        model.TelemetryQuality
	RawValue       float64
	ProcessedValue float64
	Value          float64
}

// TelemetryInjectionRepository manages Superadmin-only telemetry injection batches, record links, and rollback
type TelemetryInjectionRepository struct {
	db *gorm.DB
}

// NewTelemetryInjectionRepository creates a new TelemetryInjectionRepository
func NewTelemetryInjectionRepository(db *gorm.DB) *TelemetryInjectionRepository {
	return &TelemetryInjectionRepository{db: db}
}

// GetBatchByID retrieves a single injection batch by ID
func (r *TelemetryInjectionRepository) GetBatchByID(ctx context.Context, id uint) (*model.TelemetryInjectionBatch, error) {
	var batch model.TelemetryInjectionBatch
	if err := r.db.WithContext(ctx).First(&batch, id).Error; err != nil {
		return nil, err
	}
	return &batch, nil
}

// UpdateBatchSummary updates post-calculation counters (aggregated_buckets, refreshed_outbound) on a batch
func (r *TelemetryInjectionRepository) UpdateBatchSummary(ctx context.Context, batchID uint, aggBuckets, refreshedOutbound int) error {
	return r.db.WithContext(ctx).
		Model(&model.TelemetryInjectionBatch{}).
		Where("id = ?", batchID).
		Updates(map[string]interface{}{
			"aggregated_buckets": aggBuckets,
			"refreshed_outbound": refreshedOutbound,
		}).Error
}

// ListBatches retrieves paginated Superadmin injection batches ordered newest first
func (r *TelemetryInjectionRepository) ListBatches(ctx context.Context, deviceID uint, page, pageSize int) ([]model.TelemetryInjectionBatch, int64, error) {
	q := r.db.WithContext(ctx).Model(&model.TelemetryInjectionBatch{})
	if deviceID > 0 {
		q = q.Where("device_id = ?", deviceID)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	offset := (page - 1) * pageSize

	var batches []model.TelemetryInjectionBatch
	if err := q.Order("created_at DESC, id DESC").Limit(pageSize).Offset(offset).Find(&batches).Error; err != nil {
		return nil, 0, err
	}
	return batches, total, nil
}

// GetExistingRawInWindow fetches lightweight existing RawData summaries across [startTime, endTime] for conflict detection
func (r *TelemetryInjectionRepository) GetExistingRawInWindow(
	ctx context.Context,
	deviceID uint,
	paramIDs []uint,
	startTime, endTime time.Time,
) ([]ExistingRawSummary, error) {
	if deviceID == 0 || len(paramIDs) == 0 {
		return nil, nil
	}
	var rows []ExistingRawSummary
	err := r.db.WithContext(ctx).
		Model(&model.RawData{}).
		Select("id, device_id, parameter_id, received_at, quality, raw_value, processed_value, value").
		Where("device_id = ? AND parameter_id IN (?) AND received_at >= ? AND received_at <= ?", deviceID, paramIDs, startTime, endTime).
		Order("received_at ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	return rows, nil
}

// PersistInjectionBatch atomically writes the batch record, optional overwrite cleanup, chunked RawData rows,
// TelemetryInjectionRecord links, and conditional Parameter latest-value updates.
func (r *TelemetryInjectionRepository) PersistInjectionBatch(
	ctx context.Context,
	batch *model.TelemetryInjectionBatch,
	rawRows []*model.RawData,
	overwriteSpec *OverwriteWindowSpec,
) (int64, error) {
	var overwrittenCount int64

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if overwriteSpec != nil && overwriteSpec.DeviceID > 0 && len(overwriteSpec.ParameterIDs) > 0 {
			var existingIDs []uint64
			delQuery := tx.Model(&model.RawData{}).
				Where("device_id = ? AND parameter_id IN (?) AND received_at >= ?",
					overwriteSpec.DeviceID, overwriteSpec.ParameterIDs, overwriteSpec.StartTime)
			if overwriteSpec.InclusiveEnd {
				delQuery = delQuery.Where("received_at <= ?", overwriteSpec.EndTime)
			} else {
				delQuery = delQuery.Where("received_at < ?", overwriteSpec.EndTime)
			}

			if err := delQuery.Pluck("id", &existingIDs).Error; err != nil {
				return fmt.Errorf("failed reading existing raw IDs for overwrite: %w", err)
			}
			if len(existingIDs) > 0 {
				for start := 0; start < len(existingIDs); start += 500 {
					end := start + 500
					if end > len(existingIDs) {
						end = len(existingIDs)
					}
					chunk := existingIDs[start:end]
					res := tx.Where("id IN (?)", chunk).Delete(&model.RawData{})
					if res.Error != nil {
						return fmt.Errorf("failed deleting overwritten raw_data: %w", res.Error)
					}
					overwrittenCount += res.RowsAffected
					_ = tx.Where("raw_data_id IN (?)", chunk).Delete(&model.TelemetryInjectionRecord{}).Error
				}
			}
		}

		batch.OverwrittenCount = int(overwrittenCount)
		batch.InjectedRawCount = len(rawRows)
		if err := tx.Create(batch).Error; err != nil {
			return fmt.Errorf("failed creating injection batch: %w", err)
		}

		if len(rawRows) == 0 {
			return nil
		}

		// Chunked insert of RawData (200 rows per batch to stay well under SQLite/MySQL bind variable limits)
		const rawChunkSize = 200
		for start := 0; start < len(rawRows); start += rawChunkSize {
			end := start + rawChunkSize
			if end > len(rawRows) {
				end = len(rawRows)
			}
			chunk := rawRows[start:end]
			if err := tx.Create(&chunk).Error; err != nil {
				return fmt.Errorf("failed inserting injected raw_data chunk [%d:%d]: %w", start, end, err)
			}
		}

		// Link each inserted RawData primary key in telemetry_injection_records
		links := make([]model.TelemetryInjectionRecord, 0, len(rawRows))
		latestPerParam := make(map[uint]*model.RawData)
		var maxReceivedAt time.Time

		for _, row := range rawRows {
			if row == nil || row.ID == 0 {
				continue
			}
			links = append(links, model.TelemetryInjectionRecord{
				BatchID:     batch.ID,
				BatchCode:   batch.BatchCode,
				RawDataID:   row.ID,
				DeviceID:    row.DeviceID,
				ParameterID: row.ParameterID,
				ReceivedAt:  row.ReceivedAt,
			})
			if prev, ok := latestPerParam[row.ParameterID]; !ok || row.ReceivedAt.After(prev.ReceivedAt) {
				latestPerParam[row.ParameterID] = row
			}
			if row.ReceivedAt.After(maxReceivedAt) {
				maxReceivedAt = row.ReceivedAt
			}
		}

		const linkChunkSize = 500
		for start := 0; start < len(links); start += linkChunkSize {
			end := start + linkChunkSize
			if end > len(links) {
				end = len(links)
			}
			chunk := links[start:end]
			if err := tx.Create(&chunk).Error; err != nil {
				return fmt.Errorf("failed inserting telemetry_injection_records chunk: %w", err)
			}
		}

		// Update Parameter latest values ONLY when parameter has no newer reading
		for paramID, latestRow := range latestPerParam {
			var param model.Parameter
			if err := tx.Select("id, current_received_at").Where("id = ?", paramID).First(&param).Error; err != nil {
				continue
			}
			if param.CurrentReceivedAt != nil && !latestRow.ReceivedAt.After(*param.CurrentReceivedAt) {
				continue
			}
			valNum := latestRow.Value
			updates := map[string]interface{}{
				"current_value":            &latestRow.Value,
				"current_value_numeric":    &valNum,
				"current_processed_value":  &latestRow.ProcessedValue,
				"current_formula_value":    latestRow.FormulaValue,
				"is_current_held":          false,
				"current_quality":          latestRow.Quality,
				"current_quality_reason":   latestRow.QualityReason,
				"current_quality_flags":    "",
				"current_raw_hex":          latestRow.RawHex,
				"last_updated":             &latestRow.ReceivedAt,
				"current_received_at":      &latestRow.ReceivedAt,
				"current_device_timestamp": latestRow.DeviceTimestamp,
			}
			_ = tx.Model(&model.Parameter{}).Where("id = ?", paramID).Updates(updates).Error
		}

		if batch.DeviceID > 0 && !maxReceivedAt.IsZero() {
			var dev model.Device
			if err := tx.Select("id, last_data_at").Where("id = ?", batch.DeviceID).First(&dev).Error; err == nil {
				if dev.LastDataAt == nil || maxReceivedAt.After(*dev.LastDataAt) {
					_ = tx.Model(&model.Device{}).Where("id = ?", batch.DeviceID).Update("last_data_at", maxReceivedAt).Error
				}
			}
		}

		return nil
	})

	return overwrittenCount, err
}

// RollbackBatch deletes all RawData rows created by a batch, removes link rows, marks batch ROLLED_BACK,
// and refreshes Parameter current state if its latest reading belonged to the rolled-back window.
func (r *TelemetryInjectionRepository) RollbackBatch(
	ctx context.Context,
	batchID uint,
	rolledBackBy string,
) (*model.TelemetryInjectionBatch, []uint, int64, error) {
	batch, err := r.GetBatchByID(ctx, batchID)
	if err != nil {
		return nil, nil, 0, err
	}
	if batch.Status == "ROLLED_BACK" {
		return nil, nil, 0, fmt.Errorf("batch %s sudah di-rollback sebelumnya", batch.BatchCode)
	}

	var records []model.TelemetryInjectionRecord
	if err := r.db.WithContext(ctx).Where("batch_id = ?", batchID).Find(&records).Error; err != nil {
		return nil, nil, 0, err
	}

	rawIDs := make([]uint64, 0, len(records))
	paramSet := make(map[uint]bool)
	for _, rec := range records {
		if rec.RawDataID > 0 {
			rawIDs = append(rawIDs, rec.RawDataID)
		}
		if rec.ParameterID > 0 {
			paramSet[rec.ParameterID] = true
		}
	}

	paramIDs := make([]uint, 0, len(paramSet))
	for pid := range paramSet {
		paramIDs = append(paramIDs, pid)
	}
	sort.Slice(paramIDs, func(i, j int) bool { return paramIDs[i] < paramIDs[j] })

	var deletedRaw int64
	now := time.Now().UTC()

	err = r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for start := 0; start < len(rawIDs); start += 500 {
			end := start + 500
			if end > len(rawIDs) {
				end = len(rawIDs)
			}
			chunk := rawIDs[start:end]
			res := tx.Where("id IN (?)", chunk).Delete(&model.RawData{})
			if res.Error != nil {
				return res.Error
			}
			deletedRaw += res.RowsAffected
		}

		if err := tx.Where("batch_id = ?", batchID).Delete(&model.TelemetryInjectionRecord{}).Error; err != nil {
			return err
		}

		if err := tx.Model(&model.TelemetryInjectionBatch{}).
			Where("id = ?", batchID).
			Updates(map[string]interface{}{
				"status":         "ROLLED_BACK",
				"rolled_back_at": &now,
				"rolled_back_by": rolledBackBy,
			}).Error; err != nil {
			return err
		}

		// Reconcile Parameter current state from newest remaining RawData if affected
		for _, pid := range paramIDs {
			var latest model.RawData
			qErr := tx.Where("device_id = ? AND parameter_id = ? AND quality <> ?", batch.DeviceID, pid, model.QualityBad).
				Order("received_at DESC").
				First(&latest).Error
			if qErr == nil {
				valNum := latest.Value
				_ = tx.Model(&model.Parameter{}).Where("id = ?", pid).Updates(map[string]interface{}{
					"current_value":           &latest.Value,
					"current_value_numeric":   &valNum,
					"current_processed_value": &latest.ProcessedValue,
					"current_formula_value":   latest.FormulaValue,
					"current_quality":         latest.Quality,
					"current_quality_reason":  latest.QualityReason,
					"current_raw_hex":         latest.RawHex,
					"last_updated":            &latest.ReceivedAt,
					"current_received_at":     &latest.ReceivedAt,
				}).Error
			} else if qErr == gorm.ErrRecordNotFound {
				_ = tx.Model(&model.Parameter{}).Where("id = ?", pid).Updates(map[string]interface{}{
					"current_value":           nil,
					"current_value_numeric":   nil,
					"current_processed_value": nil,
					"current_formula_value":   nil,
					"current_quality":         model.QualityUnknown,
					"current_quality_reason":  model.ReasonNone,
					"current_raw_hex":         "",
					"current_received_at":     nil,
				}).Error
			}
		}

		return nil
	})
	if err != nil {
		return nil, nil, 0, err
	}

	batch.Status = "ROLLED_BACK"
	batch.RolledBackAt = &now
	batch.RolledBackBy = rolledBackBy
	return batch, paramIDs, deletedRaw, nil
}

// FindInjectedRawMap returns a map of RawDataID -> BatchCode for Superadmin badge annotation
func (r *TelemetryInjectionRepository) FindInjectedRawMap(ctx context.Context, rawIDs []uint64) (map[uint64]string, error) {
	result := make(map[uint64]string)
	if len(rawIDs) == 0 {
		return result, nil
	}

	type rowPair struct {
		RawDataID uint64
		BatchCode string
	}

	for start := 0; start < len(rawIDs); start += 500 {
		end := start + 500
		if end > len(rawIDs) {
			end = len(rawIDs)
		}
		chunk := rawIDs[start:end]
		var pairs []rowPair
		if err := r.db.WithContext(ctx).
			Model(&model.TelemetryInjectionRecord{}).
			Select("raw_data_id, batch_code").
			Where("raw_data_id IN (?)", chunk).
			Scan(&pairs).Error; err != nil {
			return result, err
		}
		for _, p := range pairs {
			result[p.RawDataID] = p.BatchCode
		}
	}
	return result, nil
}

// FindInjectedTimestampsByDeviceParam queries injected sample timestamps across [minStart, maxEnd)
// for fast in-memory binary-search annotation of AggregationResult buckets (Superadmin only).
func (r *TelemetryInjectionRepository) FindInjectedTimestampsByDeviceParam(
	ctx context.Context,
	deviceIDs []uint,
	paramIDs []uint,
	minStart, maxEnd time.Time,
) (map[[2]uint][]time.Time, error) {
	grouped := make(map[[2]uint][]time.Time)
	if len(deviceIDs) == 0 || len(paramIDs) == 0 || !maxEnd.After(minStart) {
		return grouped, nil
	}

	var recs []model.TelemetryInjectionRecord
	err := r.db.WithContext(ctx).
		Select("device_id, parameter_id, received_at").
		Where("device_id IN (?) AND parameter_id IN (?) AND received_at >= ? AND received_at < ?",
			deviceIDs, paramIDs, minStart, maxEnd).
		Order("received_at ASC").
		Find(&recs).Error
	if err != nil {
		return grouped, err
	}

	for _, rec := range recs {
		key := [2]uint{rec.DeviceID, rec.ParameterID}
		grouped[key] = append(grouped[key], rec.ReceivedAt)
	}
	return grouped, nil
}

