package character_draft

import (
	"context"
	"sync"

	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	dnderr "github.com/KirkDiggler/dnd-bot-discord/internal/errors"
)

// InMemoryRepository is an in-memory implementation of the character draft repository
type InMemoryRepository struct {
	mu     sync.RWMutex
	drafts map[string]*character.CharacterDraft
}

// NewInMemoryRepository creates a new in-memory draft repository
func NewInMemoryRepository() Repository {
	return &InMemoryRepository{
		drafts: make(map[string]*character.CharacterDraft),
	}
}

// Create stores a new character draft
func (r *InMemoryRepository) Create(ctx context.Context, draft *character.CharacterDraft) error {
	if draft == nil {
		return dnderr.InvalidArgument("draft cannot be nil")
	}

	if draft.ID == "" {
		return dnderr.InvalidArgument("draft ID is required")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.drafts[draft.ID]; exists {
		return dnderr.AlreadyExistsf("draft with ID '%s' already exists", draft.ID).
			WithMeta("draft_id", draft.ID)
	}

	// Store the draft
	r.drafts[draft.ID] = draft

	return nil
}

// Get retrieves a character draft by ID
func (r *InMemoryRepository) Get(ctx context.Context, id string) (*character.CharacterDraft, error) {
	if id == "" {
		return nil, dnderr.InvalidArgument("draft ID is required")
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	draft, exists := r.drafts[id]
	if !exists {
		return nil, dnderr.NotFoundf("draft with ID '%s' not found", id).
			WithMeta("draft_id", id)
	}

	return draft, nil
}

// GetByOwnerAndRealm retrieves the active draft for an owner in a realm
func (r *InMemoryRepository) GetByOwnerAndRealm(ctx context.Context, ownerID, realmID string) (*character.CharacterDraft, error) {
	if ownerID == "" {
		return nil, dnderr.InvalidArgument("owner ID is required")
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	// Find the first draft for this owner/realm (simple implementation)
	for _, draft := range r.drafts {
		if draft.OwnerID == ownerID && draft.Character != nil && draft.Character.RealmID == realmID {
			return draft, nil
		}
	}

	return nil, dnderr.NotFoundf("no draft found for owner '%s' in realm '%s'", ownerID, realmID).
		WithMeta("owner_id", ownerID).
		WithMeta("realm_id", realmID)
}

// Update updates an existing character draft
func (r *InMemoryRepository) Update(ctx context.Context, draft *character.CharacterDraft) error {
	if draft == nil {
		return dnderr.InvalidArgument("draft cannot be nil")
	}

	if draft.ID == "" {
		return dnderr.InvalidArgument("draft ID is required")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.drafts[draft.ID]; !exists {
		return dnderr.NotFoundf("draft with ID '%s' not found", draft.ID).
			WithMeta("draft_id", draft.ID)
	}

	// Update the draft
	r.drafts[draft.ID] = draft

	return nil
}

// GetByCharacterID retrieves a draft by the character ID it contains
func (r *InMemoryRepository) GetByCharacterID(ctx context.Context, characterID string) (*character.CharacterDraft, error) {
	if characterID == "" {
		return nil, dnderr.InvalidArgument("character ID is required")
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	// Search through all drafts to find one with matching character ID
	for _, draft := range r.drafts {
		if draft.Character != nil && draft.Character.ID == characterID {
			return draft, nil
		}
	}

	return nil, dnderr.NotFoundf("draft for character ID '%s' not found", characterID).
		WithMeta("character_id", characterID)
}

// Delete removes a character draft
func (r *InMemoryRepository) Delete(ctx context.Context, id string) error {
	if id == "" {
		return dnderr.InvalidArgument("draft ID is required")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.drafts[id]; !exists {
		return dnderr.NotFoundf("draft with ID '%s' not found", id).
			WithMeta("draft_id", id)
	}

	delete(r.drafts, id)
	return nil
}
