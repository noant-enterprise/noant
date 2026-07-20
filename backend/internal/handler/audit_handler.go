package handler

import (
	"net/http"

	"noant/internal/infrastructure"
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
	userID, _ := c.Get("userID")
	limit := 50 // default limit

	logs, err := h.service.ListByUser(c.Request.Context(), userID.(string), limit)
	if err != nil {
		h.logger.Error("Failed to list audit logs", "error", err)
		utils.RespondInternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"audit_logs": logs, "count": len(logs)})
}
