package handler

import (
	"net/http"
	"strconv"

	"datalogger/internal/service"
	"datalogger/pkg/response"

	"github.com/gin-gonic/gin"
)

type CategoryHandler struct {
	categoryService *service.CategoryService
}

func NewCategoryHandler(categoryService *service.CategoryService) *CategoryHandler {
	return &CategoryHandler{categoryService: categoryService}
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

// ListCategories returns all parameter categories, optionally filtered by enabled status and domain
func (h *CategoryHandler) ListCategories(c *gin.Context) {
	enabledOnly := c.Query("enabled_only") == "true"
	domain := c.Query("domain")
	categories, err := h.categoryService.ListCategoriesWithDomain(enabledOnly, domain)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve categories: "+err.Error())
		return
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

	cat, err := h.categoryService.GetCategory(uint(id))
	if err != nil {
		response.NotFound(c, "Category not found")
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
