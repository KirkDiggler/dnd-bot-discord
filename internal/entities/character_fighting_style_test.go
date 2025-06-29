package entities

import (
	"testing"

	"github.com/KirkDiggler/dnd-bot-discord/internal/entities/damage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFightingStyleArchery(t *testing.T) {
	// Create a fighter with archery fighting style
	char := &Character{
		Name:  "Archer",
		Level: 1,
		Class: &Class{Key: "fighter"},
		Attributes: map[Attribute]*AbilityScore{
			AttributeDexterity: {Score: 16, Bonus: 3},
			AttributeStrength:  {Score: 10, Bonus: 0},
		},
		Features: []*CharacterFeature{
			{
				Key:  "fighting_style",
				Name: "Fighting Style",
				Metadata: map[string]any{
					"style": "archery",
				},
			},
		},
		EquippedSlots: make(map[Slot]Equipment),
	}

	// Equip a ranged weapon
	longbow := &Weapon{
		Base: BasicEquipment{
			Key:  "longbow",
			Name: "Longbow",
		},
		WeaponCategory: "martial",
		WeaponRange:    "Ranged",
		Damage: &damage.Damage{
			DiceCount:  1,
			DiceSize:   8,
			DamageType: damage.TypePiercing,
		},
	}
	char.EquippedSlots[SlotMainHand] = longbow

	// Add martial weapon proficiency
	char.Proficiencies = map[ProficiencyType][]*Proficiency{
		ProficiencyTypeWeapon: {
			{Key: "martial-weapons", Name: "Martial Weapons"},
		},
	}

	// Attack with the bow
	results, err := char.Attack()
	require.NoError(t, err)
	require.Len(t, results, 1)

	// Check attack roll includes bonuses: base roll + DEX(3) + proficiency(2) + archery(2)
	// We can't check exact value due to dice roll, but we can verify the attack happened
	assert.Greater(t, results[0].AttackRoll, 0)
	// Damage should include DEX bonus (3)
	assert.GreaterOrEqual(t, results[0].DamageRoll, 4) // At least 1 damage + 3 DEX
}

func TestFightingStyleDueling(t *testing.T) {
	// Create a fighter with dueling fighting style
	char := &Character{
		Name:  "Duelist",
		Level: 1,
		Class: &Class{Key: "fighter"},
		Attributes: map[Attribute]*AbilityScore{
			AttributeStrength:  {Score: 16, Bonus: 3},
			AttributeDexterity: {Score: 10, Bonus: 0},
		},
		Features: []*CharacterFeature{
			{
				Key:  "fighting_style",
				Name: "Fighting Style",
				Metadata: map[string]any{
					"style": "dueling",
				},
			},
		},
		EquippedSlots: make(map[Slot]Equipment),
	}

	// Equip a one-handed weapon
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
		Properties: []*ReferenceItem{
			{Key: "versatile"},
		},
	}
	char.EquippedSlots[SlotMainHand] = longsword

	// Add martial weapon proficiency
	char.Proficiencies = map[ProficiencyType][]*Proficiency{
		ProficiencyTypeWeapon: {
			{Key: "martial-weapons", Name: "Martial Weapons"},
		},
	}

	// Attack with dueling style (no off-hand weapon)
	results, err := char.Attack()
	require.NoError(t, err)
	require.Len(t, results, 1)

	// Check damage includes dueling bonus: base damage + STR(3) + dueling(2)
	assert.GreaterOrEqual(t, results[0].DamageRoll, 6) // At least 1 damage + 5 bonus

	// Now equip a shield in off-hand (still gets dueling bonus)
	shield := &Armor{
		Base: BasicEquipment{
			Key:  "shield",
			Name: "Shield",
		},
		ArmorClass: &ArmorClass{
			Base:     2,
			DexBonus: false,
		},
		ArmorCategory: "shield",
	}
	char.EquippedSlots[SlotOffHand] = shield

	// Attack should still get dueling bonus with shield
	results2, err := char.Attack()
	require.NoError(t, err)
	require.Len(t, results2, 1)
	assert.GreaterOrEqual(t, results2[0].DamageRoll, 6, "Should still get dueling bonus with shield")

	// Now equip a weapon in off-hand (loses dueling bonus)
	dagger := &Weapon{
		Base: BasicEquipment{
			Key:  "dagger",
			Name: "Dagger",
		},
		WeaponCategory: "simple",
		WeaponRange:    "Melee",
		Damage: &damage.Damage{
			DiceCount:  1,
			DiceSize:   4,
			DamageType: damage.TypePiercing,
		},
		Properties: []*ReferenceItem{
			{Key: "light"},
		},
	}
	char.EquippedSlots[SlotOffHand] = dagger

	// Attack should NOT get dueling bonus with two weapons
	results3, err := char.Attack()
	require.NoError(t, err)
	require.Len(t, results3, 2) // Two attacks now
	// Main hand damage should only have STR(3), no dueling bonus
	// Check it's less than with dueling (no guaranteed way without mocking dice)
	assert.Greater(t, results3[0].DamageRoll, 0, "Should still do damage")
}

