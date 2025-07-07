package character

import (
	"context"
	"testing"

	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/rulebook/dnd5e"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/shared"
)

// TestFlowWithRaceAndClassOptions tests that the flow builder includes race and class options
func TestFlowWithRaceAndClassOptions(t *testing.T) {
	// Create a new character with no race or class
	char := &character.Character{
		ID:      "test-new-char",
		OwnerID: "user123",
		Status:  shared.CharacterStatusDraft,
	}

	// Create flow builder with nil client (will use fallback)
	flowBuilder := NewFlowBuilder(nil)

	// Build the flow
	ctx := context.Background()
	flow, err := flowBuilder.BuildFlow(ctx, char)
	if err != nil {
		t.Fatalf("Failed to build flow: %v", err)
	}

	// Should have at least race and class steps
	if len(flow.Steps) < 2 {
		t.Fatalf("Expected at least 2 steps, got %d", len(flow.Steps))
	}

	// First step should be race selection
	if flow.Steps[0].Type != character.StepTypeRaceSelection {
		t.Errorf("Expected first step to be race selection, got %s", flow.Steps[0].Type)
	}

	// Second step should be class selection
	if flow.Steps[1].Type != character.StepTypeClassSelection {
		t.Errorf("Expected second step to be class selection, got %s", flow.Steps[1].Type)
	}

	// No other steps should be present yet
	if len(flow.Steps) > 2 {
		t.Errorf("Expected only 2 steps for new character, got %d", len(flow.Steps))
	}
}

// TestFlowProgressionAfterRaceAndClass tests that ability scores appear after race and class are selected
func TestFlowProgressionAfterRaceAndClass(t *testing.T) {
	// Create a character with race and class selected
	char := &character.Character{
		ID:      "test-char-with-race-class",
		OwnerID: "user123",
		Status:  shared.CharacterStatusDraft,
		Race:    &rulebook.Race{Key: "human", Name: "Human"},
		Class:   &rulebook.Class{Key: "fighter", Name: "Fighter"},
	}

	flowBuilder := NewFlowBuilder(nil)
	ctx := context.Background()
	flow, err := flowBuilder.BuildFlow(ctx, char)
	if err != nil {
		t.Fatalf("Failed to build flow: %v", err)
	}

	// Should have ability scores and other steps now
	foundAbilityScores := false
	foundFightingStyle := false

	for _, step := range flow.Steps {
		switch step.Type {
		case character.StepTypeAbilityScores:
			foundAbilityScores = true
		case character.StepTypeFightingStyleSelection:
			foundFightingStyle = true
		}
	}

	if !foundAbilityScores {
		t.Error("Expected ability scores step after race and class selection")
	}

	if !foundFightingStyle {
		t.Error("Expected fighting style step for fighter")
	}
}

// TestRaceSelectionProcessing tests that race selection is properly processed
func TestRaceSelectionProcessing(t *testing.T) {
	// Test the step result processing
	step := character.CreationStep{
		Type:       character.StepTypeRaceSelection,
		MinChoices: 1,
		MaxChoices: 1,
		Required:   true,
	}

	// Valid race selection
	result := &character.CreationStepResult{
		StepType:   character.StepTypeRaceSelection,
		Selections: []string{"elf"},
	}

	if !step.IsComplete(result) {
		t.Error("Valid race selection should mark step as complete")
	}

	// No selection
	emptyResult := &character.CreationStepResult{
		StepType:   character.StepTypeRaceSelection,
		Selections: []string{},
	}

	if step.IsComplete(emptyResult) {
		t.Error("Empty race selection should not mark step as complete")
	}
}
