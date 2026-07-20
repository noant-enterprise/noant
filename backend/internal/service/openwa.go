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
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"noant/config"
	"noant/internal/infrastructure"
)

// OpenWARateLimitHeaders tracks rate limit info from OpenWA responses
type OpenWARateLimitHeaders struct {
	mu        sync.Mutex
	remaining map[string]int // sessionID -> remaining
	resetAt   map[string]time.Time
}

func NewOpenWARateLimitHeaders() *OpenWARateLimitHeaders {
	return &OpenWARateLimitHeaders{
		remaining: make(map[string]int),
		resetAt:   make(map[string]time.Time),
	}
}

func (rlh *OpenWARateLimitHeaders) Update(sessionID string, remaining int, resetAt time.Time) {
	rlh.mu.Lock()
	defer rlh.mu.Unlock()
	rlh.remaining[sessionID] = remaining
	rlh.resetAt[sessionID] = resetAt
}

func (rlh *OpenWARateLimitHeaders) GetRemaining(sessionID string) (int, bool) {
	rlh.mu.Lock()
	defer rlh.mu.Unlock()
	r, ok := rlh.remaining[sessionID]
	return r, ok
}

func (rlh *OpenWARateLimitHeaders) ShouldBackoff(sessionID string) bool {
	rlh.mu.Lock()
	defer rlh.mu.Unlock()

	remaining, ok := rlh.remaining[sessionID]
	if !ok {
		return false
	}
	if remaining < 5 {
		if resetAt, ok := rlh.resetAt[sessionID]; ok {
			return time.Now().Before(resetAt)
		}
		return true
	}
	return false
}

// Circuit breaker for OpenWA API calls
type OpenWACircuitBreaker struct {
	mu          sync.Mutex
	failures    int
	lastFailure time.Time
	state       string // closed, open, half-open
}

func NewOpenWACircuitBreaker() *OpenWACircuitBreaker {
	return &OpenWACircuitBreaker{state: "closed"}
}

func (cb *OpenWACircuitBreaker) IsOpen() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if cb.state == "open" {
		if time.Since(cb.lastFailure) > 60*time.Second {
			cb.state = "half-open"
			return false
		}
		return true
	}
	return false
}

func (cb *OpenWACircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failures++
	cb.lastFailure = time.Now()
	if cb.failures >= 5 {
		cb.state = "open"
	}
}

func (cb *OpenWACircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failures = 0
	cb.state = "closed"
}

// ========== OpenWA SERVICE ==========

type OpenWAService struct {
	cfg              *config.Config
	httpClient       *http.Client
	logger           *infrastructure.Logger
	rateLimitHeaders *OpenWARateLimitHeaders
	circuitBreaker   *OpenWACircuitBreaker

	queue       *SendQueue
	rateLimiter *MessageRateLimiter
	workerPool  *WorkerPool
	sessionMgr  *SessionManager
	mediaHandler *MediaHandler
}

func NewOpenWAService(cfg *config.Config, logger *infrastructure.Logger) *OpenWAService {
	transport := &http.Transport{
		MaxIdleConns:        cfg.OpenWAConnPoolSize,
		MaxIdleConnsPerHost: cfg.OpenWAConnPoolSize,
		IdleConnTimeout:     90 * time.Second,
		DialContext: (&net.Dialer{
			Timeout:   cfg.OpenWAConnTimeout,
			KeepAlive: 30 * time.Second,
		}).DialContext,
	}

	svc := &OpenWAService{
		cfg: cfg,
		httpClient: &http.Client{
			Transport: transport,
			Timeout:   cfg.OpenWAReqTimeout,
		},
		logger:           logger,
		rateLimitHeaders: NewOpenWARateLimitHeaders(),
		circuitBreaker:   NewOpenWACircuitBreaker(),
	}

	// Initialize subsystems after OpenWAService creation
	svc.rateLimiter = NewMessageRateLimiter(cfg)
	svc.queue = NewSendQueue(cfg, nil, logger)
	svc.workerPool = NewWorkerPool(svc, svc.queue, svc.rateLimiter, logger)
	svc.sessionMgr = NewSessionManager(cfg, svc, nil, logger, svc.queue, svc.workerPool)
	svc.mediaHandler = NewMediaHandler(cfg, svc, nil, logger)

	return svc
}

