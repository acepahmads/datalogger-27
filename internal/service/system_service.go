package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"datalogger/internal/config"
	"datalogger/internal/database"
	"datalogger/internal/model"
	"datalogger/internal/repository"
	"datalogger/pkg/sysinfo"
)

type SystemService struct {
	systemRepo       *repository.SystemRepository
	phaseRepo        *repository.PhaseRepository
	cfg              *config.Config
	telemetryService *TelemetryService
	backupService    *BackupService
	alarmService     *AlarmService
}

func NewSystemService(systemRepo *repository.SystemRepository, phaseRepo *repository.PhaseRepository, cfg *config.Config) *SystemService {
	return &SystemService{
		systemRepo: systemRepo,
		phaseRepo:  phaseRepo,
		cfg:        cfg,
	}
}

func (s *SystemService) SetTelemetryService(ts *TelemetryService) {
	s.telemetryService = ts
}

func (s *SystemService) SetBackupService(bs *BackupService) {
	s.backupService = bs
}

func (s *SystemService) SetAlarmService(as *AlarmService) {
	s.alarmService = as
}

func (s *SystemService) GetAlarmService() *AlarmService {
	return s.alarmService
}

func (s *SystemService) GetSystemStatusOverview() (*model.SystemStatusOverview, error) {
	info := sysinfo.GetInfo()

	// Devices summary
	devices, _ := s.systemRepo.GetAllDevices()
	onlineDev := 0
	offlineDev := 0
	errorDev := 0
	degradedDev := 0
	reconnDev := 0
	var lastComm time.Time
	var totalSuccess int64
	var totalFailed int64
	var avgLatency int

	for _, d := range devices {
		switch d.Status {
		case model.DeviceOnline:
			onlineDev++
		case model.DeviceOffline:
			offlineDev++
		case model.DeviceError:
			errorDev++
		case "DEGRADED":
			degradedDev++
		case "RECONNECTING":
			reconnDev++
		}
		if d.LastCommunication != nil && d.LastCommunication.After(lastComm) {
			lastComm = *d.LastCommunication
		}
		totalSuccess += d.SuccessCount
		totalFailed += d.FailedCount
		avgLatency += d.LatencyMs
	}

	if len(devices) > 0 {
		avgLatency = avgLatency / len(devices)
	}

	successRate := 99.8
	failedRate := 0.2
	totalComm := totalSuccess + totalFailed
	if totalComm > 0 {
		successRate = (float64(totalSuccess) / float64(totalComm)) * 100
		failedRate = (float64(totalFailed) / float64(totalComm)) * 100
	}

	// Alarms summary
	alarms, _ := s.systemRepo.GetAllAlarms("")
	activeAlarms := 0
	warningAlarms := 0
	criticalAlarms := 0
	ackAlarms := 0
	for _, a := range alarms {
		if a.Status == model.AlarmActive {
			activeAlarms++
			if a.Severity == model.SeverityWarning {
				warningAlarms++
			} else if a.Severity == model.SeverityCritical {
				criticalAlarms++
			}
		} else if a.Status == model.AlarmAcked {
			ackAlarms++
		}
	}

	lastCommStr := "Just now"
	if !lastComm.IsZero() {
		lastCommStr = lastComm.Format("15:04:05")
	}

	dbMetrics := database.GetMetrics()
	dbStatusStr := "MariaDB Edge (Healthy)"
	if dbMetrics.Status != "" {
		dbStatusStr = fmt.Sprintf("MariaDB Edge (%s)", dbMetrics.Status)
	}

	overview := &model.SystemStatusOverview{
		CPUPercent:        info.CPUPercent,
		ProcessCPUPercent: info.ProcessCPUPercent,
		CPUPerCore:        info.CPUPerCore,
		RAMPercent:        info.RAMPercent,
		RAMUsedMB:         float64(info.RAMUsedBytes) / (1024 * 1024),
		RAMTotalMB:        float64(info.RAMTotalBytes) / (1024 * 1024),
		DiskPercent:       info.DiskPercent,
		DiskUsedGB:        float64(info.DiskUsedBytes) / (1024 * 1024 * 1024),
		DiskTotalGB:       float64(info.DiskTotalBytes) / (1024 * 1024 * 1024),
		UptimeHuman:       info.UptimeHuman,
		UptimeSeconds:     info.UptimeSeconds,
		OS:                fmt.Sprintf("%s / %s", info.OS, info.Arch),
		Arch:              info.Arch,
		Hostname:          info.Hostname,
		NumCPU:            info.NumCPU,
		Goroutines:        info.Goroutines,
		AppVersion:        s.cfg.Version,

		ServiceStatus:   "RUNNING",
		WorkerStatus:    "READY",
		SchedulerStatus: "ACTIVE",
		QueueStatus:     "OPTIMAL",
		DatabaseStatus:  dbStatusStr,
		DBLatencyMs:     dbMetrics.LatencyMs,
		DBVersion:       dbMetrics.Version,
		DBHost:          dbMetrics.Host,
		DBPort:          dbMetrics.Port,
		DBDatabase:      dbMetrics.Database,

		TotalDevices:        len(devices),
		OnlineDevices:       onlineDev,
		OfflineDevices:      offlineDev,
		DegradedDevices:     degradedDev,
		ReconnectingDevices: reconnDev,
		ErrorDevices:        errorDev,
		LastDeviceComm:      lastCommStr,
		PersistenceStatus:   "HEALTHY",

		DataReceivedCount: totalSuccess,
		DataPerSec:        0.0,
		DataQuality:       fmt.Sprintf("%.1f%% GOOD", successRate),
		MissingDataCount:  0,
		InvalidDataCount:  totalFailed,
		LastDataReceived:  time.Now().Format("15:04:05"),

		TotalConnections: len(devices),
		SuccessRate:      successRate,
		FailedRate:       failedRate,
		AvgLatencyMs:     avgLatency,

		ActiveAlarms:       activeAlarms,
		WarningAlarms:      warningAlarms,
		CriticalAlarms:     criticalAlarms,
		AcknowledgedAlarms: ackAlarms,
	}

	if s.telemetryService != nil {
		metrics := s.telemetryService.GetMetrics()
		overview.PersistenceStatus = metrics.PersistenceStatus
		overview.QueuePendingRecords = metrics.QueuePendingRecords
		overview.QueueSpoolSizeBytes = metrics.QueueSpoolSizeBytes
		overview.QueueTotalReplayed = metrics.QueueTotalReplayed
		overview.QueueChecksumErrors = metrics.QueueChecksumErrors
		overview.QueueDiskPercent = metrics.QueueDiskPercent
		if metrics.QueuePendingRecords > 0 {
			overview.QueueStatus = fmt.Sprintf("REPLAYING (%d pending)", metrics.QueuePendingRecords)
		} else {
			overview.QueueStatus = "HEALTHY"
		}

		if metrics.IngestedCount > 0 {
			overview.DataReceivedCount = int64(metrics.IngestedCount)
			if info.UptimeSeconds > 0 {
				overview.DataPerSec = float64(metrics.IngestedCount) / float64(info.UptimeSeconds)
			}
			if metrics.IngestedCount >= metrics.ErrorCount {
				qPct := (float64(metrics.IngestedCount-metrics.ErrorCount) / float64(metrics.IngestedCount)) * 100.0
				overview.DataQuality = fmt.Sprintf("%.1f%% GOOD", qPct)
			}
		} else if onlineDev > 0 {
			overview.DataPerSec = float64(onlineDev) * 1.0
		}
	} else if onlineDev > 0 {
		overview.DataPerSec = float64(onlineDev) * 1.0
	}

	if s.backupService != nil {
		bStats := s.backupService.GetStatus()
		if st, ok := bStats["status"].(string); ok {
			overview.BackupStatus = st
		}
		if t, ok := bStats["last_backup_time"].(string); ok {
			overview.LastBackupTime = t
		}
		if st, ok := bStats["last_backup_status"].(model.BackupStatus); ok {
			overview.LastBackupStatus = string(st)
		} else if st, ok := bStats["last_backup_status"].(string); ok {
			overview.LastBackupStatus = st
		}
		if cnt, ok := bStats["total_valid_backups"].(int64); ok {
			overview.TotalBackupsCount = int(cnt)
		}
		if mb, ok := bStats["storage_used_mb"].(float64); ok {
			overview.BackupStorageUsedMB = mb
		}
		if act, ok := bStats["active_operation"].(string); ok {
			overview.ActiveBackupJob = act
		}
	} else {
		overview.BackupStatus = "READY"
	}

	return overview, nil
}

