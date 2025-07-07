package character_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/KirkDiggler/dnd-bot-discord/internal/clients/dnd5e"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/rulebook/dnd5e"
	"github.com/KirkDiggler/dnd-bot-discord/internal/repositories/characters"
	"github.com/KirkDiggler/dnd-bot-discord/internal/services/character"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFullCasterWeaponProficiencyUsage verifies that full casters get and can use their weapon proficiencies
func TestFullCasterWeaponProficiencyUsage(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()

	// Test that each caster gets the right weapon proficiencies
	testCases := []struct {
		name              string
		classKey          string
		shouldHaveWeapons []string // Weapons they should be proficient with
		shouldNotHave     []string // Weapons they should NOT be proficient with
	}{
		{
			name:              "Wizard weapon proficiencies",
			classKey:          "wizard",
			shouldHaveWeapons: []string{"daggers", "darts", "slings", "quarterstaffs", "crossbows-light"},
			shouldNotHave:     []string{"longswords", "maces", "clubs"},
		},
		{
			name:              "Sorcerer weapon proficiencies",
			classKey:          "sorcerer",
			shouldHaveWeapons: []string{"daggers", "darts", "slings", "quarterstaffs", "crossbows-light"},
			shouldNotHave:     []string{"longswords", "rapiers", "shortswords"},
		},
		{
			name:              "Warlock weapon proficiencies",
			classKey:          "warlock",
			shouldHaveWeapons: []string{"simple-weapons"}, // Has category proficiency
			shouldNotHave:     []string{"longswords", "rapiers", "martial-weapons"},
		},
		{
			name:              "Cleric weapon proficiencies",
			classKey:          "cleric",
			shouldHaveWeapons: []string{"simple-weapons"}, // Has category proficiency
			shouldNotHave:     []string{"longswords", "martial-weapons"},
		},
		{
			name:              "Druid weapon proficiencies",
			classKey:          "druid",
			shouldHaveWeapons: []string{"clubs", "daggers", "scimitars", "sickles", "spears"},
			shouldNotHave:     []string{"longswords", "rapiers", "simple-weapons"}, // No category, just specific
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create service with real API
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

			// Create and finalize character
			draft, err := svc.GetOrCreateDraftCharacter(ctx, "test-user", "test-realm")
			require.NoError(t, err)

			// Set race and class
			raceKey := "human"
			_, err = svc.UpdateDraftCharacter(ctx, draft.ID, &character.UpdateDraftInput{
				RaceKey:  &raceKey,
				ClassKey: &tc.classKey,
			})
			require.NoError(t, err)

			// Set abilities
			abilities := map[string]int{
				"STR": 10, "DEX": 14, "CON": 13,
				"INT": 16, "WIS": 15, "CHA": 12,
			}
			for ability, score := range abilities {
				_, err = svc.UpdateDraftCharacter(ctx, draft.ID, &character.UpdateDraftInput{
					AbilityScores: map[string]int{ability: score},
				})
				require.NoError(t, err)
			}

			// Add required skills
			var skillChoice []string
			switch tc.classKey {
			case "cleric", "druid":
				skillChoice = []string{"skill-medicine", "skill-insight"}
			case "warlock", "sorcerer":
				skillChoice = []string{"skill-deception", "skill-intimidation"}
			default:
				skillChoice = []string{"skill-arcana", "skill-investigation"}
			}

			_, err = svc.UpdateDraftCharacter(ctx, draft.ID, &character.UpdateDraftInput{
				Proficiencies: skillChoice,
			})
			require.NoError(t, err)

			// Set name and finalize
			name := "Test " + tc.classKey
			_, err = svc.UpdateDraftCharacter(ctx, draft.ID, &character.UpdateDraftInput{
				Name: &name,
			})
			require.NoError(t, err)

			finalized, err := svc.FinalizeDraftCharacter(ctx, draft.ID)
			require.NoError(t, err)
			require.NotNil(t, finalized)

			// Check weapon proficiencies
			weaponProfs, exists := finalized.Proficiencies[rulebook.ProficiencyTypeWeapon]
			assert.True(t, exists, "%s should have weapon proficiencies", tc.classKey)

			// Build a map of proficiencies for easy checking
			profMap := make(map[string]bool)
			for _, prof := range weaponProfs {
				profMap[prof.Key] = true
			}

			// Verify expected proficiencies
			for _, expected := range tc.shouldHaveWeapons {
				assert.True(t, profMap[expected],
					"%s should have %s proficiency", tc.classKey, expected)
			}

			// Verify they don't have unexpected proficiencies
			for _, notExpected := range tc.shouldNotHave {
				assert.False(t, profMap[notExpected],
					"%s should NOT have %s proficiency", tc.classKey, notExpected)
			}

			// Log all weapon proficiencies for debugging
			t.Logf("%s weapon proficiencies:", tc.classKey)
			for _, prof := range weaponProfs {
				t.Logf("  - %s (%s)", prof.Name, prof.Key)
			}
		})
	}
}

