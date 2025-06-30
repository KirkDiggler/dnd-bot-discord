package character

import (
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/damage"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/equipment"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/rulebook"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/shared"
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestCharacter_AddAbilityBonus(t *testing.T) {
	type fields struct {
		AbilityScores map[Attribute]*AbilityScore
	}
	type args struct {
		abilityBonus  *AbilityBonus
		expectedScore int
		expectedBonus int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "TestCharacter_AddAbilityBonus",
			fields: fields{
				AbilityScores: map[Attribute]*AbilityScore{
					AttributeStrength:     {Score: 10, Bonus: 0},
					AttributeDexterity:    {Score: 10},
					AttributeConstitution: {Score: 10},
					AttributeIntelligence: {Score: 10},
					AttributeWisdom:       {Score: 10},
					AttributeCharisma:     {Score: 10},
				},
			},
			args: args{
				abilityBonus: &AbilityBonus{
					Attribute: AttributeStrength,
					Bonus:     2,
				},
				expectedScore: 12,
				expectedBonus: 1, // (12-10)/2 = 1
			},
		},
		{
			name: "TestCharacter_AddToExistingAbilityBonus",
			fields: fields{
				AbilityScores: map[Attribute]*AbilityScore{
					AttributeStrength:     {Score: 12, Bonus: 1},
					AttributeDexterity:    {Score: 10},
					AttributeConstitution: {Score: 10},
					AttributeIntelligence: {Score: 10},
					AttributeWisdom:       {Score: 10},
					AttributeCharisma:     {Score: 10},
				},
			},
			args: args{
				abilityBonus: &AbilityBonus{
					Attribute: AttributeStrength,
					Bonus:     2,
				},
				expectedScore: 14,
				expectedBonus: 2, // (14-10)/2 = 2
			},
		},
		{
			name: "TestCharacter_AddToNewAttributeAbilityBonus",
			fields: fields{
				AbilityScores: map[Attribute]*AbilityScore{
					AttributeDexterity:    {Score: 10},
					AttributeConstitution: {Score: 10},
					AttributeIntelligence: {Score: 10},
					AttributeWisdom:       {Score: 10},
					AttributeCharisma:     {Score: 10},
				},
			},
			args: args{
				abilityBonus: &AbilityBonus{
					Attribute: AttributeStrength,
					Bonus:     2,
				},
				expectedScore: 2,
				expectedBonus: -4, // (2-10)/2 = -4
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Character{
				Attributes: tt.fields.AbilityScores,
			}

			c.AddAbilityBonus(tt.args.abilityBonus)

			if c.Attributes[tt.args.abilityBonus.Attribute].Score != tt.args.expectedScore {
				t.Errorf("expected score to be %d, got %d", tt.args.expectedScore, c.Attributes[tt.args.abilityBonus.Attribute].Score)
			}
			if c.Attributes[tt.args.abilityBonus.Attribute].Bonus != tt.args.expectedBonus {
				t.Errorf("expected bonus to be %d, got %d", tt.args.expectedBonus, c.Attributes[tt.args.abilityBonus.Attribute].Bonus)
			}
		})
	}
}

func TestCharacter_AddAttribute(t *testing.T) {
	type fields struct {
		AbilityScores map[Attribute]*AbilityScore
	}
	type args struct {
		attribute    Attribute
		abilityScore *AbilityScore
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "TestCharacter_AddAttribute",
			fields: fields{
				AbilityScores: map[Attribute]*AbilityScore{
					AttributeStrength:     {Score: 0, Bonus: 2},
					AttributeDexterity:    {Score: 10},
					AttributeConstitution: {Score: 10},
					AttributeIntelligence: {Score: 10},
					AttributeWisdom:       {Score: 10},
					AttributeCharisma:     {Score: 10},
				},
			},
			args: args{
				attribute: AttributeStrength,
				abilityScore: &AbilityScore{
					Score: 10,
					Bonus: 0, // (10-10)/2 = 0
				},
			},
		},
		{
			name: "TestCharacter_AddAttribute",
			fields: fields{
				AbilityScores: map[Attribute]*AbilityScore{
					AttributeStrength:     {Score: 0, Bonus: 0},
					AttributeDexterity:    {Score: 10},
					AttributeConstitution: {Score: 10},
					AttributeIntelligence: {Score: 10},
					AttributeWisdom:       {Score: 10},
					AttributeCharisma:     {Score: 10},
				},
			},
			args: args{
				attribute: AttributeStrength,
				abilityScore: &AbilityScore{
					Score: 12,
					Bonus: 1,
				},
			},
		},
		{
			name: "TestCharacter_AddAttribute",
			fields: fields{
				AbilityScores: map[Attribute]*AbilityScore{
					AttributeStrength:     {Score: 10},
					AttributeDexterity:    {Score: 10},
					AttributeConstitution: {Score: 10},
					AttributeIntelligence: {Score: 10},
					AttributeWisdom:       {Score: 10},
				},
			},
			args: args{
				attribute: AttributeCharisma,
				abilityScore: &AbilityScore{
					Score: 10,
					Bonus: 0,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Character{
				Attributes: tt.fields.AbilityScores,
			}

			c.AddAttribute(tt.args.attribute, tt.args.abilityScore.Score)
			_ = c.String()
			if _, ok := c.Attributes[tt.args.attribute]; !ok {
				t.Errorf("expected attribute %s to be present", tt.args.attribute)
			}

			if c.Attributes[tt.args.attribute].Score != tt.args.abilityScore.Score {
				t.Errorf("expected score to be %d, got %d", tt.args.abilityScore.Score, c.Attributes[tt.args.attribute].Score)
			}

			if c.Attributes[tt.args.attribute].Bonus != tt.args.abilityScore.Bonus {
				t.Errorf("expected bonus to be %d, got %d", tt.args.abilityScore.Bonus, c.Attributes[tt.args.attribute].Bonus)
			}
		})
	}
}

