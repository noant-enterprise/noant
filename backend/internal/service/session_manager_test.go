package service

import (
	"testing"
	"time"

	"noant/config"
	"noant/internal/infrastructure"
)

func newTestSessionManager() *SessionManager {
	cfg := &config.Config{OpenWAQueueDepth: 100}
	queue := NewSendQueue(cfg, nil, infrastructure.NewNullLogger())
	return &SessionManager{
		cfg:      cfg,
		sessions: make(map[string]*SessionHealth),
		logger:   infrastructure.NewNullLogger(),
		workerPool: &WorkerPool{
			workers: make(map[string]*SessionWorker),
			queue:   queue,
			logger:  infrastructure.NewNullLogger(),
		},
	}
}

func TestRegisterSession(t *testing.T) {
	sm := newTestSessionManager()
	sm.RegisterSession("s1", "u1")

	sh := sm.GetSession("s1")
	if sh == nil {
		t.Fatal("expected session to be registered")
	}
	if sh.SessionID != "s1" {
		t.Fatalf("expected sessionID s1, got %s", sh.SessionID)
	}
	if sh.UserID != "u1" {
		t.Fatalf("expected userID u1, got %s", sh.UserID)
	}
	if sh.State != SessionStateUnknown {
		t.Fatalf("expected initial state unknown, got %s", sh.State)
	}
}

func TestRegisterSessionDuplicateIsNoop(t *testing.T) {
	sm := newTestSessionManager()
	sm.RegisterSession("s1", "u1")
	sm.RegisterSession("s1", "u2") // duplicate with different userID

	sh := sm.GetSession("s1")
	if sh == nil {
		t.Fatal("expected session to exist")
	}
	// First registration wins
	if sh.UserID != "u1" {
		t.Fatalf("expected userID u1 (first registration), got %s", sh.UserID)
	}
}

func TestUnregisterSession(t *testing.T) {
	sm := newTestSessionManager()
	sm.RegisterSession("s1", "u1")
	sm.UnregisterSession("s1")

	if sh := sm.GetSession("s1"); sh != nil {
		t.Fatal("expected session to be removed")
	}
}

func TestGetSessionByUserID(t *testing.T) {
	sm := newTestSessionManager()
	sm.RegisterSession("s1", "u1")
	sm.RegisterSession("s2", "u2")

	sh := sm.GetSessionByUserID("u1")
	if sh == nil || sh.SessionID != "s1" {
		t.Fatalf("expected session s1 for user u1, got %v", sh)
	}

	sh = sm.GetSessionByUserID("u2")
	if sh == nil || sh.SessionID != "s2" {
		t.Fatalf("expected session s2 for user u2, got %v", sh)
	}

	if sh := sm.GetSessionByUserID("u3"); sh != nil {
		t.Fatal("expected nil for unknown user")
	}
}

func TestUpdateStateConnected(t *testing.T) {
	sm := newTestSessionManager()
	sm.RegisterSession("s1", "u1")

	sm.UpdateState("s1", SessionStateConnected)
	sh := sm.GetSession("s1")
	if sh.State != SessionStateConnected {
		t.Fatalf("expected connected, got %s", sh.State)
	}
	if sh.ConsecutiveFailures != 0 {
		t.Fatalf("expected 0 consecutive failures after connect, got %d", sh.ConsecutiveFailures)
	}
	if sh.LastConnected.IsZero() {
		t.Fatal("expected LastConnected to be set")
	}
}

func TestUpdateStateDisconnectedIncrementsFailures(t *testing.T) {
	sm := newTestSessionManager()
	sm.RegisterSession("s1", "u1")

	sm.UpdateState("s1", SessionStateConnected)
	sm.UpdateState("s1", SessionStateDisconnected)

	sh := sm.GetSession("s1")
	if sh.State != SessionStateDisconnected {
		t.Fatalf("expected disconnected, got %s", sh.State)
	}
	if sh.ConsecutiveFailures != 1 {
		t.Fatalf("expected 1 consecutive failure, got %d", sh.ConsecutiveFailures)
	}
	if sh.LastDisconnected.IsZero() {
		t.Fatal("expected LastDisconnected to be set")
	}
}

func TestUpdateStateReconnectResetsFailures(t *testing.T) {
	sm := newTestSessionManager()
	sm.RegisterSession("s1", "u1")

	sm.UpdateState("s1", SessionStateConnected)
	sm.UpdateState("s1", SessionStateDisconnected)
	sm.UpdateState("s1", SessionStateDisconnected)
	sm.UpdateState("s1", SessionStateDisconnected)
	sm.UpdateState("s1", SessionStateConnected)

	sh := sm.GetSession("s1")
	if sh.ConsecutiveFailures != 0 {
		t.Fatalf("expected 0 consecutive failures after reconnect, got %d", sh.ConsecutiveFailures)
	}
}

func TestUpdateStateUnknownSessionIsNoop(t *testing.T) {
	sm := newTestSessionManager()
	// Should not panic
	sm.UpdateState("nonexistent", SessionStateConnected)
}

func TestListSessions(t *testing.T) {
	sm := newTestSessionManager()
	sm.RegisterSession("s1", "u1")
	sm.RegisterSession("s2", "u2")

	sessions := sm.ListSessions()
	if len(sessions) != 2 {
		t.Fatalf("expected 2 sessions, got %d", len(sessions))
	}
}

func TestUpdateStateTracksDowntime(t *testing.T) {
	sm := newTestSessionManager()
	sm.RegisterSession("s1", "u1")

	sm.UpdateState("s1", SessionStateConnected)
	// Simulate a short connection
	sh := sm.GetSession("s1")
	sh.LastConnected = time.Now().Add(-5 * time.Minute) // backdate

	sm.UpdateState("s1", SessionStateDisconnected)

	if sh.TotalDowntime <= 0 {
		t.Fatal("expected downtime > 0 after disconnect")
	}
}
