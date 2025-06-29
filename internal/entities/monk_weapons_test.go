package entities

import (
	"testing"

	mockdice "github.com/KirkDiggler/dnd-bot-discord/internal/dice/mock"
	"github.com/KirkDiggler/dnd-bot-discord/internal/entities/damage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWeapon_IsMonkWeapon(t *testing.T) {
	tests := []struct {
		name           string
		weapon         *Weapon
		expectedResult bool
		description    string
	}{
		{
			name: "Shortsword is always a monk weapon",
			weapon: &Weapon{
				Base:           BasicEquipment{Key: "shortsword", Name: "Shortsword"},
				WeaponCategory: "Simple",
				WeaponRange:    "Melee",
				Properties:     []*ReferenceItem{},
			},
			expectedResult: true,
			description:    "Shortswords are specifically listed as monk weapons",
		},
		{
			name: "Simple melee weapon without properties",
			weapon: &Weapon{
				Base:           BasicEquipment{Key: "club", Name: "Club"},
				WeaponCategory: "Simple",
				WeaponRange:    "Melee",
				Properties:     []*ReferenceItem{},
			},
			expectedResult: true,
			description:    "Simple melee weapons without two-handed or heavy are monk weapons",
		},
		{
			name: "Simple melee weapon with light property",
			weapon: &Weapon{
				Base:           BasicEquipment{Key: "dagger", Name: "Dagger"},
				WeaponCategory: "Simple",
				WeaponRange:    "Melee",
				Properties:     []*ReferenceItem{{Key: "light"}},
			},
			expectedResult: true,
			description:    "Light property doesn't prevent monk weapon status",
		},
		{
			name: "Simple melee weapon with two-handed property",
			weapon: &Weapon{
				Base:           BasicEquipment{Key: "quarterstaff", Name: "Quarterstaff"},
				WeaponCategory: "Simple",
				WeaponRange:    "Melee",
				Properties:     []*ReferenceItem{{Key: "two-handed"}},
			},
			expectedResult: false,
			description:    "Two-handed weapons are not monk weapons",
		},
		{
			name: "Simple melee weapon with heavy property",
			weapon: &Weapon{
				Base:           BasicEquipment{Key: "greatclub", Name: "Greatclub"},
				WeaponCategory: "Simple",
				WeaponRange:    "Melee",
				Properties:     []*ReferenceItem{{Key: "heavy"}},
			},
			expectedResult: false,
			description:    "Heavy weapons are not monk weapons",
		},
		{
			name: "Martial melee weapon",
			weapon: &Weapon{
				Base:           BasicEquipment{Key: "longsword", Name: "Longsword"},
				WeaponCategory: "Martial",
				WeaponRange:    "Melee",
				Properties:     []*ReferenceItem{},
			},
			expectedResult: false,
			description:    "Martial weapons are not monk weapons (except shortswords)",
		},
		{
			name: "Simple ranged weapon",
			weapon: &Weapon{
				Base:           BasicEquipment{Key: "shortbow", Name: "Shortbow"},
				WeaponCategory: "Simple",
				WeaponRange:    "Ranged",
				Properties:     []*ReferenceItem{},
			},
			expectedResult: false,
			description:    "Ranged weapons are not monk weapons",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.weapon.IsMonkWeapon()
			assert.Equal(t, tt.expectedResult, result, tt.description)
		})
	}
}

