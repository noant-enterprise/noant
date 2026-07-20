package service

import (
	"context"
	"fmt"
	"time"

	"noant/config"
	"noant/internal/domain"
	"noant/internal/infrastructure"
	"noant/internal/repository"
)

type PlanService struct {
	cfg    *config.Config
	repos  *repository.Repositories
	redis  *infrastructure.RedisClient
	logger *infrastructure.Logger
	credit *CreditService
}

func NewPlanService(cfg *config.Config, repos *repository.Repositories, redis *infrastructure.RedisClient, logger *infrastructure.Logger, credit *CreditService) *PlanService {
	return &PlanService{
		cfg:    cfg,
		repos:  repos,
		redis:  redis,
		logger: logger,
		credit: credit,
	}
}

// GetLimits returns the limits for a given plan ID
func (s *PlanService) GetLimits(planID string) domain.PlanLimit {
	switch planID {
	case "pulse":
		return domain.PlanLimit{
			PlanID:               "pulse",
			MaxResponses:         0, // Unlimited via credits
			MaxHandoffs:          0, // Unlimited
			MaxInventoryItems:    0, // Unlimited
			HasNotification:      true,
			PriceNGN:             0, // Pay-as-you-go
			Description:          "Pay-as-you-go response packs",
		}
	case "pro":
		return domain.PlanLimit{
			PlanID:               "pro",
			MaxResponses:         0, // Unlimited
			MaxHandoffs:          0, // Unlimited
			MaxInventoryItems:    0, // Unlimited
			HasNotification:      true,
			PriceNGN:             21999, // ₦21,999/month
			Description:          "Unlimited responses, white-label, priority support",
		}
	case "enterprise":
		return domain.PlanLimit{
			PlanID:               "enterprise",
			MaxResponses:         0, // Unlimited
			MaxHandoffs:          0, // Unlimited
			MaxInventoryItems:    0, // Unlimited
			HasNotification:      true,
			PriceNGN:             0, // Custom pricing
			Description:          "Talk to sales for custom solution",
		}
	default: // free plan
		return domain.PlanLimit{
			PlanID:               "free",
			MaxResponses:         100, // per week
			MaxHandoffs:          0, // Unlimited but no notification
			MaxInventoryItems:    10,
			HasNotification:      false, // No handoff notifications on free plan
			PriceNGN:             0,
			Description:          "100 AI responses/week, limited features",
		}
	}
}

// CanGenerateResponse checks if the user can generate an AI response based on their plan
func (s *PlanService) CanGenerateResponse(ctx context.Context, userID, planID string) (canGenerate bool, reason string, err error) {
	switch planID {
	case "free":
		// Check weekly free counter in Redis
		key := fmt.Sprintf("free_weekly:%s", userID)
		count, err := s.redis.GetInt(ctx, key)
		if err != nil {
			// If key doesn't exist, treat as 0
			if err == infrastructure.ErrRedisNil {
				count = 0
			} else {
				return false, "", err
			}
		}
		
		if count >= 100 {
			return false, "You've exceeded your weekly limit of 100 AI responses. Upgrade your plan for more responses.", nil
		}
		
		// Increment the counter
		if _, err := s.redis.Incr(ctx, key); err != nil {
			return false, "", err
		}
		
		// Set expiry to Monday midnight if this is the first increment of the week
		ttl, ttlErr := s.redis.TTL(ctx, key)
		if ttl < 0 || ttlErr != nil {
			now := time.Now()
			daysUntilMonday := ((7 - int(now.Weekday())) % 7)
			if daysUntilMonday == 0 {
				daysUntilMonday = 7
			}
			midnight := time.Date(now.Year(), now.Month(), now.Day()+daysUntilMonday, 0, 0, 0, 0, time.Local)
			_ = s.redis.SetEx(ctx, key, count+1, time.Until(midnight))
		}
		
		return true, "", nil
	case "pulse":
		// Check if user has sufficient credit balance
		if err := s.credit.Deduct(ctx, userID); err != nil {
			if err.Error() == "insufficient credit balance" {
				return false, "You've run out of response credits. Purchase a response pack to continue using AI responses.", nil
			} else if err.Error() == "credit balance has expired" {
				return false, "Your response credits have expired. Purchase a new pack to continue using AI responses.", nil
			}
			return false, "", err
		}
		return true, "", nil
	case "pro", "enterprise":
		// Unlimited responses
		return true, "", nil
	default:
		return false, "Unknown plan", nil
	}
}

// CanCreateHandoff checks if the user can create a handoff and whether they get notifications
func (s *PlanService) CanCreateHandoff(ctx context.Context, userID, planID string) (canHandoff, hasNotification bool, err error) {
	switch planID {
	case "free":
		// Free plan allows handoffs but does NOT send notifications (conversion friction)
		return true, false, nil
	case "pulse", "pro", "enterprise":
		// Paid plans allow handoffs AND send notifications
		return true, true, nil
	default:
		return false, false, fmt.Errorf("unknown plan: %s", planID)
	}
}

// CanAddInventory checks if the user can add another inventory item based on their plan
func (s *PlanService) CanAddInventory(ctx context.Context, userID, planID string) (allowed bool, remaining int, err error) {
	switch planID {
	case "free":
		// Free plan limited to 10 inventory items
		count, err := s.repos.Inventory.CountByUser(ctx, userID)
		if err != nil {
			return false, 0, err
		}
		if count >= 10 {
			return false, count, fmt.Errorf("free plan limited to 10 inventory items")
		}
		return true, count, nil
	case "pulse", "pro", "enterprise":
		// Paid plans have unlimited inventory
		return true, 0, nil
	default:
		return false, 0, fmt.Errorf("unknown plan: %s", planID)
	}
}

// GetLimitsByUserID fetches the user and returns their plan limits
func (s *PlanService) GetLimitsByUserID(ctx context.Context, userID string) (domain.PlanLimit, error) {
	user, err := s.repos.User.GetByID(ctx, userID)
	if err != nil {
		return domain.PlanLimit{}, err
	}
	if user == nil {
		return domain.PlanLimit{}, fmt.Errorf("user not found")
	}
	return s.GetLimits(user.PlanID), nil
}

// GetFreeWeeklyUsage returns the current weekly usage count for free users
func (s *PlanService) GetFreeWeeklyUsage(ctx context.Context, userID string) (int, error) {
	key := fmt.Sprintf("free_weekly:%s", userID)
	count, err := s.redis.GetInt(ctx, key)
	if err != nil {
		if err == infrastructure.ErrRedisNil {
			return 0, nil
		}
		return 0, err
	}
	return count, nil
}