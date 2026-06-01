package service

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io"
	"net/http"
	"strings"
	"time"

	"noant/config"
	"noant/internal/infrastructure"
)

// ========== OpenWA SERVICE ==========

type OpenWAService struct {
	cfg        *config.Config
	httpClient *http.Client
	logger     *infrastructure.Logger
}

func NewOpenWAService(cfg *config.Config, logger *infrastructure.Logger) *OpenWAService {
	return &OpenWAService{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: 90 * time.Second,
		},
		logger: logger,
	}
}

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
	ID       string `json:"id"`
	From     string `json:"from"` // phone@s.whatsapp.net
	To       string `json:"to"`
	Body     string `json:"body"`
	Type     string `json:"type"` // text, image, etc.
	// Timestamp flexible: OpenWA sends string or number
	Timestamp interface{} `json:"timestamp"`
	FromMe   bool   `json:"fromMe"`
	HasMedia bool   `json:"hasMedia"`
}

type OpenWAStatusData struct {
	ID     string `json:"id"`
	Status string `json:"status"` // sent, delivered, read
}

// SendTextMessage sends a text reply to a customer via OpenWA
func (s *OpenWAService) SendTextMessage(sessionID string, chatID string, text string) error {
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
		s.logger.Error("OpenWA send failed", "error", err, "chatID", chatID, "sessionID", sessionID)
		return fmt.Errorf("openwa request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		s.logger.Error("OpenWA send error", "status", resp.StatusCode, "body", string(body), "chatID", chatID, "sessionID", sessionID)
		return fmt.Errorf("openwa returned status %d: %s", resp.StatusCode, string(body))
	}

	s.logger.Info("OpenWA message sent", "chatID", chatID, "sessionID", sessionID, "length", len(text))
	return nil
}

// SendMediaMessage sends an image/document via OpenWA
func (s *OpenWAService) SendMediaMessage(sessionID string, chatID string, mediaURL string, caption string) error {
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
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("openwa media send failed: %d %s", resp.StatusCode, string(body))
	}

	return nil
}

// GetSessionStatus checks if the WhatsApp session is connected
func (s *OpenWAService) GetSessionStatus() (string, error) {
	if !s.cfg.OpenWAEnabled {
		return "disabled", nil
	}

	url := fmt.Sprintf("%s/api/sessions/%s",
		s.cfg.OpenWABaseURL, s.cfg.OpenWASessionID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	if s.cfg.OpenWAApiKey != "" {
		req.Header.Set("X-API-Key", s.cfg.OpenWAApiKey)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "disconnected", err
	}
	defer resp.Body.Close()

	var result struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "unknown", err
	}

	return result.Status, nil
}

// RestartSession reconnects a disconnected WhatsApp session
func (s *OpenWAService) RestartSession() error {
	if !s.cfg.OpenWAEnabled {
		return nil
	}

	url := fmt.Sprintf("%s/api/sessions/%s/restart",
		s.cfg.OpenWABaseURL, s.cfg.OpenWASessionID)

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return err
	}

	if s.cfg.OpenWAApiKey != "" {
		req.Header.Set("X-API-Key", s.cfg.OpenWAApiKey)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("openwa restart failed: %d %s", resp.StatusCode, string(body))
	}

	return nil
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

