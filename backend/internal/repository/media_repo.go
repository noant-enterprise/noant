package repository

import (
	"context"
	"database/sql"

	"noant/internal/domain"
	"noant/internal/infrastructure"
)

type MediaMessageRepository struct {
	db *sql.DB
}

func NewMediaMessageRepository(db *sql.DB, redis *infrastructure.RedisClient) *MediaMessageRepository {
	return &MediaMessageRepository{db: db}
}

func (r *MediaMessageRepository) Create(ctx context.Context, m *domain.MediaMessage) error {
	if m.ID == "" {
		m.ID = generateUUID()
	}
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO media_messages (id, user_id, conversation_id, message_id, session_id, media_type, mime_type, file_size, file_name, file_path, thumb_path, width, height, duration, caption, remote_url, created_at, expires_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), ?)`,
		m.ID, m.UserID, m.ConversationID, m.MessageID, m.SessionID, m.MediaType, m.MimeType, m.FileSize, m.FileName, m.FilePath, m.ThumbPath, m.Width, m.Height, m.Duration, m.Caption, m.RemoteURL, m.ExpiresAt)
	return err
}

func (r *MediaMessageRepository) GetByConversation(ctx context.Context, conversationID string) ([]domain.MediaMessage, error) {
	var media []domain.MediaMessage
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, user_id, conversation_id, message_id, session_id, media_type, mime_type, file_size, file_name, file_path, thumb_path, width, height, duration, caption, remote_url, created_at, expires_at
		FROM media_messages WHERE conversation_id = ? ORDER BY created_at`, conversationID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		var m domain.MediaMessage
		if err := rows.Scan(&m.ID, &m.UserID, &m.ConversationID, &m.MessageID, &m.SessionID, &m.MediaType, &m.MimeType, &m.FileSize, &m.FileName, &m.FilePath, &m.ThumbPath, &m.Width, &m.Height, &m.Duration, &m.Caption, &m.RemoteURL, &m.CreatedAt, &m.ExpiresAt); err != nil {
			continue
		}
		media = append(media, m)
	}
	return media, nil
}

func (r *MediaMessageRepository) CleanupExpired(ctx context.Context) (int64, error) {
	result, err := r.db.ExecContext(ctx,
		`DELETE FROM media_messages WHERE expires_at < NOW()`)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