type suiteEquip struct {
	suite.Suite
	char *Character
}

func (s *suiteEquip) SetupTest() {
	s.char = &Character{
		Inventory: map[equipment.EquipmentType][]equipment.Equipment{
			equipment.EquipmentTypeWeapon: {
				&equipment.Weapon{
					Base: equipment.BasicEquipment{
						Key:  "sword",
						Name: "Sword",
					},
				},
				&equipment.Weapon{
					Base: equipment.BasicEquipment{
						Key:  "greatsword",
						Name: "Greatsword",
					},
					Properties: []*shared.ReferenceItem{{
						Key:  "two-handed",
						Name: "Two-Handed",
					}},
				},
			},
			equipment.EquipmentTypeArmor: {
				&equipment.Armor{
					Base: equipment.BasicEquipment{
						Key:  "leather-armor",
						Name: "Leather Armor",
					},
					ArmorClass: &equipment.ArmorClass{
						Base:     12,
						DexBonus: true,
					},
				},
				&equipment.Armor{
					Base: equipment.BasicEquipment{
						Key:  "shield",
						Name: "Shield",
					},
					ArmorCategory: equipment.ArmorCategoryShield,
					ArmorClass:    &equipment.ArmorClass{Base: 2},
				},
			},
		},
	}
	for _, v := range Attributes {
		s.char.AddAttribute(v, 10)
	}
}

func (s *suiteEquip) TestEquip() {
	actual := s.char.Equip(s.char.Inventory[equipment.EquipmentTypeWeapon][0].GetKey())
	s.Equal(10, s.char.AC)
	s.Equal(true, actual)
}

func (s *suiteEquip) TestEquipArmor() {
	s.char.Equip("leather-armor")

	s.Equal(12, s.char.AC)

}

func (s *suiteEquip) TestEquipArmorWithDexBonus() {
	s.char.AddAttribute(AttributeDexterity, 14)
	s.char.Equip("leather-armor")

	s.Equal(14, s.char.AC)
}

func (s *suiteEquip) TestEquipArmorAndShield() {
	s.char.Equip("leather-armor")
	s.char.Equip("shield")
	s.Equal(14, s.char.AC)
}

func (s *suiteEquip) TestEquipTwoItemsWithShield() {
	shield := s.char.Inventory[equipment.EquipmentTypeArmor][1]
	sword := s.char.Inventory[equipment.EquipmentTypeWeapon][0]

	s.char.Equip("shield")
	s.char.Equip("sword")

	s.Equal(12, s.char.AC)
	s.Equal(sword, s.char.EquippedSlots[SlotMainHand])
	s.Equal(shield, s.char.EquippedSlots[SlotOffHand])
}

func (s *suiteEquip) TestTwoHandedOverwritesSlots() {

	shield := s.char.Inventory[equipment.EquipmentTypeArmor][1]
	sword := s.char.Inventory[equipment.EquipmentTypeWeapon][0]
	greatsword := s.char.Inventory[equipment.EquipmentTypeWeapon][1]

	s.char.Equip("shield")
	s.char.Equip("sword")

	s.Equal(12, s.char.AC)
	s.Equal(sword, s.char.EquippedSlots[SlotMainHand])
	s.Equal(shield, s.char.EquippedSlots[SlotOffHand])

	s.char.Equip("greatsword")

	s.Equal(10, s.char.AC)
	s.Equal(greatsword, s.char.EquippedSlots[SlotTwoHanded])
	s.Nil(s.char.EquippedSlots[SlotOffHand])
	s.Nil(s.char.EquippedSlots[SlotMainHand])
}

func TestSuiteEquip(t *testing.T) {
	suite.Run(t, new(suiteEquip))
}

