package character

import (
	"context"
	"fmt"

	"github.com/KirkDiggler/dnd-bot-discord/internal/clients/dnd5e"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/rulebook/dnd5e"
)

// FlowBuilderImpl implements the FlowBuilder interface
type FlowBuilderImpl struct {
	dndClient dnd5e.Client
}

// NewFlowBuilder creates a new flow builder
func NewFlowBuilder(dndClient dnd5e.Client) character.FlowBuilder {
	return &FlowBuilderImpl{
		dndClient: dndClient,
	}
}

// BuildFlow creates a complete character creation flow based on character state
func (b *FlowBuilderImpl) BuildFlow(ctx context.Context, char *character.Character) (*character.CreationFlow, error) {
	var steps []character.CreationStep

	// Base steps for all characters
	steps = append(steps,
		// 1. Race Selection
		character.CreationStep{
			Type:        character.StepTypeRaceSelection,
			Title:       "Choose Your Race",
			Description: "Select your character's race, which determines starting abilities and traits.",
			Required:    true,
		},
		// 2. Class Selection
		character.CreationStep{
			Type:        character.StepTypeClassSelection,
			Title:       "Choose Your Class",
			Description: "Select your character's class, which determines abilities, proficiencies, and features.",
			Required:    true,
		},
		// 3. Ability Scores
		character.CreationStep{
			Type:        character.StepTypeAbilityScores,
			Title:       "Roll Ability Scores",
			Description: "Generate your character's six ability scores.",
			Required:    true,
		},
		// 4. Ability Assignment
		character.CreationStep{
			Type:        character.StepTypeAbilityAssignment,
			Title:       "Assign Ability Scores",
			Description: "Assign your rolled scores to the six abilities.",
			Required:    true,
		},
	)

	// 5. Class-specific features (dynamic based on class)
	if char.Class != nil {
		classSteps, err := b.buildClassSpecificSteps(ctx, char)
		if err != nil {
			return nil, fmt.Errorf("failed to build class steps: %w", err)
		}
		steps = append(steps, classSteps...)
	}

	// Final steps for all characters
	steps = append(steps,
		// 6. Proficiency Selection
		character.CreationStep{
			Type:        character.StepTypeProficiencySelection,
			Title:       "Choose Proficiencies",
			Description: "Select your character's skill and tool proficiencies.",
			Required:    true,
		},
		// 7. Equipment Selection
		character.CreationStep{
			Type:        character.StepTypeEquipmentSelection,
			Title:       "Choose Equipment",
			Description: "Select your starting equipment and gear.",
			Required:    true,
		},
		// 8. Character Details
		character.CreationStep{
			Type:        character.StepTypeCharacterDetails,
			Title:       "Character Details",
			Description: "Choose your character's name and other details.",
			Required:    true,
		},
	)

	return &character.CreationFlow{
		Steps: steps,
	}, nil
}

// buildClassSpecificSteps creates steps specific to the character's class
func (b *FlowBuilderImpl) buildClassSpecificSteps(ctx context.Context, char *character.Character) ([]character.CreationStep, error) {
	var steps []character.CreationStep

	switch char.Class.Key {
	case "cleric":
		steps = append(steps, b.buildClericSteps(ctx, char)...)
	case "fighter":
		steps = append(steps, b.buildFighterSteps(ctx, char)...)
	case "ranger":
		steps = append(steps, b.buildRangerSteps(ctx, char)...)
		// Add other classes as needed
	}

	return steps, nil
}

