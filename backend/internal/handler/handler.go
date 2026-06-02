package handler

import (
	"context"
	"io"
	"net/http"
	"strconv"
	"time"

	"noant/config"
	"noant/internal/domain"
	"noant/internal/infrastructure"
	"noant/internal/middleware"
	"noant/internal/service"
	"noant/internal/utils"

	"github.com/gin-gonic/gin"
)

type Handlers struct {
	Auth         *AuthHandler
	Chat         *ChatHandler
	Training     *TrainingHandler
	Analytics    *AnalyticsHandler
	Integration  *IntegrationHandler
	Settings     *SettingsHandler
	Archive      *ArchiveHandler
	Payment      *PaymentHandler
	Audit        *AuditHandler
	Notification *NotificationHandler
	Widget       *WidgetHandler
	Inventory    *InventoryHandler
	Handoff      *HandoffHandler
	OpenWA       *OpenWAHandler
	Telegram     *TelegramHandler
}

func NewHandlers(cfg *config.Config, services *service.Services, logger *infrastructure.Logger, wsHub *WebSocketHub) *Handlers {
	return &Handlers{
		Auth:         NewAuthHandler(services.Auth, logger),
		Chat:         NewChatHandler(services.Chat, logger, wsHub),
		Training:     NewTrainingHandler(services.Training, logger),
		Analytics:    NewAnalyticsHandler(services.Analytics, logger),
		Integration:  NewIntegrationHandler(services.Integration, logger),
		Settings:     NewSettingsHandler(services.Settings, logger),
		Archive:      NewArchiveHandler(services.Archive, logger),
		Payment:      NewPaymentHandler(services.Payment, logger),
		Audit:        NewAuditHandler(services.Audit, logger),
		Notification: NewNotificationHandler(services.Notification, logger),
		Widget:       NewWidgetHandler(services.Widget, logger),
		Inventory:    NewInventoryHandler(services.Inventory, logger),
		Handoff:      NewHandoffHandler(services.Handoff, logger),
		OpenWA:       NewOpenWAHandler(cfg, services.OpenWA, services.Chat, logger, wsHub),
		Telegram:     NewTelegramHandler(services.Integration, logger),
	}
}

func (h *Handlers) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"version":   "2.0.0",
	})
}

// ========== AUTH HANDLER ==========

type AuthHandler struct {
	service *service.AuthService
	logger  *infrastructure.Logger
}

func NewAuthHandler(service *service.AuthService, logger *infrastructure.Logger) *AuthHandler {
	return &AuthHandler{service: service, logger: logger}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req struct {
		Email       string `json:"email" binding:"required,email"`
		Password    string `json:"password" binding:"required,min=8"`
		FirstName   string `json:"first_name" binding:"required"`
		LastName    string `json:"last_name" binding:"required"`
		CompanyName string `json:"company_name"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondValidationError(c, err.Error())
		return
	}

	user, err := h.service.Register(c.Request.Context(), req.Email, req.Password, req.FirstName, req.LastName, req.CompanyName)
	if err != nil {
		h.logger.Error("Registration failed", "error", err)
		utils.RespondConflict(c, err.Error())
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "User registered successfully",
		"user":    user,
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondValidationError(c, err.Error())
		return
	}

	user, token, refreshToken, err := h.service.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		h.logger.Error("Login failed", "error", err)
		utils.RespondUnauthorized(c, "Invalid email or password")
		return
	}

	middleware.SetAuthCookies(c, token, refreshToken, 24*time.Hour, 7*24*time.Hour)
	c.Header("Cache-Control", "no-store")

	// Compute trial info for response
	var trialInfo map[string]interface{}
	if user.TrialExpiresAt != nil {
		trialInfo = map[string]interface{}{
			"trial_expires_at": user.TrialExpiresAt.Format(time.RFC3339),
			"trial_ended":      time.Now().After(*user.TrialExpiresAt),
			"trial_days_left":  int(time.Until(*user.TrialExpiresAt).Hours() / 24),
		}
		if trialInfo["trial_days_left"].(int) < 0 {
			trialInfo["trial_days_left"] = 0
		}
	} else {
		trialInfo = map[string]interface{}{
			"trial_ended": false,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"user":       user,
		"trial_info": trialInfo,
	})
}

func (h *AuthHandler) RefreshToken(c *gin.Context) {
	refreshToken := middleware.GetRefreshTokenFromRequest(c)
	if refreshToken == "" {
		utils.RespondUnauthorized(c, "refresh token required")
		return
	}

	token, newRefreshToken, err := h.service.RefreshToken(c.Request.Context(), refreshToken)
	if err != nil {
		utils.RespondUnauthorized(c, err.Error())
		return
	}

	middleware.SetAuthCookies(c, token, newRefreshToken, 24*time.Hour, 7*24*time.Hour)
	c.Header("Cache-Control", "no-store")
	c.JSON(http.StatusOK, gin.H{"message": "Session refreshed"})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	token := middleware.GetAccessTokenFromRequest(c)
	refreshToken := middleware.GetRefreshTokenFromRequest(c)
	if err := h.service.Logout(c.Request.Context(), token, refreshToken); err != nil {
		utils.RespondInternalError(c, "Failed to log out")
		return
	}
	middleware.ClearAuthCookies(c)
	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

func (h *AuthHandler) ChangePassword(c *gin.Context) {
	var req struct {
		CurrentPassword string `json:"current_password" binding:"required"`
		NewPassword     string `json:"new_password" binding:"required,min=8"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondValidationError(c, err.Error())
		return
	}

	userID, _ := c.Get("userID")
	if err := h.service.ChangePassword(c.Request.Context(), userID.(string), req.CurrentPassword, req.NewPassword); err != nil {
		utils.RespondValidationError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password changed successfully"})
}

func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondValidationError(c, err.Error())
		return
	}

	if err := h.service.ForgotPassword(c.Request.Context(), req.Email); err != nil {
		h.logger.Error("Forgot password failed", "error", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "If the email exists, a reset link has been sent"})
}

