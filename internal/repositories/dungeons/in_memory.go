package dungeons

import (
	"context"
	"fmt"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/game/exploration"
	"sync"
)

// inMemoryRepository implements Repository using in-memory storage
type inMemoryRepository struct {
	mu       sync.RWMutex
	dungeons map[string]*exploration.Dungeon
}

// NewInMemoryRepository creates a new in-memory dungeon repository
func NewInMemoryRepository() Repository {
	return &inMemoryRepository{
		dungeons: make(map[string]*exploration.Dungeon),
	}
}

// Create creates a new dungeon
func (r *inMemoryRepository) Create(ctx context.Context, dungeon *exploration.Dungeon) error {
	if dungeon == nil {
		return fmt.Errorf("dungeon cannot be nil")
	}
	if dungeon.ID == "" {
		return fmt.Errorf("dungeon ID cannot be empty")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.dungeons[dungeon.ID]; exists {
		return fmt.Errorf("dungeon with ID %s already exists", dungeon.ID)
	}

	// Deep copy to avoid external modifications
	dungeonCopy := *dungeon
	r.dungeons[dungeon.ID] = &dungeonCopy

	return nil
}

// Get retrieves a dungeon by ID
func (r *inMemoryRepository) Get(ctx context.Context, id string) (*exploration.Dungeon, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	dungeon, exists := r.dungeons[id]
	if !exists {
		return nil, fmt.Errorf("dungeon not found: %s", id)
	}

	// Return a copy to avoid external modifications
	dungeonCopy := *dungeon
	return &dungeonCopy, nil
}

// Update updates an existing dungeon
func (r *inMemoryRepository) Update(ctx context.Context, dungeon *exploration.Dungeon) error {
	if dungeon == nil {
		return fmt.Errorf("dungeon cannot be nil")
	}
	if dungeon.ID == "" {
		return fmt.Errorf("dungeon ID cannot be empty")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.dungeons[dungeon.ID]; !exists {
		return fmt.Errorf("dungeon not found: %s", dungeon.ID)
	}

	// Deep copy to avoid external modifications
	dungeonCopy := *dungeon
	r.dungeons[dungeon.ID] = &dungeonCopy

	return nil
}

// Delete removes a dungeon
func (r *inMemoryRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.dungeons[id]; !exists {
		return fmt.Errorf("dungeon not found: %s", id)
	}

	delete(r.dungeons, id)
	return nil
}

// GetBySession retrieves a dungeon by session ID
func (r *inMemoryRepository) GetBySession(ctx context.Context, sessionID string) (*exploration.Dungeon, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, dungeon := range r.dungeons {
		if dungeon.SessionID == sessionID && dungeon.IsActive() {
			// Return a copy
			dungeonCopy := *dungeon
			return &dungeonCopy, nil
		}
	}

	return nil, fmt.Errorf("no active dungeon found for session: %s", sessionID)
}

// ListActive retrieves all active dungeons
func (r *inMemoryRepository) ListActive(ctx context.Context) ([]*exploration.Dungeon, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var active []*exploration.Dungeon
	for _, dungeon := range r.dungeons {
		if dungeon.IsActive() {
			// Add a copy
			dungeonCopy := *dungeon
			active = append(active, &dungeonCopy)
		}
	}

	return active, nil
}
