package character

import (
	"context"
)

// DraftRepository defines the interface for character draft persistence
type DraftRepository interface {
	// Create stores a new character draft
	Create(ctx context.Context, draft *CharacterDraft) error

	// Get retrieves a character draft by ID
	Get(ctx context.Context, id string) (*CharacterDraft, error)

	// GetByOwnerAndRealm retrieves the active draft for an owner in a realm
	GetByOwnerAndRealm(ctx context.Context, ownerID, realmID string) (*CharacterDraft, error)

	// Update updates an existing character draft
	Update(ctx context.Context, draft *CharacterDraft) error

	// Delete removes a character draft
	Delete(ctx context.Context, id string) error
}
