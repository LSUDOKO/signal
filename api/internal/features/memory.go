package features

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/LSUDOKOS/signal/internal/store"
)

const (
	memoryPrefix  = "user:memory:"
	memoryMaxSize = 10
	memoryTTL     = 7 * 24 * time.Hour // 1 week
)

// MemoryEntry is one item in the user's AI memory.
type MemoryEntry struct {
	Text      string    `json:"text"`
	Timestamp time.Time `json:"ts"`
}

// MemoryService persists short-term interaction context per user in Redis.
// It stores up to memoryMaxSize recent interactions and surfaces them as
// context to AI prompts, making all features more personalized over time.
type MemoryService struct {
	cache store.CacheClient
}

// NewMemoryService creates a new MemoryService.
func NewMemoryService(cache store.CacheClient) *MemoryService {
	return &MemoryService{cache: cache}
}

// AddInteraction stores a new interaction for the user, evicting the oldest if full.
func (m *MemoryService) AddInteraction(ctx context.Context, slackUserID, text string) {
	entries := m.loadEntries(ctx, slackUserID)

	// Prepend new entry
	entries = append([]MemoryEntry{{Text: text, Timestamp: time.Now()}}, entries...)

	// Keep only the latest memoryMaxSize entries
	if len(entries) > memoryMaxSize {
		entries = entries[:memoryMaxSize]
	}

	m.saveEntries(ctx, slackUserID, entries)
}

// GetContext returns a formatted string of recent interactions for AI context injection.
// Returns empty string if no memory exists.
func (m *MemoryService) GetContext(ctx context.Context, slackUserID string) string {
	entries := m.loadEntries(ctx, slackUserID)
	if len(entries) == 0 {
		return ""
	}

	var lines []string
	for _, e := range entries {
		lines = append(lines, fmt.Sprintf("- [%s] %s", e.Timestamp.Format("Jan 2 15:04"), e.Text))
	}
	return "Recent user activity:\n" + strings.Join(lines, "\n")
}

// ClearMemory wipes all stored interactions for a user.
func (m *MemoryService) ClearMemory(ctx context.Context, slackUserID string) {
	entries := []MemoryEntry{}
	m.saveEntries(ctx, slackUserID, entries)
}

// GetGitHubUser returns the stored GitHub username for a Slack user (if set via /github link).
func (m *MemoryService) GetGitHubUser(ctx context.Context, slackUserID string) string {
	type sessionGetter interface {
		GetSession(ctx context.Context, key string) (string, error)
	}
	if sg, ok := m.cache.(sessionGetter); ok {
		val, err := sg.GetSession(ctx, "github:user:"+slackUserID)
		if err == nil {
			return val
		}
	}
	return ""
}

// SetGitHubUser links a GitHub username to a Slack user.
func (m *MemoryService) SetGitHubUser(ctx context.Context, slackUserID, githubUsername string) {
	type sessionSetter interface {
		SetSession(ctx context.Context, key string, value string, ttl time.Duration) error
	}
	if ss, ok := m.cache.(sessionSetter); ok {
		_ = ss.SetSession(ctx, "github:user:"+slackUserID, githubUsername, 30*24*time.Hour)
	}
}

func (m *MemoryService) loadEntries(ctx context.Context, slackUserID string) []MemoryEntry {
	key := memoryPrefix + slackUserID
	// Access the raw Redis client via the CacheClient interface
	// We use a type assertion to access the Raw() method on the concrete redis.Client
	type rawRedis interface {
		GetRaw(ctx context.Context, key string) (string, error)
	}

	// Use the cache's CheckAILimit as a proxy for raw access — instead, we store
	// in the session store which supports arbitrary string values
	// We use the session store (SetSession/GetSession) with a memory-specific key
	type sessionGetter interface {
		GetSession(ctx context.Context, key string) (string, error)
	}

	if sg, ok := m.cache.(sessionGetter); ok {
		val, err := sg.GetSession(ctx, key)
		if err != nil || val == "" {
			return nil
		}
		var entries []MemoryEntry
		if err := json.Unmarshal([]byte(val), &entries); err != nil {
			return nil
		}
		return entries
	}
	return nil
}

func (m *MemoryService) saveEntries(ctx context.Context, slackUserID string, entries []MemoryEntry) {
	key := memoryPrefix + slackUserID
	data, err := json.Marshal(entries)
	if err != nil {
		slog.Error("memory: failed to marshal entries", "error", err)
		return
	}

	type sessionSetter interface {
		SetSession(ctx context.Context, key string, value string, ttl time.Duration) error
	}

	if ss, ok := m.cache.(sessionSetter); ok {
		if err := ss.SetSession(ctx, key, string(data), memoryTTL); err != nil {
			slog.Error("memory: failed to save entries", "error", err)
		}
	}
}
