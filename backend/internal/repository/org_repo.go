package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"noant/internal/domain"
	"noant/internal/infrastructure"
)

type OrgRepository struct {
	db    *sql.DB
	redis *infrastructure.RedisClient
}

func NewOrgRepository(db *sql.DB, redis *infrastructure.RedisClient) *OrgRepository {
	return &OrgRepository{db: db, redis: redis}
}

func (r *OrgRepository) Create(ctx context.Context, org *domain.Organization) error {
	settingsJSON := "{}"
	if org.Settings != nil {
		b, _ := json.Marshal(org.Settings)
		settingsJSON = string(b)
	}
	query := `INSERT INTO organizations (id, name, slug, owner_id, plan_id, settings, created_at, updated_at)
			  VALUES (?, ?, ?, ?, ?, ?, NOW(), NOW())`
	_, err := r.db.ExecContext(ctx, query, org.ID, org.Name, org.Slug, org.OwnerID, org.PlanID, settingsJSON)
	if err != nil {
		return fmt.Errorf("failed to create organization: %w", err)
	}
	return nil
}

func (r *OrgRepository) GetByID(ctx context.Context, id string) (*domain.Organization, error) {
	query := `SELECT id, name, slug, owner_id, plan_id, settings, created_at, updated_at
			  FROM organizations WHERE id = ?`
	row := r.db.QueryRowContext(ctx, query, id)
	var org domain.Organization
	var settingsStr string
	err := row.Scan(&org.ID, &org.Name, &org.Slug, &org.OwnerID, &org.PlanID, &settingsStr, &org.CreatedAt, &org.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	_ = json.Unmarshal([]byte(settingsStr), &org.Settings)
	return &org, nil
}

func (r *OrgRepository) GetByOwnerID(ctx context.Context, ownerID string) (*domain.Organization, error) {
	query := `SELECT id, name, slug, owner_id, plan_id, settings, created_at, updated_at
			  FROM organizations WHERE owner_id = ? LIMIT 1`
	row := r.db.QueryRowContext(ctx, query, ownerID)
	var org domain.Organization
	var settingsStr string
	err := row.Scan(&org.ID, &org.Name, &org.Slug, &org.OwnerID, &org.PlanID, &settingsStr, &org.CreatedAt, &org.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	_ = json.Unmarshal([]byte(settingsStr), &org.Settings)
	return &org, nil
}

func (r *OrgRepository) Update(ctx context.Context, org *domain.Organization) error {
	settingsJSON := "{}"
	if org.Settings != nil {
		b, _ := json.Marshal(org.Settings)
		settingsJSON = string(b)
	}
	query := `UPDATE organizations SET name = ?, plan_id = ?, settings = ?, updated_at = NOW() WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, org.Name, org.PlanID, settingsJSON, org.ID)
	return err
}
