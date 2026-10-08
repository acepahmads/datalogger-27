package handler

import (
	"strconv"
	"strings"

	"datalogger/internal/communication"
	"datalogger/internal/model"
	"datalogger/internal/repository"
	"datalogger/internal/service"
	"datalogger/pkg/formula"
	"datalogger/pkg/response"

	"github.com/gin-gonic/gin"
)

type DeviceHandler struct {
	deviceService *service.DeviceService
	pollingEngine *communication.PollingEngine
}

func NewDeviceHandler(deviceService *service.DeviceService, pollingEngine ...*communication.PollingEngine) *DeviceHandler {
	h := &DeviceHandler{deviceService: deviceService}
	if len(pollingEngine) > 0 {
		h.pollingEngine = pollingEngine[0]
	}
	return h
}

func (h *DeviceHandler) SetPollingEngine(pe *communication.PollingEngine) {
	h.pollingEngine = pe
}

// ListDevices handles GET /api/devices with pagination, search, and filtering
func (h *DeviceHandler) ListDevices(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if lStr := c.Query("limit"); lStr != "" && pageSize == 20 {
		if l, err := strconv.Atoi(lStr); err == nil {
			pageSize = l
		}
	}

	search := c.Query("search")
	status := c.Query("status")
	connStatus := c.Query("connection_status")
	devType := c.Query("device_type")
	protocol := c.Query("protocol")
	sortBy := c.DefaultQuery("sort_by", "id")
	sortDir := c.DefaultQuery("sort_dir", "desc")

	var enabledPtr *bool
	if enStr := c.Query("enabled"); enStr != "" {
		if b, err := strconv.ParseBool(enStr); err == nil {
			enabledPtr = &b
		}
	}

	params := repository.DeviceFilterParams{
		Search:           search,
		Status:           status,
		ConnectionStatus: connStatus,
		DeviceType:       devType,
		Protocol:         protocol,
		Enabled:          enabledPtr,
		SortBy:           sortBy,
		SortDir:          sortDir,
		Page:             page,
		PageSize:         pageSize,
	}

	devices, total, err := h.deviceService.ListDevices(params)
	if err != nil {
		response.InternalError(c, "Failed to retrieve devices: "+err.Error())
		return
	}

	// If page_size is 0 or all items requested, return array directly for legacy clients
	if c.Query("all") == "true" || pageSize <= 0 {
		response.OK(c, devices)
		return
	}

	response.OK(c, response.PaginatedData{
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		Items:    devices,
	})
}

// GetDeviceByID handles GET /api/devices/:id
func (h *DeviceHandler) GetDeviceByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid device ID")
		return
	}

	device, err := h.deviceService.GetDeviceByID(uint(id))
	if err != nil {
		response.NotFound(c, "Device not found")
		return
	}

	response.OK(c, device)
}

// CreateDevice handles POST /api/devices
func (h *DeviceHandler) CreateDevice(c *gin.Context) {
	var req service.CreateDeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request payload: "+err.Error())
		return
	}

	username := h.getUsername(c)
	ip := c.ClientIP()
	ua := c.Request.UserAgent()

	device, err := h.deviceService.CreateDevice(&req, username, ip, ua)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if h.pollingEngine != nil && device.Enabled && device.Status == model.DeviceStatusActive {
		h.pollingEngine.StartDeviceWorker(device)
	}

	response.Created(c, device)
}

// UpdateDevice handles PUT /api/devices/:id
func (h *DeviceHandler) UpdateDevice(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid device ID")
		return
	}

	var req service.UpdateDeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid update payload: "+err.Error())
		return
	}

	username := h.getUsername(c)
	ip := c.ClientIP()
	ua := c.Request.UserAgent()

	device, err := h.deviceService.UpdateDevice(uint(id), &req, username, ip, ua)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if h.pollingEngine != nil {
		if device.Enabled && device.Status == model.DeviceStatusActive {
			h.pollingEngine.StartDeviceWorker(device)
		} else {
			h.pollingEngine.StopDeviceWorker(device.ID)
		}
	}

	response.OK(c, device)
}

