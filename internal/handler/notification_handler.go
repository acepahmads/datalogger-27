package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"datalogger/internal/model"
	"datalogger/internal/repository"
	"datalogger/internal/service"
	"datalogger/internal/service/notification"
	"datalogger/pkg/response"

	"github.com/gin-gonic/gin"
)

// NotificationHandler manages HTTP endpoints for notification delivery and channel configuration
type NotificationHandler struct {
	engine *service.NotificationEngine
	repo   *repository.NotificationRepository
}

// NewNotificationHandler constructs a NotificationHandler
func NewNotificationHandler(engine *service.NotificationEngine, repo *repository.NotificationRepository) *NotificationHandler {
	return &NotificationHandler{
		engine: engine,
		repo:   repo,
	}
}

// ChannelDTO represents a channel with secrets safely redacted
type ChannelDTO struct {
	ID               uint                          `json:"id"`
	Name             string                        `json:"name"`
	Type             model.NotificationChannelType `json:"type"`
	Enabled          bool                          `json:"enabled"`
	ConfigSanitized  map[string]interface{}        `json:"config"`
	MinSeverity      model.AlarmSeverity           `json:"min_severity"`
	RateLimitPerMin  int                           `json:"rate_limit_per_min"`
	MaxRetries       int                           `json:"max_retries"`
	RetryIntervalSec int                           `json:"retry_interval_sec"`
	QuietHoursStart  string                        `json:"quiet_hours_start,omitempty"`
	QuietHoursEnd    string                        `json:"quiet_hours_end,omitempty"`
	Timezone         string                        `json:"timezone"`
	IsConfigured     bool                          `json:"is_configured"`
	CreatedAt        time.Time                     `json:"created_at"`
	UpdatedAt        time.Time                     `json:"updated_at"`
}

func toChannelDTO(ch *model.NotificationChannel) ChannelDTO {
	sanitized := notification.SanitizeConfig(ch.ConfigJSON)

	isConfigured := false
	switch ch.Type {
	case model.ChannelWeb:
		isConfigured = true
	case model.ChannelEmail:
		var cfg notification.EmailConfig
		if err := json.Unmarshal([]byte(ch.ConfigJSON), &cfg); err == nil && strings.TrimSpace(cfg.SMTPHost) != "" {
			isConfigured = true
		}
	case model.ChannelTelegram:
		var cfg notification.TelegramConfig
		if err := json.Unmarshal([]byte(ch.ConfigJSON), &cfg); err == nil && strings.TrimSpace(cfg.BotToken) != "" {
			isConfigured = true
		}
	case model.ChannelWebhook:
		var cfg notification.WebhookConfig
		if err := json.Unmarshal([]byte(ch.ConfigJSON), &cfg); err == nil && strings.TrimSpace(cfg.WebhookURL) != "" {
			isConfigured = true
		}
	}

	return ChannelDTO{
		ID:               ch.ID,
		Name:             ch.Name,
		Type:             ch.Type,
		Enabled:          ch.Enabled,
		ConfigSanitized:  sanitized,
		MinSeverity:      ch.MinSeverity,
		RateLimitPerMin:  ch.RateLimitPerMin,
		MaxRetries:       ch.MaxRetries,
		RetryIntervalSec: ch.RetryIntervalSec,
		QuietHoursStart:  ch.QuietHoursStart,
		QuietHoursEnd:    ch.QuietHoursEnd,
		Timezone:         ch.Timezone,
		IsConfigured:     isConfigured,
		CreatedAt:        ch.CreatedAt,
		UpdatedAt:        ch.UpdatedAt,
	}
}

// ListChannels returns all notification channels with secrets redacted
func (h *NotificationHandler) ListChannels(c *gin.Context) {
	channels, err := h.repo.ListChannels(c.Request.Context())
	if err != nil {
		response.InternalError(c, "Failed to query notification channels: "+err.Error())
		return
	}

	dtos := make([]ChannelDTO, 0, len(channels))
	for i := range channels {
		dtos = append(dtos, toChannelDTO(&channels[i]))
	}

	response.OK(c, dtos)
}

