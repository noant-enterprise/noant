package service

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func newTestCircuitBreaker() *CircuitBreaker {
	return &CircuitBreaker{state: "closed"}
}

func tripBreaker(cb *CircuitBreaker, count int) {
	for i := 0; i < count; i++ {
		cb.RecordFailure()
	}
}

func TestChaos_CircuitBreaker_OpenAfterConsecutiveFailures(t *testing.T) {
	cb := newTestCircuitBreaker()

	if !cb.Allow() {
		t.Fatal("closed breaker should allow requests")
	}

	for i := 0; i < circuitBreakerThreshold-1; i++ {
		cb.RecordFailure()
		if !cb.Allow() {
			t.Fatalf("breaker should still be closed after %d failures", i+1)
		}
	}

	cb.RecordFailure()
	if cb.Allow() {
		t.Fatal("breaker should be open after reaching threshold")
	}
}

func TestChaos_CircuitBreaker_HalfOpenAfterRecoveryWindow(t *testing.T) {
	cb := newTestCircuitBreaker()
	tripBreaker(cb, circuitBreakerThreshold)

	if cb.Allow() {
		t.Fatal("breaker should be open immediately after tripping")
	}

	cb.mutex.Lock()
	cb.lastFailure = time.Now().Add(-61 * time.Second)
	cb.mutex.Unlock()

	if !cb.Allow() {
		t.Fatal("breaker should transition to half-open after recovery window")
	}

	cb.mutex.RLock()
	state := cb.state
	cb.mutex.RUnlock()

	if state != "half-open" {
		t.Fatalf("expected half-open state, got %q", state)
	}
}

func TestChaos_CircuitBreaker_CloseAfterSuccessfulRecovery(t *testing.T) {
	cb := newTestCircuitBreaker()
	tripBreaker(cb, circuitBreakerThreshold)

	if cb.Allow() {
		t.Fatal("breaker should be open")
	}

	cb.mutex.Lock()
	cb.lastFailure = time.Now().Add(-61 * time.Second)
	cb.mutex.Unlock()

	cb.Allow()
	cb.RecordSuccess()

	cb.mutex.RLock()
	state := cb.state
	failures := cb.failures
	cb.mutex.RUnlock()

	if state != "closed" {
		t.Fatalf("expected closed state after success, got %q", state)
	}
	if failures != 0 {
		t.Fatalf("expected 0 failures after success, got %d", failures)
	}
}

func TestChaos_CircuitBreaker_ConcurrentAccessNoRace(t *testing.T) {
	cb := newTestCircuitBreaker()
	var wg sync.WaitGroup
	var panics atomic.Int64

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					panics.Add(1)
				}
			}()

			if cb.Allow() {
				if id%3 == 0 {
					cb.RecordFailure()
				} else {
					cb.RecordSuccess()
				}
			}
		}(i)
	}

	wg.Wait()

	if panics.Load() > 0 {
		t.Fatalf("got %d panics during concurrent access", panics.Load())
	}
}

func TestChaos_CircuitBreaker_RapidFireNoPanics(t *testing.T) {
	cb := newTestCircuitBreaker()
	var panics atomic.Int64

	for i := 0; i < 1000; i++ {
		func() {
			defer func() {
				if r := recover(); r != nil {
					panics.Add(1)
				}
			}()

			cb.Allow()
			cb.RecordFailure()
			cb.RecordSuccess()
		}()
	}

	if panics.Load() > 0 {
		t.Fatalf("got %d panics during rapid fire", panics.Load())
	}
}

func TestChaos_CircuitBreaker_RecoveryThenSuccess(t *testing.T) {
	cb := newTestCircuitBreaker()
	tripBreaker(cb, circuitBreakerThreshold)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				cb.mutex.Lock()
				cb.lastFailure = time.Now().Add(-61 * time.Second)
				cb.mutex.Unlock()
				time.Sleep(10 * time.Millisecond)
			}
		}
	}()

	deadline := time.After(1 * time.Second)
	for {
		select {
		case <-deadline:
			t.Fatal("breaker did not reach half-open in time")
		default:
			if cb.Allow() {
				cb.RecordSuccess()
				cb.mutex.RLock()
				state := cb.state
				cb.mutex.RUnlock()

				if state != "closed" {
					t.Fatalf("expected closed after recovery success, got %q", state)
				}
				return
			}
			time.Sleep(5 * time.Millisecond)
		}
	}
}

func TestChaos_CircuitBreaker_FaultInjectionLifecycle(t *testing.T) {
	cb := newTestCircuitBreaker()

	err := errors.New("simulated fault")
	if err == nil {
		t.Fatal("expected non-nil error")
	}

	tripBreaker(cb, circuitBreakerThreshold)

	cb.mutex.RLock()
	failures := cb.failures
	cb.mutex.RUnlock()

	if failures != circuitBreakerThreshold {
		t.Fatalf("expected %d failures, got %d", circuitBreakerThreshold, failures)
	}
}
