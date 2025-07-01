package character

import (
	"context"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/shared"
	"log"

	dnderr "github.com/KirkDiggler/dnd-bot-discord/internal/errors"
)

// RepairCharacterAttributes fixes characters that have AbilityAssignments but no Attributes
func (s *service) RepairCharacterAttributes(ctx context.Context, characterID string) error {
	// Get the character
	char, err := s.repository.Get(ctx, characterID)
	if err != nil {
		return dnderr.Wrapf(err, "failed to get character '%s'", characterID).
			WithMeta("character_id", characterID)
	}

	// Check if repair is needed
	if len(char.Attributes) > 0 || len(char.AbilityAssignments) == 0 || len(char.AbilityRolls) == 0 {
		// No repair needed
		return nil
	}

	log.Printf("Repairing character %s (%s) - converting AbilityAssignments to Attributes", char.Name, char.ID)

	// Create roll ID to value map
	rollValues := make(map[string]int)
	for _, roll := range char.AbilityRolls {
		rollValues[roll.ID] = roll.Value
	}

	// Initialize attributes map
	char.Attributes = make(map[shared.Attribute]*character.AbilityScore)

	// Convert assignments to attributes
	for abilityStr, rollID := range char.AbilityAssignments {
		if _, ok := rollValues[rollID]; !ok {
			log.Printf("Roll ID %s not found for character %s", rollID, char.ID)
			continue
		}
		rollValue := rollValues[rollID]
		// Parse ability string to Attribute type
		var attr shared.Attribute
		switch abilityStr {
		case "STR":
			attr = shared.AttributeStrength
		case "DEX":
			attr = shared.AttributeDexterity
		case "CON":
			attr = shared.AttributeConstitution
		case "INT":
			attr = shared.AttributeIntelligence
		case "WIS":
			attr = shared.AttributeWisdom
		case "CHA":
			attr = shared.AttributeCharisma
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

	// Recalculate HP if needed
	if char.MaxHitPoints == 0 && char.Class != nil {
		conMod := 0
		if char.Attributes != nil {
			if con, ok := char.Attributes[shared.AttributeConstitution]; ok && con != nil {
				conMod = con.Bonus
			}
		}
		char.MaxHitPoints = char.Class.HitDie + conMod
		char.CurrentHitPoints = char.MaxHitPoints
	}

	// Recalculate AC if needed
	if char.AC == 0 {
		baseAC := 10
		dexMod := 0

		if char.Attributes != nil {
			if dex, ok := char.Attributes[shared.AttributeDexterity]; ok && dex != nil {
				dexMod = dex.Bonus
			}
		}

		char.AC = baseAC + dexMod
	}

	// Save the repaired character
	if err := s.repository.Update(ctx, char); err != nil {
		return dnderr.Wrap(err, "failed to save repaired character").
			WithMeta("character_id", characterID)
	}

	log.Printf("Character %s (%s) repaired successfully", char.Name, char.ID)
	return nil
}

// RepairAllCharacters repairs all characters for a given owner
func (s *service) RepairAllCharacters(ctx context.Context, ownerID string) error {
	chars, err := s.repository.GetByOwner(ctx, ownerID)
	if err != nil {
		return dnderr.Wrapf(err, "failed to list characters for owner '%s'", ownerID).
			WithMeta("owner_id", ownerID)
	}

	repaired := 0
	for _, char := range chars {
		if len(char.Attributes) == 0 && len(char.AbilityAssignments) > 0 {
			if err := s.RepairCharacterAttributes(ctx, char.ID); err != nil {
				log.Printf("Failed to repair character %s: %v", char.ID, err)
			} else {
				repaired++
			}
		}
	}

	log.Printf("Repaired %d characters for owner %s", repaired, ownerID)
	return nil
}
