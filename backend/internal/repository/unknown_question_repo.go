package repository

import (
	"context"
	"database/sql"
	"strings"

	"noant/internal/domain"
	"noant/internal/infrastructure"
)

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

func (r *UnknownQuestionRepository) List(ctx context.Context, userID string, status string, limit int, offset int) ([]domain.UnknownQuestion, error) {
	query := `SELECT id, question, conversation_id, channel, status, suggested_answer, category_id, created_at FROM unknown_questions WHERE user_id = ?`
	args := []interface{}{userID}
	if status != "" {
		query += " AND status = ?"
		args = append(args, status)
	}
	query += " ORDER BY created_at DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	questions := make([]domain.UnknownQuestion, 0)
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

func (r *UnknownQuestionRepository) BatchTrain(ctx context.Context, userID, answer, categoryID string, ids []string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, id := range ids {
		var question string
		err := tx.QueryRowContext(ctx, `SELECT question FROM unknown_questions WHERE id = ? AND user_id = ?`, id, userID).Scan(&question)
		if err != nil {
			continue
		}
		qaID := generateUUID()
		_, err = tx.ExecContext(ctx, `INSERT INTO qa_pairs (id, user_id, category_id, question, answer, is_active, created_at) VALUES (?, ?, ?, ?, ?, true, NOW())`, qaID, userID, categoryID, question, answer)
		if err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx, `UPDATE unknown_questions SET status = 'trained', suggested_answer = ?, category_id = ? WHERE id = ? AND user_id = ?`, answer, categoryID, id, userID)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *UnknownQuestionRepository) BatchIgnore(ctx context.Context, userID string, ids []string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, id := range ids {
		_, err := tx.ExecContext(ctx, `UPDATE unknown_questions SET status = 'ignored' WHERE id = ? AND user_id = ?`, id, userID)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *UnknownQuestionRepository) ExistsPending(ctx context.Context, userID string, question string) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM unknown_questions WHERE user_id = ? AND LOWER(question) = ? AND status = 'pending'`, userID, strings.ToLower(question)).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *UnknownQuestionRepository) UpdateStatus(ctx context.Context, id string, userID string, status string, answer *string, categoryID *string) error {
	query := `UPDATE unknown_questions SET status = ?, suggested_answer = ?, category_id = ? WHERE id = ? AND user_id = ?`
	_, err := r.db.ExecContext(ctx, query, status, answer, categoryID, id, userID)
	return err
}

func (r *UnknownQuestionRepository) Clear(ctx context.Context, userID string) error {
	query := `DELETE FROM unknown_questions WHERE user_id = ?`
	_, err := r.db.ExecContext(ctx, query, userID)
	return err
}

func (r *UnknownQuestionRepository) CountByStatus(ctx context.Context, userID string) (map[string]int, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT status, COUNT(*) as count FROM unknown_questions WHERE user_id = ? GROUP BY status`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := map[string]int{"pending": 0, "trained": 0, "ignored": 0}
	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err == nil {
			result[status] = count
		}
	}
	return result, nil
}

func (r *UnknownQuestionRepository) MostPopular(ctx context.Context, userID string, limit int) ([]map[string]interface{}, error) {
	query := `SELECT question, COUNT(*) as count FROM unknown_questions WHERE user_id = ? AND status = 'pending' GROUP BY question ORDER BY count DESC LIMIT ?`
	rows, err := r.db.QueryContext(ctx, query, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []map[string]interface{}
	for rows.Next() {
		var question string
		var count int
		if err := rows.Scan(&question, &count); err == nil {
			result = append(result, map[string]interface{}{"question": question, "count": count})
		}
	}
	return result, nil
}

func (r *UnknownQuestionRepository) CountByFilter(ctx context.Context, userID string, status string) (int, error) {
	query := `SELECT COUNT(*) FROM unknown_questions WHERE user_id = ?`
	args := []interface{}{userID}
	if status != "" {
		query += " AND status = ?"
		args = append(args, status)
	}
	var total int
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&total)
	if err != nil {
		return 0, err
	}
	return total, nil
}

func (r *UnknownQuestionRepository) CountByDate(ctx context.Context, userID string, days int) ([]map[string]interface{}, error) {
	query := `SELECT DATE(created_at) as date, COUNT(*) as count FROM unknown_questions WHERE user_id = ? AND created_at >= DATE_SUB(NOW(), INTERVAL ? DAY) GROUP BY DATE(created_at) ORDER BY date`
	rows, err := r.db.QueryContext(ctx, query, userID, days)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []map[string]interface{}
	for rows.Next() {
		var dateStr string
		var count int
		if err := rows.Scan(&dateStr, &count); err == nil {
			if len(dateStr) > 10 {
				dateStr = dateStr[:10]
			}
			result = append(result, map[string]interface{}{"date": dateStr, "unknown_count": count})
		}
	}
	return result, nil
}

func (r *UnknownQuestionRepository) CleanupStale(ctx context.Context, days int) (int64, error) {
	result, err := r.db.ExecContext(ctx,
		`DELETE FROM unknown_questions WHERE status IN ('trained', 'ignored') AND created_at < NOW() - INTERVAL ? DAY`, days)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
