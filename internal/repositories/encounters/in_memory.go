package encounters

import (
	"context"
	"fmt"
	"sync"
	
	"github.com/KirkDiggler/dnd-bot-discord/internal/entities"
)

type inMemoryRepository struct {
	mu         sync.RWMutex
	encounters map[string]*entities.Encounter
	bySession  map[string][]string // sessionID -> encounter IDs
	byMessage  map[string]string   // messageID -> encounter ID
}

// NewInMemoryRepository creates a new in-memory encounter repository
func NewInMemoryRepository() Repository {
	return &inMemoryRepository{
		encounters: make(map[string]*entities.Encounter),
		bySession:  make(map[string][]string),
		byMessage:  make(map[string]string),
	}
}

// Create stores a new encounter
func (r *inMemoryRepository) Create(ctx context.Context, encounter *entities.Encounter) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if _, exists := r.encounters[encounter.ID]; exists {
		return fmt.Errorf("encounter with ID %s already exists", encounter.ID)
	}
	
	r.encounters[encounter.ID] = encounter
	
	// Add to session index
	r.bySession[encounter.SessionID] = append(r.bySession[encounter.SessionID], encounter.ID)
	
	// Add to message index if message ID is set
	if encounter.MessageID != "" {
		r.byMessage[encounter.MessageID] = encounter.ID
	}
	
	return nil
}

// Get retrieves an encounter by ID
func (r *inMemoryRepository) Get(ctx context.Context, id string) (*entities.Encounter, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	encounter, exists := r.encounters[id]
	if !exists {
		return nil, fmt.Errorf("encounter not found: %s", id)
	}
	
	return encounter, nil
}

// Update modifies an existing encounter
func (r *inMemoryRepository) Update(ctx context.Context, encounter *entities.Encounter) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if _, exists := r.encounters[encounter.ID]; !exists {
		return fmt.Errorf("encounter not found: %s", encounter.ID)
	}
	
	// Update message index if changed
	oldEncounter := r.encounters[encounter.ID]
	if oldEncounter.MessageID != encounter.MessageID {
		if oldEncounter.MessageID != "" {
			delete(r.byMessage, oldEncounter.MessageID)
		}
		if encounter.MessageID != "" {
			r.byMessage[encounter.MessageID] = encounter.ID
		}
	}
	
	r.encounters[encounter.ID] = encounter
	return nil
}

// Delete removes an encounter
func (r *inMemoryRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	encounter, exists := r.encounters[id]
	if !exists {
		return fmt.Errorf("encounter not found: %s", id)
	}
	
	delete(r.encounters, id)
	
	// Remove from session index
	sessionEncounters := r.bySession[encounter.SessionID]
	for i, eid := range sessionEncounters {
		if eid == id {
			r.bySession[encounter.SessionID] = append(sessionEncounters[:i], sessionEncounters[i+1:]...)
			break
		}
	}
	
	// Remove from message index
	if encounter.MessageID != "" {
		delete(r.byMessage, encounter.MessageID)
	}
	
	return nil
}

// GetBySession retrieves all encounters for a session
func (r *inMemoryRepository) GetBySession(ctx context.Context, sessionID string) ([]*entities.Encounter, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	encounterIDs := r.bySession[sessionID]
	encounters := make([]*entities.Encounter, 0, len(encounterIDs))
	
	for _, id := range encounterIDs {
		if encounter, exists := r.encounters[id]; exists {
			encounters = append(encounters, encounter)
		}
	}
	
	return encounters, nil
}

// GetActiveBySession retrieves the active encounter for a session
func (r *inMemoryRepository) GetActiveBySession(ctx context.Context, sessionID string) (*entities.Encounter, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	encounterIDs := r.bySession[sessionID]
	
	for _, id := range encounterIDs {
		if encounter, exists := r.encounters[id]; exists {
			if encounter.Status == entities.EncounterStatusActive || 
			   encounter.Status == entities.EncounterStatusSetup ||
			   encounter.Status == entities.EncounterStatusRolling {
				return encounter, nil
			}
		}
	}
	
	return nil, fmt.Errorf("no active encounter in session: %s", sessionID)
}

// GetByMessage retrieves an encounter by Discord message ID
func (r *inMemoryRepository) GetByMessage(ctx context.Context, messageID string) (*entities.Encounter, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	encounterID, exists := r.byMessage[messageID]
	if !exists {
		return nil, fmt.Errorf("encounter not found for message: %s", messageID)
	}
	
	encounter, exists := r.encounters[encounterID]
	if !exists {
		return nil, fmt.Errorf("encounter not found: %s", encounterID)
	}
	
	return encounter, nil
}