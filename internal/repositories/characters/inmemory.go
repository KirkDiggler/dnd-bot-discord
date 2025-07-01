package characters

import (
	"context"
	"sync"

	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"

	dnderr "github.com/KirkDiggler/dnd-bot-discord/internal/errors"
)

// InMemoryRepository is an in-memory implementation of the character repository
// Useful for testing and development
type InMemoryRepository struct {
	mu         sync.RWMutex
	characters map[string]*character.Character
}

// NewInMemoryRepository creates a new in-memory repository
func NewInMemoryRepository() Repository {
	return &InMemoryRepository{
		characters: make(map[string]*character.Character),
	}
}

// Create stores a new character
func (r *InMemoryRepository) Create(ctx context.Context, char *character.Character) error {
	if char == nil {
		return dnderr.InvalidArgument("character cannot be nil")
	}

	if char.ID == "" {
		return dnderr.InvalidArgument("character ID is required")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.characters[char.ID]; exists {
		return dnderr.AlreadyExistsf("character with ID '%s' already exists", char.ID).
			WithMeta("character_id", char.ID)
	}

	// Create a copy to avoid external modifications
	r.characters[char.ID] = char.Clone()

	return nil
}

// Get retrieves a character by ID
func (r *InMemoryRepository) Get(ctx context.Context, id string) (*character.Character, error) {
	if id == "" {
		return nil, dnderr.InvalidArgument("character ID is required")
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	char, exists := r.characters[id]
	if !exists {
		return nil, dnderr.NotFoundf("character with ID '%s' not found", id).
			WithMeta("character_id", id)
	}

	// Return a copy to avoid external modifications
	return char.Clone(), nil
}

// GetByOwner retrieves all characters for a specific owner
func (r *InMemoryRepository) GetByOwner(ctx context.Context, ownerID string) ([]*character.Character, error) {
	if ownerID == "" {
		return nil, dnderr.InvalidArgument("owner ID is required")
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*character.Character
	for _, char := range r.characters {
		if char.OwnerID == ownerID {
			// Create a copy
			result = append(result, char.Clone())
		}
	}

	return result, nil
}

// GetByOwnerAndRealm retrieves all characters for a specific owner in a realm
func (r *InMemoryRepository) GetByOwnerAndRealm(ctx context.Context, ownerID, realmID string) ([]*character.Character, error) {
	if ownerID == "" {
		return nil, dnderr.InvalidArgument("owner ID is required")
	}

	if realmID == "" {
		return nil, dnderr.InvalidArgument("realm ID is required")
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*character.Character
	for _, char := range r.characters {
		if char.OwnerID == ownerID && char.RealmID == realmID {
			// Create a copy
			result = append(result, char.Clone())
		}
	}

	return result, nil
}

// Update updates an existing character
func (r *InMemoryRepository) Update(ctx context.Context, char *character.Character) error {
	if char == nil {
		return dnderr.InvalidArgument("character cannot be nil")
	}

	if char.ID == "" {
		return dnderr.InvalidArgument("character ID is required")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.characters[char.ID]; !exists {
		return dnderr.NotFoundf("character with ID '%s' not found", char.ID).
			WithMeta("character_id", char.ID)
	}

	// Create a copy to avoid external modifications
	r.characters[char.ID] = char.Clone()

	return nil
}

// Delete removes a character
func (r *InMemoryRepository) Delete(ctx context.Context, id string) error {
	if id == "" {
		return dnderr.InvalidArgument("character ID is required")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.characters[id]; !exists {
		return dnderr.NotFoundf("character with ID '%s' not found", id).
			WithMeta("character_id", id)
	}

	delete(r.characters, id)
	return nil
}
