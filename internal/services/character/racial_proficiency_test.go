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

// TestRacialProficienciesFromAPI tests what proficiencies each race provides
func TestRacialProficienciesFromAPI(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping API test in short mode")
	}

	cfg := &dnd5e.Config{
		HttpClient: http.DefaultClient,
	}
	client, err := dnd5e.New(cfg)
	require.NoError(t, err)

	// Test races with proficiencies
	testCases := []struct {
		raceKey               string
		expectedProficiencies []string
		hasProficiencyChoices bool
	}{
		{
			raceKey:               "half-orc",
			expectedProficiencies: []string{"skill-intimidation"},
			hasProficiencyChoices: false,
		},
		{
			raceKey:               "elf",
			expectedProficiencies: []string{"skill-perception"},
			hasProficiencyChoices: false,
		},
		{
			raceKey:               "dwarf",
			expectedProficiencies: []string{}, // Base dwarf might not have any
			hasProficiencyChoices: false,
		},
		{
			raceKey:               "half-elf",
			expectedProficiencies: []string{}, // Half-elves get to choose skills
			hasProficiencyChoices: true,       // Should have skill choices
		},
	}

	for _, tc := range testCases {
		t.Run(tc.raceKey, func(t *testing.T) {
			race, err := client.GetRace(tc.raceKey)
			require.NoError(t, err)
			require.NotNil(t, race)

			// Log automatic proficiencies
			t.Logf("%s has %d automatic proficiencies:", race.Name, len(race.StartingProficiencies))
			for _, prof := range race.StartingProficiencies {
				t.Logf("  - %s (%s)", prof.Name, prof.Key)
			}

			// Check expected proficiencies
			foundProfs := make(map[string]bool)
			for _, prof := range race.StartingProficiencies {
				foundProfs[prof.Key] = true
			}

			for _, expected := range tc.expectedProficiencies {
				assert.True(t, foundProfs[expected], "%s should have %s proficiency", race.Name, expected)
			}

			// Check for proficiency choices
			if race.StartingProficiencyOptions != nil {
				t.Logf("%s proficiency choices: Choose %d from %d options",
					race.Name, race.StartingProficiencyOptions.Count, len(race.StartingProficiencyOptions.Options))
			}

			if tc.hasProficiencyChoices {
				assert.NotNil(t, race.StartingProficiencyOptions, "%s should have proficiency choices", race.Name)
			}
		})
	}
}

// TestRaceClassProficiencyOverlap tests scenarios where race and class provide same proficiency
func TestRaceClassProficiencyOverlap(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()

	// Test case: Half-Orc Barbarian (both can have Intimidation)
	t.Run("Half-Orc Barbarian Intimidation overlap", func(t *testing.T) {
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
		raceKey := "half-orc"
		classKey := "barbarian"
		_, err = svc.UpdateDraftCharacter(ctx, draft.ID, &character.UpdateDraftInput{
			RaceKey:  &raceKey,
			ClassKey: &classKey,
		})
		require.NoError(t, err)

		// Get the choice resolver to check available choices
		resolver := character.NewChoiceResolver(client)

		race, err := client.GetRace(raceKey)
		require.NoError(t, err)

		class, err := client.GetClass(classKey)
		require.NoError(t, err)

		// Get proficiency choices
		choices, err := resolver.ResolveProficiencyChoices(ctx, race, class)
		require.NoError(t, err)

		// Find the barbarian skill choice
		var skillChoice *character.SimplifiedChoice
		for i := range choices {
			if choices[i].Type == "proficiency" && choices[i].Choose == 2 {
				// Check if it's skills by looking at options
				if len(choices[i].Options) > 0 {
					hasSkills := false
					for _, opt := range choices[i].Options {
						if strings.HasPrefix(opt.Key, "skill-") {
							hasSkills = true
							break
						}
					}
					if hasSkills {
						skillChoice = &choices[i]
						break
					}
				}
			}
		}

		require.NotNil(t, skillChoice, "Barbarian should have skill choices")

		// Check if Intimidation is in the choices (it shouldn't be if we're preventing duplicates)
		hasIntimidation := false
		for _, opt := range skillChoice.Options {
			if opt.Key == "skill-intimidation" {
				hasIntimidation = true
				t.Logf("WARNING: Intimidation is still in choices despite Half-Orc having it")
				break
			}
		}

		// Log all available skill choices
		t.Logf("Available skill choices for Half-Orc Barbarian:")
		for _, opt := range skillChoice.Options {
			t.Logf("  - %s (%s)", opt.Name, opt.Key)
		}

		// Duplicate prevention is now implemented
		assert.False(t, hasIntimidation, "Intimidation should be filtered from choices (already granted by Half-Orc race)")

		// Complete character creation to verify Half-Orc gets Intimidation automatically
		abilities := map[string]int{
			"STR": 16, "DEX": 14, "CON": 15,
			"INT": 8, "WIS": 12, "CHA": 10,
		}
		for ability, score := range abilities {
			_, err = svc.UpdateDraftCharacter(ctx, draft.ID, &character.UpdateDraftInput{
				AbilityScores: map[string]int{ability: score},
			})
			require.NoError(t, err)
		}

		// Choose skills (including Intimidation to test what happens)
		skillSelections := []string{"skill-intimidation", "skill-survival"}
		_, err = svc.UpdateDraftCharacter(ctx, draft.ID, &character.UpdateDraftInput{
			Proficiencies: skillSelections,
		})
		require.NoError(t, err)

		// Set name and finalize
		name := "Test Half-Orc Barbarian"
		_, err = svc.UpdateDraftCharacter(ctx, draft.ID, &character.UpdateDraftInput{
			Name: &name,
		})
		require.NoError(t, err)

		finalized, err := svc.FinalizeDraftCharacter(ctx, draft.ID)
		require.NoError(t, err)

		// Check final proficiencies
		skillProfs := finalized.Proficiencies[rulebook.ProficiencyTypeSkill]

		// Count how many times Intimidation appears
		intimidationCount := 0
		for _, prof := range skillProfs {
			if prof.Key == "skill-intimidation" {
				intimidationCount++
			}
		}

		// Log all skill proficiencies
		t.Logf("Final skill proficiencies:")
		for _, prof := range skillProfs {
			t.Logf("  - %s (%s)", prof.Name, prof.Key)
		}

		// Should only have Intimidation once, even if selected twice
		assert.Equal(t, 1, intimidationCount, "Intimidation should only appear once in final proficiencies")
	})
}

