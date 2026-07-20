package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"noant/config"
	"noant/internal/infrastructure"
)

// ========== SESSION STATE ==========

const (
	SessionStateConnected    = "connected"
	SessionStateConnecting   = "connecting"
	SessionStateQRReady     = "qr_ready"
	SessionStateDisconnected = "disconnected"
	SessionStateExpired     = "expired"
	SessionStateFailed      = "failed"
	SessionStateUnknown     = "unknown"

	MaxConsecutiveFailures = 3
)

// SessionHealth holds runtime health data for a WhatsApp session
type SessionHealth struct {
	SessionID          string
	UserID             string
	State              string
	ConsecutiveFailures int
	ReconnectAttempts   int
	LastConnected      time.Time
	LastDisconnected   time.Time
	TotalDowntime      time.Duration
	IsReconnecting     bool
	StopReconnect      chan struct{}
	mu                 sync.Mutex
}

type SessionManager struct {
	cfg        *config.Config
	openwa     *OpenWAService
	redis      *infrastructure.RedisClient
	logger     *infrastructure.Logger
	queue      *SendQueue
	workerPool *WorkerPool

	mu       sync.Mutex
	sessions map[string]*SessionHealth

	stopCh chan struct{}
	doneCh chan struct{}
}

func NewSessionManager(cfg *config.Config, openwa *OpenWAService, redis *infrastructure.RedisClient, logger *infrastructure.Logger, queue *SendQueue, workerPool *WorkerPool) *SessionManager {
	return &SessionManager{
		cfg:        cfg,
		openwa:     openwa,
		redis:      redis,
		logger:     logger,
		queue:      queue,
		workerPool: workerPool,
		sessions:   make(map[string]*SessionHealth),
		stopCh:     make(chan struct{}),
		doneCh:     make(chan struct{}),
	}
}

func (sm *SessionManager) Start() {
	go sm.healthLoop()
	sm.logger.Info("Session health monitor started")
}

func (sm *SessionManager) Stop() {
	close(sm.stopCh)
	<-sm.doneCh
	sm.logger.Info("Session health monitor stopped")
}

func (sm *SessionManager) RegisterSession(sessionID, userID string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if _, exists := sm.sessions[sessionID]; exists {
		return
	}

	sm.sessions[sessionID] = &SessionHealth{
		SessionID:     sessionID,
		UserID:        userID,
		State:         SessionStateUnknown,
		StopReconnect: make(chan struct{}),
	}
	sm.workerPool.EnsureWorker(sessionID)
	sm.logger.Info("Session registered for health monitoring", "sessionID", sessionID, "userID", userID)
}

func (sm *SessionManager) UnregisterSession(sessionID string) {
	sm.mu.Lock()
	sh, exists := sm.sessions[sessionID]
	if exists {
		close(sh.StopReconnect)
		delete(sm.sessions, sessionID)
	}
	sm.mu.Unlock()

	if exists {
		sm.workerPool.StopWorker(sessionID)
		sm.logger.Info("Session unregistered from health monitoring", "sessionID", sessionID)
	}
}

func (sm *SessionManager) GetSession(sessionID string) *SessionHealth {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	return sm.sessions[sessionID]
}

func (sm *SessionManager) GetSessionByUserID(userID string) *SessionHealth {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	for _, sh := range sm.sessions {
		if sh.UserID == userID {
			return sh
		}
	}
	return nil
}

func (sm *SessionManager) ListSessions() []*SessionHealth {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	result := make([]*SessionHealth, 0, len(sm.sessions))
	for _, sh := range sm.sessions {
		result = append(result, sh)
	}
	return result
}

func (sm *SessionManager) UpdateState(sessionID, state string) {
	sm.mu.Lock()
	sh, exists := sm.sessions[sessionID]
	sm.mu.Unlock()

	if !exists {
		return
	}

	sh.mu.Lock()
	prevState := sh.State
	sh.State = state

	now := time.Now()
	switch state {
	case SessionStateConnected:
		sh.ConsecutiveFailures = 0
		sh.ReconnectAttempts = 0
		sh.IsReconnecting = false
		if prevState != SessionStateConnected {
			sh.LastConnected = now
			sm.logger.Info("Session connected", "sessionID", sessionID)
			sm.storeSessionQR(sessionID, "")
		}
	case SessionStateDisconnected, SessionStateFailed, SessionStateExpired:
		sh.ConsecutiveFailures++
		if prevState == SessionStateConnected {
			sh.LastDisconnected = now
			if !sh.LastConnected.IsZero() {
				sh.TotalDowntime += now.Sub(sh.LastConnected)
			}
		}
	}
	sh.mu.Unlock()
}

