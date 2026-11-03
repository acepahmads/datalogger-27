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

// WQMSHandler handles HTTP endpoints for WQMS Pollution Index (IP) and Water Quality Index (IKA).
type WQMSHandler struct {
	wqmsService *service.WQMSService
}

// NewWQMSHandler creates a new WQMSHandler.
func NewWQMSHandler(wqmsService *service.WQMSService) *WQMSHandler {
	return &WQMSHandler{wqmsService: wqmsService}
}

// GetLatest returns the latest WQMS IP calculation, IKA summary, active water class config, and standards.
func (h *WQMSHandler) GetLatest(c *gin.Context) {
	var deviceID uint
	if v := c.Query("device_id"); v != "" {
		if id, err := strconv.ParseUint(v, 10, 32); err == nil {
			deviceID = uint(id)
		}
	}
	mode := c.DefaultQuery("mode", "REGULATORY")
	bundle, err := h.wqmsService.GetLatestSummary(c.Request.Context(), deviceID, mode)
	if err != nil {
		response.InternalError(c, "failed to get latest WQMS IP & IKA: "+err.Error())
		return
	}
	response.OK(c, bundle)
}

// GetStandards returns the full PP No. 22 Tahun 2021 Class I-IV water quality standards reference matrix.
func (h *WQMSHandler) GetStandards(c *gin.Context) {
	response.OK(c, h.wqmsService.GetStandards())
}

// GetConfig returns the persisted station WQMS configuration.
func (h *WQMSHandler) GetConfig(c *gin.Context) {
	cfg, err := h.wqmsService.GetStationConfig(c.Request.Context(), 0)
	if err != nil {
		response.InternalError(c, "failed to get WQMS config: "+err.Error())
		return
	}
	response.OK(c, cfg)
}

// UpdateConfig updates the station water class (CLASS_1..CLASS_4) or IKA window and recalculates immediately.
func (h *WQMSHandler) UpdateConfig(c *gin.Context) {
	var req model.WQMSStationConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid WQMS config payload: "+err.Error())
		return
	}
	actor := "system"
	if u, ok := c.Get("username"); ok {
		if s, okStr := u.(string); okStr && s != "" {
			actor = s
		}
	}
	bundle, err := h.wqmsService.UpdateStationConfig(c.Request.Context(), req, actor, c.ClientIP())
	if err != nil {
		response.InternalError(c, "failed to update WQMS config: "+err.Error())
		return
	}
	response.OK(c, bundle)
}

// GetTrend returns chronological WQMS IP trend points.
func (h *WQMSHandler) GetTrend(c *gin.Context) {
	var deviceID uint
	if v := c.Query("device_id"); v != "" {
		if id, err := strconv.ParseUint(v, 10, 32); err == nil {
			deviceID = uint(id)
		}
	}
	rangeStr := c.DefaultQuery("range", "30d")
	mode := c.DefaultQuery("mode", "REGULATORY")
	trend, err := h.wqmsService.GetTrend(c.Request.Context(), deviceID, rangeStr, mode)
	if err != nil {
		response.InternalError(c, "failed to get WQMS trend: "+err.Error())
		return
	}
	response.OK(c, trend)
}

// GetHistory returns paginated historical WQMS IP records.
func (h *WQMSHandler) GetHistory(c *gin.Context) {
	filter := repository.WQMSHistoryFilter{
		CalculationMode:   c.DefaultQuery("mode", "REGULATORY"),
		WaterClass:        c.Query("water_class"),
		Category:          c.Query("category"),
		DominantParameter: c.Query("dominant_parameter"),
		Status:            c.Query("status"),
	}
	if v := c.Query("device_id"); v != "" {
		if id, err := strconv.ParseUint(v, 10, 32); err == nil {
			filter.DeviceID = uint(id)
		}
	}
	if s := c.Query("start_time"); s != "" {
		if t, err := time.Parse(time.RFC3339, s); err == nil {
			filter.StartTime = &t
		}
	}
	if s := c.Query("end_time"); s != "" {
		if t, err := time.Parse(time.RFC3339, s); err == nil {
			filter.EndTime = &t
		}
	}
	if p := c.Query("page"); p != "" {
		if n, err := strconv.Atoi(p); err == nil && n > 0 {
			filter.Page = n
		}
	}
	if l := c.Query("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 {
			filter.PageSize = n
		}
	}

	records, total, err := h.wqmsService.GetHistory(c.Request.Context(), filter)
	if err != nil {
		response.InternalError(c, "failed to get WQMS history: "+err.Error())
		return
	}
	response.OK(c, gin.H{
		"records": records,
		"total":   total,
		"page":    filter.Page,
		"limit":   filter.PageSize,
	})
}

// GetDetail returns full parameter breakdown for a single WQMS calculation ID.
func (h *WQMSHandler) GetDetail(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || id == 0 {
		response.BadRequest(c, "invalid calculation ID")
		return
	}
	rec, err := h.wqmsService.GetCalculationDetail(c.Request.Context(), uint(id))
	if err != nil {
		response.NotFound(c, "WQMS calculation record not found")
		return
	}
	response.OK(c, rec)
}

