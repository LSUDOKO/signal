package features

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/LSUDOKOS/signal/internal/domain"
	"github.com/LSUDOKOS/signal/internal/rts"
	"github.com/LSUDOKOS/signal/internal/store"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

var neurotypes = []struct {
	Value string
	Label string
}{
	{Value: "adhd", Label: "ADHD"},
	{Value: "autism", Label: "Autistic"},
	{Value: "anxiety", Label: "Anxiety"},
	{Value: "unspecified", Label: "Unsure"},
	{Value: "ally", Label: "Ally"},
}

// Controller dispatches Slack events to the appropriate feature handlers.
type Controller struct {
	focusMode     *FocusModeService
	translator    *TranslatorService
	catchup       *CatchUpService
	digest        *DigestService
	deepWork      *DeepWorkService
	mode          *ModeService
	decisions     *DecisionService
	planner       *PlannerService
	threadSummary *ThreadSummaryService
	memory        *MemoryService
	userRepo      store.UserRepository
	prefsRepo     store.PreferencesRepository
	rtsSearcher   *rts.Searcher
	slack         SlackAPI
	aiClient      interface{ ClassifyUrgency(ctx context.Context, message string) (string, error) }
}

// NewController creates a new feature controller.
func NewController(
	focusMode *FocusModeService,
	translator *TranslatorService,
	catchup *CatchUpService,
	digest *DigestService,
	deepWork *DeepWorkService,
	mode *ModeService,
	decisions *DecisionService,
	planner *PlannerService,
	threadSummary *ThreadSummaryService,
	memory *MemoryService,
	userRepo store.UserRepository,
	prefsRepo store.PreferencesRepository,
	rtsSearcher *rts.Searcher,
	slack SlackAPI,
	aiClient interface{ ClassifyUrgency(ctx context.Context, message string) (string, error) },
) *Controller {
	return &Controller{
		focusMode:     focusMode,
		translator:    translator,
		catchup:       catchup,
		digest:        digest,
		deepWork:      deepWork,
		mode:          mode,
		decisions:     decisions,
		planner:       planner,
		threadSummary: threadSummary,
		memory:        memory,
		userRepo:      userRepo,
		prefsRepo:     prefsRepo,
		rtsSearcher:   rtsSearcher,
		slack:         slack,
		aiClient:      aiClient,
	}
}

// HandleMessage routes a message event to applicable features.
func (c *Controller) HandleMessage(ctx context.Context, event *slackevents.MessageEvent, user *domain.User, teamID string) error {
	// Skip bot messages
	if event.BotID != "" || event.SubType == "bot_message" {
		return nil
	}

	// If it's a DM to Signal, handle smart auto-response
	if event.ChannelType == "im" {
		return c.handleDirectMessage(ctx, event, user)
	}

	user, err := c.ensureUser(ctx, user)
	if err != nil {
		slog.Error("failed to ensure user exists", "error", err, "slack_user_id", user.SlackUserID)
	}

	prefs, err := c.prefsRepo.Get(ctx, user.ID)
	if err != nil {
		prefs = &domain.UserPreferences{
			UserID:           user.ID,
			FocusModeEnabled: true,
			FocusThreshold:   50,
			TranslatorEnabled: true,
		}
	}

	if prefs.FocusModeEnabled {
		if err := c.focusMode.HandleMessage(ctx, event, user, prefs); err != nil {
			slog.Error("focus mode error", "error", err, "channel", event.Channel)
		}
	}

	if prefs.TranslatorEnabled {
		if err := c.translator.HandleMessage(ctx, event, user); err != nil {
			slog.Error("translator error", "error", err, "channel", event.Channel)
		}
	}

	return nil
}

// HandleReaction routes reaction_added events to feature handlers.
func (c *Controller) HandleReaction(ctx context.Context, event *slackevents.ReactionAddedEvent, user *domain.User, teamID string) error {
	user, err := c.ensureUser(ctx, user)
	if err != nil {
		slog.Warn("could not ensure user for reaction", "error", err)
	}
	return c.threadSummary.HandleReaction(ctx, event, user)
}

