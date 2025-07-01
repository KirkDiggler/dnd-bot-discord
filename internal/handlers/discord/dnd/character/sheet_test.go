package character

import (
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/equipment"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/rulebook"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/shared"
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildCharacterSheetEmbed(t *testing.T) {
	// Create a test character
	char := &character.Character{
		ID:               "test-char-1",
		Name:             "Aragorn",
		Level:            5,
		CurrentHitPoints: 45,
		MaxHitPoints:     45,
		AC:               16,
		Class: &rulebook.Class{
			Name: "Ranger",
		},
		Race: &rulebook.Race{
			Name: "Human",
		},
		Attributes: map[shared.Attribute]*character.AbilityScore{
			shared.AttributeStrength:     {Score: 16, Bonus: 3},
			shared.AttributeDexterity:    {Score: 14, Bonus: 2},
			shared.AttributeConstitution: {Score: 14, Bonus: 2},
			shared.AttributeIntelligence: {Score: 12, Bonus: 1},
			shared.AttributeWisdom:       {Score: 13, Bonus: 1},
			shared.AttributeCharisma:     {Score: 10, Bonus: 0},
		},
		EquippedSlots: map[shared.Slot]equipment.Equipment{
			shared.SlotMainHand: &equipment.Weapon{
				Base: equipment.BasicEquipment{
					Key:  "longsword",
					Name: "Longsword",
				},
			},
			shared.SlotBody: &equipment.Armor{
				Base: equipment.BasicEquipment{
					Key:  "chain_mail",
					Name: "Chain Mail",
				},
			},
		},
		Proficiencies: map[rulebook.ProficiencyType][]*rulebook.Proficiency{
			rulebook.ProficiencyTypeWeapon: {
				{Key: "simple-weapons", Name: "Simple Weapons"},
				{Key: "martial-weapons", Name: "Martial Weapons"},
			},
			rulebook.ProficiencyTypeArmor: {
				{Key: "light-armor", Name: "Light Armor"},
				{Key: "medium-armor", Name: "Medium Armor"},
				{Key: "shields", Name: "Shields"},
			},
			rulebook.ProficiencyTypeSkill: {
				{Key: "survival", Name: "Survival"},
				{Key: "perception", Name: "Perception"},
			},
		},
		Features: []*rulebook.CharacterFeature{
			{
				Key:         "favored_enemy",
				Name:        "Favored Enemy",
				Description: "You have advantage on Wisdom (Survival) checks to track your favored enemies.",
				Type:        rulebook.FeatureTypeClass,
				Level:       1,
				Source:      "Ranger",
				Metadata: map[string]any{
					"enemy_type": "orc",
				},
			},
			{
				Key:         "natural_explorer",
				Name:        "Natural Explorer",
				Description: "You are particularly familiar with one type of natural environment.",
				Type:        rulebook.FeatureTypeClass,
				Level:       1,
				Source:      "Ranger",
			},
		},
	}

	embed := BuildCharacterSheetEmbed(char)

	// Verify embed properties
	assert.Equal(t, "Human Aragorn - Level 5 Ranger", embed.Title)
	assert.Contains(t, embed.Description, "**HP:** 45/45")
	assert.Contains(t, embed.Description, "**AC:** 16")
	assert.Contains(t, embed.Description, "**Initiative:** +2")

	// Verify fields (5 because ranger has features/effects)
	require.Len(t, embed.Fields, 5)

	// Check ability scores field
	assert.Equal(t, "📊 Ability Scores", embed.Fields[0].Name)
	assert.Contains(t, embed.Fields[0].Value, "**STR:** 16 (+3)")
	assert.Contains(t, embed.Fields[0].Value, "**DEX:** 14 (+2)")

	// Check equipment field
	assert.Equal(t, "⚔️ Equipment", embed.Fields[1].Name)
	assert.Contains(t, embed.Fields[1].Value, "**Main Hand:** Longsword")
	assert.Contains(t, embed.Fields[1].Value, "**Armor:** Chain Mail")

	// Check proficiencies field
	assert.Equal(t, "📚 Proficiencies", embed.Fields[2].Name)
	assert.Contains(t, embed.Fields[2].Value, "**Weapons:** Simple Weapons, Martial Weapons")
	assert.Contains(t, embed.Fields[2].Value, "**Skills:** Survival, Perception")

	// Check features field
	assert.Equal(t, "✨ Features", embed.Fields[3].Name)
	assert.Contains(t, embed.Fields[3].Value, "**Class Features:**")
	assert.Contains(t, embed.Fields[3].Value, "• Favored Enemy")
	assert.Contains(t, embed.Fields[3].Value, "• Natural Explorer")
}

func TestBuildCharacterSheetComponents(t *testing.T) {
	characterID := "test-char-123"
	components := BuildCharacterSheetComponents(characterID)

	require.Len(t, components, 1)

	// Check that we have an action row
	actionRow, ok := components[0].(discordgo.ActionsRow)
	require.True(t, ok)
	require.Len(t, actionRow.Components, 3)

	// Verify button custom IDs
	button1, ok := actionRow.Components[0].(discordgo.Button)
	require.True(t, ok)
	assert.Equal(t, "character:inventory:test-char-123", button1.CustomID)

	button2, ok := actionRow.Components[1].(discordgo.Button)
	require.True(t, ok)
	assert.Equal(t, "character:edit_menu:test-char-123", button2.CustomID)

	button3, ok := actionRow.Components[2].(discordgo.Button)
	require.True(t, ok)
	assert.Equal(t, "character:sheet_refresh:test-char-123", button3.CustomID)
}
