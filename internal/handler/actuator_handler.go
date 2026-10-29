package handler

import (
	"fmt"
	"strconv"

	"datalogger/internal/model"
	"datalogger/internal/service"
	"datalogger/pkg/response"

	"github.com/gin-gonic/gin"
)

type ActuatorHandler struct {
	actuationService *service.ActuationService
}

func NewActuatorHandler(actuationService *service.ActuationService) *ActuatorHandler {
	return &ActuatorHandler{actuationService: actuationService}
}

// ListByDevice handles GET /api/devices/:id/actuators
func (h *ActuatorHandler) ListByDevice(c *gin.Context) {
	deviceID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid device ID")
		return
	}

	actuators, err := h.actuationService.ListActuatorsByDevice(uint(deviceID))
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.OK(c, actuators)
}

// ListAllActive handles GET /api/actuators
func (h *ActuatorHandler) ListAllActive(c *gin.Context) {
	actuators, err := h.actuationService.ListAllActiveActuators()
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.OK(c, actuators)
}

// GetByID handles GET /api/actuators/:id
func (h *ActuatorHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid actuator ID")
		return
	}

	actuator, err := h.actuationService.GetActuator(uint(id))
	if err != nil {
		response.NotFound(c, "Actuator not found")
		return
	}

	response.OK(c, actuator)
}

// Create handles POST /api/devices/:id/actuators
func (h *ActuatorHandler) Create(c *gin.Context) {
	deviceID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid device ID")
		return
	}

	var actuator model.DeviceActuator
	if err := c.ShouldBindJSON(&actuator); err != nil {
		response.BadRequest(c, "Invalid payload: "+err.Error())
		return
	}

	actuator.DeviceID = uint(deviceID)
	if err := h.actuationService.CreateActuator(&actuator); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Created(c, actuator)
}

// Update handles PUT /api/actuators/:id
func (h *ActuatorHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid actuator ID")
		return
	}

	var actuator model.DeviceActuator
	if err := c.ShouldBindJSON(&actuator); err != nil {
		response.BadRequest(c, "Invalid payload: "+err.Error())
		return
	}

	actuator.ID = uint(id)
	if err := h.actuationService.UpdateActuator(&actuator); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.OK(c, actuator)
}

// Delete handles DELETE /api/actuators/:id
func (h *ActuatorHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid actuator ID")
		return
	}

	if err := h.actuationService.DeleteActuator(uint(id)); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.OK(c, gin.H{"deleted": true, "id": id})
}

// TestWriteExisting handles POST /api/actuators/:id/test-write
func (h *ActuatorHandler) TestWriteExisting(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid actuator ID")
		return
	}

	var req service.ActuatorTestWriteRequest
	_ = c.ShouldBindJSON(&req)
	req.ActuatorID = uint(id)

	res, err := h.actuationService.TestWrite(c.Request.Context(), req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.OK(c, res)
}

// TestWriteRaw handles POST /api/actuators/test-write (for probing before saving)
func (h *ActuatorHandler) TestWriteRaw(c *gin.Context) {
	var req service.ActuatorTestWriteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid test write request: "+err.Error())
		return
	}

	res, err := h.actuationService.TestWrite(c.Request.Context(), req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.OK(c, res)
}

// Actuate handles POST /api/actuators/:id/actuate (Manual Start/Stop)
func (h *ActuatorHandler) Actuate(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid actuator ID")
		return
	}

	var body struct {
		Action string `json:"action" binding:"required"` // "ON" or "OFF"
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "Action 'ON' or 'OFF' is required")
		return
	}

	username := c.GetString("username")
	if username == "" {
		username = "operator"
	}
	triggeredBy := fmt.Sprintf("MANUAL: %s", username)

	action := model.ActuationAction(body.Action)
	logEntry, err := h.actuationService.Actuate(c.Request.Context(), uint(id), action, triggeredBy, nil)
	if err != nil {
		response.BadRequest(c, fmt.Sprintf("Actuation failed: %v", err))
		return
	}

	response.OK(c, logEntry)
}

// -------------------------------------------------------------
// Schedule Handlers
// -------------------------------------------------------------

// ListSchedules handles GET /api/actuators/:id/schedules
func (h *ActuatorHandler) ListSchedules(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid actuator ID")
		return
	}

	schedules, err := h.actuationService.ListSchedulesByActuator(uint(id))
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.OK(c, schedules)
}

// CreateSchedule handles POST /api/actuators/:id/schedules
func (h *ActuatorHandler) CreateSchedule(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid actuator ID")
		return
	}

	var schedule model.ActuationSchedule
	if err := c.ShouldBindJSON(&schedule); err != nil {
		response.BadRequest(c, "Invalid schedule data: "+err.Error())
		return
	}

	schedule.ActuatorID = uint(id)
	if err := h.actuationService.CreateSchedule(&schedule); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Created(c, schedule)
}

// UpdateSchedule handles PUT /api/schedules/:id
func (h *ActuatorHandler) UpdateSchedule(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid schedule ID")
		return
	}

	var schedule model.ActuationSchedule
	if err := c.ShouldBindJSON(&schedule); err != nil {
		response.BadRequest(c, "Invalid schedule data: "+err.Error())
		return
	}

	schedule.ID = uint(id)
	if err := h.actuationService.UpdateSchedule(&schedule); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.OK(c, schedule)
}

// DeleteSchedule handles DELETE /api/schedules/:id
func (h *ActuatorHandler) DeleteSchedule(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid schedule ID")
		return
	}

	if err := h.actuationService.DeleteSchedule(uint(id)); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.OK(c, gin.H{"deleted": true, "id": id})
}

// ListLogs handles GET /api/actuators/logs
func (h *ActuatorHandler) ListLogs(c *gin.Context) {
	var actuatorIDPtr *uint
	if actStr := c.Query("actuator_id"); actStr != "" {
		if id, err := strconv.ParseUint(actStr, 10, 32); err == nil {
			uID := uint(id)
			actuatorIDPtr = &uID
		}
	}

	var deviceIDPtr *uint
	if devStr := c.Query("device_id"); devStr != "" {
		if id, err := strconv.ParseUint(devStr, 10, 32); err == nil {
			uID := uint(id)
			deviceIDPtr = &uID
		}
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	logs, err := h.actuationService.ListLogs(actuatorIDPtr, deviceIDPtr, limit)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.OK(c, logs)
}
