package character

import (
	"context"
	"testing"

	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/rulebook/dnd5e"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/shared"
)

// TestKnowledgeDomainStepGeneration tests that Knowledge Domain clerics get the right steps
func TestKnowledgeDomainStepGeneration(t *testing.T) {
	// Create a Knowledge Domain cleric
	char := &character.Character{
		ID:    "test-cleric",
		Race:  &rulebook.Race{Key: "human", Name: "Human"},
		Class: &rulebook.Class{Key: "cleric", Name: "Cleric"},
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

	// Create test flow builder
	flowBuilder := &SimpleTestFlowBuilder{}

	// Build flow for this character
	flow, err := flowBuilder.BuildFlow(context.Background(), char)
	if err != nil {
		t.Fatalf("Failed to build flow: %v", err)
	}

	// Verify Knowledge Domain specific steps are included
	foundSkillStep := false
	foundLanguageStep := false

	for _, step := range flow.Steps {
		switch step.Type {
		case character.StepTypeSkillSelection:
			foundSkillStep = true
			// Verify skill step details
			if step.Title != "Choose Knowledge Domain Skills" {
				t.Errorf("Expected skill step title 'Choose Knowledge Domain Skills', got: %s", step.Title)
			}
			if step.MinChoices != 2 || step.MaxChoices != 2 {
				t.Errorf("Expected skill step min=2, max=2, got min=%d, max=%d", step.MinChoices, step.MaxChoices)
			}
			if len(step.Options) != 4 {
				t.Errorf("Expected 4 skill options, got %d", len(step.Options))
			}

		case character.StepTypeLanguageSelection:
			foundLanguageStep = true
			// Verify language step details
			if step.Title != "Choose Knowledge Domain Languages" {
				t.Errorf("Expected language step title 'Choose Knowledge Domain Languages', got: %s", step.Title)
			}
			if step.MinChoices != 2 || step.MaxChoices != 2 {
				t.Errorf("Expected language step min=2, max=2, got min=%d, max=%d", step.MinChoices, step.MaxChoices)
			}
		}
	}

	if !foundSkillStep {
		t.Error("Knowledge Domain cleric should have skill selection step")
	}
	if !foundLanguageStep {
		t.Error("Knowledge Domain cleric should have language selection step")
	}
}

// TestOtherDomainNoExtraSteps tests that other domains don't get Knowledge-specific steps
func TestOtherDomainNoExtraSteps(t *testing.T) {
	// Create a Life Domain cleric
	char := &character.Character{
		ID:    "test-life-cleric",
		Race:  &rulebook.Race{Key: "human", Name: "Human"},
		Class: &rulebook.Class{Key: "cleric", Name: "Cleric"},
		Features: []*rulebook.CharacterFeature{
			{
				Key:  "divine_domain",
				Name: "Divine Domain",
				Type: rulebook.FeatureTypeClass,
				Metadata: map[string]any{
					"domain":            "life",
					"selection_display": "Life Domain",
				},
			},
		},
	}

	flowBuilder := &SimpleTestFlowBuilder{}
	flow, err := flowBuilder.BuildFlow(context.Background(), char)
	if err != nil {
		t.Fatalf("Failed to build flow: %v", err)
	}

	// Verify NO Knowledge Domain specific steps
	for _, step := range flow.Steps {
		if step.Type == character.StepTypeSkillSelection {
			t.Error("Life Domain cleric should not have skill selection step")
		}
		if step.Type == character.StepTypeLanguageSelection {
			t.Error("Life Domain cleric should not have language selection step")
		}
	}
}

// TestStepResultProcessing tests processing step results
func TestStepResultProcessing(t *testing.T) {
	// Test that CreationStepResult validation works
	step := character.CreationStep{
		Type:       character.StepTypeSkillSelection,
		MinChoices: 2,
		MaxChoices: 2,
		Required:   true,
	}

	// Valid result
	validResult := &character.CreationStepResult{
		StepType:   character.StepTypeSkillSelection,
		Selections: []string{"arcana", "history"},
	}

	if !step.IsComplete(validResult) {
		t.Error("Valid result should mark step as complete")
	}

	// Invalid result - wrong number of selections
	invalidResult := &character.CreationStepResult{
		StepType:   character.StepTypeSkillSelection,
		Selections: []string{"arcana"}, // Only 1 selection, need 2
	}

	if step.IsComplete(invalidResult) {
		t.Error("Invalid result (too few selections) should not mark step as complete")
	}

	// Invalid result - wrong step type
	wrongTypeResult := &character.CreationStepResult{
		StepType:   character.StepTypeLanguageSelection, // Wrong type
		Selections: []string{"arcana", "history"},
	}

	if step.IsComplete(wrongTypeResult) {
		t.Error("Invalid result (wrong type) should not mark step as complete")
	}
}

func TestCreationOptionGeneration(t *testing.T) {
	// Test that we can create proper creation options
	option := character.CreationOption{
		Key:         "arcana",
		Name:        "Arcana",
		Description: "Your knowledge of magic and magical theory",
		Metadata: map[string]any{
			"skill_type": "intelligence",
		},
	}

	if option.Key != "arcana" {
		t.Errorf("Expected key 'arcana', got %s", option.Key)
	}
	if option.Name != "Arcana" {
		t.Errorf("Expected name 'Arcana', got %s", option.Name)
	}
	if skillType, ok := option.Metadata["skill_type"].(string); !ok || skillType != "intelligence" {
		t.Errorf("Expected skill_type 'intelligence', got %v", skillType)
	}
}

// SimpleTestFlowBuilder creates flows for testing
type SimpleTestFlowBuilder struct{}

func (b *SimpleTestFlowBuilder) BuildFlow(ctx context.Context, char *character.Character) (*character.CreationFlow, error) {
	var steps []character.CreationStep

	// Basic steps
	steps = append(steps,
		character.CreationStep{Type: character.StepTypeRaceSelection, Required: true},
		character.CreationStep{Type: character.StepTypeClassSelection, Required: true},
		character.CreationStep{Type: character.StepTypeAbilityScores, Required: true},
		character.CreationStep{Type: character.StepTypeAbilityAssignment, Required: true},
	)

	// Add Knowledge Domain steps if applicable
	if char.Class != nil && char.Class.Key == "cleric" {
		// Divine Domain
		steps = append(steps, character.CreationStep{
			Type:     character.StepTypeDivineDomainSelection,
			Required: true,
		})

		// Check if Knowledge Domain is selected
		for _, feature := range char.Features {
			if feature.Key == "divine_domain" && feature.Metadata != nil {
				if domain, ok := feature.Metadata["domain"].(string); ok && domain == "knowledge" {
					// Add Knowledge Domain specific steps
					steps = append(steps,
						character.CreationStep{
							Type:       character.StepTypeSkillSelection,
							Title:      "Choose Knowledge Domain Skills",
							MinChoices: 2,
							MaxChoices: 2,
							Required:   true,
							Options: []character.CreationOption{
								{Key: "arcana", Name: "Arcana", Description: "Your knowledge of magic and magical theory"},
								{Key: "history", Name: "History", Description: "Your knowledge of historical events and lore"},
								{Key: "nature", Name: "Nature", Description: "Your knowledge of the natural world"},
								{Key: "religion", Name: "Religion", Description: "Your knowledge of deities and religious practices"},
							},
						},
						character.CreationStep{
							Type:       character.StepTypeLanguageSelection,
							Title:      "Choose Knowledge Domain Languages",
							MinChoices: 2,
							MaxChoices: 2,
							Required:   true,
							Options: []character.CreationOption{
								{Key: "draconic", Name: "Draconic", Description: "The language of dragons"},
								{Key: "celestial", Name: "Celestial", Description: "The language of celestials"},
								{Key: "abyssal", Name: "Abyssal", Description: "The language of demons"},
								{Key: "infernal", Name: "Infernal", Description: "The language of devils"},
							},
						},
					)
				}
			}
		}
	}

	// Final steps
	steps = append(steps,
		character.CreationStep{Type: character.StepTypeProficiencySelection, Required: true},
		character.CreationStep{Type: character.StepTypeEquipmentSelection, Required: true},
		character.CreationStep{Type: character.StepTypeCharacterDetails, Required: true},
	)

	return &character.CreationFlow{Steps: steps}, nil
}
