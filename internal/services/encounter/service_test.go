package encounter_test

import (
	"context"
	"testing"

	"fmt"

	"github.com/KirkDiggler/dnd-bot-discord/internal/entities"
	"github.com/KirkDiggler/dnd-bot-discord/internal/services/encounter"
	mockcharacter "github.com/KirkDiggler/dnd-bot-discord/internal/services/character/mock"
	mocksession "github.com/KirkDiggler/dnd-bot-discord/internal/services/session/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// MockUUIDGenerator is a simple UUID generator for testing
type MockUUIDGenerator struct {
	prefix string
	counter int
}

func NewMockUUIDGenerator(prefix string) *MockUUIDGenerator {
	return &MockUUIDGenerator{prefix: prefix, counter: 0}
}

func (m *MockUUIDGenerator) New() string {
	m.counter++
	return fmt.Sprintf("%s-%d", m.prefix, m.counter)
}

// MockRepository is a simple in-memory repository for testing
type MockRepository struct {
	encounters map[string]*entities.Encounter
}

func NewMockRepository() *MockRepository {
	return &MockRepository{
		encounters: make(map[string]*entities.Encounter),
	}
}

func (m *MockRepository) Create(_ context.Context, encounter *entities.Encounter) error {
	m.encounters[encounter.ID] = encounter
	return nil
}

func (m *MockRepository) Get(_ context.Context, id string) (*entities.Encounter, error) {
	encounter, exists := m.encounters[id]
	if !exists {
		return nil, nil
	}
	return encounter, nil
}

func (m *MockRepository) Update(_ context.Context, encounter *entities.Encounter) error {
	m.encounters[encounter.ID] = encounter
	return nil
}

func (m *MockRepository) Delete(_ context.Context, id string) error {
	delete(m.encounters, id)
	return nil
}

func (m *MockRepository) GetActiveBySession(_ context.Context, sessionID string) (*entities.Encounter, error) {
	for _, encounter := range m.encounters {
		if encounter.SessionID == sessionID && 
			(encounter.Status == entities.EncounterStatusActive ||
			 encounter.Status == entities.EncounterStatusSetup ||
			 encounter.Status == entities.EncounterStatusRolling) {
			return encounter, nil
		}
	}
	return nil, nil // Important: return nil, nil when no active encounter exists
}

func (m *MockRepository) GetBySession(_ context.Context, sessionID string) ([]*entities.Encounter, error) {
	var encounters []*entities.Encounter
	for _, encounter := range m.encounters {
		if encounter.SessionID == sessionID {
			encounters = append(encounters, encounter)
		}
	}
	return encounters, nil
}

func (m *MockRepository) GetByMessage(_ context.Context, messageID string) (*entities.Encounter, error) {
	for _, encounter := range m.encounters {
		if encounter.MessageID == messageID {
			return encounter, nil
		}
	}
	return nil, nil
}

func TestCreateEncounter_DungeonSession(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSessionService := mocksession.NewMockService(ctrl)
	mockCharService := mockcharacter.NewMockService(ctrl)
	mockRepo := NewMockRepository()
	
	svc := encounter.NewService(&encounter.ServiceConfig{
		Repository:       mockRepo,
		SessionService:   mockSessionService,
		CharacterService: mockCharService,
		UUIDGenerator:    NewMockUUIDGenerator("enc"),
	})

	// Test case: Bot creates encounter for dungeon session
	t.Run("Bot can create encounter for dungeon session", func(t *testing.T) {
		botUserID := "bot-user-123"
		sessionID := "session-123"
		channelID := "channel-123"

		// Create a session with sessionType=dungeon in metadata
		dungeonSession := &entities.Session{
			ID:   sessionID,
			Name: "Test Dungeon",
			Members: map[string]*entities.SessionMember{
				// Bot is not a member or DM
			},
			Metadata: map[string]interface{}{
				"sessionType": "dungeon",
			},
		}

		// Mock session service to return dungeon session
		mockSessionService.EXPECT().
			GetSession(gomock.Any(), sessionID).
			Return(dungeonSession, nil)

		// Create encounter as bot
		encounter, err := svc.CreateEncounter(context.Background(), &encounter.CreateEncounterInput{
			SessionID:   sessionID,
			ChannelID:   channelID,
			Name:        "Dungeon Room 1",
			Description: "A dark room",
			UserID:      botUserID,
		})

		// Should succeed for dungeon sessions
		require.NoError(t, err)
		assert.NotNil(t, encounter)
		assert.Equal(t, "Dungeon Room 1", encounter.Name)
		assert.Equal(t, sessionID, encounter.SessionID)
		assert.Equal(t, botUserID, encounter.CreatedBy)
	})

	// Test case: Non-DM cannot create encounter for regular session
	t.Run("Non-DM cannot create encounter for regular session", func(t *testing.T) {
		playerUserID := "player-123"
		sessionID := "session-456"
		channelID := "channel-456"

		// Create a regular session without dungeon metadata
		regularSession := &entities.Session{
			ID:   sessionID,
			Name: "Regular Campaign",
			Members: map[string]*entities.SessionMember{
				playerUserID: {
					UserID: playerUserID,
					Role:   entities.SessionRolePlayer,
				},
			},
			Metadata: map[string]interface{}{},
		}

		// Mock session service to return regular session
		mockSessionService.EXPECT().
			GetSession(gomock.Any(), sessionID).
			Return(regularSession, nil)

		// Try to create encounter as player
		_, err := svc.CreateEncounter(context.Background(), &encounter.CreateEncounterInput{
			SessionID:   sessionID,
			ChannelID:   channelID,
			Name:        "Test Encounter",
			Description: "Should fail",
			UserID:      playerUserID,
		})

		// Should fail with permission denied
		require.Error(t, err)
		assert.Contains(t, err.Error(), "only the DM can create encounters")
	})

	// Test case: DM can create encounter for regular session
	t.Run("DM can create encounter for regular session", func(t *testing.T) {
		dmUserID := "dm-123"
		sessionID := "session-789"
		channelID := "channel-789"

		// Create a regular session with DM
		regularSession := &entities.Session{
			ID:   sessionID,
			Name: "DM Campaign",
			Members: map[string]*entities.SessionMember{
				dmUserID: {
					UserID: dmUserID,
					Role:   entities.SessionRoleDM,
				},
			},
			Metadata: map[string]interface{}{},
		}

		// Mock session service to return session with DM
		mockSessionService.EXPECT().
			GetSession(gomock.Any(), sessionID).
			Return(regularSession, nil)

		// Create encounter as DM
		encounter, err := svc.CreateEncounter(context.Background(), &encounter.CreateEncounterInput{
			SessionID:   sessionID,
			ChannelID:   channelID,
			Name:        "DM Encounter",
			Description: "Created by DM",
			UserID:      dmUserID,
		})

		// Should succeed for DM
		require.NoError(t, err)
		assert.NotNil(t, encounter)
		assert.Equal(t, "DM Encounter", encounter.Name)
		assert.Equal(t, dmUserID, encounter.CreatedBy)
	})
}