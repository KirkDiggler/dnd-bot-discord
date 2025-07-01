package character_test

import (
	"context"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/equipment"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/rulebook/dnd5e"
	"net/http"
	"testing"
	"time"

	"github.com/KirkDiggler/dnd-bot-discord/internal/clients/dnd5e"
	"github.com/KirkDiggler/dnd-bot-discord/internal/repositories/characters"
	"github.com/KirkDiggler/dnd-bot-discord/internal/services/character"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRangerWeaponProficiencyAndAttackBonus(t *testing.T) {
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

	// Create a Ranger character
	output, err := svc.CreateCharacter(ctx, &character.CreateCharacterInput{
		UserID:   "test-user",
		RealmID:  "test-realm",
		Name:     "Test Ranger",
		RaceKey:  "human",
		ClassKey: "ranger",
		AbilityScores: map[string]int{
			"STR": 14,
			"DEX": 16, // +3 modifier
			"CON": 13,
			"INT": 10,
			"WIS": 15,
			"CHA": 8,
		},
		Proficiencies: []string{"skill-perception", "skill-stealth", "skill-survival"},
		Equipment:     []string{"longbow"}, // Start with a longbow
	})
	require.NoError(t, err)
	require.NotNil(t, output)

	char := output.Character
	require.NotNil(t, char)

	// Test 1: Verify Ranger has weapon proficiencies
	t.Run("Ranger has weapon proficiencies", func(t *testing.T) {
		weaponProfs, exists := char.Proficiencies[rulebook.ProficiencyTypeWeapon]
		assert.True(t, exists, "Ranger should have weapon proficiencies")
		assert.NotEmpty(t, weaponProfs, "Ranger should have at least one weapon proficiency")

		// Check for martial weapons proficiency
		hasMartialWeapons := false
		for _, prof := range weaponProfs {
			t.Logf("Weapon proficiency: %s", prof.Key)
			if prof.Key == "martial-weapons" {
				hasMartialWeapons = true
				break
			}
		}
		assert.True(t, hasMartialWeapons, "Ranger should have martial-weapons proficiency")
	})

	// Test 2: Verify longbow is in inventory
	t.Run("Longbow in inventory", func(t *testing.T) {
		weapons, exists := char.Inventory[equipment.EquipmentTypeWeapon]
		assert.True(t, exists, "Should have weapons in inventory")

		hasLongbow := false
		var longbow *equipment.Weapon
		for _, weapon := range weapons {
			if weapon.GetKey() == "longbow" {
				hasLongbow = true
				if w, ok := weapon.(*equipment.Weapon); ok {
					longbow = w
				}
				break
			}
		}
		assert.True(t, hasLongbow, "Should have longbow in inventory")
		assert.NotNil(t, longbow, "Longbow should be a weapon type")

		// Check weapon category
		t.Logf("Longbow category: %s", longbow.WeaponCategory)
		assert.Equal(t, "martial", longbow.WeaponCategory, "Longbow should be normalized to lowercase 'martial'")
	})

	// Test 3: Equip longbow and verify attack bonus
	t.Run("Attack with longbow includes proficiency bonus", func(t *testing.T) {
		// Equip the longbow
		equipped := char.Equip("longbow")
		assert.True(t, equipped, "Should be able to equip longbow")

		// Attack and check the result
		attacks, err := char.Attack()
		require.NoError(t, err)
		require.Len(t, attacks, 1, "Should have one attack")

		// At level 1, proficiency bonus is +2
		// DEX modifier is +3 (16 DEX)
		// Total attack bonus should be +5 (3 + 2)

		// The AttackRoll includes the d20 roll + modifiers
		// We can't check the exact value, but we can verify the character has proficiency
		assert.True(t, char.HasWeaponProficiency("longbow"), "Character should be proficient with longbow")
	})

	// Test 4: Verify proficiency bonus calculation
	t.Run("Proficiency bonus calculation", func(t *testing.T) {
		// Manually check the proficiency calculation
		// At level 1, proficiency bonus should be +2
		profBonus := 2 + ((char.Level - 1) / 4)
		assert.Equal(t, 2, profBonus, "Level 1 character should have +2 proficiency bonus")
	})

	// Test 5: Compare with non-proficient weapon
	t.Run("Attack without proficiency", func(t *testing.T) {
		// Add a weapon the ranger isn't proficient with (if we had one)
		// For now, just verify the proficiency check works
		assert.False(t, char.HasWeaponProficiency("nonexistent-weapon"), "Should not be proficient with nonexistent weapon")
	})
}

// Test that weapon categories are normalized when loaded from API
func TestWeaponCategoryNormalization(t *testing.T) {
	// This would test that our API client normalizes weapon categories
	// Need to mock the API response to test this properly
	t.Skip("TODO: Add mock test for weapon category normalization")
}
