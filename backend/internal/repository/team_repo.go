package repository

import (
	"context"
	"database/sql"

	"noant/internal/domain"
	"noant/internal/infrastructure"
)

type TeamRepository struct {
	db    *sql.DB
	redis *infrastructure.RedisClient
}

func NewTeamRepository(db *sql.DB, redis *infrastructure.RedisClient) *TeamRepository {
	return &TeamRepository{db: db, redis: redis}
}

func (r *TeamRepository) ListByOrg(ctx context.Context, orgID string) ([]domain.TeamMember, error) {
	query := `SELECT t.id, t.user_id, u.email, u.first_name, u.last_name, t.role, t.is_active, t.joined_at
	FROM team_members t JOIN users u ON t.user_id = u.id WHERE t.org_id = ?`
	rows, err := r.db.QueryContext(ctx, query, orgID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
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

func (r *TeamRepository) Create(ctx context.Context, orgID string, member *domain.TeamMember) error {
	if member.ID == "" {
		member.ID = generateUUID()
	}
	query := `INSERT INTO team_members (id, org_id, user_id, role, is_active, joined_at)
	VALUES (?, ?, ?, ?, ?, NOW())`
	_, err := r.db.ExecContext(ctx, query, member.ID, orgID, member.UserID, member.Role, member.IsActive)
	return err
}

func (r *TeamRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM team_members WHERE id = ?`, id)
	return err
}

func (r *TeamRepository) GetByID(ctx context.Context, id string) (*domain.TeamMember, error) {
	query := `SELECT t.id, t.user_id, u.email, u.first_name, u.last_name, t.role, t.is_active, t.joined_at
	FROM team_members t JOIN users u ON t.user_id = u.id WHERE t.id = ?`
	row := r.db.QueryRowContext(ctx, query, id)
	var m domain.TeamMember
	err := row.Scan(&m.ID, &m.UserID, &m.Email, &m.FirstName, &m.LastName, &m.Role, &m.IsActive, &m.JoinedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &m, nil
}

func (r *TeamRepository) GetByEmailAndOrg(ctx context.Context, email, orgID string) (*domain.TeamMember, error) {
	query := `SELECT t.id, t.user_id, u.email, u.first_name, u.last_name, t.role, t.is_active, t.joined_at
	FROM team_members t JOIN users u ON t.user_id = u.id WHERE u.email = ? AND t.org_id = ?`
	row := r.db.QueryRowContext(ctx, query, email, orgID)
	var m domain.TeamMember
	err := row.Scan(&m.ID, &m.UserID, &m.Email, &m.FirstName, &m.LastName, &m.Role, &m.IsActive, &m.JoinedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &m, nil
}
