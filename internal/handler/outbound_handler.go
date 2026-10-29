package handler

import (
	"strconv"
	"strings"
	"time"

	"datalogger/internal/model"
	"datalogger/internal/repository"
	"datalogger/internal/service"
	"datalogger/pkg/response"

	"github.com/gin-gonic/gin"
)

// OutboundHandler provides REST endpoints for destinations, configurable JSON payload schemas, and delivery queue inspection
type OutboundHandler struct {
	service service.OutboundService
}

// NewOutboundHandler creates a new handler instance
func NewOutboundHandler(service service.OutboundService) *OutboundHandler {
	return &OutboundHandler{service: service}
}

// GetDestinations handles GET /api/destinations
func (h *OutboundHandler) GetDestinations(c *gin.Context) {
	enabledParam := c.Query("enabled_only")
	enabledOnly := enabledParam == "true" || enabledParam == "1"

	// Non-superadmin users get redacted secrets
	role, _ := c.Get("role")
	roleStr, _ := role.(string)
	redactSecrets := strings.ToUpper(roleStr) != "SUPERADMIN"

	dests, err := h.service.GetDestinations(c.Request.Context(), enabledOnly, redactSecrets)
	if err != nil {
		response.InternalError(c, "Failed to retrieve destinations: "+err.Error())
		return
	}
	response.OK(c, dests)
}

// GetDestinationByID handles GET /api/destinations/:id
func (h *OutboundHandler) GetDestinationByID(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid destination ID")
		return
	}

	role, _ := c.Get("role")
	roleStr, _ := role.(string)
	redactSecrets := strings.ToUpper(roleStr) != "SUPERADMIN"

	dest, err := h.service.GetDestinationByID(c.Request.Context(), uint(id), redactSecrets)
	if err != nil {
		response.NotFound(c, "Destination not found")
		return
	}
	response.OK(c, dest)
}

// CreateDestination handles POST /api/destinations
func (h *OutboundHandler) CreateDestination(c *gin.Context) {
	var dest model.OutputDestination
	if err := c.ShouldBindJSON(&dest); err != nil {
		response.BadRequest(c, "Invalid request body: "+err.Error())
		return
	}

	if dest.Name == "" || dest.Endpoint == "" || dest.Type == "" {
		response.BadRequest(c, "Name, Endpoint, and Type are required fields")
		return
	}

	if err := h.service.CreateDestination(c.Request.Context(), &dest); err != nil {
		response.InternalError(c, "Failed to create destination: "+err.Error())
		return
	}
	response.Created(c, dest)
}

// UpdateDestination handles PUT /api/destinations/:id
func (h *OutboundHandler) UpdateDestination(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid destination ID")
		return
	}

	var dest model.OutputDestination
	if err := c.ShouldBindJSON(&dest); err != nil {
		response.BadRequest(c, "Invalid request body: "+err.Error())
		return
	}
	dest.ID = uint(id)

	if err := h.service.UpdateDestination(c.Request.Context(), &dest); err != nil {
		response.InternalError(c, "Failed to update destination: "+err.Error())
		return
	}
	response.OK(c, dest)
}

// DeleteDestination handles DELETE /api/destinations/:id
func (h *OutboundHandler) DeleteDestination(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid destination ID")
		return
	}

	if err := h.service.DeleteDestination(c.Request.Context(), uint(id)); err != nil {
		response.InternalError(c, "Failed to delete destination: "+err.Error())
		return
	}
	response.OK(c, gin.H{"deleted": true})
}

// GetPayloadConfig handles GET /api/destinations/:id/payload-config
func (h *OutboundHandler) GetPayloadConfig(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid destination ID")
		return
	}

	cfg, err := h.service.GetPayloadConfig(c.Request.Context(), uint(id))
	if err != nil {
		// Return default scaffold if config record doesn't exist yet
		cfg = &model.DestinationPayloadConfig{
			DestinationID:   uint(id),
			TemplateType:    "STANDARD",
			StationIDField:  "station_id",
			TimestampField:  "timestamp",
			TimestampFormat: "ISO8601",
			IntervalField:   "interval_minutes",
			Structure:       "FLAT",
			NestedDataKey:   "data",
			IncludeUnits:    false,
			IncludeQuality:  false,
			IncludeMetadata: true,
			Version:         1,
		}
	}
	response.OK(c, cfg)
}

// UpdatePayloadConfig handles PUT /api/destinations/:id/payload-config
func (h *OutboundHandler) UpdatePayloadConfig(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid destination ID")
		return
	}

	var cfg model.DestinationPayloadConfig
	if err := c.ShouldBindJSON(&cfg); err != nil {
		response.BadRequest(c, "Invalid request body: "+err.Error())
		return
	}
	cfg.DestinationID = uint(id)

	if err := h.service.UpdatePayloadConfig(c.Request.Context(), &cfg); err != nil {
		response.BadRequest(c, "Configuration validation failed: "+err.Error())
		return
	}
	response.OK(c, cfg)
}

// GeneratePreview handles GET and POST /api/destinations/:id/preview
func (h *OutboundHandler) GeneratePreview(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid destination ID")
		return
	}

	var deviceID *uint
	if devParam := c.Query("device_id"); devParam != "" {
		if dID, err := strconv.ParseUint(devParam, 10, 32); err == nil {
			uDID := uint(dID)
			deviceID = &uDID
		}
	}

	var customCfg *model.DestinationPayloadConfig
	if c.Request.ContentLength > 0 {
		var cfg model.DestinationPayloadConfig
		if err := c.ShouldBindJSON(&cfg); err == nil {
			customCfg = &cfg
		}
	}

	preview, err := h.service.GeneratePreviewWithConfig(c.Request.Context(), uint(id), deviceID, customCfg)
	if err != nil {
		response.BadRequest(c, "Failed to generate preview: "+err.Error())
		return
	}
	response.OK(c, gin.H{"preview": preview})
}

