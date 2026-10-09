package handler

import (
	"fmt"
	"strconv"

	"datalogger/internal/service"
	"datalogger/pkg/response"

	"github.com/gin-gonic/gin"
)

type RetentionHandler struct {
	retentionService   *service.RetentionService
	retentionScheduler *service.RetentionScheduler
}

func NewRetentionHandler(retentionService *service.RetentionService, retentionScheduler *service.RetentionScheduler) *RetentionHandler {
	return &RetentionHandler{
		retentionService:   retentionService,
		retentionScheduler: retentionScheduler,
	}
}

// ListPolicies returns all configured retention policies
func (h *RetentionHandler) ListPolicies(c *gin.Context) {
	policies, err := h.retentionService.GetAllPolicies()
	if err != nil {
		response.InternalError(c, fmt.Sprintf("Failed retrieving retention policies: %v", err))
		return
	}
	response.OK(c, policies)
}

// GetPolicy returns a single policy by ID
func (h *RetentionHandler) GetPolicy(c *gin.Context) {
	id := c.Param("id")
	policy, err := h.retentionService.GetPolicyByID(id)
	if err != nil {
		response.NotFound(c, "Retention policy not found")
		return
	}
	response.OK(c, policy)
}

type UpdatePolicyRequest struct {
	Name                  string `json:"name"`
	Description           string `json:"description"`
	Enabled               bool   `json:"enabled"`
	RetentionDays         int    `json:"retention_days"`
	MinimumAgeHours       int    `json:"minimum_age_hours"`
	ProtectedPeriodDays   int    `json:"protected_period_days"`
	RequireBackup         bool   `json:"require_backup"`
	RequireRollup         bool   `json:"require_rollup"`
	BatchSize             int    `json:"batch_size"`
	MaxDeletePerRun       int64  `json:"max_delete_per_run"`
	Priority              int    `json:"priority"`
	ScheduleIntervalHours int    `json:"schedule_interval_hours"`
	ScheduleTime          string `json:"schedule_time"`
}

// UpdatePolicy updates editable policy configuration
func (h *RetentionHandler) UpdatePolicy(c *gin.Context) {
	id := c.Param("id")
	policy, err := h.retentionService.GetPolicyByID(id)
	if err != nil {
		response.NotFound(c, "Retention policy not found")
		return
	}

	var req UpdatePolicyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, fmt.Sprintf("Invalid request payload: %v", err))
		return
	}

	username := c.GetString("username")
	if username == "" {
		username = "admin"
	}

	if req.Name != "" {
		policy.Name = req.Name
	}
	if req.Description != "" {
		policy.Description = req.Description
	}
	policy.Enabled = req.Enabled
	if req.RetentionDays > 0 {
		policy.RetentionDays = req.RetentionDays
	}
	if req.MinimumAgeHours > 0 {
		policy.MinimumAgeHours = req.MinimumAgeHours
	}
	if req.ProtectedPeriodDays >= 0 {
		policy.ProtectedPeriodDays = req.ProtectedPeriodDays
	}
	policy.RequireBackup = req.RequireBackup
	policy.RequireRollup = req.RequireRollup
	if req.BatchSize > 0 {
		policy.BatchSize = req.BatchSize
	}
	if req.MaxDeletePerRun > 0 {
		policy.MaxDeletePerRun = req.MaxDeletePerRun
	}
	if req.Priority > 0 {
		policy.Priority = req.Priority
	}
	if req.ScheduleIntervalHours > 0 {
		policy.ScheduleIntervalHours = req.ScheduleIntervalHours
	}
	if req.ScheduleTime != "" {
		policy.ScheduleTime = req.ScheduleTime
	}

	if err := h.retentionService.SavePolicy(policy, username); err != nil {
		response.BadRequest(c, fmt.Sprintf("Cannot save policy: %v", err))
		return
	}

	response.OK(c, policy)
}

type TogglePolicyRequest struct {
	Enabled bool `json:"enabled"`
}

// TogglePolicy enables or disables a policy
func (h *RetentionHandler) TogglePolicy(c *gin.Context) {
	id := c.Param("id")
	var req TogglePolicyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request payload")
		return
	}

	username := c.GetString("username")
	if username == "" {
		username = "admin"
	}

	if err := h.retentionService.TogglePolicy(id, req.Enabled, username); err != nil {
		response.InternalError(c, fmt.Sprintf("Failed toggling policy: %v", err))
		return
	}

	response.OK(c, gin.H{"id": id, "enabled": req.Enabled})
}

// DryRun previews the effect of running a retention policy without modifying any records
func (h *RetentionHandler) DryRun(c *gin.Context) {
	id := c.Param("id")
	result, err := h.retentionService.DryRun(c.Request.Context(), id)
	if err != nil {
		response.BadRequest(c, fmt.Sprintf("Dry-run preview failed: %v", err))
		return
	}
	response.OK(c, result)
}

type ExecutePolicyRequest struct {
	Confirm       bool `json:"confirm"`
	ForceOverride bool `json:"force_override"`
}

// ExecutePolicy initiates explicit bounded deletion of records older than retention threshold
func (h *RetentionHandler) ExecutePolicy(c *gin.Context) {
	id := c.Param("id")

	var req ExecutePolicyRequest
	if err := c.ShouldBindJSON(&req); err != nil || !req.Confirm {
		response.BadRequest(c, "Explicit confirmation (confirm: true) is strictly required for destructive retention execution")
		return
	}

	username := c.GetString("username")
	if username == "" {
		username = "admin"
	}

	execLog, err := h.retentionService.ExecutePolicy(c.Request.Context(), id, username, req.ForceOverride)
	if err != nil {
		c.JSON(200, gin.H{
			"success": false,
			"error":   err.Error(),
			"data":    execLog,
		})
		return
	}

	response.OK(c, execLog)
}

// TriggerHousekeeping runs a manual housekeeping sweep across all enabled policies
func (h *RetentionHandler) TriggerHousekeeping(c *gin.Context) {
	username := c.GetString("username")
	if username == "" {
		username = "admin"
	}

	if h.retentionScheduler == nil {
		response.InternalError(c, "Housekeeping scheduler is not initialized")
		return
	}

	results := h.retentionScheduler.RunHousekeepingCycle(c.Request.Context(), username)
	response.OK(c, gin.H{
		"results":     results,
		"total_runs":  len(results),
		"executed_by": username,
	})
}

// GetStorageOverview returns comprehensive database, table, WAL, and filesystem metrics
func (h *RetentionHandler) GetStorageOverview(c *gin.Context) {
	overview, err := h.retentionService.GetStorageOverview()
	if err != nil {
		response.InternalError(c, fmt.Sprintf("Failed compiling storage overview: %v", err))
		return
	}
	response.OK(c, overview)
}

// GetHistory returns paginated execution history
func (h *RetentionHandler) GetHistory(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	policyID := c.Query("policy_id")

	logs, total, err := h.retentionService.GetExecutionLogs(page, pageSize, policyID)
	if err != nil {
		response.InternalError(c, fmt.Sprintf("Failed fetching execution history: %v", err))
		return
	}

	response.OK(c, response.PaginatedData{
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		Items:    logs,
	})
}
