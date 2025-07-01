package encounter

import (
	"context"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/game/combat"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/game/session"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/rulebook"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/shared"
	"testing"

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
	char := &character.Character{
		ID:               characterID,
		OwnerID:          playerID,
		Name:             "Grognak",
		Level:            3,
		CurrentHitPoints: 20,
		MaxHitPoints:     30,
		AC:               14,
		Class: &rulebook.Class{
			Key:  "barbarian",
			Name: "Barbarian",
		},
		Attributes: map[shared.Attribute]*character.AbilityScore{
			shared.AttributeDexterity: {Score: 14, Bonus: 2},
		},
		Resources: &character.CharacterResources{
			HP: shared.HPResource{
				Current: 20,
				Max:     30,
			},
			Abilities: map[string]*shared.ActiveAbility{
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
	dungeonSession := &session.Session{
		ID: sessionID,
		Metadata: map[string]interface{}{
			"sessionType": "dungeon",
		},
	}

	encounter := &combat.Encounter{
		ID:         encounterID,
		SessionID:  sessionID,
		Combatants: make(map[string]*combat.Combatant),
	}

	// Setup expectations
	mockRepo.EXPECT().Get(ctx, encounterID).Return(encounter, nil)
	mockCharService.EXPECT().GetByID(characterID).Return(char, nil)
	mockSessionService.EXPECT().GetSession(ctx, sessionID).Return(dungeonSession, nil)

	// Expect the character to be saved after long rest
	mockCharService.EXPECT().UpdateEquipment(gomock.Any()).DoAndReturn(func(char *character.Character) error {
		// Verify the character's resources were reset
		assert.Equal(t, 3, char.Resources.Abilities["rage"].UsesRemaining, "Rage uses should be reset to max")
		assert.Equal(t, 30, char.Resources.HP.Current, "HP should be restored to max")
		return nil
	})

	// Expect the character to be saved after action economy reset
	mockCharService.EXPECT().UpdateEquipment(gomock.Any()).Return(nil)

	mockUUID.EXPECT().New().Return(combatantID)
	mockRepo.EXPECT().Update(ctx, gomock.Any()).Return(nil)

	// Execute
	combatant, err := svc.AddPlayer(ctx, encounterID, playerID, characterID)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, combatant)
	assert.Equal(t, "Grognak", combatant.Name)

	// Verify the character's resources were reset (in memory)
	assert.Equal(t, 3, char.Resources.Abilities["rage"].UsesRemaining, "Rage uses should be reset to max")
	assert.Equal(t, 30, char.Resources.HP.Current, "HP should be restored to max")
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
	char := &character.Character{
		ID:               characterID,
		OwnerID:          playerID,
		Name:             "Grognak",
		Level:            3,
		CurrentHitPoints: 20,
		MaxHitPoints:     30,
		AC:               14,
		Class: &rulebook.Class{
			Key:  "barbarian",
			Name: "Barbarian",
		},
		Attributes: map[shared.Attribute]*character.AbilityScore{
			shared.AttributeDexterity: {Score: 14, Bonus: 2},
		},
		Resources: &character.CharacterResources{
			HP: shared.HPResource{
				Current: 20,
				Max:     30,
			},
			Abilities: map[string]*shared.ActiveAbility{
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
	regularSession := &session.Session{
		ID: sessionID,
		Metadata: map[string]interface{}{
			"sessionType": "combat",
		},
	}

	encounter := &combat.Encounter{
		ID:         encounterID,
		SessionID:  sessionID,
		Combatants: make(map[string]*combat.Combatant),
	}

	// Setup expectations
	mockRepo.EXPECT().Get(ctx, encounterID).Return(encounter, nil)
	mockCharService.EXPECT().GetByID(characterID).Return(char, nil)
	mockSessionService.EXPECT().GetSession(ctx, sessionID).Return(regularSession, nil)

	// Expect the character to be saved after action economy reset (but no long rest)
	mockCharService.EXPECT().UpdateEquipment(gomock.Any()).Return(nil)

	mockUUID.EXPECT().New().Return(combatantID)
	mockRepo.EXPECT().Update(ctx, gomock.Any()).Return(nil)

	// Execute
	combatant, err := svc.AddPlayer(ctx, encounterID, playerID, characterID)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, combatant)

	// Verify the character's resources were NOT reset
	assert.Equal(t, 1, char.Resources.Abilities["rage"].UsesRemaining, "Rage uses should not be reset")
	assert.Equal(t, 20, char.Resources.HP.Current, "HP should not be restored")
}
