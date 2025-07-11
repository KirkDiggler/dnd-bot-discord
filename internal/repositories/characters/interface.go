package characters

//go:generate mockgen -destination=mock/mock.go -package=mockcharacters -source=interface.go

import (
	"context"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"

	dnderr "github.com/KirkDiggler/dnd-bot-discord/internal/errors"
)

// Repository defines the interface for character persistence
type Repository interface {
	// Create stores a new character
	Create(ctx context.Context, character *character.Character) error

	// Get retrieves a character by ID
	Get(ctx context.Context, id string) (*character.Character, error)

	// GetByOwner retrieves all characters for a specific owner
	GetByOwner(ctx context.Context, ownerID string) ([]*character.Character, error)

	// GetByOwnerAndRealm retrieves all characters for a specific owner in a realm
	GetByOwnerAndRealm(ctx context.Context, ownerID, realmID string) ([]*character.Character, error)

	// Update updates an existing character
	Update(ctx context.Context, character *character.Character) error

	// Delete removes a character
	Delete(ctx context.Context, id string) error
}

// Deprecated: Use dnderr.NotFound instead
type NotFoundError = dnderr.Error

// Deprecated: Use dnderr.AlreadyExists instead
type AlreadyExistsError = dnderr.Error
