package repository

import (
	"context"
	"database/sql"
	"time"

	"noant/internal/domain"
	"noant/internal/infrastructure"
)

// ========== NOTIFICATION REPOSITORY ==========

type NotificationRepository struct {
	db    *sql.DB
	redis *infrastructure.RedisClient
}

func NewNotificationRepository(db *sql.DB, redis *infrastructure.RedisClient) *NotificationRepository {
	return &NotificationRepository{db: db, redis: redis}
}

func (r *NotificationRepository) Create(ctx context.Context, n *domain.Notification) error {
	if n.ID == "" {
		n.ID = generateUUID()
	}
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO notifications (id, user_id, type, title, body, link, is_read, created_at)
		VALUES (?, ?, ?, ?, ?, ?, FALSE, NOW())`,
		n.ID, n.UserID, n.Type, n.Title, n.Body, n.Link)
	return err
}

func (r *NotificationRepository) ListByUser(ctx context.Context, userID string, limit int) ([]*domain.Notification, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, user_id, type, title, body, link, is_read, created_at
		FROM notifications WHERE user_id = ?
		ORDER BY created_at DESC LIMIT ?`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*domain.Notification
	for rows.Next() {
		n := &domain.Notification{}
		var link sql.NullString
		if err := rows.Scan(&n.ID, &n.UserID, &n.Type, &n.Title, &n.Body, &link, &n.IsRead, &n.CreatedAt); err != nil {
			return nil, err
		}
		if link.Valid {
			n.Link = link.String
		}
		results = append(results, n)
	}
	return results, nil
}

func (r *NotificationRepository) UnreadCount(ctx context.Context, userID string) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM notifications WHERE user_id = ? AND is_read = FALSE`, userID).Scan(&count)
	return count, err
}

func (r *NotificationRepository) MarkRead(ctx context.Context, id, userID string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE notifications SET is_read = TRUE WHERE id = ? AND user_id = ?`, id, userID)
	return err
}

func (r *NotificationRepository) MarkAllRead(ctx context.Context, userID string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE notifications SET is_read = TRUE WHERE user_id = ?`, userID)
	return err
}

// ========== WIDGET CONFIG REPOSITORY ==========

type WidgetConfigRepository struct {
	db    *sql.DB
	redis *infrastructure.RedisClient
}

func NewWidgetConfigRepository(db *sql.DB, redis *infrastructure.RedisClient) *WidgetConfigRepository {
	return &WidgetConfigRepository{db: db, redis: redis}
}

func (r *WidgetConfigRepository) Get(ctx context.Context, userID string) (*domain.WidgetConfig, error) {
	cfg := &domain.WidgetConfig{}
	err := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, brand_color, greeting, bot_name, position, widget_api_key, is_active, created_at, updated_at
		FROM widget_configs WHERE user_id = ?`, userID).Scan(
		&cfg.ID, &cfg.UserID, &cfg.BrandColor, &cfg.Greeting, &cfg.BotName, &cfg.Position,
		&cfg.WidgetAPIKey, &cfg.IsActive, &cfg.CreatedAt, &cfg.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return cfg, err
}

func (r *WidgetConfigRepository) GetByAPIKey(ctx context.Context, apiKey string) (*domain.WidgetConfig, error) {
	cfg := &domain.WidgetConfig{}
	err := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, brand_color, greeting, bot_name, position, widget_api_key, is_active, created_at, updated_at
		FROM widget_configs WHERE widget_api_key = ? AND is_active = TRUE`, apiKey).Scan(
		&cfg.ID, &cfg.UserID, &cfg.BrandColor, &cfg.Greeting, &cfg.BotName, &cfg.Position,
		&cfg.WidgetAPIKey, &cfg.IsActive, &cfg.CreatedAt, &cfg.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return cfg, err
}

func (r *WidgetConfigRepository) Upsert(ctx context.Context, cfg *domain.WidgetConfig) error {
	if cfg.ID == "" {
		cfg.ID = generateUUID()
	}
	cfg.UpdatedAt = time.Now()
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO widget_configs (id, user_id, brand_color, greeting, bot_name, position, widget_api_key, is_active, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, TRUE, NOW(), NOW())
		ON DUPLICATE KEY UPDATE brand_color=VALUES(brand_color), greeting=VALUES(greeting),
		bot_name=VALUES(bot_name), position=VALUES(position), updated_at=NOW()`,
		cfg.ID, cfg.UserID, cfg.BrandColor, cfg.Greeting, cfg.BotName, cfg.Position, cfg.WidgetAPIKey)
	return err
}

// ========== NOTIFICATION PREFERENCES ==========

type NotifPrefs struct {
	Escalation      bool `json:"escalation" db:"notif_escalation"`
	UnknownQs       bool `json:"unknown_questions" db:"notif_unknown_questions"`
	Payment         bool `json:"payment" db:"notif_payment"`
	Security        bool `json:"security" db:"notif_security"`
	TeamInvite      bool `json:"team_invite" db:"notif_team_invite"`
	LanguagePref    string `json:"language_preference" db:"language_preference"`
}

func (r *UserRepository) GetNotifPrefs(ctx context.Context, userID string) (*NotifPrefs, error) {
	prefs := &NotifPrefs{}
	err := r.db.QueryRowContext(ctx, `
		SELECT notif_escalation, notif_unknown_questions, notif_payment, notif_security, notif_team_invite, language_preference
		FROM users WHERE id = ?`, userID).Scan(
		&prefs.Escalation, &prefs.UnknownQs, &prefs.Payment, &prefs.Security,
		&prefs.TeamInvite, &prefs.LanguagePref)
	if err != nil {
		// Default all true if columns don't exist yet
		return &NotifPrefs{Escalation: true, UnknownQs: true, Payment: true, Security: true, TeamInvite: true, LanguagePref: "en"}, nil
	}
	return prefs, nil
}

func (r *UserRepository) UpdateNotifPrefs(ctx context.Context, userID string, prefs *NotifPrefs) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE users SET notif_escalation=?, notif_unknown_questions=?, notif_payment=?,
		notif_security=?, notif_team_invite=?, language_preference=? WHERE id=?`,
		prefs.Escalation, prefs.UnknownQs, prefs.Payment, prefs.Security,
		prefs.TeamInvite, prefs.LanguagePref, userID)
	return err
}

func (r *UserRepository) Delete(ctx context.Context, userID string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM users WHERE id = ?`, userID)
	return err
}

// ExportUserData returns all data belonging to a user for GDPR export
func (r *UserRepository) ExportUserData(ctx context.Context, userID string) (map[string]interface{}, error) {
	user := &domain.User{}
	err := r.db.QueryRowContext(ctx, `
		SELECT id, email, first_name, last_name, company_name, phone, plan_id, is_active, created_at, updated_at
		FROM users WHERE id = ?`, userID).Scan(
		&user.ID, &user.Email, &user.FirstName, &user.LastName, &user.CompanyName,
		&user.Phone, &user.PlanID, &user.IsActive, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"user":        user,
		"exported_at": time.Now().UTC().Format(time.RFC3339),
	}, nil
}

func (r *UserRepository) UpdateProfile(ctx context.Context, userID string, firstName, lastName, companyName, phone string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE users SET first_name = ?, last_name = ?, company_name = ?, phone = ?, updated_at = NOW()
		WHERE id = ?`, firstName, lastName, companyName, phone, userID)
	return err
}
