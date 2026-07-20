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

// Ensure min from parent package is used
// (defined in service.go)

// ========== MESSAGE TYPES ==========

const (
	MsgTypeText     = "text"
	MsgTypeMedia    = "media"
	MsgTypeTemplate = "template"

	PriorityUrgent = 0
	PriorityNormal = 1
	PriorityBulk   = 2

	QueueStatusQueued     = "queued"
	QueueStatusSending    = "sending"
	QueueStatusSent       = "sent"
	QueueStatusFailed     = "failed"
	QueueStatusDeadLetter = "dead_letter"

	MaxRetries = 5
)

// retryDelays in seconds for each retry attempt
var retryDelays = []time.Duration{
	0,           // attempt 1: immediate
	30 * time.Second,  // attempt 2
	2 * time.Minute,   // attempt 3
	5 * time.Minute,   // attempt 4
	15 * time.Minute,  // attempt 5
}

// ========== QUEUE ENTRY ==========

type QueueEntry struct {
	ID         string     `json:"id"`
	SessionID  string     `json:"session_id"`
	UserID     string     `json:"user_id"`
	ChatID     string     `json:"chat_id"`
	MsgType    string     `json:"msg_type"`
	Content    string     `json:"content"`
	MediaURL   string     `json:"media_url,omitempty"`
	Caption    string     `json:"caption,omitempty"`
	Priority   int        `json:"priority"`
	Status     string     `json:"status"`
	RetryCount int        `json:"retry_count"`
	LastError  string     `json:"last_error,omitempty"`
	NextRetry  *time.Time `json:"next_retry,omitempty"`
	ScheduledAt time.Time `json:"scheduled_at"`
	CreatedAt  time.Time  `json:"created_at"`
}

// ========== RATE LIMITER ==========

type MessageRateLimiter struct {
	mu         sync.Mutex
	windows    map[string][]time.Time // sessionID -> timestamps
	textLimit  int
	mediaLimit int
	tplLimit   int
	burstLimit int
}

func NewMessageRateLimiter(cfg *config.Config) *MessageRateLimiter {
	return &MessageRateLimiter{
		windows:    make(map[string][]time.Time),
		textLimit:  cfg.OpenWARateLimitText,
		mediaLimit: cfg.OpenWARateLimitMedia,
		tplLimit:   cfg.OpenWARateLimitTemplate,
		burstLimit: cfg.OpenWARateLimitBurst,
	}
}

func (rl *MessageRateLimiter) Allow(sessionID, msgType string) (allowed bool, retryAfter int) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-1 * time.Minute)

	var limit int
	switch msgType {
	case MsgTypeText:
		limit = rl.textLimit
	case MsgTypeMedia:
		limit = rl.mediaLimit
	case MsgTypeTemplate:
		limit = rl.tplLimit
	default:
		limit = rl.textLimit
	}

	timestamps := rl.windows[sessionID]

	// Prune timestamps older than 1 minute
	var active []time.Time
	for _, ts := range timestamps {
		if ts.After(windowStart) {
			active = append(active, ts)
		}
	}

	// Burst check: allow burstLimit rapid messages even if at limit
	if len(active) >= limit {
		if len(active) > 0 {
			oldestInWindow := active[0]
			if now.Sub(oldestInWindow) < 5*time.Second && len(active) < limit+rl.burstLimit {
				active = append(active, now)
				rl.windows[sessionID] = active
				return true, limit - len(active) + 1
			}
		}
		return false, limit - len(active)
	}

	active = append(active, now)
	rl.windows[sessionID] = active
	return true, limit - len(active)
}

// ========== SEND QUEUE ==========

type SendQueue struct {
	mu       sync.Mutex
	entries  []*QueueEntry
	bySession map[string][]*QueueEntry // sessionID -> entries for quick lookup
	redis    *infrastructure.RedisClient
	logger   *infrastructure.Logger
	cfg      *config.Config
}

