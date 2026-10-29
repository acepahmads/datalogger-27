package model

import (
	"time"
)

// ControlType enumerates supported Modbus write function codes
type ControlType string

const (
	ControlTypeCoil              ControlType = "COIL_FC05"
	ControlTypeHoldingRegister   ControlType = "HOLDING_REGISTER_FC06"
	ControlTypeMultipleCoils     ControlType = "MULTIPLE_COILS_FC15"
	ControlTypeMultipleRegisters ControlType = "MULTIPLE_REGISTERS_FC16"
)

// ActuatorState reflects the current physical/logical state of an actuator
type ActuatorState string

const (
	ActuatorStateOff     ActuatorState = "OFF"
	ActuatorStateOn      ActuatorState = "ON"
	ActuatorStateUnknown ActuatorState = "UNKNOWN"
)

// ScheduleType specifies the automation recurrence mode
type ScheduleType string

const (
	ScheduleTypeInterval  ScheduleType = "INTERVAL"
	ScheduleTypeTimeOfDay ScheduleType = "TIME_OF_DAY"
	ScheduleTypeCron      ScheduleType = "CRON"
)

// ActuationAction indicates the direction of command
type ActuationAction string

const (
	ActionTurnOn    ActuationAction = "ON"
	ActionTurnOff   ActuationAction = "OFF"
	ActionTestWrite ActuationAction = "TEST_WRITE"
)

// ActuationStatus represents the execution result from the hardware/bus
type ActuationStatus string

const (
	ActuationStatusSuccess   ActuationStatus = "SUCCESS"
	ActuationStatusFailed    ActuationStatus = "FAILED"
	ActuationStatusTimeout   ActuationStatus = "TIMEOUT"
	ActuationStatusException ActuationStatus = "EXCEPTION"
)

// DeviceActuator represents a controllable output equipment (pump, valve, relay, fan, etc.)
type DeviceActuator struct {
	ID                   uint          `gorm:"primaryKey;autoIncrement" json:"id"`
	DeviceID             uint          `gorm:"not null;index" json:"device_id"`
	Device               *Device       `gorm:"foreignKey:DeviceID" json:"device,omitempty"`
	ActuatorCode         string        `gorm:"type:varchar(64);not null;uniqueIndex:idx_dev_act_code" json:"actuator_code"`
	ActuatorName         string        `gorm:"type:varchar(128);not null" json:"actuator_name"`
	Description          string        `gorm:"type:varchar(255)" json:"description"`
	Icon                 string        `gorm:"type:varchar(32);default:'pump'" json:"icon"`

	// Modbus Write Configuration
	ControlType          ControlType   `gorm:"type:varchar(32);not null;default:'COIL_FC05'" json:"control_type"`
	SlaveID              byte          `gorm:"not null;default:1" json:"slave_id"`
	RegisterAddress      uint16        `gorm:"not null;default:0" json:"register_address"`
	OnPayload            uint16        `gorm:"default:65280" json:"on_payload"` // 0xFF00 for FC05, or 1 for FC06
	OffPayload           uint16        `gorm:"default:0" json:"off_payload"`    // 0x0000 for FC05, or 0 for FC06

	// Safety & Watchdog Interlock
	SafetyTimeoutSeconds int           `gorm:"default:900" json:"safety_timeout_seconds"` // 15 mins default auto-off
	SafetyWatchdogActive bool          `gorm:"default:true" json:"safety_watchdog_active"`

	// Optional Readback Parameter for true closed-loop physical feedback
	FeedbackParameterID  *uint         `gorm:"index" json:"feedback_parameter_id,omitempty"`
	FeedbackParameter    *Parameter    `gorm:"foreignKey:FeedbackParameterID" json:"feedback_parameter,omitempty"`

	// Real-time Runtime State
	CurrentState         ActuatorState `gorm:"type:varchar(16);default:'UNKNOWN'" json:"current_state"`
	StateChangedAt       *time.Time    `json:"state_changed_at,omitempty"`
	AutoOffDueAt         *time.Time    `json:"auto_off_due_at,omitempty"`
	LastActuatedBy       string        `gorm:"type:varchar(64)" json:"last_actuated_by,omitempty"`
	LastActuationStatus  string        `gorm:"type:varchar(32)" json:"last_actuation_status,omitempty"`

	Enabled              bool          `gorm:"default:true" json:"enabled"`
	CreatedAt            time.Time     `json:"created_at"`
	UpdatedAt            time.Time     `json:"updated_at"`

	Schedules            []ActuationSchedule `gorm:"foreignKey:ActuatorID" json:"schedules,omitempty"`
}

