package encounter

import (
	"context"
	"testing"

	mockdice "github.com/KirkDiggler/dnd-bot-discord/internal/dice/mock"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/conditions"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/events"
	gameSession "github.com/KirkDiggler/dnd-bot-discord/internal/domain/game/session"
	rulebook "github.com/KirkDiggler/dnd-bot-discord/internal/domain/rulebook/dnd5e"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/shared"
	"github.com/KirkDiggler/dnd-bot-discord/internal/repositories/encounters"
	mockchar "github.com/KirkDiggler/dnd-bot-discord/internal/services/character/mock"
	condService "github.com/KirkDiggler/dnd-bot-discord/internal/services/condition"
	mocksess "github.com/KirkDiggler/dnd-bot-discord/internal/services/session/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestEncounterService_ConditionIntegration(t *testing.T) {
	// Create gomock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create event bus
	eventBus := events.NewEventBus()

	// Create condition service
	condSvc := condService.NewService(eventBus)

	// Create characters
	char1 := &character.Character{
		ID:               "char-1",
		OwnerID:          "player-1",
		Name:             "TestChar",
		CurrentHitPoints: 20,
		MaxHitPoints:     20,
		Level:            3,
		Class:            &rulebook.Class{Key: "fighter"},
		Attributes: map[shared.Attribute]*character.AbilityScore{
			shared.AttributeDexterity: {Score: 14, Bonus: 2},
		},
	}

	// Set up mock character service
	mockCharSvc := mockchar.NewMockService(ctrl)
	mockCharSvc.EXPECT().GetByID("char-1").Return(char1, nil).AnyTimes()
	mockCharSvc.EXPECT().UpdateEquipment(gomock.Any()).Return(nil).AnyTimes()

	// Set up mock session service
	mockSessSvc := mocksess.NewMockService(ctrl)
	mockSession := &gameSession.Session{
		ID:   "session-1",
		DMID: "dm-1",
		Members: map[string]*gameSession.SessionMember{
			"dm-1": {
				UserID: "dm-1",
				Role:   gameSession.SessionRoleDM,
			},
		},
	}
	mockSessSvc.EXPECT().GetSession(gomock.Any(), "session-1").Return(mockSession, nil).AnyTimes()

	// Create dice roller
	diceRoller := mockdice.NewManualMockRoller()
	// Set enough rolls for initiative and other needs
	diceRoller.SetRolls([]int{10, 10, 10, 10, 10, 10, 10, 10, 10, 10})

	// Create encounter service
	svc := NewService(&ServiceConfig{
		Repository:       encounters.NewInMemoryRepository(),
		SessionService:   mockSessSvc,
		CharacterService: mockCharSvc,
		ConditionService: condSvc,
		DiceRoller:       diceRoller,
		EventBus:         eventBus,
	})

	// Create encounter
	encounter, err := svc.CreateEncounter(context.Background(), &CreateEncounterInput{
		SessionID: "session-1",
		Name:      "Test Combat",
		UserID:    "dm-1",
	})
	require.NoError(t, err)

	// Add player
	combatant, err := svc.AddPlayer(context.Background(), encounter.ID, "player-1", "char-1")
	require.NoError(t, err)
	assert.Equal(t, "TestChar", combatant.Name)

	// Add a monster so turns can advance
	monster, err := svc.AddMonster(context.Background(), encounter.ID, "dm-1", &AddMonsterInput{
		Name:  "Goblin",
		MaxHP: 10,
		AC:    13,
	})
	require.NoError(t, err)
	assert.Equal(t, "Goblin", monster.Name)

	// Apply a condition directly through the service
	t.Run("apply condition directly", func(t *testing.T) {
		// Apply condition directly through the service
		cond, err := condSvc.AddCondition("char-1", conditions.Poisoned, "test", conditions.DurationRounds, 3)
		require.NoError(t, err)
		assert.NotNil(t, cond)

		// Check if condition was applied
		conds := condSvc.GetConditions("char-1")
		require.Len(t, conds, 1)
		assert.Equal(t, conditions.Poisoned, conds[0].Type)
		assert.Equal(t, 3, conds[0].Remaining)
	})

	// Start combat and process rounds
	t.Run("process condition durations", func(t *testing.T) {
		// Roll initiative first
		err := svc.RollInitiative(context.Background(), encounter.ID, "dm-1")
		require.NoError(t, err)

		// Start combat
		err = svc.StartEncounter(context.Background(), encounter.ID, "dm-1")
		require.NoError(t, err)

		// Process a round by advancing turn
		err = svc.NextTurn(context.Background(), encounter.ID, "dm-1")
		require.NoError(t, err)

		// Check condition - should still be active (3 rounds remaining)
		conds := condSvc.GetConditions("char-1")
		require.Len(t, conds, 1)
		assert.Equal(t, 3, conds[0].Remaining)

		// Advance to complete first round (we have 2 combatants, so 2 turns = 1 round)
		err = svc.NextTurn(context.Background(), encounter.ID, "dm-1")
		require.NoError(t, err)

		// Now a full round has passed - check condition should have 2 rounds remaining
		conds = condSvc.GetConditions("char-1")
		require.Len(t, conds, 1)
		assert.Equal(t, 2, conds[0].Remaining)
	})

	// Test damage-based condition removal
	t.Run("remove condition on damage", func(t *testing.T) {
		// Apply a condition that ends on damage
		_, err := condSvc.AddCondition("char-1", conditions.DisadvantageOnNextAttack, "spell", conditions.DurationUntilDamaged, 0)
		require.NoError(t, err)

		// Verify condition exists
		assert.True(t, condSvc.HasCondition("char-1", conditions.DisadvantageOnNextAttack))

		// Apply damage
		err = svc.ApplyDamage(context.Background(), encounter.ID, combatant.ID, "dm-1", 5)
		require.NoError(t, err)

		// Condition should be removed
		assert.False(t, condSvc.HasCondition("char-1", conditions.DisadvantageOnNextAttack))
	})

	// Test condition effects
	t.Run("get active effects", func(t *testing.T) {
		// Clear previous conditions
		err := condSvc.RemoveConditionByType("char-1", conditions.Poisoned)
		require.NoError(t, err)
		err = condSvc.RemoveConditionByType("char-1", conditions.DisadvantageOnNextAttack)
		require.NoError(t, err)

		// Apply stunned condition
		_, err = condSvc.AddCondition("char-1", conditions.Stunned, "spell", conditions.DurationRounds, 1)
		require.NoError(t, err)

		// Get active effects
		effects := condSvc.GetActiveEffects("char-1")
		assert.NotNil(t, effects)
		assert.True(t, effects.Incapacitated)
		assert.True(t, effects.CantAct)
		assert.True(t, effects.CantReact)
	})
}
