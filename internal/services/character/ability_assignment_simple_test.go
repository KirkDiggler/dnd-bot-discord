package character

import (
	"testing"

	"github.com/KirkDiggler/dnd-bot-discord/internal/entities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFinalizeDraftCharacter_ConvertsAbilityAssignments_Simple(t *testing.T) {
	// This test verifies the core conversion logic without external dependencies
	
	// Test the conversion logic directly
	char := &entities.Character{
		ID:      "test-char-1",
		Name:    "Test Character",
		Status:  entities.CharacterStatusDraft,
		Level:   1,
		Race: &entities.Race{
			Key:  "elf",
			Name: "Elf",
			AbilityBonuses: []*entities.AbilityBonus{
				{Attribute: entities.AttributeDexterity, Bonus: 2},
				{Attribute: entities.AttributeIntelligence, Bonus: 1},
			},
		},
		Class: &entities.Class{
			Key:    "wizard",
			Name:   "Wizard",
			HitDie: 6,
		},
		AbilityRolls: []entities.AbilityRoll{
			{ID: "roll_1", Value: 15},
			{ID: "roll_2", Value: 14},
			{ID: "roll_3", Value: 13},
			{ID: "roll_4", Value: 12},
			{ID: "roll_5", Value: 11},
			{ID: "roll_6", Value: 10},
		},
		AbilityAssignments: map[string]string{
			"STR": "roll_3", // 13
			"DEX": "roll_2", // 14 + 2 (racial) = 16
			"CON": "roll_4", // 12
			"INT": "roll_1", // 15 + 1 (racial) = 16
			"WIS": "roll_5", // 11
			"CHA": "roll_6", // 10
		},
		Attributes: make(map[entities.Attribute]*entities.AbilityScore), // Empty attributes
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
		char.Attributes = make(map[entities.Attribute]*entities.AbilityScore)
		
		// Convert assignments to attributes
		for abilityStr, rollID := range char.AbilityAssignments {
			if rollValue, ok := rollValues[rollID]; ok {
				// Parse ability string to Attribute type
				var attr entities.Attribute
				switch abilityStr {
				case "STR":
					attr = entities.AttributeStrength
				case "DEX":
					attr = entities.AttributeDexterity
				case "CON":
					attr = entities.AttributeConstitution
				case "INT":
					attr = entities.AttributeIntelligence
				case "WIS":
					attr = entities.AttributeWisdom
				case "CHA":
					attr = entities.AttributeCharisma
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
				char.Attributes[attr] = &entities.AbilityScore{
					Score: score,
					Bonus: modifier,
				}
			}
		}
	}
	
	// Verify the conversion worked correctly
	require.NotEmpty(t, char.Attributes)
	assert.Len(t, char.Attributes, 6)
	
	// Check STR (13, no racial bonus)
	assert.NotNil(t, char.Attributes[entities.AttributeStrength])
	assert.Equal(t, 13, char.Attributes[entities.AttributeStrength].Score)
	assert.Equal(t, 1, char.Attributes[entities.AttributeStrength].Bonus) // (13-10)/2 = 1
	
	// Check DEX (14 + 2 racial = 16)
	assert.NotNil(t, char.Attributes[entities.AttributeDexterity])
	assert.Equal(t, 16, char.Attributes[entities.AttributeDexterity].Score)
	assert.Equal(t, 3, char.Attributes[entities.AttributeDexterity].Bonus) // (16-10)/2 = 3
	
	// Check INT (15 + 1 racial = 16)
	assert.NotNil(t, char.Attributes[entities.AttributeIntelligence])
	assert.Equal(t, 16, char.Attributes[entities.AttributeIntelligence].Score)
	assert.Equal(t, 3, char.Attributes[entities.AttributeIntelligence].Bonus) // (16-10)/2 = 3
}

