package slack

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/LSUDOKOS/signal/internal/domain"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
)

// EventHandler handles Slack events and dispatches to features.
type EventHandler struct {
	client      *socketmode.Client
	api         *slack.Client
	botUserID   string
	botUserName string
	featureCtrl FeatureController
}

// FeatureController defines the interface for feature-level event handling.
type FeatureController interface {
	HandleMessage(ctx context.Context, event *slackevents.MessageEvent, user *domain.User, teamID string) error
	HandleAppMention(ctx context.Context, event *slackevents.AppMentionEvent, user *domain.User, teamID string) error
	HandleBlockAction(ctx context.Context, action *slack.InteractionCallback, user *domain.User, teamID string) error
	HandleCommand(ctx context.Context, cmd *slack.SlashCommand, user *domain.User) error
	HandleAppHomeOpened(ctx context.Context, event *slackevents.AppHomeOpenedEvent, user *domain.User, teamID string) error
}

// NewEventHandler creates a new Slack event handler.
func NewEventHandler(client *socketmode.Client, api *slack.Client, featureCtrl FeatureController) *EventHandler {
	return &EventHandler{
		client:      client,
		api:         api,
		featureCtrl: featureCtrl,
	}
}

// SetBotInfo sets the bot's user ID and name for reference.
func (h *EventHandler) SetBotInfo(userID, userName string) {
	h.botUserID = userID
	h.botUserName = userName
}

// SetFeatureCtrl sets the feature controller after initialization.
func (h *EventHandler) SetFeatureCtrl(fc FeatureController) {
	h.featureCtrl = fc
}

// GetAPI returns the underlying Slack API client.
func (h *EventHandler) GetAPI() *slack.Client {
	return h.api
}

// Start begins listening for Slack events via Socket Mode.
func (h *EventHandler) Start(ctx context.Context) error {
	slog.Info("starting slack socket mode handler")

	// Run the socket mode client - it handles the event loop internally
	go h.eventLoop(ctx)

	return h.client.RunContext(ctx)
}

// eventLoop processes events from the socket mode client.
func (h *EventHandler) eventLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			slog.Info("slack event handler shutting down")
			return
		case event := <-h.client.Events:
			h.handleEvent(ctx, event)
		}
	}
}

func (h *EventHandler) handleEvent(ctx context.Context, event socketmode.Event) {
	switch event.Type {
	case socketmode.EventTypeEventsAPI:
		if event.Request != nil {
			h.handleEventsAPI(ctx, event)
		}
	case socketmode.EventTypeInteractive:
		if event.Request != nil {
			h.handleInteractive(ctx, event)
		}
	case socketmode.EventTypeSlashCommand:
		if event.Request != nil {
			h.handleSlashCommand(ctx, event)
		}
	case socketmode.EventTypeConnected:
		slog.Info("slack socket mode connected")
	default:
		slog.Debug("unhandled socket event type", "type", event.Type)
		if event.Request != nil {
			h.client.Ack(*event.Request)
		}
	}
}

func (h *EventHandler) handleEventsAPI(ctx context.Context, se socketmode.Event) {
	eventData, ok := se.Data.(slackevents.EventsAPIEvent)
	if !ok {
		slog.Error("failed to cast events API event")
		return
	}

	// Ack synchronously — must be fast to keep Slack happy
	h.client.Ack(*se.Request)

	// Process events asynchronously so the event loop stays responsive
	asyncCtx := context.WithoutCancel(ctx)
	switch eventData.Type {
	case slackevents.CallbackEvent:
		innerEvent := eventData.InnerEvent
		go h.handleCallbackEvent(asyncCtx, innerEvent)
	default:
		slog.Debug("unhandled events API type", "type", eventData.Type)
	}
}

func (h *EventHandler) handleCallbackEvent(ctx context.Context, innerEvent slackevents.EventsAPIInnerEvent) {
	switch event := innerEvent.Data.(type) {
	case *slackevents.MessageEvent:
		h.handleMessageEvent(ctx, event)
	case *slackevents.AppMentionEvent:
		h.handleAppMentionEvent(ctx, event)
	case *slackevents.AppHomeOpenedEvent:
		h.handleAppHomeOpenedEvent(ctx, event)
	case *slackevents.MemberJoinedChannelEvent:
		slog.Info("member joined channel", "channel", event.Channel, "user", event.User)
	default:
		slog.Debug("unhandled callback event", "type", innerEvent.Type)
	}
}

func (h *EventHandler) handleMessageEvent(ctx context.Context, event *slackevents.MessageEvent) {
	// Skip bot messages
	if event.BotID != "" || event.User == h.botUserID {
		return
	}

	// Skip message subtypes we don't handle (message_changed, message_deleted, etc.)
	if event.SubType != "" && event.SubType != "bot_message" {
		return
	}

	if h.featureCtrl == nil {
		slog.Warn("feature controller is nil, skipping message")
		return
	}

	user := &domain.User{
		SlackUserID: event.User,
	}

	if err := h.featureCtrl.HandleMessage(ctx, event, user, ""); err != nil {
		slog.Error("error handling message", "error", err, "channel", event.Channel, "user", event.User)
	}
}

func (h *EventHandler) handleAppMentionEvent(ctx context.Context, event *slackevents.AppMentionEvent) {
	if h.featureCtrl == nil {
		return
	}

	user := &domain.User{
		SlackUserID: event.User,
	}

	if err := h.featureCtrl.HandleAppMention(ctx, event, user, ""); err != nil {
		slog.Error("error handling app mention", "error", err)
	}
}

