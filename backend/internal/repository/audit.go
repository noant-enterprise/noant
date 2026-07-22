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

type AuditFilter struct {
	UserID       string
	Action       string
	ResourceType string
	StartDate    string
	EndDate      string
	Limit        int
	Offset       int
}

type AuditListResult struct {
	Logs  []domain.AuditLog `json:"logs"`
	Total int               `json:"total"`
}

func (r *AuditRepository) ListWithFilters(ctx context.Context, filter *AuditFilter) (*AuditListResult, error) {
	where := "1=1"
	args := []interface{}{}

	if filter.UserID != "" {
		where += " AND user_id = ?"
		args = append(args, filter.UserID)
	}
	if filter.Action != "" {
		where += " AND action LIKE ?"
		args = append(args, "%"+filter.Action+"%")
	}
	if filter.ResourceType != "" {
		where += " AND resource_type LIKE ?"
		args = append(args, "%"+filter.ResourceType+"%")
	}
	if filter.StartDate != "" {
		where += " AND created_at >= ?"
		args = append(args, filter.StartDate)
	}
	if filter.EndDate != "" {
		where += " AND created_at <= ?"
		args = append(args, filter.EndDate)
	}

	// Count total
	countQuery := "SELECT COUNT(*) FROM audit_logs WHERE " + where
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("failed to count audit logs: %w", err)
	}

	// Fetch page
	limit := filter.Limit
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	query := `SELECT id, user_id, action, resource_type, resource_id, details, ip_address, user_agent, created_at 
			  FROM audit_logs WHERE ` + where + ` ORDER BY created_at DESC LIMIT ? OFFSET ?`
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list audit logs: %w", err)
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

	return &AuditListResult{Logs: logs, Total: total}, nil
}