func (s *SystemService) GetDevices() ([]model.Device, error) {
	return s.systemRepo.GetAllDevices()
}

func (s *SystemService) GetDeviceByID(id uint) (*model.Device, error) {
	return s.systemRepo.GetDeviceByID(id)
}

func (s *SystemService) GetAlarms(status string) ([]model.Alarm, error) {
	if s.alarmService != nil {
		return s.alarmService.GetAlarmsByStatus(context.Background(), status)
	}
	return s.systemRepo.GetAllAlarms(status)
}

func (s *SystemService) GetAlarmByID(id uint) (*model.Alarm, error) {
	if s.alarmService != nil {
		return s.alarmService.GetAlarmByID(context.Background(), id)
	}
	return nil, errors.New("alarm service not configured")
}

func (s *SystemService) CreateAlarm(alarm *model.Alarm) error {
	if s.alarmService != nil {
		return s.alarmService.CreateAlarm(context.Background(), alarm)
	}
	return errors.New("alarm service not configured")
}

func (s *SystemService) AcknowledgeAlarm(id uint, username string, notes ...string) error {
	if s.alarmService != nil {
		_, err := s.alarmService.AcknowledgeAlarm(context.Background(), id, username, notes...)
		return err
	}
	return s.systemRepo.AcknowledgeAlarm(id, username)
}

