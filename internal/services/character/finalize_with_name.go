package character

import (
	"context"
	"log"
	"strings"

	"github.com/KirkDiggler/dnd-bot-discord/internal/entities"
	dnderr "github.com/KirkDiggler/dnd-bot-discord/internal/errors"
)

// FinalizeCharacterWithName sets the name and finalizes a draft character in one operation
func (s *service) FinalizeCharacterWithName(ctx context.Context, characterID, name, raceKey, classKey string) (*entities.Character, error) {
	if strings.TrimSpace(characterID) == "" {
		return nil, dnderr.InvalidArgument("character ID is required")
	}
	if strings.TrimSpace(name) == "" {
		return nil, dnderr.InvalidArgument("character name is required")
	}

	log.Printf("Finalizing character %s with name: %s", characterID, name)

	// Update the draft with name, race, and class
	updates := &UpdateDraftInput{
		Name: &name,
	}
	
	// Only set race/class if provided (they might already be set)
	if raceKey != "" {
		updates.RaceKey = &raceKey
	}
	if classKey != "" {
		updates.ClassKey = &classKey
	}

	// Update the character
	_, err := s.UpdateDraftCharacter(ctx, characterID, updates)
	if err != nil {
		return nil, dnderr.Wrap(err, "failed to update character with name")
	}

	log.Printf("Character updated with name, now finalizing...")

	// Finalize the character
	finalChar, err := s.FinalizeDraftCharacter(ctx, characterID)
	if err != nil {
		return nil, dnderr.Wrap(err, "failed to finalize character")
	}

	log.Printf("Character %s finalized successfully with %d attributes", finalChar.Name, len(finalChar.Attributes))

	return finalChar, nil
}