func NewSendQueue(cfg *config.Config, redis *infrastructure.RedisClient, logger *infrastructure.Logger) *SendQueue {
	sq := &SendQueue{
		entries:   make([]*QueueEntry, 0),
		bySession: make(map[string][]*QueueEntry),
		redis:     redis,
		logger:    logger,
		cfg:       cfg,
	}
	if redis != nil {
		sq.loadFromRedis()
	}
	return sq
}

func (sq *SendQueue) loadFromRedis() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Use SCAN to find all queue keys
	iter := sq.redis.Scan(ctx, "openwa:queue:*", 100)
	if iter == nil {
		sq.logger.Warn("Redis SCAN not available, skipping queue load")
		return
	}

	var count int
	for iter.Next(ctx) {
		key := iter.Val()
		data, err := sq.redis.Get(ctx, key)
		if err != nil {
			continue
		}
		var entry QueueEntry
		if err := json.Unmarshal([]byte(data), &entry); err != nil {
			continue
		}
		if entry.Status == QueueStatusQueued || entry.Status == QueueStatusFailed {
			sq.mu.Lock()
			sq.entries = append(sq.entries, &entry)
			sq.bySession[entry.SessionID] = append(sq.bySession[entry.SessionID], &entry)
			sq.mu.Unlock()
			count++
		}
	}
	sq.logger.Info("Loaded queue entries from Redis", "count", count)
}

func (sq *SendQueue) persist(ctx context.Context, entry *QueueEntry) {
	if sq.redis == nil {
		return
	}
	data, err := json.Marshal(entry)
	if err != nil {
		return
	}
	key := fmt.Sprintf("openwa:queue:%s", entry.ID)
	_ = sq.redis.SetEx(ctx, key, string(data), 72*time.Hour)
}

func (sq *SendQueue) removeFromRedis(ctx context.Context, entryID string) {
	if sq.redis == nil {
		return
	}
	_ = sq.redis.Delete(ctx, fmt.Sprintf("openwa:queue:%s", entryID))
}

// Enqueue adds a message to the queue with FIFO ordering within priority
func (sq *SendQueue) Enqueue(entry *QueueEntry) error {
	sq.mu.Lock()
	defer sq.mu.Unlock()

	// Check queue depth per session
	if len(sq.bySession[entry.SessionID]) >= sq.cfg.OpenWAQueueDepth {
		return fmt.Errorf("queue depth exceeded for session %s: max %d", entry.SessionID, sq.cfg.OpenWAQueueDepth)
	}

	// Insert in priority order (lower number = higher priority)
	// Within same priority, FIFO (append)
	insertIdx := len(sq.entries)
	for i, e := range sq.entries {
		if e.Priority > entry.Priority {
			insertIdx = i
			break
		}
	}

	if insertIdx >= len(sq.entries) {
		sq.entries = append(sq.entries, entry)
	} else {
		sq.entries = append(sq.entries, nil)
		copy(sq.entries[insertIdx+1:], sq.entries[insertIdx:])
		sq.entries[insertIdx] = entry
	}

	sq.bySession[entry.SessionID] = append(sq.bySession[entry.SessionID], entry)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	sq.persist(ctx, entry)

	sq.logger.Debug("Message queued", "id", entry.ID, "session", entry.SessionID, "type", entry.MsgType, "priority", entry.Priority, "queueDepth", len(sq.entries))
	return nil
}

// Dequeue retrieves the next ready message (respects next_retry timing)
func (sq *SendQueue) Dequeue() *QueueEntry {
	sq.mu.Lock()
	defer sq.mu.Unlock()

	now := time.Now()
	for i, entry := range sq.entries {
		if entry.Status == QueueStatusQueued || entry.Status == QueueStatusFailed {
			if entry.NextRetry == nil || entry.NextRetry.Before(now) {
				entry.Status = QueueStatusSending
				return entry
			}
		}
		_ = i
	}
	return nil
}