// Accessor methods for handler layer

func (s *OpenWAService) GetQueueStats() map[string]interface{} {
	if s.queue == nil {
		return map[string]interface{}{"total": 0, "queued": 0}
	}
	return s.queue.Stats()
}

func (s *OpenWAService) ListManagedSessions() []map[string]interface{} {
	if s.sessionMgr == nil {
		return []map[string]interface{}{}
	}
	sessions := s.sessionMgr.ListSessions()
	result := make([]map[string]interface{}, len(sessions))
	for i, sh := range sessions {
		sh.mu.Lock()
		result[i] = map[string]interface{}{
			"session_id":           sh.SessionID,
			"state":                sh.State,
			"is_reconnecting":      sh.IsReconnecting,
			"reconnect_attempts":   sh.ReconnectAttempts,
			"consecutive_failures": sh.ConsecutiveFailures,
			"last_connected":       sh.LastConnected,
		}
		sh.mu.Unlock()
	}
	return result
}

func (s *OpenWAService) GetSessionMetrics(sessionID string) map[string]interface{} {
	if s.sessionMgr == nil {
		return map[string]interface{}{}
	}
	sh := s.sessionMgr.GetSession(sessionID)
	if sh == nil {
		return map[string]interface{}{"error": "session not found"}
	}
	sh.mu.Lock()
	defer sh.mu.Unlock()
	return map[string]interface{}{
		"session_id":           sh.SessionID,
		"state":                sh.State,
		"is_reconnecting":      sh.IsReconnecting,
		"reconnect_attempts":   sh.ReconnectAttempts,
		"consecutive_failures": sh.ConsecutiveFailures,
		"total_downtime":       sh.TotalDowntime.String(),
		"last_connected":       sh.LastConnected,
		"last_disconnected":    sh.LastDisconnected,
	}
}

func (s *OpenWAService) GetQueue() *SendQueue {
	return s.queue
}

func (s *OpenWAService) GetRateLimiter() *MessageRateLimiter {
	return s.rateLimiter
}

func (s *OpenWAService) GetWorkerPool() *WorkerPool {
	return s.workerPool
}

func (s *OpenWAService) GetSessionManager() *SessionManager {
	return s.sessionMgr
}

func (s *OpenWAService) GetMediaHandler() *MediaHandler {
	return s.mediaHandler
}

// StartSessionManager starts the background session health monitor
func (s *OpenWAService) StartSessionManager() {
	if s.sessionMgr != nil {
		s.sessionMgr.Start()
	}
}

// StopSessionManager stops the background session health monitor
func (s *OpenWAService) StopSessionManager() {
	if s.sessionMgr != nil {
		s.sessionMgr.Stop()
	}
}

