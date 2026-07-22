package handler

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"noant/config"
	"noant/internal/infrastructure"
	"noant/internal/repository"
	"noant/internal/service"
	"noant/internal/utils"

	"github.com/gin-gonic/gin"
)

type OpenWAHandler struct {
	cfg    *config.Config
	openwa *service.OpenWAService
	chat   *service.ChatService
	repos  *repository.Repositories
	logger *infrastructure.Logger
	wsHub  *WebSocketHub
}

func NewOpenWAHandler(cfg *config.Config, openwa *service.OpenWAService, chat *service.ChatService, repos *repository.Repositories, logger *infrastructure.Logger, wsHub *WebSocketHub) *OpenWAHandler {
	return &OpenWAHandler{cfg: cfg, openwa: openwa, chat: chat, repos: repos, logger: logger, wsHub: wsHub}
}

// WhatsAppWebhook receives incoming messages from OpenWA
func (h *OpenWAHandler) WhatsAppWebhook(c *gin.Context) {
	const maxBodySize = 10 << 20 // 10 MB
	rawBody, err := io.ReadAll(io.LimitReader(c.Request.Body, maxBodySize+1))
	if err != nil {
		h.logger.Error("Failed to read OpenWA webhook body", "error", err)
		utils.RespondValidationError(c, "Failed to read request body")
		return
	}
	if int64(len(rawBody)) > maxBodySize {
		utils.RespondError(c, http.StatusRequestEntityTooLarge, "BODY_TOO_LARGE", "Request body too large", false)
		return
	}

	// Verify HMAC signature — required when secret is configured
	signature := c.GetHeader("X-Hub-Signature-256")
	if signature == "" {
		h.logger.Warn("OpenWA webhook missing signature header")
		utils.RespondUnauthorized(c, "Missing signature")
		return
	}
	if !h.openwa.VerifyWebhookSignature(rawBody, signature) {
		h.logger.Warn("OpenWA webhook signature verification failed")
		utils.RespondUnauthorized(c, "Invalid signature")
		return
	}

	// Parse webhook event
	event, err := h.openwa.ParseWebhookEvent(rawBody)
	if err != nil {
		h.logger.Error("Failed to parse OpenWA webhook", "error", err)
		utils.RespondValidationError(c, "Invalid payload")
		return
	}

	h.logger.Info("OpenWA webhook received", "event", event.Event, "session", event.SessionID)

	switch event.Event {
	case "message.received":
		h.handleIncomingMessage(c, event)
	case "message.status":
		h.handleMessageStatus(c, event)
	default:
		h.logger.Info("Unhandled OpenWA event", "event", event.Event)
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// handleIncomingMessage processes an incoming WhatsApp message
func (h *OpenWAHandler) handleIncomingMessage(c *gin.Context, event *service.OpenWAWebhookPayload) {
	msg, err := h.openwa.ParseMessageData(event.Data)
	if err != nil {
		h.logger.Error("Failed to parse incoming message", "error", err)
		return
	}

	// Ignore messages sent by us
	if msg.FromMe {
		return
	}

	chatID := msg.From
	customerPhone := service.CleanPhoneNumber(chatID)

	integration, err := h.chat.GetWhatsAppIntegrationBySessionID(c.Request.Context(), event.SessionID)
	if err != nil {
		h.logger.Error("Failed to resolve WhatsApp integration", "error", err, "session", event.SessionID)
		return
	}
	if integration == nil {
		h.logger.Warn("No WhatsApp integration found for session", "session", event.SessionID)
		return
	}
	userID := integration.UserID
	if userID == "" {
		h.logger.Warn("WhatsApp integration has no user owner", "session", event.SessionID)
		return
	}

	// Resolve identity using multiple fallbacks so a partial payload still becomes a usable customer profile.
	identity, err := h.chat.ResolveWhatsAppIdentity(c.Request.Context(), userID, event.SessionID, msg)
	if err != nil {
		h.logger.Warn("WhatsApp identity resolution failed, using basic fallback", "error", err)
	}
	customerName := customerPhone
	customerAvatar := ""
	if identity != nil {
		if identity.Name != "" {
			customerName = identity.Name
		}
		if identity.Phone != "" {
			customerPhone = identity.Phone
		}
		if identity.Avatar != "" {
			customerAvatar = identity.Avatar
		}
		h.logger.Info("WhatsApp identity resolved", "methods", identity.Methods, "name", customerName, "phone", customerPhone, "avatar", customerAvatar != "")
	} else {
		if msg.Sender.Pushname != "" {
			customerName = msg.Sender.Pushname
		} else if msg.Sender.Name != "" {
			customerName = msg.Sender.Name
		}
		customerAvatar = msg.Sender.ProfilePicThumbObj.Eurl
	}
	if customerName == "" {
		customerName = customerPhone
	}

	// Handle media messages (image, document, audio, video, etc.)
	if msg.HasMedia && msg.MediaURL != "" {
		h.logger.Info("Processing incoming media message", "type", msg.Type, "mediaType", msg.MediaType, "from", customerPhone)
		mediaHandler := h.openwa.GetMediaHandler()
		filePath, thumbPath, err := mediaHandler.HandleIncomingMedia(c.Request.Context(), event.SessionID, userID, &service.OpenWAMediaData{
			HasMedia:  true,
			MediaURL:  msg.MediaURL,
			MediaType: msg.MediaType,
			MimeType:  msg.MimeType,
			FileName:  msg.FileName,
			FileSize:  msg.FileSize,
			Width:     msg.Width,
			Height:    msg.Height,
			Duration:  msg.Duration,
		})
		if err != nil {
			h.logger.Error("Failed to process incoming media", "error", err)
		} else {
			conv, _ := h.chat.EnsureConversation(c.Request.Context(), userID, customerName, customerPhone, "whatsapp", customerAvatar)
			if conv != nil {
				_ = h.chat.StoreMediaRecord(c.Request.Context(), conv.ID, userID, event.SessionID, msg)
			}
			h.logger.Info("Incoming media stored", "path", filePath, "thumb", thumbPath)
		}
		// Don't send AI reply for media-only messages
		return
	}

	// Ignore non-text messages that aren't media
	if msg.Type != "text" && msg.Type != "chat" && msg.Type != "" {
		h.logger.Info("Ignoring non-text message", "type", msg.Type)
		return
	}

	content := msg.Body
	h.logger.Info("OpenWA incoming message", "from", customerPhone, "body", content)

	// Process through chat service (same flow as web widget)
	conv, aiResp, err := h.chat.DirectChat(c.Request.Context(), userID, customerName, customerPhone, content, "whatsapp", customerAvatar)
	if err != nil {
		h.logger.Error("Failed to process OpenWA message", "error", err)
		return
	}

	// Send AI reply back via OpenWA (routed through queue for rate-limit + retry protection)
	if aiResp != nil && aiResp.Content != "" {
		queue := h.openwa.GetQueue()
		workerPool := h.openwa.GetWorkerPool()
		if queue != nil && workerPool != nil {
			workerPool.EnsureWorker(event.SessionID)
			entry := &service.QueueEntry{
				ID:        fmt.Sprintf("ai_%s_%d", conv.ID, time.Now().UnixNano()),
				SessionID: event.SessionID,
				UserID:    userID,
				ChatID:    chatID,
				MsgType:   service.MsgTypeText,
				Content:   aiResp.Content,
				Priority:  service.PriorityNormal,
			}
			if err := queue.Enqueue(entry); err != nil {
				h.logger.Error("Failed to enqueue AI reply", "error", err, "chatID", chatID)
			}
		} else {
			// Fallback to direct send if queue unavailable
			if err := h.openwa.SendTextMessage(event.SessionID, chatID, aiResp.Content); err != nil {
				h.logger.Error("Failed to send OpenWA reply", "error", err, "chatID", chatID)
			}
		}
	}

	// Broadcast to WebSocket dashboard
	if h.wsHub != nil && conv != nil && aiResp != nil {
		h.wsHub.BroadcastMessage(WebSocketMessage{
			ConversationID: conv.ID,
			Type:           "new_message",
			Data: map[string]interface{}{
				"content":     aiResp.Content,
				"sender_type": "ai",
				"customer":    customerName,
				"channel":     "whatsapp",
			},
		})
	}
}

// handleMessageStatus handles delivery/read status updates
func (h *OpenWAHandler) handleMessageStatus(c *gin.Context, event *service.OpenWAWebhookPayload) {
	status, err := h.openwa.ParseStatusData(event.Data)
	if err != nil {
		h.logger.Error("Failed to parse status update", "error", err)
		return
	}

	h.logger.Info("WhatsApp delivery status",
		"messageID", status.ID,
		"status", status.Status,
		"sessionID", event.SessionID,
	)

	// Fix 2: persist delivery status to database
	h.persistDeliveryStatus(c.Request.Context(), status.ID, status.Status, event.SessionID)
}

// persistDeliveryStatus stores the delivery status in the message_deliveries table
func (h *OpenWAHandler) persistDeliveryStatus(ctx context.Context, messageID, status, sessionID string) {
	if messageID == "" {
		return
	}

	// Update conversation messages table if messageID matches
	if h.repos != nil && h.repos.DB != nil {
		query := `UPDATE messages SET delivery_status = ? WHERE external_id = ? AND delivery_status != ?`
		if _, err := h.repos.DB.ExecContext(ctx, query, status, messageID, status); err != nil {
			h.logger.Warn("Failed to update message delivery status", "messageID", messageID, "status", status, "error", err)
		}
	}

	// Track metrics
	infrastructure.OpenWADeliveryStatusTotal.WithLabelValues(status).Inc()
}

// GetSessionStatus returns the WhatsApp session status
func (h *OpenWAHandler) GetSessionStatus(c *gin.Context) {
	if userID, ok := c.Get("userID"); ok {
		if integration, err := h.chat.GetWhatsAppIntegration(c.Request.Context(), userID.(string)); err == nil && integration != nil {
			if sessionID, _ := integration.Config["session_id"].(string); sessionID != "" {
				status, err := h.openwa.GetSessionStatusByID(sessionID)
				if err != nil {
					h.logger.Error("Failed to get OpenWA session status", "error", err, "sessionID", sessionID)
					utils.RespondInternalError(c, err.Error())
					return
				}
				c.JSON(http.StatusOK, gin.H{"status": status, "session_id": sessionID})
				return
			}
		}
	}

	status, err := h.openwa.GetSessionStatus()
	if err != nil {
		h.logger.Error("Failed to get OpenWA session status", "error", err)
		utils.RespondInternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": status})
}

// RestartSession restarts the WhatsApp session
func (h *OpenWAHandler) RestartSession(c *gin.Context) {
	if userID, ok := c.Get("userID"); ok {
		if integration, err := h.chat.GetWhatsAppIntegration(c.Request.Context(), userID.(string)); err == nil && integration != nil {
			if sessionID, _ := integration.Config["session_id"].(string); sessionID != "" {
				if err := h.openwa.RestartSessionByID(sessionID); err != nil {
					h.logger.Error("Failed to restart OpenWA session", "error", err, "sessionID", sessionID)
					utils.RespondInternalError(c, err.Error())
					return
				}
				c.JSON(http.StatusOK, gin.H{"message": "Session restart initiated", "session_id": sessionID})
				return
			}
		}
	}

	if err := h.openwa.RestartSession(); err != nil {
		h.logger.Error("Failed to restart OpenWA session", "error", err)
		utils.RespondInternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Session restart initiated"})
}

// HealthCheck returns OpenWA server status
func (h *OpenWAHandler) HealthCheck(c *gin.Context) {
	err := h.openwa.Ping()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"status": "unhealthy",
			"error":  err.Error(),
			"openwa": h.cfg.OpenWABaseURL,
		})
		return
	}

	sessions, _ := h.openwa.ListSessions()
	c.JSON(http.StatusOK, gin.H{
		"status":   "healthy",
		"openwa":   h.cfg.OpenWABaseURL,
		"sessions": len(sessions),
	})
}

// PhonePing sends a test message to verify WhatsApp connection
func (h *OpenWAHandler) PhonePing(c *gin.Context) {
	var req struct {
		Phone string `json:"phone" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondValidationError(c, "Phone number is required")
		return
	}
	utils.SanitizeStruct(&req)

	// Find active session
	userID, ok := c.Get("userID")
	if !ok {
		utils.RespondUnauthorized(c, "Unauthorized")
		return
	}

	integration, err := h.chat.GetWhatsAppIntegration(c.Request.Context(), userID.(string))
	if err != nil {
		h.logger.Error("Failed to load WhatsApp integration", "error", err)
		utils.RespondInternalError(c, "Failed to load WhatsApp integration")
		return
	}
	if integration == nil {
		utils.RespondInternalError(c, "No active WhatsApp integration. Connect first.")
		return
	}

	sessionID, _ := integration.Config["session_id"].(string)
	if sessionID == "" {
		utils.RespondInternalError(c, "WhatsApp integration is missing a session ID")
		return
	}

	// Check if the number is on WhatsApp first
	exists, err := h.openwa.CheckNumberExists(sessionID, req.Phone)
	if err != nil {
		h.logger.Error("Failed to check if number exists on WhatsApp", "error", err)
		utils.RespondInternalError(c, "Failed to verify number on WhatsApp: "+err.Error())
		return
	}
	if !exists {
		utils.RespondValidationError(c, "The phone number is not registered on WhatsApp. Please check the number and try again.")
		return
	}

	chatID := cleanPhone(req.Phone) + "@s.whatsapp.net"
	h.logger.Info("Phone ping", "phone", req.Phone, "chatID", chatID, "session", sessionID)

	err = h.openwa.SendTextMessage(sessionID, chatID, "Hello! This is a test message from NOANT AI. Your WhatsApp is connected!")
	if err != nil {
		h.logger.Error("Phone ping failed", "error", err)
		utils.RespondInternalError(c, "Failed to send test message: "+err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Test message sent to " + req.Phone,
	})
}

// CheckNumber checks if a phone number is registered on WhatsApp
func (h *OpenWAHandler) CheckNumber(c *gin.Context) {
	var req struct {
		Phone string `json:"phone" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondValidationError(c, "Phone number is required")
		return
	}
	utils.SanitizeStruct(&req)

	// Find active session
	userID, ok := c.Get("userID")
	if !ok {
		utils.RespondUnauthorized(c, "Unauthorized")
		return
	}

	integration, err := h.chat.GetWhatsAppIntegration(c.Request.Context(), userID.(string))
	if err != nil {
		h.logger.Error("Failed to load WhatsApp integration", "error", err)
		utils.RespondInternalError(c, "Failed to load WhatsApp integration")
		return
	}
	if integration == nil {
		utils.RespondInternalError(c, "No active WhatsApp integration. Connect first.")
		return
	}

	sessionID, _ := integration.Config["session_id"].(string)
	if sessionID == "" {
		utils.RespondInternalError(c, "WhatsApp integration is missing a session ID")
		return
	}

	exists, err := h.openwa.CheckNumberExists(sessionID, req.Phone)
	if err != nil {
		h.logger.Error("Failed to check WhatsApp number existence", "error", err)
		utils.RespondInternalError(c, "Failed to verify number on WhatsApp: "+err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"phone":  req.Phone,
		"exists": exists,
	})
}

// ========== SIMPLIFIED WHATSAPP CHANNEL ENDPOINTS ==========

// ConnectWhatsApp creates an OpenWA session and returns QR code
func (h *OpenWAHandler) ConnectWhatsApp(c *gin.Context) {
	var req struct {
		Phone string `json:"phone" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondValidationError(c, "Phone number is required")
		return
	}
	utils.SanitizeStruct(&req)

	sessionName := "noant-" + cleanPhone(req.Phone)
	h.logger.Info("Connecting WhatsApp", "phone", req.Phone, "session", sessionName)

	// Step 1: Check if OpenWA is reachable
	if err := h.openwa.Ping(); err != nil {
		h.logger.Error("OpenWA not reachable", "error", err)
		utils.RespondInternalError(c, "OpenWA server is not running. Start Docker container: docker compose up -d")
		return
	}

	userID, ok := c.Get("userID")
	if !ok {
		utils.RespondUnauthorized(c, "Unauthorized")
		return
	}

	// Step 2: Clean up only this user's previous WhatsApp session
	if existing, err := h.chat.GetWhatsAppIntegration(c.Request.Context(), userID.(string)); err == nil && existing != nil {
		if oldSessionID, _ := existing.Config["session_id"].(string); oldSessionID != "" {
			h.logger.Info("Logging out and deleting previous session for user", "sessionID", oldSessionID)
			_ = h.openwa.LogoutSession(oldSessionID)
			_ = h.openwa.DeleteSession(oldSessionID)
		}
	}

	// Step 3: Create fresh session with retry
	var sessionID string
	var createErr error
	for i := 0; i < 3; i++ {
		sessionID, createErr = h.openwa.CreateSession(sessionName)
		if createErr == nil {
			break
		}
		h.logger.Warn("Session create attempt failed", "attempt", i+1, "error", createErr)
		time.Sleep(2 * time.Second)
	}
	if createErr != nil {
		h.logger.Error("Failed to create session after 3 attempts", "error", createErr)
		utils.RespondInternalError(c, "Failed to create OpenWA session after 3 attempts")
		return
	}
	h.logger.Info("Session created", "sessionID", sessionID)

	// Step 4: Store integration initially as "connecting" until scanned
	h.chat.StoreWhatsAppIntegrationWithStatus(c.Request.Context(), userID.(string), sessionID, req.Phone, "connecting")

	// Step 5: Start session and configure webhook asynchronously in background
	go func(sid string, uid string, phone string) {
		defer func() {
			if r := recover(); r != nil {
				h.logger.Error("Panic in WhatsApp session setup goroutine", "sessionID", sid, "error", r)
			}
		}()
		h.logger.Info("Background: Starting session and configuring webhook", "sessionID", sid)

		// Start session with retry
		var startErr error
		for i := 0; i < 3; i++ {
			startErr = h.openwa.StartSession(sid)
			if startErr == nil {
				break
			}
			h.logger.Warn("Background: Session start attempt failed", "attempt", i+1, "error", startErr, "sessionID", sid)
			time.Sleep(2 * time.Second)
		}
		if startErr != nil {
			h.logger.Error("Background: Start session failed", "error", startErr, "sessionID", sid)
			return
		}
		h.logger.Info("Background: Session started successfully", "sessionID", sid)

		// Configure webhook with multiple fallback URLs
		webhookURLs := []string{
			"http://host.docker.internal:8080/api/v1/openwa/webhook",
			"http://172.19.0.1:8080/api/v1/openwa/webhook",
			"http://localhost:8080/api/v1/openwa/webhook",
		}
		var webhookOK bool
		for _, webhookURL := range webhookURLs {
			h.logger.Info("Background: Configuring webhook", "url", webhookURL, "sessionID", sid)
			if err := h.openwa.ConfigureWebhook(sid, webhookURL, h.cfg.OpenWAWebhookSecret); err != nil {
				h.logger.Warn("Background: Webhook URL failed", "url", webhookURL, "error", err, "sessionID", sid)
			} else {
				h.logger.Info("Background: Webhook configured successfully", "url", webhookURL, "sessionID", sid)
				webhookOK = true
				break
			}
		}
		if !webhookOK {
			h.logger.Error("Background: All webhook URLs failed for session", "sessionID", sid)
		}

		// Register session with health monitor
		if mgr := h.openwa.GetSessionManager(); mgr != nil {
			mgr.RegisterSession(sid, uid)
		}
	}(sessionID, userID.(string), req.Phone)

	c.JSON(http.StatusOK, gin.H{
		"session_id": sessionID,
		"qr_code":    "",
		"phone":      req.Phone,
		"status":     "initializing",
	})
}

// GetWhatsAppStatus returns the status of a WhatsApp session
func (h *OpenWAHandler) GetWhatsAppStatus(c *gin.Context) {
	sessionID := c.Param("sessionId")
	if sessionID == "" {
		utils.RespondValidationError(c, "Session ID is required")
		return
	}

	// ?force=true allows the frontend "Done" button to force-confirm connection
	// when the phone shows it's logged in but the status API still shows qr_ready
	forceConnect := c.Query("force") == "true"
	poll := c.Query("poll") != "false"

	status, err := h.openwa.GetSessionStatusByID(sessionID)
	if err != nil {
		h.logger.Error("Failed to get WhatsApp status", "error", err)
		utils.RespondInternalError(c, err.Error())
		return
	}

	h.logger.Info("WhatsApp status check", "sessionID", sessionID, "status", status, "force", forceConnect)

	// Also try to get QR code if status is not connected
	var qrCode string
	isConnected := status == "connected"
	if !isConnected {
		qr, _ := h.openwa.GetQRCode(sessionID)
		qrCode = qr
	}

	// Long poll: if not connected, not expired, not failed, not disconnected, not force-connecting, and poll is true, wait up to 25 seconds for updates
	if !isConnected && status != "expired" && status != "failed" && status != "disconnected" && !forceConnect && poll {
		h.logger.Info("Long-polling WhatsApp status starting", "sessionID", sessionID, "initialStatus", status)

		timeout := time.After(25 * time.Second)
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		initialStatus := status
		initialQR := qrCode

	LongPollLoop:
		for {
			select {
			case <-c.Request.Context().Done():
				h.logger.Info("Long-polling WhatsApp status canceled by client connection close", "sessionID", sessionID)
				return
			case <-timeout:
				h.logger.Info("Long-polling WhatsApp status timed out (no change)", "sessionID", sessionID)
				break LongPollLoop
			case <-ticker.C:
				currentStatus, err := h.openwa.GetSessionStatusByID(sessionID)
				if err != nil {
					h.logger.Warn("Failed to get WhatsApp status during long-poll", "error", err)
					continue
				}

				var currentQR string
				if currentStatus != "connected" {
					currentQR, _ = h.openwa.GetQRCode(sessionID)
				}

				if currentStatus != initialStatus || currentQR != initialQR {
					h.logger.Info("WhatsApp status/QR changed during long-poll", "sessionID", sessionID, "oldStatus", initialStatus, "newStatus", currentStatus, "qrChanged", currentQR != initialQR)
					status = currentStatus
					qrCode = currentQR
					isConnected = (currentStatus == "connected")
					break LongPollLoop
				}
			}
		}
	}

	// Force connect: user's phone shows logged in — trust the user and mark as connected
	if forceConnect && !isConnected {
		h.logger.Warn("Force-connect bypass used — session may not actually be connected", "sessionID", sessionID)
		isConnected = true
		status = "connected"
		qrCode = ""

		// Background verification: check actual status after a delay
		go func() {
			time.Sleep(10 * time.Second)
			actualStatus, err := h.openwa.GetSessionStatusByID(sessionID)
			if err != nil || actualStatus != "connected" {
				h.logger.Error("Force-connect verification failed — session is NOT actually connected",
					"sessionID", sessionID, "actualStatus", actualStatus, "error", err)
			}
		}()
	}

	// Update integration status if connected, and notify dashboard via WebSocket
	if isConnected {
		userID, ok := c.Get("userID")
		if ok {
			h.chat.StoreWhatsAppIntegrationWithStatus(c.Request.Context(), userID.(string), sessionID, "", "connected")

			// Register session with health monitor
			if mgr := h.openwa.GetSessionManager(); mgr != nil {
				mgr.RegisterSession(sessionID, userID.(string))
			}

			// Broadcast real-time status update to frontend
			if h.wsHub != nil {
				h.wsHub.BroadcastMessage(WebSocketMessage{
					ConversationID: "",
					Type:           "integration_update",
					Data: map[string]interface{}{
						"channel": "whatsapp",
						"status":  "connected",
					},
				})
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    status,
		"qr_code":   qrCode,
		"session":   sessionID,
		"connected": isConnected,
	})
}

// RefreshWhatsAppQR refreshes the QR code for a session
func (h *OpenWAHandler) RefreshWhatsAppQR(c *gin.Context) {
	sessionID := c.Param("sessionId")
	if sessionID == "" {
		utils.RespondValidationError(c, "Session ID is required")
		return
	}

	h.logger.Info("Refreshing QR", "sessionID", sessionID)

	// Look up the integration to get the phone number (so we reuse the correct session name)
	userID, ok := c.Get("userID")
	if !ok {
		utils.RespondUnauthorized(c, "Unauthorized")
		return
	}

	integration, err := h.chat.GetWhatsAppIntegration(c.Request.Context(), userID.(string))
	if err != nil || integration == nil {
		utils.RespondInternalError(c, "No WhatsApp integration found")
		return
	}

	phone, _ := integration.Config["phone"].(string)
	if phone == "" {
		utils.RespondInternalError(c, "Integration has no phone number — please reconnect")
		return
	}

	// Reuse the same session name pattern as ConnectWhatsApp
	sessionName := "noant-" + cleanPhone(phone)
	h.logger.Info("QR refresh: recreating session", "name", sessionName, "oldID", sessionID)

	// Delete stale session (best-effort — may already be gone)
	_ = h.openwa.DeleteSession(sessionID)
	time.Sleep(2 * time.Second)

	// Recreate with the same name
	newID, err := h.openwa.CreateSession(sessionName)
	if err != nil {
		h.logger.Error("Failed to recreate session for QR refresh", "error", err)
		utils.RespondInternalError(c, "Failed to refresh QR")
		return
	}
	h.logger.Info("QR refresh: new session", "id", newID)

	// Update DB integration to use the new session ID
	h.chat.StoreWhatsAppIntegrationWithStatus(c.Request.Context(), userID.(string), newID, phone, "connecting")

	// Start the new session and configure webhook asynchronously in background
	go func(nid string, uid string, phoneNum string) {
		defer func() {
			if r := recover(); r != nil {
				h.logger.Error("Panic in WhatsApp QR refresh goroutine", "sessionID", nid, "error", r)
			}
		}()
		h.logger.Info("Background QR Refresh: Starting session and configuring webhook", "sessionID", nid)

		// Start the new session with retry
		var startErr error
		for i := 0; i < 3; i++ {
			startErr = h.openwa.StartSession(nid)
			if startErr == nil {
				break
			}
			h.logger.Warn("Background QR Refresh: Session start attempt failed", "attempt", i+1, "error", startErr, "sessionID", nid)
			time.Sleep(2 * time.Second)
		}

		// Configure webhook on new session
		webhookURL := "http://host.docker.internal:8080/api/v1/openwa/webhook"
		if wErr := h.openwa.ConfigureWebhook(nid, webhookURL, h.cfg.OpenWAWebhookSecret); wErr != nil {
			h.logger.Warn("Background QR Refresh: Webhook config failed, trying localhost fallback", "error", wErr, "sessionID", nid)
			_ = h.openwa.ConfigureWebhook(nid, "http://localhost:8080/api/v1/openwa/webhook", h.cfg.OpenWAWebhookSecret)
		}

		// Register session with health monitor
		if mgr := h.openwa.GetSessionManager(); mgr != nil {
			mgr.RegisterSession(nid, uid)
		}
	}(newID, userID.(string), phone)

	c.JSON(http.StatusOK, gin.H{
		"qr_code":    "",
		"session_id": newID,
	})
}

// DisconnectWhatsApp disconnects a WhatsApp session
func (h *OpenWAHandler) DisconnectWhatsApp(c *gin.Context) {
	var req struct {
		SessionID string `json:"session_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondValidationError(c, "Session ID is required")
		return
	}
	utils.SanitizeStruct(&req)

	// Remove integration and disconnect session completely (logging out credentials)
	userID := getUserID(c)
	if userID != "" {
		h.chat.DisconnectWhatsAppSession(c.Request.Context(), userID)
		h.chat.RemoveWhatsAppIntegration(c.Request.Context(), userID)
	} else {
		// Fallback: delete session only if userID is somehow missing
		_ = h.openwa.LogoutSession(req.SessionID)
		_ = h.openwa.DeleteSession(req.SessionID)
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// cleanPhone removes all non-digit characters
func cleanPhone(phone string) string {
	result := make([]byte, 0, len(phone))
	for _, c := range phone {
		if c >= '0' && c <= '9' {
			result = append(result, byte(c))
		}
	}
	return string(result)
}
