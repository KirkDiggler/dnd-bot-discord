//go:build integration
// +build integration

package character_test

import (
	"context"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/KirkDiggler/dnd-bot-discord/internal/clients/dnd5e"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	charactersRepo "github.com/KirkDiggler/dnd-bot-discord/internal/repositories/characters"
	charService "github.com/KirkDiggler/dnd-bot-discord/internal/services/character"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRacePreviewIntegration(t *testing.T) {
	// Set up Redis client
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "redis://localhost:6379/15" // Use test database
	}

	opts, err := redis.ParseURL(redisURL)
	require.NoError(t, err)

	client := redis.NewClient(opts)
	defer func() {
		if closeErr := client.Close(); closeErr != nil {
			t.Logf("Error closing Redis client: %v", closeErr)
		}
	}()

	// Verify Redis is available
	ctx := context.Background()
	_, err = client.Ping(ctx).Result()
	if err != nil {
		t.Skipf("Redis not available: %v", err)
	}

	// Clean up test data
	defer func() { _ = client.FlushDB(ctx) }()

	// Create repository
	repo := charactersRepo.NewRedisRepository(&charactersRepo.RedisRepoConfig{
		Client: client,
	})

	// Create real D&D API client
	dndClient, err := dnd5e.New(&dnd5e.Config{
		HttpClient: &http.Client{},
	})
	require.NoError(t, err)

	// Create service
	svc := charService.NewService(&charService.ServiceConfig{
		DNDClient:  dndClient,
		Repository: repo,
	})

	// Create flow builder and flow service
	flowBuilder := charService.NewFlowBuilder(dndClient)
	flowService := charService.NewCreationFlowService(svc, flowBuilder)

	t.Run("race selection step should have populated options", func(t *testing.T) {
		// Create a draft character
		char, err := svc.GetOrCreateDraftCharacter(ctx, "test_user", "test_realm")
		require.NoError(t, err)

		// Build the flow to get race options
		flow, err := flowBuilder.BuildFlow(ctx, char)
		require.NoError(t, err)
		require.NotNil(t, flow)
		require.NotEmpty(t, flow.Steps)

		// First step should be race selection
		raceStep := flow.Steps[0]
		assert.Equal(t, character.StepTypeRaceSelection, raceStep.Type)
		assert.NotEmpty(t, raceStep.Options)

		// Check specific races for proper data
		t.Run("elf should have ability bonuses", func(t *testing.T) {
			var elfOption *character.CreationOption
			for _, opt := range raceStep.Options {
				if opt.Key == "elf" {
					elfOption = &opt
					break
				}
			}
			require.NotNil(t, elfOption, "Elf option not found")

			// Check description isn't "No special traits"
			assert.NotEqual(t, "No special traits", elfOption.Description)
			assert.Contains(t, strings.ToUpper(elfOption.Description), "DEX", "Elf should show DEX bonus")

			// Check metadata
			assert.NotNil(t, elfOption.Metadata)
			assert.NotNil(t, elfOption.Metadata["race"])
			assert.NotNil(t, elfOption.Metadata["bonuses"])
		})

		t.Run("dwarf should have ability bonuses", func(t *testing.T) {
			var dwarfOption *character.CreationOption
			for _, opt := range raceStep.Options {
				if opt.Key == "dwarf" {
					dwarfOption = &opt
					break
				}
			}
			require.NotNil(t, dwarfOption, "Dwarf option not found")

			assert.NotEqual(t, "No special traits", dwarfOption.Description)
			assert.Contains(t, strings.ToUpper(dwarfOption.Description), "CON", "Dwarf should show CON bonus")
		})
	})

	t.Run("preview step result should populate race correctly", func(t *testing.T) {
		// Create a draft character
		char, err := svc.GetOrCreateDraftCharacter(ctx, "test_user2", "test_realm")
		require.NoError(t, err)

		// Create a race selection result
		result := &character.CreationStepResult{
			StepType:   character.StepTypeRaceSelection,
			Selections: []string{"elf"},
		}

		// Preview the result
		previewChar, err := flowService.PreviewStepResult(ctx, char.ID, result)
		require.NoError(t, err)
		require.NotNil(t, previewChar)
		require.NotNil(t, previewChar.Race)

		assert.Equal(t, "elf", previewChar.Race.Key)
		assert.Equal(t, "Elf", previewChar.Race.Name)
		assert.NotEmpty(t, previewChar.Race.AbilityBonuses)

		// Check for DEX bonus
		foundDexBonus := false
		for _, bonus := range previewChar.Race.AbilityBonuses {
			t.Logf("Elf ability bonus: Attribute=%s, Bonus=%d", bonus.Attribute, bonus.Bonus)
			// Check various formats
			attrUpper := strings.ToUpper(string(bonus.Attribute))
			if attrUpper == "DEXTERITY" || attrUpper == "DEX" {
				foundDexBonus = true
				assert.Greater(t, bonus.Bonus, 0)
			}
		}
		assert.True(t, foundDexBonus, "Elf should have DEX bonus")
	})
}
