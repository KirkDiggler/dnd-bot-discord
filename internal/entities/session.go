package entities

import (
	"time"
)

// SessionStatus represents the current state of a game session
type SessionStatus string

const (
	SessionStatusPlanning SessionStatus = "planning" // Session is being set up
	SessionStatusActive   SessionStatus = "active"   // Session is in progress
	SessionStatusPaused   SessionStatus = "paused"   // Session is temporarily paused
	SessionStatusEnded    SessionStatus = "ended"    // Session has concluded
)

// SessionMemberRole represents a member's role in the session
type SessionMemberRole string

const (
	SessionRoleDM        SessionMemberRole = "dm"        // Dungeon Master
	SessionRolePlayer    SessionMemberRole = "player"    // Player
	SessionRoleSpectator SessionMemberRole = "spectator" // Observer (can't participate)
)

// Session represents a D&D game session
type Session struct {
	ID                string                    `json:"id"`
	Name              string                    `json:"name"`
	Description       string                    `json:"description"`
	RealmID           string                    `json:"realm_id"`   // Discord server ID
	ChannelID         string                    `json:"channel_id"` // Discord channel where session is active
	CreatorID         string                    `json:"creator_id"` // User who created the session
	DMID              string                    `json:"dm_id"`      // Current DM (can be different from creator)
	Status            SessionStatus             `json:"status"`
	InviteCode        string                    `json:"invite_code"` // Unique code for joining
	Members           map[string]*SessionMember `json:"members"`     // UserID -> Member
	Settings          *SessionSettings          `json:"settings"`
	Encounters        []string                  `json:"encounters"` // List of encounter IDs
	ActiveEncounterID *string                   `json:"active_encounter_id,omitempty"`
	Metadata          map[string]interface{}    `json:"metadata,omitempty"` // Custom metadata for the session
	CreatedAt         time.Time                 `json:"created_at"`
	StartedAt         *time.Time                `json:"started_at"`
	EndedAt           *time.Time                `json:"ended_at"`
	LastActive        time.Time                 `json:"last_active"`
}

// SessionMember represents a participant in a session
type SessionMember struct {
	UserID      string            `json:"user_id"`
	CharacterID string            `json:"character_id"` // Selected character for this session
	Role        SessionMemberRole `json:"role"`
	JoinedAt    time.Time         `json:"joined_at"`
	LastActive  time.Time         `json:"last_active"`
	IsActive    bool              `json:"is_active"`  // Currently in the session
	TurnOrder   int               `json:"turn_order"` // Initiative order (-1 if not in combat)
}

// SessionSettings holds configuration for a session
type SessionSettings struct {
	MaxPlayers        int      `json:"max_players"`
	AllowSpectators   bool     `json:"allow_spectators"`
	RequireInvite     bool     `json:"require_invite"`       // If false, anyone can join with code
	AutoEndAfterHours int      `json:"auto_end_after_hours"` // Auto-end session after inactivity
	AllowLateJoin     bool     `json:"allow_late_join"`      // Can players join after session starts
	RestrictedContent []string `json:"restricted_content"`   // Restricted sourcebooks/content
}

// NewSession creates a new session with default settings
func NewSession(id, name, realmID, channelID, creatorID string) *Session {
	now := time.Now()
	return &Session{
		ID:         id,
		Name:       name,
		RealmID:    realmID,
		ChannelID:  channelID,
		CreatorID:  creatorID,
		DMID:       creatorID, // Creator is DM by default
		Status:     SessionStatusPlanning,
		Members:    make(map[string]*SessionMember),
		Settings:   DefaultSessionSettings(),
		CreatedAt:  now,
		LastActive: now,
	}
}

// DefaultSessionSettings returns default session configuration
func DefaultSessionSettings() *SessionSettings {
	return &SessionSettings{
		MaxPlayers:        6,
		AllowSpectators:   true,
		RequireInvite:     false,
		AutoEndAfterHours: 24,
		AllowLateJoin:     true,
		RestrictedContent: []string{},
	}
}

// AddMember adds a new member to the session
func (s *Session) AddMember(userID string, role SessionMemberRole) *SessionMember {
	member := &SessionMember{
		UserID:     userID,
		Role:       role,
		JoinedAt:   time.Now(),
		LastActive: time.Now(),
		IsActive:   true,
		TurnOrder:  -1,
	}
	s.Members[userID] = member
	s.LastActive = time.Now()
	return member
}

// RemoveMember removes a member from the session
func (s *Session) RemoveMember(userID string) {
	delete(s.Members, userID)
	s.LastActive = time.Now()
}

// SetCharacter assigns a character to a session member
func (s *Session) SetCharacter(userID, characterID string) bool {
	if member, ok := s.Members[userID]; ok {
		member.CharacterID = characterID
		member.LastActive = time.Now()
		s.LastActive = time.Now()
		return true
	}
	return false
}

// Start begins the session
func (s *Session) Start() bool {
	if s.Status != SessionStatusPlanning {
		return false
	}

	now := time.Now()
	s.Status = SessionStatusActive
	s.StartedAt = &now
	s.LastActive = now
	return true
}

// End concludes the session
func (s *Session) End() bool {
	if s.Status == SessionStatusEnded {
		return false
	}

	now := time.Now()
	s.Status = SessionStatusEnded
	s.EndedAt = &now
	s.LastActive = now
	return true
}

// Pause temporarily pauses the session
func (s *Session) Pause() bool {
	if s.Status != SessionStatusActive {
		return false
	}

	s.Status = SessionStatusPaused
	s.LastActive = time.Now()
	return true
}

// Resume resumes a paused session
func (s *Session) Resume() bool {
	if s.Status != SessionStatusPaused {
		return false
	}

	s.Status = SessionStatusActive
	s.LastActive = time.Now()
	return true
}

// GetActivePlayers returns all active players in the session
func (s *Session) GetActivePlayers() []*SessionMember {
	var players []*SessionMember
	for _, member := range s.Members {
		if member.Role == SessionRolePlayer && member.IsActive {
			players = append(players, member)
		}
	}
	return players
}

// GetDM returns the current DM member
func (s *Session) GetDM() *SessionMember {
	for _, member := range s.Members {
		if member.Role == SessionRoleDM {
			return member
		}
	}
	return nil
}

// CanJoin checks if a user can join the session
func (s *Session) CanJoin() bool {
	if s.Status == SessionStatusEnded {
		return false
	}

	if s.Status == SessionStatusActive && !s.Settings.AllowLateJoin {
		return false
	}

	activePlayerCount := 0
	for _, member := range s.Members {
		if member.Role == SessionRolePlayer && member.IsActive {
			activePlayerCount++
		}
	}

	return activePlayerCount < s.Settings.MaxPlayers
}

// IsUserInSession checks if a user is already in the session
func (s *Session) IsUserInSession(userID string) bool {
	_, exists := s.Members[userID]
	return exists
}

// UpdateActivity updates the last active timestamp
func (s *Session) UpdateActivity() {
	s.LastActive = time.Now()
}