func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req struct {
		Token       string `json:"token" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=8"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondValidationError(c, err.Error())
		return
	}

	if err := h.service.ResetPassword(c.Request.Context(), req.Token, req.NewPassword); err != nil {
		h.logger.Error("Reset password failed", "error", err)
		utils.RespondInternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password reset successfully"})
}

func (h *AuthHandler) Me(c *gin.Context) {
	userID, _ := c.Get("userID")
	id, _ := userID.(string)
	user, err := h.service.GetUser(c.Request.Context(), id)
	if err != nil {
		utils.RespondInternalError(c, "Failed to retrieve user")
		return
	}
	if user == nil {
		utils.RespondUnauthorized(c, "User not found")
		return
	}

	c.Header("Cache-Control", "no-store")

	// Compute trial info for response
	var trialInfo map[string]interface{}
	if user.TrialExpiresAt != nil {
		trialInfo = map[string]interface{}{
			"trial_expires_at": user.TrialExpiresAt.Format(time.RFC3339),
			"trial_ended":      time.Now().After(*user.TrialExpiresAt),
			"trial_days_left":  int(time.Until(*user.TrialExpiresAt).Hours() / 24),
		}
		if trialInfo["trial_days_left"].(int) < 0 {
			trialInfo["trial_days_left"] = 0
		}
	} else {
		trialInfo = map[string]interface{}{
			"trial_ended": false,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"user":       user,
		"trial_info": trialInfo,
	})
}

// ========== CHAT HANDLER ==========

type ChatHandler struct {
	service *service.ChatService
	logger  *infrastructure.Logger
	wsHub   *WebSocketHub
}

func NewChatHandler(service *service.ChatService, logger *infrastructure.Logger, wsHub *WebSocketHub) *ChatHandler {
	return &ChatHandler{service: service, logger: logger, wsHub: wsHub}
}

func (h *ChatHandler) DirectChat(c *gin.Context) {
	var req struct {
		CustomerName string `json:"customer_name"`
		Message      string `json:"message" binding:"required"`
		Channel      string `json:"channel" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondValidationError(c, err.Error())
		return
	}

	userID, _ := c.Get("userID")
	conv, msg, err := h.service.DirectChat(c.Request.Context(), userID.(string), req.CustomerName, req.CustomerName, req.Message, req.Channel, "")
	if err != nil {
		h.logger.Error("Direct chat failed", "error", err)
		utils.RespondInternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"conversation": conv,
		"message":      msg,
	})
}

func (h *ChatHandler) ClearChats(c *gin.Context) {
	userID, _ := c.Get("userID")

	if err := h.service.ClearChats(c.Request.Context(), userID.(string)); err != nil {
		h.logger.Error("Clear chats failed", "error", err)
		utils.RespondInternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Chats cleared successfully"})
}

func (h *ChatHandler) ListConversations(c *gin.Context) {
	userID, _ := c.Get("userID")
	status := c.Query("status")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	conversations, total, err := h.service.ListConversations(c.Request.Context(), userID.(string), status, page, limit)
	if err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}

	hasMore := page*limit < total

	c.JSON(http.StatusOK, gin.H{
		"conversations": conversations,
		"total":         total,
		"page":          page,
		"limit":         limit,
		"has_more":      hasMore,
	})
}

func (h *ChatHandler) GetConversation(c *gin.Context) {
	id := c.Param("id")
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "30"))
	if limit < 1 || limit > 100 {
		limit = 30
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * limit

	conv, messages, total, err := h.service.GetConversationPaginated(c.Request.Context(), userID.(string), id, limit, offset)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	hasMore := page*limit < total

	c.JSON(http.StatusOK, gin.H{
		"conversation": conv,
		"messages":     messages,
		"total":        total,
		"has_more":     hasMore,
		"page":         page,
	})
}

