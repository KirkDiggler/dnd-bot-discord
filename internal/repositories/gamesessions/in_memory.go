package gamesessions

import (
	"context"
	"fmt"
	"sync"

	"github.com/KirkDiggler/dnd-bot-discord/internal/entities"
)

// inMemoryRepository implements Repository using in-memory storage
type inMemoryRepository struct {
	mu          sync.RWMutex
	sessions    map[string]*entities.Session
	inviteCodes map[string]string // inviteCode -> sessionID
}

// NewInMemoryRepository creates a new in-memory session repository
func NewInMemoryRepository() Repository {
	return &inMemoryRepository{
		sessions:    make(map[string]*entities.Session),
		inviteCodes: make(map[string]string),
	}
}

// Create creates a new session
func (r *inMemoryRepository) Create(ctx context.Context, session *entities.Session) error {
	if session == nil {
		return fmt.Errorf("session cannot be nil")
	}
	if session.ID == "" {
		return fmt.Errorf("session ID cannot be empty")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.sessions[session.ID]; exists {
		return fmt.Errorf("session with ID %s already exists", session.ID)
	}

	if session.InviteCode != "" {
		if existingID, exists := r.inviteCodes[session.InviteCode]; exists {
			return fmt.Errorf("invite code %s already in use by session %s", session.InviteCode, existingID)
		}
		r.inviteCodes[session.InviteCode] = session.ID
	}

	// Create a deep copy to avoid external modifications
	sessionCopy := *session
	r.sessions[session.ID] = &sessionCopy

	return nil
}

// Get retrieves a session by ID
func (r *inMemoryRepository) Get(ctx context.Context, id string) (*entities.Session, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	session, exists := r.sessions[id]
	if !exists {
		return nil, fmt.Errorf("session not found: %s", id)
	}

	// Return a copy to avoid external modifications
	sessionCopy := *session
	return &sessionCopy, nil
}

// GetByInviteCode retrieves a session by its invite code
func (r *inMemoryRepository) GetByInviteCode(ctx context.Context, code string) (*entities.Session, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	sessionID, exists := r.inviteCodes[code]
	if !exists {
		return nil, fmt.Errorf("no session found with invite code: %s", code)
	}

	session, exists := r.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}

	// Return a copy to avoid external modifications
	sessionCopy := *session
	return &sessionCopy, nil
}

// Update updates an existing session
func (r *inMemoryRepository) Update(ctx context.Context, session *entities.Session) error {
	if session == nil {
		return fmt.Errorf("session cannot be nil")
	}
	if session.ID == "" {
		return fmt.Errorf("session ID cannot be empty")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	existing, exists := r.sessions[session.ID]
	if !exists {
		return fmt.Errorf("session not found: %s", session.ID)
	}

	// Handle invite code changes
	if existing.InviteCode != session.InviteCode {
		// Remove old invite code mapping
		if existing.InviteCode != "" {
			delete(r.inviteCodes, existing.InviteCode)
		}

		// Add new invite code mapping
		if session.InviteCode != "" {
			if existingID, exists := r.inviteCodes[session.InviteCode]; exists && existingID != session.ID {
				return fmt.Errorf("invite code %s already in use by session %s", session.InviteCode, existingID)
			}
			r.inviteCodes[session.InviteCode] = session.ID
		}
	}

	// Update with a copy
	sessionCopy := *session
	r.sessions[session.ID] = &sessionCopy

	return nil
}

// Delete removes a session
func (r *inMemoryRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	session, exists := r.sessions[id]
	if !exists {
		return fmt.Errorf("session not found: %s", id)
	}

	// Remove invite code mapping
	if session.InviteCode != "" {
		delete(r.inviteCodes, session.InviteCode)
	}

	delete(r.sessions, id)
	return nil
}

// GetByRealm retrieves all sessions for a realm
func (r *inMemoryRepository) GetByRealm(ctx context.Context, realmID string) ([]*entities.Session, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var sessions []*entities.Session
	for _, session := range r.sessions {
		if session.RealmID == realmID {
			sessionCopy := *session
			sessions = append(sessions, &sessionCopy)
		}
	}

	return sessions, nil
}

// GetByUser retrieves all sessions a user is part of
func (r *inMemoryRepository) GetByUser(ctx context.Context, userID string) ([]*entities.Session, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var sessions []*entities.Session
	for _, session := range r.sessions {
		if _, exists := session.Members[userID]; exists {
			sessionCopy := *session
			sessions = append(sessions, &sessionCopy)
		}
	}

	return sessions, nil
}

// GetActiveByRealm retrieves all active sessions for a realm
func (r *inMemoryRepository) GetActiveByRealm(ctx context.Context, realmID string) ([]*entities.Session, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var sessions []*entities.Session
	for _, session := range r.sessions {
		if session.RealmID == realmID &&
			(session.Status == entities.SessionStatusPlanning ||
				session.Status == entities.SessionStatusActive ||
				session.Status == entities.SessionStatusPaused) {
			sessionCopy := *session
			sessions = append(sessions, &sessionCopy)
		}
	}

	return sessions, nil
}

// GetActiveByUser retrieves all active sessions a user is part of
func (r *inMemoryRepository) GetActiveByUser(ctx context.Context, userID string) ([]*entities.Session, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var sessions []*entities.Session
	for _, session := range r.sessions {
		if _, exists := session.Members[userID]; exists &&
			(session.Status == entities.SessionStatusPlanning ||
				session.Status == entities.SessionStatusActive ||
				session.Status == entities.SessionStatusPaused) {
			sessionCopy := *session
			sessions = append(sessions, &sessionCopy)
		}
	}

	return sessions, nil
}
