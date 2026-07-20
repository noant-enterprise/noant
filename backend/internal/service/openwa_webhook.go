package service

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"bytes"
	"io"
)

// OpenWA webhook payload structures

type OpenWAWebhookPayload struct {
	Event     string          `json:"event"`
	SessionID string          `json:"sessionId"`
	Data      json.RawMessage `json:"data"`
	// Timestamp is declared as interface{} because OpenWA may send it
	// as either a JSON number or a JSON string depending on the version.
	Timestamp interface{} `json:"timestamp"`
}

type OpenWAMessageData struct {
	ID        string       `json:"id"`
	From      string       `json:"from"` // phone@s.whatsapp.net
	To        string       `json:"to"`
	Body      string       `json:"body"`
	Type      string       `json:"type"` // text, image, etc.
	Timestamp interface{}  `json:"timestamp"`
	FromMe    bool         `json:"fromMe"`
	HasMedia  bool         `json:"hasMedia"`
	MediaType string       `json:"mediaType"`
	MimeType  string       `json:"mimeType"`
	FileName  string       `json:"fileName"`
	FileSize  int64        `json:"fileSize"`
	MediaURL  string       `json:"mediaUrl"`
	Width     int          `json:"width"`
	Height    int          `json:"height"`
	Duration  int          `json:"duration"`
	Latitude  float64      `json:"latitude"`
	Longitude float64      `json:"longitude"`
	Address   string       `json:"address"`
	VCard     string       `json:"vcard"`
	Sender    OpenWASender `json:"sender"`
}

type OpenWASender struct {
	ID                 string           `json:"id"`
	Name               string           `json:"name"`
	ShortName          string           `json:"shortName"`
	Pushname           string           `json:"pushname"`
	FormattedName      string           `json:"formattedName"`
	ProfilePicThumbObj OpenWAProfilePic `json:"profilePicThumbObj"`
}

type OpenWAProfilePic struct {
	Eurl string `json:"eurl"`
	Tag  string `json:"tag"`
}

type OpenWAStatusData struct {
	ID     string `json:"id"`
	Status string `json:"status"` // sent, delivered, read
}

// VerifyWebhookSignature verifies HMAC-SHA256 signature from OpenWA webhook
func (s *OpenWAService) VerifyWebhookSignature(payload []byte, signature string) bool {
	if s.cfg.OpenWAWebhookSecret == "" {
		return true // No secret configured, skip verification
	}

	if len(signature) > 7 && signature[:7] == "sha256=" {
		signature = signature[7:]
	}

	mac := hmac.New(sha256.New, []byte(s.cfg.OpenWAWebhookSecret))
	mac.Write(payload)
	expectedMAC := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(signature), []byte(expectedMAC))
}

// ParseWebhookEvent extracts the event type and parses the data
func (s *OpenWAService) ParseWebhookEvent(payload []byte) (*OpenWAWebhookPayload, error) {
	var event OpenWAWebhookPayload
	if err := json.Unmarshal(payload, &event); err != nil {
		return nil, fmt.Errorf("failed to parse webhook payload: %w", err)
	}
	return &event, nil
}

// ParseMessageData parses the message data from a message.received event
func (s *OpenWAService) ParseMessageData(data json.RawMessage) (*OpenWAMessageData, error) {
	var msg OpenWAMessageData
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, fmt.Errorf("failed to parse message data: %w", err)
	}
	return &msg, nil
}

// ParseStatusData parses the status data from a message.status event
func (s *OpenWAService) ParseStatusData(data json.RawMessage) (*OpenWAStatusData, error) {
	var status OpenWAStatusData
	if err := json.Unmarshal(data, &status); err != nil {
		return nil, fmt.Errorf("failed to parse status data: %w", err)
	}
	return &status, nil
}

// ConfigureWebhook sets the webhook URL for a session
func (s *OpenWAService) ConfigureWebhook(sessionID, webhookURL, secret string) error {
	url := fmt.Sprintf("%s/api/sessions/%s/webhooks", s.cfg.OpenWABaseURL, sessionID)

	payload := map[string]interface{}{
		"url":    webhookURL,
		"events": []string{"message.received", "message.status", "session.status"},
	}
	if secret != "" {
		payload["secret"] = secret
	}

	jsonPayload, _ := json.Marshal(payload)
	s.logger.Info("Configuring webhook", "url", url, "payload", string(jsonPayload))

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
		s.circuitBreaker.RecordFailure()
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	body, _ := io.ReadAll(resp.Body)
	s.logger.Info("Webhook config response", "status", resp.StatusCode, "body", string(body))

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
		s.logger.Info("Webhook configured successfully", "sessionID", sessionID, "url", webhookURL)
		return nil
	}

	return fmt.Errorf("webhook config failed: %d %s", resp.StatusCode, string(body))
}