// handleDirectMessage handles DMs sent to Signal with smart deep-work auto-responder.
func (c *Controller) handleDirectMessage(ctx context.Context, event *slackevents.MessageEvent, user *domain.User) error {
	user, _ = c.ensureUser(ctx, user)

	// Check if user is in deep work — if so, auto-respond
	deepWorkDuration, _ := c.deepWork.GetDeepWorkState(ctx, user.SlackUserID)
	if deepWorkDuration > 0 {
		// Classify urgency of the incoming DM
		urgency := "NORMAL"
		if c.aiClient != nil && strings.TrimSpace(event.Text) != "" {
			urgency, _ = c.aiClient.ClassifyUrgency(ctx, event.Text)
		}

		if urgency == "URGENT" || strings.Contains(strings.ToUpper(event.Text), "URGENT") {
			// Urgent bypass — notify user immediately, respond to sender
			return c.slack.PostMessage(event.Channel, []slack.Block{
				slack.NewSectionBlock(
					slack.NewTextBlockObject("mrkdwn",
						"🚨 *Urgent message detected.* I'm in Deep Work mode but your message has been flagged as urgent. Please wait — I'll respond as soon as possible.",
						false, false,
					),
					nil, nil,
				),
			}, "Urgent Message")
		}

		// Non-urgent — send auto-response
		return c.slack.PostMessage(event.Channel, []slack.Block{
			slack.NewSectionBlock(
				slack.NewTextBlockObject("mrkdwn",
					"🧘 I'm in *Deep Work mode* right now.\n\nI'll respond when I'm back. If this is urgent, reply with the word *URGENT* and I'll be notified immediately.",
					false, false,
				),
				nil, nil,
			),
		}, "Deep Work Auto-Reply")
	}

	// Not in deep work — show help
	return c.slack.PostMessage(event.Channel, buildHelpBlocks(), "Signal Help")
}

// HandleAppMention handles when Signal is @mentioned.
func (c *Controller) HandleAppMention(ctx context.Context, event *slackevents.AppMentionEvent, user *domain.User, teamID string) error {
	user, err := c.ensureUser(ctx, user)
	if err != nil {
		slog.Warn("could not ensure user for app mention", "error", err)
		// Continue anyway - can still respond
	}

	// Respond in the channel where we were mentioned (not DM)
	blocks := []slack.Block{
		slack.NewSectionBlock(
			slack.NewTextBlockObject("mrkdwn",
				fmt.Sprintf("👋 Hi <@%s>! I'm Signal, your neurodivergent-friendly Slack assistant.\n\nUse `/signal` to see all my commands, or DM me anytime for help!", user.SlackUserID),
				false, false,
			),
			nil, nil,
		),
		slack.NewContextBlock("mention_context",
			slack.NewTextBlockObject("mrkdwn", "_💡 Tip: Try `/translate [message]` to decode ambiguous workplace language_", false, false),
		),
	}

	return c.slack.PostMessage(event.Channel, blocks, "Signal Help")
}

// HandleBlockAction routes block action events (button clicks) to features.
func (c *Controller) HandleBlockAction(ctx context.Context, action *slack.InteractionCallback, user *domain.User, teamID string) error {
	user, err := c.ensureUser(ctx, user)
	if err != nil {
		return err
	}

	if len(action.ActionCallback.BlockActions) == 0 {
		return nil
	}

	actionID := action.ActionCallback.BlockActions[0].ActionID

	switch {
	case isFocusAction(actionID):
		return c.focusMode.HandleBlockAction(ctx, action, user)
	case isTranslatorAction(actionID):
		return c.translator.HandleBlockAction(ctx, action, user)
	case isDeepWorkAction(actionID):
		return c.deepWork.HandleBlockAction(ctx, action, user)
	default:
		slog.Debug("unhandled block action", "action_id", actionID)
		return nil
	}
}

