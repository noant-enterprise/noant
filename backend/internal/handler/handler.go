package handler

import (
	"context"
	"io"
	"net/http"
	"strconv"
	"time"

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
}

func NewHandlers(services *service.Services, logger *infrastructure.Logger, wsHub *WebSocketHub) *Handlers {
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
	conv, msg, err := h.service.DirectChat(c.Request.Context(), userID.(string), req.CustomerName, req.Message, req.Channel)
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
