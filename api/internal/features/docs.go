package features

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/LSUDOKOS/signal/internal/ai"
	"github.com/LSUDOKOS/signal/internal/domain"
	"github.com/slack-go/slack"
)

// DocsService handles /docs command — search Notion workspace via real Notion API.
type DocsService struct {
	slack        SlackAPI
	ai           *ai.Client
	memory       *MemoryService
	httpClient   *http.Client
	notionToken  string // Notion Internal Integration Token
}

// NewDocsService creates a new DocsService.
// notionToken: Notion internal integration token (starts with "secret_")
func NewDocsService(slack SlackAPI, ai *ai.Client, memory *MemoryService, notionToken string) *DocsService {
	return &DocsService{
		slack:       slack,
		ai:          ai,
		memory:      memory,
		httpClient:  &http.Client{Timeout: 15 * time.Second},
		notionToken: notionToken,
	}
}

// IsConfigured returns true if Notion token is set.
func (d *DocsService) IsConfigured() bool {
	return d.notionToken != ""
}

// notionSearchResult is a Notion search result page/database.
type notionSearchResult struct {
	Object         string `json:"object"`
	ID             string `json:"id"`
	URL            string `json:"url"`
	Properties     map[string]json.RawMessage `json:"properties"`
	Title          []struct {
		PlainText string `json:"plain_text"`
	} `json:"title,omitempty"`
	LastEditedTime string `json:"last_edited_time"`
	Parent         struct {
		Type string `json:"type"`
	} `json:"parent"`
}

// HandleSlashCommand handles /docs [query].
// Examples:
//   /docs onboarding guide
//   /docs API documentation
//   /docs how to deploy
func (d *DocsService) HandleSlashCommand(ctx context.Context, cmd *slack.SlashCommand, user *domain.User, responseURL string) error {
	if !d.IsConfigured() {
		return d.slack.PostWebhook(responseURL, []slack.Block{
			slack.NewSectionBlock(
				slack.NewTextBlockObject("mrkdwn",
					"⚙️ *Docs MCP not configured*\n\nAsk your admin to add `NOTION_TOKEN` to Signal's configuration.\n\n*How to get a Notion token:*\n1. Go to https://www.notion.so/my-integrations\n2. Create a new integration\n3. Copy the Internal Integration Token\n4. Share your pages/databases with the integration\n\nOnce configured:\n• `/docs [search query]` — search your Notion workspace",
					false, false,
				),
				nil, nil,
			),
		}, "Docs Not Configured")
	}

	query := strings.TrimSpace(cmd.Text)
	if query == "" {
		return d.slack.PostWebhook(responseURL, []slack.Block{
			slack.NewSectionBlock(
				slack.NewTextBlockObject("mrkdwn",
					"📚 *Signal Docs MCP*\n\nUsage: `/docs [search query]`\n\n*Examples:*\n• `/docs onboarding guide`\n• `/docs API documentation`\n• `/docs how to deploy to production`\n• `/docs Q3 roadmap`",
					false, false,
				),
				nil, nil,
			),
		}, "Docs Help")
	}

	// Search Notion
	results, err := d.searchNotion(ctx, query)
	if err != nil {
		slog.Error("docs: notion search failed", "error", err)
		return d.slack.PostWebhook(responseURL, []slack.Block{
			slack.NewSectionBlock(
				slack.NewTextBlockObject("mrkdwn",
					"❌ Notion search failed. Check that the token is valid and pages are shared with the integration.",
					false, false,
				),
				nil, nil,
			),
		}, "Docs Error")
	}

	if len(results) == 0 {
		return d.slack.PostWebhook(responseURL, []slack.Block{
			slack.NewSectionBlock(
				slack.NewTextBlockObject("mrkdwn",
					fmt.Sprintf("📚 No Notion pages found for *\"%s\"*.\n\nMake sure:\n• The page is shared with your Signal integration\n• The query matches the page title or content", query),
					false, false,
				),
				nil, nil,
			),
		}, "No Results")
	}

	// Save to memory
	d.memory.AddInteraction(ctx, user.SlackUserID, fmt.Sprintf("searched docs: %s", query))

	blocks := []slack.Block{
		slack.NewHeaderBlock(
			slack.NewTextBlockObject("plain_text", fmt.Sprintf("📚 Docs: \"%s\" (%d results)", query, len(results)), true, false),
		),
	}

	for i, result := range results {
		if i >= 6 {
			break
		}
		title := d.extractTitle(result)
		if title == "" {
			title = "Untitled"
		}
		lastEdited := formatNotionDate(result.LastEditedTime)
		objType := result.Object
		if objType == "page" {
			objType = "📄 Page"
		} else if objType == "database" {
			objType = "🗄️ Database"
		}

		blocks = append(blocks, slack.NewSectionBlock(
			slack.NewTextBlockObject("mrkdwn",
				fmt.Sprintf("*<%s|%s>*\n%s  •  Last edited: %s", result.URL, title, objType, lastEdited),
				false, false,
			),
			nil, nil,
		))
	}

	if len(results) > 6 {
		blocks = append(blocks, slack.NewContextBlock("docs_more",
			slack.NewTextBlockObject("mrkdwn", fmt.Sprintf("_%d more results. Refine your search to narrow down._", len(results)-6), false, false),
		))
	}

	blocks = append(blocks, slack.NewContextBlock("docs_footer",
		slack.NewTextBlockObject("mrkdwn", "_Results from your Notion workspace. Use `/docs [query]` to search again._", false, false),
	))

	return d.slack.PostWebhook(responseURL, blocks, "Docs Search Results")
}

