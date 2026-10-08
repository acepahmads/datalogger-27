package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"datalogger/internal/logger"
	"datalogger/internal/model"
	"datalogger/internal/repository"
	"datalogger/pkg/formula"
)

// Request DTOs
type ConnectionConfigDTO struct {
	Protocol        model.ProtocolType `json:"protocol"`
	ConnectionType  string             `json:"connection_type"`
	Host            string             `json:"host"`
	Port            int                `json:"port"`
	SerialPort      string             `json:"serial_port"`
	BaudRate        int                `json:"baud_rate"`
	DataBits        int                `json:"data_bits"`
	Parity          string             `json:"parity"`
	StopBits        int                `json:"stop_bits"`
	Timeout         int                `json:"timeout"`
	RetryCount      int                `json:"retry_count"`
	PollingInterval int                `json:"polling_interval"`
	Enabled         bool               `json:"enabled"`
	SlaveID         int                `json:"slave_id"`
	ByteOrder       string             `json:"byte_order"`
	ExtraConfig     string             `json:"extra_config,omitempty"`
}

type CreateDeviceRequest struct {
	DeviceCode      string               `json:"device_code"`
	DeviceName      string               `json:"device_name"`
	DeviceType      string               `json:"device_type"`
	Manufacturer    string               `json:"manufacturer"`
	Model           string               `json:"model"`
	SerialNumber    string               `json:"serial_number"`
	FirmwareVersion string               `json:"firmware_version"`
	Description     string               `json:"description"`
	Location        string               `json:"location"`
	Latitude        *float64             `json:"latitude"`
	Longitude       *float64             `json:"longitude"`
	Timezone        string               `json:"timezone"`
	Status          model.DeviceAdminStatus `json:"status"`
	Enabled         *bool                `json:"enabled"`
	Connection      *ConnectionConfigDTO `json:"connection,omitempty"`
}

type UpdateDeviceRequest struct {
	DeviceCode      *string              `json:"device_code"`
	DeviceName      *string              `json:"device_name"`
	DeviceType      *string              `json:"device_type"`
	Manufacturer    *string              `json:"manufacturer"`
	Model           *string              `json:"model"`
	SerialNumber    *string              `json:"serial_number"`
	FirmwareVersion *string              `json:"firmware_version"`
	Description     *string              `json:"description"`
	Location        *string              `json:"location"`
	Latitude        *float64             `json:"latitude"`
	Longitude       *float64             `json:"longitude"`
	Timezone        *string              `json:"timezone"`
	Status          *model.DeviceAdminStatus `json:"status"`
	Enabled         *bool                `json:"enabled"`
	Connection      *ConnectionConfigDTO `json:"connection,omitempty"`
}

type CreateParameterRequest struct {
	ParameterCode            string                  `json:"parameter_code"`
	ParameterName            string                  `json:"parameter_name"`
	DataType                 model.ParameterDataType `json:"data_type"`
	Unit                     string                  `json:"unit"`
	Description              string                  `json:"description"`
	MinValue                 *float64                `json:"min_value"`
	MaxValue                 *float64                `json:"max_value"`
	WarningLow               *float64                `json:"warning_low"`
	WarningHigh              *float64                `json:"warning_high"`
	QualityValidationEnabled *bool                   `json:"quality_validation_enabled"`
	ProcessingEnabled        *bool                   `json:"processing_enabled"`
	StaleTimeoutSeconds      *int                    `json:"stale_timeout_seconds"`
	SpikeDetectionEnabled    *bool                   `json:"spike_detection_enabled"`
	SpikeThreshold           *float64                `json:"spike_threshold"`
	SpikeWindowSize          *int                    `json:"spike_window_size"`
	Precision                int                     `json:"precision"`
	Scale                    float64                 `json:"scale"`
	Offset                   float64                 `json:"offset"`
	RegisterAddress          int                     `json:"register_address"`
	RegisterType             string                  `json:"register_type"`
	ByteOrder                string                  `json:"byte_order"`
	Enabled                  *bool                   `json:"enabled"`
	Formula                  string                  `json:"formula"`
	HoldLastValueEnabled     *bool                   `json:"hold_last_value_enabled"`
	HoldLastValueSeconds     int                     `json:"hold_last_value_seconds"`
}