func (h *ChatHandler) SendMessage(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Content string `json:"content" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondValidationError(c, err.Error())
		return
	}

	// Store customer message
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	_, err := h.service.SendMessage(c.Request.Context(), userID.(string), id, "customer", req.Content)
	if err != nil {
		h.logger.Error("Failed to store message", "error", err)
		utils.RespondInternalError(c, err.Error())
		return
	}

	// Broadcast typing indicator - AI is thinking
	if h.wsHub != nil {
		h.wsHub.BroadcastMessage(WebSocketMessage{
			ConversationID: id,
			Type:           "typing_indicator",
			Data: map[string]interface{}{
				"conversation_id": id,
				"is_typing":       true,
			},
		})
	}

	// Generate AI response asynchronously
	go func() {
		aiMsg, err := h.service.GenerateAIResponse(context.Background(), id, req.Content)
		if err != nil {
			h.logger.Error("AI generation failed in goroutine", "error", err)
			// Remove typing indicator on error
			if h.wsHub != nil {
				h.wsHub.BroadcastMessage(WebSocketMessage{
					ConversationID: id,
					Type:           "typing_indicator",
					Data: map[string]interface{}{
						"conversation_id": id,
						"is_typing":       false,
					},
				})
			}
			return
		}
		if h.wsHub != nil {
			h.wsHub.BroadcastMessage(WebSocketMessage{
				ConversationID: id,
				Type:           "new_message",
				Data: map[string]interface{}{
					"id":              aiMsg.ID,
					"conversation_id": aiMsg.ConversationID,
					"content":         aiMsg.Content,
					"role":            aiMsg.Role,
					"created_at":      aiMsg.CreatedAt,
					"metadata":        aiMsg.Metadata,
					"confidence":      aiMsg.Confidence,
					"source":          aiMsg.Source,
				},
			})
			// Stop typing indicator when response arrives
			h.wsHub.BroadcastMessage(WebSocketMessage{
				ConversationID: id,
				Type:           "typing_indicator",
				Data: map[string]interface{}{
					"conversation_id": id,
					"is_typing":       false,
				},
			})
		}
	}()

	c.JSON(http.StatusOK, gin.H{"message": "Message sent"})
}

func (h *ChatHandler) HumanTakeover(c *gin.Context) {
	id := c.Param("id")
	agentID, _ := c.Get("userID")

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	if err := h.service.HumanTakeover(c.Request.Context(), userID.(string), id, agentID.(string)); err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Conversation taken over by human agent"})
}

func (h *ChatHandler) Escalate(c *gin.Context) {
	var req struct {
		Reason string `json:"reason"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondValidationError(c, err.Error())
		return
	}

	id := c.Param("id")
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	if err := h.service.Escalate(c.Request.Context(), userID.(string), id, req.Reason); err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Conversation escalated"})
}

// ========== TRAINING HANDLER ==========

type TrainingHandler struct {
	service *service.TrainingService
	logger  *infrastructure.Logger
}

func NewTrainingHandler(service *service.TrainingService, logger *infrastructure.Logger) *TrainingHandler {
	return &TrainingHandler{service: service, logger: logger}
}

func (h *TrainingHandler) ListCategories(c *gin.Context) {
	userID, _ := c.Get("userID")
	categories, err := h.service.ListCategories(c.Request.Context(), userID.(string))
	if err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"categories": categories})
}

func (h *TrainingHandler) CreateCategory(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
		Color       string `json:"color"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondValidationError(c, err.Error())
		return
	}

	userID, _ := c.Get("userID")
	category, err := h.service.CreateCategory(c.Request.Context(), userID.(string), req.Name, req.Description, req.Color)
	if err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Category created", "id": category.ID})
}

func (h *TrainingHandler) BulkImport(c *gin.Context) {
	var req struct {
		CategoryID string `json:"category_id" binding:"required"`
		QAPairs    []struct {
			Question string `json:"question"`
			Answer   string `json:"answer"`
		} `json:"qa_pairs" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondValidationError(c, err.Error())
		return
	}

	var pairs []domain.QAPair
	for _, p := range req.QAPairs {
		pairs = append(pairs, domain.QAPair{
			Question: p.Question,
			Answer:   p.Answer,
		})
	}

	userID, _ := c.Get("userID")
	if err := h.service.BulkImport(c.Request.Context(), userID.(string), req.CategoryID, pairs); err != nil {
		h.logger.Error("Bulk import failed", "error", err)
		utils.RespondInternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Bulk import successful", "count": len(pairs)})
}

func (h *TrainingHandler) UploadCSV(c *gin.Context) {
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}
	defer file.Close()

	categoryID := c.PostForm("category_id")
	if categoryID == "" {
		categoryID = "default"
	}

	data, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file"})
		return
	}

	userID, _ := c.Get("userID")
	count, err := h.service.UploadCSV(c.Request.Context(), userID.(string), categoryID, data)
	if err != nil {
		h.logger.Error("CSV upload failed", "error", err)
		utils.RespondInternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "CSV uploaded successfully", "count": count})
}

