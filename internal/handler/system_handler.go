package handler

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"datalogger/internal/model"
	"datalogger/internal/repository"
	"datalogger/internal/service"
	"datalogger/pkg/response"
	"datalogger/pkg/sysinfo"

	"github.com/gin-gonic/gin"
)

type SystemHandler struct {
	systemService *service.SystemService
}

func NewSystemHandler(systemService *service.SystemService) *SystemHandler {
	return &SystemHandler{systemService: systemService}
}

// System Status
func (h *SystemHandler) GetStatus(c *gin.Context) {
	status, err := h.systemService.GetSystemStatusOverview()
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, status)
}

// System Health
func (h *SystemHandler) GetHealth(c *gin.Context) {
	info := sysinfo.GetInfo()
	response.OK(c, gin.H{
		"status":              "HEALTHY",
		"timestamp":           time.Now(),
		"cpu_percent":         info.CPUPercent,
		"process_cpu_percent": info.ProcessCPUPercent,
		"ram_percent":         info.RAMPercent,
		"disk_percent":        info.DiskPercent,
		"uptime":              info.UptimeHuman,
		"goroutines":          info.Goroutines,
	})
}

// System Resources
func (h *SystemHandler) GetResources(c *gin.Context) {
	info := sysinfo.GetInfo()
	response.OK(c, info)
}

// GetSerialPorts lists available hardware/virtual serial COM ports
func (h *SystemHandler) GetSerialPorts(c *gin.Context) {
	ports := sysinfo.GetAvailableSerialPorts()
	response.OK(c, ports)
}

// GetDetailedSerialPorts lists available serial ports with persistent metadata and recommendations
func (h *SystemHandler) GetDetailedSerialPorts(c *gin.Context) {
	ports := sysinfo.GetDetailedSerialPorts()
	response.OK(c, ports)
}

// Devices
func (h *SystemHandler) GetDevices(c *gin.Context) {
	devices, err := h.systemService.GetDevices()
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, devices)
}

func (h *SystemHandler) GetDeviceByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid device ID")
		return
	}

	dev, err := h.systemService.GetDeviceByID(uint(id))
	if err != nil {
		response.NotFound(c, "Device not found")
		return
	}
	response.OK(c, dev)
}

func (h *SystemHandler) GetDeviceStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid device ID")
		return
	}

	dev, err := h.systemService.GetDeviceByID(uint(id))
	if err != nil {
		response.NotFound(c, "Device not found")
		return
	}

	response.OK(c, gin.H{
		"id":                 dev.ID,
		"code":               dev.Code,
		"name":               dev.Name,
		"status":             dev.Status,
		"latency_ms":         dev.LatencyMs,
		"last_communication": dev.LastCommunication,
		"success_count":      dev.SuccessCount,
		"failed_count":       dev.FailedCount,
	})
}

// Data endpoints
func (h *SystemHandler) GetData(c *gin.Context) {
	// Returns latest operational telemetry readings from parameters
	devices, err := h.systemService.GetDevices()
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	type ParamData struct {
		DeviceID     uint     `json:"device_id"`
		DeviceName   string   `json:"device_name"`
		Code         string   `json:"code"`
		Name         string   `json:"name"`
		Value        *float64 `json:"value"`
		Unit         string   `json:"unit"`
		LastUpdated  string   `json:"last_updated"`
	}

	var results []ParamData
	for _, d := range devices {
		for _, p := range d.Parameters {
			results = append(results, ParamData{
				DeviceID:    d.ID,
				DeviceName:  d.Name,
				Code:        p.Code,
				Name:        p.Name,
				Value:       p.CurrentValue,
				Unit:        p.Unit,
				LastUpdated: time.Now().Format("15:04:05"),
			})
		}
	}
	response.OK(c, results)
}

func (h *SystemHandler) GetLatestData(c *gin.Context) {
	h.GetData(c)
}

