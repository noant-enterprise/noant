package repository

import (
	"context"
	"database/sql"

	"noant/internal/domain"
)

type IOrgRepo interface {
	Create(ctx context.Context, org *domain.Organization) error
	GetByID(ctx context.Context, id string) (*domain.Organization, error)
	GetByOwnerID(ctx context.Context, ownerID string) (*domain.Organization, error)
	Update(ctx context.Context, org *domain.Organization) error
}

type IUserRepo interface {
	Create(ctx context.Context, user *domain.User) error
	RunInTx(ctx context.Context, fn func(tx *sql.Tx) error) error
	CreateTx(ctx context.Context, tx *sql.Tx, user *domain.User) error
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	GetByID(ctx context.Context, id string) (*domain.User, error)
	UpdateLastLogin(ctx context.Context, id string) error
	UpdatePassword(ctx context.Context, id, hashedPassword string) error
	UpdatePlan(ctx context.Context, userID, planID string) error
	UpdateVerificationStatus(ctx context.Context, id string, verified bool) error
	UpdateVerificationCode(ctx context.Context, id, code string) error
	GetOwnerWhatsApp(ctx context.Context, userID string) (string, error)
	CleanupExpiredTrials(ctx context.Context, days int) (int64, error)
	GetNotifPrefs(ctx context.Context, userID string) (*NotifPrefs, error)
	UpdateNotifPrefs(ctx context.Context, userID string, prefs *NotifPrefs) error
	Delete(ctx context.Context, userID string) error
	ExportUserData(ctx context.Context, userID string) (map[string]interface{}, error)
	UpdateProfile(ctx context.Context, userID, firstName, lastName, companyName, phone string) error
	GetOnboardingStatus(ctx context.Context, userID string) (*string, error)
	UpdateOnboardingStatus(ctx context.Context, userID, status string, industry *string) error
}

type IConversationRepo interface {
	GetByID(ctx context.Context, id string) (*domain.Conversation, error)
	Create(ctx context.Context, conv *domain.Conversation) error
	List(ctx context.Context, userID, status string, limit, offset int) ([]domain.Conversation, int, error)
	GetByIDAndUser(ctx context.Context, id, userID string) (*domain.Conversation, error)
	UpdateStatus(ctx context.Context, id, userID, status string) error
	UpdateCustomerInfo(ctx context.Context, id, name, avatar string) error
	FindActiveByCustomer(ctx context.Context, userID, customerName, channel string) (*domain.Conversation, error)
	Takeover(ctx context.Context, id, userID, agentID string) error
	ClearChats(ctx context.Context, userID string) error
	GetOverview(ctx context.Context, userID string) (map[string]interface{}, error)
	CountByChannel(ctx context.Context, userID string) (map[string]int, error)
	CountByIntent(ctx context.Context, userID string) ([]map[string]interface{}, error)
	CountByHour(ctx context.Context, userID string) ([]map[string]interface{}, error)
	CountByDate(ctx context.Context, userID string, days int) ([]map[string]interface{}, error)
	RecordCSAT(ctx context.Context, userID, conversationID string, score int, comment *string) error
	GetCSATAverage(ctx context.Context, userID string) (avg float64, total int, err error)
	GetCSATDistribution(ctx context.Context, userID string) (map[int]int, error)
	GetCSATTrend(ctx context.Context, userID string, days int) ([]map[string]interface{}, error)
	CountMessagesByDate(ctx context.Context, userID string, days int) ([]map[string]interface{}, error)
	GetUptimeStats(ctx context.Context, userID string) (int, error)
	CleanupOldResolved(ctx context.Context, days int) (int64, error)
	CleanupAbandoned(ctx context.Context, days int) (int64, error)
}

