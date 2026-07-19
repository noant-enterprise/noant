package repository

import (
	"context"
	"database/sql"
	"fmt"

	"noant/internal/domain"
	"noant/internal/infrastructure"
)

type MessageRepository struct {
	db    *sql.DB
	redis *infrastructure.RedisClient
}

func NewMessageRepository(db *sql.DB, redis *infrastructure.RedisClient) *MessageRepository {
	return &MessageRepository{db: db, redis: redis}
}

func (r *MessageRepository) Create(ctx context.Context, msg *domain.Message) error {
	if msg.ID == "" {
		msg.ID = generateUUID()
	}
	var matchedQA interface{} = nil
	var escalationReason interface{} = nil
	var language interface{} = nil
	if msg.Metadata != nil {
		if msg.Metadata.MatchedQAID != nil {
			matchedQA = *msg.Metadata.MatchedQAID
		}
		if msg.Metadata.EscalationReason != "" {
			escalationReason = msg.Metadata.EscalationReason
		}
		if msg.Metadata.Language != "" {
			language = msg.Metadata.Language
		}
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	// Get next sequence number for this conversation (atomic within transaction)
	var maxSeq sql.NullInt64
	if err := tx.QueryRowContext(ctx, `SELECT MAX(sequence) FROM messages WHERE conversation_id = ? FOR UPDATE`, msg.ConversationID).Scan(&maxSeq); err != nil {
		return fmt.Errorf("get max sequence: %w", err)
	}
	msg.Sequence = 1
	if maxSeq.Valid {
		msg.Sequence = int(maxSeq.Int64) + 1
	}

	query := `INSERT INTO messages (id, conversation_id, sender_type, sender_id, content, is_read, confidence, matched_qa_id, escalation_reason, language, source, sequence, created_at)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW())`
	if _, err := tx.ExecContext(ctx, query, msg.ID, msg.ConversationID, msg.Role, msg.SenderID, msg.Content, msg.IsRead, msg.Confidence, matchedQA, escalationReason, language, msg.Source, msg.Sequence); err != nil {
		return fmt.Errorf("insert message: %w", err)
	}
	// Update conversation's updated_at timestamp!
	updateConvQuery := `UPDATE conversations SET updated_at = NOW() WHERE id = ?`
	if _, err := tx.ExecContext(ctx, updateConvQuery, msg.ConversationID); err != nil {
		return fmt.Errorf("update conversation: %w", err)
	}
	return tx.Commit()
}

func (r *MessageRepository) ListByConversation(ctx context.Context, conversationID string, limit int) ([]domain.Message, error) {
	query := `SELECT id, conversation_id, sender_type, sender_id, content, is_read, confidence, matched_qa_id, escalation_reason, language, source, sequence, created_at
	FROM messages WHERE conversation_id = ? ORDER BY sequence ASC LIMIT ?`
	rows, err := r.db.QueryContext(ctx, query, conversationID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var messages []domain.Message
	for rows.Next() {
		var msg domain.Message
		var senderID sql.NullString
		var confidence sql.NullFloat64
		var matchedQA sql.NullString
		var escalationReason sql.NullString
		var language sql.NullString
		var source sql.NullString
		err := rows.Scan(&msg.ID, &msg.ConversationID, &msg.Role, &senderID, &msg.Content, &msg.IsRead, &confidence, &matchedQA, &escalationReason, &language, &source, &msg.Sequence, &msg.CreatedAt)
		if err != nil {
			continue
		}
		if senderID.Valid {
			msg.SenderID = &senderID.String
		}
		msg.Metadata = &domain.MessageMetadata{}
		if confidence.Valid {
			msg.Confidence = confidence.Float64
		}
		if matchedQA.Valid {
			v := matchedQA.String
			msg.Metadata.MatchedQAID = &v
		}
		if escalationReason.Valid {
			msg.Metadata.EscalationReason = escalationReason.String
		}
		if language.Valid {
			msg.Metadata.Language = language.String
		}
		if source.Valid {
			msg.Source = source.String
		}
		messages = append(messages, msg)
	}
	return messages, nil
}

func (r *MessageRepository) ListByConversationPaginated(ctx context.Context, conversationID string, limit, offset int) ([]domain.Message, int, error) {
	var total int
	countQuery := `SELECT COUNT(*) FROM messages WHERE conversation_id = ?`
	err := r.db.QueryRowContext(ctx, countQuery, conversationID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	query := `SELECT id, conversation_id, sender_type, sender_id, content, is_read, confidence, matched_qa_id, escalation_reason, language, source, sequence, created_at
	FROM messages WHERE conversation_id = ? ORDER BY sequence DESC LIMIT ? OFFSET ?`
	rows, err := r.db.QueryContext(ctx, query, conversationID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var messages []domain.Message
	for rows.Next() {
		var msg domain.Message
		var senderID sql.NullString
		var confidence sql.NullFloat64
		var matchedQA sql.NullString
		var escalationReason sql.NullString
		var language sql.NullString
		var source sql.NullString
		err := rows.Scan(&msg.ID, &msg.ConversationID, &msg.Role, &senderID, &msg.Content, &msg.IsRead, &confidence, &matchedQA, &escalationReason, &language, &source, &msg.Sequence, &msg.CreatedAt)
		if err != nil {
			continue
		}
		if senderID.Valid {
			msg.SenderID = &senderID.String
		}
		msg.Metadata = &domain.MessageMetadata{}
		if confidence.Valid {
			msg.Confidence = confidence.Float64
		}
		if matchedQA.Valid {
			v := matchedQA.String
			msg.Metadata.MatchedQAID = &v
		}
		if escalationReason.Valid {
			msg.Metadata.EscalationReason = escalationReason.String
		}
		if language.Valid {
			msg.Metadata.Language = language.String
		}
		if source.Valid {
			msg.Source = source.String
		}
		messages = append(messages, msg)
	}
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}
	return messages, total, nil
}

func (r *MessageRepository) GetLastMessage(ctx context.Context, conversationID string) (*domain.Message, error) {
	query := `SELECT id, conversation_id, sender_type, content, is_read, created_at FROM messages WHERE conversation_id = ? ORDER BY created_at DESC LIMIT 1`
	row := r.db.QueryRowContext(ctx, query, conversationID)
	msg := &domain.Message{}
	err := row.Scan(&msg.ID, &msg.ConversationID, &msg.Role, &msg.Content, &msg.IsRead, &msg.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return msg, nil
}

func (r *MessageRepository) CountUnread(ctx context.Context, conversationID string) (int, error) {
	query := `SELECT COUNT(*) FROM messages WHERE conversation_id = ? AND is_read = FALSE AND sender_type = 'customer'`
	var count int
	err := r.db.QueryRowContext(ctx, query, conversationID).Scan(&count)
	return count, err
}

func (r *MessageRepository) MarkRead(ctx context.Context, conversationID string) error {
	query := `UPDATE messages SET is_read = TRUE WHERE conversation_id = ? AND is_read = FALSE`
	_, err := r.db.ExecContext(ctx, query, conversationID)
	return err
}

func (r *MessageRepository) CleanupOrphaned(ctx context.Context) (int64, error) {
	result, err := r.db.ExecContext(ctx,
		`DELETE FROM messages WHERE conversation_id NOT IN (SELECT id FROM conversations)`)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