func (h *SystemHandler) GetDataTrend(c *gin.Context) {
	// Synthesize trend datapoints for Phase 1 visualization
	now := time.Now()
	var points []gin.H
	for i := 20; i >= 0; i-- {
		t := now.Add(-time.Duration(i*5) * time.Minute)
		points = append(points, gin.H{
			"timestamp": t.Format("15:04"),
			"voltage":   220.0 + (float64(i%5) * 0.4),
			"current":   44.0 + (float64(i%3) * 0.6),
			"temp":      24.0 + (float64(i%4) * 0.2),
		})
	}
	response.OK(c, points)
}

// Alarms
func (h *SystemHandler) GetAlarms(c *gin.Context) {
	alarmService := h.systemService.GetAlarmService()
	pageStr := c.Query("page")
	pageSizeStr := c.Query("page_size")

	// If paginated query is requested and alarmService is available
	if (pageStr != "" || pageSizeStr != "") && alarmService != nil {
		page, _ := strconv.Atoi(pageStr)
		pageSize, _ := strconv.Atoi(pageSizeStr)
		var devID, paramID uint
		if dStr := c.Query("device_id"); dStr != "" {
			if v, err := strconv.ParseUint(dStr, 10, 32); err == nil {
				devID = uint(v)
			}
		}
		if pStr := c.Query("parameter_id"); pStr != "" {
			if v, err := strconv.ParseUint(pStr, 10, 32); err == nil {
				paramID = uint(v)
			}
		}

		filter := repository.AlarmFilterParams{
			Status:         c.Query("status"),
			Severity:       c.Query("severity"),
			DeviceID:       devID,
			ParameterID:    paramID,
			AlarmCode:      c.Query("alarm_code"),
			AcknowledgedBy: c.Query("acknowledged_by"),
			Search:         c.Query("search"),
			StartTime:      parseFlexibleTime(c.Query("start_time")),
			EndTime:        parseFlexibleTime(c.Query("end_time")),
			Page:           page,
			PageSize:       pageSize,
		}

		alarms, total, err := alarmService.ListAlarms(c.Request.Context(), filter)
		if err != nil {
			response.InternalError(c, err.Error())
			return
		}

		response.OK(c, gin.H{
			"items":     alarms,
			"total":     total,
			"page":      filter.Page,
			"page_size": filter.PageSize,
		})
		return
	}

	// Default unpaginated query (backward-compatible slice output for web UI)
	status := c.Query("status")
	alarms, err := h.systemService.GetAlarms(status)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, alarms)
}

func (h *SystemHandler) GetActiveAlarms(c *gin.Context) {
	alarms, err := h.systemService.GetAlarms("ACTIVE")
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, alarms)
}

