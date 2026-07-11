package postgres

import (
	"context"

	"github.com/LSUDOKOS/signal/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ChannelRepo struct {
	pool *pgxpool.Pool
}

func NewChannelRepository(db *DB) *ChannelRepo {
	return &ChannelRepo{pool: db.Pool()}
}

func (r *ChannelRepo) Create(ctx context.Context, ch *domain.Channel) error {
	query := `INSERT INTO channels (slack_channel_id, slack_team_id, name, is_dm, focus_mode_enabled)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at`
	return r.pool.QueryRow(ctx, query,
		ch.SlackChannelID, ch.SlackTeamID, ch.Name, ch.IsDM, ch.FocusModeEnabled,
	).Scan(&ch.ID, &ch.CreatedAt)
}

func (r *ChannelRepo) GetBySlackID(ctx context.Context, slackChannelID, slackTeamID string) (*domain.Channel, error) {
	query := `SELECT id, slack_channel_id, slack_team_id, name, is_dm, focus_mode_enabled, created_at
		FROM channels WHERE slack_channel_id = $1 AND slack_team_id = $2`
	ch := &domain.Channel{}
	err := r.pool.QueryRow(ctx, query, slackChannelID, slackTeamID).Scan(
		&ch.ID, &ch.SlackChannelID, &ch.SlackTeamID, &ch.Name, &ch.IsDM, &ch.FocusModeEnabled, &ch.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return ch, nil
}

func (r *ChannelRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Channel, error) {
	query := `SELECT id, slack_channel_id, slack_team_id, name, is_dm, focus_mode_enabled, created_at
		FROM channels WHERE id = $1`
	ch := &domain.Channel{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&ch.ID, &ch.SlackChannelID, &ch.SlackTeamID, &ch.Name, &ch.IsDM, &ch.FocusModeEnabled, &ch.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return ch, nil
}

func (r *ChannelRepo) Update(ctx context.Context, ch *domain.Channel) error {
	query := `UPDATE channels SET name=$1, focus_mode_enabled=$2 WHERE id=$3`
	_, err := r.pool.Exec(ctx, query, ch.Name, ch.FocusModeEnabled, ch.ID)
	return err
}

func (r *ChannelRepo) Upsert(ctx context.Context, ch *domain.Channel) error {
	query := `INSERT INTO channels (slack_channel_id, slack_team_id, name, is_dm, focus_mode_enabled)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (slack_channel_id) DO UPDATE SET
			name = EXCLUDED.name,
			is_dm = EXCLUDED.is_dm,
			updated_at = NOW()
		RETURNING id, created_at`
	return r.pool.QueryRow(ctx, query,
		ch.SlackChannelID, ch.SlackTeamID, ch.Name, ch.IsDM, ch.FocusModeEnabled,
	).Scan(&ch.ID, &ch.CreatedAt)
}