// searchNotion calls the Notion Search API.
func (d *DocsService) searchNotion(ctx context.Context, query string) ([]notionSearchResult, error) {
	payload := fmt.Sprintf(`{"query":%s,"page_size":10}`, jsonString(query))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://api.notion.com/v1/search",
		strings.NewReader(payload),
	)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+d.notionToken)
	req.Header.Set("Notion-Version", "2022-06-28")
	req.Header.Set("Content-Type", "application/json")

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("notion api: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, fmt.Errorf("invalid Notion token (401)")
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("notion api status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Results []notionSearchResult `json:"results"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return result.Results, nil
}

func (d *DocsService) extractTitle(r notionSearchResult) string {
	// Try title array (databases and pages have different structures)
	if len(r.Title) > 0 {
		return r.Title[0].PlainText
	}
	// Try properties.Name or properties.title
	for key, raw := range r.Properties {
		if strings.EqualFold(key, "name") || strings.EqualFold(key, "title") {
			var titleProp struct {
				Title []struct {
					PlainText string `json:"plain_text"`
				} `json:"title"`
			}
			if err := json.Unmarshal(raw, &titleProp); err == nil && len(titleProp.Title) > 0 {
				return titleProp.Title[0].PlainText
			}
		}
	}
	return ""
}

func formatNotionDate(s string) string {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return s
	}
	return t.Format("Jan 2, 2006")
}

func jsonString(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}

// GetNotionPageContent fetches the text content of a Notion page block tree.
// Used for AI-powered summarization of a specific page.
func (d *DocsService) GetNotionPageContent(ctx context.Context, pageID string) (string, error) {
	apiURL := fmt.Sprintf("https://api.notion.com/v1/blocks/%s/children?page_size=50", url.PathEscape(pageID))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+d.notionToken)
	req.Header.Set("Notion-Version", "2022-06-28")

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("notion blocks api: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("notion blocks status %d", resp.StatusCode)
	}

	var result struct {
		Results []struct {
			Type      string          `json:"type"`
			Paragraph *struct {
				RichText []struct {
					PlainText string `json:"plain_text"`
				} `json:"rich_text"`
			} `json:"paragraph,omitempty"`
			Heading1 *struct {
				RichText []struct {
					PlainText string `json:"plain_text"`
				} `json:"rich_text"`
			} `json:"heading_1,omitempty"`
			BulletedListItem *struct {
				RichText []struct {
					PlainText string `json:"plain_text"`
				} `json:"rich_text"`
			} `json:"bulleted_list_item,omitempty"`
		} `json:"results"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	var lines []string
	for _, block := range result.Results {
		switch block.Type {
		case "paragraph":
			if block.Paragraph != nil {
				for _, rt := range block.Paragraph.RichText {
					if rt.PlainText != "" {
						lines = append(lines, rt.PlainText)
					}
				}
			}
		case "heading_1":
			if block.Heading1 != nil {
				for _, rt := range block.Heading1.RichText {
					if rt.PlainText != "" {
						lines = append(lines, "# "+rt.PlainText)
					}
				}
			}
		case "bulleted_list_item":
			if block.BulletedListItem != nil {
				for _, rt := range block.BulletedListItem.RichText {
					if rt.PlainText != "" {
						lines = append(lines, "• "+rt.PlainText)
					}
				}
			}
		}
	}

	return strings.Join(lines, "\n"), nil
}
