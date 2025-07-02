package encounter

import (
	"context"
	"testing"

	mockdice "github.com/KirkDiggler/dnd-bot-discord/internal/dice/mock"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/conditions"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/damage"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/events"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/game/combat"
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

func TestMonsterAttackWithDisadvantage(t *testing.T) {
	// Create gomock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create event bus
	eventBus := events.NewEventBus()

	// Create condition service
	condSvc := condService.NewService(eventBus)

	// Create a test character
	testChar := &character.Character{
		ID:               "char-1",
		OwnerID:          "player-1",
		Name:             "TestHero",
		CurrentHitPoints: 30,
		MaxHitPoints:     30,
		Level:            1,
		AC:               15, // Set explicit AC
		Class:            &rulebook.Class{Key: "fighter"},
		Attributes: map[shared.Attribute]*character.AbilityScore{
			shared.AttributeStrength:  {Score: 16, Bonus: 3},
			shared.AttributeDexterity: {Score: 14, Bonus: 2},
		},
	}

	// Set up mock character service
	mockCharSvc := mockchar.NewMockService(ctrl)
	mockCharSvc.EXPECT().GetByID("char-1").Return(testChar, nil).AnyTimes()
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
				CharacterID: "char-1",
			},
		},
	}
	mockSessSvc.EXPECT().GetSession(gomock.Any(), "session-1").Return(mockSession, nil).AnyTimes()

	// Create manual dice roller for predictable results
	diceRoller := mockdice.NewManualMockRoller()

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
		Name:      "Test Monster Disadvantage",
		UserID:    "dm-1",
	})
	require.NoError(t, err)

	// Add the player
	playerCombatant, err := svc.AddPlayer(context.Background(), encounter.ID, "player-1", "char-1")
	require.NoError(t, err)

	// Add a skeleton monster with an attack
	skeleton, err := svc.AddMonster(context.Background(), encounter.ID, "dm-1", &AddMonsterInput{
		Name:  "Skeleton",
		MaxHP: 13,
		AC:    13,
		Actions: []*combat.MonsterAction{
			{
				Name:        "Shortsword",
				AttackBonus: 4,
				Damage: []*damage.Damage{
					{
						DiceCount:  1,
						DiceSize:   6,
						Bonus:      2,
						DamageType: damage.TypePiercing,
					},
				},
			},
		},
	})
	require.NoError(t, err)

	// Roll initiative (player: 15, skeleton: 10)
	diceRoller.SetRolls([]int{15, 10})
	err = svc.RollInitiative(context.Background(), encounter.ID, "dm-1")
	require.NoError(t, err)

	// Start encounter
	err = svc.StartEncounter(context.Background(), encounter.ID, "dm-1")
	require.NoError(t, err)

	// Apply disadvantage condition to the skeleton (simulating Vicious Mockery)
	_, err = condSvc.AddCondition(
		skeleton.ID,
		conditions.DisadvantageOnNextAttack,
		"Vicious Mockery",
		conditions.DurationEndOfNextTurn,
		1,
	)
	require.NoError(t, err)

	// Verify condition was applied
	assert.True(t, condSvc.HasCondition(skeleton.ID, conditions.DisadvantageOnNextAttack))

	// Advance to skeleton's turn
	err = svc.NextTurn(context.Background(), encounter.ID, "dm-1")
	require.NoError(t, err)

	// Verify it's now the skeleton's turn
	updatedEnc, err := svc.GetEncounter(context.Background(), encounter.ID)
	require.NoError(t, err)
	currentCombatant := updatedEnc.GetCurrentCombatant()
	require.NotNil(t, currentCombatant)
	assert.Equal(t, skeleton.ID, currentCombatant.ID, "Should be skeleton's turn (current: %s)", currentCombatant.Name)

	// Set up dice rolls for disadvantage attack
	// First roll: 18 (would hit), Second roll: 5 (miss), Damage roll: 4
	diceRoller.SetRolls([]int{18, 5, 4})

	// Perform skeleton's attack on the player
	result, err := svc.PerformAttack(context.Background(), &AttackInput{
		EncounterID: encounter.ID,
		AttackerID:  skeleton.ID,
		TargetID:    playerCombatant.ID,
		UserID:      "dm-1",
		ActionIndex: 0,
	})
	require.NoError(t, err)

	// Verify the attack used disadvantage
	assert.Equal(t, 5, result.AttackRoll, "Should use the lower roll (5) due to disadvantage")
	assert.Equal(t, 4, result.AttackBonus)
	assert.Equal(t, 9, result.TotalAttack, "Total should be 5 + 4 = 9")
	assert.False(t, result.Hit, "Attack should miss (9 < 15 AC)")

	// Verify condition was removed after the attack
	assert.False(t, condSvc.HasCondition(skeleton.ID, conditions.DisadvantageOnNextAttack))
}

