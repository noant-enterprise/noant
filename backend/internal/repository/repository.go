package repository

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"noant/internal/domain"
	"noant/internal/infrastructure"
)

type Repositories struct {
	User         *UserRepository
	Conversation *ConversationRepository
	Message      *MessageRepository
	QAPair       *QAPairRepository
	Category     *CategoryRepository
	UnknownQ     *UnknownQuestionRepository
	Integration  *IntegrationRepository
	Team         *TeamRepository
	APIKey       *APIKeyRepository
	Archive      *ArchiveRepository
	Subscription *SubscriptionRepository
	Audit        *AuditRepository
	Notification *NotificationRepository
	WidgetConfig *WidgetConfigRepository
	Inventory    *InventoryRepository
	Handoff      *HandoffRepository
}

func NewRepositories(db *sql.DB, redis *infrastructure.RedisClient) *Repositories {
	return &Repositories{
		User:         NewUserRepository(db, redis),
		Conversation: NewConversationRepository(db, redis),
		Message:      NewMessageRepository(db, redis),
		QAPair:       NewQAPairRepository(db, redis),
		Category:     NewCategoryRepository(db, redis),
		UnknownQ:     NewUnknownQuestionRepository(db, redis),
		Integration:  NewIntegrationRepository(db, redis),
		Team:         NewTeamRepository(db, redis),
		APIKey:       NewAPIKeyRepository(db, redis),
		Archive:      NewArchiveRepository(db, redis),
		Subscription: NewSubscriptionRepository(db, redis),
		Audit:        NewAuditRepository(db, redis),
		Notification: NewNotificationRepository(db, redis),
		WidgetConfig: NewWidgetConfigRepository(db, redis),
		Inventory:    NewInventoryRepository(db, redis),
		Handoff:      NewHandoffRepository(db, redis),
	}
}

// ConversationRepository additional methods
func (r *ConversationRepository) GetByID(ctx context.Context, id string) (*domain.Conversation, error) {
	conv := &domain.Conversation{}
	row := r.db.QueryRowContext(ctx, `SELECT id, user_id, customer_name, customer_phone, customer_email, channel, status, intent, priority, is_ai_transferred, taken_over_by, taken_over_at, resolved_at, folder_id, created_at, updated_at FROM conversations WHERE id = ?`, id)
	err := row.Scan(&conv.ID, &conv.UserID, &conv.CustomerName, &conv.CustomerPhone, &conv.CustomerEmail, &conv.Channel, &conv.Status, &conv.Intent, &conv.Priority, &conv.IsAITransferred, &conv.TakenOverBy, &conv.TakenOverAt, &conv.ResolvedAt, &conv.FolderID, &conv.CreatedAt, &conv.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return conv, nil
}
func generateUUID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40 // Version 4
	b[8] = (b[8] & 0x3f) | 0x80 // Variant 10
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

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
	query := `INSERT INTO users (id, email, password_hash, first_name, last_name, role, company_name, phone, plan_id, is_active, must_change_password, created_at, updated_at)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())`
	_, err := r.db.ExecContext(ctx, query, user.ID, user.Email, user.Password, user.FirstName, user.LastName, user.Role, user.CompanyName, user.Phone, user.PlanID, user.IsActive, user.MustChangePassword)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `SELECT id, email, password_hash, first_name, last_name, role, company_name, phone, avatar, plan_id, is_active, must_change_password, last_login_at, created_at, updated_at FROM users WHERE email = ?`
	row := r.db.QueryRowContext(ctx, query, email)
	user := &domain.User{}
	err := row.Scan(&user.ID, &user.Email, &user.Password, &user.FirstName, &user.LastName, &user.Role, &user.CompanyName, &user.Phone, &user.Avatar, &user.PlanID, &user.IsActive, &user.MustChangePassword, &user.LastLoginAt, &user.CreatedAt, &user.UpdatedAt)
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
	query := `SELECT id, email, password_hash, first_name, last_name, role, company_name, phone, avatar, plan_id, is_active, must_change_password, last_login_at, created_at, updated_at FROM users WHERE id = ?`
	row := r.db.QueryRowContext(ctx, query, id)
	user := &domain.User{}
	err := row.Scan(&user.ID, &user.Email, &user.Password, &user.FirstName, &user.LastName, &user.Role, &user.CompanyName, &user.Phone, &user.Avatar, &user.PlanID, &user.IsActive, &user.MustChangePassword, &user.LastLoginAt, &user.CreatedAt, &user.UpdatedAt)
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

func (r *UserRepository) UpdatePassword(ctx context.Context, id string, hashedPassword string) error {
	query := `UPDATE users SET password_hash = ?, must_change_password = false WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, hashedPassword, id)
	return err
}

func (r *UserRepository) UpdatePlan(ctx context.Context, userID string, planID string) error {
	query := `UPDATE users SET plan_id = ? WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, planID, userID)
	return err
}

type ConversationRepository struct {
	db    *sql.DB
	redis *infrastructure.RedisClient
}

func NewConversationRepository(db *sql.DB, redis *infrastructure.RedisClient) *ConversationRepository {
	return &ConversationRepository{db: db, redis: redis}
}

func (r *ConversationRepository) Create(ctx context.Context, conv *domain.Conversation) error {
	if conv.ID == "" {
		conv.ID = generateUUID()
	}
	query := `INSERT INTO conversations (id, user_id, customer_name, customer_phone, customer_email, channel, status, intent, priority, is_ai_transferred, created_at, updated_at)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())`
	_, err := r.db.ExecContext(ctx, query, conv.ID, conv.UserID, conv.CustomerName, conv.CustomerPhone, conv.CustomerEmail, conv.Channel, conv.Status, conv.Intent, conv.Priority, conv.IsAITransferred)
	if err != nil {
		return fmt.Errorf("failed to create conversation: %w", err)
	}
	return nil
}

