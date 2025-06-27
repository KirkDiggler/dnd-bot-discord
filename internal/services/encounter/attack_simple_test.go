package encounter_test

import (
	"context"
	"testing"
	"time"

	"github.com/KirkDiggler/dnd-bot-discord/internal/dice"
	"github.com/KirkDiggler/dnd-bot-discord/internal/entities"
	"github.com/KirkDiggler/dnd-bot-discord/internal/entities/damage"
	"github.com/KirkDiggler/dnd-bot-discord/internal/repositories/characters"
	"github.com/KirkDiggler/dnd-bot-discord/internal/repositories/encounters"
	"github.com/KirkDiggler/dnd-bot-discord/internal/repositories/gamesessions"
	"github.com/KirkDiggler/dnd-bot-discord/internal/services/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/services/encounter"
	"github.com/KirkDiggler/dnd-bot-discord/internal/services/session"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPerformAttack_MonsterVsMonster_WithMockDice(t *testing.T) {
	ctx := context.Background()
	mockDice := dice.NewMockRoller()

	// Set up deterministic rolls
	mockDice.SetRolls([]int{
		15, // Attack roll (15 + 4 = 19 vs AC 15 = HIT)
		6,  // Damage roll (1d6)
	})

	// Create services
	charRepo := characters.NewInMemoryRepository()
	charService := character.NewService(&character.ServiceConfig{
		Repository: charRepo,
	})

	sessionRepo := gamesessions.NewInMemoryRepository()
	sessionService := session.NewService(&session.ServiceConfig{
		Repository:       sessionRepo,
		CharacterService: charService,
	})

	encounterService := encounter.NewService(&encounter.ServiceConfig{
		Repository:       encounters.NewInMemoryRepository(),
		SessionService:   sessionService,
		CharacterService: charService,
		DiceRoller:       mockDice,
	})

	// Create a test session
	sess := &entities.Session{
		ID:        "test-session",
		Name:      "Test Session",
		ChannelID: "channel-1",
		CreatorID: "dm-1",
		DMID:      "dm-1",
		Members: map[string]*entities.SessionMember{
			"dm-1": {UserID: "dm-1", Role: entities.SessionRoleDM},
		},
		Status:     entities.SessionStatusActive,
		CreatedAt:  time.Now(),
		LastActive: time.Now(),
	}
	err := sessionRepo.Create(ctx, sess)
	require.NoError(t, err)

	// Create encounter
	enc, err := encounterService.CreateEncounter(ctx, &encounter.CreateEncounterInput{
		SessionID: "test-session",
		ChannelID: "channel-1",
		Name:      "Test Combat",
		UserID:    "dm-1",
	})
	require.NoError(t, err)

	// Add attacker monster with action
	attacker, err := encounterService.AddMonster(ctx, enc.ID, "dm-1", &encounter.AddMonsterInput{
		Name:  "Goblin",
		MaxHP: 7,
		AC:    15,
		Actions: []*entities.MonsterAction{
			{
				Name:        "Scimitar",
				AttackBonus: 4,
				Damage: []*damage.Damage{
					{
						DiceCount:  1,
						DiceSize:   6,
						Bonus:      2,
						DamageType: damage.TypeSlashing,
					},
				},
			},
		},
	})
	require.NoError(t, err)

	// Add target monster
	target, err := encounterService.AddMonster(ctx, enc.ID, "dm-1", &encounter.AddMonsterInput{
		Name:  "Orc",
		MaxHP: 15,
		AC:    13,
	})
	require.NoError(t, err)

	// Get updated encounter and manually set it up for combat
	enc, err = encounterService.GetEncounter(ctx, enc.ID)
	require.NoError(t, err)
	enc.Status = entities.EncounterStatusActive
	enc.Turn = 0
	enc.TurnOrder = []string{attacker.ID, target.ID}
	enc.Combatants[attacker.ID].Initiative = 15
	enc.Combatants[target.ID].Initiative = 10

	// Perform attack
	result, err := encounterService.PerformAttack(ctx, &encounter.AttackInput{
		EncounterID: enc.ID,
		AttackerID:  attacker.ID,
		TargetID:    target.ID,
		UserID:      "dm-1",
		ActionIndex: 0,
	})

	require.NoError(t, err)
	assert.NotNil(t, result)

	// Verify attack details
	assert.Equal(t, "Goblin", result.AttackerName)
	assert.Equal(t, "Orc", result.TargetName)
	assert.Equal(t, "Scimitar", result.WeaponName)
	assert.Equal(t, 15, result.AttackRoll)
	assert.Equal(t, 4, result.AttackBonus)
	assert.Equal(t, 19, result.TotalAttack)
	assert.Equal(t, 13, result.TargetAC)
	assert.True(t, result.Hit)
	assert.False(t, result.Critical)
	assert.Equal(t, 8, result.Damage) // 6 (roll) + 2 (bonus)
	assert.Equal(t, "slashing", result.DamageType)
	assert.Equal(t, 7, result.TargetNewHP) // 15 - 8 = 7
	assert.False(t, result.TargetDefeated)

	// Check the log entry with spoiler format
	assert.Contains(t, result.LogEntry, "⚔️ **Goblin** → **Orc**")
	assert.Contains(t, result.LogEntry, "HIT 🩸 **8**")
	assert.Contains(t, result.LogEntry, "||d20:15+4=19")
	assert.Contains(t, result.LogEntry, "vs AC:13")
	assert.Contains(t, result.LogEntry, "dmg:1d6: [6]||") // damage roll in spoiler
}

func TestPerformAttack_UnarmedStrike_WithMockDice(t *testing.T) {
	ctx := context.Background()
	mockDice := dice.NewMockRoller()

	// Set up deterministic rolls
	mockDice.SetRolls([]int{
		10, // Attack roll
		3,  // Damage roll (1d4)
	})

	// Create services
	charRepo := characters.NewInMemoryRepository()
	charService := character.NewService(&character.ServiceConfig{
		Repository: charRepo,
	})

	sessionRepo := gamesessions.NewInMemoryRepository()
	sessionService := session.NewService(&session.ServiceConfig{
		Repository:       sessionRepo,
		CharacterService: charService,
	})

	encounterService := encounter.NewService(&encounter.ServiceConfig{
		Repository:       encounters.NewInMemoryRepository(),
		SessionService:   sessionService,
		CharacterService: charService,
		DiceRoller:       mockDice,
	})

	// Create a test session
	sess := &entities.Session{
		ID:        "test-session",
		Name:      "Test Session",
		ChannelID: "channel-1",
		CreatorID: "dm-1",
		DMID:      "dm-1",
		Members: map[string]*entities.SessionMember{
			"dm-1": {UserID: "dm-1", Role: entities.SessionRoleDM},
		},
		Status:     entities.SessionStatusActive,
		CreatedAt:  time.Now(),
		LastActive: time.Now(),
	}
	err := sessionRepo.Create(ctx, sess)
	require.NoError(t, err)

	// Create encounter
	enc, err := encounterService.CreateEncounter(ctx, &encounter.CreateEncounterInput{
		SessionID: "test-session",
		ChannelID: "channel-1",
		Name:      "Test Combat",
		UserID:    "dm-1",
	})
	require.NoError(t, err)

	// Add two monsters without weapons (will use unarmed strike)
	attacker, err := encounterService.AddMonster(ctx, enc.ID, "dm-1", &encounter.AddMonsterInput{
		Name:  "Commoner",
		MaxHP: 4,
		AC:    10,
	})
	require.NoError(t, err)

	target, err := encounterService.AddMonster(ctx, enc.ID, "dm-1", &encounter.AddMonsterInput{
		Name:  "Peasant",
		MaxHP: 4,
		AC:    10,
	})
	require.NoError(t, err)

	// Get updated encounter and manually set it up for combat
	enc, err = encounterService.GetEncounter(ctx, enc.ID)
	require.NoError(t, err)
	enc.Status = entities.EncounterStatusActive
	enc.Turn = 0
	enc.TurnOrder = []string{attacker.ID, target.ID}
	enc.Combatants[attacker.ID].Initiative = 10
	enc.Combatants[target.ID].Initiative = 5

	// Perform unarmed attack (no ActionIndex since there are no actions)
	result, err := encounterService.PerformAttack(ctx, &encounter.AttackInput{
		EncounterID: enc.ID,
		AttackerID:  attacker.ID,
		TargetID:    target.ID,
		UserID:      "dm-1",
		ActionIndex: -1, // -1 or omit for unarmed
	})

	require.NoError(t, err)
	assert.NotNil(t, result)

	// Verify unarmed strike details
	assert.Equal(t, "Commoner", result.AttackerName)
	assert.Equal(t, "Peasant", result.TargetName)
	assert.Equal(t, "Unarmed Strike", result.WeaponName)
	assert.Equal(t, 10, result.AttackRoll)
	assert.Equal(t, 0, result.AttackBonus)
	assert.Equal(t, 10, result.TotalAttack)
	assert.Equal(t, 10, result.TargetAC)
	assert.True(t, result.Hit) // Tie goes to attacker in 5e
	assert.Equal(t, 3, result.Damage)
	assert.Equal(t, "bludgeoning", result.DamageType)
	assert.Equal(t, 1, result.TargetNewHP) // 4 - 3 = 1
	assert.False(t, result.TargetDefeated)
}
