package character_test

import (
	"context"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/rulebook"
	"net/http"
	"testing"
	"time"

	"github.com/KirkDiggler/dnd-bot-discord/internal/clients/dnd5e"
	"github.com/KirkDiggler/dnd-bot-discord/internal/repositories/characters"
	"github.com/KirkDiggler/dnd-bot-discord/internal/services/character"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRangerDraftFinalization_MissingProficiencies(t *testing.T) {
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

	// Simulate the draft creation flow
	// 1. Create draft character
	draft, err := svc.GetOrCreateDraftCharacter(ctx, "test-user", "test-realm")
	require.NoError(t, err)
	require.NotNil(t, draft)

	// 2. Update with race and class (simulating UI selections)
	raceKey := "elf"
	classKey := "ranger"
	updated, err := svc.UpdateDraftCharacter(ctx, draft.ID, &character.UpdateDraftInput{
		RaceKey:  &raceKey,
		ClassKey: &classKey,
	})
	require.NoError(t, err)
	require.NotNil(t, updated)

	// 3. Set ability scores
	abilities := map[string]int{
		"STR": 14,
		"DEX": 16,
		"CON": 13,
		"INT": 10,
		"WIS": 15,
		"CHA": 8,
	}
	for ability, score := range abilities {
		_, err = svc.UpdateDraftCharacter(ctx, draft.ID, &character.UpdateDraftInput{
			AbilityScores: map[string]int{ability: score},
		})
		require.NoError(t, err)
	}

	// 4. Add selected proficiencies (skills)
	skillProfs := []string{"skill-perception", "skill-survival", "skill-stealth"}
	_, err = svc.UpdateDraftCharacter(ctx, draft.ID, &character.UpdateDraftInput{
		Proficiencies: skillProfs,
	})
	require.NoError(t, err)

	// 5. Set name
	name := "Test Elf Ranger"
	_, err = svc.UpdateDraftCharacter(ctx, draft.ID, &character.UpdateDraftInput{
		Name: &name,
	})
	require.NoError(t, err)

	// 6. Finalize the draft
	finalized, err := svc.FinalizeDraftCharacter(ctx, draft.ID)
	require.NoError(t, err)
	require.NotNil(t, finalized)

	// Debug: Print all proficiencies
	t.Logf("=== Character Proficiencies After Finalization ===")
	for profType, profs := range finalized.Proficiencies {
		t.Logf("%s:", profType)
		for _, prof := range profs {
			t.Logf("  - %s (%s)", prof.Name, prof.Key)
		}
	}

	// Test: Verify class proficiencies were added
	t.Run("Has weapon proficiencies", func(t *testing.T) {
		weaponProfs, exists := finalized.Proficiencies[rulebook.ProficiencyTypeWeapon]
		assert.True(t, exists, "Should have weapon proficiencies")
		assert.NotEmpty(t, weaponProfs, "Should have at least one weapon proficiency")

		// Check for specific proficiencies Rangers should have
		hasSimple := false
		hasMartial := false
		for _, prof := range weaponProfs {
			if prof.Key == "simple-weapons" {
				hasSimple = true
			}
			if prof.Key == "martial-weapons" {
				hasMartial = true
			}
		}
		assert.True(t, hasSimple, "Ranger should have simple weapon proficiency")
		assert.True(t, hasMartial, "Ranger should have martial weapon proficiency")
	})

	t.Run("Has armor proficiencies", func(t *testing.T) {
		armorProfs, exists := finalized.Proficiencies[rulebook.ProficiencyTypeArmor]
		assert.True(t, exists, "Should have armor proficiencies")
		assert.GreaterOrEqual(t, len(armorProfs), 3, "Should have light, medium armor and shields")
	})

	t.Run("Has saving throw proficiencies", func(t *testing.T) {
		saveProfs, exists := finalized.Proficiencies[rulebook.ProficiencyTypeSavingThrow]
		assert.True(t, exists, "Should have saving throw proficiencies")
		assert.GreaterOrEqual(t, len(saveProfs), 2, "Rangers have STR and DEX saves")
	})

	t.Run("Still has selected skill proficiencies", func(t *testing.T) {
		skillProfs, exists := finalized.Proficiencies[rulebook.ProficiencyTypeSkill]
		assert.True(t, exists, "Should have skill proficiencies")
		assert.Equal(t, 3, len(skillProfs), "Should have the 3 selected skills")
	})
}
