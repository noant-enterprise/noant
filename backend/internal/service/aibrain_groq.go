package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"noant/internal/infrastructure"
)

func (b *AIBrain) getNextAPIKey() string {
	b.keyMutex.Lock()
	defer b.keyMutex.Unlock()
	if len(b.cfg.GroqAPIKeys) == 0 {
		return ""
	}
	key := b.cfg.GroqAPIKeys[b.keyIndex]
	b.keyIndex = (b.keyIndex + 1) % len(b.cfg.GroqAPIKeys)
	return key
}

// allowGroqCall checks if the user has remaining Groq API calls in the current window.
// Returns true if the call is allowed, false if rate limited.
func (b *AIBrain) allowGroqCall(ctx context.Context, userID string) bool {
	if b.redis == nil || userID == "" {
		return true
	}
	allowed, err := b.redis.RateLimit(ctx, "groq_rate:"+userID, 20, time.Minute)
	if err != nil {
		b.logger.Warn("Groq rate limit check failed, allowing call", "error", err)
		return true
	}
	if !allowed {
		b.logger.Warn("Groq rate limit exceeded for user", "userID", userID)
		infrastructure.NoantGroqRateLimited.Inc()
	}
	return allowed
}

func (b *AIBrain) callGroqWithFallback(ctx context.Context, messages []MessageTurn) (content string, confidence float64, err error) {
	if !b.cb.Allow() {
		return "", 0, fmt.Errorf("circuit breaker open: Groq API temporarily unavailable")
	}
	apiKey := b.getNextAPIKey()
	if apiKey == "" {
		b.cb.RecordFailure()
		return "", 0, fmt.Errorf("no Groq API keys configured")
	}
	payload := map[string]interface{}{
		"model":       "llama-3.3-70b-versatile",
		"messages":    messages,
		"temperature": 0.1,
		"max_tokens":  500,
		"top_p":       0.9,
	}
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return "", 0, fmt.Errorf("failed to marshal request payload: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.groq.com/openai/v1/chat/completions", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return "", 0, err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := b.httpClient.Do(req)
	if err != nil {
		b.cb.RecordFailure()
		return "", 0, err
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		b.cb.RecordFailure()
		return "", 0, fmt.Errorf("failed to read response body: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		b.cb.RecordFailure()
		snippet := string(body)
		if len(snippet) > 500 {
			snippet = snippet[:500]
		}
		return "", 0, fmt.Errorf("groq API error: %s - %s", resp.Status, snippet)
	}
	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
			FinishReason string `json:"finish_reason"`
		} `json:"choices"`
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
		} `json:"usage"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		b.cb.RecordFailure()
		return "", 0, err
	}
	if len(result.Choices) == 0 {
		b.cb.RecordFailure()
		return "", 0, fmt.Errorf("no response from Groq")
	}
	content = result.Choices[0].Message.Content
	b.cb.RecordSuccess()
	confidence = 0.85
	if result.Choices[0].FinishReason != "stop" {
		confidence = 0.5
	}
	if result.Usage.CompletionTokens < 10 {
		confidence = 0.4
	}
	return content, confidence, nil
}