func (s *SystemService) AcknowledgeAlarmWithUser(ctx context.Context, id uint, userID *uint, username, displayName string, notes ...string) (*model.Alarm, error) {
	if s.alarmService != nil {
		return s.alarmService.AcknowledgeAlarmWithUser(ctx, id, userID, username, displayName, notes...)
	}
	return nil, errors.New("alarm service not configured")
}

func (s *SystemService) ClearAlarm(id uint) (*model.Alarm, error) {
	if s.alarmService != nil {
		return s.alarmService.ClearAlarm(context.Background(), id)
	}
	return nil, errors.New("alarm service not configured")
}

func (s *SystemService) GetLogs(limit int) ([]model.SystemLog, error) {
	if limit <= 0 {
		limit = 50
	}
	return s.systemRepo.GetSystemLogs(limit)
}

func (s *SystemService) GetLogsFiltered(params repository.SystemLogFilterParams) ([]model.SystemLog, int64, error) {
	return s.systemRepo.GetSystemLogsFiltered(params)
}

func (s *SystemService) GetAuditTrails(limit int) ([]model.AuditTrail, error) {
	if limit <= 0 {
		limit = 50
	}
	return s.systemRepo.GetAuditTrails(limit)
}

func (s *SystemService) GetAuditTrailsFiltered(params repository.AuditTrailFilterParams) ([]model.AuditTrail, int64, error) {
	return s.systemRepo.GetAuditTrailsFiltered(params)
}

func (s *SystemService) GetCommunicationLogs(limit int) ([]model.CommunicationLog, error) {
	if limit <= 0 {
		limit = 50
	}
	return s.systemRepo.GetCommunicationLogs(limit)
}

func (s *SystemService) GetCommunicationLogsFiltered(params repository.CommunicationLogFilterParams) ([]model.CommunicationLog, int64, error) {
	return s.systemRepo.GetCommunicationLogsFiltered(params)
}

func (s *SystemService) AddCommunicationLog(log *model.CommunicationLog) error {
	return s.systemRepo.AddCommunicationLog(log)
}

func (s *SystemService) GetDatabaseMetrics() database.DatabaseMetrics {
	return database.GetMetrics()
}

func (s *SystemService) TestDatabaseConnection(host, port, user, password, dbName string) (bool, string, float64, error) {
	return database.TestConnection(host, port, user, password, dbName)
}

func (s *SystemService) SaveDatabaseConfig(host, port, user, password, dbName string) error {
	cfg := config.Get()
	cfg.DBType = "mariadb"
	cfg.DBHost = host
	cfg.DBPort = port
	cfg.DBUser = user
	if password != "" && password != "••••••••" && password != "********" {
		cfg.DBPassword = password
	}
	cfg.DBName = dbName

	return config.Save("config.json", cfg)
}

