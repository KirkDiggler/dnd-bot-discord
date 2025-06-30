package entities

import (
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/damage"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/equipment"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/rulebook"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/shared"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFightingStyleArchery(t *testing.T) {
	// Create a fighter with archery fighting style
	char := &character.Character{
		Name:  "Archer",
		Level: 1,
		Class: &rulebook.Class{Key: "fighter"},
		Attributes: map[shared.Attribute]*character.AbilityScore{
			shared.AttributeDexterity: {Score: 16, Bonus: 3},
			shared.AttributeStrength:  {Score: 10, Bonus: 0},
		},
		Features: []*rulebook.CharacterFeature{
			{
				Key:  "fighting_style",
				Name: "Fighting Style",
				Metadata: map[string]any{
					"style": "archery",
				},
			},
		},
		EquippedSlots: make(map[shared.Slot]equipment.Equipment),
	}

	// Equip a ranged weapon
	longbow := &equipment.Weapon{
		Base: equipment.BasicEquipment{
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
	char.EquippedSlots[shared.SlotMainHand] = longbow

	// Add martial weapon proficiency
	char.Proficiencies = map[rulebook.ProficiencyType][]*rulebook.Proficiency{
		rulebook.ProficiencyTypeWeapon: {
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
	char := &character.Character{
		Name:  "Duelist",
		Level: 1,
		Class: &rulebook.Class{Key: "fighter"},
		Attributes: map[shared.Attribute]*character.AbilityScore{
			shared.AttributeStrength:  {Score: 16, Bonus: 3},
			shared.AttributeDexterity: {Score: 10, Bonus: 0},
		},
		Features: []*rulebook.CharacterFeature{
			{
				Key:  "fighting_style",
				Name: "Fighting Style",
				Metadata: map[string]any{
					"style": "dueling",
				},
			},
		},
		EquippedSlots: make(map[shared.Slot]equipment.Equipment),
	}

	// Equip a one-handed weapon
	longsword := &equipment.Weapon{
		Base: equipment.BasicEquipment{
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
		Properties: []*shared.ReferenceItem{
			{Key: "versatile"},
		},
	}
	char.EquippedSlots[shared.SlotMainHand] = longsword

	// Add martial weapon proficiency
	char.Proficiencies = map[rulebook.ProficiencyType][]*rulebook.Proficiency{
		rulebook.ProficiencyTypeWeapon: {
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
	shield := &equipment.Armor{
		Base: equipment.BasicEquipment{
			Key:  "shield",
			Name: "Shield",
		},
		ArmorClass: &equipment.ArmorClass{
			Base:     2,
			DexBonus: false,
		},
		ArmorCategory: "shield",
	}
	char.EquippedSlots[shared.SlotOffHand] = shield

	// Attack should still get dueling bonus with shield
	results2, err := char.Attack()
	require.NoError(t, err)
	require.Len(t, results2, 1)
	assert.GreaterOrEqual(t, results2[0].DamageRoll, 6, "Should still get dueling bonus with shield")

	// Now equip a weapon in off-hand (loses dueling bonus)
	dagger := &equipment.Weapon{
		Base: equipment.BasicEquipment{
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
		Properties: []*shared.ReferenceItem{
			{Key: "light"},
		},
	}
	char.EquippedSlots[shared.SlotOffHand] = dagger

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
	char := &character.Character{
		Name:  "Dual Wielder",
		Level: 1,
		Class: &rulebook.Class{Key: "fighter"},
		Attributes: map[shared.Attribute]*character.AbilityScore{
			shared.AttributeStrength:  {Score: 16, Bonus: 3},
			shared.AttributeDexterity: {Score: 14, Bonus: 2},
		},
		Features: []*rulebook.CharacterFeature{
			{
				Key:  "fighting_style",
				Name: "Fighting Style",
				Metadata: map[string]any{
					"style": "two_weapon",
				},
			},
		},
		EquippedSlots: make(map[shared.Slot]equipment.Equipment),
	}

	// Equip two light weapons
	shortsword := &equipment.Weapon{
		Base: equipment.BasicEquipment{
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
		Properties: []*shared.ReferenceItem{
			{Key: "light"},
		},
	}
	char.EquippedSlots[shared.SlotMainHand] = shortsword
	char.EquippedSlots[shared.SlotOffHand] = shortsword

	// Add martial weapon proficiency
	char.Proficiencies = map[rulebook.ProficiencyType][]*rulebook.Proficiency{
		rulebook.ProficiencyTypeWeapon: {
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
	char := &character.Character{
		Name:  "Defender",
		Level: 1,
		Class: &rulebook.Class{Key: "fighter"},
		Attributes: map[shared.Attribute]*character.AbilityScore{
			shared.AttributeDexterity:    {Score: 14, Bonus: 2},
			shared.AttributeConstitution: {Score: 14, Bonus: 2},
		},
		Features: []*rulebook.CharacterFeature{
			{
				Key:  "fighting_style",
				Name: "Fighting Style",
				Metadata: map[string]any{
					"style": "defense",
				},
			},
		},
		EquippedSlots: make(map[shared.Slot]equipment.Equipment),
	}

	// Calculate AC without armor (no defense bonus)
	char.calculateAC()
	assert.Equal(t, 10, char.AC) // 10 base, no DEX without armor in 5e by default

	// Equip armor
	chainMail := &equipment.Armor{
		Base: equipment.BasicEquipment{
			Key:  "chain-mail",
			Name: "Chain Mail",
		},
		ArmorClass: &equipment.ArmorClass{
			Base:     16,
			DexBonus: false,
		},
		ArmorCategory: "heavy",
	}
	char.EquippedSlots[shared.SlotBody] = chainMail

	// Calculate AC with armor (gets defense bonus)
	char.calculateAC()
	assert.Equal(t, 17, char.AC) // 16 (chain mail) + 1 (defense)
}

func TestFightingStyleGreatWeapon(t *testing.T) {
	// Create a fighter with great weapon fighting style
	char := &character.Character{
		Name:  "Great Weapon Fighter",
		Level: 1,
		Class: &rulebook.Class{Key: "fighter"},
		Attributes: map[shared.Attribute]*character.AbilityScore{
			shared.AttributeStrength:  {Score: 18, Bonus: 4},
			shared.AttributeDexterity: {Score: 10, Bonus: 0},
		},
		Features: []*rulebook.CharacterFeature{
			{
				Key:  "fighting_style",
				Name: "Fighting Style",
				Metadata: map[string]any{
					"style": "great_weapon",
				},
			},
		},
		EquippedSlots: make(map[shared.Slot]equipment.Equipment),
	}

	// Equip a two-handed weapon
	greatsword := &equipment.Weapon{
		Base: equipment.BasicEquipment{
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
		Properties: []*shared.ReferenceItem{
			{Key: "two-handed"},
		},
	}
	char.EquippedSlots[shared.SlotTwoHanded] = greatsword

	// Add martial weapon proficiency
	char.Proficiencies = map[rulebook.ProficiencyType][]*rulebook.Proficiency{
		rulebook.ProficiencyTypeWeapon: {
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
