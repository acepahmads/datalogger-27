package handler

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
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
	systemService   *service.SystemService
	outboundService service.OutboundService
}

func NewSystemHandler(systemService *service.SystemService) *SystemHandler {
	return &SystemHandler{systemService: systemService}
}

// SetOutboundService injects outbound integration service for network & delivery monitoring
func (h *SystemHandler) SetOutboundService(outboundService service.OutboundService) {
	h.outboundService = outboundService
}

// GetRemovableStorage reports detected removable USB drives / flash drives
func (h *SystemHandler) GetRemovableStorage(c *gin.Context) {
	status := sysinfo.GetRemovableDrives()
	response.OK(c, status)
}

type SaveToUSBRequest struct {
	MountPoint    string `json:"mount_point"`
	Filename      string `json:"filename"`
	ContentBase64 string `json:"content_base64"`
}

// SaveToRemovableStorage saves exported telemetry file directly to a detected USB mount point on edge system
func (h *SystemHandler) SaveToRemovableStorage(c *gin.Context) {
	var req SaveToUSBRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request payload: "+err.Error())
		return
	}

	if req.Filename == "" || req.ContentBase64 == "" {
		response.BadRequest(c, "Filename and content_base64 are required")
		return
	}

	// Verify that mount point is an actual detected removable drive
	status := sysinfo.GetRemovableDrives()
	if !status.Detected || len(status.Drives) == 0 {
		response.BadRequest(c, "No removable USB storage detected on edge hardware")
		return
	}

	targetMount := status.Drives[0].MountPoint
	if req.MountPoint != "" {
		found := false
		for _, d := range status.Drives {
			if strings.EqualFold(d.MountPoint, req.MountPoint) || strings.EqualFold(d.Device, req.MountPoint) {
				targetMount = d.MountPoint
				found = true
				break
			}
		}
		if !found {
			response.BadRequest(c, "Specified mount point is not a recognized removable drive")
			return
		}
	}

	// Sanitize filename to prevent directory traversal
	cleanFilename := filepath.Base(req.Filename)
	outDir := filepath.Join(targetMount, "telemetry_exports")
	_ = os.MkdirAll(outDir, 0755)

	outPath := filepath.Join(outDir, cleanFilename)
	data, err := base64.StdEncoding.DecodeString(req.ContentBase64)
	if err != nil {
		response.BadRequest(c, "Invalid base64 content: "+err.Error())
		return
	}

	if err := os.WriteFile(outPath, data, 0644); err != nil {
		response.InternalError(c, "Failed to write file to USB drive: "+err.Error())
		return
	}

	response.OK(c, gin.H{
		"saved":       true,
		"file_path":   outPath,
		"mount_point": targetMount,
		"filename":    cleanFilename,
		"bytes":       len(data),
	})
}

