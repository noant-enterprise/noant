package handler

import (
	"net/http"

	"noant/internal/infrastructure"
	"noant/internal/service"
	"noant/internal/utils"

	"github.com/gin-gonic/gin"
)

type AssistantHandler struct {
	svc    *service.AssistantService
	logger *infrastructure.Logger
}

func NewAssistantHandler(svc *service.AssistantService, logger *infrastructure.Logger) *AssistantHandler {
	return &AssistantHandler{svc: svc, logger: logger}
}

func (h *AssistantHandler) Chat(c *gin.Context) {
	userID := getUserID(c)
	userEmail, _ := c.Get("email")

	var req struct {
		Message string `json:"message" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondValidationError(c, "Message is required")
		return
	}

	resp, err := h.svc.Chat(c.Request.Context(), req.Message)
	if err != nil {
		h.logger.Error("Assistant chat failed", "error", err, "user_id", userID, "email", userEmail)
		utils.RespondInternalError(c, "")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"content":     resp.Content,
		"action":      resp.Action,
		"steps":       resp.Steps,
		"suggestions": resp.Suggestions,
	})
}
