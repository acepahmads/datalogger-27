package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"datalogger/internal/communication"
	"datalogger/internal/model"
	"datalogger/internal/repository"
	"datalogger/internal/service"
	"datalogger/pkg/response"
)

// ActivityMonitorHandler provides HTTP endpoints for the Superadmin System Activity Monitor
type ActivityMonitorHandler struct {
	monitor       *communication.ActivityMonitor
	deviceService *service.DeviceService
	commTracer    *communication.CommTracer
}

// NewActivityMonitorHandler creates a new ActivityMonitorHandler
func NewActivityMonitorHandler(
	monitor *communication.ActivityMonitor,
	deviceService *service.DeviceService,
	commTracer *communication.CommTracer,
) *ActivityMonitorHandler {
	return &ActivityMonitorHandler{
		monitor:       monitor,
		deviceService: deviceService,
		commTracer:    commTracer,
	}
}

// GetOverview returns executive summary metrics for the System Overview bar
func (h *ActivityMonitorHandler) GetOverview(c *gin.Context) {
	if h.monitor == nil {
		response.InternalError(c, "Activity monitor service unavailable")
		return
	}

	var devices []model.Device
	if h.deviceService != nil {
		devList, _, err := h.deviceService.ListDevices(repository.DeviceFilterParams{
			Page:     1,
			PageSize: 500,
		})
		if err == nil {
			devices = devList
		}
	}

	overview := h.monitor.GetSystemOverview(devices)
	response.OK(c, overview)
}

// GetEvents returns filtered activity events from the in-memory ring buffer
func (h *ActivityMonitorHandler) GetEvents(c *gin.Context) {
	if h.monitor == nil {
		response.InternalError(c, "Activity monitor service unavailable")
		return
	}

	filter := communication.ActivityFilter{
		Search:   c.Query("search"),
		Severity: c.Query("severity"),
		Category: c.Query("category"),
	}

	if dIDStr := c.Query("device_id"); dIDStr != "" {
		if id, err := strconv.ParseUint(dIDStr, 10, 32); err == nil {
			filter.DeviceID = uint(id)
		}
	}

	if sIDStr := c.Query("since_id"); sIDStr != "" {
		if sid, err := strconv.ParseInt(sIDStr, 10, 64); err == nil {
			filter.SinceID = sid
		}
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			filter.Limit = l
		}
	} else {
		filter.Limit = 50
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			filter.Offset = o
		}
	}

	if stStr := c.Query("start_time"); stStr != "" {
		if t, err := time.Parse(time.RFC3339, stStr); err == nil {
			filter.StartTime = &t
		}
	}

	if etStr := c.Query("end_time"); etStr != "" {
		if t, err := time.Parse(time.RFC3339, etStr); err == nil {
			filter.EndTime = &t
		}
	}

	events, total := h.monitor.GetEvents(filter)
	response.OK(c, gin.H{
		"events": events,
		"total":  total,
		"limit":  filter.Limit,
		"offset": filter.Offset,
	})
}

// GetPollingSnapshots returns per-device live polling activity states
func (h *ActivityMonitorHandler) GetPollingSnapshots(c *gin.Context) {
	if h.monitor == nil {
		response.InternalError(c, "Activity monitor service unavailable")
		return
	}

	snapshots := h.monitor.GetAllPollingSnapshots()

	// Augment with device metadata if missing in snapshot
	if h.deviceService != nil {
		devList, _, err := h.deviceService.ListDevices(repository.DeviceFilterParams{
			Page:     1,
			PageSize: 500,
		})
		if err == nil {
			snapMap := make(map[uint]*communication.DevicePollingSnapshot)
			for i := range snapshots {
				snapMap[snapshots[i].DeviceID] = &snapshots[i]
			}

			// Ensure every enabled device has an entry
			for _, dev := range devList {
				if snap, exists := snapMap[dev.ID]; exists {
					if snap.DeviceCode == "" {
						snap.DeviceCode = dev.DeviceCode
					}
					if snap.DeviceName == "" {
						snap.DeviceName = dev.DeviceName
					}
					if snap.Protocol == "" && dev.Connection != nil {
						snap.Protocol = string(dev.Connection.Protocol)
					}
					if snap.Transport == "" && dev.Connection != nil {
						if dev.Connection.Protocol == model.ProtocolModbusRTU {
							snap.Transport = fmt.Sprintf("Serial (%s)", dev.Connection.SerialPort)
						} else {
							snap.Transport = fmt.Sprintf("TCP (%s:%d)", dev.Connection.Host, dev.Connection.Port)
						}
					}
					if snap.CommunicationStatus == "UNKNOWN" || snap.CommunicationStatus == "" {
						snap.CommunicationStatus = string(dev.ConnectionStatus)
					}
				} else {
					// Create snapshot for device if not yet tracked
					transport := ""
					protocol := ""
					if dev.Connection != nil {
						protocol = string(dev.Connection.Protocol)
						if dev.Connection.Protocol == model.ProtocolModbusRTU {
							transport = fmt.Sprintf("Serial (%s)", dev.Connection.SerialPort)
						} else {
							transport = fmt.Sprintf("TCP (%s:%d)", dev.Connection.Host, dev.Connection.Port)
						}
					}
					state := communication.PollingStateWaiting
					if !dev.Enabled || dev.Status != model.DeviceStatusActive {
						state = communication.PollingStateStopped
					}
					snapshots = append(snapshots, communication.DevicePollingSnapshot{
						DeviceID:            dev.ID,
						DeviceCode:          dev.DeviceCode,
						DeviceName:          dev.DeviceName,
						Protocol:            protocol,
						Transport:           transport,
						CommunicationStatus: string(dev.ConnectionStatus),
						PollingState:        state,
						RecoveryState:       "NORMAL",
						LastSuccessfulResponse: dev.LastDataAt,
						LastSuccessfulTelemetry: dev.LastDataAt,
					})
				}
			}
		}
	}

	response.OK(c, snapshots)
}

