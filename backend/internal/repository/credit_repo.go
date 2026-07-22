package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"noant/internal/domain"
	"noant/internal/infrastructure"
)

type CreditRepository struct {
	db *sql.DB
}

// CampaignRepository database operations
func NewCreditRepository(db *sql.DB, redis *infrastructure.RedisClient) *CreditRepository {
	return &CreditRepository{db: db}
}

func (r *CreditRepository) GetByOrgID(ctx context.Context, orgID string) (*domain.UserCredit, error) {
	uc := &domain.UserCredit{}
	err := r.db.QueryRowContext(ctx, `SELECT id, user_id, balance, expires_at, last_updated_at FROM user_credits WHERE user_id = ?`, orgID).
		Scan(&uc.ID, &uc.UserID, &uc.Balance, &uc.ExpiresAt, &uc.LastUpdatedAt)
	if err == sql.ErrNoRows {
		return &domain.UserCredit{
			ID:           "",
			UserID:       orgID,
			Balance:      0,
			ExpiresAt:    nil,
			LastUpdatedAt: time.Now(),
		}, nil
	}
	if err != nil {
		return nil, err
	}
	return uc, nil
}

func (r *CreditRepository) Upsert(ctx context.Context, credit *domain.UserCredit) error {
	if credit.ID == "" {
		// Insert new record
		result, err := r.db.ExecContext(ctx, 
			`INSERT INTO user_credits (user_id, balance, expires_at, last_updated_at) VALUES (?, ?, ?, ?)`,
			credit.UserID, credit.Balance, credit.ExpiresAt, credit.LastUpdatedAt)
		if err != nil {
			return err
		}
		id, err := result.LastInsertId()
		if err != nil {
			return err
		}
		credit.ID = fmt.Sprintf("%d", id)
		return nil
	} else {
		// Update existing record
		_, err := r.db.ExecContext(ctx,
			`UPDATE user_credits SET balance = ?, expires_at = ?, last_updated_at = ? WHERE id = ?`,
			credit.Balance, credit.ExpiresAt, credit.LastUpdatedAt, credit.ID)
		return err
	}
}

func (r *CreditRepository) Deduct(ctx context.Context, userID string, amount int) error {
	return retryOnDeadlock(func() error {
		tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
		if err != nil {
			return fmt.Errorf("begin tx: %w", err)
		}
		defer func() { _ = tx.Rollback() }()

		// Get current credit record FOR UPDATE to prevent concurrent overspend
		var currentBalance int
		var expiresAt *time.Time
		err = tx.QueryRowContext(ctx, `SELECT balance, expires_at FROM user_credits WHERE user_id = ? FOR UPDATE`, userID).
			Scan(&currentBalance, &expiresAt)
		if err != nil {
			if err == sql.ErrNoRows {
				return fmt.Errorf("no credit balance found for user %s", userID)
			}
			return fmt.Errorf("select balance: %w", err)
		}

		// Check if expired
		if expiresAt != nil && expiresAt.Before(time.Now()) {
			return fmt.Errorf("credit balance has expired for user %s", userID)
		}

		// Check sufficient balance
		if currentBalance < amount {
			return fmt.Errorf("insufficient credit balance: have %d, need %d", currentBalance, amount)
		}

		// Deduct amount
		newBalance := currentBalance - amount
		_, err = tx.ExecContext(ctx, `UPDATE user_credits SET balance = ?, last_updated_at = NOW() WHERE user_id = ?`,
			newBalance, userID)
		if err != nil {
			return fmt.Errorf("update balance: %w", err)
		}

		return tx.Commit()
	})
}

func (r *CreditRepository) GetExpiring(ctx context.Context, days int) ([]domain.UserCredit, error) {
	var credits []domain.UserCredit
	expiryDate := time.Now().AddDate(0, 0, days)
	
	rows, err := r.db.QueryContext(ctx, 
		`SELECT id, user_id, balance, expires_at, last_updated_at FROM user_credits WHERE expires_at BETWEEN NOW() AND ?`, 
		expiryDate)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var uc domain.UserCredit
		if err := rows.Scan(&uc.ID, &uc.UserID, &uc.Balance, &uc.ExpiresAt, &uc.LastUpdatedAt); err != nil {
			continue
		}
		credits = append(credits, uc)
	}
	return credits, nil
}

func (r *CreditRepository) CreatePurchase(ctx context.Context, p *domain.CreditPurchase) error {
	if p.ID == "" {
		p.ID = generateUUID()
	}
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO credit_purchases (id, user_id, checkout_id, pack_type, amount, status, purchased_at, expires_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		p.ID, p.UserID, p.CheckoutID, p.PackType, p.Amount, p.Status, p.PurchasedAt, p.ExpiresAt)
	return err
}

func (r *CreditRepository) GetPurchaseHistory(ctx context.Context, userID string) ([]domain.CreditPurchase, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, user_id, checkout_id, pack_type, amount, status, purchased_at, expires_at
		 FROM credit_purchases WHERE user_id = ? ORDER BY purchased_at DESC LIMIT 50`, userID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var purchases []domain.CreditPurchase
	for rows.Next() {
		var p domain.CreditPurchase
		if err := rows.Scan(&p.ID, &p.UserID, &p.CheckoutID, &p.PackType, &p.Amount, &p.Status, &p.PurchasedAt, &p.ExpiresAt); err != nil {
			continue
		}
		purchases = append(purchases, p)
	}
	return purchases, nil
}

func (r *CreditRepository) CleanupExpired(ctx context.Context) (int64, error) {
	result, err := r.db.ExecContext(ctx,
		`DELETE FROM user_credits WHERE expires_at IS NOT NULL AND expires_at < NOW()`)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (r *CreditRepository) CleanupStalePurchases(ctx context.Context, days int) (int64, error) {
	result, err := r.db.ExecContext(ctx,
		`DELETE FROM credit_purchases WHERE created_at < NOW() - INTERVAL ? DAY`, days)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
