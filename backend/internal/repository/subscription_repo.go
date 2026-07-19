package repository

import (
	"context"
	"database/sql"

	"noant/internal/domain"
	"noant/internal/infrastructure"
)

type SubscriptionRepository struct {
	db    *sql.DB
	redis *infrastructure.RedisClient
}

func NewSubscriptionRepository(db *sql.DB, redis *infrastructure.RedisClient) *SubscriptionRepository {
	return &SubscriptionRepository{db: db, redis: redis}
}

func (r *SubscriptionRepository) GetActive(ctx context.Context, userID string) (*domain.Subscription, error) {
	query := `SELECT id, user_id, plan_id, status, current_period_start, current_period_end, created_at, updated_at
	FROM subscriptions WHERE user_id = ? AND status = 'active' ORDER BY created_at DESC LIMIT 1`
	row := r.db.QueryRowContext(ctx, query, userID)
	sub := &domain.Subscription{}
	err := row.Scan(&sub.ID, &sub.UserID, &sub.PlanID, &sub.Status, &sub.CurrentPeriodStart, &sub.CurrentPeriodEnd, &sub.CreatedAt, &sub.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return sub, nil
}

func (r *SubscriptionRepository) Create(ctx context.Context, sub *domain.Subscription) error {
	if sub.ID == "" {
		sub.ID = generateUUID()
	}
	query := `INSERT INTO subscriptions (id, user_id, plan_id, status, current_period_start, current_period_end, created_at, updated_at)
	VALUES (?, ?, ?, ?, ?, ?, NOW(), NOW())`
	_, err := r.db.ExecContext(ctx, query, sub.ID, sub.UserID, sub.PlanID, sub.Status, sub.CurrentPeriodStart, sub.CurrentPeriodEnd)
	return err
}

func (r *SubscriptionRepository) CreateOrUpdate(ctx context.Context, sub *domain.Subscription) error {
	existing, err := r.GetActive(ctx, sub.UserID)
	if err != nil {
		return err
	}
	if existing != nil {
		query := `UPDATE subscriptions SET plan_id = ?, status = ?, current_period_start = ?, current_period_end = ?, updated_at = NOW() WHERE id = ?`
		_, err := r.db.ExecContext(ctx, query, sub.PlanID, sub.Status, sub.CurrentPeriodStart, sub.CurrentPeriodEnd, existing.ID)
		return err
	}
	return r.Create(ctx, sub)
}

func (r *SubscriptionRepository) Cancel(ctx context.Context, userID string) error {
	query := `UPDATE subscriptions SET status = 'cancelled', updated_at = NOW() WHERE user_id = ? AND status = 'active'`
	_, err := r.db.ExecContext(ctx, query, userID)
	return err
}