func (h *SystemHandler) GetAlarmByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid alarm ID")
		return
	}

	alarm, err := h.systemService.GetAlarmByID(uint(id))
	if err != nil {
		if errors.Is(err, repository.ErrAlarmNotFound) {
			response.NotFound(c, "Alarm not found")
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.OK(c, alarm)
}

func (h *SystemHandler) CreateAlarm(c *gin.Context) {
	var alarm model.Alarm
	if err := c.ShouldBindJSON(&alarm); err != nil {
		response.BadRequest(c, "Invalid alarm payload: "+err.Error())
		return
	}

	if err := h.systemService.CreateAlarm(&alarm); err != nil {
		if errors.Is(err, repository.ErrMissingAlarmCode) ||
			errors.Is(err, repository.ErrMissingAlarmMessage) ||
			errors.Is(err, repository.ErrInvalidAlarmSeverity) {
			response.BadRequest(c, err.Error())
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.Created(c, alarm)
}

func (h *SystemHandler) AcknowledgeAlarm(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid alarm ID")
		return
	}

	// Server-derived user identity; never trust arbitrary user ID or username from request payload
	var userID *uint
	if uidVal, exists := c.Get("userID"); exists {
		if uid, ok := uidVal.(uint); ok && uid > 0 {
			userID = &uid
		}
	}

	username := "admin"
	if u, exists := c.Get("username"); exists {
		if uStr, ok := u.(string); ok && strings.TrimSpace(uStr) != "" {
			username = uStr
		}
	}

	displayName := username
	if fn, exists := c.Get("fullName"); exists {
		if fnStr, ok := fn.(string); ok && strings.TrimSpace(fnStr) != "" {
			displayName = fnStr
		}
	}

	var req struct {
		Notes string `json:"notes"`
	}
	_ = c.ShouldBindJSON(&req)

	alarm, err := h.systemService.AcknowledgeAlarmWithUser(c.Request.Context(), uint(id), userID, username, displayName, req.Notes)
	if err != nil {
		if errors.Is(err, repository.ErrAlarmNotFound) {
			response.NotFound(c, "Alarm not found")
			return
		}
		if errors.Is(err, repository.ErrCannotAcknowledgeCleared) || errors.Is(err, repository.ErrAlarmAlreadyAcknowledged) {
			response.BadRequest(c, err.Error())
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	respData := gin.H{
		"acknowledged":            true,
		"id":                      alarm.ID,
		"alarm_code":              alarm.AlarmCode,
		"status":                  alarm.Status,
		"acknowledged_at":         alarm.AcknowledgedAt,
		"acknowledged_by":         alarm.AcknowledgedBy,
		"acknowledged_by_user_id": alarm.AcknowledgedByUserID,
		"acknowledged_by_name":    alarm.AcknowledgedByName,
		"notes":                   alarm.Notes,
		"alarm":                   alarm,
	}

	response.OK(c, respData)
}

// GetAlarmHistory provides searchable, paginated historical alarm records with rich context
func (h *SystemHandler) GetAlarmHistory(c *gin.Context) {
	alarmService := h.systemService.GetAlarmService()
	if alarmService == nil {
		response.InternalError(c, "Alarm service unavailable")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 1000 {
		pageSize = 1000
	}

	var devID, paramID uint
	if dStr := c.Query("device_id"); dStr != "" {
		if v, err := strconv.ParseUint(dStr, 10, 32); err == nil {
			devID = uint(v)
		}
	}
	if pStr := c.Query("parameter_id"); pStr != "" {
		if v, err := strconv.ParseUint(pStr, 10, 32); err == nil {
			paramID = uint(v)
		}
	}

	filter := repository.AlarmFilterParams{
		Status:         c.Query("status"),
		Severity:       c.Query("severity"),
		DeviceID:       devID,
		ParameterID:    paramID,
		AlarmCode:      c.Query("alarm_code"),
		AcknowledgedBy: c.Query("acknowledged_by"),
		Search:         c.Query("search"),
		StartTime:      parseFlexibleTime(c.Query("start_time")),
		EndTime:        parseFlexibleTime(c.Query("end_time")),
		Page:           page,
		PageSize:       pageSize,
	}

	alarms, total, err := alarmService.ListAlarms(c.Request.Context(), filter)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	now := time.Now().UTC()
	items := make([]gin.H, len(alarms))
	for i, a := range alarms {
		isOngoing := a.ClearedAt == nil || a.ClearedAt.IsZero()
		var durationSecs int64
		if !isOngoing {
			durationSecs = int64(a.ClearedAt.Sub(a.TriggeredAt).Seconds())
		} else {
			durationSecs = int64(now.Sub(a.TriggeredAt).Seconds())
		}
		if durationSecs < 0 {
			durationSecs = 0
		}

		devCode := ""
		devName := ""
		if a.Device != nil {
			devCode = a.Device.Code
			devName = a.Device.Name
		}
		paramName := ""
		if a.Parameter != nil {
			paramName = a.Parameter.Name
		}

		recoveryReason := ""
		if !isOngoing {
			if a.Notes != "" && strings.Contains(a.Notes, "Recovery:") {
				recoveryReason = a.Notes
			} else {
				recoveryReason = "Condition normalized / auto-cleared"
			}
		}

		items[i] = gin.H{
			"id":                      a.ID,
			"device_id":               a.DeviceID,
			"device_code":             devCode,
			"device_name":             devName,
			"parameter_id":            a.ParameterID,
			"parameter_name":          paramName,
			"alarm_code":              a.AlarmCode,
			"severity":                a.Severity,
			"status":                  a.Status,
			"message":                 a.Message,
			"notes":                   a.Notes,
			"recovery_reason":         recoveryReason,
			"trigger_value":           a.TriggerValue,
			"threshold_value":         a.ThresholdValue,
			"triggered_at":            a.TriggeredAt,
			"acknowledged_at":         a.AcknowledgedAt,
			"acknowledged_by":         a.AcknowledgedBy,
			"acknowledged_by_user_id": a.AcknowledgedByUserID,
			"acknowledged_by_name":    a.AcknowledgedByName,
			"cleared_at":              a.ClearedAt,
			"duration_seconds":        durationSecs,
			"duration_formatted":      formatAlarmDuration(durationSecs),
			"is_ongoing":              isOngoing,
			"device":                  a.Device,
			"parameter":               a.Parameter,
		}
	}

	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))
	if totalPages == 0 && total == 0 {
		totalPages = 1
	}

	response.OK(c, gin.H{
		"items":       items,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": totalPages,
		"has_next":    page < totalPages,
		"has_prev":    page > 1,
	})
}

func (h *SystemHandler) ClearAlarm(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid alarm ID")
		return
	}

	alarm, err := h.systemService.ClearAlarm(uint(id))
	if err != nil {
		if errors.Is(err, repository.ErrAlarmNotFound) {
			response.NotFound(c, "Alarm not found")
			return
		}
		if errors.Is(err, repository.ErrAlarmAlreadyCleared) {
			response.BadRequest(c, err.Error())
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.OK(c, gin.H{
		"cleared":    true,
		"id":         id,
		"cleared_at": alarm.ClearedAt,
		"status":     alarm.Status,
	})
}

// Redaction & Date Helpers
var sensitiveLogPattern = regexp.MustCompile(`(?i)(password|secret|token|api_key|bot_token|auth_token|bearer)\s*[:=]\s*["']?([^\s"',;]+)["']?`)

func redactSensitiveString(s string) string {
	if s == "" {
		return ""
	}
	return sensitiveLogPattern.ReplaceAllString(s, `$1: "[REDACTED]"`)
}

func parseFlexibleDate(str string) *time.Time {
	if str == "" {
		return nil
	}
	formats := []string{
		time.RFC3339,
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02",
	}
	for _, fmtStr := range formats {
		if t, err := time.Parse(fmtStr, str); err == nil {
			return &t
		}
	}
	return nil
}

// Logs
func (h *SystemHandler) GetLogs(c *gin.Context) {
	pageStr := c.Query("page")
	pageSizeStr := c.Query("page_size")
	level := c.Query("level")
	component := c.Query("component")
	search := c.Query("search")
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	isPaginatedRequest := pageStr != "" || pageSizeStr != "" || level != "" || component != "" || search != "" || startDateStr != "" || endDateStr != ""

	if isPaginatedRequest {
		page, _ := strconv.Atoi(pageStr)
		if page < 1 {
			page = 1
		}
		pageSize, _ := strconv.Atoi(pageSizeStr)
		if pageSize < 1 || pageSize > 500 {
			pageSize = 50
		}
		startTime := parseFlexibleDate(startDateStr)
		endTime := parseFlexibleDate(endDateStr)

		params := repository.SystemLogFilterParams{
			Level:     level,
			Component: component,
			Search:    search,
			StartTime: startTime,
			EndTime:   endTime,
			Page:      page,
			PageSize:  pageSize,
		}

		logs, total, err := h.systemService.GetLogsFiltered(params)
		if err != nil {
			response.InternalError(c, err.Error())
			return
		}

		for i := range logs {
			logs[i].Message = redactSensitiveString(logs[i].Message)
			logs[i].Details = redactSensitiveString(logs[i].Details)
		}

		totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))
		if totalPages == 0 && total == 0 {
			totalPages = 1
		}

		response.OK(c, gin.H{
			"items":       logs,
			"total":       total,
			"page":        page,
			"page_size":   pageSize,
			"total_pages": totalPages,
			"has_next":    page < totalPages,
			"has_prev":    page > 1,
		})
		return
	}

	// Legacy unpaginated endpoint
	limit := 50
	if lStr := c.Query("limit"); lStr != "" {
		if l, err := strconv.Atoi(lStr); err == nil {
			limit = l
		}
	}
	logs, err := h.systemService.GetLogs(limit)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	for i := range logs {
		logs[i].Message = redactSensitiveString(logs[i].Message)
		logs[i].Details = redactSensitiveString(logs[i].Details)
	}
	response.OK(c, logs)
}

func (h *SystemHandler) GetAuditTrails(c *gin.Context) {
	pageStr := c.Query("page")
	pageSizeStr := c.Query("page_size")
	username := c.Query("username")
	action := c.Query("action")
	resourceName := c.Query("resource")
	search := c.Query("search")
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	isPaginatedRequest := pageStr != "" || pageSizeStr != "" || username != "" || action != "" || resourceName != "" || search != "" || startDateStr != "" || endDateStr != ""

	if isPaginatedRequest {
		page, _ := strconv.Atoi(pageStr)
		if page < 1 {
			page = 1
		}
		pageSize, _ := strconv.Atoi(pageSizeStr)
		if pageSize < 1 || pageSize > 500 {
			pageSize = 50
		}
		startTime := parseFlexibleDate(startDateStr)
		endTime := parseFlexibleDate(endDateStr)

		params := repository.AuditTrailFilterParams{
			Username:  username,
			Action:    action,
			Resource:  resourceName,
			Search:    search,
			StartTime: startTime,
			EndTime:   endTime,
			Page:      page,
			PageSize:  pageSize,
		}

		trails, total, err := h.systemService.GetAuditTrailsFiltered(params)
		if err != nil {
			response.InternalError(c, err.Error())
			return
		}

		for i := range trails {
			trails[i].Details = redactSensitiveString(trails[i].Details)
		}

		totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))
		if totalPages == 0 && total == 0 {
			totalPages = 1
		}

		response.OK(c, gin.H{
			"items":       trails,
			"total":       total,
			"page":        page,
			"page_size":   pageSize,
			"total_pages": totalPages,
			"has_next":    page < totalPages,
			"has_prev":    page > 1,
		})
		return
	}

	// Legacy unpaginated endpoint
	limit := 50
	if lStr := c.Query("limit"); lStr != "" {
		if l, err := strconv.Atoi(lStr); err == nil {
			limit = l
		}
	}
	trails, err := h.systemService.GetAuditTrails(limit)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	for i := range trails {
		trails[i].Details = redactSensitiveString(trails[i].Details)
	}
	response.OK(c, trails)
}

