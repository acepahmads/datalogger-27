package handler

import (
	"strconv"
	"strings"

	"datalogger/internal/service"
	"datalogger/pkg/response"

	"github.com/gin-gonic/gin"
)

// isStrictSuperAdmin verifies that the caller holds SUPERADMIN (or legacy Administrator) role,
// strictly excluding ADMIN, ENGINEER, and OPERATOR so injection metadata never leaks to non-Superadmins.
func isStrictSuperAdmin(c *gin.Context) bool {
	roleVal, exists := c.Get("role")
	if !exists {
		return false
	}
	role, ok := roleVal.(string)
	return ok && (strings.EqualFold(role, "SUPERADMIN") || strings.EqualFold(role, "ADMINISTRATOR"))
}

// TelemetryInjectionHandler exposes Superadmin-only REST endpoints for telemetry injection, preview, history, and rollback
type TelemetryInjectionHandler struct {
	injService *service.TelemetryInjectionService
}

// NewTelemetryInjectionHandler creates a new TelemetryInjectionHandler
func NewTelemetryInjectionHandler(injService *service.TelemetryInjectionService) *TelemetryInjectionHandler {
	return &TelemetryInjectionHandler{injService: injService}
}

// GetDeviceHints handles GET /api/superadmin/telemetry-inject/devices/:id/hints
func (h *TelemetryInjectionHandler) GetDeviceHints(c *gin.Context) {
	if !isStrictSuperAdmin(c) {
		response.Forbidden(c, "Access denied: Superadmin privileges required")
		return
	}
	devID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || devID == 0 {
		response.BadRequest(c, "Invalid device ID")
		return
	}
	hints, err := h.injService.GetDeviceParameterHints(c.Request.Context(), uint(devID))
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, hints)
}

// Preview handles POST /api/superadmin/telemetry-inject/preview
func (h *TelemetryInjectionHandler) Preview(c *gin.Context) {
	if !isStrictSuperAdmin(c) {
		response.Forbidden(c, "Access denied: Superadmin privileges required")
		return
	}
	var req service.TelemetryInjectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Payload tidak valid: "+err.Error())
		return
	}
	preview, err := h.injService.PreviewInjection(c.Request.Context(), req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, preview)
}

// Execute handles POST /api/superadmin/telemetry-inject/execute
func (h *TelemetryInjectionHandler) Execute(c *gin.Context) {
	if !isStrictSuperAdmin(c) {
		response.Forbidden(c, "Access denied: Superadmin privileges required")
		return
	}
	var req service.TelemetryInjectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Payload tidak valid: "+err.Error())
		return
	}
	uVal, _ := c.Get("username")
	username, _ := uVal.(string)

	res, err := h.injService.ExecuteInjection(c.Request.Context(), req, username)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Created(c, res)
}

// ListBatches handles GET /api/superadmin/telemetry-inject/batches
func (h *TelemetryInjectionHandler) ListBatches(c *gin.Context) {
	if !isStrictSuperAdmin(c) {
		response.Forbidden(c, "Access denied: Superadmin privileges required")
		return
	}
	var devID uint
	if dStr := c.Query("device_id"); dStr != "" {
		if dVal, err := strconv.ParseUint(dStr, 10, 32); err == nil {
			devID = uint(dVal)
		}
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	items, total, err := h.injService.ListBatches(c.Request.Context(), devID, page, pageSize)
	if err != nil {
		response.InternalError(c, "Gagal memuat riwayat batch suntikan: "+err.Error())
		return
	}
	response.OK(c, response.PaginatedData{
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		Items:    items,
	})
}

// Rollback handles POST /api/superadmin/telemetry-inject/batches/:id/rollback (and DELETE /:id)
func (h *TelemetryInjectionHandler) Rollback(c *gin.Context) {
	if !isStrictSuperAdmin(c) {
		response.Forbidden(c, "Access denied: Superadmin privileges required")
		return
	}
	idVal, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || idVal == 0 {
		response.BadRequest(c, "ID batch tidak valid")
		return
	}
	uVal, _ := c.Get("username")
	username, _ := uVal.(string)

	res, err := h.injService.RollbackBatch(c.Request.Context(), uint(idVal), username)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, res)
}

// SyncOutbound handles POST /api/superadmin/telemetry-inject/batches/:id/sync-outbound
func (h *TelemetryInjectionHandler) SyncOutbound(c *gin.Context) {
	if !isStrictSuperAdmin(c) {
		response.Forbidden(c, "Access denied: Superadmin privileges required")
		return
	}
	idVal, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || idVal == 0 {
		response.BadRequest(c, "ID batch tidak valid")
		return
	}

	var body struct {
		OutboundDestinationIDs []uint `json:"outbound_destination_ids"`
	}
	_ = c.ShouldBindJSON(&body)

	res, err := h.injService.SyncBatchOutbound(c.Request.Context(), uint(idVal), body.OutboundDestinationIDs)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, res)
}


