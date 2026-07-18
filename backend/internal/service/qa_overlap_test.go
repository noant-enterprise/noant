package service

import "testing"

func TestQaWordOverlapShortQuery(t *testing.T) {
	if qaWordOverlap("hi", "shipping policy") != 1.0 {
		t.Fatal("short queries (1-2 words) should return 1.0")
	}
	if qaWordOverlap("my order", "track order status") != 1.0 {
		t.Fatal("two-word queries should return 1.0")
	}
}

func TestQaWordOverlapExactSubstringMatch(t *testing.T) {
	got := qaWordOverlap("where are you located", "I am located in Lagos")
	// meaningful: where, are, you, located → 4 words
	// "located" appears in "located" → matched=1
	// result = 1/4 = 0.25
	if got != 0.25 {
		t.Fatalf("expected 0.25 (1/4), got %f", got)
	}
}

func TestQaWordOverlapHighMatch(t *testing.T) {
	got := qaWordOverlap("refund policy", "What is your refund policy")
	// query is 2 words → short circuit → 1.0
	if got != 1.0 {
		t.Fatalf("expected 1.0 (short query), got %f", got)
	}
}

func TestQaWordOverlapThreeWordQuery(t *testing.T) {
	got := qaWordOverlap("shipping policy info", "What is your shipping policy")
	// meaningful (>=3 chars): shipping, policy, info → 3 words
	// "shipping" appears in "shipping" → matched=1
	// "policy" appears in "policy" → matched=1
	// "info" does NOT appear in "shipping policy" → matched=0
	// result = 2/3 = 0.667
	if got < 0.6 {
		t.Fatalf("expected >= 0.6, got %f", got)
	}
}

func TestQaWordOverlapNoMatch(t *testing.T) {
	got := qaWordOverlap("hello there", "shipping policy")
	// 2-word query → short circuit → 1.0
	if got != 1.0 {
		t.Fatalf("expected 1.0 (short query), got %f", got)
	}
}

func TestQaWordOverlapZeroMeaningful(t *testing.T) {
	got := qaWordOverlap("a an is", "shipping policy")
	// all words are < 3 chars → meaningful=[], return 1.0
	if got != 1.0 {
		t.Fatalf("expected 1.0 (no meaningful words), got %f", got)
	}
}

func TestQaWordOverlapPunctuationTrimmed(t *testing.T) {
	got := qaWordOverlap("where's my order?", "order status")
	// query words after Fields: "where's", "my", "order?"
	// after Trim: "wheres", "my", "order"
	// meaningful: wheres(6), order(5) → "my" is 2 chars, skipped
	// "wheres" not in "order status" → 0
	// "order" in "order status" → 1
	// result = 1/2 = 0.5
	if got != 0.5 {
		t.Fatalf("expected 0.5, got %f", got)
	}
}
