package handler

import (
	"strconv"
	"time"

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
		"status":      "HEALTHY",
		"timestamp":   time.Now(),
		"cpu_percent": info.CPUPercent,
		"ram_percent": info.RAMPercent,
		"disk_percent": info.DiskPercent,
		"uptime":      info.UptimeHuman,
		"goroutines":  info.Goroutines,
	})
}

// System Resources
func (h *SystemHandler) GetResources(c *gin.Context) {
	info := sysinfo.GetInfo()
	response.OK(c, info)
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

func (h *SystemHandler) AcknowledgeAlarm(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid alarm ID")
		return
	}

	username := "admin"
	if u, exists := c.Get("username"); exists {
		if uStr, ok := u.(string); ok {
			username = uStr
		}
	}

	if err := h.systemService.AcknowledgeAlarm(uint(id), username); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.OK(c, gin.H{"acknowledged": true})
}

// Logs
func (h *SystemHandler) GetLogs(c *gin.Context) {
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
	response.OK(c, logs)
}

func (h *SystemHandler) GetAuditTrails(c *gin.Context) {
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
	response.OK(c, trails)
}

func (h *SystemHandler) GetCommunicationLogs(c *gin.Context) {
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

