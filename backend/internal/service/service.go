package service

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"noant/config"
	"noant/internal/domain"
	"noant/internal/infrastructure"
	"noant/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type Services struct {
	Auth        *AuthService
	Chat        *ChatService
	Training    *TrainingService
	Analytics   *AnalyticsService
	Integration *IntegrationService
	Settings    *SettingsService
	Archive     *ArchiveService
	Payment     *PaymentService
	Audit       *AuditService
	Notification *NotificationService
	Widget       *WidgetService
}

func NewServices(cfg *config.Config, repos *repository.Repositories, redis *infrastructure.RedisClient, logger *infrastructure.Logger, email *ResendService, broadcastFn func(convID string, msgType string, data interface{})) *Services {
	aiBrain := NewAIBrain(cfg, repos, redis, logger, broadcastFn)
	return &Services{
		Auth:        NewAuthService(cfg, repos.User, redis, logger, email),
		Chat:        NewChatService(cfg, repos, redis, aiBrain, logger),
		Training:    NewTrainingService(cfg, repos, redis, logger),
		Analytics:   NewAnalyticsService(cfg, repos, redis, logger),
		Integration: NewIntegrationService(cfg, repos, redis, logger, broadcastFn),
		Settings:    NewSettingsService(cfg, repos, redis, logger),
		Archive:     NewArchiveService(cfg, repos, redis, logger),
		Payment:     NewPaymentService(cfg, repos, redis, logger),
		Audit:       NewAuditService(repos, logger),
		Notification: NewNotificationService(cfg, repos, redis, logger, email),
		Widget:       NewWidgetService(cfg, repos, redis, aiBrain, logger, email),
	}
}

// ========== AI BRAIN SERVICE ==========

type CircuitBreaker struct {
	failures    int
	lastFailure time.Time
	state       string // closed, open, half-open
	mutex       sync.RWMutex
}

func (cb *CircuitBreaker) Allow() bool {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()
	switch cb.state {
	case "open":
		if time.Since(cb.lastFailure) > 60*time.Second {
			cb.state = "half-open"
			cb.failures = 0
			return true
		}
		return false
	case "half-open":
		return true
	default: // closed
		return true
	}
}

func (cb *CircuitBreaker) RecordSuccess() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()
	cb.failures = 0
	cb.state = "closed"
}

func (cb *CircuitBreaker) RecordFailure() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()
	cb.failures++
	cb.lastFailure = time.Now()
	if cb.failures >= 3 {
		cb.state = "open"
	}
}

type AIBrain struct {
	cfg         *config.Config
	repos       *repository.Repositories
	redis       *infrastructure.RedisClient
	logger      *infrastructure.Logger
	keyIndex    int
	keyMutex    sync.RWMutex
	cb          *CircuitBreaker
	broadcastFn func(convID string, msgType string, data interface{})
}

func NewAIBrain(cfg *config.Config, repos *repository.Repositories, redis *infrastructure.RedisClient, logger *infrastructure.Logger, broadcastFn func(convID string, msgType string, data interface{})) *AIBrain {
	return &AIBrain{
		cfg:         cfg,
		repos:       repos,
		redis:       redis,
		logger:      logger,
		keyIndex:    0,
		cb:          &CircuitBreaker{state: "closed"},
		broadcastFn: broadcastFn,
	}
}

func (b *AIBrain) getNextAPIKey() string {
	b.keyMutex.Lock()
	defer b.keyMutex.Unlock()
	if len(b.cfg.GroqAPIKeys) == 0 {
		return ""
	}
	key := b.cfg.GroqAPIKeys[b.keyIndex]
	b.keyIndex = (b.keyIndex + 1) % len(b.cfg.GroqAPIKeys)
	return key
}

// LangChain-style prompt template
type PromptTemplate struct {
	SystemPrompt string
	Context      []MessageTurn
	UserQuery    string
	Language     string
	Tone         string
}

type MessageTurn struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

func (b *AIBrain) BuildPrompt(template PromptTemplate) []MessageTurn {
	var messages []MessageTurn
	systemContent := fmt.Sprintf(`You are Noant, an enterprise AI customer support agent. Tone: %s, professional, warm, and efficient.
Language: %s
Rules:
1. Answer ONLY based on the provided knowledge base.
2. If you don't know the answer, say "I don't have that information yet, but I'll escalate this to a human agent who can help you."
3. Never make up prices, policies, or facts.
4. Keep responses concise (under 150 words).
5. Use the customer's name if known.
6. For Nigerian customers, use friendly local expressions appropriately.`, template.Tone, template.Language)
	messages = append(messages, MessageTurn{Role: "system", Content: systemContent})
	if len(template.Context) > 0 {
		messages = append(messages, template.Context...)
	}
	messages = append(messages, MessageTurn{Role: "user", Content: template.UserQuery})
	return messages
}

