package calculators_test

import (
	"testing"

	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/equipment"
	rulebook "github.com/KirkDiggler/dnd-bot-discord/internal/domain/rulebook/dnd5e"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/rulebook/dnd5e/calculators"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/shared"
	"github.com/stretchr/testify/assert"
)

func TestDnD5eACCalculator_Calculate(t *testing.T) {
	calculator := calculators.NewDnD5eACCalculator()

	tests := []struct {
		name     string
		setup    func() *character.Character
		expected int
	}{
		{
			name: "base AC with no equipment",
			setup: func() *character.Character {
				char := &character.Character{
					ID:         "test-1",
					Attributes: map[shared.Attribute]*character.AbilityScore{},
				}
				// No DEX bonus
				char.Attributes[shared.AttributeDexterity] = &character.AbilityScore{
					Score: 10,
					Bonus: 0,
				}
				return char
			},
			expected: 10,
		},
		{
			name: "base AC with DEX bonus",
			setup: func() *character.Character {
				char := &character.Character{
					ID:         "test-2",
					Attributes: map[shared.Attribute]*character.AbilityScore{},
				}
				// +3 DEX bonus
				char.Attributes[shared.AttributeDexterity] = &character.AbilityScore{
					Score: 16,
					Bonus: 3,
				}
				return char
			},
			expected: 13, // 10 + 3 DEX
		},
		{
			name: "leather armor with DEX bonus",
			setup: func() *character.Character {
				char := &character.Character{
					ID:            "test-3",
					Attributes:    map[shared.Attribute]*character.AbilityScore{},
					EquippedSlots: map[shared.Slot]equipment.Equipment{},
				}
				char.Attributes[shared.AttributeDexterity] = &character.AbilityScore{
					Score: 14,
					Bonus: 2,
				}
				// Add leather armor
				char.EquippedSlots[shared.SlotBody] = &equipment.Armor{
					Base: equipment.BasicEquipment{
						Name: "Leather Armor",
						Key:  "leather-armor",
					},
					ArmorCategory: equipment.ArmorCategoryLight,
					ArmorClass: &equipment.ArmorClass{
						Base:     11,
						DexBonus: true,
					},
				}
				return char
			},
			expected: 13, // 11 base + 2 DEX
		},
		{
			name: "medium armor with DEX cap",
			setup: func() *character.Character {
				char := &character.Character{
					ID:            "test-4",
					Attributes:    map[shared.Attribute]*character.AbilityScore{},
					EquippedSlots: map[shared.Slot]equipment.Equipment{},
				}
				// High DEX that will be capped
				char.Attributes[shared.AttributeDexterity] = &character.AbilityScore{
					Score: 18,
					Bonus: 4,
				}
				// Add half plate (max +2 DEX)
				char.EquippedSlots[shared.SlotBody] = &equipment.Armor{
					Base: equipment.BasicEquipment{
						Name: "Half Plate",
						Key:  "half-plate",
					},
					ArmorCategory: equipment.ArmorCategoryMedium,
					ArmorClass: &equipment.ArmorClass{
						Base:     15,
						DexBonus: true,
						MaxBonus: 2,
					},
				}
				return char
			},
			expected: 17, // 15 base + 2 DEX (capped from 4)
		},
		{
			name: "heavy armor ignores DEX",
			setup: func() *character.Character {
				char := &character.Character{
					ID:            "test-5",
					Attributes:    map[shared.Attribute]*character.AbilityScore{},
					EquippedSlots: map[shared.Slot]equipment.Equipment{},
				}
				char.Attributes[shared.AttributeDexterity] = &character.AbilityScore{
					Score: 16,
					Bonus: 3,
				}
				// Add plate armor
				char.EquippedSlots[shared.SlotBody] = &equipment.Armor{
					Base: equipment.BasicEquipment{
						Name: "Plate Armor",
						Key:  "plate-armor",
					},
					ArmorCategory: equipment.ArmorCategoryHeavy,
					ArmorClass: &equipment.ArmorClass{
						Base:     18,
						DexBonus: false,
					},
				}
				return char
			},
			expected: 18, // 18 base, no DEX
		},
		{
			name: "armor with shield",
			setup: func() *character.Character {
				char := &character.Character{
					ID:            "test-6",
					Attributes:    map[shared.Attribute]*character.AbilityScore{},
					EquippedSlots: map[shared.Slot]equipment.Equipment{},
				}
				char.Attributes[shared.AttributeDexterity] = &character.AbilityScore{
					Score: 12,
					Bonus: 1,
				}
				// Add chain mail
				char.EquippedSlots[shared.SlotBody] = &equipment.Armor{
					Base: equipment.BasicEquipment{
						Name: "Chain Mail",
						Key:  "chain-mail",
					},
					ArmorCategory: equipment.ArmorCategoryHeavy,
					ArmorClass: &equipment.ArmorClass{
						Base:     16,
						DexBonus: false,
					},
				}
				// Add shield (shield is armor with shield category)
				char.EquippedSlots[shared.SlotOffHand] = &equipment.Armor{
					Base: equipment.BasicEquipment{
						Name: "Shield",
						Key:  "shield",
					},
					ArmorCategory: equipment.ArmorCategoryShield,
				}
				return char
			},
			expected: 18, // 16 base + 2 shield
		},
		{
			name: "monk unarmored defense",
			setup: func() *character.Character {
				char := &character.Character{
					ID:         "test-7",
					Level:      1,
					Attributes: map[shared.Attribute]*character.AbilityScore{},
					Class: &rulebook.Class{
						Key:  "monk",
						Name: "Monk",
					},
				}
				char.Attributes[shared.AttributeDexterity] = &character.AbilityScore{
					Score: 16,
					Bonus: 3,
				}
				char.Attributes[shared.AttributeWisdom] = &character.AbilityScore{
					Score: 14,
					Bonus: 2,
				}
				return char
			},
			expected: 15, // 10 + 3 DEX + 2 WIS
		},
		{
			name: "barbarian unarmored defense",
			setup: func() *character.Character {
				char := &character.Character{
					ID:         "test-8",
					Level:      1,
					Attributes: map[shared.Attribute]*character.AbilityScore{},
					Class: &rulebook.Class{
						Key:  "barbarian",
						Name: "Barbarian",
					},
				}
				char.Attributes[shared.AttributeDexterity] = &character.AbilityScore{
					Score: 14,
					Bonus: 2,
				}
				char.Attributes[shared.AttributeConstitution] = &character.AbilityScore{
					Score: 16,
					Bonus: 3,
				}
				return char
			},
			expected: 15, // 10 + 2 DEX + 3 CON
		},
		{
			name: "defense fighting style with armor",
			setup: func() *character.Character {
				char := &character.Character{
					ID:            "test-9",
					Level:         1,
					Attributes:    map[shared.Attribute]*character.AbilityScore{},
					EquippedSlots: map[shared.Slot]equipment.Equipment{},
					Features:      []*rulebook.CharacterFeature{},
				}
				char.Attributes[shared.AttributeDexterity] = &character.AbilityScore{
					Score: 12,
					Bonus: 1,
				}
				// Add chain shirt
				char.EquippedSlots[shared.SlotBody] = &equipment.Armor{
					Base: equipment.BasicEquipment{
						Name: "Chain Shirt",
						Key:  "chain-shirt",
					},
					ArmorCategory: equipment.ArmorCategoryMedium,
					ArmorClass: &equipment.ArmorClass{
						Base:     13,
						DexBonus: true,
						MaxBonus: 2,
					},
				}
				// Add defense fighting style
				char.Features = append(char.Features, &rulebook.CharacterFeature{
					Key:      "fighting_style",
					Name:     "Fighting Style",
					Type:     rulebook.FeatureTypeClass,
					Metadata: map[string]interface{}{"style": "defense"},
				})
				return char
			},
			expected: 15, // 13 base + 1 DEX + 1 defense
		},
		{
			name: "active effect AC bonus",
			setup: func() *character.Character {
				char := &character.Character{
					ID:         "test-10",
					Attributes: map[shared.Attribute]*character.AbilityScore{},
					Resources:  &character.CharacterResources{},
				}
				char.Attributes[shared.AttributeDexterity] = &character.AbilityScore{
					Score: 10,
					Bonus: 0,
				}
				// Add shield of faith effect (+2 AC)
				char.Resources.ActiveEffects = []*shared.ActiveEffect{
					{
						ID:          "shield-of-faith",
						Name:        "Shield of Faith",
						Description: "+2 AC",
						Modifiers: []shared.Modifier{
							{
								Type:  shared.ModifierTypeACBonus,
								Value: 2,
							},
						},
					},
				}
				return char
			},
			expected: 12, // 10 base + 0 DEX + 2 effect
		},
		{
			name: "nil character returns base AC",
			setup: func() *character.Character {
				return nil
			},
			expected: 10,
		},
		{
			name: "armor fallback for chain mail",
			setup: func() *character.Character {
				char := &character.Character{
					ID:            "test-11",
					Attributes:    map[shared.Attribute]*character.AbilityScore{},
					EquippedSlots: map[shared.Slot]equipment.Equipment{},
				}
				char.Attributes[shared.AttributeDexterity] = &character.AbilityScore{
					Score: 14,
					Bonus: 2,
				}
				// Add chain mail without AC data (fallback test)
				char.EquippedSlots[shared.SlotBody] = &equipment.Armor{
					Base: equipment.BasicEquipment{
						Name: "Chain Mail",
						Key:  "chain-mail",
					},
					ArmorCategory: equipment.ArmorCategoryHeavy,
					// No ArmorClass data - will use fallback
				}
				return char
			},
			expected: 16, // Fallback: 16 base, no DEX for heavy armor
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			char := tt.setup()
			result := calculator.Calculate(char)
			assert.Equal(t, tt.expected, result, "AC calculation mismatch")
		})
	}
}