// DeleteDevice handles DELETE /api/devices/:id
func (h *DeviceHandler) DeleteDevice(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid device ID")
		return
	}

	username := h.getUsername(c)
	ip := c.ClientIP()
	ua := c.Request.UserAgent()

	if err := h.deviceService.DeleteDevice(uint(id), username, ip, ua); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if h.pollingEngine != nil {
		h.pollingEngine.StopDeviceWorker(uint(id))
	}

	response.OK(c, gin.H{"deleted": true, "id": id})
}

// ToggleDeviceEnabled handles PUT /api/devices/:id/enable
func (h *DeviceHandler) ToggleDeviceEnabled(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid device ID")
		return
	}

	var req struct {
		Enabled bool `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid payload")
		return
	}

	username := h.getUsername(c)
	ip := c.ClientIP()
	ua := c.Request.UserAgent()

	if err := h.deviceService.ToggleDeviceEnabled(uint(id), req.Enabled, username, ip, ua); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if h.pollingEngine != nil {
		if dev, err := h.deviceService.GetDeviceByID(uint(id)); err == nil {
			if dev.Enabled && dev.Status == model.DeviceStatusActive {
				h.pollingEngine.StartDeviceWorker(dev)
			} else {
				h.pollingEngine.StopDeviceWorker(dev.ID)
			}
		}
	}

	response.OK(c, gin.H{"id": id, "enabled": req.Enabled})
}

// UpdateDeviceStatus handles PUT /api/devices/:id/status
func (h *DeviceHandler) UpdateDeviceStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid device ID")
		return
	}

	var req struct {
		Status model.DeviceAdminStatus `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid payload")
		return
	}

	username := h.getUsername(c)
	ip := c.ClientIP()
	ua := c.Request.UserAgent()

	if err := h.deviceService.UpdateDeviceStatus(uint(id), req.Status, username, ip, ua); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if h.pollingEngine != nil {
		if dev, err := h.deviceService.GetDeviceByID(uint(id)); err == nil {
			if dev.Enabled && dev.Status == model.DeviceStatusActive {
				h.pollingEngine.StartDeviceWorker(dev)
			} else {
				h.pollingEngine.StopDeviceWorker(dev.ID)
			}
		}
	}

	response.OK(c, gin.H{"id": id, "status": req.Status})
}

// UpdateConnection handles PUT /api/devices/:id/connection
func (h *DeviceHandler) UpdateConnection(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid device ID")
		return
	}

	var req service.ConnectionConfigDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid connection payload: "+err.Error())
		return
	}

	username := h.getUsername(c)
	ip := c.ClientIP()
	ua := c.Request.UserAgent()

	conn, err := h.deviceService.UpdateConnection(uint(id), &req, username, ip, ua)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if h.pollingEngine != nil {
		if dev, err := h.deviceService.GetDeviceByID(uint(id)); err == nil && dev.Enabled && dev.Status == model.DeviceStatusActive {
			h.pollingEngine.StartDeviceWorker(dev)
		}
	}

	response.OK(c, conn)
}

// GetDeviceActivity handles GET /api/devices/:id/activity
func (h *DeviceHandler) GetDeviceActivity(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid device ID")
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	trails, err := h.deviceService.GetDeviceAuditTrail(uint(id), limit)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.OK(c, trails)
}

// Parameters
func (h *DeviceHandler) ListParameters(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid device ID")
		return
	}

	params, err := h.deviceService.ListParameters(uint(id))
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.OK(c, params)
}

func (h *DeviceHandler) GetParameter(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid device ID")
		return
	}
	paramID, err := strconv.ParseUint(c.Param("paramId"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid parameter ID")
		return
	}

	param, err := h.deviceService.GetParameter(uint(id), uint(paramID))
	if err != nil {
		response.NotFound(c, "Parameter not found")
		return
	}

	response.OK(c, param)
}

