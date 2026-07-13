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

	"github.com/LSUDOKOS/signal/internal/domain"
	"github.com/LSUDOKOS/signal/internal/ai"
	"github.com/slack-go/slack"
)

// GitHubService handles /github command — search PRs, issues, repos via real GitHub API.
type GitHubService struct {
	slack      SlackAPI
	ai         *ai.Client
	memory     *MemoryService
	httpClient *http.Client
	token      string // GitHub personal access token
	defaultOrg string // default GitHub org/user to search
}

// NewGitHubService creates a new GitHubService.
// token: GitHub personal access token (fine-grained or classic with repo scope)
// defaultOrg: default GitHub org or username to scope searches to
func NewGitHubService(slack SlackAPI, ai *ai.Client, memory *MemoryService, token, defaultOrg string) *GitHubService {
	return &GitHubService{
		slack:      slack,
		ai:         ai,
		memory:     memory,
		httpClient: &http.Client{Timeout: 15 * time.Second},
		token:      token,
		defaultOrg: defaultOrg,
	}
}

// IsConfigured returns true if GitHub token is set.
func (g *GitHubService) IsConfigured() bool {
	return g.token != ""
}

// githubIssue is a minimal GitHub issue/PR struct.
type githubIssue struct {
	Number    int    `json:"number"`
	Title     string `json:"title"`
	State     string `json:"state"`
	HTMLURL   string `json:"html_url"`
	Body      string `json:"body"`
	User      struct {
		Login string `json:"login"`
	} `json:"user"`
	PullRequest *struct{} `json:"pull_request,omitempty"` // non-nil if this is a PR
	CreatedAt   string    `json:"created_at"`
	UpdatedAt   string    `json:"updated_at"`
	Labels      []struct {
		Name string `json:"name"`
	} `json:"labels"`
}

// githubRepo is a minimal GitHub repo struct.
type githubRepo struct {
	FullName    string `json:"full_name"`
	Description string `json:"description"`
	HTMLURL     string `json:"html_url"`
	Language    string `json:"language"`
	OpenIssues  int    `json:"open_issues_count"`
	StarCount   int    `json:"stargazers_count"`
	UpdatedAt   string `json:"updated_at"`
}

// HandleSlashCommand handles /github [query].
// Examples:
//   /github open PRs
//   /github issues assigned to me
//   /github my open PRs in signal
//   /github repo signal
func (g *GitHubService) HandleSlashCommand(ctx context.Context, cmd *slack.SlashCommand, user *domain.User, responseURL string) error {
	if !g.IsConfigured() {
		return g.slack.PostWebhook(responseURL, []slack.Block{
			slack.NewSectionBlock(
				slack.NewTextBlockObject("mrkdwn",
					"⚙️ *GitHub MCP not configured*\n\nAsk your admin to add `GITHUB_TOKEN` and `GITHUB_ORG` to Signal's configuration.\n\nOnce configured, you can:\n• `/github open PRs` — see your open pull requests\n• `/github issues` — see your assigned issues\n• `/github repo [name]` — get repo info",
					false, false,
				),
				nil, nil,
			),
		}, "GitHub Not Configured")
	}

	query := strings.TrimSpace(cmd.Text)
	if query == "" {
		return g.showGitHubHelp(responseURL)
	}

	// Parse intent from query
	queryLower := strings.ToLower(query)

	switch {
	case strings.Contains(queryLower, "repo ") || strings.HasPrefix(queryLower, "repo"):
		repoName := strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(queryLower, "repo "), "repo"))
		return g.handleRepoInfo(ctx, cmd, user, responseURL, repoName)
	case strings.Contains(queryLower, "pr") || strings.Contains(queryLower, "pull request"):
		return g.handlePRSearch(ctx, cmd, user, responseURL, query)
	case strings.Contains(queryLower, "issue") || strings.Contains(queryLower, "bug"):
		return g.handleIssueSearch(ctx, cmd, user, responseURL, query)
	default:
		// Use AI to understand the intent and do a general search
		return g.handleGeneralSearch(ctx, cmd, user, responseURL, query)
	}
}

