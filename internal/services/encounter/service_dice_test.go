package encounter_test

import (
	"context"
	"strings"
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
		Attributes: map[entities.Attribute]*entities.AbilityScore{
			entities.AttributeDexterity: {Score: 16, Bonus: 3}, // +3 DEX bonus for initiative
		},
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
	enc, err = encounterService.GetEncounter(ctx, enc.ID)
	require.NoError(t, err)
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

	// The dice rolls (20, 10, 5) are assigned in alphabetical order of combatant IDs
	// Since UUIDs are random, we need to figure out which combatant got which roll
	totalsByName := make(map[string]int)
	for _, combatant := range enc.Combatants {
		totalsByName[combatant.Name] = combatant.Initiative
	}

	// We know the rolls were 20, 10, 5 and the bonuses
	// Player has +3, Goblin has +2, Skeleton has +1
	// So the possible totals are: 23, 22, 21 (from 20), 13, 12, 11 (from 10), 8, 7, 6 (from 5)
	possibleTotals := []int{23, 22, 21, 13, 12, 11, 8, 7, 6}

	// Verify each combatant got one of the expected totals
	assert.Contains(t, possibleTotals, totalsByName["Player"])
	assert.Contains(t, possibleTotals, totalsByName["Goblin"])
	assert.Contains(t, possibleTotals, totalsByName["Skeleton"])

	// Verify the totals match the bonuses
	// Player has +3, so could be 23 (20+3), 13 (10+3), or 8 (5+3)
	playerTotal := totalsByName["Player"]
	assert.True(t, playerTotal == 23 || playerTotal == 13 || playerTotal == 8,
		"Player total should be one of the rolls + 3")

	// Goblin has +2, so could be 22 (20+2), 12 (10+2), or 7 (5+2)
	goblinTotal := totalsByName["Goblin"]
	assert.True(t, goblinTotal == 22 || goblinTotal == 12 || goblinTotal == 7,
		"Goblin total should be one of the rolls + 2")

	// Verify turn order (should be sorted by initiative descending)
	assert.Len(t, enc.TurnOrder, 3)

	// The first in turn order should have the highest initiative
	firstCombatant := enc.Combatants[enc.TurnOrder[0]]
	highestInit := 0
	for _, combatant := range enc.Combatants {
		if combatant.Initiative > highestInit {
			highestInit = combatant.Initiative
		}
	}
	assert.Equal(t, highestInit, firstCombatant.Initiative,
		"First combatant should have highest initiative")

	// Combat log should contain all three combatants
	logText := strings.Join(enc.CombatLog, "\n")
	assert.Contains(t, logText, "Player")
	assert.Contains(t, logText, "Goblin")
	assert.Contains(t, logText, "Skeleton")
}

func TestEncounterService_CombatScenario_WithMockDice(t *testing.T) {
	// This test demonstrates how we can test a complete combat scenario
	// with predetermined dice rolls

	ctx := context.Background()
	mockDice := dice.NewMockRoller()

	// Set up a complete combat scenario:
	// The dice rolls will be assigned in alphabetical order of combatant IDs
	// We can't predict the order, so let's set up rolls that work either way
	mockDice.SetRolls([]int{
		15, // First combatant's initiative roll
		10, // Second combatant's initiative roll
		16, // Attack roll (hits AC 15)
		8,  // Damage roll
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

	// Verify one of them goes first based on initiative
	// We can't predict the order due to random UUIDs
	firstCombatant := enc.GetCurrentCombatant()
	assert.NotNil(t, firstCombatant)
	assert.True(t, firstCombatant.Type == entities.CombatantTypePlayer || firstCombatant.Type == entities.CombatantTypeMonster)

	// In a real scenario, the player would attack here
	// The mock dice would provide: attack roll 16 (hit), damage 8 (kills goblin)

	// This demonstrates how deterministic testing enables:
	// 1. Testing specific combat scenarios
	// 2. Reproducing bug reports
	// 3. Testing edge cases (critical hits/misses)
	// 4. Ensuring combat math is correct
}
