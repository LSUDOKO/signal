package features

import (
	"context"
	"testing"

	"github.com/LSUDOKOS/signal/internal/domain"
	"github.com/slack-go/slack"
)

// TestParseDuration tests the duration string parser.
func TestParseDuration(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    int
		wantErr bool
	}{
		{"2 hours", "2h", 120, false},
		{"1 hour", "1h", 60, false},
		{"0 hours", "0h", 0, false},
		{"90 minutes", "90min", 90, false},
		{"30 minutes", "30min", 30, false},
		{"plain number 60", "60", 60, false},
		{"plain number 120", "120", 120, false},
		{"plain number 0", "0", 0, false},
		{"with spaces", "  2h  ", 120, false},
		{"uppercase H", "2H", 120, false},
		{"uppercase MIN", "90MIN", 90, false},
		// Error cases
		{"invalid string", "abc", 0, true},
		{"empty string", "", 0, true},
		{"negative number", "-30", -30, false},
		{"invalid format", "2hours", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseDuration(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseDuration(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("parseDuration(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

// TestIsDeepWorkAction verifies action ID detection.
func TestIsDeepWorkAction(t *testing.T) {
	tests := []struct {
		name     string
		actionID string
		want     bool
	}{
		{"start", "deepwork_start", true},
		{"stop", "deepwork_stop", true},
		{"extend", "deepwork_extend", true},
		{"other", "focus_mute_30", false},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isDeepWorkAction(tt.actionID); got != tt.want {
				t.Errorf("isDeepWorkAction(%q) = %v, want %v", tt.actionID, got, tt.want)
			}
		})
	}
}

// TestDeepWorkService_IsInDeepWork verifies deep work state detection.
func TestDeepWorkService_IsInDeepWork(t *testing.T) {
	mock := &mockCache{}
	svc := &DeepWorkService{
		cache: mock,
	}

	if svc.IsInDeepWork(context.Background(), "U123") {
		t.Error("should not be in deep work initially")
	}

	mock.deepWork = "active"

	if !svc.IsInDeepWork(context.Background(), "U123") {
		t.Error("should be in deep work after setting state")
	}
}

// TestDeepWorkService_HandleBlockAction_Empty verifies no panic on empty actions.
func TestDeepWorkService_HandleBlockAction_Empty(t *testing.T) {
	svc := &DeepWorkService{
		slack: &mockSlackAPI{},
		cache: &mockCache{},
	}

	err := svc.HandleBlockAction(context.Background(), &slack.InteractionCallback{}, &domain.User{SlackUserID: "U123"})
	if err != nil {
		t.Errorf("HandleBlockAction with empty actions should not error: %v", err)
	}
}