func (h *TrainingHandler) ListUnknownQuestions(c *gin.Context) {
	status := c.Query("status")
	limit := 50

	userID, _ := c.Get("userID")
	questions, err := h.service.ListUnknownQuestions(c.Request.Context(), userID.(string), status, limit)
	if err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"questions": questions})
}

func (h *TrainingHandler) TrainUnknown(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Answer     string `json:"answer" binding:"required"`
		CategoryID string `json:"category_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondValidationError(c, err.Error())
		return
	}

	userID, _ := c.Get("userID")
	if err := h.service.TrainUnknown(c.Request.Context(), userID.(string), id, req.Answer, req.CategoryID); err != nil {
		if err.Error() == "unknown question not found" || err.Error() == "not found" {
			utils.RespondNotFound(c, err.Error())
			return
		}
		utils.RespondInternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Question trained successfully"})
}

func (h *TrainingHandler) IgnoreUnknown(c *gin.Context) {
	id := c.Param("id")
	userID, _ := c.Get("userID")

	if err := h.service.IgnoreUnknown(c.Request.Context(), userID.(string), id); err != nil {
		if err.Error() == "unknown question not found" || err.Error() == "not found" {
			utils.RespondNotFound(c, err.Error())
			return
		}
		h.logger.Error("Ignore unknown question failed", "error", err)
		utils.RespondInternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Question ignored successfully"})
}

func (h *TrainingHandler) ClearUnknown(c *gin.Context) {
	userID, _ := c.Get("userID")

	if err := h.service.ClearUnknownQuestions(c.Request.Context(), userID.(string)); err != nil {
		h.logger.Error("Clear unknown questions failed", "error", err)
		utils.RespondInternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Unknown questions cleared successfully"})
}

func (h *TrainingHandler) ListQAPairs(c *gin.Context) {
	categoryID := c.Param("id")
	userID, _ := c.Get("userID")

	qaPairs, err := h.service.ListQAPairs(c.Request.Context(), userID.(string), categoryID)
	if err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"qa_pairs": qaPairs})
}

func (h *TrainingHandler) CreateQAPair(c *gin.Context) {
	var req struct {
		CategoryID string `json:"category_id" binding:"required"`
		Question   string `json:"question" binding:"required"`
		Answer     string `json:"answer" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondValidationError(c, err.Error())
		return
	}

	userID, _ := c.Get("userID")
	qa, err := h.service.CreateQAPair(c.Request.Context(), userID.(string), req.CategoryID, req.Question, req.Answer)
	if err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Q&A pair created successfully", "qa_pair": qa})
}

func (h *TrainingHandler) UpdateQAPair(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		CategoryID string `json:"category_id" binding:"required"`
		Question   string `json:"question" binding:"required"`
		Answer     string `json:"answer" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondValidationError(c, err.Error())
		return
	}

	userID, _ := c.Get("userID")
	err := h.service.UpdateQAPair(c.Request.Context(), userID.(string), id, req.CategoryID, req.Question, req.Answer)
	if err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Q&A pair updated successfully"})
}

func (h *TrainingHandler) DeleteQAPair(c *gin.Context) {
	id := c.Param("id")
	userID, _ := c.Get("userID")

	err := h.service.DeleteQAPair(c.Request.Context(), userID.(string), id)
	if err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Q&A pair deleted successfully"})
}

func (h *TrainingHandler) DeleteCategory(c *gin.Context) {
	id := c.Param("id")
	userID, _ := c.Get("userID")

	err := h.service.DeleteCategory(c.Request.Context(), userID.(string), id)
	if err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Category and all associated Q&A pairs deleted successfully"})
}

func (h *TrainingHandler) SearchQAPairs(c *gin.Context) {
	query := c.Query("q")
	userID, _ := c.Get("userID")

	qaPairs, err := h.service.SearchQAPairs(c.Request.Context(), userID.(string), query)
	if err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"qa_pairs": qaPairs})
}

// ========== ANALYTICS HANDLER ==========

type AnalyticsHandler struct {
	service *service.AnalyticsService
	logger  *infrastructure.Logger
}

func NewAnalyticsHandler(service *service.AnalyticsService, logger *infrastructure.Logger) *AnalyticsHandler {
	return &AnalyticsHandler{service: service, logger: logger}
}

func (h *AnalyticsHandler) Overview(c *gin.Context) {
	userID, _ := c.Get("userID")
	overview, err := h.service.Overview(c.Request.Context(), userID.(string))
	if err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, overview)
}

func (h *AnalyticsHandler) ChannelDistribution(c *gin.Context) {
	userID, _ := c.Get("userID")
	distribution, err := h.service.ChannelDistribution(c.Request.Context(), userID.(string))
	if err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"distribution": distribution})
}

func (h *AnalyticsHandler) Insights(c *gin.Context) {
	userID, _ := c.Get("userID")
	insights, err := h.service.Insights(c.Request.Context(), userID.(string))
	if err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, insights)
}

func (h *AnalyticsHandler) Trends(c *gin.Context) {
	userID, _ := c.Get("userID")
	days := 7

	trends, err := h.service.Trends(c.Request.Context(), userID.(string), days)
	if err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"trends": trends})
}

// ========== INTEGRATION HANDLER ==========

type IntegrationHandler struct {
	service *service.IntegrationService
	logger  *infrastructure.Logger
}

func NewIntegrationHandler(service *service.IntegrationService, logger *infrastructure.Logger) *IntegrationHandler {
	return &IntegrationHandler{service: service, logger: logger}
}

func (h *IntegrationHandler) List(c *gin.Context) {
	userID, _ := c.Get("userID")
	integrations, err := h.service.List(c.Request.Context(), userID.(string))
	if err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"integrations": integrations})
}

func (h *IntegrationHandler) Connect(c *gin.Context) {
	var req struct {
		Channel string                 `json:"channel" binding:"required"`
		Config  map[string]interface{} `json:"config"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondValidationError(c, err.Error())
		return
	}

	userID, _ := c.Get("userID")
	integration, err := h.service.Connect(c.Request.Context(), userID.(string), req.Channel, req.Config)
	if err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusCreated, gin.H{"integration": integration})
}

