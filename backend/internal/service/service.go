// Package service contains the business logic layer for NOANT.
// Services orchestrate domain operations, coordinate between repositories,
// and implement the AI Brain, chat processing, training, billing, and
// integration subsystems. Each domain has its own file; service.go
// aggregates them via the Services struct.
package service

import (
	"noant/config"
	"noant/internal/infrastructure"
	"noant/internal/repository"
)

type Services struct {
	Auth         *AuthService
	Chat         *ChatService
	Training     *TrainingService
	Analytics    *AnalyticsService
	Integration  *IntegrationService
	Settings     *SettingsService
	Archive      *ArchiveService
	Payment      *PaymentService
	Audit        *AuditService
	Notification *NotificationService
	Widget       *WidgetService
	Inventory    *InventoryService
	Handoff      *HandoffService
	OpenWA       *OpenWAService
	Telegram     *TelegramService
	Credit       *CreditService
	Plan         *PlanService
	Campaign     *CampaignService
	DBManager    *DBManagerService
	Background   *BackgroundWorker
	Template     *TemplateService
	Assistant    *AssistantService
	Onboarding   *OnboardingService
	Push         *PushNotificationService
}

func NewServices(cfg *config.Config, repos *repository.Repositories, redis *infrastructure.RedisClient, logger *infrastructure.Logger, email *EmailService, polarSvc *PolarService, broadcastFn func(convID string, msgType string, data interface{})) *Services {
	aiBrain := NewAIBrain(cfg, repos, redis, logger, broadcastFn)
	embeddings := aiBrain.embeddings

	telegramSvc := NewTelegramService(cfg, logger)
	openwaSvc := NewOpenWAService(cfg, logger)
	chatSvc := NewChatService(cfg, repos, redis, aiBrain, logger, openwaSvc, telegramSvc)
	creditSvc := NewCreditService(cfg, repos, redis, logger)
	planSvc := NewPlanService(cfg, repos, redis, logger, creditSvc)
	campaignSvc := NewCampaignService(cfg, repos, redis, logger, creditSvc)
	dbManagerSvc := NewDBManagerService(repos, logger)
	bgWorker := NewBackgroundWorker(logger, dbManagerSvc, 3)
	templateSvc := NewTemplateService(cfg, openwaSvc, redis, logger, repos)
	onboardingSvc := NewOnboardingService(cfg, repos, redis, logger)
	pushSvc := NewPushNotificationService(cfg, repos, logger)
	notifSvc := NewNotificationService(cfg, repos, redis, logger, email, pushSvc)
	return &Services{
		Auth:         NewAuthService(cfg, repos.User, repos.Org, redis, logger, email),
		Chat:         chatSvc,
		Training:     NewTrainingService(cfg, repos, redis, logger, embeddings),
		Analytics:    NewAnalyticsService(cfg, repos, redis, logger),
		Integration:  NewIntegrationService(cfg, repos, redis, logger, chatSvc, telegramSvc, broadcastFn),
		Settings:     NewSettingsService(cfg, repos, redis, logger, email),
		Archive:      NewArchiveService(cfg, repos, redis, logger),
		Payment:      NewPaymentService(cfg, repos, redis, logger, polarSvc, creditSvc),
		Audit:        NewAuditService(repos, logger),
		Notification: notifSvc,
		Widget:       NewWidgetService(cfg, repos, redis, aiBrain, logger, email),
		Inventory:    NewInventoryService(cfg, repos, redis, logger, embeddings),
		Handoff:      NewHandoffService(cfg, repos, redis, logger, broadcastFn, planSvc),
		OpenWA:       openwaSvc,
		Telegram:     telegramSvc,
		Credit:       creditSvc,
		Plan:         planSvc,
		Campaign:     campaignSvc,
		DBManager:    dbManagerSvc,
		Background:   bgWorker,
		Template:     templateSvc,
		Assistant:    NewAssistantService(aiBrain, logger),
		Onboarding:   onboardingSvc,
		Push:         pushSvc,
	}
}
