package repository

import (
	"context"
	"database/sql"

	"noant/internal/domain"
	"noant/internal/infrastructure"
)

type PushSubscriptionRepository struct {
	db    *sql.DB
	redis *infrastructure.RedisClient
}

func NewPushSubscriptionRepository(db *sql.DB, redis *infrastructure.RedisClient) *PushSubscriptionRepository {
	return &PushSubscriptionRepository{db: db, redis: redis}
}

func (r *PushSubscriptionRepository) Create(ctx context.Context, sub *domain.PushSubscription) error {
	if sub.ID == "" {
		sub.ID = generateUUID()
	}
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO push_subscriptions (id, user_id, endpoint, auth, p256dh, user_agent, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, NOW(), NOW())
		ON DUPLICATE KEY UPDATE auth=VALUES(auth), p256dh=VALUES(p256dh), user_agent=VALUES(user_agent), updated_at=NOW()`,
		sub.ID, sub.UserID, sub.Endpoint, sub.Auth, sub.P256dh, sub.UserAgent)
	return err
}

func (r *PushSubscriptionRepository) Delete(ctx context.Context, userID, endpoint string) error {
	_, err := r.db.ExecContext(ctx, `
		DELETE FROM push_subscriptions WHERE user_id = ? AND endpoint = ?`, userID, endpoint)
	return err
}

func (r *PushSubscriptionRepository) DeleteAllByUser(ctx context.Context, userID string) error {
	_, err := r.db.ExecContext(ctx, `
		DELETE FROM push_subscriptions WHERE user_id = ?`, userID)
	return err
}

func (r *PushSubscriptionRepository) ListByUser(ctx context.Context, userID string) ([]*domain.PushSubscription, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, user_id, endpoint, auth, p256dh, user_agent, created_at, updated_at
		FROM push_subscriptions WHERE user_id = ?`, userID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var results []*domain.PushSubscription
	for rows.Next() {
		sub := &domain.PushSubscription{}
		var userAgent sql.NullString
		if err := rows.Scan(&sub.ID, &sub.UserID, &sub.Endpoint, &sub.Auth, &sub.P256dh, &userAgent, &sub.CreatedAt, &sub.UpdatedAt); err != nil {
			continue
		}
		if userAgent.Valid {
			sub.UserAgent = userAgent.String
		}
		results = append(results, sub)
	}
	return results, nil
}

func (r *PushSubscriptionRepository) ListByUserIDs(ctx context.Context, userIDs []string) ([]*domain.PushSubscription, error) {
	if len(userIDs) == 0 {
		return nil, nil
	}
	query := `SELECT id, user_id, endpoint, auth, p256dh, user_agent, created_at, updated_at
		FROM push_subscriptions WHERE user_id IN (` + placeholders(len(userIDs)) + `)`
	args := make([]interface{}, len(userIDs))
	for i, id := range userIDs {
		args[i] = id
	}
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var results []*domain.PushSubscription
	for rows.Next() {
		sub := &domain.PushSubscription{}
		var userAgent sql.NullString
		if err := rows.Scan(&sub.ID, &sub.UserID, &sub.Endpoint, &sub.Auth, &sub.P256dh, &userAgent, &sub.CreatedAt, &sub.UpdatedAt); err != nil {
			continue
		}
		if userAgent.Valid {
			sub.UserAgent = userAgent.String
		}
		results = append(results, sub)
	}
	return results, nil
}

func (r *PushSubscriptionRepository) DeleteByID(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM push_subscriptions WHERE id = ?`, id)
	return err
}

func placeholders(n int) string {
	if n <= 0 {
		return ""
	}
	b := make([]byte, 0, n*2-1)
	for i := 0; i < n; i++ {
		if i > 0 {
			b = append(b, ',')
		}
		b = append(b, '?')
	}
	return string(b)
}
