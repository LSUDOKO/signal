package features

import (
	"context"
	"testing"

	"github.com/LSUDOKOS/signal/internal/domain"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

// TestTranslatorService_containsAmbiguousPhrase tests all 18 regex patterns.
func TestTranslatorService_containsAmbiguousPhrase(t *testing.T) {
	svc := &TranslatorService{}

	tests := []struct {
		name  string
		text  string
		match bool
	}{
		// Should match — 18 patterns
		{"per my last", "Per my last email, we need this done.", true},
		{"as I mentioned", "As I mentioned in the meeting yesterday...", true},
		{"just following up", "Just following up on this request.", true},
		{"let's take this offline", "Let's take this offline.", true},
		{"moving forward", "Moving forward, please update the report.", true},
		{"with all due respect", "With all due respect, that's incorrect.", true},
		{"friendly reminder", "Friendly reminder that the deadline is today.", true},
		{"circle back", "Let's circle back on this next week.", true},
		{"touch base", "I'd like to touch base about the project.", true},
		{"loop in", "Let me loop in Sarah on this thread.", true},
		{"per our conversation", "Per our conversation, I've attached the file.", true},
		{"going forward", "Going forward, use the new template.", true},
		{"as you know", "As you know, the deadline was extended.", true},
		{"not sure if you saw", "Not sure if you saw my earlier message.", true},
		{"just checking in", "Just checking in on the status.", true},
		{"per our discussion", "Per our discussion, here are the next steps.", true},
		{"any updates on this", "Any updates on this ticket?", true},
		{"per usual", "Per usual, the build is failing.", true},
		// Case insensitive
		{"case insensitive", "PER MY LAST email...", true},
		{"mixed case", "Just Following Up on this.", true},
		// Should not match — normal phrases
		{"normal thanks", "Thanks for your help!", false},
		{"normal question", "Can you review this PR?", false},
		{"normal update", "Here are the changes I made.", false},
		{"normal greeting", "Good morning team!", false},
		{"normal compliment", "Great work on the presentation.", false},
		{"empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := svc.containsAmbiguousPhrase(tt.text)
			if got != tt.match {
				t.Errorf("containsAmbiguousPhrase(%q) = %v, want %v", tt.text, got, tt.match)
			}
		})
	}
}

// TestTranslatorService_extractMentionedUsers tests user mention extraction.
func TestTranslatorService_extractMentionedUsers(t *testing.T) {
	svc := &TranslatorService{}

	tests := []struct {
		name string
		text string
		want []string
	}{
		{"single mention", "Hey <@U12345> can you review this?", []string{"U12345"}},
		{"multiple mentions", "<@U12345> and <@U67890> please review", []string{"U12345", "U67890"}},
		{"no mentions", "Hello everyone, please review.", nil},
		{"duplicate mentions", "<@U12345> and <@U12345> again", []string{"U12345"}},
		{"empty text", "", nil},
		{"with ambiguous phrase", "Per my last <@U12345>", []string{"U12345"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := svc.extractMentionedUsers(tt.text)
			if len(got) != len(tt.want) {
				t.Errorf("extractMentionedUsers(%q) = %v, want %v", tt.text, got, tt.want)
				return
			}
			for i, v := range got {
				if v != tt.want[i] {
					t.Errorf("extractMentionedUsers(%q)[%d] = %q, want %q", tt.text, i, v, tt.want[i])
				}
			}
		})
	}
}

// TestTranslatorService_HandleBlockAction verifies acknowledgment handling.
func TestTranslatorService_HandleBlockAction(t *testing.T) {
	svc := &TranslatorService{
		slack: &mockSlackAPI{},
	}

	tests := []struct {
		name     string
		actionID string
		wantErr  bool
	}{
		{"acknowledge", "translator_ack", false},
		{"unknown action", "some_action", false},
		{"empty action", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			action := &slack.InteractionCallback{}
			if tt.actionID != "" {
				action.ActionCallback.BlockActions = []*slack.BlockAction{
					{ActionID: tt.actionID, Value: "ack"},
				}
			}
			err := svc.HandleBlockAction(context.Background(), action, &domain.User{SlackUserID: "U123"})
			if (err != nil) != tt.wantErr {
				t.Errorf("HandleBlockAction() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestIsTranslatorAction verifies translator action ID detection.
func TestIsTranslatorAction(t *testing.T) {
	tests := []struct {
		name     string
		actionID string
		want     bool
	}{
		{"acknowledge", "translator_ack", true},
		{"other", "focus_get_summary", false},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isTranslatorAction(tt.actionID); got != tt.want {
				t.Errorf("isTranslatorAction(%q) = %v, want %v", tt.actionID, got, tt.want)
			}
		})
	}
}

// mockSlackAPI implements SlackAPI for testing.
type mockSlackAPI struct{}

func (m *mockSlackAPI) PostMessage(channelID string, blocks []slack.Block, text string) error {
	return nil
}

func (m *mockSlackAPI) PostEphemeral(channelID, userID string, blocks []slack.Block, text string) error {
	return nil
}

func (m *mockSlackAPI) OpenDMChannel(userID string) (string, error) {
	return "DM" + userID, nil
}

func (m *mockSlackAPI) GetUser(userID string) (*slack.User, error) {
	return &slack.User{ID: userID, Name: "test_user"}, nil
}

func (m *mockSlackAPI) GetChannelHistory(channelID string, limit int) ([]slack.Message, error) {
	return []slack.Message{}, nil
}

func (m *mockSlackAPI) SearchMessages(query string, params slack.SearchParameters) (*slack.SearchMessages, error) {
	return &slack.SearchMessages{}, nil
}

func (m *mockSlackAPI) SetUserStatus(userID, statusText, statusEmoji string, expiration int) error {
	return nil
}

// TestHandleMessage_NoAmbiguousPhrase verifies messages without ambiguous phrases are skipped.
func TestHandleMessage_NoAmbiguousPhrase(t *testing.T) {
	svc := &TranslatorService{}
	event := &slackevents.MessageEvent{
		Text:    "Great work on the project!",
		Channel: "C123",
		User:    "U123",
	}

	err := svc.HandleMessage(context.Background(), event, &domain.User{SlackUserID: "U456"})
	if err != nil {
		t.Errorf("HandleMessage with clean text should not error: %v", err)
	}
}

// TestHandleMessage_AmbiguousPhraseNoMention verifies no DM sent when no @mentions.
func TestHandleMessage_AmbiguousPhraseNoMention(t *testing.T) {
	svc := &TranslatorService{}
	event := &slackevents.MessageEvent{
		Text:    "Per my last email, this needs to be done.",
		Channel: "C123",
		User:    "U123",
	}

	err := svc.HandleMessage(context.Background(), event, &domain.User{SlackUserID: "U456"})
	if err != nil {
		t.Errorf("HandleMessage with no mentions should not error: %v", err)
	}
}
