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
