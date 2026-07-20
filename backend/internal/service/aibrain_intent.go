package service

import (
	"context"
	"strings"
)

// classifyIntent uses LLM to classify user intent (replaces keyword matching)
func (b *AIBrain) classifyIntent(ctx context.Context, query string) string {
	// Fast path: check for clear handoff signals first (no LLM needed)
	lower := strings.ToLower(query)
	handoffTriggers := []string{"i want to buy", "i'll take it", "place the order", "checkout", "pay now", "send account number", "how do i pay", "send me your account"}
	for _, t := range handoffTriggers {
		if strings.Contains(lower, t) {
			return "handoff"
		}
	}

	salesTriggers := []string{"price", "cost", "how much", "available", "do you have", "show me", "recommend", "suggest", "discount", "cheaper", "stock", "product", "service", "package", "what do you sell"}
	for _, t := range salesTriggers {
		if strings.Contains(lower, t) {
			return "sales"
		}
	}

	supportTriggers := []string{"complaint", "problem", "issue", "refund", "return", "help", "support", "order status", "not working", "broken", "delay"}
	for _, t := range supportTriggers {
		if strings.Contains(lower, t) {
			return "support"
		}
	}

	// Use LLM for ambiguous cases
	prompt := []MessageTurn{
		{Role: "system", Content: `You are an intent classifier for a shop that sells products and services.
Classify the customer message into ONE intent:

- "sales": customer wants to BUY something, asks about products, prices, stock, availability, services, says "i need", "i want", "do you have", "show me", "how much"
- "handoff": customer is READY TO PAY, says "i want to buy", "deal", "checkout", "pay now", "send account"
- "support": customer has a complaint, asks about order status, return policy, or general help NOT about buying

IMPORTANT: "i need X" or "i want X" = sales (they want to buy).
Reply with ONLY the intent word.`},
		{Role: "user", Content: query},
	}

	response, _, err := b.callGroqWithFallback(ctx, prompt)
	if err != nil {
		// Fallback to keyword detection if LLM fails
		return b.keywordIntent(query)
	}

	response = strings.ToLower(strings.TrimSpace(response))
	if strings.Contains(response, "handoff") {
		return "handoff"
	}
	if strings.Contains(response, "sales") {
		return "sales"
	}
	return "support"
}

// keywordIntent is the fallback when LLM is unavailable
func (b *AIBrain) keywordIntent(query string) string {
	lower := strings.ToLower(query)
	salesTriggers := []string{"i need", "i want", "price", "how much", "cost", "stock", "available", "do you have", "what do you have", "show me", "product", "package", "do you sell", "discount", "cheaper", "last price", "service"}
	for _, t := range salesTriggers {
		if strings.Contains(lower, t) {
			return "sales"
		}
	}
	return "support"
}

func (b *AIBrain) isGreetingQuery(query string) bool {
	lower := strings.ToLower(strings.TrimSpace(query))
	if lower == "" {
		return false
	}

	greetings := []string{
		"hi",
		"hello",
		"hey",
		"good morning",
		"good afternoon",
		"good evening",
		"how are you",
	}
	for _, greeting := range greetings {
		if lower == greeting {
			return true
		}
	}

	words := strings.Fields(lower)
	if len(words) <= 2 {
		for _, greeting := range []string{"hi", "hello", "hey"} {
			if strings.Contains(lower, greeting) {
				return true
			}
		}
	}

	return false
}
