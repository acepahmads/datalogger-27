package handler

import (
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

	roleName := "Viewer"
	if user.Role != nil {
		roleName = user.Role.Name
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
