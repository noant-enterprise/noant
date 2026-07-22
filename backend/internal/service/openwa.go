package service

import (
	"context"
	"net"
	"net/http"
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

// ========== FIX 5: PER-SESSION CIRCUIT BREAKER ==========

type SessionCircuitBreaker struct {
	mu          sync.Mutex
	failures    int
	lastFailure time.Time
	state       string // closed, open, half-open
}

func NewSessionCircuitBreaker() *SessionCircuitBreaker {
	return &SessionCircuitBreaker{state: "closed"}
}

func (cb *SessionCircuitBreaker) IsOpen() bool {
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

func (cb *SessionCircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failures++
	cb.lastFailure = time.Now()
	if cb.failures >= 5 {
		cb.state = "open"
	}
}

func (cb *SessionCircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failures = 0
	cb.state = "closed"
}

// CircuitBreakerManager manages per-session circuit breakers
type CircuitBreakerManager struct {
	mu       sync.Mutex
	breakers map[string]*SessionCircuitBreaker
}

func NewCircuitBreakerManager() *CircuitBreakerManager {
	return &CircuitBreakerManager{
		breakers: make(map[string]*SessionCircuitBreaker),
	}
}

func (m *CircuitBreakerManager) Get(sessionID string) *SessionCircuitBreaker {
	m.mu.Lock()
	defer m.mu.Unlock()

	cb, ok := m.breakers[sessionID]
	if !ok {
		cb = NewSessionCircuitBreaker()
		m.breakers[sessionID] = cb
	}
	return cb
}

func (m *CircuitBreakerManager) IsOpen(sessionID string) bool {
	return m.Get(sessionID).IsOpen()
}

func (m *CircuitBreakerManager) RecordFailure(sessionID string) {
	m.Get(sessionID).RecordFailure()
}

func (m *CircuitBreakerManager) RecordSuccess(sessionID string) {
	m.Get(sessionID).RecordSuccess()
}

// ========== OpenWA SERVICE ==========

type OpenWAService struct {
	cfg              *config.Config
	httpClient       *http.Client
	logger           *infrastructure.Logger
	rateLimitHeaders *OpenWARateLimitHeaders
	cbManager        *CircuitBreakerManager

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
		cbManager:        NewCircuitBreakerManager(),
	}

	svc.rateLimiter = NewMessageRateLimiter(cfg)
	svc.queue = NewSendQueue(cfg, nil, logger)
	svc.workerPool = NewWorkerPool(svc, svc.queue, svc.rateLimiter, logger)
	svc.sessionMgr = NewSessionManager(cfg, svc, nil, logger, svc.queue, svc.workerPool)
	svc.mediaHandler = NewMediaHandler(cfg, svc, nil, logger)

	// Wire up message completion callback for session rotation tracking
	svc.queue.OnComplete = func(entry *QueueEntry) {
		if svc.sessionMgr != nil {
			svc.sessionMgr.RecordMessageSent(entry.SessionID)
		}
	}

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

func (s *OpenWAService) ShouldBackoff(sessionID string) bool {
	return s.rateLimitHeaders.ShouldBackoff(sessionID)
}

func (s *OpenWAService) StartSessionManager() {
	if s.sessionMgr != nil {
		s.sessionMgr.Start()
	}
}

func (s *OpenWAService) StopSessionManager() {
	if s.sessionMgr != nil {
		s.sessionMgr.Stop()
	}
}

func (s *OpenWAService) InjectDependencies(redis *infrastructure.RedisClient) {
	if redis == nil {
		return
	}
	s.queue = NewSendQueue(s.cfg, redis, s.logger)
	s.queue.OnComplete = func(entry *QueueEntry) {
		if s.sessionMgr != nil {
			s.sessionMgr.RecordMessageSent(entry.SessionID)
		}
	}
	s.sessionMgr = NewSessionManager(s.cfg, s, redis, s.logger, s.queue, s.workerPool)
	s.mediaHandler = NewMediaHandler(s.cfg, s, redis, s.logger)
}

// Shutdown performs a graceful drain of the queue before exiting (Fix 1)
func (s *OpenWAService) Shutdown(ctx context.Context) {
	s.logger.Info("OpenWA: starting graceful shutdown")

	// Stop session manager first
	s.StopSessionManager()

	// Stop workers (they stop dequeuing)
	if s.workerPool != nil {
		s.workerPool.StopAll()
	}

	// Drain remaining queued messages
	if s.queue != nil {
		drained := s.queue.DrainQueue(ctx, func(entry *QueueEntry) error {
			return s.sendTextMessageInternal(entry.SessionID, entry.ChatID, entry.Content)
		})
		s.logger.Info("OpenWA: queue drained", "sent", drained, "remaining", s.queue.Depth())

		// Stop the queue eviction loop
		s.queue.Stop()
	}
}
