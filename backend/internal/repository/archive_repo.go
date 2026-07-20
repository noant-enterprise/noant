package repository

import (
	"context"
	"database/sql"

	"noant/internal/domain"
	"noant/internal/infrastructure"
)

type ArchiveRepository struct {
	db    *sql.DB
	redis *infrastructure.RedisClient
}

func NewArchiveRepository(db *sql.DB, redis *infrastructure.RedisClient) *ArchiveRepository {
	return &ArchiveRepository{db: db, redis: redis}
}

func (r *ArchiveRepository) CreateFolder(ctx context.Context, folder *domain.ArchiveFolder) error {
	if folder.ID == "" {
		folder.ID = generateUUID()
	}
	query := `INSERT INTO archive_folders (id, user_id, name, type, color, created_at)
	VALUES (?, ?, ?, ?, ?, NOW())`
	_, err := r.db.ExecContext(ctx, query, folder.ID, folder.UserID, folder.Name, folder.Type, folder.Color)
	return err
}

func (r *ArchiveRepository) ListFolders(ctx context.Context, userID, folderType string) ([]domain.ArchiveFolder, error) {
	query := `SELECT id, user_id, name, type, color, item_count, created_at FROM archive_folders WHERE user_id = ?`
	args := []interface{}{userID}
	if folderType != "" {
		query += " AND type = ?"
		args = append(args, folderType)
	}
	query += " ORDER BY created_at DESC"
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var folders []domain.ArchiveFolder
	for rows.Next() {
		var f domain.ArchiveFolder
		err := rows.Scan(&f.ID, &f.UserID, &f.Name, &f.Type, &f.Color, &f.ItemCount, &f.CreatedAt)
		if err != nil {
			continue
		}
		folders = append(folders, f)
	}
	return folders, nil
}

func (r *ArchiveRepository) MoveChat(ctx context.Context, conversationID, userID, folderID string) error {
	query := `UPDATE conversations SET folder_id = ? WHERE id = ? AND user_id = ?`
	_, err := r.db.ExecContext(ctx, query, folderID, conversationID, userID)
	return err
}
