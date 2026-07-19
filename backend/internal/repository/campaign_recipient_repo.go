package repository

import (
	"context"
	"database/sql"

	"noant/internal/domain"
	"noant/internal/infrastructure"
)

type CampaignRecipientRepository struct {
	db *sql.DB
}

func NewCampaignRecipientRepository(db *sql.DB, redis *infrastructure.RedisClient) *CampaignRecipientRepository {
	return &CampaignRecipientRepository{db: db}
}

func (r *CampaignRecipientRepository) Create(ctx context.Context, cr *domain.CampaignRecipient) error {
	if cr.ID == "" {
		cr.ID = generateUUID()
	}
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO campaign_recipients (id, campaign_id, user_id, phone, name, status, created_at)
		VALUES (?, ?, ?, ?, ?, ?, NOW())`,
		cr.ID, cr.CampaignID, cr.UserID, cr.Phone, cr.Name, cr.Status)
	return err
}

func (r *CampaignRecipientRepository) ListByCampaign(ctx context.Context, campaignID string) ([]domain.CampaignRecipient, error) {
	var recipients []domain.CampaignRecipient
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, campaign_id, user_id, phone, name, status, error, sent_at, delivered_at, read_at, created_at
		FROM campaign_recipients WHERE campaign_id = ? ORDER BY created_at`, campaignID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var cr domain.CampaignRecipient
		if err := rows.Scan(&cr.ID, &cr.CampaignID, &cr.UserID, &cr.Phone, &cr.Name, &cr.Status, &cr.Error, &cr.SentAt, &cr.DeliveredAt, &cr.ReadAt, &cr.CreatedAt); err != nil {
			continue
		}
		recipients = append(recipients, cr)
	}
	return recipients, nil
}

func (r *CampaignRecipientRepository) UpdateStatus(ctx context.Context, id, status string, errInfo *string) error {
	query := `UPDATE campaign_recipients SET status = ?, error = ?`
	args := []interface{}{status, errInfo}
	if status == "sent" {
		query += ", sent_at = NOW()"
	} else if status == "delivered" {
		query += ", delivered_at = NOW()"
	} else if status == "read" {
		query += ", read_at = NOW()"
	}
	query += " WHERE id = ?"
	args = append(args, id)
	_, err := r.db.ExecContext(ctx, query, args...)
	return err
}

func (r *CampaignRecipientRepository) MarkOptedOut(ctx context.Context, userID, phone string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE campaign_recipients SET status = 'opted_out' WHERE user_id = ? AND phone = ? AND status IN ('pending', 'sent')`,
		userID, phone)
	return err
}

func (r *CampaignRecipientRepository) IsOptedOut(ctx context.Context, userID, phone string) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM campaign_recipients WHERE user_id = ? AND phone = ? AND status = 'opted_out'`,
		userID, phone).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
