package ai

import (
	"context"
	"fmt"
	"strings"

	"github.com/LSUDOKOS/signal/internal/domain"
	"github.com/sashabaranov/go-openai"
)

// Client wraps the OpenAI API for Signal-specific AI operations.
type Client struct {
	client *openai.Client
	model  string
}

// NewClient creates a new AI client.
// baseURL should be the full API base URL (e.g. "https://api.openai.com/v1" for OpenAI,
// or "https://api.x.ai/v1" for Grok by xAI, which uses an OpenAI-compatible API).
func NewClient(apiKey, model, baseURL string) *Client {
	config := openai.DefaultConfig(apiKey)
	config.BaseURL = baseURL
	return &Client{
		client: openai.NewClientWithConfig(config),
		model:  model,
	}
}

// SummarizeFocus generates a decision tree summary from channel messages.
func (c *Client) SummarizeFocus(ctx context.Context, messages []string) (*domain.FocusSummaryResult, error) {
	prompt := buildFocusPrompt(messages)

	resp, err := c.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: c.model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: "You are a workplace communication summarizer for neurodivergent professionals. Extract ONLY decisions and action items. Ignore greetings, small talk, emojis, reactions, and off-topic jokes. Use plain language. No corporate buzzwords.",
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
		Temperature: 0.3,
		MaxTokens:   1000,
	})
	if err != nil {
		return nil, fmt.Errorf("openai focus summary: %w", err)
	}

	content, err := extractContent(resp)
	if err != nil {
		return nil, fmt.Errorf("focus summary: %w", err)
	}
	return parseFocusResponse(content), nil
}

// AnalyzeTone performs tone analysis on an ambiguous workplace message.
func (c *Client) AnalyzeTone(ctx context.Context, message string) (*domain.ToneAnalysis, error) {
	resp, err := c.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: c.model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: "You are a direct, kind translator of workplace subtext for autistic and ADHD adults. You decode ambiguous messages into literal meaning without judgment. Respond in this exact format:\n- Tone: [single word or short phrase]\n- Intent: [1 sentence, what they want]\n- Action: [1 sentence, what you should do]\n- Note: [1-2 sentences explaining any hidden social context]\n\nRules:\n- Never say 'they might be' — be confident but not rude.\n- Never use 'just' or 'simply' — these are patronizing.\n- If the message is genuinely neutral, say so clearly.\n- If the message is passive-aggressive, name it directly but kindly.",
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: fmt.Sprintf("Analyze this message: %q", message),
			},
		},
		Temperature: 0.3,
		MaxTokens:   500,
	})
	if err != nil {
		return nil, fmt.Errorf("openai tone analysis: %w", err)
	}

	content, err := extractContent(resp)
	if err != nil {
		return nil, fmt.Errorf("tone analysis: %w", err)
	}
	return parseToneResponse(content), nil
}

// CatchUpSummary generates a "What You Missed" digest from search results.
func (c *Client) CatchUpSummary(ctx context.Context, messages []string) (string, error) {
	prompt := buildCatchUpPrompt(messages)

	resp, err := c.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: c.model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: "You are a 'What You Missed' assistant for a neurodivergent professional. Summarize these messages into topics. Highlight decisions, action items, and anything requiring the user's input.",
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
		Temperature: 0.3,
		MaxTokens:   1500,
	})
	if err != nil {
		return "", fmt.Errorf("openai catchup summary: %w", err)
	}

	return extractContent(resp)
}

// GenerateDigestContent categorizes messages into urgent/action/FYI buckets.
func (c *Client) GenerateDigestContent(ctx context.Context, messages []string) (string, error) {
	prompt := fmt.Sprintf(`Categorize these Slack messages into:
🔴 URGENT (needs response today — contains words like "urgent", "asap", "deadline", "today")
🟡 ACTION REQUIRED (this week)
🟢 FYI (no action needed)

Messages:
%s

Format each category as a bullet list. If empty, say "None."`, strings.Join(messages, "\n"))

	resp, err := c.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: c.model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: "You organize Slack messages by urgency for an ADHD professional. Be concise.",
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
		Temperature: 0.3,
		MaxTokens:   1000,
	})
	if err != nil {
		return "", fmt.Errorf("openai digest: %w", err)
	}

	return extractContent(resp)
}

