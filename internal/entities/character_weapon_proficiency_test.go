package entities

import (
	"testing"

	"github.com/KirkDiggler/dnd-bot-discord/internal/entities/damage"
	"github.com/stretchr/testify/assert"
)

func TestCharacter_WeaponCategoryProficiency(t *testing.T) {
	tests := []struct {
		name           string
		proficiencies  []*Proficiency
		weaponCategory string
		expected       bool
	}{
		{
			name: "Has simple weapon proficiency",
			proficiencies: []*Proficiency{
				{Key: "simple-weapons", Name: "Simple Weapons"},
			},
			weaponCategory: "simple",
			expected:       true,
		},
		{
			name: "Has martial weapon proficiency",
			proficiencies: []*Proficiency{
				{Key: "martial-weapons", Name: "Martial Weapons"},
			},
			weaponCategory: "martial",
			expected:       true,
		},
		{
			name: "No martial proficiency when only simple",
			proficiencies: []*Proficiency{
				{Key: "simple-weapons", Name: "Simple Weapons"},
			},
			weaponCategory: "martial",
			expected:       false,
		},
		{
			name:           "No proficiencies",
			proficiencies:  []*Proficiency{},
			weaponCategory: "simple",
			expected:       false,
		},
		{
			name: "Unknown weapon category",
			proficiencies: []*Proficiency{
				{Key: "simple-weapons", Name: "Simple Weapons"},
				{Key: "martial-weapons", Name: "Martial Weapons"},
			},
			weaponCategory: "exotic",
			expected:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			char := &Character{
				Proficiencies: map[ProficiencyType][]*Proficiency{
					ProficiencyTypeWeapon: tt.proficiencies,
				},
			}

			result := char.hasWeaponCategoryProficiency(tt.weaponCategory)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCharacter_Attack_WithCategoryProficiency(t *testing.T) {
	// Create a ranger with simple and martial weapon proficiencies
	ranger := &Character{
		ID:    "test-ranger",
		Name:  "Test Ranger",
		Class: &Class{Key: "ranger"},
		Level: 1,
		Attributes: map[Attribute]*AbilityScore{
			AttributeStrength:  {Score: 14, Bonus: 2},
			AttributeDexterity: {Score: 16, Bonus: 3},
		},
		Proficiencies: map[ProficiencyType][]*Proficiency{
			ProficiencyTypeWeapon: {
				{Key: "simple-weapons", Name: "Simple Weapons"},
				{Key: "martial-weapons", Name: "Martial Weapons"},
			},
		},
		EquippedSlots: make(map[Slot]Equipment),
		Resources:     &CharacterResources{},
	}

	// Equip a martial weapon (longsword)
	longsword := &Weapon{
		Base: BasicEquipment{
			Key:  "longsword",
			Name: "Longsword",
		},
		WeaponCategory: "martial",
		WeaponRange:    "Melee",
		Damage: &damage.Damage{
			DiceCount:  1,
			DiceSize:   8,
			DamageType: damage.TypeSlashing,
		},
	}
	ranger.EquippedSlots[SlotMainHand] = longsword

	// Attack should include proficiency bonus
	attacks, err := ranger.Attack()
	assert.NoError(t, err)
	assert.Len(t, attacks, 1)

	// Attack bonus should be +4 (2 STR + 2 proficiency)
	// We can't test the exact roll, but we can verify the attack was made
	assert.NotNil(t, attacks[0])
}

func TestCharacter_Attack_WithoutCategoryProficiency(t *testing.T) {
	// Create a wizard with only simple weapon proficiency
	wizard := &Character{
		ID:    "test-wizard",
		Name:  "Test Wizard",
		Class: &Class{Key: "wizard"},
		Level: 1,
		Attributes: map[Attribute]*AbilityScore{
			AttributeStrength:     {Score: 8, Bonus: -1},
			AttributeDexterity:    {Score: 14, Bonus: 2},
			AttributeIntelligence: {Score: 16, Bonus: 3},
		},
		Proficiencies: map[ProficiencyType][]*Proficiency{
			ProficiencyTypeWeapon: {
				{Key: "dagger", Name: "Dagger"},
				{Key: "dart", Name: "Dart"},
				{Key: "sling", Name: "Sling"},
				{Key: "quarterstaff", Name: "Quarterstaff"},
				{Key: "light-crossbow", Name: "Light Crossbow"},
			},
		},
		EquippedSlots: make(map[Slot]Equipment),
		Resources:     &CharacterResources{},
	}

	// Equip a martial weapon (longsword) - wizard is NOT proficient
	longsword := &Weapon{
		Base: BasicEquipment{
			Key:  "longsword",
			Name: "Longsword",
		},
		WeaponCategory: "martial",
		WeaponRange:    "Melee",
		Damage: &damage.Damage{
			DiceCount:  1,
			DiceSize:   8,
			DamageType: damage.TypeSlashing,
		},
	}
	wizard.EquippedSlots[SlotMainHand] = longsword

	// Attack should NOT include proficiency bonus
	attacks, err := wizard.Attack()
	assert.NoError(t, err)
	assert.Len(t, attacks, 1)

	// Attack bonus should be -1 (just STR modifier, no proficiency)
	// We can't test the exact roll, but we can verify the attack was made
	assert.NotNil(t, attacks[0])
}
