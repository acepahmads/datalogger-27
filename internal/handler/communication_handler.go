package handler

import (
	"strconv"

	"datalogger/internal/communication"
	"datalogger/internal/service"
	"datalogger/pkg/response"

	"github.com/gin-gonic/gin"
)

// CommunicationHandler exposes REST API endpoints for communication management and diagnostics
type CommunicationHandler struct {
	connManager   *communication.ConnectionManager
	pollingEngine *communication.PollingEngine
	deviceService *service.DeviceService
}

// NewCommunicationHandler constructs a new CommunicationHandler
func NewCommunicationHandler(
	connManager *communication.ConnectionManager,
	pollingEngine *communication.PollingEngine,
	deviceService *service.DeviceService,
) *CommunicationHandler {
	return &CommunicationHandler{
		connManager:   connManager,
		pollingEngine: pollingEngine,
		deviceService: deviceService,
	}
}

// Connect handles POST /api/devices/:id/communication/connect
func (h *CommunicationHandler) Connect(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid device ID")
		return
	}

	dev, err := h.deviceService.GetDeviceByID(uint(id))
	if err != nil {
		response.NotFound(c, "Device not found")
		return
	}

	if err := h.connManager.ConnectDevice(c.Request.Context(), dev); err != nil {
		response.InternalError(c, "Failed to connect: "+err.Error())
		return
	}

	if h.pollingEngine != nil && dev.Enabled {
		h.pollingEngine.StartDeviceWorker(dev)
	}

	response.OK(c, gin.H{
		"device_id": id,
		"status":    "CONNECTED",
		"message":   "Device communication established successfully",
	})
}

// Disconnect handles POST /api/devices/:id/communication/disconnect
func (h *CommunicationHandler) Disconnect(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid device ID")
		return
	}

	if err := h.connManager.DisconnectDevice(uint(id)); err != nil {
		response.InternalError(c, "Failed to disconnect: "+err.Error())
		return
	}

	if h.pollingEngine != nil {
		h.pollingEngine.StopDeviceWorker(uint(id))
	}

	response.OK(c, gin.H{
		"device_id": id,
		"status":    "DISCONNECTED",
		"message":   "Device disconnected successfully",
	})
}

// Reconnect handles POST /api/devices/:id/communication/reconnect
func (h *CommunicationHandler) Reconnect(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid device ID")
		return
	}

	dev, err := h.deviceService.GetDeviceByID(uint(id))
	if err != nil {
		response.NotFound(c, "Device not found")
		return
	}

	if err := h.connManager.ReconnectDevice(c.Request.Context(), dev); err != nil {
		response.InternalError(c, "Failed to reconnect: "+err.Error())
		return
	}

	if h.pollingEngine != nil && dev.Enabled {
		h.pollingEngine.StartDeviceWorker(dev)
	}

	response.OK(c, gin.H{
		"device_id": id,
		"status":    "CONNECTED",
		"message":   "Device reconnected successfully",
	})
}

// TestConnection handles POST /api/devices/:id/communication/test
func (h *CommunicationHandler) TestConnection(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid device ID")
		return
	}

	dev, err := h.deviceService.GetDeviceByID(uint(id))
	if err != nil {
		response.NotFound(c, "Device not found")
		return
	}

	latencyMs, err := h.connManager.TestConnection(c.Request.Context(), dev)
	if err != nil {
		c.JSON(200, gin.H{
			"success":    false,
			"device_id":  id,
			"latency_ms": latencyMs,
			"error":      err.Error(),
		})
		return
	}

	response.OK(c, gin.H{
		"success":    true,
		"device_id":  id,
		"latency_ms": latencyMs,
		"message":    "Communication test successful",
	})
}

// GetStatus handles GET /api/devices/:id/communication/status
func (h *CommunicationHandler) GetStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid device ID")
		return
	}

	dev, err := h.deviceService.GetDeviceByID(uint(id))
	if err != nil {
		response.NotFound(c, "Device not found")
		return
	}

	adapterStatus, exists := h.connManager.GetAdapterStatus(uint(id))

	response.OK(c, gin.H{
		"device_id":          dev.ID,
		"device_code":        dev.DeviceCode,
		"connection_status":  dev.ConnectionStatus,
		"protocol":           dev.DeviceType,
		"adapter_active":     exists,
		"adapter_state":      adapterStatus.State,
		"connected_since":    adapterStatus.ConnectedSince,
		"last_communication": dev.LastSeenAt,
		"last_data":          dev.LastDataAt,
		"latency_ms":         dev.LatencyMs,
		"retry_count":        adapterStatus.RetryCount,
		"last_error":         adapterStatus.LastError,
		"success_count":      dev.SuccessCount,
		"failed_count":       dev.FailedCount,
	})
}

// TestReadParameter handles POST /api/devices/:id/parameters/:paramId/test-read
func (h *CommunicationHandler) TestReadParameter(c *gin.Context) {
	deviceID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid device ID")
		return
	}

	paramID, err := strconv.ParseUint(c.Param("paramId"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid parameter ID")
		return
	}

	result, err := h.pollingEngine.ReadParameterOnDemand(c.Request.Context(), uint(deviceID), uint(paramID))
	if err != nil {
		if result != nil {
			response.OK(c, result)
			return
		}
		response.InternalError(c, "Diagnostic read failed: "+err.Error())
		return
	}

	response.OK(c, result)
}
