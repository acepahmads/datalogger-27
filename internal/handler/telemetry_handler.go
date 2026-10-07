package handler

import (
	"strconv"
	"time"

	"datalogger/internal/repository"
	"datalogger/internal/service"
	"datalogger/pkg/response"

	"github.com/gin-gonic/gin"
)

// TelemetryHandler handles REST endpoints for instantaneous and historical telemetry data
type TelemetryHandler struct {
	telemetryService *service.TelemetryService
}

// NewTelemetryHandler constructs a new telemetry handler
func NewTelemetryHandler(telemetryService *service.TelemetryService) *TelemetryHandler {
	return &TelemetryHandler{telemetryService: telemetryService}
}

// GetLatestForDevice handles GET /api/devices/:id/telemetry/latest
func (h *TelemetryHandler) GetLatestForDevice(c *gin.Context) {
	idVal, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid device ID format")
		return
	}

	results, err := h.telemetryService.GetLatestForDevice(c.Request.Context(), uint(idVal))
	if err != nil {
		response.InternalError(c, "Failed to retrieve latest telemetry: "+err.Error())
		return
	}

	response.OK(c, results)
}

// GetLatestForParameter handles GET /api/devices/:id/parameters/:paramId/telemetry/latest
func (h *TelemetryHandler) GetLatestForParameter(c *gin.Context) {
	devID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid device ID format")
		return
	}

	paramID, err := strconv.ParseUint(c.Param("paramId"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid parameter ID format")
		return
	}

	result, err := h.telemetryService.GetLatestForParameter(c.Request.Context(), uint(devID), uint(paramID))
	if err != nil {
		response.NotFound(c, "Parameter telemetry not found: "+err.Error())
		return
	}

	response.OK(c, result)
}

// GetHistorical handles GET /api/devices/:id/telemetry/history
func (h *TelemetryHandler) GetHistorical(c *gin.Context) {
	devID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid device ID format")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "50"))
	if lStr := c.Query("limit"); lStr != "" && pageSize == 50 {
		if l, err := strconv.Atoi(lStr); err == nil {
			pageSize = l
		}
	}

	var paramID uint
	if pStr := c.Query("parameter_id"); pStr != "" {
		if pVal, err := strconv.ParseUint(pStr, 10, 32); err == nil {
			paramID = uint(pVal)
		}
	}

	quality := c.Query("quality")

	var startTime *time.Time
	if startStr := c.Query("start_time"); startStr != "" {
		if t, err := time.Parse(time.RFC3339, startStr); err == nil {
			startTime = &t
		}
	}

	var endTime *time.Time
	if endStr := c.Query("end_time"); endStr != "" {
		if t, err := time.Parse(time.RFC3339, endStr); err == nil {
			endTime = &t
		}
	}

	records, total, err := h.telemetryService.GetHistorical(c.Request.Context(), repository.TelemetryFilterParams{
		DeviceID:    uint(devID),
		ParameterID: paramID,
		Quality:     quality,
		StartTime:   startTime,
		EndTime:     endTime,
		Page:        page,
		PageSize:    pageSize,
	})
	if err != nil {
		response.InternalError(c, "Failed to retrieve historical telemetry: "+err.Error())
		return
	}

	response.OK(c, response.PaginatedData{
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		Items:    records,
	})
}

// GetRawTelemetry handles GET /api/devices/:id/telemetry/raw for diagnostic telemetry inspection
func (h *TelemetryHandler) GetRawTelemetry(c *gin.Context) {
	devID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid device ID format")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "50"))
	if lStr := c.Query("limit"); lStr != "" && pageSize == 50 {
		if l, err := strconv.Atoi(lStr); err == nil {
			pageSize = l
		}
	}

	var paramID uint
	if pStr := c.Query("parameter_id"); pStr != "" {
		if pVal, err := strconv.ParseUint(pStr, 10, 32); err == nil {
			paramID = uint(pVal)
		}
	}

	quality := c.Query("quality")

	var startTime *time.Time
	if startStr := c.Query("start_time"); startStr != "" {
		if t, err := time.Parse(time.RFC3339, startStr); err == nil {
			startTime = &t
		}
	}

	var endTime *time.Time
	if endStr := c.Query("end_time"); endStr != "" {
		if t, err := time.Parse(time.RFC3339, endStr); err == nil {
			endTime = &t
		}
	}

	records, total, err := h.telemetryService.GetHistorical(c.Request.Context(), repository.TelemetryFilterParams{
		DeviceID:    uint(devID),
		ParameterID: paramID,
		Quality:     quality,
		StartTime:   startTime,
		EndTime:     endTime,
		Page:        page,
		PageSize:    pageSize,
	})
	if err != nil {
		response.InternalError(c, "Failed to query raw telemetry: "+err.Error())
		return
	}

	response.OK(c, response.PaginatedData{
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		Items:    records,
	})
}

// GetMetrics handles GET /api/telemetry/metrics for buffer and throughput diagnostics
func (h *TelemetryHandler) GetMetrics(c *gin.Context) {
	metrics := h.telemetryService.GetMetrics()
	response.OK(c, metrics)
}
