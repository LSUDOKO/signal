package features

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/LSUDOKOS/signal/internal/domain"
	"github.com/LSUDOKOS/signal/internal/store"
	"github.com/slack-go/slack"
)

// ModeService handles the /mode command — neurotype-aware personalization.
type ModeService struct {
	slack     SlackAPI
	userRepo  store.UserRepository
	prefsRepo store.PreferencesRepository
}

// NewModeService creates a new ModeService.
func NewModeService(slack SlackAPI, userRepo store.UserRepository, prefsRepo store.PreferencesRepository) *ModeService {
	return &ModeService{slack: slack, userRepo: userRepo, prefsRepo: prefsRepo}
}

// neurotypeMeta describes each mode for display and AI prompting.
var neurotypeMeta = map[string]struct {
	emoji       string
	label       string
	description string
	tips        []string
}{
	"adhd": {
		emoji: "⚡",
		label: "ADHD Mode",
		description: "Summaries are short and action-focused. Focus Mode triggers earlier (30 messages). Digests highlight only urgent items. AI responses skip preamble.",
		tips: []string{
			"• Summaries: bullet points only, 5 items max",
			"• Focus Mode triggers at 30 messages (not 50)",
			"• Digest: urgent-only view",
			"• All AI replies: lead with the answer, not the context",
		},
	},
	"autism": {
		emoji: "🧩",
		label: "Autism Mode",
		description: "Social Translator is always on. Tone analysis includes explicit subtext. AI responses are literal and unambiguous. Passive-aggressive detection is heightened.",
		tips: []string{
			"• Translator: auto-detects more phrase patterns",
			"• Tone analysis: names subtext directly",
			"• Summaries: explicit cause-and-effect, no implied meaning",
			"• Action items: stated as direct instructions",
		},
	},
	"anxiety": {
		emoji: "🌿",
		label: "Anxiety Mode",
		description: "Reassuring tone in all AI responses. Digest de-emphasizes urgency labels. Focus Guard activates automatically. Translator adds a 'this is routine' note for neutral messages.",
		tips: []string{
			"• AI tone: calm, reassuring, never alarmist",
			"• Digest: removes 🔴 URGENT labels (shows priority instead)",
			"• Focus Guard: always active",
			"• Translator: neutral messages confirmed as non-threatening",
		},
	},
	"unspecified": {
		emoji: "🧘",
		label: "Default Mode",
		description: "Balanced settings. All features use standard thresholds.",
		tips:        []string{"• All features at default settings"},
	},
	"ally": {
		emoji: "🤝",
		label: "Ally Mode",
		description: "Optimized for neurotypical team members supporting neurodivergent colleagues. Provides context on why teammates may communicate differently.",
		tips: []string{
			"• Focus summaries include 'why this matters' context",
			"• Translator includes 'ally perspective' notes",
			"• Digest: full view including all priority levels",
		},
	},
}

