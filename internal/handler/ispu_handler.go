package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"datalogger/internal/model"
	"datalogger/internal/repository"
	"datalogger/internal/service"
	"datalogger/pkg/response"

	"github.com/gin-gonic/gin"
)

// ISPUHandler handles HTTP requests for ISPU calculations, history, mappings, and export.
type ISPUHandler struct {
	ispuService *service.ISPUService
}

// NewISPUHandler creates a new ISPUHandler.
func NewISPUHandler(ispuService *service.ISPUService) *ISPUHandler {
	return &ISPUHandler{ispuService: ispuService}
}

// GetLatest returns the latest ISPU calculation summary and parameter breakdown for a station.
func (h *ISPUHandler) GetLatest(c *gin.Context) {
	var deviceID uint
	if deviceIDStr := c.Query("device_id"); deviceIDStr != "" {
		if id, err := strconv.ParseUint(deviceIDStr, 10, 32); err == nil {
			deviceID = uint(id)
		}
	}

	mode := c.DefaultQuery("mode", "REGULATORY")
	result, err := h.ispuService.GetLatestSummary(c.Request.Context(), uint(deviceID), mode)
	if err != nil {
		response.InternalError(c, "failed to get latest ISPU: "+err.Error())
		return
	}

	response.OK(c, result)
}

// GetTrend returns historical calculation results for time-series charts.
func (h *ISPUHandler) GetTrend(c *gin.Context) {
	var deviceID uint
	if deviceIDStr := c.Query("device_id"); deviceIDStr != "" {
		if id, err := strconv.ParseUint(deviceIDStr, 10, 32); err == nil {
			deviceID = uint(id)
		}
	}

	rangeStr := c.DefaultQuery("range", "24h")
	mode := c.DefaultQuery("mode", "REGULATORY")
	trend, err := h.ispuService.GetTrend(c.Request.Context(), uint(deviceID), rangeStr, mode)
	if err != nil {
		response.InternalError(c, "failed to get ISPU trend: "+err.Error())
		return
	}

	response.OK(c, trend)
}

// GetHistory returns paginated historical ISPU calculations with multi-criteria filtering.
func (h *ISPUHandler) GetHistory(c *gin.Context) {
	filter := repository.ISPUHistoryFilter{
		CalculationMode:   c.DefaultQuery("mode", "REGULATORY"),
		Category:          c.Query("category"),
		DominantPollutant: c.Query("dominant_pollutant"),
		Status:            c.Query("status"),
	}

	if devIDStr := c.Query("device_id"); devIDStr != "" {
		if id, err := strconv.ParseUint(devIDStr, 10, 32); err == nil {
			filter.DeviceID = uint(id)
		}
	}
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
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil {
			filter.Page = p
		}
	}
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			filter.PageSize = l
		}
	}

	records, total, err := h.ispuService.GetHistory(c.Request.Context(), filter)
	if err != nil {
		response.InternalError(c, "failed to get ISPU history: "+err.Error())
		return
	}

	response.OK(c, response.PaginatedData{
		Total:    total,
		Page:     filter.Page,
		PageSize: filter.PageSize,
		Items:    records,
	})
}

// GetDetail returns single calculation breakdown with full formulas.
func (h *ISPUHandler) GetDetail(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "invalid calculation ID")
		return
	}

	record, err := h.ispuService.GetCalculationDetail(c.Request.Context(), uint(id))
	if err != nil {
		response.NotFound(c, "ISPU calculation record not found")
		return
	}

	response.OK(c, record)
}

// GetAuditTracking returns outlier details, peak timing, and top spikes for an ISPU calculation.
func (h *ISPUHandler) GetAuditTracking(c *gin.Context) {
	calcIDStr := c.Query("calculation_result_id")
	if calcIDStr == "" {
		calcIDStr = c.Query("id")
	}
	if calcIDStr == "" {
		response.BadRequest(c, "calculation_result_id is required")
		return
	}

	calcID, err := strconv.ParseUint(calcIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "invalid calculation_result_id")
		return
	}

	pollutant := c.Query("pollutant")
	details, err := h.ispuService.GetAuditTracking(c.Request.Context(), uint(calcID), pollutant)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	response.OK(c, details)
}


