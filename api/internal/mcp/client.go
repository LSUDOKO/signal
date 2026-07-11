package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
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

// callTool sends a tool execution request to the MCP server and parses the response.
func (h *HostClient) callTool(ctx context.Context, toolName string, args map[string]interface{}, result interface{}) error {
	body, err := json.Marshal(args)
	if err != nil {
		return fmt.Errorf("marshal args: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		h.serverURL+"/tools/"+toolName,
		bytes.NewReader(body),
	)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("mcp request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("mcp returned %d: %s", resp.StatusCode, string(respBody))
	}

	// Parse response body into the result struct
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read mcp response: %w", err)
	}

	if result != nil {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("parse mcp response: %w", err)
		}
	}

	slog.Debug("mcp call successful", "tool", toolName, "response", string(respBody))
	return nil
}

// BlockFocusTime blocks focus time on the user's calendar via MCP.
func (h *HostClient) BlockFocusTime(ctx context.Context, userID string, durationMinutes int, title string) (*FocusTimeResult, error) {
	if title == "" {
		title = "Deep Work"
	}

	result := &FocusTimeResult{}
	err := h.callTool(ctx, "block_focus_time", map[string]interface{}{
		"user_id":          userID,
		"duration_minutes": durationMinutes,
		"title":            title,
	}, result)
	if err != nil {
		return nil, fmt.Errorf("mcp block focus time: %w", err)
	}

	slog.Info("focus time blocked via mcp", "user", userID, "duration", durationMinutes, "event_id", result.EventID)
	return result, nil
}

// GetUserStatus checks the user's current calendar status via MCP.
func (h *HostClient) GetUserStatus(ctx context.Context, userID string, checkNextMinutes int) (*UserStatusResult, error) {
	result := &UserStatusResult{}
	err := h.callTool(ctx, "get_user_status", map[string]interface{}{
		"user_id":            userID,
		"check_next_minutes": checkNextMinutes,
	}, result)
	if err != nil {
		return nil, fmt.Errorf("mcp get user status: %w", err)
	}

	return result, nil
}

// SetSlackStatus sets the user's Slack status via MCP.
func (h *HostClient) SetSlackStatus(ctx context.Context, userID, statusText, statusEmoji string, expirationMinutes int) error {
	result := make(map[string]interface{})
	err := h.callTool(ctx, "set_slack_status", map[string]interface{}{
		"user_id":            userID,
		"status_text":        statusText,
		"status_emoji":       statusEmoji,
		"expiration_minutes": expirationMinutes,
	}, &result)
	if err != nil {
		return fmt.Errorf("mcp set slack status: %w", err)
	}

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
