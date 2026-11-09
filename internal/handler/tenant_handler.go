package handler

import (
	"strconv"
	"strings"

	"datalogger/internal/middleware"
	"datalogger/internal/service"
	"datalogger/pkg/response"
	"datalogger/pkg/tenantctx"

	"github.com/gin-gonic/gin"
)

type TenantHandler struct {
	tenantService *service.TenantService
}

func NewTenantHandler(tenantService *service.TenantService) *TenantHandler {
	return &TenantHandler{tenantService: tenantService}
}

// GetContext returns the caller's resolved tenant/site context and accessible tenants & sites for the UI switcher.
func (h *TenantHandler) GetContext(c *gin.Context) {
	scope := middleware.GetScope(c)
	tenants, err := h.tenantService.ListTenants(scope)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	sites, err := h.tenantService.ListSites(scope, scope.TenantID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.OK(c, gin.H{
		"is_super_admin":      scope.IsSuperAdmin,
		"effective_tenant_id": scope.TenantID,
		"effective_site_id":   scope.SiteID,
		"user_tenant_id":      scope.UserTenantID,
		"allowed_site_ids":    scope.AllowedSiteIDs,
		"tenants":             tenants,
		"sites":               sites,
	})
}

func (h *TenantHandler) ListTenants(c *gin.Context) {
	scope := middleware.GetScope(c)
	tenants, err := h.tenantService.ListTenants(scope)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, tenants)
}

func (h *TenantHandler) GetCurrentTenant(c *gin.Context) {
	scope := middleware.GetScope(c)
	tid := scope.TenantID
	if tid == 0 {
		tid = tenantctx.DefaultTenantID
	}
	t, err := h.tenantService.GetTenantByID(tid)
	if err != nil || t == nil {
		response.NotFound(c, "Tenant not found")
		return
	}
	if !scope.CanAccessTenant(t.ID) {
		response.Forbidden(c, "Access denied: tenant belongs to another organization")
		return
	}
	response.OK(c, t)
}

func (h *TenantHandler) CreateTenant(c *gin.Context) {
	scope := middleware.GetScope(c)
	if !scope.IsSuperAdmin {
		response.Forbidden(c, "Hanya Superadmin platform yang dapat membuat tenant baru")
		return
	}
	var req service.CreateTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid tenant payload: "+err.Error())
		return
	}
	created, err := h.tenantService.CreateTenant(req, scope.Username, c.ClientIP())
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Created(c, created)
}

func (h *TenantHandler) UpdateCurrentTenant(c *gin.Context) {
	scope := middleware.GetScope(c)
	if !scope.IsSuperAdmin && !strings.EqualFold(scope.Role, "ADMIN") {
		response.Forbidden(c, "Hanya Superadmin atau Admin Tenant yang dapat memperbarui profil organisasi")
		return
	}
	tid := scope.TenantID
	if tid == 0 {
		tid = tenantctx.DefaultTenantID
	}
	var req service.UpdateTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid tenant update payload: "+err.Error())
		return
	}
	updated, err := h.tenantService.UpdateTenant(tid, req, scope, c.ClientIP())
	if err != nil {
		lower := strings.ToLower(err.Error())
		if strings.Contains(lower, "access denied") {
			response.Forbidden(c, err.Error())
			return
		}
		if strings.Contains(lower, "not found") {
			response.NotFound(c, err.Error())
			return
		}
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, updated)
}

func (h *TenantHandler) UpdateTenant(c *gin.Context) {
	if strings.EqualFold(strings.TrimSpace(c.Param("id")), "current") {
		h.UpdateCurrentTenant(c)
		return
	}
	scope := middleware.GetScope(c)
	if !scope.IsSuperAdmin && !strings.EqualFold(scope.Role, "ADMIN") {
		response.Forbidden(c, "Hanya Superadmin atau Admin Tenant yang dapat memperbarui profil organisasi")
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || id == 0 {
		response.BadRequest(c, "Invalid tenant ID")
		return
	}
	var req service.UpdateTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid tenant update payload: "+err.Error())
		return
	}
	updated, err := h.tenantService.UpdateTenant(uint(id), req, scope, c.ClientIP())
	if err != nil {
		lower := strings.ToLower(err.Error())
		if strings.Contains(lower, "access denied") {
			response.Forbidden(c, err.Error())
			return
		}
		if strings.Contains(lower, "not found") {
			response.NotFound(c, err.Error())
			return
		}
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, updated)
}

func (h *TenantHandler) DeleteTenant(c *gin.Context) {
	scope := middleware.GetScope(c)
	if !scope.IsSuperAdmin {
		response.Forbidden(c, "Hanya Superadmin platform yang dapat menghapus tenant")
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || id == 0 {
		response.BadRequest(c, "Invalid tenant ID")
		return
	}
	if err := h.tenantService.DeleteTenant(uint(id), scope.Username, c.ClientIP()); err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			response.NotFound(c, err.Error())
			return
		}
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, gin.H{"deleted": true, "id": uint(id)})
}