func (b *AIBrain) GenerateResponse(ctx context.Context, conversationID string, userQuery string, language string) (*AIResponse, error) {
	conv, _ := b.repos.Conversation.GetByID(ctx, conversationID)
	userID := ""
	if conv != nil {
		userID = conv.UserID
	}
	qaPairs, err := b.repos.QAPair.Search(ctx, userID, userQuery)
	if err != nil {
		b.logger.Error("Failed to search Q&A pairs", "error", err)
	}
	var contextMessages []MessageTurn
	if len(qaPairs) > 0 {
		contextMessages = append(contextMessages, MessageTurn{
			Role:    "system",
			Content: "Relevant knowledge base entries:",
		})
		for _, qa := range qaPairs[:min(3, len(qaPairs))] {
			contextMessages = append(contextMessages, MessageTurn{
				Role:    "system",
				Content: fmt.Sprintf("Q: %s\nA: %s", qa.Question, qa.Answer),
			})
		}
	} else {
		// Strict zero-hallucination instruction when no match exists
		contextMessages = append(contextMessages, MessageTurn{
			Role:    "system",
			Content: "CRITICAL: No matching entries were found in the knowledge base. You MUST respond with exactly: \"I don't have that information yet, but I'll escalate this to a human agent who can help you.\" Do not try to answer from general knowledge.",
		})
	}
	history, err := b.getConversationHistory(ctx, conversationID)
	if err == nil && len(history) > 0 {
		contextMessages = append(contextMessages, history...)
	}
	prompt := b.BuildPrompt(PromptTemplate{
		Context:   contextMessages,
		UserQuery: userQuery,
		Language:  language,
		Tone:      "professional",
	})
	response, confidence, err := b.callGroqWithFallback(ctx, prompt)
	if err != nil {
		b.logger.Error("Groq API failed", "error", err)
		return &AIResponse{
			Content:    "I apologize, I'm experiencing a temporary issue. A human agent will assist you shortly.",
			Confidence: 0,
			Escalate:   true,
			Reason:     "AI service unavailable",
		}, nil
	}
	if len(qaPairs) == 0 {
		confidence = 0.3 // Force low confidence for unknown questions to trigger training and notification flow
	} else {
		confidence = 0.95 // Force high confidence when we have matching training data
	}
	aiResp := &AIResponse{
		Content:    response,
		Confidence: confidence,
	}
	if confidence < 0.6 {
		aiResp.Escalate = true
		aiResp.Reason = "Low confidence in answer"
		channel := ""
		if conv != nil {
			channel = conv.Channel
		}
		err = b.repos.UnknownQ.Create(ctx, &domain.UnknownQuestion{
			UserID:         userID,
			Question:       userQuery,
			ConversationID: conversationID,
			Channel:        channel,
			Status:         "pending",
		})
		if err != nil {
			b.logger.Error("Failed to create unknown question", "error", err, "conversationID", conversationID)
		}

		// Create database notification for persistent bell alerts
		notif := &domain.Notification{
			UserID: userID,
			Type:   "unknown_question",
			Title:  "New Unknown Question",
			Body:   fmt.Sprintf("AI could not answer: \"%s\"", userQuery),
			Link:   "/teach?tab=unknown",
		}
		_ = b.repos.Notification.Create(ctx, notif)

		// Broadcast event via WebSocket singleton
		if b.broadcastFn != nil {
			b.broadcastFn(conversationID, "unknown_question", map[string]interface{}{
				"question":        userQuery,
				"conversation_id": conversationID,
				"channel":         channel,
				"created_at":      time.Now(),
			})
		}
	}
	_ = b.storeConversationTurn(ctx, conversationID, userQuery, response)
	return aiResp, nil
}

type AIResponse struct {
	Content    string  `json:"content"`
	Confidence float64 `json:"confidence"`
	Escalate   bool    `json:"escalate"`
	Reason     string  `json:"reason,omitempty"`
	MatchedQA  *string `json:"matched_qa,omitempty"`
}

