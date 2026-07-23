package service

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"noant/config"
	"noant/internal/domain"
	"noant/internal/infrastructure"
	"noant/internal/repository"
)

// ========== INTEGRATION SERVICE ==========

type IntegrationService struct {
	cfg              *config.Config
	repos            *repository.Repositories
	redis            *infrastructure.RedisClient
	logger           *infrastructure.Logger
	chat             *ChatService
	telegram         *TelegramService
	broadcastFn      func(convID string, msgType string, data interface{})
	telegramPollers  map[string]context.CancelFunc
	telegramPollerMu sync.Mutex
}

func NewIntegrationService(cfg *config.Config, repos *repository.Repositories, redis *infrastructure.RedisClient, logger *infrastructure.Logger, chat *ChatService, telegram *TelegramService, broadcastFn func(convID string, msgType string, data interface{})) *IntegrationService {
	return &IntegrationService{
		cfg:             cfg,
		repos:           repos,
		redis:           redis,
		logger:          logger,
		chat:            chat,
		telegram:        telegram,
		broadcastFn:     broadcastFn,
		telegramPollers: map[string]context.CancelFunc{},
	}
}

func (s *IntegrationService) List(ctx context.Context, userID string) ([]domain.Integration, error) {
	return s.repos.Integration.ListByOrg(ctx, userID)
}

