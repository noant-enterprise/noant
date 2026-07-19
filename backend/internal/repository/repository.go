package repository

import (
	"database/sql"

	"noant/internal/infrastructure"
)

type Repositories struct {
	User                *UserRepository
	Conversation        *ConversationRepository
	Message             *MessageRepository
	QAPair              *QAPairRepository
	Category            *CategoryRepository
	UnknownQ            *UnknownQuestionRepository
	Integration         *IntegrationRepository
	Team                *TeamRepository
	APIKey              *APIKeyRepository
	Archive             *ArchiveRepository
	Subscription        *SubscriptionRepository
	Audit               *AuditRepository
	Notification        *NotificationRepository
	WidgetConfig        *WidgetConfigRepository
	Inventory           *InventoryRepository
	Handoff             *HandoffRepository
	Credit              *CreditRepository
	Campaign            *CampaignRepository
	WhatsAppTemplate    *WhatsAppTemplateRepository
	CampaignRecipient   *CampaignRecipientRepository
	MediaMessage        *MediaMessageRepository
	PushSubscription    *PushSubscriptionRepository
}

func NewRepositories(db *sql.DB, redis *infrastructure.RedisClient) *Repositories {
	return &Repositories{
		User:                NewUserRepository(db, redis),
		Conversation:        NewConversationRepository(db, redis),
		Message:             NewMessageRepository(db, redis),
		QAPair:              NewQAPairRepository(db, redis),
		Category:            NewCategoryRepository(db, redis),
		UnknownQ:            NewUnknownQuestionRepository(db, redis),
		Integration:         NewIntegrationRepository(db, redis),
		Team:                NewTeamRepository(db, redis),
		APIKey:              NewAPIKeyRepository(db, redis),
		Archive:             NewArchiveRepository(db, redis),
		Subscription:        NewSubscriptionRepository(db, redis),
		Audit:               NewAuditRepository(db, redis),
		Notification:        NewNotificationRepository(db, redis),
		WidgetConfig:        NewWidgetConfigRepository(db, redis),
		Inventory:           NewInventoryRepository(db, redis),
		Handoff:             NewHandoffRepository(db, redis),
		Credit:              NewCreditRepository(db, redis),
		Campaign:            NewCampaignRepository(db, redis),
		WhatsAppTemplate:    NewWhatsAppTemplateRepository(db, redis),
		CampaignRecipient:   NewCampaignRecipientRepository(db, redis),
		MediaMessage:        NewMediaMessageRepository(db, redis),
		PushSubscription:    NewPushSubscriptionRepository(db, redis),
	}
}
