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

// HandleMessage checks for ambiguous language and triggers auto-translation.
func (t *TranslatorService) HandleMessage(ctx context.Context, event *slackevents.MessageEvent, user *domain.User) error {
	if !t.containsAmbiguousPhrase(event.Text) {
		if event.ThreadTimeStamp == "" {
			return nil
		}
	}

	analysis, err := t.ai.AnalyzeTone(ctx, event.Text)
	if err != nil {
		slog.Error("ai tone analysis failed", "error", err)
		return nil
	}

	mentionedUsers := t.extractMentionedUsers(event.Text)
	if len(mentionedUsers) > 0 {
		for _, uid := range mentionedUsers {
			if err := t.sendTranslationDM(ctx, uid, event, analysis); err != nil {
				slog.Error("failed to send translation dm", "error", err, "user", uid)
			}
		}
	} else {
		if err := t.sendTranslationDM(ctx, user.SlackUserID, event, analysis); err != nil {
			slog.Error("failed to send translation dm to user", "error", err)
		}
	}
	return nil
}

// HandleSlashCommand handles the /translate command via response_url (no token needed).
func (t *TranslatorService) HandleSlashCommand(ctx context.Context, cmd *slack.SlashCommand, user *domain.User, responseURL string) error {
	message := strings.TrimSpace(cmd.Text)
	if message == "" {
		return t.slack.PostWebhook(responseURL, []slack.Block{
			slack.NewSectionBlock(
				slack.NewTextBlockObject("mrkdwn",
					"🔍 *Signal Translator*\n\nUsage: `/translate [message]`\n\nPaste any ambiguous workplace message and I'll decode the tone, intent, and what you should do.\n\n*Example:* `/translate Per my last email, we need this by EOD.`",
					false, false,
				),
				nil, nil,
			),
		}, "Translation Help")
	}

	// Immediately acknowledge with a processing message
	_ = t.slack.PostWebhook(responseURL, []slack.Block{
		slack.NewSectionBlock(
			slack.NewTextBlockObject("mrkdwn", "🔍 Analyzing tone...", false, false),
			nil, nil,
		),
	}, "Analyzing...")

	analysis, err := t.ai.AnalyzeTone(ctx, message)
	if err != nil {
		slog.Error("ai tone analysis failed for slash command", "error", err)
		return t.slack.PostWebhook(responseURL, []slack.Block{
			slack.NewSectionBlock(
				slack.NewTextBlockObject("mrkdwn", "❌ Couldn't analyze that right now. Please try again in a moment.", false, false),
				nil, nil,
			),
		}, "Translation Error")
	}

	return t.slack.PostWebhook(responseURL, t.buildTranslationBlocks(message, analysis), "Signal Translation")
}

// HandleBlockAction handles translator button clicks (e.g., "Got it").
func (t *TranslatorService) HandleBlockAction(ctx context.Context, action *slack.InteractionCallback, user *domain.User) error {
	if len(action.ActionCallback.BlockActions) == 0 {
		return nil
	}
	if action.ActionCallback.BlockActions[0].ActionID == "translator_ack" {
		// Reply in the same channel/DM where the button was clicked
		return t.slack.PostMessage(action.Channel.ID, []slack.Block{
			slack.NewSectionBlock(
				slack.NewTextBlockObject("mrkdwn", "👍 Glad I could help! Feel free to translate another message anytime.", false, false),
				nil, nil,
			),
		}, "Translation Acknowledged")
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

func (t *TranslatorService) sendTranslationDM(ctx context.Context, userID string, event *slackevents.MessageEvent, analysis *domain.ToneAnalysis) error {
	channelID, err := t.slack.OpenDMChannel(userID)
	if err != nil {
		return fmt.Errorf("open dm: %w", err)
	}
	return t.slack.PostMessage(channelID, t.buildTranslationBlocks(event.Text, analysis), "Signal Translation")
}

func (t *TranslatorService) buildTranslationBlocks(originalText string, analysis *domain.ToneAnalysis) []slack.Block {
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

	escapedText := strings.ReplaceAll(originalText, ">", "\\>")

	return []slack.Block{
		slack.NewHeaderBlock(
			slack.NewTextBlockObject("plain_text", "🔍 Signal Translation", true, false),
		),
		slack.NewSectionBlock(
			slack.NewTextBlockObject("mrkdwn",
				fmt.Sprintf("*Original message:*\n> %s", escapedText),
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
			slack.NewButtonBlockElement("translator_ack", "ack",
				slack.NewTextBlockObject("plain_text", "Got it ✓", true, false),
			).WithStyle(slack.StylePrimary),
		),
	}
}