func (s *IntegrationService) Connect(ctx context.Context, userID, channel string, cfg map[string]interface{}) (*domain.Integration, error) {
	existing, err := s.repos.Integration.GetByOrgAndChannel(ctx, userID, channel)
	if err != nil {
		return nil, err
	}

	mergedConfig := mergeIntegrationConfig(nil, cfg)
	var integration *domain.Integration
	if existing != nil {
		existing.Status = "active"
		existing.Config = mergeIntegrationConfig(existing.Config, mergedConfig)
		existing.WebhookURL = fmt.Sprintf("%s/api/v1/webhooks/%s", s.cfg.APIURL, channel)
		integration = existing
	} else {
		integration = &domain.Integration{
			UserID:     userID,
			OrgID:      userID,
			Channel:    channel,
			Status:     "active",
			Config:     mergedConfig,
			WebhookURL: fmt.Sprintf("%s/api/v1/webhooks/%s", s.cfg.APIURL, channel),
		}
	}

	if channel == "telegram" {
		updated, err := s.configureTelegramIntegration(ctx, integration, cfg)
		if err != nil {
			return nil, err
		}
		integration = updated
	}

	if existing != nil {
		if err := s.repos.Integration.Update(ctx, integration); err != nil {
			return nil, err
		}
	} else {
		if err := s.repos.Integration.Create(ctx, integration); err != nil {
			return nil, err
		}
	}

	if channel == "telegram" {
		s.applyTelegramDeliveryMode(ctx, integration)
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
	if channel == "whatsapp" && s.chat != nil {
		s.chat.DisconnectWhatsAppSession(ctx, userID)
	}
	if channel == "telegram" && s.telegram != nil {
		if integration, err := s.repos.Integration.GetByOrgAndChannel(ctx, userID, channel); err == nil && integration != nil {
			s.stopTelegramPolling(integration.ID)
			if token, _ := integration.Config["bot_token"].(string); strings.TrimSpace(token) != "" {
				if err := s.telegram.DeleteWebhook(ctx, token); err != nil {
					s.logger.Warn("Failed to delete Telegram webhook", "error", err)
				}
			}
		}
	}

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

// SyncTelegramWebhooks re-applies webhook configuration for active Telegram integrations.
func (s *IntegrationService) SyncTelegramWebhooks(ctx context.Context) error {
	integrations, err := s.repos.Integration.ListActive(ctx)
	if err != nil {
		return err
	}

	for i := range integrations {
		integration := &integrations[i]
		if integration.Channel != "telegram" {
			continue
		}

		updated, err := s.configureTelegramIntegration(ctx, integration, integration.Config)
		if err != nil {
			s.logger.Warn("Failed to sync Telegram webhook", "integrationID", integration.ID, "userID", integration.UserID, "error", err)
			continue
		}
		if err := s.repos.Integration.Update(ctx, updated); err != nil {
			s.logger.Warn("Failed to persist Telegram webhook sync", "integrationID", integration.ID, "userID", integration.UserID, "error", err)
			continue
		}
		s.applyTelegramDeliveryMode(ctx, updated)
	}

	return nil
}

func (s *IntegrationService) configureTelegramIntegration(ctx context.Context, integration *domain.Integration, cfg map[string]interface{}) (*domain.Integration, error) {
	if s.telegram == nil {
		return nil, fmt.Errorf("telegram service is not available")
	}

	token := ""
	if cfg != nil {
		if v, ok := cfg["bot_token"].(string); ok {
			token = strings.TrimSpace(v)
		}
	}
	if token == "" {
		token = strings.TrimSpace(s.cfg.TelegramBotToken)
	}
	if token == "" {
		return nil, fmt.Errorf("telegram bot token is required")
	}

	info, err := s.telegram.GetBotInfo(ctx, token)
	if err != nil {
		return nil, err
	}

	secret := ""
	if cfg != nil {
		if v, ok := cfg["webhook_secret"].(string); ok {
			secret = strings.TrimSpace(v)
		}
	}
	if integration.Config != nil {
		if v, ok := integration.Config["webhook_secret"].(string); ok {
			secret = strings.TrimSpace(v)
		}
	}
	if secret == "" {
		secret = generateRandomString(32)
	}

	webhookURL := strings.TrimSpace(s.cfg.TelegramWebhookURL)
	if cfg != nil {
		if v, ok := cfg["webhook_url"].(string); ok && strings.TrimSpace(v) != "" {
			webhookURL = strings.TrimSpace(v)
		}
	}
	if webhookURL == "" {
		webhookURL = fmt.Sprintf("%s/api/v1/telegram/webhook", s.cfg.APIURL)
	}

	deliveryMode := "webhook"
	if !isPublicTelegramWebhookURL(webhookURL) {
		deliveryMode = "polling"
		if err := s.telegram.DeleteWebhook(ctx, token); err != nil {
			s.logger.Warn("Failed to delete Telegram webhook before enabling polling", "error", err)
		}
	}

	if deliveryMode == "webhook" {
		if err := s.telegram.SetWebhook(ctx, token, webhookURL, secret); err != nil {
			return nil, err
		}
	}

	if integration.Config == nil {
		integration.Config = map[string]interface{}{}
	}
	integration.Config["bot_token"] = token
	integration.Config["bot_username"] = info.Result.Username
	integration.Config["bot_first_name"] = info.Result.FirstName
	integration.Config["delivery_mode"] = deliveryMode
	if deliveryMode == "webhook" {
		integration.Config["webhook_secret"] = secret
		integration.Config["webhook_url"] = webhookURL
		integration.WebhookURL = webhookURL
	} else {
		integration.Config["webhook_secret"] = ""
		integration.Config["webhook_url"] = ""
		integration.WebhookURL = ""
	}
	integration.Status = "active"

	return integration, nil
}

func mergeIntegrationConfig(existing, updates map[string]interface{}) map[string]interface{} {
	merged := cloneIntegrationConfig(existing)
	for key, value := range updates {
		merged[key] = value
	}
	return merged
}

func cloneIntegrationConfig(src map[string]interface{}) map[string]interface{} {
	if len(src) == 0 {
		return map[string]interface{}{}
	}

	dst := make(map[string]interface{}, len(src))
	for key, value := range src {
		dst[key] = value
	}
	return dst
}

func isPublicTelegramWebhookURL(raw string) bool {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return false
	}

	parsed, err := url.Parse(raw)
	if err != nil {
		return false
	}
	if !strings.EqualFold(parsed.Scheme, "https") {
		return false
	}

	host := strings.ToLower(parsed.Hostname())
	switch host {
	case "localhost", "127.0.0.1", "0.0.0.0", "::1":
		return false
	}

	if ip := net.ParseIP(host); ip != nil {
		if ip.IsLoopback() || ip.IsPrivate() || ip.IsUnspecified() {
			return false
		}
	}

	return true
}

func (s *IntegrationService) HandleTelegramIncoming(ctx context.Context, integration *domain.Integration, incoming *TelegramIncomingMessage) (*domain.Conversation, *domain.Message, error) {
	if s.chat == nil {
		return nil, nil, fmt.Errorf("chat service is not available")
	}
	if s.telegram == nil {
		return nil, nil, fmt.Errorf("telegram service is not available")
	}
	if integration == nil {
		return nil, nil, fmt.Errorf("telegram integration is required")
	}
	if incoming == nil {
		return nil, nil, fmt.Errorf("telegram message is required")
	}

	botToken := ""
	if integration.Config != nil {
		if v, ok := integration.Config["bot_token"].(string); ok {
			botToken = strings.TrimSpace(v)
		}
	}
	if botToken == "" {
		botToken = strings.TrimSpace(s.cfg.TelegramBotToken)
	}
	if botToken == "" {
		return nil, nil, fmt.Errorf("telegram bot token is required")
	}

	customerKey := strconv.FormatInt(incoming.ChatID, 10)
	conv, aiMsg, err := s.chat.DirectChat(ctx, integration.UserID, incoming.DisplayName, customerKey, incoming.Text, "telegram", "")
	if err != nil {
		return nil, nil, err
	}

	if aiMsg != nil && strings.TrimSpace(aiMsg.Content) != "" {
		if err := s.telegram.SendTextMessage(ctx, botToken, incoming.ChatID, aiMsg.Content); err != nil {
			s.logger.Error("Failed to send Telegram reply", "error", err, "chatID", incoming.ChatID, "userID", integration.UserID)
		}
	}

	if s.broadcastFn != nil && conv != nil && aiMsg != nil {
		s.broadcastFn(conv.ID, "new_message", map[string]interface{}{
			"content":     aiMsg.Content,
			"sender_type": "ai",
			"customer":    incoming.DisplayName,
			"customer_id": customerKey,
			"channel":     "telegram",
		})
	}

	return conv, aiMsg, nil
}

func (s *IntegrationService) GetTelegramIntegrationByWebhookSecret(ctx context.Context, secret string) (*domain.Integration, error) {
	return s.repos.Integration.GetByChannelAndWebhookSecret(ctx, "telegram", secret)
}

func (s *IntegrationService) applyTelegramDeliveryMode(ctx context.Context, integration *domain.Integration) {
	if integration == nil || integration.Channel != "telegram" {
		return
	}

	mode := ""
	if integration.Config != nil {
		if v, ok := integration.Config["delivery_mode"].(string); ok {
			mode = strings.ToLower(strings.TrimSpace(v))
		}
	}

	switch mode {
	case "polling":
		s.startTelegramPolling(integration)
	default:
		s.stopTelegramPolling(integration.ID)
	}
}

func (s *IntegrationService) startTelegramPolling(integration *domain.Integration) {
	if integration == nil || integration.Channel != "telegram" {
		return
	}
	if s.telegram == nil || s.chat == nil {
		s.logger.Warn("Telegram polling requested but services are unavailable", "integrationID", integration.ID)
		return
	}

	botToken := ""
	if integration.Config != nil {
		if v, ok := integration.Config["bot_token"].(string); ok {
			botToken = strings.TrimSpace(v)
		}
	}
	if botToken == "" {
		botToken = strings.TrimSpace(s.cfg.TelegramBotToken)
	}
	if botToken == "" {
		s.logger.Warn("Telegram polling not started because bot token is missing", "integrationID", integration.ID)
		return
	}

	s.stopTelegramPolling(integration.ID)

	pollIntegration := &domain.Integration{
		ID:         integration.ID,
		UserID:     integration.UserID,
		OrgID:      integration.OrgID,
		Channel:    integration.Channel,
		Status:     integration.Status,
		Config:     cloneIntegrationConfig(integration.Config),
		WebhookURL: integration.WebhookURL,
	}

	ctx, cancel := context.WithCancel(context.Background())
	s.telegramPollerMu.Lock()
	s.telegramPollers[pollIntegration.ID] = cancel
	s.telegramPollerMu.Unlock()

	go s.runTelegramPoller(ctx, pollIntegration, botToken)
	s.logger.Info("Telegram polling started", "integrationID", integration.ID, "userID", integration.UserID)
}

func (s *IntegrationService) stopTelegramPolling(integrationID string) {
	if strings.TrimSpace(integrationID) == "" {
		return
	}

	s.telegramPollerMu.Lock()
	cancel, ok := s.telegramPollers[integrationID]
	if ok {
		delete(s.telegramPollers, integrationID)
	}
	s.telegramPollerMu.Unlock()

	if ok && cancel != nil {
		cancel()
	}
}

func (s *IntegrationService) runTelegramPoller(ctx context.Context, integration *domain.Integration, botToken string) {
	defer s.stopTelegramPolling(integration.ID)

	offset := getConfigInt64(integration.Config, "polling_offset")
	if offset < 0 {
		offset = 0
	}

	backoff := time.Second
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		updates, err := s.telegram.GetUpdates(ctx, botToken, offset, 30)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			s.logger.Warn("Telegram polling failed", "integrationID", integration.ID, "userID", integration.UserID, "error", err)
			time.Sleep(backoff)
			if backoff < 15*time.Second {
				backoff *= 2
			}
			continue
		}

		backoff = time.Second
		for _, update := range updates {
			if update.UpdateID >= offset {
				offset = update.UpdateID + 1
			}

			if incoming, ok := update.IncomingMessage(); ok && incoming != nil && !incoming.IsBot {
				if _, _, err := s.HandleTelegramIncoming(ctx, integration, incoming); err != nil {
					s.logger.Error("Failed to process Telegram update", "integrationID", integration.ID, "userID", integration.UserID, "updateID", update.UpdateID, "error", err)
				}
			}

			s.persistTelegramPollingOffset(ctx, integration, offset)
		}
	}
}