type SystemConfigDTO struct {
	AppName                     string `json:"app_name"`
	Version                     string `json:"version"`
	Environment                 string `json:"environment"`
	Port                        string `json:"port"`
	Host                        string `json:"host"`
	DBType                      string `json:"db_type"`
	DBHost                      string `json:"db_host"`
	DBPort                      string `json:"db_port"`
	DBUser                      string `json:"db_user"`
	DBPassword                  string `json:"db_password"` // Masked
	DBName                      string `json:"db_name"`
	LogLevel                    string `json:"log_level"`
	JWTExpirationHours          int    `json:"jwt_expiration_hours"`
	SerialPort                  string `json:"serial_port"`
	SerialBaudRate              int    `json:"serial_baud_rate"`
	SerialDataBits              int    `json:"serial_data_bits"`
	SerialStopBits              int    `json:"serial_stop_bits"`
	SerialParity                string `json:"serial_parity"`
	SerialTimeoutMs             int    `json:"serial_timeout_ms"`
	QueueEnabled                bool   `json:"queue_enabled"`
	QueueBatchSize              int    `json:"queue_batch_size"`
	QueueFlushIntervalMs        int    `json:"queue_flush_interval_ms"`
	BackupScheduleEnabled       bool   `json:"backup_schedule_enabled"`
	BackupScheduleIntervalHours int    `json:"backup_schedule_interval_hours"`
	BackupScheduleTime          string `json:"backup_schedule_time"`
	BackupCompression           string `json:"backup_compression"`
}

type SaveSystemConfigRequest struct {
	Environment        string `json:"environment"`
	Port               string `json:"port"`
	Host               string `json:"host"`
	DBHost             string `json:"db_host"`
	DBPort             string `json:"db_port"`
	DBUser             string `json:"db_user"`
	DBPassword         string `json:"db_password"`
	DBName             string `json:"db_name"`
	LogLevel           string `json:"log_level"`
	JWTExpirationHours int    `json:"jwt_expiration_hours"`
	SerialPort         string `json:"serial_port"`
	SerialBaudRate     int    `json:"serial_baud_rate"`
	SerialDataBits     int    `json:"serial_data_bits"`
	SerialStopBits     int    `json:"serial_stop_bits"`
	SerialParity       string `json:"serial_parity"`
	SerialTimeoutMs    int    `json:"serial_timeout_ms"`
}

type ConfigSaveResult struct {
	Saved                 bool     `json:"saved"`
	RestartRequired       bool     `json:"restart_required"`
	AcquisitionDisruption bool     `json:"acquisition_disruption"`
	RestartRequiredFields []string `json:"restart_required_fields"`
	DisruptedFields       []string `json:"disrupted_fields"`
	Message               string   `json:"message"`
}

func (s *SystemService) GetSystemConfig() *SystemConfigDTO {
	cfg := config.Get()
	return &SystemConfigDTO{
		AppName:                     cfg.AppName,
		Version:                     cfg.Version,
		Environment:                 cfg.Environment,
		Port:                        cfg.Port,
		Host:                        cfg.Host,
		DBType:                      cfg.DBType,
		DBHost:                      cfg.DBHost,
		DBPort:                      cfg.DBPort,
		DBUser:                      cfg.DBUser,
		DBPassword:                  "••••••••",
		DBName:                      cfg.DBName,
		LogLevel:                    cfg.LogLevel,
		JWTExpirationHours:          cfg.JWTExpirationHours,
		SerialPort:                  cfg.SerialPort,
		SerialBaudRate:              cfg.SerialBaudRate,
		SerialDataBits:              cfg.SerialDataBits,
		SerialStopBits:              cfg.SerialStopBits,
		SerialParity:                cfg.SerialParity,
		SerialTimeoutMs:             cfg.SerialTimeoutMs,
		QueueEnabled:                cfg.QueueEnabled,
		QueueBatchSize:              cfg.QueueBatchSize,
		QueueFlushIntervalMs:        cfg.QueueFlushIntervalMs,
		BackupScheduleEnabled:       cfg.BackupScheduleEnabled,
		BackupScheduleIntervalHours: cfg.BackupScheduleIntervalHours,
		BackupScheduleTime:          cfg.BackupScheduleTime,
		BackupCompression:           cfg.BackupCompression,
	}
}

