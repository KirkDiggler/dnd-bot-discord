package entities

import (
	"testing"

	"github.com/KirkDiggler/dnd-bot-discord/internal/effects"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCharacter_RangerInitialization(t *testing.T) {
	// Create a ranger character
	ranger := &Character{
		ID:    "ranger_1",
		Name:  "Legolas",
		Level: 1,
		Class: &Class{
			Key:  "ranger",
			Name: "Ranger",
		},
		Attributes: map[Attribute]*AbilityScore{
			AttributeStrength:     {Score: 14, Bonus: 2},
			AttributeDexterity:    {Score: 16, Bonus: 3},
			AttributeConstitution: {Score: 13, Bonus: 1},
			AttributeIntelligence: {Score: 10, Bonus: 0},
			AttributeWisdom:       {Score: 15, Bonus: 2},
			AttributeCharisma:     {Score: 8, Bonus: -1},
		},
		MaxHitPoints:     11, // d10 + 1 CON
		CurrentHitPoints: 11,
	}

	// Initialize resources which should add ranger status effects
	ranger.InitializeResources()

	// Check that the ranger has the favored enemy effect
	activeEffects := ranger.GetActiveStatusEffects()
	require.NotEmpty(t, activeEffects, "Ranger should have active status effects")

	// Find the favored enemy effect
	var favoredEnemyEffect *effects.StatusEffect
	for _, effect := range activeEffects {
		if effect.Name == "Favored Enemy" {
			favoredEnemyEffect = effect
			break
		}
	}

	require.NotNil(t, favoredEnemyEffect, "Ranger should have Favored Enemy effect")
	assert.Equal(t, effects.SourceFeature, favoredEnemyEffect.Source)
	assert.Equal(t, effects.DurationPermanent, favoredEnemyEffect.Duration.Type)

	// Check that the effect has the expected conditions (default to orc for now)
	hasOrcCondition := false
	for _, cond := range favoredEnemyEffect.Conditions {
		if cond.Type == "enemy_type" && cond.Value == "orc" {
			hasOrcCondition = true
			break
		}
	}
	assert.True(t, hasOrcCondition, "Favored Enemy should have orc as the enemy type")

	// Verify the effect provides advantage on survival checks
	skillMods := ranger.GetEffectManager().GetModifiers(effects.TargetSkillCheck, map[string]string{
		"enemy_type": "orc",
		"skill":      "survival",
	})

	hasAdvantage := false
	for _, mod := range skillMods {
		if mod.Value == "advantage" && mod.SubTarget == "survival" {
			hasAdvantage = true
			break
		}
	}
	assert.True(t, hasAdvantage, "Should have advantage on survival checks vs orcs")

	// Verify the effect provides advantage on intelligence checks
	intMods := ranger.GetEffectManager().GetModifiers(effects.TargetAbilityScore, map[string]string{
		"enemy_type": "orc",
	})

	hasIntAdvantage := false
	for _, mod := range intMods {
		if mod.Value == "advantage" && mod.SubTarget == "intelligence" {
			hasIntAdvantage = true
			break
		}
	}
	assert.True(t, hasIntAdvantage, "Should have advantage on intelligence checks vs orcs")

	// Verify no advantage against other enemy types
	otherMods := ranger.GetEffectManager().GetModifiers(effects.TargetSkillCheck, map[string]string{
		"enemy_type": "goblin",
		"skill":      "survival",
	})
	assert.Empty(t, otherMods, "Should not have advantage against non-favored enemies")
}

func TestCharacter_RangerMultipleEffects(t *testing.T) {
	// Create a ranger
	ranger := &Character{
		ID:    "ranger_1",
		Name:  "Aragorn",
		Level: 1,
		Class: &Class{
			Key:  "ranger",
			Name: "Ranger",
		},
	}

	// Initialize to get favored enemy
	ranger.InitializeResources()

	// Add another effect (like Bless)
	blessEffect := effects.BuildBlessEffect()
	err := ranger.AddStatusEffect(blessEffect)
	require.NoError(t, err)

	// Verify both effects are active
	activeEffects := ranger.GetActiveStatusEffects()
	assert.Len(t, activeEffects, 2, "Should have 2 active effects")

	effectNames := make(map[string]bool)
	for _, effect := range activeEffects {
		effectNames[effect.Name] = true
	}

	assert.True(t, effectNames["Favored Enemy"], "Should have Favored Enemy")
	assert.True(t, effectNames["Bless"], "Should have Bless")
}