func (sm *SessionManager) healthLoop() {
	ticker := time.NewTicker(sm.cfg.OpenWASessionHealthInterval)
	defer ticker.Stop()
	defer close(sm.doneCh)

	for {
		select {
		case <-sm.stopCh:
			return
		case <-ticker.C:
			sm.checkAllSessions()
		}
	}
}

func (sm *SessionManager) checkAllSessions() {
	sm.mu.Lock()
	sessions := make([]*SessionHealth, 0, len(sm.sessions))
	for _, sh := range sm.sessions {
		sessions = append(sessions, sh)
	}
	sm.mu.Unlock()

	for _, sh := range sessions {
		sm.checkSession(sh)
	}
}

func (sm *SessionManager) checkSession(sh *SessionHealth) {
	sh.mu.Lock()
	if sh.IsReconnecting {
		sh.mu.Unlock()
		return
	}
	sh.mu.Unlock()

	status, err := sm.openwa.GetSessionStatusByID(sh.SessionID)
	if err != nil {
		sm.logger.Error("Session health check failed", "sessionID", sh.SessionID, "error", err)
		sm.UpdateState(sh.SessionID, SessionStateDisconnected)
		return
	}

	normalized := normalizeSessionStatus(status)
	sm.UpdateState(sh.SessionID, normalized)

	// Check if reconnection is needed
	sh.mu.Lock()
	consecutiveFails := sh.ConsecutiveFailures
	shouldReconnect := consecutiveFails >= MaxConsecutiveFailures && normalized != SessionStateConnected
	sh.mu.Unlock()

	if shouldReconnect {
		sm.triggerReconnect(sh.SessionID)
	}

	// Verify webhook on health check
	if normalized == SessionStateConnected {
		sm.verifyWebhook(sh.SessionID, sh.UserID)
	}
}

func (sm *SessionManager) triggerReconnect(sessionID string) {
	sm.mu.Lock()
	sh, exists := sm.sessions[sessionID]
	if !exists {
		sm.mu.Unlock()
		return
	}

	if sh.IsReconnecting {
		sm.mu.Unlock()
		return
	}
	sh.IsReconnecting = true
	sh.StopReconnect = make(chan struct{})
	sm.mu.Unlock()

	go sm.reconnectLoop(sh)
}

func (sm *SessionManager) reconnectLoop(sh *SessionHealth) {
	backoff := []time.Duration{
		30 * time.Second,
		1 * time.Minute,
		2 * time.Minute,
		5 * time.Minute,
		10 * time.Minute,
		30 * time.Minute,
	}

	maxAttempts := sm.cfg.OpenWAMaxReconnectAttempts
	attempt := 0

	for attempt < maxAttempts {
		select {
		case <-sh.StopReconnect:
			sm.logger.Info("Reconnect canceled", "sessionID", sh.SessionID)
			return
		default:
		}

		sm.logger.Info("Attempting session reconnect", "sessionID", sh.SessionID, "attempt", attempt+1)
		err := sm.openwa.RestartSessionByID(sh.SessionID)
		if err != nil {
			sm.logger.Warn("Reconnect attempt failed", "sessionID", sh.SessionID, "attempt", attempt+1, "error", err)

			sh.mu.Lock()
			sh.ReconnectAttempts++
			sh.mu.Unlock()

			attempt++
			if attempt >= maxAttempts {
				break
			}

			delay := backoff[min(attempt-1, len(backoff)-1)]
			select {
			case <-sh.StopReconnect:
				return
			case <-time.After(delay):
			}
			continue
		}

		// Success — wait briefly for state to update then verify
		time.Sleep(5 * time.Second)

		status, err := sm.openwa.GetSessionStatusByID(sh.SessionID)
		if err == nil && normalizeSessionStatus(status) == SessionStateConnected {
			sm.UpdateState(sh.SessionID, SessionStateConnected)
			sm.logger.Info("Session reconnected successfully", "sessionID", sh.SessionID, "attempts", attempt+1)

			// Send admin notification
			sm.notifyAdmin(sh.UserID, "WhatsApp session reconnected", fmt.Sprintf("Session %s reconnected after %d attempts", sh.SessionID, attempt+1))
			return
		}

		// Connected but not yet confirmed — check for QR regeneration
		qr, _ := sm.openwa.GetQRCode(sh.SessionID)
		if qr != "" {
			sm.storeSessionQR(sh.SessionID, qr)
			sm.notifyAdmin(sh.UserID, "WhatsApp re-authentication required", fmt.Sprintf("Session %s needs QR scan", sh.SessionID))
		}

		attempt++
		if attempt >= maxAttempts {
			break
		}

		delay := backoff[min(attempt-1, len(backoff)-1)]
		select {
		case <-sh.StopReconnect:
			return
		case <-time.After(delay):
		}
	}

	// All attempts exhausted
	sm.mu.Lock()
	sh.IsReconnecting = false
	sm.mu.Unlock()

	sm.logger.Error("Session reconnect exhausted all attempts", "sessionID", sh.SessionID, "maxAttempts", maxAttempts)
	sm.notifyAdmin(sh.UserID, "WhatsApp session disconnected", fmt.Sprintf("Session %s failed to reconnect after %d attempts. Manual intervention required.", sh.SessionID, maxAttempts))
}