func (h *IntegrationHandler) Disconnect(c *gin.Context) {
	channel := c.Param("channel")
	userID, _ := c.Get("userID")
	if err := h.service.Disconnect(c.Request.Context(), userID.(string), channel); err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Integration disconnected"})
}

func (h *IntegrationHandler) Test(c *gin.Context) {
	channel := c.Param("channel")

	// Optionally parse config credentials from the request body (for pre-connect testing)
	var req struct {
		Config map[string]interface{} `json:"config"`
	}
	// Ignore bind errors – the body is optional
	c.ShouldBindJSON(&req)

	success, message := h.service.Test(c.Request.Context(), channel, req.Config)

	if success {
		c.JSON(http.StatusOK, gin.H{"status": "success", "message": message})
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": message})
	}
}

// ========== SETTINGS HANDLER ==========

type SettingsHandler struct {
	service *service.SettingsService
	logger  *infrastructure.Logger
}

func NewSettingsHandler(service *service.SettingsService, logger *infrastructure.Logger) *SettingsHandler {
	return &SettingsHandler{service: service, logger: logger}
}

func (h *SettingsHandler) GetProfile(c *gin.Context) {
	userID, _ := c.Get("userID")
	profile, err := h.service.GetProfile(c.Request.Context(), userID.(string))
	if err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, profile)
}

func (h *SettingsHandler) UpdateProfile(c *gin.Context) {
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondValidationError(c, err.Error())
		return
	}

	userID, _ := c.Get("userID")
	if err := h.service.UpdateProfile(c.Request.Context(), userID.(string), req); err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Profile updated"})
}

func (h *SettingsHandler) ListAPIKeys(c *gin.Context) {
	userID, _ := c.Get("userID")
	keys, err := h.service.ListAPIKeys(c.Request.Context(), userID.(string))
	if err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"api_keys": keys})
}

func (h *SettingsHandler) CreateAPIKey(c *gin.Context) {
	var req struct {
		Name string `json:"name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondValidationError(c, err.Error())
		return
	}

	userID, _ := c.Get("userID")
	key, err := h.service.CreateAPIKey(c.Request.Context(), userID.(string), req.Name)
	if err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusCreated, gin.H{"api_key": key, "id": key.ID})
}

func (h *SettingsHandler) RevokeAPIKey(c *gin.Context) {
	id := c.Param("id")
	userID, _ := c.Get("userID")
	if err := h.service.RevokeAPIKey(c.Request.Context(), userID.(string), id); err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "API key revoked"})
}

func (h *SettingsHandler) ListTeam(c *gin.Context) {
	userID, _ := c.Get("userID")
	members, err := h.service.ListTeam(c.Request.Context(), userID.(string))
	if err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"team": members})
}

func (h *SettingsHandler) InviteTeamMember(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
		Role  string `json:"role" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondValidationError(c, err.Error())
		return
	}

	userID, _ := c.Get("userID")
	member, err := h.service.InviteTeamMember(c.Request.Context(), userID.(string), req.Email, req.Role)
	if err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Invitation sent", "id": member.ID})
}

func (h *SettingsHandler) RemoveTeamMember(c *gin.Context) {
	id := c.Param("id")
	if err := h.service.RemoveTeamMember(c.Request.Context(), id); err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Team member removed"})
}

// ========== ARCHIVE HANDLER ==========

type ArchiveHandler struct {
	service *service.ArchiveService
	logger  *infrastructure.Logger
}

func NewArchiveHandler(service *service.ArchiveService, logger *infrastructure.Logger) *ArchiveHandler {
	return &ArchiveHandler{service: service, logger: logger}
}

