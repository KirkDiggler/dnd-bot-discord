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

func TestFullCasterClassFinalization_Proficiencies(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()

	// Test data for full caster classes
	testCases := []struct {
		name             string
		classKey         string
		expectedSaves    []string
		skillCount       int
		skillChoiceKeys  []string
		hasLightArmor    bool
		hasMediumArmor   bool
		hasShields       bool
		hasSimpleWeapons bool
		specificWeapons  []string // For classes that get specific weapons instead of simple
		hasHerbalism     bool     // For Druid
	}{
		{
			name:             "Wizard",
			classKey:         "wizard",
			expectedSaves:    []string{"saving-throw-int", "saving-throw-wis"},
			skillCount:       2,
			skillChoiceKeys:  []string{"skill-arcana", "skill-investigation"},
			hasLightArmor:    false,
			hasMediumArmor:   false,
			hasShields:       false,
			hasSimpleWeapons: false,
			specificWeapons:  []string{"daggers", "darts", "slings", "quarterstaffs", "crossbows-light"},
		},
		{
			name:             "Sorcerer",
			classKey:         "sorcerer",
			expectedSaves:    []string{"saving-throw-con", "saving-throw-cha"},
			skillCount:       2,
			skillChoiceKeys:  []string{"skill-deception", "skill-persuasion"},
			hasLightArmor:    false,
			hasMediumArmor:   false,
			hasShields:       false,
			hasSimpleWeapons: false,
			specificWeapons:  []string{"daggers", "darts", "slings", "quarterstaffs", "crossbows-light"},
		},
		{
			name:             "Warlock",
			classKey:         "warlock",
			expectedSaves:    []string{"saving-throw-wis", "saving-throw-cha"},
			skillCount:       2,
			skillChoiceKeys:  []string{"skill-investigation", "skill-deception"},
			hasLightArmor:    true,
			hasMediumArmor:   false,
			hasShields:       false,
			hasSimpleWeapons: true,
			specificWeapons:  []string{}, // Gets simple weapons category
		},
		{
			name:             "Cleric",
			classKey:         "cleric",
			expectedSaves:    []string{"saving-throw-wis", "saving-throw-cha"},
			skillCount:       2,
			skillChoiceKeys:  []string{"skill-insight", "skill-medicine"},
			hasLightArmor:    true,
			hasMediumArmor:   true,
			hasShields:       true,
			hasSimpleWeapons: true,
			specificWeapons:  []string{}, // Gets simple weapons category
		},
		{
			name:            "Druid",
			classKey:        "druid",
			expectedSaves:   []string{"saving-throw-int", "saving-throw-wis"},
			skillCount:      2,
			skillChoiceKeys: []string{"skill-nature", "skill-perception"},
			hasLightArmor:   true,
			hasMediumArmor:  true,
			hasShields:      true,
			hasHerbalism:    true,
			// Druid gets specific weapons, not simple category
			hasSimpleWeapons: false,
			specificWeapons: []string{
				"clubs", "daggers", "darts", "javelins", "maces",
				"quarterstaffs", "scimitars", "sickles", "slings", "spears",
			},
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
			raceKey := "human"
			updated, err := svc.UpdateDraftCharacter(ctx, draft.ID, &character.UpdateDraftInput{
				RaceKey:  &raceKey,
				ClassKey: &tc.classKey,
			})
			require.NoError(t, err)
			require.NotNil(t, updated)

			// 3. Set ability scores (appropriate for casters)
			abilities := map[string]int{
				"STR": 8,
				"DEX": 14,
				"CON": 13,
				"INT": 16, // High for wizards
				"WIS": 15, // High for clerics/druids
				"CHA": 12, // High for sorcerers/warlocks
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

			// 5. Set name
			name := "Test " + tc.name
			_, err = svc.UpdateDraftCharacter(ctx, draft.ID, &character.UpdateDraftInput{
				Name: &name,
			})
			require.NoError(t, err)

			// 6. Finalize the draft
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

				if tc.hasSimpleWeapons {
					assert.True(t, exists, "%s should have weapon proficiencies", tc.name)
					hasSimple := false
					for _, prof := range weaponProfs {
						if prof.Key == "simple-weapons" {
							hasSimple = true
							break
						}
					}
					assert.True(t, hasSimple, "%s should have simple weapon proficiency", tc.name)
				} else if len(tc.specificWeapons) > 0 {
					assert.True(t, exists, "%s should have weapon proficiencies", tc.name)

					// Check for specific weapons
					foundWeapons := make(map[string]bool)
					for _, prof := range weaponProfs {
						foundWeapons[prof.Key] = true
					}

					for _, weapon := range tc.specificWeapons {
						assert.True(t, foundWeapons[weapon], "%s should have %s proficiency", tc.name, weapon)
					}
				}
			})

			// Test armor proficiencies
			t.Run("Has armor proficiencies", func(t *testing.T) {
				armorProfs, exists := finalized.Proficiencies[rulebook.ProficiencyTypeArmor]

				if tc.hasLightArmor || tc.hasMediumArmor || tc.hasShields {
					assert.True(t, exists, "%s should have armor proficiencies", tc.name)

					foundArmor := make(map[string]bool)
					for _, prof := range armorProfs {
						foundArmor[prof.Key] = true
					}

					if tc.hasLightArmor {
						assert.True(t, foundArmor["light-armor"], "%s should have light armor proficiency", tc.name)
					}
					if tc.hasMediumArmor {
						assert.True(t, foundArmor["medium-armor"], "%s should have medium armor proficiency", tc.name)
					}
					if tc.hasShields {
						assert.True(t, foundArmor["shields"], "%s should have shield proficiency", tc.name)
					}
				} else if exists {
					// Wizard and Sorcerer have no armor proficiencies
					assert.Empty(t, armorProfs, "%s should have no armor proficiencies", tc.name)
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

			// Test tool proficiencies (Druid herbalism kit)
			if tc.hasHerbalism {
				t.Run("Has herbalism kit proficiency", func(t *testing.T) {
					// Check both Tool and Unknown (Other) proficiency types
					toolProfs, toolExists := finalized.Proficiencies[rulebook.ProficiencyTypeTool]
					otherProfs, otherExists := finalized.Proficiencies[rulebook.ProficiencyTypeUnknown]

					hasHerbalism := false

					// Check in Tool proficiencies
					if toolExists {
						for _, prof := range toolProfs {
							if prof.Key == "herbalism-kit" {
								hasHerbalism = true
								break
							}
						}
					}

					// If not found in Tool, check in Unknown (Other) proficiencies
					if !hasHerbalism && otherExists {
						for _, prof := range otherProfs {
							if prof.Key == "herbalism-kit" {
								hasHerbalism = true
								break
							}
						}
					}

					assert.True(t, hasHerbalism, "Druid should have herbalism kit proficiency")
				})
			}

			// Test skill proficiencies
			t.Run("Has selected skill proficiencies", func(t *testing.T) {
				skillProfs, exists := finalized.Proficiencies[rulebook.ProficiencyTypeSkill]
				assert.True(t, exists, "%s should have skill proficiencies", tc.name)

				// Should have at least the class skills
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
		})
	}
}

// TestDruidNonMetalRestriction verifies that Druids have the non-metal armor restriction
func TestDruidNonMetalRestriction(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	t.Log("Druids will not wear armor or use shields made of metal")
	t.Log("This is a roleplay restriction that might need special handling")

	// TODO: Implement check for metal armor/shields when equipment system supports material types
	// For now, this is documented as a known limitation
}
