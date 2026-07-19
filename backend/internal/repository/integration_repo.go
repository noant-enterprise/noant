package repository

import (
	"context"
	"database/sql"
	"encoding/json"

	"noant/internal/domain"
	"noant/internal/infrastructure"
)

type IntegrationRepository struct {
	db    *sql.DB
	redis *infrastructure.RedisClient
}

func NewIntegrationRepository(db *sql.DB, redis *infrastructure.RedisClient) *IntegrationRepository {
	return &IntegrationRepository{db: db, redis: redis}
}

func (r *IntegrationRepository) Create(ctx context.Context, integration *domain.Integration) error {
	if integration.ID == "" {
		integration.ID = generateUUID()
	}
	query := `INSERT INTO integrations (id, user_id, channel, status, config, webhook_url, created_at, updated_at)
	VALUES (?, ?, ?, ?, ?, ?, NOW(), NOW())`
	configJSON := []byte("{}")
	if integration.Config != nil {
		b, err := json.Marshal(integration.Config)
		if err == nil {
			configJSON = b
		}
	}
	_, err := r.db.ExecContext(ctx, query, integration.ID, integration.UserID, integration.Channel, integration.Status, string(configJSON), integration.WebhookURL)
	return err
}

func (r *IntegrationRepository) ListByUser(ctx context.Context, userID string) ([]domain.Integration, error) {
	query := `SELECT id, user_id, channel, status, config, webhook_url, last_error, created_at, updated_at FROM integrations WHERE user_id = ?`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var integrations []domain.Integration
	for rows.Next() {
		var i domain.Integration
		var configStr string
		err := rows.Scan(&i.ID, &i.UserID, &i.Channel, &i.Status, &configStr, &i.WebhookURL, &i.LastError, &i.CreatedAt, &i.UpdatedAt)
		if err != nil {
			continue
		}
		if configStr != "" && configStr != "{}" {
			_ = json.Unmarshal([]byte(configStr), &i.Config)
		} else {
			i.Config = map[string]interface{}{}
		}
		integrations = append(integrations, i)
	}
	return integrations, nil
}

func (r *IntegrationRepository) ListActive(ctx context.Context) ([]domain.Integration, error) {
	query := `SELECT id, user_id, channel, status, config, webhook_url, last_error, created_at, updated_at FROM integrations WHERE status = 'active'`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var integrations []domain.Integration
	for rows.Next() {
		var i domain.Integration
		var configStr string
		err := rows.Scan(&i.ID, &i.UserID, &i.Channel, &i.Status, &configStr, &i.WebhookURL, &i.LastError, &i.CreatedAt, &i.UpdatedAt)
		if err != nil {
			continue
		}
		if configStr != "" && configStr != "{}" {
			_ = json.Unmarshal([]byte(configStr), &i.Config)
		} else {
			i.Config = map[string]interface{}{}
		}
		integrations = append(integrations, i)
	}
	return integrations, nil
}

func (r *IntegrationRepository) ListByChannel(ctx context.Context, channel string) ([]domain.Integration, error) {
	query := `SELECT id, user_id, channel, status, config, webhook_url, last_error, created_at, updated_at FROM integrations WHERE channel = ?`
	rows, err := r.db.QueryContext(ctx, query, channel)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var integrations []domain.Integration
	for rows.Next() {
		var i domain.Integration
		var configStr string
		err := rows.Scan(&i.ID, &i.UserID, &i.Channel, &i.Status, &configStr, &i.WebhookURL, &i.LastError, &i.CreatedAt, &i.UpdatedAt)
		if err != nil {
			continue
		}
		if configStr != "" && configStr != "{}" {
			_ = json.Unmarshal([]byte(configStr), &i.Config)
		} else {
			i.Config = map[string]interface{}{}
		}
		integrations = append(integrations, i)
	}
	return integrations, nil
}

