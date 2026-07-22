package repository

import (
	"context"
	"database/sql"

	"noant/internal/domain"
	"noant/internal/infrastructure"
)

type WhatsAppTemplateRepository struct {
	db *sql.DB
}

func NewWhatsAppTemplateRepository(db *sql.DB, redis *infrastructure.RedisClient) *WhatsAppTemplateRepository {
	return &WhatsAppTemplateRepository{db: db}
}

func (r *WhatsAppTemplateRepository) Create(ctx context.Context, tpl *domain.WhatsAppTemplate) error {
	if tpl.ID == "" {
		tpl.ID = generateUUID()
	}
	query := `INSERT INTO whatsapp_templates (id, user_id, org_id, name, language, category, status, header_type, header_value, body_text, footer_text, buttons, namespace, created_at, updated_at)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())`
	_, err := r.db.ExecContext(ctx, query, tpl.ID, tpl.UserID, tpl.OrgID, tpl.Name, tpl.Language, tpl.Category, tpl.Status, tpl.HeaderType, tpl.HeaderValue, tpl.BodyText, tpl.FooterText, tpl.Buttons, tpl.Namespace)
	return err
}

func (r *WhatsAppTemplateRepository) ListByOrg(ctx context.Context, orgID string) ([]domain.WhatsAppTemplate, error) {
	var templates []domain.WhatsAppTemplate
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, user_id, name, language, category, status, header_type, header_value, body_text, footer_text, buttons, namespace, rejection_reason, created_at, updated_at
		FROM whatsapp_templates WHERE org_id = ? ORDER BY created_at DESC`, orgID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		var t domain.WhatsAppTemplate
		if err := rows.Scan(&t.ID, &t.UserID, &t.Name, &t.Language, &t.Category, &t.Status, &t.HeaderType, &t.HeaderValue, &t.BodyText, &t.FooterText, &t.Buttons, &t.Namespace, &t.RejectionReason, &t.CreatedAt, &t.UpdatedAt); err != nil {
			continue
		}
		templates = append(templates, t)
	}
	return templates, nil
}

func (r *WhatsAppTemplateRepository) GetByID(ctx context.Context, id, orgID string) (*domain.WhatsAppTemplate, error) {
	query := `SELECT id, user_id, name, language, category, status, header_type, header_value, body_text, footer_text, buttons, namespace, rejection_reason, created_at, updated_at
	FROM whatsapp_templates WHERE id = ?`
	args := []interface{}{id}
	if orgID != "" {
		query += " AND org_id = ?"
		args = append(args, orgID)
	}
	var t domain.WhatsAppTemplate
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&t.ID, &t.UserID, &t.Name, &t.Language, &t.Category, &t.Status, &t.HeaderType, &t.HeaderValue, &t.BodyText, &t.FooterText, &t.Buttons, &t.Namespace, &t.RejectionReason, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &t, nil
}

func (r *WhatsAppTemplateRepository) Update(ctx context.Context, tpl *domain.WhatsAppTemplate) error {
	query := `UPDATE whatsapp_templates SET name = ?, language = ?, category = ?, status = ?, header_type = ?, header_value = ?, body_text = ?, footer_text = ?, buttons = ?, namespace = ?, rejection_reason = ?, updated_at = NOW() WHERE id = ? AND org_id = ?`
	_, err := r.db.ExecContext(ctx, query, tpl.Name, tpl.Language, tpl.Category, tpl.Status, tpl.HeaderType, tpl.HeaderValue, tpl.BodyText, tpl.FooterText, tpl.Buttons, tpl.Namespace, tpl.RejectionReason, tpl.ID, tpl.OrgID)
	return err
}

func (r *WhatsAppTemplateRepository) Delete(ctx context.Context, id, orgID string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM whatsapp_templates WHERE id = ? AND org_id = ?`, id, orgID)
	return err
}

func (r *WhatsAppTemplateRepository) GetByStatus(ctx context.Context, status string) ([]domain.WhatsAppTemplate, error) {
	var templates []domain.WhatsAppTemplate
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, user_id, name, language, category, status, header_type, header_value, body_text, footer_text, buttons, namespace, rejection_reason, created_at, updated_at
		FROM whatsapp_templates WHERE status = ?`, status)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		var t domain.WhatsAppTemplate
		if err := rows.Scan(&t.ID, &t.UserID, &t.Name, &t.Language, &t.Category, &t.Status, &t.HeaderType, &t.HeaderValue, &t.BodyText, &t.FooterText, &t.Buttons, &t.Namespace, &t.RejectionReason, &t.CreatedAt, &t.UpdatedAt); err != nil {
			continue
		}
		templates = append(templates, t)
	}
	return templates, nil
}