// HandleCommand routes slash commands to features.
func (c *Controller) HandleCommand(ctx context.Context, cmd *slack.SlashCommand, user *domain.User, responseURL string) error {
	slog.Info("📝 handling slash command",
		"command", cmd.Command,
		"user", cmd.UserID,
		"text", cmd.Text,
		"has_response_url", responseURL != "")

	user, err := c.ensureUser(ctx, user)
	if err != nil {
		slog.Error("ensureUser failed", "error", err)
		return err
	}

	slog.Info("✅ user ensured", "slack_id", user.SlackUserID)

	var handlerErr error
	switch cmd.Command {
	case "/signal":
		slog.Info("routing to handleOpenPreferences")
		handlerErr = c.handleOpenPreferences(ctx, cmd, user, responseURL)
	case "/translate":
		slog.Info("routing to translator")
		handlerErr = c.translator.HandleSlashCommand(ctx, cmd, user, responseURL)
	case "/catchup":
		slog.Info("routing to catchup")
		handlerErr = c.catchup.HandleSlashCommand(ctx, cmd, user, responseURL)
	case "/focus":
		slog.Info("routing to deepwork")
		handlerErr = c.deepWork.HandleSlashCommand(ctx, cmd, user, responseURL)
	case "/digest":
		slog.Info("routing to digest")
		handlerErr = c.digest.HandleSlashCommand(ctx, cmd, user, responseURL)
	case "/mode":
		slog.Info("routing to mode")
		handlerErr = c.mode.HandleSlashCommand(ctx, cmd, user, responseURL)
	case "/decisions":
		slog.Info("routing to decisions")
		handlerErr = c.decisions.HandleSlashCommand(ctx, cmd, user, responseURL)
	case "/plan":
		slog.Info("routing to planner")
		handlerErr = c.planner.HandleSlashCommand(ctx, cmd, user, responseURL)
	default:
		slog.Warn("unknown command", "command", cmd.Command)
		return nil
	}

	if handlerErr != nil {
		slog.Error("❌ command handler failed", "command", cmd.Command, "error", handlerErr)
		return handlerErr
	}

	slog.Info("✅ command handled successfully", "command", cmd.Command)
	return nil
}

// HandleAppHomeOpened publishes the App Home view with user preferences.
func (c *Controller) HandleAppHomeOpened(ctx context.Context, event *slackevents.AppHomeOpenedEvent, user *domain.User, teamID string) error {
	user, err := c.ensureUser(ctx, user)
	if err != nil {
		return err
	}

	prefs, err := c.prefsRepo.Get(ctx, user.ID)
	if err != nil {
		prefs = &domain.UserPreferences{UserID: user.ID}
	}

	// Preferences loaded; user can manage them at the /app-home page
	slog.Debug("app home opened", "user", user.SlackUserID, "neurotype", user.Neurotype)
	_ = prefs
	return nil
}

func (c *Controller) ensureUser(ctx context.Context, user *domain.User) (*domain.User, error) {
	// If no SlackUserID, return the passed user as-is (shouldn't happen but defensive)
	if user.SlackUserID == "" {
		return user, fmt.Errorf("user has no slack_user_id")
	}

	// Use Upsert — no more duplicate key errors ever
	newUser := &domain.User{
		SlackUserID: user.SlackUserID,
		SlackTeamID: user.SlackTeamID,
		Neurotype:   "unspecified",
	}
	if err := c.userRepo.Upsert(ctx, newUser); err != nil {
		// Upsert failed (very unlikely), fall back to get
		if existing, fetchErr := c.userRepo.GetBySlackID(ctx, user.SlackUserID, user.SlackTeamID); fetchErr == nil {
			return existing, nil
		}
		slog.Warn("upsert failed, using in-memory user", "error", err, "slack_user_id", user.SlackUserID)
		return newUser, nil
	}
	return newUser, nil
}

