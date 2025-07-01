package character

import (
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/rulebook/dnd5e"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestCharacterProficiencyDuplicationBug reproduces issue #73
func TestCharacterProficiencyDuplicationBug(t *testing.T) {
	t.Run("AddProficiency should not add duplicates", func(t *testing.T) {
		char := &Character{
			Proficiencies: make(map[rulebook.ProficiencyType][]*rulebook.Proficiency),
		}

		// Add a proficiency
		prof := &rulebook.Proficiency{
			Key:  "skill-acrobatics",
			Name: "Acrobatics",
			Type: rulebook.ProficiencyTypeSkill,
		}
		char.AddProficiency(prof)

		// Try to add the same proficiency again
		char.AddProficiency(prof)

		// Should only have one instance
		assert.Len(t, char.Proficiencies[rulebook.ProficiencyTypeSkill], 1, "Should not have duplicate proficiencies")
	})

	t.Run("SetProficiencies should replace proficiencies of same type", func(t *testing.T) {
		char := &Character{
			Proficiencies: make(map[rulebook.ProficiencyType][]*rulebook.Proficiency),
		}

		// Add initial proficiencies
		initial := []*rulebook.Proficiency{
			{Key: "skill-acrobatics", Name: "Acrobatics", Type: rulebook.ProficiencyTypeSkill},
			{Key: "skill-athletics", Name: "Athletics", Type: rulebook.ProficiencyTypeSkill},
		}
		for _, p := range initial {
			char.AddProficiency(p)
		}

		// Now set new proficiencies (simulating a new selection)
		newProfs := []*rulebook.Proficiency{
			{Key: "skill-perception", Name: "Perception", Type: rulebook.ProficiencyTypeSkill},
			{Key: "skill-stealth", Name: "Stealth", Type: rulebook.ProficiencyTypeSkill},
		}
		char.SetProficiencies(rulebook.ProficiencyTypeSkill, newProfs)

		// Should only have the new proficiencies
		assert.Len(t, char.Proficiencies[rulebook.ProficiencyTypeSkill], 2)
		assert.Equal(t, "skill-perception", char.Proficiencies[rulebook.ProficiencyTypeSkill][0].Key)
		assert.Equal(t, "skill-stealth", char.Proficiencies[rulebook.ProficiencyTypeSkill][1].Key)
	})

	t.Run("Mixed proficiency types should be preserved", func(t *testing.T) {
		char := &Character{
			Proficiencies: make(map[rulebook.ProficiencyType][]*rulebook.Proficiency),
		}

		// Add different types of proficiencies
		char.AddProficiency(&rulebook.Proficiency{Key: "armor-light", Name: "Light Armor", Type: rulebook.ProficiencyTypeArmor})
		char.AddProficiency(&rulebook.Proficiency{Key: "weapon-simple", Name: "Simple Weapons", Type: rulebook.ProficiencyTypeWeapon})
		char.AddProficiency(&rulebook.Proficiency{Key: "skill-athletics", Name: "Athletics", Type: rulebook.ProficiencyTypeSkill})

		// Replace only skill proficiencies
		newSkills := []*rulebook.Proficiency{
			{Key: "skill-perception", Name: "Perception", Type: rulebook.ProficiencyTypeSkill},
		}
		char.SetProficiencies(rulebook.ProficiencyTypeSkill, newSkills)

		// Other proficiency types should be preserved
		assert.Len(t, char.Proficiencies[rulebook.ProficiencyTypeArmor], 1)
		assert.Len(t, char.Proficiencies[rulebook.ProficiencyTypeWeapon], 1)
		assert.Len(t, char.Proficiencies[rulebook.ProficiencyTypeSkill], 1)
		assert.Equal(t, "skill-perception", char.Proficiencies[rulebook.ProficiencyTypeSkill][0].Key)
	})
}

// TestProficiencyChoiceCategories tests that we can distinguish between base and chosen proficiencies
func TestProficiencyChoiceCategories(t *testing.T) {
	t.Run("Should distinguish base proficiencies from chosen ones", func(t *testing.T) {
		char := &Character{
			Proficiencies: make(map[rulebook.ProficiencyType][]*rulebook.Proficiency),
		}

		// Base proficiencies from race/class (these should be preserved)
		baseProfs := []*rulebook.Proficiency{
			{Key: "armor-light", Name: "Light Armor", Type: rulebook.ProficiencyTypeArmor},
			{Key: "weapon-simple", Name: "Simple Weapons", Type: rulebook.ProficiencyTypeWeapon},
			{Key: "saving-throw-dex", Name: "Dexterity", Type: rulebook.ProficiencyTypeSavingThrow},
			{Key: "saving-throw-int", Name: "Intelligence", Type: rulebook.ProficiencyTypeSavingThrow},
		}

		// Add base proficiencies
		for _, p := range baseProfs {
			char.AddProficiency(p)
		}

		// Chosen skill proficiencies (these should be replaceable)
		chosenSkills := []*rulebook.Proficiency{
			{Key: "skill-acrobatics", Name: "Acrobatics", Type: rulebook.ProficiencyTypeSkill},
			{Key: "skill-stealth", Name: "Stealth", Type: rulebook.ProficiencyTypeSkill},
		}

		// Set chosen skills
		char.SetProficiencies(rulebook.ProficiencyTypeSkill, chosenSkills)

		// Verify base proficiencies are preserved
		assert.Len(t, char.Proficiencies[rulebook.ProficiencyTypeArmor], 1)
		assert.Len(t, char.Proficiencies[rulebook.ProficiencyTypeWeapon], 1)
		assert.Len(t, char.Proficiencies[rulebook.ProficiencyTypeSavingThrow], 2)

		// Verify chosen skills are set correctly
		assert.Len(t, char.Proficiencies[rulebook.ProficiencyTypeSkill], 2)
	})
}