func (h *SystemHandler) GetCommunicationLogs(c *gin.Context) {
	pageStr := c.Query("page")
	pageSizeStr := c.Query("page_size")
	deviceIDStr := c.Query("device_id")
	protocol := c.Query("protocol")
	direction := c.Query("direction")
	status := c.Query("status")
	search := c.Query("search")
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	isPaginatedRequest := pageStr != "" || pageSizeStr != "" || deviceIDStr != "" || protocol != "" || direction != "" || status != "" || search != "" || startDateStr != "" || endDateStr != ""

	if isPaginatedRequest {
		page, _ := strconv.Atoi(pageStr)
		if page < 1 {
			page = 1
		}
		pageSize, _ := strconv.Atoi(pageSizeStr)
		if pageSize < 1 || pageSize > 500 {
			pageSize = 50
		}
		var devID *uint
		if deviceIDStr != "" {
			if d, err := strconv.ParseUint(deviceIDStr, 10, 32); err == nil && d > 0 {
				u := uint(d)
				devID = &u
			}
		}
		startTime := parseFlexibleDate(startDateStr)
		endTime := parseFlexibleDate(endDateStr)

		params := repository.CommunicationLogFilterParams{
			DeviceID:  devID,
			Protocol:  protocol,
			Direction: direction,
			Status:    status,
			Search:    search,
			StartTime: startTime,
			EndTime:   endTime,
			Page:      page,
			PageSize:  pageSize,
		}

		logs, total, err := h.systemService.GetCommunicationLogsFiltered(params)
		if err != nil {
			response.InternalError(c, err.Error())
			return
		}

		for i := range logs {
			logs[i].ErrorMessage = redactSensitiveString(logs[i].ErrorMessage)
		}

		totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))
		if totalPages == 0 && total == 0 {
			totalPages = 1
		}

		response.OK(c, gin.H{
			"items":       logs,
			"total":       total,
			"page":        page,
			"page_size":   pageSize,
			"total_pages": totalPages,
			"has_next":    page < totalPages,
			"has_prev":    page > 1,
		})
		return
	}

	limit := 50
	if lStr := c.Query("limit"); lStr != "" {
		if l, err := strconv.Atoi(lStr); err == nil {
			limit = l
		}
	}
	logs, err := h.systemService.GetCommunicationLogs(limit)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	for i := range logs {
		logs[i].ErrorMessage = redactSensitiveString(logs[i].ErrorMessage)
	}
	response.OK(c, logs)
}

