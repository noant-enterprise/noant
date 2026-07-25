// Package handler implements HTTP request handlers for the NOANT API.
// Each domain (auth, chat, training, etc.) has its own file with handler methods
// that validate input, call service-layer methods, and return JSON responses.
// Handlers use the gin.Context for request/response and are wired to routes in main.go.
package handler

import (
	"net/http"
	"time"

	"noant/config"
	"noant/internal/infrastructure"
	"noant/internal/repository"
	"noant/internal/service"

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
	Credit       *CreditHandler
	Campaign     *CampaignHandler
	DBManager    *DBManagerHandler
	Background   *BackgroundHandler
	Template     *TemplateHandler
	Assistant    *AssistantHandler
	Onboarding   *OnboardingHandler
	Push         *PushHandler
	Admin        *AdminHandler
}

func NewHandlers(cfg *config.Config, services *service.Services, repos *repository.Repositories, auditRepo *repository.AuditRepository, logger *infrastructure.Logger, wsHub *WebSocketHub) *Handlers {
	return &Handlers{
		Auth:         NewAuthHandler(services.Auth, auditRepo, logger),
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
		OpenWA:       NewOpenWAHandler(cfg, services.OpenWA, services.Chat, repos, logger, wsHub),
		Telegram:     NewTelegramHandler(services.Integration, logger),
		Credit:       NewCreditHandler(services.Credit, services.Plan, logger),
		Campaign:     NewCampaignHandler(services.Campaign, logger),
		DBManager:    NewDBManagerHandler(services.DBManager, logger),
		Background:   NewBackgroundHandler(services.Background, logger),
		Template:     NewTemplateHandler(services.Template, logger),
		Assistant:    NewAssistantHandler(services.Assistant, logger),
		Onboarding:   NewOnboardingHandler(services.Onboarding, logger),
		Push:         NewPushHandler(services.Push, logger),
		Admin:        NewAdminHandler(repos, logger),
	}
}

func (h *Handlers) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"version":   "2.0.0",
	})
}
