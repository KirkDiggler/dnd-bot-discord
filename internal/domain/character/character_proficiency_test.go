package character

import (
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/rulebook"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/shared"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCharacterProficiencyBonus(t *testing.T) {
	tests := []struct {
		name     string
		level    int
		expected int
	}{
		{"Level 0 (default)", 0, 2},
		{"Level 1", 1, 2},
		{"Level 4", 4, 2},
		{"Level 5", 5, 3},
		{"Level 8", 8, 3},
		{"Level 9", 9, 4},
		{"Level 12", 12, 4},
		{"Level 13", 13, 5},
		{"Level 16", 16, 5},
		{"Level 17", 17, 6},
		{"Level 20", 20, 6},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			char := &Character{Level: tt.level}
			assert.Equal(t, tt.expected, char.GetProficiencyBonus())
		})
	}
}

func TestCharacterSavingThrows(t *testing.T) {
	// Create a test character
	char := &Character{
		Level: 5, // +3 proficiency bonus
		Attributes: map[shared.Attribute]*AbilityScore{
			shared.AttributeStrength:     {Score: 16, Bonus: 3},
			shared.AttributeDexterity:    {Score: 14, Bonus: 2},
			shared.AttributeConstitution: {Score: 13, Bonus: 1},
			shared.AttributeIntelligence: {Score: 10, Bonus: 0},
			shared.AttributeWisdom:       {Score: 12, Bonus: 1},
			shared.AttributeCharisma:     {Score: 8, Bonus: -1},
		},
		Proficiencies: map[rulebook.ProficiencyType][]*rulebook.Proficiency{
			rulebook.ProficiencyTypeSavingThrow: {
				{Key: "saving-throw-str", Name: "Strength", Type: rulebook.ProficiencyTypeSavingThrow},
				{Key: "saving-throw-con", Name: "Constitution", Type: rulebook.ProficiencyTypeSavingThrow},
			},
		},
	}

	t.Run("HasSavingThrowProficiency", func(t *testing.T) {
		assert.True(t, char.HasSavingThrowProficiency(shared.AttributeStrength))
		assert.True(t, char.HasSavingThrowProficiency(shared.AttributeConstitution))
		assert.False(t, char.HasSavingThrowProficiency(shared.AttributeDexterity))
		assert.False(t, char.HasSavingThrowProficiency(shared.AttributeWisdom))
	})

	t.Run("GetSavingThrowBonus", func(t *testing.T) {
		// With proficiency: ability mod + prof bonus
		assert.Equal(t, 6, char.GetSavingThrowBonus(shared.AttributeStrength))     // 3 + 3
		assert.Equal(t, 4, char.GetSavingThrowBonus(shared.AttributeConstitution)) // 1 + 3

		// Without proficiency: just ability mod
		assert.Equal(t, 2, char.GetSavingThrowBonus(shared.AttributeDexterity)) // 2 + 0
		assert.Equal(t, 1, char.GetSavingThrowBonus(shared.AttributeWisdom))    // 1 + 0
		assert.Equal(t, -1, char.GetSavingThrowBonus(shared.AttributeCharisma)) // -1 + 0
	})

	t.Run("RollSavingThrow", func(t *testing.T) {
		// Test that saving throw rolls work
		roll, total, err := char.RollSavingThrow(shared.AttributeStrength)
		require.NoError(t, err)
		require.NotNil(t, roll)

		// Total should be roll result + bonus (6)
		expectedMin := 1 + 6  // Natural 1 + bonus
		expectedMax := 20 + 6 // Natural 20 + bonus
		assert.GreaterOrEqual(t, total, expectedMin)
		assert.LessOrEqual(t, total, expectedMax)
	})
}

func TestCharacterSkillChecks(t *testing.T) {
	// Create a test character
	char := &Character{
		Level: 5, // +3 proficiency bonus
		Attributes: map[shared.Attribute]*AbilityScore{
			shared.AttributeStrength:  {Score: 16, Bonus: 3},
			shared.AttributeDexterity: {Score: 14, Bonus: 2},
			shared.AttributeWisdom:    {Score: 12, Bonus: 1},
		},
		Proficiencies: map[rulebook.ProficiencyType][]*rulebook.Proficiency{
			rulebook.ProficiencyTypeSkill: {
				{Key: "skill-athletics", Name: "Athletics", Type: rulebook.ProficiencyTypeSkill},
				{Key: "skill-acrobatics", Name: "Acrobatics", Type: rulebook.ProficiencyTypeSkill},
				{Key: "skill-perception", Name: "Perception", Type: rulebook.ProficiencyTypeSkill},
			},
		},
	}

	t.Run("HasSkillProficiency", func(t *testing.T) {
		assert.True(t, char.HasSkillProficiency("skill-athletics"))
		assert.True(t, char.HasSkillProficiency("skill-acrobatics"))
		assert.True(t, char.HasSkillProficiency("skill-perception"))
		assert.False(t, char.HasSkillProficiency("skill-stealth"))
		assert.False(t, char.HasSkillProficiency("skill-intimidation"))
	})

	t.Run("GetSkillBonus", func(t *testing.T) {
		// With proficiency
		assert.Equal(t, 6, char.GetSkillBonus("skill-athletics", shared.AttributeStrength))   // 3 + 3
		assert.Equal(t, 5, char.GetSkillBonus("skill-acrobatics", shared.AttributeDexterity)) // 2 + 3
		assert.Equal(t, 4, char.GetSkillBonus("skill-perception", shared.AttributeWisdom))    // 1 + 3

		// Without proficiency
		assert.Equal(t, 2, char.GetSkillBonus("skill-stealth", shared.AttributeDexterity))     // 2 + 0
		assert.Equal(t, 3, char.GetSkillBonus("skill-intimidation", shared.AttributeStrength)) // 3 + 0
	})

	t.Run("RollSkillCheck", func(t *testing.T) {
		// Test that skill check rolls work
		roll, total, err := char.RollSkillCheck("skill-athletics", shared.AttributeStrength)
		require.NoError(t, err)
		require.NotNil(t, roll)

		// Total should be roll result + bonus (6)
		expectedMin := 1 + 6  // Natural 1 + bonus
		expectedMax := 20 + 6 // Natural 20 + bonus
		assert.GreaterOrEqual(t, total, expectedMin)
		assert.LessOrEqual(t, total, expectedMax)
	})
}

func TestCharacterNilSafety(t *testing.T) {
	// Test with empty character
	char := &Character{Level: 1}

	t.Run("NilProficiencies", func(t *testing.T) {
		assert.False(t, char.HasSavingThrowProficiency(shared.AttributeStrength))
		assert.False(t, char.HasSkillProficiency("skill-athletics"))
		assert.Equal(t, 0, char.GetSavingThrowBonus(shared.AttributeStrength))
		assert.Equal(t, 0, char.GetSkillBonus("skill-athletics", shared.AttributeStrength))
	})

	t.Run("NilAttributes", func(t *testing.T) {
		char.Proficiencies = map[rulebook.ProficiencyType][]*rulebook.Proficiency{
			rulebook.ProficiencyTypeSavingThrow: {
				{Key: "saving-throw-str", Name: "Strength", Type: rulebook.ProficiencyTypeSavingThrow},
			},
		}
		// Should get proficiency bonus but no ability modifier
		assert.Equal(t, 2, char.GetSavingThrowBonus(shared.AttributeStrength)) // 0 + 2
	})
}