// FormatChatID converts a phone number to OpenWA chat ID format
func FormatChatID(phone string) string {
	// OpenWA uses format: phone@s.whatsapp.net
	phone = CleanPhoneNumber(phone)
	return phone + "@s.whatsapp.net"
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

// ========== SESSION MANAGEMENT ==========

type sessionInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Ping checks if OpenWA server is reachable
func (s *OpenWAService) Ping() error {
	url := fmt.Sprintf("%s/api/sessions", s.cfg.OpenWABaseURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	if s.cfg.OpenWAApiKey != "" {
		req.Header.Set("X-API-Key", s.cfg.OpenWAApiKey)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("OpenWA not reachable: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 500 {
		return fmt.Errorf("OpenWA server error: %d", resp.StatusCode)
	}

	return nil
}

// findSessionByName lists all sessions and finds the one matching the name
func (s *OpenWAService) findSessionByName(name string) (string, error) {
	url := fmt.Sprintf("%s/api/sessions", s.cfg.OpenWABaseURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	if s.cfg.OpenWAApiKey != "" {
		req.Header.Set("X-API-Key", s.cfg.OpenWAApiKey)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var sessions []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&sessions); err != nil {
		return "", err
	}

	for _, sess := range sessions {
		if sess.Name == name {
			return sess.ID, nil
		}
	}
	return "", fmt.Errorf("session not found by name: %s", name)
}

// CreateSession creates a new OpenWA session
func (s *OpenWAService) CreateSession(sessionName string) (string, error) {
	url := fmt.Sprintf("%s/api/sessions", s.cfg.OpenWABaseURL)

	payload := map[string]string{"name": sessionName}
	jsonPayload, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	if s.cfg.OpenWAApiKey != "" {
		req.Header.Set("X-API-Key", s.cfg.OpenWAApiKey)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	// Parse response to get session ID
	var result struct {
		ID      string `json:"id"`
		Name    string `json:"name"`
		Message string `json:"message"`
	}
	_ = json.Unmarshal(body, &result)

	// 201 = created, 409 = already exists (both are OK)
	if resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusOK {
		if result.ID != "" {
			return result.ID, nil
		}
		return sessionName, nil
	}

	if resp.StatusCode == http.StatusConflict {
		// Session already exists — try to get its ID from the list
		s.logger.Info("Session already exists, finding ID", "name", sessionName)
		id, err := s.findSessionByName(sessionName)
		if err == nil && id != "" {
			return id, nil
		}
		// Fallback: use name as ID (some OpenWA versions accept names)
		return sessionName, nil
	}

	return "", fmt.Errorf("create session failed: %d %s", resp.StatusCode, string(body))
}

// StartSession starts an OpenWA session (generates QR code)
func (s *OpenWAService) StartSession(sessionID string) error {
	url := fmt.Sprintf("%s/api/sessions/%s/start", s.cfg.OpenWABaseURL, sessionID)

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return err
	}
	if s.cfg.OpenWAApiKey != "" {
		req.Header.Set("X-API-Key", s.cfg.OpenWAApiKey)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to start session: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		body, _ := io.ReadAll(resp.Body)
		if resp.StatusCode == http.StatusBadRequest && bytes.Contains(bytes.ToLower(body), []byte("already started")) {
			s.logger.Info("OpenWA session already started", "sessionID", sessionID)
			return nil
		}
		return fmt.Errorf("start session failed: %d %s", resp.StatusCode, string(body))
	}

	return nil
}

// GetQRCode retrieves the QR code for a session
func (s *OpenWAService) GetQRCode(sessionID string) (string, error) {
	url := fmt.Sprintf("%s/api/sessions/%s/qr", s.cfg.OpenWABaseURL, sessionID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	if s.cfg.OpenWAApiKey != "" {
		req.Header.Set("X-API-Key", s.cfg.OpenWAApiKey)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to get QR: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	s.logger.Info("OpenWA QR response", "status", resp.StatusCode, "body", string(body)[:min(500, len(body))])

	if resp.StatusCode != http.StatusOK {
		// If "not ready yet" or "not started", return empty (caller will retry)
		if resp.StatusCode == http.StatusBadRequest {
			return "", nil
		}
		return "", fmt.Errorf("get QR failed: %d %s", resp.StatusCode, string(body))
	}

	// Try multiple response formats
	var result struct {
		QR      string `json:"qr"`
		Image   string `json:"image"`
		QRCode  string `json:"qrCode"`
		Data    string `json:"data"`
		Base64  string `json:"base64"`
		URL     string `json:"url"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		// Response might be a plain string (base64 QR)
		trimmed := string(body)
		if len(trimmed) > 100 {
			return trimmed, nil
		}
		return "", fmt.Errorf("failed to parse QR response: %w", err)
	}

	// Check all possible fields
	var rawQR string
	if result.Image != "" {
		rawQR = result.Image
	} else if result.QR != "" {
		rawQR = result.QR
	} else if result.QRCode != "" {
		rawQR = result.QRCode
	} else if result.Data != "" {
		rawQR = result.Data
	} else if result.Base64 != "" {
		rawQR = result.Base64
	}

	if rawQR == "" {
		return "", nil
	}

	// Overlay logo on QR code
	return s.OverlayLogo(rawQR), nil
}

// OverlayLogo adds a logo to the center of a QR code image
func (s *OpenWAService) OverlayLogo(qrBase64 string) string {
	// Decode QR code image
	qrData := qrBase64
	if len(qrData) > 22 && qrData[:22] == "data:image/png;base64," {
		qrData = qrData[22:]
	}

	qrBytes, err := base64.StdEncoding.DecodeString(qrData)
	if err != nil {
		return qrBase64
	}

	qrImg, _, err := image.Decode(bytes.NewReader(qrBytes))
	if err != nil {
		return qrBase64
	}

	qrBounds := qrImg.Bounds()
	qrWidth := qrBounds.Dx()
	logoSize := qrWidth / 4 // Logo area = 25% of QR code

	// Create output image (copy of QR)
	output := image.NewRGBA(qrBounds)
	draw.Draw(output, qrBounds, qrImg, qrBounds.Min, draw.Src)

	centerX := qrWidth / 2
	centerY := qrWidth / 2

	// 1. Draw white circle background
	white := color.RGBA{255, 255, 255, 255}
	radius := logoSize / 2
	for y := -radius; y <= radius; y++ {
		for x := -radius; x <= radius; x++ {
			if x*x+y*y <= radius*radius {
				px := centerX + x
				py := centerY + y
				if px >= qrBounds.Min.X && px < qrBounds.Max.X && py >= qrBounds.Min.Y && py < qrBounds.Max.Y {
					output.Set(px, py, white)
				}
			}
		}
	}

	// 2. Draw black inner circle
	black := color.RGBA{0, 0, 0, 255}
	innerRadius := logoSize / 3
	for y := -innerRadius; y <= innerRadius; y++ {
		for x := -innerRadius; x <= innerRadius; x++ {
			if x*x+y*y <= innerRadius*innerRadius {
				px := centerX + x
				py := centerY + y
				if px >= qrBounds.Min.X && px < qrBounds.Max.X && py >= qrBounds.Min.Y && py < qrBounds.Max.Y {
					output.Set(px, py, black)
				}
			}
		}
	}

	// 3. Draw 3 white dots inside the black circle (left, center, right)
	dotColor := color.RGBA{255, 255, 255, 255}
	dotSizes := []int{innerRadius / 5, innerRadius / 4, innerRadius / 3} // small, medium, large
	dotOffsets := []int{-innerRadius / 3, 0, innerRadius / 3}

	for i, dotR := range dotSizes {
		dotX := centerX + dotOffsets[i]
		dotY := centerY
		for y := -dotR; y <= dotR; y++ {
			for x := -dotR; x <= dotR; x++ {
				if x*x+y*y <= dotR*dotR {
					px := dotX + x
					py := dotY + y
					if px >= qrBounds.Min.X && px < qrBounds.Max.X && py >= qrBounds.Min.Y && py < qrBounds.Max.Y {
						output.Set(px, py, dotColor)
					}
				}
			}
		}
	}

	// Encode back to base64
	var buf bytes.Buffer
	png.Encode(&buf, output)
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(buf.Bytes())
}

// GetSessionStatusByID gets the status of a specific session
func (s *OpenWAService) GetSessionStatusByID(sessionID string) (string, error) {
	url := fmt.Sprintf("%s/api/sessions/%s", s.cfg.OpenWABaseURL, sessionID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	if s.cfg.OpenWAApiKey != "" {
		req.Header.Set("X-API-Key", s.cfg.OpenWAApiKey)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "disconnected", err
	}
	defer resp.Body.Close()

	// 404 = session no longer exists in OpenWA (QR expired / session dropped)
	if resp.StatusCode == http.StatusNotFound {
		return "expired", nil
	}

	body, _ := io.ReadAll(resp.Body)
	var result struct {
		Status string `json:"status"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "unknown", err
	}

	// Normalize different OpenWA connected status variants to a consistent "connected"
	normalized := normalizeSessionStatus(result.Status)
	return normalized, nil
}

// CheckNumberExists checks if a phone number exists on WhatsApp
func (s *OpenWAService) CheckNumberExists(sessionID string, phone string) (bool, error) {
	if !s.cfg.OpenWAEnabled {
		return false, nil
	}

	cleaned := CleanPhoneNumber(phone)
	url := fmt.Sprintf("%s/api/sessions/%s/contacts/check/%s", s.cfg.OpenWABaseURL, sessionID, cleaned)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false, err
	}
	if s.cfg.OpenWAApiKey != "" {
		req.Header.Set("X-API-Key", s.cfg.OpenWAApiKey)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return false, fmt.Errorf("failed to check number existence: %d %s", resp.StatusCode, string(body))
	}

	var result struct {
		Exists bool `json:"exists"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, err
	}

	return result.Exists, nil
}


// normalizeSessionStatus maps all variants of a connected session status to "connected"
func normalizeSessionStatus(status string) string {
	lower := strings.ToLower(strings.TrimSpace(status))
	switch lower {
	case "connected", "authenticated", "ready", "open":
		return "connected"
	// qr_read = phone scanned, session is confirming — treat as connected
	case "qr_read":
		return "connected"
	case "qr_ready", "scan_qr_code", "waitforlogin":
		return "qr_ready"
	case "starting", "initializing", "connecting":
		return "initializing"
	case "failed", "timeout", "error":
		return "failed"
	case "disconnected", "stopped", "inactive":
		return "disconnected"
	// expired = OpenWA returned 404 (session no longer exists)
	case "expired":
		return "expired"
	default:
		if status == "" {
			return "unknown"
		}
		return status
	}
}

// RestartSessionByID restarts a specific session
func (s *OpenWAService) RestartSessionByID(sessionID string) error {
	url := fmt.Sprintf("%s/api/sessions/%s/restart", s.cfg.OpenWABaseURL, sessionID)

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return err
	}
	if s.cfg.OpenWAApiKey != "" {
		req.Header.Set("X-API-Key", s.cfg.OpenWAApiKey)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("restart failed: %d %s", resp.StatusCode, string(body))
	}

	return nil
}

// DeleteSession deletes an OpenWA session
func (s *OpenWAService) DeleteSession(sessionID string) error {
	url := fmt.Sprintf("%s/api/sessions/%s", s.cfg.OpenWABaseURL, sessionID)

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}
	if s.cfg.OpenWAApiKey != "" {
		req.Header.Set("X-API-Key", s.cfg.OpenWAApiKey)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("delete session failed: %d %s", resp.StatusCode, string(body))
	}

	return nil
}

// DeleteAllSessions removes all existing sessions
func (s *OpenWAService) DeleteAllSessions() error {
	sessions, err := s.ListSessions()
	if err != nil {
		return err
	}
	for _, sess := range sessions {
		// Only delete sessions that start with "noant-" (our sessions)
		if len(sess.Name) > 6 && sess.Name[:6] == "noant-" {
			if err := s.DeleteSession(sess.ID); err != nil {
				s.logger.Warn("Failed to delete session", "id", sess.ID, "error", err)
			}
		}
	}
	return nil
}

// ListSessions returns all sessions (exported)
func (s *OpenWAService) ListSessions() ([]sessionInfo, error) {
	url := fmt.Sprintf("%s/api/sessions", s.cfg.OpenWABaseURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	if s.cfg.OpenWAApiKey != "" {
		req.Header.Set("X-API-Key", s.cfg.OpenWAApiKey)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var sessions []sessionInfo
	if err := json.NewDecoder(resp.Body).Decode(&sessions); err != nil {
		return nil, err
	}

	return sessions, nil
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
		s.logger.Error("Webhook config request failed", "error", err)
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	s.logger.Info("Webhook config response", "status", resp.StatusCode, "body", string(body))

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
		s.logger.Info("Webhook configured successfully", "sessionID", sessionID, "url", webhookURL)
		return nil
	}

	return fmt.Errorf("webhook config failed: %d %s", resp.StatusCode, string(body))
}