func (h *DeviceHandler) CreateParameter(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid device ID")
		return
	}

	var req service.CreateParameterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid parameter payload: "+err.Error())
		return
	}

	username := h.getUsername(c)
	ip := c.ClientIP()
	ua := c.Request.UserAgent()

	param, err := h.deviceService.CreateParameter(uint(id), &req, username, ip, ua)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if h.pollingEngine != nil {
		if dev, err := h.deviceService.GetDeviceByID(uint(id)); err == nil && dev.Enabled && dev.Status == model.DeviceStatusActive {
			h.pollingEngine.StartDeviceWorker(dev)
		}
	}

	response.Created(c, param)
}

func (h *DeviceHandler) UpdateParameter(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid device ID")
		return
	}
	paramID, err := strconv.ParseUint(c.Param("paramId"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid parameter ID")
		return
	}

	var req service.UpdateParameterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid parameter payload: "+err.Error())
		return
	}

	username := h.getUsername(c)
	ip := c.ClientIP()
	ua := c.Request.UserAgent()

	param, err := h.deviceService.UpdateParameter(uint(id), uint(paramID), &req, username, ip, ua)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if h.pollingEngine != nil {
		if dev, err := h.deviceService.GetDeviceByID(uint(id)); err == nil && dev.Enabled && dev.Status == model.DeviceStatusActive {
			h.pollingEngine.StartDeviceWorker(dev)
		}
	}

	response.OK(c, param)
}

func (h *DeviceHandler) DeleteParameter(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid device ID")
		return
	}
	paramID, err := strconv.ParseUint(c.Param("paramId"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid parameter ID")
		return
	}

	username := h.getUsername(c)
	ip := c.ClientIP()
	ua := c.Request.UserAgent()

	if err := h.deviceService.DeleteParameter(uint(id), uint(paramID), username, ip, ua); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if h.pollingEngine != nil {
		if dev, err := h.deviceService.GetDeviceByID(uint(id)); err == nil && dev.Enabled && dev.Status == model.DeviceStatusActive {
			h.pollingEngine.StartDeviceWorker(dev)
		}
	}

	response.OK(c, gin.H{"deleted": true, "device_id": id, "parameter_id": paramID})
}

func (h *DeviceHandler) ToggleParameterEnabled(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid device ID")
		return
	}
	paramID, err := strconv.ParseUint(c.Param("paramId"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid parameter ID")
		return
	}

	var req struct {
		Enabled bool `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid payload")
		return
	}

	username := h.getUsername(c)
	ip := c.ClientIP()
	ua := c.Request.UserAgent()

	if err := h.deviceService.ToggleParameterEnabled(uint(id), uint(paramID), req.Enabled, username, ip, ua); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if h.pollingEngine != nil {
		if dev, err := h.deviceService.GetDeviceByID(uint(id)); err == nil && dev.Enabled && dev.Status == model.DeviceStatusActive {
			h.pollingEngine.StartDeviceWorker(dev)
		}
	}

	response.OK(c, gin.H{"device_id": id, "parameter_id": paramID, "enabled": req.Enabled})
}

func (h *DeviceHandler) getUsername(c *gin.Context) string {
	if u, exists := c.Get("username"); exists {
		if uStr, ok := u.(string); ok && strings.TrimSpace(uStr) != "" {
			return uStr
		}
	}
	return "admin"
}

type ValidateFormulaRequest struct {
	Formula     string   `json:"formula"`
	SampleValue *float64 `json:"sample_value"`
	SampleRaw   *float64 `json:"sample_raw"`
}

func (h *DeviceHandler) ValidateFormula(c *gin.Context) {
	var req ValidateFormulaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	trimmed := strings.TrimSpace(req.Formula)
	if trimmed == "" {
		response.OK(c, gin.H{
			"valid":   true,
			"formula": "",
			"result":  nil,
		})
		return
	}

	val := 25.0
	if req.SampleValue != nil {
		val = *req.SampleValue
	}
	raw := 100.0
	if req.SampleRaw != nil {
		raw = *req.SampleRaw
	}

	vars := map[string]float64{
		"x":     val,
		"val":   val,
		"value": val,
		"raw":   raw,
	}

	res, err := formula.Evaluate(trimmed, vars)
	if err != nil {
		response.BadRequest(c, "Invalid formula: "+err.Error())
		return
	}

	response.OK(c, gin.H{
		"valid":   true,
		"formula": trimmed,
		"result":  res,
	})
}

