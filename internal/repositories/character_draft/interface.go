package character_draft

//go:generate mockgen -destination=mock/mock.go -package=mockcharacterdraft -source=interface.go

import (
	"context"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
)

// Repository defines the interface for character draft persistence
type Repository interface {
	// Create stores a new character draft
	Create(ctx context.Context, draft *character.CharacterDraft) error

	// Get retrieves a character draft by ID
	Get(ctx context.Context, id string) (*character.CharacterDraft, error)

	// GetByOwnerAndRealm retrieves the active draft for an owner in a realm
	GetByOwnerAndRealm(ctx context.Context, ownerID, realmID string) (*character.CharacterDraft, error)

	// GetByCharacterID retrieves a draft by the character ID it contains
	GetByCharacterID(ctx context.Context, characterID string) (*character.CharacterDraft, error)

	// Update updates an existing character draft
	Update(ctx context.Context, draft *character.CharacterDraft) error

	// Delete removes a character draft
	Delete(ctx context.Context, id string) error
}
