package character_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/KirkDiggler/dnd-bot-discord/internal/clients/dnd5e"
	charDomain "github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/shared"
	"github.com/KirkDiggler/dnd-bot-discord/internal/repositories/characters"
	"github.com/KirkDiggler/dnd-bot-discord/internal/services/character"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPaladinCreationFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()

	// Use in-memory repository for testing
	repo := characters.NewInMemoryRepository()

	// Use real D&D 5e API client
	client, err := dnd5e.New(&dnd5e.Config{
		HttpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	})
	require.NoError(t, err)

	// Create service
	svc := character.NewService(&character.ServiceConfig{
		DNDClient:  client,
		Repository: repo,
	})

	t.Run("Paladin creation with proper ability assignment", func(t *testing.T) {
		// 1. Create draft character
		draft, err := svc.GetOrCreateDraftCharacter(ctx, "test-user", "test-realm")
		require.NoError(t, err)

		// 2. Set race and class first (as in normal flow)
		raceKey := "dwarf"
		classKey := "paladin"
		_, err = svc.UpdateDraftCharacter(ctx, draft.ID, &character.UpdateDraftInput{
			RaceKey:  &raceKey,
			ClassKey: &classKey,
		})
		require.NoError(t, err)

		// 3. Set up ability rolls and assignments (simulate normal character creation flow)
		abilityRolls := []charDomain.AbilityRoll{
			{ID: "roll_1", Value: 16}, // For STR
			{ID: "roll_2", Value: 15}, // For CHA
			{ID: "roll_3", Value: 14}, // For CON
			{ID: "roll_4", Value: 12}, // For WIS
			{ID: "roll_5", Value: 10}, // For DEX
			{ID: "roll_6", Value: 10}, // For INT
		}

		assignments := map[string]string{
			"STR": "roll_1", // 16
			"CHA": "roll_2", // 15
			"CON": "roll_3", // 14
			"WIS": "roll_4", // 12
			"DEX": "roll_5", // 10
			"INT": "roll_6", // 10
		}

		// Apply rolls and assignments
		_, err = svc.UpdateDraftCharacter(ctx, draft.ID, &character.UpdateDraftInput{
			AbilityRolls:       abilityRolls,
			AbilityAssignments: assignments,
		})
		require.NoError(t, err)

		// 4. Set name
		name := "Test Paladin"
		_, err = svc.UpdateDraftCharacter(ctx, draft.ID, &character.UpdateDraftInput{
			Name: &name,
		})
		require.NoError(t, err)

		// 5. Finalize the draft
		finalized, err := svc.FinalizeDraftCharacter(ctx, draft.ID)
		require.NoError(t, err)
		require.NotNil(t, finalized)

		// Verify features
		t.Run("Has class features", func(t *testing.T) {
			hasDivineSense := false
			hasLayOnHands := false
			for _, feature := range finalized.Features {
				if feature.Key == "divine_sense" {
					hasDivineSense = true
				}
				if feature.Key == "lay_on_hands" {
					hasLayOnHands = true
				}
			}
			assert.True(t, hasDivineSense, "Paladin should have Divine Sense feature")
			assert.True(t, hasLayOnHands, "Paladin should have Lay on Hands feature")
		})

		// Verify abilities with proper modifiers
		t.Run("Has abilities with correct uses", func(t *testing.T) {
			resources := finalized.GetResources()
			require.NotNil(t, resources)
			require.NotNil(t, resources.Abilities)

			// Check CHA modifier for Divine Sense calculation
			chaScore := finalized.Attributes[shared.AttributeCharisma].Score
			chaModifier := (chaScore - 10) / 2
			t.Logf("CHA Score: %d, Modifier: %d", chaScore, chaModifier)

			// Divine Sense uses = 1 + CHA modifier
			divineSense, exists := resources.Abilities[shared.AbilityKeyDivineSense]
			assert.True(t, exists, "Paladin should have Divine Sense ability")
			if exists {
				expectedUses := 1 + chaModifier
				assert.Equal(t, expectedUses, divineSense.UsesMax,
					"Divine Sense should have %d uses (1 + %d CHA mod)", expectedUses, chaModifier)
				assert.Equal(t, expectedUses, divineSense.UsesRemaining)
			}

			// Lay on Hands pool = 5 × level
			layOnHands, exists := resources.Abilities[shared.AbilityKeyLayOnHands]
			assert.True(t, exists, "Paladin should have Lay on Hands ability")
			if exists {
				assert.Equal(t, 5, layOnHands.UsesMax, "Lay on Hands should have 5 HP pool at level 1")
				assert.Equal(t, 5, layOnHands.UsesRemaining)
			}
		})
	})
}