// TestFullCasterSkillChoicesPresented verifies that skill choices are presented during character creation
func TestFullCasterSkillChoicesPresented(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()

	testCases := []struct {
		classKey       string
		expectedCount  int
		expectedInList []string // Some skills that should be in the list
	}{
		{
			classKey:       "wizard",
			expectedCount:  2,
			expectedInList: []string{"skill-arcana", "skill-history", "skill-investigation"},
		},
		{
			classKey:       "sorcerer",
			expectedCount:  2,
			expectedInList: []string{"skill-arcana", "skill-deception", "skill-persuasion"},
		},
		{
			classKey:       "warlock",
			expectedCount:  2,
			expectedInList: []string{"skill-arcana", "skill-deception", "skill-investigation"},
		},
		{
			classKey:       "cleric",
			expectedCount:  2,
			expectedInList: []string{"skill-history", "skill-insight", "skill-medicine"},
		},
		{
			classKey:       "druid",
			expectedCount:  2,
			expectedInList: []string{"skill-animal-handling", "skill-nature", "skill-perception"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.classKey+" skill choices", func(t *testing.T) {
			// Create service
			client, err := dnd5e.New(&dnd5e.Config{
				HttpClient: &http.Client{
					Timeout: 30 * time.Second,
				},
			})
			require.NoError(t, err)

			resolver := character.NewChoiceResolver(client)

			// Get class data
			class, err := client.GetClass(tc.classKey)
			require.NoError(t, err)

			// Get race (human has no extra choices)
			race, err := client.GetRace("human")
			require.NoError(t, err)

			// Resolve proficiency choices
			choices, err := resolver.ResolveProficiencyChoices(ctx, race, class)
			require.NoError(t, err)

			// Find skill choice
			var skillChoice *character.SimplifiedChoice
			for _, choice := range choices {
				if choice.Type == "proficiency" && choice.Choose == tc.expectedCount {
					// Check if this is likely a skill choice by looking at options
					hasSkills := false
					for _, opt := range choice.Options {
						if opt.Key != "" && len(opt.Key) > 6 && opt.Key[:6] == "skill-" {
							hasSkills = true
							break
						}
					}
					if hasSkills {
						skillChoice = &choice
						break
					}
				}
			}

			assert.NotNil(t, skillChoice, "%s should have skill proficiency choice", tc.classKey)
			if skillChoice != nil {
				assert.Equal(t, tc.expectedCount, skillChoice.Choose,
					"%s should choose %d skills", tc.classKey, tc.expectedCount)

				// Build option map
				optionMap := make(map[string]bool)
				for _, opt := range skillChoice.Options {
					optionMap[opt.Key] = true
				}

				// Verify expected skills are in the list
				for _, expected := range tc.expectedInList {
					assert.True(t, optionMap[expected],
						"%s should have %s in skill choices", tc.classKey, expected)
				}

				t.Logf("%s can choose %d skills from %d options",
					tc.classKey, skillChoice.Choose, len(skillChoice.Options))
			}
		})
	}
}
