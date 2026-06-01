package handler

import (
	"encoding/json"
	"io"
	"net/http"

	"noant/internal/infrastructure"
	"noant/internal/service"
	"noant/internal/utils"

	"github.com/gin-gonic/gin"
)

type TelegramHandler struct {
	integration *service.IntegrationService
	logger      *infrastructure.Logger
}

func NewTelegramHandler(integration *service.IntegrationService, logger *infrastructure.Logger) *TelegramHandler {
	return &TelegramHandler{integration: integration, logger: logger}
}

// TelegramWebhook receives incoming messages from Telegram bots.
func (h *TelegramHandler) Webhook(c *gin.Context) {
	rawBody, err := io.ReadAll(c.Request.Body)
	if err != nil {
		h.logger.Error("Failed to read Telegram webhook body", "error", err)
		utils.RespondValidationError(c, "Failed to read request body")
		return
	}

	var update service.TelegramUpdate
	if err := json.Unmarshal(rawBody, &update); err != nil {
		h.logger.Error("Failed to parse Telegram webhook", "error", err)
		utils.RespondValidationError(c, "Invalid payload")
		return
	}

	incoming, ok := update.IncomingMessage()
	if !ok {
		c.JSON(http.StatusOK, gin.H{"success": true})
		return
	}
	if incoming.IsBot {
		c.JSON(http.StatusOK, gin.H{"success": true})
		return
	}

	secret := c.GetHeader("X-Telegram-Bot-Api-Secret-Token")
	if secret == "" {
		h.logger.Warn("Telegram webhook missing secret token")
		utils.RespondUnauthorized(c, "Missing webhook secret")
		return
	}

	integration, resolveErr := h.integration.GetTelegramIntegrationByWebhookSecret(c.Request.Context(), secret)
	if resolveErr != nil {
		h.logger.Error("Failed to resolve Telegram integration", "error", resolveErr)
		utils.RespondInternalError(c, resolveErr.Error())
		return
	}
	if integration == nil {
		h.logger.Warn("No Telegram integration found for webhook secret")
		utils.RespondUnauthorized(c, "Invalid webhook secret")
		return
	}

	_, _, processErr := h.integration.HandleTelegramIncoming(c.Request.Context(), integration, incoming)
	if processErr != nil {
		h.logger.Error("Failed to process Telegram message", "error", processErr)
		utils.RespondInternalError(c, processErr.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}
