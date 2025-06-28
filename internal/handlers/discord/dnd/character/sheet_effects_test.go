package character

import (
	"strings"
	"testing"

	"github.com/KirkDiggler/dnd-bot-discord/internal/effects"
	"github.com/KirkDiggler/dnd-bot-discord/internal/entities"
	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildActiveEffectsDisplay(t *testing.T) {
	t.Run("no active effects", func(t *testing.T) {
		char := &entities.Character{
			ID:   "test_char",
			Name: "Test Character",
		}

		lines := buildActiveEffectsDisplay(char)
		assert.Equal(t, []string{"*No active effects*"}, lines)
	})

	t.Run("single rage effect", func(t *testing.T) {
		char := &entities.Character{
			ID:    "barbarian_char",
			Name:  "Grog",
			Level: 1,
			Class: &entities.Class{Key: "barbarian"},
		}

		// Initialize resources to get the effect manager
		char.InitializeResources()

		// Add rage effect
		rageEffect := effects.BuildRageEffect(1)
		err := char.AddStatusEffect(rageEffect)
		require.NoError(t, err)

		lines := buildActiveEffectsDisplay(char)

		// Should show the rage effect under Abilities
		assert.Contains(t, lines, "**Abilities:**")
		assert.Contains(t, lines, "• **Rage** (10 rounds)")
	})

	t.Run("ranger favored enemy effect", func(t *testing.T) {
		char := &entities.Character{
			ID:    "ranger_char",
			Name:  "Legolas",
			Level: 1,
			Class: &entities.Class{Key: "ranger"},
		}

		// Initialize resources - this should add the favored enemy effect
		char.InitializeResources()

		lines := buildActiveEffectsDisplay(char)

		// Should show the favored enemy effect under Features (permanent effects don't show duration)
		assert.Contains(t, lines, "**Features:**")
		assert.Contains(t, lines, "• **Favored Enemy**")

		// Permanent effects shouldn't show duration
		joined := strings.Join(lines, "\n")
		assert.NotContains(t, joined, "permanent")
	})

	t.Run("multiple effects from different sources", func(t *testing.T) {
		char := &entities.Character{
			ID:    "multi_char",
			Name:  "Multi",
			Level: 1,
		}

		// Initialize to get effect manager
		char.InitializeResources()

		// Add effects from different sources
		rageEffect := effects.BuildRageEffect(1)
		blessEffect := effects.BuildBlessEffect()

		err := char.AddStatusEffect(rageEffect)
		require.NoError(t, err)

		err = char.AddStatusEffect(blessEffect)
		require.NoError(t, err)

		lines := buildActiveEffectsDisplay(char)
		joined := strings.Join(lines, "\n")

		// Should have both abilities and spells sections
		assert.Contains(t, joined, "**Abilities:**")
		assert.Contains(t, joined, "• **Rage** (10 rounds)")

		assert.Contains(t, joined, "**Spells:**")
		assert.Contains(t, joined, "• **Bless** (10 rounds, concentration)")
	})

	t.Run("concentration effects show properly", func(t *testing.T) {
		char := &entities.Character{
			ID:   "caster_char",
			Name: "Wizard",
		}

		char.InitializeResources()

		// Add a bless effect (requires concentration)
		blessEffect := effects.BuildBlessEffect()
		err := char.AddStatusEffect(blessEffect)
		require.NoError(t, err)

		lines := buildActiveEffectsDisplay(char)
		joined := strings.Join(lines, "\n")

		// Should show concentration in the duration
		assert.Contains(t, joined, "concentration")
	})
}

func TestCharacterSheetWithActiveEffects(t *testing.T) {
	// Create a barbarian with rage effect
	char := &entities.Character{
		ID:    "test_barbarian",
		Name:  "Conan",
		Level: 1,
		Class: &entities.Class{
			Key:  "barbarian",
			Name: "Barbarian",
		},
		Race: &entities.Race{
			Key:  "human",
			Name: "Human",
		},
		Attributes: map[entities.Attribute]*entities.AbilityScore{
			entities.AttributeStrength:     {Score: 16, Bonus: 3},
			entities.AttributeDexterity:    {Score: 14, Bonus: 2},
			entities.AttributeConstitution: {Score: 15, Bonus: 2},
			entities.AttributeIntelligence: {Score: 10, Bonus: 0},
			entities.AttributeWisdom:       {Score: 12, Bonus: 1},
			entities.AttributeCharisma:     {Score: 8, Bonus: -1},
		},
		MaxHitPoints:     13, // d12 + 2 CON
		CurrentHitPoints: 13,
		AC:               12, // 10 + 2 DEX
	}

	// Initialize and add rage effect
	char.InitializeResources()
	rageEffect := effects.BuildRageEffect(1)
	err := char.AddStatusEffect(rageEffect)
	require.NoError(t, err)

	// Build character sheet embed
	embed := BuildCharacterSheetEmbed(char)

	// Check that active effects field exists
	var activeEffectsField *discordgo.MessageEmbedField
	for _, field := range embed.Fields {
		if field.Name == "🔮 Active Effects" {
			activeEffectsField = field
			break
		}
	}

	require.NotNil(t, activeEffectsField, "Character sheet should have Active Effects field")
	assert.Contains(t, activeEffectsField.Value, "**Abilities:**")
	assert.Contains(t, activeEffectsField.Value, "• **Rage** (10 rounds)")
}