// GetNetworkStatus reports local network interfaces, gateway connectivity, internet probe, and outbound links
func (h *SystemHandler) GetNetworkStatus(c *gin.Context) {
	diag := sysinfo.GetNetworkDiagnostics()

	outboundOverview := gin.H{
		"total_links":    0,
		"healthy_count":  0,
		"degraded_count": 0,
		"failed_count":   0,
		"disabled_count": 0,
		"unknown_count":  0,
		"overall_status": "UNKNOWN",
		"queue_summary": gin.H{
			"pending":      0,
			"sending":      0,
			"delivered":    0,
			"retrying":     0,
			"failed":       0,
			"acknowledged": 0,
			"total":        0,
		},
		"destinations": []gin.H{},
	}

	if h.outboundService != nil {
		ctx := c.Request.Context()
		destinations, err := h.outboundService.GetDestinations(ctx, false, true)
		if err == nil {
			destList := make([]gin.H, 0, len(destinations))
			var healthyCount, degradedCount, failedCount, disabledCount, unknownCount int

			// Overall queue summary
			if qs, qsErr := h.outboundService.GetQueueSummary(ctx); qsErr == nil && qs != nil {
				outboundOverview["queue_summary"] = qs
			}

			for _, d := range destinations {
				destID := d.ID
				healthStatus := sysinfo.MapOutboundStatus(d.Enabled, d.Status, d.SuccessDeliveries, d.FailedDeliveries)

				switch healthStatus {
				case "HEALTHY":
					healthyCount++
				case "DEGRADED":
					degradedCount++
				case "FAILED":
					failedCount++
				case "DISABLED":
					disabledCount++
				case "UNKNOWN":
					unknownCount++
				}

				totalDeliv := d.SuccessDeliveries + d.FailedDeliveries
				successRate := 0.0
				if totalDeliv > 0 {
					successRate = (float64(d.SuccessDeliveries) / float64(totalDeliv)) * 100.0
				}

				var lastError string
				var lastAttempt *time.Time
				var destQueueCount int64

				qItems, qTotal, qErr := h.outboundService.GetQueueItems(ctx, repository.QueueFilter{
					DestinationID: &destID,
					Page:          1,
					PageSize:      1,
				})
				if qErr == nil {
					destQueueCount = qTotal
					if len(qItems) > 0 {
						lastError = sysinfo.SanitizeError(qItems[0].LastError)
						lastAttempt = qItems[0].LastAttemptAt
					}
				}

				destList = append(destList, gin.H{
					"id":                 d.ID,
					"name":               d.Name,
					"type":               d.Type,
					"endpoint":           sysinfo.SanitizeEndpoint(d.Endpoint),
					"enabled":            d.Enabled,
					"status":             healthStatus,
					"raw_status":         d.Status,
					"format":             d.Format,
					"interval_seconds":   d.IntervalSeconds,
					"success_deliveries": d.SuccessDeliveries,
					"failed_deliveries":  d.FailedDeliveries,
					"success_rate":       successRate,
					"last_delivery":      d.LastDelivery,
					"last_attempt_at":    lastAttempt,
					"last_error":         lastError,
					"queue_count":        destQueueCount,
				})
			}

			overallStatus := "HEALTHY"
			if len(destinations) == 0 {
				overallStatus = "UNKNOWN"
			} else if disabledCount == len(destinations) {
				overallStatus = "DISABLED"
			} else if failedCount > 0 {
				overallStatus = "FAILED"
			} else if degradedCount > 0 {
				overallStatus = "DEGRADED"
			} else if unknownCount > 0 && healthyCount == 0 {
				overallStatus = "UNKNOWN"
			}

			outboundOverview["total_links"] = len(destinations)
			outboundOverview["healthy_count"] = healthyCount
			outboundOverview["degraded_count"] = degradedCount
			outboundOverview["failed_count"] = failedCount
			outboundOverview["disabled_count"] = disabledCount
			outboundOverview["unknown_count"] = unknownCount
			outboundOverview["overall_status"] = overallStatus
			outboundOverview["destinations"] = destList
		}
	}

	response.OK(c, gin.H{
		"local_network": diag.LocalNetwork,
		"internet":      diag.Internet,
		"outbound":      outboundOverview,
		"timestamp":     diag.Timestamp,
	})
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
		"process_rss_bytes":   info.ProcessRSSBytes,
		"process_rss_mb":      info.ProcessRSSMB,
		"go_heap_alloc_bytes": info.GoHeapAllocBytes,
		"go_heap_sys_bytes":   info.GoHeapSysBytes,
		"go_num_goroutine":    info.GoNumGoroutine,
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

// CheckDomainCompatibility validates parameters and category alignment before domain switch
func (h *SystemHandler) CheckDomainCompatibility(c *gin.Context) {
	target := c.Query("target")
	if target == "" {
		response.BadRequest(c, "Target domain parameter 'target' is required (e.g. 'AQMS', 'WQMS', 'CUSTOM')")
		return
	}

	report, err := h.systemService.CheckDomainCompatibility(target)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.OK(c, report)
}

// GetBranding retrieves database-persisted organization branding (customer + maker)
func (h *SystemHandler) GetBranding(c *gin.Context) {
	branding, err := h.systemService.GetBranding()
	if err != nil {
		response.InternalError(c, "Failed to load branding: "+err.Error())
		return
	}
	response.OK(c, branding)
}

// SaveBranding updates database-persisted organization branding (customer + maker)
func (h *SystemHandler) SaveBranding(c *gin.Context) {
	var req service.UpdateSystemBrandingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid branding payload: "+err.Error())
		return
	}

	branding, err := h.systemService.SaveBranding(req)
	if err != nil {
		response.InternalError(c, "Failed to save branding: "+err.Error())
		return
	}
	response.OK(c, branding)
}

// GetStationLocation retrieves the persistent station coordinates and configuration
func (h *SystemHandler) GetStationLocation(c *gin.Context) {
	loc, err := h.systemService.GetStationLocation()
	if err != nil {
		response.InternalError(c, "Failed to load station location: "+err.Error())
		return
	}
	response.OK(c, loc)
}

