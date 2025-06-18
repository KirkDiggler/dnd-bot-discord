package character

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/KirkDiggler/dnd-bot-discord/internal/entities"
	dnderr "github.com/KirkDiggler/dnd-bot-discord/internal/errors"
)

// sessionMutex protects the in-memory session store
var sessionMutex sync.RWMutex

// StartCharacterCreation starts a new character creation session
func (s *service) StartCharacterCreation(ctx context.Context, userID, guildID string) (*entities.CharacterCreationSession, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, dnderr.InvalidArgument("user ID is required")
	}
	if strings.TrimSpace(guildID) == "" {
		return nil, dnderr.InvalidArgument("guild ID is required")
	}

	// Clean up any existing sessions for this user
	sessionMutex.Lock()
	for id, session := range s.sessions {
		if session.UserID == userID && session.GuildID == guildID {
			log.Printf("Cleaning up existing session %s for user %s", id, userID)
			delete(s.sessions, id)
		}
	}
	sessionMutex.Unlock()

	// Create a new draft character
	character := &entities.Character{
		ID:      generateID(),
		OwnerID: userID,
		RealmID: guildID,
		Name:    "Draft Character",
		Status:  entities.CharacterStatusDraft,
		Level:   1,
		// Initialize empty maps
		Attributes:    make(map[entities.Attribute]*entities.AbilityScore),
		Proficiencies: make(map[entities.ProficiencyType][]*entities.Proficiency),
		Inventory:     make(map[entities.EquipmentType][]entities.Equipment),
		EquippedSlots: make(map[entities.Slot]entities.Equipment),
	}

	// Save to repository
	if err := s.repository.Create(ctx, character); err != nil {
		return nil, dnderr.Wrap(err, "failed to create draft character").
			WithMeta("character_id", character.ID).
			WithMeta("owner_id", userID)
	}

	// Create session
	session := &entities.CharacterCreationSession{
		ID:          fmt.Sprintf("session_%d", time.Now().UnixNano()),
		UserID:      userID,
		GuildID:     guildID,
		CharacterID: character.ID,
		CurrentStep: entities.StepRaceSelection,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		ExpiresAt:   time.Now().Add(1 * time.Hour), // 1 hour expiration
	}

	// Store session
	sessionMutex.Lock()
	s.sessions[session.ID] = session
	sessionMutex.Unlock()

	log.Printf("Started character creation session %s for user %s with character %s", session.ID, userID, character.ID)
	return session, nil
}

// GetCharacterCreationSession retrieves an active session
func (s *service) GetCharacterCreationSession(ctx context.Context, sessionID string) (*entities.CharacterCreationSession, error) {
	if strings.TrimSpace(sessionID) == "" {
		return nil, dnderr.InvalidArgument("session ID is required")
	}

	sessionMutex.RLock()
	session, exists := s.sessions[sessionID]
	sessionMutex.RUnlock()

	if !exists {
		return nil, dnderr.NotFound("session not found").
			WithMeta("session_id", sessionID)
	}

	// Check if expired
	if session.IsExpired() {
		sessionMutex.Lock()
		delete(s.sessions, sessionID)
		sessionMutex.Unlock()
		return nil, dnderr.InvalidArgument("session has expired").
			WithMeta("session_id", sessionID)
	}

	return session, nil
}

// UpdateCharacterCreationSession updates the session step
func (s *service) UpdateCharacterCreationSession(ctx context.Context, sessionID, step string) error {
	if strings.TrimSpace(sessionID) == "" {
		return dnderr.InvalidArgument("session ID is required")
	}
	if strings.TrimSpace(step) == "" {
		return dnderr.InvalidArgument("step is required")
	}

	sessionMutex.Lock()
	defer sessionMutex.Unlock()

	session, exists := s.sessions[sessionID]
	if !exists {
		return dnderr.NotFound("session not found").
			WithMeta("session_id", sessionID)
	}

	// Check if expired
	if session.IsExpired() {
		delete(s.sessions, sessionID)
		return dnderr.InvalidArgument("session has expired").
			WithMeta("session_id", sessionID)
	}

	// Update step
	session.UpdateStep(step)
	log.Printf("Updated session %s to step %s", sessionID, step)
	return nil
}

// GetCharacterFromSession gets the character associated with a session
func (s *service) GetCharacterFromSession(ctx context.Context, sessionID string) (*entities.Character, error) {
	// Get the session
	session, err := s.GetCharacterCreationSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	// Get the character
	char, err := s.repository.Get(ctx, session.CharacterID)
	if err != nil {
		return nil, dnderr.Wrapf(err, "failed to get character from session").
			WithMeta("session_id", sessionID).
			WithMeta("character_id", session.CharacterID)
	}

	return char, nil
}

// cleanupExpiredSessions removes expired sessions (should be called periodically)
func (s *service) cleanupExpiredSessions() {
	sessionMutex.Lock()
	defer sessionMutex.Unlock()

	now := time.Now()
	for id, session := range s.sessions {
		if now.After(session.ExpiresAt) {
			log.Printf("Cleaning up expired session %s", id)
			delete(s.sessions, id)
		}
	}
}