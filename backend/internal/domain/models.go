package domain

import (
	"time"
)

// User represents a platform user
type User struct {
	ID           string    `json:"id" db:"id"`
	Email        string    `json:"email" db:"email"`
	Password     string    `json:"-" db:"password_hash"`
	FirstName    string    `json:"first_name" db:"first_name"`
	LastName     string    `json:"last_name" db:"last_name"`
	Role         string    `json:"role" db:"role"` // owner, admin, agent
	CompanyName  string    `json:"company_name" db:"company_name"`
	Phone        string    `json:"phone" db:"phone"`
	Avatar       *string   `json:"avatar" db:"avatar"`
	PlanID       string    `json:"plan_id" db:"plan_id"`
	IsActive     bool      `json:"is_active" db:"is_active"`
	MustChangePassword bool `json:"must_change_password" db:"must_change_password"`
	TrialExpiresAt *time.Time `json:"trial_expires_at" db:"trial_expires_at"`
	LastLoginAt  *time.Time `json:"last_login_at" db:"last_login_at"`
	OwnerWhatsapp *string  `json:"owner_whatsapp,omitempty" db:"owner_whatsapp"`
	IsVerified   bool      `json:"is_verified" db:"is_verified"`
	VerificationCode *string `json:"-" db:"verification_code"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
}

// Conversation represents a chat session
type Conversation struct {
	ID              string    `json:"id" db:"id"`
	UserID          string    `json:"user_id" db:"user_id"`
	CustomerName    string    `json:"customer_name" db:"customer_name"`
	CustomerPhone   string    `json:"customer_phone" db:"customer_phone"`
	CustomerEmail   string    `json:"customer_email" db:"customer_email"`
	CustomerAvatar  string    `json:"customer_avatar" db:"customer_avatar"`
	Channel         string    `json:"channel" db:"channel"` // telegram, whatsapp, web, instagram
	Status          string    `json:"status" db:"status"`   // active, resolved, escalated, archived
	Intent          string    `json:"intent" db:"intent"`   // buying, complaining, inquiry, support
	Priority        string    `json:"priority" db:"priority"` // low, medium, high, urgent
	IsAITransferred bool      `json:"is_ai_transferred" db:"is_ai_transferred"`
	TakenOverBy     *string   `json:"taken_over_by" db:"taken_over_by"`
	TakenOverAt     *time.Time `json:"taken_over_at" db:"taken_over_at"`
	ResolvedAt      *time.Time `json:"resolved_at" db:"resolved_at"`
	FolderID        *string   `json:"folder_id" db:"folder_id"`
	Location        *Location `json:"location" db:"location"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
	LastMessage     string     `json:"last_message"`
	Unread          int        `json:"unread"`
}

type Location struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	City      string  `json:"city"`
	Country   string  `json:"country"`
}

