package middleware

import (
	"net/http"
	"strings"

	"datalogger/internal/logger"
	"datalogger/internal/service"
	"datalogger/pkg/response"

	"github.com/gin-gonic/gin"
)

// CORS handles cross-origin requests for local and browser access
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")
		c.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Disposition, Content-Length")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// JWTAuth verifies the Bearer token in the Authorization header
func JWTAuth(authService *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "Authorization header required")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			response.Unauthorized(c, "Authorization header format must be Bearer {token}")
			c.Abort()
			return
		}

		claims, err := authService.ValidateToken(parts[1])
		if err != nil {
			response.Unauthorized(c, "Invalid or expired token")
			c.Abort()
			return
		}

		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)
		c.Next()
	}
}

// RequirePermission checks if the authenticated user has the specified permission
func RequirePermission(permissionCode string) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleVal, exists := c.Get("role")
		if !exists {
			response.Unauthorized(c, "Authentication required")
			c.Abort()
			return
		}

		role, _ := roleVal.(string)

		// Administrator always has full system permissions
		if strings.EqualFold(role, "Administrator") || strings.EqualFold(role, "admin") {
			c.Next()
			return
		}

		// Engineer has permissions for creating, updating, managing, viewing devices, backup operations, and viewing retention
		if strings.EqualFold(role, "Engineer") {
			if permissionCode == "device.view" || permissionCode == "device.create" ||
				permissionCode == "device.update" || permissionCode == "device.manage" ||
				permissionCode == "device.communication.view" || permissionCode == "device.communication.manage" ||
				permissionCode == "device.communication.test" ||
				permissionCode == "backup.view" || permissionCode == "backup.create" || permissionCode == "backup.validate" ||
				permissionCode == "retention.view" {
				c.Next()
				return
			}
		}

		// Operator only has read/view permissions
		if strings.EqualFold(role, "Operator") {
			if permissionCode == "device.view" || permissionCode == "device.communication.view" {
				c.Next()
				return
			}
		}

		response.Forbidden(c, "Access denied: missing required permission '"+permissionCode+"'")
		c.Abort()
	}
}

// RequestLogger logs incoming HTTP requests with latency
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		method := c.Request.Method

		c.Next()

		status := c.Writer.Status()
		if path != "/ws" && !strings.HasPrefix(path, "/static") {
			logger.Debug("%s %s -> %d", method, path, status)
		}
	}
}
