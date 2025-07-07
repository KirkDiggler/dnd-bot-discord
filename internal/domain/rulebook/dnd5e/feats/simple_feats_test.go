package feats_test

import (
	"sync"
	"testing"

	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/rulebook/dnd5e"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/rulebook/dnd5e/feats"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var registerOnce sync.Once

func setupRegistry() {
	registerOnce.Do(func() {
		feats.RegisterAll()
	})
}

func TestFeatApplication(t *testing.T) {
	// Initialize the registry once
	setupRegistry()

	t.Run("Alert feat application", func(t *testing.T) {
		char := &character.Character{
			ID:       "test-char",
			Name:     "Test Character",
			Features: []*rulebook.CharacterFeature{},
		}

		alertFeat := feats.NewAlertFeat()
		err := alertFeat.Apply(char)
		require.NoError(t, err)

		// Verify feat was added
		assert.Len(t, char.Features, 1)
		assert.Equal(t, "alert", char.Features[0].Key)
		assert.Equal(t, "Alert", char.Features[0].Name)
		assert.Equal(t, rulebook.FeatureTypeFeat, char.Features[0].Type)
	})

	t.Run("Great Weapon Master feat application", func(t *testing.T) {
		char := &character.Character{
			ID:       "test-char",
			Name:     "Test Character",
			Features: []*rulebook.CharacterFeature{},
		}

		gwmFeat := feats.NewGreatWeaponMasterFeat()
		err := gwmFeat.Apply(char)
		require.NoError(t, err)

		// Verify feat was added
		assert.Len(t, char.Features, 1)
		assert.Equal(t, "great_weapon_master", char.Features[0].Key)
		assert.Equal(t, "Great Weapon Master", char.Features[0].Name)
		assert.Equal(t, rulebook.FeatureTypeFeat, char.Features[0].Type)
	})

	t.Run("Lucky feat application", func(t *testing.T) {
		char := &character.Character{
			ID:       "test-char",
			Name:     "Test Character",
			Features: []*rulebook.CharacterFeature{},
			Resources: &character.CharacterResources{
				Abilities: make(map[string]*shared.ActiveAbility),
			},
		}

		luckyFeat := feats.NewLuckyFeat()
		err := luckyFeat.Apply(char)
		require.NoError(t, err)

		// Verify feat was added
		assert.Len(t, char.Features, 1)
		assert.Equal(t, "lucky", char.Features[0].Key)
		assert.Equal(t, 3, char.Features[0].Metadata["luck_points"])

		// Lucky feat doesn't add an ability to Resources, it tracks points in metadata
		// This is different from other abilities like Rage or Second Wind
	})

	t.Run("Sharpshooter feat application", func(t *testing.T) {
		char := &character.Character{
			ID:       "test-char",
			Name:     "Test Character",
			Features: []*rulebook.CharacterFeature{},
		}

		ssFeat := feats.NewSharpshooterFeat()
		err := ssFeat.Apply(char)
		require.NoError(t, err)

		// Verify feat was added
		assert.Len(t, char.Features, 1)
		assert.Equal(t, "sharpshooter", char.Features[0].Key)
		assert.Equal(t, "Sharpshooter", char.Features[0].Name)
		assert.Equal(t, rulebook.FeatureTypeFeat, char.Features[0].Type)
	})

	t.Run("Tough feat HP calculation", func(t *testing.T) {
		// Test at different levels
		levels := []struct {
			level         int
			baseHP        int
			expectedBonus int
		}{
			{1, 10, 2},
			{5, 30, 10},
			{10, 60, 20},
			{20, 120, 40},
		}

		for _, tc := range levels {
			t.Run(string(rune(tc.level))+" level", func(t *testing.T) {
				char := &character.Character{
					ID:           "test-char",
					Name:         "Test Character",
					Level:        tc.level,
					MaxHitPoints: tc.baseHP,
					Features:     []*rulebook.CharacterFeature{},
				}

				toughFeat := feats.NewToughFeat()
				err := toughFeat.Apply(char)
				require.NoError(t, err)

				// Verify HP bonus
				assert.Equal(t, tc.expectedBonus, char.Features[0].Metadata["hp_bonus"])
				assert.Equal(t, tc.baseHP+tc.expectedBonus, char.MaxHitPoints)
			})
		}
	})

	t.Run("War Caster feat application", func(t *testing.T) {
		char := &character.Character{
			ID:       "test-char",
			Name:     "Test Character",
			Features: []*rulebook.CharacterFeature{},
		}

		wcFeat := feats.NewWarCasterFeat()
		err := wcFeat.Apply(char)
		require.NoError(t, err)

		// Verify feat was added
		assert.Len(t, char.Features, 1)
		assert.Equal(t, "war_caster", char.Features[0].Key)
		assert.Equal(t, "War Caster", char.Features[0].Name)
		assert.Equal(t, rulebook.FeatureTypeFeat, char.Features[0].Type)
	})
}

func TestFeatRegistry(t *testing.T) {
	// Initialize the registry once
	setupRegistry()

	t.Run("All feats registered", func(t *testing.T) {
		// Get all feats from registry
		allFeats := feats.GlobalRegistry.List()

		// Expected feat keys
		expectedFeats := []string{
			"alert",
			"great_weapon_master",
			"lucky",
			"sharpshooter",
			"tough",
			"war_caster",
		}

		// Verify all expected feats are registered
		featMap := make(map[string]bool)
		for _, feat := range allFeats {
			featMap[feat.Key()] = true
		}

		for _, key := range expectedFeats {
			assert.True(t, featMap[key], "Feat %s should be registered", key)
		}

		// Verify count
		assert.GreaterOrEqual(t, len(allFeats), len(expectedFeats))
	})

	t.Run("Get feat by key", func(t *testing.T) {
		// Test getting existing feat
		alertFeat, exists := feats.GlobalRegistry.Get("alert")
		assert.True(t, exists)
		assert.NotNil(t, alertFeat)
		assert.Equal(t, "alert", alertFeat.Key())
		assert.Equal(t, "Alert", alertFeat.Name())

		// Test getting non-existent feat
		nilFeat, exists := feats.GlobalRegistry.Get("nonexistent")
		assert.False(t, exists)
		assert.Nil(t, nilFeat)
	})

	t.Run("Apply feat through registry", func(t *testing.T) {
		// Create character
		char := &character.Character{
			ID:       "test-char",
			Name:     "Test Character",
			Features: []*rulebook.CharacterFeature{},
			Resources: &character.CharacterResources{
				Abilities: make(map[string]*shared.ActiveAbility),
			},
		}

		// Apply feat by key
		err := feats.GlobalRegistry.ApplyFeat("lucky", char, nil)
		require.NoError(t, err)

		// Verify feat was applied
		assert.Len(t, char.Features, 1)
		assert.Equal(t, "lucky", char.Features[0].Key)

		// Test applying same feat again
		err = feats.GlobalRegistry.ApplyFeat("lucky", char, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already has feat")

		// Test applying non-existent feat
		err = feats.GlobalRegistry.ApplyFeat("nonexistent", char, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("Available feats for character", func(t *testing.T) {
		// Create character with one feat
		char := &character.Character{
			ID:   "test-char",
			Name: "Test Character",
			Features: []*rulebook.CharacterFeature{
				{
					Key:  "lucky",
					Name: "Lucky",
					Type: rulebook.FeatureTypeFeat,
				},
			},
		}

		// Get available feats
		available := feats.GlobalRegistry.AvailableFeats(char)

		// Should not include Lucky since character already has it
		for _, feat := range available {
			assert.NotEqual(t, "lucky", feat.Key(), "Lucky should not be in available feats")
		}

		// Should include other feats
		foundAlert := false
		for _, feat := range available {
			if feat.Key() == "alert" {
				foundAlert = true
				break
			}
		}
		assert.True(t, foundAlert, "Alert should be in available feats")
	})
}

func TestFeatMetadata(t *testing.T) {
	t.Run("Great Weapon Master power attack toggle", func(t *testing.T) {
		char := &character.Character{
			ID:   "test-char",
			Name: "GWM Character",
			Features: []*rulebook.CharacterFeature{
				{
					Key:  "great_weapon_master",
					Name: "Great Weapon Master",
					Type: rulebook.FeatureTypeFeat,
					Metadata: map[string]interface{}{
						"power_attack": true,
					},
				},
			},
		}

		// Check metadata
		powerAttack, ok := char.Features[0].Metadata["power_attack"].(bool)
		assert.True(t, ok, "power_attack should be a bool")
		assert.True(t, powerAttack)

		// Toggle off
		char.Features[0].Metadata["power_attack"] = false
		powerAttack, ok = char.Features[0].Metadata["power_attack"].(bool)
		assert.True(t, ok, "power_attack should be a bool")
		assert.False(t, powerAttack)
	})

	t.Run("Sharpshooter power shot toggle", func(t *testing.T) {
		char := &character.Character{
			ID:   "test-char",
			Name: "Sharpshooter Character",
			Features: []*rulebook.CharacterFeature{
				{
					Key:  "sharpshooter",
					Name: "Sharpshooter",
					Type: rulebook.FeatureTypeFeat,
					Metadata: map[string]interface{}{
						"power_shot": true,
					},
				},
			},
		}

		// Check metadata
		powerShot, ok := char.Features[0].Metadata["power_shot"].(bool)
		assert.True(t, ok, "power_shot should be a bool")
		assert.True(t, powerShot)

		// Toggle off
		char.Features[0].Metadata["power_shot"] = false
		powerShot, ok = char.Features[0].Metadata["power_shot"].(bool)
		assert.True(t, ok, "power_shot should be a bool")
		assert.False(t, powerShot)
	})

	t.Run("Lucky feat point tracking", func(t *testing.T) {
		char := &character.Character{
			ID:   "test-char",
			Name: "Lucky Character",
			Features: []*rulebook.CharacterFeature{
				{
					Key:  "lucky",
					Name: "Lucky",
					Type: rulebook.FeatureTypeFeat,
					Metadata: map[string]interface{}{
						"luck_points": 3,
					},
				},
			},
			Resources: &character.CharacterResources{
				Abilities: make(map[string]*shared.ActiveAbility),
			},
		}

		// Lucky feat tracks points in metadata, not as an ability
		assert.Equal(t, 3, char.Features[0].Metadata["luck_points"])

		// Simulate using a luck point
		char.Features[0].Metadata["luck_points"] = 2
		assert.Equal(t, 2, char.Features[0].Metadata["luck_points"])
	})
}

func TestFeatPrerequisites(t *testing.T) {
	t.Run("Feats with no prerequisites", func(t *testing.T) {
		char := &character.Character{
			ID:       "test-char",
			Name:     "Test Character",
			Features: []*rulebook.CharacterFeature{},
		}

		// These feats have no prerequisites (except war_caster requires spellcasting)
		noPrereqFeats := []string{"alert", "great_weapon_master", "lucky", "sharpshooter", "tough"}

		for _, featKey := range noPrereqFeats {
			feat, exists := feats.GlobalRegistry.Get(featKey)
			assert.True(t, exists, "Feat %s should exist", featKey)
			assert.True(t, feat.CanTake(char), "Character should be able to take %s", featKey)
		}
	})

	t.Run("War Caster requires spellcasting", func(t *testing.T) {
		// Character without spellcasting
		nonCaster := &character.Character{
			ID:       "test-char",
			Name:     "Fighter",
			Features: []*rulebook.CharacterFeature{},
		}

		wcFeat, exists := feats.GlobalRegistry.Get("war_caster")
		assert.True(t, exists)
		assert.False(t, wcFeat.CanTake(nonCaster), "Non-caster should not be able to take War Caster")

		// Character with spellcasting (has spell slots)
		caster := &character.Character{
			ID:       "test-char",
			Name:     "Wizard",
			Features: []*rulebook.CharacterFeature{},
			Resources: &character.CharacterResources{
				SpellSlots: map[int]shared.SpellSlotInfo{
					1: {Max: 2, Remaining: 2, Source: "spellcasting"},
				},
			},
		}

		assert.True(t, wcFeat.CanTake(caster), "Caster should be able to take War Caster")
	})
}
