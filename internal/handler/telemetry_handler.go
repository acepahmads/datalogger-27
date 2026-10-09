package handler

import (
	"strconv"
	"strings"
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

func (h *TelemetryHandler) parseTelemetryFilterParams(c *gin.Context, defaultDevID uint) repository.TelemetryFilterParams {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "50"))
	if lStr := c.Query("limit"); lStr != "" && pageSize == 50 {
		if l, err := strconv.Atoi(lStr); err == nil {
			pageSize = l
		}
	}

	filter := repository.TelemetryFilterParams{
		Page:     page,
		PageSize: pageSize,
		Quality:  strings.TrimSpace(c.Query("quality")),
	}

	// Device ID & Device IDs
	if defaultDevID > 0 {
		filter.DeviceID = defaultDevID
	} else {
		if dStr := c.Query("device_id"); dStr != "" {
			if dVal, err := strconv.ParseUint(dStr, 10, 32); err == nil {
				filter.DeviceID = uint(dVal)
			}
		}
		if dIDsStr := c.Query("device_ids"); dIDsStr != "" {
			parts := strings.Split(dIDsStr, ",")
			for _, p := range parts {
				if v, err := strconv.ParseUint(strings.TrimSpace(p), 10, 32); err == nil {
					filter.DeviceIDs = append(filter.DeviceIDs, uint(v))
				}
			}
		}
	}

	// Parameter ID & Parameter IDs
	if pStr := c.Query("parameter_id"); pStr != "" {
		if pVal, err := strconv.ParseUint(pStr, 10, 32); err == nil {
			filter.ParameterID = uint(pVal)
		}
	}
	if pIDsStr := c.Query("parameter_ids"); pIDsStr != "" {
		parts := strings.Split(pIDsStr, ",")
		for _, p := range parts {
			if v, err := strconv.ParseUint(strings.TrimSpace(p), 10, 32); err == nil {
				filter.ParameterIDs = append(filter.ParameterIDs, uint(v))
			}
		}
	}

	// Time range
	if startStr := c.Query("start_time"); startStr != "" {
		if t, err := time.Parse(time.RFC3339, startStr); err == nil {
			filter.StartTime = &t
		}
	}
	if endStr := c.Query("end_time"); endStr != "" {
		if t, err := time.Parse(time.RFC3339, endStr); err == nil {
			filter.EndTime = &t
		}
	}

	return filter
}

// GetHistorical handles GET /api/devices/:id/telemetry/history
func (h *TelemetryHandler) GetHistorical(c *gin.Context) {
	devID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid device ID format")
		return
	}

	params := h.parseTelemetryFilterParams(c, uint(devID))
	records, total, err := h.telemetryService.GetHistorical(c.Request.Context(), params)
	if err != nil {
		response.InternalError(c, "Failed to retrieve historical telemetry: "+err.Error())
		return
	}

	response.OK(c, response.PaginatedData{
		Total:    total,
		Page:     params.Page,
		PageSize: params.PageSize,
		Items:    records,
	})
}

// GetAllHistorical handles GET /api/telemetry/history for unified multi-device analysis
func (h *TelemetryHandler) GetAllHistorical(c *gin.Context) {
	params := h.parseTelemetryFilterParams(c, 0)
	records, total, err := h.telemetryService.GetHistorical(c.Request.Context(), params)
	if err != nil {
		response.InternalError(c, "Failed to retrieve historical telemetry: "+err.Error())
		return
	}

	response.OK(c, response.PaginatedData{
		Total:    total,
		Page:     params.Page,
		PageSize: params.PageSize,
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

	params := h.parseTelemetryFilterParams(c, uint(devID))
	records, total, err := h.telemetryService.GetHistorical(c.Request.Context(), params)
	if err != nil {
		response.InternalError(c, "Failed to query raw telemetry: "+err.Error())
		return
	}

	response.OK(c, response.PaginatedData{
		Total:    total,
		Page:     params.Page,
		PageSize: params.PageSize,
		Items:    records,
	})
}

// GetAllRawTelemetry handles GET /api/telemetry/raw for unified multi-device raw telemetry inspection
func (h *TelemetryHandler) GetAllRawTelemetry(c *gin.Context) {
	params := h.parseTelemetryFilterParams(c, 0)
	records, total, err := h.telemetryService.GetHistorical(c.Request.Context(), params)
	if err != nil {
		response.InternalError(c, "Failed to query raw telemetry: "+err.Error())
		return
	}

	response.OK(c, response.PaginatedData{
		Total:    total,
		Page:     params.Page,
		PageSize: params.PageSize,
		Items:    records,
	})
}

// GetMetrics handles GET /api/telemetry/metrics for buffer and throughput diagnostics
func (h *TelemetryHandler) GetMetrics(c *gin.Context) {
	metrics := h.telemetryService.GetMetrics()
	response.OK(c, metrics)
}

// GetQualitySummary handles GET /api/telemetry/quality-summary
func (h *TelemetryHandler) GetQualitySummary(c *gin.Context) {
	var devID *uint
	if dStr := c.Query("device_id"); dStr != "" {
		if idVal, err := strconv.ParseUint(dStr, 10, 32); err == nil {
			uID := uint(idVal)
			devID = &uID
		}
	}

	summary, err := h.telemetryService.GetQualitySummary(c.Request.Context(), devID)
	if err != nil {
		response.InternalError(c, "Failed to retrieve quality summary: "+err.Error())
		return
	}

	response.OK(c, summary)
}

// GetDeviceQualitySummary handles GET /api/devices/:id/telemetry/quality-summary
func (h *TelemetryHandler) GetDeviceQualitySummary(c *gin.Context) {
	idVal, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid device ID format")
		return
	}
	uID := uint(idVal)

	summary, err := h.telemetryService.GetQualitySummary(c.Request.Context(), &uID)
	if err != nil {
		response.InternalError(c, "Failed to retrieve device quality summary: "+err.Error())
		return
	}

	response.OK(c, summary)
}