func (r *ConversationRepository) List(ctx context.Context, userID string, status string, limit, offset int) ([]domain.Conversation, int, error) {
	countQuery := "SELECT COUNT(*) FROM conversations WHERE user_id = ?"
	countArgs := []interface{}{userID}
	if status != "" {
		countQuery += " AND status = ?"
		countArgs = append(countArgs, status)
	}
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, countArgs...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}
	query := `SELECT id, user_id, customer_name, customer_phone, customer_email, channel, status, intent, priority, is_ai_transferred, taken_over_by, taken_over_at, resolved_at, folder_id, created_at, updated_at
	FROM conversations WHERE user_id = ?`
	args := []interface{}{userID}
	if status != "" {
		query += " AND status = ?"
		args = append(args, status)
	}
	query += " ORDER BY updated_at DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var conversations []domain.Conversation
	for rows.Next() {
		var conv domain.Conversation
		err := rows.Scan(&conv.ID, &conv.UserID, &conv.CustomerName, &conv.CustomerPhone, &conv.CustomerEmail, &conv.Channel, &conv.Status, &conv.Intent, &conv.Priority, &conv.IsAITransferred, &conv.TakenOverBy, &conv.TakenOverAt, &conv.ResolvedAt, &conv.FolderID, &conv.CreatedAt, &conv.UpdatedAt)
		if err != nil {
			continue
		}
		conversations = append(conversations, conv)
	}
	return conversations, total, nil
}

