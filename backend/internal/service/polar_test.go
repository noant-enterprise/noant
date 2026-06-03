package service

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"noant/config"
	"strconv"
	"testing"
	"time"
)

func TestVerifyWebhook(t *testing.T) {
	secret := "whsec_dGVzdF9zZWNyZXRfa2V5XzEyMzQ1Njc4OTBfZ29vZA==" // base64 for "test_secret_key_1234567890_good"
	secretBytes, err := base64.StdEncoding.DecodeString("dGVzdF9zZWNyZXRfa2V5XzEyMzQ1Njc4OTBfZ29vZA==")
	if err != nil {
		t.Fatalf("failed to decode test secret: %v", err)
	}

	cfg := &config.Config{
		PolarWebhookSecret: secret,
	}
	s := NewPolarService(cfg)

	payload := []byte(`{"type":"order.created","data":{"id":"ord_123"}}`)
	webhookID := "msg_123"
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)

	// Reconstruct payload: id.timestamp.payload
	signedContent := fmt.Sprintf("%s.%s.%s", webhookID, timestamp, string(payload))

	// Compute signature
	mac := hmac.New(sha256.New, secretBytes)
	mac.Write([]byte(signedContent))
	computedBytes := mac.Sum(nil)
	computedBase64 := base64.StdEncoding.EncodeToString(computedBytes)

	headers := map[string]string{
		"webhook-id":        webhookID,
		"webhook-timestamp": timestamp,
		"webhook-signature": "v1," + computedBase64,
	}

	if !s.VerifyWebhook(payload, headers) {
		t.Error("VerifyWebhook failed with valid signature")
	}

	// Test invalid signature
	headers["webhook-signature"] = "v1,invalid_signature"
	if s.VerifyWebhook(payload, headers) {
		t.Error("VerifyWebhook succeeded with invalid signature")
	}

	// Test expired timestamp (more than 5 mins ago)
	oldTimestamp := strconv.FormatInt(time.Now().Unix()-301, 10)
	signedContentOld := fmt.Sprintf("%s.%s.%s", webhookID, oldTimestamp, string(payload))
	macOld := hmac.New(sha256.New, secretBytes)
	macOld.Write([]byte(signedContentOld))
	computedBase64Old := base64.StdEncoding.EncodeToString(macOld.Sum(nil))

	headersOld := map[string]string{
		"webhook-id":        webhookID,
		"webhook-timestamp": oldTimestamp,
		"webhook-signature": "v1," + computedBase64Old,
	}
	if s.VerifyWebhook(payload, headersOld) {
		t.Error("VerifyWebhook succeeded with expired timestamp")
	}
}
