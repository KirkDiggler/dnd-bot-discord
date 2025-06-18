package dungeon_test

import (
	"context"
	"testing"

	"github.com/KirkDiggler/dnd-bot-discord/internal/entities"
	"github.com/KirkDiggler/dnd-bot-discord/internal/repositories/dungeons"
	"github.com/KirkDiggler/dnd-bot-discord/internal/services/dungeon"
	mockencounter "github.com/KirkDiggler/dnd-bot-discord/internal/services/encounter/mock"
	mockloot "github.com/KirkDiggler/dnd-bot-discord/internal/services/loot/mock"
	mockmonster "github.com/KirkDiggler/dnd-bot-discord/internal/services/monster/mock"
	mocksession "github.com/KirkDiggler/dnd-bot-discord/internal/services/session/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestDungeonService_CreateDungeon_WithMonsterService(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mocks
	repo := dungeons.NewInMemoryRepository()
	sessionService := mocksession.NewMockService(ctrl)
	encounterService := mockencounter.NewMockService(ctrl)
	monsterService := mockmonster.NewMockService(ctrl)

	// Set up expectations
	sessionService.EXPECT().GetSession(gomock.Any(), "session-123").Return(&entities.Session{
		ID: "session-123",
	}, nil)
	sessionService.EXPECT().SaveSession(gomock.Any(), gomock.Any()).Return(nil)

	// Expect monster service to be called for room generation
	monsterService.EXPECT().GetRandomMonsters(gomock.Any(), "medium", gomock.Any()).Return([]*entities.MonsterTemplate{
		{Key: "goblin", Name: "Goblin"},
		{Key: "orc", Name: "Orc"},
	}, nil).AnyTimes()

	// Create service with monster service
	service := dungeon.NewService(&dungeon.ServiceConfig{
		Repository:       repo,
		SessionService:   sessionService,
		EncounterService: encounterService,
		MonsterService:   monsterService,
	})

	// Create dungeon
	result, err := service.CreateDungeon(context.Background(), &dungeon.CreateDungeonInput{
		SessionID:  "session-123",
		Difficulty: "medium",
		CreatorID:  "user-123",
	})

	// Assert
	require.NoError(t, err)
	assert.NotEmpty(t, result.ID)
	assert.Equal(t, "session-123", result.SessionID)
	assert.Equal(t, entities.DungeonStateAwaitingParty, result.State)
	assert.Equal(t, 1, result.RoomNumber)
	assert.NotNil(t, result.CurrentRoom)
	assert.Equal(t, entities.RoomTypeCombat, result.CurrentRoom.Type)

	// Should have dynamic monsters from the service
	assert.Contains(t, result.CurrentRoom.Monsters, "goblin")
	assert.Contains(t, result.CurrentRoom.Monsters, "orc")
}

func TestDungeonService_CreateDungeon_WithLootService(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mocks
	repo := dungeons.NewInMemoryRepository()
	sessionService := mocksession.NewMockService(ctrl)
	encounterService := mockencounter.NewMockService(ctrl)
	lootService := mockloot.NewMockService(ctrl)

	// Set up expectations
	sessionService.EXPECT().GetSession(gomock.Any(), "session-123").Return(&entities.Session{
		ID: "session-123",
	}, nil)
	sessionService.EXPECT().SaveSession(gomock.Any(), gomock.Any()).Return(nil)

	// Create service with loot service
	service := dungeon.NewService(&dungeon.ServiceConfig{
		Repository:       repo,
		SessionService:   sessionService,
		EncounterService: encounterService,
		LootService:      lootService,
	})

	// Create dungeon
	result, err := service.CreateDungeon(context.Background(), &dungeon.CreateDungeonInput{
		SessionID:  "session-123",
		Difficulty: "hard",
		CreatorID:  "user-123",
	})

	// Assert
	require.NoError(t, err)
	assert.NotEmpty(t, result.ID)

	// First room should be combat
	assert.Equal(t, entities.RoomTypeCombat, result.CurrentRoom.Type)

	// TODO: Add test for treasure room generation when we have proper room progression
	// This would require simulating combat completion and room clearing
}

func TestDungeonService_CreateDungeon_FallbackWithoutServices(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mocks
	repo := dungeons.NewInMemoryRepository()
	sessionService := mocksession.NewMockService(ctrl)
	encounterService := mockencounter.NewMockService(ctrl)

	// Set up expectations
	sessionService.EXPECT().GetSession(gomock.Any(), "session-123").Return(&entities.Session{
		ID: "session-123",
	}, nil)
	sessionService.EXPECT().SaveSession(gomock.Any(), gomock.Any()).Return(nil)

	// Create service WITHOUT monster and loot services
	service := dungeon.NewService(&dungeon.ServiceConfig{
		Repository:       repo,
		SessionService:   sessionService,
		EncounterService: encounterService,
		// No MonsterService or LootService - should fallback to hardcoded
	})

	// Create dungeon
	result, err := service.CreateDungeon(context.Background(), &dungeon.CreateDungeonInput{
		SessionID:  "session-123",
		Difficulty: "easy",
		CreatorID:  "user-123",
	})

	// Assert
	require.NoError(t, err)
	assert.NotEmpty(t, result.ID)
	assert.NotNil(t, result.CurrentRoom)

	// Should have hardcoded monsters
	possibleMonsters := []string{"goblin", "skeleton", "orc", "dire-wolf"}
	for _, monster := range result.CurrentRoom.Monsters {
		assert.Contains(t, possibleMonsters, monster)
	}
}