type IMessageRepo interface {
	Create(ctx context.Context, msg *domain.Message) error
	ListByConversation(ctx context.Context, conversationID string, limit int) ([]domain.Message, error)
	ListByConversationPaginated(ctx context.Context, conversationID string, limit, offset int) ([]domain.Message, int, error)
	GetLastMessage(ctx context.Context, conversationID string) (*domain.Message, error)
	CountUnread(ctx context.Context, conversationID string) (int, error)
	MarkRead(ctx context.Context, conversationID string) error
	CleanupOrphaned(ctx context.Context) (int64, error)
}

type IQAPairRepo interface {
	Create(ctx context.Context, qa *domain.QAPair) error
	BulkCreate(ctx context.Context, qas []domain.QAPair) error
	ListByCategory(ctx context.Context, categoryID string) ([]domain.QAPair, error)
	ListByCategoryAndUser(ctx context.Context, categoryID, userID string) ([]domain.QAPair, error)
	Search(ctx context.Context, userID, query string) ([]domain.QAPair, error)
	ListByUser(ctx context.Context, userID, categoryID string) ([]domain.QAPair, error)
	GetByID(ctx context.Context, id string) (*domain.QAPair, error)
	GetByQuestion(ctx context.Context, userID, question string) (*domain.QAPair, error)
	Update(ctx context.Context, qa *domain.QAPair) error
	IncrementUsage(ctx context.Context, id string) error
	CountByUser(ctx context.Context, userID string) (int, error)
	Delete(ctx context.Context, id, userID string) error
}

type ICategoryRepo interface {
	GetByName(ctx context.Context, userID, name string) (*domain.Category, error)
	Create(ctx context.Context, cat *domain.Category) error
	List(ctx context.Context, userID string) ([]domain.Category, error)
	Delete(ctx context.Context, id, userID string) error
}

type IIntegrationRepo interface {
	Create(ctx context.Context, integration *domain.Integration) error
	ListByUser(ctx context.Context, userID string) ([]domain.Integration, error)
	ListActive(ctx context.Context) ([]domain.Integration, error)
	ListByChannel(ctx context.Context, channel string) ([]domain.Integration, error)
	UpdateStatus(ctx context.Context, id, status string, lastError *string) error
	GetByUserAndChannel(ctx context.Context, userID, channel string) (*domain.Integration, error)
	GetByChannelAndSessionID(ctx context.Context, channel, sessionID string) (*domain.Integration, error)
	GetByChannelAndWebhookSecret(ctx context.Context, channel, secret string) (*domain.Integration, error)
	Update(ctx context.Context, integration *domain.Integration) error
	Disconnect(ctx context.Context, userID, channel string) error
	CleanupStaleInactive(ctx context.Context, days int) (int64, error)
}

type IHandoffRepo interface {
	Create(ctx context.Context, h *domain.Handoff) error
	GetByID(ctx context.Context, id, userID string) (*domain.Handoff, error)
	List(ctx context.Context, userID, status string, limit int) ([]domain.Handoff, error)
	UpdateStatus(ctx context.Context, id, userID, status, notes string) error
	GetPending(ctx context.Context, userID string) ([]domain.Handoff, error)
	GetReadyForReminder(ctx context.Context) ([]domain.Handoff, error)
	IncrementReminder(ctx context.Context, id string) error
	Expire(ctx context.Context, id string) error
	CleanupExpired(ctx context.Context, days int) (int64, error)
}

type INotificationRepo interface {
	Create(ctx context.Context, n *domain.Notification) error
	ListByUser(ctx context.Context, userID string, limit int) ([]*domain.Notification, error)
	UnreadCount(ctx context.Context, userID string) (int, error)
	MarkRead(ctx context.Context, id, userID string) error
	MarkAllRead(ctx context.Context, userID string) error
	CleanupOld(ctx context.Context, days int) (int64, error)
}

type IWidgetConfigRepo interface {
	Get(ctx context.Context, userID string) (*domain.WidgetConfig, error)
	GetByAPIKey(ctx context.Context, apiKey string) (*domain.WidgetConfig, error)
	Upsert(ctx context.Context, cfg *domain.WidgetConfig) error
}

