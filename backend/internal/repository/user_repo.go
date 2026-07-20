package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"noant/internal/domain"
	"noant/internal/infrastructure"
)

type UserRepository struct {
	db    *sql.DB
	redis *infrastructure.RedisClient
}

func NewUserRepository(db *sql.DB, redis *infrastructure.RedisClient) *UserRepository {
	return &UserRepository{db: db, redis: redis}
}

func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	if user.ID == "" {
		user.ID = generateUUID()
	}
	query := `INSERT INTO users (id, email, password_hash, first_name, last_name, role, company_name, phone, plan_id, is_active, must_change_password, is_verified, verification_code, created_at, updated_at)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())`
	_, err := r.db.ExecContext(ctx, query, user.ID, user.Email, user.Password, user.FirstName, user.LastName, user.Role, user.CompanyName, user.Phone, user.PlanID, user.IsActive, user.MustChangePassword, user.IsVerified, user.VerificationCode)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

func (r *UserRepository) RunInTx(ctx context.Context, fn func(tx *sql.Tx) error) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	if err := fn(tx); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *UserRepository) CreateTx(ctx context.Context, tx *sql.Tx, user *domain.User) error {
	if user.ID == "" {
		user.ID = generateUUID()
	}
	query := `INSERT INTO users (id, email, password_hash, first_name, last_name, role, company_name, phone, plan_id, is_active, must_change_password, is_verified, verification_code, created_at, updated_at)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())`
	_, err := tx.ExecContext(ctx, query, user.ID, user.Email, user.Password, user.FirstName, user.LastName, user.Role, user.CompanyName, user.Phone, user.PlanID, user.IsActive, user.MustChangePassword, user.IsVerified, user.VerificationCode)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `SELECT id, email, password_hash, first_name, last_name, role, company_name, phone, avatar, plan_id, is_active, must_change_password, last_login_at, is_verified, verification_code, created_at, updated_at FROM users WHERE email = ?`
	row := r.db.QueryRowContext(ctx, query, email)
	user := &domain.User{}
	err := row.Scan(&user.ID, &user.Email, &user.Password, &user.FirstName, &user.LastName, &user.Role, &user.CompanyName, &user.Phone, &user.Avatar, &user.PlanID, &user.IsActive, &user.MustChangePassword, &user.LastLoginAt, &user.IsVerified, &user.VerificationCode, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	if r.redis != nil {
		cached, err := r.redis.Get(ctx, "user:"+id)
		if err == nil && cached != "" {
			_ = cached
		}
	}
	query := `SELECT id, email, password_hash, first_name, last_name, role, company_name, phone, avatar, plan_id, is_active, must_change_password, last_login_at, is_verified, verification_code, onboarding_status, industry, created_at, updated_at FROM users WHERE id = ?`
	row := r.db.QueryRowContext(ctx, query, id)
	user := &domain.User{}
	err := row.Scan(&user.ID, &user.Email, &user.Password, &user.FirstName, &user.LastName, &user.Role, &user.CompanyName, &user.Phone, &user.Avatar, &user.PlanID, &user.IsActive, &user.MustChangePassword, &user.LastLoginAt, &user.IsVerified, &user.VerificationCode, &user.OnboardingStatus, &user.Industry, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	if r.redis != nil {
		_ = r.redis.Set(ctx, "user:"+id, "cached", 5*time.Minute)
	}
	return user, nil
}

func (r *UserRepository) UpdateLastLogin(ctx context.Context, id string) error {
	query := `UPDATE users SET last_login_at = NOW() WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *UserRepository) UpdatePassword(ctx context.Context, id, hashedPassword string) error {
	query := `UPDATE users SET password_hash = ?, must_change_password = false WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, hashedPassword, id)
	return err
}

func (r *UserRepository) UpdatePlan(ctx context.Context, userID, planID string) error {
	query := `UPDATE users SET plan_id = ? WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, planID, userID)
	return err
}

func (r *UserRepository) UpdateVerificationStatus(ctx context.Context, id string, verified bool) error {
	query := `UPDATE users SET is_verified = ?, verification_code = NULL WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, verified, id)
	return err
}

func (r *UserRepository) UpdateVerificationCode(ctx context.Context, id, code string) error {
	query := `UPDATE users SET verification_code = ? WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, code, id)
	return err
}

func (r *UserRepository) GetOwnerWhatsApp(ctx context.Context, userID string) (string, error) {
	var phone string
	err := r.db.QueryRowContext(ctx, "SELECT COALESCE(owner_whatsapp, '') FROM users WHERE id = ?", userID).Scan(&phone)
	if err != nil {
		return "", err
	}
	return phone, nil
}

func (r *UserRepository) CleanupExpiredTrials(ctx context.Context, days int) (int64, error) {
	result, err := r.db.ExecContext(ctx,
		`UPDATE users SET is_active = FALSE
		 WHERE plan_id = 'free' AND trial_expires_at IS NOT NULL
		 AND trial_expires_at < NOW() - INTERVAL ? DAY
		 AND is_active = TRUE`, days)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