// ActuationSchedule defines an automated timed actuation routine
type ActuationSchedule struct {
	ID              uint         `gorm:"primaryKey;autoIncrement" json:"id"`
	ActuatorID      uint         `gorm:"not null;index" json:"actuator_id"`
	Actuator        *DeviceActuator `gorm:"foreignKey:ActuatorID" json:"actuator,omitempty"`
	ScheduleName    string       `gorm:"type:varchar(128);not null" json:"schedule_name"`
	ScheduleType    ScheduleType `gorm:"type:varchar(32);not null;default:'INTERVAL'" json:"schedule_type"`

	// Interval Mode (e.g. repeat every 2 hours, run for 5 minutes)
	IntervalSeconds int          `gorm:"default:7200" json:"interval_seconds"` // repeat interval (e.g. 2h = 7200s)
	DurationSeconds int          `gorm:"default:300" json:"duration_seconds"`   // ON duration (e.g. 5m = 300s)

	// Time of Day Mode (e.g. HH:MM)
	StartTimeStr    string       `gorm:"type:varchar(8)" json:"start_time_str"` // e.g. "06:00"
	EndTimeStr      string       `gorm:"type:varchar(8)" json:"end_time_str"`   // e.g. "06:15"

	// Cron Expression (optional for custom schedules)
	CronExpression  string       `gorm:"type:varchar(64)" json:"cron_expression"`

	DaysOfWeek      string       `gorm:"type:varchar(32);default:'1,2,3,4,5,6,7'" json:"days_of_week"`

	LastRunAt       *time.Time   `json:"last_run_at,omitempty"`
	NextRunAt       *time.Time   `json:"next_run_at,omitempty"`
	LastStatus      string       `gorm:"type:varchar(32)" json:"last_status,omitempty"`

	Enabled         bool         `gorm:"default:true" json:"enabled"`
	CreatedAt       time.Time    `json:"created_at"`
	UpdatedAt       time.Time    `json:"updated_at"`
}

// ActuationLog provides an immutable audit trail of every actuation command
type ActuationLog struct {
	ID             uint            `gorm:"primaryKey;autoIncrement" json:"id"`
	ActuatorID     uint            `gorm:"not null;index" json:"actuator_id"`
	ScheduleID     *uint           `gorm:"index" json:"schedule_id,omitempty"`
	DeviceID       uint            `gorm:"not null;index" json:"device_id"`
	Action         ActuationAction `gorm:"type:varchar(16);not null" json:"action"`
	TriggeredBy    string          `gorm:"type:varchar(64);not null" json:"triggered_by"` // "SCHEDULER: Name" or "MANUAL: username"

	// Modbus Frame Details
	FunctionCode   byte            `json:"function_code"`
	SlaveID        byte            `json:"slave_id"`
	Address        uint16          `json:"address"`
	PayloadWritten uint16          `json:"payload_written"`

	Status         ActuationStatus `gorm:"type:varchar(24);not null" json:"status"`
	LatencyMs      int             `json:"latency_ms"`
	RawRequestHex  string          `gorm:"type:varchar(64)" json:"raw_request_hex"`
	RawResponseHex string          `gorm:"type:varchar(64)" json:"raw_response_hex"`
	ErrorMessage   string          `gorm:"type:text" json:"error_message,omitempty"`

	ExecutedAt     time.Time       `gorm:"not null;index" json:"executed_at"`
}