func (s *SystemService) SaveSystemConfig(req SaveSystemConfigRequest, actorID uint, actorUsername, ip, ua string) (*ConfigSaveResult, error) {
	cfg := config.Get()

	// 1. Validation
	if req.Port != "" {
		p, err := strconv.Atoi(req.Port)
		if err != nil || p < 1 || p > 65535 {
			return nil, errors.New("HTTP Port must be a valid port number between 1 and 65535")
		}
	}
	if req.DBPort != "" {
		p, err := strconv.Atoi(req.DBPort)
		if err != nil || p < 1 || p > 65535 {
			return nil, errors.New("Database Port must be a valid port number between 1 and 65535")
		}
	}
	if req.LogLevel != "" {
		validLevels := map[string]bool{"debug": true, "info": true, "warn": true, "error": true}
		if !validLevels[strings.ToLower(req.LogLevel)] {
			return nil, errors.New("invalid log level: must be 'debug', 'info', 'warn', or 'error'")
		}
	}
	if req.SerialBaudRate > 0 {
		validBauds := map[int]bool{1200: true, 2400: true, 4800: true, 9600: true, 19200: true, 38400: true, 57600: true, 115200: true}
		if !validBauds[req.SerialBaudRate] {
			return nil, errors.New("invalid serial baud rate: must be standard rate (e.g. 9600, 19200, 115200)")
		}
	}
	if req.SerialDataBits > 0 && req.SerialDataBits != 7 && req.SerialDataBits != 8 {
		return nil, errors.New("serial data bits must be 7 or 8")
	}
	if req.SerialStopBits > 0 && req.SerialStopBits != 1 && req.SerialStopBits != 2 {
		return nil, errors.New("serial stop bits must be 1 or 2")
	}
	if req.SerialParity != "" {
		p := strings.ToUpper(req.SerialParity)
		if p != "N" && p != "E" && p != "O" && p != "NONE" && p != "EVEN" && p != "ODD" {
			return nil, errors.New("serial parity must be 'N' (None), 'E' (Even), or 'O' (Odd)")
		}
	}

	var restartFields []string
	var disruptedFields []string
	var changedNotes []string

	// Detect changes
	if req.DBHost != "" && req.DBHost != cfg.DBHost {
		restartFields = append(restartFields, "db_host")
		changedNotes = append(changedNotes, fmt.Sprintf("db_host: %s -> %s", cfg.DBHost, req.DBHost))
		cfg.DBHost = req.DBHost
	}
	if req.DBPort != "" && req.DBPort != cfg.DBPort {
		restartFields = append(restartFields, "db_port")
		changedNotes = append(changedNotes, fmt.Sprintf("db_port: %s -> %s", cfg.DBPort, req.DBPort))
		cfg.DBPort = req.DBPort
	}
	if req.DBName != "" && req.DBName != cfg.DBName {
		restartFields = append(restartFields, "db_name")
		changedNotes = append(changedNotes, fmt.Sprintf("db_name: %s -> %s", cfg.DBName, req.DBName))
		cfg.DBName = req.DBName
	}
	if req.DBUser != "" && req.DBUser != cfg.DBUser {
		restartFields = append(restartFields, "db_user")
		changedNotes = append(changedNotes, fmt.Sprintf("db_user: %s -> %s", cfg.DBUser, req.DBUser))
		cfg.DBUser = req.DBUser
	}
	if req.DBPassword != "" && req.DBPassword != "••••••••" && req.DBPassword != "********" {
		changedNotes = append(changedNotes, "db_password: [REDACTED_CHANGED]")
		cfg.DBPassword = req.DBPassword
	}
	if req.Port != "" && req.Port != cfg.Port {
		restartFields = append(restartFields, "port")
		changedNotes = append(changedNotes, fmt.Sprintf("port: %s -> %s", cfg.Port, req.Port))
		cfg.Port = req.Port
	}
	if req.Host != "" && req.Host != cfg.Host {
		restartFields = append(restartFields, "host")
		changedNotes = append(changedNotes, fmt.Sprintf("host: %s -> %s", cfg.Host, req.Host))
		cfg.Host = req.Host
	}
	if req.Environment != "" && req.Environment != cfg.Environment {
		changedNotes = append(changedNotes, fmt.Sprintf("environment: %s -> %s", cfg.Environment, req.Environment))
		cfg.Environment = req.Environment
	}
	if req.LogLevel != "" && req.LogLevel != cfg.LogLevel {
		changedNotes = append(changedNotes, fmt.Sprintf("log_level: %s -> %s", cfg.LogLevel, req.LogLevel))
		cfg.LogLevel = req.LogLevel
	}
	if req.JWTExpirationHours > 0 && req.JWTExpirationHours != cfg.JWTExpirationHours {
		changedNotes = append(changedNotes, fmt.Sprintf("jwt_expiration_hours: %d -> %d", cfg.JWTExpirationHours, req.JWTExpirationHours))
		cfg.JWTExpirationHours = req.JWTExpirationHours
	}

	// Serial communication settings
	if req.SerialPort != "" && req.SerialPort != cfg.SerialPort {
		disruptedFields = append(disruptedFields, "serial_port")
		changedNotes = append(changedNotes, fmt.Sprintf("serial_port: %s -> %s", cfg.SerialPort, req.SerialPort))
		cfg.SerialPort = req.SerialPort
	}
	if req.SerialBaudRate > 0 && req.SerialBaudRate != cfg.SerialBaudRate {
		disruptedFields = append(disruptedFields, "serial_baud_rate")
		changedNotes = append(changedNotes, fmt.Sprintf("serial_baud_rate: %d -> %d", cfg.SerialBaudRate, req.SerialBaudRate))
		cfg.SerialBaudRate = req.SerialBaudRate
	}
	if req.SerialDataBits > 0 && req.SerialDataBits != cfg.SerialDataBits {
		disruptedFields = append(disruptedFields, "serial_data_bits")
		changedNotes = append(changedNotes, fmt.Sprintf("serial_data_bits: %d -> %d", cfg.SerialDataBits, req.SerialDataBits))
		cfg.SerialDataBits = req.SerialDataBits
	}
	if req.SerialStopBits > 0 && req.SerialStopBits != cfg.SerialStopBits {
		disruptedFields = append(disruptedFields, "serial_stop_bits")
		changedNotes = append(changedNotes, fmt.Sprintf("serial_stop_bits: %d -> %d", cfg.SerialStopBits, req.SerialStopBits))
		cfg.SerialStopBits = req.SerialStopBits
	}
	if req.SerialParity != "" && req.SerialParity != cfg.SerialParity {
		disruptedFields = append(disruptedFields, "serial_parity")
		changedNotes = append(changedNotes, fmt.Sprintf("serial_parity: %s -> %s", cfg.SerialParity, req.SerialParity))
		cfg.SerialParity = req.SerialParity
	}
	if req.SerialTimeoutMs > 0 && req.SerialTimeoutMs != cfg.SerialTimeoutMs {
		disruptedFields = append(disruptedFields, "serial_timeout_ms")
		changedNotes = append(changedNotes, fmt.Sprintf("serial_timeout_ms: %d -> %d", cfg.SerialTimeoutMs, req.SerialTimeoutMs))
		cfg.SerialTimeoutMs = req.SerialTimeoutMs
	}

	if err := config.Save("config.json", cfg); err != nil {
		return nil, errors.New("failed to save configuration file: " + err.Error())
	}

	// Audit trail
	details := "Configuration updated."
	if len(changedNotes) > 0 {
		details = "Configuration updated: " + strings.Join(changedNotes, ", ")
	}

	_ = s.systemRepo.AddAuditTrail(&model.AuditTrail{
		UserID:    &actorID,
		Username:  actorUsername,
		Action:    "CONFIG_UPDATE",
		Resource:  "CONFIG",
		Details:   details,
		IPAddress: ip,
		UserAgent: ua,
		CreatedAt: time.Now(),
	})

	restartRequired := len(restartFields) > 0
	acqDisruption := len(disruptedFields) > 0 || restartRequired

	msg := "System configuration saved successfully."
	if restartRequired {
		msg += " Note: Some changes (" + strings.Join(restartFields, ", ") + ") require an application restart to take effect."
	}
	if len(disruptedFields) > 0 {
		msg += " Active serial acquisition ports may need to be reconnected."
	}

	return &ConfigSaveResult{
		Saved:                 true,
		RestartRequired:       restartRequired,
		AcquisitionDisruption: acqDisruption,
		RestartRequiredFields: restartFields,
		DisruptedFields:       disruptedFields,
		Message:               msg,
	}, nil
}


