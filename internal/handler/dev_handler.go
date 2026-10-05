package handler

import (
	"strconv"
	"time"

	"datalogger/internal/model"
	"datalogger/internal/service"
	"datalogger/internal/websocket"
	"datalogger/pkg/response"

	"github.com/gin-gonic/gin"
)

type DevHandler struct {
	phaseService *service.PhaseService
	hub          *websocket.Hub
}

func NewDevHandler(phaseService *service.PhaseService, hub *websocket.Hub) *DevHandler {
	return &DevHandler{phaseService: phaseService, hub: hub}
}

func (h *DevHandler) GetPhases(c *gin.Context) {
	phases, err := h.phaseService.GetAllPhases()
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, phases)
}

func (h *DevHandler) GetPhaseByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid phase ID")
		return
	}

	phase, err := h.phaseService.GetPhaseByID(uint(id))
	if err != nil {
		response.NotFound(c, "Phase not found")
		return
	}

	response.OK(c, phase)
}

func (h *DevHandler) GetProgress(c *gin.Context) {
	progress, err := h.phaseService.GetProgressSummary()
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, progress)
}

func (h *DevHandler) GetTasks(c *gin.Context) {
	var phaseID uint
	if pStr := c.Query("phase_id"); pStr != "" {
		if pid, err := strconv.ParseUint(pStr, 10, 32); err == nil {
			phaseID = uint(pid)
		}
	}
	status := c.Query("status")

	tasks, err := h.phaseService.GetAllTasks(phaseID, status)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, tasks)
}

func (h *DevHandler) GetTaskByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid task ID")
		return
	}

	task, err := h.phaseService.GetTaskByID(uint(id))
	if err != nil {
		response.NotFound(c, "Task not found")
		return
	}
	response.OK(c, task)
}

func (h *DevHandler) CreateTask(c *gin.Context) {
	var task model.DevelopmentTask
	if err := c.ShouldBindJSON(&task); err != nil {
		response.BadRequest(c, "Invalid task payload: "+err.Error())
		return
	}

	username := "admin"
	if u, exists := c.Get("username"); exists {
		if uStr, ok := u.(string); ok {
			username = uStr
		}
	}

	if err := h.phaseService.CreateTask(&task, username); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	// Broadcast task creation via WebSocket
	if h.hub != nil {
		h.hub.Broadcast(websocket.WSMessage{
			Type:      "TASK_UPDATE",
			Timestamp: time.Now(),
			Data:      task,
		})
	}

	response.Created(c, task)
}

func (h *DevHandler) UpdateTask(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid task ID")
		return
	}

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		response.BadRequest(c, "Invalid update payload")
		return
	}

	username := "admin"
	if u, exists := c.Get("username"); exists {
		if uStr, ok := u.(string); ok {
			username = uStr
		}
	}

	task, err := h.phaseService.UpdateTask(uint(id), updates, username)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	// Broadcast task update via WebSocket
	if h.hub != nil {
		h.hub.Broadcast(websocket.WSMessage{
			Type:      "TASK_UPDATE",
			Timestamp: time.Now(),
			Data:      task,
		})
	}

	response.OK(c, task)
}

func (h *DevHandler) DeleteTask(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid task ID")
		return
	}

	username := "admin"
	if u, exists := c.Get("username"); exists {
		if uStr, ok := u.(string); ok {
			username = uStr
		}
	}

	if err := h.phaseService.DeleteTask(uint(id), username); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	// Broadcast task update via WebSocket
	if h.hub != nil {
		h.hub.Broadcast(websocket.WSMessage{
			Type:      "TASK_UPDATE",
			Timestamp: time.Now(),
			Data:      gin.H{"deleted_id": id},
		})
	}

	response.OK(c, gin.H{"deleted": true})
}

func (h *DevHandler) GetActivity(c *gin.Context) {
	limit := 30
	if lStr := c.Query("limit"); lStr != "" {
		if l, err := strconv.Atoi(lStr); err == nil {
			limit = l
		}
	}

	activities, err := h.phaseService.GetActivityTimeline(limit)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, activities)
}
