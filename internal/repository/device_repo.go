package repository

import (
	"strconv"
	"strings"
	"time"

	"datalogger/internal/model"

	"gorm.io/gorm"
)

type DeviceFilterParams struct {
	Search           string
	Status           string
	ConnectionStatus string
	DeviceType       string
	Protocol         string
	Enabled          *bool
	SortBy           string
	SortDir          string
	Page             int
	PageSize         int
}

type DeviceRepository struct {
	db *gorm.DB
}

func NewDeviceRepository(db *gorm.DB) *DeviceRepository {
	return &DeviceRepository{db: db}
}

// List queries devices with search, filtering, pagination, and sorting
func (r *DeviceRepository) List(params DeviceFilterParams) ([]model.Device, int64, error) {
	var devices []model.Device
	var total int64

	query := r.db.Model(&model.Device{}).Where("deleted_at IS NULL")

	// Search filter
	if params.Search != "" {
		searchTerm := "%" + strings.TrimSpace(params.Search) + "%"
		query = query.Where(
			"device_code LIKE ? OR device_name LIKE ? OR code LIKE ? OR name LIKE ? OR manufacturer LIKE ? OR model LIKE ? OR location LIKE ?",
			searchTerm, searchTerm, searchTerm, searchTerm, searchTerm, searchTerm, searchTerm,
		)
	}

	// Status filter (Administrative)
	if params.Status != "" {
		query = query.Where("status = ?", strings.ToUpper(strings.TrimSpace(params.Status)))
	}

	// Connection status filter (Communication)
	if params.ConnectionStatus != "" {
		query = query.Where("connection_status = ?", strings.ToUpper(strings.TrimSpace(params.ConnectionStatus)))
	}

	// Device Type filter
	if params.DeviceType != "" {
		query = query.Where("device_type = ?", params.DeviceType)
	}

	// Enabled filter
	if params.Enabled != nil {
		query = query.Where("enabled = ?", *params.Enabled)
	}

	// Protocol filter
	if params.Protocol != "" {
		query = query.Joins("LEFT JOIN device_connections ON device_connections.device_id = devices.id").
			Where("device_connections.protocol = ?", strings.ToUpper(strings.TrimSpace(params.Protocol)))
	}

	// Count total matching items before pagination
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Sorting
	sortBy := "id"
	switch strings.ToLower(params.SortBy) {
	case "device_code", "code":
		sortBy = "device_code"
	case "device_name", "name":
		sortBy = "device_name"
	case "created_at":
		sortBy = "created_at"
	case "updated_at":
		sortBy = "updated_at"
	case "status":
		sortBy = "status"
	case "connection_status":
		sortBy = "connection_status"
	case "last_seen_at":
		sortBy = "last_seen_at"
	}

	sortDir := "desc"
	if strings.ToLower(params.SortDir) == "asc" {
		sortDir = "asc"
	}
	query = query.Order(sortBy + " " + sortDir)

	// Pagination
	if params.PageSize > 0 {
		page := params.Page
		if page < 1 {
			page = 1
		}
		offset := (page - 1) * params.PageSize
		query = query.Offset(offset).Limit(params.PageSize)
	}

	// Execute query with preloads
	err := query.Preload("Connection").Preload("Parameters").Preload("TypeRel").Find(&devices).Error
	return devices, total, err
}

// GetByID fetches a device by its ID
func (r *DeviceRepository) GetByID(id uint) (*model.Device, error) {
	var dev model.Device
	err := r.db.Where("deleted_at IS NULL").
		Preload("Connection").
		Preload("Parameters", "deleted_at IS NULL").
		Preload("TypeRel").
		First(&dev, id).Error
	if err != nil {
		return nil, err
	}
	return &dev, nil
}

// GetByCode fetches a device by its unique device code
func (r *DeviceRepository) GetByCode(code string) (*model.Device, error) {
	var dev model.Device
	cleanCode := strings.TrimSpace(code)
	err := r.db.Where("deleted_at IS NULL AND (LOWER(device_code) = LOWER(?) OR LOWER(code) = LOWER(?))", cleanCode, cleanCode).
		Preload("Connection").
		Preload("Parameters", "deleted_at IS NULL").
		First(&dev).Error
	if err != nil {
		return nil, err
	}
	return &dev, nil
}

// Create inserts a new device and optional connection
func (r *DeviceRepository) Create(device *model.Device) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		conn := device.Connection
		if err := tx.Omit("Connection", "Parameters").Create(device).Error; err != nil {
			return err
		}

		if conn != nil {
			conn.DeviceID = device.ID
			if err := tx.Create(conn).Error; err != nil {
				return err
			}
			device.Connection = conn
		}
		return nil
	})
}

// Update updates device details and optionally updates/creates connection
func (r *DeviceRepository) Update(device *model.Device) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		conn := device.Connection
		// Update device attributes
		if err := tx.Omit("Connection", "Parameters").Save(device).Error; err != nil {
			return err
		}

		// Update or create connection if supplied
		if conn != nil {
			conn.DeviceID = device.ID
			var existingConn model.DeviceConnection
			if err := tx.Where("device_id = ?", device.ID).First(&existingConn).Error; err == nil {
				conn.ID = existingConn.ID
				if err := tx.Save(conn).Error; err != nil {
					return err
				}
			} else {
				if err := tx.Create(conn).Error; err != nil {
					return err
				}
			}
			device.Connection = conn
		}
		return nil
	})
}

