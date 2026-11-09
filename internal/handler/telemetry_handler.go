package handler

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"datalogger/internal/config"
	"datalogger/internal/middleware"
	"datalogger/internal/model"
	"datalogger/internal/repository"
	"datalogger/internal/service"
	"datalogger/pkg/response"

	"github.com/gin-gonic/gin"
)

// TelemetryHandler handles REST endpoints for instantaneous and historical telemetry data
type TelemetryHandler struct {
	telemetryService *service.TelemetryService
	injectionService *service.TelemetryInjectionService
}

// NewTelemetryHandler constructs a new telemetry handler
func NewTelemetryHandler(telemetryService *service.TelemetryService) *TelemetryHandler {
	return &TelemetryHandler{telemetryService: telemetryService}
}

// SetInjectionService attaches the Superadmin-only telemetry injection service for badge annotation
func (h *TelemetryHandler) SetInjectionService(injService *service.TelemetryInjectionService) {
	h.injectionService = injService
}

func canAccessRawTelemetry(c *gin.Context) bool {
	roleVal, exists := c.Get("role")
	if !exists {
		return false
	}
	role, ok := roleVal.(string)
	return ok && (strings.EqualFold(role, "SUPERADMIN") || strings.EqualFold(role, "Administrator") || strings.EqualFold(role, "ADMIN"))
}

func sanitizeRawData(records []model.RawData) {
	for i := range records {
		records[i].RawValue = 0
		records[i].RawHex = ""
		records[i].RawBytes = nil
	}
}

func sanitizeLatest(items []service.LatestTelemetry) {
	for i := range items {
		items[i].RawValue = 0
		items[i].RawHex = ""
	}
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
		if strings.HasPrefix(err.Error(), "access denied") {
			response.Forbidden(c, err.Error())
			return
		}
		response.InternalError(c, "Failed to retrieve latest telemetry: "+err.Error())
		return
	}

	if !canAccessRawTelemetry(c) {
		sanitizeLatest(results)
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
		if strings.HasPrefix(err.Error(), "access denied") {
			response.Forbidden(c, err.Error())
			return
		}
		response.NotFound(c, "Parameter telemetry not found: "+err.Error())
		return
	}

	if !canAccessRawTelemetry(c) && result != nil {
		result.RawValue = 0
		result.RawHex = ""
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

	skipCountRaw := strings.TrimSpace(strings.ToLower(c.Query("skip_count")))
	scope := middleware.GetScope(c)
	filter := repository.TelemetryFilterParams{
		TenantID:       scope.TenantID,
		SiteID:         scope.SiteID,
		AllowedSiteIDs: scope.AllowedSiteIDs,
		Page:           page,
		PageSize:       pageSize,
		Quality:        strings.TrimSpace(c.Query("quality")),
		SkipCount:      skipCountRaw == "1" || skipCountRaw == "true",
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

// enforceRawWindowLimit validates and bounds the query window against the configured RawTelemetryMaxRangeHours.
func (h *TelemetryHandler) enforceRawWindowLimit(params *repository.TelemetryFilterParams) error {
	maxHours := config.Get().EffectiveRawTelemetryMaxRangeHours()
	maxDur := time.Duration(maxHours) * time.Hour
	grace := 5 * time.Minute

	if params.StartTime != nil && params.EndTime != nil {
		if params.EndTime.Before(*params.StartTime) {
			return fmt.Errorf("end_time must be greater than or equal to start_time")
		}
		if params.EndTime.Sub(*params.StartTime) > maxDur+grace {
			return fmt.Errorf("Requested time range exceeds maximum configured limit of %d hours", maxHours)
		}
		return nil
	}

	if params.StartTime != nil && params.EndTime == nil {
		if time.Now().UTC().Sub(*params.StartTime) > maxDur+grace {
			return fmt.Errorf("Requested time range exceeds maximum configured limit of %d hours", maxHours)
		}
		return nil
	}

	// Unbounded start_time: clamp to configured max window relative to end_time (or now)
	refEnd := time.Now().UTC()
	if params.EndTime != nil {
		refEnd = *params.EndTime
	}
	boundedStart := refEnd.Add(-maxDur)
	params.StartTime = &boundedStart
	return nil
}

// GetHistorical handles GET /api/devices/:id/telemetry/history
func (h *TelemetryHandler) GetHistorical(c *gin.Context) {
	devID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid device ID format")
		return
	}

	params := h.parseTelemetryFilterParams(c, uint(devID))
	if c.Query("raw_window") == "1" {
		if err := h.enforceRawWindowLimit(&params); err != nil {
			response.BadRequest(c, err.Error())
			return
		}
	}
	records, total, err := h.telemetryService.GetHistorical(c.Request.Context(), params)
	if err != nil {
		if strings.HasPrefix(err.Error(), "access denied") {
			response.Forbidden(c, err.Error())
			return
		}
		response.InternalError(c, "Failed to retrieve historical telemetry: "+err.Error())
		return
	}

	if !canAccessRawTelemetry(c) {
		sanitizeRawData(records)
	}
	if isStrictSuperAdmin(c) && h.injectionService != nil {
		h.injectionService.AnnotateRawDataForSuperadmin(c.Request.Context(), records)
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
	if c.Query("raw_window") == "1" {
		if err := h.enforceRawWindowLimit(&params); err != nil {
			response.BadRequest(c, err.Error())
			return
		}
	}
	records, total, err := h.telemetryService.GetHistorical(c.Request.Context(), params)
	if err != nil {
		if strings.HasPrefix(err.Error(), "access denied") {
			response.Forbidden(c, err.Error())
			return
		}
		response.InternalError(c, "Failed to retrieve historical telemetry: "+err.Error())
		return
	}

	if !canAccessRawTelemetry(c) {
		sanitizeRawData(records)
	}
	if isStrictSuperAdmin(c) && h.injectionService != nil {
		h.injectionService.AnnotateRawDataForSuperadmin(c.Request.Context(), records)
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
	if !canAccessRawTelemetry(c) {
		response.Forbidden(c, "Access denied: raw telemetry inspection requires Admin or Superadmin privileges")
		return
	}

	devID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid device ID format")
		return
	}

	params := h.parseTelemetryFilterParams(c, uint(devID))
	if err := h.enforceRawWindowLimit(&params); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	records, total, err := h.telemetryService.GetHistorical(c.Request.Context(), params)
	if err != nil {
		if strings.HasPrefix(err.Error(), "access denied") {
			response.Forbidden(c, err.Error())
			return
		}
		response.InternalError(c, "Failed to query raw telemetry: "+err.Error())
		return
	}

	if isStrictSuperAdmin(c) && h.injectionService != nil {
		h.injectionService.AnnotateRawDataForSuperadmin(c.Request.Context(), records)
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
	if !canAccessRawTelemetry(c) {
		response.Forbidden(c, "Access denied: raw telemetry inspection requires Admin or Superadmin privileges")
		return
	}

	params := h.parseTelemetryFilterParams(c, 0)
	if err := h.enforceRawWindowLimit(&params); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	records, total, err := h.telemetryService.GetHistorical(c.Request.Context(), params)
	if err != nil {
		if strings.HasPrefix(err.Error(), "access denied") {
			response.Forbidden(c, err.Error())
			return
		}
		response.InternalError(c, "Failed to query raw telemetry: "+err.Error())
		return
	}

	if isStrictSuperAdmin(c) && h.injectionService != nil {
		h.injectionService.AnnotateRawDataForSuperadmin(c.Request.Context(), records)
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
