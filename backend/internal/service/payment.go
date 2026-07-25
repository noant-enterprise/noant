package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"noant/config"
	"noant/internal/domain"
	"noant/internal/infrastructure"
	"noant/internal/repository"
)

// ========== PAYMENT SERVICE ==========

type PaymentService struct {
	cfg      *config.Config
	repos    *repository.Repositories
	redis    *infrastructure.RedisClient
	logger   *infrastructure.Logger
	polarSvc *PolarService
	credit   *CreditService
}

func NewPaymentService(cfg *config.Config, repos *repository.Repositories, redis *infrastructure.RedisClient, logger *infrastructure.Logger, polarSvc *PolarService, credit *CreditService) *PaymentService {
	return &PaymentService{cfg: cfg, repos: repos, redis: redis, logger: logger, polarSvc: polarSvc, credit: credit}
}

func (s *PaymentService) ListPlans(ctx context.Context) ([]domain.PaymentPlan, error) {
	return []domain.PaymentPlan{
		{
			ID:          "free",
			Name:        "Free",
			PriceNGN:    0,
			AIResponses: 100, // per week
			Channels:    []string{"whatsapp", "web"},
			Features:    []string{"100 AI responses/week", "Web Widget + WhatsApp", "10 inventory items", "1 team member", "Basic AI responses", "Handoff system enabled"},
		},
		{
			ID:          "pulse",
			Name:        "Pulse",
			PriceNGN:    2999, // Starts at NGN 2,999
			AIResponses: 500,  // minimum pack size
			Channels:    []string{"telegram", "web", "whatsapp", "email"},
			Features:    []string{"Pay as you go", "All 4 channels", "Unlimited inventory", "Full handoff system", "Instant notifications", "AI price negotiation"},
		},
		{
			ID:          "pro",
			Name:        "Pro",
			PriceNGN:    21999,
			AIResponses: 0, // Unlimited
			Channels:    []string{"telegram", "web", "whatsapp", "email"},
			Features:    []string{"Unlimited AI responses", "Unlimited team members", "All 4 channels", "Full inventory & handoff", "AI price negotiation", "White-label widget", "Campaign Mode"},
			IsPopular:   true,
		},
		{
			ID:          "enterprise",
			Name:        "Enterprise",
			PriceNGN:    99999,
			AIResponses: 0, // Unlimited
			Channels:    []string{"telegram", "web", "whatsapp", "email", "instagram", "messenger"},
			Features:    []string{"Unlimited everything", "Custom AI training", "API access", "White-label platform", "SLA guarantee", "Dedicated account manager"},
		},
	}, nil
}

func (s *PaymentService) Subscribe(ctx context.Context, userID, planID string) (string, error) {
	// Handle free plan - no payment needed
	if planID == "free" {
		if err := s.repos.User.UpdatePlan(ctx, userID, "free"); err != nil {
			s.logger.Error("Failed to update user plan", "error", err)
			return "", err
		}
		s.logger.Info("User plan set to free", "user", userID, "plan", "free")
		return "", nil
	}

	// Determine planName / planID
	planName := planID
	switch planID {
	case "pulse", "pro", "enterprise":
	default:
		return "", fmt.Errorf("invalid plan ID: %s", planID)
	}

	// Try to get configured static URL
	var urlStr string
	switch planName {
	case "pulse":
		urlStr = s.cfg.PolarPulseSmallURL
	case "pro":
		urlStr = s.cfg.PolarProMonthlyURL
	case "enterprise":
		urlStr = s.cfg.PolarEnterpriseURL
	}

	if urlStr != "" {
		// Append metadata search params
		if strings.Contains(urlStr, "?") {
			urlStr = fmt.Sprintf("%s&metadata[user_id]=%s&metadata[plan_id]=%s", urlStr, userID, planName)
		} else {
			urlStr = fmt.Sprintf("%s?metadata[user_id]=%s&metadata[plan_id]=%s", urlStr, userID, planName)
		}
		s.logger.Info("Returning static Polar checkout URL with metadata", "user", userID, "plan", planName, "url", urlStr)
		return urlStr, nil
	}

	// Fallback to dynamic checkout if server URL is configured and access token is present
	if s.polarSvc != nil && s.cfg.PolarAccessToken != "" {
		checkoutURL, err := s.polarSvc.CreateCheckout(ctx, userID, planName)
		if err == nil && checkoutURL != "" {
			s.logger.Info("Polar checkout created dynamically", "user", userID, "plan", planName, "url", checkoutURL)
			return checkoutURL, nil
		}
		s.logger.Warn("Polar checkout creation failed", "user", userID, "plan", planName, "error", err)
	}

	// Local database fallback if Polar is not configured
	now := time.Now()
	periodEnd := now.AddDate(0, 1, 0)

	if err := s.repos.RunInTx(ctx, func(tx *sql.Tx) error {
		// Upsert subscription within the transaction
		existing, _ := s.repos.Subscription.GetActive(ctx, userID)
		if existing != nil {
			if _, err := tx.ExecContext(ctx, `UPDATE subscriptions SET plan_id = ?, status = 'active', current_period_start = ?, current_period_end = ?, updated_at = NOW() WHERE id = ?`, planName, now, periodEnd, existing.ID); err != nil {
				return fmt.Errorf("failed to update subscription: %w", err)
			}
		} else {
			if _, err := tx.ExecContext(ctx, `INSERT INTO subscriptions (id, user_id, plan_id, status, current_period_start, current_period_end, created_at, updated_at) VALUES (UUID(), ?, ?, 'active', ?, ?, NOW(), NOW())`, userID, planName, now, periodEnd); err != nil {
				return fmt.Errorf("failed to create subscription: %w", err)
			}
		}
		if _, err := tx.ExecContext(ctx, `UPDATE users SET plan_id = ? WHERE id = ?`, planName, userID); err != nil {
			return fmt.Errorf("failed to update user plan: %w", err)
		}
		return nil
	}); err != nil {
		s.logger.Error("Failed to complete subscription in transaction", "error", err)
		return "", err
	}

	s.logger.Info("Local subscription fallback created", "user", userID, "plan", planName, "period_end", periodEnd)
	return "", nil
}

