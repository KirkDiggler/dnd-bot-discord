package encounters

//go:generate mockgen -destination=mock/mock_repository.go -package=mockencrepo -source=repository.go

import (
	"context"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/game/combat"
)

// Repository defines the interface for encounter storage operations
type Repository interface {
	// Create stores a new encounter
	Create(ctx context.Context, encounter *combat.Encounter) error

	// Get retrieves an encounter by ID
	Get(ctx context.Context, id string) (*combat.Encounter, error)

	// Update modifies an existing encounter
	Update(ctx context.Context, encounter *combat.Encounter) error

	// Delete removes an encounter
	Delete(ctx context.Context, id string) error

	// GetBySession retrieves all encounters for a session
	GetBySession(ctx context.Context, sessionID string) ([]*combat.Encounter, error)

	// GetActiveBySession retrieves the active encounter for a session
	GetActiveBySession(ctx context.Context, sessionID string) (*combat.Encounter, error)

	// GetByMessage retrieves an encounter by Discord message ID
	GetByMessage(ctx context.Context, messageID string) (*combat.Encounter, error)
}
