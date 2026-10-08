package model

import (
	"time"
)

type SystemHealth struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	Timestamp      time.Time `gorm:"index;not null" json:"timestamp"`
	CPUPercent     float64   `json:"cpu_percent"`
	RAMUsedBytes   uint64    `json:"ram_used_bytes"`
	RAMTotalBytes  uint64    `json:"ram_total_bytes"`
	RAMPercent     float64   `json:"ram_percent"`
	DiskUsedBytes  uint64    `json:"disk_used_bytes"`
	DiskTotalBytes uint64    `json:"disk_total_bytes"`
	DiskPercent    float64   `json:"disk_percent"`
	UptimeSeconds  uint64    `json:"uptime_seconds"`
	GoroutineCount int       `json:"goroutine_count"`
	DBStatus       string    `gorm:"size:32" json:"db_status"`
	ServiceStatus  string    `gorm:"size:32" json:"service_status"`
	WorkerStatus   string    `gorm:"size:32" json:"worker_status"`
	QueueStatus    string    `gorm:"size:32" json:"queue_status"`
	CreatedAt      time.Time `json:"created_at"`
}

type SystemStatusOverview struct {
	// System resources
	CPUPercent        float64   `json:"cpu_percent"`
	ProcessCPUPercent float64   `json:"process_cpu_percent"`
	CPUPerCore        []float64 `json:"cpu_per_core,omitempty"`
	RAMPercent        float64   `json:"ram_percent"`
	RAMUsedMB      float64 `json:"ram_used_mb"`
	RAMTotalMB     float64 `json:"ram_total_mb"`
	DiskPercent    float64 `json:"disk_percent"`
	DiskUsedGB     float64 `json:"disk_used_gb"`
	DiskTotalGB    float64 `json:"disk_total_gb"`
	UptimeHuman    string  `json:"uptime_human"`
	UptimeSeconds  uint64  `json:"uptime_seconds"`
	OS             string  `json:"os"`
	Arch           string  `json:"arch"`
	Hostname       string  `json:"hostname"`
	NumCPU         int     `json:"num_cpu"`
	Goroutines     int     `json:"goroutines"`
	AppVersion     string  `json:"app_version"`

	// Datalogger service status
	ServiceStatus   string `json:"service_status"`
	WorkerStatus    string `json:"worker_status"`
	SchedulerStatus string `json:"scheduler_status"`
	QueueStatus     string `json:"queue_status"`
	DatabaseStatus  string `json:"database_status"`

	// Devices
	TotalDevices    int    `json:"total_devices"`
	OnlineDevices   int    `json:"online_devices"`
	OfflineDevices  int    `json:"offline_devices"`
	ErrorDevices    int    `json:"error_devices"`
	LastDeviceComm  string `json:"last_device_comm"`

	// Data
	DataReceivedCount int64   `json:"data_received_count"`
	DataPerSec        float64 `json:"data_per_sec"`
	DataQuality       string  `json:"data_quality"`
	MissingDataCount  int64   `json:"missing_data_count"`
	InvalidDataCount  int64   `json:"invalid_data_count"`
	LastDataReceived  string  `json:"last_data_received"`

	// Communication
	TotalConnections  int     `json:"total_connections"`
	SuccessRate       float64 `json:"success_rate"`
	FailedRate        float64 `json:"failed_rate"`
	AvgLatencyMs      int     `json:"avg_latency_ms"`

	// Alarm
	ActiveAlarms        int `json:"active_alarms"`
	WarningAlarms       int `json:"warning_alarms"`
	CriticalAlarms      int `json:"critical_alarms"`
	AcknowledgedAlarms  int `json:"acknowledged_alarms"`
}
