package handler

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"datalogger/internal/communication"
	"datalogger/pkg/response"
	"datalogger/internal/service"
)

// CommTraceHandler handles live communication inspection and sniffer toggle endpoints
type CommTraceHandler struct {
	tracer        *communication.CommTracer
	deviceService *service.DeviceService
}

// NewCommTraceHandler creates a new CommTraceHandler
func NewCommTraceHandler(tracer *communication.CommTracer, deviceService *service.DeviceService) *CommTraceHandler {
	return &CommTraceHandler{
		tracer:        tracer,
		deviceService: deviceService,
	}
}

// ToggleTraceRequest specifies parameters to enable or disable the live trace
type ToggleTraceRequest struct {
	Active          bool `json:"active"`
	TargetDeviceID  uint `json:"target_device_id"`
	DurationMinutes int  `json:"duration_minutes"` // default 10 minutes
}

// GetStatus returns the current live communication trace status
func (h *CommTraceHandler) GetStatus(c *gin.Context) {
	if h.tracer == nil {
		response.OK(c, gin.H{
			"active":             false,
			"remaining_seconds":  0,
			"target_device_id":   0,
			"total_frames":      0,
			"can_toggle":        false,
		})
		return
	}

	status := h.tracer.GetStatus()

	// Check if caller is SUPERADMIN
	roleVal, _ := c.Get("role")
	role, _ := roleVal.(string)
	canToggle := strings.EqualFold(role, "SUPERADMIN") || strings.EqualFold(role, "Administrator")

	response.OK(c, gin.H{
		"active":            status.Active,
		"activated_by":      status.ActivatedBy,
		"activated_at":      status.ActivatedAt,
		"expires_at":        status.ExpiresAt,
		"remaining_seconds": status.RemainingSeconds,
		"target_device_id":  status.TargetDeviceID,
		"total_frames":      status.TotalFrames,
		"can_toggle":        canToggle,
	})
}

// GetFrames returns recent captured transaction frames from memory
func (h *CommTraceHandler) GetFrames(c *gin.Context) {
	if h.tracer == nil {
		response.OK(c, []communication.CommTraceFrame{})
		return
	}

	limit := 100
	if lStr := c.Query("limit"); lStr != "" {
		if l, err := strconv.Atoi(lStr); err == nil && l > 0 {
			limit = l
		}
	}
	if limit > 200 {
		limit = 200
	}

	var devID uint
	if dStr := c.Query("device_id"); dStr != "" {
		if d, err := strconv.ParseUint(dStr, 10, 32); err == nil {
			devID = uint(d)
		}
	}

	frames := h.tracer.GetRecentFrames(limit, devID)
	response.OK(c, frames)
}

// ToggleTrace activates or deactivates live frame capture (SUPERADMIN only)
func (h *CommTraceHandler) ToggleTrace(c *gin.Context) {
	if h.tracer == nil {
		response.InternalError(c, "CommTracer not initialized")
		return
	}

	var req ToggleTraceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Default to activating if empty body
		req.Active = true
	}

	usernameVal, _ := c.Get("username")
	username, _ := usernameVal.(string)
	if username == "" {
		username = "superadmin"
	}

	if req.Active {
		durMin := req.DurationMinutes
		if durMin <= 0 {
			durMin = 10 // Safe default 10 minutes
		}
		if durMin > 30 {
			durMin = 30
		}
		duration := time.Duration(durMin) * time.Minute

		status := h.tracer.Start(username, duration, req.TargetDeviceID)
		response.OK(c, gin.H{
			"message": "Live sensor communication trace activated",
			"status":  status,
		})
	} else {
		status := h.tracer.Stop(username)
		response.OK(c, gin.H{
			"message": "Live sensor communication trace stopped",
			"status":  status,
		})
	}
}

// ClearFrames empties the in-memory trace ring buffer (SUPERADMIN only)
func (h *CommTraceHandler) ClearFrames(c *gin.Context) {
	if h.tracer != nil {
		h.tracer.ClearBuffer()
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Frame capture buffer cleared",
	})
}