// Database Monitoring & Configuration
func (h *SystemHandler) GetDatabaseStatus(c *gin.Context) {
	metrics := h.systemService.GetDatabaseMetrics()
	response.OK(c, metrics)
}

type DatabaseConnTestRequest struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	Database string `json:"database"`
}

func (h *SystemHandler) TestDatabaseConnection(c *gin.Context) {
	var req DatabaseConnTestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request payload")
		return
	}

	ok, ver, latency, err := h.systemService.TestDatabaseConnection(req.Host, req.Port, req.Username, req.Password, req.Database)
	if err != nil {
		c.JSON(200, gin.H{
			"success": false,
			"error":   err.Error(),
			"latency": latency,
		})
		return
	}

	response.OK(c, gin.H{
		"connected":  ok,
		"version":    ver,
		"latency_ms": latency,
		"message":    "Successfully connected to MariaDB database",
	})
}

func (h *SystemHandler) SaveDatabaseConfig(c *gin.Context) {
	var req DatabaseConnTestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request payload")
		return
	}

	if err := h.systemService.SaveDatabaseConfig(req.Host, req.Port, req.Username, req.Password, req.Database); err != nil {
		response.InternalError(c, "Failed to save database config: "+err.Error())
		return
	}

	response.OK(c, gin.H{
		"saved":   true,
		"message": "Database configuration saved successfully. Will apply on application restart.",
	})
}

