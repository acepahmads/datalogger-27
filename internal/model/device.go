package model

import (
	"time"

	"gorm.io/gorm"
)

// Administrative status constants (Section 4)
type DeviceAdminStatus string

const (
	DeviceStatusActive      DeviceAdminStatus = "ACTIVE"
	DeviceStatusInactive    DeviceAdminStatus = "INACTIVE"
	DeviceStatusMaintenance DeviceAdminStatus = "MAINTENANCE"
	DeviceStatusDisabled    DeviceAdminStatus = "DISABLED"
)

// Communication status constants (Section 4)
type DeviceConnectionStatus string

const (
	DeviceConnOnline     DeviceConnectionStatus = "ONLINE"
	DeviceConnOffline    DeviceConnectionStatus = "OFFLINE"
	DeviceConnConnecting DeviceConnectionStatus = "CONNECTING"
	DeviceConnError      DeviceConnectionStatus = "ERROR"
	DeviceConnUnknown    DeviceConnectionStatus = "UNKNOWN"
)

// Phase 1 Status Compatibility Aliases
const (
	DeviceOnline  DeviceAdminStatus = "ONLINE"
	DeviceOffline DeviceAdminStatus = "OFFLINE"
	DeviceError   DeviceAdminStatus = "ERROR"
	DeviceStandby DeviceAdminStatus = "STANDBY"
)

type DeviceStatus = DeviceAdminStatus

// Supported Communication Protocols (Section 7)
type ProtocolType string

const (
	ProtocolModbusRTU ProtocolType = "MODBUS_RTU"
	ProtocolModbusTCP ProtocolType = "MODBUS_TCP"
	ProtocolTCP       ProtocolType = "TCP"
	ProtocolUDP       ProtocolType = "UDP"
	ProtocolHTTP      ProtocolType = "HTTP"
	ProtocolMQTT      ProtocolType = "MQTT"
	ProtocolSerial    ProtocolType = "SERIAL"
	ProtocolWebSocket ProtocolType = "WEBSOCKET"
	ProtocolCustom    ProtocolType = "CUSTOM"
)

type DeviceType struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Code        string    `gorm:"size:64;uniqueIndex;not null" json:"code"`
	Name        string    `gorm:"size:128;not null" json:"name"`
	Protocol    string    `gorm:"size:64;not null" json:"protocol"` // MODBUS_RTU, MODBUS_TCP, MQTT, TCP, UDP, REST, SERIAL
	Description string    `gorm:"size:255" json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// DeviceConnection represents connection parameters for a device (Section 7)
type DeviceConnection struct {
	ID              uint         `gorm:"primaryKey" json:"id"`
	DeviceID        uint         `gorm:"uniqueIndex:idx_dev_conn;not null" json:"device_id"`
	Protocol        ProtocolType `gorm:"size:32;default:'MODBUS_TCP';not null" json:"protocol"`
	ConnectionType  string       `gorm:"size:64;default:'ETHERNET'" json:"connection_type"`
	Host            string       `gorm:"size:255" json:"host"`
	Port            int          `gorm:"default:502" json:"port"`
	SerialPort      string       `gorm:"size:64" json:"serial_port"`
	BaudRate        int          `gorm:"default:9600" json:"baud_rate"`
	DataBits        int          `gorm:"default:8" json:"data_bits"`
	Parity          string       `gorm:"size:16;default:'N'" json:"parity"`
	StopBits        int          `gorm:"default:1" json:"stop_bits"`
	Timeout         int          `gorm:"default:1000" json:"timeout"`
	RetryCount      int          `gorm:"default:3" json:"retry_count"`
	PollingInterval int          `gorm:"default:1000" json:"polling_interval"`
	Enabled         bool         `gorm:"default:true" json:"enabled"`
	SlaveID         int          `gorm:"default:1" json:"slave_id"`
	ByteOrder       string       `gorm:"size:32;default:'ABCD'" json:"byte_order"`
	ExtraConfig     string       `gorm:"type:text" json:"extra_config,omitempty"`

	// Compatibility aliases with Phase 1
	Address        string `gorm:"size:128" json:"address,omitempty"`
	TimeoutMs      int    `gorm:"default:1000" json:"timeout_ms,omitempty"`
	PollIntervalMs int    `gorm:"default:1000" json:"poll_interval_ms,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (c *DeviceConnection) BeforeSave(tx *gorm.DB) error {
	if c.SlaveID == 0 {
		c.SlaveID = 1
	}
	if c.ByteOrder == "" {
		c.ByteOrder = "ABCD"
	}
	if c.Host == "" && c.Address != "" {
		c.Host = c.Address
	}
	if c.Address == "" && c.Host != "" {
		c.Address = c.Host
	}
	if c.Timeout == 0 && c.TimeoutMs != 0 {
		c.Timeout = c.TimeoutMs
	}
	if c.TimeoutMs == 0 && c.Timeout != 0 {
		c.TimeoutMs = c.Timeout
	}
	if c.PollingInterval == 0 && c.PollIntervalMs != 0 {
		c.PollingInterval = c.PollIntervalMs
	}
	if c.PollIntervalMs == 0 && c.PollingInterval != 0 {
		c.PollIntervalMs = c.PollingInterval
	}
	if c.Protocol == "" {
		c.Protocol = ProtocolModbusTCP
	}
	return nil
}

