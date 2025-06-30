package session

import (
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/shared"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSession_IsDungeon(t *testing.T) {
	tests := []struct {
		name     string
		session  *Session
		expected bool
	}{
		{
			name: "dungeon session",
			session: &Session{
				Metadata: shared.Metadata{
					string(MetadataKeySessionType): string(SessionTypeDungeon),
				},
			},
			expected: true,
		},
		{
			name: "combat session",
			session: &Session{
				Metadata: shared.Metadata{
					string(MetadataKeySessionType): string(SessionTypeCombat),
				},
			},
			expected: false,
		},
		{
			name:     "nil metadata",
			session:  &Session{},
			expected: false,
		},
		{
			name: "empty metadata",
			session: &Session{
				Metadata: shared.Metadata{},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.session.IsDungeon())
		})
	}
}

func TestSession_GetSetSessionType(t *testing.T) {
	session := &Session{}

	// Initially empty
	assert.Equal(t, SessionType(""), session.GetSessionType())

	// Set dungeon type
	session.SetSessionType(SessionTypeDungeon)
	assert.Equal(t, SessionTypeDungeon, session.GetSessionType())
	assert.True(t, session.IsDungeon())

	// Change to combat
	session.SetSessionType(SessionTypeCombat)
	assert.Equal(t, SessionTypeCombat, session.GetSessionType())
	assert.False(t, session.IsDungeon())
}

func TestSession_GetDifficulty(t *testing.T) {
	tests := []struct {
		name     string
		session  *Session
		expected string
	}{
		{
			name: "has difficulty",
			session: &Session{
				Metadata: shared.Metadata{
					string(MetadataKeyDifficulty): "hard",
				},
			},
			expected: "hard",
		},
		{
			name:     "nil metadata returns default",
			session:  &Session{},
			expected: "medium",
		},
		{
			name: "empty metadata returns default",
			session: &Session{
				Metadata: shared.Metadata{},
			},
			expected: "medium",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.session.GetDifficulty())
		})
	}
}

func TestSession_GetRoomNumber(t *testing.T) {
	tests := []struct {
		name     string
		session  *Session
		expected int
	}{
		{
			name: "has room number",
			session: &Session{
				Metadata: shared.Metadata{
					string(MetadataKeyRoomNumber): 5,
				},
			},
			expected: 5,
		},
		{
			name: "float64 room number (from JSON)",
			session: &Session{
				Metadata: shared.Metadata{
					string(MetadataKeyRoomNumber): float64(3),
				},
			},
			expected: 3,
		},
		{
			name:     "nil metadata returns default",
			session:  &Session{},
			expected: 1,
		},
		{
			name: "empty metadata returns default",
			session: &Session{
				Metadata: shared.Metadata{},
			},
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.session.GetRoomNumber())
		})
	}
}

func TestSessionType_IsValid(t *testing.T) {
	tests := []struct {
		sessionType SessionType
		expected    bool
	}{
		{SessionTypeDungeon, true},
		{SessionTypeCombat, true},
		{SessionTypeRoleplay, true},
		{SessionTypeOneShot, true},
		{SessionType("invalid"), false},
		{SessionType(""), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.sessionType), func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.sessionType.IsValid())
		})
	}
}