func TestMonkMartialArts_WeaponAttacks(t *testing.T) {
	tests := []struct {
		name           string
		weapon         *Weapon
		features       []*CharacterFeature
		strScore       int
		dexScore       int
		level          int
		expectedAttack int // Expected attack bonus (not including d20)
		expectedDamage int // Expected damage bonus (not including weapon damage dice)
		description    string
	}{
		{
			name: "Monk with shortsword uses DEX",
			weapon: &Weapon{
				Base:           BasicEquipment{Key: "shortsword", Name: "Shortsword"},
				WeaponCategory: "Simple",
				WeaponRange:    "Melee",
				Damage:         &damage.Damage{DiceCount: 1, DiceSize: 6, DamageType: damage.TypeSlashing},
				Properties:     []*ReferenceItem{},
			},
			features:       []*CharacterFeature{{Key: "martial-arts", Name: "Martial Arts"}},
			strScore:       10, // +0 bonus
			dexScore:       16, // +3 bonus
			level:          1,
			expectedAttack: 5, // DEX(3) + proficiency(2)
			expectedDamage: 3, // DEX(3)
			description:    "Monks should use DEX for shortswords",
		},
		{
			name: "Monk with club uses higher of STR/DEX",
			weapon: &Weapon{
				Base:           BasicEquipment{Key: "club", Name: "Club"},
				WeaponCategory: "Simple",
				WeaponRange:    "Melee",
				Damage:         &damage.Damage{DiceCount: 1, DiceSize: 4, DamageType: damage.TypeBludgeoning},
				Properties:     []*ReferenceItem{},
			},
			features:       []*CharacterFeature{{Key: "martial-arts", Name: "Martial Arts"}},
			strScore:       18, // +4 bonus
			dexScore:       14, // +2 bonus
			level:          1,
			expectedAttack: 6, // STR(4) + proficiency(2)
			expectedDamage: 4, // STR(4)
			description:    "Monks can still use STR if it's higher",
		},
		{
			name: "Non-monk with shortsword must use STR",
			weapon: &Weapon{
				Base:           BasicEquipment{Key: "shortsword", Name: "Shortsword"},
				WeaponCategory: "Simple",
				WeaponRange:    "Melee",
				Damage:         &damage.Damage{DiceCount: 1, DiceSize: 6, DamageType: damage.TypeSlashing},
				Properties:     []*ReferenceItem{},
			},
			features:       []*CharacterFeature{}, // No martial arts
			strScore:       10,                    // +0 bonus
			dexScore:       18,                    // +4 bonus
			level:          1,
			expectedAttack: 2, // STR(0) + proficiency(2)
			expectedDamage: 0, // STR(0)
			description:    "Non-monks must use STR for melee weapons",
		},
		{
			name: "Monk with non-monk weapon (longsword) uses STR",
			weapon: &Weapon{
				Base:           BasicEquipment{Key: "longsword", Name: "Longsword"},
				WeaponCategory: "Martial",
				WeaponRange:    "Melee",
				Damage:         &damage.Damage{DiceCount: 1, DiceSize: 8, DamageType: damage.TypeSlashing},
				Properties:     []*ReferenceItem{},
			},
			features:       []*CharacterFeature{{Key: "martial-arts", Name: "Martial Arts"}},
			strScore:       12, // +1 bonus
			dexScore:       16, // +3 bonus
			level:          1,
			expectedAttack: 3, // STR(1) + proficiency(2)
			expectedDamage: 1, // STR(1)
			description:    "Monks can't use DEX with non-monk weapons",
		},
		{
			name: "Level 5 monk gets higher proficiency",
			weapon: &Weapon{
				Base:           BasicEquipment{Key: "shortsword", Name: "Shortsword"},
				WeaponCategory: "Simple",
				WeaponRange:    "Melee",
				Damage:         &damage.Damage{DiceCount: 1, DiceSize: 6, DamageType: damage.TypeSlashing},
				Properties:     []*ReferenceItem{},
			},
			features:       []*CharacterFeature{{Key: "martial-arts", Name: "Martial Arts"}},
			strScore:       10, // +0 bonus
			dexScore:       16, // +3 bonus
			level:          5,
			expectedAttack: 6, // DEX(3) + proficiency(3)
			expectedDamage: 3, // DEX(3)
			description:    "Proficiency bonus increases with level",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create character with specified attributes
			char := &Character{
				Level:    tt.level,
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
						{Key: "simple-weapons", Name: "Simple Weapons"},
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
			assert.Equal(t, 15+tt.expectedAttack, result.AttackRoll, "Attack roll should be d20 + ability + proficiency")
			assert.Equal(t, tt.weapon.Damage.DiceSize/2+tt.expectedDamage, result.DamageRoll, "Damage roll should be weapon dice + ability bonus")
		})
	}
}

func TestMonkMartialArts_DualWielding(t *testing.T) {
	// Create a level 1 monk with two shortswords
	monk := &Character{
		Level: 1,
		Features: []*CharacterFeature{
			{Key: "martial-arts", Name: "Martial Arts"},
		},
		Attributes: map[Attribute]*AbilityScore{
			AttributeStrength: {
				Score: 10, // +0 bonus
				Bonus: 0,
			},
			AttributeDexterity: {
				Score: 16, // +3 bonus
				Bonus: 3,
			},
		},
		EquippedSlots: map[Slot]Equipment{
			SlotMainHand: &Weapon{
				Base:           BasicEquipment{Key: "shortsword", Name: "Shortsword"},
				WeaponCategory: "Simple",
				WeaponRange:    "Melee",
				Damage:         &damage.Damage{DiceCount: 1, DiceSize: 6, DamageType: damage.TypeSlashing},
				Properties:     []*ReferenceItem{{Key: "light"}},
			},
			SlotOffHand: &Weapon{
				Base:           BasicEquipment{Key: "shortsword", Name: "Shortsword"},
				WeaponCategory: "Simple",
				WeaponRange:    "Melee",
				Damage:         &damage.Damage{DiceCount: 1, DiceSize: 6, DamageType: damage.TypeSlashing},
				Properties:     []*ReferenceItem{{Key: "light"}},
			},
		},
		Proficiencies: map[ProficiencyType][]*Proficiency{
			ProficiencyTypeWeapon: {
				{Key: "shortsword", Name: "Shortsword"},
				{Key: "simple-weapons", Name: "Simple Weapons"},
			},
		},
	}

	// Mock dice roller
	mockRoller := mockdice.NewManualMockRoller()
	mockRoller.SetRolls([]int{
		12, // Main hand attack roll
		4,  // Main hand damage roll (1d6)
		8,  // Off-hand attack roll
		2,  // Off-hand damage roll (1d6)
	})
	monk = monk.WithDiceRoller(mockRoller)

	results, err := monk.Attack()
	require.NoError(t, err)
	require.Len(t, results, 2, "Should have two attack results for dual wielding")

	// Main hand: d20(12) + DEX(3) + prof(2) = 17
	assert.Equal(t, 17, results[0].AttackRoll, "Main hand should use DEX for monk weapon")
	assert.Equal(t, 7, results[0].DamageRoll, "Main hand damage should be 1d6(4) + DEX(3)")

	// Off-hand: d20(8) + DEX(3) + prof(2) = 13
	assert.Equal(t, 13, results[1].AttackRoll, "Off-hand should use DEX for monk weapon")
	assert.Equal(t, 5, results[1].DamageRoll, "Off-hand damage should be 1d6(2) + DEX(3)")
}
