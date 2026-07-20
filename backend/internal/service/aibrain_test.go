package service

import (
	"context"
	"strings"
	"testing"

	"noant/internal/domain"
)

func TestLocalPlatformAnswerQAMatch(t *testing.T) {
	brain := &AIBrain{}
	qa := []domain.QAPair{
		{ID: "qa1", Question: "What is your shipping policy", Answer: "We ship within 24 hours"},
	}
	resp := brain.localPlatformAnswer("u1", "shipping policy", qa, nil)
	if resp == nil {
		t.Fatal("expected a response from QA match")
	}
	if resp.Content != "We ship within 24 hours" {
		t.Fatalf("expected QA answer, got %q", resp.Content)
	}
	if resp.Source != "training" {
		t.Fatalf("expected source 'training', got %q", resp.Source)
	}
}

func TestLocalPlatformAnswerShortQueryPassesOverlap(t *testing.T) {
	brain := &AIBrain{}
	// Short query (1-2 words) gets automatic 1.0 overlap, so it matches
	qa := []domain.QAPair{
		{ID: "qa1", Question: "What is your return policy", Answer: "Returns are accepted within 30 days"},
	}
	resp := brain.localPlatformAnswer("u1", "hi", qa, nil)
	if resp == nil {
		t.Fatal("expected response — short queries auto-pass overlap check")
	}
	if resp.Content != "Returns are accepted within 30 days" {
		t.Fatalf("expected QA answer, got %q", resp.Content)
	}
}

func TestLocalPlatformAnswerNoMatch(t *testing.T) {
	brain := &AIBrain{}
	qa := []domain.QAPair{
		{ID: "qa1", Question: "What is your return policy", Answer: "Returns are accepted within 30 days"},
	}
	resp := brain.localPlatformAnswer("u1", "hello world test query", qa, nil)
	if resp != nil {
		t.Fatalf("expected nil when overlap < 0.3, got %v", resp)
	}
}

func TestLocalPlatformAnswerInventoryFallback(t *testing.T) {
	brain := &AIBrain{}
	inv := []domain.InventoryItem{
		{Name: "Starter Pack", Price: 1000},
	}
	resp := brain.localPlatformAnswer("u1", "iphone", nil, inv)
	if resp == nil {
		t.Fatal("expected response from inventory")
	}
	if resp.Source != "inventory" {
		t.Fatalf("expected source 'inventory', got %q", resp.Source)
	}
	if !contains(resp.Content, "Starter Pack") {
		t.Fatalf("expected 'Starter Pack' in response, got %q", resp.Content)
	}
}

func TestLocalPlatformAnswerNegotiation(t *testing.T) {
	brain := &AIBrain{}
	minPrice := 800.0
	inv := []domain.InventoryItem{
		{Name: "Starter Pack", Price: 1000, MinPrice: &minPrice},
	}
	resp := brain.localPlatformAnswer("u1", "can you do cheaper price", nil, inv)
	if resp == nil {
		t.Fatal("expected response from negotiation")
	}
	if !contains(resp.Content, "₦800") {
		t.Fatalf("expected discounted price ₦800 in response, got %q", resp.Content)
	}
}

func TestLocalPlatformAnswerNilWhenNoData(t *testing.T) {
	brain := &AIBrain{}
	resp := brain.localPlatformAnswer("u1", "something", nil, nil)
	if resp != nil {
		t.Fatal("expected nil when no QA and no inventory")
	}
}

func TestAllowGroqCallNoRedis(t *testing.T) {
	brain := &AIBrain{redis: nil}
	if !brain.allowGroqCall(context.TODO(), "u1") {
		t.Fatal("expected allow when redis is nil")
	}
}

func TestAllowGroqCallEmptyUserID(t *testing.T) {
	brain := &AIBrain{redis: nil}
	if !brain.allowGroqCall(context.TODO(), "") {
		t.Fatal("expected allow when userID is empty")
	}
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
