package entities

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestCharacter_AddAbilityBonus(t *testing.T) {
	type fields struct {
		AbilityScores map[Attribute]*AbilityScore
	}
	type args struct {
		abilityBonus  *AbilityBonus
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
					AttributeStrength:     {Score: 10},
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
				expectedBonus: 2,
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
				expectedBonus: 3,
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
				expectedBonus: 2,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Character{
				Attributes: tt.fields.AbilityScores,
			}

			c.AddAbilityBonus(tt.args.abilityBonus)

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
					Bonus: 2,
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
		Inventory: map[EquipmentType][]Equipment{
			EquipmentTypeWeapon: {
				&Weapon{
					Base: BasicEquipment{
						Key:  "sword",
						Name: "Sword",
					},
				},
				&Weapon{
					Base: BasicEquipment{
						Key:  "greatsword",
						Name: "Greatsword",
					},
					Properties: []*ReferenceItem{{
						Key:  "two-handed",
						Name: "Two-Handed",
					}},
				},
			},
			EquipmentTypeArmor: {
				&Armor{
					Base: BasicEquipment{
						Key:  "leather-armor",
						Name: "Leather Armor",
					},
					ArmorClass: &ArmorClass{
						Base:     12,
						DexBonus: true,
					},
				},
				&Armor{
					Base: BasicEquipment{
						Key:  "shield",
						Name: "Shield",
					},
					ArmorCategory: ArmorCategoryShield,
					ArmorClass:    &ArmorClass{Base: 2},
				},
			},
		},
	}
	for _, v := range Attributes {
		s.char.AddAttribute(v, 10)
	}
}

func (s *suiteEquip) TestEquip() {
	actual := s.char.Equip(s.char.Inventory[EquipmentTypeWeapon][0].GetKey())
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
	shield := s.char.Inventory[EquipmentTypeArmor][1]
	sword := s.char.Inventory[EquipmentTypeWeapon][0]

	s.char.Equip("shield")
	s.char.Equip("sword")

	s.Equal(12, s.char.AC)
	s.Equal(sword, s.char.EquippedSlots[SlotMainHand])
	s.Equal(shield, s.char.EquippedSlots[SlotOffHand])
}

func (s *suiteEquip) TestTwoHandedOverwritesSlots() {

	shield := s.char.Inventory[EquipmentTypeArmor][1]
	sword := s.char.Inventory[EquipmentTypeWeapon][0]
	greatsword := s.char.Inventory[EquipmentTypeWeapon][1]

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
