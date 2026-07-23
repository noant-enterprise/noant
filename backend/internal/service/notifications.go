package service

import (
	"context"
	crypto_rand "crypto/rand"
	"errors"
	"fmt"
	"time"
	
	"noant/config"
	"noant/internal/domain"
	"noant/internal/infrastructure"
	"noant/internal/repository"
)

// ========== NOTIFICATION SERVICE ==========

type NotificationService struct {
	cfg    *config.Config
	repos  *repository.Repositories
	redis  *infrastructure.RedisClient
	logger *infrastructure.Logger
	email  *EmailService
	push   *PushNotificationService
}

func NewNotificationService(cfg *config.Config, repos *repository.Repositories, redis *infrastructure.RedisClient, logger *infrastructure.Logger, email *EmailService, push *PushNotificationService) *NotificationService {
	return &NotificationService{cfg: cfg, repos: repos, redis: redis, logger: logger, email: email, push: push}
}

func (s *NotificationService) Create(ctx context.Context, n *domain.Notification) error {
	if err := s.repos.Notification.Create(ctx, n); err != nil {
		return err
	}
	// Also send push notification if VAPID keys are configured
	if n.UserID != "" && n.Title != "" {
		go func() {
			defer func() {
				if r := recover(); r != nil {
					s.logger.Error("Panic in push notification goroutine", "error", r)
				}
			}()
			pushCtx := context.Background()
			if err := s.push.SendToUser(pushCtx, n.UserID, n.Title, n.Body, n.Link); err != nil {
				s.logger.Warn("Push notification send failed", "error", err)
			}
		}()
	}
	return nil
}

func (s *NotificationService) List(ctx context.Context, userID string, limit int) ([]*domain.Notification, error) {
	if limit <= 0 {
		limit = 50
	}
	return s.repos.Notification.ListByUser(ctx, userID, limit)
}

func (s *NotificationService) UnreadCount(ctx context.Context, userID string) (int, error) {
	return s.repos.Notification.UnreadCount(ctx, userID)
}

func (s *NotificationService) MarkRead(ctx context.Context, id, userID string) error {
	return s.repos.Notification.MarkRead(ctx, id, userID)
}

func (s *NotificationService) MarkAllRead(ctx context.Context, userID string) error {
	return s.repos.Notification.MarkAllRead(ctx, userID)
}

// ========== WIDGET SERVICE ==========

type WidgetService struct {
	cfg     *config.Config
	repos   *repository.Repositories
	redis   *infrastructure.RedisClient
	aiBrain *AIBrain
	logger  *infrastructure.Logger
	email   *EmailService
}

func NewWidgetService(cfg *config.Config, repos *repository.Repositories, redis *infrastructure.RedisClient, aiBrain *AIBrain, logger *infrastructure.Logger, email *EmailService) *WidgetService {
	return &WidgetService{cfg: cfg, repos: repos, redis: redis, aiBrain: aiBrain, logger: logger, email: email}
}

func (s *WidgetService) Get(ctx context.Context, userID string) (*domain.WidgetConfig, error) {
	cfg, err := s.repos.WidgetConfig.Get(ctx, userID)
	if err != nil {
		return nil, err
	}
	if cfg == nil {
		// Return a default config if none exists yet
		cfg = &domain.WidgetConfig{
			UserID:       userID,
			OrgID:        userID,
			BrandColor:   "#3b82f6",
			Greeting:     "Hello! How can we help you today?",
			BotName:      "Noant Bot",
			Position:     "right",
			WidgetAPIKey: "widget_" + generateRandomString(24),
			IsActive:     true,
		}
		err = s.repos.WidgetConfig.Upsert(ctx, cfg)
		if err != nil {
			return nil, err
		}
	}
	return cfg, nil
}

func (s *WidgetService) GetByAPIKey(ctx context.Context, apiKey string) (*domain.WidgetConfig, error) {
	return s.repos.WidgetConfig.GetByAPIKey(ctx, apiKey)
}

func (s *WidgetService) Upsert(ctx context.Context, cfg *domain.WidgetConfig) error {
	if cfg.WidgetAPIKey == "" {
		cfg.WidgetAPIKey = "widget_" + generateRandomString(24)
	}
	return s.repos.WidgetConfig.Upsert(ctx, cfg)
}

