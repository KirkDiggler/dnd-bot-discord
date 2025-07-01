package character_test

import (
	character2 "github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/rulebook/dnd5e"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/shared"
	"testing"

	"github.com/KirkDiggler/dnd-bot-discord/internal/handlers/discord/dnd/character"
	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShowCharacterSheetAfterCreation(t *testing.T) {
	// This test verifies the character creation success response
	// and full character sheet display to address issue #166

	t.Run("BuildCharacterSheetEmbed shows complete character info", func(t *testing.T) {
		// Create a test character that represents a newly finalized character
		finalChar := &character2.Character{
			ID:      "test-char-123",
			Name:    "Thorin Ironforge",
			OwnerID: "user123",
			Level:   1,
			Status:  shared.CharacterStatusActive,
			Race: &rulebook.Race{
				Key:   "dwarf",
				Name:  "Dwarf",
				Speed: 25,
			},
			Class: &rulebook.Class{
				Key:    "fighter",
				Name:   "Fighter",
				HitDie: 10,
			},
			CurrentHitPoints: 12,
			MaxHitPoints:     12,
			AC:               16,
			HitDie:           10,
			Attributes: map[shared.Attribute]*character2.AbilityScore{
				shared.AttributeStrength:     {Score: 16, Bonus: 3},
				shared.AttributeDexterity:    {Score: 14, Bonus: 2},
				shared.AttributeConstitution: {Score: 15, Bonus: 2},
				shared.AttributeIntelligence: {Score: 13, Bonus: 1},
				shared.AttributeWisdom:       {Score: 12, Bonus: 1},
				shared.AttributeCharisma:     {Score: 10, Bonus: 0},
			},
			Features: []*rulebook.CharacterFeature{
				{
					Key:  "fighting_style",
					Name: "Fighting Style",
					Type: rulebook.FeatureTypeClass,
					Metadata: map[string]any{
						"style": "defense",
					},
				},
				{
					Key:  "dwarven_resilience",
					Name: "Dwarven Resilience",
					Type: rulebook.FeatureTypeRacial,
				},
			},
			Proficiencies: map[rulebook.ProficiencyType][]*rulebook.Proficiency{
				rulebook.ProficiencyTypeWeapon: {
					{Key: "simple-weapons", Name: "Simple Weapons"},
					{Key: "martial-weapons", Name: "Martial Weapons"},
				},
				rulebook.ProficiencyTypeArmor: {
					{Key: "light-armor", Name: "Light Armor"},
					{Key: "medium-armor", Name: "Medium Armor"},
					{Key: "heavy-armor", Name: "Heavy Armor"},
					{Key: "shields", Name: "Shields"},
				},
			},
		}

		// Build the character sheet embed
		embed := character.BuildCharacterSheetEmbed(finalChar)

		// Verify the embed has all the expected information
		require.NotNil(t, embed)

		// Check title includes name, level, race and class
		assert.Contains(t, embed.Title, "Thorin Ironforge")
		assert.Contains(t, embed.Title, "Level 1")
		assert.Contains(t, embed.Title, "Dwarf")
		assert.Contains(t, embed.Title, "Fighter")

		// Check that we have fields
		require.NotEmpty(t, embed.Fields)

		// Check for specific fields
		hasAbilityScores := false
		hasFeatures := false
		hasProficiencies := false
		hasEquipment := false

		for _, field := range embed.Fields {
			if field.Name == "📊 Ability Scores" {
				hasAbilityScores = true
				// Should contain ability scores
				assert.Contains(t, field.Value, "STR")
				assert.Contains(t, field.Value, "16")
			}
			if field.Name == "✨ Features" {
				hasFeatures = true
				// Should contain fighting style
				assert.Contains(t, field.Value, "Fighting Style")
			}
			if field.Name == "📚 Proficiencies" {
				hasProficiencies = true
				// Should contain weapon and armor proficiencies
				assert.Contains(t, field.Value, "Weapons:")
				assert.Contains(t, field.Value, "Armor:")
			}
			if field.Name == "⚔️ Equipment" {
				hasEquipment = true
			}
		}

		assert.True(t, hasAbilityScores, "Character sheet should show ability scores")
		assert.True(t, hasFeatures, "Character sheet should show features")
		assert.True(t, hasProficiencies, "Character sheet should show proficiencies")
		assert.True(t, hasEquipment, "Character sheet should show equipment section")
	})

	t.Run("BuildCharacterSheetComponents provides interactive buttons", func(t *testing.T) {
		characterID := "test-char-456"

		// Build the components
		components := character.BuildCharacterSheetComponents(characterID)

		// Verify we have components
		require.NotEmpty(t, components)

		// Check the first component is an action row
		actionRow, ok := components[0].(discordgo.ActionsRow)
		require.True(t, ok, "First component should be an action row")

		// Check we have buttons
		require.NotEmpty(t, actionRow.Components)

		// Verify button custom IDs include the character ID
		buttonCount := 0
		for _, comp := range actionRow.Components {
			if button, ok := comp.(discordgo.Button); ok {
				assert.Contains(t, button.CustomID, characterID)
				buttonCount++
			}
		}
		assert.Equal(t, 3, buttonCount, "Should have 3 buttons: Manage Equipment, Edit Character, Refresh")
	})

	t.Run("BuildCreationSuccessResponse provides simple success with view button", func(t *testing.T) {
		// Create a test character
		finalChar := &character2.Character{
			ID:     "test-char-789",
			Name:   "Gimli",
			Level:  1,
			Status: shared.CharacterStatusActive,
			Race: &rulebook.Race{
				Key:  "dwarf",
				Name: "Dwarf",
			},
			Class: &rulebook.Class{
				Key:    "fighter",
				Name:   "Fighter",
				HitDie: 10,
			},
			CurrentHitPoints: 12,
			MaxHitPoints:     12,
			AC:               16,
		}

		// Build the success response
		embed, components := character.BuildCreationSuccessResponse(finalChar)

		// Verify embed
		require.NotNil(t, embed)
		assert.Contains(t, embed.Title, "Character Created Successfully")
		assert.Contains(t, embed.Description, "Gimli")

		// Verify we have the view button
		require.Len(t, components, 1, "Should have one action row")
		actionRow, ok := components[0].(discordgo.ActionsRow)
		require.True(t, ok)
		require.Len(t, actionRow.Components, 1, "Should have one button")

		button, ok := actionRow.Components[0].(discordgo.Button)
		require.True(t, ok)
		assert.Equal(t, "View Character Sheet", button.Label)
		assert.Equal(t, "character:sheet_show:test-char-789", button.CustomID)
	})
}
