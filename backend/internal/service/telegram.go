package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"noant/config"
	"noant/internal/infrastructure"
)

type TelegramService struct {
	cfg    *config.Config
	logger *infrastructure.Logger
	client *http.Client
}

func NewTelegramService(cfg *config.Config, logger *infrastructure.Logger) *TelegramService {
	return &TelegramService{
		cfg:    cfg,
		logger: logger,
		client: &http.Client{Timeout: 35 * time.Second},
	}
}

type TelegramBotInfo struct {
	OK          bool   `json:"ok"`
	Description string `json:"description"`
	Result      struct {
		ID        int64  `json:"id"`
		IsBot     bool   `json:"is_bot"`
		FirstName string `json:"first_name"`
		Username  string `json:"username"`
	} `json:"result"`
}

type TelegramAPIResponse struct {
	OK          bool   `json:"ok"`
	Description string `json:"description"`
}

type TelegramUpdatesResponse struct {
	OK          bool             `json:"ok"`
	Description string           `json:"description"`
	Result      []TelegramUpdate `json:"result"`
}

type TelegramUpdate struct {
	UpdateID      int64            `json:"update_id"`
	Message       *TelegramMessage `json:"message,omitempty"`
	EditedMessage *TelegramMessage `json:"edited_message,omitempty"`
}

type TelegramMessage struct {
	MessageID int64         `json:"message_id"`
	From      *TelegramUser `json:"from,omitempty"`
	Chat      TelegramChat  `json:"chat"`
	Date      int64         `json:"date"`
	Text      string        `json:"text,omitempty"`
	Caption   string        `json:"caption,omitempty"`
}

type TelegramChat struct {
	ID       int64  `json:"id"`
	Type     string `json:"type"`
	Title    string `json:"title,omitempty"`
	Username string `json:"username,omitempty"`
}

type TelegramUser struct {
	ID           int64  `json:"id"`
	IsBot        bool   `json:"is_bot"`
	FirstName    string `json:"first_name,omitempty"`
	LastName     string `json:"last_name,omitempty"`
	Username     string `json:"username,omitempty"`
	LanguageCode string `json:"language_code,omitempty"`
}

type TelegramIncomingMessage struct {
	ChatID      int64
	DisplayName string
	Username    string
	Text        string
	IsBot       bool
	Raw         *TelegramMessage
}

func (u *TelegramUser) DisplayName() string {
	if u == nil {
		return "Telegram User"
	}
	if strings.TrimSpace(u.Username) != "" {
		return "@" + strings.TrimSpace(u.Username)
	}
	fullName := strings.TrimSpace(strings.TrimSpace(u.FirstName + " " + u.LastName))
	if fullName != "" {
		return fullName
	}
	return "Telegram User"
}

func (m *TelegramMessage) DisplayName() string {
	if m == nil {
		return "Telegram User"
	}
	if m.From != nil {
		return m.From.DisplayName()
	}
	if strings.TrimSpace(m.Chat.Username) != "" {
		return "@" + strings.TrimSpace(m.Chat.Username)
	}
	if strings.TrimSpace(m.Chat.Title) != "" {
		return strings.TrimSpace(m.Chat.Title)
	}
	return "Telegram User"
}

func (u *TelegramUpdate) IncomingMessage() (*TelegramIncomingMessage, bool) {
	if u == nil {
		return nil, false
	}
	msg := u.Message
	if msg == nil {
		msg = u.EditedMessage
	}
	if msg == nil {
		return nil, false
	}
	text := strings.TrimSpace(msg.Text)
	if text == "" {
		text = strings.TrimSpace(msg.Caption)
	}
	if text == "" {
		return nil, false
	}
	return &TelegramIncomingMessage{
		ChatID:      msg.Chat.ID,
		DisplayName: msg.DisplayName(),
		Username: func() string {
			if msg.From != nil {
				return msg.From.Username
			}
			return ""
		}(),
		Text:  text,
		IsBot: msg.From != nil && msg.From.IsBot,
		Raw:   msg,
	}, true
}