func (h *EventHandler) handleAppHomeOpenedEvent(ctx context.Context, event *slackevents.AppHomeOpenedEvent) {
	if h.featureCtrl == nil {
		return
	}

	user := &domain.User{
		SlackUserID: event.User,
	}

	if err := h.featureCtrl.HandleAppHomeOpened(ctx, event, user, ""); err != nil {
		slog.Error("error handling app home opened", "error", err)
	}
}

func (h *EventHandler) handleInteractive(ctx context.Context, se socketmode.Event) {
	actionEvent, ok := se.Data.(slack.InteractionCallback)
	if !ok {
		return
	}

	// Ack synchronously — must be fast to keep Slack happy
	h.client.Ack(*se.Request)

	if h.featureCtrl == nil {
		return
	}

	user := &domain.User{
		SlackUserID: actionEvent.User.ID,
		SlackTeamID: actionEvent.Team.ID,
	}

	// Process block action asynchronously so the event loop stays responsive
	asyncCtx := context.WithoutCancel(ctx)
	actionCopy := actionEvent
	go func() {
		if err := h.featureCtrl.HandleBlockAction(asyncCtx, &actionCopy, user, actionCopy.Team.ID); err != nil {
			slog.Error("error handling block action", "error", err)
		}
	}()
}

func (h *EventHandler) handleSlashCommand(ctx context.Context, se socketmode.Event) {
	cmd, ok := se.Data.(slack.SlashCommand)
	if !ok {
		return
	}

	// Acknowledge immediately with an ephemeral "processing" message
	// (Slack requires a response within 3 seconds or it shows "app did not respond")
	h.client.Ack(*se.Request, map[string]interface{}{
		"response_type": "ephemeral",
		"text":          "⏳ Working on your request...",
	})

	if h.featureCtrl == nil {
		return
	}

	user := &domain.User{
		SlackUserID: cmd.UserID,
		SlackTeamID: cmd.TeamID,
	}

	// Process the command asynchronously — the ack already confirmed receipt
	// to Slack, and the final result will be posted via chat.postMessage
	// by the feature handler. Running in a goroutine keeps the event loop
	// responsive so Slack doesn't drop the socket connection.
	asyncCtx := context.WithoutCancel(ctx)
	cmdCopy := cmd
	go func() {
		if err := h.featureCtrl.HandleCommand(asyncCtx, &cmdCopy, user); err != nil {
			slog.Error("error handling command", "error", err, "command", cmdCopy.Command)
		}
	}()
}

// PostMessage sends a message to a Slack channel.
func (h *EventHandler) PostMessage(channelID string, blocks []slack.Block, text string) error {
	_, _, err := h.api.PostMessage(channelID,
		slack.MsgOptionBlocks(blocks...),
		slack.MsgOptionText(text, false),
	)
	return err
}

// PostEphemeral sends an ephemeral message to a specific user in a channel.
func (h *EventHandler) PostEphemeral(channelID, userID string, blocks []slack.Block, text string) error {
	_, err := h.api.PostEphemeral(channelID, userID,
		slack.MsgOptionBlocks(blocks...),
		slack.MsgOptionText(text, false),
	)
	return err
}

// OpenDMChannel opens (or finds) a DM channel with a user.
func (h *EventHandler) OpenDMChannel(userID string) (string, error) {
	channel, _, _, err := h.api.OpenConversation(
		&slack.OpenConversationParameters{
			Users: []string{userID},
		},
	)
	if err != nil {
		return "", fmt.Errorf("open dm: %w", err)
	}
	if channel != nil {
		return channel.ID, nil
	}
	return "", fmt.Errorf("open dm: no channel returned")
}

// GetUser retrieves user info from Slack API.
func (h *EventHandler) GetUser(userID string) (*slack.User, error) {
	return h.api.GetUserInfo(userID)
}

// GetChannelHistory fetches recent messages from a channel.
func (h *EventHandler) GetChannelHistory(channelID string, limit int) ([]slack.Message, error) {
	history, err := h.api.GetConversationHistory(
		&slack.GetConversationHistoryParameters{
			ChannelID: channelID,
			Limit:     limit,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("get history: %w", err)
	}
	return history.Messages, nil
}

// SearchMessages performs a RTS search via Slack API.
func (h *EventHandler) SearchMessages(query string, params slack.SearchParameters) (*slack.SearchMessages, error) {
	result, err := h.api.SearchMessages(query, params)
	if err != nil {
		return nil, fmt.Errorf("search messages: %w", err)
	}
	return result, nil
}

// SetUserStatus sets a user's Slack status.
func (h *EventHandler) SetUserStatus(userID, statusText, statusEmoji string, expiration int) error {
	_ = userID // SetUserCustomStatus applies to the signed-in user's status
	return h.api.SetUserCustomStatus(statusText, statusEmoji, int64(expiration))
}

// PublishView publishes a Home tab view for a user.
func (h *EventHandler) PublishView(userID string, blocks []slack.Block) error {
	view := slack.HomeTabViewRequest{
		Type:   "home",
		Blocks: slack.Blocks{BlockSet: blocks},
	}
	_, err := h.api.PublishView(userID, view, "")
	if err != nil {
		return fmt.Errorf("publish view: %w", err)
	}
	return nil
}

// UnmarshalJSON is a helper to parse raw JSON into a typed event.
func UnmarshalJSON(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}