func (h *ArchiveHandler) ListFolders(c *gin.Context) {
	userID, _ := c.Get("userID")
	folderType := c.Query("type")

	folders, err := h.service.ListFolders(c.Request.Context(), userID.(string), folderType)
	if err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"folders": folders})
}

func (h *ArchiveHandler) CreateFolder(c *gin.Context) {
	var req struct {
		Name  string `json:"name" binding:"required"`
		Type  string `json:"type"`
		Color string `json:"color"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondValidationError(c, err.Error())
		return
	}

	if req.Type == "" {
		req.Type = "custom"
	}

	userID, _ := c.Get("userID")
	folder, err := h.service.CreateFolder(c.Request.Context(), userID.(string), req.Name, req.Type, req.Color)
	if err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Folder created", "id": folder.ID})
}

func (h *ArchiveHandler) DeleteFolder(c *gin.Context) {
	id := c.Param("id")
	if err := h.service.DeleteFolder(c.Request.Context(), id); err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Folder deleted"})
}

func (h *ArchiveHandler) MoveChat(c *gin.Context) {
	var req struct {
		ConversationID string `json:"conversation_id" binding:"required"`
		FolderID       string `json:"folder_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondValidationError(c, err.Error())
		return
	}

	userID, _ := c.Get("userID")
	if err := h.service.MoveChat(c.Request.Context(), userID.(string), req.ConversationID, req.FolderID); err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Chat moved"})
}

func (h *ArchiveHandler) RemoveFromArchive(c *gin.Context) {
	var req struct {
		ConversationID string `json:"conversation_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondValidationError(c, err.Error())
		return
	}

	userID, _ := c.Get("userID")
	if err := h.service.RemoveFromArchive(c.Request.Context(), userID.(string), req.ConversationID); err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Chat removed from archive"})
}

func (h *ArchiveHandler) GetStatus(c *gin.Context) {
	userID, _ := c.Get("userID")
	status, err := h.service.GetStatus(c.Request.Context(), userID.(string))
	if err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, status)
}

// ========== PAYMENT HANDLER ==========

type PaymentHandler struct {
	service *service.PaymentService
	logger  *infrastructure.Logger
}

func NewPaymentHandler(service *service.PaymentService, logger *infrastructure.Logger) *PaymentHandler {
	return &PaymentHandler{service: service, logger: logger}
}

func (h *PaymentHandler) ListPlans(c *gin.Context) {
	plans, err := h.service.ListPlans(c.Request.Context())
	if err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"plans": plans})
}

func (h *PaymentHandler) Subscribe(c *gin.Context) {
	// Accept both 'plan_id' and 'plan' keys for frontend compatibility
	var req struct {
		PlanID   string `json:"plan_id"`
		Plan     string `json:"plan"`
		Currency string `json:"currency"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondValidationError(c, err.Error())
		return
	}

	// Prefer plan_id, fall back to plan
	planID := req.PlanID
	if planID == "" {
		planID = req.Plan
	}
	if planID == "" {
		utils.RespondValidationError(c, "plan or plan_id is required")
		return
	}

	userID, _ := c.Get("userID")
	checkoutURL, err := h.service.Subscribe(c.Request.Context(), userID.(string), planID)
	if err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}

	resp := gin.H{"message": "Subscription initiated"}
	if checkoutURL != "" {
		resp["checkout_url"] = checkoutURL
	}
	c.JSON(http.StatusOK, resp)
}

func (h *PaymentHandler) Webhook(c *gin.Context) {
	payload, _ := c.GetRawData()
	if err := h.service.Webhook(c.Request.Context(), payload); err != nil {
		utils.RespondValidationError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *PaymentHandler) Status(c *gin.Context) {
	userID, _ := c.Get("userID")
	status, err := h.service.Status(c.Request.Context(), userID.(string))
	if err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, status)
}

// ========== AUDIT HANDLER ==========

type AuditHandler struct {
	service *service.AuditService
	logger  *infrastructure.Logger
}

