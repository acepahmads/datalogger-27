package handler

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"datalogger/internal/model"
	"datalogger/internal/service"
	"datalogger/pkg/response"

	"github.com/gin-gonic/gin"
)

type BackupHandler struct {
	backupService *service.BackupService
	scheduler     *service.BackupScheduler
}

func NewBackupHandler(backupService *service.BackupService, scheduler *service.BackupScheduler) *BackupHandler {
	return &BackupHandler{
		backupService: backupService,
		scheduler:     scheduler,
	}
}

// ListBackups returns paginated list of catalog backups
func (h *BackupHandler) ListBackups(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	status := c.Query("status")

	records, total, err := h.backupService.ListBackups(page, pageSize, status)
	if err != nil {
		response.InternalError(c, fmt.Sprintf("Failed listing backups: %v", err))
		return
	}

	response.OK(c, response.PaginatedData{
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		Items:    records,
	})
}

// GetBackup returns detail and manifest of a single backup
func (h *BackupHandler) GetBackup(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "Backup ID is required")
		return
	}

	record, manifest, err := h.backupService.GetBackupByID(id)
	if err != nil {
		response.NotFound(c, fmt.Sprintf("Backup %s not found", id))
		return
	}

	response.OK(c, gin.H{
		"record":   record,
		"manifest": manifest,
	})
}

type CreateBackupRequest struct {
	Type        string `json:"type"`        // "MANUAL" (default)
	Description string `json:"description"` // optional note
	Async       bool   `json:"async"`       // default true
}

// CreateBackup triggers a new backup operation
func (h *BackupHandler) CreateBackup(c *gin.Context) {
	var req CreateBackupRequest
	_ = c.ShouldBindJSON(&req)

	backupType := model.BackupTypeManual
	if strings.ToUpper(req.Type) == "SCHEDULED" {
		backupType = model.BackupTypeScheduled
	}

	username := c.GetString("username")
	if username == "" {
		username = "admin"
	}

	// Always default to async for smooth UI progress tracking
	jobID, err := h.backupService.StartBackupJob(backupType, username)
	if err != nil {
		if err == service.ErrOperationInProgress {
			c.JSON(http.StatusConflict, response.APIResponse{
				Success: false,
				Error:   err.Error(),
			})
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.Message(c, "Backup operation started in background", gin.H{
		"job_id": jobID,
		"status": "RUNNING",
	})
}

// ValidateBackup validates the integrity and artifacts of a backup archive
func (h *BackupHandler) ValidateBackup(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "Backup ID is required")
		return
	}

	manifest, err := h.backupService.ValidateBackup(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, response.APIResponse{
			Success: false,
			Error:   fmt.Sprintf("Backup validation failed: %v", err),
		})
		return
	}

	response.Message(c, "Backup validated successfully", gin.H{
		"backup_id": id,
		"valid":     true,
		"manifest":  manifest,
	})
}

// GetRestorePreview inspects the backup and compares against active database
func (h *BackupHandler) GetRestorePreview(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "Backup ID is required")
		return
	}

	preview, err := h.backupService.GetRestorePreview(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Error:   fmt.Sprintf("Restore preview failed: %v", err),
		})
		return
	}

	response.OK(c, preview)
}

type RestoreRequest struct {
	Confirm            bool `json:"confirm"`
	CreateSafetyBackup bool `json:"create_safety_backup"`
}

// RestoreBackup executes an authenticated restore workflow
func (h *BackupHandler) RestoreBackup(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "Backup ID is required")
		return
	}

	var req RestoreRequest
	if err := c.ShouldBindJSON(&req); err != nil || !req.Confirm {
		response.BadRequest(c, "Explicit confirmation (confirm: true) is required to restore database")
		return
	}

	username := c.GetString("username")
	if username == "" {
		username = "admin"
	}

	jobID, err := h.backupService.StartRestoreJob(id, username, req.CreateSafetyBackup)
	if err != nil {
		if err == service.ErrOperationInProgress {
			c.JSON(http.StatusConflict, response.APIResponse{
				Success: false,
				Error:   err.Error(),
			})
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.Message(c, "Restore operation started in background", gin.H{
		"job_id":    jobID,
		"backup_id": id,
		"status":    "RUNNING",
	})
}

// GetStatus returns the operational status of the backup engine
func (h *BackupHandler) GetStatus(c *gin.Context) {
	status := h.backupService.GetStatus()
	if h.scheduler != nil {
		schedCfg := h.scheduler.GetScheduleConfig()
		status["schedule"] = schedCfg
		if schedCfg.NextRunTime != nil {
			status["next_scheduled_run"] = schedCfg.NextRunTime.Format("2006-01-02T15:04:05Z07:00")
		}
	}
	response.OK(c, status)
}

// GetJob returns progress and result of an asynchronous operation
func (h *BackupHandler) GetJob(c *gin.Context) {
	jobID := c.Param("jobId")
	if jobID == "" {
		response.BadRequest(c, "Job ID is required")
		return
	}

	job, err := h.backupService.GetJob(jobID)
	if err != nil {
		response.NotFound(c, "Job not found")
		return
	}

	response.OK(c, job)
}

// DeleteBackup removes a backup file and catalog entry
func (h *BackupHandler) DeleteBackup(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "Backup ID is required")
		return
	}

	username := c.GetString("username")
	if username == "" {
		username = "admin"
	}

	if err := h.backupService.DeleteBackup(id, username); err != nil {
		response.InternalError(c, fmt.Sprintf("Failed deleting backup: %v", err))
		return
	}

	response.Message(c, fmt.Sprintf("Backup %s deleted successfully", id), nil)
}

// DownloadBackup streams the backup archive file to authorized client
func (h *BackupHandler) DownloadBackup(c *gin.Context) {
	id := c.Param("id")
	record, _, err := h.backupService.GetBackupByID(id)
	if err != nil || record.FilePath == "" {
		response.NotFound(c, "Backup file not found")
		return
	}

	if _, err := os.Stat(record.FilePath); err != nil {
		response.NotFound(c, "Backup file missing from storage")
		return
	}

	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", record.Filename))
	c.Header("Content-Type", "application/gzip")
	c.File(record.FilePath)
}

// GetSchedule returns current scheduler configuration
func (h *BackupHandler) GetSchedule(c *gin.Context) {
	if h.scheduler == nil {
		response.OK(c, gin.H{"enabled": false})
		return
	}
	response.OK(c, h.scheduler.GetScheduleConfig())
}

type UpdateScheduleRequest struct {
	Enabled       bool   `json:"enabled"`
	IntervalHours int    `json:"interval_hours"`
	TimeOfDay     string `json:"time_of_day"`
	KeepMaxCount  int    `json:"keep_max_count"`
}

// UpdateSchedule updates schedule configuration
func (h *BackupHandler) UpdateSchedule(c *gin.Context) {
	if h.scheduler == nil {
		response.InternalError(c, "Backup scheduler not initialized")
		return
	}

	var req UpdateScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid schedule configuration payload")
		return
	}

	h.scheduler.UpdateScheduleConfig(req.Enabled, req.IntervalHours, req.TimeOfDay, req.KeepMaxCount)
	response.Message(c, "Backup schedule updated successfully", h.scheduler.GetScheduleConfig())
}
