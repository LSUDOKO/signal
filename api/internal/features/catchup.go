package features

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/LSUDOKOS/signal/internal/ai"
	"github.com/LSUDOKOS/signal/internal/domain"
	"github.com/LSUDOKOS/signal/internal/rts"
	"github.com/slack-go/slack"
)

// CatchUpService implements the Catch-Up semantic search feature.
type CatchUpService struct {
	slack    SlackAPI
	ai       *ai.Client
	searcher *rts.Searcher
}

// NewCatchUpService creates a new Catch-Up service.
func NewCatchUpService(slack SlackAPI, ai *ai.Client, searcher *rts.Searcher) *CatchUpService {
	return &CatchUpService{slack: slack, ai: ai, searcher: searcher}
}

// HandleSlashCommand processes the /catchup command.
func (c *CatchUpService) HandleSlashCommand(ctx context.Context, cmd *slack.SlashCommand, user *domain.User) error {
	query := strings.TrimSpace(cmd.Text)
	if query == "" {
		// Send help to DM
		dmChannel, err := c.slack.OpenDMChannel(cmd.UserID)
		if err != nil {
			slog.Error("failed to open dm for catchup help", "error", err)
			return err
		}
		blocks := []slack.Block{
			slack.NewSectionBlock(
				slack.NewTextBlockObject("mrkdwn",
					"🔍 *Signal Catch-Up*\nUsage: `/catchup [what you missed]`\n\nExamples:\n• `/catchup What did we decide about the budget?`\n• `/catchup Any updates on the design system?`\n• `/catchup What happened in engineering this week?`",
					false, false,
				),
				nil, nil,
			),
		}
		return c.slack.PostMessage(dmChannel, blocks, "Catch-Up Help")
	}

	// Perform semantic search (might take 2-3 seconds)
	result, err := c.searchAndSummarize(ctx, cmd.UserID, query, 7)
	if err != nil {
		slog.Error("catchup search failed", "error", err, "query", query)
		// Send error to user's DM
		dmChannel, dmErr := c.slack.OpenDMChannel(cmd.UserID)
		if dmErr != nil {
			return fmt.Errorf("catchup search: %w, open dm: %w", err, dmErr)
		}
		return c.slack.PostMessage(dmChannel,
			[]slack.Block{
				slack.NewSectionBlock(
					slack.NewTextBlockObject("mrkdwn", "❌ Sorry, I couldn't search for that right now. Please try again in a moment.", false, false),
					nil, nil,
				),
			},
			"Catch-Up Error",
		)
	}

	// Post result to DM
	dmChannel, err := c.slack.OpenDMChannel(cmd.UserID)
	if err != nil {
		return fmt.Errorf("open dm: %w", err)
	}

	if result == nil || result.MessageCount == 0 {
		return c.slack.PostMessage(dmChannel,
			[]slack.Block{
				slack.NewSectionBlock(
					slack.NewTextBlockObject("mrkdwn",
						fmt.Sprintf("I couldn't find anything about *%s* in your accessible channels. Try rephrasing or expanding the date range?", query),
						false, false,
					),
					nil, nil,
				),
			},
			"No Results",
		)
	}

	return c.slack.PostMessage(dmChannel, c.buildResultBlocks(query, result, ""), "Catch-Up Results")
}

// searchAndSummarize performs a Slack search and AI summarization.
func (c *CatchUpService) searchAndSummarize(ctx context.Context, userID, query string, daysBack int) (*domain.CatchUpResult, error) {
	// Use the RTS client for semantic search
	result, err := c.searcher.SemanticCatchup(ctx, userID, query, daysBack)
	if err != nil {
		return nil, fmt.Errorf("semantic search: %w", err)
	}

	if result.TotalCount == 0 {
		return &domain.CatchUpResult{MessageCount: 0}, nil
	}

	// Extract message text for AI summarization
	var messageTexts []string
	for _, msg := range result.Messages {
		if msg.Text != "" {
			messageTexts = append(messageTexts, msg.Text)
		}
	}

	if len(messageTexts) == 0 {
		return &domain.CatchUpResult{MessageCount: 0}, nil
	}

	// Generate AI summary
	summary, err := c.ai.CatchUpSummary(ctx, messageTexts)
	if err != nil {
		return nil, fmt.Errorf("ai summary: %w", err)
	}

	return &domain.CatchUpResult{
		Topics: []domain.CatchUpTopic{
			{
				Name:    "Search Results",
				Decision: summary,
				Action:  "Review the full context in Slack",
				Context: fmt.Sprintf("Found %d relevant messages from the last %d days.", len(messageTexts), daysBack),
			},
		},
		MessageCount: len(messageTexts),
	}, nil
}

func (c *CatchUpService) buildResultBlocks(query string, result *domain.CatchUpResult, _ string) []slack.Block {
	var blocks []slack.Block

	blocks = append(blocks,
		slack.NewHeaderBlock(
			slack.NewTextBlockObject("plain_text", "📋 What You Missed", true, false),
		),
		slack.NewSectionBlock(
			slack.NewTextBlockObject("mrkdwn",
				fmt.Sprintf("*Query:* \"%s\"\n*Messages found:* %d", query, result.MessageCount),
				false, false,
			),
			nil, nil,
		),
		slack.NewDividerBlock(),
	)

	summary := "No structured summary available."
	if len(result.Topics) > 0 {
		summary = result.Topics[0].Decision
	}

	blocks = append(blocks,
		slack.NewSectionBlock(
			slack.NewTextBlockObject("mrkdwn", summary, false, false),
			nil, nil,
		),
	)

	return blocks
}

// SemanticQuery parses a natural language query into a Slack search query.
func (c *CatchUpService) SemanticQuery(userID, naturalQuery string, daysBack int) string {
	dateFilter := time.Now().AddDate(0, 0, -daysBack).Format("2006-01-02")
	return fmt.Sprintf("from:@%s OR to:@%s %s after:%s",
		userID, userID, naturalQuery, dateFilter,
	)
}
