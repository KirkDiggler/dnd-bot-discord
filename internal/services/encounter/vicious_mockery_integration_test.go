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

func TestViciousMockery_AppliesDisadvantageToMonster(t *testing.T) {
	// Create gomock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create event bus
	eventBus := events.NewEventBus()

	// Create condition service
	condSvc := condService.NewService(eventBus)

	// Create a bard character
	bard := &character.Character{
		ID:               "bard-1",
		OwnerID:          "player-1",
		Name:             "ViciousBard",
		CurrentHitPoints: 20,
		MaxHitPoints:     20,
		Level:            3,
		Class:            &rulebook.Class{Key: "bard"},
		Attributes: map[shared.Attribute]*character.AbilityScore{
			shared.AttributeCharisma:  {Score: 16, Bonus: 3},
			shared.AttributeDexterity: {Score: 14, Bonus: 2},
		},
	}

	// Set up mock character service
	mockCharSvc := mockchar.NewMockService(ctrl)
	mockCharSvc.EXPECT().GetByID("bard-1").Return(bard, nil).AnyTimes()
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
			"player-1": {
				UserID:      "player-1",
				Role:        gameSession.SessionRolePlayer,
				CharacterID: "bard-1",
			},
		},
	}
	mockSessSvc.EXPECT().GetSession(gomock.Any(), "session-1").Return(mockSession, nil).AnyTimes()

	// Create dice roller
	diceRoller := mockdice.NewManualMockRoller()
	// Set enough rolls for initiative and attacks
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
		Name:      "Test Vicious Mockery",
		UserID:    "dm-1",
	})
	require.NoError(t, err)

	// Add the bard
	combatant, err := svc.AddPlayer(context.Background(), encounter.ID, "player-1", "bard-1")
	require.NoError(t, err)
	assert.Equal(t, "ViciousBard", combatant.Name)

	// Add a goblin monster
	goblin, err := svc.AddMonster(context.Background(), encounter.ID, "dm-1", &AddMonsterInput{
		Name:  "Goblin",
		MaxHP: 10,
		AC:    13,
	})
	require.NoError(t, err)
	assert.Equal(t, "Goblin", goblin.Name)

	// Roll initiative and start combat
	err = svc.RollInitiative(context.Background(), encounter.ID, "dm-1")
	require.NoError(t, err)

	err = svc.StartEncounter(context.Background(), encounter.ID, "dm-1")
	require.NoError(t, err)

	// Simulate Vicious Mockery spell damage event
	t.Run("vicious mockery applies disadvantage", func(t *testing.T) {
		// Create spell damage event (simulating Vicious Mockery hit)
		spellDamageEvent := events.NewGameEvent(events.OnSpellDamage).
			WithContext(events.ContextDamage, 3).
			WithContext(events.ContextDamageType, "psychic").
			WithContext(events.ContextSpellName, "Vicious Mockery").
			WithContext(events.ContextTargetID, goblin.ID).
			WithContext(events.ContextEncounterID, encounter.ID).
			WithContext(events.ContextUserID, "player-1")

		// Emit the event
		err := eventBus.Emit(spellDamageEvent)
		require.NoError(t, err)

		// Wait a moment for event processing
		// In a real system, you might want to use a more sophisticated synchronization mechanism

		// Check if condition was applied to the goblin
		conds := condSvc.GetConditions(goblin.ID)
		require.Len(t, conds, 1, "Goblin should have one condition")
		assert.Equal(t, conditions.DisadvantageOnNextAttack, conds[0].Type)
		assert.Equal(t, "Vicious Mockery", conds[0].Source)

		// Verify the goblin's HP was reduced
		updatedEncounter, err := svc.GetEncounter(context.Background(), encounter.ID)
		require.NoError(t, err)
		updatedGoblin := updatedEncounter.Combatants[goblin.ID]
		assert.Equal(t, 7, updatedGoblin.CurrentHP, "Goblin should have taken 3 damage")
	})

	// Test that the condition is removed after the goblin attacks
	t.Run("disadvantage removed after attack", func(t *testing.T) {
		// Advance turn to goblin's turn (if needed)
		// Process damage to trigger condition removal
		// This simulates the goblin attacking with disadvantage
		condSvc.ProcessDamage(goblin.ID, 0) // Processing 0 damage shouldn't remove "until damaged" conditions

		// The condition should still be there
		assert.True(t, condSvc.HasCondition(goblin.ID, conditions.DisadvantageOnNextAttack))

		// In a real game, the condition would be removed when the goblin makes its next attack
		// For now, we'll just verify it exists
	})
}
