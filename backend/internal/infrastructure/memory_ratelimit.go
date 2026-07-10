package infrastructure

import (
	"sync"
	"time"
)

type MemoryRateLimiter struct {
	mu       sync.Mutex
	entries  map[string]*memRateEntry
	interval time.Duration
}

type memRateEntry struct {
	count    int
	window   time.Duration
	firstHit time.Time
}

func NewMemoryRateLimiter(cleanupInterval time.Duration) *MemoryRateLimiter {
	rl := &MemoryRateLimiter{
		entries:  make(map[string]*memRateEntry),
		interval: cleanupInterval,
	}
	if cleanupInterval > 0 {
		go rl.cleanupLoop()
	}
	return rl
}

func (rl *MemoryRateLimiter) Allow(key string, limit int, window time.Duration) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	entry, exists := rl.entries[key]
	if !exists || now.Sub(entry.firstHit) > entry.window {
		rl.entries[key] = &memRateEntry{
			count:    1,
			window:   window,
			firstHit: now,
		}
		return true
	}

	entry.count++
	return entry.count <= limit
}

func (rl *MemoryRateLimiter) cleanupLoop() {
	defer func() {
		if r := recover(); r != nil {
			// restart cleanup on panic
			go rl.cleanupLoop()
		}
	}()
	ticker := time.NewTicker(rl.interval)
	defer ticker.Stop()
	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for k, e := range rl.entries {
			if now.Sub(e.firstHit) > e.window {
				delete(rl.entries, k)
			}
		}
		rl.mu.Unlock()
	}
}
