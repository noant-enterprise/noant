package handler

import (
	"net/http"

	"noant/internal/infrastructure"
	"noant/internal/service"

	"github.com/gin-gonic/gin"
)

type DBManagerHandler struct {
	dbManager *service.DBManagerService
	logger    *infrastructure.Logger
}

func NewDBManagerHandler(dbManager *service.DBManagerService, logger *infrastructure.Logger) *DBManagerHandler {
	return &DBManagerHandler{dbManager: dbManager, logger: logger}
}

func (h *DBManagerHandler) RunAllCleanups(c *gin.Context) {
	config := service.DefaultCleanupConfig()
	report := h.dbManager.RunAllCleanups(c.Request.Context(), config)
	status := http.StatusOK
	if report.HasErrors {
		status = http.StatusOK
	}
	c.JSON(status, gin.H{
		"report": report,
	})
}

func (h *DBManagerHandler) RunCleanupTask(c *gin.Context) {
	var req struct {
		TaskName string `json:"task_name" binding:"required"`
		Days     int    `json:"days"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "task_name is required"})
		return
	}

	config := service.DefaultCleanupConfig()
	if req.Days > 0 {
		switch req.TaskName {
		case "old_resolved_conversations":
			config.OldConversationsDays = req.Days
		case "abandoned_conversations":
			config.AbandonedConversationsDays = req.Days
		case "stale_unknown_questions":
			config.UnknownQuestionsDays = req.Days
		case "expired_handoffs":
			config.HandoffsDays = req.Days
		case "old_audit_logs":
			config.AuditLogsDays = req.Days
		case "old_notifications":
			config.NotificationsDays = req.Days
		case "stale_inactive_integrations":
			config.InactiveIntegrationDays = req.Days
		case "expired_trials":
			config.ExpiredTrialDays = req.Days
		case "stale_credit_purchases":
			config.CreditPurchasesDays = req.Days
		case "completed_campaigns":
			config.CompletedCampaignsDays = req.Days
		}
	}

	result := h.dbManager.RunTask(c.Request.Context(), req.TaskName, config)
	status := http.StatusOK
	if result.Status == "failed" {
		status = http.StatusInternalServerError
	}
	c.JSON(status, gin.H{
		"task": result,
	})
}

func (h *DBManagerHandler) GetCleanupConfig(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"config": service.DefaultCleanupConfig(),
	})
}

func (h *DBManagerHandler) ListCleanupTasks(c *gin.Context) {
	tasks := []map[string]interface{}{
		{"name": "all", "description": "Run all cleanup tasks"},
		{"name": "old_resolved_conversations", "description": "Delete resolved conversations older than N days"},
		{"name": "abandoned_conversations", "description": "Resolve active conversations with no recent messages"},
		{"name": "orphaned_messages", "description": "Delete messages with no parent conversation"},
		{"name": "stale_unknown_questions", "description": "Delete trained/ignored unknown questions older than N days"},
		{"name": "expired_handoffs", "description": "Auto-expire pending handoffs older than N days"},
		{"name": "old_audit_logs", "description": "Delete audit logs older than N days"},
		{"name": "old_notifications", "description": "Delete old notifications older than N days"},
		{"name": "stale_inactive_integrations", "description": "Delete old inactive integrations"},
		{"name": "expired_trials", "description": "Deactivate expired free trials"},
		{"name": "expired_credits", "description": "Delete expired credit balances"},
		{"name": "stale_credit_purchases", "description": "Delete old credit purchase records"},
		{"name": "completed_campaigns", "description": "Delete completed/canceled campaigns older than N days"},
	}
	c.JSON(http.StatusOK, gin.H{"tasks": tasks})
}