func TestCharacter_Attack_WithRageBonus(t *testing.T) {
	// Create a barbarian character with a weapon
	char := &Character{
		Name:  "Ragnar",
		Level: 1,
		Class: &rulebook.Class{Key: "barbarian", Name: "Barbarian"},
		Attributes: map[Attribute]*AbilityScore{
			AttributeStrength: {Score: 16, Bonus: 3}, // +3 STR modifier
		},
		EquippedSlots: map[Slot]equipment.Equipment{
			SlotMainHand: &equipment.Weapon{
				Base: equipment.BasicEquipment{
					Key:  "greataxe",
					Name: "Greataxe",
				},
				WeaponRange: "Melee",
				Damage: &damage.Damage{
					DiceCount:  1,
					DiceSize:   12,
					Bonus:      0,
					DamageType: damage.TypeSlashing,
				},
			},
		},
		Proficiencies: map[rulebook.ProficiencyType][]*rulebook.Proficiency{
			rulebook.ProficiencyTypeWeapon: {
				{Key: "greataxe", Name: "Greataxe"},
			},
		},
	}

	// Initialize resources
	char.InitializeResources()

	// Verify rage ability exists
	rage, exists := char.Resources.Abilities["rage"]
	if !exists {
		t.Fatal("Rage ability not found")
	}

	// Activate rage (simulate the effect being added)
	char.Resources.AddEffect(&shared.ActiveEffect{
		ID:           "rage-effect",
		Name:         "Rage",
		Source:       "rage",
		DurationType: shared.DurationTypeRounds,
		Duration:     10,
		Modifiers: []shared.Modifier{
			{
				Type:        shared.ModifierTypeDamageBonus,
				Value:       2,
				DamageTypes: []string{"melee"},
			},
		},
	})

	// Mark rage as active
	rage.IsActive = true

	// Perform attack
	attacks, err := char.Attack()
	if err != nil {
		t.Fatalf("Attack failed: %v", err)
	}

	if len(attacks) != 1 {
		t.Fatalf("Expected 1 attack, got %d", len(attacks))
	}

	attack := attacks[0]

	// Attack roll should include:
	// - STR bonus (+3)
	// - Proficiency bonus (+2 at level 1)
	// Total attack bonus should be +5
	// Note: Natural 1 is automatic miss (attack roll set to 0)
	if attack.AttackResult.Rolls[0] != 1 && attack.AttackResult.Rolls[0] != 20 {
		// Only check attack bonus on non-critical rolls
		expectedAttackBonus := 5
		actualAttackBonus := attack.AttackRoll - attack.AttackResult.Total
		if actualAttackBonus != expectedAttackBonus {
			t.Errorf("Expected attack bonus +%d, got +%d", expectedAttackBonus, actualAttackBonus)
		}
	}

	// Damage roll should include:
	// - STR bonus (+3)
	// - Rage bonus (+2)
	// Total damage bonus should be +5
	expectedDamageBonus := 5
	actualDamageBonus := attack.DamageRoll - attack.DamageResult.Total
	if actualDamageBonus != expectedDamageBonus {
		t.Errorf("Expected damage bonus +%d, got +%d (damage roll: %d, dice total: %d)",
			expectedDamageBonus, actualDamageBonus, attack.DamageRoll, attack.DamageResult.Total)
	}
}

func TestCharacter_Attack_WithoutRage(t *testing.T) {
	// Create a barbarian character with a weapon but no rage active
	char := &Character{
		Name:  "Ragnar",
		Level: 1,
		Class: &rulebook.Class{Key: "barbarian", Name: "Barbarian"},
		Attributes: map[Attribute]*AbilityScore{
			AttributeStrength: {Score: 16, Bonus: 3}, // +3 STR modifier
		},
		EquippedSlots: map[Slot]equipment.Equipment{
			SlotMainHand: &equipment.Weapon{
				Base: equipment.BasicEquipment{
					Key:  "greataxe",
					Name: "Greataxe",
				},
				WeaponRange: "Melee",
				Damage: &damage.Damage{
					DiceCount:  1,
					DiceSize:   12,
					Bonus:      0,
					DamageType: damage.TypeSlashing,
				},
			},
		},
		Proficiencies: map[rulebook.ProficiencyType][]*rulebook.Proficiency{
			rulebook.ProficiencyTypeWeapon: {
				{Key: "greataxe", Name: "Greataxe"},
			},
		},
	}

	// Initialize resources but don't activate rage
	char.InitializeResources()

	// Perform attack
	attacks, err := char.Attack()
	if err != nil {
		t.Fatalf("Attack failed: %v", err)
	}

	if len(attacks) != 1 {
		t.Fatalf("Expected 1 attack, got %d", len(attacks))
	}

	attack := attacks[0]

	// Damage roll should only include STR bonus (+3), no rage bonus
	expectedDamageBonus := 3

	// For critical hits, we need to account for doubled dice
	if attack.AttackResult.Rolls[0] == 20 {
		// Critical hit - damage includes doubled dice plus bonus once
		// Skip this check for criticals as the math is different
		t.Logf("Critical hit rolled - damage includes doubled dice")
	} else {
		actualDamageBonus := attack.DamageRoll - attack.DamageResult.Total
		if actualDamageBonus != expectedDamageBonus {
			t.Errorf("Expected damage bonus +%d, got +%d", expectedDamageBonus, actualDamageBonus)
		}
	}
}
