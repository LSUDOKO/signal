package mcp

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"golang.org/x/oauth2/google"
	calendar "google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

// GoogleCalendarClient implements CalendarAPI using the Google Calendar API.
// It authenticates via a service account with domain-wide delegation.
type GoogleCalendarClient struct {
	service    *calendar.Service
	calendarID string // usually "primary" or a specific calendar email
}

// NewGoogleCalendarClient creates a new Google Calendar client from a service account credentials file.
// If credentialsPath is empty, it returns nil (caller should handle gracefully).
// The credentials file should be a Google Cloud service account JSON key with Calendar API access.
func NewGoogleCalendarClient(ctx context.Context, credentialsPath, calendarID string) (*GoogleCalendarClient, error) {
	if credentialsPath == "" {
		slog.Warn("no google calendar credentials path set, calendar features will use mock")
		return nil, nil
	}

	if calendarID == "" {
		calendarID = "primary"
	}

	// Try explicit credentials file first (from GOOGLE_CALENDAR_CREDENTIALS env var in mcp-server)
	srv, err := calendar.NewService(ctx, option.WithCredentialsFile(credentialsPath))
	if err == nil {
		slog.Info("google calendar client initialized from credentials file")
		return &GoogleCalendarClient{
			service:    srv,
			calendarID: calendarID,
		}, nil
	}

	// Fallback: try GOOGLE_APPLICATION_CREDENTIALS env var or default credentials
	creds, err := google.FindDefaultCredentials(ctx, calendar.CalendarEventsScope)
	if err != nil {
		return nil, fmt.Errorf("no google calendar credentials found (tried %q and GOOGLE_APPLICATION_CREDENTIALS): %w",
			credentialsPath, err)
	}

	srv, err = calendar.NewService(ctx, option.WithCredentials(creds))
	if err != nil {
		return nil, fmt.Errorf("create calendar service from default creds: %w", err)
	}

	slog.Info("google calendar client initialized from default credentials")
	return &GoogleCalendarClient{
		service:    srv,
		calendarID: calendarID,
	}, nil
}

// CreateEvent creates a calendar event for focus time.
func (c *GoogleCalendarClient) CreateEvent(ctx context.Context, summary string, durationMinutes int, userID string) (string, error) {
	if c == nil || c.service == nil {
		return "", fmt.Errorf("calendar client not initialized")
	}

	startTime := time.Now()
	endTime := startTime.Add(time.Duration(durationMinutes) * time.Minute)

	event := &calendar.Event{
		Summary:     summary,
		Description: fmt.Sprintf("Focus time session initiated by Signal Slack Agent for user %s", userID),
		Start: &calendar.EventDateTime{
			DateTime: startTime.Format(time.RFC3339),
			TimeZone: "UTC",
		},
		End: &calendar.EventDateTime{
			DateTime: endTime.Format(time.RFC3339),
			TimeZone: "UTC",
		},
		Transparency: "opaque", // Shows as "busy"
		ColorId:      "2",      // Green color
	}

	created, err := c.service.Events.Insert(c.calendarID, event).Context(ctx).Do()
	if err != nil {
		return "", fmt.Errorf("create calendar event: %w", err)
	}

	slog.Info("google calendar event created",
		"summary", summary,
		"event_id", created.Id,
		"start", startTime.Format(time.RFC3339),
		"end", endTime.Format(time.RFC3339),
	)

	return created.Id, nil
}

// GetCurrentEvent checks if the user has any current or upcoming events.
func (c *GoogleCalendarClient) GetCurrentEvent(ctx context.Context, userID string) (string, error) {
	if c == nil || c.service == nil {
		return "", fmt.Errorf("calendar client not initialized")
	}

	now := time.Now()
	endTime := now.Add(60 * time.Minute)

	events, err := c.service.Events.List(c.calendarID).
		ShowDeleted(false).
		SingleEvents(true).
		TimeMin(now.Format(time.RFC3339)).
		TimeMax(endTime.Format(time.RFC3339)).
		OrderBy("startTime").
		MaxResults(5).
		Context(ctx).
		Do()
	if err != nil {
		return "", fmt.Errorf("list calendar events: %w", err)
	}

	for _, event := range events.Items {
		// Skip declined events and all-day events that aren't today
		if event.Status == "cancelled" || event.Transparency == "transparent" {
			continue
		}

		// Check if the event is actually happening now or starting soon
		if event.Start != nil {
			eventStart, err := time.Parse(time.RFC3339, event.Start.DateTime)
			if err != nil {
				continue
			}
			if eventStart.Before(endTime) {
				slog.Debug("found current/upcoming event",
					"summary", event.Summary,
					"start", event.Start.DateTime,
				)
				return event.Summary, nil
			}
		}
	}

	return "", nil
}