// GetChannel retrieves a single notification channel by ID
func (h *NotificationHandler) GetChannel(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid channel ID")
		return
	}

	ch, err := h.repo.GetChannelByID(c.Request.Context(), uint(id))
	if err != nil {
		if err == repository.ErrChannelNotFound {
			response.NotFound(c, "Notification channel not found")
			return
		}
		response.InternalError(c, "Failed to retrieve channel: "+err.Error())
		return
	}

	response.OK(c, toChannelDTO(ch))
}

// UpdateChannelRequest models the payload for updating a channel configuration
type UpdateChannelRequest struct {
	Name             string                 `json:"name"`
	Enabled          *bool                  `json:"enabled"`
	Config           map[string]interface{} `json:"config"`
	MinSeverity      string                 `json:"min_severity"`
	RateLimitPerMin  int                    `json:"rate_limit_per_min"`
	MaxRetries       int                    `json:"max_retries"`
	RetryIntervalSec int                    `json:"retry_interval_sec"`
	QuietHoursStart  string                 `json:"quiet_hours_start"`
	QuietHoursEnd    string                 `json:"quiet_hours_end"`
	Timezone         string                 `json:"timezone"`
}

// UpdateChannel updates settings and credentials for a notification channel
func (h *NotificationHandler) UpdateChannel(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid channel ID")
		return
	}

	existing, err := h.repo.GetChannelByID(c.Request.Context(), uint(id))
	if err != nil {
		if err == repository.ErrChannelNotFound {
			response.NotFound(c, "Notification channel not found")
			return
		}
		response.InternalError(c, "Failed to retrieve channel: "+err.Error())
		return
	}

	var req UpdateChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body: "+err.Error())
		return
	}

	if req.Name != "" {
		existing.Name = strings.TrimSpace(req.Name)
	}
	if req.Enabled != nil {
		existing.Enabled = *req.Enabled
	}
	if req.MinSeverity != "" {
		sev := model.AlarmSeverity(strings.ToUpper(strings.TrimSpace(req.MinSeverity)))
		if sev == model.SeverityInfo || sev == model.SeverityWarning || sev == model.SeverityCritical {
			existing.MinSeverity = sev
		}
	}
	if req.RateLimitPerMin > 0 {
		existing.RateLimitPerMin = req.RateLimitPerMin
	}
	if req.MaxRetries > 0 {
		existing.MaxRetries = req.MaxRetries
	}
	if req.RetryIntervalSec > 0 {
		existing.RetryIntervalSec = req.RetryIntervalSec
	}
	existing.QuietHoursStart = strings.TrimSpace(req.QuietHoursStart)
	existing.QuietHoursEnd = strings.TrimSpace(req.QuietHoursEnd)
	if req.Timezone != "" {
		existing.Timezone = strings.TrimSpace(req.Timezone)
	}

	// Merge secrets if config is provided
	if req.Config != nil {
		newConfigBytes, _ := json.Marshal(req.Config)
		existing.ConfigJSON = notification.MergeSecrets(string(newConfigBytes), existing.ConfigJSON)
	}

	// Validate provider configuration only when the channel is enabled (active),
	// allowing administrators to freely disable/inactivate unconfigured channels.
	if existing.Enabled {
		provider := h.engine.GetProvider(existing.Type)
		if provider != nil {
			if vErr := provider.ValidateConfig(existing.ConfigJSON); vErr != nil {
				response.BadRequest(c, "Invalid channel configuration: "+vErr.Error())
				return
			}
		}
	}

	if err := h.repo.UpdateChannel(c.Request.Context(), existing); err != nil {
		response.InternalError(c, "Failed to update channel: "+err.Error())
		return
	}

	if !existing.Enabled {
		_, _ = h.repo.CancelPendingByChannelID(c.Request.Context(), existing.ID)
	}

	response.OK(c, toChannelDTO(existing))
}