func (s *IntegrationService) persistTelegramPollingOffset(ctx context.Context, integration *domain.Integration, offset int64) {
	if integration == nil {
		return
	}

	if integration.Config == nil {
		integration.Config = map[string]interface{}{}
	}
	integration.Config["polling_offset"] = offset

	if err := s.repos.Integration.Update(ctx, integration); err != nil {
		s.logger.Warn("Failed to persist Telegram polling offset", "integrationID", integration.ID, "userID", integration.UserID, "error", err)
	}
}

func getConfigInt64(cfgMap map[string]interface{}, key string) int64 {
	if cfgMap == nil {
		return 0
	}

	value, ok := cfgMap[key]
	if !ok {
		return 0
	}

	switch v := value.(type) {
	case int:
		return int64(v)
	case int8:
		return int64(v)
	case int16:
		return int64(v)
	case int32:
		return int64(v)
	case int64:
		return v
	case float32:
		return int64(v)
	case float64:
		return int64(v)
	case json.Number:
		if parsed, err := v.Int64(); err == nil {
			return parsed
		}
	case string:
		if parsed, err := strconv.ParseInt(strings.TrimSpace(v), 10, 64); err == nil {
			return parsed
		}
	}

	return 0
}

func (s *IntegrationService) Test(ctx context.Context, channel string, cfg map[string]interface{}) (ok bool, msg string) {
	client := &http.Client{Timeout: 10 * time.Second}

	switch channel {
	case "telegram":
		// Prefer token from the provided config, fall back to environment config
		token := ""
		if cfg != nil {
			if t, ok := cfg["bot_token"].(string); ok {
				token = t
			}
		}
		if token == "" {
			token = s.cfg.TelegramBotToken
		}
		if token == "" {
			return false, "No Telegram Bot Token provided"
		}
		apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/getMe", token)
		resp, err := client.Get(apiURL)
		if err != nil {
			return false, fmt.Sprintf("Connection failed: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()
		var result struct {
			OK     bool `json:"ok"`
			Result struct {
				Username  string `json:"username"`
				FirstName string `json:"first_name"`
			} `json:"result"`
			Description string `json:"description"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return false, "Invalid response from Telegram API"
		}
		if !result.OK {
			return false, fmt.Sprintf("Telegram API error: %s", result.Description)
		}
		return true, fmt.Sprintf("✓ Connected as @%s (%s)", result.Result.Username, result.Result.FirstName)

	case "email":
		toEmail := ""
		subject := "NOANT email integration test"
		body := "If you received this, your SMTP/Gmail email integration is working."
		if cfg != nil {
			if v, ok := cfg["to_email"].(string); ok {
				toEmail = v
			}
			if v, ok := cfg["subject"].(string); ok && strings.TrimSpace(v) != "" {
				subject = v
			}
			if v, ok := cfg["body"].(string); ok && strings.TrimSpace(v) != "" {
				body = v
			}
		}
		if strings.TrimSpace(toEmail) == "" {
			return false, "Recipient email (to_email) is required"
		}
		settings := smtpSettingsFromConfig(s.cfg, cfg)
		if _, err := sendSMTPMessage(ctx, settings, toEmail, subject, fmt.Sprintf("<html><body><p>%s</p></body></html>", html.EscapeString(body))); err != nil {
			return false, fmt.Sprintf("Email test failed: %v", err)
		}
		return true, fmt.Sprintf("✓ Test email sent to %s", toEmail)

	case "whatsapp":
		if s.cfg.OpenWAEnabled && s.chat != nil && s.chat.openwa != nil {
			err := s.chat.openwa.Ping()
			if err != nil {
				return false, fmt.Sprintf("OpenWA server unreachable: %v", err)
			}
			return true, fmt.Sprintf("✓ OpenWA WhatsApp channel healthy (session: %s)", s.cfg.OpenWASessionID)
		}

		phoneNumberID := ""
		accessToken := ""
		if cfg != nil {
			if v, ok := cfg["phone_number_id"].(string); ok {
				phoneNumberID = v
			}
			if v, ok := cfg["access_token"].(string); ok {
				accessToken = v
			}
		}
		if phoneNumberID == "" || accessToken == "" {
			if s.cfg.MetaAccessToken != "" && s.cfg.MetaPhoneNumberID != "" {
				phoneNumberID = s.cfg.MetaPhoneNumberID
				accessToken = s.cfg.MetaAccessToken
			} else {
				return false, "Phone Number ID and Access Token are required"
			}
		}
		apiURL := fmt.Sprintf("https://graph.facebook.com/v21.0/%s", phoneNumberID)
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, http.NoBody)
		if err != nil {
			return false, "Failed to create request"
		}
		req.Header.Set("Authorization", "Bearer "+accessToken)
		resp, err := client.Do(req)
		if err != nil {
			return false, fmt.Sprintf("Connection failed: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()
		var result struct {
			ID          string `json:"id"`
			DisplayName string `json:"display_phone_number"`
			Error       struct {
				Message string `json:"message"`
			} `json:"error"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return false, "Invalid response from Meta Graph API"
		}
		if resp.StatusCode != http.StatusOK {
			return false, fmt.Sprintf("Meta API error: %s", result.Error.Message)
		}
		return true, fmt.Sprintf("✓ WhatsApp number verified: %s (ID: %s)", result.DisplayName, result.ID)

	case "facebook":
		pageID := ""
		pageToken := ""
		if cfg != nil {
			if v, ok := cfg["page_id"].(string); ok {
				pageID = v
			}
			if v, ok := cfg["page_access_token"].(string); ok {
				pageToken = v
			}
		}
		if pageID == "" || pageToken == "" {
			if s.cfg.MetaAccessToken != "" && s.cfg.MetaPageID != "" {
				pageID = s.cfg.MetaPageID
				pageToken = s.cfg.MetaAccessToken
			} else {
				return false, "Page ID and Page Access Token are required"
			}
		}
		apiURL := fmt.Sprintf("https://graph.facebook.com/v21.0/%s?fields=id,name&access_token=%s", pageID, pageToken)
		resp, err := client.Get(apiURL)
		if err != nil {
			return false, fmt.Sprintf("Connection failed: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()
		var result struct {
			ID    string `json:"id"`
			Name  string `json:"name"`
			Error struct {
				Message string `json:"message"`
			} `json:"error"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return false, "Invalid response from Meta Graph API"
		}
		if resp.StatusCode != http.StatusOK {
			return false, fmt.Sprintf("Meta API error: %s", result.Error.Message)
		}
		return true, fmt.Sprintf("✓ Facebook Page verified: %s (ID: %s)", result.Name, result.ID)

	case "instagram":
		instagramID := ""
		pageToken := ""
		if cfg != nil {
			if v, ok := cfg["instagram_id"].(string); ok {
				instagramID = v
			}
			if v, ok := cfg["page_access_token"].(string); ok {
				pageToken = v
			}
		}
		if instagramID == "" || pageToken == "" {
			if s.cfg.MetaAccessToken != "" && s.cfg.InstagramAccountID != "" {
				instagramID = s.cfg.InstagramAccountID
				pageToken = s.cfg.MetaAccessToken
			} else {
				return false, "Instagram Account ID and Page Access Token are required"
			}
		}
		apiURL := fmt.Sprintf("https://graph.facebook.com/v21.0/%s?fields=id,username&access_token=%s", instagramID, pageToken)
		resp, err := client.Get(apiURL)
		if err != nil {
			return false, fmt.Sprintf("Connection failed: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()
		var result struct {
			ID       string `json:"id"`
			Username string `json:"username"`
			Error    struct {
				Message string `json:"message"`
			} `json:"error"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return false, "Invalid response from Meta Graph API"
		}
		if resp.StatusCode != http.StatusOK {
			return false, fmt.Sprintf("Meta API error: %s", result.Error.Message)
		}
		return true, fmt.Sprintf("✓ Instagram account verified: @%s (ID: %s)", result.Username, result.ID)

	case "web":
		return true, "✓ Web chat widget is ready"
	default:
		return false, "Unsupported channel"
	}
}
