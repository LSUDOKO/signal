package features

import (
	"context"
	"fmt"
	"log/slog"

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
	focusMode   *FocusModeService
	translator  *TranslatorService
	catchup     *CatchUpService
	digest      *DigestService
	deepWork    *DeepWorkService
	userRepo    store.UserRepository
	prefsRepo   store.PreferencesRepository
	rtsSearcher *rts.Searcher
	slack       SlackAPI
}

// NewController creates a new feature controller.
func NewController(
	focusMode *FocusModeService,
	translator *TranslatorService,
	catchup *CatchUpService,
	digest *DigestService,
	deepWork *DeepWorkService,
	userRepo store.UserRepository,
	prefsRepo store.PreferencesRepository,
	rtsSearcher *rts.Searcher,
	slack SlackAPI,
) *Controller {
	return &Controller{
		focusMode:   focusMode,
		translator:  translator,
		catchup:     catchup,
		digest:      digest,
		deepWork:    deepWork,
		userRepo:    userRepo,
		prefsRepo:   prefsRepo,
		rtsSearcher: rtsSearcher,
		slack:       slack,
	}
}

// HandleMessage routes a message event to applicable features.
func (c *Controller) HandleMessage(ctx context.Context, event *slackevents.MessageEvent, user *domain.User, teamID string) error {
	// Skip bot messages and thread replies (handled separately if needed)
	if event.BotID != "" || event.SubType == "bot_message" {
		return nil
	}

	// Ensure user exists
	user, err := c.ensureUser(ctx, user)
	if err != nil {
		return err
	}

	// Get user preferences
	prefs, err := c.prefsRepo.Get(ctx, user.ID)
	if err != nil {
		// Default preferences if not set
		prefs = &domain.UserPreferences{
			UserID:            user.ID,
			FocusModeEnabled:  true,
			FocusThreshold:    50,
			TranslatorEnabled: true,
		}
	}

	// Route to features
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

// HandleAppMention handles when Signal is @mentioned.
func (c *Controller) HandleAppMention(ctx context.Context, event *slackevents.AppMentionEvent, user *domain.User, teamID string) error {
	user, err := c.ensureUser(ctx, user)
	if err != nil {
		return err
	}

	// Provide help message or process command in mention
	_ = user // Future: could route to help system
	return nil
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
func (c *Controller) HandleCommand(ctx context.Context, cmd *slack.SlashCommand, user *domain.User) error {
	user, err := c.ensureUser(ctx, user)
	if err != nil {
		return err
	}

	switch cmd.Command {
	case "/signal":
		return c.handleOpenPreferences(ctx, cmd, user)
	case "/translate":
		return c.translator.HandleSlashCommand(ctx, cmd, user)
	case "/catchup":
		return c.catchup.HandleSlashCommand(ctx, cmd, user)
	case "/focus":
		return c.deepWork.HandleSlashCommand(ctx, cmd, user)
	case "/digest":
		return c.digest.HandleSlashCommand(ctx, cmd, user)
	default:
		return nil
	}
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
	existing, err := c.userRepo.GetBySlackID(ctx, user.SlackUserID, user.SlackTeamID)
	if err != nil {
		// Create new user
		newUser := &domain.User{
			SlackUserID: user.SlackUserID,
			SlackTeamID: user.SlackTeamID,
		}
		if err := c.userRepo.Create(ctx, newUser); err != nil {
			return nil, err
		}
		return newUser, nil
	}
	return existing, nil
}

func (c *Controller) handleOpenPreferences(ctx context.Context, cmd *slack.SlashCommand, user *domain.User) error {
	// Send help message with available commands and preferences link
	blocks := []slack.Block{
		slack.NewHeaderBlock(
			slack.NewTextBlockObject("plain_text", "🧘 Signal — Calm Slack for Neurodivergent Professionals", true, false),
		),
		slack.NewSectionBlock(
			slack.NewTextBlockObject("mrkdwn",
				"Here are the commands you can use:",
				false, false,
			),
			nil, nil,
		),
		slack.NewDividerBlock(),
		slack.NewSectionBlock(
			nil,
			[]*slack.TextBlockObject{
				slack.NewTextBlockObject("mrkdwn",
					"*/signal*\nOpen this help menu", false, false),
				slack.NewTextBlockObject("mrkdwn",
					"*/translate [message]*\nTranslate ambiguous workplace language", false, false),
			},
			nil,
		),
		slack.NewSectionBlock(
			nil,
			[]*slack.TextBlockObject{
				slack.NewTextBlockObject("mrkdwn",
					"*/catchup [topic]*\nGet AI summary of what you missed", false, false),
				slack.NewTextBlockObject("mrkdwn",
					"*/focus [duration]*\nStart deep work mode (e.g., /focus 2h)", false, false),
			},
			nil,
		),
		slack.NewSectionBlock(
			slack.NewTextBlockObject("mrkdwn",
				"*/digest*\nForce-send your Quiet Hours Digest now\n\n*Need help?* Visit the support page at /support or open an issue on GitHub.",
				false, false,
			),
			nil, nil,
		),
	}

	return c.slack.PostMessage(cmd.ChannelID, blocks, "Signal Help")
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
	PostEphemeral(channelID, userID string, blocks []slack.Block, text string) error
	OpenDMChannel(userID string) (string, error)
	GetUser(userID string) (*slack.User, error)
	GetChannelHistory(channelID string, limit int) ([]slack.Message, error)
	SearchMessages(query string, params slack.SearchParameters) (*slack.SearchMessages, error)
	SetUserStatus(userID, statusText, statusEmoji string, expiration int) error
	PublishView(userID string, blocks []slack.Block) error
}
