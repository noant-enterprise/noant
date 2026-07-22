package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"

	"noant/internal/domain"
	"noant/internal/infrastructure"
)

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
	query := `INSERT INTO qa_pairs (id, user_id, org_id, category_id, question, answer, variations, is_active, created_at, updated_at)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())`
	variationsJSON := []byte("[]")
	if len(qa.Variations) > 0 {
		b, err := json.Marshal(qa.Variations)
		if err == nil {
			variationsJSON = b
		}
	}
	_, err := r.db.ExecContext(ctx, query, qa.ID, qa.UserID, qa.OrgID, qa.CategoryID, qa.Question, qa.Answer, string(variationsJSON), qa.IsActive)
	return err
}

func (r *QAPairRepository) BulkCreate(ctx context.Context, qas []domain.QAPair) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	query := `INSERT INTO qa_pairs (id, user_id, org_id, category_id, question, answer, variations, is_active, created_at, updated_at)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())`
	for i := range qas {
		if qas[i].ID == "" {
			qas[i].ID = generateUUID()
		}
		variationsJSON := []byte("[]")
		if len(qas[i].Variations) > 0 {
			b, err := json.Marshal(qas[i].Variations)
			if err == nil {
				variationsJSON = b
			}
		}
		_, err := tx.ExecContext(ctx, query, qas[i].ID, qas[i].UserID, qas[i].OrgID, qas[i].CategoryID, qas[i].Question, qas[i].Answer, string(variationsJSON), qas[i].IsActive)
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
	defer func() { _ = rows.Close() }()
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

func (r *QAPairRepository) ListByCategoryAndOrg(ctx context.Context, categoryID, orgID string) ([]domain.QAPair, error) {
	query := `SELECT id, category_id, question, answer, variations, is_active, usage_count, created_at, updated_at FROM qa_pairs WHERE category_id = ? AND org_id = ? AND is_active = true`
	rows, err := r.db.QueryContext(ctx, query, categoryID, orgID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
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
		qa.OrgID = orgID
		qas = append(qas, qa)
	}
	return qas, nil
}

func (r *QAPairRepository) Search(ctx context.Context, orgID, query string) ([]domain.QAPair, error) {
	// Clean query by removing common punctuation and trim
	cleanQuery := query
	for _, char := range []string{"?", "!", ".", ",", ";", ":"} {
		cleanQuery = strings.ReplaceAll(cleanQuery, char, "")
	}
	cleanQuery = strings.TrimSpace(cleanQuery)

	sqlQuery := `SELECT id, category_id, question, answer, variations, is_active, usage_count, created_at, updated_at
	FROM qa_pairs 
	WHERE org_id = ? 
	  AND is_active = true 
	  AND (
	      LOWER(question) LIKE LOWER(?) 
	      OR LOWER(?) LIKE CONCAT('%', LOWER(REPLACE(REPLACE(question, '?', ''), '!', '')), '%')
	      OR LOWER(REPLACE(REPLACE(question, '?', ''), '!', '')) LIKE LOWER(?)
	      OR LOWER(answer) LIKE LOWER(?)
	  ) LIMIT 10`

	searchTerm := "%" + cleanQuery + "%"
	rows, err := r.db.QueryContext(ctx, sqlQuery, orgID, searchTerm, cleanQuery, searchTerm, searchTerm)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
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

func (r *QAPairRepository) ListByOrg(ctx context.Context, orgID, categoryID string) ([]domain.QAPair, error) {
	query := `SELECT id, category_id, question, answer, variations, is_active, usage_count, created_at, updated_at FROM qa_pairs WHERE org_id = ? AND is_active = true`
	args := []interface{}{orgID}
	if categoryID != "" {
		query += " AND category_id = ?"
		args = append(args, categoryID)
	}
	query += " ORDER BY created_at DESC"
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
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

func (r *QAPairRepository) GetByQuestion(ctx context.Context, orgID, question string) (*domain.QAPair, error) {
	query := `SELECT id, category_id, question, answer, variations, is_active, usage_count, created_at, updated_at 
	FROM qa_pairs WHERE org_id = ? AND question = ? LIMIT 1`
	var qa domain.QAPair
	var variationsJSON sql.NullString
	err := r.db.QueryRowContext(ctx, query, orgID, question).Scan(
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
	qa.OrgID = orgID
	return &qa, nil
}

func (r *QAPairRepository) Update(ctx context.Context, qa *domain.QAPair) error {
	query := `UPDATE qa_pairs SET category_id = ?, question = ?, answer = ?, variations = ?, is_active = ?, updated_at = NOW() WHERE id = ? AND org_id = ?`
	variationsJSON := []byte("[]")
	if len(qa.Variations) > 0 {
		b, err := json.Marshal(qa.Variations)
		if err == nil {
			variationsJSON = b
		}
	}
	_, err := r.db.ExecContext(ctx, query, qa.CategoryID, qa.Question, qa.Answer, string(variationsJSON), qa.IsActive, qa.ID, qa.OrgID)
	return err
}

func (r *QAPairRepository) IncrementUsage(ctx context.Context, id string) error {
	query := `UPDATE qa_pairs SET usage_count = usage_count + 1 WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *QAPairRepository) CountByOrg(ctx context.Context, orgID string) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM qa_pairs WHERE org_id = ? AND is_active = true`, orgID).Scan(&count)
	return count, err
}

func (r *QAPairRepository) Delete(ctx context.Context, id, orgID string) error {
	query := `DELETE FROM qa_pairs WHERE id = ? AND org_id = ?`
	_, err := r.db.ExecContext(ctx, query, id, orgID)
	return err
}
