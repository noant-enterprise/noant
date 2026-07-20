package infrastructure

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
)

// Bottleneck provides adaptive rate limiting and concurrency control
type Bottleneck struct {
	mu            sync.RWMutex
	maxConcurrent int64
	active        int64
	waiting       int64
	maxQueue      int64
	tokenBucket   chan struct{}
	stats         BottleneckStats
}

// BottleneckStats tracks bottleneck performance
type BottleneckStats struct {
	RequestsTotal     int64
	RequestsRejected  int64
	RequestsCompleted int64
	AvgWaitTime       time.Duration
	PeakConcurrent    int64
}

// BottleneckOption configures the bottleneck
type BottleneckOption func(*Bottleneck)

func WithMaxConcurrent(n int) BottleneckOption {
	return func(b *Bottleneck) {
		b.maxConcurrent = int64(n)
	}
}

func WithMaxQueue(n int) BottleneckOption {
	return func(b *Bottleneck) {
		b.maxQueue = int64(n)
	}
}

// NewBottleneck creates a new concurrency limiter
func NewBottleneck(opts ...BottleneckOption) *Bottleneck {
	b := &Bottleneck{
		maxConcurrent: 100,
		maxQueue:      500,
	}

	for _, opt := range opts {
		opt(b)
	}

	// Create token bucket with the correct buffer size
	b.tokenBucket = make(chan struct{}, b.maxConcurrent)

	// Pre-fill token bucket
	for i := int64(0); i < b.maxConcurrent; i++ {
		b.tokenBucket <- struct{}{}
	}

	return b
}

// Acquire blocks until a slot is available or context is canceled
func (b *Bottleneck) Acquire(ctx context.Context) error {
	atomic.AddInt64(&b.waiting, 1)

	// Check if queue is full
	if atomic.LoadInt64(&b.waiting) > b.maxQueue {
		atomic.AddInt64(&b.waiting, -1)
		atomic.AddInt64(&b.stats.RequestsRejected, 1)
		return fmt.Errorf("server busy, try again later")
	}

	select {
	case <-b.tokenBucket:
		atomic.AddInt64(&b.waiting, -1)
		active := atomic.AddInt64(&b.active, 1)
		atomic.AddInt64(&b.stats.RequestsTotal, 1)

		// Track peak concurrency
		if active > atomic.LoadInt64(&b.stats.PeakConcurrent) {
			atomic.StoreInt64(&b.stats.PeakConcurrent, active)
		}
		return nil

	case <-ctx.Done():
		atomic.AddInt64(&b.waiting, -1)
		return ctx.Err()
	}
}

// Release frees a slot
func (b *Bottleneck) Release() {
	atomic.AddInt64(&b.active, -1)
	atomic.AddInt64(&b.stats.RequestsCompleted, 1)
	b.tokenBucket <- struct{}{}
}

// Stats returns bottleneck statistics
func (b *Bottleneck) Stats() BottleneckStats {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return BottleneckStats{
		RequestsTotal:     atomic.LoadInt64(&b.stats.RequestsTotal),
		RequestsRejected:  atomic.LoadInt64(&b.stats.RequestsRejected),
		RequestsCompleted: atomic.LoadInt64(&b.stats.RequestsCompleted),
		PeakConcurrent:    atomic.LoadInt64(&b.stats.PeakConcurrent),
	}
}

// BottleneckMiddleware creates a Gin middleware for rate limiting
func BottleneckMiddleware(b *Bottleneck) func(ctx *gin.Context) {
	return func(c *gin.Context) {
		if err := b.Acquire(c.Request.Context()); err != nil {
			c.JSON(429, gin.H{"error": "too many concurrent requests, please retry"})
			c.Abort()
			return
		}
		defer b.Release()
		c.Next()
	}
}

// GroupBottleneck limits concurrent operations per resource group
type GroupBottleneck struct {
	mu         sync.RWMutex
	groups     map[string]*Bottleneck
	maxPerGroup int64
}

// NewGroupBottleneck creates a new per-group bottleneck
func NewGroupBottleneck(maxPerGroup int) *GroupBottleneck {
	return &GroupBottleneck{
		groups:      make(map[string]*Bottleneck),
		maxPerGroup: int64(maxPerGroup),
	}
}

// Acquire acquires a slot for a specific group
func (gb *GroupBottleneck) Acquire(ctx context.Context, group string) error {
	gb.mu.Lock()
	b, exists := gb.groups[group]
	if !exists {
		b = NewBottleneck(WithMaxConcurrent(int(gb.maxPerGroup)))
		gb.groups[group] = b
	}
	gb.mu.Unlock()
	return b.Acquire(ctx)
}

// Release releases a slot for a specific group
func (gb *GroupBottleneck) Release(group string) {
	gb.mu.RLock()
	b, exists := gb.groups[group]
	gb.mu.RUnlock()
	if exists {
		b.Release()
	}
}