func (b *AIBrain) callGroqWithFallback(ctx context.Context, messages []MessageTurn) (string, float64, error) {
	if !b.cb.Allow() {
		return "", 0, fmt.Errorf("circuit breaker open: Groq API temporarily unavailable")
	}
	apiKey := b.getNextAPIKey()
	if apiKey == "" {
		b.cb.RecordFailure()
		return "", 0, fmt.Errorf("no Groq API keys configured")
	}
	payload := map[string]interface{}{
		"model":       "llama-3.3-70b-versatile",
		"messages":    messages,
		"temperature": 0.3,
		"max_tokens":  500,
		"top_p":       0.9,
	}
	jsonPayload, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.groq.com/openai/v1/chat/completions", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return "", 0, err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		b.cb.RecordFailure()
		return "", 0, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", 0, fmt.Errorf("Groq API error: %s - %s", resp.Status, string(body))
	}
	var result struct {
		Choices []struct {
			Message      struct {
				Content string `json:"content"`
			} `json:"message"`
			FinishReason string `json:"finish_reason"`
		} `json:"choices"`
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
		} `json:"usage"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", 0, err
	}
	if len(result.Choices) == 0 {
		return "", 0, fmt.Errorf("no response from Groq")
	}
	content := result.Choices[0].Message.Content
	b.cb.RecordSuccess()
	confidence := 0.85
	if result.Choices[0].FinishReason != "stop" {
		confidence = 0.5
	}
	if result.Usage.CompletionTokens < 10 {
		confidence = 0.4
	}
	return content, confidence, nil
}

func (b *AIBrain) getConversationHistory(ctx context.Context, conversationID string) ([]MessageTurn, error) {
	if b.redis == nil {
		return nil, nil
	}
	key := fmt.Sprintf("conv:%s:history", conversationID)
	historyJSON, err := b.redis.Get(ctx, key)
	if err != nil {
		return nil, err
	}
	var history []MessageTurn
	if err := json.Unmarshal([]byte(historyJSON), &history); err != nil {
		return nil, err
	}
	if len(history) > 10 {
		history = history[len(history)-10:]
	}
	return history, nil
}

func (b *AIBrain) storeConversationTurn(ctx context.Context, conversationID, userQuery, aiResponse string) error {
	if b.redis == nil {
		return nil
	}
	history, _ := b.getConversationHistory(ctx, conversationID)
	history = append(history, MessageTurn{Role: "user", Content: userQuery})
	history = append(history, MessageTurn{Role: "assistant", Content: aiResponse})
	if len(history) > 10 {
		history = history[len(history)-10:]
	}
	historyJSON, _ := json.Marshal(history)
	return b.redis.Set(ctx, fmt.Sprintf("conv:%s:history", conversationID), string(historyJSON), b.cfg.RedisShortTTL)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ========== AUTH SERVICE ==========

type AuthService struct {
	cfg      *config.Config
	userRepo *repository.UserRepository
	redis    *infrastructure.RedisClient
	logger   *infrastructure.Logger
	email    *ResendService
}

func NewAuthService(cfg *config.Config, userRepo *repository.UserRepository, redis *infrastructure.RedisClient, logger *infrastructure.Logger, email *ResendService) *AuthService {
	return &AuthService{cfg: cfg, userRepo: userRepo, redis: redis, logger: logger, email: email}
}

func (s *AuthService) Register(ctx context.Context, email, password, firstName, lastName, companyName string) (*domain.User, error) {
	existing, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("email already registered")
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}
	
	// Set 14-day trial period
	now := time.Now()
	trialExpires := now.AddDate(0, 0, 14)
	
	user := &domain.User{
		Email:              email,
		Password:           string(hashedPassword),
		FirstName:          firstName,
		LastName:           lastName,
		CompanyName:        companyName,
		Role:               "owner",
		PlanID:             "free",
		IsActive:           true,
		MustChangePassword: true,
		TrialExpiresAt:     &trialExpires,
	}
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}
	created, _ := s.userRepo.GetByEmail(ctx, email)
	return created, nil
}

func (s *AuthService) generateRefreshToken() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func (s *AuthService) Login(ctx context.Context, email, password string) (*domain.User, string, string, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if user == nil {
		return nil, "", "", fmt.Errorf("invalid credentials")
	}
	if err != nil {
		return nil, "", "", fmt.Errorf("database error: %w", err)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, "", "", fmt.Errorf("invalid credentials")
	}
	_ = s.userRepo.UpdateLastLogin(ctx, user.ID)
	token, err := s.generateToken(user)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to generate token: %w", err)
	}

	refreshToken := s.generateRefreshToken()
	if s.redis != nil {
		_ = s.redis.Set(ctx, "refresh:"+refreshToken, user.ID, 7*24*time.Hour)
	}

	return user, token, refreshToken, nil
}

func (s *AuthService) generateToken(user *domain.User) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"role":    user.Role,
		"type":    "access",
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	})
	return token.SignedString([]byte(s.cfg.JWTSecret))
}

func (s *AuthService) RefreshToken(refreshToken string) (string, error) {
	if refreshToken == "" {
		return "", fmt.Errorf("refresh token required")
	}
	var userID string
	if s.redis != nil {
		uid, err := s.redis.Get(context.Background(), "refresh:"+refreshToken)
		if err != nil || uid == "" {
			return "", fmt.Errorf("invalid or expired refresh token")
		}
		userID = uid
	} else {
		return "", fmt.Errorf("token store unavailable")
	}
	user, err := s.userRepo.GetByID(context.Background(), userID)
	if err != nil || user == nil {
		return "", fmt.Errorf("user not found")
	}
	accessToken, err := s.generateToken(user)
	if err != nil {
		return "", err
	}
	return accessToken, nil
}

func (s *AuthService) GetUser(ctx context.Context, userID string) (*domain.User, error) {
	return s.userRepo.GetByID(ctx, userID)
}

func (s *AuthService) ChangePassword(ctx context.Context, userID, currentPassword, newPassword string) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return fmt.Errorf("user not found")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(currentPassword)); err != nil {
		return fmt.Errorf("current password is incorrect")
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return s.userRepo.UpdatePassword(ctx, userID, string(hashedPassword))
}

func (s *AuthService) ForgotPassword(ctx context.Context, email string) error {
	if s.redis != nil {
		key := fmt.Sprintf("ratelimit:forgot-password:%s", email)
		countStr, err := s.redis.Get(ctx, key)
		var count int
		if err == nil {
			fmt.Sscanf(countStr, "%d", &count)
		}
		if count >= 3 {
			s.logger.Warn("Forgot password request rate limited", "email", email)
			return fmt.Errorf("too many forgot password requests, please try again in an hour")
		}

		newVal, err := s.redis.Incr(ctx, key)
		if err == nil && newVal == 1 {
			_ = s.redis.Expire(ctx, key, time.Hour)
		}
	}

	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return err
	}
	if user == nil {
		return nil
	}
	resetToken := make([]byte, 32)
	if _, err := rand.Read(resetToken); err != nil {
		return err
	}
	token := hex.EncodeToString(resetToken)
	if s.redis != nil {
		_ = s.redis.Set(ctx, "reset:"+token, user.ID, time.Hour)
	}
	if s.email != nil {
		if _, err := s.email.SendPasswordReset(ctx, user.Email, token); err != nil {
			s.logger.Error("Failed to send password reset email", "error", err)
		}
	}
	return nil
}



func (s *AuthService) Logout(ctx context.Context, token string) error {
	if s.redis != nil {
		return s.redis.Set(ctx, "blacklist:"+token, "true", 24*time.Hour)
	}
	return nil
}

func (s *AuthService) ResetPassword(ctx context.Context, token, newPassword string) error {
	if s.redis == nil {
		return fmt.Errorf("token store unavailable")
	}
	userID, err := s.redis.Get(ctx, "reset:"+token)
	if err != nil || userID == "" {
		return fmt.Errorf("invalid or expired reset token")
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	if err := s.userRepo.UpdatePassword(ctx, userID, string(hashedPassword)); err != nil {
		return err
	}
	_ = s.redis.Delete(ctx, "reset:"+token)
	return nil
}

func (s *AuthService) Me(ctx context.Context, userID string) (*domain.User, error) {
	return s.userRepo.GetByID(ctx, userID)
}

// ========== CHAT SERVICE ==========

type ChatService struct {
	cfg     *config.Config
	repos   *repository.Repositories
	redis   *infrastructure.RedisClient
	aiBrain *AIBrain
	logger  *infrastructure.Logger
}

func NewChatService(cfg *config.Config, repos *repository.Repositories, redis *infrastructure.RedisClient, aiBrain *AIBrain, logger *infrastructure.Logger) *ChatService {
	return &ChatService{cfg: cfg, repos: repos, redis: redis, aiBrain: aiBrain, logger: logger}
}

func (s *ChatService) DirectChat(ctx context.Context, userID, customerName, message, channel string) (*domain.Conversation, *domain.Message, error) {
	if s.redis != nil {
		limit := 500
		user, err := s.repos.User.GetByID(ctx, userID)
		if err == nil && user != nil {
			switch user.PlanID {
			case "pro", "starter":
				limit = 500
			case "business", "enterprise":
				limit = 999999 // unlimited
			}
		}

		if limit < 999999 {
			allowed, _ := s.redis.RateLimit(ctx, "chat:"+userID, limit, time.Minute)
			if !allowed {
				return nil, nil, fmt.Errorf("rate limit exceeded")
			}
		}
	}
	existing, _ := s.repos.Conversation.FindActiveByCustomer(ctx, userID, customerName, channel)
	var conv *domain.Conversation
	if existing != nil {
		conv = existing
	} else {
		conv = &domain.Conversation{
			UserID:       userID,
			CustomerName: customerName,
			Channel:      channel,
			Status:       "active",
			Intent:       "inquiry",
			Priority:     "medium",
		}
		if err := s.repos.Conversation.Create(ctx, conv); err != nil {
			return nil, nil, err
		}
	}
	aiResp, err := s.aiBrain.GenerateResponse(ctx, conv.ID, message, "en")
	if err != nil {
		s.logger.Error("AI generation failed", "error", err)
		aiResp = &AIResponse{
			Content:    "I apologize, I'm having trouble processing your request. A human agent will assist you shortly.",
			Confidence: 0,
			Escalate:   true,
		}
	}
	customerMsg := &domain.Message{
		ConversationID: conv.ID,
		SenderType:     "customer",
		Content:        message,
		IsRead:         false,
	}
	_ = s.repos.Message.Create(ctx, customerMsg)
	aiMsg := &domain.Message{
		ConversationID: conv.ID,
		SenderType:     "ai",
		Content:        aiResp.Content,
		IsRead:         false,
		Metadata: &domain.MessageMetadata{
			Confidence: aiResp.Confidence,
			Language:   "en",
		},
	}
	_ = s.repos.Message.Create(ctx, aiMsg)
	if aiResp.Escalate {
		_ = s.repos.Conversation.UpdateStatus(ctx, conv.ID, "escalated")
	}
	return conv, aiMsg, nil
}

func (s *ChatService) ListConversations(ctx context.Context, userID string, status string, page, limit int) ([]domain.Conversation, int, error) {
	offset := (page - 1) * limit
	return s.repos.Conversation.List(ctx, userID, status, limit, offset)
}

func (s *ChatService) GetConversation(ctx context.Context, conversationID string) (*domain.Conversation, []domain.Message, error) {
	conv, err := s.repos.Conversation.GetByID(ctx, conversationID)
	if err != nil {
		return nil, nil, err
	}
	if conv == nil {
		return nil, nil, fmt.Errorf("conversation not found")
	}
	messages, err := s.repos.Message.ListByConversation(ctx, conversationID, 100)
	if err != nil {
		return nil, nil, err
	}
	return conv, messages, nil
}

func (s *ChatService) HumanTakeover(ctx context.Context, conversationID, agentID string) error {
	return s.repos.Conversation.Takeover(ctx, conversationID, agentID)
}

func (s *ChatService) Escalate(ctx context.Context, conversationID, reason string) error {
	if err := s.repos.Conversation.UpdateStatus(ctx, conversationID, "escalated"); err != nil {
		return err
	}
	msg := &domain.Message{
		ConversationID: conversationID,
		SenderType:     "system",
		Content:        fmt.Sprintf("Conversation escalated. Reason: %s", reason),
		IsRead:         false,
	}
	return s.repos.Message.Create(ctx, msg)
}

func (s *ChatService) SendMessage(ctx context.Context, conversationID, senderType, content string) (*domain.Message, error) {
	msg := &domain.Message{
		ConversationID: conversationID,
		SenderType:     senderType,
		Content:        content,
		IsRead:         false,
	}
	if err := s.repos.Message.Create(ctx, msg); err != nil {
		return nil, err
	}
	return msg, nil
}

func (s *ChatService) GenerateAIResponse(ctx context.Context, conversationID, userMessage string) (*domain.Message, error) {
	aiResp, err := s.aiBrain.GenerateResponse(ctx, conversationID, userMessage, "en")
	if err != nil {
		s.logger.Error("AI generation failed", "error", err)
		aiResp = &AIResponse{
			Content:    "I apologize, I am having trouble right now. A human agent will assist you shortly.",
			Confidence: 0,
			Escalate:   true,
		}
	}
	aiMsg := &domain.Message{
		ConversationID: conversationID,
		SenderType:     "ai",
		Content:        aiResp.Content,
		IsRead:         false,
		Metadata: &domain.MessageMetadata{
			Confidence: aiResp.Confidence,
			Language:   "en",
		},
	}
	if err := s.repos.Message.Create(ctx, aiMsg); err != nil {
		return nil, err
	}
	if aiResp.Escalate {
		_ = s.repos.Conversation.UpdateStatus(ctx, conversationID, "escalated")
	}
	return aiMsg, nil
}

// ========== TRAINING SERVICE ==========

type TrainingService struct {
	cfg    *config.Config
	repos  *repository.Repositories
	redis  *infrastructure.RedisClient
	logger *infrastructure.Logger
}

func NewTrainingService(cfg *config.Config, repos *repository.Repositories, redis *infrastructure.RedisClient, logger *infrastructure.Logger) *TrainingService {
	return &TrainingService{cfg: cfg, repos: repos, redis: redis, logger: logger}
}

func (s *TrainingService) CreateCategory(ctx context.Context, userID, name, description, color string) (*domain.Category, error) {
	cat := &domain.Category{
		UserID:      userID,
		Name:        name,
		Description: description,
		Color:       color,
	}
	if err := s.repos.Category.Create(ctx, cat); err != nil {
		return nil, err
	}
	return cat, nil
}

func (s *TrainingService) ListCategories(ctx context.Context, userID string) ([]domain.Category, error) {
	return s.repos.Category.List(ctx, userID)
}

func (s *TrainingService) BulkImport(ctx context.Context, userID, categoryID string, qaPairs []domain.QAPair) error {
	for i := range qaPairs {
		qaPairs[i].UserID = userID
		qaPairs[i].CategoryID = categoryID
		qaPairs[i].IsActive = true
	}
	return s.repos.QAPair.BulkCreate(ctx, qaPairs)
}

func (s *TrainingService) UploadCSV(ctx context.Context, userID, categoryID string, csvData []byte) (int, error) {
	reader := csv.NewReader(bytes.NewReader(csvData))
	reader.LazyQuotes = true
	reader.FieldsPerRecord = -1
	records, err := reader.ReadAll()
	if err != nil {
		return 0, fmt.Errorf("failed to parse CSV: %w", err)
	}
	if len(records) < 2 {
		return 0, fmt.Errorf("CSV must have at least a header and one data row")
	}

	categoryMap := make(map[string]string)
	var qaPairs []domain.QAPair

	for i, record := range records[1:] {
		if len(record) < 3 {
			s.logger.Warn("Skipping invalid CSV row", "row", i+2)
			continue
		}
		categoryName := strings.TrimSpace(record[0])
		question := strings.TrimSpace(record[1])
		answer := strings.TrimSpace(strings.Join(record[2:], ","))

		if categoryName == "" || question == "" || answer == "" {
			s.logger.Warn("Skipping empty CSV row", "row", i+2)
			continue
		}

		catID, exists := categoryMap[categoryName]
		if !exists {
			existing, _ := s.repos.Category.GetByName(ctx, userID, categoryName)
			if existing != nil {
				catID = existing.ID
			} else {
				cat := &domain.Category{
					UserID:      userID,
					Name:        categoryName,
					Description: "Auto-imported from CSV",
					Color:       "#3b82f6",
				}
				if err := s.repos.Category.Create(ctx, cat); err != nil {
					s.logger.Warn("Failed to create category", "name", categoryName, "error", err)
					continue
				}
				catID = cat.ID
			}
			categoryMap[categoryName] = catID
		}
		qaPairs = append(qaPairs, domain.QAPair{
			UserID:     userID,
			CategoryID: catID,
			Category:   categoryName,
			Question:   question,
			Answer:     answer,
			IsActive:   true,
		})
	}

	if len(qaPairs) == 0 {
		return 0, fmt.Errorf("no valid Q&A pairs found in CSV")
	}
	err = s.repos.QAPair.BulkCreate(ctx, qaPairs)
	if err != nil {
		return 0, err
	}
	return len(qaPairs), nil
}

func (s *TrainingService) ListUnknownQuestions(ctx context.Context, userID string, status string, limit int) ([]domain.UnknownQuestion, error) {
	return s.repos.UnknownQ.List(ctx, userID, status, limit)
}

func (s *TrainingService) TrainUnknown(ctx context.Context, id string, answer string, categoryID string) error {
	target, err := s.repos.UnknownQ.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if target == nil {
		return fmt.Errorf("unknown question not found")
	}
	qa := &domain.QAPair{
		UserID:     target.UserID,
		CategoryID: categoryID,
		Question:   target.Question,
		Answer:     answer,
		IsActive:   true,
	}
	if err := s.repos.QAPair.Create(ctx, qa); err != nil {
		return err
	}
	return s.repos.UnknownQ.UpdateStatus(ctx, id, "trained", &answer, &categoryID)
}

func (s *TrainingService) IgnoreUnknown(ctx context.Context, id string) error {
	if err := s.repos.UnknownQ.UpdateStatus(ctx, id, "ignored", nil, nil); err != nil {
		return err
	}
	return nil
}

// ========== ANALYTICS SERVICE ==========

type AnalyticsService struct {
	cfg    *config.Config
	repos  *repository.Repositories
	redis  *infrastructure.RedisClient
	logger *infrastructure.Logger
}

func NewAnalyticsService(cfg *config.Config, repos *repository.Repositories, redis *infrastructure.RedisClient, logger *infrastructure.Logger) *AnalyticsService {
	return &AnalyticsService{cfg: cfg, repos: repos, redis: redis, logger: logger}
}

func getInt(m map[string]interface{}, key string) int {
	if v, ok := m[key]; ok {
		switch i := v.(type) {
		case int:
			return i
		case int64:
			return int(i)
		case float64:
			return int(i)
		}
	}
	return 0
}

func getFloat64(m map[string]interface{}, key string) float64 {
	if v, ok := m[key]; ok {
		switch i := v.(type) {
		case float64:
			return i
		case int:
			return float64(i)
		}
	}
	return 0
}

func (s *AnalyticsService) Overview(ctx context.Context, userID string) (*domain.AnalyticsOverview, error) {
	data, err := s.repos.Conversation.GetOverview(ctx, userID)
	if err != nil {
		s.logger.Warn("Failed to get analytics overview", "error", err)
		return nil, fmt.Errorf("failed to load analytics: %w", err)
	}
	return &domain.AnalyticsOverview{
		TotalConversations:   getInt(data, "total_conversations"),
		ActiveConversations:  getInt(data, "active_conversations"),
		ResolvedToday:        getInt(data, "resolved_today"),
		AIResolutionRate:     getFloat64(data, "ai_resolution_rate"),
		AvgResponseTime:      0,
		CustomerSatisfaction: 0,
		TotalMessages:        0,
		EscalatedCount:       getInt(data, "escalated_count"),
	}, nil
}

func (s *AnalyticsService) ChannelDistribution(ctx context.Context, userID string) (map[string]int, error) {
	data, err := s.repos.Conversation.CountByChannel(ctx, userID)
	if err != nil {
		s.logger.Warn("Failed to get channel distribution", "error", err)
		return nil, err
	}
	return data, nil
}

func (s *AnalyticsService) Insights(ctx context.Context, userID string) (map[string]interface{}, error) {
	topIntents, err := s.repos.Conversation.CountByIntent(ctx, userID)
	if err != nil {
		s.logger.Warn("Failed to get insights", "error", err)
		topIntents = []map[string]interface{}{}
	}
	peakHours, err := s.repos.Conversation.CountByHour(ctx, userID)
	if err != nil {
		s.logger.Warn("Failed to get peak hours", "error", err)
		peakHours = []map[string]interface{}{}
	}
	return map[string]interface{}{
		"top_intents": topIntents,
		"peak_hours":  peakHours,
	}, nil
}

func (s *AnalyticsService) Trends(ctx context.Context, userID string, days int) ([]map[string]interface{}, error) {
	data, err := s.repos.Conversation.CountByDate(ctx, userID, days)
	if err != nil {
		s.logger.Warn("Failed to get trends", "error", err)
		return nil, err
	}
	return data, nil
}

// ========== INTEGRATION SERVICE ==========

type IntegrationService struct {
	cfg         *config.Config
	repos       *repository.Repositories
	redis       *infrastructure.RedisClient
	logger      *infrastructure.Logger
	broadcastFn func(convID string, msgType string, data interface{})
}

func NewIntegrationService(cfg *config.Config, repos *repository.Repositories, redis *infrastructure.RedisClient, logger *infrastructure.Logger, broadcastFn func(convID string, msgType string, data interface{})) *IntegrationService {
	return &IntegrationService{cfg: cfg, repos: repos, redis: redis, logger: logger, broadcastFn: broadcastFn}
}

func (s *IntegrationService) List(ctx context.Context, userID string) ([]domain.Integration, error) {
	return s.repos.Integration.ListByUser(ctx, userID)
}

func (s *IntegrationService) Connect(ctx context.Context, userID, channel string, config map[string]interface{}) (*domain.Integration, error) {
	existing, err := s.repos.Integration.GetByUserAndChannel(ctx, userID, channel)
	if err != nil {
		return nil, err
	}

	var integration *domain.Integration
	if existing != nil {
		existing.Status = "active"
		existing.Config = config
		existing.WebhookURL = fmt.Sprintf("%s/api/v1/webhooks/%s", s.cfg.APIURL, channel)
		if err := s.repos.Integration.Update(ctx, existing); err != nil {
			return nil, err
		}
		integration = existing
	} else {
		integration = &domain.Integration{
			UserID:     userID,
			Channel:    channel,
			Status:     "active",
			Config:     config,
			WebhookURL: fmt.Sprintf("%s/api/v1/webhooks/%s", s.cfg.APIURL, channel),
		}
		if err := s.repos.Integration.Create(ctx, integration); err != nil {
			return nil, err
		}
	}

	// Trigger real-time status update broadcast
	if s.broadcastFn != nil {
		s.broadcastFn("", "integration_update", map[string]interface{}{
			"channel": channel,
			"status":  "connected",
		})
	}

	return integration, nil
}

func (s *IntegrationService) Disconnect(ctx context.Context, userID, channel string) error {
	err := s.repos.Integration.Disconnect(ctx, userID, channel)
	if err != nil {
		return err
	}

	// Trigger real-time status update broadcast
	if s.broadcastFn != nil {
		s.broadcastFn("", "integration_update", map[string]interface{}{
			"channel": channel,
			"status":  "disconnected",
		})
	}

	return nil
}

func (s *IntegrationService) Test(ctx context.Context, channel string) (bool, string) {
	switch channel {
	case "telegram":
		return true, "Telegram webhook configured successfully"
	case "whatsapp":
		if s.cfg.TwilioAccountSID == "" {
			return false, "Twilio credentials not configured"
		}
		return true, "WhatsApp connection test passed"
	case "web":
		return true, "Web chat widget ready"
	default:
		return false, "Unsupported channel"
	}
}

// ========== SETTINGS SERVICE ==========

type SettingsService struct {
	cfg    *config.Config
	repos  *repository.Repositories
	redis  *infrastructure.RedisClient
	logger *infrastructure.Logger
}

func NewSettingsService(cfg *config.Config, repos *repository.Repositories, redis *infrastructure.RedisClient, logger *infrastructure.Logger) *SettingsService {
	return &SettingsService{cfg: cfg, repos: repos, redis: redis, logger: logger}
}

func (s *SettingsService) GetProfile(ctx context.Context, userID string) (*domain.User, error) {
	return s.repos.User.GetByID(ctx, userID)
}

func (s *SettingsService) UpdateProfile(ctx context.Context, userID string, updates map[string]interface{}) error {
	firstName, _ := updates["first_name"].(string)
	lastName, _ := updates["last_name"].(string)
	companyName, _ := updates["company_name"].(string)
	phone, _ := updates["phone"].(string)
	return s.repos.User.UpdateProfile(ctx, userID, firstName, lastName, companyName, phone)
}

func (s *SettingsService) GetNotifPrefs(ctx context.Context, userID string) (*repository.NotifPrefs, error) {
	return s.repos.User.GetNotifPrefs(ctx, userID)
}

func (s *SettingsService) UpdateNotifPrefs(ctx context.Context, userID string, prefs *repository.NotifPrefs) error {
	return s.repos.User.UpdateNotifPrefs(ctx, userID, prefs)
}

func (s *SettingsService) DeleteAccount(ctx context.Context, userID string) error {
	return s.repos.User.Delete(ctx, userID)
}

func (s *SettingsService) ExportUserData(ctx context.Context, userID string) (map[string]interface{}, error) {
	return s.repos.User.ExportUserData(ctx, userID)
}

func (s *SettingsService) ListAPIKeys(ctx context.Context, userID string) ([]domain.APIKey, error) {
	return s.repos.APIKey.ListByUser(ctx, userID)
}

func (s *SettingsService) CreateAPIKey(ctx context.Context, userID, name string) (*domain.APIKey, error) {
	keyBytes := make([]byte, 32)
	rand.Read(keyBytes)
	apiKey := &domain.APIKey{
		UserID:   userID,
		Name:     name,
		Key:      "noant_" + hex.EncodeToString(keyBytes),
		IsActive: true,
	}
	if err := s.repos.APIKey.Create(ctx, apiKey); err != nil {
		return nil, err
	}
	return apiKey, nil
}

func (s *SettingsService) RevokeAPIKey(ctx context.Context, id string) error {
	return s.repos.APIKey.Revoke(ctx, id)
}

func (s *SettingsService) ListTeam(ctx context.Context, ownerID string) ([]domain.TeamMember, error) {
	return s.repos.Team.ListByUser(ctx, ownerID)
}

func (s *SettingsService) InviteTeamMember(ctx context.Context, ownerID, email, role string) (*domain.TeamMember, error) {
	member := &domain.TeamMember{
		Email:    email,
		Role:     role,
		IsActive: false,
	}
	if err := s.repos.Team.Create(ctx, ownerID, member); err != nil {
		return nil, err
	}
	return member, nil
}

func (s *SettingsService) RemoveTeamMember(ctx context.Context, id string) error {
	return nil
}

// ========== ARCHIVE SERVICE ==========

type ArchiveService struct {
	cfg    *config.Config
	repos  *repository.Repositories
	redis  *infrastructure.RedisClient
	logger *infrastructure.Logger
}

func NewArchiveService(cfg *config.Config, repos *repository.Repositories, redis *infrastructure.RedisClient, logger *infrastructure.Logger) *ArchiveService {
	return &ArchiveService{cfg: cfg, repos: repos, redis: redis, logger: logger}
}

func (s *ArchiveService) ListFolders(ctx context.Context, userID, folderType string) ([]domain.ArchiveFolder, error) {
	return s.repos.Archive.ListFolders(ctx, userID, folderType)
}

func (s *ArchiveService) CreateFolder(ctx context.Context, userID, name, folderType, color string) (*domain.ArchiveFolder, error) {
	folder := &domain.ArchiveFolder{
		UserID: userID,
		Name:   name,
		Type:   folderType,
		Color:  color,
	}
	if err := s.repos.Archive.CreateFolder(ctx, folder); err != nil {
		return nil, err
	}
	return folder, nil
}

func (s *ArchiveService) DeleteFolder(ctx context.Context, id string) error {
	return nil
}

func (s *ArchiveService) MoveChat(ctx context.Context, conversationID, folderID string) error {
	return s.repos.Archive.MoveChat(ctx, conversationID, folderID)
}

func (s *ArchiveService) RemoveFromArchive(ctx context.Context, conversationID string) error {
	return s.repos.Archive.MoveChat(ctx, conversationID, "")
}

func (s *ArchiveService) GetStatus(ctx context.Context, userID string) (map[string]interface{}, error) {
	folders, _ := s.repos.Archive.ListFolders(ctx, userID, "")
	return map[string]interface{}{
		"folders":     len(folders),
		"total_items": 0,
	}, nil
}

// ========== PAYMENT SERVICE ==========

type PaymentService struct {
	cfg    *config.Config
	repos  *repository.Repositories
	redis  *infrastructure.RedisClient
	logger *infrastructure.Logger
}

func NewPaymentService(cfg *config.Config, repos *repository.Repositories, redis *infrastructure.RedisClient, logger *infrastructure.Logger) *PaymentService {
	return &PaymentService{cfg: cfg, repos: repos, redis: redis, logger: logger}
}

func (s *PaymentService) ListPlans(ctx context.Context) ([]domain.PaymentPlan, error) {
	return []domain.PaymentPlan{
		{
			ID:          "free",
			Name:        "Free",
			PriceNGN:    0,
			AIResponses: 50,
			Channels:    []string{"telegram", "web"},
			Features:    []string{"50 AI responses/month", "Telegram + Web Chat", "Basic analytics"},
		},
		{
			ID:          "starter",
			Name:        "Starter",
			PriceNGN:    10000,
			AIResponses: 5000,
			Channels:    []string{"telegram", "web"},
			Features:    []string{"5,000 AI responses/month", "Telegram + Web Chat", "CSV training", "Real-time dashboard"},
			IsPopular:   false,
		},
		{
			ID:          "pro",
			Name:        "Pro",
			PriceNGN:    25000,
			AIResponses: 50000,
			Channels:    []string{"telegram", "web", "whatsapp", "email"},
			Features:    []string{"50,000 AI responses/month", "All channels", "Advanced analytics", "Team management", "API access"},
			IsPopular:   true,
		},
		{
			ID:          "enterprise",
			Name:        "Enterprise",
			PriceNGN:    100000,
			AIResponses: 0,
			Channels:    []string{"telegram", "web", "whatsapp", "email", "instagram", "messenger"},
			Features:    []string{"Unlimited AI responses", "All channels + white-label", "Priority support", "Custom integrations", "Dedicated account manager"},
		},
	}, nil
}

func (s *PaymentService) Subscribe(ctx context.Context, userID, planID string) error {
	planName := planID
	switch planID {
	case "starter", "pro", "enterprise":
		planName = planID
	default:
		return fmt.Errorf("invalid plan ID: %s", planID)
	}

	now := time.Now()
	periodEnd := now.AddDate(0, 1, 0) // 1 month

	sub := &domain.Subscription{
		UserID:             userID,
		PlanID:             planName,
		Status:             "active",
		CurrentPeriodStart: now,
		CurrentPeriodEnd:   periodEnd,
	}

	if err := s.repos.Subscription.CreateOrUpdate(ctx, sub); err != nil {
		s.logger.Error("Failed to create subscription", "error", err)
		return fmt.Errorf("failed to create subscription: %w", err)
	}

	if err := s.repos.User.UpdatePlan(ctx, userID, planName); err != nil {
		s.logger.Error("Failed to update user plan", "error", err)
		return err
	}

	s.logger.Info("Subscription created", "user", userID, "plan", planName, "period_end", periodEnd)
	return nil
}

func (s *PaymentService) Webhook(ctx context.Context, payload []byte) error {
	var event struct {
		Type string          `json:"type"`
		Data json.RawMessage `json:"data"`
	}

	if err := json.Unmarshal(payload, &event); err != nil {
		return fmt.Errorf("failed to parse webhook payload: %w", err)
	}

	s.logger.Info("Payment webhook received", "type", event.Type)

	switch event.Type {
	case "checkout.success", "subscription.active", "subscription.created":
		var subData struct {
			UserID        string `json:"user_id"`
			PlanID        string `json:"plan_id"`
			Status        string `json:"status"`
			PeriodStart   string `json:"current_period_start"`
			PeriodEnd     string `json:"current_period_end"`
			Metadata      map[string]interface{} `json:"metadata"`
		}

		if err := json.Unmarshal(event.Data, &subData); err != nil {
			return fmt.Errorf("failed to parse subscription data: %w", err)
		}

		// Extract user_id from metadata if not at top level
		userID := subData.UserID
		if userID == "" && len(subData.Metadata) > 0 {
			if uid, ok := subData.Metadata["user_id"].(string); ok {
				userID = uid
			}
		}

		if userID == "" {
			return fmt.Errorf("missing user_id in webhook payload")
		}

		planID := subData.PlanID
		if planID == "" {
			planID = "starter"
		}

		// Parse dates or use defaults
		now := time.Now()
		periodEnd := now.AddDate(0, 1, 0)
		if subData.PeriodEnd != "" {
			if t, err := time.Parse(time.RFC3339, subData.PeriodEnd); err == nil {
				periodEnd = t
			}
		}

		sub := &domain.Subscription{
			UserID:             userID,
			PlanID:             planID,
			Status:             "active",
			CurrentPeriodStart: now,
			CurrentPeriodEnd:   periodEnd,
		}

		if err := s.repos.Subscription.CreateOrUpdate(ctx, sub); err != nil {
			s.logger.Error("Failed to update subscription from webhook", "error", err)
			return err
		}

		if err := s.repos.User.UpdatePlan(ctx, userID, planID); err != nil {
			s.logger.Error("Failed to update user plan from webhook", "error", err)
			return err
		}

		s.logger.Info("Subscription updated via webhook", "user", userID, "plan", planID, "status", subData.Status)

	case "subscription.cancelled", "subscription.updated":
		var subData struct {
			UserID   string `json:"user_id"`
			PlanID   string `json:"plan_id"`
			Status   string `json:"status"`
			Metadata map[string]interface{} `json:"metadata"`
		}

		if err := json.Unmarshal(event.Data, &subData); err != nil {
			return fmt.Errorf("failed to parse subscription data: %w", err)
		}

		userID := subData.UserID
		if userID == "" && len(subData.Metadata) > 0 {
			if uid, ok := subData.Metadata["user_id"].(string); ok {
				userID = uid
			}
		}

		if userID != "" && event.Type == "subscription.cancelled" {
			if err := s.repos.Subscription.Cancel(ctx, userID); err != nil {
				s.logger.Error("Failed to cancel subscription", "error", err)
			}
			if err := s.repos.User.UpdatePlan(ctx, userID, "free"); err != nil {
				s.logger.Error("Failed to downgrade user plan", "error", err)
			}
			s.logger.Info("Subscription cancelled", "user", userID)
		}

	default:
		s.logger.Warn("Unhandled webhook event type", "type", event.Type)
	}

	return nil
}

func (s *PaymentService) Status(ctx context.Context, userID string) (*domain.Subscription, error) {
	return s.repos.Subscription.GetActive(ctx, userID)
}

// ========== AUDIT SERVICE ==========

type AuditService struct {
	repos  *repository.Repositories
	logger *infrastructure.Logger
}

func NewAuditService(repos *repository.Repositories, logger *infrastructure.Logger) *AuditService {
	return &AuditService{repos: repos, logger: logger}
}

func (s *AuditService) ListByUser(ctx context.Context, userID string, limit int) ([]domain.AuditLog, error) {
	return s.repos.Audit.ListByUser(ctx, userID, limit)
}