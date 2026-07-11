package rts

import (
	"strings"
	"testing"
)

// TestBuildSearchQuery verifies the search query builder.
func TestBuildSearchQuery(t *testing.T) {
	tests := []struct {
		name   string
		userID string
		query  string
		days   int
		check  func(string) bool
	}{
		{
			name:   "budget query",
			userID: "U12345",
			query:  "Q3 budget",
			days:   7,
			check: func(q string) bool {
				return strings.Contains(q, "from:@U12345") &&
					strings.Contains(q, "Q3 budget") &&
					strings.Contains(q, "after:")
			},
		},
		{
			name:   "empty query",
			userID: "U67890",
			query:  "",
			days:   1,
			check: func(q string) bool {
				return strings.Contains(q, "from:@U67890") &&
					strings.Contains(q, "after:")
			},
		},
		{
			name:   "zero days",
			userID: "U11111",
			query:  "urgent",
			days:   0,
			check: func(q string) bool {
				return strings.Contains(q, "from:@U11111") &&
					strings.Contains(q, "after:")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildSearchQuery(tt.userID, tt.query, tt.days)
			if !tt.check(got) {
				t.Errorf("BuildSearchQuery(%q, %q, %d) = %q, doesn't satisfy check", tt.userID, tt.query, tt.days, got)
			}
		})
	}
}

// TestBuildSearchQuery_DateFormat verifies the date filter is valid.
func TestBuildSearchQuery_DateFormat(t *testing.T) {
	query := BuildSearchQuery("U123", "test", 7)
	// Should contain "after:20" (current year prefix for the date)
	if !strings.Contains(query, "after:20") {
		t.Errorf("query should contain date after:20..., got: %s", query)
	}
	// Should be in YYYY-MM-DD format
	parts := strings.Split(query, "after:")
	if len(parts) > 1 {
		datePart := parts[len(parts)-1]
		if len(datePart) != 10 {
			t.Errorf("date part should be 10 chars (YYYY-MM-DD), got %q (len=%d)", datePart, len(datePart))
		}
	}
}
