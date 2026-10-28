package middleware

import (
	"net/http"
	"strings"
	"sync"

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

// JWTAuth verifies the Bearer token in the Authorization header or query parameter 'token'
func JWTAuth(authService *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := ""
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			parts := strings.SplitN(authHeader, " ", 2)
			if !(len(parts) == 2 && strings.EqualFold(parts[0], "Bearer")) {
				response.Unauthorized(c, "Authorization header format must be Bearer {token}")
				c.Abort()
				return
			}
			tokenString = parts[1]
		} else {
			tokenString = c.Query("token")
		}

		if tokenString == "" {
			response.Unauthorized(c, "Authorization header required")
			c.Abort()
			return
		}

		claims, err := authService.ValidateToken(tokenString)
		if err != nil {
			response.Unauthorized(c, "Invalid or expired token")
			c.Abort()
			return
		}

		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("fullName", claims.FullName)
		c.Set("role", claims.Role)
		c.Next()
	}
}

var (
	permCheckerMu sync.RWMutex
	permChecker   func(role string, permission string) bool
)

// SetPermissionChecker registers a dynamic permission checker function
func SetPermissionChecker(checker func(role string, permission string) bool) {
	permCheckerMu.Lock()
	defer permCheckerMu.Unlock()
	permChecker = checker
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

		// SUPERADMIN (and legacy Administrator) always has full system permissions
		if strings.EqualFold(role, "SUPERADMIN") || strings.EqualFold(role, "Administrator") {
			c.Next()
			return
		}

		// Strict restriction: Non-superadmin CANNOT access development.*
		if strings.HasPrefix(permissionCode, "development.") {
			response.Forbidden(c, "Access denied: missing required permission '"+permissionCode+"'")
			c.Abort()
			return
		}

		// Mode raw value / telemetry.raw restriction: only ADMIN and SUPERADMIN/Administrator
		if permissionCode == "telemetry.raw" {
			if !strings.EqualFold(role, "ADMIN") {
				response.Forbidden(c, "Access denied: missing required permission 'telemetry.raw' (Admin and Superadmin only)")
				c.Abort()
				return
			}
		}

		// Check dynamic permission checker if registered
		permCheckerMu.RLock()
		checker := permChecker
		permCheckerMu.RUnlock()
		if checker != nil && checker(role, permissionCode) {
			c.Next()
			return
		}

		// ADMIN fallback: Operational & administrative permissions (strictly no dev or role.manage; telemetry.raw is permitted)
		if strings.EqualFold(role, "ADMIN") {
			if !strings.HasPrefix(permissionCode, "development.") && permissionCode != "role.manage" {
				c.Next()
				return
			}
		}

		// ENGINEER fallback
		if strings.EqualFold(role, "ENGINEER") || strings.EqualFold(role, "Engineer") {
			engAllowed := map[string]bool{
				"device.view": true, "device.create": true, "device.update": true, "device.manage": true,
				"device.communication.view": true, "device.communication.manage": true, "device.communication.test": true,
				"telemetry.view": true,
				"category.view": true, "category.manage": true, "category.assign": true,
				"backup.view": true, "backup.create": true, "backup.validate": true,
				"retention.view": true,
				"alarm.view": true, "alarm.acknowledge": true, "alarm.manage": true,
				"notification.view": true, "notification.manage": true,
				"report.view": true, "report.generate": true,
				"ispu.view": true, "ispu.export": true, "ispu.recalculate": true, "ispu.manage_mapping": true,
				"config.view": true,
			}
			if engAllowed[permissionCode] {
				c.Next()
				return
			}
		}

		// OPERATOR fallback
		if strings.EqualFold(role, "OPERATOR") || strings.EqualFold(role, "Operator") {
			opAllowed := map[string]bool{
				"device.view": true, "device.communication.view": true,
				"telemetry.view": true,
				"category.view": true,
				"alarm.view": true, "alarm.acknowledge": true,
				"notification.view": true,
				"report.view":       true,
				"ispu.view":         true, "ispu.export": true,
			}
			if opAllowed[permissionCode] {
				c.Next()
				return
			}
		}

		response.Forbidden(c, "Access denied: missing required permission '"+permissionCode+"'")
		c.Abort()
	}
}

// RequireRole checks if the authenticated user has one of the specified roles
func RequireRole(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleVal, exists := c.Get("role")
		if !exists {
			response.Unauthorized(c, "Authentication required")
			c.Abort()
			return
		}

		role, _ := roleVal.(string)
		for _, allowed := range allowedRoles {
			if strings.EqualFold(role, allowed) || (strings.EqualFold(allowed, "SUPERADMIN") && strings.EqualFold(role, "Administrator")) {
				c.Next()
				return
			}
		}

		response.Forbidden(c, "Access denied: role '"+role+"' is not authorized for this resource")
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
