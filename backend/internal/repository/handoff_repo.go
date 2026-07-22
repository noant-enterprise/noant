package repository

import (
	"context"
	"database/sql"
	"time"

	"noant/internal/domain"
	"noant/internal/infrastructure"
)

type HandoffRepository struct {
	db    *sql.DB
	redis *infrastructure.RedisClient
}

func NewHandoffRepository(db *sql.DB, redis *infrastructure.RedisClient) *HandoffRepository {
	return &HandoffRepository{db: db, redis: redis}
}

func (r *HandoffRepository) Create(ctx context.Context, h *domain.Handoff) error {
	if h.ID == "" {
		h.ID = generateUUID()
	}
	if h.Quantity == 0 {
		h.Quantity = 1
	}
	query := `INSERT INTO handoffs (id, user_id, org_id, conversation_id, customer_name, customer_phone, customer_whatsapp, customer_location, product_name, original_price, agreed_price, quantity, status, created_at, updated_at)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())`
	_, err := r.db.ExecContext(ctx, query, h.ID, h.UserID, h.OrgID, h.ConversationID, h.CustomerName, h.CustomerPhone, h.CustomerWhatsapp, h.CustomerLocation, h.ProductName, h.OriginalPrice, h.AgreedPrice, h.Quantity, h.Status)
	return err
}

func (r *HandoffRepository) GetByID(ctx context.Context, id, orgID string) (*domain.Handoff, error) {
	h := &domain.Handoff{}
	row := r.db.QueryRowContext(ctx, `SELECT id, user_id, conversation_id, customer_name, customer_phone, customer_whatsapp, customer_location, product_name, original_price, agreed_price, quantity, status, final_price, owner_notes, owner_notified_at, reminder_count, next_reminder_at, created_at, updated_at FROM handoffs WHERE id = ? AND org_id = ?`, id, orgID)
	err := row.Scan(&h.ID, &h.UserID, &h.ConversationID, &h.CustomerName, &h.CustomerPhone, &h.CustomerWhatsapp, &h.CustomerLocation, &h.ProductName, &h.OriginalPrice, &h.AgreedPrice, &h.Quantity, &h.Status, &h.FinalPrice, &h.OwnerNotes, &h.OwnerNotifiedAt, &h.ReminderCount, &h.NextReminderAt, &h.CreatedAt, &h.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return h, nil
}

func (r *HandoffRepository) List(ctx context.Context, orgID, status string, limit int) ([]domain.Handoff, error) {
	query := `SELECT id, user_id, conversation_id, customer_name, customer_phone, customer_whatsapp, customer_location, product_name, original_price, agreed_price, quantity, status, final_price, owner_notes, owner_notified_at, reminder_count, next_reminder_at, created_at, updated_at FROM handoffs WHERE org_id = ?`
	args := []interface{}{orgID}
	if status != "" {
		query += " AND status = ?"
		args = append(args, status)
	}
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	query += " ORDER BY created_at DESC LIMIT ?"
	args = append(args, limit)
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var handoffs []domain.Handoff
	for rows.Next() {
		var h domain.Handoff
		if err := rows.Scan(&h.ID, &h.UserID, &h.ConversationID, &h.CustomerName, &h.CustomerPhone, &h.CustomerWhatsapp, &h.CustomerLocation, &h.ProductName, &h.OriginalPrice, &h.AgreedPrice, &h.Quantity, &h.Status, &h.FinalPrice, &h.OwnerNotes, &h.OwnerNotifiedAt, &h.ReminderCount, &h.NextReminderAt, &h.CreatedAt, &h.UpdatedAt); err != nil {
			continue
		}
		handoffs = append(handoffs, h)
	}
	return handoffs, nil
}

func (r *HandoffRepository) UpdateStatus(ctx context.Context, id, orgID, status, notes string) error {
	query := `UPDATE handoffs SET status=?, owner_notes=?, updated_at=NOW() WHERE id=? AND org_id=?`
	_, err := r.db.ExecContext(ctx, query, status, notes, id, orgID)
	return err
}

func (r *HandoffRepository) GetPending(ctx context.Context, orgID string) ([]domain.Handoff, error) {
	return r.List(ctx, orgID, "pending", 100)
}

func (r *HandoffRepository) GetReadyForReminder(ctx context.Context) ([]domain.Handoff, error) {
	query := `SELECT id, user_id, conversation_id, customer_name, customer_phone, customer_whatsapp, customer_location, product_name, original_price, agreed_price, quantity, status, final_price, owner_notes, owner_notified_at, reminder_count, next_reminder_at, created_at, updated_at FROM handoffs WHERE status = 'pending' AND next_reminder_at IS NOT NULL AND next_reminder_at <= NOW() AND reminder_count < 3`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var handoffs []domain.Handoff
	for rows.Next() {
		var h domain.Handoff
		if err := rows.Scan(&h.ID, &h.UserID, &h.ConversationID, &h.CustomerName, &h.CustomerPhone, &h.CustomerWhatsapp, &h.CustomerLocation, &h.ProductName, &h.OriginalPrice, &h.AgreedPrice, &h.Quantity, &h.Status, &h.FinalPrice, &h.OwnerNotes, &h.OwnerNotifiedAt, &h.ReminderCount, &h.NextReminderAt, &h.CreatedAt, &h.UpdatedAt); err != nil {
			continue
		}
		handoffs = append(handoffs, h)
	}
	return handoffs, nil
}

func (r *HandoffRepository) IncrementReminder(ctx context.Context, id string) error {
	next := time.Now().Add(15 * time.Minute)
	_, err := r.db.ExecContext(ctx, "UPDATE handoffs SET reminder_count = reminder_count + 1, next_reminder_at = ?, owner_notified_at = NOW(), updated_at = NOW() WHERE id = ?", next, id)
	return err
}

func (r *HandoffRepository) Expire(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, "UPDATE handoffs SET status = 'expired', updated_at = NOW() WHERE id = ?", id)
	return err
}

func (r *HandoffRepository) CleanupExpired(ctx context.Context, days int) (int64, error) {
	result, err := r.db.ExecContext(ctx,
		`UPDATE handoffs SET status = 'expired', updated_at = NOW()
		 WHERE status = 'pending' AND created_at < NOW() - INTERVAL ? DAY`, days)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
