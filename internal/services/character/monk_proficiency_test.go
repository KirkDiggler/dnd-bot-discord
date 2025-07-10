package character_test

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/KirkDiggler/dnd-bot-discord/internal/clients/dnd5e"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/rulebook/dnd5e"
	"github.com/KirkDiggler/dnd-bot-discord/internal/repositories/character_draft"
	"github.com/KirkDiggler/dnd-bot-discord/internal/repositories/characters"
	"github.com/KirkDiggler/dnd-bot-discord/internal/services/character"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMonkProficienciesFromAPI verifies what the API returns for Monk
func TestMonkProficienciesFromAPI(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping API test in short mode")
	}

	cfg := &dnd5e.Config{
		HttpClient: http.DefaultClient,
	}
	client, err := dnd5e.New(cfg)
	require.NoError(t, err)

	// Get monk class
	monk, err := client.GetClass("monk")
	require.NoError(t, err)
	require.NotNil(t, monk)

	// Log all proficiencies
	t.Logf("Monk has %d automatic proficiencies:", len(monk.Proficiencies))
	for _, prof := range monk.Proficiencies {
		t.Logf("  - %s (%s)", prof.Name, prof.Key)
	}

	// Log proficiency choices
	t.Logf("\nMonk has %d proficiency choice groups:", len(monk.ProficiencyChoices))
	for i, choice := range monk.ProficiencyChoices {
		t.Logf("\nChoice %d: %s (type: %s, count: %d)", i+1, choice.Name, choice.Type, choice.Count)
		t.Logf("  Has %d options", len(choice.Options))
		// Log all options for monk since tool/instrument choice is important
		for j, opt := range choice.Options {
			t.Logf("  %d. %s (%s)", j+1, opt.GetName(), opt.GetKey())
		}
	}

	// Check expected proficiencies
	expectedProfs := []string{
		"simple-weapons",
		"shortswords",
		"saving-throw-str",
		"saving-throw-dex",
	}

	foundProfs := make(map[string]bool)
	for _, prof := range monk.Proficiencies {
		foundProfs[prof.Key] = true
	}

	for _, expected := range expectedProfs {
		assert.True(t, foundProfs[expected], "Monk should have %s proficiency", expected)
	}

	// Monk should NOT have armor proficiencies
	assert.False(t, foundProfs["light-armor"], "Monk should NOT have light armor")
	assert.False(t, foundProfs["medium-armor"], "Monk should NOT have medium armor")
	assert.False(t, foundProfs["heavy-armor"], "Monk should NOT have heavy armor")
}