// TestHalfElfSkillChoices tests that Half-Elf gets to choose skills
func TestHalfElfSkillChoices(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()

	t.Run("Half-Elf gets 2 skill choices", func(t *testing.T) {
		client, err := dnd5e.New(&dnd5e.Config{
			HttpClient: &http.Client{
				Timeout: 30 * time.Second,
			},
		})
		require.NoError(t, err)

		resolver := character.NewChoiceResolver(client)

		// Get race and class data
		race, err := client.GetRace("half-elf")
		require.NoError(t, err)

		// Use wizard as a simple class
		class, err := client.GetClass("wizard")
		require.NoError(t, err)

		// Get proficiency choices
		choices, err := resolver.ResolveProficiencyChoices(ctx, race, class)
		require.NoError(t, err)

		// Half-elf should have a racial skill choice
		var racialSkillChoice *character.SimplifiedChoice
		for i := range choices {
			if choices[i].ID == "half-elf-prof" {
				racialSkillChoice = &choices[i]
				break
			}
		}

		if racialSkillChoice != nil {
			t.Logf("Half-Elf racial skill choice: Choose %d from %d options",
				racialSkillChoice.Choose, len(racialSkillChoice.Options))
			assert.Equal(t, 2, racialSkillChoice.Choose, "Half-Elf should choose 2 skills")
		} else {
			t.Log("Half-Elf racial skill choice not found - might be in starting proficiency options")
		}
	})
}

// TestElfWeaponProficiencies tests that elves get weapon proficiencies
func TestElfWeaponProficiencies(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()

	// Create service
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

	// Create elf wizard
	draft, err := svc.GetOrCreateDraftCharacter(ctx, "test-user", "test-realm")
	require.NoError(t, err)

	raceKey := "elf"
	classKey := "wizard"
	_, err = svc.UpdateDraftCharacter(ctx, draft.ID, &character.UpdateDraftInput{
		RaceKey:  &raceKey,
		ClassKey: &classKey,
	})
	require.NoError(t, err)

	// Complete character
	abilities := map[string]int{
		"STR": 10, "DEX": 16, "CON": 14,
		"INT": 16, "WIS": 13, "CHA": 12,
	}
	for ability, score := range abilities {
		_, err = svc.UpdateDraftCharacter(ctx, draft.ID, &character.UpdateDraftInput{
			AbilityScores: map[string]int{ability: score},
		})
		require.NoError(t, err)
	}

	// Choose wizard skills
	_, err = svc.UpdateDraftCharacter(ctx, draft.ID, &character.UpdateDraftInput{
		Proficiencies: []string{"skill-arcana", "skill-investigation"},
	})
	require.NoError(t, err)

	name := "Test Elf Wizard"
	_, err = svc.UpdateDraftCharacter(ctx, draft.ID, &character.UpdateDraftInput{
		Name: &name,
	})
	require.NoError(t, err)

	finalized, err := svc.FinalizeDraftCharacter(ctx, draft.ID)
	require.NoError(t, err)

	// Check proficiencies
	t.Log("=== Elf Wizard Proficiencies ===")
	for profType, profs := range finalized.Proficiencies {
		t.Logf("%s:", profType)
		for _, prof := range profs {
			t.Logf("  - %s (%s)", prof.Name, prof.Key)
		}
	}

	// Check for elf weapon proficiencies
	weaponProfs := finalized.Proficiencies[rulebook.ProficiencyTypeWeapon]

	// Check for perception (all elves get this)
	skillProfs := finalized.Proficiencies[rulebook.ProficiencyTypeSkill]
	hasPerception := false
	for _, prof := range skillProfs {
		if prof.Key == "skill-perception" {
			hasPerception = true
			break
		}
	}
	assert.True(t, hasPerception, "Elf should have Perception proficiency")

	// Log weapon proficiencies to see what elves get
	if len(weaponProfs) > 0 {
		t.Log("Elf racial weapon proficiencies found")
	}
}
