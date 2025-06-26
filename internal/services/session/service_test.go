package session_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/KirkDiggler/dnd-bot-discord/internal/entities"
	mockchar "github.com/KirkDiggler/dnd-bot-discord/internal/services/character/mock"
	"github.com/KirkDiggler/dnd-bot-discord/internal/services/session"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// MockRepository is a simple in-memory repository for testing
type MockRepository struct {
	sessions map[string]*entities.Session
}

func NewMockRepository() *MockRepository {
	return &MockRepository{
		sessions: make(map[string]*entities.Session),
	}
}

func (m *MockRepository) Create(_ context.Context, sess *entities.Session) error {
	m.sessions[sess.ID] = sess
	return nil
}

func (m *MockRepository) Get(_ context.Context, id string) (*entities.Session, error) {
	sess, exists := m.sessions[id]
	if !exists {
		return nil, fmt.Errorf("session not found")
	}
	return sess, nil
}

func (m *MockRepository) Update(_ context.Context, sess *entities.Session) error {
	m.sessions[sess.ID] = sess
	return nil
}

func (m *MockRepository) Delete(_ context.Context, id string) error {
	delete(m.sessions, id)
	return nil
}

func (m *MockRepository) ListByOwner(_ context.Context, ownerID string) ([]*entities.Session, error) {
	var result []*entities.Session
	for _, sess := range m.sessions {
		if sess.CreatorID == ownerID {
			result = append(result, sess)
		}
	}
	return result, nil
}

func (m *MockRepository) GetActiveByRealm(_ context.Context, realmID string) ([]*entities.Session, error) {
	var result []*entities.Session
	for _, sess := range m.sessions {
		if sess.RealmID == realmID && sess.Status == entities.SessionStatusActive {
			result = append(result, sess)
		}
	}
	return result, nil
}

func (m *MockRepository) GetActiveByUser(_ context.Context, userID string) ([]*entities.Session, error) {
	var result []*entities.Session
	for _, sess := range m.sessions {
		if sess.Status == entities.SessionStatusActive {
			// Check if user is a member
			if _, isMember := sess.Members[userID]; isMember {
				result = append(result, sess)
			}
			// Also check if user is the creator
			if sess.CreatorID == userID {
				result = append(result, sess)
			}
		}
	}
	return result, nil
}

func (m *MockRepository) GetByRealm(_ context.Context, realmID string) ([]*entities.Session, error) {
	var result []*entities.Session
	for _, sess := range m.sessions {
		if sess.RealmID == realmID {
			result = append(result, sess)
		}
	}
	return result, nil
}

func (m *MockRepository) GetByUser(_ context.Context, userID string) ([]*entities.Session, error) {
	var result []*entities.Session
	for _, sess := range m.sessions {
		// Check if user is a member
		if _, isMember := sess.Members[userID]; isMember {
			result = append(result, sess)
		}
		// Also check if user is the creator
		if sess.CreatorID == userID {
			result = append(result, sess)
		}
	}
	return result, nil
}

func (m *MockRepository) GetByInviteCode(_ context.Context, code string) (*entities.Session, error) {
	for _, sess := range m.sessions {
		if sess.InviteCode == code {
			return sess, nil
		}
	}
	return nil, nil
}

func (m *MockRepository) ListUserSessions(_ context.Context, userID string) ([]*entities.Session, error) {
	var result []*entities.Session
	for _, sess := range m.sessions {
		if _, exists := sess.Members[userID]; exists {
			result = append(result, sess)
		}
	}
	return result, nil
}

// MockUUIDGenerator for testing
type MockUUIDGenerator struct {
	prefix  string
	counter int
}

func NewMockUUIDGenerator(prefix string) *MockUUIDGenerator {
	return &MockUUIDGenerator{prefix: prefix, counter: 0}
}

func (m *MockUUIDGenerator) New() string {
	m.counter++
	return m.prefix + "-" + string(rune('0'+m.counter))
}

