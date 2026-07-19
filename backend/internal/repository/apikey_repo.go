package repository

import (
	"context"
	"database/sql"

	"noant/internal/domain"
	"noant/internal/infrastructure"
)

type APIKeyRepository struct {
	db    *sql.DB
	redis *infrastructure.RedisClient
}

func NewAPIKeyRepository(db *sql.DB, redis *infrastructure.RedisClient) *APIKeyRepository {
	return &APIKeyRepository{db: db, redis: redis}
}

func (r *APIKeyRepository) Create(ctx context.Context, key *domain.APIKey) error {
	if key.ID == "" {
		key.ID = generateUUID()
	}
	query := `INSERT INTO api_keys (id, user_id, name, key_hash, is_active, created_at)
	VALUES (?, ?, ?, ?, ?, NOW())`
	_, err := r.db.ExecContext(ctx, query, key.ID, key.UserID, key.Name, key.Key, key.IsActive)
	return err
}

func (r *APIKeyRepository) ListByUser(ctx context.Context, userID string) ([]domain.APIKey, error) {
	query := `SELECT id, user_id, name, key_hash, last_used, is_active, created_at FROM api_keys WHERE user_id = ? AND is_active = true`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var keys []domain.APIKey
	for rows.Next() {
		var k domain.APIKey
		err := rows.Scan(&k.ID, &k.UserID, &k.Name, &k.Key, &k.LastUsed, &k.IsActive, &k.CreatedAt)
		if err != nil {
			continue
		}
		keys = append(keys, k)
	}
	return keys, nil
}

func (r *APIKeyRepository) Revoke(ctx context.Context, id string, userID string) error {
	query := `UPDATE api_keys SET is_active = false WHERE id = ? AND user_id = ?`
	_, err := r.db.ExecContext(ctx, query, id, userID)
	return err
}
