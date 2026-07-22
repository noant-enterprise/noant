package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"noant/config"
	"noant/internal/domain"
	"noant/internal/infrastructure"
	"noant/internal/repository"
)

type CreditService struct {
	cfg      *config.Config
	repos    *repository.Repositories
	redis    *infrastructure.RedisClient
	logger   *infrastructure.Logger
}

func NewCreditService(cfg *config.Config, repos *repository.Repositories, redis *infrastructure.RedisClient, logger *infrastructure.Logger) *CreditService {
	return &CreditService{
		cfg:    cfg,
		repos:  repos,
		redis:  redis,
		logger: logger,
	}
}

// PurchasePack returns the Polar checkout URL for a response pack
func (s *CreditService) PurchasePack(ctx context.Context, userID, packType string) (checkoutURL string, err error) {
	// Map pack types to Polar checkout URLs from config
	var urlStr string
	switch packType {
	case "small":
		urlStr = s.cfg.PolarPulseSmallURL
	case "medium":
		urlStr = s.cfg.PolarPulseMediumURL
	case "large":
		urlStr = s.cfg.PolarPulseLargeURL
	default:
		return "", fmt.Errorf("invalid pack type: %s", packType)
	}

	if urlStr == "" {
		return "", fmt.Errorf("polar checkout URL not configured for pack type: %s", packType)
	}

	// Append user_id and pack_type to static checkout URL
	if strings.Contains(urlStr, "?") {
		urlStr = fmt.Sprintf("%s&metadata[user_id]=%s&metadata[pack_type]=%s", urlStr, userID, packType)
	} else {
		urlStr = fmt.Sprintf("%s?metadata[user_id]=%s&metadata[pack_type]=%s", urlStr, userID, packType)
	}

	return urlStr, nil
}

// ActivatePurchase activates a completed credit purchase from webhook
func (s *CreditService) ActivatePurchase(ctx context.Context, checkoutID, userID, packType string) error {
	// Determine credit amount based on pack type
	var amount int
	switch packType {
	case "small":
		amount = 500 // ₦2,999 pack
	case "medium":
		amount = 1250 // ₦5,999 pack
	case "large":
		amount = 2500 // ₦9,999 pack
	default:
		return fmt.Errorf("invalid pack type: %s", packType)
	}

	// Get current credit balance
	currentCredit, err := s.repos.Credit.GetByUserID(ctx, userID)
	if err != nil {
		return err
	}

	// Calculate new balance and expiry (30 days from now)
	newBalance := currentCredit.Balance + amount
	expiry := time.Now().AddDate(0, 1, 0) // 30 days

	// Update credit record
	credit := &domain.UserCredit{
		ID:           currentCredit.ID,
		UserID:       userID,
		Balance:      newBalance,
		ExpiresAt:    &expiry,
		LastUpdatedAt: time.Now(),
	}

	if err := s.repos.Credit.Upsert(ctx, credit); err != nil {
		return err
	}

	// Create purchase record
	purchase := &domain.CreditPurchase{
		UserID:      userID,
		CheckoutID:  checkoutID,
		PackType:    packType,
		Amount:      amount,
		Status:      "completed",
		PurchasedAt: time.Now(),
		ExpiresAt:   expiry,
	}
	if err := s.repos.Credit.CreatePurchase(ctx, purchase); err != nil {
		// Non-fatal: credit balance is already updated, just log the failure
		s.logger.Warn("Failed to save credit purchase history record", "error", err, "userID", userID, "packType", packType)
	}

	s.logger.Info("Credit purchase activated", "userID", userID, "packType", packType, "amount", amount, "newBalance", newBalance)

	return nil
}

// Deduct deducts one response credit from user's balance atomically.
// The repository uses a serializable transaction with row-level locking,
// so balance checks and deduction happen atomically.
func (s *CreditService) Deduct(ctx context.Context, userID string) error {
	if err := s.repos.Credit.Deduct(ctx, userID, 1); err != nil {
		if strings.Contains(err.Error(), "insufficient balance") {
			return fmt.Errorf("insufficient credit balance")
		}
		if strings.Contains(err.Error(), "expired") {
			return fmt.Errorf("credit balance has expired")
		}
		return err
	}
	return nil
}

// GetBalance returns the user's current credit balance and expiry
func (s *CreditService) GetBalance(ctx context.Context, userID string) (*domain.UserCredit, error) {
	return s.repos.Credit.GetByUserID(ctx, userID)
}

// GetPurchaseHistory returns the user's credit purchase history from the database
func (s *CreditService) GetPurchaseHistory(ctx context.Context, userID string) ([]domain.CreditPurchase, error) {
	return s.repos.Credit.GetPurchaseHistory(ctx, userID)
}

// CheckAndNotifyExpiry is a background job that checks for credits expiring in 3 days
// and sends notification emails to users
func (s *CreditService) CheckAndNotifyExpiry(ctx context.Context) error {
	// Find credits expiring in 3 days
	expiringCredits, err := s.repos.Credit.GetExpiring(ctx, 3)
	if err != nil {
		return err
	}

	// For each expiring credit, notify the user
	for _, credit := range expiringCredits {
		if credit.Balance > 0 && credit.ExpiresAt != nil {
			daysLeft := int(time.Until(*credit.ExpiresAt).Hours() / 24)
			if daysLeft <= 3 && daysLeft > 0 {
				s.logger.Info("Credit expiry notification", "userID", credit.UserID, "daysLeft", daysLeft, "balance", credit.Balance)
				// In a full implementation, we would send an email here using EmailService
				// For example:
				// err := s.emailService.SendCreditExpiryWarning(credit.UserID, credit.Balance, daysLeft)
				// if err != nil {
				//     s.logger.Error("Failed to send credit expiry warning", "error", err)
				// }
			}
		}
	}

	return nil
}