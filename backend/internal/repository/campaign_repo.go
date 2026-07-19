package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"noant/internal/domain"
	"noant/internal/infrastructure"
)

type CampaignRepository struct {
	db *sql.DB
}

func NewCampaignRepository(db *sql.DB, redis *infrastructure.RedisClient) *CampaignRepository {
	return &CampaignRepository{db: db}
}

func (r *CampaignRepository) Create(ctx context.Context, campaign *domain.CampaignSchedule) error {
	if campaign.ID == "" {
		campaign.ID = fmt.Sprintf("%d", time.Now().UnixNano())
	}
	_, err := r.db.ExecContext(ctx, 
		`INSERT INTO campaign_schedules (id, user_id, name, start_date, end_date, status, created_at, updated_at) 
		 VALUES (?, ?, ?, ?, ?, ?, NOW(), NOW())`,
		campaign.ID, campaign.UserID, campaign.Name, campaign.StartDate, campaign.EndDate, campaign.Status)
	return err
}

func (r *CampaignRepository) ListByUser(ctx context.Context, userID string) ([]domain.CampaignSchedule, error) {
	var campaigns []domain.CampaignSchedule
	rows, err := r.db.QueryContext(ctx, 
		`SELECT id, user_id, name, start_date, end_date, status, created_at, updated_at FROM campaign_schedules WHERE user_id = ? ORDER BY created_at DESC`, 
		userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var c domain.CampaignSchedule
		if err := rows.Scan(&c.ID, &c.UserID, &c.Name, &c.StartDate, &c.EndDate, &c.Status, &c.CreatedAt, &c.UpdatedAt); err != nil {
			continue
		}
		campaigns = append(campaigns, c)
	}
	return campaigns, nil
}

func (r *CampaignRepository) GetScheduledForToday(ctx context.Context) ([]domain.CampaignSchedule, error) {
	var campaigns []domain.CampaignSchedule
	today := time.Now().Format("2006-01-02")
	
	rows, err := r.db.QueryContext(ctx, 
		`SELECT id, user_id, name, start_date, end_date, status, created_at, updated_at FROM campaign_schedules WHERE start_date = ? AND status = 'draft'`, 
		today)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var c domain.CampaignSchedule
		if err := rows.Scan(&c.ID, &c.UserID, &c.Name, &c.StartDate, &c.EndDate, &c.Status, &c.CreatedAt, &c.UpdatedAt); err != nil {
			continue
		}
		campaigns = append(campaigns, c)
	}
	return campaigns, nil
}

func (r *CampaignRepository) GetEndingToday(ctx context.Context) ([]domain.CampaignSchedule, error) {
	var campaigns []domain.CampaignSchedule
	today := time.Now().Format("2006-01-02")
	
	rows, err := r.db.QueryContext(ctx, 
		`SELECT id, user_id, name, start_date, end_date, status, created_at, updated_at FROM campaign_schedules WHERE end_date = ? AND status = 'active'`, 
		today)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var c domain.CampaignSchedule
		if err := rows.Scan(&c.ID, &c.UserID, &c.Name, &c.StartDate, &c.EndDate, &c.Status, &c.CreatedAt, &c.UpdatedAt); err != nil {
			continue
		}
		campaigns = append(campaigns, c)
	}
	return campaigns, nil
}

func (r *CampaignRepository) UpdateStatus(ctx context.Context, id string, status string) error {
	_, err := r.db.ExecContext(ctx, 
		`UPDATE campaign_schedules SET status = ?, updated_at = NOW() WHERE id = ?`, 
		status, id)
	return err
}

func (r *CampaignRepository) CleanupCompleted(ctx context.Context, days int) (int64, error) {
	result, err := r.db.ExecContext(ctx,
		`DELETE FROM campaign_schedules WHERE status IN ('completed', 'cancelled') AND updated_at < NOW() - INTERVAL ? DAY`, days)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