// GetDeviceDetail returns detailed diagnostic status and recent frames for a selected device
func (h *ActivityMonitorHandler) GetDeviceDetail(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid device ID")
		return
	}
	deviceID := uint(id)

	var dev *model.Device
	if h.deviceService != nil {
		dev, err = h.deviceService.GetDeviceByID(deviceID)
		if err != nil || dev == nil {
			response.NotFound(c, "Device not found")
			return
		}
	}

	snapshot, _ := h.monitor.GetDevicePollingSnapshot(deviceID)

	var recentFrames []communication.CommTraceFrame
	if h.commTracer != nil {
		recentFrames = h.commTracer.GetRecentFrames(20, deviceID)
	}

	var params []model.Parameter
	if h.deviceService != nil {
		if pList, pErr := h.deviceService.ListParameters(deviceID); pErr == nil {
			params = pList
		}
	}

	// Safe connection view without any sensitive secrets
	var safeConn map[string]interface{}
	if dev != nil && dev.Connection != nil {
		safeConn = map[string]interface{}{
			"id":                 dev.Connection.ID,
			"protocol":           dev.Connection.Protocol,
			"serial_port":        dev.Connection.SerialPort,
			"baud_rate":          dev.Connection.BaudRate,
			"data_bits":          dev.Connection.DataBits,
			"stop_bits":          dev.Connection.StopBits,
			"parity":             dev.Connection.Parity,
			"host":               dev.Connection.Host,
			"port":               dev.Connection.Port,
			"slave_id":           dev.Connection.SlaveID,
			"timeout_ms":         dev.Connection.Timeout,
			"retry_count":        dev.Connection.RetryCount,
			"polling_interval":   dev.Connection.PollingInterval,
			"batch_read_enabled": dev.Connection.BatchReadEnabled,
			"max_batch_registers": dev.Connection.MaxBatchRegisters,
			"enabled":            dev.Connection.Enabled,
		}
	}

	response.OK(c, gin.H{
		"device": gin.H{
			"id":                 dev.ID,
			"device_code":        dev.DeviceCode,
			"device_name":        dev.DeviceName,
			"status":             dev.Status,
			"connection_status":  dev.ConnectionStatus,
			"enabled":            dev.Enabled,
			"last_data_at":       dev.LastDataAt,
		},
		"connection":    safeConn,
		"snapshot":      snapshot,
		"parameters":    params,
		"recent_frames": recentFrames,
	})
}

// GetBackgroundJobs returns the execution status of all background services
func (h *ActivityMonitorHandler) GetBackgroundJobs(c *gin.Context) {
	if h.monitor == nil {
		response.InternalError(c, "Activity monitor service unavailable")
		return
	}

	jobs := h.monitor.GetAllBackgroundJobs()
	response.OK(c, jobs)
}

// ClearEvents clears the in-memory ring buffer (SUPERADMIN only)
func (h *ActivityMonitorHandler) ClearEvents(c *gin.Context) {
	if h.monitor == nil {
		response.InternalError(c, "Activity monitor service unavailable")
		return
	}

	h.monitor.ClearEvents()
	response.OK(c, gin.H{"message": "Activity log ring buffer cleared successfully"})
}

// ExportEvents exports activity events to CSV or JSON
func (h *ActivityMonitorHandler) ExportEvents(c *gin.Context) {
	if h.monitor == nil {
		response.InternalError(c, "Activity monitor service unavailable")
		return
	}

	format := strings.ToLower(c.DefaultQuery("format", "csv"))
	filter := communication.ActivityFilter{
		Search:   c.Query("search"),
		Severity: c.Query("severity"),
		Category: c.Query("category"),
		Limit:    1000,
	}

	if dIDStr := c.Query("device_id"); dIDStr != "" {
		if id, err := strconv.ParseUint(dIDStr, 10, 32); err == nil {
			filter.DeviceID = uint(id)
		}
	}

	data, contentType, err := h.monitor.ExportEvents(filter, format)
	if err != nil {
		response.InternalError(c, fmt.Sprintf("Failed to export events: %v", err))
		return
	}

	filename := fmt.Sprintf("system_activity_%s.%s", time.Now().Format("20060102_150405"), format)
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	c.Data(http.StatusOK, contentType, data)
}