func (s *WidgetService) PublicChat(ctx context.Context, apiKey, message, conversationID string) (aiResp, convID string, err error) {
	cfg, err := s.repos.WidgetConfig.GetByAPIKey(ctx, apiKey)
	if err != nil {
		return "", "", err
	}
	if cfg == nil {
		return "", "", errors.New("invalid or inactive widget API key")
	}

	userID := cfg.UserID

	// Find or create conversation
	var conv *domain.Conversation
	if conversationID != "" {
		conv, err = s.repos.Conversation.GetByID(ctx, conversationID)
		if err != nil {
			return "", "", err
		}
	}

	if conv == nil {
		conv = &domain.Conversation{
			UserID:       userID,
			CustomerName: "Web Visitor",
			Channel:      "web",
			Status:       "active",
			Intent:       "inquiry",
			Priority:     "medium",
		}
		if err := s.repos.Conversation.Create(ctx, conv); err != nil {
			return "", "", err
		}
	}

	// Rate limit check
	if s.redis != nil {
		allowed, _ := s.redis.RateLimit(ctx, "chat:"+userID, 100, time.Minute)
		if !allowed {
			return "", conv.ID, errors.New("rate limit exceeded")
		}
	}

	// Create customer message
	customerMsg := &domain.Message{
		ConversationID: conv.ID,
		Role:           "customer",
		Content:        message,
		IsRead:         false,
	}
	if err := s.repos.Message.Create(ctx, customerMsg); err != nil {
		s.logger.Error("Widget: failed to save customer message", "error", err, "conv_id", conv.ID)
	}

	// Generate response
	aiResult, err := s.aiBrain.GenerateResponse(ctx, conv.ID, message, "en")
	if err != nil {
		s.logger.Error("Widget AI generation failed", "error", err)
		aiResult = &AIResponse{
			Content:    "I apologize, I am having trouble processing your request. A human agent will assist you shortly.",
			Confidence: 0,
			Escalate:   true,
		}
	}

	// Create AI message
	aiMsg := &domain.Message{
		ConversationID: conv.ID,
		Role:           "ai",
		Content:        aiResult.Content,
		IsRead:         false,
		Metadata: &domain.MessageMetadata{
			Confidence: aiResult.Confidence,
			Language:   "en",
		},
	}
	if err := s.repos.Message.Create(ctx, aiMsg); err != nil {
		s.logger.Error("Widget: failed to save AI message", "error", err, "conv_id", conv.ID)
	}

	if aiResult.Escalate {
		if err := s.repos.Conversation.UpdateStatus(ctx, conv.ID, "escalated", userID); err != nil {
			s.logger.Error("Widget: failed to escalate conversation", "error", err, "conv_id", conv.ID)
		}
		
		// Send notification of escalation if preference is set
		prefs, err := s.repos.User.GetNotifPrefs(ctx, userID)
		if err == nil && prefs.Escalation {
			// Trigger immediate email notification via Resend!
			user, uerr := s.repos.User.GetByID(ctx, userID)
			if uerr == nil && user != nil {
				// Let's create a notification in DB first
				notif := &domain.Notification{
					UserID: userID,
					Type:   "escalation",
					Title:  "Chat Escalated to Human",
					Body:   fmt.Sprintf("A web chat with %s requires your attention.", conv.CustomerName),
					Link:   fmt.Sprintf("/chats?id=%s", conv.ID),
				}
				if err := s.repos.Notification.Create(ctx, notif); err != nil {
					s.logger.Error("Widget: failed to create escalation notification", "error", err)
				}
				
				// Send email if configured
			if s.email != nil {
				go func() {
					defer func() {
						if r := recover(); r != nil {
							s.logger.Error("Panic in escalation email goroutine", "error", r)
						}
					}()
					_, emailErr := s.email.SendNotificationEmail(
							context.Background(),
							user.Email,
							"NOANT Escalation Alert: Support agent needed",
							fmt.Sprintf("Hello %s,\n\nA customer chat on your web widget has been escalated to a human agent because the AI confidence was low. Please log in to your NOANT dashboard to take over the conversation.\n\nLink: %s/chats?id=%s", user.FirstName, s.cfg.AppURL, conv.ID),
						)
						if emailErr != nil {
							s.logger.Error("Failed to send escalation email", "error", emailErr)
						}
					}()
				}
			}
		}
	}

	return aiResult.Content, conv.ID, nil
}

func generateRandomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	bytes := make([]byte, n)
	if _, err := crypto_rand.Read(bytes); err != nil {
		return time.Now().Format("20060102150405") + "fallback"
	}
	for i, b := range bytes {
		bytes[i] = letters[b%byte(len(letters))]
	}
	return string(bytes)
}
