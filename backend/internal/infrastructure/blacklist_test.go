package infrastructure

import (
	"testing"
	"time"
)

func TestMemoryBlacklist_AddAndExists(t *testing.T) {
	b := NewMemoryBlacklist()

	if b.Exists("token1") {
		t.Error("expected new token to not be blacklisted")
	}

	b.Add("token1", time.Hour)

	if !b.Exists("token1") {
		t.Error("expected token to be blacklisted after Add")
	}
}

func TestMemoryBlacklist_Expired(t *testing.T) {
	b := NewMemoryBlacklist()

	b.Add("expired-token", 1*time.Millisecond)

	time.Sleep(5 * time.Millisecond)

	if b.Exists("expired-token") {
		t.Error("expected token to be expired")
	}
}

func TestMemoryBlacklist_MultipleTokens(t *testing.T) {
	b := NewMemoryBlacklist()

	b.Add("good", time.Hour)
	b.Add("bad", 1*time.Millisecond)

	time.Sleep(5 * time.Millisecond)

	if !b.Exists("good") {
		t.Error("expected good token to still be valid")
	}

	if b.Exists("bad") {
		t.Error("expected bad token to be expired")
	}
}

func TestMemoryBlacklist_EmptyToken(t *testing.T) {
	b := NewMemoryBlacklist()

	// Adding an empty token should not cause issues
	b.Add("", time.Hour)

	// Non-existent token should not be found
	if b.Exists("nonexistent") {
		t.Error("expected nonexistent token to not be found")
	}
}
