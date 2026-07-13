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
	return fmt.Sprintf(`Summarize these Slack messages into a plain "What You Missed" digest.

IMPORTANT: Use ONLY Slack mrkdwn formatting:
- Use *bold* for topic names (NOT ## markdown headers)
- Use bullet points with •
- NO markdown ## headers
- NO [link](#anchor) style links
- Keep each topic to 3-4 lines max

Format each topic like this:
*Topic Name*
• Decision: what was decided (or "None")
• Context: 1-2 sentence background
• Your action: what you need to do (or "None needed")

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

// ExtractDecisions extracts decisions and action items from channel messages.
func (c *Client) ExtractDecisions(ctx context.Context, messages []string, channelID string, daysBack int) (string, error) {
	prompt := fmt.Sprintf(`You are analyzing a Slack channel's recent message history to extract decisions.

Channel: <#%s> — Last %d days — %d messages

RULES:
- Only list actual decisions that were made (agreed upon, approved, or confirmed)
- Include the approximate context (who was involved, what triggered it)
- List action items that followed each decision
- If no decisions were made, say "No formal decisions found — discussion was exploratory."
- Use Slack mrkdwn: *bold*, bullet points with •, NO markdown headers

FORMAT each decision as:
✅ *[Decision]*
• Context: [1 sentence]
• Action items: [list owners if mentioned, or "None assigned"]

Messages:
%s`, channelID, daysBack, len(messages), strings.Join(messages, "\n"))

	resp, err := c.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: c.model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: "You extract formal decisions from Slack conversations for neurodivergent professionals. Be precise, literal, and concise.",
			},
			{Role: openai.ChatMessageRoleUser, Content: prompt},
		},
		Temperature: 0.2,
		MaxTokens:   1200,
	})
	if err != nil {
		return "", fmt.Errorf("openai decisions: %w", err)
	}
	return extractContent(resp)
}

// GenerateActionPlan creates a structured action plan from a natural language goal.
func (c *Client) GenerateActionPlan(ctx context.Context, goal, neurotype, memoryContext string) (string, error) {
	// Neurotype-specific framing
	framingMap := map[string]string{
		"adhd":        "The user has ADHD. Use very short tasks (1-2 minutes each). Lead with the most important task. Use energy labels (🔋 Low / ⚡ Medium / 🚀 High energy).",
		"autism":      "The user is autistic. Be completely literal and specific. Include exact steps. No ambiguous instructions like 'reach out' — say exactly 'send a Slack DM saying X'.",
		"anxiety":     "The user has anxiety. Start with the easiest task to build momentum. Include a ✅ checkpoint after every 2-3 tasks. Don't use urgent language.",
		"unspecified": "Use a clear, structured format with time estimates.",
		"ally":        "Standard professional format.",
	}
	framing := framingMap[neurotype]
	if framing == "" {
		framing = framingMap["unspecified"]
	}

	contextSection := ""
	if memoryContext != "" {
		contextSection = fmt.Sprintf("\n\nUser context:\n%s\n", memoryContext)
	}

	prompt := fmt.Sprintf(`Create a concrete action plan for this goal: %q
%s
%s
FORMAT (use Slack mrkdwn — *bold*, bullet •, NO ## headers):
*Task 1 — [Task name]* (~X min)
• What to do: [specific action]
• Done when: [clear completion criteria]

*Task 2 — ...*
...

End with:
*Total estimated time:* X minutes
*Best time to start:* [suggestion based on task type]`, goal, contextSection, framing)

	resp, err := c.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: c.model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: "You create practical, neurodivergent-friendly action plans. Be specific, not vague.",
			},
			{Role: openai.ChatMessageRoleUser, Content: prompt},
		},
		Temperature: 0.3,
		MaxTokens:   1200,
	})
	if err != nil {
		return "", fmt.Errorf("openai plan: %w", err)
	}
	return extractContent(resp)
}

// SummarizeThread creates a short AI summary of a Slack thread for the reacting user.
func (c *Client) SummarizeThread(ctx context.Context, messages []string, neurotype string) (string, error) {
	framingMap := map[string]string{
		"adhd":   "3 bullet points max. Lead with any action required.",
		"autism": "Be literal. State what was decided explicitly. No implied meaning.",
		"anxiety": "Calm tone. Confirm if action is required or not.",
	}
	framing := framingMap[neurotype]
	if framing == "" {
		framing = "Clear and concise. Max 5 bullet points."
	}

	prompt := fmt.Sprintf(`Summarize this Slack thread. %s

Use Slack mrkdwn (*bold*, bullet •). NO markdown headers.

Format:
*What happened:* [1 sentence]
*Key points:*
• [point]
• [point]
*Your action:* [specific action, or "None required"]

Thread messages:
%s`, framing, strings.Join(messages, "\n---\n"))

	resp, err := c.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: c.model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: "You summarize Slack threads concisely for neurodivergent professionals.",
			},
			{Role: openai.ChatMessageRoleUser, Content: prompt},
		},
		Temperature: 0.3,
		MaxTokens:   600,
	})
	if err != nil {
		return "", fmt.Errorf("openai thread summary: %w", err)
	}
	return extractContent(resp)
}

// ClassifyUrgency classifies an incoming message as urgent, normal, or low priority.
// Used by Smart Notifications during quiet hours / deep work.
func (c *Client) ClassifyUrgency(ctx context.Context, message string) (string, error) {
	resp, err := c.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: c.model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: "Classify a Slack message as: URGENT, NORMAL, or LOW. Reply with ONLY that one word.\n\nURGENT = needs immediate response (incident, deadline today, explicit 'urgent'/'asap', someone is blocked).\nNORMAL = needs a response but can wait a few hours.\nLOW = informational, no action needed, social.",
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: fmt.Sprintf("Message: %q", message),
			},
		},
		Temperature: 0,
		MaxTokens:   5,
	})
	if err != nil {
		return "NORMAL", fmt.Errorf("openai urgency: %w", err)
	}
	content, err := extractContent(resp)
	if err != nil {
		return "NORMAL", err
	}
	result := strings.TrimSpace(strings.ToUpper(content))
	if result != "URGENT" && result != "LOW" {
		return "NORMAL", nil
	}
	return result, nil
}

// NeurotypicAwareAnalyzeTone is an enhanced AnalyzeTone that adapts output for the user's neurotype.
func (c *Client) NeurotypicAwareAnalyzeTone(ctx context.Context, message, neurotype string) (*domain.ToneAnalysis, error) {
	addons := map[string]string{
		"autism":  "\n\nADDITIONAL RULE: Spell out any implied social norms explicitly. If the message is passive-aggressive, say exactly what the sender is feeling. Never use vague social phrases.",
		"anxiety": "\n\nADDITIONAL RULE: If the message is neutral or positive, explicitly state that in the Note. Never leave ambiguity about whether the sender is upset.",
		"adhd":    "\n\nADDITIONAL RULE: Put the Action first. Keep Note to one sentence maximum.",
	}

	systemPrompt := "You are a direct, kind translator of workplace subtext for autistic and ADHD adults. You decode ambiguous messages into literal meaning without judgment. Respond in this exact format:\n- Tone: [single word or short phrase]\n- Intent: [1 sentence, what they want]\n- Action: [1 sentence, what you should do]\n- Note: [1-2 sentences explaining any hidden social context]\n\nRules:\n- Never say 'they might be' — be confident but not rude.\n- Never use 'just' or 'simply' — these are patronizing.\n- If the message is genuinely neutral, say so clearly.\n- If the message is passive-aggressive, name it directly but kindly."

	if addon, ok := addons[neurotype]; ok {
		systemPrompt += addon
	}

	resp, err := c.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: c.model,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: systemPrompt},
			{Role: openai.ChatMessageRoleUser, Content: fmt.Sprintf("Analyze this message: %q", message)},
		},
		Temperature: 0.3,
		MaxTokens:   500,
	})
	if err != nil {
		return nil, fmt.Errorf("openai neurotypic tone: %w", err)
	}
	content, err := extractContent(resp)
	if err != nil {
		return nil, err
	}
	return parseToneResponse(content), nil
}
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