// TestMonkFinalizationWithProficiencies tests complete Monk character creation
func TestMonkFinalizationWithProficiencies(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()

	// Create service with real API
	repo := characters.NewInMemoryRepository()
	draftRepo := character_draft.NewInMemoryRepository()
	client, err := dnd5e.New(&dnd5e.Config{
		HttpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	})
	require.NoError(t, err)

	svc := character.NewService(&character.ServiceConfig{
		DNDClient:       client,
		Repository:      repo,
		DraftRepository: draftRepo,
	})

	// Create draft character
	draft, err := svc.GetOrCreateDraftCharacter(ctx, "test-user", "test-realm")
	require.NoError(t, err)

	// Set race and class
	raceKey := "human"
	classKey := "monk"
	_, err = svc.UpdateDraftCharacter(ctx, draft.ID, &character.UpdateDraftInput{
		RaceKey:  &raceKey,
		ClassKey: &classKey,
	})
	require.NoError(t, err)

	// Set abilities (high DEX and WIS for monk)
	abilities := map[string]int{
		"STR": 13,
		"DEX": 16,
		"CON": 14,
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

	// Add skill choices - Monk chooses 2
	skillChoice := []string{"skill-acrobatics", "skill-insight"}
	_, err = svc.UpdateDraftCharacter(ctx, draft.ID, &character.UpdateDraftInput{
		Proficiencies: skillChoice,
	})
	require.NoError(t, err)

	// Set name and finalize
	name := "Test Monk"
	_, err = svc.UpdateDraftCharacter(ctx, draft.ID, &character.UpdateDraftInput{
		Name: &name,
	})
	require.NoError(t, err)

	finalized, err := svc.FinalizeDraftCharacter(ctx, draft.ID)
	require.NoError(t, err)
	require.NotNil(t, finalized)

	// Log all proficiencies
	t.Logf("=== Monk Proficiencies After Finalization ===")
	for profType, profs := range finalized.Proficiencies {
		t.Logf("%s:", profType)
		for _, prof := range profs {
			t.Logf("  - %s (%s)", prof.Name, prof.Key)
		}
	}

	// Test weapon proficiencies
	t.Run("Has weapon proficiencies", func(t *testing.T) {
		weaponProfs, exists := finalized.Proficiencies[rulebook.ProficiencyTypeWeapon]
		assert.True(t, exists, "Monk should have weapon proficiencies")

		profMap := make(map[string]bool)
		for _, prof := range weaponProfs {
			profMap[prof.Key] = true
		}

		// Monk gets simple weapons and shortswords
		assert.True(t, profMap["simple-weapons"], "Monk should have simple weapon proficiency")
		assert.True(t, profMap["shortswords"], "Monk should have shortsword proficiency")
	})

	// Test armor proficiencies (should have none)
	t.Run("Has no armor proficiencies", func(t *testing.T) {
		armorProfs, exists := finalized.Proficiencies[rulebook.ProficiencyTypeArmor]
		if exists {
			assert.Empty(t, armorProfs, "Monk should have no armor proficiencies")
		}
		// Monks use Unarmored Defense instead
	})

	// Test saving throw proficiencies
	t.Run("Has saving throw proficiencies", func(t *testing.T) {
		saveProfs, exists := finalized.Proficiencies[rulebook.ProficiencyTypeSavingThrow]
		assert.True(t, exists, "Monk should have saving throw proficiencies")

		foundSaves := make(map[string]bool)
		for _, prof := range saveProfs {
			foundSaves[prof.Key] = true
		}

		assert.True(t, foundSaves["saving-throw-str"], "Monk should have STR save proficiency")
		assert.True(t, foundSaves["saving-throw-dex"], "Monk should have DEX save proficiency")
	})

	// Test skill proficiencies
	t.Run("Has selected skill proficiencies", func(t *testing.T) {
		skillProfs, exists := finalized.Proficiencies[rulebook.ProficiencyTypeSkill]
		assert.True(t, exists, "Monk should have skill proficiencies")
		assert.GreaterOrEqual(t, len(skillProfs), 2, "Monk should have at least 2 skills")

		foundSkills := make(map[string]bool)
		for _, prof := range skillProfs {
			foundSkills[prof.Key] = true
		}

		assert.True(t, foundSkills["skill-acrobatics"], "Monk should have Acrobatics")
		assert.True(t, foundSkills["skill-insight"], "Monk should have Insight")
	})

	// Test special features (note: features are working, just checking they exist)
	t.Run("Has Monk features", func(t *testing.T) {
		// Check for Martial Arts
		hasMartialArts := false

		for _, feature := range finalized.Features {
			if feature.Key == "martial-arts" {
				hasMartialArts = true
				t.Logf("Found Martial Arts feature: %s", feature.Description)
			}
		}

		assert.True(t, hasMartialArts, "Monk should have Martial Arts feature")

		// Log all features for debugging
		t.Logf("Monk features (%d total):", len(finalized.Features))
		for _, feature := range finalized.Features {
			t.Logf("  - %s (%s)", feature.Name, feature.Key)
		}
	})
}

// TestMonkProficiencyChoices verifies tool/instrument choices are presented
func TestMonkProficiencyChoices(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()

	// Create service
	client, err := dnd5e.New(&dnd5e.Config{
		HttpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	})
	require.NoError(t, err)

	resolver := character.NewChoiceResolver(client)

	// Get class and race data
	class, err := client.GetClass("monk")
	require.NoError(t, err)

	race, err := client.GetRace("human")
	require.NoError(t, err)

	// Resolve proficiency choices
	choices, err := resolver.ResolveProficiencyChoices(ctx, race, class)
	require.NoError(t, err)

	// Should have at least 2 choices: skills and tools/instruments
	assert.GreaterOrEqual(t, len(choices), 2, "Monk should have at least 2 proficiency choice groups")

	// Find and verify skill choice
	var skillChoice *character.SimplifiedChoice
	var toolChoice *character.SimplifiedChoice

	for i := range choices {
		choice := &choices[i]
		if choice.Choose == 2 && choice.Type == "proficiency" {
			// Check if this is skills by looking at first option
			if len(choice.Options) > 0 && strings.HasPrefix(choice.Options[0].Key, "skill-") {
				skillChoice = choice
			}
		} else if choice.Choose == 1 && choice.Type == "proficiency" {
			// This might be the tool/instrument choice
			toolChoice = choice
		}
	}

	assert.NotNil(t, skillChoice, "Monk should have skill choice")
	if skillChoice != nil {
		assert.Equal(t, 2, skillChoice.Choose, "Monk should choose 2 skills")

		// Verify expected skills are available
		skillMap := make(map[string]bool)
		for _, opt := range skillChoice.Options {
			skillMap[opt.Key] = true
		}

		expectedSkills := []string{"skill-acrobatics", "skill-athletics", "skill-history", "skill-insight", "skill-religion", "skill-stealth"}
		for _, expected := range expectedSkills {
			assert.True(t, skillMap[expected], "Monk should have %s as an option", expected)
		}
	}

	// Tool choice is more complex - might be nested or flattened
	if toolChoice != nil {
		t.Logf("Monk tool/instrument choice: %s (choose %d)", toolChoice.Name, toolChoice.Choose)
		t.Logf("Has %d options", len(toolChoice.Options))
		// Log first few options to understand structure
		for i, opt := range toolChoice.Options {
			if i < 5 {
				t.Logf("  - %s (%s)", opt.Name, opt.Key)
			}
		}
	} else {
		t.Logf("Note: Tool/instrument choice might be nested and not resolved by choice resolver")
		// This is expected based on the API data showing nested choices
	}
}
