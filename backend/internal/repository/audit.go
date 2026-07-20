package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"noant/internal/domain"
	"noant/internal/infrastructure"
)

type AuditRepository struct {
	db    *sql.DB
	redis *infrastructure.RedisClient
}

func NewAuditRepository(db *sql.DB, redis *infrastructure.RedisClient) *AuditRepository {
	return &AuditRepository{db: db, redis: redis}
}

func (r *AuditRepository) Create(ctx context.Context, log *domain.AuditLog) error {
	detailsJSON := "{}"
	if log.Details != nil {
		b, _ := json.Marshal(log.Details)
		detailsJSON = string(b)
	}

	query := `INSERT INTO audit_logs (id, user_id, action, resource_type, resource_id, details, ip_address, user_agent, created_at)
			  VALUES (UUID(), ?, ?, ?, ?, ?, ?, ?, NOW())`

	_, err := r.db.ExecContext(ctx, query, log.UserID, log.Action, log.ResourceType, log.ResourceID, detailsJSON, log.IPAddress, log.UserAgent)
	if err != nil {
		return fmt.Errorf("failed to create audit log: %w", err)
	}
	return nil
}

func (r *AuditRepository) ListByUser(ctx context.Context, userID string, limit int) ([]domain.AuditLog, error) {
	query := `SELECT id, user_id, action, resource_type, resource_id, details, ip_address, user_agent, created_at 
			  FROM audit_logs WHERE user_id = ? ORDER BY created_at DESC LIMIT ?`

	rows, err := r.db.QueryContext(ctx, query, userID, limit)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var logs []domain.AuditLog
	for rows.Next() {
		var log domain.AuditLog
		var detailsStr string
		var resourceID, ip, ua sql.NullString
		err := rows.Scan(&log.ID, &log.UserID, &log.Action, &log.ResourceType, &resourceID, &detailsStr, &ip, &ua, &log.CreatedAt)
		if err != nil {
			continue
		}
		if resourceID.Valid {
			log.ResourceID = &resourceID.String
		}
		if ip.Valid {
			log.IPAddress = &ip.String
		}
		if ua.Valid {
			log.UserAgent = &ua.String
		}
		_ = json.Unmarshal([]byte(detailsStr), &log.Details)
		logs = append(logs, log)
	}
	if logs == nil {
		logs = []domain.AuditLog{}
	}
	return logs, nil
}

func (r *AuditRepository) CleanupOld(ctx context.Context, days int) (int64, error) {
	result, err := r.db.ExecContext(ctx,
		`DELETE FROM audit_logs WHERE created_at < NOW() - INTERVAL ? DAY`, days)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
