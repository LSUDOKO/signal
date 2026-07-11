package postgres

import (
	"context"

	"github.com/LSUDOKOS/signal/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PreferencesRepo struct {
	pool *pgxpool.Pool
}

func NewPreferencesRepository(db *DB) *PreferencesRepo {
	return &PreferencesRepo{pool: db.Pool()}
}

func (r *PreferencesRepo) Get(ctx context.Context, userID uuid.UUID) (*domain.UserPreferences, error) {
	query := `SELECT user_id, focus_mode_enabled, focus_threshold, translator_enabled,
		digest_enabled, digest_hour, deep_work_auto_detect, quiet_hours_start, quiet_hours_end,
		created_at, updated_at
		FROM user_preferences WHERE user_id = $1`
	p := &domain.UserPreferences{}
	err := r.pool.QueryRow(ctx, query, userID).Scan(
		&p.UserID, &p.FocusModeEnabled, &p.FocusThreshold, &p.TranslatorEnabled,
		&p.DigestEnabled, &p.DigestHour, &p.DeepWorkAutoDetect,
		&p.QuietHoursStart, &p.QuietHoursEnd, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (r *PreferencesRepo) Update(ctx context.Context, prefs *domain.UserPreferences) error {
	query := `UPDATE user_preferences SET
		focus_mode_enabled = $1, focus_threshold = $2, translator_enabled = $3,
		digest_enabled = $4, digest_hour = $5, deep_work_auto_detect = $6,
		quiet_hours_start = $7, quiet_hours_end = $8, updated_at = NOW()
		WHERE user_id = $9`
	_, err := r.pool.Exec(ctx, query,
		prefs.FocusModeEnabled, prefs.FocusThreshold, prefs.TranslatorEnabled,
		prefs.DigestEnabled, prefs.DigestHour, prefs.DeepWorkAutoDetect,
		prefs.QuietHoursStart, prefs.QuietHoursEnd, prefs.UserID,
	)
	return err
}

func (r *PreferencesRepo) Upsert(ctx context.Context, prefs *domain.UserPreferences) error {
	query := `INSERT INTO user_preferences (user_id, focus_mode_enabled, focus_threshold,
		translator_enabled, digest_enabled, digest_hour, deep_work_auto_detect,
		quiet_hours_start, quiet_hours_end)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (user_id) DO UPDATE SET
			focus_mode_enabled = EXCLUDED.focus_mode_enabled,
			focus_threshold = EXCLUDED.focus_threshold,
			translator_enabled = EXCLUDED.translator_enabled,
			digest_enabled = EXCLUDED.digest_enabled,
			digest_hour = EXCLUDED.digest_hour,
			deep_work_auto_detect = EXCLUDED.deep_work_auto_detect,
			quiet_hours_start = EXCLUDED.quiet_hours_start,
			quiet_hours_end = EXCLUDED.quiet_hours_end,
			updated_at = NOW()`
	_, err := r.pool.Exec(ctx, query,
		prefs.FocusModeEnabled, prefs.FocusThreshold, prefs.TranslatorEnabled,
		prefs.DigestEnabled, prefs.DigestHour, prefs.DeepWorkAutoDetect,
		prefs.QuietHoursStart, prefs.QuietHoursEnd, prefs.UserID,
	)
	return err
}

func (r *PreferencesRepo) GetByDigestHour(ctx context.Context, hour int) ([]domain.UserPreferences, error) {
	query := `SELECT user_id, focus_mode_enabled, focus_threshold, translator_enabled,
		digest_enabled, digest_hour, deep_work_auto_detect, quiet_hours_start, quiet_hours_end,
		created_at, updated_at
		FROM user_preferences WHERE digest_enabled = true AND digest_hour = $1`
	rows, err := r.pool.Query(ctx, query, hour)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prefs []domain.UserPreferences
	for rows.Next() {
		var p domain.UserPreferences
		if err := rows.Scan(
			&p.UserID, &p.FocusModeEnabled, &p.FocusThreshold, &p.TranslatorEnabled,
			&p.DigestEnabled, &p.DigestHour, &p.DeepWorkAutoDetect,
			&p.QuietHoursStart, &p.QuietHoursEnd, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, err
		}
		prefs = append(prefs, p)
	}
	return prefs, nil
}