func NewAuditHandler(service *service.AuditService, logger *infrastructure.Logger) *AuditHandler {
	return &AuditHandler{service: service, logger: logger}
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

// ========== INVENTORY HANDLER ==========

type InventoryHandler struct {
	service *service.InventoryService
	logger  *infrastructure.Logger
}

func NewInventoryHandler(service *service.InventoryService, logger *infrastructure.Logger) *InventoryHandler {
	return &InventoryHandler{service: service, logger: logger}
}

func (h *InventoryHandler) Create(c *gin.Context) {
	var req struct {
		Type          string   `json:"type" binding:"required"`
		Name          string   `json:"name" binding:"required"`
		Description   string   `json:"description"`
		Price         float64  `json:"price" binding:"required"`
		MinPrice      *float64 `json:"min_price"`
		StockQuantity *int     `json:"stock_quantity"`
		ImageURL      *string  `json:"image_url"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondValidationError(c, err.Error())
		return
	}

	userID, _ := c.Get("userID")
	item := &domain.InventoryItem{
		Type:          req.Type,
		Name:          req.Name,
		Description:   req.Description,
		Price:         req.Price,
		MinPrice:      req.MinPrice,
		StockQuantity: req.StockQuantity,
		ImageURL:      req.ImageURL,
	}

	if err := h.service.Create(c.Request.Context(), userID.(string), item); err != nil {
		h.logger.Error("Failed to create inventory item", "error", err)
		utils.RespondInternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusCreated, gin.H{"item": item})
}

func (h *InventoryHandler) List(c *gin.Context) {
	userID, _ := c.Get("userID")
	itemType := c.Query("type")

	items, err := h.service.List(c.Request.Context(), userID.(string), itemType)
	if err != nil {
		h.logger.Error("Failed to list inventory", "error", err)
		utils.RespondInternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"items": items, "count": len(items)})
}

func (h *InventoryHandler) GetByID(c *gin.Context) {
	userID, _ := c.Get("userID")
	id := c.Param("id")

	item, err := h.service.GetByID(c.Request.Context(), id, userID.(string))
	if err != nil {
		h.logger.Error("Failed to get inventory item", "error", err)
		utils.RespondInternalError(c, err.Error())
		return
	}
	if item == nil {
		utils.RespondNotFound(c, "Item not found")
		return
	}

	c.JSON(http.StatusOK, gin.H{"item": item})
}

func (h *InventoryHandler) Update(c *gin.Context) {
	var req struct {
		ID            string   `json:"id" binding:"required"`
		Type          string   `json:"type"`
		Name          string   `json:"name"`
		Description   string   `json:"description"`
		Price         float64  `json:"price"`
		MinPrice      *float64 `json:"min_price"`
		StockQuantity *int     `json:"stock_quantity"`
		ImageURL      *string  `json:"image_url"`
		IsActive      *bool    `json:"is_active"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondValidationError(c, err.Error())
		return
	}

	userID, _ := c.Get("userID")
	item, err := h.service.GetByID(c.Request.Context(), req.ID, userID.(string))
	if err != nil || item == nil {
		utils.RespondNotFound(c, "Item not found")
		return
	}

	if req.Type != "" {
		item.Type = req.Type
	}
	if req.Name != "" {
		item.Name = req.Name
	}
	if req.Description != "" {
		item.Description = req.Description
	}
	if req.Price > 0 {
		item.Price = req.Price
	}
	if req.MinPrice != nil {
		item.MinPrice = req.MinPrice
	}
	if req.StockQuantity != nil {
		item.StockQuantity = req.StockQuantity
	}
	if req.ImageURL != nil {
		item.ImageURL = req.ImageURL
	}
	if req.IsActive != nil {
		item.IsActive = *req.IsActive
	}

	if err := h.service.Update(c.Request.Context(), item); err != nil {
		h.logger.Error("Failed to update inventory item", "error", err)
		utils.RespondInternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"item": item})
}

func (h *InventoryHandler) Delete(c *gin.Context) {
	userID, _ := c.Get("userID")
	id := c.Param("id")

	if err := h.service.Delete(c.Request.Context(), id, userID.(string)); err != nil {
		h.logger.Error("Failed to delete inventory item", "error", err)
		utils.RespondInternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Item deleted"})
}

func (h *InventoryHandler) Search(c *gin.Context) {
	userID, _ := c.Get("userID")
	q := c.Query("q")

	items, err := h.service.Search(c.Request.Context(), userID.(string), q)
	if err != nil {
		h.logger.Error("Failed to search inventory", "error", err)
		utils.RespondInternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"items": items, "count": len(items)})
}

// ========== HANDOFF HANDLER ==========

type HandoffHandler struct {
	service *service.HandoffService
	logger  *infrastructure.Logger
}

func NewHandoffHandler(service *service.HandoffService, logger *infrastructure.Logger) *HandoffHandler {
	return &HandoffHandler{service: service, logger: logger}
}

func (h *HandoffHandler) List(c *gin.Context) {
	userID, _ := c.Get("userID")
	status := c.Query("status")

	handoffs, err := h.service.List(c.Request.Context(), userID.(string), status)
	if err != nil {
		h.logger.Error("Failed to list handoffs", "error", err)
		utils.RespondInternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"handoffs": handoffs, "count": len(handoffs)})
}

func (h *HandoffHandler) GetByID(c *gin.Context) {
	userID, _ := c.Get("userID")
	id := c.Param("id")

	handoff, err := h.service.GetByID(c.Request.Context(), id, userID.(string))
	if err != nil {
		h.logger.Error("Failed to get handoff", "error", err)
		utils.RespondInternalError(c, err.Error())
		return
	}
	if handoff == nil {
		utils.RespondNotFound(c, "Handoff not found")
		return
	}

	c.JSON(http.StatusOK, gin.H{"handoff": handoff})
}

