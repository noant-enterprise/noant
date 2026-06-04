package service

import (
	"strings"
	"testing"

	"noant/internal/domain"
)

func TestDetectSalesStage(t *testing.T) {
	cases := []struct {
		name      string
		query     string
		want      string
		inventory []domain.InventoryItem
	}{
		{
			name:  "price negotiation",
			query: "can you do cheaper?",
			want:  "price negotiation",
		},
		{
			name:  "ready to buy",
			query: "i want to buy now",
			want:  "ready to buy",
		},
		{
			name:  "comparison",
			query: "which one should i choose?",
			want:  "comparison and recommendation",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := detectSalesStage(tc.query, tc.inventory, nil, nil)
			if got != tc.want {
				t.Fatalf("detectSalesStage(%q) = %q, want %q", tc.query, got, tc.want)
			}
		})
	}
}

func TestBroadInventoryQuery(t *testing.T) {
	if !isBroadInventoryQuery("what do you have?") {
		t.Fatal("expected broad inventory query to be true")
	}
	if isBroadInventoryQuery("Samsung A05") {
		t.Fatal("expected specific inventory query to be false")
	}
}

func TestLocalPlatformAnswerBroadInventory(t *testing.T) {
	brain := &AIBrain{}
	resp := brain.localPlatformAnswer("user-1", "what do you have?", nil, []domain.InventoryItem{
		{Name: "Starter Pack", Price: 1000},
		{Name: "Pro Pack", Price: 2000},
	})
	if resp == nil {
		t.Fatal("expected a response")
	}
	if !strings.Contains(resp.Content, "I found a few good options") {
		t.Fatalf("unexpected response: %s", resp.Content)
	}
}