// Recalculate triggers on-demand recalculation for a station and time window.
func (h *ISPUHandler) Recalculate(c *gin.Context) {
	var req struct {
		DeviceID   uint   `json:"device_id"`
		TargetTime string `json:"target_time"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid recalculate payload: "+err.Error())
		return
	}

	var targetTime *time.Time
	if req.TargetTime != "" {
		if t, err := time.Parse(time.RFC3339, req.TargetTime); err == nil {
			targetTime = &t
		}
	}

	actor := "operator"
	if username, exists := c.Get("username"); exists {
		actor = username.(string)
	}

	result, err := h.ispuService.Recalculate(c.Request.Context(), req.DeviceID, targetTime, actor, c.ClientIP())
	if err != nil {
		response.InternalError(c, "recalculation failed: "+err.Error())
		return
	}

	response.OK(c, result)
}

// GetMappings lists station parameter mappings.
func (h *ISPUHandler) GetMappings(c *gin.Context) {
	var deviceID uint
	if deviceIDStr := c.Query("device_id"); deviceIDStr != "" {
		if id, err := strconv.ParseUint(deviceIDStr, 10, 32); err == nil {
			deviceID = uint(id)
		}
	}

	mappings, err := h.ispuService.GetMappings(c.Request.Context(), deviceID)
	if err != nil {
		response.InternalError(c, "failed to get mappings: "+err.Error())
		return
	}

	response.OK(c, mappings)
}

// SaveMapping creates or updates parameter mapping.
func (h *ISPUHandler) SaveMapping(c *gin.Context) {
	var mapping model.ISPUParameterMapping
	if err := c.ShouldBindJSON(&mapping); err != nil {
		response.BadRequest(c, "invalid mapping payload: "+err.Error())
		return
	}

	if mapping.ParameterID == 0 || mapping.PollutantCode == "" {
		response.BadRequest(c, "parameter_id and pollutant_code are required")
		return
	}

	actor := "operator"
	if username, exists := c.Get("username"); exists {
		actor = username.(string)
	}

	if err := h.ispuService.SaveMapping(c.Request.Context(), &mapping, actor, c.ClientIP()); err != nil {
		response.InternalError(c, "failed to save mapping: "+err.Error())
		return
	}

	response.OK(c, mapping)
}

// DeleteMapping deletes a mapping.
func (h *ISPUHandler) DeleteMapping(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "invalid mapping ID")
		return
	}

	actor := "operator"
	if username, exists := c.Get("username"); exists {
		actor = username.(string)
	}

	if err := h.ispuService.DeleteMapping(c.Request.Context(), uint(id), actor, c.ClientIP()); err != nil {
		response.InternalError(c, "failed to delete mapping: "+err.Error())
		return
	}

	response.Message(c, "mapping deleted successfully", nil)
}

// ExportCSV exports ISPU records to CSV format.
func (h *ISPUHandler) ExportCSV(c *gin.Context) {
	filter := repository.ISPUHistoryFilter{
		CalculationMode:   c.DefaultQuery("mode", "REGULATORY"),
		Category:          c.Query("category"),
		DominantPollutant: c.Query("dominant_pollutant"),
		Status:            c.Query("status"),
	}
	if devIDStr := c.Query("device_id"); devIDStr != "" {
		if id, err := strconv.ParseUint(devIDStr, 10, 32); err == nil {
			filter.DeviceID = uint(id)
		}
	}

	csvData, err := h.ispuService.ExportCSV(c.Request.Context(), filter)
	if err != nil {
		response.InternalError(c, "failed to export CSV: "+err.Error())
		return
	}

	filename := fmt.Sprintf("ispu_export_%s.csv", time.Now().Format("20060102_150405"))
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Data(http.StatusOK, "text/csv; charset=utf-8", csvData)
}

// ExportExcel exports ISPU records to Excel-compatible CSV format with UTF-8 BOM.
func (h *ISPUHandler) ExportExcel(c *gin.Context) {
	filter := repository.ISPUHistoryFilter{
		CalculationMode:   c.DefaultQuery("mode", "REGULATORY"),
		Category:          c.Query("category"),
		DominantPollutant: c.Query("dominant_pollutant"),
		Status:            c.Query("status"),
	}
	if devIDStr := c.Query("device_id"); devIDStr != "" {
		if id, err := strconv.ParseUint(devIDStr, 10, 32); err == nil {
			filter.DeviceID = uint(id)
		}
	}

	excelData, err := h.ispuService.ExportExcel(c.Request.Context(), filter)
	if err != nil {
		response.InternalError(c, "failed to export Excel: "+err.Error())
		return
	}

	filename := fmt.Sprintf("ispu_export_%s.csv", time.Now().Format("20060102_150405"))
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Data(http.StatusOK, "text/csv; charset=utf-8", excelData)
}

// Backfill triggers historical trend backfill for a station over lookback hours.
func (h *ISPUHandler) Backfill(c *gin.Context) {
	var req struct {
		DeviceID uint `json:"device_id"`
		Hours    int  `json:"hours"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request: "+err.Error())
		return
	}
	if req.Hours <= 0 {
		req.Hours = 24
	}

	count, err := h.ispuService.BackfillStationHourlyTrend(c.Request.Context(), req.DeviceID, req.Hours)
	if err != nil {
		response.InternalError(c, "failed to backfill trend: "+err.Error())
		return
	}

	response.OK(c, gin.H{
		"device_id":          req.DeviceID,
		"hours":              req.Hours,
		"backfilled_records": count,
	})
}