// SaveStationLocation updates station coordinates with validation and audit logging
func (h *SystemHandler) SaveStationLocation(c *gin.Context) {
	var req service.UpdateStationLocationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid location payload: "+err.Error())
		return
	}

	var actorID uint
	if uid, exists := c.Get("userID"); exists {
		if id, ok := uid.(uint); ok {
			actorID = id
		}
	}
	actorUsername := "operator"
	if u, exists := c.Get("username"); exists {
		if un, ok := u.(string); ok {
			actorUsername = un
		}
	}
	ip := c.ClientIP()
	ua := c.Request.UserAgent()

	loc, err := h.systemService.SaveStationLocation(req, actorID, actorUsername, ip, ua)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, loc)
}

// GetIPGeolocationEstimate checks configured IP geolocation provider status without leaking IP or inventing fake coordinates
func (h *SystemHandler) GetIPGeolocationEstimate(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"available":   false,
		"provider":    "NONE",
		"status":      "UNAVAILABLE",
		"message":     "Internet IP geolocation provider is not configured or disabled. No external provider requests performed.",
		"coordinates": nil,
		"disclaimer":  "Internet Geolocation berbasis IP hanya merupakan estimasi perkiraan dan tidak seakurat GPS fisik.",
	})
}

// LocationSearchResult represents a location item found via geocoding or coordinates
type LocationSearchResult struct {
	Name        string  `json:"name"`
	DisplayName string  `json:"display_name"`
	Lat         float64 `json:"lat"`
	Lon         float64 `json:"lon"`
	Type        string  `json:"type"`
}

// SearchLocation resolves places and coordinates for station location mapping
func (h *SystemHandler) SearchLocation(c *gin.Context) {
	q := strings.TrimSpace(c.Query("q"))
	if q == "" {
		response.OK(c, []LocationSearchResult{})
		return
	}

	// 1. Direct coordinate parsing: e.g. "-6.9218, 107.6070" or "-6.9218 107.6070"
	coordPattern := regexp.MustCompile(`^([-+]?\d{1,2}(?:\.\d+)?)[,\s;]+([-+]?\d{1,3}(?:\.\d+)?)$`)
	if matches := coordPattern.FindStringSubmatch(q); len(matches) == 3 {
		lat, err1 := strconv.ParseFloat(matches[1], 64)
		lng, err2 := strconv.ParseFloat(matches[2], 64)
		if err1 == nil && err2 == nil && lat >= -90.0 && lat <= 90.0 && lng >= -180.0 && lng <= 180.0 {
			res := []LocationSearchResult{
				{
					Name:        fmt.Sprintf("Coordinates (%.6f, %.6f)", lat, lng),
					DisplayName: fmt.Sprintf("Latitude: %.6f, Longitude: %.6f", lat, lng),
					Lat:         lat,
					Lon:         lng,
					Type:        "coordinate",
				},
			}
			response.OK(c, res)
			return
		}
	}

	// 2. Query Nominatim OpenStreetMap Geocoding API
	escapedQuery := url.QueryEscape(q)
	nominatimURL := fmt.Sprintf("https://nominatim.openstreetmap.org/search?format=json&q=%s&limit=6&addressdetails=1", escapedQuery)

	req, err := http.NewRequestWithContext(c.Request.Context(), http.MethodGet, nominatimURL, nil)
	if err != nil {
		response.OK(c, []LocationSearchResult{})
		return
	}
	req.Header.Set("User-Agent", "Datalogger-CBI/1.0 (LocationSearch; contact: admin@datalogger.local)")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 4 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		response.OK(c, []LocationSearchResult{})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		response.OK(c, []LocationSearchResult{})
		return
	}

	var rawResults []struct {
		PlaceID     interface{} `json:"place_id"`
		Name        string      `json:"name"`
		DisplayName string      `json:"display_name"`
		Lat         string      `json:"lat"`
		Lon         string      `json:"lon"`
		Type        string      `json:"type"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&rawResults); err != nil {
		response.OK(c, []LocationSearchResult{})
		return
	}

	results := make([]LocationSearchResult, 0, len(rawResults))
	for _, item := range rawResults {
		lat, _ := strconv.ParseFloat(item.Lat, 64)
		lng, _ := strconv.ParseFloat(item.Lon, 64)
		name := item.Name
		if name == "" {
			parts := strings.Split(item.DisplayName, ",")
			if len(parts) > 0 {
				name = strings.TrimSpace(parts[0])
			}
		}
		results = append(results, LocationSearchResult{
			Name:        name,
			DisplayName: item.DisplayName,
			Lat:         lat,
			Lon:         lng,
			Type:        item.Type,
		})
	}

	response.OK(c, results)
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

