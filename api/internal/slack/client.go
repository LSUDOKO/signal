package slack

import (
	"fmt"
	"log/slog"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
)

// NewClient creates a Slack API client and Socket Mode client.
func NewClient(botToken, appToken string, featureCtrl FeatureController) (*EventHandler, error) {
	api := slack.New(
		botToken,
		slack.OptionAppLevelToken(appToken),
	)

	socketClient := socketmode.New(
		api,
	)

	handler := NewEventHandler(socketClient, api, featureCtrl)

	// Get bot user info
	authTest, err := api.AuthTest()
	if err != nil {
		return nil, fmt.Errorf("auth test: %w", err)
	}
	handler.SetBotInfo(authTest.UserID, authTest.User)
	slog.Info("connected to slack", "bot_user", authTest.UserID, "team", authTest.Team)

	return handler, nil
}

// HandlerLogger adapts slog to socketmode's logger interface.
type HandlerLogger struct{}

func (l HandlerLogger) Output(calldepth int, s string) error {
	slog.Debug(s)
	return nil
}
