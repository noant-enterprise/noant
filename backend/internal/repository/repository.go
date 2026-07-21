// Package repository provides data access for NOANT's domain models.
// Each domain (user, conversation, message, etc.) has its own file with
// CRUD operations backed by TiDB/MySQL and Redis caching. The package
// also includes interface definitions and mock implementations for testing.
package repository

import (
	"database/sql"

	"noant/internal/infrastructure"
)

type Repositories struct {
	User                IUserRepo
	Conversation        IConversationRepo
	Message             IMessageRepo
	QAPair              IQAPairRepo
	Category            ICategoryRepo
	UnknownQ            IUnknownQuestionRepo
	Integration         IIntegrationRepo
	Team                ITeamRepo
	APIKey              IAPIKeyRepo
	Archive             IArchiveRepo
	Subscription        ISubscriptionRepo
	Audit               IAuditRepo
	Notification        INotificationRepo
	WidgetConfig        IWidgetConfigRepo
	Inventory           IInventoryRepo
	Handoff             IHandoffRepo
	Credit              ICreditRepo
	Campaign            ICampaignRepo
	WhatsAppTemplate    IWhatsAppTemplateRepo
	CampaignRecipient   ICampaignRecipientRepo
	MediaMessage        IMediaMessageRepo
	PushSubscription    IPushSubscriptionRepo
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
