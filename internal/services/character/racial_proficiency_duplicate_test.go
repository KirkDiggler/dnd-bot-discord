package character_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/KirkDiggler/dnd-bot-discord/internal/clients/dnd5e"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/rulebook/dnd5e"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/shared"
	"github.com/KirkDiggler/dnd-bot-discord/internal/repositories/characters"
	"github.com/KirkDiggler/dnd-bot-discord/internal/services/character"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDuplicateProficiencyPrevention tests that the service prevents duplicate proficiency selection
func TestDuplicateProficiencyPrevention(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()

	testCases := []struct {
		name              string
		raceKey           string
		classKey          string
		racialProficiency string
		shouldBeInChoices bool // Should the racial proficiency appear in class choices?
	}{
		{
			name:              "Half-Orc Barbarian - Intimidation",
			raceKey:           "half-orc",
			classKey:          "barbarian",
			racialProficiency: "skill-intimidation",
			shouldBeInChoices: false, // Should be filtered out
		},
		{
			name:              "Elf Ranger - Perception",
			raceKey:           "elf",
			classKey:          "ranger",
			racialProficiency: "skill-perception",
			shouldBeInChoices: false, // Should be filtered out
		},
		{
			name:              "Human Fighter - No overlap",
			raceKey:           "human",
			classKey:          "fighter",
			racialProficiency: "",   // Humans don't get skill proficiencies
			shouldBeInChoices: true, // All skills should be available
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create service
			repo := characters.NewInMemoryRepository()
			client, err := dnd5e.New(&dnd5e.Config{
				HttpClient: &http.Client{
					Timeout: 30 * time.Second,
				},
			})
			require.NoError(t, err)

			svc := character.NewService(&character.ServiceConfig{
				DNDClient:  client,
				Repository: repo,
			})

			// Create draft and set race/class
			draft, err := svc.GetOrCreateDraftCharacter(ctx, "test-user", "test-realm")
			require.NoError(t, err)

			updated, err := svc.UpdateDraftCharacter(ctx, draft.ID, &character.UpdateDraftInput{
				RaceKey:  &tc.raceKey,
				ClassKey: &tc.classKey,
			})
			require.NoError(t, err)

			// Check if the racial proficiency exists
			if tc.racialProficiency != "" {
				// Log racial proficiencies from the race
				t.Logf("%s racial proficiencies:", tc.raceKey)
				if updated.Race != nil {
					for _, prof := range updated.Race.StartingProficiencies {
						t.Logf("  - %s (%s)", prof.Name, prof.Key)
					}
				}
			}

			// Get available proficiency choices through the resolver
			resolver := character.NewChoiceResolver(client)

			race, err := client.GetRace(tc.raceKey)
			require.NoError(t, err)

			class, err := client.GetClass(tc.classKey)
			require.NoError(t, err)

			choices, err := resolver.ResolveProficiencyChoices(ctx, race, class)
			require.NoError(t, err)

			// Find skill choices
			var skillChoice *character.SimplifiedChoice
			for i := range choices {
				if choices[i].Type == "proficiency" {
					// Check if it's skills
					hasSkills := false
					for _, opt := range choices[i].Options {
						if len(opt.Key) > 6 && opt.Key[:6] == "skill-" {
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

			if skillChoice != nil && tc.racialProficiency != "" {
				// Check if racial proficiency is in the choices
				found := false
				for _, opt := range skillChoice.Options {
					if opt.Key == tc.racialProficiency {
						found = true
						break
					}
				}

				if tc.shouldBeInChoices {
					assert.True(t, found, "%s should be in skill choices", tc.racialProficiency)
				} else {
					// Duplicate prevention is now implemented
					assert.False(t, found, "%s should be filtered from choices (already granted by race)", tc.racialProficiency)
					t.Logf("✓ %s correctly filtered from choices", tc.racialProficiency)
				}
			}
		})
	}
}

// TestCharacterFinalizationDeduplication verifies that finalization removes duplicate proficiencies
func TestCharacterFinalizationDeduplication(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()

	// Create service
	repo := characters.NewInMemoryRepository()
	client, err := dnd5e.New(&dnd5e.Config{
		HttpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	})
	require.NoError(t, err)

	svc := character.NewService(&character.ServiceConfig{
		DNDClient:  client,
		Repository: repo,
	})

	// Create Half-Orc Barbarian
	draft, err := svc.GetOrCreateDraftCharacter(ctx, "test-user", "test-realm")
	require.NoError(t, err)

	raceKey := "half-orc"
	classKey := "barbarian"
	_, err = svc.UpdateDraftCharacter(ctx, draft.ID, &character.UpdateDraftInput{
		RaceKey:  &raceKey,
		ClassKey: &classKey,
	})
	require.NoError(t, err)

	// Set abilities
	abilities := map[string]int{
		"STR": 17, "DEX": 14, "CON": 16,
		"INT": 8, "WIS": 12, "CHA": 10,
	}
	for ability, score := range abilities {
		_, err = svc.UpdateDraftCharacter(ctx, draft.ID, &character.UpdateDraftInput{
			AbilityScores: map[string]int{ability: score},
		})
		require.NoError(t, err)
	}

	// Intentionally select Intimidation (which Half-Orc already has)
	_, err = svc.UpdateDraftCharacter(ctx, draft.ID, &character.UpdateDraftInput{
		Proficiencies: []string{"skill-intimidation", "skill-survival"},
	})
	require.NoError(t, err)

	// Set name and finalize
	name := "Grokk the Intimidating"
	_, err = svc.UpdateDraftCharacter(ctx, draft.ID, &character.UpdateDraftInput{
		Name: &name,
	})
	require.NoError(t, err)

	finalized, err := svc.FinalizeDraftCharacter(ctx, draft.ID)
	require.NoError(t, err)

	// Count proficiencies
	skillProfs := finalized.Proficiencies[rulebook.ProficiencyTypeSkill]

	// Map to track unique proficiencies
	uniqueProfs := make(map[string]int)
	for _, prof := range skillProfs {
		uniqueProfs[prof.Key]++
	}

	// Check that each proficiency appears only once
	for key, count := range uniqueProfs {
		assert.Equal(t, 1, count, "Proficiency %s should appear exactly once, but appears %d times", key, count)
	}

	// Verify we have the expected proficiencies
	assert.Equal(t, 1, uniqueProfs["skill-intimidation"], "Should have Intimidation exactly once")
	assert.Equal(t, 1, uniqueProfs["skill-survival"], "Should have Survival exactly once")

	t.Logf("Final unique skill proficiencies: %d", len(uniqueProfs))
	for key := range uniqueProfs {
		t.Logf("  - %s", key)
	}
}

// TestRacialAbilityScoreImprovements verifies that racial ability bonuses are applied
func TestRacialAbilityScoreImprovements(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()

	testCases := []struct {
		raceKey         string
		expectedBonuses map[string]int
	}{
		{
			raceKey: "half-orc",
			expectedBonuses: map[string]int{
				"STR": 2,
				"CON": 1,
			},
		},
		{
			raceKey: "elf",
			expectedBonuses: map[string]int{
				"DEX": 2,
			},
		},
		{
			raceKey: "human",
			expectedBonuses: map[string]int{
				"STR": 1,
				"DEX": 1,
				"CON": 1,
				"INT": 1,
				"WIS": 1,
				"CHA": 1,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.raceKey+" ability bonuses", func(t *testing.T) {
			// Create service
			repo := characters.NewInMemoryRepository()
			client, err := dnd5e.New(&dnd5e.Config{
				HttpClient: &http.Client{
					Timeout: 30 * time.Second,
				},
			})
			require.NoError(t, err)

			svc := character.NewService(&character.ServiceConfig{
				DNDClient:  client,
				Repository: repo,
			})

			// Create character with base 10 in all stats
			draft, err := svc.GetOrCreateDraftCharacter(ctx, "test-user", "test-realm")
			require.NoError(t, err)

			// Set race and a simple class
			classKey := "fighter"
			_, err = svc.UpdateDraftCharacter(ctx, draft.ID, &character.UpdateDraftInput{
				RaceKey:  &tc.raceKey,
				ClassKey: &classKey,
			})
			require.NoError(t, err)

			// Set all abilities to 10
			baseAbilities := map[string]int{
				"STR": 10, "DEX": 10, "CON": 10,
				"INT": 10, "WIS": 10, "CHA": 10,
			}
			for ability, score := range baseAbilities {
				_, err = svc.UpdateDraftCharacter(ctx, draft.ID, &character.UpdateDraftInput{
					AbilityScores: map[string]int{ability: score},
				})
				require.NoError(t, err)
			}

			// Add required proficiencies and name
			_, err = svc.UpdateDraftCharacter(ctx, draft.ID, &character.UpdateDraftInput{
				Proficiencies: []string{"skill-athletics", "skill-intimidation"},
			})
			require.NoError(t, err)

			name := "Test " + tc.raceKey
			_, err = svc.UpdateDraftCharacter(ctx, draft.ID, &character.UpdateDraftInput{
				Name: &name,
			})
			require.NoError(t, err)

			// Finalize
			finalized, err := svc.FinalizeDraftCharacter(ctx, draft.ID)
			require.NoError(t, err)

			// Check ability scores include racial bonuses
			t.Logf("%s final ability scores:", tc.raceKey)

			// Check each ability score
			for ability, baseScore := range baseAbilities {
				var finalScore int
				switch ability {
				case "STR":
					if finalized.Attributes[shared.AttributeStrength] != nil {
						finalScore = finalized.Attributes[shared.AttributeStrength].Score
					}
				case "DEX":
					if finalized.Attributes[shared.AttributeDexterity] != nil {
						finalScore = finalized.Attributes[shared.AttributeDexterity].Score
					}
				case "CON":
					if finalized.Attributes[shared.AttributeConstitution] != nil {
						finalScore = finalized.Attributes[shared.AttributeConstitution].Score
					}
				case "INT":
					if finalized.Attributes[shared.AttributeIntelligence] != nil {
						finalScore = finalized.Attributes[shared.AttributeIntelligence].Score
					}
				case "WIS":
					if finalized.Attributes[shared.AttributeWisdom] != nil {
						finalScore = finalized.Attributes[shared.AttributeWisdom].Score
					}
				case "CHA":
					if finalized.Attributes[shared.AttributeCharisma] != nil {
						finalScore = finalized.Attributes[shared.AttributeCharisma].Score
					}
				}

				expectedBonus := tc.expectedBonuses[ability]
				expectedTotal := baseScore + expectedBonus

				t.Logf("  %s: %d (base %d + racial %d)", ability, finalScore, baseScore, expectedBonus)
				assert.Equal(t, expectedTotal, finalScore, "%s %s should be %d", tc.raceKey, ability, expectedTotal)
			}
		})
	}
}
