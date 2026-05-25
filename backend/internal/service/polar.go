package service

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"noant/config"
)

type PolarService struct {
	cfg    *config.Config
	client *http.Client
}

func NewPolarService(cfg *config.Config) *PolarService {
	return &PolarService{
		cfg:    cfg,
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

func (s *PolarService) CreateCheckout(ctx context.Context, userID, planID string) (string, error) {
	payload := map[string]interface{}{
		"product_price_id": planID,
		"metadata": map[string]string{
			"user_id": userID,
		},
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal checkout payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", s.cfg.PolarServerURL+"/v1/checkouts/", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+s.cfg.PolarAccessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("polar API error: %s - %s", resp.Status, string(body))
	}

	var result struct {
		URL string `json:"url"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	return result.URL, nil
}

func (s *PolarService) VerifyWebhook(payload []byte, signature string) bool {
	if s.cfg.PolarWebhookSecret == "" {
		return false
	}

	mac := hmac.New(sha256.New, []byte(s.cfg.PolarWebhookSecret))
	mac.Write(payload)
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(signature), []byte(expected))
}

type PolarWebhookEvent struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

type SubscriptionData struct {
	Status   string `json:"status"`
	Metadata struct {
		UserID string `json:"user_id"`
	} `json:"metadata"`
	CurrentPeriodStart string `json:"current_period_start"`
	CurrentPeriodEnd   string `json:"current_period_end"`
}

func (s *PolarService) ProcessWebhook(ctx context.Context, payload []byte, signature string) (*SubscriptionData, error) {
	if !s.VerifyWebhook(payload, signature) {
		return nil, fmt.Errorf("invalid webhook signature")
	}

	var event PolarWebhookEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return nil, fmt.Errorf("failed to unmarshal webhook: %w", err)
	}

	// Only process subscription events
	validEvents := map[string]bool{
		"subscription.created":   true,
		"subscription.updated":   true,
		"subscription.active":    true,
		"subscription.cancelled": true,
	}

	if !validEvents[event.Type] {
		return nil, fmt.Errorf("unhandled event type: %s", event.Type)
	}

	var subData SubscriptionData
	if err := json.Unmarshal(event.Data, &subData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal subscription data: %w", err)
	}

	if subData.Metadata.UserID == "" {
		return nil, fmt.Errorf("webhook missing user_id in metadata")
	}

	return &subData, nil
}