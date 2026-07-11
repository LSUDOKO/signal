package ai

import (
	"testing"
)

// TestParseFocusResponse_DecisionTree verifies focus response parsing with decisions.
func TestParseFocusResponse_DecisionTree_WithDecisions(t *testing.T) {
	input := `✅ Moved Q3 budget to cloud servers
   ↳ Cost breakdown due Friday — Owner: @sarah — Due: Friday
   ↳ Implementation plan — Owner: @mike — Due: Next Monday
✅ Approved new design system
   ↳ No action needed — Owner: Unassigned — Due: None`

	result := parseFocusResponse(input)
	if result.NoDecisions {
		t.Error("should have found decisions")
	}
	if len(result.Decisions) != 2 {
		t.Errorf("expected 2 decisions, got %d", len(result.Decisions))
	}
	if len(result.Decisions) > 0 && result.Decisions[0].Decision != "Moved Q3 budget to cloud servers" {
		t.Errorf("decision 1 = %q, want %q", result.Decisions[0].Decision, "Moved Q3 budget to cloud servers")
	}
	if len(result.Decisions) > 0 && len(result.Decisions[0].ActionItems) != 2 {
		t.Errorf("expected 2 action items, got %d", len(result.Decisions[0].ActionItems))
	}
}

// TestParseFocusResponse_NoDecisions verifies focus response parsing without decisions.
func TestParseFocusResponse_NoDecisions(t *testing.T) {
	input := "No decisions found. Discussion was exploratory."

	result := parseFocusResponse(input)
	if !result.NoDecisions {
		t.Error("should have NoDecisions = true")
	}
	if len(result.Decisions) != 0 {
		t.Errorf("should have 0 decisions, got %d", len(result.Decisions))
	}
}

// TestParseFocusResponse_Empty verifies empty response handling.
func TestParseFocusResponse_Empty(t *testing.T) {
	result := parseFocusResponse("")
	if result == nil {
		t.Fatal("parseFocusResponse should not return nil")
	}
}

// TestParseToneResponse_FullAnalysis verifies tone analysis parsing.
func TestParseToneResponse_Full(t *testing.T) {
	input := `- Tone: Frustrated
- Intent: They want you to complete the task immediately
- Action: Reply with status update and estimated completion time
- Note: The sender has been waiting longer than expected. A quick update will reduce tension.`

	result := parseToneResponse(input)
	if result.Tone != "Frustrated" {
		t.Errorf("Tone = %q, want 'Frustrated'", result.Tone)
	}
	if result.Intent != "They want you to complete the task immediately" {
		t.Errorf("Intent = %q, want 'They want you to complete the task immediately'", result.Intent)
	}
	if result.Action != "Reply with status update and estimated completion time" {
		t.Errorf("Action = %q, want 'Reply with status update...'", result.Action)
	}
	if result.Note != "The sender has been waiting longer than expected. A quick update will reduce tension." {
		t.Errorf("Note = %q, want 'The sender has been waiting...'", result.Note)
	}
}

// TestParseToneResponse_Neutral verifies neutral tone parsing.
func TestParseToneResponse_Neutral(t *testing.T) {
	input := `- Tone: Neutral
- Intent: They are sharing information
- Action: Acknowledge receipt
- Note: No hidden meaning. This is a straightforward information update.`

	result := parseToneResponse(input)
	if result.Tone != "Neutral" {
		t.Errorf("Tone = %q, want 'Neutral'", result.Tone)
	}
}

// TestParseToneResponse_Empty verifies empty response handling.
func TestParseToneResponse_Empty(t *testing.T) {
	result := parseToneResponse("")
	if result == nil {
		t.Fatal("parseToneResponse should not return nil")
	}
	if result.Tone != "" {
		t.Errorf("empty response should have empty tone, got %q", result.Tone)
	}
}

// TestBuildFocusPrompt verifies focus prompt building.
func TestBuildFocusPrompt(t *testing.T) {
	messages := []string{"Decision: Use cloud servers", "Action item: Sarah to provide cost breakdown"}
	prompt := buildFocusPrompt(messages)
	if prompt == "" {
		t.Error("buildFocusPrompt should return a non-empty prompt")
	}
	if !contains(prompt, "cloud servers") {
		t.Error("prompt should contain message content")
	}
}

// TestBuildCatchUpPrompt verifies catchup prompt building.
func TestBuildCatchUpPrompt(t *testing.T) {
	messages := []string{"We decided to move to AWS", "Q3 budget approved at $50K"}
	prompt := buildCatchUpPrompt(messages)
	if prompt == "" {
		t.Error("buildCatchUpPrompt should return a non-empty prompt")
	}
	if !contains(prompt, "AWS") {
		t.Error("prompt should contain message content")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
