package rts

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/slack-go/slack"
)

// Searcher performs semantic search via Slack's Real-Time Search API.
type Searcher struct {
	client *slack.Client
}

// NewSearcher creates a new RTS search client.
func NewSearcher(client *slack.Client) *Searcher {
	return &Searcher{
		client: client,
	}
}

// SearchResult holds the results of a semantic search.
type SearchResult struct {
	Messages    []SearchMessage `json:"messages"`
	TotalCount  int             `json:"total_count"`
}

// SearchMessage represents a single search result message.
type SearchMessage struct {
	Text      string `json:"text"`
	User      string `json:"user"`
	ChannelID string `json:"channel_id"`
	Timestamp string `json:"ts"`
	Permalink string `json:"permalink"`
}

// SemanticCatchup performs a natural-language search across accessible channels.
func (s *Searcher) SemanticCatchup(ctx context.Context, userID, query string, daysBack int) (*SearchResult, error) {
	searchQuery := BuildSearchQuery(userID, query, daysBack)

	slog.Debug("rts search",
		"user", userID,
		"query", searchQuery,
		"days_back", daysBack,
	)

	params := slack.SearchParameters{
		Sort:  "timestamp",
		Count: 20,
	}

	results, err := s.client.SearchMessages(searchQuery, params)
	if err != nil {
		return nil, fmt.Errorf("rts search: %w", err)
	}

	result := &SearchResult{
		TotalCount: results.Total,
	}

	for _, match := range results.Matches {
		msg := SearchMessage{
			Text:      match.Text,
			User:      match.Username,
			ChannelID: match.Channel.ID,
			Timestamp: match.Timestamp,
		}
		result.Messages = append(result.Messages, msg)
	}

	slog.Debug("rts search complete",
		"total", result.TotalCount,
		"returned", len(result.Messages),
	)

	return result, nil
}

// BuildSearchQuery constructs a Slack search query from natural language.
func BuildSearchQuery(userID, query string, daysBack int) string {
	dateFilter := time.Now().AddDate(0, 0, -daysBack).Format("2006-01-02")
	return fmt.Sprintf(
		"from:@%s OR to:@%s %s after:%s",
		userID, userID, query, dateFilter,
	)
}
