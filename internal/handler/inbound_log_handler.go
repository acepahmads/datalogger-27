package handler

import (
	"strconv"
	"strings"
	"time"

	"datalogger/internal/config"
	"datalogger/internal/repository"
	"datalogger/pkg/response"

	"github.com/gin-gonic/gin"
)

// InboundLogHandler exposes Cloud-mode Superadmin-only endpoints for inspecting raw inbound JSON telemetry payloads across all tenants.
type InboundLogHandler struct {
	repo *repository.InboundLogRepository
}

func NewInboundLogHandler(repo *repository.InboundLogRepository) *InboundLogHandler {
	return &InboundLogHandler{repo: repo}
}

func (h *InboundLogHandler) guardCloudSuperAdmin(c *gin.Context) bool {
	if !isStrictSuperAdmin(c) {
		response.Forbidden(c, "Access denied: Inbound Data monitor requires SUPERADMIN privileges")
		return false
	}
	cfg := config.Get()
	if cfg == nil || !cfg.IsCloudMode() {
		response.Forbidden(c, "Inbound Data monitor is only available in Cloud Mode")
		return false
	}
	return true
}

func parseInboundLogFilter(c *gin.Context) repository.InboundLogFilter {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "25"))
	f := repository.InboundLogFilter{
		SiteCode:       strings.TrimSpace(c.Query("site_code")),
		DeviceCode:     strings.TrimSpace(c.Query("device_code")),
		Status:         strings.TrimSpace(c.Query("status")),
		Source:         strings.TrimSpace(c.Query("source")),
		LoggerType:     strings.TrimSpace(c.Query("logger_type")),
		DetectedFormat: strings.TrimSpace(c.Query("format")),
		Endpoint:       strings.TrimSpace(c.Query("endpoint")),
		Search:         strings.TrimSpace(c.Query("search")),
		Page:           page,
		PageSize:       pageSize,
	}
	if tidStr := strings.TrimSpace(c.Query("tenant_id")); tidStr != "" {
		if tid, err := strconv.ParseUint(tidStr, 10, 32); err == nil {
			f.TenantID = uint(tid)
		}
	}
	if startStr := strings.TrimSpace(c.Query("start_time")); startStr != "" {
		if t, err := time.Parse(time.RFC3339, startStr); err == nil {
			f.StartTime = &t
		}
	}
	if endStr := strings.TrimSpace(c.Query("end_time")); endStr != "" {
		if t, err := time.Parse(time.RFC3339, endStr); err == nil {
			f.EndTime = &t
		}
	}
	return f
}

// List handles GET /api/superadmin/inbound-logs
func (h *InboundLogHandler) List(c *gin.Context) {
	if !h.guardCloudSuperAdmin(c) {
		return
	}
	filter := parseInboundLogFilter(c)
	items, total, err := h.repo.List(c.Request.Context(), filter)
	if err != nil {
		response.InternalError(c, "Failed to query inbound logs: "+err.Error())
		return
	}
	summary, err := h.repo.GetSummary(c.Request.Context(), filter)
	if err != nil {
		summary = &repository.InboundLogSummary{}
	}
	page := filter.Page
	if page < 1 {
		page = 1
	}
	pageSize := filter.PageSize
	if pageSize < 1 {
		pageSize = 25
	} else if pageSize > 100 {
		pageSize = 100
	}
	response.OK(c, gin.H{
		"items":     items,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
		"summary":   summary,
	})
}

// GetByID handles GET /api/superadmin/inbound-logs/:id
func (h *InboundLogHandler) GetByID(c *gin.Context) {
	if !h.guardCloudSuperAdmin(c) {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		response.BadRequest(c, "Invalid inbound log ID")
		return
	}
	item, err := h.repo.GetByID(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, "Inbound log entry not found")
		return
	}
	response.OK(c, item)
}

// Purge handles DELETE /api/superadmin/inbound-logs
func (h *InboundLogHandler) Purge(c *gin.Context) {
	if !h.guardCloudSuperAdmin(c) {
		return
	}
	clearAll := strings.EqualFold(strings.TrimSpace(c.Query("all")), "true") || strings.TrimSpace(c.Query("all")) == "1"
	days := 7
	if dStr := strings.TrimSpace(c.Query("older_than_days")); dStr != "" {
		if v, err := strconv.Atoi(dStr); err == nil && v >= 1 {
			days = v
		}
	}
	deleted, err := h.repo.Purge(c.Request.Context(), days, clearAll)
	if err != nil {
		response.InternalError(c, "Failed to purge inbound logs: "+err.Error())
		return
	}
	response.OK(c, gin.H{
		"deleted":         deleted,
		"clear_all":       clearAll,
		"older_than_days": days,
	})
}
