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

// ========== NOTIFICATION HANDLER ==========

type NotificationHandler struct {
	service *service.NotificationService
	logger  *infrastructure.Logger
}

func NewNotificationHandler(svc *service.NotificationService, logger *infrastructure.Logger) *NotificationHandler {
	return &NotificationHandler{service: svc, logger: logger}
}

func (h *NotificationHandler) List(c *gin.Context) {
	userID := getUserID(c)
	if userID == "" {
		utils.RespondUnauthorized(c, "Unauthorized")
		return
	}
	limit := 50
	if lStr := c.Query("limit"); lStr != "" {
		if l, err := strconv.Atoi(lStr); err == nil && l > 0 {
			limit = l
		}
	}
	if limit > 200 {
		limit = 200
	}

	notifs, err := h.service.List(c.Request.Context(), userID, limit)
	if err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"notifications": notifs})
}

func (h *NotificationHandler) UnreadCount(c *gin.Context) {
	userID := getUserID(c)
	if userID == "" {
		utils.RespondUnauthorized(c, "Unauthorized")
		return
	}
	count, err := h.service.UnreadCount(c.Request.Context(), userID)
	if err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"count": count})
}

func (h *NotificationHandler) MarkRead(c *gin.Context) {
	userID := getUserID(c)
	if userID == "" {
		utils.RespondUnauthorized(c, "Unauthorized")
		return
	}
	id := c.Param("id")

	err := h.service.MarkRead(c.Request.Context(), id, userID)
	if err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Notification marked as read"})
}

func (h *NotificationHandler) MarkAllRead(c *gin.Context) {
	userID := getUserID(c)
	if userID == "" {
		utils.RespondUnauthorized(c, "Unauthorized")
		return
	}

	err := h.service.MarkAllRead(c.Request.Context(), userID)
	if err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "All notifications marked as read"})
}

// ========== WIDGET HANDLER ==========

type WidgetHandler struct {
	service *service.WidgetService
	logger  *infrastructure.Logger
}

func NewWidgetHandler(svc *service.WidgetService, logger *infrastructure.Logger) *WidgetHandler {
	return &WidgetHandler{service: svc, logger: logger}
}

func (h *WidgetHandler) Get(c *gin.Context) {
	userID := getUserID(c)
	if userID == "" {
		utils.RespondUnauthorized(c, "Unauthorized")
		return
	}
	cfg, err := h.service.Get(c.Request.Context(), userID)
	if err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, cfg)
}

func (h *WidgetHandler) Upsert(c *gin.Context) {
	userID := getUserID(c)
	if userID == "" {
		utils.RespondUnauthorized(c, "Unauthorized")
		return
	}
	var req struct {
		BrandColor string `json:"brand_color"`
		Greeting   string `json:"greeting"`
		BotName    string `json:"bot_name"`
		Position   string `json:"position"`
		IsActive   bool   `json:"is_active"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondValidationError(c, err.Error())
		return
	}
	utils.SanitizeStruct(&req)

	cfg, err := h.service.Get(c.Request.Context(), userID)
	if err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}

	cfg.BrandColor = req.BrandColor
	cfg.Greeting = req.Greeting
	cfg.BotName = req.BotName
	cfg.Position = req.Position
	cfg.IsActive = req.IsActive

	err = h.service.Upsert(c.Request.Context(), cfg)
	if err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Widget configuration updated", "config": cfg})
}

func (h *WidgetHandler) GetPublic(c *gin.Context) {
	apiKey := c.Query("api_key")
	if apiKey == "" {
		utils.RespondValidationError(c, "api_key is required")
		return
	}

	cfg, err := h.service.GetByAPIKey(c.Request.Context(), apiKey)
	if err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}
	if cfg == nil {
		utils.RespondNotFound(c, "Widget configuration not found or inactive")
		return
	}

	c.JSON(http.StatusOK, cfg)
}

func (h *WidgetHandler) PublicChat(c *gin.Context) {
	var req struct {
		APIKey         string `json:"api_key" binding:"required"`
		Message        string `json:"message" binding:"required"`
		ConversationID string `json:"conversation_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondValidationError(c, err.Error())
		return
	}
	utils.SanitizeStruct(&req)

	reply, convID, err := h.service.PublicChat(c.Request.Context(), req.APIKey, req.Message, req.ConversationID)
	if err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"reply":           reply,
		"conversation_id": convID,
	})
}

// ========== SETTINGS EXTENSIONS ==========

func (h *SettingsHandler) GetNotifPrefs(c *gin.Context) {
	userID := getUserID(c)
	if userID == "" {
		utils.RespondUnauthorized(c, "Unauthorized")
		return
	}
	prefs, err := h.service.GetNotifPrefs(c.Request.Context(), userID)
	if err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, prefs)
}

func (h *SettingsHandler) UpdateNotifPrefs(c *gin.Context) {
	userID := getUserID(c)
	if userID == "" {
		utils.RespondUnauthorized(c, "Unauthorized")
		return
	}
	var req repository.NotifPrefs
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondValidationError(c, err.Error())
		return
	}
	utils.SanitizeStruct(&req)

	err := h.service.UpdateNotifPrefs(c.Request.Context(), userID, &req)
	if err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Notification preferences updated"})
}

func (h *SettingsHandler) DeleteAccount(c *gin.Context) {
	userID := getUserID(c)
	if userID == "" {
		utils.RespondUnauthorized(c, "Unauthorized")
		return
	}
	err := h.service.DeleteAccount(c.Request.Context(), userID)
	if err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Account deleted successfully"})
}

func (h *SettingsHandler) ExportData(c *gin.Context) {
	userID := getUserID(c)
	if userID == "" {
		utils.RespondUnauthorized(c, "Unauthorized")
		return
	}
	data, err := h.service.ExportUserData(c.Request.Context(), userID)
	if err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, data)
}
