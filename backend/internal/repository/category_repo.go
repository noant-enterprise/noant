package repository

import (
	"context"
	"database/sql"

	"noant/internal/domain"
	"noant/internal/infrastructure"
)

type CategoryRepository struct {
	db    *sql.DB
	redis *infrastructure.RedisClient
}

func NewCategoryRepository(db *sql.DB, redis *infrastructure.RedisClient) *CategoryRepository {
	return &CategoryRepository{db: db, redis: redis}
}

func (r *CategoryRepository) GetByName(ctx context.Context, userID, name string) (*domain.Category, error) {
	query := "SELECT id, name, description, color, created_at FROM categories WHERE user_id = ? AND name = ? LIMIT 1"
	row := r.db.QueryRowContext(ctx, query, userID, name)
	cat := &domain.Category{}
	err := row.Scan(&cat.ID, &cat.Name, &cat.Description, &cat.Color, &cat.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return cat, nil
}

func (r *CategoryRepository) Create(ctx context.Context, cat *domain.Category) error {
	if cat.ID == "" {
		cat.ID = generateUUID()
	}
	query := `INSERT INTO categories (id, user_id, name, description, color, created_at) VALUES (?, ?, ?, ?, ?, NOW())`
	_, err := r.db.ExecContext(ctx, query, cat.ID, cat.UserID, cat.Name, cat.Description, cat.Color)
	return err
}

func (r *CategoryRepository) List(ctx context.Context, userID string) ([]domain.Category, error) {
	query := `SELECT c.id, c.name, c.description, c.color, c.created_at, COUNT(q.id) as qa_count
	FROM categories c LEFT JOIN qa_pairs q ON c.id = q.category_id AND q.user_id = ?
	WHERE c.user_id = ?
	GROUP BY c.id ORDER BY c.created_at DESC`
	rows, err := r.db.QueryContext(ctx, query, userID, userID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var categories []domain.Category
	for rows.Next() {
		var cat domain.Category
		err := rows.Scan(&cat.ID, &cat.Name, &cat.Description, &cat.Color, &cat.CreatedAt, &cat.QACount)
		if err != nil {
			continue
		}
		categories = append(categories, cat)
	}
	return categories, nil
}

func (r *CategoryRepository) Delete(ctx context.Context, id, userID string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	// 1. Delete all associated Q&As
	_, err = tx.ExecContext(ctx, `DELETE FROM qa_pairs WHERE category_id = ? AND user_id = ?`, id, userID)
	if err != nil {
		return err
	}

	// 2. Delete the category itself
	_, err = tx.ExecContext(ctx, `DELETE FROM categories WHERE id = ? AND user_id = ?`, id, userID)
	if err != nil {
		return err
	}

	return tx.Commit()
}
