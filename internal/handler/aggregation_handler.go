package handler

import (
	"strconv"
	"time"

	"datalogger/internal/model"
	"datalogger/internal/repository"
	"datalogger/internal/service"
	"datalogger/pkg/response"

	"github.com/gin-gonic/gin"
)

// AggregationHandler exposes REST endpoints for aggregation definitions, rollups, and customer queries
type AggregationHandler struct {
	aggService *service.AggregationService
}

// NewAggregationHandler creates a new AggregationHandler
func NewAggregationHandler(aggService *service.AggregationService) *AggregationHandler {
	return &AggregationHandler{aggService: aggService}
}

// ListDefinitions handles GET /api/aggregations/definitions
func (h *AggregationHandler) ListDefinitions(c *gin.Context) {
	var devID uint
	if dStr := c.Query("device_id"); dStr != "" {
		if d, err := strconv.ParseUint(dStr, 10, 32); err == nil {
			devID = uint(d)
		}
	}
	enabledOnly := c.Query("enabled") == "true"

	defs, err := h.aggService.ListDefinitions(c.Request.Context(), devID, enabledOnly)
	if err != nil {
		response.InternalError(c, "Failed to list aggregation definitions: "+err.Error())
		return
	}
	response.OK(c, defs)
}

// GetDefinition handles GET /api/aggregations/definitions/:id
func (h *AggregationHandler) GetDefinition(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid definition ID format")
		return
	}

	def, err := h.aggService.GetDefinition(c.Request.Context(), uint(id))
	if err != nil {
		response.NotFound(c, "Aggregation definition not found")
		return
	}
	response.OK(c, def)
}

// CreateDefinition handles POST /api/aggregations/definitions
func (h *AggregationHandler) CreateDefinition(c *gin.Context) {
	var req service.CreateAggregationDefinitionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request payload: "+err.Error())
		return
	}

	username, _ := c.Get("username")
	uStr, _ := username.(string)
	ip := c.ClientIP()
	ua := c.Request.UserAgent()

	def, err := h.aggService.CreateDefinition(c.Request.Context(), &req, uStr, ip, ua)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Created(c, def)
}

// UpdateDefinition handles PUT /api/aggregations/definitions/:id
func (h *AggregationHandler) UpdateDefinition(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid definition ID format")
		return
	}

	var req service.UpdateAggregationDefinitionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request payload: "+err.Error())
		return
	}

	username, _ := c.Get("username")
	uStr, _ := username.(string)
	ip := c.ClientIP()
	ua := c.Request.UserAgent()

	def, err := h.aggService.UpdateDefinition(c.Request.Context(), uint(id), &req, uStr, ip, ua)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, def)
}

// DeleteDefinition handles DELETE /api/aggregations/definitions/:id
func (h *AggregationHandler) DeleteDefinition(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid definition ID format")
		return
	}

	username, _ := c.Get("username")
	uStr, _ := username.(string)
	ip := c.ClientIP()
	ua := c.Request.UserAgent()

	if err := h.aggService.DeleteDefinition(c.Request.Context(), uint(id), uStr, ip, ua); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, gin.H{"message": "Aggregation definition deleted successfully"})
}

// TriggerBucket handles POST /api/aggregations/definitions/:id/run for manual recalculation
func (h *AggregationHandler) TriggerBucket(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid definition ID format")
		return
	}

	var body struct {
		Timestamp *time.Time `json:"timestamp"`
	}
	_ = c.ShouldBindJSON(&body)

	targetTime := time.Now().UTC()
	if body.Timestamp != nil && !body.Timestamp.IsZero() {
		targetTime = *body.Timestamp
	}

	result, err := h.aggService.CalculateBucket(c.Request.Context(), uint(id), targetTime)
	if err != nil {
		response.InternalError(c, "Failed to calculate aggregation bucket: "+err.Error())
		return
	}
	response.OK(c, result)
}

// RunBuckets handles POST /api/aggregations/run for manual recalculation of definitions
func (h *AggregationHandler) RunBuckets(c *gin.Context) {
	var body struct {
		DeviceID     uint `json:"device_id"`
		DefinitionID uint `json:"definition_id"`
	}
	_ = c.ShouldBindJSON(&body)

	if body.DefinitionID > 0 {
		def, err := h.aggService.GetDefinition(c.Request.Context(), body.DefinitionID)
		if err != nil {
			response.BadRequest(c, "Aggregation definition not found")
			return
		}
		count, err := h.aggService.ProcessPendingBucketsForDefinition(c.Request.Context(), def, 24*time.Hour)
		if err != nil {
			response.InternalError(c, "Failed to calculate aggregation bucket: "+err.Error())
			return
		}
		response.OK(c, gin.H{"processed_buckets": count, "definition_id": body.DefinitionID})
		return
	}

	if body.DeviceID > 0 {
		defs, err := h.aggService.ListDefinitions(c.Request.Context(), body.DeviceID, true)
		if err != nil {
			response.InternalError(c, "Failed to list definitions: "+err.Error())
			return
		}
		totalProcessed := 0
		for _, def := range defs {
			count, _ := h.aggService.ProcessPendingBucketsForDefinition(c.Request.Context(), &def, 24*time.Hour)
			totalProcessed += count
		}
		response.OK(c, gin.H{"processed_buckets": totalProcessed, "device_id": body.DeviceID})
		return
	}

	total, err := h.aggService.RunAllActiveDefinitions(c.Request.Context())
	if err != nil {
		response.InternalError(c, "Failed to run aggregations: "+err.Error())
		return
	}
	response.OK(c, gin.H{"processed_buckets": total})
}

