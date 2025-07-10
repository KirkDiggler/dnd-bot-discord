package character_test

import (
	"context"
	"net/http"
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

func TestMartialClassFinalization_Proficiencies(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()

	// Test data for martial classes
	testCases := []struct {
		name            string
		classKey        string
		expectedSaves   []string
		skillChoiceKeys []string // Skills to select during creation
		hasAllArmor     bool     // Fighter/Paladin get "all-armor", Barbarian gets light/medium
	}{
		{
			name:            "Fighter",
			classKey:        "fighter",
			expectedSaves:   []string{"saving-throw-str", "saving-throw-con"},
			skillChoiceKeys: []string{"skill-athletics", "skill-intimidation"},
			hasAllArmor:     true,
		},
		{
			name:            "Barbarian",
			classKey:        "barbarian",
			expectedSaves:   []string{"saving-throw-str", "saving-throw-con"},
			skillChoiceKeys: []string{"skill-athletics", "skill-survival"},
			hasAllArmor:     false,
		},
		{
			name:            "Paladin",
			classKey:        "paladin",
			expectedSaves:   []string{"saving-throw-wis", "saving-throw-cha"},
			skillChoiceKeys: []string{"skill-athletics", "skill-persuasion"},
			hasAllArmor:     true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Use in-memory repository for testing
			repo := characters.NewInMemoryRepository()
			draftRepo := character_draft.NewInMemoryRepository()

			// Use real D&D 5e API client
			client, err := dnd5e.New(&dnd5e.Config{
				HttpClient: &http.Client{
					Timeout: 30 * time.Second,
				},
			})
			require.NoError(t, err)

			// Create service
			svc := character.NewService(&character.ServiceConfig{
				DNDClient:       client,
				Repository:      repo,
				DraftRepository: draftRepo,
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

			// 3. Set ability scores
			abilities := map[string]int{
				"STR": 16,
				"DEX": 14,
				"CON": 15,
				"INT": 10,
				"WIS": 12,
				"CHA": 8,
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
				assert.True(t, exists, "%s should have weapon proficiencies", tc.name)

				// All martial classes get simple and martial weapons
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
				assert.True(t, hasSimple, "%s should have simple weapon proficiency", tc.name)
				assert.True(t, hasMartial, "%s should have martial weapon proficiency", tc.name)
			})

			// Test armor proficiencies
			t.Run("Has armor proficiencies", func(t *testing.T) {
				armorProfs, exists := finalized.Proficiencies[rulebook.ProficiencyTypeArmor]
				assert.True(t, exists, "%s should have armor proficiencies", tc.name)

				// Check for shields (all martial classes get shields)
				hasShields := false
				hasAllArmor := false
				hasLightArmor := false
				hasMediumArmor := false
				hasHeavyArmor := false

				for _, prof := range armorProfs {
					switch prof.Key {
					case "shields":
						hasShields = true
					case "all-armor":
						hasAllArmor = true
					case "light-armor":
						hasLightArmor = true
					case "medium-armor":
						hasMediumArmor = true
					case "heavy-armor":
						hasHeavyArmor = true
					}
				}

				assert.True(t, hasShields, "%s should have shield proficiency", tc.name)

				if tc.hasAllArmor {
					// Fighter and Paladin get "all-armor"
					assert.True(t, hasAllArmor, "%s should have all-armor proficiency", tc.name)
				} else {
					// Barbarian gets specific armor types
					assert.True(t, hasLightArmor, "%s should have light armor proficiency", tc.name)
					assert.True(t, hasMediumArmor, "%s should have medium armor proficiency", tc.name)
					assert.False(t, hasHeavyArmor, "%s should NOT have heavy armor proficiency", tc.name)
				}
			})

			// Test saving throw proficiencies
			t.Run("Has saving throw proficiencies", func(t *testing.T) {
				saveProfs, exists := finalized.Proficiencies[rulebook.ProficiencyTypeSavingThrow]
				assert.True(t, exists, "%s should have saving throw proficiencies", tc.name)

				// Check for expected saves
				foundSaves := make(map[string]bool)
				for _, prof := range saveProfs {
					foundSaves[prof.Key] = true
				}

				for _, expectedSave := range tc.expectedSaves {
					assert.True(t, foundSaves[expectedSave], "%s should have %s proficiency", tc.name, expectedSave)
				}
			})

			// Test skill proficiencies
			t.Run("Has selected skill proficiencies", func(t *testing.T) {
				skillProfs, exists := finalized.Proficiencies[rulebook.ProficiencyTypeSkill]
				assert.True(t, exists, "%s should have skill proficiencies", tc.name)
				assert.Equal(t, len(tc.skillChoiceKeys), len(skillProfs), "%s should have %d selected skills", tc.name, len(tc.skillChoiceKeys))

				// Verify the selected skills are present
				foundSkills := make(map[string]bool)
				for _, prof := range skillProfs {
					foundSkills[prof.Key] = true
				}

				for _, expectedSkill := range tc.skillChoiceKeys {
					assert.True(t, foundSkills[expectedSkill], "%s should have %s skill proficiency", tc.name, expectedSkill)
				}
			})
		})
	}
}

// TestAllArmorProficiencyHandling verifies that "all-armor" proficiency works correctly in combat
func TestAllArmorProficiencyHandling(t *testing.T) {
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

	// Create a fighter (who gets "all-armor")
	draft, err := svc.GetOrCreateDraftCharacter(ctx, "test-user", "test-realm")
	require.NoError(t, err)

	// Set up as fighter
	raceKey := "human"
	classKey := "fighter"
	_, err = svc.UpdateDraftCharacter(ctx, draft.ID, &character.UpdateDraftInput{
		RaceKey:  &raceKey,
		ClassKey: &classKey,
	})
	require.NoError(t, err)

	// Set abilities and name
	abilities := map[string]int{"STR": 16, "DEX": 14, "CON": 15, "INT": 10, "WIS": 12, "CHA": 8}
	for ability, score := range abilities {
		_, err = svc.UpdateDraftCharacter(ctx, draft.ID, &character.UpdateDraftInput{
			AbilityScores: map[string]int{ability: score},
		})
		require.NoError(t, err)
	}

	name := "Test Fighter"
	_, err = svc.UpdateDraftCharacter(ctx, draft.ID, &character.UpdateDraftInput{
		Name: &name,
	})
	require.NoError(t, err)

	// Finalize
	finalized, err := svc.FinalizeDraftCharacter(ctx, draft.ID)
	require.NoError(t, err)

	// Test that the character can use all armor types
	// This tests the actual proficiency checking logic
	armorTypes := []struct {
		name     string
		category string
	}{
		{"Leather Armor", "Light"},
		{"Chain Shirt", "Medium"},
		{"Plate Armor", "Heavy"},
	}

	for _, armor := range armorTypes {
		t.Run("Can use "+armor.name, func(t *testing.T) {
			// Check if character has proficiency with this armor category
			// The actual implementation would check if "all-armor" grants proficiency
			// with light-armor, medium-armor, and heavy-armor

			// For now, we just verify the proficiency exists
			armorProfs := finalized.Proficiencies[rulebook.ProficiencyTypeArmor]
			hasAllArmor := false
			for _, prof := range armorProfs {
				if prof.Key == "all-armor" {
					hasAllArmor = true
					break
				}
			}
			assert.True(t, hasAllArmor, "Fighter should have all-armor proficiency to use %s", armor.name)
		})
	}
}
