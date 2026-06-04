package service

import "testing"

func TestCleanWhatsAppID(t *testing.T) {
	got := cleanWhatsAppID("waid:+234-801-234-5678@s.whatsapp.net")
	if got != "2348012345678" {
		t.Fatalf("cleanWhatsAppID() = %q, want %q", got, "2348012345678")
	}
}

func TestFirstNonEmpty(t *testing.T) {
	got := firstNonEmpty("", "  ", "Alice", "Bob")
	if got != "Alice" {
		t.Fatalf("firstNonEmpty() = %q, want %q", got, "Alice")
	}
}

func TestResolveWhatsAppIdentityBasicFallback(t *testing.T) {
	svc := &ChatService{}
	identity, err := svc.ResolveWhatsAppIdentity(
		nil,
		"user-1",
		"session-1",
		&OpenWAMessageData{
			From: "+2348012345678@s.whatsapp.net",
			Sender: OpenWASender{
				Pushname: "Grace",
				ID:       "2348012345678@s.whatsapp.net",
			},
		},
	)
	if err != nil {
		t.Fatalf("ResolveWhatsAppIdentity() error = %v", err)
	}
	if identity.Name != "Grace" {
		t.Fatalf("expected name Grace, got %q", identity.Name)
	}
	if identity.Phone != "2348012345678" {
		t.Fatalf("expected phone 2348012345678, got %q", identity.Phone)
	}
	if len(identity.Methods) == 0 {
		t.Fatal("expected methods to be recorded")
	}
}
