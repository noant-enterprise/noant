package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"noant/internal/infrastructure"
)

// SendTextMessage sends a text reply to a customer via OpenWA
func (s *OpenWAService) SendTextMessage(sessionID, chatID, text string) error {
	if !s.cfg.OpenWAEnabled {
		s.logger.Warn("OpenWA is disabled, skipping message send")
		return nil
	}

	url := fmt.Sprintf("%s/api/sessions/%s/messages/send-text",
		s.cfg.OpenWABaseURL, sessionID)

	payload := map[string]string{
		"chatId": chatID,
		"text":   text,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal OpenWA payload: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("failed to create OpenWA request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if s.cfg.OpenWAApiKey != "" {
		req.Header.Set("X-API-Key", s.cfg.OpenWAApiKey)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		infrastructure.OpenWAMessagesSentTotal.WithLabelValues("text", "error").Inc()
		s.logger.Error("OpenWA send failed", "error", err, "chatID", chatID, "sessionID", sessionID)
		return fmt.Errorf("openwa request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		s.logger.Error("OpenWA send error", "status", resp.StatusCode, "body", string(body), "chatID", chatID, "sessionID", sessionID)
		infrastructure.OpenWAMessagesSentTotal.WithLabelValues("text", "error").Inc()
		return fmt.Errorf("openwa returned status %d: %s", resp.StatusCode, string(body))
	}

	infrastructure.OpenWAMessagesSentTotal.WithLabelValues("text", "success").Inc()
	s.logger.Info("OpenWA message sent", "chatID", chatID, "sessionID", sessionID, "length", len(text))
	return nil
}

// SendMediaMessage sends an image/document via OpenWA
func (s *OpenWAService) SendMediaMessage(sessionID, chatID, mediaURL, caption string) error {
	if !s.cfg.OpenWAEnabled {
		return nil
	}

	url := fmt.Sprintf("%s/api/sessions/%s/messages/send-media",
		s.cfg.OpenWABaseURL, sessionID)

	payload := map[string]string{
		"chatId":  chatID,
		"file":    mediaURL,
		"caption": caption,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	if s.cfg.OpenWAApiKey != "" {
		req.Header.Set("X-API-Key", s.cfg.OpenWAApiKey)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		infrastructure.OpenWAMessagesSentTotal.WithLabelValues("media", "error").Inc()
		return fmt.Errorf("openwa media send failed: %d %s", resp.StatusCode, string(body))
	}

	infrastructure.OpenWAMessagesSentTotal.WithLabelValues("media", "success").Inc()
	return nil
}

// ========== INTERNAL SEND METHODS (used by queue worker) ==========

// sendTextMessageInternal sends a text message and tracks rate limit headers
func (s *OpenWAService) sendTextMessageInternal(sessionID, chatID, text string) error {
	if s.cbManager.IsOpen(sessionID) {
		return fmt.Errorf("circuit breaker open for session %s", sessionID)
	}

	url := fmt.Sprintf("%s/api/sessions/%s/messages/send-text", s.cfg.OpenWABaseURL, sessionID)
	payload := map[string]string{"chatId": chatID, "text": text}
	jsonPayload, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if s.cfg.OpenWAApiKey != "" {
		req.Header.Set("X-API-Key", s.cfg.OpenWAApiKey)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		s.cbManager.RecordFailure(sessionID)
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	remaining := resp.Header.Get("X-RateLimit-Remaining")
	if remaining != "" {
		var rem int
		if _, err := fmt.Sscanf(remaining, "%d", &rem); err == nil {
			s.rateLimitHeaders.Update(sessionID, rem, time.Now().Add(1*time.Minute))
		}
	}

	if resp.StatusCode == http.StatusTooManyRequests {
		s.cbManager.RecordFailure(sessionID)
		return fmt.Errorf("rate limited by OpenWA (429)")
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		s.cbManager.RecordFailure(sessionID)
		return fmt.Errorf("openwa returned status %d: %s", resp.StatusCode, string(body))
	}

	s.cbManager.RecordSuccess(sessionID)
	return nil
}

// sendMediaMessageInternal sends a media message internally (used by queue worker)
func (s *OpenWAService) sendMediaMessageInternal(sessionID, chatID, mediaURL, caption string) error {
	if s.cbManager.IsOpen(sessionID) {
		return fmt.Errorf("circuit breaker open for session %s", sessionID)
	}

	url := fmt.Sprintf("%s/api/sessions/%s/messages/send-media", s.cfg.OpenWABaseURL, sessionID)
	payload := map[string]string{"chatId": chatID, "file": mediaURL, "caption": caption}
	jsonPayload, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if s.cfg.OpenWAApiKey != "" {
		req.Header.Set("X-API-Key", s.cfg.OpenWAApiKey)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		s.cbManager.RecordFailure(sessionID)
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusTooManyRequests {
		s.cbManager.RecordFailure(sessionID)
		return fmt.Errorf("rate limited by OpenWA (429)")
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		s.cbManager.RecordFailure(sessionID)
		return fmt.Errorf("openwa media send failed: %d %s", resp.StatusCode, string(body))
	}

	s.cbManager.RecordSuccess(sessionID)
	return nil
}

// sendTemplateMessageInternal sends a template message via OpenWA
func (s *OpenWAService) sendTemplateMessageInternal(sessionID, chatID string, params map[string]interface{}) error {
	if s.cbManager.IsOpen(sessionID) {
		return fmt.Errorf("circuit breaker open for session %s", sessionID)
	}

	url := fmt.Sprintf("%s/api/sessions/%s/messages/send-template", s.cfg.OpenWABaseURL, sessionID)

	payload := map[string]interface{}{
		"chatId": chatID,
	}
	if tpl, ok := params["template"].(string); ok {
		payload["template"] = tpl
	}
	if lang, ok := params["language"].(string); ok {
		payload["language"] = lang
	}
	if ns, ok := params["namespace"].(string); ok {
		payload["namespace"] = ns
	}
	if vars, ok := params["variables"].(map[string]string); ok {
		payload["variables"] = vars
	}

	return s.sendRawMessage(url, payload)
}

// sendRawMessage sends a generic JSON payload to OpenWA
func (s *OpenWAService) sendRawMessage(url string, payload map[string]interface{}) error {
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if s.cfg.OpenWAApiKey != "" {
		req.Header.Set("X-API-Key", s.cfg.OpenWAApiKey)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		s.cbManager.RecordFailure("")
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusTooManyRequests {
		s.cbManager.RecordFailure("")
		return fmt.Errorf("rate limited by OpenWA (429)")
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("openwa request failed: %d %s", resp.StatusCode, string(body))
	}

	s.cbManager.RecordSuccess("")
	return nil
}

// EnqueueMessage adds a message to the outbound queue (main entry point for sending)
func (s *OpenWAService) EnqueueMessage(sessionID, userID, chatID, text string) error {
	entry := &QueueEntry{
		ID:        fmt.Sprintf("msg_%s_%d", sessionID, time.Now().UnixNano()),
		SessionID: sessionID,
		UserID:    userID,
		ChatID:    chatID,
		MsgType:   MsgTypeText,
		Content:   text,
		Priority:  PriorityNormal,
	}
	return s.queue.Enqueue(entry)
}

// EnqueueMediaMessage adds a media message to the outbound queue
func (s *OpenWAService) EnqueueMediaMessage(sessionID, userID, chatID, mediaURL, caption string) error {
	entry := &QueueEntry{
		ID:        fmt.Sprintf("media_%s_%d", sessionID, time.Now().UnixNano()),
		SessionID: sessionID,
		UserID:    userID,
		ChatID:    chatID,
		MsgType:   MsgTypeMedia,
		MediaURL:  mediaURL,
		Caption:   caption,
		Priority:  PriorityNormal,
	}
	return s.queue.Enqueue(entry)
}

// EnqueueTemplateMessage adds a template message to the outbound queue
func (s *OpenWAService) EnqueueTemplateMessage(sessionID, userID, chatID string, params map[string]interface{}) error {
	content, _ := json.Marshal(params)
	entry := &QueueEntry{
		ID:        fmt.Sprintf("tpl_%s_%d", sessionID, time.Now().UnixNano()),
		SessionID: sessionID,
		UserID:    userID,
		ChatID:    chatID,
		MsgType:   MsgTypeTemplate,
		Content:   string(content),
		Priority:  PriorityNormal,
	}
	return s.queue.Enqueue(entry)
}

// FormatChatID converts a phone number to OpenWA chat ID format for MESSAGING
func FormatChatID(phone string) string {
	// For sending messages, OpenWA uses: phone@s.whatsapp.net
	phone = CleanPhoneNumber(phone)
	return phone + "@s.whatsapp.net"
}

// FormatContactID converts a phone number to the contact ID format for the CONTACTS API
func FormatContactID(phone string) string {
	// For the contacts API, the format is: phone@c.us
	phone = CleanPhoneNumber(phone)
	return phone + "@c.us"
}

// CleanPhoneNumber removes all non-digit characters from a phone number
func CleanPhoneNumber(phone string) string {
	result := make([]byte, 0, len(phone))
	for _, c := range phone {
		if c >= '0' && c <= '9' {
			result = append(result, byte(c))
		}
	}
	return string(result)
}
