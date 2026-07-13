package features

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/LSUDOKOS/signal/internal/ai"
	"github.com/LSUDOKOS/signal/internal/domain"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

// ThreadSummaryService summarizes Slack threads when a user adds a 📋 reaction.
// It sends the summary as a DM to the reacting user.
type ThreadSummaryService struct {
	slack  SlackAPI
	ai     *ai.Client
	memory *MemoryService
}

// SummaryReactionEmoji is the emoji that triggers thread summarization.
const SummaryReactionEmoji = "memo"    // 📋 emoji name in Slack

// NewThreadSummaryService creates a new ThreadSummaryService.
func NewThreadSummaryService(slack SlackAPI, ai *ai.Client, memory *MemoryService) *ThreadSummaryService {
	return &ThreadSummaryService{slack: slack, ai: ai, memory: memory}
}

// HandleReaction handles a reaction_added event.
// If the reaction is 📋 (memo), it summarizes the thread and DMs the user.
func (t *ThreadSummaryService) HandleReaction(ctx context.Context, event *slackevents.ReactionAddedEvent, user *domain.User) error {
	if event.Reaction != SummaryReactionEmoji {
		return nil
	}

	// Only works on channel messages (not files or other items)
	if event.Item.Type != "message" {
		return nil
	}

	channelID := event.Item.Channel
	messageTS := event.Item.Timestamp

	// Fetch the thread (the original message + all replies)
	history, err := t.slack.GetThreadMessages(channelID, messageTS)
	if err != nil {
		slog.Error("thread summary: failed to get thread", "error", err, "channel", channelID, "ts", messageTS)
		return t.slack.PostMessage(channelID, []slack.Block{
			slack.NewSectionBlock(
				slack.NewTextBlockObject("mrkdwn",
					fmt.Sprintf("<@%s> ❌ I couldn't read that thread. Make sure I'm a member of this channel.", user.SlackUserID),
					false, false,
				),
				nil, nil,
			),
		}, "Thread Summary Error")
	}

	if len(history) == 0 {
		return nil
	}

	// Extract message texts
	var texts []string
	for _, msg := range history {
		if msg.BotID != "" || msg.Text == "" {
			continue
		}
		texts = append(texts, fmt.Sprintf("[%s]: %s", msg.User, msg.Text))
	}

	if len(texts) == 0 {
		return nil
	}

	// Generate summary
	summary, err := t.ai.SummarizeThread(ctx, texts, user.Neurotype)
	if err != nil {
		slog.Error("thread summary: AI failed", "error", err)
		return nil
	}

	// Save to memory
	t.memory.AddInteraction(ctx, user.SlackUserID, fmt.Sprintf("summarized thread in <#%s>", channelID))

	// Post summary as ephemeral (only visible to the user who reacted) in the channel
	blocks := []slack.Block{
		slack.NewHeaderBlock(
			slack.NewTextBlockObject("plain_text", "📋 Thread Summary", true, false),
		),
		slack.NewSectionBlock(
			slack.NewTextBlockObject("mrkdwn",
				fmt.Sprintf("*Thread in <#%s>* — %d messages", channelID, len(texts)),
				false, false,
			),
			nil, nil,
		),
		slack.NewDividerBlock(),
		slack.NewSectionBlock(
			slack.NewTextBlockObject("mrkdwn", summary, false, false),
			nil, nil,
		),
		slack.NewContextBlock("thread_summary_footer",
			slack.NewTextBlockObject("mrkdwn", "_React with 📋 on any message to get a thread summary_", false, false),
		),
	}

	// Post as ephemeral in the channel so only the reactor sees it
	return t.slack.PostEphemeral(channelID, user.SlackUserID, blocks, "Thread Summary")
}
