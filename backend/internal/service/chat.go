package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"noant/config"
	"noant/internal/domain"
	"noant/internal/infrastructure"
	"noant/internal/repository"
	"noant/internal/utils"
)

type ChatService struct {
	cfg      *config.Config
	repos    *repository.Repositories
	redis    *infrastructure.RedisClient
	aiBrain  *AIBrain
	logger   *infrastructure.Logger
	openwa   *OpenWAService
	telegram *TelegramService
	replyMu  sync.Mutex
	replies  map[string]*replyGateState
}

type replyGateState struct {
	lastKey     string
	inFlightKey string
	inFlightAt  time.Time
	lastReplyAt time.Time
}

// NewChatService creates a ChatService that manages conversations, messages,
// and real-time WebSocket broadcasting. The aiBrain parameter handles AI response
// generation; openwa and telegram handle outbound channel delivery.
func NewChatService(cfg *config.Config, repos *repository.Repositories, redis *infrastructure.RedisClient, aiBrain *AIBrain, logger *infrastructure.Logger, openwa *OpenWAService, telegram *TelegramService) *ChatService {
	return &ChatService{
		cfg:      cfg,
		repos:    repos,
		redis:    redis,
		aiBrain:  aiBrain,
		logger:   logger,
		openwa:   openwa,
		telegram: telegram,
		replies:  make(map[string]*replyGateState),
	}
}

func normalizeReplyKey(message string) string {
	message = strings.ToLower(strings.TrimSpace(message))
	if message == "" {
		return ""
	}
	return strings.Join(strings.Fields(message), " ")
}

func (s *ChatService) beginAIReply(conversationID, message string) bool {
	key := normalizeReplyKey(message)
	if key == "" {
		return true
	}

	s.replyMu.Lock()
	defer s.replyMu.Unlock()

	state, ok := s.replies[conversationID]
	if !ok {
		state = &replyGateState{}
		s.replies[conversationID] = state
	}

	now := time.Now()
	const cooldown = 5 * time.Second

	if state.inFlightKey == key && now.Sub(state.inFlightAt) < cooldown {
		return false
	}
	if state.lastKey == key && now.Sub(state.lastReplyAt) < cooldown {
		return false
	}

	state.inFlightKey = key
	state.inFlightAt = now
	state.lastKey = key
	return true
}

func (s *ChatService) completeAIReply(conversationID, message string) {
	key := normalizeReplyKey(message)
	s.replyMu.Lock()
	defer s.replyMu.Unlock()

	state, ok := s.replies[conversationID]
	if !ok {
		state = &replyGateState{}
		s.replies[conversationID] = state
	}

	state.inFlightKey = ""
	state.lastKey = key
	state.lastReplyAt = time.Now()
}

func (s *ChatService) abortAIReply(conversationID string) {
	s.replyMu.Lock()
	defer s.replyMu.Unlock()

	if state, ok := s.replies[conversationID]; ok {
		state.inFlightKey = ""
	}
}

