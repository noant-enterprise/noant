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

	dedupKeyPrefix = "openwa:dedup:"
	dedupTTL       = 2 * time.Hour
)

// retryDelays in seconds for each retry attempt
var retryDelays = []time.Duration{
	0,                // attempt 1: immediate
	30 * time.Second, // attempt 2
	2 * time.Minute,  // attempt 3
	5 * time.Minute,  // attempt 4
	15 * time.Minute, // attempt 5
}

// ========== QUEUE ENTRY ==========

type QueueEntry struct {
	ID          string     `json:"id"`
	SessionID   string     `json:"session_id"`
	UserID      string     `json:"user_id"`
	ChatID      string     `json:"chat_id"`
	MsgType     string     `json:"msg_type"`
	Content     string     `json:"content"`
	MediaURL    string     `json:"media_url,omitempty"`
	Caption     string     `json:"caption,omitempty"`
	Priority    int        `json:"priority"`
	Status      string     `json:"status"`
	RetryCount  int        `json:"retry_count"`
	LastError   string     `json:"last_error,omitempty"`
	NextRetry   *time.Time `json:"next_retry,omitempty"`
	ScheduledAt time.Time  `json:"scheduled_at"`
	CreatedAt   time.Time  `json:"created_at"`
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
	rl := &MessageRateLimiter{
		windows:    make(map[string][]time.Time),
		textLimit:  cfg.OpenWARateLimitText,
		mediaLimit: cfg.OpenWARateLimitMedia,
		tplLimit:   cfg.OpenWARateLimitTemplate,
		burstLimit: cfg.OpenWARateLimitBurst,
	}
	go rl.evictLoop()
	return rl
}

func (rl *MessageRateLimiter) evictLoop() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		cutoff := now.Add(-5 * time.Minute)
		for sid, timestamps := range rl.windows {
			if len(timestamps) == 0 || timestamps[len(timestamps)-1].Before(cutoff) {
				delete(rl.windows, sid)
			}
		}
		rl.mu.Unlock()
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
	mu        sync.Mutex
	entries   []*QueueEntry
	bySession map[string][]*QueueEntry // sessionID -> entries for quick lookup
	byUser    map[string]int           // userID -> queued message count (Fix 9)
	redis     *infrastructure.RedisClient
	logger    *infrastructure.Logger
	cfg       *config.Config

	// Fix 8: condition variable for efficient worker wake-up
	cond     *sync.Cond
	stopCh   chan struct{}
	doneCh   chan struct{}
}

func NewSendQueue(cfg *config.Config, redis *infrastructure.RedisClient, logger *infrastructure.Logger) *SendQueue {
	sq := &SendQueue{
		entries:   make([]*QueueEntry, 0),
		bySession: make(map[string][]*QueueEntry),
		byUser:    make(map[string]int),
		redis:     redis,
		logger:    logger,
		cfg:       cfg,
		stopCh:    make(chan struct{}),
		doneCh:    make(chan struct{}),
	}
	sq.cond = sync.NewCond(&sq.mu)
	if redis != nil {
		sq.loadFromRedis()
	}
	go sq.evictExpiredLoop()
	return sq
}

// evictExpiredLoop removes entries older than MessageMaxAge (Fix 7)
func (sq *SendQueue) evictExpiredLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	defer close(sq.doneCh)

	for {
		select {
		case <-sq.stopCh:
			return
		case <-ticker.C:
			sq.evictExpired()
		}
	}
}

func (sq *SendQueue) evictExpired() {
	if sq.cfg.OpenWAMessageMaxAge <= 0 {
		return
	}
	sq.mu.Lock()
	defer sq.mu.Unlock()

	cutoff := time.Now().Add(-sq.cfg.OpenWAMessageMaxAge)
	var remaining []*QueueEntry
	for _, e := range sq.entries {
		if e.Status == QueueStatusQueued && e.CreatedAt.Before(cutoff) {
			sq.logger.Warn("Evicting expired queue entry", "id", e.ID, "age", time.Since(e.CreatedAt).String())
			sq.removeFromRedisUnlocked(context.Background(), e.ID)
			sq.byUser[e.UserID]--
			if sq.byUser[e.UserID] <= 0 {
				delete(sq.byUser, e.UserID)
			}
			continue
		}
		remaining = append(remaining, e)
	}
	if len(remaining) < len(sq.entries) {
		sq.entries = remaining
		sq.rebuildSessionIndex()
		sq.cond.Broadcast()
	}
}

func (sq *SendQueue) rebuildSessionIndex() {
	sq.bySession = make(map[string][]*QueueEntry)
	for _, e := range sq.entries {
		sq.bySession[e.SessionID] = append(sq.bySession[e.SessionID], e)
	}
}

// Stop shuts down the eviction loop
func (sq *SendQueue) Stop() {
	close(sq.stopCh)
	<-sq.doneCh
}

func (sq *SendQueue) loadFromRedis() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

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
			sq.byUser[entry.UserID]++
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
	if err := sq.redis.SetEx(ctx, key, string(data), 72*time.Hour); err != nil {
		sq.logger.Warn("Failed to persist queue entry to Redis", "id", entry.ID, "error", err)
	}
}

func (sq *SendQueue) removeFromRedisUnlocked(ctx context.Context, entryID string) {
	if sq.redis == nil {
		return
	}
	if err := sq.redis.Delete(ctx, fmt.Sprintf("openwa:queue:%s", entryID)); err != nil {
		sq.logger.Warn("Failed to remove queue entry from Redis", "id", entryID, "error", err)
	}
}

