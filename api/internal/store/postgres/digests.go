package postgres

import (
	"context"

	"github.com/LSUDOKOS/signal/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DigestRepo struct {
	pool *pgxpool.Pool
}

func NewDigestRepository(db *DB) *DigestRepo {
	return &DigestRepo{pool: db.Pool()}
}

func NewFocusSummaryRepository(db *DB) *FocusSummaryRepo {
	return &FocusSummaryRepo{pool: db.Pool()}
}

func NewTranslationRepository(db *DB) *TranslationRepo {
	return &TranslationRepo{pool: db.Pool()}
}

func (r *DigestRepo) Create(ctx context.Context, d *domain.Digest) error {
	query := `INSERT INTO digests (user_id, mention_count, thread_reply_count, content, status)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, sent_at`
	return r.pool.QueryRow(ctx, query,
		d.UserID, d.MentionCount, d.ThreadReplyCount, d.Content, d.Status,
	).Scan(&d.ID, &d.SentAt)
}

func (r *DigestRepo) GetByUser(ctx context.Context, userID uuid.UUID, limit int) ([]domain.Digest, error) {
	query := `SELECT id, user_id, sent_at, mention_count, thread_reply_count, content, status
		FROM digests WHERE user_id = $1 ORDER BY sent_at DESC LIMIT $2`
	rows, err := r.pool.Query(ctx, query, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var digests []domain.Digest
	for rows.Next() {
		var d domain.Digest
		if err := rows.Scan(&d.ID, &d.UserID, &d.SentAt, &d.MentionCount, &d.ThreadReplyCount, &d.Content, &d.Status); err != nil {
			return nil, err
		}
		digests = append(digests, d)
	}
	return digests, nil
}

func (r *DigestRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	_, err := r.pool.Exec(ctx, `UPDATE digests SET status=$1 WHERE id=$2`, status, id)
	return err
}

type FocusSummaryRepo struct {
	pool *pgxpool.Pool
}

func NewFocusSummaryRepo(pool *pgxpool.Pool) *FocusSummaryRepo {
	return &FocusSummaryRepo{pool: pool}
}

func (r *FocusSummaryRepo) Create(ctx context.Context, fs *domain.FocusSummary) error {
	query := `INSERT INTO focus_summaries (channel_id, message_count, summary_text, ai_model, raw_messages)
		VALUES ($1, $2, $3, $4, $5) RETURNING id, triggered_at`
	return r.pool.QueryRow(ctx, query,
		fs.ChannelID, fs.MessageCount, fs.SummaryText, fs.AIModel, fs.RawMessages,
	).Scan(&fs.ID, &fs.TriggeredAt)
}

func (r *FocusSummaryRepo) GetByChannel(ctx context.Context, channelID uuid.UUID, limit int) ([]domain.FocusSummary, error) {
	query := `SELECT id, channel_id, triggered_at, message_count, summary_text, ai_model, raw_messages
		FROM focus_summaries WHERE channel_id = $1 ORDER BY triggered_at DESC LIMIT $2`
	rows, err := r.pool.Query(ctx, query, channelID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var summaries []domain.FocusSummary
	for rows.Next() {
		var fs domain.FocusSummary
		if err := rows.Scan(&fs.ID, &fs.ChannelID, &fs.TriggeredAt, &fs.MessageCount, &fs.SummaryText, &fs.AIModel, &fs.RawMessages); err != nil {
			return nil, err
		}
		summaries = append(summaries, fs)
	}
	return summaries, nil
}

type TranslationRepo struct {
	pool *pgxpool.Pool
}

func NewTranslationRepo(pool *pgxpool.Pool) *TranslationRepo {
	return &TranslationRepo{pool: pool}
}

func (r *TranslationRepo) Create(ctx context.Context, t *domain.Translation) error {
	query := `INSERT INTO translations (user_id, original_message_ts, original_channel_id, original_text, translation_text, tone, intent, action)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id, created_at`
	return r.pool.QueryRow(ctx, query,
		t.UserID, t.OriginalMessageTs, t.OriginalChannelID, t.OriginalText,
		t.TranslationText, t.Tone, t.Intent, t.Action,
	).Scan(&t.ID, &t.CreatedAt)
}

func (r *TranslationRepo) GetByUser(ctx context.Context, userID uuid.UUID, limit int) ([]domain.Translation, error) {
	query := `SELECT id, user_id, original_message_ts, original_channel_id, original_text, translation_text, tone, intent, action, created_at
		FROM translations WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2`
	rows, err := r.pool.Query(ctx, query, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var translations []domain.Translation
	for rows.Next() {
		var t domain.Translation
		if err := rows.Scan(&t.ID, &t.UserID, &t.OriginalMessageTs, &t.OriginalChannelID,
			&t.OriginalText, &t.TranslationText, &t.Tone, &t.Intent, &t.Action, &t.CreatedAt); err != nil {
			return nil, err
		}
		translations = append(translations, t)
	}
	return translations, nil
}
