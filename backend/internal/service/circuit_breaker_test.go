package service

import (
	"sync"
	"testing"
	"time"
)

func TestAICircuitBreaker_InitialState(t *testing.T) {
	cb := &CircuitBreaker{state: "closed"}
	if cb.state != "closed" {
		t.Fatalf("expected initial state 'closed', got %q", cb.state)
	}
	if !cb.Allow() {
		t.Fatal("closed circuit breaker should allow requests")
	}
}

func TestAICircuitBreaker_OpenAfterFailures(t *testing.T) {
	cb := &CircuitBreaker{state: "closed"}
	for i := 0; i < 2; i++ {
		cb.RecordFailure()
		if cb.state != "closed" {
			t.Fatalf("expected closed after %d failures, got %q", i+1, cb.state)
		}
	}
	cb.RecordFailure()
	if cb.state != "open" {
		t.Fatalf("expected open after 3 failures, got %q", cb.state)
	}
}

func TestAICircuitBreaker_OpenBlocksRequests(t *testing.T) {
	cb := &CircuitBreaker{state: "closed"}
	for i := 0; i < 3; i++ {
		cb.RecordFailure()
	}
	if cb.Allow() {
		t.Fatal("open circuit breaker should block requests")
	}
}

func TestAICircuitBreaker_HalfOpenAfterTimeout(t *testing.T) {
	cb := &CircuitBreaker{state: "open"}
	cb.failures = 3
	cb.lastFailure = time.Now().Add(-61 * time.Second)

	if !cb.Allow() {
		t.Fatal("should transition to half-open after timeout")
	}
	if cb.state != "half-open" {
		t.Fatalf("expected half-open, got %q", cb.state)
	}
}

func TestAICircuitBreaker_SuccessResetsToClosed(t *testing.T) {
	cb := &CircuitBreaker{state: "half-open"}
	cb.failures = 0

	cb.RecordSuccess()
	if cb.state != "closed" {
		t.Fatalf("expected closed after success in half-open, got %q", cb.state)
	}
	if cb.failures != 0 {
		t.Fatalf("expected 0 failures after success, got %d", cb.failures)
	}
}

func TestAICircuitBreaker_FailureInHalfOpenReopens(t *testing.T) {
	cb := &CircuitBreaker{state: "half-open"}
	cb.failures = 0

	for i := 0; i < 2; i++ {
		cb.RecordFailure()
		if cb.state != "half-open" {
			t.Fatalf("expected half-open after %d failures, got %q", i+1, cb.state)
		}
	}
	cb.RecordFailure()
	if cb.state != "open" {
		t.Fatalf("expected open after 3 failures in half-open, got %q", cb.state)
	}
}

func TestAICircuitBreaker_ConcurrentAccess(t *testing.T) {
	cb := &CircuitBreaker{state: "closed"}
	var wg sync.WaitGroup
	const goroutines = 100

	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			cb.Allow()
			cb.RecordFailure()
			cb.RecordSuccess()
		}()
	}
	wg.Wait()

	if cb.state != "closed" && cb.state != "open" {
		t.Fatalf("unexpected state after concurrent access: %q", cb.state)
	}
}