func (r *IntegrationRepository) UpdateStatus(ctx context.Context, id string, status string, lastError *string) error {
	query := `UPDATE integrations SET status = ?, last_error = ? WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, status, lastError, id)
	return err
}

func (r *IntegrationRepository) GetByUserAndChannel(ctx context.Context, userID, channel string) (*domain.Integration, error) {
	query := `SELECT id, user_id, channel, status, config, webhook_url, last_error, created_at, updated_at 
	FROM integrations WHERE user_id = ? AND channel = ? LIMIT 1`
	var i domain.Integration
	var configStr string
	err := r.db.QueryRowContext(ctx, query, userID, channel).Scan(
		&i.ID, &i.UserID, &i.Channel, &i.Status, &configStr, &i.WebhookURL, &i.LastError, &i.CreatedAt, &i.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	if configStr != "" && configStr != "{}" {
		_ = json.Unmarshal([]byte(configStr), &i.Config)
	} else {
		i.Config = map[string]interface{}{}
	}
	return &i, nil
}

func (r *IntegrationRepository) GetByChannelAndSessionID(ctx context.Context, channel, sessionID string) (*domain.Integration, error) {
	query := `SELECT id, user_id, channel, status, config, webhook_url, last_error, created_at, updated_at
	FROM integrations WHERE channel = ? AND JSON_EXTRACT(config, '$.session_id') = ?`
	var i domain.Integration
	var configStr string
	err := r.db.QueryRowContext(ctx, query, channel, sessionID).Scan(&i.ID, &i.UserID, &i.Channel, &i.Status, &configStr, &i.WebhookURL, &i.LastError, &i.CreatedAt, &i.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	if configStr != "" && configStr != "{}" {
		_ = json.Unmarshal([]byte(configStr), &i.Config)
	} else {
		i.Config = map[string]interface{}{}
	}
	return &i, nil
}

func (r *IntegrationRepository) GetByChannelAndWebhookSecret(ctx context.Context, channel, secret string) (*domain.Integration, error) {
	query := `SELECT id, user_id, channel, status, config, webhook_url, last_error, created_at, updated_at
	FROM integrations WHERE channel = ? AND status IN ('active', 'connected')`
	rows, err := r.db.QueryContext(ctx, query, channel)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var i domain.Integration
		var configStr string
		if err := rows.Scan(&i.ID, &i.UserID, &i.Channel, &i.Status, &configStr, &i.WebhookURL, &i.LastError, &i.CreatedAt, &i.UpdatedAt); err != nil {
			continue
		}
		if configStr != "" && configStr != "{}" {
			_ = json.Unmarshal([]byte(configStr), &i.Config)
		} else {
			i.Config = map[string]interface{}{}
		}
		if cfgSecret, ok := i.Config["webhook_secret"].(string); ok && cfgSecret == secret {
			return &i, nil
		}
	}

	return nil, nil
}

func (r *IntegrationRepository) Update(ctx context.Context, integration *domain.Integration) error {
	query := `UPDATE integrations SET status = ?, config = ?, webhook_url = ?, updated_at = NOW() WHERE id = ?`
	configJSON := []byte("{}")
	if integration.Config != nil {
		b, err := json.Marshal(integration.Config)
		if err == nil {
			configJSON = b
		}
	}
	_, err := r.db.ExecContext(ctx, query, integration.Status, string(configJSON), integration.WebhookURL, integration.ID)
	return err
}

func (r *IntegrationRepository) Disconnect(ctx context.Context, userID, channel string) error {
	query := `UPDATE integrations SET status = 'inactive', updated_at = NOW() WHERE user_id = ? AND channel = ?`
	_, err := r.db.ExecContext(ctx, query, userID, channel)
	return err
}

func (r *IntegrationRepository) CleanupStaleInactive(ctx context.Context, days int) (int64, error) {
	result, err := r.db.ExecContext(ctx,
		`DELETE FROM integrations WHERE status = 'inactive' AND updated_at < NOW() - INTERVAL ? DAY`, days)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
