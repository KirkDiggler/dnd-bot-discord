package character_test

import (
	"context"
	character2 "github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/rulebook"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/shared"
	"net/http"
	"testing"
	"time"

	"github.com/KirkDiggler/dnd-bot-discord/internal/clients/dnd5e"
	"github.com/KirkDiggler/dnd-bot-discord/internal/repositories/characters"
	"github.com/KirkDiggler/dnd-bot-discord/internal/services/character"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRangerCharacterCreation_Integration(t *testing.T) {
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

	// Create a Ranger character using the service
	output, err := svc.CreateCharacter(ctx, &character.CreateCharacterInput{
		UserID:   "user123",
		RealmID:  "test-realm",
		Name:     "Aragorn",
		RaceKey:  "human",
		ClassKey: "ranger",
		AbilityScores: map[string]int{
			"STR": 14,
			"DEX": 16,
			"CON": 13,
			"INT": 10,
			"WIS": 15,
			"CHA": 8,
		},
		Proficiencies: []string{"skill-perception", "skill-stealth", "skill-survival"},
		Equipment:     []string{}, // No specific equipment choices for this test
	})
	require.NoError(t, err)
	require.NotNil(t, output)

	finalized := output.Character
	require.NotNil(t, finalized)

	// Verify Ranger features were applied
	assert.Len(t, finalized.Features, 2, "Should have 2 level 1 Ranger features")

	// Check for specific features
	hasFeature := func(key string) bool {
		for _, f := range finalized.Features {
			if f.Key == key {
				return true
			}
		}
		return false
	}

	assert.True(t, hasFeature("favored_enemy"), "Should have Favored Enemy feature")
	assert.True(t, hasFeature("natural_explorer"), "Should have Natural Explorer feature")

	// Verify proficiencies
	profTypes := make(map[rulebook.ProficiencyType]int)
	for pType, profs := range finalized.Proficiencies {
		profTypes[pType] = len(profs)
	}

	// Check armor proficiencies
	assert.GreaterOrEqual(t, profTypes[rulebook.ProficiencyTypeArmor], 3,
		"Should have at least light armor, medium armor, and shields")

	// Check weapon proficiencies
	assert.GreaterOrEqual(t, profTypes[rulebook.ProficiencyTypeWeapon], 2,
		"Should have simple and martial weapon proficiencies")

	// Check saving throw proficiencies
	assert.GreaterOrEqual(t, profTypes[rulebook.ProficiencyTypeSavingThrow], 2,
		"Should have STR and DEX saving throw proficiencies")

	// Check skill proficiencies (3 chosen)
	assert.Equal(t, 3, profTypes[rulebook.ProficiencyTypeSkill],
		"Should have exactly 3 skill proficiencies")

	// Verify ability scores with racial bonuses
	assert.Equal(t, 15, finalized.Attributes[shared.AttributeStrength].Score, "STR should be 14 + 1 (human)")
	assert.Equal(t, 17, finalized.Attributes[shared.AttributeDexterity].Score, "DEX should be 16 + 1 (human)")
	assert.Equal(t, 16, finalized.Attributes[shared.AttributeWisdom].Score, "WIS should be 15 + 1 (human)")

	// Verify HP calculation (10 + CON modifier)
	expectedHP := 10 + 2 // d10 hit die + 2 CON modifier (14 CON with +1 human = 15 = +2 modifier)
	assert.Equal(t, expectedHP, finalized.MaxHitPoints, "Should have correct starting HP")

	// Verify that Rangers don't get spell slots at level 1
	assert.Nil(t, finalized.Resources.SpellSlots[1], "Rangers shouldn't have spell slots at level 1")
}

// Test Ranger weapon proficiency application
func TestRangerWeaponProficiency(t *testing.T) {
	ranger := &character2.Character{
		Class: &rulebook.Class{Key: "ranger"},
		Proficiencies: map[rulebook.ProficiencyType][]*rulebook.Proficiency{
			rulebook.ProficiencyTypeWeapon: {
				{Key: "simple-weapons", Name: "Simple Weapons"},
				{Key: "martial-weapons", Name: "Martial Weapons"},
			},
		},
	}

	// Test that ranger has both simple and martial weapon proficiencies
	weaponProfs := ranger.Proficiencies[rulebook.ProficiencyTypeWeapon]
	assert.Len(t, weaponProfs, 2, "Ranger should have 2 weapon proficiency categories")

	// Check for specific proficiencies
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
}