func (g *GitHubService) handlePRSearch(ctx context.Context, cmd *slack.SlashCommand, user *domain.User, responseURL, query string) error {
	queryLower := strings.ToLower(query)

	// Build GitHub search query
	searchQ := fmt.Sprintf("is:pr is:open org:%s", g.defaultOrg)
	if strings.Contains(queryLower, "my") || strings.Contains(queryLower, "mine") || strings.Contains(queryLower, "assigned") {
		// Get GitHub username from memory or use Slack display name
		ghUser := g.memory.GetGitHubUser(ctx, user.SlackUserID)
		if ghUser != "" {
			searchQ += " author:" + ghUser
		}
	}
	if strings.Contains(queryLower, "review") {
		searchQ += " review:required"
	}

	issues, err := g.searchIssues(ctx, searchQ, 10)
	if err != nil {
		slog.Error("github: PR search failed", "error", err)
		return g.slack.PostWebhook(responseURL, []slack.Block{
			slack.NewSectionBlock(
				slack.NewTextBlockObject("mrkdwn", "❌ GitHub search failed. Check that the token has `repo` scope.", false, false),
				nil, nil,
			),
		}, "GitHub Error")
	}

	return g.postIssuesResult(responseURL, "Open Pull Requests", searchQ, issues, true)
}

func (g *GitHubService) handleIssueSearch(ctx context.Context, cmd *slack.SlashCommand, user *domain.User, responseURL, query string) error {
	queryLower := strings.ToLower(query)

	searchQ := fmt.Sprintf("is:issue is:open org:%s", g.defaultOrg)
	if strings.Contains(queryLower, "my") || strings.Contains(queryLower, "assigned to me") {
		ghUser := g.memory.GetGitHubUser(ctx, user.SlackUserID)
		if ghUser != "" {
			searchQ += " assignee:" + ghUser
		}
	}
	if strings.Contains(queryLower, "bug") {
		searchQ += " label:bug"
	}
	if strings.Contains(queryLower, "urgent") || strings.Contains(queryLower, "priority") {
		searchQ += " label:priority"
	}

	issues, err := g.searchIssues(ctx, searchQ, 10)
	if err != nil {
		slog.Error("github: issue search failed", "error", err)
		return g.slack.PostWebhook(responseURL, []slack.Block{
			slack.NewSectionBlock(
				slack.NewTextBlockObject("mrkdwn", "❌ GitHub search failed.", false, false),
				nil, nil,
			),
		}, "GitHub Error")
	}

	return g.postIssuesResult(responseURL, "Open Issues", searchQ, issues, false)
}

func (g *GitHubService) handleRepoInfo(ctx context.Context, cmd *slack.SlashCommand, user *domain.User, responseURL, repoName string) error {
	// If no slash in repo name, prepend default org
	if !strings.Contains(repoName, "/") && g.defaultOrg != "" {
		repoName = g.defaultOrg + "/" + repoName
	}

	repo, err := g.getRepo(ctx, repoName)
	if err != nil {
		slog.Error("github: repo fetch failed", "error", err, "repo", repoName)
		return g.slack.PostWebhook(responseURL, []slack.Block{
			slack.NewSectionBlock(
				slack.NewTextBlockObject("mrkdwn", fmt.Sprintf("❌ Couldn't find repo `%s`. Check the name and try again.", repoName), false, false),
				nil, nil,
			),
		}, "Repo Not Found")
	}

	desc := repo.Description
	if desc == "" {
		desc = "_No description_"
	}

	blocks := []slack.Block{
		slack.NewHeaderBlock(
			slack.NewTextBlockObject("plain_text", "📦 GitHub Repository", true, false),
		),
		slack.NewSectionBlock(
			slack.NewTextBlockObject("mrkdwn",
				fmt.Sprintf("*<%s|%s>*\n%s", repo.HTMLURL, repo.FullName, desc),
				false, false,
			),
			nil, nil,
		),
		slack.NewSectionBlock(
			nil,
			[]*slack.TextBlockObject{
				slack.NewTextBlockObject("mrkdwn", fmt.Sprintf("*Language:*\n%s", orDefault(repo.Language, "N/A")), false, false),
				slack.NewTextBlockObject("mrkdwn", fmt.Sprintf("*Open Issues:*\n%d", repo.OpenIssues), false, false),
			},
			nil,
		),
		slack.NewSectionBlock(
			nil,
			[]*slack.TextBlockObject{
				slack.NewTextBlockObject("mrkdwn", fmt.Sprintf("*Stars:*\n⭐ %d", repo.StarCount), false, false),
				slack.NewTextBlockObject("mrkdwn", fmt.Sprintf("*Updated:*\n%s", formatGitHubDate(repo.UpdatedAt)), false, false),
			},
			nil,
		),
		slack.NewContextBlock("github_repo_footer",
			slack.NewTextBlockObject("mrkdwn", fmt.Sprintf("_Use `/github issues` or `/github open PRs` to dig deeper_"), false, false),
		),
	}

	return g.slack.PostWebhook(responseURL, blocks, "GitHub Repository")
}