// DequeueBySession retrieves next ready message for a specific session
func (sq *SendQueue) DequeueBySession(sessionID string) *QueueEntry {
	sq.mu.Lock()
	defer sq.mu.Unlock()

	now := time.Now()
	for _, entry := range sq.bySession[sessionID] {
		if entry.Status == QueueStatusQueued || entry.Status == QueueStatusFailed {
			if entry.NextRetry == nil || entry.NextRetry.Before(now) {
				entry.Status = QueueStatusSending
				return entry
			}
		}
	}
	return nil
}

// Complete marks a message as sent and removes it from the queue
func (sq *SendQueue) Complete(entryID string) {
	sq.mu.Lock()
	defer sq.mu.Unlock()

	for i, entry := range sq.entries {
		if entry.ID == entryID {
			entry.Status = QueueStatusSent
			sq.removeEntry(i, entry)
			return
		}
	}
}

// Fail marks a message as failed and schedules retry or dead letter
func (sq *SendQueue) Fail(entryID string, err error) {
	sq.mu.Lock()
	defer sq.mu.Unlock()

	for i, entry := range sq.entries {
		if entry.ID != entryID {
			continue
		}
		entry.RetryCount++
		entry.LastError = err.Error()

		if entry.RetryCount > MaxRetries {
			entry.Status = QueueStatusDeadLetter
			sq.logger.Error("Message moved to dead letter queue", "id", entryID, "session", entry.SessionID, "error", entry.LastError)
			sq.removeEntry(i, entry)
			return
		}

		entry.Status = QueueStatusFailed
		delay := retryDelays[min(entry.RetryCount, len(retryDelays)-1)]
		nextRetry := time.Now().Add(delay)
		entry.NextRetry = &nextRetry

		sq.logger.Warn("Message queued for retry", "id", entryID, "attempt", entry.RetryCount, "nextRetry", nextRetry)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		sq.persist(ctx, entry)
		cancel()
		return
	}
}

