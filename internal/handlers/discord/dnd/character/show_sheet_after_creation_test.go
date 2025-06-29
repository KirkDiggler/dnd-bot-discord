package character_test

import (
	"testing"

	"github.com/KirkDiggler/dnd-bot-discord/internal/entities"
	"github.com/KirkDiggler/dnd-bot-discord/internal/handlers/discord/dnd/character"
	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShowCharacterSheetAfterCreation(t *testing.T) {
	// This test verifies that we can properly build a character sheet embed
	// for a newly created character to address issue #166

	t.Run("BuildCharacterSheetEmbed shows complete character info", func(t *testing.T) {
		// Create a test character that represents a newly finalized character
		finalChar := &entities.Character{
			ID:      "test-char-123",
			Name:    "Thorin Ironforge",
			OwnerID: "user123",
			Level:   1,
			Status:  entities.CharacterStatusActive,
			Race: &entities.Race{
				Key:   "dwarf",
				Name:  "Dwarf",
				Speed: 25,
			},
			Class: &entities.Class{
				Key:    "fighter",
				Name:   "Fighter",
				HitDie: 10,
			},
			CurrentHitPoints: 12,
			MaxHitPoints:     12,
			AC:               16,
			HitDie:           10,
			Attributes: map[entities.Attribute]*entities.AbilityScore{
				entities.AttributeStrength:     {Score: 16, Bonus: 3},
				entities.AttributeDexterity:    {Score: 14, Bonus: 2},
				entities.AttributeConstitution: {Score: 15, Bonus: 2},
				entities.AttributeIntelligence: {Score: 13, Bonus: 1},
				entities.AttributeWisdom:       {Score: 12, Bonus: 1},
				entities.AttributeCharisma:     {Score: 10, Bonus: 0},
			},
			Features: []*entities.CharacterFeature{
				{
					Key:  "fighting_style",
					Name: "Fighting Style",
					Type: entities.FeatureTypeClass,
					Metadata: map[string]any{
						"style": "defense",
					},
				},
				{
					Key:  "dwarven_resilience",
					Name: "Dwarven Resilience",
					Type: entities.FeatureTypeRacial,
				},
			},
			Proficiencies: map[entities.ProficiencyType][]*entities.Proficiency{
				entities.ProficiencyTypeWeapon: {
					{Key: "simple-weapons", Name: "Simple Weapons"},
					{Key: "martial-weapons", Name: "Martial Weapons"},
				},
				entities.ProficiencyTypeArmor: {
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
		for _, comp := range actionRow.Components {
			if button, ok := comp.(discordgo.Button); ok {
				assert.Contains(t, button.CustomID, characterID)
			}
		}
	})
}
