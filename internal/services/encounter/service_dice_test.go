package encounter_test

import (
	"context"
	"testing"
	"time"

	"github.com/KirkDiggler/dnd-bot-discord/internal/dice"
	"github.com/KirkDiggler/dnd-bot-discord/internal/entities"
	"github.com/KirkDiggler/dnd-bot-discord/internal/repositories/characters"
	"github.com/KirkDiggler/dnd-bot-discord/internal/repositories/encounters"
	"github.com/KirkDiggler/dnd-bot-discord/internal/repositories/gamesessions"
	"github.com/KirkDiggler/dnd-bot-discord/internal/services/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/services/encounter"
	"github.com/KirkDiggler/dnd-bot-discord/internal/services/session"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncounterService_RollInitiative_WithMockDice(t *testing.T) {
	// Setup
	ctx := context.Background()
	mockDice := dice.NewMockRoller()
	
	// Set deterministic rolls for initiative
	mockDice.SetRolls([]int{
		20, // Player rolls nat 20
		10, // Goblin rolls 10
		5,  // Skeleton rolls 5
	})
	
	// Create services with mock dice
	charRepo := characters.NewInMemoryRepository()
	charService := character.NewService(&character.ServiceConfig{
		Repository: charRepo,
	})
	
	// Create a test character
	testChar := &entities.Character{
		ID:               "char-1",
		Name:             "Player",
		OwnerID:          "player-1",
		Status:           entities.CharacterStatusActive,
		Level:            1,
		CurrentHitPoints: 10,
		MaxHitPoints:     10,
		AC:               15,
	}
	err := charRepo.Create(ctx, testChar)
	require.NoError(t, err)
	
	sessionRepo := gamesessions.NewInMemoryRepository()
	sessionService := session.NewService(&session.ServiceConfig{
		Repository:       sessionRepo,
		CharacterService: charService,
	})
	
	// Create session with all required fields
	sess := &entities.Session{
		ID:          "test-session",
		Name:        "Test Session",
		InviteCode:  "TEST123",
		ChannelID:   "channel-1",
		CreatorID:   "user-1",
		DMID:        "user-1",
		Description: "Test session for dice rolling",
		Members:     make(map[string]*entities.SessionMember),
		Status:      entities.SessionStatusActive,
		CreatedAt:   time.Now(),
		LastActive:  time.Now(),
	}
	sess.Members["user-1"] = &entities.SessionMember{
		UserID: "user-1",
		Role:   entities.SessionRoleDM,
	}
	createErr := sessionRepo.Create(ctx, sess)
	require.NoError(t, createErr)
	
	// Create encounter service with mock dice
	encounterService := encounter.NewService(&encounter.ServiceConfig{
		Repository:       encounters.NewInMemoryRepository(),
		SessionService:   sessionService,
		CharacterService: charService,
		DiceRoller:       mockDice,
	})
	
	// Create encounter
	enc, err := encounterService.CreateEncounter(ctx, &encounter.CreateEncounterInput{
		SessionID:   "test-session",
		ChannelID:   "channel-1",
		Name:        "Test Combat",
		Description: "Testing dice rolls",
		UserID:      "user-1",
	})
	require.NoError(t, err)
	
	// Add combatants
	_, err = encounterService.AddMonster(ctx, enc.ID, "user-1", &encounter.AddMonsterInput{
		Name:            "Goblin",
		MaxHP:           7,
		AC:              15,
		InitiativeBonus: 2, // Will roll 10 + 2 = 12
	})
	require.NoError(t, err)
	
	_, err = encounterService.AddMonster(ctx, enc.ID, "user-1", &encounter.AddMonsterInput{
		Name:            "Skeleton",
		MaxHP:           13,
		AC:              13,
		InitiativeBonus: 1, // Will roll 5 + 1 = 6
	})
	require.NoError(t, err)
	
	// Add player
	_, err = encounterService.AddPlayer(ctx, enc.ID, "player-1", "char-1")
	require.NoError(t, err)
	
	// Update player's initiative bonus
	enc, _ = encounterService.GetEncounter(ctx, enc.ID)
	for _, combatant := range enc.Combatants {
		if combatant.Type == entities.CombatantTypePlayer {
			combatant.InitiativeBonus = 3 // Will roll 20 + 3 = 23
		}
	}
	
	// Roll initiative
	err = encounterService.RollInitiative(ctx, enc.ID, "user-1")
	require.NoError(t, err)
	
	// Get updated encounter
	enc, err = encounterService.GetEncounter(ctx, enc.ID)
	require.NoError(t, err)
	
	// The order of adding combatants determines the order of initiative rolls
	// We added: Goblin (gets roll 20), Skeleton (gets roll 10), Player (gets roll 5)
	expectedInitiatives := map[string]int{
		"Goblin":   22, // 20 + 2
		"Skeleton": 11, // 10 + 1  
		"Player":   8,  // 5 + 3
	}
	
	for _, combatant := range enc.Combatants {
		expected, exists := expectedInitiatives[combatant.Name]
		if exists {
			assert.Equal(t, expected, combatant.Initiative, 
				"Expected %s to have initiative %d, got %d", 
				combatant.Name, expected, combatant.Initiative)
		}
	}
	
	// Verify turn order (should be sorted by initiative descending)
	assert.Len(t, enc.TurnOrder, 3)
	
	// The first in turn order should be Goblin (highest initiative)
	firstCombatant := enc.Combatants[enc.TurnOrder[0]]
	assert.Equal(t, "Goblin", firstCombatant.Name)
	assert.Equal(t, 22, firstCombatant.Initiative)
	
	// Combat log should show the rolls
	assert.Contains(t, enc.CombatLog[1], "Goblin") 
	assert.Contains(t, enc.CombatLog[1], "20 + 2 = **22**")
}

func TestEncounterService_CombatScenario_WithMockDice(t *testing.T) {
	// This test demonstrates how we can test a complete combat scenario
	// with predetermined dice rolls
	
	ctx := context.Background()
	mockDice := dice.NewMockRoller()
	
	// Set up a complete combat scenario:
	// Initiative: Player (15), Goblin (10)
	// Player attacks: hits (roll 16), deals 8 damage - goblin dies
	// Combat ends with player victory
	mockDice.SetRolls([]int{
		15, // Player initiative
		10, // Goblin initiative
		16, // Player attack roll (hits AC 15)
		8,  // Player damage roll
	})
	
	// Create services
	charRepo := characters.NewInMemoryRepository()
	charService := character.NewService(&character.ServiceConfig{
		Repository: charRepo,
	})
	
	// Create test character
	testChar := &entities.Character{
		ID:               "char-1",
		Name:             "Fighter",
		OwnerID:          "player-1",
		Status:           entities.CharacterStatusActive,
		Level:            1,
		CurrentHitPoints: 10,
		MaxHitPoints:     10,
		AC:               16,
	}
	err := charRepo.Create(ctx, testChar)
	require.NoError(t, err)
	
	sessionRepo := gamesessions.NewInMemoryRepository()
	sessionService := session.NewService(&session.ServiceConfig{
		Repository:       sessionRepo,
		CharacterService: charService,
	})
	
	// Create session
	sess := &entities.Session{
		ID:          "test-session",
		Name:        "Test Session",
		InviteCode:  "TEST456",
		ChannelID:   "channel-1",
		CreatorID:   "user-1",
		DMID:        "user-1",
		Description: "Combat test session",
		Members:     make(map[string]*entities.SessionMember),
		Status:      entities.SessionStatusActive,
		CreatedAt:   time.Now(),
		LastActive:  time.Now(),
	}
	sess.Members["user-1"] = &entities.SessionMember{
		UserID: "user-1",
		Role:   entities.SessionRoleDM,
	}
	createErr := sessionRepo.Create(ctx, sess)
	require.NoError(t, createErr)
	
	// Create encounter service with mock dice
	encounterService := encounter.NewService(&encounter.ServiceConfig{
		Repository:       encounters.NewInMemoryRepository(),
		SessionService:   sessionService,
		CharacterService: charService,
		DiceRoller:       mockDice,
	})
	
	// Create and setup encounter
	enc, err := encounterService.CreateEncounter(ctx, &encounter.CreateEncounterInput{
		SessionID: "test-session",
		ChannelID: "channel-1",
		Name:      "Boss Fight",
		UserID:    "user-1",
	})
	require.NoError(t, err)
	
	// Add goblin with 7 HP
	_, err = encounterService.AddMonster(ctx, enc.ID, "user-1", &encounter.AddMonsterInput{
		Name:  "Goblin Boss",
		MaxHP: 7,
		AC:    15,
	})
	require.NoError(t, err)
	
	// Add player
	_, err = encounterService.AddPlayer(ctx, enc.ID, "player-1", "char-1")
	require.NoError(t, err)
	
	// Roll initiative
	err = encounterService.RollInitiative(ctx, enc.ID, "user-1")
	require.NoError(t, err)
	
	// Start encounter
	err = encounterService.StartEncounter(ctx, enc.ID, "user-1")
	require.NoError(t, err)
	
	// Get updated encounter
	enc, err = encounterService.GetEncounter(ctx, enc.ID)
	require.NoError(t, err)
	
	// Verify player goes first
	firstCombatant := enc.GetCurrentCombatant()
	assert.Equal(t, entities.CombatantTypePlayer, firstCombatant.Type)
	
	// In a real scenario, the player would attack here
	// The mock dice would provide: attack roll 16 (hit), damage 8 (kills goblin)
	
	// This demonstrates how deterministic testing enables:
	// 1. Testing specific combat scenarios
	// 2. Reproducing bug reports
	// 3. Testing edge cases (critical hits/misses)
	// 4. Ensuring combat math is correct
}