package character

import (
	"context"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
)

// buildBardSteps creates bard-specific steps
func (b *FlowBuilderImpl) buildBardSteps(ctx context.Context, char *character.Character) []character.CreationStep {
	var steps []character.CreationStep

	// Expertise selection (2 skills)
	steps = append(steps, character.CreationStep{
		Type:        character.StepTypeExpertiseSelection,
		Title:       "Choose Your Expertise",
		Description: "Your proficiency bonus is doubled for ability checks using these skills. Choose 2 skills you are proficient with.",
		MinChoices:  2,
		MaxChoices:  2,
		Required:    true,
		UIHints: &character.StepUIHints{
			Actions: []character.StepAction{
				{
					ID:    "select_expertise",
					Label: "Choose Skills",
					Style: "primary",
					Icon:  "⭐",
				},
			},
			Layout:       "list",
			ShowProgress: true,
			Color:        0xA855F7, // Purple for bards
		},
	})

	// Cantrip selection (2 cantrips)
	cantripStep := character.CreationStep{
		Type:        character.StepTypeCantripsSelection,
		Title:       "Choose Your Cantrips",
		Description: "Cantrips are simple spells you can cast at will. Choose 2 cantrips from the bard spell list.",
		MinChoices:  2,
		MaxChoices:  2,
		Required:    true,
		UIHints: &character.StepUIHints{
			Layout:          "grid",
			ShowRecommended: true,
			Color:           0xA855F7,
		},
	}

	// Populate cantrips options from D&D API
	if cantrips, err := b.dndClient.ListSpellsByClassAndLevel("bard", 0); err == nil {
		for _, cantrip := range cantrips {
			cantripStep.Options = append(cantripStep.Options, character.CreationOption{
				Key:         cantrip.Key,
				Name:        cantrip.Name,
				Description: "Cantrip spell",
			})
		}
	}

	// Add both cantrip and spell steps
	steps = append(steps, cantripStep, character.CreationStep{
		Type:        character.StepTypeSpellsKnownSelection,
		Title:       "Choose Your Spells",
		Description: "Bards know a limited number of spells. Choose 4 1st-level spells from the bard spell list.",
		MinChoices:  4,
		MaxChoices:  4,
		Required:    true,
		UIHints: &character.StepUIHints{
			Actions: []character.StepAction{
				{
					ID:    "select_spells",
					Label: "Choose Spells",
					Style: "primary",
					Icon:  "🎵",
				},
				{
					ID:          "quick_build",
					Label:       "Suggested Spells",
					Style:       "secondary",
					Icon:        "💡",
					Description: "Popular bard spell choices",
				},
			},
			Layout:          "list",
			ShowProgress:    true,
			ProgressFormat:  "%d/%d spells selected",
			ShowRecommended: true,
			Color:           0xA855F7,
		},
	})

	return steps
}

// buildDruidSteps creates druid-specific steps
func (b *FlowBuilderImpl) buildDruidSteps(ctx context.Context, char *character.Character) []character.CreationStep {
	var steps []character.CreationStep

	// Cantrip selection (2 cantrips)
	druidCantripStep := character.CreationStep{
		Type:        character.StepTypeCantripsSelection,
		Title:       "Choose Your Cantrips",
		Description: "Cantrips are simple spells you can cast at will. Choose 2 cantrips from the druid spell list.",
		MinChoices:  2,
		MaxChoices:  2,
		Required:    true,
		UIHints: &character.StepUIHints{
			Layout:          "grid",
			ShowRecommended: true,
			Color:           0x10B981, // Green for druids
		},
	}

	// Populate cantrips options from D&D API
	if cantrips, err := b.dndClient.ListSpellsByClassAndLevel("druid", 0); err == nil {
		for _, cantrip := range cantrips {
			druidCantripStep.Options = append(druidCantripStep.Options, character.CreationOption{
				Key:         cantrip.Key,
				Name:        cantrip.Name,
				Description: "Cantrip spell",
			})
		}
	}

	steps = append(steps, druidCantripStep)

	// Note: Druids prepare spells, they don't have spells known
	// They can prepare Wisdom modifier + druid level spells each day

	return steps
}

// buildRogueSteps creates rogue-specific steps
func (b *FlowBuilderImpl) buildRogueSteps(ctx context.Context, char *character.Character) []character.CreationStep {
	var steps []character.CreationStep

	// Expertise selection (2 skills)
	steps = append(steps, character.CreationStep{
		Type:        character.StepTypeExpertiseSelection,
		Title:       "Choose Your Expertise",
		Description: "Your proficiency bonus is doubled for ability checks using these skills. Choose 2 skills you are proficient with.",
		MinChoices:  2,
		MaxChoices:  2,
		Required:    true,
		UIHints: &character.StepUIHints{
			Actions: []character.StepAction{
				{
					ID:    "select_expertise",
					Label: "Choose Skills",
					Style: "primary",
					Icon:  "🗡️",
				},
			},
			Layout:       "list",
			ShowProgress: true,
			Color:        0x64748B, // Gray for rogues
		},
	})

	// Thieves' Cant is automatic, no choice needed

	return steps
}

