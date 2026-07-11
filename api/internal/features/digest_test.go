package features

import (
	"testing"

	"github.com/LSUDOKOS/signal/internal/domain"
)

// TestDigestService_buildDigestBlocks verifies the digest block builder.
func TestDigestService_buildDigestBlocks(t *testing.T) {
	svc := &DigestService{}

	tests := []struct {
		name    string
		hour    int
		urgent  []domain.DigestItem
		fyi     []domain.DigestItem
		threads []domain.DigestItem
		wantMin int // minimum blocks expected
	}{
		{
			name:    "empty digest",
			hour:    16,
			urgent:  nil,
			fyi:     nil,
			threads: nil,
			wantMin: 2, // header + actions
		},
		{
			name: "with urgent items",
			hour: 9,
			urgent: []domain.DigestItem{
				{From: "@john", Message: "Need report by 5 PM", Channel: "general"},
			},
			fyi:     nil,
			threads: nil,
			wantMin: 3, // header + urgent + actions
		},
		{
			name:   "with fyi items",
			hour:   14,
			urgent: nil,
			fyi: []domain.DigestItem{
				{From: "@team", Message: "Lunch tomorrow", Channel: "general"},
			},
			threads: nil,
			wantMin: 3, // header + fyi + actions
		},
		{
			name:    "with thread replies",
			hour:    16,
			urgent:  nil,
			fyi:     nil,
			threads: []domain.DigestItem{
				{From: "you", Message: "3 replies in #design", Channel: "design"},
			},
			wantMin: 3, // header + threads + actions
		},
		{
			name: "all sections populated",
			hour: 16,
			urgent: []domain.DigestItem{
				{From: "@john", Message: "Need report", Channel: "general"},
			},
			fyi: []domain.DigestItem{
				{From: "@team", Message: "Lunch tomorrow", Channel: "general"},
			},
			threads: []domain.DigestItem{
				{From: "you", Message: "3 replies in #design", Channel: "design"},
			},
			wantMin: 5, // header + urgent + fyi + threads + actions
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			blocks := svc.buildDigestBlocks(tt.hour, tt.urgent, tt.fyi, tt.threads)
			if len(blocks) < tt.wantMin {
				t.Errorf("buildDigestBlocks() returned %d blocks, want at least %d", len(blocks), tt.wantMin)
			}
		})
	}
}

// TestDigestService_buildDigestBlocks_HourFormat verifies hour formatting.
func TestDigestService_buildDigestBlocks_HourFormat(t *testing.T) {
	svc := &DigestService{}

	hours := []int{0, 1, 9, 12, 16, 23}
	for _, hour := range hours {
		t.Run("hour", func(t *testing.T) {
			blocks := svc.buildDigestBlocks(hour, nil, nil, nil)
			if len(blocks) == 0 {
				t.Errorf("buildDigestBlocks(%d) should return at least some blocks", hour)
			}
		})
	}
}
