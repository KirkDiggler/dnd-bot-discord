package features_test

import (
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/rulebook"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/rulebook/features"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPassiveFeatures(t *testing.T) {
	t.Run("KeenSenses grants Perception proficiency", func(t *testing.T) {
		// Create a character without Perception proficiency
		char := &character.Character{
			ID:    "test-char",
			Name:  "Test Elf",
			Level: 1,
			Features: []*rulebook.CharacterFeature{
				{
					Key:    "keen_senses",
					Name:   "Keen Senses",
					Type:   rulebook.FeatureTypeRacial,
					Source: "Elf",
				},
			},
			Proficiencies: make(map[rulebook.ProficiencyType][]*rulebook.Proficiency),
		}

		// Apply passive effects
		err := features.DefaultRegistry.ApplyAllPassiveEffects(char)
		require.NoError(t, err)

		// Verify Perception proficiency was granted
		assert.True(t, char.HasSkillProficiency("skill-perception"), "Should have Perception proficiency from Keen Senses")
	})

	t.Run("Darkvision provides display info", func(t *testing.T) {
		char := &character.Character{
			ID:    "test-char",
			Name:  "Test Dwarf",
			Level: 1,
			Features: []*rulebook.CharacterFeature{
				{
					Key:    "darkvision",
					Name:   "Darkvision",
					Type:   rulebook.FeatureTypeRacial,
					Source: "Dwarf",
				},
			},
		}

		// Get passive display info
		displayInfo := features.DefaultRegistry.GetAllPassiveDisplayInfo(char)
		assert.Len(t, displayInfo, 1, "Should have one display info item")
		assert.Contains(t, displayInfo[0], "Darkvision", "Should contain Darkvision info")
		assert.Contains(t, displayInfo[0], "60 feet", "Should mention 60 feet range")
	})

	t.Run("Multiple passive features work together", func(t *testing.T) {
		char := &character.Character{
			ID:    "test-char",
			Name:  "Test Elf",
			Level: 1,
			Features: []*rulebook.CharacterFeature{
				{
					Key:    "keen_senses",
					Name:   "Keen Senses",
					Type:   rulebook.FeatureTypeRacial,
					Source: "Elf",
				},
				{
					Key:    "darkvision",
					Name:   "Darkvision",
					Type:   rulebook.FeatureTypeRacial,
					Source: "Elf",
				},
				{
					Key:    "fey_ancestry",
					Name:   "Fey Ancestry",
					Type:   rulebook.FeatureTypeRacial,
					Source: "Elf",
				},
			},
			Proficiencies: make(map[rulebook.ProficiencyType][]*rulebook.Proficiency),
		}

		// Apply passive effects
		err := features.DefaultRegistry.ApplyAllPassiveEffects(char)
		require.NoError(t, err)

		// Verify Perception proficiency was granted
		assert.True(t, char.HasSkillProficiency("skill-perception"), "Should have Perception proficiency")

		// Get all display info
		displayInfo := features.DefaultRegistry.GetAllPassiveDisplayInfo(char)
		assert.GreaterOrEqual(t, len(displayInfo), 2, "Should have multiple display items")

		// Check that expected features are represented
		infoText := ""
		for _, info := range displayInfo {
			infoText += info + " "
		}
		assert.Contains(t, infoText, "Darkvision", "Should contain Darkvision")
		assert.Contains(t, infoText, "Fey Ancestry", "Should contain Fey Ancestry")
	})

	t.Run("Unknown features don't cause errors", func(t *testing.T) {
		char := &character.Character{
			ID:    "test-char",
			Name:  "Test Character",
			Level: 1,
			Features: []*rulebook.CharacterFeature{
				{
					Key:    "unknown_feature",
					Name:   "Unknown Feature",
					Type:   rulebook.FeatureTypeRacial,
					Source: "Unknown",
				},
				{
					Key:    "keen_senses",
					Name:   "Keen Senses",
					Type:   rulebook.FeatureTypeRacial,
					Source: "Elf",
				},
			},
			Proficiencies: make(map[rulebook.ProficiencyType][]*rulebook.Proficiency),
		}

		// Apply passive effects - should not error even with unknown feature
		err := features.DefaultRegistry.ApplyAllPassiveEffects(char)
		require.NoError(t, err)

		// Known feature should still work
		assert.True(t, char.HasSkillProficiency("skill-perception"), "Should have Perception proficiency from known feature")
	})

	t.Run("Feature handlers can be retrieved", func(t *testing.T) {
		// Test that we can get specific handlers
		handler, exists := features.DefaultRegistry.GetHandler("keen_senses")
		assert.True(t, exists, "Should find keen_senses handler")
		assert.Equal(t, "keen_senses", handler.GetKey(), "Handler should return correct key")

		handler, exists = features.DefaultRegistry.GetHandler("darkvision")
		assert.True(t, exists, "Should find darkvision handler")
		assert.Equal(t, "darkvision", handler.GetKey(), "Handler should return correct key")

		// Test non-existent handler
		_, exists = features.DefaultRegistry.GetHandler("non_existent")
		assert.False(t, exists, "Should not find non-existent handler")
	})
}