// isDuplicate checks if a message ID has been processed recently (Fix 4)
func (sq *SendQueue) isDuplicate(messageID string) bool {
	if sq.redis == nil || messageID == "" {
		return false
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	key := dedupKeyPrefix + messageID
	exists, err := sq.redis.Exists(ctx, key)
	if err != nil {
		return false
	}
	return exists
}

// markProcessed records a message ID as processed (Fix 4)
func (sq *SendQueue) markProcessed(messageID string) {
	if sq.redis == nil || messageID == "" {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_ = sq.redis.SetEx(ctx, dedupKeyPrefix+messageID, "1", dedupTTL)
}

// Enqueue adds a message to the queue with FIFO ordering within priority
// Validates: message size (Fix 6), dedup (Fix 4), per-user rate (Fix 9)
func (sq *SendQueue) Enqueue(entry *QueueEntry) error {
	sq.mu.Lock()
	defer sq.mu.Unlock()

	// Fix 6: message size validation
	maxSize := sq.cfg.OpenWAMaxMessageSize
	if maxSize <= 0 {
		maxSize = 65536
	}
	if len(entry.Content) > maxSize {
		return fmt.Errorf("message too large: %d bytes exceeds limit of %d", len(entry.Content), maxSize)
	}

	// Fix 4: dedup check on message ID
	if sq.isDuplicate(entry.ID) {
		return fmt.Errorf("duplicate message: %s already processed", entry.ID)
	}

	// Check queue depth per session
	if len(sq.bySession[entry.SessionID]) >= sq.cfg.OpenWAQueueDepth {
		return fmt.Errorf("queue depth exceeded for session %s: max %d", entry.SessionID, sq.cfg.OpenWAQueueDepth)
	}

	// Fix 9: per-user rate limit
	perUserLimit := sq.cfg.OpenWAPerUserRateLimit
	if perUserLimit > 0 && entry.UserID != "" {
		if sq.byUser[entry.UserID] >= perUserLimit {
			return fmt.Errorf("per-user queue limit exceeded for user %s: max %d", entry.UserID, perUserLimit)
		}
	}

	// Set creation time if zero
	if entry.CreatedAt.IsZero() {
		entry.CreatedAt = time.Now()
	}
	if entry.ScheduledAt.IsZero() {
		entry.ScheduledAt = time.Now()
	}

	// Insert in priority order (lower number = higher priority)
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
	sq.byUser[entry.UserID]++

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	sq.persist(ctx, entry)

	sq.logger.Debug("Message queued", "id", entry.ID, "session", entry.SessionID, "type", entry.MsgType, "priority", entry.Priority, "queueDepth", len(sq.entries))

	// Wake up waiting workers (Fix 8)
	sq.cond.Broadcast()
	return nil
}

// Dequeue retrieves the next ready message (respects next_retry timing and message TTL)
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
		if entry.ID != entryID {
			continue
		}
		entry.Status = QueueStatusSent
		sq.markProcessed(entryID)
		sq.removeEntry(i, entry)
		sq.cond.Broadcast()
		return
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
			sq.cond.Broadcast()
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

	sessionEntries := sq.bySession[entry.SessionID]
	for j, se := range sessionEntries {
		if se.ID == entry.ID {
			sq.bySession[entry.SessionID] = append(sessionEntries[:j], sessionEntries[j+1:]...)
			break
		}
	}

	sq.byUser[entry.UserID]--
	if sq.byUser[entry.UserID] <= 0 {
		delete(sq.byUser, entry.UserID)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	sq.removeFromRedisUnlocked(ctx, entry.ID)
}

// DrainQueue processes all remaining queued entries (Fix 1: graceful shutdown)
func (sq *SendQueue) DrainQueue(ctx context.Context, sender func(entry *QueueEntry) error) int {
	var processed int
	for {
		select {
		case <-ctx.Done():
			sq.logger.Warn("Queue drain context expired", "remaining", sq.Depth())
			return processed
		default:
		}

		entry := sq.Dequeue()
		if entry == nil {
			return processed
		}

		if err := sender(entry); err != nil {
			sq.Fail(entry.ID, err)
			sq.logger.Warn("Drain: message failed", "id", entry.ID, "error", err)
		} else {
			sq.Complete(entry.ID)
			processed++
		}
	}
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

// ========== SEND WORKER (Fix 8: sync.Cond instead of busy poll) ==========

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
	// Wake up the worker so it can check stopCh
	w.queue.cond.Broadcast()
	<-w.doneCh
}

func (w *SessionWorker) run() {
	defer close(w.doneCh)

	for {
		w.processNext()

		// Wait efficiently for new messages
		w.queue.mu.Lock()
		select {
		case <-w.stopCh:
			w.queue.mu.Unlock()
			return
		default:
		}
		// sync.Cond.Wait atomically unlocks the mutex and suspends the goroutine
		w.queue.cond.Wait()
		w.queue.mu.Unlock()
	}
}

func (w *SessionWorker) processNext() {
	if w.queue == nil || w.openwa == nil {
		return
	}

	// Check OpenWA API rate limit headers before dequeuing
	if w.openwa.ShouldBackoff(w.sessionID) {
		w.logger.Warn("Backing off due to OpenWA rate limit headers", "session", w.sessionID)
		return
	}

	entry := w.queue.DequeueBySession(w.sessionID)
	if entry == nil {
		return
	}

	// Rate limit check
	allowed, remaining := w.rateLimiter.Allow(w.sessionID, entry.MsgType)
	if !allowed {
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

// WorkerCount returns the number of active workers
func (wp *WorkerPool) WorkerCount() int {
	wp.mu.Lock()
	defer wp.mu.Unlock()
	return len(wp.workers)
}
