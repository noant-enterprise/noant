package handler

import (
	"net/http"
	"strconv"

	"noant/internal/infrastructure"
	"noant/internal/repository"
	"noant/internal/service"
	"noant/internal/utils"

	"github.com/gin-gonic/gin"
)

type AuditHandler struct {
	service *service.AuditService
	logger  *infrastructure.Logger
}

func NewAuditHandler(svc *service.AuditService, logger *infrastructure.Logger) *AuditHandler {
	return &AuditHandler{service: svc, logger: logger}
}

func (h *AuditHandler) ListLogs(c *gin.Context) {
	userID := getScopeID(c)
	if userID == "" {
		utils.RespondUnauthorized(c, "Unauthorized")
		return
	}
	limit := 50

	logs, err := h.service.ListByUser(c.Request.Context(), userID, limit)
	if err != nil {
		h.logger.Error("Failed to list audit logs", "error", err)
		utils.RespondInternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"audit_logs": logs, "count": len(logs)})
}

func (h *AuditHandler) SearchLogs(c *gin.Context) {
	userID := getUserID(c)
	if userID == "" {
		utils.RespondUnauthorized(c, "Unauthorized")
		return
	}

	filter := &repository.AuditFilter{
		OrgID:        getOrgID(c),
		UserID:       userID,
		Action:       c.Query("action"),
		ResourceType: c.Query("resource_type"),
		StartDate:    c.Query("start_date"),
		EndDate:      c.Query("end_date"),
	}

	if v := c.Query("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			filter.Limit = n
		}
	}
	if v := c.Query("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			filter.Offset = n
		}
	}

	result, err := h.service.ListWithFilters(c.Request.Context(), filter)
	if err != nil {
		h.logger.Error("Failed to search audit logs", "error", err)
		utils.RespondInternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"audit_logs": result.Logs,
		"total":      result.Total,
		"limit":      filter.Limit,
		"offset":     filter.Offset,
	})
}