// Message represents a single message in a conversation
type Message struct {
	ID             string    `json:"id" db:"id"`
	ConversationID string    `json:"conversation_id" db:"conversation_id"`
	Role           string    `json:"role" db:"sender_type"` // ai, human, customer, system
	SenderID       *string   `json:"sender_id" db:"sender_id"`
	Content        string    `json:"content" db:"content"`
	IsRead         bool      `json:"is_read" db:"is_read"`
	Confidence     float64   `json:"confidence,omitempty" db:"confidence"`
	Source         string    `json:"source,omitempty" db:"source"`
	Metadata       *MessageMetadata `json:"metadata" db:"metadata"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
}

type MessageMetadata struct {
	Confidence    float64 `json:"confidence,omitempty"`
	MatchedQAID   *string `json:"matched_qa_id,omitempty"`
	EscalationReason string `json:"escalation_reason,omitempty"`
	Language      string  `json:"language,omitempty"`
	Source        string  `json:"source,omitempty"`
}

// QAPair represents a question-answer pair for training
type QAPair struct {
    ID         string    `json:"id" db:"id"`
    UserID     string    `json:"user_id" db:"user_id"`
    CategoryID string    `json:"category_id" db:"category_id"`
    Category   string    `json:"category" db:"category"`
    Question   string    `json:"question" db:"question"`
    Answer     string    `json:"answer" db:"answer"`
    Variations []string  `json:"variations" db:"variations"`
    Embedding  []float32 `json:"-" db:"embedding"`
    IsActive   bool      `json:"is_active" db:"is_active"`
    UsageCount int       `json:"usage_count" db:"usage_count"`
    CreatedAt  time.Time `json:"created_at" db:"created_at"`
    UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}

// Category represents a Q&A category
type Category struct {
    ID          string    `json:"id" db:"id"`
    UserID      string    `json:"user_id" db:"user_id"`
    Name        string    `json:"name" db:"name"`
    Description string    `json:"description" db:"description"`
    Color       string    `json:"color" db:"color"`
    QACount     int       `json:"qa_count" db:"qa_count"`
    CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// UnknownQuestion represents a question the AI couldn't answer
type UnknownQuestion struct {
    ID             string    `json:"id" db:"id"`
    UserID         string    `json:"user_id" db:"user_id"`
    Question       string    `json:"question" db:"question"`
    ConversationID string    `json:"conversation_id" db:"conversation_id"`
    Channel        string    `json:"channel" db:"channel"`
    Status         string    `json:"status" db:"status"` // pending, trained, ignored
    SuggestedAnswer *string  `json:"suggested_answer" db:"suggested_answer"`
    CategoryID     *string   `json:"category_id" db:"category_id"`
    CreatedAt      time.Time `json:"created_at" db:"created_at"`
}

// Integration represents a connected channel
type Integration struct {
	ID        string    `json:"id" db:"id"`
	UserID    string    `json:"user_id" db:"user_id"`
	Channel   string    `json:"channel" db:"channel"`
	Status    string    `json:"status" db:"status"` // active, inactive, error
	Config    map[string]interface{} `json:"config" db:"config"`
	WebhookURL string   `json:"webhook_url" db:"webhook_url"`
	LastError *string   `json:"last_error" db:"last_error"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// AnalyticsOverview represents dashboard metrics
type AnalyticsOverview struct {
	TotalConversations   int     `json:"total_conversations"`
	ConversationsToday   int     `json:"conversations_today"`
	ActiveConversations  int     `json:"active_conversations"`
	UnreadConversations  int     `json:"unread_conversations"` // Open convos with no recent agent reply
	ResolvedToday        int     `json:"resolved_today"`
	AIResolutionRate     float64 `json:"ai_resolution_rate"`
	AvgResponseTime      float64 `json:"avg_response_time"`
	CustomerSatisfaction float64 `json:"customer_satisfaction"`
	Satisfaction         float64 `json:"satisfaction"`
	TotalMessages        int     `json:"total_messages"`
	EscalatedCount       int     `json:"escalated_count"`
	BillingAlert         bool    `json:"billing_alert"`  // True when plan is expiring/over limit
}

// TeamMember represents a team member
type TeamMember struct {
	ID        string    `json:"id" db:"id"`
	UserID    string    `json:"user_id" db:"user_id"`
	Email     string    `json:"email" db:"email"`
	FirstName string    `json:"first_name" db:"first_name"`
	LastName  string    `json:"last_name" db:"last_name"`
	Role      string    `json:"role" db:"role"`
	IsActive  bool      `json:"is_active" db:"is_active"`
	JoinedAt  time.Time `json:"joined_at" db:"joined_at"`
}

// APIKey represents an API key for integrations
type APIKey struct {
	ID        string    `json:"id" db:"id"`
	UserID    string    `json:"user_id" db:"user_id"`
	Name      string    `json:"name" db:"name"`
	Key       string    `json:"key" db:"key_hash"`
	LastUsed  *time.Time `json:"last_used" db:"last_used"`
	IsActive  bool      `json:"is_active" db:"is_active"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// ArchiveFolder represents an archive folder
type ArchiveFolder struct {
	ID        string    `json:"id" db:"id"`
	UserID    string    `json:"user_id" db:"user_id"`
	Name      string    `json:"name" db:"name"`
	Type      string    `json:"type" db:"type"` // chats, contacts, locations
	Color     string    `json:"color" db:"color"`
	ItemCount int       `json:"item_count" db:"item_count"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// PaymentPlan represents a subscription plan
type PaymentPlan struct {
	ID            string `json:"id" db:"id"`
	Name          string `json:"name" db:"name"`
	PriceNGN      int    `json:"price_ngn" db:"price_ngn"`
	AIResponses   int    `json:"ai_responses" db:"ai_responses"`
	Channels      []string `json:"channels" db:"channels"`
	Features      []string `json:"features" db:"features"`
	IsPopular     bool   `json:"is_popular" db:"is_popular"`
}

type Subscription struct {
	ID                 string    `json:"id" db:"id"`
	UserID             string    `json:"user_id" db:"user_id"`
	PlanID             string    `json:"plan_id" db:"plan_id"`
	Status             string    `json:"status" db:"status"`
	CurrentPeriodStart time.Time `json:"current_period_start" db:"current_period_start"`
	CurrentPeriodEnd   time.Time `json:"current_period_end" db:"current_period_end"`
	CreatedAt          time.Time `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time `json:"updated_at" db:"updated_at"`
}

type Notification struct {
	ID        string    `json:"id" db:"id"`
	UserID    string    `json:"user_id" db:"user_id"`
	Type      string    `json:"type" db:"type"`
	Title     string    `json:"title" db:"title"`
	Body      string    `json:"body" db:"body"`
	Link      string    `json:"link" db:"link"`
	IsRead    bool      `json:"is_read" db:"is_read"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type WidgetConfig struct {
	ID           string    `json:"id" db:"id"`
	UserID       string    `json:"user_id" db:"user_id"`
	BrandColor   string    `json:"brand_color" db:"brand_color"`
	Greeting     string    `json:"greeting" db:"greeting"`
	BotName      string    `json:"bot_name" db:"bot_name"`
	Position     string    `json:"position" db:"position"`
	WidgetAPIKey string    `json:"widget_api_key" db:"widget_api_key"`
	IsActive     bool      `json:"is_active" db:"is_active"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// InventoryItem represents a product, service, or package
type InventoryItem struct {
	ID            string    `json:"id" db:"id"`
	UserID        string    `json:"user_id" db:"user_id"`
	Type          string    `json:"type" db:"type"` // product, service, package
	Name          string    `json:"name" db:"name"`
	Description   string    `json:"description" db:"description"`
	Price         float64   `json:"price" db:"price"`
	MinPrice      *float64  `json:"min_price" db:"min_price"`
	StockQuantity *int      `json:"stock_quantity" db:"stock_quantity"`
	ImageURL      *string   `json:"image_url" db:"image_url"`
	IsActive      bool      `json:"is_active" db:"is_active"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}

// Handoff represents a sales handoff from AI to owner
type Handoff struct {
	ID               string     `json:"id" db:"id"`
	UserID           string     `json:"user_id" db:"user_id"`
	ConversationID   string     `json:"conversation_id" db:"conversation_id"`
	CustomerName     string     `json:"customer_name" db:"customer_name"`
	CustomerPhone    string     `json:"customer_phone" db:"customer_phone"`
	CustomerWhatsapp string     `json:"customer_whatsapp" db:"customer_whatsapp"`
	CustomerLocation string     `json:"customer_location" db:"customer_location"`
	ProductName      string     `json:"product_name" db:"product_name"`
	OriginalPrice    float64    `json:"original_price" db:"original_price"`
	AgreedPrice      float64    `json:"agreed_price" db:"agreed_price"`
	Quantity         int        `json:"quantity" db:"quantity"`
	Status           string     `json:"status" db:"status"` // pending, sold, lost, expired
	FinalPrice       *float64   `json:"final_price" db:"final_price"`
	OwnerNotes       string     `json:"owner_notes" db:"owner_notes"`
	OwnerNotifiedAt  *time.Time `json:"owner_notified_at" db:"owner_notified_at"`
	ReminderCount    int        `json:"reminder_count" db:"reminder_count"`
	NextReminderAt   *time.Time `json:"next_reminder_at" db:"next_reminder_at"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at" db:"updated_at"`
}

// UserCredit tracks Pulse response balance and expiry
type UserCredit struct {
	ID            string    `json:"id" db:"id"`
	UserID        string    `json:"user_id" db:"user_id"`
	Balance       int       `json:"balance" db:"balance"`
	ExpiresAt     *time.Time `json:"expires_at" db:"expires_at"`
	LastUpdatedAt time.Time `json:"last_updated_at" db:"last_updated_at"`
}

// CreditPurchase represents a credit purchase history record
type CreditPurchase struct {
	ID           string    `json:"id" db:"id"`
	UserID       string    `json:"user_id" db:"user_id"`
	CheckoutID   string    `json:"checkout_id" db:"checkout_id"`
	PackType     string    `json:"pack_type" db:"pack_type"` // small/medium/large
	Amount       int       `json:"amount" db:"amount"`
	Status       string    `json:"status" db:"status"` // pending/completed/failed/refunded
	PurchasedAt  time.Time `json:"purchased_at" db:"purchased_at"`
	ExpiresAt    time.Time `json:"expires_at" db:"expires_at"`
}

// MediaMessage represents a media file attached to a WhatsApp message
type MediaMessage struct {
	ID           string    `json:"id" db:"id"`
	UserID       string    `json:"user_id" db:"user_id"`
	ConversationID string  `json:"conversation_id" db:"conversation_id"`
	MessageID    string    `json:"message_id" db:"message_id"`
	SessionID    string    `json:"session_id" db:"session_id"`
	MediaType    string    `json:"media_type" db:"media_type"` // image, document, audio, video, sticker, location, contact
	MimeType     string    `json:"mime_type" db:"mime_type"`
	FileSize     int64     `json:"file_size" db:"file_size"`
	FileName     string    `json:"file_name" db:"file_name"`
	FilePath     string    `json:"file_path" db:"file_path"`
	ThumbPath    string    `json:"thumb_path" db:"thumb_path"`
	Width        int       `json:"width" db:"width"`
	Height       int       `json:"height" db:"height"`
	Duration     int       `json:"duration" db:"duration"` // seconds for audio/video
	Caption      string    `json:"caption" db:"caption"`
	RemoteURL    string    `json:"remote_url" db:"remote_url"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	ExpiresAt    time.Time `json:"expires_at" db:"expires_at"`
}

// WhatsAppTemplate represents a message template for WhatsApp
type WhatsAppTemplate struct {
	ID           string    `json:"id" db:"id"`
	UserID       string    `json:"user_id" db:"user_id"`
	Name         string    `json:"name" db:"name"`
	Language     string    `json:"language" db:"language"`
	Category     string    `json:"category" db:"category"` // marketing, utility, authentication
	Status       string    `json:"status" db:"status"`     // draft, pending, approved, rejected, disabled
	HeaderType   string    `json:"header_type" db:"header_type"` // text, image, video, document, none
	HeaderValue  string    `json:"header_value" db:"header_value"`
	BodyText     string    `json:"body_text" db:"body_text"` // with {{1}}, {{2}} placeholders
	FooterText   string    `json:"footer_text" db:"footer_text"`
	Buttons      string    `json:"buttons" db:"buttons"` // JSON array of button configs
	Namespace    string    `json:"namespace" db:"namespace"`
	RejectionReason string `json:"rejection_reason" db:"rejection_reason"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// QueueMessage represents a message in the outbound queue
type QueueMessage struct {
	ID         string    `json:"id" db:"id"`
	SessionID  string    `json:"session_id" db:"session_id"`
	UserID     string    `json:"user_id" db:"user_id"`
	ChatID     string    `json:"chat_id" db:"chat_id"`
	MsgType    string    `json:"msg_type" db:"msg_type"` // text, media, template
	Content    string    `json:"content" db:"content"`   // text body or JSON payload
	Priority   int       `json:"priority" db:"priority"` // 0=urgent, 1=normal, 2=bulk
	Status     string    `json:"status" db:"status"`     // queued, sending, sent, failed, dead_letter
	RetryCount int       `json:"retry_count" db:"retry_count"`
	MaxRetries int       `json:"max_retries" db:"max_retries"`
	LastError  string    `json:"last_error" db:"last_error"`
	ScheduledAt *time.Time `json:"scheduled_at" db:"scheduled_at"`
	SentAt     *time.Time `json:"sent_at" db:"sent_at"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}

// CampaignRecipient represents a contact in a campaign broadcast
type CampaignRecipient struct {
	ID             string    `json:"id" db:"id"`
	CampaignID     string    `json:"campaign_id" db:"campaign_id"`
	UserID         string    `json:"user_id" db:"user_id"`
	Phone          string    `json:"phone" db:"phone"`
	Name           string    `json:"name" db:"name"`
	Status         string    `json:"status" db:"status"` // pending, sent, delivered, read, failed, blocked, opted_out
	Error          string    `json:"error" db:"error"`
	SentAt         *time.Time `json:"sent_at" db:"sent_at"`
	DeliveredAt    *time.Time `json:"delivered_at" db:"delivered_at"`
	ReadAt         *time.Time `json:"read_at" db:"read_at"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
}

// CampaignSchedule represents a campaign mode schedule
type CampaignSchedule struct {
	ID          string    `json:"id" db:"id"`
	UserID      string    `json:"user_id" db:"user_id"`
	Name        string    `json:"name" db:"name"`
	StartDate   string    `json:"start_date" db:"start_date"` // stored as DATE, but handled as string
	EndDate     string    `json:"end_date" db:"end_date"`     // stored as DATE, but handled as string
	Status      string    `json:"status" db:"status"` // draft/active/completed/cancelled
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// PlanLimit defines limits per plan
type PlanLimit struct {
	PlanID               string `json:"plan_id" db:"plan_id"`
	MaxResponses         int    `json:"max_responses" db:"max_responses"`
	MaxHandoffs          int    `json:"max_handoffs" db:"max_handoffs"`
	MaxInventoryItems    int    `json:"max_inventory_items" db:"max_inventory_items"`
	HasNotification      bool   `json:"has_notification" db:"has_notification"`
	PriceNGN             int    `json:"price_ngn" db:"price_ngn"`
	Description          string `json:"description" db:"description"`
}