// GetConfig returns the masked runtime system configuration (config.view)
func (h *SystemHandler) GetConfig(c *gin.Context) {
	cfg := h.systemService.GetSystemConfig()
	response.OK(c, cfg)
}

// SaveConfig safely updates runtime configuration parameters with validation and audit logging (config.manage)
func (h *SystemHandler) SaveConfig(c *gin.Context) {
	var req service.SaveSystemConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid configuration payload: "+err.Error())
		return
	}

	var actorID uint
	if uid, exists := c.Get("userID"); exists {
		if id, ok := uid.(uint); ok {
			actorID = id
		}
	}
	actorUsername := "system"
	if u, exists := c.Get("username"); exists {
		if un, ok := u.(string); ok {
			actorUsername = un
		}
	}
	ip := c.ClientIP()
	ua := c.Request.UserAgent()

	result, err := h.systemService.SaveSystemConfig(req, actorID, actorUsername, ip, ua)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.OK(c, result)
}


func formatAlarmDuration(seconds int64) string {
	if seconds < 0 {
		return "0s"
	}
	if seconds < 60 {
		return fmt.Sprintf("%ds", seconds)
	}
	if seconds < 3600 {
		return fmt.Sprintf("%dm %ds", seconds/60, seconds%60)
	}
	if seconds < 86400 {
		return fmt.Sprintf("%dh %dm", seconds/3600, (seconds%3600)/60)
	}
	return fmt.Sprintf("%dd %dh", seconds/86400, (seconds%86400)/3600)
}

