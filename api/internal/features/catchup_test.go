package features

import (
	"strings"
	"testing"

	"github.com/LSUDOKOS/signal/internal/domain"
)

// TestCatchUpService_SemanticQuery verifies the search query builder.
func TestCatchUpService_SemanticQuery(t *testing.T) {
	svc := &CatchUpService{}

	tests := []struct {
		name         string
		userID       string
		query        string
		daysBack     int
		wantPrefix   string
		wantContains string
	}{
		{
			name:         "budget query",
			userID:       "U12345",
			query:        "Q3 budget decision",
			daysBack:     7,
			wantPrefix:   "from:@U12345 OR to:@U12345 Q3 budget decision after:",
			wantContains: "U12345",
		},
		{
			name:         "design mockup query",
			userID:       "U67890",
			query:        "design mockup review",
			daysBack:     3,
			wantPrefix:   "from:@U67890 OR to:@U67890 design mockup review after:",
			wantContains: "U67890",
		},
		{
			name:         "empty query",
			userID:       "U12345",
			query:        "",
			daysBack:     1,
			wantPrefix:   "from:@U12345 OR to:@U12345  after:",
			wantContains: "U12345",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := svc.SemanticQuery(tt.userID, tt.query, tt.daysBack)
			if !strings.HasPrefix(got, tt.wantPrefix) {
				t.Errorf("SemanticQuery() = %q, want prefix %q", got, tt.wantPrefix)
			}
			if !strings.Contains(got, tt.wantContains) {
				t.Errorf("SemanticQuery() = %q, should contain %q", got, tt.wantContains)
			}
		})
	}
}

// TestCatchUpService_buildResultBlocks verifies the block kit builder.
func TestCatchUpService_buildResultBlocks(t *testing.T) {
	svc := &CatchUpService{}

	tests := []struct {
		name      string
		query     string
		result    *domain.CatchUpResult
		wantCount int // expected number of blocks
	}{
		{
			name:  "with results",
			query: "budget",
			result: &domain.CatchUpResult{
				Topics: []domain.CatchUpTopic{
					{
						Name:     "Q3 Budget",
						Decision: "Approved $50K for cloud infrastructure",
						Action:   "Submit cost breakdown by Friday",
					},
				},
				MessageCount: 5,
			},
			wantCount: 4, // header, section, divider, final section
		},
		{
			name:  "empty topics",
			query: "nothing",
			result: &domain.CatchUpResult{
				MessageCount: 0,
			},
			wantCount: 4, // still builds blocks with fallback text
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			blocks := svc.buildResultBlocks(tt.query, tt.result, "")
			if len(blocks) < tt.wantCount {
				t.Errorf("buildResultBlocks() returned %d blocks, want at least %d", len(blocks), tt.wantCount)
			}
		})
	}
}

// TestBuildSearchQuery_RTS verifies the RTS query builder works correctly.
func TestBuildSearchQuery_RTS(t *testing.T) {
	svc := &CatchUpService{}

	// Test that the query format is correct
	query := svc.SemanticQuery("U123", "test query", 7)
	if !strings.Contains(query, "from:@U123") {
		t.Errorf("query should contain from:@U123, got: %s", query)
	}
	if !strings.Contains(query, "test query") {
		t.Errorf("query should contain 'test query', got: %s", query)
	}
}