func TestMonsterAttackWithDisadvantageHits(t *testing.T) {
	// Create gomock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create event bus
	eventBus := events.NewEventBus()

	// Create condition service
	condSvc := condService.NewService(eventBus)

	// Create a test character with low AC
	testChar := &character.Character{
		ID:               "char-1",
		OwnerID:          "player-1",
		Name:             "TestHero",
		CurrentHitPoints: 30,
		MaxHitPoints:     30,
		Level:            1,
		AC:               10, // Low AC to ensure hit even with disadvantage
		Class:            &rulebook.Class{Key: "fighter"},
		Attributes: map[shared.Attribute]*character.AbilityScore{
			shared.AttributeStrength:  {Score: 16, Bonus: 3},
			shared.AttributeDexterity: {Score: 10, Bonus: 0},
		},
	}

	// Set up mock character service
	mockCharSvc := mockchar.NewMockService(ctrl)
	mockCharSvc.EXPECT().GetByID("char-1").Return(testChar, nil).AnyTimes()
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
				CharacterID: "char-1",
			},
		},
	}
	mockSessSvc.EXPECT().GetSession(gomock.Any(), "session-1").Return(mockSession, nil).AnyTimes()

	// Create manual dice roller for predictable results
	diceRoller := mockdice.NewManualMockRoller()

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
		Name:      "Test Monster Disadvantage Hit",
		UserID:    "dm-1",
	})
	require.NoError(t, err)

	// Add a skeleton monster first so it gets the first initiative roll
	skeleton, err := svc.AddMonster(context.Background(), encounter.ID, "dm-1", &AddMonsterInput{
		Name:  "Skeleton",
		MaxHP: 13,
		AC:    13,
		Actions: []*combat.MonsterAction{
			{
				Name:        "Shortsword",
				AttackBonus: 4,
				Damage: []*damage.Damage{
					{
						DiceCount:  1,
						DiceSize:   6,
						Bonus:      2,
						DamageType: damage.TypePiercing,
					},
				},
			},
		},
	})
	require.NoError(t, err)

	// Add the player second
	playerCombatant, err := svc.AddPlayer(context.Background(), encounter.ID, "player-1", "char-1")
	require.NoError(t, err)

	// Roll initiative (first roll for skeleton: 15, second roll for player: 10)
	diceRoller.SetRolls([]int{15, 10})
	err = svc.RollInitiative(context.Background(), encounter.ID, "dm-1")
	require.NoError(t, err)

	// Start encounter
	err = svc.StartEncounter(context.Background(), encounter.ID, "dm-1")
	require.NoError(t, err)

	// Apply disadvantage condition to the skeleton
	_, err = condSvc.AddCondition(
		skeleton.ID,
		conditions.DisadvantageOnNextAttack,
		"Vicious Mockery",
		conditions.DurationEndOfNextTurn,
		1,
	)
	require.NoError(t, err)

	// Verify it's the skeleton's turn (skeleton has higher initiative)
	updatedEnc, err := svc.GetEncounter(context.Background(), encounter.ID)
	require.NoError(t, err)
	currentCombatant := updatedEnc.GetCurrentCombatant()
	require.NotNil(t, currentCombatant)
	assert.Equal(t, skeleton.ID, currentCombatant.ID, "Should be skeleton's turn (current: %s)", currentCombatant.Name)

	// Set up dice rolls for disadvantage attack that still hits
	// First roll: 15, Second roll: 8 (both would hit AC 10), Damage roll: 4
	diceRoller.SetRolls([]int{15, 8, 4})

	// Perform skeleton's attack on the player
	result, err := svc.PerformAttack(context.Background(), &AttackInput{
		EncounterID: encounter.ID,
		AttackerID:  skeleton.ID,
		TargetID:    playerCombatant.ID,
		UserID:      "dm-1",
		ActionIndex: 0,
	})
	require.NoError(t, err)

	// Verify the attack used disadvantage and hit
	assert.Equal(t, 8, result.AttackRoll, "Should use the lower roll (8) due to disadvantage")
	assert.Equal(t, 4, result.AttackBonus)
	assert.Equal(t, 12, result.TotalAttack, "Total should be 8 + 4 = 12")
	assert.True(t, result.Hit, "Attack should hit (12 >= 10 AC)")
	assert.Equal(t, 6, result.Damage, "Damage should be 4 (roll) + 2 (bonus) = 6")

	// Verify condition was removed after the attack
	assert.False(t, condSvc.HasCondition(skeleton.ID, conditions.DisadvantageOnNextAttack))
}
