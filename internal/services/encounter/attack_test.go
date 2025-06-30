package encounter_test

import (
	"context"
	character2 "github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/damage"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/equipment"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/game/combat"
	session2 "github.com/KirkDiggler/dnd-bot-discord/internal/domain/game/session"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/shared"
	"testing"
	"time"

	"github.com/KirkDiggler/dnd-bot-discord/internal/dice"
	mockdice "github.com/KirkDiggler/dnd-bot-discord/internal/dice/mock"
	"github.com/KirkDiggler/dnd-bot-discord/internal/repositories/characters"
	"github.com/KirkDiggler/dnd-bot-discord/internal/repositories/encounters"
	"github.com/KirkDiggler/dnd-bot-discord/internal/repositories/gamesessions"
	"github.com/KirkDiggler/dnd-bot-discord/internal/services/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/services/encounter"
	"github.com/KirkDiggler/dnd-bot-discord/internal/services/session"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPerformAttack_PlayerVsMonster(t *testing.T) {
	t.Skip("Skipping test until dice roller is refactored to use dependency injection - Issue #116")
	ctx := context.Background()
	mockDice := mockdice.NewManualMockRoller()

	// Set up deterministic rolls
	mockDice.SetRolls([]int{
		15, // Attack roll
		8,  // Damage roll (1d8)
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
	sess := &session2.Session{
		ID:        "test-session",
		Name:      "Test Session",
		ChannelID: "channel-1",
		CreatorID: "dm-1",
		DMID:      "dm-1",
		Members: map[string]*session2.SessionMember{
			"player-1": {UserID: "player-1", Role: session2.SessionRolePlayer},
			"dm-1":     {UserID: "dm-1", Role: session2.SessionRoleDM},
		},
		Status:     session2.SessionStatusActive,
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

	// Create a test character with equipment
	testChar := &character2.Character{
		ID:               "char-1",
		Name:             "Fighter",
		OwnerID:          "player-1",
		Status:           shared.CharacterStatusActive,
		Level:            1,
		CurrentHitPoints: 10,
		MaxHitPoints:     10,
		AC:               16,
		Attributes: map[shared.Attribute]*character2.AbilityScore{
			shared.AttributeStrength: {Score: 16}, // +3 modifier
		},
		EquippedSlots: map[character2.Slot]equipment.Equipment{
			character2.SlotMainHand: &equipment.Weapon{
				Base: equipment.BasicEquipment{
					Key:  "longsword",
					Name: "Longsword",
				},
				Damage: &damage.Damage{
					DiceCount:  1,
					DiceSize:   8,
					Bonus:      0,
					DamageType: damage.TypeSlashing,
				},
				WeaponRange: "Melee",
				Properties: []*shared.ReferenceItem{
					{Key: "versatile"},
				},
			},
		},
	}
	err = charRepo.Create(ctx, testChar)
	require.NoError(t, err)

	// Add player
	playerCombatant, err := encounterService.AddPlayer(ctx, enc.ID, "player-1", "char-1")
	require.NoError(t, err)

	// Add monster
	monsterCombatant, err := encounterService.AddMonster(ctx, enc.ID, "dm-1", &encounter.AddMonsterInput{
		Name:  "Goblin",
		MaxHP: 7,
		AC:    15,
	})
	require.NoError(t, err)

	// Start encounter
	enc.Status = combat.EncounterStatusActive
	enc.Turn = 0
	enc.TurnOrder = []string{playerCombatant.ID, monsterCombatant.ID}
	enc.Combatants[playerCombatant.ID].Initiative = 15
	enc.Combatants[monsterCombatant.ID].Initiative = 10

	// Perform attack
	result, err := encounterService.PerformAttack(ctx, &encounter.AttackInput{
		EncounterID: enc.ID,
		AttackerID:  playerCombatant.ID,
		TargetID:    monsterCombatant.ID,
		UserID:      "player-1",
	})

	require.NoError(t, err)
	assert.NotNil(t, result)

	// Verify attack details
	assert.Equal(t, "Fighter", result.AttackerName)
	assert.Equal(t, "Goblin", result.TargetName)
	assert.Equal(t, "Longsword", result.WeaponName)
	assert.Equal(t, 15, result.AttackRoll)
	assert.Equal(t, 3, result.AttackBonus) // STR modifier
	assert.Equal(t, 18, result.TotalAttack)
	assert.Equal(t, 15, result.TargetAC)
	assert.True(t, result.Hit)
	assert.False(t, result.Critical)
	assert.Equal(t, 11, result.Damage) // 8 (roll) + 3 (STR)
	assert.Equal(t, "slashing", result.DamageType)
	assert.False(t, result.TargetDefeated) // 7 HP - 11 damage, but HP can't go below 0

	// Check the log entry
	assert.Contains(t, result.LogEntry, "Fighter attacks Goblin")
	assert.Contains(t, result.LogEntry, "Longsword")
	assert.Contains(t, result.LogEntry, "15 + 3 = **18**")
	assert.Contains(t, result.LogEntry, "HIT")
	assert.Contains(t, result.LogEntry, "11 slashing damage")
	assert.Contains(t, result.LogEntry, "defeated")
}

func TestPerformAttack_MonsterVsPlayer(t *testing.T) {
	t.Skip("Skipping test until dice roller is refactored to use dependency injection - Issue #116")
	ctx := context.Background()
	mockDice := mockdice.NewManualMockRoller()

	// Set up deterministic rolls
	mockDice.SetRolls([]int{
		10, // Attack roll
		4,  // Damage roll (1d6)
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
	sess := &session2.Session{
		ID:        "test-session",
		Name:      "Test Session",
		ChannelID: "channel-1",
		CreatorID: "dm-1",
		DMID:      "dm-1",
		Members: map[string]*session2.SessionMember{
			"player-1": {UserID: "player-1", Role: session2.SessionRolePlayer},
			"dm-1":     {UserID: "dm-1", Role: session2.SessionRoleDM},
		},
		Status:     session2.SessionStatusActive,
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

	// Create a test character
	testChar := &character2.Character{
		ID:               "char-1",
		Name:             "Fighter",
		OwnerID:          "player-1",
		Status:           shared.CharacterStatusActive,
		Level:            1,
		CurrentHitPoints: 10,
		MaxHitPoints:     10,
		AC:               16,
	}
	err = charRepo.Create(ctx, testChar)
	require.NoError(t, err)

	// Add player
	playerCombatant, err := encounterService.AddPlayer(ctx, enc.ID, "player-1", "char-1")
	require.NoError(t, err)

	// Add monster with attack action
	monsterCombatant, err := encounterService.AddMonster(ctx, enc.ID, "dm-1", &encounter.AddMonsterInput{
		Name:  "Goblin",
		MaxHP: 7,
		AC:    15,
		Actions: []*combat.MonsterAction{
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

	// Start encounter with goblin's turn
	enc.Status = combat.EncounterStatusActive
	enc.Turn = 1 // Goblin's turn
	enc.TurnOrder = []string{playerCombatant.ID, monsterCombatant.ID}

	// Perform monster attack
	result, err := encounterService.PerformAttack(ctx, &encounter.AttackInput{
		EncounterID: enc.ID,
		AttackerID:  monsterCombatant.ID,
		TargetID:    playerCombatant.ID,
		UserID:      "dm-1",
		ActionIndex: 0,
	})

	require.NoError(t, err)
	assert.NotNil(t, result)

	// Verify attack details
	assert.Equal(t, "Goblin", result.AttackerName)
	assert.Equal(t, "Fighter", result.TargetName)
	assert.Equal(t, "Scimitar", result.WeaponName)
	assert.Equal(t, 10, result.AttackRoll)
	assert.Equal(t, 4, result.AttackBonus)
	assert.Equal(t, 14, result.TotalAttack)
	assert.Equal(t, 16, result.TargetAC)
	assert.False(t, result.Hit) // 14 < 16 AC

	// Check the log entry
	assert.Contains(t, result.LogEntry, "Goblin attacks Fighter")
	assert.Contains(t, result.LogEntry, "Scimitar")
	assert.Contains(t, result.LogEntry, "10 + 4 = **14**")
	assert.Contains(t, result.LogEntry, "MISS")
}

func TestPerformAttack_CriticalHit(t *testing.T) {
	t.Skip("Skipping test until dice roller is refactored to use dependency injection - Issue #116")
	ctx := context.Background()
	mockDice := mockdice.NewManualMockRoller()

	// Set up deterministic rolls for critical hit
	mockDice.SetRolls([]int{
		20, // Natural 20!
		4,  // First damage die (1d4)
		4,  // Critical damage die (1d4)
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
	sess := &session2.Session{
		ID:        "test-session",
		Name:      "Test Session",
		ChannelID: "channel-1",
		CreatorID: "dm-1",
		DMID:      "dm-1",
		Members: map[string]*session2.SessionMember{
			"dm-1": {UserID: "dm-1", Role: session2.SessionRoleDM},
		},
		Status:     session2.SessionStatusActive,
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

	// Add two unarmed combatants
	attacker, err := encounterService.AddMonster(ctx, enc.ID, "dm-1", &encounter.AddMonsterInput{
		Name:  "Attacker",
		MaxHP: 10,
		AC:    10,
	})
	require.NoError(t, err)

	target, err := encounterService.AddMonster(ctx, enc.ID, "dm-1", &encounter.AddMonsterInput{
		Name:  "Target",
		MaxHP: 20,
		AC:    10,
	})
	require.NoError(t, err)

	// Start encounter
	enc.Status = combat.EncounterStatusActive
	enc.Turn = 0
	enc.TurnOrder = []string{attacker.ID, target.ID}

	// Perform attack (will use unarmed strike)
	result, err := encounterService.PerformAttack(ctx, &encounter.AttackInput{
		EncounterID: enc.ID,
		AttackerID:  attacker.ID,
		TargetID:    target.ID,
		UserID:      "dm-1",
	})

	require.NoError(t, err)
	assert.NotNil(t, result)

	// Verify critical hit
	assert.Equal(t, 20, result.AttackRoll)
	assert.True(t, result.Critical)
	assert.True(t, result.Hit)
	assert.Equal(t, 8, result.Damage) // 4 + 4 (two dice for critical)
	assert.Contains(t, result.LogEntry, "CRITICAL HIT!")
}

func TestPerformAttack_Validations(t *testing.T) {
	ctx := context.Background()

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
		DiceRoller:       dice.NewRandomRoller(),
	})

	// Create a test session
	sess := &session2.Session{
		ID:        "test-session",
		Name:      "Test Session",
		ChannelID: "channel-1",
		CreatorID: "dm-1",
		DMID:      "dm-1",
		Members: map[string]*session2.SessionMember{
			"dm-1": {UserID: "dm-1", Role: session2.SessionRoleDM},
		},
		Status:     session2.SessionStatusActive,
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

	// Add combatants
	attacker, err := encounterService.AddMonster(ctx, enc.ID, "dm-1", &encounter.AddMonsterInput{
		Name:  "Attacker",
		MaxHP: 10,
		AC:    10,
	})
	require.NoError(t, err)

	target, err := encounterService.AddMonster(ctx, enc.ID, "dm-1", &encounter.AddMonsterInput{
		Name:  "Target",
		MaxHP: 10,
		AC:    10,
	})
	require.NoError(t, err)

	// Test: Encounter not active
	_, err = encounterService.PerformAttack(ctx, &encounter.AttackInput{
		EncounterID: enc.ID,
		AttackerID:  attacker.ID,
		TargetID:    target.ID,
		UserID:      "dm-1",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not active")

	// Start encounter
	enc.Status = combat.EncounterStatusActive
	enc.Turn = 0
	enc.TurnOrder = []string{attacker.ID, target.ID}

	// Test: Invalid attacker
	_, err = encounterService.PerformAttack(ctx, &encounter.AttackInput{
		EncounterID: enc.ID,
		AttackerID:  "invalid-id",
		TargetID:    target.ID,
		UserID:      "dm-1",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")

	// Test: Inactive attacker
	enc.Combatants[attacker.ID].IsActive = false
	_, err = encounterService.PerformAttack(ctx, &encounter.AttackInput{
		EncounterID: enc.ID,
		AttackerID:  attacker.ID,
		TargetID:    target.ID,
		UserID:      "dm-1",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not active")
}
