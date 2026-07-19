package repository

import (
	"context"
	"database/sql"

	"noant/internal/domain"
	"noant/internal/infrastructure"
)

type InventoryRepository struct {
	db    *sql.DB
	redis *infrastructure.RedisClient
}

func NewInventoryRepository(db *sql.DB, redis *infrastructure.RedisClient) *InventoryRepository {
	return &InventoryRepository{db: db, redis: redis}
}

func (r *InventoryRepository) Create(ctx context.Context, item *domain.InventoryItem) error {
	if item.ID == "" {
		item.ID = generateUUID()
	}
	query := `INSERT INTO inventory_items (id, user_id, type, name, description, price, min_price, stock_quantity, image_url, is_active, created_at, updated_at)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())`
	_, err := r.db.ExecContext(ctx, query, item.ID, item.UserID, item.Type, item.Name, item.Description, item.Price, item.MinPrice, item.StockQuantity, item.ImageURL, item.IsActive)
	return err
}

func (r *InventoryRepository) GetByID(ctx context.Context, id string, userID string) (*domain.InventoryItem, error) {
	item := &domain.InventoryItem{}
	row := r.db.QueryRowContext(ctx, `SELECT id, user_id, type, name, description, price, min_price, stock_quantity, image_url, is_active, created_at, updated_at FROM inventory_items WHERE id = ? AND user_id = ?`, id, userID)
	err := row.Scan(&item.ID, &item.UserID, &item.Type, &item.Name, &item.Description, &item.Price, &item.MinPrice, &item.StockQuantity, &item.ImageURL, &item.IsActive, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return item, nil
}

func (r *InventoryRepository) List(ctx context.Context, userID string, itemType string, activeOnly bool) ([]domain.InventoryItem, error) {
	query := `SELECT id, user_id, type, name, description, price, min_price, stock_quantity, image_url, is_active, created_at, updated_at FROM inventory_items WHERE user_id = ?`
	args := []interface{}{userID}
	if itemType != "" {
		query += " AND type = ?"
		args = append(args, itemType)
	}
	if activeOnly {
		query += " AND is_active = TRUE"
	}
	query += " ORDER BY created_at DESC"
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []domain.InventoryItem
	for rows.Next() {
		var item domain.InventoryItem
		if err := rows.Scan(&item.ID, &item.UserID, &item.Type, &item.Name, &item.Description, &item.Price, &item.MinPrice, &item.StockQuantity, &item.ImageURL, &item.IsActive, &item.CreatedAt, &item.UpdatedAt); err != nil {
			continue
		}
		items = append(items, item)
	}
	return items, nil
}

func (r *InventoryRepository) Search(ctx context.Context, userID string, q string) ([]domain.InventoryItem, error) {
	query := `SELECT id, user_id, type, name, description, price, min_price, stock_quantity, image_url, is_active, created_at, updated_at FROM inventory_items WHERE user_id = ? AND is_active = TRUE AND (name LIKE ? OR description LIKE ?) ORDER BY name LIMIT 10`
	like := "%" + q + "%"
	rows, err := r.db.QueryContext(ctx, query, userID, like, like)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []domain.InventoryItem
	for rows.Next() {
		var item domain.InventoryItem
		if err := rows.Scan(&item.ID, &item.UserID, &item.Type, &item.Name, &item.Description, &item.Price, &item.MinPrice, &item.StockQuantity, &item.ImageURL, &item.IsActive, &item.CreatedAt, &item.UpdatedAt); err != nil {
			continue
		}
		items = append(items, item)
	}
	return items, nil
}

func (r *InventoryRepository) Update(ctx context.Context, item *domain.InventoryItem) error {
	query := `UPDATE inventory_items SET type=?, name=?, description=?, price=?, min_price=?, stock_quantity=?, image_url=?, is_active=?, updated_at=NOW() WHERE id=? AND user_id=?`
	_, err := r.db.ExecContext(ctx, query, item.Type, item.Name, item.Description, item.Price, item.MinPrice, item.StockQuantity, item.ImageURL, item.IsActive, item.ID, item.UserID)
	return err
}

func (r *InventoryRepository) Delete(ctx context.Context, id string, userID string) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM inventory_items WHERE id=? AND user_id=?", id, userID)
	return err
}

func (r *InventoryRepository) DecreaseStock(ctx context.Context, itemID string, quantity int) error {
	_, err := r.db.ExecContext(ctx, "UPDATE inventory_items SET stock_quantity = stock_quantity - ? WHERE id = ? AND stock_quantity >= ?", quantity, itemID, quantity)
	return err
}

func (r *InventoryRepository) CountByUser(ctx context.Context, userID string) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM inventory_items WHERE user_id = ?", userID).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}
