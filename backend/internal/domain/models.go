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
	ActiveConversations  int     `json:"active_conversations"`
	ResolvedToday        int     `json:"resolved_today"`
	AIResolutionRate     float64 `json:"ai_resolution_rate"`
	AvgResponseTime      float64 `json:"avg_response_time"`
	CustomerSatisfaction float64 `json:"customer_satisfaction"`
	TotalMessages        int     `json:"total_messages"`
	EscalatedCount       int     `json:"escalated_count"`
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