func (r *ConversationRepository) GetByIDAndUser(ctx context.Context, id string, userID string) (*domain.Conversation, error) {
	query := `SELECT id, user_id, customer_name, customer_phone, customer_email, channel, status, intent, priority, is_ai_transferred, taken_over_by, taken_over_at, resolved_at, folder_id, created_at, updated_at FROM conversations WHERE id = ? AND user_id = ?`
	row := r.db.QueryRowContext(ctx, query, id, userID)
	conv := &domain.Conversation{}
	err := row.Scan(&conv.ID, &conv.UserID, &conv.CustomerName, &conv.CustomerPhone, &conv.CustomerEmail, &conv.Channel, &conv.Status, &conv.Intent, &conv.Priority, &conv.IsAITransferred, &conv.TakenOverBy, &conv.TakenOverAt, &conv.ResolvedAt, &conv.FolderID, &conv.CreatedAt, &conv.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return conv, nil
}

func (r *ConversationRepository) UpdateStatus(ctx context.Context, id string, userID string, status string) error {
	query := `UPDATE conversations SET status = ? WHERE id = ? AND user_id = ?`
	_, err := r.db.ExecContext(ctx, query, status, id, userID)
	return err
}

func (r *ConversationRepository) FindActiveByCustomer(ctx context.Context, userID, customerName, channel string) (*domain.Conversation, error) {
	query := `SELECT id, user_id, customer_name, customer_phone, customer_email, channel, status, intent, priority, is_ai_transferred, taken_over_by, taken_over_at, resolved_at, folder_id, created_at, updated_at FROM conversations WHERE user_id = ? AND (customer_phone = ? OR customer_name = ?) AND channel = ? AND status = 'active' ORDER BY created_at DESC LIMIT 1`
	row := r.db.QueryRowContext(ctx, query, userID, customerName, customerName, channel)
	conv := &domain.Conversation{}
	err := row.Scan(&conv.ID, &conv.UserID, &conv.CustomerName, &conv.CustomerPhone, &conv.CustomerEmail, &conv.Channel, &conv.Status, &conv.Intent, &conv.Priority, &conv.IsAITransferred, &conv.TakenOverBy, &conv.TakenOverAt, &conv.ResolvedAt, &conv.FolderID, &conv.CreatedAt, &conv.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return conv, nil
}

func (r *ConversationRepository) Takeover(ctx context.Context, id string, userID string, agentID string) error {
	query := `UPDATE conversations SET status = 'escalated', taken_over_by = ?, taken_over_at = NOW() WHERE id = ? AND user_id = ?`
	_, err := r.db.ExecContext(ctx, query, agentID, id, userID)
	return err
}

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
	query := `INSERT INTO messages (id, conversation_id, sender_type, sender_id, content, is_read, confidence, matched_qa_id, escalation_reason, language, source, created_at)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW())`
	_, err := r.db.ExecContext(ctx, query, msg.ID, msg.ConversationID, msg.Role, msg.SenderID, msg.Content, msg.IsRead, msg.Confidence, matchedQA, escalationReason, language, msg.Source)
	if err != nil {
		return err
	}
	// Update conversation's updated_at timestamp!
	updateConvQuery := `UPDATE conversations SET updated_at = NOW() WHERE id = ?`
	_, _ = r.db.ExecContext(ctx, updateConvQuery, msg.ConversationID)
	return nil
}

func (r *MessageRepository) ListByConversation(ctx context.Context, conversationID string, limit int) ([]domain.Message, error) {
	query := `SELECT id, conversation_id, sender_type, sender_id, content, is_read, confidence, matched_qa_id, escalation_reason, language, source, created_at
	FROM messages WHERE conversation_id = ? ORDER BY created_at DESC LIMIT ?`
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
		err := rows.Scan(&msg.ID, &msg.ConversationID, &msg.Role, &senderID, &msg.Content, &msg.IsRead, &confidence, &matchedQA, &escalationReason, &language, &source, &msg.CreatedAt)
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
	return messages, nil
}

func (r *MessageRepository) ListByConversationPaginated(ctx context.Context, conversationID string, limit, offset int) ([]domain.Message, int, error) {
	var total int
	countQuery := `SELECT COUNT(*) FROM messages WHERE conversation_id = ?`
	err := r.db.QueryRowContext(ctx, countQuery, conversationID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	query := `SELECT id, conversation_id, sender_type, sender_id, content, is_read, confidence, matched_qa_id, escalation_reason, language, source, created_at
	FROM messages WHERE conversation_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?`
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
		err := rows.Scan(&msg.ID, &msg.ConversationID, &msg.Role, &senderID, &msg.Content, &msg.IsRead, &confidence, &matchedQA, &escalationReason, &language, &source, &msg.CreatedAt)
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

type QAPairRepository struct {
	db    *sql.DB
	redis *infrastructure.RedisClient
}

func NewQAPairRepository(db *sql.DB, redis *infrastructure.RedisClient) *QAPairRepository {
	return &QAPairRepository{db: db, redis: redis}
}

func (r *QAPairRepository) Create(ctx context.Context, qa *domain.QAPair) error {
	if qa.ID == "" {
		qa.ID = generateUUID()
	}
	query := `INSERT INTO qa_pairs (id, user_id, category_id, question, answer, variations, is_active, created_at, updated_at)
	VALUES (?, ?, ?, ?, ?, ?, ?, NOW(), NOW())`
	variationsJSON := []byte("[]")
	if len(qa.Variations) > 0 {
		b, err := json.Marshal(qa.Variations)
		if err == nil {
			variationsJSON = b
		}
	}
	_, err := r.db.ExecContext(ctx, query, qa.ID, qa.UserID, qa.CategoryID, qa.Question, qa.Answer, string(variationsJSON), qa.IsActive)
	return err
}

func (r *QAPairRepository) BulkCreate(ctx context.Context, qas []domain.QAPair) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	query := `INSERT INTO qa_pairs (id, user_id, category_id, question, answer, variations, is_active, created_at, updated_at)
	VALUES (?, ?, ?, ?, ?, ?, ?, NOW(), NOW())`
	for _, qa := range qas {
		if qa.ID == "" {
			qa.ID = generateUUID()
		}
		variationsJSON := []byte("[]")
		if len(qa.Variations) > 0 {
			b, err := json.Marshal(qa.Variations)
			if err == nil {
				variationsJSON = b
			}
		}
		_, err := tx.ExecContext(ctx, query, qa.ID, qa.UserID, qa.CategoryID, qa.Question, qa.Answer, string(variationsJSON), qa.IsActive)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *QAPairRepository) ListByCategory(ctx context.Context, categoryID string) ([]domain.QAPair, error) {
	query := `SELECT id, category_id, question, answer, variations, is_active, usage_count, created_at, updated_at FROM qa_pairs WHERE category_id = ? AND is_active = true`
	rows, err := r.db.QueryContext(ctx, query, categoryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var qas []domain.QAPair
	for rows.Next() {
		var qa domain.QAPair
		var variationsJSON sql.NullString
		err := rows.Scan(&qa.ID, &qa.CategoryID, &qa.Question, &qa.Answer, &variationsJSON, &qa.IsActive, &qa.UsageCount, &qa.CreatedAt, &qa.UpdatedAt)
		if err != nil {
			continue
		}
		if variationsJSON.Valid && variationsJSON.String != "" && variationsJSON.String != "[]" {
			_ = json.Unmarshal([]byte(variationsJSON.String), &qa.Variations)
		} else {
			qa.Variations = []string{}
		}
		qas = append(qas, qa)
	}
	return qas, nil
}

func (r *QAPairRepository) ListByCategoryAndUser(ctx context.Context, categoryID string, userID string) ([]domain.QAPair, error) {
	query := `SELECT id, category_id, question, answer, variations, is_active, usage_count, created_at, updated_at FROM qa_pairs WHERE category_id = ? AND user_id = ? AND is_active = true`
	rows, err := r.db.QueryContext(ctx, query, categoryID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var qas []domain.QAPair
	for rows.Next() {
		var qa domain.QAPair
		var variationsJSON sql.NullString
		err := rows.Scan(&qa.ID, &qa.CategoryID, &qa.Question, &qa.Answer, &variationsJSON, &qa.IsActive, &qa.UsageCount, &qa.CreatedAt, &qa.UpdatedAt)
		if err != nil {
			continue
		}
		if variationsJSON.Valid && variationsJSON.String != "" && variationsJSON.String != "[]" {
			_ = json.Unmarshal([]byte(variationsJSON.String), &qa.Variations)
		} else {
			qa.Variations = []string{}
		}
		qa.UserID = userID
		qas = append(qas, qa)
	}
	return qas, nil
}

func (r *QAPairRepository) Search(ctx context.Context, userID string, query string) ([]domain.QAPair, error) {
	// Clean query by removing common punctuation and trim
	cleanQuery := query
	for _, char := range []string{"?", "!", ".", ",", ";", ":"} {
		cleanQuery = strings.ReplaceAll(cleanQuery, char, "")
	}
	cleanQuery = strings.TrimSpace(cleanQuery)

	sqlQuery := `SELECT id, category_id, question, answer, variations, is_active, usage_count, created_at, updated_at
	FROM qa_pairs 
	WHERE user_id = ? 
	  AND is_active = true 
	  AND (
	      LOWER(question) LIKE LOWER(?) 
	      OR LOWER(?) LIKE CONCAT('%', LOWER(REPLACE(REPLACE(question, '?', ''), '!', '')), '%')
	      OR LOWER(REPLACE(REPLACE(question, '?', ''), '!', '')) LIKE LOWER(?)
	      OR LOWER(answer) LIKE LOWER(?)
	  ) LIMIT 10`

	searchTerm := "%" + cleanQuery + "%"
	rows, err := r.db.QueryContext(ctx, sqlQuery, userID, searchTerm, cleanQuery, searchTerm, searchTerm)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var qas []domain.QAPair
	for rows.Next() {
		var qa domain.QAPair
		var variationsJSON sql.NullString
		err := rows.Scan(&qa.ID, &qa.CategoryID, &qa.Question, &qa.Answer, &variationsJSON, &qa.IsActive, &qa.UsageCount, &qa.CreatedAt, &qa.UpdatedAt)
		if err != nil {
			continue
		}
		if variationsJSON.Valid && variationsJSON.String != "" && variationsJSON.String != "[]" {
			_ = json.Unmarshal([]byte(variationsJSON.String), &qa.Variations)
		} else {
			qa.Variations = []string{}
		}
		qas = append(qas, qa)
	}
	return qas, nil
}

func (r *QAPairRepository) ListByUser(ctx context.Context, userID string, categoryID string) ([]domain.QAPair, error) {
	query := `SELECT id, category_id, question, answer, variations, is_active, usage_count, created_at, updated_at FROM qa_pairs WHERE user_id = ? AND is_active = true`
	args := []interface{}{userID}
	if categoryID != "" {
		query += " AND category_id = ?"
		args = append(args, categoryID)
	}
	query += " ORDER BY created_at DESC"
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var qas []domain.QAPair
	for rows.Next() {
		var qa domain.QAPair
		var variationsJSON sql.NullString
		if err := rows.Scan(&qa.ID, &qa.CategoryID, &qa.Question, &qa.Answer, &variationsJSON, &qa.IsActive, &qa.UsageCount, &qa.CreatedAt, &qa.UpdatedAt); err != nil {
			continue
		}
		if variationsJSON.Valid && variationsJSON.String != "" && variationsJSON.String != "[]" {
			_ = json.Unmarshal([]byte(variationsJSON.String), &qa.Variations)
		} else {
			qa.Variations = []string{}
		}
		qas = append(qas, qa)
	}
	return qas, nil
}

func (r *QAPairRepository) GetByID(ctx context.Context, id string) (*domain.QAPair, error) {
	query := `SELECT id, category_id, question, answer, variations, is_active, usage_count, created_at, updated_at FROM qa_pairs WHERE id = ? LIMIT 1`
	var qa domain.QAPair
	var variationsJSON sql.NullString
	err := r.db.QueryRowContext(ctx, query, id).Scan(&qa.ID, &qa.CategoryID, &qa.Question, &qa.Answer, &variationsJSON, &qa.IsActive, &qa.UsageCount, &qa.CreatedAt, &qa.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	if variationsJSON.Valid && variationsJSON.String != "" && variationsJSON.String != "[]" {
		_ = json.Unmarshal([]byte(variationsJSON.String), &qa.Variations)
	} else {
		qa.Variations = []string{}
	}
	return &qa, nil
}

func (r *QAPairRepository) GetByQuestion(ctx context.Context, userID, question string) (*domain.QAPair, error) {
	query := `SELECT id, category_id, question, answer, variations, is_active, usage_count, created_at, updated_at 
	FROM qa_pairs WHERE user_id = ? AND question = ? LIMIT 1`
	var qa domain.QAPair
	var variationsJSON sql.NullString
	err := r.db.QueryRowContext(ctx, query, userID, question).Scan(
		&qa.ID, &qa.CategoryID, &qa.Question, &qa.Answer, &variationsJSON, 
		&qa.IsActive, &qa.UsageCount, &qa.CreatedAt, &qa.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if variationsJSON.Valid && variationsJSON.String != "" && variationsJSON.String != "[]" {
		_ = json.Unmarshal([]byte(variationsJSON.String), &qa.Variations)
	} else {
		qa.Variations = []string{}
	}
	qa.UserID = userID
	return &qa, nil
}

func (r *QAPairRepository) Update(ctx context.Context, qa *domain.QAPair) error {
	query := `UPDATE qa_pairs SET category_id = ?, question = ?, answer = ?, variations = ?, is_active = ?, updated_at = NOW() WHERE id = ? AND user_id = ?`
	variationsJSON := []byte("[]")
	if len(qa.Variations) > 0 {
		b, err := json.Marshal(qa.Variations)
		if err == nil {
			variationsJSON = b
		}
	}
	_, err := r.db.ExecContext(ctx, query, qa.CategoryID, qa.Question, qa.Answer, string(variationsJSON), qa.IsActive, qa.ID, qa.UserID)
	return err
}

func (r *QAPairRepository) IncrementUsage(ctx context.Context, id string) error {
	query := `UPDATE qa_pairs SET usage_count = usage_count + 1 WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *QAPairRepository) CountByUser(ctx context.Context, userID string) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM qa_pairs WHERE user_id = ? AND is_active = true`, userID).Scan(&count)
	return count, err
}

func (r *QAPairRepository) Delete(ctx context.Context, id string, userID string) error {
	query := `DELETE FROM qa_pairs WHERE id = ? AND user_id = ?`
	_, err := r.db.ExecContext(ctx, query, id, userID)
	return err
}

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
	defer rows.Close()
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

func (r *CategoryRepository) Delete(ctx context.Context, id string, userID string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

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

type UnknownQuestionRepository struct {
	db    *sql.DB
	redis *infrastructure.RedisClient
}

func NewUnknownQuestionRepository(db *sql.DB, redis *infrastructure.RedisClient) *UnknownQuestionRepository {
	return &UnknownQuestionRepository{db: db, redis: redis}
}

func (r *UnknownQuestionRepository) Create(ctx context.Context, uq *domain.UnknownQuestion) error {
	if uq.ID == "" {
		uq.ID = generateUUID()
	}
	query := `INSERT INTO unknown_questions (id, user_id, question, conversation_id, channel, status, created_at)
	VALUES (?, ?, ?, ?, ?, ?, NOW())`
	_, err := r.db.ExecContext(ctx, query, uq.ID, uq.UserID, uq.Question, uq.ConversationID, uq.Channel, uq.Status)
	return err
}

func (r *UnknownQuestionRepository) GetByIDAndUser(ctx context.Context, id string, userID string) (*domain.UnknownQuestion, error) {
	query := `SELECT id, user_id, question, conversation_id, channel, status, suggested_answer, category_id, created_at FROM unknown_questions WHERE id = ? AND user_id = ?`
	row := r.db.QueryRowContext(ctx, query, id, userID)
	uq := &domain.UnknownQuestion{}
	err := row.Scan(&uq.ID, &uq.UserID, &uq.Question, &uq.ConversationID, &uq.Channel, &uq.Status, &uq.SuggestedAnswer, &uq.CategoryID, &uq.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return uq, nil
}

func (r *UnknownQuestionRepository) List(ctx context.Context, userID string, status string, limit int) ([]domain.UnknownQuestion, error) {
	query := `SELECT id, question, conversation_id, channel, status, suggested_answer, category_id, created_at FROM unknown_questions WHERE user_id = ?`
	args := []interface{}{userID}
	if status != "" {
		query += " AND status = ?"
		args = append(args, status)
	}
	query += " ORDER BY created_at DESC LIMIT ?"
	args = append(args, limit)
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var questions []domain.UnknownQuestion
	for rows.Next() {
		var uq domain.UnknownQuestion
		err := rows.Scan(&uq.ID, &uq.Question, &uq.ConversationID, &uq.Channel, &uq.Status, &uq.SuggestedAnswer, &uq.CategoryID, &uq.CreatedAt)
		if err != nil {
			continue
		}
		questions = append(questions, uq)
	}
	return questions, nil
}

func (r *UnknownQuestionRepository) UpdateStatus(ctx context.Context, id string, userID string, status string, answer *string, categoryID *string) error {
	query := `UPDATE unknown_questions SET status = ?, suggested_answer = ?, category_id = ? WHERE id = ? AND user_id = ?`
	_, err := r.db.ExecContext(ctx, query, status, answer, categoryID, id, userID)
	return err
}

type IntegrationRepository struct {
	db    *sql.DB
	redis *infrastructure.RedisClient
}

func NewIntegrationRepository(db *sql.DB, redis *infrastructure.RedisClient) *IntegrationRepository {
	return &IntegrationRepository{db: db, redis: redis}
}

func (r *IntegrationRepository) Create(ctx context.Context, integration *domain.Integration) error {
	if integration.ID == "" {
		integration.ID = generateUUID()
	}
	query := `INSERT INTO integrations (id, user_id, channel, status, config, webhook_url, created_at, updated_at)
	VALUES (?, ?, ?, ?, ?, ?, NOW(), NOW())`
	configJSON := []byte("{}")
	if integration.Config != nil {
		b, err := json.Marshal(integration.Config)
		if err == nil {
			configJSON = b
		}
	}
	_, err := r.db.ExecContext(ctx, query, integration.ID, integration.UserID, integration.Channel, integration.Status, string(configJSON), integration.WebhookURL)
	return err
}

func (r *IntegrationRepository) ListByUser(ctx context.Context, userID string) ([]domain.Integration, error) {
	query := `SELECT id, user_id, channel, status, config, webhook_url, last_error, created_at, updated_at FROM integrations WHERE user_id = ?`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var integrations []domain.Integration
	for rows.Next() {
		var i domain.Integration
		var configStr string
		err := rows.Scan(&i.ID, &i.UserID, &i.Channel, &i.Status, &configStr, &i.WebhookURL, &i.LastError, &i.CreatedAt, &i.UpdatedAt)
		if err != nil {
			continue
		}
		if configStr != "" && configStr != "{}" {
			_ = json.Unmarshal([]byte(configStr), &i.Config)
		} else {
			i.Config = map[string]interface{}{}
		}
		integrations = append(integrations, i)
	}
	return integrations, nil
}

func (r *IntegrationRepository) ListActive(ctx context.Context) ([]domain.Integration, error) {
	query := `SELECT id, user_id, channel, status, config, webhook_url, last_error, created_at, updated_at FROM integrations WHERE status = 'active'`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var integrations []domain.Integration
	for rows.Next() {
		var i domain.Integration
		var configStr string
		err := rows.Scan(&i.ID, &i.UserID, &i.Channel, &i.Status, &configStr, &i.WebhookURL, &i.LastError, &i.CreatedAt, &i.UpdatedAt)
		if err != nil {
			continue
		}
		if configStr != "" && configStr != "{}" {
			_ = json.Unmarshal([]byte(configStr), &i.Config)
		} else {
			i.Config = map[string]interface{}{}
		}
		integrations = append(integrations, i)
	}
	return integrations, nil
}

func (r *IntegrationRepository) UpdateStatus(ctx context.Context, id string, status string, lastError *string) error {
	query := `UPDATE integrations SET status = ?, last_error = ? WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, status, lastError, id)
	return err
}

func (r *IntegrationRepository) GetByUserAndChannel(ctx context.Context, userID, channel string) (*domain.Integration, error) {
	query := `SELECT id, user_id, channel, status, config, webhook_url, last_error, created_at, updated_at 
	FROM integrations WHERE user_id = ? AND channel = ? LIMIT 1`
	var i domain.Integration
	var configStr string
	err := r.db.QueryRowContext(ctx, query, userID, channel).Scan(
		&i.ID, &i.UserID, &i.Channel, &i.Status, &configStr, &i.WebhookURL, &i.LastError, &i.CreatedAt, &i.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	if configStr != "" && configStr != "{}" {
		_ = json.Unmarshal([]byte(configStr), &i.Config)
	} else {
		i.Config = map[string]interface{}{}
	}
	return &i, nil
}

func (r *IntegrationRepository) GetByChannelAndSessionID(ctx context.Context, channel, sessionID string) (*domain.Integration, error) {
	query := `SELECT id, user_id, channel, status, config, webhook_url, last_error, created_at, updated_at
	FROM integrations WHERE channel = ?`
	rows, err := r.db.QueryContext(ctx, query, channel)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var i domain.Integration
		var configStr string
		if err := rows.Scan(&i.ID, &i.UserID, &i.Channel, &i.Status, &configStr, &i.WebhookURL, &i.LastError, &i.CreatedAt, &i.UpdatedAt); err != nil {
			continue
		}
		if configStr != "" && configStr != "{}" {
			_ = json.Unmarshal([]byte(configStr), &i.Config)
		} else {
			i.Config = map[string]interface{}{}
		}
		if cfgSessionID, ok := i.Config["session_id"].(string); ok && cfgSessionID == sessionID {
			return &i, nil
		}
	}

	return nil, nil
}

func (r *IntegrationRepository) GetByChannelAndWebhookSecret(ctx context.Context, channel, secret string) (*domain.Integration, error) {
	query := `SELECT id, user_id, channel, status, config, webhook_url, last_error, created_at, updated_at
	FROM integrations WHERE channel = ? AND status IN ('active', 'connected')`
	rows, err := r.db.QueryContext(ctx, query, channel)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var i domain.Integration
		var configStr string
		if err := rows.Scan(&i.ID, &i.UserID, &i.Channel, &i.Status, &configStr, &i.WebhookURL, &i.LastError, &i.CreatedAt, &i.UpdatedAt); err != nil {
			continue
		}
		if configStr != "" && configStr != "{}" {
			_ = json.Unmarshal([]byte(configStr), &i.Config)
		} else {
			i.Config = map[string]interface{}{}
		}
		if cfgSecret, ok := i.Config["webhook_secret"].(string); ok && cfgSecret == secret {
			return &i, nil
		}
	}

	return nil, nil
}

func (r *IntegrationRepository) Update(ctx context.Context, integration *domain.Integration) error {
	query := `UPDATE integrations SET status = ?, config = ?, webhook_url = ?, updated_at = NOW() WHERE id = ?`
	configJSON := []byte("{}")
	if integration.Config != nil {
		b, err := json.Marshal(integration.Config)
		if err == nil {
			configJSON = b
		}
	}
	_, err := r.db.ExecContext(ctx, query, integration.Status, string(configJSON), integration.WebhookURL, integration.ID)
	return err
}

func (r *IntegrationRepository) Disconnect(ctx context.Context, userID, channel string) error {
	query := `UPDATE integrations SET status = 'inactive', updated_at = NOW() WHERE user_id = ? AND channel = ?`
	_, err := r.db.ExecContext(ctx, query, userID, channel)
	return err
}

type TeamRepository struct {
	db    *sql.DB
	redis *infrastructure.RedisClient
}

func NewTeamRepository(db *sql.DB, redis *infrastructure.RedisClient) *TeamRepository {
	return &TeamRepository{db: db, redis: redis}
}

func (r *TeamRepository) ListByUser(ctx context.Context, ownerID string) ([]domain.TeamMember, error) {
	query := `SELECT t.id, t.user_id, u.email, u.first_name, u.last_name, t.role, t.is_active, t.joined_at
	FROM team_members t JOIN users u ON t.user_id = u.id WHERE t.owner_id = ?`
	rows, err := r.db.QueryContext(ctx, query, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var members []domain.TeamMember
	for rows.Next() {
		var m domain.TeamMember
		err := rows.Scan(&m.ID, &m.UserID, &m.Email, &m.FirstName, &m.LastName, &m.Role, &m.IsActive, &m.JoinedAt)
		if err != nil {
			continue
		}
		members = append(members, m)
	}
	return members, nil
}

func (r *TeamRepository) Create(ctx context.Context, ownerID string, member *domain.TeamMember) error {
	if member.ID == "" {
		member.ID = generateUUID()
	}
	query := `INSERT INTO team_members (id, owner_id, user_id, role, is_active, joined_at)
	VALUES (?, ?, ?, ?, ?, NOW())`
	_, err := r.db.ExecContext(ctx, query, member.ID, ownerID, member.UserID, member.Role, member.IsActive)
	return err
}

type APIKeyRepository struct {
	db    *sql.DB
	redis *infrastructure.RedisClient
}

func NewAPIKeyRepository(db *sql.DB, redis *infrastructure.RedisClient) *APIKeyRepository {
	return &APIKeyRepository{db: db, redis: redis}
}

func (r *APIKeyRepository) Create(ctx context.Context, key *domain.APIKey) error {
	if key.ID == "" {
		key.ID = generateUUID()
	}
	query := `INSERT INTO api_keys (id, user_id, name, key_hash, is_active, created_at)
	VALUES (?, ?, ?, ?, ?, NOW())`
	_, err := r.db.ExecContext(ctx, query, key.ID, key.UserID, key.Name, key.Key, key.IsActive)
	return err
}

func (r *APIKeyRepository) ListByUser(ctx context.Context, userID string) ([]domain.APIKey, error) {
	query := `SELECT id, user_id, name, key_hash, last_used, is_active, created_at FROM api_keys WHERE user_id = ? AND is_active = true`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var keys []domain.APIKey
	for rows.Next() {
		var k domain.APIKey
		err := rows.Scan(&k.ID, &k.UserID, &k.Name, &k.Key, &k.LastUsed, &k.IsActive, &k.CreatedAt)
		if err != nil {
			continue
		}
		keys = append(keys, k)
	}
	return keys, nil
}

func (r *APIKeyRepository) Revoke(ctx context.Context, id string, userID string) error {
	query := `UPDATE api_keys SET is_active = false WHERE id = ? AND user_id = ?`
	_, err := r.db.ExecContext(ctx, query, id, userID)
	return err
}

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

func (r *ArchiveRepository) ListFolders(ctx context.Context, userID string, folderType string) ([]domain.ArchiveFolder, error) {
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
	defer rows.Close()
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

func (r *ArchiveRepository) MoveChat(ctx context.Context, conversationID string, userID string, folderID string) error {
	query := `UPDATE conversations SET folder_id = ? WHERE id = ? AND user_id = ?`
	_, err := r.db.ExecContext(ctx, query, folderID, conversationID, userID)
	return err
}

type SubscriptionRepository struct {
	db    *sql.DB
	redis *infrastructure.RedisClient
}

func NewSubscriptionRepository(db *sql.DB, redis *infrastructure.RedisClient) *SubscriptionRepository {
	return &SubscriptionRepository{db: db, redis: redis}
}

func (r *SubscriptionRepository) GetActive(ctx context.Context, userID string) (*domain.Subscription, error) {
	query := `SELECT id, user_id, plan_id, status, current_period_start, current_period_end, created_at, updated_at
	FROM subscriptions WHERE user_id = ? AND status = 'active' ORDER BY created_at DESC LIMIT 1`
	row := r.db.QueryRowContext(ctx, query, userID)
	sub := &domain.Subscription{}
	err := row.Scan(&sub.ID, &sub.UserID, &sub.PlanID, &sub.Status, &sub.CurrentPeriodStart, &sub.CurrentPeriodEnd, &sub.CreatedAt, &sub.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return sub, nil
}

func (r *SubscriptionRepository) Create(ctx context.Context, sub *domain.Subscription) error {
	if sub.ID == "" {
		sub.ID = generateUUID()
	}
	query := `INSERT INTO subscriptions (id, user_id, plan_id, status, current_period_start, current_period_end, created_at, updated_at)
	VALUES (?, ?, ?, ?, ?, ?, NOW(), NOW())`
	_, err := r.db.ExecContext(ctx, query, sub.ID, sub.UserID, sub.PlanID, sub.Status, sub.CurrentPeriodStart, sub.CurrentPeriodEnd)
	return err
}

func (r *SubscriptionRepository) CreateOrUpdate(ctx context.Context, sub *domain.Subscription) error {
	existing, err := r.GetActive(ctx, sub.UserID)
	if err != nil {
		return err
	}
	if existing != nil {
		query := `UPDATE subscriptions SET plan_id = ?, status = ?, current_period_start = ?, current_period_end = ?, updated_at = NOW() WHERE id = ?`
		_, err := r.db.ExecContext(ctx, query, sub.PlanID, sub.Status, sub.CurrentPeriodStart, sub.CurrentPeriodEnd, existing.ID)
		return err
	}
	return r.Create(ctx, sub)
}

func (r *SubscriptionRepository) Cancel(ctx context.Context, userID string) error {
	query := `UPDATE subscriptions SET status = 'cancelled', updated_at = NOW() WHERE user_id = ? AND status = 'active'`
	_, err := r.db.ExecContext(ctx, query, userID)
	return err
}

// ========== ANALYTICS REPOSITORY METHODS ==========

func (r *ConversationRepository) GetOverview(ctx context.Context, userID string) (map[string]interface{}, error) {
	query := "SELECT COUNT(*) as total, COALESCE(SUM(CASE WHEN DATE(created_at) = CURDATE() THEN 1 ELSE 0 END), 0) as conversations_today, COALESCE(SUM(CASE WHEN status = 'active' THEN 1 ELSE 0 END), 0) as active, COALESCE(SUM(CASE WHEN status = 'resolved' AND DATE(resolved_at) = CURDATE() THEN 1 ELSE 0 END), 0) as resolved_today, COALESCE(SUM(CASE WHEN is_ai_transferred = false THEN 1 ELSE 0 END), 0) as ai_handled, COALESCE(COUNT(DISTINCT CASE WHEN status = 'escalated' THEN id END), 0) as escalated FROM conversations WHERE user_id = ?"
	row := r.db.QueryRowContext(ctx, query, userID)
	var total, conversationsToday, active, resolvedToday, aiHandled, escalated int
	err := row.Scan(&total, &conversationsToday, &active, &resolvedToday, &aiHandled, &escalated)
	if err != nil {
		return nil, err
	}
	aiRate := 0.0
	if total > 0 {
		aiRate = float64(aiHandled) / float64(total)
	}
	return map[string]interface{}{
		"total_conversations":  total,
		"conversations_today":  conversationsToday,
		"active_conversations": active,
		"resolved_today":       resolvedToday,
		"ai_resolution_rate":   aiRate,
		"escalated_count":      escalated,
	}, nil
}

func (r *ConversationRepository) CountByChannel(ctx context.Context, userID string) (map[string]int, error) {
	query := "SELECT channel, COUNT(*) as count FROM conversations WHERE user_id = ? GROUP BY channel"
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make(map[string]int)
	for rows.Next() {
		var channel string
		var count int
		if err := rows.Scan(&channel, &count); err == nil {
			result[channel] = count
		}
	}
	return result, nil
}

func (r *ConversationRepository) CountByIntent(ctx context.Context, userID string) ([]map[string]interface{}, error) {
	query := "SELECT intent, COUNT(*) as count FROM conversations WHERE user_id = ? GROUP BY intent ORDER BY count DESC LIMIT 5"
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []map[string]interface{}
	for rows.Next() {
		var intent string
		var count int
		if err := rows.Scan(&intent, &count); err == nil {
			result = append(result, map[string]interface{}{"intent": intent, "count": count})
		}
	}
	return result, nil
}

func (r *ConversationRepository) CountByHour(ctx context.Context, userID string) ([]map[string]interface{}, error) {
	query := "SELECT HOUR(created_at) as hour, COUNT(*) as count FROM conversations WHERE user_id = ? AND created_at >= DATE_SUB(NOW(), INTERVAL 7 DAY) GROUP BY HOUR(created_at) ORDER BY hour"
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []map[string]interface{}
	for rows.Next() {
		var hour, count int
		if err := rows.Scan(&hour, &count); err == nil {
			hourStr := fmt.Sprintf("%02d:00", hour)
			result = append(result, map[string]interface{}{"hour": hourStr, "volume": count})
		}
	}
	return result, nil
}

func (r *ConversationRepository) CountByDate(ctx context.Context, userID string, days int) ([]map[string]interface{}, error) {
	query := "SELECT DATE(created_at) as date, COUNT(*) as count FROM conversations WHERE user_id = ? AND created_at >= DATE_SUB(NOW(), INTERVAL ? DAY) GROUP BY DATE(created_at) ORDER BY date"
	rows, err := r.db.QueryContext(ctx, query, userID, days)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []map[string]interface{}
	for rows.Next() {
		var date string
		var count int
		if err := rows.Scan(&date, &count); err == nil {
			// Format date string to YYYY-MM-DD cleanly
			if len(date) > 10 {
				date = date[:10]
			}
			result = append(result, map[string]interface{}{"date": date, "conversations": count})
		}
	}
	return result, nil
}

// ========== USER OWNER WHATSAPP ==========

func (r *UserRepository) GetOwnerWhatsApp(ctx context.Context, userID string) (string, error) {
	var phone string
	err := r.db.QueryRowContext(ctx, "SELECT COALESCE(owner_whatsapp, '') FROM users WHERE id = ?", userID).Scan(&phone)
	if err != nil {
		return "", err
	}
	return phone, nil
}

// ========== INVENTORY REPOSITORY ==========

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

// ========== HANDOFF REPOSITORY ==========

type HandoffRepository struct {
	db    *sql.DB
	redis *infrastructure.RedisClient
}

func NewHandoffRepository(db *sql.DB, redis *infrastructure.RedisClient) *HandoffRepository {
	return &HandoffRepository{db: db, redis: redis}
}

func (r *HandoffRepository) Create(ctx context.Context, h *domain.Handoff) error {
	if h.ID == "" {
		h.ID = generateUUID()
	}
	if h.Quantity == 0 {
		h.Quantity = 1
	}
	query := `INSERT INTO handoffs (id, user_id, conversation_id, customer_name, customer_phone, customer_whatsapp, customer_location, product_name, original_price, agreed_price, quantity, status, created_at, updated_at)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())`
	_, err := r.db.ExecContext(ctx, query, h.ID, h.UserID, h.ConversationID, h.CustomerName, h.CustomerPhone, h.CustomerWhatsapp, h.CustomerLocation, h.ProductName, h.OriginalPrice, h.AgreedPrice, h.Quantity, h.Status)
	return err
}

func (r *HandoffRepository) GetByID(ctx context.Context, id string, userID string) (*domain.Handoff, error) {
	h := &domain.Handoff{}
	row := r.db.QueryRowContext(ctx, `SELECT id, user_id, conversation_id, customer_name, customer_phone, customer_whatsapp, customer_location, product_name, original_price, agreed_price, quantity, status, final_price, owner_notes, owner_notified_at, reminder_count, next_reminder_at, created_at, updated_at FROM handoffs WHERE id = ? AND user_id = ?`, id, userID)
	err := row.Scan(&h.ID, &h.UserID, &h.ConversationID, &h.CustomerName, &h.CustomerPhone, &h.CustomerWhatsapp, &h.CustomerLocation, &h.ProductName, &h.OriginalPrice, &h.AgreedPrice, &h.Quantity, &h.Status, &h.FinalPrice, &h.OwnerNotes, &h.OwnerNotifiedAt, &h.ReminderCount, &h.NextReminderAt, &h.CreatedAt, &h.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return h, nil
}

func (r *HandoffRepository) List(ctx context.Context, userID string, status string) ([]domain.Handoff, error) {
	query := `SELECT id, user_id, conversation_id, customer_name, customer_phone, customer_whatsapp, customer_location, product_name, original_price, agreed_price, quantity, status, final_price, owner_notes, owner_notified_at, reminder_count, next_reminder_at, created_at, updated_at FROM handoffs WHERE user_id = ?`
	args := []interface{}{userID}
	if status != "" {
		query += " AND status = ?"
		args = append(args, status)
	}
	query += " ORDER BY created_at DESC"
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var handoffs []domain.Handoff
	for rows.Next() {
		var h domain.Handoff
		if err := rows.Scan(&h.ID, &h.UserID, &h.ConversationID, &h.CustomerName, &h.CustomerPhone, &h.CustomerWhatsapp, &h.CustomerLocation, &h.ProductName, &h.OriginalPrice, &h.AgreedPrice, &h.Quantity, &h.Status, &h.FinalPrice, &h.OwnerNotes, &h.OwnerNotifiedAt, &h.ReminderCount, &h.NextReminderAt, &h.CreatedAt, &h.UpdatedAt); err != nil {
			continue
		}
		handoffs = append(handoffs, h)
	}
	return handoffs, nil
}

func (r *HandoffRepository) UpdateStatus(ctx context.Context, id string, userID string, status string, notes string) error {
	query := `UPDATE handoffs SET status=?, owner_notes=?, updated_at=NOW() WHERE id=? AND user_id=?`
	_, err := r.db.ExecContext(ctx, query, status, notes, id, userID)
	return err
}

func (r *HandoffRepository) GetPending(ctx context.Context, userID string) ([]domain.Handoff, error) {
	return r.List(ctx, userID, "pending")
}

func (r *HandoffRepository) GetReadyForReminder(ctx context.Context) ([]domain.Handoff, error) {
	query := `SELECT id, user_id, conversation_id, customer_name, customer_phone, customer_whatsapp, customer_location, product_name, original_price, agreed_price, quantity, status, final_price, owner_notes, owner_notified_at, reminder_count, next_reminder_at, created_at, updated_at FROM handoffs WHERE status = 'pending' AND next_reminder_at IS NOT NULL AND next_reminder_at <= NOW() AND reminder_count < 3`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var handoffs []domain.Handoff
	for rows.Next() {
		var h domain.Handoff
		if err := rows.Scan(&h.ID, &h.UserID, &h.ConversationID, &h.CustomerName, &h.CustomerPhone, &h.CustomerWhatsapp, &h.CustomerLocation, &h.ProductName, &h.OriginalPrice, &h.AgreedPrice, &h.Quantity, &h.Status, &h.FinalPrice, &h.OwnerNotes, &h.OwnerNotifiedAt, &h.ReminderCount, &h.NextReminderAt, &h.CreatedAt, &h.UpdatedAt); err != nil {
			continue
		}
		handoffs = append(handoffs, h)
	}
	return handoffs, nil
}

func (r *HandoffRepository) IncrementReminder(ctx context.Context, id string) error {
	next := time.Now().Add(15 * time.Minute)
	_, err := r.db.ExecContext(ctx, "UPDATE handoffs SET reminder_count = reminder_count + 1, next_reminder_at = ?, owner_notified_at = NOW(), updated_at = NOW() WHERE id = ?", next, id)
	return err
}

func (r *HandoffRepository) Expire(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, "UPDATE handoffs SET status = 'expired', updated_at = NOW() WHERE id = ?", id)
	return err
}
