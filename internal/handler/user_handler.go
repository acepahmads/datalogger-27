package handler

import (
	"strconv"
	"strings"

	"datalogger/internal/repository"
	"datalogger/internal/service"
	"datalogger/pkg/response"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	authService *service.AuthService
}

func NewUserHandler(authService *service.AuthService) *UserHandler {
	return &UserHandler{authService: authService}
}

func (h *UserHandler) getActorInfo(c *gin.Context) (uint, string, string, string, string) {
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
	actorRole := ""
	if r, exists := c.Get("role"); exists {
		if rs, ok := r.(string); ok {
			actorRole = rs
		}
	}
	ip := c.ClientIP()
	ua := c.Request.UserAgent()
	return actorID, actorUsername, actorRole, ip, ua
}

// ListUsers returns paginated users with optional search and status filtering
func (h *UserHandler) ListUsers(c *gin.Context) {
	search := c.Query("search")
	status := c.Query("status")
	roleIDStr := c.Query("role_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	var roleID uint
	if roleIDStr != "" {
		if rid, err := strconv.ParseUint(roleIDStr, 10, 32); err == nil {
			roleID = uint(rid)
		}
	}

	params := repository.UserFilterParams{
		Search:   search,
		Status:   status,
		RoleID:   roleID,
		Page:     page,
		PageSize: pageSize,
	}

	users, total, err := h.authService.GetUsers(params)
	if err != nil {
		response.InternalError(c, "Failed to retrieve users: "+err.Error())
		return
	}

	response.OK(c, gin.H{
		"users":     users,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetUser returns a single user by ID
func (h *UserHandler) GetUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	user, err := h.authService.GetUserByID(uint(id))
	if err != nil || user == nil {
		response.NotFound(c, "User not found")
		return
	}

	response.OK(c, user)
}

// CreateUser creates a new system user
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req service.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request payload: "+err.Error())
		return
	}

	actorID, actorUsername, actorRole, ip, ua := h.getActorInfo(c)
	user, err := h.authService.CreateUser(req, actorID, actorUsername, actorRole, ip, ua)
	if err != nil {
		if strings.HasPrefix(strings.ToLower(err.Error()), "access denied") {
			response.Forbidden(c, err.Error())
			return
		}
		response.BadRequest(c, err.Error())
		return
	}

	response.OK(c, user)
}

// UpdateUser modifies user profile details
func (h *UserHandler) UpdateUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	var req service.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request payload: "+err.Error())
		return
	}

	actorID, actorUsername, actorRole, ip, ua := h.getActorInfo(c)
	user, err := h.authService.UpdateUser(uint(id), req, actorID, actorUsername, actorRole, ip, ua)
	if err != nil {
		if strings.HasPrefix(strings.ToLower(err.Error()), "access denied") {
			response.Forbidden(c, err.Error())
			return
		}
		response.BadRequest(c, err.Error())
		return
	}

	response.OK(c, user)
}

// ToggleUserActive enables or disables a user account
func (h *UserHandler) ToggleUserActive(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	var req struct {
		IsActive bool `json:"is_active"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "is_active flag is required")
		return
	}

	actorID, actorUsername, actorRole, ip, ua := h.getActorInfo(c)
	if err := h.authService.SetUserActive(uint(id), req.IsActive, actorID, actorUsername, actorRole, ip, ua); err != nil {
		if strings.HasPrefix(strings.ToLower(err.Error()), "access denied") {
			response.Forbidden(c, err.Error())
			return
		}
		response.BadRequest(c, err.Error())
		return
	}

	response.OK(c, gin.H{
		"id":        id,
		"is_active": req.IsActive,
		"message":   "User status updated successfully",
	})
}

// ResetPassword resets a user's password
func (h *UserHandler) ResetPassword(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	var req struct {
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Password is required")
		return
	}

	actorID, actorUsername, actorRole, ip, ua := h.getActorInfo(c)
	if err := h.authService.ResetPassword(uint(id), req.Password, actorID, actorUsername, actorRole, ip, ua); err != nil {
		if strings.HasPrefix(strings.ToLower(err.Error()), "access denied") {
			response.Forbidden(c, err.Error())
			return
		}
		response.BadRequest(c, err.Error())
		return
	}

	response.OK(c, gin.H{
		"id":      id,
		"message": "User password updated successfully",
	})
}

// DeleteUser deletes a user account
func (h *UserHandler) DeleteUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	actorID, actorUsername, actorRole, ip, ua := h.getActorInfo(c)
	if err := h.authService.DeleteUser(uint(id), actorID, actorUsername, actorRole, ip, ua); err != nil {
		if strings.HasPrefix(strings.ToLower(err.Error()), "access denied") {
			response.Forbidden(c, err.Error())
			return
		}
		response.BadRequest(c, err.Error())
		return
	}

	response.OK(c, gin.H{
		"id":      id,
		"message": "User account deleted successfully",
	})
}
