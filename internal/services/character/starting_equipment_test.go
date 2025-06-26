package character_test

import (
	"context"
	"errors"
	"testing"

	"github.com/KirkDiggler/dnd-bot-discord/internal/entities"
	"github.com/KirkDiggler/dnd-bot-discord/internal/entities/damage"
	"github.com/KirkDiggler/dnd-bot-discord/internal/repositories/characters"
	"github.com/KirkDiggler/dnd-bot-discord/internal/services/character"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	mockdnd5e "github.com/KirkDiggler/dnd-bot-discord/internal/clients/dnd5e/mock"
)

func TestFinalizeDraftCharacter_AddsStartingEquipment(t *testing.T) {
	// Create test fixtures
	fighterClass := &entities.Class{
		Key:    "fighter",
		Name:   "Fighter",
		HitDie: 10,
		StartingEquipment: []*entities.StartingEquipment{
			{
				Quantity: 1,
				Equipment: &entities.ReferenceItem{
					Key:  "chain-mail",
					Name: "Chain Mail",
				},
			},
			{
				Quantity: 5,
				Equipment: &entities.ReferenceItem{
					Key:  "javelin",
					Name: "Javelin",
				},
			},
		},
	}

	rangerClass := &entities.Class{
		Key:    "ranger",
		Name:   "Ranger",
		HitDie: 10,
		StartingEquipment: []*entities.StartingEquipment{
			{
				Quantity: 1,
				Equipment: &entities.ReferenceItem{
					Key:  "scale-mail",
					Name: "Scale Mail",
				},
			},
			{
				Quantity: 1,
				Equipment: &entities.ReferenceItem{
					Key:  "shortbow",
					Name: "Shortbow",
				},
			},
			{
				Quantity: 20,
				Equipment: &entities.ReferenceItem{
					Key:  "arrow",
					Name: "Arrow",
				},
			},
		},
	}

	// Create equipment
	chainMail := &entities.Armor{
		Base: entities.BasicEquipment{
			Key:  "chain-mail",
			Name: "Chain Mail",
			Cost: &entities.Cost{
				Quantity: 75,
				Unit:     "gp",
			},
		},
		ArmorClass: &entities.ArmorClass{
			Base:     16,
			DexBonus: false,
		},
		ArmorCategory: entities.ArmorCategoryHeavy,
	}

	javelin := &entities.Weapon{
		Base: entities.BasicEquipment{
			Key:  "javelin",
			Name: "Javelin",
			Cost: &entities.Cost{
				Quantity: 5,
				Unit:     "sp",
			},
		},
		WeaponRange: "Melee",
		Properties: []*entities.ReferenceItem{
			{Key: "thrown", Name: "Thrown"},
		},
		Damage: &damage.Damage{
			DiceCount:  1,
			DiceSize:   6,
			Bonus:      0,
			DamageType: damage.TypePiercing,
		},
	}

	scaleMail := &entities.Armor{
		Base: entities.BasicEquipment{
			Key:  "scale-mail",
			Name: "Scale Mail",
			Cost: &entities.Cost{
				Quantity: 50,
				Unit:     "gp",
			},
		},
		ArmorClass: &entities.ArmorClass{
			Base:     14,
			DexBonus: true,
			MaxBonus: 2,
		},
		ArmorCategory: entities.ArmorCategoryMedium,
	}

	shortbow := &entities.Weapon{
		Base: entities.BasicEquipment{
			Key:  "shortbow",
			Name: "Shortbow",
			Cost: &entities.Cost{
				Quantity: 25,
				Unit:     "gp",
			},
		},
		WeaponRange: "Ranged",
		Properties: []*entities.ReferenceItem{
			{Key: "ammunition", Name: "Ammunition"},
		},
		Damage: &damage.Damage{
			DiceCount:  1,
			DiceSize:   6,
			Bonus:      0,
			DamageType: damage.TypePiercing,
		},
		Range: 80,
	}

	arrow := &entities.BasicEquipment{
		Key:  "arrow",
		Name: "Arrow",
		Cost: &entities.Cost{
			Quantity: 1,
			Unit:     "gp",
		},
	}

	tests := []struct {
		name                  string
		character             *entities.Character
		class                 *entities.Class
		expectedEquipmentKeys map[string]int // key -> expected quantity
	}{
		{
			name: "Fighter gets chain mail and javelins",
			character: &entities.Character{
				ID:     "test-fighter",
				Name:   "Test Fighter",
				Status: entities.CharacterStatusDraft,
				Level:  1,
				Class:  fighterClass,
				Attributes: map[entities.Attribute]*entities.AbilityScore{
					entities.AttributeStrength:     {Score: 16, Bonus: 3},
					entities.AttributeDexterity:    {Score: 12, Bonus: 1},
					entities.AttributeConstitution: {Score: 14, Bonus: 2},
					entities.AttributeIntelligence: {Score: 10, Bonus: 0},
					entities.AttributeWisdom:       {Score: 12, Bonus: 1},
					entities.AttributeCharisma:     {Score: 8, Bonus: -1},
				},
			},
			class: fighterClass,
			expectedEquipmentKeys: map[string]int{
				"chain-mail": 1,
				"javelin":    5,
			},
		},
		{
			name: "Ranger gets scale mail, shortbow and arrows",
			character: &entities.Character{
				ID:     "test-ranger",
				Name:   "Test Ranger",
				Status: entities.CharacterStatusDraft,
				Level:  1,
				Class:  rangerClass,
				Attributes: map[entities.Attribute]*entities.AbilityScore{
					entities.AttributeStrength:     {Score: 14, Bonus: 2},
					entities.AttributeDexterity:    {Score: 16, Bonus: 3},
					entities.AttributeConstitution: {Score: 13, Bonus: 1},
					entities.AttributeIntelligence: {Score: 10, Bonus: 0},
					entities.AttributeWisdom:       {Score: 14, Bonus: 2},
					entities.AttributeCharisma:     {Score: 10, Bonus: 0},
				},
			},
			class: rangerClass,
			expectedEquipmentKeys: map[string]int{
				"scale-mail": 1,
				"shortbow":   1,
				"arrow":      20,
			},
		},
		{
			name: "Character with existing equipment still gets starting equipment",
			character: &entities.Character{
				ID:     "test-with-equipment",
				Name:   "Test With Equipment",
				Status: entities.CharacterStatusDraft,
				Level:  1,
				Class:  fighterClass,
				Inventory: map[entities.EquipmentType][]entities.Equipment{
					entities.EquipmentTypeWeapon: []entities.Equipment{
						&entities.Weapon{
							Base: entities.BasicEquipment{
								Key:  "longsword",
								Name: "Longsword",
							},
							WeaponRange: "Melee",
							Damage: &damage.Damage{
								DiceCount:  1,
								DiceSize:   8,
								Bonus:      0,
								DamageType: damage.TypeSlashing,
							},
						},
					},
				},
				Attributes: map[entities.Attribute]*entities.AbilityScore{
					entities.AttributeStrength:     {Score: 16, Bonus: 3},
					entities.AttributeDexterity:    {Score: 12, Bonus: 1},
					entities.AttributeConstitution: {Score: 14, Bonus: 2},
					entities.AttributeIntelligence: {Score: 10, Bonus: 0},
					entities.AttributeWisdom:       {Score: 12, Bonus: 1},
					entities.AttributeCharisma:     {Score: 8, Bonus: -1},
				},
			},
			class: fighterClass,
			expectedEquipmentKeys: map[string]int{
				"longsword":  1,
				"chain-mail": 1,
				"javelin":    5,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock client
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockClient := mockdnd5e.NewMockClient(ctrl)

			// Set up equipment expectations
			mockClient.EXPECT().GetEquipment("chain-mail").Return(chainMail, nil).AnyTimes()
			mockClient.EXPECT().GetEquipment("javelin").Return(javelin, nil).AnyTimes()
			mockClient.EXPECT().GetEquipment("scale-mail").Return(scaleMail, nil).AnyTimes()
			mockClient.EXPECT().GetEquipment("shortbow").Return(shortbow, nil).AnyTimes()
			mockClient.EXPECT().GetEquipment("arrow").Return(arrow, nil).AnyTimes()
			mockClient.EXPECT().GetClassFeatures(tt.class.Key, 1).Return([]*entities.CharacterFeature{}, nil).AnyTimes()

			// Create repository and service
			repo := characters.NewInMemoryRepository()
			service := character.NewService(&character.ServiceConfig{
				Repository: repo,
				DNDClient:  mockClient,
			})

			// Store the character
			err := repo.Create(context.Background(), tt.character)
			require.NoError(t, err)

			// Finalize the character
			finalizedChar, err := service.FinalizeDraftCharacter(context.Background(), tt.character.ID)
			require.NoError(t, err)
			require.NotNil(t, finalizedChar)

			// Verify status changed to active
			require.Equal(t, entities.CharacterStatusActive, finalizedChar.Status)

			// Count all equipment by key
			actualEquipmentCounts := make(map[string]int)
			for _, equipmentList := range finalizedChar.Inventory {
				for _, equipment := range equipmentList {
					actualEquipmentCounts[equipment.GetKey()]++
				}
			}

			// Verify all expected equipment is in inventory with correct quantities
			for equipKey, expectedCount := range tt.expectedEquipmentKeys {
				actualCount := actualEquipmentCounts[equipKey]
				require.Equal(t, expectedCount, actualCount,
					"Wrong quantity for %s: expected %d, got %d",
					equipKey, expectedCount, actualCount)
			}

			// Verify total equipment count
			totalExpected := 0
			for _, count := range tt.expectedEquipmentKeys {
				totalExpected += count
			}
			totalActual := 0
			for _, count := range actualEquipmentCounts {
				totalActual += count
			}
			require.Equal(t, totalExpected, totalActual,
				"Total equipment count mismatch: expected %d, got %d", totalExpected, totalActual)
		})
	}
}

