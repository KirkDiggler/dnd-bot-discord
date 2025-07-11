package character

import (
	"context"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
)

// buildWizardSteps creates wizard-specific steps
func (b *FlowBuilderImpl) buildWizardSteps(ctx context.Context, char *character.Character) []character.CreationStep {
	var steps []character.CreationStep

	// Cantrip selection (3 cantrips at level 1)
	cantripStep := character.CreationStep{
		Type:        character.StepTypeCantripsSelection,
		Title:       "Choose Your Cantrips",
		Description: "Cantrips are simple spells you can cast at will. Choose 3 cantrips from the wizard spell list.",
		MinChoices:  3,
		MaxChoices:  3,
		Required:    true,
		UIHints: &character.StepUIHints{
			Actions: []character.StepAction{
				{
					ID:    "open_spell_selection",
					Label: "Browse Cantrips",
					Style: "primary",
					Icon:  "✨",
				},
				{
					ID:    "suggested_cantrips",
					Label: "Use Suggested",
					Style: "secondary",
					Icon:  "💡",
				},
			},
			Layout:          "grid",
			ShowRecommended: true,
			Color:           0x6B46C1, // Purple for arcane
		},
	}

	// Fetch wizard cantrips from D&D API
	if b.dndClient != nil {
		spells, err := b.dndClient.ListSpellsByClassAndLevel("wizard", 0) // 0 = cantrips
		if err == nil && len(spells) > 0 {
			var cantripOptions []character.CreationOption
			for _, spell := range spells {
				cantripOptions = append(cantripOptions, character.CreationOption{
					Key:         spell.Key,
					Name:        spell.Name,
					Description: "Cantrip", // Simple description for now
				})
			}
			cantripStep.Options = cantripOptions
		}
	}

	// Spell selection (6 1st-level spells for spellbook)
	spellStep := character.CreationStep{
		Type:        character.StepTypeSpellbookSelection,
		Title:       "Fill Your Spellbook",
		Description: "Your spellbook contains all the spells you know. Choose 6 1st-level spells to start with. Your Intelligence modifier (if positive) grants additional spells.",
		MinChoices:  6,
		MaxChoices:  6,
		Required:    true,
		UIHints: &character.StepUIHints{
			Actions: []character.StepAction{
				{
					ID:    "open_spell_selection",
					Label: "Browse Spell List",
					Style: "primary",
					Icon:  "📜",
				},
				{
					ID:          "quick_build",
					Label:       "Quick Build",
					Style:       "secondary",
					Icon:        "⚡",
					Description: "Recommended spell selection",
				},
			},
			Layout:          "list",
			ShowProgress:    true,
			ProgressFormat:  "%d/%d spells selected",
			ShowRecommended: true,
			Color:           0x6B46C1,
		},
	}

	// Fetch wizard 1st level spells from D&D API
	if b.dndClient != nil {
		spells, err := b.dndClient.ListSpellsByClassAndLevel("wizard", 1) // 1 = 1st level spells
		if err == nil && len(spells) > 0 {
			var spellOptions []character.CreationOption
			for _, spell := range spells {
				spellOptions = append(spellOptions, character.CreationOption{
					Key:         spell.Key,
					Name:        spell.Name,
					Description: "1st level spell", // Simple description for now
				})
			}
			spellStep.Options = spellOptions
		}
	}

	steps = append(steps, cantripStep, spellStep)

	// Note: Arcane tradition selection happens at level 2
	if char.Level >= 2 {
		steps = append(steps, character.CreationStep{
			Type:        character.StepTypeSubclassSelection,
			Title:       "Choose Your Arcane Tradition",
			Description: "At 2nd level, you choose an arcane tradition, shaping your practice of magic.",
			MinChoices:  1,
			MaxChoices:  1,
			Required:    true,
			UIHints: &character.StepUIHints{
				Actions: []character.StepAction{
					{
						ID:    "select_tradition",
						Label: "Choose Tradition",
						Style: "primary",
						Icon:  "🔮",
					},
				},
				Layout: "grid",
				Color:  0x6B46C1,
			},
		})
	}

	return steps
}
