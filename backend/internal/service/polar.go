package service

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
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

func (s *PolarService) VerifyWebhook(payload []byte, headers map[string]string) bool {
	if s.cfg.PolarWebhookSecret == "" {
		return true // Allow bypass if not configured (for dev fallback)
	}

	// Strip whsec_ prefix if present
	secret := strings.TrimPrefix(s.cfg.PolarWebhookSecret, "whsec_")
	secretBytes, err := base64.StdEncoding.DecodeString(secret)
	if err != nil {
		return false
	}

	webhookID := headers["webhook-id"]
	webhookTimestamp := headers["webhook-timestamp"]
	webhookSignature := headers["webhook-signature"]

	if webhookID == "" || webhookTimestamp == "" || webhookSignature == "" {
		return false
	}

	// Verify timestamp is within 5 minutes (300 seconds) to prevent replay attacks
	timestampInt, err := strconv.ParseInt(webhookTimestamp, 10, 64)
	if err != nil {
		return false
	}

	now := time.Now().Unix()
	diff := now - timestampInt
	if diff < -300 || diff > 300 {
		return false
	}

	// Reconstruct payload: id.timestamp.payload
	signedContent := fmt.Sprintf("%s.%s.%s", webhookID, webhookTimestamp, string(payload))

	// Calculate HMAC SHA256
	mac := hmac.New(sha256.New, secretBytes)
	mac.Write([]byte(signedContent))
	computedBytes := mac.Sum(nil)
	computedBase64 := base64.StdEncoding.EncodeToString(computedBytes)

	// webhookSignature may contain multiple space-separated signatures (e.g. "v1,sig1 v1,sig2")
	signatures := strings.Fields(webhookSignature)
	for _, sig := range signatures {
		if strings.HasPrefix(sig, "v1,") {
			signaturePart := strings.TrimPrefix(sig, "v1,")
			if hmac.Equal([]byte(signaturePart), []byte(computedBase64)) {
				return true
			}
		}
	}

	return false
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
	CurrentPeriodEnd   time.Time `json:"current_period_end"` // wait, let's keep it simple or change back
}

// Keep a compatible version of ProcessWebhook just in case
func (s *PolarService) ProcessWebhook(ctx context.Context, payload []byte, headers map[string]string) (*SubscriptionData, error) {
	if !s.VerifyWebhook(payload, headers) {
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

	var subData struct {
		Status   string `json:"status"`
		Metadata struct {
			UserID string `json:"user_id"`
		} `json:"metadata"`
		CurrentPeriodStart string `json:"current_period_start"`
		CurrentPeriodEnd   string `json:"current_period_end"`
	}
	if err := json.Unmarshal(event.Data, &subData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal subscription data: %w", err)
	}

	if subData.Metadata.UserID == "" {
		return nil, fmt.Errorf("webhook missing user_id in metadata")
	}

	res := &SubscriptionData{
		Status: subData.Status,
	}
	res.Metadata.UserID = subData.Metadata.UserID
	res.CurrentPeriodStart = subData.CurrentPeriodStart
	if subData.CurrentPeriodEnd != "" {
		if t, err := time.Parse(time.RFC3339, subData.CurrentPeriodEnd); err == nil {
			res.CurrentPeriodEnd = t
		}
	}

	return res, nil
}