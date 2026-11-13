package repository

import (
	"strconv"
	"strings"
	"time"

	"datalogger/internal/model"

	"gorm.io/gorm"
)

type DeviceFilterParams struct {
	TenantID         uint
	SiteID           uint
	AllowedSiteIDs   []uint
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

	query := r.db.Model(&model.Device{}).Where("devices.deleted_at IS NULL")

	if params.TenantID > 0 {
		query = query.Where("devices.tenant_id = ?", params.TenantID)
	}
	if len(params.AllowedSiteIDs) > 0 {
		query = query.Where("devices.site_id IN ?", params.AllowedSiteIDs)
	}
	if params.SiteID > 0 {
		query = query.Where("devices.site_id = ?", params.SiteID)
	}

	// Search filter
	if params.Search != "" {
		searchTerm := "%" + strings.TrimSpace(params.Search) + "%"
		query = query.Where(
			"devices.device_code LIKE ? OR devices.device_name LIKE ? OR devices.code LIKE ? OR devices.name LIKE ? OR devices.manufacturer LIKE ? OR devices.model LIKE ? OR devices.location LIKE ?",
			searchTerm, searchTerm, searchTerm, searchTerm, searchTerm, searchTerm, searchTerm,
		)
	}

	// Status filter (Administrative)
	if params.Status != "" {
		query = query.Where("devices.status = ?", strings.ToUpper(strings.TrimSpace(params.Status)))
	}

	// Connection status filter (Communication)
	if params.ConnectionStatus != "" {
		query = query.Where("devices.connection_status = ?", strings.ToUpper(strings.TrimSpace(params.ConnectionStatus)))
	}

	// Device Type filter
	if params.DeviceType != "" {
		query = query.Where("devices.device_type = ?", params.DeviceType)
	}

	// Enabled filter
	if params.Enabled != nil {
		query = query.Where("devices.enabled = ?", *params.Enabled)
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
	sortBy := "devices.id"
	switch strings.ToLower(params.SortBy) {
	case "device_code", "code":
		sortBy = "devices.device_code"
	case "device_name", "name":
		sortBy = "devices.device_name"
	case "created_at":
		sortBy = "devices.created_at"
	case "updated_at":
		sortBy = "devices.updated_at"
	case "status":
		sortBy = "devices.status"
	case "connection_status":
		sortBy = "devices.connection_status"
	case "last_seen_at":
		sortBy = "devices.last_seen_at"
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
	err := query.Preload("Tenant").Preload("Site").Preload("Connection").Preload("Parameters", "deleted_at IS NULL").Preload("Parameters.Category").Preload("TypeRel").Find(&devices).Error
	return devices, total, err
}

// GetByID fetches a device by its ID
func (r *DeviceRepository) GetByID(id uint) (*model.Device, error) {
	var dev model.Device
	err := r.db.Where("deleted_at IS NULL").
		Preload("Tenant").
		Preload("Site").
		Preload("Connection").
		Preload("Parameters", "deleted_at IS NULL").
		Preload("Parameters.Category").
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
		Preload("Tenant").
		Preload("Site").
		Preload("Connection").
		Preload("Parameters", "deleted_at IS NULL").
		Preload("Parameters.Category").
		First(&dev).Error
	if err != nil {
		return nil, err
	}
	return &dev, nil
}

// GetByCodeInTenant fetches a device by its unique device code within a specific tenant
func (r *DeviceRepository) GetByCodeInTenant(tenantID uint, code string) (*model.Device, error) {
	if tenantID == 0 {
		return r.GetByCode(code)
	}
	var dev model.Device
	cleanCode := strings.TrimSpace(code)
	err := r.db.Where("tenant_id = ? AND deleted_at IS NULL AND (LOWER(device_code) = LOWER(?) OR LOWER(code) = LOWER(?))", tenantID, cleanCode, cleanCode).
		Preload("Tenant").
		Preload("Site").
		Preload("Connection").
		Preload("Parameters", "deleted_at IS NULL").
		Preload("Parameters.Category").
		First(&dev).Error
	if err != nil {
		return nil, err
	}
	return &dev, nil
}

// GetByLegacyModuleIDInSite fetches an active device by its legacy Cloud module_id within a specific site.
func (r *DeviceRepository) GetByLegacyModuleIDInSite(tenantID, siteID uint, legacyModuleID string) (*model.Device, error) {
	clean := strings.TrimSpace(legacyModuleID)
	if clean == "" {
		return nil, gorm.ErrRecordNotFound
	}
	var dev model.Device
	err := r.db.Where("tenant_id = ? AND site_id = ? AND deleted_at IS NULL AND enabled = ? AND LOWER(legacy_module_id) = LOWER(?)", tenantID, siteID, true, clean).
		Preload("Tenant").
		Preload("Site").
		Preload("Connection").
		Preload("Parameters", "deleted_at IS NULL").
		Preload("Parameters.Category").
		Order("id ASC").
		First(&dev).Error
	if err != nil {
		return nil, err
	}
	return &dev, nil
}

// GetByLegacyModuleIDInTenant fetches an active device by its legacy Cloud module_id within a tenant.
func (r *DeviceRepository) GetByLegacyModuleIDInTenant(tenantID uint, legacyModuleID string) (*model.Device, error) {
	clean := strings.TrimSpace(legacyModuleID)
	if clean == "" {
		return nil, gorm.ErrRecordNotFound
	}
	var dev model.Device
	err := r.db.Where("tenant_id = ? AND deleted_at IS NULL AND enabled = ? AND LOWER(legacy_module_id) = LOWER(?)", tenantID, true, clean).
		Preload("Tenant").
		Preload("Site").
		Preload("Connection").
		Preload("Parameters", "deleted_at IS NULL").
		Preload("Parameters.Category").
		Order("id ASC").
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
		if err := tx.Omit("Tenant", "Site", "Connection", "Parameters").Create(device).Error; err != nil {
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
		// Update device attributes (never overwrite immutable created_at with zero value on MariaDB)
		if err := tx.Omit("Tenant", "Site", "Connection", "Parameters", "created_at").Save(device).Error; err != nil {
			return err
		}

		// Update or create connection if supplied
		if conn != nil {
			conn.DeviceID = device.ID
			var existingConn model.DeviceConnection
			if err := tx.Where("device_id = ?", device.ID).First(&existingConn).Error; err == nil {
				conn.ID = existingConn.ID
				conn.CreatedAt = existingConn.CreatedAt
				if err := tx.Omit("created_at").Save(conn).Error; err != nil {
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

// Delete performs soft delete of device and its associated parameters and aggregation definitions
func (r *DeviceRepository) Delete(id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if tx.Migrator().HasTable(&model.AggregationDefinition{}) {
			_ = tx.Model(&model.AggregationDefinition{}).Where("device_id = ?", id).Update("enabled", false).Error
			_ = tx.Where("device_id = ?", id).Delete(&model.AggregationDefinition{}).Error
		}
		if tx.Migrator().HasTable(&model.Parameter{}) {
			_ = tx.Where("device_id = ?", id).Delete(&model.Parameter{}).Error
		}
		return tx.Where("id = ?", id).Delete(&model.Device{}).Error
	})
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
	err := r.db.Preload("Category").Where("device_id = ? AND deleted_at IS NULL", deviceID).Order("id ASC").Find(&params).Error
	return params, err
}

func (r *DeviceRepository) GetParameterByID(deviceID, paramID uint) (*model.Parameter, error) {
	var param model.Parameter
	err := r.db.Preload("Category").Where("id = ? AND device_id = ? AND deleted_at IS NULL", paramID, deviceID).First(&param).Error
	if err != nil {
		return nil, err
	}
	return &param, nil
}

func (r *DeviceRepository) FindParameterByID(paramID uint) (*model.Parameter, error) {
	var param model.Parameter
	err := r.db.Preload("Category").Where("id = ? AND deleted_at IS NULL", paramID).First(&param).Error
	if err != nil {
		return nil, err
	}
	return &param, nil
}

func (r *DeviceRepository) GetParameterByCode(deviceID uint, code string) (*model.Parameter, error) {
	var param model.Parameter
	cleanCode := strings.TrimSpace(code)
	err := r.db.Preload("Category").Where("device_id = ? AND deleted_at IS NULL AND (LOWER(parameter_code) = LOWER(?) OR LOWER(code) = LOWER(?))",
		deviceID, cleanCode, cleanCode).First(&param).Error
	if err != nil {
		return nil, err
	}
	return &param, nil
}

func (r *DeviceRepository) CreateParameter(param *model.Parameter) error {
	if err := r.db.Create(param).Error; err != nil {
		return err
	}
	if param.CategoryID != nil && *param.CategoryID > 0 {
		_ = r.db.Preload("Category").First(param, param.ID).Error
	}
	return nil
}

func (r *DeviceRepository) UpdateParameter(param *model.Parameter) error {
	if err := r.db.Omit("created_at").Save(param).Error; err != nil {
		return err
	}
	if param.CategoryID != nil && *param.CategoryID > 0 {
		_ = r.db.Preload("Category").First(param, param.ID).Error
	}
	return nil
}

func (r *DeviceRepository) DeleteParameter(deviceID, paramID uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if tx.Migrator().HasTable(&model.AggregationDefinition{}) {
			_ = tx.Model(&model.AggregationDefinition{}).Where("device_id = ? AND parameter_id = ?", deviceID, paramID).Update("enabled", false).Error
			_ = tx.Where("device_id = ? AND parameter_id = ?", deviceID, paramID).Delete(&model.AggregationDefinition{}).Error
		}
		return tx.Where("id = ? AND device_id = ?", paramID, deviceID).Delete(&model.Parameter{}).Error
	})
}

func (r *DeviceRepository) SetParameterEnabled(deviceID, paramID uint, enabled bool) error {
	return r.db.Model(&model.Parameter{}).Where("id = ? AND device_id = ?", paramID, deviceID).Update("enabled", enabled).Error
}

func (r *DeviceRepository) SetParametersEnabledBulk(deviceID uint, paramIDs []uint, enabled bool) error {
	if deviceID == 0 || len(paramIDs) == 0 {
		return nil
	}
	return r.db.Model(&model.Parameter{}).
		Where("device_id = ? AND id IN ? AND deleted_at IS NULL", deviceID, paramIDs).
		Update("enabled", enabled).Error
}

func (r *DeviceRepository) SyncDeviceParametersStaleTimeout(deviceID uint, staleTimeoutSeconds int) error {
	if deviceID == 0 || staleTimeoutSeconds <= 0 {
		return nil
	}
	return r.db.Model(&model.Parameter{}).
		Where("device_id = ? AND deleted_at IS NULL", deviceID).
		Update("stale_timeout_seconds", staleTimeoutSeconds).Error
}

// GetDeviceAuditTrail fetches audit records related to a specific device
func (r *DeviceRepository) GetDeviceAuditTrail(deviceID uint, deviceCode string, limit int, tenantID ...uint) ([]model.AuditTrail, error) {
	var trails []model.AuditTrail
	idStr := strconv.FormatUint(uint64(deviceID), 10)
	devPattern1 := "%\"device_id\":" + idStr + "%"
	devPattern2 := "%\"device_id\": " + idStr + "%"
	resPattern := "device:" + deviceCode
	codePattern := "%" + deviceCode + "%"

	query := r.db.Where(
		"(resource = ? OR resource = 'DEVICE' OR resource = 'PARAMETER' OR resource = ?) AND (details LIKE ? OR details LIKE ? OR details LIKE ? OR resource = ?)",
		deviceCode, resPattern, devPattern1, devPattern2, codePattern, resPattern,
	)
	if len(tenantID) > 0 && tenantID[0] > 0 {
		query = query.Where("(tenant_id IS NULL OR tenant_id = ?)", tenantID[0])
	}
	query = query.Order("created_at DESC")

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
	return r.UpdateParameterCurrentReadings(paramID, nil, value, formulaVal, isHeld, t)
}

// UpdateParameterCurrentReadings persists both pre-calibration sensor value (x) and final calibrated/processed value
func (r *DeviceRepository) UpdateParameterCurrentReadings(paramID uint, uncalibratedVal *float64, value float64, formulaVal *float64, isHeld bool, t time.Time) error {
	updates := map[string]interface{}{
		"current_value":   &value,
		"is_current_held": isHeld,
		"last_updated":    &t,
	}
	if uncalibratedVal != nil {
		updates["current_uncalibrated_value"] = uncalibratedVal
	}
	if formulaVal != nil {
		updates["current_formula_value"] = formulaVal
	}
	return r.db.Model(&model.Parameter{}).Where("id = ?", paramID).Updates(updates).Error
}
