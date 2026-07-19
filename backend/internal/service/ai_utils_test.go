package service

import "testing"

func TestParseAIMetadata_WithAllTags(t *testing.T) {
	input := "Thanks for reaching out!\n[SENTIMENT:positive]\n[LANGUAGE:en]\n[SUGGESTIONS:Buy now|Check prices|Contact support]"
	clean, sentiment, language, suggestions := parseAIMetadata(input)
	if clean != "Thanks for reaching out!" {
		t.Fatalf("expected clean text 'Thanks for reaching out!', got %q", clean)
	}
	if sentiment != "positive" {
		t.Fatalf("expected sentiment 'positive', got %q", sentiment)
	}
	if language != "en" {
		t.Fatalf("expected language 'en', got %q", language)
	}
	if len(suggestions) != 3 {
		t.Fatalf("expected 3 suggestions, got %d", len(suggestions))
	}
	if suggestions[0] != "Buy now" || suggestions[1] != "Check prices" || suggestions[2] != "Contact support" {
		t.Fatalf("unexpected suggestions: %v", suggestions)
	}
}

func TestParseAIMetadata_NoTags(t *testing.T) {
	input := "This is a plain response with no metadata."
	clean, sentiment, language, suggestions := parseAIMetadata(input)
	if clean != "This is a plain response with no metadata." {
		t.Fatalf("expected clean text unchanged, got %q", clean)
	}
	if sentiment != "neutral" {
		t.Fatalf("expected default sentiment 'neutral', got %q", sentiment)
	}
	if language != "en" {
		t.Fatalf("expected default language 'en', got %q", language)
	}
	if len(suggestions) != 0 {
		t.Fatalf("expected no suggestions, got %v", suggestions)
	}
}

func TestParseAIMetadata_PartialTags(t *testing.T) {
	input := "Hello!\n[SENTIMENT:negative]"
	clean, sentiment, language, suggestions := parseAIMetadata(input)
	if clean != "Hello!" {
		t.Fatalf("expected 'Hello!', got %q", clean)
	}
	if sentiment != "negative" {
		t.Fatalf("expected 'negative', got %q", sentiment)
	}
	if language != "en" {
		t.Fatalf("expected default language 'en', got %q", language)
	}
	if len(suggestions) != 0 {
		t.Fatalf("expected no suggestions, got %v", suggestions)
	}
}

func TestMin(t *testing.T) {
	tests := []struct {
		a, b, want int
	}{
		{1, 2, 1},
		{5, 3, 3},
		{7, 7, 7},
		{0, -1, -1},
		{-5, -3, -5},
	}
	for _, tt := range tests {
		got := min(tt.a, tt.b)
		if got != tt.want {
			t.Errorf("min(%d, %d) = %d, want %d", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestQaWordOverlap_ExactMatch(t *testing.T) {
	got := qaWordOverlap("shipping policy", "shipping policy")
	if got != 1.0 {
		t.Fatalf("expected 1.0 for exact match, got %f", got)
	}
}

func TestQaWordOverlap_PartialMatch(t *testing.T) {
	got := qaWordOverlap("return refund policy", "What is your refund policy")
	// meaningful: return, refund, policy → "refund" in question, "policy" in question
	// "return" not in "what is your refund policy" (substring check)
	// result = 2/3 ≈ 0.667
	if got < 0.6 || got > 0.7 {
		t.Fatalf("expected ~0.667, got %f", got)
	}
}

func TestQaWordOverlap_NoMatch(t *testing.T) {
	got := qaWordOverlap("quantum physics theory", "What is your shipping policy")
	// meaningful: quantum, physics, theory → none appear in question
	// result = 0/3 = 0.0
	if got != 0.0 {
		t.Fatalf("expected 0.0 for no match, got %f", got)
	}
}

func TestQaWordOverlap_ShortQuery(t *testing.T) {
	if got := qaWordOverlap("hi", "anything at all"); got != 1.0 {
		t.Fatalf("expected 1.0 for single word, got %f", got)
	}
	if got := qaWordOverlap("hello world", "anything at all"); got != 1.0 {
		t.Fatalf("expected 1.0 for two words, got %f", got)
	}
}
