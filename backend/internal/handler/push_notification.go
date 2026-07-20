package handler

import (
	"net/http"

	"noant/internal/infrastructure"
	"noant/internal/service"
	"noant/internal/utils"

	"github.com/gin-gonic/gin"
)

type PushHandler struct {
	service *service.PushNotificationService
	logger  *infrastructure.Logger
}

func NewPushHandler(svc *service.PushNotificationService, logger *infrastructure.Logger) *PushHandler {
	return &PushHandler{service: svc, logger: logger}
}

func (h *PushHandler) Subscribe(c *gin.Context) {
	userID, _ := c.Get("userID")

	var req struct {
		Endpoint string `json:"endpoint" binding:"required"`
		Auth     string `json:"auth" binding:"required"`
		P256dh   string `json:"p256dh" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondValidationError(c, "endpoint, auth, and p256dh are required")
		return
	}
	utils.SanitizeStruct(&req)

	userAgent := c.GetHeader("User-Agent")

	if err := h.service.Subscribe(c.Request.Context(), userID.(string), req.Endpoint, req.Auth, req.P256dh, userAgent); err != nil {
		h.logger.Error("Push subscribe failed", "error", err)
		utils.RespondInternalError(c, "Failed to subscribe")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Subscribed to push notifications"})
}

func (h *PushHandler) Unsubscribe(c *gin.Context) {
	userID, _ := c.Get("userID")

	var req struct {
		Endpoint string `json:"endpoint"`
	}
	_ = c.ShouldBindJSON(&req)

	if req.Endpoint != "" {
		if err := h.service.Unsubscribe(c.Request.Context(), userID.(string), req.Endpoint); err != nil {
			h.logger.Error("Push unsubscribe failed", "error", err)
			utils.RespondInternalError(c, "Failed to unsubscribe")
			return
		}
	} else {
		// Delete all subscriptions for this user (fallback)
		_ = h.service.Unsubscribe(c.Request.Context(), userID.(string), "")
	}

	c.JSON(http.StatusOK, gin.H{"message": "Unsubscribed from push notifications"})
}