func TestFightingStyleTwoWeaponFighting(t *testing.T) {
	// Create a fighter with two-weapon fighting style
	char := &Character{
		Name:  "Dual Wielder",
		Level: 1,
		Class: &Class{Key: "fighter"},
		Attributes: map[Attribute]*AbilityScore{
			AttributeStrength:  {Score: 16, Bonus: 3},
			AttributeDexterity: {Score: 14, Bonus: 2},
		},
		Features: []*CharacterFeature{
			{
				Key:  "fighting_style",
				Name: "Fighting Style",
				Metadata: map[string]any{
					"style": "two_weapon",
				},
			},
		},
		EquippedSlots: make(map[Slot]Equipment),
	}

	// Equip two light weapons
	shortsword := &Weapon{
		Base: BasicEquipment{
			Key:  "shortsword",
			Name: "Shortsword",
		},
		WeaponCategory: "martial",
		WeaponRange:    "Melee",
		Damage: &damage.Damage{
			DiceCount:  1,
			DiceSize:   6,
			DamageType: damage.TypePiercing,
		},
		Properties: []*ReferenceItem{
			{Key: "light"},
		},
	}
	char.EquippedSlots[SlotMainHand] = shortsword
	char.EquippedSlots[SlotOffHand] = shortsword

	// Add martial weapon proficiency
	char.Proficiencies = map[ProficiencyType][]*Proficiency{
		ProficiencyTypeWeapon: {
			{Key: "martial-weapons", Name: "Martial Weapons"},
		},
	}

	// Attack with two weapons
	results, err := char.Attack()
	require.NoError(t, err)
	require.Len(t, results, 2)

	// Main hand: includes STR bonus to damage
	assert.GreaterOrEqual(t, results[0].DamageRoll, 4) // At least 1 + STR(3)

	// Off-hand: WITH ability modifier to damage due to fighting style
	assert.GreaterOrEqual(t, results[1].DamageRoll, 4) // At least 1 + STR(3) from two-weapon fighting
}

func TestFightingStyleDefense(t *testing.T) {
	// Create a fighter with defense fighting style
	char := &Character{
		Name:  "Defender",
		Level: 1,
		Class: &Class{Key: "fighter"},
		Attributes: map[Attribute]*AbilityScore{
			AttributeDexterity:    {Score: 14, Bonus: 2},
			AttributeConstitution: {Score: 14, Bonus: 2},
		},
		Features: []*CharacterFeature{
			{
				Key:  "fighting_style",
				Name: "Fighting Style",
				Metadata: map[string]any{
					"style": "defense",
				},
			},
		},
		EquippedSlots: make(map[Slot]Equipment),
	}

	// Calculate AC without armor (no defense bonus)
	char.calculateAC()
	assert.Equal(t, 10, char.AC) // 10 base, no DEX without armor in 5e by default

	// Equip armor
	chainMail := &Armor{
		Base: BasicEquipment{
			Key:  "chain-mail",
			Name: "Chain Mail",
		},
		ArmorClass: &ArmorClass{
			Base:     16,
			DexBonus: false,
		},
		ArmorCategory: "heavy",
	}
	char.EquippedSlots[SlotBody] = chainMail

	// Calculate AC with armor (gets defense bonus)
	char.calculateAC()
	assert.Equal(t, 17, char.AC) // 16 (chain mail) + 1 (defense)
}

func TestFightingStyleGreatWeapon(t *testing.T) {
	// Create a fighter with great weapon fighting style
	char := &Character{
		Name:  "Great Weapon Fighter",
		Level: 1,
		Class: &Class{Key: "fighter"},
		Attributes: map[Attribute]*AbilityScore{
			AttributeStrength:  {Score: 18, Bonus: 4},
			AttributeDexterity: {Score: 10, Bonus: 0},
		},
		Features: []*CharacterFeature{
			{
				Key:  "fighting_style",
				Name: "Fighting Style",
				Metadata: map[string]any{
					"style": "great_weapon",
				},
			},
		},
		EquippedSlots: make(map[Slot]Equipment),
	}

	// Equip a two-handed weapon
	greatsword := &Weapon{
		Base: BasicEquipment{
			Key:  "greatsword",
			Name: "Greatsword",
		},
		WeaponCategory: "martial",
		WeaponRange:    "Melee",
		Damage: &damage.Damage{
			DiceCount:  2,
			DiceSize:   6,
			DamageType: damage.TypeSlashing,
		},
		Properties: []*ReferenceItem{
			{Key: "two-handed"},
		},
	}
	char.EquippedSlots[SlotTwoHanded] = greatsword

	// Add martial weapon proficiency
	char.Proficiencies = map[ProficiencyType][]*Proficiency{
		ProficiencyTypeWeapon: {
			{Key: "martial-weapons", Name: "Martial Weapons"},
		},
	}

	// Attack with the greatsword
	results, err := char.Attack()
	require.NoError(t, err)
	require.Len(t, results, 1)

	// Check damage includes STR bonus
	assert.GreaterOrEqual(t, results[0].DamageRoll, 6) // At least 2 damage + STR(4)

	// Great Weapon Fighting allows rerolling 1s and 2s on damage dice
	// This is handled in the damage rolling logic, not as a flat bonus
}
