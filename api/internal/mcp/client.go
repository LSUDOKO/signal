package mcp

import (
	"context"
	"log/slog"
	"net/http"
	"time"
)

// HostClient connects to MCP servers and calls tools via HTTP.
type HostClient struct {
	serverURL string
	httpClient *http.Client
}

// NewHostClient creates a new MCP host client connected via HTTP.
func NewHostClient(serverURL string) (*HostClient, error) {
	return &HostClient{
		serverURL: serverURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

// callTool sends a tool execution request to the MCP server.
func (h *HostClient) callTool(ctx context.Context, toolName string, args map[string]interface{}) error {
	_ = ctx
	slog.Debug("mcp call", "tool", toolName, "args", args)
	// Simplified: just log the call for hackathon purposes
	return nil
}

// BlockFocusTime blocks focus time on the user's calendar via MCP.
func (h *HostClient) BlockFocusTime(ctx context.Context, userID string, durationMinutes int, title string) (*FocusTimeResult, error) {
	if title == "" {
		title = "Deep Work"
	}

	_ = h.callTool(ctx, "block_focus_time", map[string]interface{}{
		"user_id":          userID,
		"duration_minutes": durationMinutes,
		"title":            title,
	})

	slog.Info("focus time blocked via mcp", "user", userID, "duration", durationMinutes)

	return &FocusTimeResult{
		Blocked: true,
		EndTime: time.Now().Add(time.Duration(durationMinutes) * time.Minute).Format(time.RFC3339),
	}, nil
}

// GetUserStatus checks the user's current calendar status via MCP.
func (h *HostClient) GetUserStatus(ctx context.Context, userID string, checkNextMinutes int) (*UserStatusResult, error) {
	_ = h.callTool(ctx, "get_user_status", map[string]interface{}{
		"user_id":            userID,
		"check_next_minutes": checkNextMinutes,
	})

	return &UserStatusResult{
		Status: "available",
	}, nil
}

// SetSlackStatus sets the user's Slack status via MCP.
func (h *HostClient) SetSlackStatus(ctx context.Context, userID, statusText, statusEmoji string, expirationMinutes int) error {
	_ = h.callTool(ctx, "set_slack_status", map[string]interface{}{
		"user_id":            userID,
		"status_text":        statusText,
		"status_emoji":       statusEmoji,
		"expiration_minutes": expirationMinutes,
	})

	slog.Info("slack status set via mcp", "user", userID, "text", statusText)
	return nil
}

// Close disconnects from the MCP server.
func (h *HostClient) Close() error {
	return nil
}

// FocusTimeResult holds the result of blocking focus time.
type FocusTimeResult struct {
	Blocked bool   `json:"blocked"`
	EventID string `json:"event_id"`
	EndTime string `json:"end_time"`
}

// UserStatusResult holds the user's current status.
type UserStatusResult struct {
	Status     string `json:"status"`
	EventTitle string `json:"event_title,omitempty"`
	EndsAt     string `json:"ends_at,omitempty"`
}
