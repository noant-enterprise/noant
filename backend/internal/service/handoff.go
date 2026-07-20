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

// ========== HANDOFF SERVICE ==========

type HandoffService struct {
	cfg         *config.Config
	repos       *repository.Repositories
	redis       *infrastructure.RedisClient
	logger      *infrastructure.Logger
	broadcastFn func(convID string, msgType string, data interface{})
	planSvc     *PlanService
}

func NewHandoffService(cfg *config.Config, repos *repository.Repositories, redis *infrastructure.RedisClient, logger *infrastructure.Logger, broadcastFn func(convID string, msgType string, data interface{}), planSvc *PlanService) *HandoffService {
	return &HandoffService{cfg: cfg, repos: repos, redis: redis, logger: logger, broadcastFn: broadcastFn, planSvc: planSvc}
}

func (s *HandoffService) Create(ctx context.Context, h *domain.Handoff) error {
	h.Status = "pending"
	if h.Quantity == 0 {
		h.Quantity = 1
	}
	next := time.Now().Add(15 * time.Minute)
	h.NextReminderAt = &next
	if err := s.repos.Handoff.Create(ctx, h); err != nil {
		return err
	}

	// Check if this plan gets notifications
	user, _ := s.repos.User.GetByID(ctx, h.UserID)
	var hasNotification bool
	if user != nil {
		_, hasNotification, _ = s.planSvc.CanCreateHandoff(ctx, user.ID, user.PlanID)
		// For free plan specifically, we know it doesn't get notifications
		if user.PlanID == "free" {
			hasNotification = false
		}
	}

	// Notify owner via WebSocket only if plan allows it
	if hasNotification && s.broadcastFn != nil {
		s.broadcastFn("", "new_handoff", map[string]interface{}{
			"handoff_id":      h.ID,
			"customer_name":   h.CustomerName,
			"product_name":    h.ProductName,
			"agreed_price":    h.AgreedPrice,
			"conversation_id": h.ConversationID,
		})
	}

	// Create notification for owner only if plan allows it
	if hasNotification {
		notif := &domain.Notification{
			UserID: h.UserID,
			Type:   "handoff",
			Title:  "New Sale Handoff",
			Body:   fmt.Sprintf("%s wants to buy %s for ₦%.0f", h.CustomerName, h.ProductName, h.AgreedPrice),
			Link:   "/leads",
			IsRead: false,
		}
		_ = s.repos.Notification.Create(ctx, notif)
	}

	return nil
}

func (s *HandoffService) List(ctx context.Context, userID, status string) ([]domain.Handoff, error) {
	return s.repos.Handoff.List(ctx, userID, status, 100)
}

func (s *HandoffService) GetByID(ctx context.Context, id, userID string) (*domain.Handoff, error) {
	return s.repos.Handoff.GetByID(ctx, id, userID)
}

func (s *HandoffService) UpdateStatus(ctx context.Context, id, userID, status, notes string, finalPrice *float64) error {
	if err := s.repos.Handoff.UpdateStatus(ctx, id, userID, status, notes); err != nil {
		return err
	}
	if status == "sold" && finalPrice != nil {
		// Decrease inventory stock if product
		h, err := s.repos.Handoff.GetByID(ctx, id, userID)
		if err == nil && h != nil {
			items, _ := s.repos.Inventory.Search(ctx, userID, h.ProductName)
			if len(items) > 0 && items[0].StockQuantity != nil {
				_ = s.repos.Inventory.DecreaseStock(ctx, items[0].ID, h.Quantity)
			}
		}
	}
	return nil
}

func (s *HandoffService) ProcessReminders(ctx context.Context) error {
	handoffs, err := s.repos.Handoff.GetReadyForReminder(ctx)
	if err != nil {
		return err
	}
	for i := range handoffs {
		h := &handoffs[i]
		if h.ReminderCount >= 3 {
			_ = s.repos.Handoff.Expire(ctx, h.ID)
			// Auto-reply to customer
			if s.broadcastFn != nil {
				s.broadcastFn(h.ConversationID, "handoff_expired", map[string]interface{}{
					"handoff_id":    h.ID,
					"customer_name": h.CustomerName,
				})
			}
			continue
		}
		_ = s.repos.Handoff.IncrementReminder(ctx, h.ID)
		if s.broadcastFn != nil {
			s.broadcastFn("", "handoff_reminder", map[string]interface{}{
				"handoff_id":     h.ID,
				"customer_name":  h.CustomerName,
				"product_name":   h.ProductName,
				"reminder_count": h.ReminderCount + 1,
			})
		}
		notif := &domain.Notification{
			UserID: h.UserID,
			Type:   "handoff_reminder",
			Title:  "Handoff Reminder",
			Body:   fmt.Sprintf("Follow up with %s about %s", h.CustomerName, h.ProductName),
			Link:   "/leads",
			IsRead: false,
		}
		_ = s.repos.Notification.Create(ctx, notif)
	}
	return nil
}