func (s *ChatService) DirectChat(ctx context.Context, userID, customerName, customerKey, message, channel, customerAvatar string) (*domain.Conversation, *domain.Message, error) {
	customerName = utils.SanitizeName(customerName)
	customerKey = utils.SanitizeName(customerKey)
	message = utils.SanitizeXSS(message)
	channel = utils.SanitizeName(channel)
	customerAvatar = utils.SanitizeXSS(customerAvatar)

	if s.redis != nil {
		limit := 500
		user, err := s.repos.User.GetByID(ctx, userID)
		if err == nil && user != nil {
			switch user.PlanID {
			case "pulse":
				limit = 500
			case "pro", "business", "enterprise":
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
	if strings.TrimSpace(customerKey) == "" {
		customerKey = customerName
	}

	// If channel is whatsapp and we don't have a valid pushname/avatar yet,
	// query OpenWA dynamically to resolve the real contact profile details.
	if channel == "whatsapp" && s.openwa != nil && (customerName == "" || customerName == customerKey || customerAvatar == "") {
		integration, err := s.repos.Integration.GetByUserAndChannel(ctx, userID, "whatsapp")
		if err == nil && integration != nil {
			if sessionID, _ := integration.Config["session_id"].(string); sessionID != "" {
				contactID := FormatContactID(customerKey) // use @c.us for contacts API
				contact, err := s.openwa.GetContactInfo(sessionID, contactID)
				if err == nil && contact != nil {
					if contact.Pushname != "" {
						customerName = contact.Pushname
					} else if contact.Name != "" {
						customerName = contact.Name
					}
					if contact.ProfilePicUrl != "" {
						customerAvatar = contact.ProfilePicUrl
					}
				}
			}
		}
	}

	existing, _ := s.repos.Conversation.FindActiveByCustomer(ctx, userID, customerKey, channel)
	var conv *domain.Conversation
	if existing != nil {
		conv = existing
		needsUpdate := false
		if customerName != "" && customerName != customerKey && conv.CustomerName != customerName {
			conv.CustomerName = customerName
			needsUpdate = true
		}
		if customerAvatar != "" && conv.CustomerAvatar != customerAvatar {
			conv.CustomerAvatar = customerAvatar
			needsUpdate = true
		}
		if needsUpdate {
			_ = s.repos.Conversation.UpdateCustomerInfo(ctx, conv.ID, conv.CustomerName, conv.CustomerAvatar)
		}
	} else {
		conv = &domain.Conversation{
			UserID:         userID,
			CustomerName:   customerName,
			CustomerPhone:  customerKey,
			CustomerAvatar: customerAvatar,
			Channel:        channel,
			Status:         "active",
			Intent:         "inquiry",
			Priority:       "medium",
		}
		if err := s.repos.Conversation.Create(ctx, conv); err != nil {
			return nil, nil, err
		}
	}
	if !s.beginAIReply(conv.ID, message) {
		s.logger.Info("Skipping duplicate AI reply", "conversationID", conv.ID, "channel", channel)
		return conv, nil, nil
	}
	defer s.abortAIReply(conv.ID)

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
		Role:           "customer",
		Content:        message,
		IsRead:         false,
	}
	if err := s.repos.Message.Create(ctx, customerMsg); err != nil {
		s.logger.Error("Failed to save customer message", "error", err, "conv_id", conv.ID)
	}
	aiMsg := &domain.Message{
		ConversationID: conv.ID,
		Role:           "ai",
		Content:        aiResp.Content,
		IsRead:         false,
		Metadata: &domain.MessageMetadata{
			Confidence: aiResp.Confidence,
			Language:   "en",
		},
	}
	if err := s.repos.Message.Create(ctx, aiMsg); err != nil {
		s.logger.Error("Failed to save AI message", "error", err, "conv_id", conv.ID)
	}
	if aiResp.Escalate {
		if err := s.repos.Conversation.UpdateStatus(ctx, conv.ID, "escalated", userID); err != nil {
			s.logger.Error("Failed to escalate conversation", "error", err, "conv_id", conv.ID)
		}
	}
	s.completeAIReply(conv.ID, message)
	return conv, aiMsg, nil
}

type WhatsAppIdentity struct {
	Name    string
	Phone   string
	Avatar  string
	Methods []string
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func cleanWhatsAppID(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	raw = strings.TrimPrefix(raw, "waid:")
	return CleanPhoneNumber(raw)
}

func (s *ChatService) ResolveWhatsAppIdentity(ctx context.Context, userID, sessionID string, msg *OpenWAMessageData) (*WhatsAppIdentity, error) {
	if msg == nil {
		return nil, fmt.Errorf("message is required")
	}

	identity := &WhatsAppIdentity{
		Phone: cleanWhatsAppID(msg.From),
	}

	// Method 1: direct sender payload fields from the webhook.
	identity.Methods = append(identity.Methods, "sender_payload")
	identity.Name = firstNonEmpty(msg.Sender.Pushname, msg.Sender.Name, msg.Sender.FormattedName, msg.Sender.ShortName)
	identity.Avatar = firstNonEmpty(msg.Sender.ProfilePicThumbObj.Eurl)

	// Method 2: sender ID fallback from the payload.
	senderID := cleanWhatsAppID(msg.Sender.ID)
	if senderID != "" {
		identity.Methods = append(identity.Methods, "sender_id")
		if identity.Phone == "" {
			identity.Phone = senderID
		}
		if identity.Name == "" {
			identity.Name = senderID
		}
	}

	// Method 3: use the raw WhatsApp chat ID as the phone number fallback.
	if identity.Phone == "" {
		identity.Methods = append(identity.Methods, "chat_id")
		identity.Phone = cleanWhatsAppID(msg.From)
	}

	// Method 4: OpenWA contacts API using the chat ID.
	if s.openwa != nil && sessionID != "" && identity.Phone != "" {
		if contact, err := s.openwa.GetContactInfo(sessionID, FormatContactID(identity.Phone)); err == nil && contact != nil {
			identity.Methods = append(identity.Methods, "contact_lookup_from")
			identity.Name = firstNonEmpty(identity.Name, contact.Pushname, contact.Name)
			identity.Avatar = firstNonEmpty(identity.Avatar, contact.ProfilePicUrl)
			if identity.Phone == "" {
				identity.Phone = cleanWhatsAppID(contact.Number)
			}
		}
	}

	// Method 5: OpenWA contacts API using the sender ID if it differs.
	if s.openwa != nil && sessionID != "" && senderID != "" && senderID != identity.Phone {
		if contact, err := s.openwa.GetContactInfo(sessionID, FormatContactID(senderID)); err == nil && contact != nil {
			identity.Methods = append(identity.Methods, "contact_lookup_sender")
			identity.Name = firstNonEmpty(identity.Name, contact.Pushname, contact.Name)
			identity.Avatar = firstNonEmpty(identity.Avatar, contact.ProfilePicUrl)
		}
	}

	// Method 6: existing conversation fallback, in case this is a returning customer.
	if identity.Phone != "" && s.repos != nil {
		if existing, err := s.repos.Conversation.FindActiveByCustomer(ctx, userID, identity.Phone, "whatsapp"); err == nil && existing != nil {
			identity.Methods = append(identity.Methods, "existing_conversation")
			identity.Name = firstNonEmpty(identity.Name, existing.CustomerName)
			identity.Avatar = firstNonEmpty(identity.Avatar, existing.CustomerAvatar)
		}
	}

	if identity.Name == "" {
		identity.Name = identity.Phone
	}
	if identity.Name == "" {
		identity.Name = "WhatsApp User"
	}

	return identity, nil
}

func (s *ChatService) ListConversations(ctx context.Context, userID, status string, page, limit int) ([]domain.Conversation, int, error) {
	offset := (page - 1) * limit
	conversations, total, err := s.repos.Conversation.List(ctx, userID, status, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	for i := range conversations {
		// Populate last message
		lastMsg, err := s.repos.Message.GetLastMessage(ctx, conversations[i].ID)
		if err == nil && lastMsg != nil {
			conversations[i].LastMessage = lastMsg.Content
		}

		// Populate unread count
		unreadCount, err := s.repos.Message.CountUnread(ctx, conversations[i].ID)
		if err == nil {
			conversations[i].Unread = unreadCount
		}
	}

	// Synchronously resolve WhatsApp contact names/avatars for conversations
	// where the customer_name still looks like a raw phone number.
	// Done in parallel goroutines, we wait for all to finish before returning
	// so the UI always receives real names on every page load.
	if s.openwa != nil {
		integration, intErr := s.repos.Integration.GetByUserAndChannel(ctx, userID, "whatsapp")
		if intErr == nil && integration != nil {
			if sessionID, _ := integration.Config["session_id"].(string); sessionID != "" {
				var wg sync.WaitGroup
				var mu sync.Mutex
				for i := range conversations {
					conv := &conversations[i]
					needsResolve := conv.Channel == "whatsapp" &&
						(conv.CustomerName == "" || conv.CustomerName == conv.CustomerPhone || isAllDigits(conv.CustomerName))
					if !needsResolve {
						continue
					}
					wg.Add(1)
					go func(idx int, convID, phone string) {
						defer wg.Done()
						contactID := FormatContactID(phone)
						contact, err := s.openwa.GetContactInfo(sessionID, contactID)
						if err != nil || contact == nil {
							return
						}
						name := phone
						if contact.Pushname != "" {
							name = contact.Pushname
						} else if contact.Name != "" {
							name = contact.Name
						}
						avatar := contact.ProfilePicUrl
						if name == phone && avatar == "" {
							return // nothing changed
						}
						mu.Lock()
						conversations[idx].CustomerName = name
						if avatar != "" {
							conversations[idx].CustomerAvatar = avatar
						}
						mu.Unlock()
						// Persist update to DB in background (non-blocking)
						go func() {
							_ = s.repos.Conversation.UpdateCustomerInfo(context.Background(), convID, name, avatar)
						}()
					}(i, conv.ID, conv.CustomerPhone)
				}
				wg.Wait()
			}
		}
	}

	return conversations, total, nil
}

// isAllDigits returns true if s contains only digit characters (i.e. looks like a phone number)
func isAllDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

func (s *ChatService) GetConversation(ctx context.Context, userID, conversationID string) (*domain.Conversation, []domain.Message, error) {
	conv, err := s.repos.Conversation.GetByIDAndUser(ctx, conversationID, userID)
	if err != nil {
		return nil, nil, err
	}
	if conv == nil {
		return nil, nil, fmt.Errorf("conversation not found")
	}

	// Mark messages as read
	_ = s.repos.Message.MarkRead(ctx, conversationID)

	messages, err := s.repos.Message.ListByConversation(ctx, conversationID, 100)
	if err != nil {
		return nil, nil, err
	}
	return conv, messages, nil
}

func (s *ChatService) GetConversationOnly(ctx context.Context, conversationID, userID string) (*domain.Conversation, error) {
	return s.repos.Conversation.GetByIDAndUser(ctx, conversationID, userID)
}

func (s *ChatService) GetConversationPaginated(ctx context.Context, userID, conversationID string, limit, offset int) (*domain.Conversation, []domain.Message, int, error) {
	conv, err := s.repos.Conversation.GetByIDAndUser(ctx, conversationID, userID)
	if err != nil {
		return nil, nil, 0, err
	}
	if conv == nil {
		return nil, nil, 0, fmt.Errorf("conversation not found")
	}

	// Mark messages as read
	_ = s.repos.Message.MarkRead(ctx, conversationID)

	messages, total, err := s.repos.Message.ListByConversationPaginated(ctx, conversationID, limit, offset)
	if err != nil {
		return nil, nil, 0, err
	}
	return conv, messages, total, nil
}

func (s *ChatService) HumanTakeover(ctx context.Context, userID, conversationID, agentID string) error {
	conv, err := s.repos.Conversation.GetByIDAndUser(ctx, conversationID, userID)
	if err != nil {
		return err
	}
	if conv == nil {
		return fmt.Errorf("conversation not found")
	}
	return s.repos.Conversation.Takeover(ctx, conversationID, agentID, conv.UserID)
}

func (s *ChatService) Escalate(ctx context.Context, userID, conversationID, reason string) error {
	conv, err := s.repos.Conversation.GetByIDAndUser(ctx, conversationID, userID)
	if err != nil {
		return err
	}
	if conv == nil {
		return fmt.Errorf("conversation not found")
	}
	if err := s.repos.Conversation.UpdateStatus(ctx, conversationID, "escalated", conv.UserID); err != nil {
		return err
	}
	msg := &domain.Message{
		ConversationID: conversationID,
		Role:           "system",
		Content:        fmt.Sprintf("Conversation escalated. Reason: %s", reason),
		IsRead:         false,
	}
	return s.repos.Message.Create(ctx, msg)
}

func (s *ChatService) RateConversation(ctx context.Context, userID, conversationID string, score int, feedback string) error {
	if score < 1 || score > 5 {
		return fmt.Errorf("score must be between 1 and 5")
	}
	conv, err := s.repos.Conversation.GetByIDAndUser(ctx, conversationID, userID)
	if err != nil || conv == nil {
		return fmt.Errorf("conversation not found")
	}
	if s.redis == nil {
		return nil
	}
	rating := map[string]interface{}{
		"score":      score,
		"feedback":   feedback,
		"created_at": time.Now(),
	}
	data, _ := json.Marshal(rating)
	infrastructure.CSATScore.Observe(float64(score))
	s.logger.Info("CSAT rating recorded", "conversation_id", conversationID, "score", score, "feedback", feedback)
	ttl := 90 * 24 * time.Hour
	return s.redis.Set(ctx, fmt.Sprintf("conv:%s:rating", conversationID), string(data), ttl)
}

func (s *ChatService) SendMessage(ctx context.Context, userID, conversationID, senderType, content string) (*domain.Message, error) {
	conv, err := s.repos.Conversation.GetByIDAndUser(ctx, conversationID, userID)
	if err != nil {
		return nil, fmt.Errorf("conversation not found or unauthorized")
	}
	if conv == nil {
		return nil, fmt.Errorf("conversation not found")
	}

	// Determine if this is an agent/human sending the message
	role := senderType
	isAgent := senderType == "agent" || senderType == "human"

	// If the message is sent from the dashboard, treat it as an agent reply unless it is the internal AI chat
	if senderType == "customer" && conv.CustomerName != "Noant AI" {
		role = "agent"
		isAgent = true
	}

	msg := &domain.Message{
		ConversationID: conversationID,
		Role:           role,
		Content:        content,
		IsRead:         true, // Agent replies are read by default
	}
	if err := s.repos.Message.Create(ctx, msg); err != nil {
		return nil, err
	}

	// If it is an agent reply, send it to the external customer channel
	if isAgent {
		if conv.Channel == "whatsapp" && s.openwa != nil {
			// Find active WhatsApp integration to get sessionID
			integration, err := s.repos.Integration.GetByUserAndChannel(ctx, userID, "whatsapp")
			if err == nil && integration != nil {
				if sessionID, _ := integration.Config["session_id"].(string); sessionID != "" {
					chatID := FormatChatID(conv.CustomerPhone)
					s.logger.Info("Sending manual agent WhatsApp reply", "session", sessionID, "chatID", chatID)
					// Send text message asynchronously to avoid blocking the HTTP response
					go func() {
						if err := s.openwa.SendTextMessage(sessionID, chatID, content); err != nil {
							s.logger.Error("Failed to send manual agent WhatsApp message", "error", err)
						}
					}()
				}
			}
		} else if conv.Channel == "telegram" && s.telegram != nil {
			// Find active Telegram integration to get bot token
			integration, err := s.repos.Integration.GetByUserAndChannel(ctx, userID, "telegram")
			if err == nil && integration != nil {
				if botToken, _ := integration.Config["bot_token"].(string); botToken != "" {
					chatID, err := strconv.ParseInt(conv.CustomerPhone, 10, 64)
					if err == nil {
						s.logger.Info("Sending manual agent Telegram reply", "chatID", chatID)
						go func() {
							if err := s.telegram.SendTextMessage(context.Background(), botToken, chatID, content); err != nil {
								s.logger.Error("Failed to send manual agent Telegram message", "error", err)
							}
						}()
					}
				}
			}
		}
	}

	return msg, nil
}

func (s *ChatService) GenerateAIResponse(ctx context.Context, conversationID, userMessage string) (*domain.Message, error) {
	if !s.beginAIReply(conversationID, userMessage) {
		s.logger.Info("Skipping duplicate AI reply", "conversationID", conversationID)
		return nil, nil
	}
	defer s.abortAIReply(conversationID)

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
		Role:           "ai",
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
		conv, err := s.repos.Conversation.GetByID(ctx, conversationID)
		if err == nil && conv != nil {
			_ = s.repos.Conversation.UpdateStatus(ctx, conversationID, "escalated", conv.UserID)
		}
	}
	s.completeAIReply(conversationID, userMessage)
	return aiMsg, nil
}

// StoreWhatsAppIntegration stores the WhatsApp integration config
func (s *ChatService) StoreWhatsAppIntegration(ctx context.Context, userID, sessionID, phone string) {
	s.StoreWhatsAppIntegrationWithStatus(ctx, userID, sessionID, phone, "connected")
}

// StoreWhatsAppIntegrationWithStatus stores the WhatsApp integration config with a custom status
func (s *ChatService) StoreWhatsAppIntegrationWithStatus(ctx context.Context, userID, sessionID, phone, status string) {
	existing, err := s.repos.Integration.GetByUserAndChannel(ctx, userID, "whatsapp")
	phoneVal := phone
	if phoneVal == "" && existing != nil {
		if p, ok := existing.Config["phone"].(string); ok {
			phoneVal = p
		}
	}
	if status == "" {
		status = "connected"
	}
	integration := &domain.Integration{
		UserID:     userID,
		Channel:    "whatsapp",
		Status:     status,
		WebhookURL: fmt.Sprintf("%s/api/v1/openwa/webhook", s.cfg.APIURL),
		Config: map[string]interface{}{
			"session_id": sessionID,
			"phone":      phoneVal,
			"type":       "openwa",
		},
	}
	if err == nil && existing != nil {
		integration.ID = existing.ID
		_ = s.repos.Integration.Update(ctx, integration)
		return
	}
	_ = s.repos.Integration.Create(ctx, integration)
}

// EnsureConversation finds or creates a conversation for a customer on a given channel
func (s *ChatService) EnsureConversation(ctx context.Context, userID, customerName, customerKey, channel, customerAvatar string) (*domain.Conversation, error) {
	existing, _ := s.repos.Conversation.FindActiveByCustomer(ctx, userID, customerKey, channel)
	if existing != nil {
		return existing, nil
	}
	conv := &domain.Conversation{
		UserID:        userID,
		CustomerName:  customerName,
		CustomerPhone: customerKey,
		Channel:       channel,
		Status:        "active",
	}
	if err := s.repos.Conversation.Create(ctx, conv); err != nil {
		return nil, err
	}
	return conv, nil
}

// StoreMediaRecord stores a media message record in the database
func (s *ChatService) StoreMediaRecord(ctx context.Context, conversationID, userID, sessionID string, msg *OpenWAMessageData) error {
	record := &domain.MediaMessage{
		ConversationID: conversationID,
		UserID:         userID,
		SessionID:      sessionID,
		MessageID:      msg.ID,
		MediaType:      msg.MediaType,
		MimeType:       msg.MimeType,
		FileSize:       msg.FileSize,
		FileName:       msg.FileName,
		Width:          msg.Width,
		Height:         msg.Height,
		Duration:       msg.Duration,
		RemoteURL:      msg.MediaURL,
		ExpiresAt:      time.Now().Add(30 * 24 * time.Hour), // 30-day retention
	}
	if msg.Latitude != 0 || msg.Longitude != 0 {
		record.RemoteURL = fmt.Sprintf("%f,%f", msg.Latitude, msg.Longitude)
		record.MediaType = "location"
	}
	if msg.VCard != "" {
		record.RemoteURL = msg.VCard
		record.MediaType = "contact"
	}
	return s.repos.MediaMessage.Create(ctx, record)
}

// GetWhatsAppIntegration returns the WhatsApp integration for a user regardless of
// connection state, so callers can operate on connecting / qr_ready sessions too.
// Returns nil only when no record exists or a hard error occurred.
func (s *ChatService) GetWhatsAppIntegration(ctx context.Context, userID string) (*domain.Integration, error) {
	integration, err := s.repos.Integration.GetByUserAndChannel(ctx, userID, "whatsapp")
	if err != nil || integration == nil {
		return integration, err
	}
	// Exclude only hard-failed or explicitly disconnected integrations
	if integration.Status == "error" || integration.Status == "disconnected" || integration.Status == "inactive" {
		return nil, nil
	}
	return integration, nil
}

// GetWhatsAppIntegrationBySessionID returns the WhatsApp integration that owns a given OpenWA session
func (s *ChatService) GetWhatsAppIntegrationBySessionID(ctx context.Context, sessionID string) (*domain.Integration, error) {
	return s.repos.Integration.GetByChannelAndSessionID(ctx, "whatsapp", sessionID)
}

// GetTelegramIntegrationByWebhookSecret returns the Telegram integration that owns a webhook secret.
func (s *ChatService) GetTelegramIntegrationByWebhookSecret(ctx context.Context, secret string) (*domain.Integration, error) {
	return s.repos.Integration.GetByChannelAndWebhookSecret(ctx, "telegram", secret)
}

// DisconnectWhatsAppSession completely logs out, unregisters, and deletes the WhatsApp session
func (s *ChatService) DisconnectWhatsAppSession(ctx context.Context, userID string) {
	if s.openwa == nil {
		return
	}
	integration, err := s.repos.Integration.GetByUserAndChannel(ctx, userID, "whatsapp")
	if err == nil && integration != nil {
		if sessionID, _ := integration.Config["session_id"].(string); sessionID != "" {
			s.logger.Info("Logging out and deleting WhatsApp session", "sessionID", sessionID)
			if mgr := s.openwa.GetSessionManager(); mgr != nil {
				mgr.UnregisterSession(sessionID)
			}
			_ = s.openwa.LogoutSession(sessionID)
			_ = s.openwa.DeleteSession(sessionID)
		}
	}
}

// RemoveWhatsAppIntegration removes the WhatsApp integration
func (s *ChatService) RemoveWhatsAppIntegration(ctx context.Context, userID string) {
	_ = s.repos.Integration.Disconnect(ctx, userID, "whatsapp")
}

func (s *ChatService) GetMediaByConversation(ctx context.Context, convID, userID string) ([]domain.MediaMessage, error) {
	conv, err := s.repos.Conversation.GetByID(ctx, convID)
	if err != nil {
		return nil, err
	}
	if conv == nil || conv.UserID != userID {
		return nil, fmt.Errorf("conversation not found")
	}
	return s.repos.MediaMessage.GetByConversation(ctx, convID)
}

func (s *ChatService) ClearChats(ctx context.Context, userID string) error {
	return s.repos.Conversation.ClearChats(ctx, userID)
}
