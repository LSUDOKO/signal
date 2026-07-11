package features

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"
	"strings"

	"github.com/LSUDOKOS/signal/internal/ai"
	"github.com/LSUDOKOS/signal/internal/domain"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

// TranslatorService implements the Social Translator feature.
type TranslatorService struct {
	slack SlackAPI
	ai    *ai.Client
}

// NewTranslatorService creates a new Social Translator service.
func NewTranslatorService(slack SlackAPI, ai *ai.Client) *TranslatorService {
	return &TranslatorService{
		slack: slack,
		ai:    ai,
	}
}

// ambiguousPatterns matches passive-aggressive or ambiguous workplace phrases.
var ambiguousPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)\bper my last\b`),
	regexp.MustCompile(`(?i)\bas I mentioned\b`),
	regexp.MustCompile(`(?i)\bjust following up\b`),
	regexp.MustCompile(`(?i)\blet's take this offline\b`),
	regexp.MustCompile(`(?i)\bmoving forward\b`),
	regexp.MustCompile(`(?i)\bwith all due respect\b`),
	regexp.MustCompile(`(?i)\bfriendly reminder\b`),
	regexp.MustCompile(`(?i)\bcircle back\b`),
	regexp.MustCompile(`(?i)\btouch base\b`),
	regexp.MustCompile(`(?i)\bloop in\b`),
	regexp.MustCompile(`(?i)\bper our conversation\b`),
	regexp.MustCompile(`(?i)\bgoing forward\b`),
	regexp.MustCompile(`(?i)\bas you know\b`),
	regexp.MustCompile(`(?i)\bnot sure if you saw\b`),
	regexp.MustCompile(`(?i)\bjust checking in\b`),
	regexp.MustCompile(`(?i)\bper our discussion\b`),
	regexp.MustCompile(`(?i)\bany updates on this\b`),
	regexp.MustCompile(`(?i)\bper usual\b`),
}

// userMentionPattern extracts @user mentions from a message.
var userMentionPattern = regexp.MustCompile(`<@([A-Z0-9]+)>`)

// HandleMessage checks for ambiguous language and triggers translation.
func (t *TranslatorService) HandleMessage(ctx context.Context, event *slackevents.MessageEvent, user *domain.User) error {
	// Check if message contains ambiguous language
	if !t.containsAmbiguousPhrase(event.Text) {
		// Check if it's a reply in a thread to a previous message
		if event.ThreadTimeStamp == "" {
			return nil
		}
	}

	// Analyze tone
	analysis, err := t.ai.AnalyzeTone(ctx, event.Text)
	if err != nil {
		slog.Error("ai tone analysis failed", "error", err)
		return nil // Don't fail the message flow
	}

	// Extract mentioned users
	mentionedUsers := t.extractMentionedUsers(event.Text)

	// Send translation DM: to mentioned users if any, otherwise to the current user
	if len(mentionedUsers) > 0 {
		for _, mentionedUser := range mentionedUsers {
			if err := t.sendTranslationDM(ctx, mentionedUser, event, analysis); err != nil {
				slog.Error("failed to send translation dm", "error", err, "user", mentionedUser)
			}
		}
	} else {
		// No @-mentions but ambiguous language detected; send to the current user
		if err := t.sendTranslationDM(ctx, user.SlackUserID, event, analysis); err != nil {
			slog.Error("failed to send translation dm to user", "error", err)
		}
	}

	return nil
}

// HandleSlashCommand handles the /translate command.
func (t *TranslatorService) HandleSlashCommand(ctx context.Context, cmd *slack.SlashCommand, user *domain.User) error {
	message := cmd.Text
	if strings.TrimSpace(message) == "" {
		// Send help
		blocks := []slack.Block{
			slack.NewSectionBlock(
				slack.NewTextBlockObject("mrkdwn",
					"🔍 *Signal Translator*\nUsage: `/translate [message]`\n\nPaste a message you'd like translated into plain language.",
					false, false,
				),
				nil, nil,
			),
		}
		return t.slack.PostMessage(cmd.ChannelID, blocks, "Translation Help")
	}

	analysis, err := t.ai.AnalyzeTone(ctx, message)
	if err != nil {
		return fmt.Errorf("analyze tone: %w", err)
	}

	channelID, err := t.slack.OpenDMChannel(cmd.UserID)
	if err != nil {
		return err
	}

	return t.postTranslationBlocks(channelID, message, analysis)
}

