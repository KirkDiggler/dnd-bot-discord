package character

import (
	"context"
	"testing"

	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/rulebook/dnd5e"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/shared"
)

// TestKnowledgeDomainCreationFlow tests the complete creation flow for a Knowledge Domain cleric
func TestKnowledgeDomainCreationFlow(t *testing.T) {
	// Create test character with Knowledge Domain
	char := &character.Character{
		ID:      "test-knowledge-cleric",
		OwnerID: "user123",
		Name:    "Theron",
		Status:  shared.CharacterStatusDraft,
		Race:    &rulebook.Race{Key: "human", Name: "Human"},
		Class:   &rulebook.Class{Key: "cleric", Name: "Cleric"},
		Level:   1,
		Attributes: map[shared.Attribute]*character.AbilityScore{
			shared.AttributeWisdom: {Score: 16, Bonus: 3},
		},
		Features: []*rulebook.CharacterFeature{
			{
				Key:  "divine_domain",
				Name: "Divine Domain",
				Type: rulebook.FeatureTypeClass,
				Metadata: map[string]any{
					"domain":            "knowledge",
					"selection_display": "Knowledge Domain",
				},
			},
		},
	}

	// Create flow builder (nil client is ok for this test since we're not fetching data)
	flowBuilder := NewFlowBuilder(nil)

	// Build the flow
	ctx := context.Background()
	flow, err := flowBuilder.BuildFlow(ctx, char)
	if err != nil {
		t.Fatalf("Failed to build flow: %v", err)
	}

	// Find Knowledge Domain specific steps
	var foundSkillStep, foundLanguageStep bool
	for _, step := range flow.Steps {
		switch step.Type {
		case character.StepTypeSkillSelection:
			foundSkillStep = true
			// Verify skill step details
			if step.Title != "Choose Knowledge Domain Skills" {
				t.Errorf("Expected skill step title 'Choose Knowledge Domain Skills', got: %s", step.Title)
			}
			if step.MinChoices != 2 || step.MaxChoices != 2 {
				t.Errorf("Expected 2 skill choices, got min=%d, max=%d", step.MinChoices, step.MaxChoices)
			}

		case character.StepTypeLanguageSelection:
			foundLanguageStep = true
			// Verify language step details
			if step.Title != "Choose Knowledge Domain Languages" {
				t.Errorf("Expected language step title 'Choose Knowledge Domain Languages', got: %s", step.Title)
			}
			if step.MinChoices != 2 || step.MaxChoices != 2 {
				t.Errorf("Expected 2 language choices, got min=%d, max=%d", step.MinChoices, step.MaxChoices)
			}
		}
	}

	if !foundSkillStep {
		t.Error("Knowledge Domain cleric should have skill selection step")
	}
	if !foundLanguageStep {
		t.Error("Knowledge Domain cleric should have language selection step")
	}

	// Test that the steps are in correct order
	var divineDomainIndex, skillIndex, languageIndex int
	for i, step := range flow.Steps {
		switch step.Type {
		case character.StepTypeDivineDomainSelection:
			divineDomainIndex = i
		case character.StepTypeSkillSelection:
			skillIndex = i
		case character.StepTypeLanguageSelection:
			languageIndex = i
		}
	}

	// Skills and languages should come after divine domain selection
	if skillIndex <= divineDomainIndex {
		t.Error("Skill selection should come after divine domain selection")
	}
	if languageIndex <= divineDomainIndex {
		t.Error("Language selection should come after divine domain selection")
	}
	if languageIndex <= skillIndex {
		t.Error("Language selection should come after skill selection")
	}

	t.Logf("Knowledge Domain flow created successfully with %d steps", len(flow.Steps))
}

// TestProcessKnowledgeDomainSelections tests processing skill and language selections
func TestProcessKnowledgeDomainSelections(t *testing.T) {
	// Test skill selection processing
	skillStep := character.CreationStep{
		Type:       character.StepTypeSkillSelection,
		MinChoices: 2,
		MaxChoices: 2,
		Required:   true,
	}

	// Valid skill selection
	validSkillResult := &character.CreationStepResult{
		StepType:   character.StepTypeSkillSelection,
		Selections: []string{"arcana", "history"},
	}

	if !skillStep.IsComplete(validSkillResult) {
		t.Error("Valid skill selection should be complete")
	}

	// Invalid - too few selections
	invalidSkillResult := &character.CreationStepResult{
		StepType:   character.StepTypeSkillSelection,
		Selections: []string{"arcana"},
	}

	if skillStep.IsComplete(invalidSkillResult) {
		t.Error("Single skill selection should not be complete when 2 are required")
	}

	// Test language selection processing
	languageStep := character.CreationStep{
		Type:       character.StepTypeLanguageSelection,
		MinChoices: 2,
		MaxChoices: 2,
		Required:   true,
	}

	// Valid language selection
	validLanguageResult := &character.CreationStepResult{
		StepType:   character.StepTypeLanguageSelection,
		Selections: []string{"draconic", "celestial"},
	}

	if !languageStep.IsComplete(validLanguageResult) {
		t.Error("Valid language selection should be complete")
	}
}
