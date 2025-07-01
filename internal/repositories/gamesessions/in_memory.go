package gamesessions

import (
	"context"
	"fmt"
	"sync"

	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/game/session"
)

// inMemoryRepository implements Repository using in-memory storage
type inMemoryRepository struct {
	mu          sync.RWMutex
	sessions    map[string]*session.Session
	inviteCodes map[string]string // inviteCode -> sessionID
}

// NewInMemoryRepository creates a new in-memory session repository
func NewInMemoryRepository() Repository {
	return &inMemoryRepository{
		sessions:    make(map[string]*session.Session),
		inviteCodes: make(map[string]string),
	}
}

// Create creates a new session
func (r *inMemoryRepository) Create(ctx context.Context, sess *session.Session) error {
	if sess == nil {
		return fmt.Errorf("session cannot be nil")
	}
	if sess.ID == "" {
		return fmt.Errorf("session ID cannot be empty")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.sessions[sess.ID]; exists {
		return fmt.Errorf("session with ID %s already exists", sess.ID)
	}

	if sess.InviteCode != "" {
		if existingID, exists := r.inviteCodes[sess.InviteCode]; exists {
			return fmt.Errorf("invite code %s already in use by session %s", sess.InviteCode, existingID)
		}
		r.inviteCodes[sess.InviteCode] = sess.ID
	}

	// Create a deep copy to avoid external modifications
	sessionCopy := *sess
	r.sessions[sess.ID] = &sessionCopy

	return nil
}

// Get retrieves a session by ID
func (r *inMemoryRepository) Get(ctx context.Context, id string) (*session.Session, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	sess, exists := r.sessions[id]
	if !exists {
		return nil, fmt.Errorf("session not found: %s", id)
	}

	// Return a copy to avoid external modifications
	sessCopy := *sess
	return &sessCopy, nil
}

// GetByInviteCode retrieves a session by its invite code
func (r *inMemoryRepository) GetByInviteCode(ctx context.Context, code string) (*session.Session, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	sessionID, exists := r.inviteCodes[code]
	if !exists {
		return nil, fmt.Errorf("no session found with invite code: %s", code)
	}

	sess, exists := r.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}

	// Return a copy to avoid external modifications
	sessCopy := *sess
	return &sessCopy, nil
}

// Update updates an existing session
func (r *inMemoryRepository) Update(ctx context.Context, sess *session.Session) error {
	if sess == nil {
		return fmt.Errorf("session cannot be nil")
	}
	if sess.ID == "" {
		return fmt.Errorf("session ID cannot be empty")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	existing, exists := r.sessions[sess.ID]
	if !exists {
		return fmt.Errorf("session not found: %s", sess.ID)
	}

	// Handle invite code changes
	if existing.InviteCode != sess.InviteCode {
		// Remove old invite code mapping
		if existing.InviteCode != "" {
			delete(r.inviteCodes, existing.InviteCode)
		}

		// Add new invite code mapping
		if sess.InviteCode != "" {
			if existingID, exists := r.inviteCodes[sess.InviteCode]; exists && existingID != sess.ID {
				return fmt.Errorf("invite code %s already in use by session %s", sess.InviteCode, existingID)
			}
			r.inviteCodes[sess.InviteCode] = sess.ID
		}
	}

	// Update with a copy
	sessionCopy := *sess
	r.sessions[sess.ID] = &sessionCopy

	return nil
}

// Delete removes a session
func (r *inMemoryRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	sess, exists := r.sessions[id]
	if !exists {
		return fmt.Errorf("session not found: %s", id)
	}

	// Remove invite code mapping
	if sess.InviteCode != "" {
		delete(r.inviteCodes, sess.InviteCode)
	}

	delete(r.sessions, id)
	return nil
}

// GetByRealm retrieves all sessions for a realm
func (r *inMemoryRepository) GetByRealm(ctx context.Context, realmID string) ([]*session.Session, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var sessions []*session.Session
	for _, session := range r.sessions {
		if session.RealmID == realmID {
			sessionCopy := *session
			sessions = append(sessions, &sessionCopy)
		}
	}

	return sessions, nil
}

// GetByUser retrieves all sessions a user is part of
func (r *inMemoryRepository) GetByUser(ctx context.Context, userID string) ([]*session.Session, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var sessions []*session.Session
	for _, session := range r.sessions {
		if _, exists := session.Members[userID]; exists {
			sessionCopy := *session
			sessions = append(sessions, &sessionCopy)
		}
	}

	return sessions, nil
}

// GetActiveByRealm retrieves all active sessions for a realm
func (r *inMemoryRepository) GetActiveByRealm(ctx context.Context, realmID string) ([]*session.Session, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var sessions []*session.Session
	for _, sessionValue := range r.sessions {
		if sessionValue.RealmID == realmID &&
			(sessionValue.Status == session.SessionStatusPlanning ||
				sessionValue.Status == session.SessionStatusActive ||
				sessionValue.Status == session.SessionStatusPaused) {
			sessionCopy := *sessionValue
			sessions = append(sessions, &sessionCopy)
		}
	}

	return sessions, nil
}

// GetActiveByUser retrieves all active sessions a user is part of
func (r *inMemoryRepository) GetActiveByUser(ctx context.Context, userID string) ([]*session.Session, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var sessions []*session.Session
	for _, sessionValue := range r.sessions {
		if _, exists := sessionValue.Members[userID]; exists &&
			(sessionValue.Status == session.SessionStatusPlanning ||
				sessionValue.Status == session.SessionStatusActive ||
				sessionValue.Status == session.SessionStatusPaused) {
			sessionCopy := *sessionValue
			sessions = append(sessions, &sessionCopy)
		}
	}

	return sessions, nil
}