// HandleBlockAction handles translator button clicks (e.g., "Got it").
func (t *TranslatorService) HandleBlockAction(ctx context.Context, action *slack.InteractionCallback, user *domain.User) error {
	if len(action.ActionCallback.BlockActions) == 0 {
		return nil
	}
	actionID := action.ActionCallback.BlockActions[0].ActionID
	if actionID == "translator_ack" {
		// Acknowledge the translation was helpful
		channelID, err := t.slack.OpenDMChannel(user.SlackUserID)
		if err != nil {
			return err
		}
		return t.slack.PostMessage(channelID,
			[]slack.Block{
				slack.NewSectionBlock(
					slack.NewTextBlockObject("mrkdwn", "👍 Glad I could help! Let me know if you need anything else translated.", false, false),
					nil, nil,
				),
			},
			"Translation Acknowledged",
		)
	}
	return nil
}

func (t *TranslatorService) containsAmbiguousPhrase(text string) bool {
	for _, pattern := range ambiguousPatterns {
		if pattern.MatchString(text) {
			return true
		}
	}
	return false
}

func (t *TranslatorService) extractMentionedUsers(text string) []string {
	matches := userMentionPattern.FindAllStringSubmatch(text, -1)
	var users []string
	seen := make(map[string]bool)
	for _, match := range matches {
		if len(match) > 1 && !seen[match[1]] {
			users = append(users, match[1])
			seen[match[1]] = true
		}
	}
	return users
}

func (t *TranslatorService) sendTranslationDM(ctx context.Context, mentionedUser string, event *slackevents.MessageEvent, analysis *domain.ToneAnalysis) error {
	channelID, err := t.slack.OpenDMChannel(mentionedUser)
	if err != nil {
		return fmt.Errorf("open dm: %w", err)
	}

	return t.postTranslationBlocks(channelID, event.Text, analysis)
}

func (t *TranslatorService) postTranslationBlocks(channelID, originalText string, analysis *domain.ToneAnalysis) error {
	tone := analysis.Tone
	if tone == "" {
		tone = "Neutral"
	}
	intent := analysis.Intent
	if intent == "" {
		intent = "Unable to determine"
	}
	action := analysis.Action
	if action == "" {
		action = "No specific action required"
	}
	note := analysis.Note
	if note == "" {
		note = "This message appears straightforward."
	}

	blocks := []slack.Block{
		slack.NewHeaderBlock(
			slack.NewTextBlockObject("plain_text", "🔍 Signal Translation", true, false),
		),
		slack.NewSectionBlock(
			slack.NewTextBlockObject("mrkdwn",
				fmt.Sprintf("*Original message:*\n> %s", originalText),
				false, false,
			),
			nil, nil,
		),
		slack.NewDividerBlock(),
		slack.NewSectionBlock(
			nil,
			[]*slack.TextBlockObject{
				slack.NewTextBlockObject("mrkdwn", fmt.Sprintf("*Tone:*\n%s", tone), false, false),
				slack.NewTextBlockObject("mrkdwn", fmt.Sprintf("*Intent:*\n%s", intent), false, false),
			},
			nil,
		),
		slack.NewSectionBlock(
			slack.NewTextBlockObject("mrkdwn",
				fmt.Sprintf("*Action:* %s\n\n*Note:* %s", action, note),
				false, false,
			),
			nil, nil,
		),
		slack.NewActionBlock("translator_actions",
			slack.NewButtonBlockElement("translator_ack", "ack", slack.NewTextBlockObject("plain_text", "Got it ✓", false, true)).WithStyle("primary"),
		),
	}

	return t.slack.PostMessage(channelID, blocks, "Signal Translation")
}