func (sq *SendQueue) removeEntry(idx int, entry *QueueEntry) {
	sq.entries = append(sq.entries[:idx], sq.entries[idx+1:]...)

	// Remove from bySession
	sessionEntries := sq.bySession[entry.SessionID]
	for j, se := range sessionEntries {
		if se.ID == entry.ID {
			sq.bySession[entry.SessionID] = append(sessionEntries[:j], sessionEntries[j+1:]...)
			break
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	sq.removeFromRedis(ctx, entry.ID)
}

func (sq *SendQueue) Depth() int {
	sq.mu.Lock()
	defer sq.mu.Unlock()
	return len(sq.entries)
}

func (sq *SendQueue) DepthBySession(sessionID string) int {
	sq.mu.Lock()
	defer sq.mu.Unlock()
	return len(sq.bySession[sessionID])
}

func (sq *SendQueue) Stats() map[string]interface{} {
	sq.mu.Lock()
	defer sq.mu.Unlock()

	var sent, failed, queued, deadLetter int
	for _, e := range sq.entries {
		switch e.Status {
		case QueueStatusSent:
			sent++
		case QueueStatusFailed, QueueStatusQueued:
			queued++
		case QueueStatusDeadLetter:
			deadLetter++
		}
	}
	infrastructure.OpenWAQueueDepth.Set(float64(queued))
	return map[string]interface{}{
		"total":       len(sq.entries),
		"queued":      queued,
		"sent":        sent,
		"failed":      failed,
		"dead_letter": deadLetter,
	}
}

// ========== SEND WORKER ==========

type SessionWorker struct {
	sessionID  string
	openwa     *OpenWAService
	queue      *SendQueue
	rateLimiter *MessageRateLimiter
	logger     *infrastructure.Logger
	stopCh     chan struct{}
	doneCh     chan struct{}
}

func NewSessionWorker(sessionID string, openwa *OpenWAService, queue *SendQueue, rateLimiter *MessageRateLimiter, logger *infrastructure.Logger) *SessionWorker {
	return &SessionWorker{
		sessionID:   sessionID,
		openwa:      openwa,
		queue:       queue,
		rateLimiter: rateLimiter,
		logger:      logger,
		stopCh:      make(chan struct{}),
		doneCh:      make(chan struct{}),
	}
}

func (w *SessionWorker) Start() {
	go w.run()
}

func (w *SessionWorker) Stop() {
	close(w.stopCh)
	<-w.doneCh
}

func (w *SessionWorker) run() {
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()
	defer close(w.doneCh)

	for {
		select {
		case <-w.stopCh:
			return
		case <-ticker.C:
			w.processNext()
		}
	}
}

func (w *SessionWorker) processNext() {
	if w.queue == nil {
		return
	}
	entry := w.queue.DequeueBySession(w.sessionID)
	if entry == nil {
		return
	}

	// Rate limit check
	allowed, remaining := w.rateLimiter.Allow(w.sessionID, entry.MsgType)
	if !allowed {
		// Put it back as queued
		entry.Status = QueueStatusQueued
		w.logger.Warn("Rate limited, re-queuing", "id", entry.ID, "session", w.sessionID, "type", entry.MsgType)
		return
	}

	w.logger.Debug("Sending message", "id", entry.ID, "session", w.sessionID, "type", entry.MsgType, "rateLimitRemaining", remaining)

	var sendErr error
	switch entry.MsgType {
	case MsgTypeText:
		sendErr = w.openwa.sendTextMessageInternal(w.sessionID, entry.ChatID, entry.Content)
	case MsgTypeMedia:
		sendErr = w.openwa.sendMediaMessageInternal(w.sessionID, entry.ChatID, entry.MediaURL, entry.Caption)
	case MsgTypeTemplate:
		var params map[string]interface{}
		if err := json.Unmarshal([]byte(entry.Content), &params); err == nil {
			sendErr = w.openwa.sendTemplateMessageInternal(w.sessionID, entry.ChatID, params)
		} else {
			sendErr = fmt.Errorf("invalid template params: %w", err)
		}
	default:
		sendErr = w.openwa.sendTextMessageInternal(w.sessionID, entry.ChatID, entry.Content)
	}

	if sendErr != nil {
		w.queue.Fail(entry.ID, sendErr)
	} else {
		w.queue.Complete(entry.ID)
	}
}

// ========== WORKER POOL MANAGER ==========

type WorkerPool struct {
	mu          sync.Mutex
	workers     map[string]*SessionWorker
	openwa      *OpenWAService
	queue       *SendQueue
	rateLimiter *MessageRateLimiter
	logger      *infrastructure.Logger
}

func NewWorkerPool(openwa *OpenWAService, queue *SendQueue, rateLimiter *MessageRateLimiter, logger *infrastructure.Logger) *WorkerPool {
	return &WorkerPool{
		workers:     make(map[string]*SessionWorker),
		openwa:      openwa,
		queue:       queue,
		rateLimiter: rateLimiter,
		logger:      logger,
	}
}

func (wp *WorkerPool) EnsureWorker(sessionID string) {
	wp.mu.Lock()
	defer wp.mu.Unlock()

	if _, exists := wp.workers[sessionID]; exists {
		return
	}

	w := NewSessionWorker(sessionID, wp.openwa, wp.queue, wp.rateLimiter, wp.logger)
	w.Start()
	wp.workers[sessionID] = w
	wp.logger.Info("Worker started for session", "sessionID", sessionID)
}

func (wp *WorkerPool) StopWorker(sessionID string) {
	wp.mu.Lock()
	w, exists := wp.workers[sessionID]
	if exists {
		delete(wp.workers, sessionID)
	}
	wp.mu.Unlock()

	if exists {
		w.Stop()
		wp.logger.Info("Worker stopped for session", "sessionID", sessionID)
	}
}

func (wp *WorkerPool) StopAll() {
	wp.mu.Lock()
	workers := make(map[string]*SessionWorker)
	for k, v := range wp.workers {
		workers[k] = v
	}
	wp.workers = make(map[string]*SessionWorker)
	wp.mu.Unlock()

	for id, w := range workers {
		w.Stop()
		wp.logger.Info("Worker stopped", "sessionID", id)
	}
}