func (c *DeviceConnection) AfterFind(tx *gorm.DB) error {
	if c.SlaveID == 0 {
		c.SlaveID = 1
	}
	if c.ByteOrder == "" {
		c.ByteOrder = "ABCD"
	}
	if c.Host == "" && c.Address != "" {
		c.Host = c.Address
	}
	if c.Address == "" && c.Host != "" {
		c.Address = c.Host
	}
	if c.Timeout == 0 && c.TimeoutMs != 0 {
		c.Timeout = c.TimeoutMs
	}
	if c.TimeoutMs == 0 && c.Timeout != 0 {
		c.TimeoutMs = c.Timeout
	}
	if c.PollingInterval == 0 && c.PollIntervalMs != 0 {
		c.PollingInterval = c.PollIntervalMs
	}
	if c.PollIntervalMs == 0 && c.PollingInterval != 0 {
		c.PollIntervalMs = c.PollingInterval
	}
	return nil
}

// Device entity (Section 3)
type Device struct {
	ID               uint                   `gorm:"primaryKey" json:"id"`
	DeviceCode       string                 `gorm:"column:device_code;size:64;uniqueIndex;not null" json:"device_code"`
	DeviceName       string                 `gorm:"column:device_name;size:128;not null" json:"device_name"`
	DeviceType       string                 `gorm:"column:device_type;size:64;not null;default:'MODBUS_TCP'" json:"device_type"`
	Manufacturer     string                 `gorm:"size:128" json:"manufacturer"`
	Model            string                 `gorm:"size:128" json:"model"`
	SerialNumber     string                 `gorm:"size:128;index" json:"serial_number"`
	FirmwareVersion  string                 `gorm:"size:64" json:"firmware_version"`
	Description      string                 `gorm:"size:255" json:"description"`
	Location         string                 `gorm:"size:128" json:"location"`
	Latitude         *float64               `json:"latitude"`
	Longitude        *float64               `json:"longitude"`
	Timezone         string                 `gorm:"size:64;default:'UTC'" json:"timezone"`
	Status           DeviceAdminStatus      `gorm:"size:32;index;default:'ACTIVE'" json:"status"`
	Enabled          bool                   `gorm:"index;default:true" json:"enabled"`
	ConnectionStatus DeviceConnectionStatus `gorm:"size:32;default:'UNKNOWN'" json:"connection_status"`
	LastSeenAt       *time.Time             `json:"last_seen_at"`
	LastDataAt       *time.Time             `json:"last_data_at"`

	// Phase 1 Backward Compatibility Fields
	Code              string      `gorm:"size:64" json:"code,omitempty"`
	Name              string      `gorm:"size:128" json:"name,omitempty"`
	DeviceTypeID      uint        `gorm:"index" json:"device_type_id,omitempty"`
	TypeRel           *DeviceType `gorm:"foreignKey:DeviceTypeID" json:"device_type_rel,omitempty"`
	LastCommunication *time.Time  `json:"last_communication,omitempty"`
	LatencyMs         int         `gorm:"default:0" json:"latency_ms"`
	SuccessCount      int64       `gorm:"default:0" json:"success_count"`
	FailedCount       int64       `gorm:"default:0" json:"failed_count"`

	// Relationships
	Connection *DeviceConnection `gorm:"foreignKey:DeviceID" json:"connection,omitempty"`
	Parameters []Parameter       `gorm:"foreignKey:DeviceID" json:"parameters,omitempty"`

	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

func (d *Device) BeforeSave(tx *gorm.DB) error {
	if d.DeviceCode == "" && d.Code != "" {
		d.DeviceCode = d.Code
	}
	if d.Code == "" && d.DeviceCode != "" {
		d.Code = d.DeviceCode
	}
	if d.DeviceName == "" && d.Name != "" {
		d.DeviceName = d.Name
	}
	if d.Name == "" && d.DeviceName != "" {
		d.Name = d.DeviceName
	}
	if d.LastSeenAt == nil && d.LastCommunication != nil {
		d.LastSeenAt = d.LastCommunication
	}
	if d.LastCommunication == nil && d.LastSeenAt != nil {
		d.LastCommunication = d.LastSeenAt
	}
	if d.Status == "" {
		d.Status = DeviceStatusActive
	}
	if d.ConnectionStatus == "" {
		d.ConnectionStatus = DeviceConnUnknown
	}
	if d.Timezone == "" {
		d.Timezone = "UTC"
	}
	return nil
}

func (d *Device) AfterFind(tx *gorm.DB) error {
	if d.DeviceCode == "" && d.Code != "" {
		d.DeviceCode = d.Code
	}
	if d.Code == "" && d.DeviceCode != "" {
		d.Code = d.DeviceCode
	}
	if d.DeviceName == "" && d.Name != "" {
		d.DeviceName = d.Name
	}
	if d.Name == "" && d.DeviceName != "" {
		d.Name = d.DeviceName
	}
	if d.LastSeenAt == nil && d.LastCommunication != nil {
		d.LastSeenAt = d.LastCommunication
	}
	if d.LastCommunication == nil && d.LastSeenAt != nil {
		d.LastCommunication = d.LastSeenAt
	}
	return nil
}

type ParameterDataType string

const (
	DataTypeFloat32 ParameterDataType = "FLOAT32"
	DataTypeFloat64 ParameterDataType = "FLOAT64"
	DataTypeInt16   ParameterDataType = "INT16"
	DataTypeInt32   ParameterDataType = "INT32"
	DataTypeUInt16  ParameterDataType = "UINT16"
	DataTypeUInt32  ParameterDataType = "UINT32"
	DataTypeBoolean ParameterDataType = "BOOLEAN"
	DataTypeString  ParameterDataType = "STRING"
)

// Parameter represents an engineering telemetry parameter (Section 8)
type Parameter struct {
	ID            uint              `gorm:"primaryKey" json:"id"`
	DeviceID      uint              `gorm:"uniqueIndex:idx_device_param_code;index;not null" json:"device_id"`
	ParameterCode string            `gorm:"column:parameter_code;size:64;uniqueIndex:idx_device_param_code;not null" json:"parameter_code"`
	ParameterName string            `gorm:"column:parameter_name;size:128;not null" json:"parameter_name"`
	DataType      ParameterDataType `gorm:"size:32;default:'FLOAT32'" json:"data_type"`
	Unit          string            `gorm:"size:32" json:"unit"`
	Description   string            `gorm:"size:255" json:"description"`
	MinValue      *float64          `json:"min_value"`
	MaxValue      *float64          `json:"max_value"`
	Precision     int               `gorm:"default:2" json:"precision"`
	Scale         float64           `gorm:"default:1.0" json:"scale"`
	Offset        float64           `gorm:"default:0.0" json:"offset"`
	ByteOrder     string            `gorm:"size:32;default:'ABCD'" json:"byte_order"`
	Enabled       bool              `gorm:"index;default:true" json:"enabled"`

	// Phase 1 Compatibility
	Code            string     `gorm:"size:64" json:"code,omitempty"`
	Name            string     `gorm:"size:128" json:"name,omitempty"`
	RegisterAddress int        `json:"register_address"`
	RegisterType    string     `gorm:"size:32" json:"register_type"` // HOLDING_REGISTER, INPUT_REGISTER, COIL, DISCRETE_INPUT
	ScaleFactor     float64    `gorm:"default:1.0" json:"scale_factor"`
	LowLimit        *float64   `json:"low_limit,omitempty"`
	HighLimit       *float64   `json:"high_limit,omitempty"`
	WarningLow      *float64   `json:"warning_low,omitempty"`
	WarningHigh     *float64   `json:"warning_high,omitempty"`
	CriticalLow     *float64   `json:"critical_low,omitempty"`
	CriticalHigh    *float64   `json:"critical_high,omitempty"`
	CurrentValue    *float64   `json:"current_value,omitempty"`
	LastUpdated     *time.Time `json:"last_updated,omitempty"`

	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

func (p *Parameter) BeforeSave(tx *gorm.DB) error {
	if p.ByteOrder == "" {
		p.ByteOrder = "ABCD"
	}
	if p.ParameterCode == "" && p.Code != "" {
		p.ParameterCode = p.Code
	}
	if p.Code == "" && p.ParameterCode != "" {
		p.Code = p.ParameterCode
	}
	if p.ParameterName == "" && p.Name != "" {
		p.ParameterName = p.Name
	}
	if p.Name == "" && p.ParameterName != "" {
		p.Name = p.ParameterName
	}
	if p.Scale == 0 && p.ScaleFactor != 0 {
		p.Scale = p.ScaleFactor
	}
	if p.ScaleFactor == 0 && p.Scale != 0 {
		p.ScaleFactor = p.Scale
	}
	if p.MinValue == nil && p.LowLimit != nil {
		p.MinValue = p.LowLimit
	}
	if p.LowLimit == nil && p.MinValue != nil {
		p.LowLimit = p.MinValue
	}
	if p.MaxValue == nil && p.HighLimit != nil {
		p.MaxValue = p.HighLimit
	}
	if p.HighLimit == nil && p.MaxValue != nil {
		p.HighLimit = p.MaxValue
	}
	if p.DataType == "" {
		p.DataType = DataTypeFloat32
	}
	return nil
}

func (p *Parameter) AfterFind(tx *gorm.DB) error {
	if p.ByteOrder == "" {
		p.ByteOrder = "ABCD"
	}
	if p.ParameterCode == "" && p.Code != "" {
		p.ParameterCode = p.Code
	}
	if p.Code == "" && p.ParameterCode != "" {
		p.Code = p.ParameterCode
	}
	if p.ParameterName == "" && p.Name != "" {
		p.ParameterName = p.Name
	}
	if p.Name == "" && p.ParameterName != "" {
		p.Name = p.ParameterName
	}
	if p.Scale == 0 && p.ScaleFactor != 0 {
		p.Scale = p.ScaleFactor
	}
	if p.ScaleFactor == 0 && p.Scale != 0 {
		p.ScaleFactor = p.Scale
	}
	if p.MinValue == nil && p.LowLimit != nil {
		p.MinValue = p.LowLimit
	}
	if p.LowLimit == nil && p.MinValue != nil {
		p.LowLimit = p.MinValue
	}
	if p.MaxValue == nil && p.HighLimit != nil {
		p.MaxValue = p.HighLimit
	}
	if p.HighLimit == nil && p.MaxValue != nil {
		p.HighLimit = p.MaxValue
	}
	return nil
}

type Sensor struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	ParameterID uint       `gorm:"index;not null" json:"parameter_id"`
	Model       string     `gorm:"size:128" json:"model"`
	SerialNo    string     `gorm:"size:128" json:"serial_no"`
	Calibration *time.Time `json:"calibration"`
	Accuracy    string     `gorm:"size:64" json:"accuracy"`
	RangeMin    float64    `json:"range_min"`
	RangeMax    float64    `json:"range_max"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

