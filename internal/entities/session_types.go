package entities

// SessionType represents the type of game session
type SessionType string

const (
	// SessionTypeDungeon represents a dungeon crawl session
	SessionTypeDungeon SessionType = "dungeon"

	// SessionTypeCombat represents a standard combat encounter
	SessionTypeCombat SessionType = "combat"

	// SessionTypeRoleplay represents a roleplay-focused session
	SessionTypeRoleplay SessionType = "roleplay"

	// SessionTypeOneShot represents a one-shot adventure
	SessionTypeOneShot SessionType = "oneshot"
)

// String returns the string representation of the session type
func (st SessionType) String() string {
	return string(st)
}

// IsValid checks if the session type is valid
func (st SessionType) IsValid() bool {
	switch st {
	case SessionTypeDungeon, SessionTypeCombat, SessionTypeRoleplay, SessionTypeOneShot:
		return true
	default:
		return false
	}
}

// MetadataKeys defines standard metadata keys used across the application
type MetadataKey string

const (
	// Session metadata keys
	MetadataKeySessionType  MetadataKey = "sessionType"
	MetadataKeyDifficulty   MetadataKey = "difficulty"
	MetadataKeyRoomNumber   MetadataKey = "roomNumber"
	MetadataKeyLobbyMessage MetadataKey = "lobbyMessageID"
	MetadataKeyLobbyChannel MetadataKey = "lobbyChannelID"

	// Character metadata keys
	MetadataKeyLevel      MetadataKey = "level"
	MetadataKeyExperience MetadataKey = "experience"

	// Combat metadata keys
	MetadataKeyInitiative MetadataKey = "initiative"
	MetadataKeyRound      MetadataKey = "round"
)
