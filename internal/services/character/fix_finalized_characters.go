package character

import (
	"context"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/rulebook/features"
	"log"
)

// FixCharacterAttributes fixes characters that were finalized without proper attribute conversion
// This is a utility method to fix the bug where characters have AbilityAssignments but no Attributes
func (s *service) FixCharacterAttributes(ctx context.Context, characterID string) (*character.Character, error) {
	// Get the character
	char, err := s.repository.Get(ctx, characterID)
	if err != nil {
		return nil, err
	}

	// Check if this character needs fixing
	if len(char.Attributes) > 0 || len(char.AbilityAssignments) == 0 || len(char.AbilityRolls) == 0 {
		// Character doesn't need fixing
		return char, nil
	}

	log.Printf("Fixing character %s (ID: %s) - has AbilityAssignments but no Attributes", char.Name, char.ID)

	// Apply the same conversion logic from FinalizeDraftCharacter
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
			log.Printf("Unknown ability string: %s", abilityStr)
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

		log.Printf("Created attribute %s: score=%d, modifier=%d", attr, score, modifier)
	}

	log.Printf("Fixed character now has %d attributes", len(char.Attributes))

	// Recalculate HP if needed
	if char.MaxHitPoints == 0 && char.Class != nil {
		conMod := 0
		if con, ok := char.Attributes[character.AttributeConstitution]; ok && con != nil {
			conMod = con.Bonus
		}
		char.MaxHitPoints = char.Class.HitDie + conMod
		char.CurrentHitPoints = char.MaxHitPoints
		log.Printf("Calculated HP: %d", char.MaxHitPoints)
	}

	// Recalculate AC if needed
	if char.AC == 0 {
		char.AC = features.CalculateAC(char)
		log.Printf("Calculated AC: %d", char.AC)
	}

	// Save the fixed character
	if err := s.repository.Update(ctx, char); err != nil {
		return nil, err
	}

	return char, nil
}
