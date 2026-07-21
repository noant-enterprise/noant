package service

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
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
	allowed, err := b.redis.RateLimit(ctx, "groq_rate:"+userID, groqRateLimitPerMin, time.Minute)
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
		"model":       groqModel,
		"messages":    messages,
		"temperature": groqTemperature,
		"max_tokens":  groqMaxTokens,
		"top_p":       groqTopP,
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

// callGroqStreaming sends a request to Groq with streaming enabled.
// It calls onChunk for each text chunk received and returns the full content when done.
func (b *AIBrain) callGroqStreaming(ctx context.Context, messages []MessageTurn, onChunk func(chunk string)) (content string, confidence float64, err error) {
	if !b.cb.Allow() {
		return "", 0, fmt.Errorf("circuit breaker open: Groq API temporarily unavailable")
	}
	apiKey := b.getNextAPIKey()
	if apiKey == "" {
		b.cb.RecordFailure()
		return "", 0, fmt.Errorf("no Groq API keys configured")
	}
	payload := map[string]interface{}{
		"model":       groqModel,
		"messages":    messages,
		"temperature": groqTemperature,
		"max_tokens":  groqMaxTokens,
		"top_p":       groqTopP,
		"stream":      true,
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

	if resp.StatusCode != http.StatusOK {
		b.cb.RecordFailure()
		body, _ := io.ReadAll(resp.Body)
		snippet := string(body)
		if len(snippet) > 500 {
			snippet = snippet[:500]
		}
		return "", 0, fmt.Errorf("groq API error: %s - %s", resp.Status, snippet)
	}

	var builder strings.Builder
	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			break
		}
		var chunk struct {
			Choices []struct {
				Delta struct {
					Content string `json:"content"`
				} `json:"delta"`
				FinishReason *string `json:"finish_reason"`
			} `json:"choices"`
		}
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			continue
		}
		if len(chunk.Choices) == 0 {
			continue
		}
		delta := chunk.Choices[0].Delta.Content
		if delta != "" {
			builder.WriteString(delta)
			if onChunk != nil {
				onChunk(delta)
			}
		}
	}

	content = builder.String()
	b.cb.RecordSuccess()
	confidence = 0.85
	if len(content) < 10 {
		confidence = 0.4
	}
	return content, confidence, nil
}
