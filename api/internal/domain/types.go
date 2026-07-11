package domain

import (
	"time"

	"github.com/google/uuid"
)

// User represents a Slack user who has installed Signal.
type User struct {
	ID          uuid.UUID `json:"id" db:"id"`
	SlackUserID string    `json:"slack_user_id" db:"slack_user_id"`
	SlackTeamID string    `json:"slack_team_id" db:"slack_team_id"`
	Email       string    `json:"email,omitempty" db:"email"`
	DisplayName string    `json:"display_name,omitempty" db:"display_name"`
	Neurotype   string    `json:"neurotype,omitempty" db:"neurotype"` // adhd, autism, anxiety, unspecified, ally
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// UserPreferences stores per-user Signal configuration.
type UserPreferences struct {
	UserID            uuid.UUID `json:"user_id" db:"user_id"`
	FocusModeEnabled  bool      `json:"focus_mode_enabled" db:"focus_mode_enabled"`
	FocusThreshold    int       `json:"focus_threshold" db:"focus_threshold"`           // messages per 10 min
	TranslatorEnabled bool      `json:"translator_enabled" db:"translator_enabled"`
	DigestEnabled     bool      `json:"digest_enabled" db:"digest_enabled"`
	DigestHour        int       `json:"digest_hour" db:"digest_hour"` // 0-23
	DeepWorkAutoDetect bool     `json:"deep_work_auto_detect" db:"deep_work_auto_detect"`
	QuietHoursStart   string    `json:"quiet_hours_start" db:"quiet_hours_start"` // HH:MM
	QuietHoursEnd     string    `json:"quiet_hours_end" db:"quiet_hours_end"`     // HH:MM
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
}

// Channel represents a Slack channel that Signal operates in.
type Channel struct {
	ID               uuid.UUID `json:"id" db:"id"`
	SlackChannelID   string    `json:"slack_channel_id" db:"slack_channel_id"`
	SlackTeamID      string    `json:"slack_team_id" db:"slack_team_id"`
	Name             string    `json:"name,omitempty" db:"name"`
	IsDM             bool      `json:"is_dm" db:"is_dm"`
	FocusModeEnabled bool      `json:"focus_mode_enabled" db:"focus_mode_enabled"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
}

// ChannelSubscription tracks per-user channel settings.
type ChannelSubscription struct {
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	ChannelID uuid.UUID `json:"channel_id" db:"channel_id"`
	Muted     bool      `json:"muted" db:"muted"`
}

// Digest represents a batched digest sent to a user.
type Digest struct {
	ID              uuid.UUID  `json:"id" db:"id"`
	UserID          uuid.UUID  `json:"user_id" db:"user_id"`
	SentAt          time.Time  `json:"sent_at" db:"sent_at"`
	MentionCount    int        `json:"mention_count" db:"mention_count"`
	ThreadReplyCount int       `json:"thread_reply_count" db:"thread_reply_count"`
	Content         []byte     `json:"content" db:"content"` // JSONB
	Status          string     `json:"status" db:"status"`   // pending, sent, read
}

// FocusSummary stores an AI-generated summary trigger.
type FocusSummary struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	ChannelID    uuid.UUID  `json:"channel_id" db:"channel_id"`
	TriggeredAt  time.Time  `json:"triggered_at" db:"triggered_at"`
	MessageCount int        `json:"message_count" db:"message_count"`
	SummaryText  string     `json:"summary_text" db:"summary_text"`
	AIModel      string     `json:"ai_model" db:"ai_model"`
	RawMessages  []byte     `json:"raw_messages" db:"raw_messages"` // JSONB
}

// Translation stores a social translation record.
type Translation struct {
	ID                  uuid.UUID `json:"id" db:"id"`
	UserID              uuid.UUID `json:"user_id" db:"user_id"`
	OriginalMessageTs   string    `json:"original_message_ts" db:"original_message_ts"`
	OriginalChannelID    string    `json:"original_channel_id" db:"original_channel_id"`
	OriginalText        string    `json:"original_text" db:"original_text"`
	TranslationText     string    `json:"translation_text" db:"translation_text"`
	Tone                string    `json:"tone" db:"tone"`
	Intent              string    `json:"intent" db:"intent"`
	Action              string    `json:"action" db:"action"`
	CreatedAt           time.Time `json:"created_at" db:"created_at"`
}

// ToneAnalysis is the structured result from AI tone analysis.
type ToneAnalysis struct {
	Tone   string `json:"tone"`
	Intent string `json:"intent"`
	Action string `json:"action"`
	Note   string `json:"note"`
}

// FocusSummaryResult is the AI-generated decision tree summary.
type FocusSummaryResult struct {
	Decisions   []DecisionItem `json:"decisions,omitempty"`
	NoDecisions bool           `json:"no_decisions,omitempty"`
}

// DecisionItem is a single decision + action items.
type DecisionItem struct {
	Decision    string       `json:"decision"`
	ActionItems []ActionItem `json:"action_items,omitempty"`
}

// ActionItem represents one action from a decision.
type ActionItem struct {
	Description string `json:"description"`
	Owner       string `json:"owner,omitempty"`
	Due         string `json:"due,omitempty"`
}

// CatchUpResult is the structured semantic search result.
type CatchUpResult struct {
	Topics       []CatchUpTopic `json:"topics"`
	MessageCount int            `json:"message_count"`
}

// CatchUpTopic is one topic in a catch-up digest.
type CatchUpTopic struct {
	Name    string `json:"name"`
	Decision string `json:"decision"`
	Action  string `json:"action"`
	Context string `json:"context"`
}

// DigestContent holds the structured content of a digest.
type DigestContent struct {
	Urgent       []DigestItem `json:"urgent"`
	ActionRequired []DigestItem `json:"action_required"`
	FYI          []DigestItem `json:"fyi"`
	ThreadReplies []DigestItem `json:"thread_replies"`
}

// DigestItem is a single item in a digest.
type DigestItem struct {
	From    string `json:"from"`
	Message string `json:"message"`
	Link    string `json:"link,omitempty"`
	Channel string `json:"channel,omitempty"`
}
