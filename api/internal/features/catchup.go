package features

import (
	"context"
	"fmt"
	"log/slog"
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

// HandleSlashCommand processes the /catchup command via response_url.
func (c *CatchUpService) HandleSlashCommand(ctx context.Context, cmd *slack.SlashCommand, user *domain.User, responseURL string) error {
	query := strings.TrimSpace(cmd.Text)
	if query == "" {
		return c.slack.PostWebhook(responseURL, []slack.Block{
			slack.NewSectionBlock(
				slack.NewTextBlockObject("mrkdwn",
					"📋 *Signal Catch-Up*\n\nUsage: `/catchup [what you missed]`\n\n*Examples:*\n• `/catchup What did we decide about the budget?`\n• `/catchup Any updates on the design system?`\n• `/catchup What happened in engineering this week?`",
					false, false,
				),
				nil, nil,
			),
		}, "Catch-Up Help")
	}

	// Immediately acknowledge
	_ = c.slack.PostWebhook(responseURL, []slack.Block{
		slack.NewSectionBlock(
			slack.NewTextBlockObject("mrkdwn", fmt.Sprintf("🔍 Searching for *\"%s\"*...", query), false, false),
			nil, nil,
		),
	}, "Searching...")

	result, err := c.searchAndSummarize(ctx, cmd.UserID, query, 7)
	if err != nil {
		slog.Error("catchup search failed", "error", err, "query", query)
		return c.slack.PostWebhook(responseURL, []slack.Block{
			slack.NewSectionBlock(
				slack.NewTextBlockObject("mrkdwn", "❌ Search failed. Please try again in a moment.", false, false),
				nil, nil,
			),
		}, "Catch-Up Error")
	}

	if result == nil || result.MessageCount == 0 {
		return c.slack.PostWebhook(responseURL, []slack.Block{
			slack.NewSectionBlock(
				slack.NewTextBlockObject("mrkdwn",
					fmt.Sprintf("🔍 I couldn't find anything about *\"%s\"* in your accessible channels.\n\nTry rephrasing or expanding the date range.", query),
					false, false,
				),
				nil, nil,
			),
		}, "No Results")
	}

	return c.slack.PostWebhook(responseURL, c.buildResultBlocks(query, result, ""), "Catch-Up Results")
}

// searchAndSummarize performs a Slack search and AI summarization.
// Falls back to searching accessible channels via history if RTS fails.
func (c *CatchUpService) searchAndSummarize(ctx context.Context, userID, query string, daysBack int) (*domain.CatchUpResult, error) {
	// Try RTS semantic search first
	result, err := c.searcher.SemanticCatchup(ctx, userID, query, daysBack)
	if err != nil {
		// RTS failed (missing_scope or other) — fall back to AI-only answer
		slog.Warn("rts search failed, falling back to AI-only response", "error", err)
		return c.aiOnlySummary(ctx, query)
	}

	if result.TotalCount == 0 {
		return &domain.CatchUpResult{MessageCount: 0}, nil
	}

	var messageTexts []string
	for _, msg := range result.Messages {
		if msg.Text != "" {
			messageTexts = append(messageTexts, msg.Text)
		}
	}
	if len(messageTexts) == 0 {
		return &domain.CatchUpResult{MessageCount: 0}, nil
	}

	summary, err := c.ai.CatchUpSummary(ctx, messageTexts)
	if err != nil {
		return nil, fmt.Errorf("ai summary: %w", err)
	}

	return &domain.CatchUpResult{
		Topics: []domain.CatchUpTopic{
			{
				Name:     "Search Results",
				Decision: summary,
				Action:   "Review the full context in Slack",
				Context:  fmt.Sprintf("Found %d relevant messages from the last %d days.", len(messageTexts), daysBack),
			},
		},
		MessageCount: len(messageTexts),
	}, nil
}

// aiOnlySummary uses the AI to answer the query without search results.
// Used as fallback when RTS search scope is unavailable.
func (c *CatchUpService) aiOnlySummary(ctx context.Context, query string) (*domain.CatchUpResult, error) {
	// Ask AI to provide a helpful response about what to look for
	prompt := fmt.Sprintf("A user wants to catch up on Slack and asked: %q\n\nSearch is unavailable. Give them a helpful 2-3 sentence response about what keywords or channels they might check manually, and what kind of information they're likely looking for.", query)
	summary, err := c.ai.CatchUpSummary(ctx, []string{prompt})
	if err != nil {
		return nil, fmt.Errorf("ai fallback: %w", err)
	}
	return &domain.CatchUpResult{
		Topics: []domain.CatchUpTopic{
			{
				Name:     "AI Guidance",
				Decision: summary,
				Action:   "Check the channels or keywords mentioned above",
				Context:  "Note: Full Slack search is not available. This is AI guidance based on your query.",
			},
		},
		MessageCount: 1, // Non-zero so the result gets displayed
	}, nil
}

func (c *CatchUpService) buildResultBlocks(query string, result *domain.CatchUpResult, _ string) []slack.Block {
	var blocks []slack.Block

	blocks = append(blocks,
		slack.NewHeaderBlock(
			slack.NewTextBlockObject("plain_text", "📋 What You Missed", true, false),
		),
	)

	summary := "No structured summary available."
	note := ""
	if len(result.Topics) > 0 {
		summary = result.Topics[0].Decision
		note = result.Topics[0].Context
	}

	queryLine := fmt.Sprintf("*Your query:* \"%s\"", query)
	if result.MessageCount > 1 {
		queryLine += fmt.Sprintf(" — *%d messages found*", result.MessageCount)
	}

	blocks = append(blocks,
		slack.NewSectionBlock(
			slack.NewTextBlockObject("mrkdwn", queryLine, false, false),
			nil, nil,
		),
		slack.NewDividerBlock(),
		slack.NewSectionBlock(
			slack.NewTextBlockObject("mrkdwn", summary, false, false),
			nil, nil,
		),
	)

	if note != "" {
		blocks = append(blocks,
			slack.NewContextBlock("catchup_note",
				slack.NewTextBlockObject("mrkdwn", "_"+note+"_", false, false),
			),
		)
	}

	return blocks
}

// SemanticQuery parses a natural language query into a Slack search query.
func (c *CatchUpService) SemanticQuery(userID, naturalQuery string, daysBack int) string {
	dateFilter := time.Now().AddDate(0, 0, -daysBack).Format("2006-01-02")
	return fmt.Sprintf("from:@%s OR to:@%s %s after:%s",
		userID, userID, naturalQuery, dateFilter,
	)
}