func TestSelectCharacter(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockRepository()
	mockCharService := mockchar.NewMockService(ctrl)

	svc := session.NewService(&session.ServiceConfig{
		Repository:       mockRepo,
		UUIDGenerator:    NewMockUUIDGenerator("uuid"),
		CharacterService: mockCharService,
	})

	// Create a test session with a player
	sessionID := "session-123"
	playerID := "player-123"
	testSession := &entities.Session{
		ID:        sessionID,
		Name:      "Test Dungeon",
		CreatorID: "dm-123",
		Members: map[string]*entities.SessionMember{
			playerID: {
				UserID: playerID,
				Role:   entities.SessionRolePlayer,
				// CharacterID is initially empty
			},
		},
		Status: entities.SessionStatusActive,
	}
	mockRepo.Create(context.Background(), testSession)

	t.Run("Successfully selects character for player", func(t *testing.T) {
		characterID := "char-123"

		// Mock the character service to return a character that belongs to the player
		mockCharService.EXPECT().
			GetByID(characterID).
			Return(&entities.Character{
				ID:      characterID,
				OwnerID: playerID,
			}, nil)

		// Select character
		err := svc.SelectCharacter(context.Background(), sessionID, playerID, characterID)
		require.NoError(t, err)

		// Verify character was set
		sess, err := mockRepo.Get(context.Background(), sessionID)
		require.NoError(t, err)
		require.NotNil(t, sess)

		member, exists := sess.Members[playerID]
		require.True(t, exists)
		assert.Equal(t, characterID, member.CharacterID)
	})

	t.Run("Updates existing character selection", func(t *testing.T) {
		newCharacterID := "char-456"

		// Mock the character service to return a character that belongs to the player
		mockCharService.EXPECT().
			GetByID(newCharacterID).
			Return(&entities.Character{
				ID:      newCharacterID,
				OwnerID: playerID,
			}, nil)

		// Select different character
		err := svc.SelectCharacter(context.Background(), sessionID, playerID, newCharacterID)
		require.NoError(t, err)

		// Verify character was updated
		sess, err := mockRepo.Get(context.Background(), sessionID)
		require.NoError(t, err)

		member := sess.Members[playerID]
		assert.Equal(t, newCharacterID, member.CharacterID)
	})

	t.Run("Fails when session doesn't exist", func(t *testing.T) {
		// Mock the character service - it gets called before session check
		mockCharService.EXPECT().
			GetByID("char-789").
			Return(&entities.Character{
				ID:      "char-789",
				OwnerID: playerID,
			}, nil)

		err := svc.SelectCharacter(context.Background(), "invalid-session", playerID, "char-789")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "session not found")
	})

	t.Run("Fails when player not in session", func(t *testing.T) {
		// Mock the character service to return a character
		mockCharService.EXPECT().
			GetByID("char-789").
			Return(&entities.Character{
				ID:      "char-789",
				OwnerID: "not-a-member",
			}, nil)

		err := svc.SelectCharacter(context.Background(), sessionID, "not-a-member", "char-789")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not in the session")
	})

	t.Run("Fails when character doesn't belong to player", func(t *testing.T) {
		// Mock the character service to return a character that belongs to a different player
		mockCharService.EXPECT().
			GetByID("char-890").
			Return(&entities.Character{
				ID:      "char-890",
				OwnerID: "different-player",
			}, nil)

		err := svc.SelectCharacter(context.Background(), sessionID, playerID, "char-890")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "does not belong to user")
	})

	t.Run("Preserves character selection across session updates", func(t *testing.T) {
		// Set a character
		characterID := "char-persistent"
		
		// Mock the character service to return a character that belongs to the player
		mockCharService.EXPECT().
			GetByID(characterID).
			Return(&entities.Character{
				ID:      characterID,
				OwnerID: playerID,
			}, nil)
		
		err := svc.SelectCharacter(context.Background(), sessionID, playerID, characterID)
		require.NoError(t, err)

		// Update session (simulate other operations)
		sess, _ := mockRepo.Get(context.Background(), sessionID)
		sess.Status = entities.SessionStatusPaused
		mockRepo.Update(context.Background(), sess)

		// Character should still be set
		updatedSess, _ := mockRepo.Get(context.Background(), sessionID)
		member := updatedSess.Members[playerID]
		assert.Equal(t, characterID, member.CharacterID)
	})
}

func TestSessionMemberCharacterPersistence(t *testing.T) {
	// Test that character IDs persist properly in session members

	t.Run("Character ID set in member struct", func(t *testing.T) {
		member := &entities.SessionMember{
			UserID:      "user-123",
			Role:        entities.SessionRolePlayer,
			CharacterID: "char-123",
		}

		assert.Equal(t, "char-123", member.CharacterID)
	})

	t.Run("Empty character ID for new members", func(t *testing.T) {
		member := &entities.SessionMember{
			UserID: "user-456",
			Role:   entities.SessionRolePlayer,
			// CharacterID not set
		}

		assert.Empty(t, member.CharacterID)
	})
}
