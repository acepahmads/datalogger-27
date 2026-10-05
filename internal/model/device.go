package model

import (
	"time"
)

type DeviceStatus string

const (
	DeviceOnline  DeviceStatus = "ONLINE"
	DeviceOffline DeviceStatus = "OFFLINE"
	DeviceError   DeviceStatus = "ERROR"
	DeviceStandby DeviceStatus = "STANDBY"
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

type DeviceConnection struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	DeviceID       uint      `gorm:"index;not null" json:"device_id"`
	ConnectionType string    `gorm:"size:64;not null" json:"connection_type"` // SERIAL, TCP, UDP, MQTT, HTTP
	Address        string    `gorm:"size:128" json:"address"`                 // IP or COM port /dev/ttyUSB0
	Port           int       `json:"port"`
	BaudRate       int       `json:"baud_rate"`
	DataBits       int       `json:"data_bits"`
	StopBits       int       `json:"stop_bits"`
	Parity         string    `gorm:"size:16" json:"parity"`
	TimeoutMs      int       `gorm:"default:1000" json:"timeout_ms"`
	PollIntervalMs int       `gorm:"default:1000" json:"poll_interval_ms"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type Device struct {
	ID                uint              `gorm:"primaryKey" json:"id"`
	Code              string            `gorm:"size:64;uniqueIndex;not null" json:"code"`
	Name              string            `gorm:"size:128;not null" json:"name"`
	DeviceTypeID      uint              `gorm:"index" json:"device_type_id"`
	DeviceType        *DeviceType       `gorm:"foreignKey:DeviceTypeID" json:"device_type,omitempty"`
	Status            DeviceStatus      `gorm:"size:32;default:'STANDBY'" json:"status"`
	Enabled           bool              `gorm:"default:true" json:"enabled"`
	Location          string            `gorm:"size:128" json:"location"`
	LastCommunication *time.Time        `json:"last_communication"`
	LatencyMs         int               `gorm:"default:0" json:"latency_ms"`
	SuccessCount      int64             `gorm:"default:0" json:"success_count"`
	FailedCount       int64             `gorm:"default:0" json:"failed_count"`
	Connection        *DeviceConnection `gorm:"foreignKey:DeviceID" json:"connection,omitempty"`
	Parameters        []Parameter       `gorm:"foreignKey:DeviceID" json:"parameters,omitempty"`
	CreatedAt         time.Time         `json:"created_at"`
	UpdatedAt         time.Time         `json:"updated_at"`
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

type Parameter struct {
	ID             uint              `gorm:"primaryKey" json:"id"`
	DeviceID       uint              `gorm:"index;not null" json:"device_id"`
	Code           string            `gorm:"size:64;not null" json:"code"`
	Name           string            `gorm:"size:128;not null" json:"name"`
	Unit           string            `gorm:"size:32" json:"unit"`
	DataType       ParameterDataType `gorm:"size:32;default:'FLOAT32'" json:"data_type"`
	RegisterAddress int              `json:"register_address"`
	RegisterType   string            `gorm:"size:32" json:"register_type"` // HOLDING_REGISTER, INPUT_REGISTER, COIL, DISCRETE_INPUT
	ScaleFactor    float64           `gorm:"default:1.0" json:"scale_factor"`
	Offset         float64           `gorm:"default:0.0" json:"offset"`
	LowLimit       *float64          `json:"low_limit"`
	HighLimit      *float64          `json:"high_limit"`
	WarningLow     *float64          `json:"warning_low"`
	WarningHigh    *float64          `json:"warning_high"`
	CriticalLow    *float64          `json:"critical_low"`
	CriticalHigh   *float64          `json:"critical_high"`
	CurrentValue   *float64          `json:"current_value"`
	LastUpdated    *time.Time        `json:"last_updated"`
	Enabled        bool              `gorm:"default:true" json:"enabled"`
	CreatedAt      time.Time         `json:"created_at"`
	UpdatedAt      time.Time         `json:"updated_at"`
}

type Sensor struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	ParameterID uint      `gorm:"index;not null" json:"parameter_id"`
	Model       string    `gorm:"size:128" json:"model"`
	SerialNo    string    `gorm:"size:128" json:"serial_no"`
	Calibration *time.Time `json:"calibration"`
	Accuracy    string    `gorm:"size:64" json:"accuracy"`
	RangeMin    float64   `json:"range_min"`
	RangeMax    float64   `json:"range_max"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
