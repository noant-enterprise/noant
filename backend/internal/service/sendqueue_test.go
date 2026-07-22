package service

import (
	"errors"
	"sync"
	"testing"
	"time"

	"noant/config"
	"noant/internal/infrastructure"
)

func newTestQueue() *SendQueue {
	cfg := &config.Config{OpenWAQueueDepth: 100, OpenWAMaxMessageSize: 65536, OpenWAPerUserRateLimit: 50}
	sq := &SendQueue{
		entries:   make([]*QueueEntry, 0),
		bySession: make(map[string][]*QueueEntry),
		byUser:    make(map[string]int),
		cfg:       cfg,
		logger:    infrastructure.NewNullLogger(),
		stopCh:    make(chan struct{}),
		doneCh:    make(chan struct{}),
	}
	sq.cond = sync.NewCond(&sq.mu)
	return sq
}

func TestEnqueueFIFO(t *testing.T) {
	q := newTestQueue()
	e1 := &QueueEntry{ID: "1", Priority: PriorityNormal, Status: QueueStatusQueued, SessionID: "s1", ChatID: "c1"}
	e2 := &QueueEntry{ID: "2", Priority: PriorityNormal, Status: QueueStatusQueued, SessionID: "s1", ChatID: "c1"}
	e3 := &QueueEntry{ID: "3", Priority: PriorityNormal, Status: QueueStatusQueued, SessionID: "s1", ChatID: "c1"}

	_ = q.Enqueue(e1)
	_ = q.Enqueue(e2)
	_ = q.Enqueue(e3)

	d1 := q.Dequeue()
	if d1 == nil || d1.ID != "1" {
		t.Fatalf("expected first dequeued to be '1', got %v", d1)
	}
	q.Complete(d1.ID)

	d2 := q.Dequeue()
	if d2 == nil || d2.ID != "2" {
		t.Fatalf("expected second dequeued to be '2', got %v", d2)
	}
	q.Complete(d2.ID)

	d3 := q.Dequeue()
	if d3 == nil || d3.ID != "3" {
		t.Fatalf("expected third dequeued to be '3', got %v", d3)
	}
	q.Complete(d3.ID)
}

func TestEnqueuePriority(t *testing.T) {
	q := newTestQueue()
	eNormal := &QueueEntry{ID: "normal", Priority: PriorityNormal, Status: QueueStatusQueued, SessionID: "s1", ChatID: "c1"}
	eUrgent := &QueueEntry{ID: "urgent", Priority: PriorityUrgent, Status: QueueStatusQueued, SessionID: "s1", ChatID: "c1"}
	eBulk := &QueueEntry{ID: "bulk", Priority: PriorityBulk, Status: QueueStatusQueued, SessionID: "s1", ChatID: "c1"}

	_ = q.Enqueue(eNormal)
	_ = q.Enqueue(eUrgent)
	_ = q.Enqueue(eBulk)

	first := q.Dequeue()
	if first == nil || first.ID != "urgent" {
		t.Fatalf("expected urgent first, got %v", first)
	}
	q.Complete(first.ID)

	second := q.Dequeue()
	if second == nil || second.ID != "normal" {
		t.Fatalf("expected normal second, got %v", second)
	}
	q.Complete(second.ID)

	third := q.Dequeue()
	if third == nil || third.ID != "bulk" {
		t.Fatalf("expected bulk third, got %v", third)
	}
	q.Complete(third.ID)
}

func TestEnqueuePriorityFIFOWithinSame(t *testing.T) {
	q := newTestQueue()
	e1 := &QueueEntry{ID: "1", Priority: PriorityNormal, Status: QueueStatusQueued, SessionID: "s1", ChatID: "c1"}
	e2 := &QueueEntry{ID: "2", Priority: PriorityNormal, Status: QueueStatusQueued, SessionID: "s1", ChatID: "c1"}
	e3 := &QueueEntry{ID: "3", Priority: PriorityUrgent, Status: QueueStatusQueued, SessionID: "s1", ChatID: "c1"}
	e4 := &QueueEntry{ID: "4", Priority: PriorityNormal, Status: QueueStatusQueued, SessionID: "s1", ChatID: "c1"}

	_ = q.Enqueue(e1)
	_ = q.Enqueue(e2)
	_ = q.Enqueue(e3)
	_ = q.Enqueue(e4)

	// Expected order: urgent(3), then normal 1,2,4 (FIFO)
	got := q.Dequeue()
	if got.ID != "3" {
		t.Fatalf("expected 3 first (urgent), got %s", got.ID)
	}
	q.Complete(got.ID)

	got = q.Dequeue()
	if got.ID != "1" {
		t.Fatalf("expected 1 second (FIFO within normal), got %s", got.ID)
	}
	q.Complete(got.ID)

	got = q.Dequeue()
	if got.ID != "2" {
		t.Fatalf("expected 2 third, got %s", got.ID)
	}
	q.Complete(got.ID)

	got = q.Dequeue()
	if got.ID != "4" {
		t.Fatalf("expected 4 fourth, got %s", got.ID)
	}
	q.Complete(got.ID)
}

