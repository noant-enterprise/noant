package service

import (
    "context"
    "database/sql"
    "time"

    "noant/internal/infrastructure"
)

type RetentionService struct {
    db     *sql.DB
    logger *infrastructure.Logger
}

func NewRetentionService(db *sql.DB, logger *infrastructure.Logger) *RetentionService {
    return &RetentionService{db: db, logger: logger}
}

// CleanupOldConversations deletes conversations older than retention days
func (s *RetentionService) CleanupOldConversations(ctx context.Context, retentionDays int) (int64, error) {
    cutoff := time.Now().AddDate(0, 0, -retentionDays)
    query := "DELETE FROM conversations WHERE created_at < ? AND status = 'resolved'"
    
    result, err := s.db.ExecContext(ctx, query, cutoff)
    if err != nil {
        return 0, err
    }
    
    return result.RowsAffected()
}

// CleanupOldMessages deletes messages older than retention days for archived conversations
func (s *RetentionService) CleanupOldMessages(ctx context.Context, retentionDays int) (int64, error) {
    cutoff := time.Now().AddDate(0, 0, -retentionDays)
    query := "DELETE FROM messages WHERE created_at < ? AND conversation_id IN (SELECT id FROM conversations WHERE status = 'archived')"
    
    result, err := s.db.ExecContext(ctx, query, cutoff)
    if err != nil {
        return 0, err
    }
    
    return result.RowsAffected()
}