func buildFocusPrompt(messages []string) string {
	return fmt.Sprintf(`Extract a decision tree from this Slack channel history.
Only include: decisions made, action items with owners, and deadlines.
Ignore small talk, greetings, emojis, and reactions.

Format strictly as:
✅ [Decision made]
   ↳ [Action item] — Owner: @Name — Due: [Date or "None"]
   ↳ [Action item] — Owner: @Name — Due: [Date or "None"]

If no decisions were made, say: "No decisions found. Discussion was exploratory."

History:
%s`, strings.Join(messages, "\n"))
}

func buildCatchUpPrompt(messages []string) string {
	return fmt.Sprintf(`Summarize these Slack messages into a "What You Missed" digest.
Organize by topic. Highlight decisions, action items, and anything requiring the user's input.

For each topic:
## [Topic Name]
- **Decision:** [What was decided]
- **Context:** [2-sentence background]
- **Your Action:** [What you need to do, or "None"]
- **Link:** [Jump to message]

Messages:
%s`, strings.Join(messages, "\n"))
}

func parseFocusResponse(content string) *domain.FocusSummaryResult {
	if strings.Contains(content, "No decisions found") {
		return &domain.FocusSummaryResult{NoDecisions: true}
	}

	result := &domain.FocusSummaryResult{}
	lines := strings.Split(content, "\n")

	var currentDecision *domain.DecisionItem
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "✅") {
			if currentDecision != nil {
				result.Decisions = append(result.Decisions, *currentDecision)
			}
			currentDecision = &domain.DecisionItem{
				Decision: strings.TrimSpace(strings.TrimPrefix(line, "✅")),
			}
		} else if strings.HasPrefix(line, "↳") && currentDecision != nil {
			item := strings.TrimSpace(strings.TrimPrefix(line, "↳"))
			parts := strings.Split(item, "—")
			actionItem := domain.ActionItem{}
			for _, part := range parts {
				part = strings.TrimSpace(part)
				if strings.HasPrefix(part, "Owner:") {
					actionItem.Owner = strings.TrimPrefix(part, "Owner:")
				} else if strings.HasPrefix(part, "Due:") {
					actionItem.Due = strings.TrimPrefix(part, "Due:")
				} else if actionItem.Description == "" {
					actionItem.Description = part
				}
			}
			currentDecision.ActionItems = append(currentDecision.ActionItems, actionItem)
		}
	}

	if currentDecision != nil {
		result.Decisions = append(result.Decisions, *currentDecision)
	}

	return result
}

func parseToneResponse(content string) *domain.ToneAnalysis {
	analysis := &domain.ToneAnalysis{}
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(line, "- Tone:") || strings.HasPrefix(line, "- Tone :"):
			analysis.Tone = strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(line, "- Tone:"), "- Tone :"))
		case strings.HasPrefix(line, "- Intent:") || strings.HasPrefix(line, "- Intent :"):
			analysis.Intent = strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(line, "- Intent:"), "- Intent :"))
		case strings.HasPrefix(line, "- Action:") || strings.HasPrefix(line, "- Action :"):
			analysis.Action = strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(line, "- Action:"), "- Action :"))
		case strings.HasPrefix(line, "- Note:") || strings.HasPrefix(line, "- Note :"):
			analysis.Note = strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(line, "- Note:"), "- Note :"))
		}
	}

	return analysis
}

// Chat is a generic chat completion method for flexibility.
func (c *Client) Chat(ctx context.Context, systemPrompt, userPrompt string, maxTokens int) (string, error) {
	resp, err := c.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: c.model,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: systemPrompt},
			{Role: openai.ChatMessageRoleUser, Content: userPrompt},
		},
		Temperature: 0.3,
		MaxTokens:   maxTokens,
	})
	if err != nil {
		return "", fmt.Errorf("openai chat: %w", err)
	}
	return extractContent(resp)
}

// RawClient returns the underlying OpenAI client for use by the HTTP API layer.
func (c *Client) RawClient() *openai.Client {
	return c.client
}

// extractContent safely extracts the first choice's content, guarding against empty choices.
func extractContent(resp openai.ChatCompletionResponse) (string, error) {
	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no choices returned by AI model")
	}
	return resp.Choices[0].Message.Content, nil
}


