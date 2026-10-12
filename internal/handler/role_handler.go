package handler

import (
	"strconv"

	"datalogger/internal/service"
	"datalogger/pkg/response"

	"github.com/gin-gonic/gin"
)

type RoleHandler struct {
	authService *service.AuthService
}

func NewRoleHandler(authService *service.AuthService) *RoleHandler {
	return &RoleHandler{authService: authService}
}

func (h *RoleHandler) getActorInfo(c *gin.Context) (uint, string, string, string) {
	var actorID uint
	if uid, exists := c.Get("userID"); exists {
		if id, ok := uid.(uint); ok {
			actorID = id
		}
	}
	actorUsername := "system"
	if u, exists := c.Get("username"); exists {
		if un, ok := u.(string); ok {
			actorUsername = un
		}
	}
	ip := c.ClientIP()
	ua := c.Request.UserAgent()
	return actorID, actorUsername, ip, ua
}

// ListRoles returns all roles with their assigned permissions
func (h *RoleHandler) ListRoles(c *gin.Context) {
	roles, err := h.authService.GetRoles()
	if err != nil {
		response.InternalError(c, "Failed to retrieve roles: "+err.Error())
		return
	}
	response.OK(c, roles)
}

// GetRole returns a role by ID
func (h *RoleHandler) GetRole(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid role ID")
		return
	}

	role, err := h.authService.GetRoleByID(uint(id))
	if err != nil || role == nil {
		response.NotFound(c, "Role not found")
		return
	}
	response.OK(c, role)
}

// CreateRole creates a new role with optional permission IDs
func (h *RoleHandler) CreateRole(c *gin.Context) {
	var input service.RoleInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, "Invalid request payload: "+err.Error())
		return
	}

	actorID, actorUsername, ip, ua := h.getActorInfo(c)
	role, err := h.authService.CreateRole(input, actorID, actorUsername, ip, ua)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, role)
}

// UpdateRole updates role name and description
func (h *RoleHandler) UpdateRole(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid role ID")
		return
	}

	var input service.RoleInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, "Invalid request payload: "+err.Error())
		return
	}

	actorID, actorUsername, ip, ua := h.getActorInfo(c)
	role, err := h.authService.UpdateRole(uint(id), input, actorID, actorUsername, ip, ua)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, role)
}

// UpdateRolePermissions updates the permission matrix assignment for a role
func (h *RoleHandler) UpdateRolePermissions(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid role ID")
		return
	}

	var req struct {
		PermissionIDs []uint `json:"permission_ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid permission payload: "+err.Error())
		return
	}

	actorID, actorUsername, ip, ua := h.getActorInfo(c)
	if err := h.authService.UpdateRolePermissions(uint(id), req.PermissionIDs, actorID, actorUsername, ip, ua); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	role, _ := h.authService.GetRoleByID(uint(id))
	response.OK(c, gin.H{
		"role":    role,
		"message": "Role permissions updated successfully",
	})
}

// DeleteRole deletes a custom role
func (h *RoleHandler) DeleteRole(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid role ID")
		return
	}

	actorID, actorUsername, ip, ua := h.getActorInfo(c)
	if err := h.authService.DeleteRole(uint(id), actorID, actorUsername, ip, ua); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.OK(c, gin.H{
		"id":      id,
		"message": "Role deleted successfully",
	})
}

// ListPermissions lists all system permissions categorized by resource
func (h *RoleHandler) ListPermissions(c *gin.Context) {
	perms, err := h.authService.GetPermissions()
	if err != nil {
		response.InternalError(c, "Failed to retrieve permissions: "+err.Error())
		return
	}

	// Group by category
	type CategoryGroup struct {
		Category    string             `json:"category"`
		Permissions []interface{}      `json:"permissions"`
	}
	categoryMap := make(map[string][]interface{})
	var categories []string

	for _, p := range perms {
		cat := p.Category
		if cat == "" {
			cat = "SYSTEM"
		}
		if _, exists := categoryMap[cat]; !exists {
			categories = append(categories, cat)
		}
		categoryMap[cat] = append(categoryMap[cat], p)
	}

	var grouped []CategoryGroup
	for _, cat := range categories {
		grouped = append(grouped, CategoryGroup{
			Category:    cat,
			Permissions: categoryMap[cat],
		})
	}

	response.OK(c, gin.H{
		"permissions": perms,
		"grouped":     grouped,
	})
}