// TestPayload handles POST /api/destinations/:id/test
func (h *OutboundHandler) TestPayload(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid destination ID")
		return
	}

	result, err := h.service.TestPayload(c.Request.Context(), uint(id))
	if err != nil {
		response.BadRequest(c, "Test execution failed: "+err.Error())
		return
	}
	response.OK(c, result)
}

// GetQueueItems handles GET /api/integration/queue
func (h *OutboundHandler) GetQueueItems(c *gin.Context) {
	filter := repository.QueueFilter{
		Status: c.Query("status"),
		Search: c.Query("search"),
	}

	if destIDStr := c.Query("destination_id"); destIDStr != "" {
		if dID, err := strconv.ParseUint(destIDStr, 10, 32); err == nil {
			uID := uint(dID)
			filter.DestinationID = &uID
		}
	}

	if devIDStr := c.Query("device_id"); devIDStr != "" {
		if dID, err := strconv.ParseUint(devIDStr, 10, 32); err == nil {
			uID := uint(dID)
			filter.DeviceID = &uID
		}
	}

	if intStr := c.Query("interval_seconds"); intStr != "" {
		if interval, err := strconv.Atoi(intStr); err == nil {
			filter.IntervalSeconds = &interval
		}
	}

	if startStr := c.Query("start_date"); startStr != "" {
		if t, err := time.Parse(time.RFC3339, startStr); err == nil {
			filter.StartDate = &t
		}
	}

	if endStr := c.Query("end_date"); endStr != "" {
		if t, err := time.Parse(time.RFC3339, endStr); err == nil {
			filter.EndDate = &t
		}
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	filter.Page = page
	filter.PageSize = pageSize

	items, total, err := h.service.GetQueueItems(c.Request.Context(), filter)
	if err != nil {
		response.InternalError(c, "Failed to query queue: "+err.Error())
		return
	}

	response.OK(c, response.PaginatedData{
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		Items:    items,
	})
}

// GetQueueSummary handles GET /api/integration/queue/summary
func (h *OutboundHandler) GetQueueSummary(c *gin.Context) {
	summary, err := h.service.GetQueueSummary(c.Request.Context())
	if err != nil {
		response.InternalError(c, "Failed to get queue summary: "+err.Error())
		return
	}
	response.OK(c, summary)
}

// GetQueueItemDetail handles GET /api/integration/queue/:id
func (h *OutboundHandler) GetQueueItemDetail(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid queue item ID")
		return
	}

	detail, err := h.service.GetQueueItemDetail(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, "Queue item not found")
		return
	}
	response.OK(c, detail)
}

// RetryQueueItem handles POST /api/integration/queue/:id/retry
func (h *OutboundHandler) RetryQueueItem(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid queue item ID")
		return
	}

	if err := h.service.RetryQueueItem(c.Request.Context(), id); err != nil {
		response.InternalError(c, "Failed to retry queue item: "+err.Error())
		return
	}
	response.OK(c, gin.H{"status": "PENDING", "message": "Queue item reset for immediate retry"})
}

// ReplayQueueItem handles POST /api/integration/queue/:id/replay
func (h *OutboundHandler) ReplayQueueItem(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid queue item ID")
		return
	}

	newItem, err := h.service.ReplayQueueItem(c.Request.Context(), id)
	if err != nil {
		response.InternalError(c, "Failed to replay queue item: "+err.Error())
		return
	}
	response.Created(c, newItem)
}

// CancelQueueItem handles POST /api/integration/queue/:id/cancel
func (h *OutboundHandler) CancelQueueItem(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid queue item ID")
		return
	}

	if err := h.service.CancelQueueItem(c.Request.Context(), id); err != nil {
		response.InternalError(c, "Failed to cancel queue item: "+err.Error())
		return
	}
	response.OK(c, gin.H{"status": "CANCELLED", "message": "Queue item cancelled successfully"})
}

// DeleteQueueItem handles DELETE /api/integration/queue/:id
func (h *OutboundHandler) DeleteQueueItem(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid queue item ID")
		return
	}

	if err := h.service.DeleteQueueItem(c.Request.Context(), id); err != nil {
		response.InternalError(c, "Failed to delete queue item: "+err.Error())
		return
	}
	response.OK(c, gin.H{"message": "Queue item deleted successfully"})
}

// PurgeQueueRequest defines payload for POST /api/integration/queue/purge
type PurgeQueueRequest struct {
	DestinationID *uint      `json:"destination_id"`
	Status        string     `json:"status"` // ALL, FAILED, DELIVERED, CANCELLED, etc.
	OlderThan     *time.Time `json:"older_than"`
	PurgeLogs     bool       `json:"purge_logs"`
}

// PurgeQueue handles POST /api/integration/queue/purge
func (h *OutboundHandler) PurgeQueue(c *gin.Context) {
	var req PurgeQueueRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// allow empty body as well (purge all)
		req.Status = "ALL"
	}

	filter := repository.QueuePurgeFilter{
		DestinationID: req.DestinationID,
		Status:        req.Status,
		OlderThan:     req.OlderThan,
		PurgeLogs:     req.PurgeLogs,
	}

	deletedQueue, deletedLogs, err := h.service.PurgeQueue(c.Request.Context(), filter)
	if err != nil {
		response.InternalError(c, "Failed to purge queue: "+err.Error())
		return
	}

	response.OK(c, gin.H{
		"message":       "Queue purge completed successfully",
		"deleted_queue": deletedQueue,
		"deleted_logs":  deletedLogs,
	})
}
