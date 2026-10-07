package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

type PaginatedData struct {
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
	Items    interface{} `json:"items"`
}

func OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    data,
	})
}

func Message(c *gin.Context, msg string, data interface{}) {
	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: msg,
		Data:    data,
	})
}

func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, APIResponse{
		Success: true,
		Data:    data,
	})
}

func Error(c *gin.Context, statusCode int, errMsg string) {
	c.JSON(statusCode, APIResponse{
		Success: false,
		Error:   errMsg,
	})
}

func BadRequest(c *gin.Context, errMsg string) {
	Error(c, http.StatusBadRequest, errMsg)
}

func Unauthorized(c *gin.Context, errMsg string) {
	if errMsg == "" {
		errMsg = "Unauthorized access"
	}
	Error(c, http.StatusUnauthorized, errMsg)
}

func Forbidden(c *gin.Context, errMsg string) {
	if errMsg == "" {
		errMsg = "Forbidden: insufficient permissions"
	}
	Error(c, http.StatusForbidden, errMsg)
}

func NotFound(c *gin.Context, errMsg string) {
	if errMsg == "" {
		errMsg = "Resource not found"
	}
	Error(c, http.StatusNotFound, errMsg)
}

func InternalError(c *gin.Context, errMsg string) {
	if errMsg == "" {
		errMsg = "Internal server error"
	}
	Error(c, http.StatusInternalServerError, errMsg)
}