// buildSorcererSteps creates sorcerer-specific steps
func (b *FlowBuilderImpl) buildSorcererSteps(ctx context.Context, char *character.Character) []character.CreationStep {
	var steps []character.CreationStep

	// Sorcerous Origin selection (subclass at level 1!)
	steps = append(steps, character.CreationStep{
		Type:        character.StepTypeSubclassSelection,
		Title:       "Choose Your Sorcerous Origin",
		Description: "Choose the source of your innate magical power.",
		MinChoices:  1,
		MaxChoices:  1,
		Required:    true,
		UIHints: &character.StepUIHints{
			Actions: []character.StepAction{
				{
					ID:    "select_origin",
					Label: "Choose Origin",
					Style: "primary",
					Icon:  "🔥",
				},
			},
			Layout: "grid",
			Color:  0xDC2626, // Red for sorcerers
		},
	})

	// Cantrip selection (4 cantrips)
	sorcererCantripStep := character.CreationStep{
		Type:        character.StepTypeCantripsSelection,
		Title:       "Choose Your Cantrips",
		Description: "Cantrips are simple spells you can cast at will. Choose 4 cantrips from the sorcerer spell list.",
		MinChoices:  4,
		MaxChoices:  4,
		Required:    true,
		UIHints: &character.StepUIHints{
			Layout:          "grid",
			ShowRecommended: true,
			Color:           0xDC2626,
		},
	}

	// Populate cantrips options from D&D API
	if cantrips, err := b.dndClient.ListSpellsByClassAndLevel("sorcerer", 0); err == nil {
		for _, cantrip := range cantrips {
			sorcererCantripStep.Options = append(sorcererCantripStep.Options, character.CreationOption{
				Key:         cantrip.Key,
				Name:        cantrip.Name,
				Description: "Cantrip spell",
			})
		}
	}

	// Add both cantrip and spell steps
	steps = append(steps, sorcererCantripStep, character.CreationStep{
		Type:        character.StepTypeSpellsKnownSelection,
		Title:       "Choose Your Spells",
		Description: "Sorcerers know a limited number of spells. Choose 2 1st-level spells from the sorcerer spell list.",
		MinChoices:  2,
		MaxChoices:  2,
		Required:    true,
		UIHints: &character.StepUIHints{
			Actions: []character.StepAction{
				{
					ID:    "select_spells",
					Label: "Choose Spells",
					Style: "primary",
					Icon:  "🔮",
				},
			},
			Layout:          "list",
			ShowProgress:    true,
			ProgressFormat:  "%d/%d spells selected",
			ShowRecommended: true,
			Color:           0xDC2626,
		},
	})

	return steps
}

// buildWarlockSteps creates warlock-specific steps
func (b *FlowBuilderImpl) buildWarlockSteps(ctx context.Context, char *character.Character) []character.CreationStep {
	var steps []character.CreationStep

	// Otherworldly Patron selection (subclass at level 1!)
	steps = append(steps, character.CreationStep{
		Type:        character.StepTypeSubclassSelection,
		Title:       "Choose Your Otherworldly Patron",
		Description: "Choose the otherworldly being that has granted you power.",
		MinChoices:  1,
		MaxChoices:  1,
		Required:    true,
		UIHints: &character.StepUIHints{
			Actions: []character.StepAction{
				{
					ID:    "select_patron",
					Label: "Choose Patron",
					Style: "primary",
					Icon:  "👁️",
				},
			},
			Layout: "grid",
			Color:  0x7C3AED, // Dark purple for warlocks
		},
	})

	// Cantrip selection (2 cantrips)
	warlockCantripStep := character.CreationStep{
		Type:        character.StepTypeCantripsSelection,
		Title:       "Choose Your Cantrips",
		Description: "Cantrips are simple spells you can cast at will. Choose 2 cantrips from the warlock spell list.",
		MinChoices:  2,
		MaxChoices:  2,
		Required:    true,
		UIHints: &character.StepUIHints{
			Layout:          "grid",
			ShowRecommended: true,
			Color:           0x7C3AED,
		},
	}

	// Populate cantrips options from D&D API
	if cantrips, err := b.dndClient.ListSpellsByClassAndLevel("warlock", 0); err == nil {
		for _, cantrip := range cantrips {
			warlockCantripStep.Options = append(warlockCantripStep.Options, character.CreationOption{
				Key:         cantrip.Key,
				Name:        cantrip.Name,
				Description: "Cantrip spell",
			})
		}
	}

	// Add both cantrip and spell steps
	steps = append(steps, warlockCantripStep, character.CreationStep{
		Type:        character.StepTypeSpellsKnownSelection,
		Title:       "Choose Your Spells",
		Description: "Warlocks know a limited number of spells but cast them through Pact Magic. Choose 2 1st-level spells from the warlock spell list.",
		MinChoices:  2,
		MaxChoices:  2,
		Required:    true,
		UIHints: &character.StepUIHints{
			Actions: []character.StepAction{
				{
					ID:    "select_spells",
					Label: "Choose Spells",
					Style: "primary",
					Icon:  "📜",
				},
			},
			Layout:          "list",
			ShowProgress:    true,
			ProgressFormat:  "%d/%d spells selected",
			ShowRecommended: true,
			Color:           0x7C3AED,
		},
	})

	return steps
}