func (s *PaymentService) Webhook(ctx context.Context, payload []byte, headers map[string]string) error {
	// First verify the signature
	if s.polarSvc != nil {
		if !s.polarSvc.VerifyWebhook(payload, headers) {
			return fmt.Errorf("invalid webhook signature")
		}
	}

	var event struct {
		Type string          `json:"type"`
		Data json.RawMessage `json:"data"`
	}

	if err := json.Unmarshal(payload, &event); err != nil {
		return fmt.Errorf("failed to parse webhook payload: %w", err)
	}

	s.logger.Info("Payment webhook received", "type", event.Type)

	switch event.Type {
	case "order.created":
		// Handle order payments (both one-time credit packs and subscription payments)
		var orderData struct {
			ID       string                 `json:"id"`
			Metadata map[string]interface{} `json:"metadata"`
		}

		if err := json.Unmarshal(event.Data, &orderData); err != nil {
			return fmt.Errorf("failed to parse order data: %w", err)
		}

		// Extract metadata
		var userID, packType, planID string
		if len(orderData.Metadata) > 0 {
			if uid, ok := orderData.Metadata["user_id"].(string); ok {
				userID = uid
			}
			if pt, ok := orderData.Metadata["pack_type"].(string); ok {
				packType = pt
			}
			if pid, ok := orderData.Metadata["plan_id"].(string); ok {
				planID = pid
			}
		}

		if userID == "" {
			s.logger.Warn("order.created event missing user_id in metadata", "order_id", orderData.ID)
			return nil // Don't fail the request, just ignore
		}

		if packType != "" {
			// Activate credit pack purchase
			if err := s.credit.ActivatePurchase(ctx, orderData.ID, userID, packType); err != nil {
				s.logger.Error("Failed to activate credit purchase from order webhook", "error", err, "userID", userID, "packType", packType)
				return err
			}
			s.logger.Info("Credit purchase activated via order.created webhook", "userID", userID, "packType", packType, "orderID", orderData.ID)
		} else if planID != "" {
			// Sync subscription plan from the order payment
			now := time.Now()
			periodEnd := now.AddDate(0, 1, 0) // Default 1 month

			if err := s.repos.RunInTx(ctx, func(tx *sql.Tx) error {
				existing, _ := s.repos.Subscription.GetActive(ctx, userID)
				if existing != nil {
					if _, err := tx.ExecContext(ctx, `UPDATE subscriptions SET plan_id = ?, status = 'active', current_period_start = ?, current_period_end = ?, updated_at = NOW() WHERE id = ?`, planID, now, periodEnd, existing.ID); err != nil {
						return err
					}
				} else {
					if _, err := tx.ExecContext(ctx, `INSERT INTO subscriptions (id, user_id, plan_id, status, current_period_start, current_period_end, created_at, updated_at) VALUES (UUID(), ?, ?, 'active', ?, ?, NOW(), NOW())`, userID, planID, now, periodEnd); err != nil {
						return err
					}
				}
				if _, err := tx.ExecContext(ctx, `UPDATE users SET plan_id = ? WHERE id = ?`, planID, userID); err != nil {
					return err
				}
				return nil
			}); err != nil {
				s.logger.Error("Failed to update subscription from order webhook", "error", err)
				return err
			}

			s.logger.Info("Subscription/plan updated via order.created webhook", "user", userID, "plan", planID, "orderID", orderData.ID)
		}

	case "subscription.created", "subscription.active", "subscription.updated":
		var subData struct {
			ID                 string                 `json:"id"`
			Status             string                 `json:"status"`
			CurrentPeriodStart string                 `json:"current_period_start"`
			CurrentPeriodEnd   string                 `json:"current_period_end"`
			Metadata           map[string]interface{} `json:"metadata"`
		}

		if err := json.Unmarshal(event.Data, &subData); err != nil {
			return fmt.Errorf("failed to parse subscription data: %w", err)
		}

		var userID, planID string
		if len(subData.Metadata) > 0 {
			if uid, ok := subData.Metadata["user_id"].(string); ok {
				userID = uid
			}
			if pid, ok := subData.Metadata["plan_id"].(string); ok {
				planID = pid
			}
		}

		if userID == "" {
			s.logger.Warn("Subscription event missing user_id in metadata", "sub_id", subData.ID, "type", event.Type)
			return nil
		}

		if planID == "" {
			planID = "pro" // Default fallback
		}

		// Parse dates or use defaults
		now := time.Now()
		periodEnd := now.AddDate(0, 1, 0)
		if subData.CurrentPeriodEnd != "" {
			if t, err := time.Parse(time.RFC3339, subData.CurrentPeriodEnd); err == nil {
				periodEnd = t
			}
		}

		// Handle cancellation / non-active status
		status := "active"
	if subData.Status == "canceled" || subData.Status == "revoked" || subData.Status == "cancelled" { //nolint:misspell // external API status values
		status = "cancelled" //nolint:misspell // DB status value
		}

		userPlan := planID
		if status == "cancelled" { //nolint:misspell // DB status value
			userPlan = "free"
		}

		if err := s.repos.RunInTx(ctx, func(tx *sql.Tx) error {
			existing, _ := s.repos.Subscription.GetActive(ctx, userID)
			if existing != nil {
				if _, err := tx.ExecContext(ctx, `UPDATE subscriptions SET plan_id = ?, status = ?, current_period_start = ?, current_period_end = ?, updated_at = NOW() WHERE id = ?`, planID, status, now, periodEnd, existing.ID); err != nil {
					return err
				}
			} else {
				if _, err := tx.ExecContext(ctx, `INSERT INTO subscriptions (id, user_id, plan_id, status, current_period_start, current_period_end, created_at, updated_at) VALUES (UUID(), ?, ?, ?, ?, ?, NOW(), NOW())`, userID, planID, status, now, periodEnd); err != nil {
					return err
				}
			}
			if _, err := tx.ExecContext(ctx, `UPDATE users SET plan_id = ? WHERE id = ?`, userPlan, userID); err != nil {
				return err
			}
			return nil
		}); err != nil {
			s.logger.Error("Failed to update subscription from webhook", "error", err)
			return err
		}

		s.logger.Info("Subscription updated via webhook", "user", userID, "plan", userPlan, "status", status, "subID", subData.ID)

	case "subscription.revoked", "subscription.cancelled":
		var subData struct {
			ID       string                 `json:"id"`
			Metadata map[string]interface{} `json:"metadata"`
		}

		if err := json.Unmarshal(event.Data, &subData); err != nil {
			return fmt.Errorf("failed to parse subscription data: %w", err)
		}

		var userID string
		if len(subData.Metadata) > 0 {
			if uid, ok := subData.Metadata["user_id"].(string); ok {
				userID = uid
			}
		}

		if userID != "" {
			if err := s.repos.RunInTx(ctx, func(tx *sql.Tx) error {
				if _, err := tx.ExecContext(ctx, `UPDATE subscriptions SET status = 'cancelled', updated_at = NOW() WHERE user_id = ? AND status = 'active'`, userID); err != nil {
					s.logger.Error("Failed to cancel subscription", "error", err)
				}
				if _, err := tx.ExecContext(ctx, `UPDATE users SET plan_id = 'free' WHERE id = ?`, userID); err != nil {
					return err
				}
				return nil
			}); err != nil {
				s.logger.Error("Failed to revoke subscription in transaction", "error", err)
			}
			s.logger.Info("Subscription revoked/canceled via webhook", "user", userID, "subID", subData.ID)
		}

	default:
		s.logger.Warn("Unhandled webhook event type", "type", event.Type)
	}

	return nil
}

func (s *PaymentService) Status(ctx context.Context, userID string) (*domain.Subscription, error) {
	return s.repos.Subscription.GetActive(ctx, userID)
}
