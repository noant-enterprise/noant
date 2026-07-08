package infrastructure

import (
	"testing"
	"time"
)

func TestMemoryRateLimiter_Allow(t *testing.T) {
	rl := NewMemoryRateLimiter(0)

	if !rl.Allow("key1", 3, time.Minute) {
		t.Error("expected first request to be allowed")
	}

	if !rl.Allow("key1", 3, time.Minute) {
		t.Error("expected second request to be allowed")
	}

	if !rl.Allow("key1", 3, time.Minute) {
		t.Error("expected third request to be allowed")
	}

	if rl.Allow("key1", 3, time.Minute) {
		t.Error("expected fourth request to be denied")
	}
}

func TestMemoryRateLimiter_WindowReset(t *testing.T) {
	rl := NewMemoryRateLimiter(0)

	if !rl.Allow("key1", 1, 50*time.Millisecond) {
		t.Error("expected first request to be allowed")
	}

	if rl.Allow("key1", 1, 50*time.Millisecond) {
		t.Error("expected second request to be denied within window")
	}

	time.Sleep(60 * time.Millisecond)

	if !rl.Allow("key1", 1, 50*time.Millisecond) {
		t.Error("expected request to be allowed after window expiry")
	}
}

func TestMemoryRateLimiter_IndependentKeys(t *testing.T) {
	rl := NewMemoryRateLimiter(0)

	if !rl.Allow("alice", 1, time.Minute) {
		t.Error("expected alice first to be allowed")
	}

	if rl.Allow("alice", 1, time.Minute) {
		t.Error("expected alice second to be denied")
	}

	if !rl.Allow("bob", 1, time.Minute) {
		t.Error("expected bob first to be allowed")
	}
}

func TestMemoryRateLimiter_Cleanup(t *testing.T) {
	rl := NewMemoryRateLimiter(50 * time.Millisecond)

	rl.Allow("stale", 1, 10*time.Millisecond)

	time.Sleep(100 * time.Millisecond)

	rl.mu.Lock()
	_, exists := rl.entries["stale"]
	rl.mu.Unlock()

	if exists {
		t.Error("expected stale entry to be cleaned up")
	}
}
