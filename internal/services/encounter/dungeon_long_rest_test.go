package encounter

import (
	"context"
	"testing"

	"github.com/KirkDiggler/dnd-bot-discord/internal/entities"
	encountermock "github.com/KirkDiggler/dnd-bot-discord/internal/repositories/encounters/mock"
	mockcharacters "github.com/KirkDiggler/dnd-bot-discord/internal/services/character/mock"
	sessionmock "github.com/KirkDiggler/dnd-bot-discord/internal/services/session/mock"
	"github.com/KirkDiggler/dnd-bot-discord/internal/uuid/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestAddPlayer_DungeonLongRest(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := encountermock.NewMockRepository(ctrl)
	mockCharService := mockcharacters.NewMockService(ctrl)
	mockSessionService := sessionmock.NewMockService(ctrl)
	mockUUID := mocks.NewMockGenerator(ctrl)

	svc := &service{
		repository:       mockRepo,
		characterService: mockCharService,
		sessionService:   mockSessionService,
		uuidGenerator:    mockUUID,
	}

	ctx := context.Background()
	encounterID := "enc123"
	sessionID := "session123"
	playerID := "player123"
	characterID := "char123"
	combatantID := "combatant123"

	// Create a barbarian character with used rage
	character := &entities.Character{
		ID:               characterID,
		OwnerID:          playerID,
		Name:             "Grognak",
		Level:            3,
		CurrentHitPoints: 20,
		MaxHitPoints:     30,
		AC:               14,
		Class: &entities.Class{
			Key:  "barbarian",
			Name: "Barbarian",
		},
		Attributes: map[entities.Attribute]*entities.AbilityScore{
			entities.AttributeDexterity: {Score: 14, Bonus: 2},
		},
		Resources: &entities.CharacterResources{
			HP: entities.HPResource{
				Current: 20,
				Max:     30,
			},
			Abilities: map[string]*entities.ActiveAbility{
				"rage": {
					Key:           "rage",
					Name:          "Rage",
					UsesMax:       3, // Level 3 barbarian gets 3 rages
					UsesRemaining: 1, // Used 2 rages already
				},
			},
		},
	}

	// Create a dungeon session
	dungeonSession := &entities.Session{
		ID: sessionID,
		Metadata: map[string]interface{}{
			"sessionType": "dungeon",
		},
	}

	encounter := &entities.Encounter{
		ID:         encounterID,
		SessionID:  sessionID,
		Combatants: make(map[string]*entities.Combatant),
	}

	// Setup expectations
	mockRepo.EXPECT().Get(ctx, encounterID).Return(encounter, nil)
	mockCharService.EXPECT().GetByID(characterID).Return(character, nil)
	mockSessionService.EXPECT().GetSession(ctx, sessionID).Return(dungeonSession, nil)

	// Expect the character to be saved after long rest
	mockCharService.EXPECT().UpdateEquipment(gomock.Any()).DoAndReturn(func(char *entities.Character) error {
		// Verify the character's resources were reset
		assert.Equal(t, 3, char.Resources.Abilities["rage"].UsesRemaining, "Rage uses should be reset to max")
		assert.Equal(t, 30, char.Resources.HP.Current, "HP should be restored to max")
		return nil
	})

	mockUUID.EXPECT().New().Return(combatantID)
	mockRepo.EXPECT().Update(ctx, gomock.Any()).Return(nil)

	// Execute
	combatant, err := svc.AddPlayer(ctx, encounterID, playerID, characterID)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, combatant)
	assert.Equal(t, "Grognak", combatant.Name)

	// Verify the character's resources were reset (in memory)
	assert.Equal(t, 3, character.Resources.Abilities["rage"].UsesRemaining, "Rage uses should be reset to max")
	assert.Equal(t, 30, character.Resources.HP.Current, "HP should be restored to max")
}

func TestAddPlayer_NonDungeonNoLongRest(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := encountermock.NewMockRepository(ctrl)
	mockCharService := mockcharacters.NewMockService(ctrl)
	mockSessionService := sessionmock.NewMockService(ctrl)
	mockUUID := mocks.NewMockGenerator(ctrl)

	svc := &service{
		repository:       mockRepo,
		characterService: mockCharService,
		sessionService:   mockSessionService,
		uuidGenerator:    mockUUID,
	}

	ctx := context.Background()
	encounterID := "enc123"
	sessionID := "session123"
	playerID := "player123"
	characterID := "char123"
	combatantID := "combatant123"

	// Create a barbarian character with used rage
	character := &entities.Character{
		ID:               characterID,
		OwnerID:          playerID,
		Name:             "Grognak",
		Level:            3,
		CurrentHitPoints: 20,
		MaxHitPoints:     30,
		AC:               14,
		Class: &entities.Class{
			Key:  "barbarian",
			Name: "Barbarian",
		},
		Attributes: map[entities.Attribute]*entities.AbilityScore{
			entities.AttributeDexterity: {Score: 14, Bonus: 2},
		},
		Resources: &entities.CharacterResources{
			HP: entities.HPResource{
				Current: 20,
				Max:     30,
			},
			Abilities: map[string]*entities.ActiveAbility{
				"rage": {
					Key:           "rage",
					Name:          "Rage",
					UsesMax:       3,
					UsesRemaining: 1, // Used 2 rages
				},
			},
		},
	}

	// Create a regular combat session (not dungeon)
	regularSession := &entities.Session{
		ID: sessionID,
		Metadata: map[string]interface{}{
			"sessionType": "combat",
		},
	}

	encounter := &entities.Encounter{
		ID:         encounterID,
		SessionID:  sessionID,
		Combatants: make(map[string]*entities.Combatant),
	}

	// Setup expectations
	mockRepo.EXPECT().Get(ctx, encounterID).Return(encounter, nil)
	mockCharService.EXPECT().GetByID(characterID).Return(character, nil)
	mockSessionService.EXPECT().GetSession(ctx, sessionID).Return(regularSession, nil)

	// Should NOT save character since no long rest is performed
	// mockCharService.EXPECT().UpdateEquipment(gomock.Any()).Times(0)

	mockUUID.EXPECT().New().Return(combatantID)
	mockRepo.EXPECT().Update(ctx, gomock.Any()).Return(nil)

	// Execute
	combatant, err := svc.AddPlayer(ctx, encounterID, playerID, characterID)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, combatant)

	// Verify the character's resources were NOT reset
	assert.Equal(t, 1, character.Resources.Abilities["rage"].UsesRemaining, "Rage uses should not be reset")
	assert.Equal(t, 20, character.Resources.HP.Current, "HP should not be restored")
}