// HandleSlashCommand handles /mode [adhd|autism|anxiety|unspecified|ally].
func (m *ModeService) HandleSlashCommand(ctx context.Context, cmd *slack.SlashCommand, user *domain.User, responseURL string) error {
	arg := strings.TrimSpace(strings.ToLower(cmd.Text))

	// /mode with no arg — show current mode + options
	if arg == "" {
		return m.showModeMenu(responseURL, user.Neurotype)
	}

	// Validate neurotype
	meta, ok := neurotypeMeta[arg]
	if !ok {
		return m.slack.PostWebhook(responseURL, []slack.Block{
			slack.NewSectionBlock(
				slack.NewTextBlockObject("mrkdwn",
					"❓ Unknown mode. Choose one of:\n`/mode adhd` `/ mode autism` `/mode anxiety` `/mode ally` `/mode unspecified`",
					false, false,
				),
				nil, nil,
			),
		}, "Unknown Mode")
	}

	// Update user neurotype
	user.Neurotype = arg
	if err := m.userRepo.Update(ctx, user); err != nil {
		slog.Error("failed to update neurotype", "error", err, "user", user.SlackUserID)
		// Don't fail — still show success
	}

	// Update preferences based on neurotype
	prefs, err := m.prefsRepo.Get(ctx, user.ID)
	if err != nil {
		prefs = &domain.UserPreferences{UserID: user.ID}
	}
	applyNeurotypePrefDefaults(prefs, arg)
	if err := m.prefsRepo.Upsert(ctx, prefs); err != nil {
		slog.Error("failed to upsert prefs for neurotype", "error", err)
	}

	tipsText := strings.Join(meta.tips, "\n")
	blocks := []slack.Block{
		slack.NewHeaderBlock(
			slack.NewTextBlockObject("plain_text", fmt.Sprintf("%s %s Activated", meta.emoji, meta.label), true, false),
		),
		slack.NewSectionBlock(
			slack.NewTextBlockObject("mrkdwn", meta.description, false, false),
			nil, nil,
		),
		slack.NewDividerBlock(),
		slack.NewSectionBlock(
			slack.NewTextBlockObject("mrkdwn", fmt.Sprintf("*What changed:*\n%s", tipsText), false, false),
			nil, nil,
		),
		slack.NewContextBlock("mode_context",
			slack.NewTextBlockObject("mrkdwn", "_Use `/mode` anytime to switch modes or see your current setting._", false, false),
		),
	}

	return m.slack.PostWebhook(responseURL, blocks, fmt.Sprintf("%s %s Activated", meta.emoji, meta.label))
}

func (m *ModeService) showModeMenu(responseURL, currentNeurotype string) error {
	current := currentNeurotype
	if current == "" {
		current = "unspecified"
	}
	currentMeta := neurotypeMeta[current]

	blocks := []slack.Block{
		slack.NewHeaderBlock(
			slack.NewTextBlockObject("plain_text", "🎛️ Signal Mode Settings", true, false),
		),
		slack.NewSectionBlock(
			slack.NewTextBlockObject("mrkdwn",
				fmt.Sprintf("*Current mode:* %s %s\n\n%s", currentMeta.emoji, currentMeta.label, currentMeta.description),
				false, false,
			),
			nil, nil,
		),
		slack.NewDividerBlock(),
		slack.NewSectionBlock(
			slack.NewTextBlockObject("mrkdwn",
				"*Available modes:*\n• `/mode adhd` — Short summaries, early focus triggers\n• `/mode autism` — Literal AI, enhanced tone detection\n• `/mode anxiety` — Calm framing, auto focus guard\n• `/mode ally` — Ally perspective for supporting teammates\n• `/mode unspecified` — Balanced defaults",
				false, false,
			),
			nil, nil,
		),
	}
	return m.slack.PostWebhook(responseURL, blocks, "Signal Mode Settings")
}

// applyNeurotypePrefDefaults adjusts preferences based on selected neurotype.
func applyNeurotypePrefDefaults(prefs *domain.UserPreferences, neurotype string) {
	switch neurotype {
	case "adhd":
		prefs.FocusModeEnabled = true
		prefs.FocusThreshold = 30   // Trigger earlier
		prefs.TranslatorEnabled = true
		prefs.DigestEnabled = true
		prefs.DigestHour = 16
		prefs.DeepWorkAutoDetect = true
	case "autism":
		prefs.FocusModeEnabled = true
		prefs.FocusThreshold = 50
		prefs.TranslatorEnabled = true // Always on
		prefs.DigestEnabled = true
		prefs.DigestHour = 17
		prefs.DeepWorkAutoDetect = false
	case "anxiety":
		prefs.FocusModeEnabled = true
		prefs.FocusThreshold = 40
		prefs.TranslatorEnabled = true
		prefs.DigestEnabled = true
		prefs.DigestHour = 16
		prefs.DeepWorkAutoDetect = true
	case "ally":
		prefs.FocusModeEnabled = false
		prefs.FocusThreshold = 50
		prefs.TranslatorEnabled = false
		prefs.DigestEnabled = false
	default: // unspecified
		prefs.FocusModeEnabled = true
		prefs.FocusThreshold = 50
		prefs.TranslatorEnabled = true
		prefs.DigestEnabled = false
	}
}
