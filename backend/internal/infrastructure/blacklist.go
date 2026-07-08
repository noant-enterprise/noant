package infrastructure

import (
	"sync"
	"time"
)

type MemoryBlacklist struct {
	entries map[string]time.Time
	mu      sync.RWMutex
}

func NewMemoryBlacklist() *MemoryBlacklist {
	b := &MemoryBlacklist{entries: make(map[string]time.Time)}
	go func() {
		ticker := time.NewTicker(10 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			b.mu.Lock()
			now := time.Now()
			for k, exp := range b.entries {
				if now.After(exp) {
					delete(b.entries, k)
				}
			}
			b.mu.Unlock()
		}
	}()
	return b
}

func (b *MemoryBlacklist) Add(token string, ttl time.Duration) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.entries[token] = time.Now().Add(ttl)
}

func (b *MemoryBlacklist) Exists(token string) bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	exp, ok := b.entries[token]
	if !ok {
		return false
	}
	if time.Now().After(exp) {
		return false
	}
	return true
}