func (r *CreateParameterRequest) UnmarshalJSON(data []byte) error {
	type Alias CreateParameterRequest
	aux := struct {
		RawMin         interface{} `json:"min_value"`
		RawMax         interface{} `json:"max_value"`
		RawWarnLow     interface{} `json:"warning_low"`
		RawWarnHigh    interface{} `json:"warning_high"`
		RawScale       interface{} `json:"scale"`
		RawOffset      interface{} `json:"offset"`
		RawSpikeThresh interface{} `json:"spike_threshold"`
		*Alias
	}{
		Alias: (*Alias)(r),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	var rawMap map[string]interface{}
	_ = json.Unmarshal(data, &rawMap)

	if val, ok := rawMap["min_value"]; ok {
		r.MinValue = parseNullableFloat(val, nil)
	}
	if val, ok := rawMap["max_value"]; ok {
		r.MaxValue = parseNullableFloat(val, nil)
	}
	if val, ok := rawMap["warning_low"]; ok {
		r.WarningLow = parseNullableFloat(val, nil)
	}
	if val, ok := rawMap["warning_high"]; ok {
		r.WarningHigh = parseNullableFloat(val, nil)
	}
	if val, ok := rawMap["scale"]; ok && val != nil {
		if f := parseNullableFloat(val, nil); f != nil {
			r.Scale = *f
		}
	}
	if val, ok := rawMap["offset"]; ok && val != nil {
		if f := parseNullableFloat(val, nil); f != nil {
			r.Offset = *f
		}
	}
	if val, ok := rawMap["spike_threshold"]; ok && val != nil {
		if f := parseNullableFloat(val, nil); f != nil {
			r.SpikeThreshold = f
		}
	}
	return nil
}

type UpdateParameterRequest struct {
	ParameterCode            *string                  `json:"parameter_code"`
	ParameterName            *string                  `json:"parameter_name"`
	DataType                 *model.ParameterDataType `json:"data_type"`
	Unit                     *string                  `json:"unit"`
	Description              *string                  `json:"description"`
	MinValue                 *float64                 `json:"min_value"`
	MaxValue                 *float64                 `json:"max_value"`
	WarningLow               *float64                 `json:"warning_low"`
	WarningHigh              *float64                 `json:"warning_high"`
	ClearMinValue            bool                     `json:"clear_min_value"`
	ClearMaxValue            bool                     `json:"clear_max_value"`
	ClearWarningLow          bool                     `json:"clear_warning_low"`
	ClearWarningHigh         bool                     `json:"clear_warning_high"`
	QualityValidationEnabled *bool                    `json:"quality_validation_enabled"`
	ProcessingEnabled        *bool                    `json:"processing_enabled"`
	StaleTimeoutSeconds      *int                     `json:"stale_timeout_seconds"`
	SpikeDetectionEnabled    *bool                    `json:"spike_detection_enabled"`
	SpikeThreshold           *float64                 `json:"spike_threshold"`
	SpikeWindowSize          *int                     `json:"spike_window_size"`
	Precision                *int                     `json:"precision"`
	Scale                    *float64                 `json:"scale"`
	Offset                   *float64                 `json:"offset"`
	RegisterAddress          *int                     `json:"register_address"`
	RegisterType             *string                  `json:"register_type"`
	ByteOrder                *string                  `json:"byte_order"`
	Enabled                  *bool                    `json:"enabled"`
	Formula                  *string                  `json:"formula"`
	HoldLastValueEnabled     *bool                    `json:"hold_last_value_enabled"`
	HoldLastValueSeconds     *int                     `json:"hold_last_value_seconds"`
}

func (r *UpdateParameterRequest) UnmarshalJSON(data []byte) error {
	type Alias UpdateParameterRequest
	aux := struct {
		RawMin         interface{} `json:"min_value"`
		RawMax         interface{} `json:"max_value"`
		RawWarnLow     interface{} `json:"warning_low"`
		RawWarnHigh    interface{} `json:"warning_high"`
		RawScale       interface{} `json:"scale"`
		RawOffset      interface{} `json:"offset"`
		RawSpikeThresh interface{} `json:"spike_threshold"`
		*Alias
	}{
		Alias: (*Alias)(r),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	var rawMap map[string]interface{}
	_ = json.Unmarshal(data, &rawMap)

	if val, ok := rawMap["min_value"]; ok {
		r.MinValue = parseNullableFloat(val, &r.ClearMinValue)
	}
	if val, ok := rawMap["max_value"]; ok {
		r.MaxValue = parseNullableFloat(val, &r.ClearMaxValue)
	}
	if val, ok := rawMap["warning_low"]; ok {
		r.WarningLow = parseNullableFloat(val, &r.ClearWarningLow)
	}
	if val, ok := rawMap["warning_high"]; ok {
		r.WarningHigh = parseNullableFloat(val, &r.ClearWarningHigh)
	}
	if val, ok := rawMap["scale"]; ok && val != nil {
		if f := parseNullableFloat(val, nil); f != nil {
			r.Scale = f
		}
	}
	if val, ok := rawMap["offset"]; ok && val != nil {
		if f := parseNullableFloat(val, nil); f != nil {
			r.Offset = f
		}
	}
	if val, ok := rawMap["spike_threshold"]; ok && val != nil {
		if f := parseNullableFloat(val, nil); f != nil {
			r.SpikeThreshold = f
		}
	}
	return nil
}

// ParameterQualityDTO encapsulates Phase 3.2 data quality configuration for a parameter
type ParameterQualityDTO struct {
	ParameterID              uint     `json:"parameter_id"`
	ParameterCode            string   `json:"parameter_code"`
	ParameterName            string   `json:"parameter_name"`
	QualityValidationEnabled bool     `json:"quality_validation_enabled"`
	ProcessingEnabled        bool     `json:"processing_enabled"`
	MinValue                 *float64 `json:"min_value"`
	MaxValue                 *float64 `json:"max_value"`
	WarningLow               *float64 `json:"warning_low"`
	WarningHigh              *float64 `json:"warning_high"`
	StaleTimeoutSeconds      int      `json:"stale_timeout_seconds"`
	SpikeDetectionEnabled    bool     `json:"spike_detection_enabled"`
	SpikeThreshold           float64  `json:"spike_threshold"`
	SpikeWindowSize          int      `json:"spike_window_size"`
}

// UpdateParameterQualityRequest allows updating only quality & processing parameters
type UpdateParameterQualityRequest struct {
	QualityValidationEnabled *bool    `json:"quality_validation_enabled"`
	ProcessingEnabled        *bool    `json:"processing_enabled"`
	MinValue                 *float64 `json:"min_value"`
	MaxValue                 *float64 `json:"max_value"`
	WarningLow               *float64 `json:"warning_low"`
	WarningHigh              *float64 `json:"warning_high"`
	ClearMinValue            bool     `json:"clear_min_value"`
	ClearMaxValue            bool     `json:"clear_max_value"`
	ClearWarningLow          bool     `json:"clear_warning_low"`
	ClearWarningHigh         bool     `json:"clear_warning_high"`
	StaleTimeoutSeconds      *int     `json:"stale_timeout_seconds"`
	SpikeDetectionEnabled    *bool    `json:"spike_detection_enabled"`
	SpikeThreshold           *float64 `json:"spike_threshold"`
	SpikeWindowSize          *int     `json:"spike_window_size"`
}

func (r *UpdateParameterQualityRequest) UnmarshalJSON(data []byte) error {
	type Alias UpdateParameterQualityRequest
	aux := struct {
		RawMin         interface{} `json:"min_value"`
		RawMax         interface{} `json:"max_value"`
		RawWarnLow     interface{} `json:"warning_low"`
		RawWarnHigh    interface{} `json:"warning_high"`
		RawSpikeThresh interface{} `json:"spike_threshold"`
		*Alias
	}{
		Alias: (*Alias)(r),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	var rawMap map[string]interface{}
	_ = json.Unmarshal(data, &rawMap)

	if val, ok := rawMap["min_value"]; ok {
		r.MinValue = parseNullableFloat(val, &r.ClearMinValue)
	}
	if val, ok := rawMap["max_value"]; ok {
		r.MaxValue = parseNullableFloat(val, &r.ClearMaxValue)
	}
	if val, ok := rawMap["warning_low"]; ok {
		r.WarningLow = parseNullableFloat(val, &r.ClearWarningLow)
	}
	if val, ok := rawMap["warning_high"]; ok {
		r.WarningHigh = parseNullableFloat(val, &r.ClearWarningHigh)
	}
	if val, ok := rawMap["spike_threshold"]; ok && val != nil {
		if f := parseNullableFloat(val, nil); f != nil {
			r.SpikeThreshold = f
		}
	}
	return nil
}

func parseNullableFloat(val interface{}, clear *bool) *float64 {
	if val == nil {
		if clear != nil {
			*clear = true
		}
		return nil
	}
	switch v := val.(type) {
	case float64:
		return &v
	case float32:
		f := float64(v)
		return &f
	case int:
		f := float64(v)
		return &f
	case int64:
		f := float64(v)
		return &f
	case json.Number:
		if f, err := v.Float64(); err == nil {
			return &f
		}
	case string:
		trimmed := strings.TrimSpace(v)
		if trimmed == "" {
			if clear != nil {
				*clear = true
			}
			return nil
		}
		if parsed, err := strconv.ParseFloat(trimmed, 64); err == nil {
			return &parsed
		}
	}
	return nil
}

type DeviceService struct {
	repo        *repository.DeviceRepository
	systemRepo  *repository.SystemRepository
	commMu      sync.Mutex
	lastCommErr map[uint]time.Time
	prevStatus  map[uint]model.DeviceConnectionStatus
}

func NewDeviceService(repo *repository.DeviceRepository, systemRepo *repository.SystemRepository) *DeviceService {
	return &DeviceService{
		repo:        repo,
		systemRepo:  systemRepo,
		lastCommErr: make(map[uint]time.Time),
		prevStatus:  make(map[uint]model.DeviceConnectionStatus),
	}
}

// ListDevices retrieves devices matching filter parameters
func (s *DeviceService) ListDevices(params repository.DeviceFilterParams) ([]model.Device, int64, error) {
	return s.repo.List(params)
}

// GetDeviceByID retrieves device by ID
func (s *DeviceService) GetDeviceByID(id uint) (*model.Device, error) {
	return s.repo.GetByID(id)
}

// GetDeviceByCode retrieves device by its code
func (s *DeviceService) GetDeviceByCode(code string) (*model.Device, error) {
	return s.repo.GetByCode(code)
}

// CreateDevice registers a new device with validation and audit trail
func (s *DeviceService) CreateDevice(req *CreateDeviceRequest, username, ipAddress, userAgent string) (*model.Device, error) {
	// 1. Validation
	code := strings.TrimSpace(req.DeviceCode)
	if code == "" {
		return nil, errors.New("device_code is required")
	}
	if len(code) > 64 {
		return nil, errors.New("device_code cannot exceed 64 characters")
	}

	name := strings.TrimSpace(req.DeviceName)
	if name == "" {
		return nil, errors.New("device_name is required")
	}
	if len(name) > 128 {
		return nil, errors.New("device_name cannot exceed 128 characters")
	}

	devType := strings.TrimSpace(req.DeviceType)
	if devType == "" {
		devType = "MODBUS_TCP"
	}

	// Check unique device_code
	if existing, _ := s.repo.GetByCode(code); existing != nil {
		return nil, fmt.Errorf("device with code '%s' already exists", code)
	}

	// Validate coordinates if provided
	if req.Latitude != nil {
		if *req.Latitude < -90.0 || *req.Latitude > 90.0 {
			return nil, errors.New("latitude must be between -90 and 90")
		}
	}
	if req.Longitude != nil {
		if *req.Longitude < -180.0 || *req.Longitude > 180.0 {
			return nil, errors.New("longitude must be between -180 and 180")
		}
	}

	// Validate status
	status := req.Status
	if status == "" {
		status = model.DeviceStatusActive
	}
	switch status {
	case model.DeviceStatusActive, model.DeviceStatusInactive, model.DeviceStatusMaintenance, model.DeviceStatusDisabled:
	default:
		return nil, fmt.Errorf("invalid status: %s (must be ACTIVE, INACTIVE, MAINTENANCE, or DISABLED)", status)
	}

	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}

	tz := strings.TrimSpace(req.Timezone)
	if tz == "" {
		tz = "UTC"
	}

	device := &model.Device{
		DeviceCode:       code,
		DeviceName:       name,
		DeviceType:       devType,
		Manufacturer:     strings.TrimSpace(req.Manufacturer),
		Model:            strings.TrimSpace(req.Model),
		SerialNumber:     strings.TrimSpace(req.SerialNumber),
		FirmwareVersion:  strings.TrimSpace(req.FirmwareVersion),
		Description:      strings.TrimSpace(req.Description),
		Location:         strings.TrimSpace(req.Location),
		Latitude:         req.Latitude,
		Longitude:        req.Longitude,
		Timezone:         tz,
		Status:           status,
		Enabled:          enabled,
		ConnectionStatus: model.DeviceConnUnknown,
	}

	// Connection config
	if req.Connection != nil {
		device.Connection = s.buildConnectionModel(req.Connection)
	}

	// Persist
	if err := s.repo.Create(device); err != nil {
		return nil, fmt.Errorf("failed to create device: %w", err)
	}

	// Audit Trail
	auditDetails, _ := json.Marshal(map[string]interface{}{
		"device_id":   device.ID,
		"device_code": device.DeviceCode,
		"device_name": device.DeviceName,
		"status":      device.Status,
		"type":        device.DeviceType,
	})
	s.recordAudit(username, "CREATE_DEVICE", "DEVICE", string(auditDetails), ipAddress, userAgent)

	return device, nil
}

// UpdateDevice updates an existing device
func (s *DeviceService) UpdateDevice(id uint, req *UpdateDeviceRequest, username, ipAddress, userAgent string) (*model.Device, error) {
	device, err := s.repo.GetByID(id)
	if err != nil {
		return nil, errors.New("device not found")
	}

	beforeState, _ := json.Marshal(map[string]interface{}{
		"device_code": device.DeviceCode,
		"device_name": device.DeviceName,
		"status":      device.Status,
		"enabled":     device.Enabled,
		"location":    device.Location,
	})

	if req.DeviceCode != nil {
		cleanCode := strings.TrimSpace(*req.DeviceCode)
		if cleanCode == "" {
			return nil, errors.New("device_code cannot be empty")
		}
		if len(cleanCode) > 64 {
			return nil, errors.New("device_code cannot exceed 64 characters")
		}
		if !strings.EqualFold(cleanCode, device.DeviceCode) {
			if existing, _ := s.repo.GetByCode(cleanCode); existing != nil && existing.ID != device.ID {
				return nil, fmt.Errorf("device code '%s' is already in use by another device", cleanCode)
			}
			device.DeviceCode = cleanCode
			device.Code = cleanCode
		}
	}

	if req.DeviceName != nil {
		cleanName := strings.TrimSpace(*req.DeviceName)
		if cleanName == "" {
			return nil, errors.New("device_name cannot be empty")
		}
		if len(cleanName) > 128 {
			return nil, errors.New("device_name cannot exceed 128 characters")
		}
		device.DeviceName = cleanName
		device.Name = cleanName
	}

	if req.DeviceType != nil {
		device.DeviceType = strings.TrimSpace(*req.DeviceType)
	}
	if req.Manufacturer != nil {
		device.Manufacturer = strings.TrimSpace(*req.Manufacturer)
	}
	if req.Model != nil {
		device.Model = strings.TrimSpace(*req.Model)
	}
	if req.SerialNumber != nil {
		device.SerialNumber = strings.TrimSpace(*req.SerialNumber)
	}
	if req.FirmwareVersion != nil {
		device.FirmwareVersion = strings.TrimSpace(*req.FirmwareVersion)
	}
	if req.Description != nil {
		device.Description = strings.TrimSpace(*req.Description)
	}
	if req.Location != nil {
		device.Location = strings.TrimSpace(*req.Location)
	}
	if req.Latitude != nil {
		if *req.Latitude < -90.0 || *req.Latitude > 90.0 {
			return nil, errors.New("latitude must be between -90 and 90")
		}
		device.Latitude = req.Latitude
	}
	if req.Longitude != nil {
		if *req.Longitude < -180.0 || *req.Longitude > 180.0 {
			return nil, errors.New("longitude must be between -180 and 180")
		}
		device.Longitude = req.Longitude
	}
	if req.Timezone != nil {
		device.Timezone = strings.TrimSpace(*req.Timezone)
	}
	if req.Status != nil {
		status := *req.Status
		switch status {
		case model.DeviceStatusActive, model.DeviceStatusInactive, model.DeviceStatusMaintenance, model.DeviceStatusDisabled:
			device.Status = status
		default:
			return nil, fmt.Errorf("invalid status: %s", status)
		}
	}
	if req.Enabled != nil {
		device.Enabled = *req.Enabled
	}

	if req.Connection != nil {
		device.Connection = s.buildConnectionModel(req.Connection)
	}

	if err := s.repo.Update(device); err != nil {
		return nil, fmt.Errorf("failed to update device: %w", err)
	}

	afterState, _ := json.Marshal(map[string]interface{}{
		"device_code": device.DeviceCode,
		"device_name": device.DeviceName,
		"status":      device.Status,
		"enabled":     device.Enabled,
		"location":    device.Location,
	})

	auditDetails, _ := json.Marshal(map[string]interface{}{
		"device_id": device.ID,
		"before":    string(beforeState),
		"after":     string(afterState),
	})
	s.recordAudit(username, "UPDATE_DEVICE", "DEVICE", string(auditDetails), ipAddress, userAgent)

	return device, nil
}

// DeleteDevice performs soft delete of device
func (s *DeviceService) DeleteDevice(id uint, username, ipAddress, userAgent string) error {
	device, err := s.repo.GetByID(id)
	if err != nil {
		return errors.New("device not found")
	}

	if err := s.repo.Delete(id); err != nil {
		return fmt.Errorf("failed to delete device: %w", err)
	}

	auditDetails, _ := json.Marshal(map[string]interface{}{
		"device_id":   id,
		"device_code": device.DeviceCode,
		"device_name": device.DeviceName,
		"action":      "soft_delete",
	})
	s.recordAudit(username, "DELETE_DEVICE", "DEVICE", string(auditDetails), ipAddress, userAgent)

	return nil
}

// ToggleDeviceEnabled enables or disables a device
func (s *DeviceService) ToggleDeviceEnabled(id uint, enabled bool, username, ipAddress, userAgent string) error {
	device, err := s.repo.GetByID(id)
	if err != nil {
		return errors.New("device not found")
	}

	if err := s.repo.SetEnabled(id, enabled); err != nil {
		return err
	}

	action := "ENABLE_DEVICE"
	if !enabled {
		action = "DISABLE_DEVICE"
	}

	auditDetails, _ := json.Marshal(map[string]interface{}{
		"device_id":   id,
		"device_code": device.DeviceCode,
		"enabled":     enabled,
	})
	s.recordAudit(username, action, "DEVICE", string(auditDetails), ipAddress, userAgent)

	return nil
}

// UpdateDeviceStatus updates device administrative status
func (s *DeviceService) UpdateDeviceStatus(id uint, status model.DeviceAdminStatus, username, ipAddress, userAgent string) error {
	switch status {
	case model.DeviceStatusActive, model.DeviceStatusInactive, model.DeviceStatusMaintenance, model.DeviceStatusDisabled:
	default:
		return fmt.Errorf("invalid status: %s", status)
	}

	device, err := s.repo.GetByID(id)
	if err != nil {
		return errors.New("device not found")
	}

	if err := s.repo.SetStatus(id, status); err != nil {
		return err
	}

	auditDetails, _ := json.Marshal(map[string]interface{}{
		"device_id":   id,
		"device_code": device.DeviceCode,
		"old_status":  device.Status,
		"new_status":  status,
	})
	s.recordAudit(username, "UPDATE_DEVICE", "DEVICE", string(auditDetails), ipAddress, userAgent)

	return nil
}

// UpdateConnection updates or configures device communication foundation
func (s *DeviceService) UpdateConnection(deviceID uint, req *ConnectionConfigDTO, username, ipAddress, userAgent string) (*model.DeviceConnection, error) {
	device, err := s.repo.GetByID(deviceID)
	if err != nil {
		return nil, errors.New("device not found")
	}

	conn := s.buildConnectionModel(req)
	conn.DeviceID = deviceID

	device.Connection = conn
	if err := s.repo.Update(device); err != nil {
		return nil, err
	}

	auditDetails, _ := json.Marshal(map[string]interface{}{
		"device_id":       deviceID,
		"device_code":     device.DeviceCode,
		"protocol":        conn.Protocol,
		"connection_type": conn.ConnectionType,
		"host":            conn.Host,
		"port":            conn.Port,
		"serial_port":     conn.SerialPort,
	})
	s.recordAudit(username, "CHANGE_CONNECTION", "DEVICE", string(auditDetails), ipAddress, userAgent)

	return conn, nil
}

// Health Foundation Methods (Section 9)
func (s *DeviceService) UpdateLastSeen(deviceID uint) error {
	return s.repo.UpdateLastSeen(deviceID, time.Now())
}

func (s *DeviceService) UpdateLastData(deviceID uint) error {
	now := time.Now()
	_ = s.repo.UpdateLastSeen(deviceID, now)
	return s.repo.UpdateLastData(deviceID, now)
}

func (s *DeviceService) SetOnline(deviceID uint) error {
	now := time.Now()
	_ = s.repo.UpdateLastSeen(deviceID, now)

	s.commMu.Lock()
	wasError := (s.prevStatus[deviceID] == model.DeviceConnError)
	s.prevStatus[deviceID] = model.DeviceConnOnline
	s.commMu.Unlock()

	if wasError {
		logger.Info("Device ID %d communication restored: ONLINE", deviceID)
		if s.systemRepo != nil {
			dev, err := s.repo.GetByID(deviceID)
			devCode := fmt.Sprintf("#%d", deviceID)
			if err == nil && dev != nil {
				devCode = dev.DeviceCode
			}
			_ = s.systemRepo.AddSystemLog(&model.SystemLog{
				Level:     "INFO",
				Component: "COMMUNICATION",
				Message:   fmt.Sprintf("Device %s communication restored successfully", devCode),
				Details:   "Auto-reconnect succeeded; telemetry polling active.",
				CreatedAt: now,
			})
			_ = s.systemRepo.AddAuditTrail(&model.AuditTrail{
				Username:  "system",
				Action:    "COMMUNICATION_RESTORED",
				Resource:  fmt.Sprintf("device:%s", devCode),
				Details:   "Device communication link re-established; operational telemetry active.",
				CreatedAt: now,
			})
		}
	}

	return s.repo.SetConnectionStatus(deviceID, model.DeviceConnOnline)
}

func (s *DeviceService) SetOffline(deviceID uint) error {
	s.commMu.Lock()
	s.prevStatus[deviceID] = model.DeviceConnOffline
	s.commMu.Unlock()
	return s.repo.SetConnectionStatus(deviceID, model.DeviceConnOffline)
}

func (s *DeviceService) SetConnectionError(deviceID uint, errMessage string) error {
	now := time.Now()
	s.commMu.Lock()
	prev := s.prevStatus[deviceID]
	lastLogged := s.lastCommErr[deviceID]
	shouldLog := (prev != model.DeviceConnError) || now.Sub(lastLogged) >= 60*time.Second
	if shouldLog {
		s.lastCommErr[deviceID] = now
	}
	s.prevStatus[deviceID] = model.DeviceConnError
	s.commMu.Unlock()

	if shouldLog {
		devCode := fmt.Sprintf("#%d", deviceID)
		protocol := "SERIAL/TCP"
		if dev, err := s.repo.GetByID(deviceID); err == nil && dev != nil {
			devCode = dev.DeviceCode
			if dev.Connection != nil {
				protocol = string(dev.Connection.Protocol)
			}
		}

		logger.Warn("Device %s (%s) communication error: %s (auto-reconnecting in background...)", devCode, protocol, errMessage)

		if s.systemRepo != nil {
			_ = s.systemRepo.AddSystemLog(&model.SystemLog{
				Level:     "WARN",
				Component: "COMMUNICATION",
				Message:   fmt.Sprintf("Device %s communication error: %s", devCode, errMessage),
				Details:   fmt.Sprintf("Protocol: %s. Background engine is continuously attempting auto-reconnect.", protocol),
				CreatedAt: now,
			})
			_ = s.systemRepo.AddAuditTrail(&model.AuditTrail{
				Username:  "system",
				Action:    "COMMUNICATION_ERROR",
				Resource:  fmt.Sprintf("device:%s", devCode),
				Details:   fmt.Sprintf("Communication failed: %s. Automatic background reconnect is active.", errMessage),
				CreatedAt: now,
			})
		}
	}

	return s.repo.SetConnectionStatus(deviceID, model.DeviceConnError)
}

func (s *DeviceService) RecordCommunicationResult(deviceID uint, success bool, latencyMs int) error {
	return s.repo.RecordCommunicationResult(deviceID, success, latencyMs)
}

func (s *DeviceService) UpdateParameterCurrentValue(paramID uint, value float64, formulaVal *float64, isHeld bool) error {
	return s.repo.UpdateParameterCurrentValue(paramID, value, formulaVal, isHeld, time.Now())
}

// Parameters Management (Section 8 & 16)
func (s *DeviceService) ListParameters(deviceID uint) ([]model.Parameter, error) {
	if _, err := s.repo.GetByID(deviceID); err != nil {
		return nil, errors.New("device not found")
	}
	return s.repo.ListParameters(deviceID)
}

func (s *DeviceService) GetParameter(deviceID, paramID uint) (*model.Parameter, error) {
	return s.repo.GetParameterByID(deviceID, paramID)
}

func (s *DeviceService) CreateParameter(deviceID uint, req *CreateParameterRequest, username, ipAddress, userAgent string) (*model.Parameter, error) {
	dev, err := s.repo.GetByID(deviceID)
	if err != nil {
		return nil, errors.New("device not found")
	}

	code := strings.TrimSpace(req.ParameterCode)
	if code == "" {
		return nil, errors.New("parameter_code is required")
	}
	if len(code) > 64 {
		return nil, errors.New("parameter_code cannot exceed 64 characters")
	}

	name := strings.TrimSpace(req.ParameterName)
	if name == "" {
		return nil, errors.New("parameter_name is required")
	}
	if len(name) > 128 {
		return nil, errors.New("parameter_name cannot exceed 128 characters")
	}

	// Check uniqueness per device
	if existing, _ := s.repo.GetParameterByCode(deviceID, code); existing != nil {
		return nil, fmt.Errorf("parameter '%s' already exists for this device", code)
	}

	if req.MinValue != nil && req.MaxValue != nil && *req.MinValue > *req.MaxValue {
		return nil, errors.New("min_value cannot be greater than max_value")
	}

	scale := req.Scale
	if scale == 0 {
		scale = 1.0
	}

	precision := req.Precision
	if precision <= 0 {
		precision = 2
	}

	dataType := req.DataType
	if dataType == "" {
		dataType = model.DataTypeFloat32
	}

	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}

	byteOrder := req.ByteOrder
	if byteOrder == "" {
		byteOrder = "ABCD"
	}

	cleanFormula := strings.TrimSpace(req.Formula)
	if cleanFormula != "" {
		if err := formula.Validate(cleanFormula); err != nil {
			return nil, fmt.Errorf("invalid formula syntax: %w", err)
		}
	}

	holdLastVal := false
	if req.HoldLastValueEnabled != nil {
		holdLastVal = *req.HoldLastValueEnabled
	}

	holdSeconds := req.HoldLastValueSeconds
	if holdSeconds <= 0 {
		holdSeconds = 120
	}

	qualEnabled := true
	if req.QualityValidationEnabled != nil {
		qualEnabled = *req.QualityValidationEnabled
	}
	procEnabled := true
	if req.ProcessingEnabled != nil {
		procEnabled = *req.ProcessingEnabled
	}
	staleTimeout := 120
	if req.StaleTimeoutSeconds != nil && *req.StaleTimeoutSeconds > 0 {
		staleTimeout = *req.StaleTimeoutSeconds
	}
	spikeEnabled := false
	if req.SpikeDetectionEnabled != nil {
		spikeEnabled = *req.SpikeDetectionEnabled
	}
	spikeThresh := 0.0
	if req.SpikeThreshold != nil && *req.SpikeThreshold > 0 {
		spikeThresh = *req.SpikeThreshold
	}
	spikeWindow := 3
	if req.SpikeWindowSize != nil && *req.SpikeWindowSize > 0 {
		spikeWindow = *req.SpikeWindowSize
	}

	param := &model.Parameter{
		DeviceID:                 deviceID,
		ParameterCode:            code,
		ParameterName:            name,
		DataType:                 dataType,
		Unit:                     strings.TrimSpace(req.Unit),
		Description:              strings.TrimSpace(req.Description),
		MinValue:                 req.MinValue,
		MaxValue:                 req.MaxValue,
		WarningLow:               req.WarningLow,
		WarningHigh:              req.WarningHigh,
		QualityValidationEnabled: qualEnabled,
		ProcessingEnabled:        procEnabled,
		StaleTimeoutSeconds:      staleTimeout,
		SpikeDetectionEnabled:    spikeEnabled,
		SpikeThreshold:           spikeThresh,
		SpikeWindowSize:          spikeWindow,
		Precision:                precision,
		Scale:                    scale,
		Offset:                   req.Offset,
		RegisterAddress:          req.RegisterAddress,
		RegisterType:             strings.TrimSpace(req.RegisterType),
		ByteOrder:                byteOrder,
		Enabled:                  enabled,
		Formula:                  cleanFormula,
		HoldLastValueEnabled:     holdLastVal,
		HoldLastValueSeconds:     holdSeconds,
	}

	if err := s.repo.CreateParameter(param); err != nil {
		return nil, fmt.Errorf("failed to create parameter: %w", err)
	}

	auditDetails, _ := json.Marshal(map[string]interface{}{
		"device_id":      deviceID,
		"device_code":    dev.DeviceCode,
		"parameter_id":   param.ID,
		"parameter_code": param.ParameterCode,
		"parameter_name": param.ParameterName,
	})
	s.recordAudit(username, "CREATE_PARAMETER", "PARAMETER", string(auditDetails), ipAddress, userAgent)

	return param, nil
}

func (s *DeviceService) UpdateParameter(deviceID, paramID uint, req *UpdateParameterRequest, username, ipAddress, userAgent string) (*model.Parameter, error) {
	param, err := s.repo.GetParameterByID(deviceID, paramID)
	if err != nil {
		return nil, errors.New("parameter not found")
	}

	beforeState, _ := json.Marshal(map[string]interface{}{
		"parameter_code": param.ParameterCode,
		"parameter_name": param.ParameterName,
		"scale":          param.Scale,
		"offset":         param.Offset,
		"enabled":        param.Enabled,
	})

	if req.ParameterCode != nil {
		cleanCode := strings.TrimSpace(*req.ParameterCode)
		if cleanCode == "" {
			return nil, errors.New("parameter_code cannot be empty")
		}
		if !strings.EqualFold(cleanCode, param.ParameterCode) {
			if existing, _ := s.repo.GetParameterByCode(deviceID, cleanCode); existing != nil && existing.ID != param.ID {
				return nil, fmt.Errorf("parameter code '%s' is already in use for this device", cleanCode)
			}
			param.ParameterCode = cleanCode
			param.Code = cleanCode
		}
	}

	if req.ParameterName != nil {
		cleanName := strings.TrimSpace(*req.ParameterName)
		if cleanName == "" {
			return nil, errors.New("parameter_name cannot be empty")
		}
		param.ParameterName = cleanName
		param.Name = cleanName
	}

	if req.DataType != nil {
		param.DataType = *req.DataType
	}
	if req.Unit != nil {
		param.Unit = strings.TrimSpace(*req.Unit)
	}
	if req.Description != nil {
		param.Description = strings.TrimSpace(*req.Description)
	}
	if req.ClearMinValue {
		param.MinValue = nil
		param.LowLimit = nil
	} else if req.MinValue != nil {
		param.MinValue = req.MinValue
		param.LowLimit = req.MinValue
	}
	if req.ClearMaxValue {
		param.MaxValue = nil
		param.HighLimit = nil
	} else if req.MaxValue != nil {
		param.MaxValue = req.MaxValue
		param.HighLimit = req.MaxValue
	}
	if param.MinValue != nil && param.MaxValue != nil && *param.MinValue > *param.MaxValue {
		return nil, errors.New("min_value cannot be greater than max_value")
	}
	if req.Precision != nil {
		param.Precision = *req.Precision
	}
	if req.Scale != nil {
		if *req.Scale == 0 {
			return nil, errors.New("scale factor cannot be zero")
		}
		param.Scale = *req.Scale
		param.ScaleFactor = *req.Scale
	}
	if req.Offset != nil {
		param.Offset = *req.Offset
	}
	if req.RegisterAddress != nil {
		param.RegisterAddress = *req.RegisterAddress
	}
	if req.RegisterType != nil {
		param.RegisterType = strings.TrimSpace(*req.RegisterType)
	}
	if req.ByteOrder != nil && *req.ByteOrder != "" {
		param.ByteOrder = strings.TrimSpace(*req.ByteOrder)
	}
	if req.Enabled != nil {
		param.Enabled = *req.Enabled
	}
	if req.Formula != nil {
		cleanFormula := strings.TrimSpace(*req.Formula)
		if cleanFormula != "" {
			if err := formula.Validate(cleanFormula); err != nil {
				return nil, fmt.Errorf("invalid formula syntax: %w", err)
			}
			param.Formula = cleanFormula
			if param.CurrentValue != nil {
				if res, err := formula.EvalWithXRaw(cleanFormula, *param.CurrentValue, *param.CurrentValue); err == nil {
					param.CurrentFormulaValue = &res
					param.CurrentValue = &res
				}
			}
		} else {
			param.Formula = ""
			param.CurrentFormulaValue = nil
		}
	}
	if req.HoldLastValueEnabled != nil {
		param.HoldLastValueEnabled = *req.HoldLastValueEnabled
	}
	if req.HoldLastValueSeconds != nil {
		sec := *req.HoldLastValueSeconds
		if sec <= 0 {
			sec = 120
		}
		param.HoldLastValueSeconds = sec
	}
	if req.ClearWarningLow {
		param.WarningLow = nil
	} else if req.WarningLow != nil {
		param.WarningLow = req.WarningLow
	}
	if req.ClearWarningHigh {
		param.WarningHigh = nil
	} else if req.WarningHigh != nil {
		param.WarningHigh = req.WarningHigh
	}
	if param.WarningLow != nil && param.WarningHigh != nil && *param.WarningLow > *param.WarningHigh {
		return nil, errors.New("warning_low cannot be greater than warning_high")
	}
	if req.QualityValidationEnabled != nil {
		param.QualityValidationEnabled = *req.QualityValidationEnabled
	}
	if req.ProcessingEnabled != nil {
		param.ProcessingEnabled = *req.ProcessingEnabled
	}
	if req.StaleTimeoutSeconds != nil {
		sec := *req.StaleTimeoutSeconds
		if sec <= 0 {
			sec = 120
		}
		param.StaleTimeoutSeconds = sec
	}
	if req.SpikeDetectionEnabled != nil {
		param.SpikeDetectionEnabled = *req.SpikeDetectionEnabled
	}
	if req.SpikeThreshold != nil {
		param.SpikeThreshold = *req.SpikeThreshold
	}
	if req.SpikeWindowSize != nil {
		win := *req.SpikeWindowSize
		if win <= 0 {
			win = 3
		}
		param.SpikeWindowSize = win
	}

	if err := s.repo.UpdateParameter(param); err != nil {
		return nil, fmt.Errorf("failed to update parameter: %w", err)
	}

	afterState, _ := json.Marshal(map[string]interface{}{
		"parameter_code":             param.ParameterCode,
		"parameter_name":             param.ParameterName,
		"scale":                      param.Scale,
		"offset":                     param.Offset,
		"enabled":                    param.Enabled,
		"warning_low":                param.WarningLow,
		"warning_high":               param.WarningHigh,
		"quality_validation_enabled": param.QualityValidationEnabled,
		"stale_timeout_seconds":      param.StaleTimeoutSeconds,
	})

	auditDetails, _ := json.Marshal(map[string]interface{}{
		"device_id":    deviceID,
		"parameter_id": param.ID,
		"before":       string(beforeState),
		"after":        string(afterState),
	})
	s.recordAudit(username, "UPDATE_PARAMETER", "PARAMETER", string(auditDetails), ipAddress, userAgent)

	return param, nil
}

// GetParameterQualityConfig retrieves Phase 3.2 quality configuration for a parameter
func (s *DeviceService) GetParameterQualityConfig(deviceID, paramID uint) (*ParameterQualityDTO, error) {
	param, err := s.repo.GetParameterByID(deviceID, paramID)
	if err != nil {
		return nil, errors.New("parameter not found")
	}

	return &ParameterQualityDTO{
		ParameterID:              param.ID,
		ParameterCode:            param.ParameterCode,
		ParameterName:            param.ParameterName,
		QualityValidationEnabled: param.QualityValidationEnabled,
		ProcessingEnabled:        param.ProcessingEnabled,
		MinValue:                 param.MinValue,
		MaxValue:                 param.MaxValue,
		WarningLow:               param.WarningLow,
		WarningHigh:              param.WarningHigh,
		StaleTimeoutSeconds:      param.StaleTimeoutSeconds,
		SpikeDetectionEnabled:    param.SpikeDetectionEnabled,
		SpikeThreshold:           param.SpikeThreshold,
		SpikeWindowSize:          param.SpikeWindowSize,
	}, nil
}

// UpdateParameterQualityConfig updates parameter quality validation settings and records audit trail
func (s *DeviceService) UpdateParameterQualityConfig(deviceID, paramID uint, req *UpdateParameterQualityRequest, username, ipAddress, userAgent string) (*ParameterQualityDTO, error) {
	param, err := s.repo.GetParameterByID(deviceID, paramID)
	if err != nil {
		return nil, errors.New("parameter not found")
	}

	beforeState, _ := json.Marshal(map[string]interface{}{
		"quality_validation_enabled": param.QualityValidationEnabled,
		"processing_enabled":        param.ProcessingEnabled,
		"min_value":                 param.MinValue,
		"max_value":                 param.MaxValue,
		"warning_low":               param.WarningLow,
		"warning_high":              param.WarningHigh,
		"stale_timeout_seconds":      param.StaleTimeoutSeconds,
		"spike_detection_enabled":    param.SpikeDetectionEnabled,
		"spike_threshold":           param.SpikeThreshold,
		"spike_window_size":          param.SpikeWindowSize,
	})

	if req.QualityValidationEnabled != nil {
		param.QualityValidationEnabled = *req.QualityValidationEnabled
	}
	if req.ProcessingEnabled != nil {
		param.ProcessingEnabled = *req.ProcessingEnabled
	}
	if req.ClearMinValue {
		param.MinValue = nil
		param.LowLimit = nil
	} else if req.MinValue != nil {
		param.MinValue = req.MinValue
		param.LowLimit = req.MinValue
	}
	if req.ClearMaxValue {
		param.MaxValue = nil
		param.HighLimit = nil
	} else if req.MaxValue != nil {
		param.MaxValue = req.MaxValue
		param.HighLimit = req.MaxValue
	}
	if param.MinValue != nil && param.MaxValue != nil && *param.MinValue > *param.MaxValue {
		return nil, errors.New("min_value cannot be greater than max_value")
	}
	if req.ClearWarningLow {
		param.WarningLow = nil
	} else if req.WarningLow != nil {
		param.WarningLow = req.WarningLow
	}
	if req.ClearWarningHigh {
		param.WarningHigh = nil
	} else if req.WarningHigh != nil {
		param.WarningHigh = req.WarningHigh
	}
	if param.WarningLow != nil && param.WarningHigh != nil && *param.WarningLow > *param.WarningHigh {
		return nil, errors.New("warning_low cannot be greater than warning_high")
	}
	if req.StaleTimeoutSeconds != nil {
		sec := *req.StaleTimeoutSeconds
		if sec <= 0 {
			sec = 120
		}
		param.StaleTimeoutSeconds = sec
	}
	if req.SpikeDetectionEnabled != nil {
		param.SpikeDetectionEnabled = *req.SpikeDetectionEnabled
	}
	if req.SpikeThreshold != nil {
		param.SpikeThreshold = *req.SpikeThreshold
	}
	if req.SpikeWindowSize != nil {
		win := *req.SpikeWindowSize
		if win <= 0 {
			win = 3
		}
		param.SpikeWindowSize = win
	}

	if err := s.repo.UpdateParameter(param); err != nil {
		return nil, fmt.Errorf("failed to update parameter quality: %w", err)
	}

	afterState, _ := json.Marshal(map[string]interface{}{
		"quality_validation_enabled": param.QualityValidationEnabled,
		"processing_enabled":        param.ProcessingEnabled,
		"min_value":                 param.MinValue,
		"max_value":                 param.MaxValue,
		"warning_low":               param.WarningLow,
		"warning_high":              param.WarningHigh,
		"stale_timeout_seconds":      param.StaleTimeoutSeconds,
		"spike_detection_enabled":    param.SpikeDetectionEnabled,
		"spike_threshold":           param.SpikeThreshold,
		"spike_window_size":          param.SpikeWindowSize,
	})

	auditDetails, _ := json.Marshal(map[string]interface{}{
		"device_id":    deviceID,
		"parameter_id": param.ID,
		"before":       string(beforeState),
		"after":        string(afterState),
	})
	s.recordAudit(username, "UPDATE_PARAMETER_QUALITY", "PARAMETER", string(auditDetails), ipAddress, userAgent)

	return &ParameterQualityDTO{
		ParameterID:              param.ID,
		ParameterCode:            param.ParameterCode,
		ParameterName:            param.ParameterName,
		QualityValidationEnabled: param.QualityValidationEnabled,
		ProcessingEnabled:        param.ProcessingEnabled,
		MinValue:                 param.MinValue,
		MaxValue:                 param.MaxValue,
		WarningLow:               param.WarningLow,
		WarningHigh:              param.WarningHigh,
		StaleTimeoutSeconds:      param.StaleTimeoutSeconds,
		SpikeDetectionEnabled:    param.SpikeDetectionEnabled,
		SpikeThreshold:           param.SpikeThreshold,
		SpikeWindowSize:          param.SpikeWindowSize,
	}, nil
}

func (s *DeviceService) DeleteParameter(deviceID, paramID uint, username, ipAddress, userAgent string) error {
	param, err := s.repo.GetParameterByID(deviceID, paramID)
	if err != nil {
		return errors.New("parameter not found")
	}

	if err := s.repo.DeleteParameter(deviceID, paramID); err != nil {
		return fmt.Errorf("failed to delete parameter: %w", err)
	}

	auditDetails, _ := json.Marshal(map[string]interface{}{
		"device_id":      deviceID,
		"parameter_id":   paramID,
		"parameter_code": param.ParameterCode,
		"parameter_name": param.ParameterName,
	})
	s.recordAudit(username, "DELETE_PARAMETER", "PARAMETER", string(auditDetails), ipAddress, userAgent)

	return nil
}

func (s *DeviceService) ToggleParameterEnabled(deviceID, paramID uint, enabled bool, username, ipAddress, userAgent string) error {
	param, err := s.repo.GetParameterByID(deviceID, paramID)
	if err != nil {
		return errors.New("parameter not found")
	}

	if err := s.repo.SetParameterEnabled(deviceID, paramID, enabled); err != nil {
		return err
	}

	auditDetails, _ := json.Marshal(map[string]interface{}{
		"device_id":      deviceID,
		"parameter_id":   paramID,
		"parameter_code": param.ParameterCode,
		"enabled":        enabled,
	})
	s.recordAudit(username, "UPDATE_PARAMETER", "PARAMETER", string(auditDetails), ipAddress, userAgent)

	return nil
}

// GetDeviceAuditTrail fetches audit trail for a device
func (s *DeviceService) GetDeviceAuditTrail(deviceID uint, limit int) ([]model.AuditTrail, error) {
	dev, err := s.repo.GetByID(deviceID)
	if err != nil {
		return nil, errors.New("device not found")
	}
	return s.repo.GetDeviceAuditTrail(deviceID, dev.DeviceCode, limit)
}

func (s *DeviceService) buildConnectionModel(dto *ConnectionConfigDTO) *model.DeviceConnection {
	proto := dto.Protocol
	if proto == "" {
		proto = model.ProtocolModbusTCP
	}

	timeout := dto.Timeout
	if timeout <= 0 {
		timeout = 1000
	}

	retry := dto.RetryCount
	if retry <= 0 {
		retry = 3
	}

	interval := dto.PollingInterval
	if interval <= 0 {
		interval = 1000
	}

	port := dto.Port
	if port <= 0 {
		if proto == model.ProtocolModbusTCP {
			port = 502
		} else if proto == model.ProtocolMQTT {
			port = 1883
		} else if proto == model.ProtocolHTTP {
			port = 80
		}
	}

	baud := dto.BaudRate
	if baud <= 0 {
		baud = 9600
	}

	dataBits := dto.DataBits
	if dataBits <= 0 {
		dataBits = 8
	}

	stopBits := dto.StopBits
	if stopBits <= 0 {
		stopBits = 1
	}

	parity := dto.Parity
	if parity == "" {
		parity = "N"
	}

	connType := dto.ConnectionType
	if connType == "" {
		if proto == model.ProtocolModbusRTU || proto == model.ProtocolSerial {
			connType = "SERIAL"
		} else {
			connType = "ETHERNET"
		}
	}

	slaveID := dto.SlaveID
	if slaveID <= 0 {
		slaveID = 1
	}

	byteOrder := dto.ByteOrder
	if byteOrder == "" {
		byteOrder = "ABCD"
	}

	return &model.DeviceConnection{
		Protocol:        proto,
		ConnectionType:  connType,
		Host:            strings.TrimSpace(dto.Host),
		Port:            port,
		SerialPort:      strings.TrimSpace(dto.SerialPort),
		BaudRate:        baud,
		DataBits:        dataBits,
		Parity:          parity,
		StopBits:        stopBits,
		Timeout:         timeout,
		RetryCount:      retry,
		PollingInterval: interval,
		Enabled:         dto.Enabled,
		SlaveID:         slaveID,
		ByteOrder:       byteOrder,
		ExtraConfig:     dto.ExtraConfig,
	}
}

func (s *DeviceService) recordAudit(username, action, resource, details, ipAddress, userAgent string) {
	if s.systemRepo != nil {
		if username == "" {
			username = "system"
		}
		_ = s.systemRepo.AddAuditTrail(&model.AuditTrail{
			Username:  username,
			Action:    action,
			Resource:  resource,
			Details:   details,
			IPAddress: ipAddress,
			UserAgent: userAgent,
			CreatedAt: time.Now(),
		})
	}
}