// GetResults handles GET /api/aggregations/results
func (h *AggregationHandler) GetResults(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "50"))

	var defID, devID, paramID uint
	if str := c.Query("definition_id"); str != "" {
		if v, err := strconv.ParseUint(str, 10, 32); err == nil {
			defID = uint(v)
		}
	}
	if str := c.Query("device_id"); str != "" {
		if v, err := strconv.ParseUint(str, 10, 32); err == nil {
			devID = uint(v)
		}
	}
	if str := c.Query("parameter_id"); str != "" {
		if v, err := strconv.ParseUint(str, 10, 32); err == nil {
			paramID = uint(v)
		}
	}

	sourceType := model.AggregationSourceType(c.Query("source_type"))
	identifier := c.Query("identifier")
	quality := c.Query("quality")

	var startTime, endTime *time.Time
	if sStr := c.Query("start_time"); sStr != "" {
		if t, err := time.Parse(time.RFC3339, sStr); err == nil {
			startTime = &t
		}
	}
	if eStr := c.Query("end_time"); eStr != "" {
		if t, err := time.Parse(time.RFC3339, eStr); err == nil {
			endTime = &t
		}
	}

	results, total, err := h.aggService.GetResults(c.Request.Context(), repository.AggregationResultFilter{
		DefinitionID: defID,
		DeviceID:     devID,
		ParameterID:  paramID,
		SourceType:   sourceType,
		Identifier:   identifier,
		StartTime:    startTime,
		EndTime:      endTime,
		Quality:      quality,
		Page:         page,
		PageSize:     pageSize,
	})
	if err != nil {
		response.InternalError(c, "Failed to query aggregation results: "+err.Error())
		return
	}

	response.OK(c, response.PaginatedData{
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		Items:    results,
	})
}

// GetCustomerAggregatedData handles GET /api/customer/aggregated-data/:identifier
func (h *AggregationHandler) GetCustomerAggregatedData(c *gin.Context) {
	identifier := c.Param("identifier")
	if identifier == "" {
		response.BadRequest(c, "Period identifier is required")
		return
	}

	data, err := h.aggService.GetCustomerAggregatedData(c.Request.Context(), identifier)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.OK(c, data)
}

// GetDownsampledHistory handles GET /api/devices/:id/telemetry/downsampled
func (h *AggregationHandler) GetDownsampledHistory(c *gin.Context) {
	devID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid device ID format")
		return
	}

	paramStr := c.Query("parameter_id")
	if paramStr == "" {
		response.BadRequest(c, "parameter_id query parameter is required")
		return
	}
	paramID, err := strconv.ParseUint(paramStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid parameter ID format")
		return
	}

	resolution := c.DefaultQuery("resolution", "auto")

	now := time.Now().UTC()
	startTime := now.Add(-24 * time.Hour)
	endTime := now

	if sStr := c.Query("start_time"); sStr != "" {
		if t, err := time.Parse(time.RFC3339, sStr); err == nil {
			startTime = t
		}
	}
	if eStr := c.Query("end_time"); eStr != "" {
		if t, err := time.Parse(time.RFC3339, eStr); err == nil {
			endTime = t
		}
	}

	data, err := h.aggService.GetDownsampledHistory(c.Request.Context(), uint(devID), uint(paramID), startTime, endTime, resolution)
	if err != nil {
		response.InternalError(c, "Downsampling calculation failed: "+err.Error())
		return
	}
	response.OK(c, data)
}

// GetResultSamples handles GET /api/aggregations/results/:id/samples and GET /api/aggregations/samples
func (h *AggregationHandler) GetResultSamples(c *gin.Context) {
	idStr := c.Param("id")
	var id uint
	if idStr != "" {
		if v, err := strconv.ParseUint(idStr, 10, 32); err == nil {
			id = uint(v)
		}
	}

	var devID, paramID uint
	if dStr := c.Query("device_id"); dStr != "" {
		if v, err := strconv.ParseUint(dStr, 10, 32); err == nil {
			devID = uint(v)
		}
	}
	if pStr := c.Query("parameter_id"); pStr != "" {
		if v, err := strconv.ParseUint(pStr, 10, 32); err == nil {
			paramID = uint(v)
		}
	}

	var startTime, endTime *time.Time
	if sStr := c.Query("period_start"); sStr != "" {
		if t, err := time.Parse(time.RFC3339, sStr); err == nil {
			startTime = &t
		}
	}
	if eStr := c.Query("period_end"); eStr != "" {
		if t, err := time.Parse(time.RFC3339, eStr); err == nil {
			endTime = &t
		}
	}

	resp, err := h.aggService.GetResultSamples(c.Request.Context(), id, devID, paramID, startTime, endTime)
	if err != nil {
		response.InternalError(c, "Failed to fetch bucket samples: "+err.Error())
		return
	}
	response.OK(c, resp)
}

