package postgres

import (
	"context"

	"github.com/LSUDOKOS/signal/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepo struct {
	pool *pgxpool.Pool
}

func NewUserRepository(db *DB) *UserRepo {
	return &UserRepo{pool: db.Pool()}
}

func (r *UserRepo) Create(ctx context.Context, user *domain.User) error {
	query := `INSERT INTO users (slack_user_id, slack_team_id, email, display_name, neurotype)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at`
	return r.pool.QueryRow(ctx, query,
		user.SlackUserID, user.SlackTeamID, user.Email, user.DisplayName, user.Neurotype,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
}

func (r *UserRepo) GetBySlackID(ctx context.Context, slackUserID, slackTeamID string) (*domain.User, error) {
	query := `SELECT id, slack_user_id, slack_team_id, email, display_name, neurotype, created_at, updated_at
		FROM users WHERE slack_user_id = $1 AND slack_team_id = $2`
	user := &domain.User{}
	err := r.pool.QueryRow(ctx, query, slackUserID, slackTeamID).Scan(
		&user.ID, &user.SlackUserID, &user.SlackTeamID, &user.Email, &user.DisplayName, &user.Neurotype,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	query := `SELECT id, slack_user_id, slack_team_id, email, display_name, neurotype, created_at, updated_at
		FROM users WHERE id = $1`
	user := &domain.User{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&user.ID, &user.SlackUserID, &user.SlackTeamID, &user.Email, &user.DisplayName, &user.Neurotype,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepo) Update(ctx context.Context, user *domain.User) error {
	query := `UPDATE users SET email=$1, display_name=$2, neurotype=$3, updated_at=NOW()
		WHERE id=$4`
	_, err := r.pool.Exec(ctx, query, user.Email, user.DisplayName, user.Neurotype, user.ID)
	return err
}

func (r *UserRepo) Upsert(ctx context.Context, user *domain.User) error {
	query := `INSERT INTO users (slack_user_id, slack_team_id, email, display_name, neurotype)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (slack_user_id) DO UPDATE SET
			email = EXCLUDED.email,
			display_name = EXCLUDED.display_name,
			neurotype = COALESCE(EXCLUDED.neurotype, users.neurotype),
			updated_at = NOW()
		RETURNING id, created_at, updated_at`
	return r.pool.QueryRow(ctx, query,
		user.SlackUserID, user.SlackTeamID, user.Email, user.DisplayName, user.Neurotype,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
}
