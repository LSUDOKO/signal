package features

import (
	"context"
	"testing"
	"time"

	"github.com/LSUDOKOS/signal/internal/domain"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

// mockCache implements store.CacheClient for testing.
type mockCache struct {
	velocity      int
	offered       bool
	deepWork      bool
	lastDigest    time.Time
	aiLimit       bool
}

func (m *mockCache) IncrChannelVelocity(ctx context.Context, channelID string) (int, error) {
	m.velocity++
	return m.velocity, nil
}

func (m *mockCache) GetChannelVelocity(ctx context.Context, channelID string) (int, error) {
	return m.velocity, nil
}

func (m *mockCache) ResetChannelVelocity(ctx context.Context, channelID string) error {
	m.velocity = 0
	return nil
}

func (m *mockCache) SetFocusOffered(ctx context.Context, channelID string, ttl time.Duration) error {
	m.offered = true
	return nil
}

func (m *mockCache) HasFocusBeenOffered(ctx context.Context, channelID string) (bool, error) {
	return m.offered, nil
}

func (m *mockCache) SetSession(ctx context.Context, slackUserID string, accessToken string, ttl time.Duration) error {
	return nil
}

func (m *mockCache) GetSession(ctx context.Context, slackUserID string) (string, error) {
	return "", nil
}

func (m *mockCache) SetDeepWork(ctx context.Context, userID string, duration time.Duration) error {
	m.deepWork = true
	return nil
}

func (m *mockCache) GetDeepWork(ctx context.Context, userID string) (time.Duration, error) {
	if m.deepWork {
		return 120 * time.Minute, nil
	}
	return 0, nil
}

func (m *mockCache) ClearDeepWork(ctx context.Context, userID string) error {
	m.deepWork = false
	return nil
}

func (m *mockCache) CheckAILimit(ctx context.Context, userID string, limit int) (bool, error) {
	return true, nil
}

func (m *mockCache) SetLastDigest(ctx context.Context, userID string, timestamp time.Time) error {
	m.lastDigest = timestamp
	return nil
}

func (m *mockCache) GetLastDigest(ctx context.Context, userID string) (time.Time, error) {
	return m.lastDigest, nil
}

// TestFocusModeService_HandleMessage_SkipDM verifies DMs are skipped.
func TestFocusModeService_HandleMessage_SkipDM(t *testing.T) {
	svc := &FocusModeService{
		cache: &mockCache{},
	}

	prefs := &domain.UserPreferences{FocusModeEnabled: true, FocusThreshold: 50}
	event := &slackevents.MessageEvent{
		ChannelType: "im",
		Text:        "Hello",
		User:        "U123",
	}

	err := svc.HandleMessage(context.Background(), event, &domain.User{}, prefs)
	if err != nil {
		t.Errorf("HandleMessage on DM should not error: %v", err)
	}
}

// TestFocusModeService_HandleMessage_BelowThreshold verifies no trigger below threshold.
func TestFocusModeService_HandleMessage_BelowThreshold(t *testing.T) {
	mock := &mockCache{}
	svc := &FocusModeService{
		cache: mock,
	}

	prefs := &domain.UserPreferences{FocusModeEnabled: true, FocusThreshold: 50}
	event := &slackevents.MessageEvent{
		ChannelType: "channel",
		Channel:     "C123",
		Text:        "Hello",
		User:        "U123",
	}

	// Send 49 messages (below threshold) - no trigger
	for i := 0; i < 49; i++ {
		err := svc.HandleMessage(context.Background(), event, &domain.User{}, prefs)
		if err != nil {
			t.Fatalf("message %d: HandleMessage error: %v", i, err)
		}
	}

	if mock.offered {
		t.Error("focus should not be offered below threshold")
	}
}

// TestFocusModeService_HandleBlockAction_EmptyActions verifies no panic on empty actions.
func TestFocusModeService_HandleBlockAction_EmptyActions(t *testing.T) {
	svc := &FocusModeService{}

	err := svc.HandleBlockAction(context.Background(), &slack.InteractionCallback{}, &domain.User{})
	if err != nil {
		t.Errorf("HandleBlockAction with empty actions should not error: %v", err)
	}
}

// TestIsFocusAction verifies action ID detection.
func TestIsFocusAction(t *testing.T) {
	tests := []struct {
		name     string
		actionID string
		want     bool
	}{
		{"get summary", "focus_get_summary", true},
		{"mute 30", "focus_mute_30", true},
		{"open thread", "focus_open_thread", true},
		{"random action", "some_other_action", false},
		{"empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isFocusAction(tt.actionID); got != tt.want {
				t.Errorf("isFocusAction(%q) = %v, want %v", tt.actionID, got, tt.want)
			}
		})
	}
}