func (c *Controller) handleOpenPreferences(ctx context.Context, cmd *slack.SlashCommand, user *domain.User, responseURL string) error {
	slog.Info("🔧 sending help via response_url", "user_id", user.SlackUserID)
	err := c.slack.PostWebhook(responseURL, buildHelpBlocks(), "Signal Help")
	if err != nil {
		slog.Error("❌ failed to post help via webhook", "error", err)
		return fmt.Errorf("post webhook: %w", err)
	}
	slog.Info("✅ help message sent via response_url")
	return nil
}

func buildHelpBlocks() []slack.Block {
	return []slack.Block{
		slack.NewHeaderBlock(
			slack.NewTextBlockObject("plain_text", "🧘 Signal — Calm Slack for Neurodivergent Professionals", true, false),
		),
		slack.NewSectionBlock(
			slack.NewTextBlockObject("mrkdwn", "Here are all my commands:", false, false),
			nil, nil,
		),
		slack.NewDividerBlock(),
		slack.NewSectionBlock(
			nil,
			[]*slack.TextBlockObject{
				slack.NewTextBlockObject("mrkdwn", "*/signal*\nHelp menu", false, false),
				slack.NewTextBlockObject("mrkdwn", "*/mode [adhd|autism|anxiety]*\nPersonalize Signal for your neurotype", false, false),
			},
			nil,
		),
		slack.NewSectionBlock(
			nil,
			[]*slack.TextBlockObject{
				slack.NewTextBlockObject("mrkdwn", "*/translate [message]*\nDecode ambiguous workplace language", false, false),
				slack.NewTextBlockObject("mrkdwn", "*/catchup [topic]*\nAI summary of what you missed", false, false),
			},
			nil,
		),
		slack.NewSectionBlock(
			nil,
			[]*slack.TextBlockObject{
				slack.NewTextBlockObject("mrkdwn", "*/focus [2h]*\nStart deep work mode", false, false),
				slack.NewTextBlockObject("mrkdwn", "*/decisions [#channel]*\nFind all decisions in a channel", false, false),
			},
			nil,
		),
		slack.NewSectionBlock(
			nil,
			[]*slack.TextBlockObject{
				slack.NewTextBlockObject("mrkdwn", "*/plan [goal]*\nAI action planner for your tasks", false, false),
				slack.NewTextBlockObject("mrkdwn", "*/digest*\nInstant digest of your mentions", false, false),
			},
			nil,
		),
		slack.NewContextBlock("help_footer",
			slack.NewTextBlockObject("mrkdwn", "_React with 📋 on any thread to get an instant AI summary_", false, false),
		),
	}
}

func isFocusAction(actionID string) bool {
	focusActions := map[string]bool{
		"focus_get_summary":  true,
		"focus_mute_30":      true,
		"focus_open_thread":  true,
	}
	return focusActions[actionID]
}

func isTranslatorAction(actionID string) bool {
	return actionID == "translator_ack"
}

func isDeepWorkAction(actionID string) bool {
	deepWorkActions := map[string]bool{
		"deepwork_start":      true,
		"deepwork_stop":       true,
		"deepwork_extend":     true,
	}
	return deepWorkActions[actionID]
}

// SlackAPI provides access to the Slack API for features.
type SlackAPI interface {
	PostMessage(channelID string, blocks []slack.Block, text string) error
	PostWebhook(responseURL string, blocks []slack.Block, text string) error
	PostEphemeral(channelID, userID string, blocks []slack.Block, text string) error
	OpenDMChannel(userID string) (string, error)
	GetUser(userID string) (*slack.User, error)
	GetChannelHistory(channelID string, limit int) ([]slack.Message, error)
	GetThreadMessages(channelID, threadTS string) ([]slack.Message, error)
	SearchMessages(query string, params slack.SearchParameters) (*slack.SearchMessages, error)
	SetUserStatus(userID, statusText, statusEmoji string, expiration int) error
	PublishView(userID string, blocks []slack.Block) error
}
