package repository

import (
	"context"
	"database/sql"
	"fmt"

	"noant/internal/domain"
	"noant/internal/infrastructure"
)

type ConversationRepository struct {
	db    *sql.DB
	redis *infrastructure.RedisClient
}

func NewConversationRepository(db *sql.DB, redis *infrastructure.RedisClient) *ConversationRepository {
	return &ConversationRepository{db: db, redis: redis}
}

func (r *ConversationRepository) GetByID(ctx context.Context, id string) (*domain.Conversation, error) {
	conv := &domain.Conversation{}
	row := r.db.QueryRowContext(ctx, `SELECT id, user_id, customer_name, customer_phone, customer_email, channel, status, intent, priority, is_ai_transferred, taken_over_by, taken_over_at, resolved_at, folder_id, customer_avatar, created_at, updated_at FROM conversations WHERE id = ?`, id)
	err := row.Scan(&conv.ID, &conv.UserID, &conv.CustomerName, &conv.CustomerPhone, &conv.CustomerEmail, &conv.Channel, &conv.Status, &conv.Intent, &conv.Priority, &conv.IsAITransferred, &conv.TakenOverBy, &conv.TakenOverAt, &conv.ResolvedAt, &conv.FolderID, &conv.CustomerAvatar, &conv.CreatedAt, &conv.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return conv, nil
}

func (r *ConversationRepository) Create(ctx context.Context, conv *domain.Conversation) error {
	if conv.ID == "" {
		conv.ID = generateUUID()
	}
	query := `INSERT INTO conversations (id, user_id, customer_name, customer_phone, customer_email, channel, status, intent, priority, is_ai_transferred, customer_avatar, created_at, updated_at)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())`
	_, err := r.db.ExecContext(ctx, query, conv.ID, conv.UserID, conv.CustomerName, conv.CustomerPhone, conv.CustomerEmail, conv.Channel, conv.Status, conv.Intent, conv.Priority, conv.IsAITransferred, conv.CustomerAvatar)
	if err != nil {
		return fmt.Errorf("failed to create conversation: %w", err)
	}
	return nil
}

func (r *ConversationRepository) List(ctx context.Context, userID, status string, limit, offset int) ([]domain.Conversation, int, error) {
	countQuery := "SELECT COUNT(*) FROM conversations WHERE user_id = ?"
	countArgs := []interface{}{userID}
	if status != "" {
		countQuery += " AND status = ?"
		countArgs = append(countArgs, status)
	}
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, countArgs...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}
	query := `SELECT id, user_id, customer_name, customer_phone, customer_email, channel, status, intent, priority, is_ai_transferred, taken_over_by, taken_over_at, resolved_at, folder_id, customer_avatar, created_at, updated_at
	FROM conversations WHERE user_id = ?`
	args := []interface{}{userID}
	if status != "" {
		query += " AND status = ?"
		args = append(args, status)
	}
	query += " ORDER BY updated_at DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()
	var conversations []domain.Conversation
	for rows.Next() {
		var conv domain.Conversation
		err := rows.Scan(&conv.ID, &conv.UserID, &conv.CustomerName, &conv.CustomerPhone, &conv.CustomerEmail, &conv.Channel, &conv.Status, &conv.Intent, &conv.Priority, &conv.IsAITransferred, &conv.TakenOverBy, &conv.TakenOverAt, &conv.ResolvedAt, &conv.FolderID, &conv.CustomerAvatar, &conv.CreatedAt, &conv.UpdatedAt)
		if err != nil {
			continue
		}
		conversations = append(conversations, conv)
	}
	return conversations, total, nil
}

