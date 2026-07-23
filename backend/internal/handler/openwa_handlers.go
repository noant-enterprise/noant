package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"noant/internal/domain"
	"noant/internal/infrastructure"
	"noant/internal/service"
	"noant/internal/utils"

	"github.com/gin-gonic/gin"
)

func getUserID(c *gin.Context) string {
	if userID, ok := c.Get("userID"); ok {
		if s, ok := userID.(string); ok {
			return s
		}
	}
	return ""
}

func getOrgID(c *gin.Context) string {
	if orgID, ok := c.Get("orgID"); ok {
		if s, ok := orgID.(string); ok {
			return s
		}
	}
	return ""
}

// ========== TEMPLATE HANDLER ==========

type TemplateHandler struct {
	templateSvc *service.TemplateService
	logger      *infrastructure.Logger
}

func NewTemplateHandler(templateSvc *service.TemplateService, logger *infrastructure.Logger) *TemplateHandler {
	return &TemplateHandler{templateSvc: templateSvc, logger: logger}
}

func (h *TemplateHandler) List(c *gin.Context) {
	userID := 	getScopeID(c)
	if userID == "" {
		utils.RespondUnauthorized(c, "Unauthorized")
		return
	}
	templates, err := h.templateSvc.List(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to list templates", "error", err)
		utils.RespondInternalError(c, "Failed to list templates")
		return
	}
	c.JSON(http.StatusOK, gin.H{"templates": templates})
}

func (h *TemplateHandler) Create(c *gin.Context) {
	userID := 	getScopeID(c)
	if userID == "" {
		utils.RespondUnauthorized(c, "Unauthorized")
		return
	}
	var req service.CreateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondValidationError(c, "Invalid request: "+err.Error())
		return
	}
	utils.SanitizeStruct(&req)
	tpl, err := h.templateSvc.Create(c.Request.Context(), userID, &req)
	if err != nil {
		h.logger.Error("Failed to create template", "error", err)
		utils.RespondInternalError(c, "Failed to create template")
		return
	}
	c.JSON(http.StatusCreated, gin.H{"template": tpl})
}

func (h *TemplateHandler) GetByID(c *gin.Context) {
	userID := 	getScopeID(c)
	if userID == "" {
		utils.RespondUnauthorized(c, "Unauthorized")
		return
	}
	id := c.Param("id")
	if id == "" {
		utils.RespondValidationError(c, "Template ID required")
		return
	}
	tpl, err := h.templateSvc.GetByID(c.Request.Context(), id, userID)
	if err != nil {
		h.logger.Error("Failed to get template", "error", err)
		utils.RespondInternalError(c, "Failed to get template")
		return
	}
	if tpl == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Template not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"template": tpl})
}

