package handler

import (
	"strings"

	"datalogger/internal/service"
	"datalogger/pkg/response"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Username and password are required")
		return
	}

	ip := c.ClientIP()
	ua := c.Request.UserAgent()
	token, user, err := h.authService.Login(req.Username, req.Password, ip, ua)
	if err != nil {
		response.Unauthorized(c, err.Error())
		return
	}

	roleName := "OPERATOR"
	if user.Role != nil {
		uName := strings.ToUpper(user.Role.Name)
		if uName == "SUPERADMIN" || uName == "ADMIN" || uName == "ENGINEER" || uName == "OPERATOR" {
			roleName = uName
		} else {
			roleName = user.Role.Name
		}
	}

	response.OK(c, gin.H{
		"token": token,
		"user": gin.H{
			"id":        user.ID,
			"username":  user.Username,
			"full_name": user.FullName,
			"role":      roleName,
			"email":     user.Email,
		},
	})
}

func (h *AuthHandler) Me(c *gin.Context) {
	username, _ := c.Get("username")
	role, _ := c.Get("role")
	userID, _ := c.Get("userID")

	response.OK(c, gin.H{
		"id":       userID,
		"username": username,
		"role":     role,
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	usernameVal, _ := c.Get("username")
	userIDVal, _ := c.Get("userID")

	var uID uint
	if id, ok := userIDVal.(uint); ok {
		uID = id
	}
	var uName string
	if name, ok := usernameVal.(string); ok {
		uName = name
	}

	ip := c.ClientIP()
	ua := c.Request.UserAgent()

	if uName != "" {
		_ = h.authService.Logout(uID, uName, ip, ua)
	}

	response.OK(c, gin.H{
		"message": "Logged out successfully",
	})
}

type SwitchAccountRequest struct {
	TargetUsername string `json:"target_username" binding:"required"`
}

func (h *AuthHandler) SwitchAccount(c *gin.Context) {
	callerRole, _ := c.Get("role")
	callerUsername, _ := c.Get("username")
	callerID, _ := c.Get("userID")

	cRole, _ := callerRole.(string)
	ucRole := strings.ToUpper(cRole)
	if ucRole != "SUPERADMIN" && ucRole != "ADMINISTRATOR" {
		response.Forbidden(c, "Only SUPERADMIN is authorized to perform quick role switching")
		return
	}

	var req SwitchAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Target username is required")
		return
	}

	var cID uint
	if id, ok := callerID.(uint); ok {
		cID = id
	}
	var cName string
	if name, ok := callerUsername.(string); ok {
		cName = name
	}

	ip := c.ClientIP()
	ua := c.Request.UserAgent()

	token, targetUser, err := h.authService.SwitchAccount(cID, cName, req.TargetUsername, ip, ua)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	roleName := "OPERATOR"
	if targetUser.Role != nil {
		uName := strings.ToUpper(targetUser.Role.Name)
		if uName == "SUPERADMIN" || uName == "ADMIN" || uName == "ENGINEER" || uName == "OPERATOR" {
			roleName = uName
		} else {
			roleName = targetUser.Role.Name
		}
	}

	response.OK(c, gin.H{
		"token": token,
		"user": gin.H{
			"id":        targetUser.ID,
			"username":  targetUser.Username,
			"full_name": targetUser.FullName,
			"role":      roleName,
			"email":     targetUser.Email,
		},
	})
}
