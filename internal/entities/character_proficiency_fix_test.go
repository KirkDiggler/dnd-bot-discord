package entities

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestCharacterProficiencyDuplicationBug reproduces issue #73
func TestCharacterProficiencyDuplicationBug(t *testing.T) {
	t.Run("AddProficiency should not add duplicates", func(t *testing.T) {
		char := &Character{
			Proficiencies: make(map[ProficiencyType][]*Proficiency),
		}

		// Add a proficiency
		prof := &Proficiency{
			Key:  "skill-acrobatics",
			Name: "Acrobatics",
			Type: ProficiencyTypeSkill,
		}
		char.AddProficiency(prof)

		// Try to add the same proficiency again
		char.AddProficiency(prof)

		// Should only have one instance
		assert.Len(t, char.Proficiencies[ProficiencyTypeSkill], 1, "Should not have duplicate proficiencies")
	})

	t.Run("SetProficiencies should replace proficiencies of same type", func(t *testing.T) {
		char := &Character{
			Proficiencies: make(map[ProficiencyType][]*Proficiency),
		}

		// Add initial proficiencies
		initial := []*Proficiency{
			{Key: "skill-acrobatics", Name: "Acrobatics", Type: ProficiencyTypeSkill},
			{Key: "skill-athletics", Name: "Athletics", Type: ProficiencyTypeSkill},
		}
		for _, p := range initial {
			char.AddProficiency(p)
		}

		// Now set new proficiencies (simulating a new selection)
		newProfs := []*Proficiency{
			{Key: "skill-perception", Name: "Perception", Type: ProficiencyTypeSkill},
			{Key: "skill-stealth", Name: "Stealth", Type: ProficiencyTypeSkill},
		}
		char.SetProficiencies(ProficiencyTypeSkill, newProfs)

		// Should only have the new proficiencies
		assert.Len(t, char.Proficiencies[ProficiencyTypeSkill], 2)
		assert.Equal(t, "skill-perception", char.Proficiencies[ProficiencyTypeSkill][0].Key)
		assert.Equal(t, "skill-stealth", char.Proficiencies[ProficiencyTypeSkill][1].Key)
	})

	t.Run("Mixed proficiency types should be preserved", func(t *testing.T) {
		char := &Character{
			Proficiencies: make(map[ProficiencyType][]*Proficiency),
		}

		// Add different types of proficiencies
		char.AddProficiency(&Proficiency{Key: "armor-light", Name: "Light Armor", Type: ProficiencyTypeArmor})
		char.AddProficiency(&Proficiency{Key: "weapon-simple", Name: "Simple Weapons", Type: ProficiencyTypeWeapon})
		char.AddProficiency(&Proficiency{Key: "skill-athletics", Name: "Athletics", Type: ProficiencyTypeSkill})

		// Replace only skill proficiencies
		newSkills := []*Proficiency{
			{Key: "skill-perception", Name: "Perception", Type: ProficiencyTypeSkill},
		}
		char.SetProficiencies(ProficiencyTypeSkill, newSkills)

		// Other proficiency types should be preserved
		assert.Len(t, char.Proficiencies[ProficiencyTypeArmor], 1)
		assert.Len(t, char.Proficiencies[ProficiencyTypeWeapon], 1)
		assert.Len(t, char.Proficiencies[ProficiencyTypeSkill], 1)
		assert.Equal(t, "skill-perception", char.Proficiencies[ProficiencyTypeSkill][0].Key)
	})
}

// TestProficiencyChoiceCategories tests that we can distinguish between base and chosen proficiencies
func TestProficiencyChoiceCategories(t *testing.T) {
	t.Run("Should distinguish base proficiencies from chosen ones", func(t *testing.T) {
		char := &Character{
			Proficiencies: make(map[ProficiencyType][]*Proficiency),
		}

		// Base proficiencies from race/class (these should be preserved)
		baseProfs := []*Proficiency{
			{Key: "armor-light", Name: "Light Armor", Type: ProficiencyTypeArmor},
			{Key: "weapon-simple", Name: "Simple Weapons", Type: ProficiencyTypeWeapon},
			{Key: "saving-throw-dex", Name: "Dexterity", Type: ProficiencyTypeSavingThrow},
			{Key: "saving-throw-int", Name: "Intelligence", Type: ProficiencyTypeSavingThrow},
		}

		// Add base proficiencies
		for _, p := range baseProfs {
			char.AddProficiency(p)
		}

		// Chosen skill proficiencies (these should be replaceable)
		chosenSkills := []*Proficiency{
			{Key: "skill-acrobatics", Name: "Acrobatics", Type: ProficiencyTypeSkill},
			{Key: "skill-stealth", Name: "Stealth", Type: ProficiencyTypeSkill},
		}

		// Set chosen skills
		char.SetProficiencies(ProficiencyTypeSkill, chosenSkills)

		// Verify base proficiencies are preserved
		assert.Len(t, char.Proficiencies[ProficiencyTypeArmor], 1)
		assert.Len(t, char.Proficiencies[ProficiencyTypeWeapon], 1)
		assert.Len(t, char.Proficiencies[ProficiencyTypeSavingThrow], 2)

		// Verify chosen skills are set correctly
		assert.Len(t, char.Proficiencies[ProficiencyTypeSkill], 2)
	})
}
