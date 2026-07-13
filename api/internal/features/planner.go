package features

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/LSUDOKOS/signal/internal/ai"
	"github.com/LSUDOKOS/signal/internal/domain"
	"github.com/slack-go/slack"
)

// PlannerService handles the /plan command — AI action planner.
type PlannerService struct {
	slack  SlackAPI
	ai     *ai.Client
	memory *MemoryService
}

// NewPlannerService creates a new PlannerService.
func NewPlannerService(slack SlackAPI, ai *ai.Client, memory *MemoryService) *PlannerService {
	return &PlannerService{slack: slack, ai: ai, memory: memory}
}

// HandleSlashCommand handles /plan [description of what needs to be done].
// Example: /plan I need to finish the Q3 report and review Alex's PR by Friday
func (p *PlannerService) HandleSlashCommand(ctx context.Context, cmd *slack.SlashCommand, user *domain.User, responseURL string) error {
	goal := strings.TrimSpace(cmd.Text)
	if goal == "" {
		return p.slack.PostWebhook(responseURL, []slack.Block{
			slack.NewSectionBlock(
				slack.NewTextBlockObject("mrkdwn",
					"📋 *Signal Action Planner*\n\nUsage: `/plan [what you need to accomplish]`\n\n*Examples:*\n• `/plan Finish the Q3 report and review Alex's PR by Friday`\n• `/plan Prepare for the design review meeting tomorrow`\n• `/plan Catch up on all messages from this week`",
					false, false,
				),
				nil, nil,
			),
		}, "Action Planner Help")
	}

	// Pull user context from memory
	memoryContext := p.memory.GetContext(ctx, user.SlackUserID)

	// Generate action plan with AI
	plan, err := p.ai.GenerateActionPlan(ctx, goal, user.Neurotype, memoryContext)
	if err != nil {
		slog.Error("planner: AI plan generation failed", "error", err)
		return p.slack.PostWebhook(responseURL, []slack.Block{
			slack.NewSectionBlock(
				slack.NewTextBlockObject("mrkdwn", "❌ Couldn't generate a plan right now. Please try again.", false, false),
				nil, nil,
			),
		}, "Plan Error")
	}

	// Save to memory
	p.memory.AddInteraction(ctx, user.SlackUserID, fmt.Sprintf("planned: %s", goal))

	blocks := []slack.Block{
		slack.NewHeaderBlock(
			slack.NewTextBlockObject("plain_text", "📋 Your Action Plan", true, false),
		),
		slack.NewSectionBlock(
			slack.NewTextBlockObject("mrkdwn", fmt.Sprintf("*Goal:* %s", goal), false, false),
			nil, nil,
		),
		slack.NewDividerBlock(),
		slack.NewSectionBlock(
			slack.NewTextBlockObject("mrkdwn", plan, false, false),
			nil, nil,
		),
		slack.NewContextBlock("plan_footer",
			slack.NewTextBlockObject("mrkdwn", "_Use `/plan` again to create a new plan. Plans are saved to your AI memory._", false, false),
		),
	}

	return p.slack.PostWebhook(responseURL, blocks, "Action Plan")
}
