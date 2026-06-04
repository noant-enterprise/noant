package handler

import (
	"net/http"
	"strconv"

	"noant/internal/infrastructure"
	"noant/internal/service"

	"github.com/gin-gonic/gin"
)

type BackgroundHandler struct {
	worker *service.BackgroundWorker
	logger *infrastructure.Logger
}

func NewBackgroundHandler(worker *service.BackgroundWorker, logger *infrastructure.Logger) *BackgroundHandler {
	return &BackgroundHandler{worker: worker, logger: logger}
}

func (h *BackgroundHandler) SubmitTask(c *gin.Context) {
	var req struct {
		TaskName string                 `json:"task_name" binding:"required"`
		Payload  map[string]interface{} `json:"payload"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "task_name is required"})
		return
	}

	taskID := h.worker.SubmitTask(req.TaskName, req.Payload)
	c.JSON(http.StatusAccepted, gin.H{
		"task_id": taskID,
		"status":  "submitted",
	})
}

func (h *BackgroundHandler) GetTaskStatus(c *gin.Context) {
	taskID := c.Param("id")
	if taskID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "task ID is required"})
		return
	}

	task := h.worker.GetTask(taskID)
	if task == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"task": task,
	})
}

func (h *BackgroundHandler) ListTasks(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "20")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 20
	}

	tasks := h.worker.ListTasks(limit)
	c.JSON(http.StatusOK, gin.H{
		"tasks": tasks,
		"count": len(tasks),
	})
}

func (h *BackgroundHandler) WorkerStats(c *gin.Context) {
	stats := h.worker.Stats()
	c.JSON(http.StatusOK, gin.H{
		"stats": stats,
	})
}