// InjectDependencies sets Redis after initialization (since Redis is optional)
func (s *OpenWAService) InjectDependencies(redis *infrastructure.RedisClient) {
	if redis == nil {
		return
	}
	s.queue = NewSendQueue(s.cfg, redis, s.logger)
	s.sessionMgr = NewSessionManager(s.cfg, s, redis, s.logger, s.queue, s.workerPool)
	s.mediaHandler = NewMediaHandler(s.cfg, s, redis, s.logger)
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

// GetSessionStatus checks if the WhatsApp session is connected
func (s *OpenWAService) GetSessionStatus() (string, error) {
	if !s.cfg.OpenWAEnabled {
		return "disabled", nil
	}

	url := fmt.Sprintf("%s/api/sessions/%s",
		s.cfg.OpenWABaseURL, s.cfg.OpenWASessionID)

	req, err := http.NewRequest("GET", url, http.NoBody)
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
	defer func() { _ = resp.Body.Close() }()

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

	req, err := http.NewRequest("POST", url, http.NoBody)
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
	defer func() { _ = resp.Body.Close() }()

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

// ========== INTERNAL SEND METHODS (used by queue worker) ==========

// sendTextMessageInternal sends a text message and tracks rate limit headers
func (s *OpenWAService) sendTextMessageInternal(sessionID, chatID, text string) error {
	if s.circuitBreaker.IsOpen() {
		return fmt.Errorf("circuit breaker open for OpenWA API calls")
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
		s.circuitBreaker.RecordFailure()
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	// Track rate limit headers
	remaining := resp.Header.Get("X-RateLimit-Remaining")
	if remaining != "" {
		var rem int
		if _, err := fmt.Sscanf(remaining, "%d", &rem); err == nil {
			s.rateLimitHeaders.Update(sessionID, rem, time.Now().Add(1*time.Minute))
		}
	}

	if resp.StatusCode == http.StatusTooManyRequests {
		s.circuitBreaker.RecordFailure()
		return fmt.Errorf("rate limited by OpenWA (429)")
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("openwa returned status %d: %s", resp.StatusCode, string(body))
	}

	s.circuitBreaker.RecordSuccess()
	return nil
}

// sendMediaMessageInternal sends a media message internally (used by queue worker)
func (s *OpenWAService) sendMediaMessageInternal(sessionID, chatID, mediaURL, caption string) error {
	if s.circuitBreaker.IsOpen() {
		return fmt.Errorf("circuit breaker open for OpenWA API calls")
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
		s.circuitBreaker.RecordFailure()
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusTooManyRequests {
		s.circuitBreaker.RecordFailure()
		return fmt.Errorf("rate limited by OpenWA (429)")
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("openwa media send failed: %d %s", resp.StatusCode, string(body))
	}

	s.circuitBreaker.RecordSuccess()
	return nil
}

// sendTemplateMessageInternal sends a template message via OpenWA
func (s *OpenWAService) sendTemplateMessageInternal(sessionID, chatID string, params map[string]interface{}) error {
	if s.circuitBreaker.IsOpen() {
		return fmt.Errorf("circuit breaker open for OpenWA API calls")
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
		s.circuitBreaker.RecordFailure()
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusTooManyRequests {
		s.circuitBreaker.RecordFailure()
		return fmt.Errorf("rate limited by OpenWA (429)")
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("openwa request failed: %d %s", resp.StatusCode, string(body))
	}

	s.circuitBreaker.RecordSuccess()
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

// ========== SESSION MANAGEMENT ==========

type sessionInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Ping checks if OpenWA server is reachable
func (s *OpenWAService) Ping() error {
	url := fmt.Sprintf("%s/api/sessions", s.cfg.OpenWABaseURL)

	req, err := http.NewRequest("GET", url, http.NoBody)
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
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 500 {
		return fmt.Errorf("OpenWA server error: %d", resp.StatusCode)
	}

	return nil
}

// findSessionByName lists all sessions and finds the one matching the name
func (s *OpenWAService) findSessionByName(name string) (string, error) {
	url := fmt.Sprintf("%s/api/sessions", s.cfg.OpenWABaseURL)

	req, err := http.NewRequest("GET", url, http.NoBody)
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
	defer func() { _ = resp.Body.Close() }()

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
	defer func() { _ = resp.Body.Close() }()

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

	req, err := http.NewRequest("POST", url, http.NoBody)
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
	defer func() { _ = resp.Body.Close() }()

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

	req, err := http.NewRequest("GET", url, http.NoBody)
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
	defer func() { _ = resp.Body.Close() }()

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
	switch {
	case result.Image != "":
		rawQR = result.Image
	case result.QR != "":
		rawQR = result.QR
	case result.QRCode != "":
		rawQR = result.QRCode
	case result.Data != "":
		rawQR = result.Data
	case result.Base64 != "":
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
	_ = png.Encode(&buf, output)
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(buf.Bytes())
}

// GetSessionStatusByID gets the status of a specific session
func (s *OpenWAService) GetSessionStatusByID(sessionID string) (string, error) {
	url := fmt.Sprintf("%s/api/sessions/%s", s.cfg.OpenWABaseURL, sessionID)

	req, err := http.NewRequest("GET", url, http.NoBody)
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
	defer func() { _ = resp.Body.Close() }()

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
func (s *OpenWAService) CheckNumberExists(sessionID, phone string) (bool, error) {
	if !s.cfg.OpenWAEnabled {
		return false, nil
	}

	cleaned := CleanPhoneNumber(phone)
	url := fmt.Sprintf("%s/api/sessions/%s/contacts/check/%s", s.cfg.OpenWABaseURL, sessionID, cleaned)

	req, err := http.NewRequest("GET", url, http.NoBody)
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
	defer func() { _ = resp.Body.Close() }()

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

type OpenWAContact struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Pushname        string `json:"pushName"` // API returns camelCase pushName
	Number         string `json:"number"`
	ProfilePicUrl  string `json:"profilePicUrl"`
}

// GetContactInfo retrieves the contact information (pushname and avatar) from OpenWA.
// contactID must be in the format: number@c.us (use FormatContactID to convert a phone number).
func (s *OpenWAService) GetContactInfo(sessionID, contactID string) (*OpenWAContact, error) {
	if !s.cfg.OpenWAEnabled {
		return nil, fmt.Errorf("OpenWA is disabled")
	}

	url := fmt.Sprintf("%s/api/sessions/%s/contacts/%s", s.cfg.OpenWABaseURL, sessionID, contactID)

	req, err := http.NewRequest("GET", url, http.NoBody)
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
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get contact info: %d %s", resp.StatusCode, string(body))
	}

	var contact OpenWAContact
	if err := json.NewDecoder(resp.Body).Decode(&contact); err != nil {
		return nil, err
	}

	// If no profilePicUrl in contact response, try the dedicated profile-picture endpoint
	if contact.ProfilePicUrl == "" {
		picURL := fmt.Sprintf("%s/api/sessions/%s/contacts/%s/profile-picture", s.cfg.OpenWABaseURL, sessionID, contactID)
		picReq, err2 := http.NewRequest("GET", picURL, http.NoBody)
		if err2 == nil {
			if s.cfg.OpenWAApiKey != "" {
				picReq.Header.Set("X-API-Key", s.cfg.OpenWAApiKey)
			}
			picResp, err2 := s.httpClient.Do(picReq)
			if err2 == nil {
				defer func() { _ = picResp.Body.Close() }()
				if picResp.StatusCode == http.StatusOK {
					var picResult struct {
						URL string `json:"url"`
					}
					if json.NewDecoder(picResp.Body).Decode(&picResult) == nil && picResult.URL != "" {
						contact.ProfilePicUrl = picResult.URL
					}
				}
			}
		}
	}

	return &contact, nil
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

	req, err := http.NewRequest("POST", url, http.NoBody)
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
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("restart failed: %d %s", resp.StatusCode, string(body))
	}

	return nil
}

// DeleteSession deletes an OpenWA session
func (s *OpenWAService) DeleteSession(sessionID string) error {
	url := fmt.Sprintf("%s/api/sessions/%s", s.cfg.OpenWABaseURL, sessionID)

	req, err := http.NewRequest("DELETE", url, http.NoBody)
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
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("delete session failed: %d %s", resp.StatusCode, string(body))
	}

	return nil
}

// LogoutSession logs out the active WhatsApp session to clear credentials
func (s *OpenWAService) LogoutSession(sessionID string) error {
	url := fmt.Sprintf("%s/api/sessions/%s/logout", s.cfg.OpenWABaseURL, sessionID)

	req, err := http.NewRequest("POST", url, http.NoBody)
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
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusNotFound {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("logout session failed: %d %s", resp.StatusCode, string(body))
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

	req, err := http.NewRequest("GET", url, http.NoBody)
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
	defer func() { _ = resp.Body.Close() }()

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
