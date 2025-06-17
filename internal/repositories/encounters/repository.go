package encounters

import (
	"context"
	"github.com/KirkDiggler/dnd-bot-discord/internal/entities"
)

// Repository defines the interface for encounter storage operations
type Repository interface {
	// Create stores a new encounter
	Create(ctx context.Context, encounter *entities.Encounter) error
	
	// Get retrieves an encounter by ID
	Get(ctx context.Context, id string) (*entities.Encounter, error)
	
	// Update modifies an existing encounter
	Update(ctx context.Context, encounter *entities.Encounter) error
	
	// Delete removes an encounter
	Delete(ctx context.Context, id string) error
	
	// GetBySession retrieves all encounters for a session
	GetBySession(ctx context.Context, sessionID string) ([]*entities.Encounter, error)
	
	// GetActiveBySession retrieves the active encounter for a session
	GetActiveBySession(ctx context.Context, sessionID string) (*entities.Encounter, error)
	
	// GetByMessage retrieves an encounter by Discord message ID
	GetByMessage(ctx context.Context, messageID string) (*entities.Encounter, error)
}