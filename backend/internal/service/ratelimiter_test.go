package service

import (
	"testing"
	"time"
)

func newTestRateLimiter() *MessageRateLimiter {
	return &MessageRateLimiter{
		windows:    make(map[string][]time.Time),
		textLimit:  5,
		mediaLimit: 3,
		tplLimit:   10,
		burstLimit: 2,
	}
}

func TestRateLimiterAllowsWithinLimit(t *testing.T) {
	rl := newTestRateLimiter()
	for i := 0; i < 5; i++ {
		ok, remaining := rl.Allow("s1", MsgTypeText)
		if !ok {
			t.Fatalf("expected allow on call %d, remaining=%d", i+1, remaining)
		}
	}
}

func TestRateLimiterBlocksWhenLimitAndBurstExhausted(t *testing.T) {
	rl := newTestRateLimiter()
	// 5 calls to fill limit + 2 burst calls = 7 allowed
	for i := 0; i < 7; i++ {
		ok, _ := rl.Allow("s1", MsgTypeText)
		if !ok {
			t.Fatalf("expected allow on call %d (within limit+burst)", i+1)
		}
	}
	// 8th call should be blocked
	ok, _ := rl.Allow("s1", MsgTypeText)
	if ok {
		t.Fatal("expected block after limit+burst exhausted")
	}
}

func TestRateLimiterDifferentSessions(t *testing.T) {
	rl := newTestRateLimiter()
	for i := 0; i < 7; i++ {
		rl.Allow("s1", MsgTypeText)
	}
	// s2 should still be allowed independently
	ok, remaining := rl.Allow("s2", MsgTypeText)
	if !ok {
		t.Fatal("expected s2 to be allowed independently")
	}
	if remaining < 4 {
		t.Fatalf("expected s2 remaining >= 4, got %d", remaining)
	}
}

func TestRateLimiterMediaLimitAndBurst(t *testing.T) {
	rl := newTestRateLimiter()
	// 3 limit + 2 burst = 5 allowed
	for i := 0; i < 5; i++ {
		ok, _ := rl.Allow("s1", MsgTypeMedia)
		if !ok {
			t.Fatalf("expected allow on media call %d", i+1)
		}
	}
	// 6th should be blocked
	ok, _ := rl.Allow("s1", MsgTypeMedia)
	if ok {
		t.Fatal("expected block after media limit+burst exhausted")
	}
}

func TestRateLimiterBurst(t *testing.T) {
	rl := newTestRateLimiter()
	// Fill text limit
	for i := 0; i < 5; i++ {
		rl.Allow("s1", MsgTypeText)
	}
	// Burst should allow 2 more within 5 seconds
	for i := 0; i < 2; i++ {
		ok, _ := rl.Allow("s1", MsgTypeText)
		if !ok {
			t.Fatalf("expected burst allow on attempt %d", i+1)
		}
	}
	// 3rd burst attempt should be blocked
	ok, _ := rl.Allow("s1", MsgTypeText)
	if ok {
		t.Fatal("expected block after burst exhausted")
	}
}

func TestRateLimiterWindowSlides(t *testing.T) {
	rl := newTestRateLimiter()
	for i := 0; i < 5; i++ {
		rl.Allow("s1", MsgTypeText)
	}
	// Move time forward past the window
	rl.windows["s1"] = []time.Time{time.Now().Add(-2 * time.Minute)}
	ok, _ := rl.Allow("s1", MsgTypeText)
	if !ok {
		t.Fatal("expected allow after window expires")
	}
}
