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

func TestSkillExpertClassFinalization_Proficiencies(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()

	// Test data for skill-expert classes
	testCases := []struct {
		name            string
		classKey        string
		expectedSaves   []string
		skillCount      int
		skillChoiceKeys []string
		hasThievesTools bool
		hasInstruments  bool
		instrumentCount int
		specificWeapons []string // Specific weapon proficiencies beyond simple
	}{
		{
			name:            "Rogue",
			classKey:        "rogue",
			expectedSaves:   []string{"saving-throw-dex", "saving-throw-int"},
			skillCount:      4,
			skillChoiceKeys: []string{"skill-stealth", "skill-investigation", "skill-acrobatics", "skill-deception"},
			hasThievesTools: true,
			hasInstruments:  false,
			specificWeapons: []string{"longswords", "rapiers", "shortswords", "hand-crossbows"},
		},
		{
			name:            "Bard",
			classKey:        "bard",
			expectedSaves:   []string{"saving-throw-dex", "saving-throw-cha"},
			skillCount:      3,
			skillChoiceKeys: []string{"skill-persuasion", "skill-performance", "skill-deception"},
			hasThievesTools: false,
			hasInstruments:  true,
			instrumentCount: 3,
			specificWeapons: []string{"longswords", "rapiers", "shortswords", "hand-crossbows"},
		},
		{
			name:            "Ranger",
			classKey:        "ranger",
			expectedSaves:   []string{"saving-throw-str", "saving-throw-dex"},
			skillCount:      3,
			skillChoiceKeys: []string{"skill-survival", "skill-perception", "skill-animal-handling"},
			hasThievesTools: false,
			hasInstruments:  false,
			specificWeapons: []string{}, // Rangers get martial weapons, no specific ones
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
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

			// 1. Create draft character
			draft, err := svc.GetOrCreateDraftCharacter(ctx, "test-user", "test-realm")
			require.NoError(t, err)

			// 2. Set race and class
			raceKey := "half-elf" // Half-elves get extra skills, good for testing
			updated, err := svc.UpdateDraftCharacter(ctx, draft.ID, &character.UpdateDraftInput{
				RaceKey:  &raceKey,
				ClassKey: &tc.classKey,
			})
			require.NoError(t, err)
			require.NotNil(t, updated)

			// 3. Set ability scores
			abilities := map[string]int{
				"STR": 12,
				"DEX": 16,
				"CON": 14,
				"INT": 13,
				"WIS": 10,
				"CHA": 15,
			}
			for ability, score := range abilities {
				_, err = svc.UpdateDraftCharacter(ctx, draft.ID, &character.UpdateDraftInput{
					AbilityScores: map[string]int{ability: score},
				})
				require.NoError(t, err)
			}

			// 4. Add selected skill proficiencies
			_, err = svc.UpdateDraftCharacter(ctx, draft.ID, &character.UpdateDraftInput{
				Proficiencies: tc.skillChoiceKeys,
			})
			require.NoError(t, err)

			// 5. For Bard, we might need to select instruments (if the API supports it)
			// This might need special handling depending on how instrument choices work

			// 6. Set name
			name := "Test " + tc.name
			_, err = svc.UpdateDraftCharacter(ctx, draft.ID, &character.UpdateDraftInput{
				Name: &name,
			})
			require.NoError(t, err)

			// 7. Finalize the draft
			finalized, err := svc.FinalizeDraftCharacter(ctx, draft.ID)
			require.NoError(t, err)
			require.NotNil(t, finalized)

			// Debug: Print all proficiencies
			t.Logf("=== %s Proficiencies After Finalization ===", tc.name)
			for profType, profs := range finalized.Proficiencies {
				t.Logf("%s:", profType)
				for _, prof := range profs {
					t.Logf("  - %s (%s)", prof.Name, prof.Key)
				}
			}

			// Test weapon proficiencies
			t.Run("Has weapon proficiencies", func(t *testing.T) {
				weaponProfs, exists := finalized.Proficiencies[rulebook.ProficiencyTypeWeapon]
				assert.True(t, exists, "%s should have weapon proficiencies", tc.name)

				// All skill-expert classes get simple weapons
				hasSimple := false
				foundSpecific := make(map[string]bool)

				for _, prof := range weaponProfs {
					if prof.Key == "simple-weapons" {
						hasSimple = true
					}
					// Check for specific weapons
					for _, specific := range tc.specificWeapons {
						if prof.Key == specific {
							foundSpecific[specific] = true
						}
					}
				}

				assert.True(t, hasSimple, "%s should have simple weapon proficiency", tc.name)

				// Check specific weapons
				for _, weapon := range tc.specificWeapons {
					assert.True(t, foundSpecific[weapon], "%s should have %s proficiency", tc.name, weapon)
				}

				// Ranger should have martial weapons
				if tc.classKey == "ranger" {
					hasMartial := false
					for _, prof := range weaponProfs {
						if prof.Key == "martial-weapons" {
							hasMartial = true
							break
						}
					}
					assert.True(t, hasMartial, "Ranger should have martial weapon proficiency")
				}
			})

			// Test armor proficiencies
			t.Run("Has armor proficiencies", func(t *testing.T) {
				armorProfs, exists := finalized.Proficiencies[rulebook.ProficiencyTypeArmor]
				assert.True(t, exists, "%s should have armor proficiencies", tc.name)

				// All skill-expert classes get light armor
				hasLightArmor := false
				for _, prof := range armorProfs {
					if prof.Key == "light-armor" {
						hasLightArmor = true
						break
					}
				}
				assert.True(t, hasLightArmor, "%s should have light armor proficiency", tc.name)

				// Ranger gets more armor
				if tc.classKey == "ranger" {
					hasMedium := false
					hasShields := false
					for _, prof := range armorProfs {
						if prof.Key == "medium-armor" {
							hasMedium = true
						}
						if prof.Key == "shields" {
							hasShields = true
						}
					}
					assert.True(t, hasMedium, "Ranger should have medium armor proficiency")
					assert.True(t, hasShields, "Ranger should have shield proficiency")
				}
			})

			// Test saving throw proficiencies
			t.Run("Has saving throw proficiencies", func(t *testing.T) {
				saveProfs, exists := finalized.Proficiencies[rulebook.ProficiencyTypeSavingThrow]
				assert.True(t, exists, "%s should have saving throw proficiencies", tc.name)

				foundSaves := make(map[string]bool)
				for _, prof := range saveProfs {
					foundSaves[prof.Key] = true
				}

				for _, expectedSave := range tc.expectedSaves {
					assert.True(t, foundSaves[expectedSave], "%s should have %s proficiency", tc.name, expectedSave)
				}
			})

			// Test tool proficiencies
			t.Run("Has correct tool proficiencies", func(t *testing.T) {
				// Check both Tool and Unknown (Other) proficiency types since thieves' tools might be categorized differently
				toolProfs, toolExists := finalized.Proficiencies[rulebook.ProficiencyTypeTool]
				otherProfs, otherExists := finalized.Proficiencies[rulebook.ProficiencyTypeUnknown]

				if tc.hasThievesTools {
					hasThievesTools := false

					// Check in Tool proficiencies
					if toolExists {
						for _, prof := range toolProfs {
							if prof.Key == "thieves-tools" {
								hasThievesTools = true
								break
							}
						}
					}

					// If not found in Tool, check in Unknown (Other) proficiencies
					if !hasThievesTools && otherExists {
						for _, prof := range otherProfs {
							if prof.Key == "thieves-tools" {
								hasThievesTools = true
								break
							}
						}
					}

					assert.True(t, hasThievesTools, "Rogue should have thieves' tools proficiency")
				}
			})

			// Test skill proficiencies
			t.Run("Has selected skill proficiencies", func(t *testing.T) {
				skillProfs, exists := finalized.Proficiencies[rulebook.ProficiencyTypeSkill]
				assert.True(t, exists, "%s should have skill proficiencies", tc.name)

				// Should have at least the class skills (might have more from race)
				assert.GreaterOrEqual(t, len(skillProfs), tc.skillCount,
					"%s should have at least %d skills", tc.name, tc.skillCount)

				// Verify the selected skills are present
				foundSkills := make(map[string]bool)
				for _, prof := range skillProfs {
					foundSkills[prof.Key] = true
				}

				for _, expectedSkill := range tc.skillChoiceKeys {
					assert.True(t, foundSkills[expectedSkill],
						"%s should have %s skill proficiency", tc.name, expectedSkill)
				}
			})

			// Test for expertise feature (Rogue only at level 1)
			if tc.classKey == "rogue" {
				t.Run("Has expertise feature", func(t *testing.T) {
					hasExpertise := false
					for _, feature := range finalized.Features {
						if feature.Key == "expertise" {
							hasExpertise = true
							t.Logf("Found expertise feature: %s", feature.Description)
							break
						}
					}
					assert.True(t, hasExpertise, "Rogue should have expertise feature at level 1")
				})
			}
		})
	}
}

// TestBardInstrumentProficiencies specifically tests Bard's musical instrument choices
func TestBardInstrumentProficiencies(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	t.Log("Bard should be able to choose 3 musical instruments")
	t.Log("This test will help determine how instrument proficiencies are handled")

	// TODO: Once we understand how instrument choices work, implement a specific test
	// The API shows Bard has a choice for "Three musical instruments of your choice"
	// Need to check if these show up as Tool proficiencies or a different type
}
