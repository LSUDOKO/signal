package store

import (
	"context"
	"time"

	"github.com/LSUDOKOS/signal/internal/domain"
	"github.com/google/uuid"
)

// UserRepository defines operations for user storage.
type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetBySlackID(ctx context.Context, slackUserID, slackTeamID string) (*domain.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
	Upsert(ctx context.Context, user *domain.User) error
}

// PreferencesRepository defines operations for user preferences.
type PreferencesRepository interface {
	Get(ctx context.Context, userID uuid.UUID) (*domain.UserPreferences, error)
	Update(ctx context.Context, prefs *domain.UserPreferences) error
	GetByDigestHour(ctx context.Context, hour int) ([]domain.UserPreferences, error)
	Upsert(ctx context.Context, prefs *domain.UserPreferences) error
}

// ChannelRepository defines operations for channel storage.
type ChannelRepository interface {
	Create(ctx context.Context, channel *domain.Channel) error
	GetBySlackID(ctx context.Context, slackChannelID, slackTeamID string) (*domain.Channel, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Channel, error)
	Update(ctx context.Context, channel *domain.Channel) error
	Upsert(ctx context.Context, channel *domain.Channel) error
}

// ChannelSubscriptionRepository defines operations for channel subscriptions.
type ChannelSubscriptionRepository interface {
	Create(ctx context.Context, sub *domain.ChannelSubscription) error
	Get(ctx context.Context, userID, channelID uuid.UUID) (*domain.ChannelSubscription, error)
	Update(ctx context.Context, sub *domain.ChannelSubscription) error
}

// DigestRepository defines operations for digests.
type DigestRepository interface {
	Create(ctx context.Context, digest *domain.Digest) error
	GetByUser(ctx context.Context, userID uuid.UUID, limit int) ([]domain.Digest, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) error
}

// FocusSummaryRepository defines operations for focus summaries.
type FocusSummaryRepository interface {
	Create(ctx context.Context, summary *domain.FocusSummary) error
	GetByChannel(ctx context.Context, channelID uuid.UUID, limit int) ([]domain.FocusSummary, error)
}

// TranslationRepository defines operations for translations.
type TranslationRepository interface {
	Create(ctx context.Context, t *domain.Translation) error
	GetByUser(ctx context.Context, userID uuid.UUID, limit int) ([]domain.Translation, error)
}

// CacheClient defines operations for Redis caching/state.
type CacheClient interface {
	// Channel velocity tracking
	IncrChannelVelocity(ctx context.Context, channelID string) (int, error)
	GetChannelVelocity(ctx context.Context, channelID string) (int, error)
	ResetChannelVelocity(ctx context.Context, channelID string) error
	SetFocusOffered(ctx context.Context, channelID string, ttl time.Duration) error
	HasFocusBeenOffered(ctx context.Context, channelID string) (bool, error)

	// Session management
	SetSession(ctx context.Context, slackUserID string, accessToken string, ttl time.Duration) error
	GetSession(ctx context.Context, slackUserID string) (string, error)

	// Deep work state
	SetDeepWork(ctx context.Context, userID string, duration time.Duration) error
	GetDeepWork(ctx context.Context, userID string) (time.Duration, error)
	ClearDeepWork(ctx context.Context, userID string) error

	// Rate limiting
	CheckAILimit(ctx context.Context, userID string, limit int) (bool, error)

	// Digest tracking
	SetLastDigest(ctx context.Context, userID string, timestamp time.Time) error
	GetLastDigest(ctx context.Context, userID string) (time.Time, error)
}
