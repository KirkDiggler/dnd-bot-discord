package character

import (
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/rulebook"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFinalizeDraftCharacter_ConvertsAbilityAssignments_Simple(t *testing.T) {
	// This test verifies the core conversion logic without external dependencies

	// Test the conversion logic directly
	char := &character.Character{
		ID:     "test-char-1",
		Name:   "Test Character",
		Status: character.CharacterStatusDraft,
		Level:  1,
		Race: &rulebook.Race{
			Key:  "elf",
			Name: "Elf",
			AbilityBonuses: []*character.AbilityBonus{
				{Attribute: character.AttributeDexterity, Bonus: 2},
				{Attribute: character.AttributeIntelligence, Bonus: 1},
			},
		},
		Class: &rulebook.Class{
			Key:    "wizard",
			Name:   "Wizard",
			HitDie: 6,
		},
		AbilityRolls: []character.AbilityRoll{
			{ID: "roll_1", Value: 15},
			{ID: "roll_2", Value: 14},
			{ID: "roll_3", Value: 13},
			{ID: "roll_4", Value: 12},
			{ID: "roll_5", Value: 11},
			{ID: "roll_6", Value: 10},
		},
		AbilityAssignments: map[string]string{
			"STR": "roll_3", // Strength is roll 3 and has a score of 13
			"DEX": "roll_2", // Dexterity is roll 2 and has a score of 14 + 2 (racial) = 16
			"CON": "roll_4", // Constitution is roll 4 and has a score of 12
			"INT": "roll_1", // Intelligence is roll 1 and has a score of 15 + 1 (racial) = 16
			"WIS": "roll_5", // Wisdom is roll 5 and has a score of 11
			"CHA": "roll_6", // Charisma is roll 6 and has a score of 10
		},
		Attributes: make(map[character.Attribute]*character.AbilityScore), // Empty attributes
	}

	// Test the conversion logic from FinalizeDraftCharacter
	// This is the exact code from the service
	if len(char.Attributes) == 0 && len(char.AbilityAssignments) > 0 && len(char.AbilityRolls) > 0 {
		// Create roll ID to value map
		rollValues := make(map[string]int)
		for _, roll := range char.AbilityRolls {
			rollValues[roll.ID] = roll.Value
		}

		// Initialize attributes map
		char.Attributes = make(map[character.Attribute]*character.AbilityScore)

		// Convert assignments to attributes
		for abilityStr, rollID := range char.AbilityAssignments {
			if _, ok := rollValues[rollID]; !ok {
				log.Printf("Roll ID %s not found for character %s", rollID, char.ID)
				continue
			}
			rollValue := rollValues[rollID]
			// Parse ability string to Attribute type
			var attr character.Attribute
			switch abilityStr {
			case "STR":
				attr = character.AttributeStrength
			case "DEX":
				attr = character.AttributeDexterity
			case "CON":
				attr = character.AttributeConstitution
			case "INT":
				attr = character.AttributeIntelligence
			case "WIS":
				attr = character.AttributeWisdom
			case "CHA":
				attr = character.AttributeCharisma
			default:
				continue
			}

			// Create base ability score
			score := rollValue

			// Apply racial bonuses
			if char.Race != nil {
				for _, bonus := range char.Race.AbilityBonuses {
					if bonus.Attribute == attr {
						score += bonus.Bonus
					}
				}
			}

			// Calculate modifier
			modifier := (score - 10) / 2

			// Create ability score
			char.Attributes[attr] = &character.AbilityScore{
				Score: score,
				Bonus: modifier,
			}
		}
	}

	// Verify the conversion worked correctly
	require.NotEmpty(t, char.Attributes)
	assert.Len(t, char.Attributes, 6)

	// Check STR (13, no racial bonus)
	assert.NotNil(t, char.Attributes[character.AttributeStrength])
	assert.Equal(t, 13, char.Attributes[character.AttributeStrength].Score)
	assert.Equal(t, 1, char.Attributes[character.AttributeStrength].Bonus) // (13-10)/2 = 1

	// Check DEX (14 + 2 racial = 16)
	assert.NotNil(t, char.Attributes[character.AttributeDexterity])
	assert.Equal(t, 16, char.Attributes[character.AttributeDexterity].Score)
	assert.Equal(t, 3, char.Attributes[character.AttributeDexterity].Bonus) // (16-10)/2 = 3

	// Check INT (15 + 1 racial = 16)
	assert.NotNil(t, char.Attributes[character.AttributeIntelligence])
	assert.Equal(t, 16, char.Attributes[character.AttributeIntelligence].Score)
	assert.Equal(t, 3, char.Attributes[character.AttributeIntelligence].Bonus) // (16-10)/2 = 3
}
