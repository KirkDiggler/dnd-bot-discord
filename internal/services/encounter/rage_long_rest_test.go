package encounter

import (
	"context"
	"testing"

	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/game/combat"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/game/session"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/rulebook/dnd5e"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/shared"
	encountermock "github.com/KirkDiggler/dnd-bot-discord/internal/repositories/encounters/mock"
	mockcharacters "github.com/KirkDiggler/dnd-bot-discord/internal/services/character/mock"
	sessionmock "github.com/KirkDiggler/dnd-bot-discord/internal/services/session/mock"
	"github.com/KirkDiggler/dnd-bot-discord/internal/uuid/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestAddPlayer_RageEffectsClearedOnDungeonEntry(t *testing.T) {
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

	// Create a barbarian with active rage and rage effects
	char := &character.Character{
		ID:               characterID,
		OwnerID:          playerID,
		Name:             "Grunk",
		Level:            5,
		CurrentHitPoints: 30,
		MaxHitPoints:     45,
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
				Current: 30,
				Max:     45,
			},
			Abilities: map[string]*shared.ActiveAbility{
				shared.AbilityKeyRage: {
					Key:           shared.AbilityKeyRage,
					Name:          "Rage",
					UsesMax:       3,
					UsesRemaining: 1,    // Used 2 rages
					IsActive:      true, // Currently raging
					Duration:      5,    // 5 rounds remaining
					RestType:      shared.RestTypeLong,
				},
			},
			ActiveEffects: []*shared.ActiveEffect{
				{
					ID:           "rage-effect-old",
					Name:         "Rage",
					Source:       "barbarian_rage",
					DurationType: shared.DurationTypeRounds,
					Duration:     5,
					Modifiers: []shared.Modifier{
						{
							Type:        shared.ModifierTypeDamageBonus,
							Value:       2,
							DamageTypes: []string{"melee"},
						},
						{
							Type:        shared.ModifierTypeDamageResistance,
							Value:       1,
							DamageTypes: []string{"bludgeoning", "piercing", "slashing"},
						},
					},
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
	mockCharService.EXPECT().UpdateEquipment(gomock.Any()).DoAndReturn(func(c *character.Character) error {
		// Verify rage is completely reset
		rage := c.Resources.Abilities[shared.AbilityKeyRage]
		assert.False(t, rage.IsActive, "Rage should be deactivated after long rest")
		assert.Equal(t, 0, rage.Duration, "Rage duration should be reset")
		assert.Equal(t, 3, rage.UsesRemaining, "Rage uses should be restored to max")

		// Verify all rage effects are cleared
		assert.Empty(t, c.Resources.ActiveEffects, "All rage effects should be cleared after long rest")

		// Verify HP is restored
		assert.Equal(t, 45, c.Resources.HP.Current, "HP should be restored to max")

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
	assert.Equal(t, "Grunk", combatant.Name)

	// Verify the character's state (in memory)
	rage := char.Resources.Abilities[shared.AbilityKeyRage]
	assert.False(t, rage.IsActive, "Rage should be deactivated")
	assert.Equal(t, 0, rage.Duration, "Rage duration should be reset")
	assert.Equal(t, 3, rage.UsesRemaining, "Rage uses should be restored")
	assert.Empty(t, char.Resources.ActiveEffects, "All rage effects should be cleared")
	assert.Equal(t, 45, char.Resources.HP.Current, "HP should be restored")
}