func (sm *SessionManager) storeSessionQR(sessionID, qrData string) {
	if sm.redis == nil || qrData == "" {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = sm.redis.SetEx(ctx, fmt.Sprintf("openwa:qr:%s", sessionID), qrData, 5*time.Minute)
}

func (sm *SessionManager) GetStoredQR(sessionID string) string {
	if sm.redis == nil {
		return ""
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	data, err := sm.redis.Get(ctx, fmt.Sprintf("openwa:qr:%s", sessionID))
	if err != nil {
		return ""
	}
	return data
}

func (sm *SessionManager) verifyWebhook(sessionID, userID string) {
	webhookURL := fmt.Sprintf("%s/api/v1/openwa/webhook", sm.cfg.APIURL)
	if err := sm.openwa.ConfigureWebhook(sessionID, webhookURL, sm.cfg.OpenWAWebhookSecret); err != nil {
		sm.logger.Warn("Webhook verification failed, will retry", "sessionID", sessionID, "error", err)
	}
}

func (sm *SessionManager) notifyAdmin(userID, title, body string) {
	sm.logger.Info("Admin notification", "userID", userID, "title", title, "body", body)
}

func (sm *SessionManager) Stats() map[string]interface{} {
	sessions := sm.ListSessions()
	var connected, disconnected, reconnecting, failed int
	for _, sh := range sessions {
		sh.mu.Lock()
		switch sh.State {
		case SessionStateConnected:
			connected++
		case SessionStateDisconnected:
			disconnected++
		case SessionStateFailed:
			failed++
		}
		if sh.IsReconnecting {
			reconnecting++
		}
		sh.mu.Unlock()
	}

	return map[string]interface{}{
		"total_sessions":      len(sessions),
		"connected":           connected,
		"disconnected":        disconnected,
		"failed":              failed,
		"reconnecting":        reconnecting,
		"queue_depth":         sm.queue.Depth(),
	}
}

// ========== SESSION METRICS ==========

type SessionMetrics struct {
	mu              sync.Mutex
	messagesSent    map[string]int64   // sessionID -> count
	messagesFailed  map[string]int64
	totalDowntime   map[string]time.Duration
	reconnectCount  map[string]int64
	downtimeStart   map[string]time.Time
}

func NewSessionMetrics() *SessionMetrics {
	return &SessionMetrics{
		messagesSent:   make(map[string]int64),
		messagesFailed: make(map[string]int64),
		totalDowntime:  make(map[string]time.Duration),
		reconnectCount: make(map[string]int64),
		downtimeStart:  make(map[string]time.Time),
	}
}

func (sm *SessionMetrics) RecordSent(sessionID string) {
	sm.mu.Lock()
	sm.messagesSent[sessionID]++
	sm.mu.Unlock()
}

func (sm *SessionMetrics) RecordFailed(sessionID string) {
	sm.mu.Lock()
	sm.messagesFailed[sessionID]++
	sm.mu.Unlock()
}

func (sm *SessionMetrics) RecordDowntimeStart(sessionID string) {
	sm.mu.Lock()
	sm.downtimeStart[sessionID] = time.Now()
	sm.mu.Unlock()
}

func (sm *SessionMetrics) RecordDowntimeEnd(sessionID string) {
	sm.mu.Lock()
	if start, ok := sm.downtimeStart[sessionID]; ok {
		sm.totalDowntime[sessionID] += time.Since(start)
		delete(sm.downtimeStart, sessionID)
	}
	sm.mu.Unlock()
}

func (sm *SessionMetrics) RecordReconnect(sessionID string) {
	sm.mu.Lock()
	sm.reconnectCount[sessionID]++
	sm.mu.Unlock()
}

func (sm *SessionMetrics) GetMetrics(sessionID string) map[string]interface{} {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	dt := sm.totalDowntime[sessionID]
	if start, ok := sm.downtimeStart[sessionID]; ok {
		dt += time.Since(start)
	}
	return map[string]interface{}{
		"messages_sent":   sm.messagesSent[sessionID],
		"messages_failed": sm.messagesFailed[sessionID],
		"total_downtime":  dt.String(),
		"reconnects":      sm.reconnectCount[sessionID],
	}
}

type QRRawMessage json.RawMessage
