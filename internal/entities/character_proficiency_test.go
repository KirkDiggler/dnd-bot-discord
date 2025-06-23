package entities

import (
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
		Attributes: map[Attribute]*AbilityScore{
			AttributeStrength:     {Score: 16, Bonus: 3},
			AttributeDexterity:    {Score: 14, Bonus: 2},
			AttributeConstitution: {Score: 13, Bonus: 1},
			AttributeIntelligence: {Score: 10, Bonus: 0},
			AttributeWisdom:       {Score: 12, Bonus: 1},
			AttributeCharisma:     {Score: 8, Bonus: -1},
		},
		Proficiencies: map[ProficiencyType][]*Proficiency{
			ProficiencyTypeSavingThrow: {
				{Key: "saving-throw-str", Name: "Strength", Type: ProficiencyTypeSavingThrow},
				{Key: "saving-throw-con", Name: "Constitution", Type: ProficiencyTypeSavingThrow},
			},
		},
	}

	t.Run("HasSavingThrowProficiency", func(t *testing.T) {
		assert.True(t, char.HasSavingThrowProficiency(AttributeStrength))
		assert.True(t, char.HasSavingThrowProficiency(AttributeConstitution))
		assert.False(t, char.HasSavingThrowProficiency(AttributeDexterity))
		assert.False(t, char.HasSavingThrowProficiency(AttributeWisdom))
	})

	t.Run("GetSavingThrowBonus", func(t *testing.T) {
		// With proficiency: ability mod + prof bonus
		assert.Equal(t, 6, char.GetSavingThrowBonus(AttributeStrength))     // 3 + 3
		assert.Equal(t, 4, char.GetSavingThrowBonus(AttributeConstitution)) // 1 + 3

		// Without proficiency: just ability mod
		assert.Equal(t, 2, char.GetSavingThrowBonus(AttributeDexterity))    // 2 + 0
		assert.Equal(t, 1, char.GetSavingThrowBonus(AttributeWisdom))       // 1 + 0
		assert.Equal(t, -1, char.GetSavingThrowBonus(AttributeCharisma))    // -1 + 0
	})

	t.Run("RollSavingThrow", func(t *testing.T) {
		// Test that saving throw rolls work
		roll, total, err := char.RollSavingThrow(AttributeStrength)
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
		Attributes: map[Attribute]*AbilityScore{
			AttributeStrength:     {Score: 16, Bonus: 3},
			AttributeDexterity:    {Score: 14, Bonus: 2},
			AttributeWisdom:       {Score: 12, Bonus: 1},
		},
		Proficiencies: map[ProficiencyType][]*Proficiency{
			ProficiencyTypeSkill: {
				{Key: "skill-athletics", Name: "Athletics", Type: ProficiencyTypeSkill},
				{Key: "skill-acrobatics", Name: "Acrobatics", Type: ProficiencyTypeSkill},
				{Key: "skill-perception", Name: "Perception", Type: ProficiencyTypeSkill},
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
		assert.Equal(t, 6, char.GetSkillBonus("skill-athletics", AttributeStrength))   // 3 + 3
		assert.Equal(t, 5, char.GetSkillBonus("skill-acrobatics", AttributeDexterity)) // 2 + 3
		assert.Equal(t, 4, char.GetSkillBonus("skill-perception", AttributeWisdom))    // 1 + 3

		// Without proficiency
		assert.Equal(t, 2, char.GetSkillBonus("skill-stealth", AttributeDexterity))     // 2 + 0
		assert.Equal(t, 3, char.GetSkillBonus("skill-intimidation", AttributeStrength)) // 3 + 0
	})

	t.Run("RollSkillCheck", func(t *testing.T) {
		// Test that skill check rolls work
		roll, total, err := char.RollSkillCheck("skill-athletics", AttributeStrength)
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
		assert.False(t, char.HasSavingThrowProficiency(AttributeStrength))
		assert.False(t, char.HasSkillProficiency("skill-athletics"))
		assert.Equal(t, 0, char.GetSavingThrowBonus(AttributeStrength))
		assert.Equal(t, 0, char.GetSkillBonus("skill-athletics", AttributeStrength))
	})

	t.Run("NilAttributes", func(t *testing.T) {
		char.Proficiencies = map[ProficiencyType][]*Proficiency{
			ProficiencyTypeSavingThrow: {
				{Key: "saving-throw-str", Name: "Strength", Type: ProficiencyTypeSavingThrow},
			},
		}
		// Should get proficiency bonus but no ability modifier
		assert.Equal(t, 2, char.GetSavingThrowBonus(AttributeStrength)) // 0 + 2
	})
}