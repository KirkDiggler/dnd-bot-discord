package dungeons

import (
	"context"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/game/exploration"
)

// Repository defines the interface for dungeon storage operations
type Repository interface {
	// Create creates a new dungeon
	Create(ctx context.Context, dungeon *exploration.Dungeon) error

	// Get retrieves a dungeon by ID
	Get(ctx context.Context, id string) (*exploration.Dungeon, error)

	// Update updates an existing dungeon
	Update(ctx context.Context, dungeon *exploration.Dungeon) error

	// Delete removes a dungeon
	Delete(ctx context.Context, id string) error

	// GetBySession retrieves a dungeon by session ID
	GetBySession(ctx context.Context, sessionID string) (*exploration.Dungeon, error)

	// ListActive retrieves all active dungeons
	ListActive(ctx context.Context) ([]*exploration.Dungeon, error)
}
