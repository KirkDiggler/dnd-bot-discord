package character

import (
	"testing"

	"github.com/KirkDiggler/dnd-bot-discord/internal/entities"
	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildCharacterSheetEmbed(t *testing.T) {
	// Create a test character
	char := &entities.Character{
		ID:               "test-char-1",
		Name:             "Aragorn",
		Level:            5,
		CurrentHitPoints: 45,
		MaxHitPoints:     45,
		AC:               16,
		Class: &entities.Class{
			Name: "Ranger",
		},
		Race: &entities.Race{
			Name: "Human",
		},
		Attributes: map[entities.Attribute]*entities.AbilityScore{
			entities.AttributeStrength:     {Score: 16, Bonus: 3},
			entities.AttributeDexterity:    {Score: 14, Bonus: 2},
			entities.AttributeConstitution: {Score: 14, Bonus: 2},
			entities.AttributeIntelligence: {Score: 12, Bonus: 1},
			entities.AttributeWisdom:       {Score: 13, Bonus: 1},
			entities.AttributeCharisma:     {Score: 10, Bonus: 0},
		},
		EquippedSlots: map[entities.Slot]entities.Equipment{
			entities.SlotMainHand: &entities.Weapon{
				Base: entities.BasicEquipment{
					Key:  "longsword",
					Name: "Longsword",
				},
			},
			entities.SlotBody: &entities.Armor{
				Base: entities.BasicEquipment{
					Key:  "chain_mail",
					Name: "Chain Mail",
				},
			},
		},
		Proficiencies: map[entities.ProficiencyType][]*entities.Proficiency{
			entities.ProficiencyTypeWeapon: {
				{Key: "simple-weapons", Name: "Simple Weapons"},
				{Key: "martial-weapons", Name: "Martial Weapons"},
			},
			entities.ProficiencyTypeArmor: {
				{Key: "light-armor", Name: "Light Armor"},
				{Key: "medium-armor", Name: "Medium Armor"},
				{Key: "shields", Name: "Shields"},
			},
			entities.ProficiencyTypeSkill: {
				{Key: "survival", Name: "Survival"},
				{Key: "perception", Name: "Perception"},
			},
		},
	}

	embed := BuildCharacterSheetEmbed(char)

	// Verify embed properties
	assert.Equal(t, "Human Aragorn - Level 5 Ranger", embed.Title)
	assert.Contains(t, embed.Description, "**HP:** 45/45")
	assert.Contains(t, embed.Description, "**AC:** 16")
	assert.Contains(t, embed.Description, "**Initiative:** +2")

	// Verify fields
	require.Len(t, embed.Fields, 3)

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
	assert.Equal(t, "character:details:test-char-123", button2.CustomID)

	button3, ok := actionRow.Components[2].(discordgo.Button)
	require.True(t, ok)
	assert.Equal(t, "character:sheet_refresh:test-char-123", button3.CustomID)
}