func TestRetryAndDeadLetter(t *testing.T) {
	q := newTestQueue()
	entry := &QueueEntry{ID: "1", Priority: PriorityNormal, Status: QueueStatusQueued, SessionID: "s1", ChatID: "c1"}
	_ = q.Enqueue(entry)

	sendErr := errors.New("network error")

	// Each fail sets NextRetry to now+delay; after MaxRetries+1 fails the entry
	// goes to dead letter and is removed. Since retry delays are in the future,
	// Dequeue returns nil after each fail (entry not ready yet).
	// Each fail increments RetryCount. At RetryCount == MaxRetries+1 (6th fail),
	// the entry goes to dead letter and is removed. Since retry delays are in
	// the future, Dequeue returns nil after each fail (entry not ready yet).
	// Override NextRetry to past to simulate retry window opening.
	for i := 0; i <= MaxRetries; i++ {
		entry.NextRetry = &pastTime
		d := q.Dequeue()
		if d == nil {
			t.Fatalf("expected entry on attempt %d", i+1)
		}
		q.Fail(d.ID, sendErr)
	}

	// After MaxRetries+1 failures, entry should be dead-lettered and removed
	if q.Depth() != 0 {
		t.Fatalf("expected depth 0 after dead letter, got %d", q.Depth())
	}
}

var pastTime = time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)

func TestRetryBackoff(t *testing.T) {
	q := newTestQueue()
	entry := &QueueEntry{ID: "1", Priority: PriorityNormal, Status: QueueStatusQueued, SessionID: "s1", ChatID: "c1"}
	_ = q.Enqueue(entry)

	sendErr := errors.New("timeout")
	d := q.Dequeue()
	q.Fail(d.ID, sendErr)

	if entry.RetryCount != 1 {
		t.Fatalf("expected retry count 1, got %d", entry.RetryCount)
	}
	if entry.NextRetry == nil {
		t.Fatal("expected NextRetry to be set")
	}
	if time.Now().After(*entry.NextRetry) {
		t.Fatal("NextRetry should be in the future")
	}
}

func TestQueueDepthLimit(t *testing.T) {
	cfg := &config.Config{OpenWAQueueDepth: 2, OpenWAMaxMessageSize: 65536, OpenWAPerUserRateLimit: 50}
	q := &SendQueue{
		entries:   make([]*QueueEntry, 0),
		bySession: make(map[string][]*QueueEntry),
		byUser:    make(map[string]int),
		cfg:       cfg,
		logger:    infrastructure.NewNullLogger(),
		stopCh:    make(chan struct{}),
		doneCh:    make(chan struct{}),
	}
	q.cond = sync.NewCond(&q.mu)
	_ = q.Enqueue(&QueueEntry{ID: "1", Priority: PriorityNormal, Status: QueueStatusQueued, SessionID: "s1", ChatID: "c1"})
	_ = q.Enqueue(&QueueEntry{ID: "2", Priority: PriorityNormal, Status: QueueStatusQueued, SessionID: "s1", ChatID: "c1"})
	err := q.Enqueue(&QueueEntry{ID: "3", Priority: PriorityNormal, Status: QueueStatusQueued, SessionID: "s1", ChatID: "c1"})
	if err == nil {
		t.Fatal("expected error when queue depth exceeded")
	}
}

func TestDequeueBySession(t *testing.T) {
	q := newTestQueue()
	_ = q.Enqueue(&QueueEntry{ID: "s1e1", Priority: PriorityNormal, Status: QueueStatusQueued, SessionID: "s1", ChatID: "c1"})
	_ = q.Enqueue(&QueueEntry{ID: "s2e1", Priority: PriorityNormal, Status: QueueStatusQueued, SessionID: "s2", ChatID: "c2"})
	_ = q.Enqueue(&QueueEntry{ID: "s1e2", Priority: PriorityNormal, Status: QueueStatusQueued, SessionID: "s1", ChatID: "c1"})

	got := q.DequeueBySession("s2")
	if got == nil || got.ID != "s2e1" {
		t.Fatalf("expected s2e1 by session s2, got %v", got)
	}
	q.Complete(got.ID)

	got = q.DequeueBySession("s1")
	if got == nil || got.ID != "s1e1" {
		t.Fatalf("expected s1e1 by session s1, got %v", got)
	}
	q.Complete(got.ID)

	got = q.DequeueBySession("s1")
	if got == nil || got.ID != "s1e2" {
		t.Fatalf("expected s1e2 by session s1, got %v", got)
	}
	q.Complete(got.ID)
}

func TestCompleteRemovesEntry(t *testing.T) {
	q := newTestQueue()
	_ = q.Enqueue(&QueueEntry{ID: "1", Priority: PriorityNormal, Status: QueueStatusQueued, SessionID: "s1", ChatID: "c1"})

	if q.Depth() != 1 {
		t.Fatalf("expected depth 1, got %d", q.Depth())
	}

	d := q.Dequeue()
	q.Complete(d.ID)

	if q.Depth() != 0 {
		t.Fatalf("expected depth 0 after complete, got %d", q.Depth())
	}
}