func TestFinalizeDraftCharacter_HandlesEquipmentErrors(t *testing.T) {
	// Create test class with equipment that will fail to load
	testClass := &entities.Class{
		Key:    "test-class",
		Name:   "Test Class",
		HitDie: 8,
		StartingEquipment: []*entities.StartingEquipment{
			{
				Quantity: 1,
				Equipment: &entities.ReferenceItem{
					Key:  "valid-equipment",
					Name: "Valid Equipment",
				},
			},
			{
				Quantity: 1,
				Equipment: &entities.ReferenceItem{
					Key:  "missing-equipment",
					Name: "Missing Equipment",
				},
			},
		},
	}

	validEquipment := &entities.BasicEquipment{
		Key:  "valid-equipment",
		Name: "Valid Equipment",
	}

	// Create character
	testChar := &entities.Character{
		ID:     "test-char",
		Name:   "Test Character",
		Status: entities.CharacterStatusDraft,
		Level:  1,
		Class:  testClass,
		Attributes: map[entities.Attribute]*entities.AbilityScore{
			entities.AttributeStrength:     {Score: 10, Bonus: 0},
			entities.AttributeDexterity:    {Score: 10, Bonus: 0},
			entities.AttributeConstitution: {Score: 10, Bonus: 0},
			entities.AttributeIntelligence: {Score: 10, Bonus: 0},
			entities.AttributeWisdom:       {Score: 10, Bonus: 0},
			entities.AttributeCharisma:     {Score: 10, Bonus: 0},
		},
	}

	// Create mock client
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mockdnd5e.NewMockClient(ctrl)
	mockClient.EXPECT().GetEquipment("valid-equipment").Return(validEquipment, nil)
	mockClient.EXPECT().GetEquipment("missing-equipment").Return(nil, errors.New("not found"))
	mockClient.EXPECT().GetClassFeatures(testClass.Key, 1).Return([]*entities.CharacterFeature{}, nil)

	// Create repository and service
	repo := characters.NewInMemoryRepository()
	service := character.NewService(&character.ServiceConfig{
		Repository: repo,
		DNDClient:  mockClient,
	})

	// Store the character
	err := repo.Create(context.Background(), testChar)
	require.NoError(t, err)

	// Finalize the character - should succeed despite one equipment error
	finalizedChar, err := service.FinalizeDraftCharacter(context.Background(), testChar.ID)
	require.NoError(t, err)
	require.NotNil(t, finalizedChar)

	// Verify status changed to active
	require.Equal(t, entities.CharacterStatusActive, finalizedChar.Status)

	// Verify only valid equipment was added
	totalEquipment := 0
	foundValidEquipment := false
	for _, equipmentList := range finalizedChar.Inventory {
		for _, equipment := range equipmentList {
			totalEquipment++
			if equipment.GetKey() == "valid-equipment" {
				foundValidEquipment = true
			}
		}
	}
	require.True(t, foundValidEquipment, "Valid equipment should have been added")
	require.Equal(t, 1, totalEquipment, "Only valid equipment should have been added")
}
