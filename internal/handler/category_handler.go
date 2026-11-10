package handler

import (
	"net/http"
	"strconv"
	"strings"

	"datalogger/internal/middleware"
	"datalogger/internal/model"
	"datalogger/internal/service"
	"datalogger/pkg/response"

	"github.com/gin-gonic/gin"
)

type CategoryHandler struct {
	categoryService *service.CategoryService
	tenantService   *service.TenantService
}

func NewCategoryHandler(categoryService *service.CategoryService) *CategoryHandler {
	return &CategoryHandler{categoryService: categoryService}
}

func (h *CategoryHandler) SetTenantService(ts *service.TenantService) {
	h.tenantService = ts
}

func (h *CategoryHandler) getActorInfo(c *gin.Context) (string, string, string) {
	username := "system"
	if u, exists := c.Get("username"); exists {
		if un, ok := u.(string); ok && un != "" {
			username = un
		}
	}
	ip := c.ClientIP()
	ua := c.Request.UserAgent()
	return username, ip, ua
}

func (h *CategoryHandler) checkCategoryAccess(c *gin.Context, id uint) (*model.ParameterCategory, bool) {
	cat, err := h.categoryService.GetCategory(id)
	if err != nil || cat == nil {
		response.NotFound(c, "Category not found")
		return nil, false
	}
	scope := middleware.GetScope(c)
	if scope.Authenticated && !scope.CanAccessTenant(cat.TenantID) {
		response.Forbidden(c, "Access denied: category belongs to another tenant")
		return nil, false
	}
	return cat, true
}

// ListCategories returns all parameter categories, optionally filtered by enabled status and domain
func (h *CategoryHandler) ListCategories(c *gin.Context) {
	enabledOnly := c.Query("enabled_only") == "true"
	domain := c.Query("domain")
	scope := middleware.GetScope(c)
	tenantID := scope.TenantID
	if tenantID == 0 && scope.IsSuperAdmin {
		if qTid, err := strconv.ParseUint(strings.TrimSpace(c.Query("tenant_id")), 10, 32); err == nil && qTid > 0 {
			tenantID = uint(qTid)
		} else if c.Query("all_tenants") != "true" {
			tenantID = 1
		}
	}
	categories, err := h.categoryService.ListCategoriesForTenant(tenantID, enabledOnly, domain)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve categories: "+err.Error())
		return
	}

	// Align water-quality categories ("WQMS") with active SPARING / ONLIMO station domain so Operator Dashboard
	// renders Water Quality, Physical Water, Chemical Parameters, and Flow & Level sections natively.
	if c.Query("raw_domain") != "true" {
		effectiveDom := strings.ToUpper(strings.TrimSpace(domain))
		if effectiveDom == "" && h.tenantService != nil && (scope.SiteID > 0 || scope.TenantID > 0) {
			if st, sErr := h.tenantService.ResolveEffectiveSite(scope); sErr == nil && st != nil {
				effectiveDom = strings.ToUpper(strings.TrimSpace(st.MonitoringDomain))
			}
		}
		if effectiveDom == "SPARING" || effectiveDom == "ONLIMO" {
			for i := range categories {
				if strings.EqualFold(strings.TrimSpace(categories[i].Domain), "WQMS") {
					categories[i].Domain = effectiveDom
				}
			}
		} else if effectiveDom == "CLOUD" {
			for i := range categories {
				categories[i].Domain = "ALL"
			}
		}
	}

	response.OK(c, categories)
}

// GetCategory returns a single category by ID
func (h *CategoryHandler) GetCategory(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid category ID")
		return
	}

	cat, ok := h.checkCategoryAccess(c, uint(id))
	if !ok {
		return
	}

	response.OK(c, cat)
}

// CreateCategory handles category creation with validation and audit logging
func (h *CategoryHandler) CreateCategory(c *gin.Context) {
	var req service.CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request payload: "+err.Error())
		return
	}

	scope := middleware.GetScope(c)
	if scope.Authenticated {
		req.TenantID = scope.WriteTenantID(req.TenantID)
	}

	user, ip, ua := h.getActorInfo(c)
	cat, err := h.categoryService.CreateCategory(&req, user, ip, ua)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Created(c, cat)
}

// UpdateCategory handles category modification
func (h *CategoryHandler) UpdateCategory(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid category ID")
		return
	}
	if _, ok := h.checkCategoryAccess(c, uint(id)); !ok {
		return
	}

	var req service.UpdateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request payload: "+err.Error())
		return
	}

	user, ip, ua := h.getActorInfo(c)
	cat, err := h.categoryService.UpdateCategory(uint(id), &req, user, ip, ua)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.OK(c, cat)
}

// ToggleCategory toggles a category's enabled state
func (h *CategoryHandler) ToggleCategory(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid category ID")
		return
	}
	if _, ok := h.checkCategoryAccess(c, uint(id)); !ok {
		return
	}

	var payload struct {
		Enabled bool `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		response.BadRequest(c, "Invalid request payload")
		return
	}

	user, ip, ua := h.getActorInfo(c)
	cat, err := h.categoryService.ToggleCategory(uint(id), payload.Enabled, user, ip, ua)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.OK(c, cat)
}

// DeleteCategory handles safe category deletion with optional parameter reassignment
func (h *CategoryHandler) DeleteCategory(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid category ID")
		return
	}
	if _, ok := h.checkCategoryAccess(c, uint(id)); !ok {
		return
	}

	var reassignID *uint
	if reassignStr := c.Query("reassign_to"); reassignStr != "" {
		if rid, err := strconv.ParseUint(reassignStr, 10, 32); err == nil && rid > 0 {
			uRid := uint(rid)
			reassignID = &uRid
		}
	}

	user, ip, ua := h.getActorInfo(c)
	if err := h.categoryService.DeleteCategory(uint(id), reassignID, user, ip, ua); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Message(c, "Category deleted successfully without data loss", nil)
}

// ReorderCategories updates display order of categories
func (h *CategoryHandler) ReorderCategories(c *gin.Context) {
	var req service.ReorderCategoriesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request payload: "+err.Error())
		return
	}

	user, ip, ua := h.getActorInfo(c)
	if err := h.categoryService.ReorderCategories(req.CategoryIDs, user, ip, ua); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Message(c, "Categories reordered successfully", nil)
}

// AssignParameters assigns a list of parameters to a category
func (h *CategoryHandler) AssignParameters(c *gin.Context) {
	idStr := c.Param("id")
	var targetCategoryID *uint
	if idStr != "" && idStr != "0" && idStr != "null" && idStr != "unassign" {
		if id, err := strconv.ParseUint(idStr, 10, 32); err == nil {
			if _, ok := h.checkCategoryAccess(c, uint(id)); !ok {
				return
			}
			uID := uint(id)
			targetCategoryID = &uID
		}
	}

	var payload struct {
		ParameterIDs []uint `json:"parameter_ids"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		response.BadRequest(c, "Invalid request payload: "+err.Error())
		return
	}

	user, ip, ua := h.getActorInfo(c)
	if err := h.categoryService.AssignParameters(targetCategoryID, payload.ParameterIDs, user, ip, ua); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Message(c, "Parameters assigned to category successfully", nil)
}