// Delete performs soft delete of device
func (r *DeviceRepository) Delete(id uint) error {
	return r.db.Where("id = ?", id).Delete(&model.Device{}).Error
}

// SetEnabled toggles enabled status
func (r *DeviceRepository) SetEnabled(id uint, enabled bool) error {
	return r.db.Model(&model.Device{}).Where("id = ?", id).Update("enabled", enabled).Error
}

// SetStatus updates administrative status
func (r *DeviceRepository) SetStatus(id uint, status model.DeviceAdminStatus) error {
	return r.db.Model(&model.Device{}).Where("id = ?", id).Update("status", status).Error
}

// UpdateLastSeen updates last seen timestamp
func (r *DeviceRepository) UpdateLastSeen(id uint, t time.Time) error {
	return r.db.Model(&model.Device{}).Where("id = ?", id).Updates(map[string]interface{}{
		"last_seen_at":       &t,
		"last_communication": &t,
	}).Error
}

// UpdateLastData updates last data received timestamp
func (r *DeviceRepository) UpdateLastData(id uint, t time.Time) error {
	return r.db.Model(&model.Device{}).Where("id = ?", id).Update("last_data_at", &t).Error
}

// SetConnectionStatus updates communication status
func (r *DeviceRepository) SetConnectionStatus(id uint, status model.DeviceConnectionStatus) error {
	return r.db.Model(&model.Device{}).Where("id = ?", id).Update("connection_status", status).Error
}

// Parameters
func (r *DeviceRepository) ListParameters(deviceID uint) ([]model.Parameter, error) {
	var params []model.Parameter
	err := r.db.Where("device_id = ? AND deleted_at IS NULL", deviceID).Order("id ASC").Find(&params).Error
	return params, err
}

func (r *DeviceRepository) GetParameterByID(deviceID, paramID uint) (*model.Parameter, error) {
	var param model.Parameter
	err := r.db.Where("id = ? AND device_id = ? AND deleted_at IS NULL", paramID, deviceID).First(&param).Error
	if err != nil {
		return nil, err
	}
	return &param, nil
}

func (r *DeviceRepository) GetParameterByCode(deviceID uint, code string) (*model.Parameter, error) {
	var param model.Parameter
	cleanCode := strings.TrimSpace(code)
	err := r.db.Where("device_id = ? AND deleted_at IS NULL AND (LOWER(parameter_code) = LOWER(?) OR LOWER(code) = LOWER(?))",
		deviceID, cleanCode, cleanCode).First(&param).Error
	if err != nil {
		return nil, err
	}
	return &param, nil
}

func (r *DeviceRepository) CreateParameter(param *model.Parameter) error {
	return r.db.Create(param).Error
}

func (r *DeviceRepository) UpdateParameter(param *model.Parameter) error {
	return r.db.Save(param).Error
}

func (r *DeviceRepository) DeleteParameter(deviceID, paramID uint) error {
	return r.db.Where("id = ? AND device_id = ?", paramID, deviceID).Delete(&model.Parameter{}).Error
}

func (r *DeviceRepository) SetParameterEnabled(deviceID, paramID uint, enabled bool) error {
	return r.db.Model(&model.Parameter{}).Where("id = ? AND device_id = ?", paramID, deviceID).Update("enabled", enabled).Error
}

// GetDeviceAuditTrail fetches audit records related to a specific device
func (r *DeviceRepository) GetDeviceAuditTrail(deviceID uint, deviceCode string, limit int) ([]model.AuditTrail, error) {
	var trails []model.AuditTrail
	idStr := strconv.FormatUint(uint64(deviceID), 10)
	devPattern1 := "%\"device_id\":" + idStr + "%"
	devPattern2 := "%\"device_id\": " + idStr + "%"
	resPattern := "device:" + deviceCode
	codePattern := "%" + deviceCode + "%"

	query := r.db.Where(
		"(resource = ? OR resource = 'DEVICE' OR resource = 'PARAMETER' OR resource = ?) AND (details LIKE ? OR details LIKE ? OR details LIKE ? OR resource = ?)",
		deviceCode, resPattern, devPattern1, devPattern2, codePattern, resPattern,
	).Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	err := query.Find(&trails).Error
	return trails, err
}

// RecordCommunicationResult updates latency and atomic success/failed counters
func (r *DeviceRepository) RecordCommunicationResult(deviceID uint, success bool, latencyMs int) error {
	updates := map[string]interface{}{
		"latency_ms": latencyMs,
	}
	if success {
		return r.db.Model(&model.Device{}).Where("id = ?", deviceID).
			Updates(updates).
			UpdateColumn("success_count", gorm.Expr("success_count + ?", 1)).Error
	}
	return r.db.Model(&model.Device{}).Where("id = ?", deviceID).
		Updates(updates).
		UpdateColumn("failed_count", gorm.Expr("failed_count + ?", 1)).Error
}

// UpdateParameterCurrentValue persists the latest decoded parameter engineering value
func (r *DeviceRepository) UpdateParameterCurrentValue(paramID uint, value float64, formulaVal *float64, isHeld bool, t time.Time) error {
	updates := map[string]interface{}{
		"current_value":   &value,
		"is_current_held": isHeld,
		"last_updated":    &t,
	}
	if formulaVal != nil {
		updates["current_formula_value"] = formulaVal
	}
	return r.db.Model(&model.Parameter{}).Where("id = ?", paramID).Updates(updates).Error
}