// GetMappings returns station WQMS sensor and laboratory parameter mappings.
func (h *WQMSHandler) GetMappings(c *gin.Context) {
	var deviceID uint
	if v := c.Query("device_id"); v != "" {
		if id, err := strconv.ParseUint(v, 10, 32); err == nil {
			deviceID = uint(id)
		}
	}
	mappings, err := h.wqmsService.GetMappings(c.Request.Context(), deviceID)
	if err != nil {
		response.InternalError(c, "failed to get WQMS mappings: "+err.Error())
		return
	}
	response.OK(c, mappings)
}

// SaveMapping creates or updates a WQMS sensor or periodic lab mapping.
func (h *WQMSHandler) SaveMapping(c *gin.Context) {
	var m model.WQMSParameterMapping
	if err := c.ShouldBindJSON(&m); err != nil {
		response.BadRequest(c, "invalid mapping payload: "+err.Error())
		return
	}
	if m.ParamCode == "" {
		response.BadRequest(c, "param_code is required")
		return
	}
	actor := "system"
	if u, ok := c.Get("username"); ok {
		if s, okStr := u.(string); okStr && s != "" {
			actor = s
		}
	}
	if err := h.wqmsService.SaveMapping(c.Request.Context(), &m, actor, c.ClientIP()); err != nil {
		response.InternalError(c, "failed to save WQMS mapping: "+err.Error())
		return
	}
	response.OK(c, m)
}

// DeleteMapping removes a WQMS mapping by ID.
func (h *WQMSHandler) DeleteMapping(c *gin.Context, ) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || id == 0 {
		response.BadRequest(c, "invalid mapping ID")
		return
	}
	actor := "system"
	if u, ok := c.Get("username"); ok {
		if s, okStr := u.(string); okStr && s != "" {
			actor = s
		}
	}
	if err := h.wqmsService.DeleteMapping(c.Request.Context(), uint(id), actor, c.ClientIP()); err != nil {
		response.InternalError(c, "failed to delete WQMS mapping: "+err.Error())
		return
	}
	response.OK(c, gin.H{"deleted": id})
}

// Recalculate triggers manual WQMS IP & IKA recalculation.
func (h *WQMSHandler) Recalculate(c *gin.Context) {
	var req struct {
		DeviceID   uint   `json:"device_id"`
		TargetTime string `json:"target_time"`
	}
	_ = c.ShouldBindJSON(&req)
	var targetTime time.Time
	if req.TargetTime != "" {
		if t, err := time.Parse(time.RFC3339, req.TargetTime); err == nil {
			targetTime = t
		}
	}
	actor := "system"
	if u, ok := c.Get("username"); ok {
		if s, okStr := u.(string); okStr && s != "" {
			actor = s
		}
	}
	bundle, err := h.wqmsService.Recalculate(c.Request.Context(), req.DeviceID, targetTime, actor, c.ClientIP())
	if err != nil {
		response.InternalError(c, "failed to recalculate WQMS IP & IKA: "+err.Error())
		return
	}
	response.OK(c, bundle)
}

// Backfill triggers retrospective hourly WQMS IP backfill from historical aggregation data.
func (h *WQMSHandler) Backfill(c *gin.Context) {
	var req struct {
		Hours int `json:"hours"`
	}
	_ = c.ShouldBindJSON(&req)
	if req.Hours <= 0 {
		req.Hours = 72
	}
	count, err := h.wqmsService.BackfillHourlyTrend(c.Request.Context(), req.Hours)
	if err != nil {
		response.InternalError(c, "failed to backfill WQMS trend: "+err.Error())
		return
	}
	response.OK(c, gin.H{"backfilled_periods": count})
}

// ExportCSV streams historical WQMS IP & IKA records as CSV.
func (h *WQMSHandler) ExportCSV(c *gin.Context) {
	filter := repository.WQMSHistoryFilter{
		CalculationMode: c.DefaultQuery("mode", "REGULATORY"),
		WaterClass:      c.Query("water_class"),
		Category:        c.Query("category"),
	}
	data, err := h.wqmsService.ExportCSV(c.Request.Context(), filter)
	if err != nil {
		response.InternalError(c, "failed to export WQMS CSV: "+err.Error())
		return
	}
	filename := fmt.Sprintf("wqms_ip_ika_report_%s.csv", time.Now().Format("20060102_150405"))
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	c.Data(http.StatusOK, "text/csv; charset=utf-8", data)
}

// ExportExcel streams historical WQMS IP & IKA records as UTF-8 BOM CSV for Excel.
func (h *WQMSHandler) ExportExcel(c *gin.Context) {
	filter := repository.WQMSHistoryFilter{
		CalculationMode: c.DefaultQuery("mode", "REGULATORY"),
		WaterClass:      c.Query("water_class"),
		Category:        c.Query("category"),
	}
	data, err := h.wqmsService.ExportExcel(c.Request.Context(), filter)
	if err != nil {
		response.InternalError(c, "failed to export WQMS Excel: "+err.Error())
		return
	}
	filename := fmt.Sprintf("wqms_ip_ika_report_%s.csv", time.Now().Format("20060102_150405"))
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	c.Data(http.StatusOK, "application/vnd.ms-excel; charset=utf-8", data)
}