type IInventoryRepo interface {
	Create(ctx context.Context, item *domain.InventoryItem) error
	GetByID(ctx context.Context, id, userID string) (*domain.InventoryItem, error)
	List(ctx context.Context, userID, itemType string, activeOnly bool) ([]domain.InventoryItem, error)
	Search(ctx context.Context, userID, q string) ([]domain.InventoryItem, error)
	Update(ctx context.Context, item *domain.InventoryItem) error
	Delete(ctx context.Context, id, userID string) error
	DecreaseStock(ctx context.Context, itemID string, quantity int) error
	CountByUser(ctx context.Context, userID string) (int, error)
}

type IArchiveRepo interface {
	CreateFolder(ctx context.Context, folder *domain.ArchiveFolder) error
	ListFolders(ctx context.Context, userID, folderType string) ([]domain.ArchiveFolder, error)
	MoveChat(ctx context.Context, conversationID, userID, folderID string) error
}

type IUnknownQuestionRepo interface {
	Create(ctx context.Context, uq *domain.UnknownQuestion) error
	GetByIDAndUser(ctx context.Context, id, userID string) (*domain.UnknownQuestion, error)
	List(ctx context.Context, userID, status string, limit, offset int) ([]domain.UnknownQuestion, error)
	BatchTrain(ctx context.Context, userID, answer, categoryID string, ids []string) error
	BatchIgnore(ctx context.Context, userID string, ids []string) error
	ExistsPending(ctx context.Context, userID, question string) (bool, error)
	UpdateStatus(ctx context.Context, id, userID, status string, answer, categoryID *string) error
	Clear(ctx context.Context, userID string) error
	CountByStatus(ctx context.Context, userID string) (map[string]int, error)
	MostPopular(ctx context.Context, userID string, limit int) ([]map[string]interface{}, error)
	CountByFilter(ctx context.Context, userID, status string) (int, error)
	CountByDate(ctx context.Context, userID string, days int) ([]map[string]interface{}, error)
	CleanupStale(ctx context.Context, days int) (int64, error)
}

type ICampaignRepo interface {
	Create(ctx context.Context, campaign *domain.CampaignSchedule) error
	ListByUser(ctx context.Context, userID string) ([]domain.CampaignSchedule, error)
	GetScheduledForToday(ctx context.Context) ([]domain.CampaignSchedule, error)
	GetEndingToday(ctx context.Context) ([]domain.CampaignSchedule, error)
	UpdateStatus(ctx context.Context, id, status string) error
	CleanupCompleted(ctx context.Context, days int) (int64, error)
}

type IAPIKeyRepo interface {
	Create(ctx context.Context, key *domain.APIKey) error
	ListByUser(ctx context.Context, userID string) ([]domain.APIKey, error)
	Revoke(ctx context.Context, id, userID string) error
}

type ICreditRepo interface {
	GetByUserID(ctx context.Context, userID string) (*domain.UserCredit, error)
	Upsert(ctx context.Context, credit *domain.UserCredit) error
	Deduct(ctx context.Context, userID string, amount int) error
	GetExpiring(ctx context.Context, days int) ([]domain.UserCredit, error)
	CreatePurchase(ctx context.Context, p *domain.CreditPurchase) error
	GetPurchaseHistory(ctx context.Context, userID string) ([]domain.CreditPurchase, error)
	CleanupExpired(ctx context.Context) (int64, error)
	CleanupStalePurchases(ctx context.Context, days int) (int64, error)
}

type ITeamRepo interface {
	ListByOrg(ctx context.Context, orgID string) ([]domain.TeamMember, error)
	Create(ctx context.Context, orgID string, member *domain.TeamMember) error
	Delete(ctx context.Context, id string) error
	GetByID(ctx context.Context, id string) (*domain.TeamMember, error)
	GetByEmailAndOrg(ctx context.Context, email, orgID string) (*domain.TeamMember, error)
}