func (g *GitHubService) handleGeneralSearch(ctx context.Context, cmd *slack.SlashCommand, user *domain.User, responseURL, query string) error {
	// Let AI interpret the query and build a search
	searchQ := fmt.Sprintf("org:%s %s", g.defaultOrg, query)

	issues, err := g.searchIssues(ctx, searchQ, 8)
	if err != nil || len(issues) == 0 {
		return g.slack.PostWebhook(responseURL, []slack.Block{
			slack.NewSectionBlock(
				slack.NewTextBlockObject("mrkdwn",
					fmt.Sprintf("🔍 No results for *\"%s\"* in `%s`.\n\n*Try:*\n• `/github open PRs`\n• `/github issues`\n• `/github repo [name]`", query, g.defaultOrg),
					false, false,
				),
				nil, nil,
			),
		}, "No Results")
	}

	return g.postIssuesResult(responseURL, "Search Results", query, issues, false)
}

func (g *GitHubService) postIssuesResult(responseURL, title, query string, issues []githubIssue, isPR bool) error {
	if len(issues) == 0 {
		label := "issues"
		if isPR {
			label = "pull requests"
		}
		return g.slack.PostWebhook(responseURL, []slack.Block{
			slack.NewSectionBlock(
				slack.NewTextBlockObject("mrkdwn",
					fmt.Sprintf("✅ No open %s found for `%s`.", label, query),
					false, false,
				),
				nil, nil,
			),
		}, "No Results")
	}

	icon := "🐛"
	if isPR {
		icon = "🔀"
	}

	blocks := []slack.Block{
		slack.NewHeaderBlock(
			slack.NewTextBlockObject("plain_text", fmt.Sprintf("%s %s (%d)", icon, title, len(issues)), true, false),
		),
	}

	for i, issue := range issues {
		if i >= 8 {
			break
		}
		labels := ""
		for _, l := range issue.Labels {
			labels += fmt.Sprintf("`%s` ", l.Name)
		}

		text := fmt.Sprintf("*<%s|#%d: %s>*\n👤 %s  •  %s  %s",
			issue.HTMLURL,
			issue.Number,
			issue.Title,
			issue.User.Login,
			strings.ToUpper(issue.State),
			labels,
		)

		blocks = append(blocks, slack.NewSectionBlock(
			slack.NewTextBlockObject("mrkdwn", text, false, false),
			nil, nil,
		))
	}

	if len(issues) > 8 {
		blocks = append(blocks, slack.NewContextBlock("github_more",
			slack.NewTextBlockObject("mrkdwn", fmt.Sprintf("_%d more results. Refine your search to narrow down._", len(issues)-8), false, false),
		))
	}

	return g.slack.PostWebhook(responseURL, blocks, title)
}

// searchIssues calls the GitHub Search API.
func (g *GitHubService) searchIssues(ctx context.Context, query string, limit int) ([]githubIssue, error) {
	apiURL := fmt.Sprintf("https://api.github.com/search/issues?q=%s&per_page=%d&sort=updated&order=desc",
		url.QueryEscape(query), limit)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	g.setHeaders(req)

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("github api: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("github api status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Items []githubIssue `json:"items"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return result.Items, nil
}

// getRepo fetches a single repo by full name (owner/repo).
func (g *GitHubService) getRepo(ctx context.Context, fullName string) (*githubRepo, error) {
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s", fullName)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, err
	}
	g.setHeaders(req)

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("github api: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("repo not found: %s", fullName)
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("github api status %d: %s", resp.StatusCode, string(body))
	}

	var repo githubRepo
	if err := json.NewDecoder(resp.Body).Decode(&repo); err != nil {
		return nil, fmt.Errorf("decode repo: %w", err)
	}
	return &repo, nil
}

func (g *GitHubService) setHeaders(req *http.Request) {
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	if g.token != "" {
		req.Header.Set("Authorization", "Bearer "+g.token)
	}
}

func (g *GitHubService) showGitHubHelp(responseURL string) error {
	return g.slack.PostWebhook(responseURL, []slack.Block{
		slack.NewHeaderBlock(
			slack.NewTextBlockObject("plain_text", "🐙 GitHub MCP", true, false),
		),
		slack.NewSectionBlock(
			slack.NewTextBlockObject("mrkdwn",
				fmt.Sprintf("*Organization:* `%s`\n\n*Available commands:*\n• `/github open PRs` — list open pull requests\n• `/github my PRs` — your open pull requests\n• `/github issues` — open issues in the org\n• `/github issues assigned to me` — your assigned issues\n• `/github repo [name]` — repository info\n• `/github [anything]` — natural language search", g.defaultOrg),
				false, false,
			),
			nil, nil,
		),
	}, "GitHub Help")
}

func formatGitHubDate(s string) string {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return s
	}
	return t.Format("Jan 2, 2006")
}

func orDefault(s, def string) string {
	if s == "" {
		return def
	}
	return s
}
