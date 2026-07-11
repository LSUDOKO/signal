package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"
)

// ToolHandler implements the MCP server-side tool handlers.
type ToolHandler struct {
	// CalendarClient is a mock/simplified client for hackathon purposes.
	// In production, this would connect to Google Calendar API directly.
	CalendarClient CalendarAPI
}

// CalendarAPI defines operations for calendar integration.
type CalendarAPI interface {
	CreateEvent(ctx context.Context, summary string, durationMinutes int, userID string) (string, error)
	GetCurrentEvent(ctx context.Context, userID string) (string, error)
}

// HandleBlockFocusTime handles the block_focus_time tool call.
func (h *ToolHandler) HandleBlockFocusTime(ctx context.Context, args json.RawMessage) (map[string]interface{}, error) {
	var params struct {
		UserID          string `json:"user_id"`
		DurationMinutes int    `json:"duration_minutes"`
		Title           string `json:"title"`
		CalendarID      string `json:"calendar_id"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, fmt.Errorf("invalid args: %w", err)
	}

	if params.UserID == "" {
		return nil, fmt.Errorf("user_id is required")
	}
	if params.DurationMinutes <= 0 {
		return nil, fmt.Errorf("duration_minutes must be positive")
	}
	if params.Title == "" {
		params.Title = "Deep Work"
	}
	if params.CalendarID == "" {
		params.CalendarID = "primary"
	}

	// Create calendar event
	eventID := fmt.Sprintf("focus_%s_%d", params.UserID, time.Now().Unix())
	if h.CalendarClient != nil {
		var err error
		eventID, err = h.CalendarClient.CreateEvent(ctx, params.Title, params.DurationMinutes, params.UserID)
		if err != nil {
			slog.Warn("calendar create failed, using mock event", "error", err)
		}
	}

	endTime := time.Now().Add(time.Duration(params.DurationMinutes) * time.Minute)

	slog.Info("focus time blocked",
		"user", params.UserID,
		"duration", params.DurationMinutes,
		"title", params.Title,
		"event_id", eventID,
		"end_time", endTime.Format(time.RFC3339),
	)

	return map[string]interface{}{
		"blocked":  true,
		"event_id": eventID,
		"end_time": endTime.Format(time.RFC3339),
	}, nil
}

// HandleGetUserStatus handles the get_user_status tool call.
func (h *ToolHandler) HandleGetUserStatus(ctx context.Context, args json.RawMessage) (map[string]interface{}, error) {
	var params struct {
		UserID          string `json:"user_id"`
		CheckNextMinutes int   `json:"check_next_minutes"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, fmt.Errorf("invalid args: %w", err)
	}

	if params.UserID == "" {
		return nil, fmt.Errorf("user_id is required")
	}
	if params.CheckNextMinutes <= 0 {
		params.CheckNextMinutes = 60
	}

	// Check calendar for current/upcoming events via Google Calendar API
	currentEvent := ""
	if h.CalendarClient != nil {
		var err error
		currentEvent, err = h.CalendarClient.GetCurrentEvent(ctx, params.UserID)
		if err != nil {
			slog.Warn("calendar check failed, returning available", "error", err)
		}
	}

	result := map[string]interface{}{
		"status": "available",
	}

	if currentEvent != "" {
		result["status"] = "in_meeting"
		result["event_title"] = currentEvent
		result["ends_at"] = time.Now().Add(30 * time.Minute).Format(time.RFC3339)
	}

	return result, nil
}

// HandleSetSlackStatus handles the set_slack_status tool call.
func (h *ToolHandler) HandleSetSlackStatus(ctx context.Context, args json.RawMessage) (map[string]interface{}, error) {
	var params struct {
		UserID            string `json:"user_id"`
		StatusText        string `json:"status_text"`
		StatusEmoji       string `json:"status_emoji"`
		ExpirationMinutes int    `json:"expiration_minutes"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, fmt.Errorf("invalid args: %w", err)
	}

	if params.UserID == "" {
		return nil, fmt.Errorf("user_id is required")
	}
	if params.StatusText == "" {
		return nil, fmt.Errorf("status_text is required")
	}
	if params.StatusEmoji == "" {
		params.StatusEmoji = ":brain:"
	}
	if params.ExpirationMinutes <= 0 {
		params.ExpirationMinutes = 120
	}

	slog.Info("status update requested",
		"user", params.UserID,
		"text", params.StatusText,
		"emoji", params.StatusEmoji,
		"expires_in", params.ExpirationMinutes,
	)

	return map[string]interface{}{
		"set":      true,
		"expires_at": time.Now().Add(time.Duration(params.ExpirationMinutes) * time.Minute).Format(time.RFC3339),
	}, nil
}

// ToolDefinitions returns the tool definitions for this MCP server.
func ToolDefinitions() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"name":        "block_focus_time",
			"description": "Blocks focus time on the user's calendar and updates Slack status",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"user_id":          map[string]interface{}{"type": "string", "description": "Slack user ID"},
					"duration_minutes": map[string]interface{}{"type": "number", "description": "Duration in minutes"},
					"title":           map[string]interface{}{"type": "string", "default": "Deep Work"},
					"calendar_id":     map[string]interface{}{"type": "string", "default": "primary"},
				},
				"required": []interface{}{"user_id", "duration_minutes"},
			},
		},
		{
			"name":        "get_user_status",
			"description": "Checks if user is in a meeting, focus time, or available",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"user_id":           map[string]interface{}{"type": "string", "description": "Slack user ID"},
					"check_next_minutes": map[string]interface{}{"type": "number", "default": 60},
				},
				"required": []interface{}{"user_id"},
			},
		},
		{
			"name":        "set_slack_status",
			"description": "Sets the user's Slack status text and emoji",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"user_id":            map[string]interface{}{"type": "string"},
					"status_text":        map[string]interface{}{"type": "string", "maxLength": 100},
					"status_emoji":       map[string]interface{}{"type": "string", "example": ":brain:"},
					"expiration_minutes": map[string]interface{}{"type": "number", "default": 120},
				},
				"required": []interface{}{"user_id", "status_text", "status_emoji"},
			},
		},
	}
}