type IAuditRepo interface {
	Create(ctx context.Context, log *domain.AuditLog) error
	ListByUser(ctx context.Context, userID string, limit int) ([]domain.AuditLog, error)
	ListWithFilters(ctx context.Context, filter *AuditFilter) (*AuditListResult, error)
	CleanupOld(ctx context.Context, days int) (int64, error)
}

type IPushSubscriptionRepo interface {
	Create(ctx context.Context, sub *domain.PushSubscription) error
	Delete(ctx context.Context, userID, endpoint string) error
	DeleteAllByUser(ctx context.Context, userID string) error
	ListByUser(ctx context.Context, userID string) ([]*domain.PushSubscription, error)
	ListByUserIDs(ctx context.Context, userIDs []string) ([]*domain.PushSubscription, error)
	DeleteByID(ctx context.Context, id string) error
}

type IWhatsAppTemplateRepo interface {
	Create(ctx context.Context, tpl *domain.WhatsAppTemplate) error
	ListByUser(ctx context.Context, userID string) ([]domain.WhatsAppTemplate, error)
	GetByID(ctx context.Context, id, userID string) (*domain.WhatsAppTemplate, error)
	Update(ctx context.Context, tpl *domain.WhatsAppTemplate) error
	Delete(ctx context.Context, id, userID string) error
	GetByStatus(ctx context.Context, status string) ([]domain.WhatsAppTemplate, error)
}

type IMediaMessageRepo interface {
	Create(ctx context.Context, m *domain.MediaMessage) error
	GetByConversation(ctx context.Context, conversationID string) ([]domain.MediaMessage, error)
	CleanupExpired(ctx context.Context) (int64, error)
}

type ICampaignRecipientRepo interface {
	Create(ctx context.Context, cr *domain.CampaignRecipient) error
	ListByCampaign(ctx context.Context, campaignID string) ([]domain.CampaignRecipient, error)
	UpdateStatus(ctx context.Context, id, status string, errInfo *string) error
	MarkOptedOut(ctx context.Context, userID, phone string) error
	IsOptedOut(ctx context.Context, userID, phone string) (bool, error)
}

type ISubscriptionRepo interface {
	GetActive(ctx context.Context, userID string) (*domain.Subscription, error)
	Create(ctx context.Context, sub *domain.Subscription) error
	CreateOrUpdate(ctx context.Context, sub *domain.Subscription) error
	Cancel(ctx context.Context, userID string) error
}

var _ IUserRepo = (*UserRepository)(nil)
var _ IConversationRepo = (*ConversationRepository)(nil)
var _ IMessageRepo = (*MessageRepository)(nil)
var _ IQAPairRepo = (*QAPairRepository)(nil)
var _ ICategoryRepo = (*CategoryRepository)(nil)
var _ IIntegrationRepo = (*IntegrationRepository)(nil)
var _ IHandoffRepo = (*HandoffRepository)(nil)
var _ INotificationRepo = (*NotificationRepository)(nil)
var _ IWidgetConfigRepo = (*WidgetConfigRepository)(nil)
var _ IInventoryRepo = (*InventoryRepository)(nil)
var _ IArchiveRepo = (*ArchiveRepository)(nil)
var _ IUnknownQuestionRepo = (*UnknownQuestionRepository)(nil)
var _ ICampaignRepo = (*CampaignRepository)(nil)
var _ IAPIKeyRepo = (*APIKeyRepository)(nil)
var _ ICreditRepo = (*CreditRepository)(nil)
var _ ITeamRepo = (*TeamRepository)(nil)
var _ IAuditRepo = (*AuditRepository)(nil)
var _ IPushSubscriptionRepo = (*PushSubscriptionRepository)(nil)
var _ IWhatsAppTemplateRepo = (*WhatsAppTemplateRepository)(nil)
var _ IMediaMessageRepo = (*MediaMessageRepository)(nil)
var _ ICampaignRecipientRepo = (*CampaignRecipientRepository)(nil)
var _ ISubscriptionRepo = (*SubscriptionRepository)(nil)
