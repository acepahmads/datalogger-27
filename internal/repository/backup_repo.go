package repository

import (
	"datalogger/internal/model"

	"gorm.io/gorm"
)

type BackupRepository struct {
	db *gorm.DB
}

func NewBackupRepository(db *gorm.DB) *BackupRepository {
	return &BackupRepository{db: db}
}

// Create stores a new backup record in the catalog
func (r *BackupRepository) Create(record *model.BackupRecord) error {
	return r.db.Create(record).Error
}

// Update updates an existing backup record
func (r *BackupRepository) Update(record *model.BackupRecord) error {
	return r.db.Save(record).Error
}

// GetByID finds a backup record by its unique ID
func (r *BackupRepository) GetByID(id string) (*model.BackupRecord, error) {
	var record model.BackupRecord
	err := r.db.Where("id = ?", id).First(&record).Error
	if err != nil {
		return nil, err
	}
	record.PopulateVirtualFields()
	return &record, nil
}

// List returns a paginated list of backup records ordered by creation date descending
func (r *BackupRepository) List(page, pageSize int, status string) ([]model.BackupRecord, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	query := r.db.Model(&model.BackupRecord{})
	if status != "" {
		query = query.Where("status = ?", status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var records []model.BackupRecord
	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&records).Error
	if err == nil {
		for i := range records {
			records[i].PopulateVirtualFields()
		}
	}
	return records, total, err
}

// Delete removes a backup record from the catalog
func (r *BackupRepository) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&model.BackupRecord{}).Error
}

// GetLatestCompleted returns the most recent successfully completed backup
func (r *BackupRepository) GetLatestCompleted() (*model.BackupRecord, error) {
	var record model.BackupRecord
	err := r.db.Where("status = ?", model.BackupStatusCompleted).Order("created_at DESC").First(&record).Error
	if err != nil {
		return nil, err
	}
	record.PopulateVirtualFields()
	return &record, nil
}

// GetTotalStorageBytes calculates total storage used by all completed backups
func (r *BackupRepository) GetTotalStorageBytes() (int64, error) {
	var total int64
	row := r.db.Model(&model.BackupRecord{}).Where("status = ?", model.BackupStatusCompleted).Select("COALESCE(SUM(size_bytes), 0)").Row()
	err := row.Scan(&total)
	return total, err
}

// CountValid returns total count of valid backups
func (r *BackupRepository) CountValid() (int64, error) {
	var count int64
	err := r.db.Model(&model.BackupRecord{}).Where("status = ? AND validation_status = ?", model.BackupStatusCompleted, model.ValidationStatusValid).Count(&count).Error
	return count, err
}

// ListOldestCompleted returns completed backups ordered from oldest to newest for pruning
func (r *BackupRepository) ListOldestCompleted(limit int) ([]model.BackupRecord, error) {
	var records []model.BackupRecord
	err := r.db.Where("status = ?", model.BackupStatusCompleted).Order("created_at ASC").Limit(limit).Find(&records).Error
	if err == nil {
		for i := range records {
			records[i].PopulateVirtualFields()
		}
	}
	return records, err
}
