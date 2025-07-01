package character

import (
	"context"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/rulebook/dnd5e"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/shared"
	"log"
)

// FixCharacterProficiencies adds missing class and racial proficiencies to existing characters
// This fixes characters created before the proficiency fix was implemented
func (s *service) FixCharacterProficiencies(ctx context.Context, characterID string) (*character.Character, error) {
	// Get the character
	char, err := s.repository.Get(ctx, characterID)
	if err != nil {
		return nil, err
	}

	// Only fix active characters with a class
	if char.Status != shared.CharacterStatusActive || char.Class == nil {
		return char, nil
	}

	// Check if character already has weapon proficiencies (indicator that fix was applied)
	if weaponProfs, exists := char.Proficiencies[rulebook.ProficiencyTypeWeapon]; exists && len(weaponProfs) > 0 {
		log.Printf("Character %s (ID: %s) already has weapon proficiencies, skipping fix", char.Name, char.ID)
		return char, nil
	}

	log.Printf("Fixing proficiencies for character %s (ID: %s) - %s %s", char.Name, char.ID, char.Race.Name, char.Class.Name)

	// Add class proficiencies
	if s.dndClient != nil && char.Class.Proficiencies != nil {
		for _, prof := range char.Class.Proficiencies {
			if prof != nil {
				proficiency, err := s.dndClient.GetProficiency(prof.Key)
				if err == nil && proficiency != nil {
					char.AddProficiency(proficiency)
					log.Printf("  Added class proficiency: %s", proficiency.Name)
				}
			}
		}
	}

	// Add racial proficiencies
	if s.dndClient != nil && char.Race != nil && char.Race.StartingProficiencies != nil {
		for _, prof := range char.Race.StartingProficiencies {
			if prof != nil {
				proficiency, err := s.dndClient.GetProficiency(prof.Key)
				if err == nil && proficiency != nil {
					char.AddProficiency(proficiency)
					log.Printf("  Added racial proficiency: %s", proficiency.Name)
				}
			}
		}
	}

	// Save the updated character
	if err := s.repository.Update(ctx, char); err != nil {
		return nil, err
	}

	log.Printf("Successfully fixed proficiencies for character %s", char.Name)
	return char, nil
}

// FixAllCharacterProficiencies fixes proficiencies for all active characters
func (s *service) FixAllCharacterProficiencies(ctx context.Context, realmID string) error {
	// Get all characters in the realm
	// Note: This would need a method to get all characters, which might not exist yet
	// For now, this is a placeholder
	log.Printf("Fix all character proficiencies called for realm %s", realmID)
	return nil
}