func (h *HandoffHandler) UpdateStatus(c *gin.Context) {
	var req struct {
		ID         string   `json:"id" binding:"required"`
		Status     string   `json:"status" binding:"required"`
		Notes      string   `json:"notes"`
		FinalPrice *float64 `json:"final_price"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondValidationError(c, err.Error())
		return
	}

	userID, _ := c.Get("userID")
	if err := h.service.UpdateStatus(c.Request.Context(), req.ID, userID.(string), req.Status, req.Notes, req.FinalPrice); err != nil {
		h.logger.Error("Failed to update handoff status", "error", err)
		utils.RespondInternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Handoff updated"})
}

// ========== OPENWA HANDLER ==========

type OpenWAHandler struct {
	cfg    *config.Config
	openwa *service.OpenWAService
	chat   *service.ChatService
	logger *infrastructure.Logger
	wsHub  *WebSocketHub
}

func NewOpenWAHandler(cfg *config.Config, openwa *service.OpenWAService, chat *service.ChatService, logger *infrastructure.Logger, wsHub *WebSocketHub) *OpenWAHandler {
	return &OpenWAHandler{cfg: cfg, openwa: openwa, chat: chat, logger: logger, wsHub: wsHub}
}

// WhatsAppWebhook receives incoming messages from OpenWA
func (h *OpenWAHandler) WhatsAppWebhook(c *gin.Context) {
	// Read raw body for signature verification
	rawBody, err := io.ReadAll(c.Request.Body)
	if err != nil {
		h.logger.Error("Failed to read OpenWA webhook body", "error", err)
		utils.RespondValidationError(c, "Failed to read request body")
		return
	}

	// Verify HMAC signature if configured
	signature := c.GetHeader("X-Hub-Signature-256")
	if signature != "" && !h.openwa.VerifyWebhookSignature(rawBody, signature) {
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

	// Ignore non-text messages for now (OpenWA uses "chat" for standard text messages)
	if msg.Type != "text" && msg.Type != "chat" && msg.Type != "" {
		h.logger.Info("Ignoring non-text message", "type", msg.Type)
		return
	}

	chatID := msg.From
	customerPhone := service.CleanPhoneNumber(chatID)
	content := msg.Body

	h.logger.Info("OpenWA incoming message", "from", customerPhone, "body", content)

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

	// Extract customer name and avatar from message sender details
	customerName := msg.Sender.Pushname
	if customerName == "" {
		customerName = msg.Sender.Name
	}
	if customerName == "" {
		customerName = customerPhone
	}
	customerAvatar := msg.Sender.ProfilePicThumbObj.Eurl

	// Process through chat service (same flow as web widget)
	conv, aiResp, err := h.chat.DirectChat(c.Request.Context(), userID, customerName, customerPhone, content, "whatsapp", customerAvatar)
	if err != nil {
		h.logger.Error("Failed to process OpenWA message", "error", err)
		return
	}

	// Send AI reply back via OpenWA
	if aiResp != nil && aiResp.Content != "" {
		if err := h.openwa.SendTextMessage(event.SessionID, chatID, aiResp.Content); err != nil {
			h.logger.Error("Failed to send OpenWA reply", "error", err, "chatID", chatID)
		}
	}

	// Broadcast to WebSocket dashboard
	if h.wsHub != nil && conv != nil {
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

	h.logger.Info("OpenWA status update", "id", status.ID, "status", status.Status)
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

		// Configure webhook
		webhookURL := "http://host.docker.internal:8080/api/v1/openwa/webhook"
		h.logger.Info("Background: Configuring webhook", "url", webhookURL, "sessionID", sid)
		if err := h.openwa.ConfigureWebhook(sid, webhookURL, h.cfg.OpenWAWebhookSecret); err != nil {
			h.logger.Warn("Background: Webhook configuration failed (non-critical)", "error", err, "sessionID", sid)
			// Try localhost fallback
			altURL := "http://localhost:8080/api/v1/openwa/webhook"
			_ = h.openwa.ConfigureWebhook(sid, altURL, h.cfg.OpenWAWebhookSecret)
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
				h.logger.Info("Long-polling WhatsApp status cancelled by client connection close", "sessionID", sessionID)
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
		h.logger.Info("Force-connect requested by user — marking session as connected", "sessionID", sessionID)
		isConnected = true
		status = "connected"
		qrCode = ""
	}

	// Update integration status if connected, and notify dashboard via WebSocket
	if isConnected {
		userID, ok := c.Get("userID")
		if ok {
			h.chat.StoreWhatsAppIntegrationWithStatus(c.Request.Context(), userID.(string), sessionID, "", "connected")

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

	// Remove integration and disconnect session completely (logging out credentials)
	userID, _ := c.Get("userID")
	if userID != nil {
		h.chat.DisconnectWhatsAppSession(c.Request.Context(), userID.(string))
		h.chat.RemoveWhatsAppIntegration(c.Request.Context(), userID.(string))
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