func (s *TelegramService) request(ctx context.Context, botToken, apiMethod, httpMethod string, payload interface{}) (body []byte, statusCode int, err error) {
	if strings.TrimSpace(botToken) == "" {
		return nil, 0, fmt.Errorf("telegram bot token is required")
	}
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/%s", botToken, apiMethod)

	var reqBody io.Reader
	contentType := ""
	if payload != nil {
		if formBody, formType, err := encodeTelegramPayload(payload); err == nil {
			reqBody = formBody
			contentType = formType
		} else {
			jsonPayload, marshalErr := json.Marshal(payload)
			if marshalErr != nil {
				return nil, 0, marshalErr
			}
			reqBody = bytes.NewBuffer(jsonPayload)
			contentType = "application/json"
		}
	}

	if httpMethod == "" {
		httpMethod = http.MethodPost
	}
	req, err := http.NewRequestWithContext(ctx, httpMethod, apiURL, reqBody)
	if err != nil {
		return nil, 0, err
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, _ := io.ReadAll(resp.Body)
	return respBody, resp.StatusCode, nil
}

func encodeTelegramPayload(payload interface{}) (io.Reader, string, error) {
	values := url.Values{}

	switch v := payload.(type) {
	case map[string]interface{}:
		for key, raw := range v {
			if err := addTelegramFormValue(values, key, raw); err != nil {
				return nil, "", err
			}
		}
	case map[string]string:
		for key, raw := range v {
			values.Set(key, raw)
		}
	case url.Values:
		values = v
	default:
		return nil, "", fmt.Errorf("unsupported telegram payload type")
	}

	return strings.NewReader(values.Encode()), "application/x-www-form-urlencoded", nil
}

func addTelegramFormValue(values url.Values, key string, raw interface{}) error {
	switch v := raw.(type) {
	case string:
		values.Set(key, v)
	case []string:
		encoded, err := json.Marshal(v)
		if err != nil {
			return err
		}
		values.Set(key, string(encoded))
	case []interface{}:
		encoded, err := json.Marshal(v)
		if err != nil {
			return err
		}
		values.Set(key, string(encoded))
	case bool:
		values.Set(key, strconv.FormatBool(v))
	case int:
		values.Set(key, strconv.Itoa(v))
	case int8:
		values.Set(key, strconv.FormatInt(int64(v), 10))
	case int16:
		values.Set(key, strconv.FormatInt(int64(v), 10))
	case int32:
		values.Set(key, strconv.FormatInt(int64(v), 10))
	case int64:
		values.Set(key, strconv.FormatInt(v, 10))
	case float32:
		values.Set(key, strconv.FormatFloat(float64(v), 'f', -1, 32))
	case float64:
		values.Set(key, strconv.FormatFloat(v, 'f', -1, 64))
	case json.Number:
		values.Set(key, v.String())
	default:
		encoded, err := json.Marshal(v)
		if err != nil {
			return err
		}
		values.Set(key, string(encoded))
	}

	return nil
}

func (s *TelegramService) GetBotInfo(ctx context.Context, botToken string) (*TelegramBotInfo, error) {
	respBody, status, err := s.request(ctx, botToken, "getMe", http.MethodGet, nil)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("telegram getMe failed: %d %s", status, string(respBody))
	}

	var result TelegramBotInfo
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to decode telegram bot info: %w", err)
	}
	if !result.OK {
		return nil, fmt.Errorf("telegram API error: %s", result.Description)
	}
	return &result, nil
}

func (s *TelegramService) GetUpdates(ctx context.Context, botToken string, offset int64, timeoutSeconds int) ([]TelegramUpdate, error) {
	if strings.TrimSpace(botToken) == "" {
		return nil, fmt.Errorf("telegram bot token is required")
	}

	query := url.Values{}
	if offset > 0 {
		query.Set("offset", strconv.FormatInt(offset, 10))
	}
	if timeoutSeconds > 0 {
		query.Set("timeout", strconv.Itoa(timeoutSeconds))
	}

	endpoint := fmt.Sprintf("https://api.telegram.org/bot%s/getUpdates", botToken)
	if encoded := query.Encode(); encoded != "" {
		endpoint += "?" + encoded
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, http.NoBody)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("telegram getUpdates failed: %d %s", resp.StatusCode, string(respBody))
	}

	var result TelegramUpdatesResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to decode telegram updates response: %w", err)
	}
	if !result.OK {
		return nil, fmt.Errorf("telegram getUpdates error: %s", result.Description)
	}

	return result.Result, nil
}

func (s *TelegramService) SetWebhook(ctx context.Context, botToken, webhookURL, secret string) error {
	payload := map[string]interface{}{
		"url":             webhookURL,
		"allowed_updates": []string{"message", "edited_message"},
	}
	if strings.TrimSpace(secret) != "" {
		payload["secret_token"] = secret
	}

	respBody, status, err := s.request(ctx, botToken, "setWebhook", http.MethodPost, payload)
	if err != nil {
		return err
	}
	if status != http.StatusOK {
		return fmt.Errorf("telegram setWebhook failed: %d %s", status, string(respBody))
	}

	var result TelegramAPIResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return fmt.Errorf("failed to decode telegram webhook response: %w", err)
	}
	if !result.OK {
		return fmt.Errorf("telegram webhook error: %s", result.Description)
	}
	return nil
}

func (s *TelegramService) DeleteWebhook(ctx context.Context, botToken string) error {
	respBody, status, err := s.request(ctx, botToken, "deleteWebhook", http.MethodPost, map[string]interface{}{
		"drop_pending_updates": false,
	})
	if err != nil {
		return err
	}
	if status != http.StatusOK {
		return fmt.Errorf("telegram deleteWebhook failed: %d %s", status, string(respBody))
	}

	var result TelegramAPIResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return fmt.Errorf("failed to decode telegram deleteWebhook response: %w", err)
	}
	if !result.OK {
		return fmt.Errorf("telegram deleteWebhook error: %s", result.Description)
	}
	return nil
}

func (s *TelegramService) SendTextMessage(ctx context.Context, botToken string, chatID int64, text string) error {
	payload := map[string]interface{}{
		"chat_id": chatID,
		"text":    text,
	}

	respBody, status, err := s.request(ctx, botToken, "sendMessage", http.MethodPost, payload)
	if err != nil {
		return err
	}
	if status != http.StatusOK {
		return fmt.Errorf("telegram sendMessage failed: %d %s", status, string(respBody))
	}

	var result TelegramAPIResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return fmt.Errorf("failed to decode telegram sendMessage response: %w", err)
	}
	if !result.OK {
		return fmt.Errorf("telegram sendMessage error: %s", result.Description)
	}
	return nil
}
