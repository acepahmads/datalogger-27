package handler

import (
	"strconv"
	"strings"

	"datalogger/internal/repository"
	"datalogger/internal/service"
	"datalogger/pkg/response"

	"github.com/gin-gonic/gin"
)

type CalibrationHandler struct {
	calService *service.CalibrationService
}

func NewCalibrationHandler(calService *service.CalibrationService) *CalibrationHandler {
	return &CalibrationHandler{calService: calService}
}

func (h *CalibrationHandler) extractActor(c *gin.Context) (username, role string) {
	if u, ok := c.Get("username"); ok {
		if s, ok2 := u.(string); ok2 && strings.TrimSpace(s) != "" {
			username = s
		}
	}
	if username == "" {
		username = "system"
	}
	if r, ok := c.Get("role"); ok {
		if s, ok2 := r.(string); ok2 {
			role = strings.ToUpper(strings.TrimSpace(s))
		}
	}
	return username, role
}

// ListParameters handles GET /api/calibrations/parameters
func (h *CalibrationHandler) ListParameters(c *gin.Context) {
	devID, _ := strconv.ParseUint(c.Query("device_id"), 10, 32)
	domain := c.Query("domain")

	items, stationDomain, err := h.calService.ListCalibratableParameters(c.Request.Context(), uint(devID), domain)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.OK(c, gin.H{
		"station_domain": stationDomain,
		"items":          items,
	})
}

// GetLiveReading handles GET /api/calibrations/parameters/:paramId/live
func (h *CalibrationHandler) GetLiveReading(c *gin.Context) {
	paramID, err := strconv.ParseUint(c.Param("paramId"), 10, 32)
	if err != nil || paramID == 0 {
		response.BadRequest(c, "ID parameter tidak valid")
		return
	}
	pollNow := strings.EqualFold(c.Query("poll"), "true") || c.Query("poll") == "1"

	live, err := h.calService.GetLiveParameterReading(c.Request.Context(), uint(paramID), pollNow)
	if err != nil {
		if strings.HasPrefix(err.Error(), "akses ditolak") {
			response.Forbidden(c, err.Error())
			return
		}
		response.NotFound(c, err.Error())
		return
	}
	response.OK(c, live)
}

// CalculatePreview handles POST /api/calibrations/calculate
func (h *CalibrationHandler) CalculatePreview(c *gin.Context) {
	var req service.CalculateCalibrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Payload perhitungan tidak valid: "+err.Error())
		return
	}
	res, err := h.calService.CalculatePreview(&req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, res)
}

// UpdateConfig handles PUT /api/calibrations/parameters/:paramId/config
func (h *CalibrationHandler) UpdateConfig(c *gin.Context) {
	paramID, err := strconv.ParseUint(c.Param("paramId"), 10, 32)
	if err != nil || paramID == 0 {
		response.BadRequest(c, "ID parameter tidak valid")
		return
	}

	var req service.UpdateCalibrationConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Payload konfigurasi tidak valid: "+err.Error())
		return
	}

	username, role := h.extractActor(c)
	dto, err := h.calService.UpdateParameterConfig(c.Request.Context(), uint(paramID), &req, username, role, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		if strings.HasPrefix(err.Error(), "akses ditolak") {
			response.Forbidden(c, err.Error())
			return
		}
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, dto)
}

// ToggleSession handles PUT /api/calibrations/parameters/:paramId/session
func (h *CalibrationHandler) ToggleSession(c *gin.Context) {
	paramID, err := strconv.ParseUint(c.Param("paramId"), 10, 32)
	if err != nil || paramID == 0 {
		response.BadRequest(c, "ID parameter tidak valid")
		return
	}

	var req service.ToggleCalibrationSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Payload sesi kalibrasi tidak valid: "+err.Error())
		return
	}

	username, role := h.extractActor(c)
	dto, err := h.calService.ToggleCalibrationSession(c.Request.Context(), uint(paramID), req.Active, username, role, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		if strings.HasPrefix(err.Error(), "akses ditolak") {
			response.Forbidden(c, err.Error())
			return
		}
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, dto)
}

// ApplyCalibration handles POST /api/calibrations/parameters/:paramId/apply
func (h *CalibrationHandler) ApplyCalibration(c *gin.Context) {
	paramID, err := strconv.ParseUint(c.Param("paramId"), 10, 32)
	if err != nil || paramID == 0 {
		response.BadRequest(c, "ID parameter tidak valid")
		return
	}

	var req service.ApplyCalibrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Payload kalibrasi tidak valid: "+err.Error())
		return
	}

	username, role := h.extractActor(c)
	dto, logEntry, err := h.calService.ApplyCalibration(c.Request.Context(), uint(paramID), &req, username, role, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		if strings.HasPrefix(err.Error(), "akses ditolak") {
			response.Forbidden(c, err.Error())
			return
		}
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, gin.H{
		"parameter": dto,
		"log":       logEntry,
	})
}

// ResetCalibration handles POST /api/calibrations/parameters/:paramId/reset
func (h *CalibrationHandler) ResetCalibration(c *gin.Context) {
	paramID, err := strconv.ParseUint(c.Param("paramId"), 10, 32)
	if err != nil || paramID == 0 {
		response.BadRequest(c, "ID parameter tidak valid")
		return
	}

	var req service.ResetCalibrationRequest
	_ = c.ShouldBindJSON(&req)

	username, role := h.extractActor(c)
	dto, logEntry, err := h.calService.ResetCalibration(c.Request.Context(), uint(paramID), req.ReasonNote, username, role, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		if strings.HasPrefix(err.Error(), "akses ditolak") {
			response.Forbidden(c, err.Error())
			return
		}
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, gin.H{
		"parameter": dto,
		"log":       logEntry,
	})
}

// ListLogs handles GET /api/calibrations/logs
func (h *CalibrationHandler) ListLogs(c *gin.Context) {
	devID, _ := strconv.ParseUint(c.Query("device_id"), 10, 32)
	paramID, _ := strconv.ParseUint(c.Query("parameter_id"), 10, 32)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "50"))

	logs, total, err := h.calService.ListCalibrationLogs(c.Request.Context(), repository.CalibrationLogFilter{
		DeviceID:    uint(devID),
		ParameterID: uint(paramID),
		Domain:      c.Query("domain"),
		Action:      c.Query("action"),
		Page:        page,
		PageSize:    pageSize,
	})
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.OK(c, gin.H{
		"items":     logs,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// RollbackLog handles POST /api/calibrations/logs/:logId/rollback
func (h *CalibrationHandler) RollbackLog(c *gin.Context) {
	logID, err := strconv.ParseUint(c.Param("logId"), 10, 32)
	if err != nil || logID == 0 {
		response.BadRequest(c, "ID log kalibrasi tidak valid")
		return
	}

	username, role := h.extractActor(c)
	dto, rbLog, err := h.calService.RollbackCalibration(c.Request.Context(), uint(logID), username, role, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		if strings.HasPrefix(err.Error(), "akses ditolak") {
			response.Forbidden(c, err.Error())
			return
		}
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, gin.H{
		"parameter": dto,
		"log":       rbLog,
	})
}