// TestSendRequest models a test notification dispatch request
type TestSendRequest struct {
	Recipient string `json:"recipient"`
}

// TestSend triggers a live diagnostic send through a configured channel
func (h *NotificationHandler) TestSend(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid channel ID")
		return
	}

	var req TestSendRequest
	_ = c.ShouldBindJSON(&req)

	result := h.engine.TestSend(c.Request.Context(), uint(id), req.Recipient)
	if !result.Success {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"success": false,
			"error":   result.ErrorMsg,
			"details": result.ProviderResponse,
		})
		return
	}

	response.OK(c, gin.H{
		"success":  true,
		"response": result.ProviderResponse,
		"message":  "Test notification accepted by provider",
	})
}

// ListNotifications returns delivery queue items and history
func (h *NotificationHandler) ListNotifications(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "50"))
	status := c.Query("status")
	channel := c.Query("channel")
	search := c.Query("search")
	alarmID, _ := strconv.ParseUint(c.Query("alarm_id"), 10, 32)

	var startTime *time.Time
	if sStr := c.Query("start_time"); sStr != "" {
		if t, err := time.Parse(time.RFC3339, sStr); err == nil {
			startTime = &t
		}
	}
	var endTime *time.Time
	if eStr := c.Query("end_time"); eStr != "" {
		if t, err := time.Parse(time.RFC3339, eStr); err == nil {
			endTime = &t
		}
	}

	filter := repository.NotificationFilterParams{
		Status:    status,
		Channel:   channel,
		AlarmID:   uint(alarmID),
		Search:    search,
		StartTime: startTime,
		EndTime:   endTime,
		Page:      page,
		PageSize:  pageSize,
	}

	items, total, err := h.repo.List(c.Request.Context(), filter)
	if err != nil {
		response.InternalError(c, "Failed to list notifications: "+err.Error())
		return
	}

	response.OK(c, gin.H{
		"items":     items,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetNotification retrieves an individual notification with attempt history
func (h *NotificationHandler) GetNotification(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid notification ID")
		return
	}

	n, err := h.repo.GetByID(c.Request.Context(), uint(id))
	if err != nil {
		if err == repository.ErrNotificationNotFound {
			response.NotFound(c, "Notification not found")
			return
		}
		response.InternalError(c, "Failed to retrieve notification: "+err.Error())
		return
	}

	response.OK(c, n)
}

// RetryNotification manually resets a notification to PENDING for immediate redelivery
func (h *NotificationHandler) RetryNotification(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid notification ID")
		return
	}

	n, err := h.engine.RetryNotification(c.Request.Context(), uint(id))
	if err != nil {
		if err == repository.ErrNotificationNotFound {
			response.NotFound(c, "Notification not found")
			return
		}
		response.InternalError(c, "Failed to retry notification: "+err.Error())
		return
	}

	response.OK(c, n)
}

// CancelNotification cancels an in-flight, pending, or retrying notification
func (h *NotificationHandler) CancelNotification(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid notification ID")
		return
	}

	n, err := h.engine.CancelNotification(c.Request.Context(), uint(id))
	if err != nil {
		if err == repository.ErrNotificationNotFound {
			response.NotFound(c, "Notification not found")
			return
		}
		response.InternalError(c, "Failed to cancel notification: "+err.Error())
		return
	}

	response.OK(c, n)
}

// GetMetrics returns real-time engine dispatch statistics
func (h *NotificationHandler) GetMetrics(c *gin.Context) {
	dispatched, delivered, failed, retried, cancelled := h.engine.GetMetrics()
	response.OK(c, gin.H{
		"dispatched_count": dispatched,
		"delivered_count":  delivered,
		"failed_count":     failed,
		"retried_count":    retried,
		"cancelled_count":  cancelled,
	})
}