// buildClericSteps creates cleric-specific steps
func (b *FlowBuilderImpl) buildClericSteps(ctx context.Context, char *character.Character) []character.CreationStep {
	var steps []character.CreationStep

	// Divine Domain Selection
	domainChoice := rulebook.GetDivineDomainChoice()
	var domainOptions []character.CreationOption
	for _, option := range domainChoice.Options {
		domainOptions = append(domainOptions, character.CreationOption{
			Key:         option.Key,
			Name:        option.Name,
			Description: option.Description,
			Metadata:    option.Metadata,
		})
	}

	steps = append(steps, character.CreationStep{
		Type:        character.StepTypeDivineDomainSelection,
		Title:       "Choose Your Divine Domain",
		Description: "Choose one domain related to your deity. Your choice grants you domain spells and other features.",
		Options:     domainOptions,
		MinChoices:  1,
		MaxChoices:  1,
		Required:    true,
	})

	// Knowledge Domain specific steps
	if b.hasSelectedDomain(char, "knowledge") {
		// Skill Selection
		skillOptions := []character.CreationOption{
			{Key: "arcana", Name: "Arcana", Description: "Your knowledge of magic and magical theory"},
			{Key: "history", Name: "History", Description: "Your knowledge of historical events and lore"},
			{Key: "nature", Name: "Nature", Description: "Your knowledge of the natural world"},
			{Key: "religion", Name: "Religion", Description: "Your knowledge of deities and religious practices"},
		}

		steps = append(steps, character.CreationStep{
			Type:        character.StepTypeSkillSelection,
			Title:       "Choose Knowledge Domain Skills",
			Description: "As a Knowledge domain cleric, choose 2 additional skill proficiencies.",
			Options:     skillOptions,
			MinChoices:  2,
			MaxChoices:  2,
			Required:    true,
			Context: map[string]any{
				"source": "knowledge_domain",
			},
		})

		// Language Selection (simplified - would need full language list)
		languageOptions := []character.CreationOption{
			{Key: "draconic", Name: "Draconic", Description: "The language of dragons and dragonborn"},
			{Key: "elvish", Name: "Elvish", Description: "The language of elves"},
			{Key: "dwarvish", Name: "Dwarvish", Description: "The language of dwarves"},
			{Key: "celestial", Name: "Celestial", Description: "The language of celestials"},
			{Key: "abyssal", Name: "Abyssal", Description: "The language of demons"},
			{Key: "infernal", Name: "Infernal", Description: "The language of devils"},
		}

		steps = append(steps, character.CreationStep{
			Type:        character.StepTypeLanguageSelection,
			Title:       "Choose Knowledge Domain Languages",
			Description: "As a Knowledge domain cleric, choose 2 additional languages.",
			Options:     languageOptions,
			MinChoices:  2,
			MaxChoices:  2,
			Required:    true,
			Context: map[string]any{
				"source": "knowledge_domain",
			},
		})
	}

	return steps
}

// buildFighterSteps creates fighter-specific steps
func (b *FlowBuilderImpl) buildFighterSteps(ctx context.Context, char *character.Character) []character.CreationStep {
	var steps []character.CreationStep

	// Fighting Style Selection
	fightingStyleChoice := rulebook.GetFightingStyleChoice("fighter")
	var styleOptions []character.CreationOption
	for _, option := range fightingStyleChoice.Options {
		styleOptions = append(styleOptions, character.CreationOption{
			Key:         option.Key,
			Name:        option.Name,
			Description: option.Description,
			Metadata:    option.Metadata,
		})
	}

	steps = append(steps, character.CreationStep{
		Type:        character.StepTypeFightingStyleSelection,
		Title:       "Choose Your Fighting Style",
		Description: "Choose a fighting style that defines your combat technique.",
		Options:     styleOptions,
		MinChoices:  1,
		MaxChoices:  1,
		Required:    true,
	})

	return steps
}

// buildRangerSteps creates ranger-specific steps
func (b *FlowBuilderImpl) buildRangerSteps(ctx context.Context, char *character.Character) []character.CreationStep {
	var steps []character.CreationStep

	// Favored Enemy Selection
	favoredEnemyChoice := rulebook.GetFavoredEnemyChoice()
	var enemyOptions []character.CreationOption
	for _, option := range favoredEnemyChoice.Options {
		enemyOptions = append(enemyOptions, character.CreationOption{
			Key:         option.Key,
			Name:        option.Name,
			Description: option.Description,
			Metadata:    option.Metadata,
		})
	}

	steps = append(steps, character.CreationStep{
		Type:        character.StepTypeFavoredEnemySelection,
		Title:       "Choose Your Favored Enemy",
		Description: "Choose the type of creature you have dedicated yourself to hunting.",
		Options:     enemyOptions,
		MinChoices:  1,
		MaxChoices:  1,
		Required:    true,
	})

	// Natural Explorer Selection
	naturalExplorerChoice := rulebook.GetNaturalExplorerChoice()
	var terrainOptions []character.CreationOption
	for _, option := range naturalExplorerChoice.Options {
		terrainOptions = append(terrainOptions, character.CreationOption{
			Key:         option.Key,
			Name:        option.Name,
			Description: option.Description,
			Metadata:    option.Metadata,
		})
	}

	steps = append(steps, character.CreationStep{
		Type:        character.StepTypeNaturalExplorerSelection,
		Title:       "Choose Your Favored Terrain",
		Description: "Choose the terrain where you feel most at home.",
		Options:     terrainOptions,
		MinChoices:  1,
		MaxChoices:  1,
		Required:    true,
	})

	return steps
}

// hasSelectedDomain checks if the character has selected a specific divine domain
func (b *FlowBuilderImpl) hasSelectedDomain(char *character.Character, domain string) bool {
	for _, feature := range char.Features {
		if feature.Key == "divine_domain" && feature.Metadata != nil {
			if selectedDomain, ok := feature.Metadata["domain"].(string); ok {
				return selectedDomain == domain
			}
		}
	}
	return false
}
