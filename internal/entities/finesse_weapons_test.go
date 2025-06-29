package entities

import (
	"testing"

	mockdice "github.com/KirkDiggler/dnd-bot-discord/internal/dice/mock"
	"github.com/KirkDiggler/dnd-bot-discord/internal/entities/damage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWeapon_IsFinesse(t *testing.T) {
	tests := []struct {
		name           string
		weapon         *Weapon
		expectedResult bool
	}{
		{
			name: "Rapier is a finesse weapon",
			weapon: &Weapon{
				Base:       BasicEquipment{Key: "rapier", Name: "Rapier"},
				Properties: []*ReferenceItem{{Key: "finesse"}},
			},
			expectedResult: true,
		},
		{
			name: "Dagger is a finesse weapon",
			weapon: &Weapon{
				Base:       BasicEquipment{Key: "dagger", Name: "Dagger"},
				Properties: []*ReferenceItem{{Key: "finesse"}, {Key: "light"}, {Key: "thrown"}},
			},
			expectedResult: true,
		},
		{
			name: "Longsword is not a finesse weapon",
			weapon: &Weapon{
				Base:       BasicEquipment{Key: "longsword", Name: "Longsword"},
				Properties: []*ReferenceItem{{Key: "versatile"}},
			},
			expectedResult: false,
		},
		{
			name: "Weapon with no properties",
			weapon: &Weapon{
				Base:       BasicEquipment{Key: "club", Name: "Club"},
				Properties: []*ReferenceItem{},
			},
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.weapon.IsFinesse()
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}

func TestFinesseWeapon_UseDexForAttack(t *testing.T) {
	tests := []struct {
		name           string
		weapon         *Weapon
		features       []*CharacterFeature
		strScore       int
		dexScore       int
		expectedAttack int // Expected attack bonus (not including d20)
		expectedDamage int // Expected damage bonus
		description    string
	}{
		{
			name: "Rapier uses DEX when higher than STR",
			weapon: &Weapon{
				Base:           BasicEquipment{Key: "rapier", Name: "Rapier"},
				WeaponCategory: "Martial",
				WeaponRange:    "Melee",
				Damage:         &damage.Damage{DiceCount: 1, DiceSize: 8, DamageType: damage.TypePiercing},
				Properties:     []*ReferenceItem{{Key: "finesse"}},
			},
			features:       []*CharacterFeature{}, // No special features needed
			strScore:       10,                    // +0 bonus
			dexScore:       16,                    // +3 bonus
			expectedAttack: 5,                     // DEX(3) + proficiency(2)
			expectedDamage: 3,                     // DEX(3)
			description:    "Finesse weapons should use DEX when it's higher",
		},
		{
			name: "Rapier uses STR when higher than DEX",
			weapon: &Weapon{
				Base:           BasicEquipment{Key: "rapier", Name: "Rapier"},
				WeaponCategory: "Martial",
				WeaponRange:    "Melee",
				Damage:         &damage.Damage{DiceCount: 1, DiceSize: 8, DamageType: damage.TypePiercing},
				Properties:     []*ReferenceItem{{Key: "finesse"}},
			},
			features:       []*CharacterFeature{},
			strScore:       18, // +4 bonus
			dexScore:       14, // +2 bonus
			expectedAttack: 6,  // STR(4) + proficiency(2)
			expectedDamage: 4,  // STR(4)
			description:    "Finesse weapons use the higher of STR or DEX",
		},
		{
			name: "Non-finesse weapon must use STR",
			weapon: &Weapon{
				Base:           BasicEquipment{Key: "longsword", Name: "Longsword"},
				WeaponCategory: "Martial",
				WeaponRange:    "Melee",
				Damage:         &damage.Damage{DiceCount: 1, DiceSize: 8, DamageType: damage.TypeSlashing},
				Properties:     []*ReferenceItem{{Key: "versatile"}},
			},
			features:       []*CharacterFeature{},
			strScore:       10, // +0 bonus
			dexScore:       18, // +4 bonus
			expectedAttack: 2,  // STR(0) + proficiency(2)
			expectedDamage: 0,  // STR(0)
			description:    "Non-finesse weapons always use STR",
		},
		{
			name: "Monk with rapier uses DEX (both finesse and martial arts apply)",
			weapon: &Weapon{
				Base:           BasicEquipment{Key: "rapier", Name: "Rapier"},
				WeaponCategory: "Martial",
				WeaponRange:    "Melee",
				Damage:         &damage.Damage{DiceCount: 1, DiceSize: 8, DamageType: damage.TypePiercing},
				Properties:     []*ReferenceItem{{Key: "finesse"}},
			},
			features:       []*CharacterFeature{{Key: "martial-arts", Name: "Martial Arts"}},
			strScore:       10, // +0 bonus
			dexScore:       16, // +3 bonus
			expectedAttack: 5,  // DEX(3) + proficiency(2)
			expectedDamage: 3,  // DEX(3)
			description:    "Monks can use finesse weapons with DEX",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create character
			char := &Character{
				Level:    1,
				Features: tt.features,
				Attributes: map[Attribute]*AbilityScore{
					AttributeStrength: {
						Score: tt.strScore,
						Bonus: (tt.strScore - 10) / 2,
					},
					AttributeDexterity: {
						Score: tt.dexScore,
						Bonus: (tt.dexScore - 10) / 2,
					},
				},
				EquippedSlots: map[Slot]Equipment{
					SlotMainHand: tt.weapon,
				},
				// Give proficiency with all weapons for testing
				Proficiencies: map[ProficiencyType][]*Proficiency{
					ProficiencyTypeWeapon: {
						{Key: tt.weapon.GetKey(), Name: tt.weapon.GetName()},
						{Key: "martial-weapons", Name: "Martial Weapons"},
					},
				},
			}

			// Mock dice roller
			mockRoller := mockdice.NewManualMockRoller()
			mockRoller.SetRolls([]int{
				15,                            // Attack roll
				tt.weapon.Damage.DiceSize / 2, // Damage roll
			})
			char = char.WithDiceRoller(mockRoller)

			// Perform attack
			results, err := char.Attack()
			require.NoError(t, err)
			require.Len(t, results, 1, "Should have one attack result")

			result := results[0]
			// Verify attack and damage calculations
			assert.Equal(t, 15+tt.expectedAttack, result.AttackRoll, "Attack roll: "+tt.description)
			assert.Equal(t, tt.weapon.Damage.DiceSize/2+tt.expectedDamage, result.DamageRoll, "Damage roll: "+tt.description)
		})
	}
}

func TestFinesseWeapon_DualWielding(t *testing.T) {
	// Create a character with two finesse weapons
	character := &Character{
		Name:  "Dual Wielder",
		Level: 1,
		Attributes: map[Attribute]*AbilityScore{
			AttributeStrength:  {Score: 12, Bonus: 1}, // +1
			AttributeDexterity: {Score: 16, Bonus: 3}, // +3
		},
		Features: []*CharacterFeature{}, // No special features
		EquippedSlots: map[Slot]Equipment{
			SlotMainHand: &Weapon{
				Base:           BasicEquipment{Key: "rapier", Name: "Rapier"},
				WeaponCategory: "Martial",
				WeaponRange:    "Melee",
				Damage:         &damage.Damage{DiceCount: 1, DiceSize: 8, DamageType: damage.TypePiercing},
				Properties:     []*ReferenceItem{{Key: "finesse"}},
			},
			SlotOffHand: &Weapon{
				Base:           BasicEquipment{Key: "dagger", Name: "Dagger"},
				WeaponCategory: "Simple",
				WeaponRange:    "Melee",
				Damage:         &damage.Damage{DiceCount: 1, DiceSize: 4, DamageType: damage.TypePiercing},
				Properties:     []*ReferenceItem{{Key: "finesse"}, {Key: "light"}},
			},
		},
		Proficiencies: map[ProficiencyType][]*Proficiency{
			ProficiencyTypeWeapon: {
				{Key: "rapier", Name: "Rapier"},
				{Key: "dagger", Name: "Dagger"},
				{Key: "simple-weapons", Name: "Simple Weapons"},
				{Key: "martial-weapons", Name: "Martial Weapons"},
			},
		},
	}

	// Mock dice roller
	mockRoller := mockdice.NewManualMockRoller()
	mockRoller.SetRolls([]int{
		12, // Main hand attack roll
		5,  // Main hand damage roll (1d8)
		15, // Off-hand attack roll
		2,  // Off-hand damage roll (1d4)
	})
	character = character.WithDiceRoller(mockRoller)

	results, err := character.Attack()
	require.NoError(t, err)
	require.Len(t, results, 2, "Should have two attack results for dual wielding")

	// Main hand rapier: d20(12) + DEX(3) + prof(2) = 17
	assert.Equal(t, 17, results[0].AttackRoll, "Main hand finesse weapon should use DEX")
	assert.Equal(t, 8, results[0].DamageRoll, "Main hand damage should be 1d8(5) + DEX(3)")

	// Off-hand dagger: d20(15) + DEX(3) + prof(2) = 20
	assert.Equal(t, 20, results[1].AttackRoll, "Off-hand finesse weapon should use DEX")
	assert.Equal(t, 5, results[1].DamageRoll, "Off-hand damage should be 1d4(2) + DEX(3)")
}