func (h *TemplateHandler) Update(c *gin.Context) {
	userID := 	getScopeID(c)
	if userID == "" {
		utils.RespondUnauthorized(c, "Unauthorized")
		return
	}
	id := c.Param("id")
	if id == "" {
		utils.RespondValidationError(c, "Template ID required")
		return
	}
	var req service.CreateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondValidationError(c, "Invalid request: "+err.Error())
		return
	}
	utils.SanitizeStruct(&req)
	tpl := &domain.WhatsAppTemplate{
		ID: id, Name: req.Name, Language: req.Language, Category: req.Category,
		HeaderType: req.HeaderType, HeaderValue: req.HeaderValue,
		BodyText: req.BodyText, FooterText: req.FooterText,
	}
	if req.Buttons != nil {
		b, _ := json.Marshal(req.Buttons)
		tpl.Buttons = string(b)
	}
	if err := h.templateSvc.Update(c.Request.Context(), userID, tpl); err != nil {
		h.logger.Error("Failed to update template", "error", err)
		utils.RespondInternalError(c, "Failed to update template")
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *TemplateHandler) Delete(c *gin.Context) {
	userID := 	getScopeID(c)
	if userID == "" {
		utils.RespondUnauthorized(c, "Unauthorized")
		return
	}
	id := c.Param("id")
	if id == "" {
		utils.RespondValidationError(c, "Template ID required")
		return
	}
	if err := h.templateSvc.Delete(c.Request.Context(), id, userID); err != nil {
		h.logger.Error("Failed to delete template", "error", err)
		utils.RespondInternalError(c, "Failed to delete template")
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *TemplateHandler) SubmitForApproval(c *gin.Context) {
	userID := 	getScopeID(c)
	if userID == "" {
		utils.RespondUnauthorized(c, "Unauthorized")
		return
	}
	id := c.Param("id")
	if id == "" {
		utils.RespondValidationError(c, "Template ID required")
		return
	}
	if err := h.templateSvc.SubmitForApproval(c.Request.Context(), id, userID); err != nil {
		h.logger.Error("Failed to submit template", "error", err)
		utils.RespondInternalError(c, "Failed to submit template")
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Template submitted for approval"})
}

func (h *TemplateHandler) Send(c *gin.Context) {
	userID := 	getScopeID(c)
	if userID == "" {
		utils.RespondUnauthorized(c, "Unauthorized")
		return
	}
	var req service.SendTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondValidationError(c, "Invalid request: "+err.Error())
		return
	}
	utils.SanitizeStruct(&req)
	if err := h.templateSvc.SendTemplate(c.Request.Context(), req); err != nil {
		h.logger.Error("Failed to send template", "error", err)
		utils.RespondInternalError(c, "Failed to send template: "+err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *TemplateHandler) GetCommon(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"templates": service.GetCommonTemplates()})
}

// ========== OPENWA EXTENDED HANDLERS ==========

func (h *OpenWAHandler) BroadcastCampaign(c *gin.Context) {
	userID := 	getScopeID(c)
	if userID == "" {
		utils.RespondUnauthorized(c, "Unauthorized")
		return
	}
	var req struct {
		CampaignID string   `json:"campaign_id"`
		SessionID  string   `json:"session_id"`
		Message    string   `json:"message"`
		Recipients []string `json:"recipients"`
		TemplateID string   `json:"template_id,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondValidationError(c, "Invalid request: "+err.Error())
		return
	}
	utils.SanitizeStruct(&req)
	if len(req.Recipients) == 0 {
		utils.RespondValidationError(c, "At least one recipient required")
		return
	}
	if req.Message == "" && req.TemplateID == "" {
		utils.RespondValidationError(c, "Message or template_id required")
		return
	}

	bridge := service.NewCampaignBridge(h.cfg, h.openwa, nil, h.logger, nil, nil, nil, nil, nil)
	err := bridge.ExecuteCampaign(c.Request.Context(), &service.BroadcastRequest{
		CampaignID: req.CampaignID,
		UserID:     userID,
		SessionID:  req.SessionID,
		Message:    req.Message,
		Recipients: req.Recipients,
		TemplateID: req.TemplateID,
	})
	if err != nil {
		h.logger.Error("Campaign broadcast failed", "error", err)
		utils.RespondInternalError(c, "Broadcast failed: "+err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "recipients": len(req.Recipients)})
}

func (h *OpenWAHandler) CampaignAnalytics(c *gin.Context) {
	userID := 	getScopeID(c)
	if userID == "" {
		utils.RespondUnauthorized(c, "Unauthorized")
		return
	}
	campaignID := c.Param("campaignID")
	if campaignID == "" {
		utils.RespondValidationError(c, "Campaign ID required")
		return
	}
	bridge := service.NewCampaignBridge(h.cfg, h.openwa, nil, h.logger, nil, nil, nil, nil, nil)
	analytics, err := bridge.GetCampaignAnalytics(c.Request.Context(), campaignID)
	if err != nil {
		h.logger.Error("Failed to get campaign analytics", "error", err)
		utils.RespondInternalError(c, "Failed to get analytics")
		return
	}
	c.JSON(http.StatusOK, analytics)
}

func (h *OpenWAHandler) UploadMedia(c *gin.Context) {
	userID := 	getScopeID(c)
	if userID == "" {
		utils.RespondUnauthorized(c, "Unauthorized")
		return
	}
	convID := c.Param("id")
	if convID == "" {
		utils.RespondValidationError(c, "Conversation ID required")
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		utils.RespondValidationError(c, "File required")
		return
	}
	defer func() { _ = file.Close() }()

	allowedMIMETypes := map[string]bool{
		"image/jpeg":         true,
		"image/png":          true,
		"image/gif":          true,
		"image/webp":         true,
		"application/pdf":    true,
		"audio/mpeg":         true,
		"audio/ogg":          true,
		"audio/wav":          true,
		"video/mp4":          true,
		"video/quicktime":    true,
	}

	allowedExtensions := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".webp": true,
		".pdf":  true,
		".mp3":  true,
		".ogg":  true,
		".wav":  true,
		".mp4":  true,
		".mov":  true,
	}

	if header.Size > 10485760 {
		utils.RespondValidationError(c, "Invalid file: file too large (max 10MB)")
		return
	}

	ext := strings.ToLower(filepath.Ext(header.Filename))
	if !allowedExtensions[ext] {
		utils.RespondValidationError(c, "Invalid file: file extension not allowed")
		return
	}

	contentType := header.Header.Get("Content-Type")
	if contentType != "" {
		mimeType, _, err := mime.ParseMediaType(contentType)
		if err != nil || !allowedMIMETypes[mimeType] {
			utils.RespondValidationError(c, "Invalid file: MIME type not allowed")
			return
		}
	}

	mediaDir := h.cfg.OpenWAMediaDir
	if err := os.MkdirAll(mediaDir, 0o750); err != nil {
		utils.RespondInternalError(c, "Failed to create media directory")
		return
	}

	filename := fmt.Sprintf("upload_%s_%d%s", userID, time.Now().UnixNano(), filepath.Ext(header.Filename))
	filePath := filepath.Join(mediaDir, filename)

	out, err := os.Create(filePath)
	if err != nil {
		utils.RespondInternalError(c, "Failed to save file")
		return
	}
	defer func() { _ = out.Close() }()

	if _, err := io.Copy(out, file); err != nil {
		utils.RespondInternalError(c, "Failed to write file")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"file_url": "/api/v1/chats/conversations/media/" + filename,
		"filename": header.Filename,
		"size":     header.Size,
	})
}

func (h *OpenWAHandler) ListMedia(c *gin.Context) {
	userID := 	getScopeID(c)
	if userID == "" {
		utils.RespondUnauthorized(c, "Unauthorized")
		return
	}
	convID := c.Param("id")
	if convID == "" {
		utils.RespondValidationError(c, "Conversation ID required")
		return
	}
	media, err := h.chat.GetMediaByConversation(c.Request.Context(), convID, userID)
	if err != nil {
		h.logger.Error("Failed to list media", "error", err)
		utils.RespondInternalError(c, "Failed to list media")
		return
	}
	c.JSON(http.StatusOK, gin.H{"media": media})
}

func (h *OpenWAHandler) GetMedia(c *gin.Context) {
	mediaID := c.Param("mediaID")
	if mediaID == "" {
		utils.RespondValidationError(c, "Media ID required")
		return
	}
	// Prevent path traversal — allow only safe filename characters
	if !regexp.MustCompile(`^[a-zA-Z0-9._-]+$`).MatchString(mediaID) {
		utils.RespondValidationError(c, "Invalid media ID")
		return
	}
	// Serve the file directly
	mediaDir := h.cfg.OpenWAMediaDir
	filePath := filepath.Join(mediaDir, mediaID)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Media not found"})
		return
	}
	c.File(filePath)
}

func (h *OpenWAHandler) GetMediaThumbnail(c *gin.Context) {
	mediaID := c.Param("mediaID")
	if mediaID == "" {
		utils.RespondValidationError(c, "Media ID required")
		return
	}
	// Prevent path traversal — allow only safe filename characters
	if !regexp.MustCompile(`^[a-zA-Z0-9._-]+$`).MatchString(mediaID) {
		utils.RespondValidationError(c, "Invalid media ID")
		return
	}
	mediaDir := h.cfg.OpenWAMediaDir
	thumbPath := filepath.Join(mediaDir, "thumb_"+mediaID)
	if _, err := os.Stat(thumbPath); os.IsNotExist(err) {
		// Fall back to original
		c.File(filepath.Join(mediaDir, mediaID))
		return
	}
	c.File(thumbPath)
}

func (h *OpenWAHandler) QueueStats(c *gin.Context) {
	stats := h.openwa.GetQueueStats()
	c.JSON(http.StatusOK, stats)
}

func (h *OpenWAHandler) ListManagedSessions(c *gin.Context) {
	sessions := h.openwa.ListManagedSessions()
	c.JSON(http.StatusOK, gin.H{"sessions": sessions})
}

func (h *OpenWAHandler) SessionMetrics(c *gin.Context) {
	sessionID := c.Param("sessionID")
	if sessionID == "" {
		utils.RespondValidationError(c, "Session ID required")
		return
	}
	metrics := h.openwa.GetSessionMetrics(sessionID)
	c.JSON(http.StatusOK, metrics)
}

func (h *OpenWAHandler) ForceReconnect(c *gin.Context) {
	sessionID := c.Param("sessionID")
	if sessionID == "" {
		utils.RespondValidationError(c, "Session ID required")
		return
	}
	if err := h.openwa.RestartSessionByID(sessionID); err != nil {
		h.logger.Error("Force reconnect failed", "error", err)
		utils.RespondInternalError(c, "Reconnect failed: "+err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Reconnect initiated"})
}

func (h *OpenWAHandler) SendListMessage(c *gin.Context) {
	var req struct {
		SessionID  string   `json:"session_id"`
		ChatID     string   `json:"chat_id"`
		Header     string   `json:"header"`
		Body       string   `json:"body"`
		Footer     string   `json:"footer"`
		ButtonText string   `json:"button_text"`
		Items      []service.ListItem `json:"items"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondValidationError(c, "Invalid request: "+err.Error())
		return
	}
	utils.SanitizeStruct(&req)

	tplSvc := service.NewTemplateService(h.cfg, h.openwa, nil, h.logger, nil)
	if err := tplSvc.SendListMessage(req.SessionID, req.ChatID, req.Header, req.Body, req.Footer, req.ButtonText, req.Items); err != nil {
		h.logger.Error("Failed to send list message", "error", err)
		utils.RespondInternalError(c, "Failed to send: "+err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *OpenWAHandler) SendButtonsMessage(c *gin.Context) {
	var req struct {
		SessionID string              `json:"session_id"`
		ChatID    string              `json:"chat_id"`
		Body      string              `json:"body"`
		Buttons   []service.ReplyButton `json:"buttons"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondValidationError(c, "Invalid request: "+err.Error())
		return
	}
	utils.SanitizeStruct(&req)

	tplSvc := service.NewTemplateService(h.cfg, h.openwa, nil, h.logger, nil)
	if err := tplSvc.SendButtonsMessage(req.SessionID, req.ChatID, req.Body, req.Buttons); err != nil {
		h.logger.Error("Failed to send buttons message", "error", err)
		utils.RespondInternalError(c, "Failed to send: "+err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}