// --- Sites / Stations ---

func (h *TenantHandler) ListSites(c *gin.Context) {
	scope := middleware.GetScope(c)
	var explicitTid uint
	if tStr := strings.TrimSpace(c.Query("tenant_id")); tStr != "" {
		if v, err := strconv.ParseUint(tStr, 10, 32); err == nil {
			explicitTid = uint(v)
		}
	}
	sites, err := h.tenantService.ListSites(scope, explicitTid)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, sites)
}

func (h *TenantHandler) CreateSite(c *gin.Context) {
	scope := middleware.GetScope(c)
	var req service.CreateSiteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid station payload: "+err.Error())
		return
	}
	created, err := h.tenantService.CreateSite(req, scope, c.ClientIP())
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Created(c, created)
}

func (h *TenantHandler) UpdateSite(c *gin.Context) {
	scope := middleware.GetScope(c)
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || id == 0 {
		response.BadRequest(c, "Invalid station ID")
		return
	}
	var req service.UpdateSiteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid station update payload: "+err.Error())
		return
	}
	updated, err := h.tenantService.UpdateSite(uint(id), req, scope, c.ClientIP())
	if err != nil {
		lower := strings.ToLower(err.Error())
		if strings.Contains(lower, "access denied") {
			response.Forbidden(c, err.Error())
			return
		}
		if strings.Contains(lower, "not found") {
			response.NotFound(c, err.Error())
			return
		}
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, updated)
}

func (h *TenantHandler) DeleteSite(c *gin.Context) {
	scope := middleware.GetScope(c)
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || id == 0 {
		response.BadRequest(c, "Invalid station ID")
		return
	}
	if err := h.tenantService.DeleteSite(uint(id), scope, c.ClientIP()); err != nil {
		lower := strings.ToLower(err.Error())
		if strings.Contains(lower, "access denied") {
			response.Forbidden(c, err.Error())
			return
		}
		if strings.Contains(lower, "not found") {
			response.NotFound(c, err.Error())
			return
		}
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, gin.H{"deleted": true, "id": uint(id)})
}

// --- Edge Ingest API Keys ---

func (h *TenantHandler) ListAPIKeys(c *gin.Context) {
	scope := middleware.GetScope(c)
	var explicitTid uint
	if tStr := strings.TrimSpace(c.Query("tenant_id")); tStr != "" {
		if v, err := strconv.ParseUint(tStr, 10, 32); err == nil {
			explicitTid = uint(v)
		}
	}
	keys, err := h.tenantService.ListAPIKeys(scope, explicitTid)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, keys)
}

func (h *TenantHandler) CreateAPIKey(c *gin.Context) {
	scope := middleware.GetScope(c)
	var req service.CreateAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid API key payload: "+err.Error())
		return
	}
	key, rawKey, err := h.tenantService.CreateAPIKey(req, scope, c.ClientIP())
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Created(c, gin.H{
		"api_key":   key,
		"key":       key,
		"raw_key":   rawKey,
		"plain_key": rawKey,
		"warning":   "Simpan API Key ini sekarang — kunci penuh hanya ditampilkan satu kali.",
	})
}

func (h *TenantHandler) ToggleAPIKey(c *gin.Context) {
	scope := middleware.GetScope(c)
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || id == 0 {
		response.BadRequest(c, "Invalid API key ID")
		return
	}
	var req struct {
		Enabled       *bool `json:"enabled"`
		AutoProvision *bool `json:"auto_provision"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid API key update payload: "+err.Error())
		return
	}
	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}
	updated, err := h.tenantService.ToggleAPIKey(uint(id), enabled, req.AutoProvision, scope, c.ClientIP())
	if err != nil {
		lower := strings.ToLower(err.Error())
		if strings.Contains(lower, "access denied") {
			response.Forbidden(c, err.Error())
			return
		}
		if strings.Contains(lower, "not found") {
			response.NotFound(c, err.Error())
			return
		}
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, updated)
}

func (h *TenantHandler) DeleteAPIKey(c *gin.Context) {
	scope := middleware.GetScope(c)
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || id == 0 {
		response.BadRequest(c, "Invalid API key ID")
		return
	}
	if err := h.tenantService.DeleteAPIKey(uint(id), scope, c.ClientIP()); err != nil {
		lower := strings.ToLower(err.Error())
		if strings.Contains(lower, "access denied") {
			response.Forbidden(c, err.Error())
			return
		}
		if strings.Contains(lower, "not found") {
			response.NotFound(c, err.Error())
			return
		}
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, gin.H{"deleted": true, "id": uint(id)})
}
