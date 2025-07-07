package character

import (
	"testing"

	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
)

func TestCreationStep_IsComplete(t *testing.T) {
	step := &character.CreationStep{
		Type:       character.StepTypeSkillSelection,
		MinChoices: 2,
		MaxChoices: 2,
		Required:   true,
	}

	// Test with valid result
	result := &character.CreationStepResult{
		StepType:   character.StepTypeSkillSelection,
		Selections: []string{"arcana", "history"},
	}

	if !step.IsComplete(result) {
		t.Error("Expected step to be complete with 2 selections")
	}

	// Test with too few selections
	result.Selections = []string{"arcana"}
	if step.IsComplete(result) {
		t.Error("Expected step to be incomplete with 1 selection")
	}

	// Test with wrong step type
	result.StepType = character.StepTypeLanguageSelection
	result.Selections = []string{"arcana", "history"}
	if step.IsComplete(result) {
		t.Error("Expected step to be incomplete with wrong step type")
	}
}

func TestFlowBuilder_BuildBasicFlow(t *testing.T) {
	// Test that we can create the basic types without compilation errors
	step := character.CreationStep{
		Type:        character.StepTypeRaceSelection,
		Title:       "Choose Your Race",
		Description: "Select your character's race",
		Required:    true,
	}

	if step.Type != character.StepTypeRaceSelection {
		t.Errorf("Expected StepTypeRaceSelection, got %v", step.Type)
	}

	// Test CreationOption
	option := character.CreationOption{
		Key:         "human",
		Name:        "Human",
		Description: "Versatile and adaptable",
	}

	if option.Key != "human" {
		t.Errorf("Expected 'human', got %v", option.Key)
	}
}