func (r *ConversationRepository) GetByIDAndUser(ctx context.Context, id, userID string) (*domain.Conversation, error) {
	query := `SELECT id, user_id, customer_name, customer_phone, customer_email, channel, status, intent, priority, is_ai_transferred, taken_over_by, taken_over_at, resolved_at, folder_id, customer_avatar, created_at, updated_at FROM conversations WHERE id = ? AND user_id = ?`
	row := r.db.QueryRowContext(ctx, query, id, userID)
	conv := &domain.Conversation{}
	err := row.Scan(&conv.ID, &conv.UserID, &conv.CustomerName, &conv.CustomerPhone, &conv.CustomerEmail, &conv.Channel, &conv.Status, &conv.Intent, &conv.Priority, &conv.IsAITransferred, &conv.TakenOverBy, &conv.TakenOverAt, &conv.ResolvedAt, &conv.FolderID, &conv.CustomerAvatar, &conv.CreatedAt, &conv.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return conv, nil
}

func (r *ConversationRepository) UpdateStatus(ctx context.Context, id, userID, status string) error {
	query := `UPDATE conversations SET status = ? WHERE id = ? AND user_id = ?`
	_, err := r.db.ExecContext(ctx, query, status, id, userID)
	return err
}

func (r *ConversationRepository) UpdateCustomerInfo(ctx context.Context, id, name, avatar string) error {
	query := `UPDATE conversations SET customer_name = ?, customer_avatar = ? WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, name, avatar, id)
	return err
}

func (r *ConversationRepository) FindActiveByCustomer(ctx context.Context, userID, customerName, channel string) (*domain.Conversation, error) {
	query := `SELECT id, user_id, customer_name, customer_phone, customer_email, channel, status, intent, priority, is_ai_transferred, taken_over_by, taken_over_at, resolved_at, folder_id, customer_avatar, created_at, updated_at FROM conversations WHERE user_id = ? AND (customer_phone = ? OR customer_name = ?) AND channel = ? AND status = 'active' ORDER BY created_at DESC LIMIT 1`
	row := r.db.QueryRowContext(ctx, query, userID, customerName, customerName, channel)
	conv := &domain.Conversation{}
	err := row.Scan(&conv.ID, &conv.UserID, &conv.CustomerName, &conv.CustomerPhone, &conv.CustomerEmail, &conv.Channel, &conv.Status, &conv.Intent, &conv.Priority, &conv.IsAITransferred, &conv.TakenOverBy, &conv.TakenOverAt, &conv.ResolvedAt, &conv.FolderID, &conv.CustomerAvatar, &conv.CreatedAt, &conv.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return conv, nil
}

func (r *ConversationRepository) Takeover(ctx context.Context, id, userID, agentID string) error {
	query := `UPDATE conversations SET status = 'escalated', taken_over_by = ?, taken_over_at = NOW() WHERE id = ? AND user_id = ?`
	_, err := r.db.ExecContext(ctx, query, agentID, id, userID)
	return err
}

func (r *ConversationRepository) ClearChats(ctx context.Context, userID string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	// 1. Delete messages
	_, err = tx.ExecContext(ctx, `DELETE FROM messages WHERE conversation_id IN (SELECT id FROM conversations WHERE user_id = ?)`, userID)
	if err != nil {
		return err
	}

	// 2. Delete conversations
	_, err = tx.ExecContext(ctx, `DELETE FROM conversations WHERE user_id = ?`, userID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// ========== ANALYTICS REPOSITORY METHODS ==========

func (r *ConversationRepository) GetOverview(ctx context.Context, userID string) (map[string]interface{}, error) {
	query := "SELECT COUNT(*) as total, COALESCE(SUM(CASE WHEN DATE(created_at) = CURDATE() THEN 1 ELSE 0 END), 0) as conversations_today, COALESCE(SUM(CASE WHEN status = 'active' THEN 1 ELSE 0 END), 0) as active, COALESCE(SUM(CASE WHEN status = 'resolved' AND DATE(resolved_at) = CURDATE() THEN 1 ELSE 0 END), 0) as resolved_today, COALESCE(SUM(CASE WHEN is_ai_transferred = false THEN 1 ELSE 0 END), 0) as ai_handled, COALESCE(COUNT(DISTINCT CASE WHEN status = 'escalated' THEN id END), 0) as escalated FROM conversations WHERE user_id = ?"
	row := r.db.QueryRowContext(ctx, query, userID)
	var total, conversationsToday, active, resolvedToday, aiHandled, escalated int
	err := row.Scan(&total, &conversationsToday, &active, &resolvedToday, &aiHandled, &escalated)
	if err != nil {
		return nil, err
	}
	aiRate := 0.0
	if total > 0 {
		aiRate = float64(aiHandled) / float64(total)
	}
	return map[string]interface{}{
		"total_conversations":  total,
		"conversations_today":  conversationsToday,
		"active_conversations": active,
		"resolved_today":       resolvedToday,
		"ai_resolution_rate":   aiRate,
		"escalated_count":      escalated,
	}, nil
}

func (r *ConversationRepository) CountByChannel(ctx context.Context, userID string) (map[string]int, error) {
	query := "SELECT channel, COUNT(*) as count FROM conversations WHERE user_id = ? GROUP BY channel"
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	result := make(map[string]int)
	for rows.Next() {
		var channel string
		var count int
		if err := rows.Scan(&channel, &count); err == nil {
			result[channel] = count
		}
	}
	return result, nil
}

func (r *ConversationRepository) CountByIntent(ctx context.Context, userID string) ([]map[string]interface{}, error) {
	query := "SELECT intent, COUNT(*) as count FROM conversations WHERE user_id = ? GROUP BY intent ORDER BY count DESC LIMIT 5"
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var result []map[string]interface{}
	for rows.Next() {
		var intent string
		var count int
		if err := rows.Scan(&intent, &count); err == nil {
			result = append(result, map[string]interface{}{"intent": intent, "count": count})
		}
	}
	return result, nil
}

func (r *ConversationRepository) CountByHour(ctx context.Context, userID string) ([]map[string]interface{}, error) {
	query := "SELECT HOUR(created_at) as hour, COUNT(*) as count FROM conversations WHERE user_id = ? AND created_at >= DATE_SUB(NOW(), INTERVAL 7 DAY) GROUP BY HOUR(created_at) ORDER BY hour"
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var result []map[string]interface{}
	for rows.Next() {
		var hour, count int
		if err := rows.Scan(&hour, &count); err == nil {
			hourStr := fmt.Sprintf("%02d:00", hour)
			result = append(result, map[string]interface{}{"hour": hourStr, "volume": count})
		}
	}
	return result, nil
}

func (r *ConversationRepository) CountByDate(ctx context.Context, userID string, days int) ([]map[string]interface{}, error) {
	query := "SELECT DATE(created_at) as date, COUNT(*) as count FROM conversations WHERE user_id = ? AND created_at >= DATE_SUB(NOW(), INTERVAL ? DAY) GROUP BY DATE(created_at) ORDER BY date"
	rows, err := r.db.QueryContext(ctx, query, userID, days)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var result []map[string]interface{}
	for rows.Next() {
		var date string
		var count int
		if err := rows.Scan(&date, &count); err == nil {
			// Format date string to YYYY-MM-DD cleanly
			if len(date) > 10 {
				date = date[:10]
			}
			result = append(result, map[string]interface{}{"date": date, "conversations": count})
		}
	}
	return result, nil
}

// ========== CSAT RATINGS ==========

func (r *ConversationRepository) RecordCSAT(ctx context.Context, userID, conversationID string, score int, comment *string) error {
	query := `INSERT INTO csat_ratings (id, user_id, conversation_id, score, comment, created_at) VALUES (?, ?, ?, ?, ?, NOW())`
	_, err := r.db.ExecContext(ctx, query, generateUUID(), userID, conversationID, score, comment)
	return err
}

func (r *ConversationRepository) GetCSATAverage(ctx context.Context, userID string) (avg float64, total int, err error) {
	var avgN sql.NullFloat64
	err = r.db.QueryRowContext(ctx, `SELECT AVG(score), COUNT(*) FROM csat_ratings WHERE user_id = ?`, userID).Scan(&avgN, &total)
	if err != nil {
		return 0, 0, err
	}
	return avgN.Float64, total, nil
}

func (r *ConversationRepository) GetCSATDistribution(ctx context.Context, userID string) (map[int]int, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT score, COUNT(*) as count FROM csat_ratings WHERE user_id = ? GROUP BY score ORDER BY score`, userID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	dist := map[int]int{1: 0, 2: 0, 3: 0, 4: 0, 5: 0}
	for rows.Next() {
		var score, count int
		if err := rows.Scan(&score, &count); err == nil {
			dist[score] = count
		}
	}
	return dist, nil
}

func (r *ConversationRepository) GetCSATTrend(ctx context.Context, userID string, days int) ([]map[string]interface{}, error) {
	query := `SELECT DATE(created_at) as date, AVG(score) as avg_score, COUNT(*) as count
		FROM csat_ratings WHERE user_id = ? AND created_at >= DATE_SUB(NOW(), INTERVAL ? DAY)
		GROUP BY DATE(created_at) ORDER BY date`
	rows, err := r.db.QueryContext(ctx, query, userID, days)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var result []map[string]interface{}
	for rows.Next() {
		var dateStr string
		var avg float64
		var count int
		if err := rows.Scan(&dateStr, &avg, &count); err == nil {
			if len(dateStr) > 10 {
				dateStr = dateStr[:10]
			}
			result = append(result, map[string]interface{}{"date": dateStr, "avg_score": avg, "count": count})
		}
	}
	return result, nil
}

func (r *ConversationRepository) CountMessagesByDate(ctx context.Context, userID string, days int) ([]map[string]interface{}, error) {
	query := `SELECT DATE(m.created_at) as date, COUNT(*) as count
		FROM messages m JOIN conversations c ON m.conversation_id = c.id
		WHERE c.user_id = ? AND m.created_at >= DATE_SUB(NOW(), INTERVAL ? DAY)
		GROUP BY DATE(m.created_at) ORDER BY date`
	rows, err := r.db.QueryContext(ctx, query, userID, days)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var result []map[string]interface{}
	for rows.Next() {
		var dateStr string
		var count int
		if err := rows.Scan(&dateStr, &count); err == nil {
			if len(dateStr) > 10 {
				dateStr = dateStr[:10]
			}
			result = append(result, map[string]interface{}{"date": dateStr, "messages": count})
		}
	}
	return result, nil
}

func (r *ConversationRepository) GetUptimeStats(ctx context.Context, userID string) (int, error) {
	// Count active days in last 30 days where the user had conversations
	var activeDays int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(DISTINCT DATE(created_at)) FROM conversations WHERE user_id = ? AND created_at >= DATE_SUB(NOW(), INTERVAL 30 DAY)`, userID).Scan(&activeDays)
	if err != nil {
		return 0, err
	}
	return activeDays, nil
}

func (r *ConversationRepository) CleanupOldResolved(ctx context.Context, days int) (int64, error) {
	result, err := r.db.ExecContext(ctx,
		`DELETE FROM conversations WHERE status = 'resolved' AND updated_at < NOW() - INTERVAL ? DAY`, days)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (r *ConversationRepository) CleanupAbandoned(ctx context.Context, days int) (int64, error) {
	result, err := r.db.ExecContext(ctx,
		`UPDATE conversations SET status = 'resolved', resolved_at = NOW()
		 WHERE status = 'active' AND updated_at < NOW() - INTERVAL ? DAY
		 AND id NOT IN (SELECT DISTINCT conversation_id FROM messages WHERE created_at > NOW() - INTERVAL ? DAY)`,
		days, days)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
