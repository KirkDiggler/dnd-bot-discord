package character_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/KirkDiggler/dnd-bot-discord/internal/entities"
	"github.com/KirkDiggler/dnd-bot-discord/internal/repositories/characters"
	"github.com/KirkDiggler/dnd-bot-discord/internal/services/character"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFighterWeaponPlusShieldFlow(t *testing.T) {
	// Setup
	ctx := context.Background()
	mockRepo := characters.NewInMemoryRepository()

	// Create mock D&D client
	mockClient := &mockDndClient{
		equipment: map[string]entities.Equipment{
			"chain-mail": &entities.Armor{
				Base: entities.BasicEquipment{
					Key:  "chain-mail",
					Name: "Chain Mail",
				},
				ArmorCategory: entities.ArmorCategoryHeavy,
			},
			"warhammer": &entities.Weapon{
				Base: entities.BasicEquipment{
					Key:  "warhammer",
					Name: "Warhammer",
				},
				WeaponCategory: "martial",
			},
			"shield": &entities.Armor{
				Base: entities.BasicEquipment{
					Key:  "shield",
					Name: "Shield",
				},
				ArmorCategory: entities.ArmorCategoryShield,
			},
		},
		classes: map[string]*entities.Class{
			"fighter": createFighterClassWithEquipmentChoices(),
		},
		races: map[string]*entities.Race{
			"human": {
				Key:  "human",
				Name: "Human",
			},
		},
	}

	service := character.NewService(&character.ServiceConfig{
		Repository: mockRepo,
		DNDClient:  mockClient,
	})

	// Create a draft character
	draftChar, err := service.GetOrCreateDraftCharacter(ctx, "test-user", "test-realm")
	require.NoError(t, err)
	require.NotNil(t, draftChar)

	// Update with race and class
	_, err = service.UpdateDraftCharacter(ctx, draftChar.ID, &character.UpdateDraftInput{
		RaceKey:  strPtr("human"),
		ClassKey: strPtr("fighter"),
	})
	require.NoError(t, err)

	// Simulate equipment selection flow:
	// 1. Select chain mail (direct equipment)
	_, err = service.UpdateDraftCharacter(ctx, draftChar.ID, &character.UpdateDraftInput{
		Equipment: []string{"chain-mail"},
	})
	require.NoError(t, err)

	// 2. Select "weapon + shield" bundle - this would be filtered out by handler
	// Handler would see "bundle-0" and skip it (as per our fix)

	// 3. Test that the choice resolver properly tracks bundle items
	choices, err := service.ResolveChoices(ctx, &character.ResolveChoicesInput{
		RaceKey:  "human",
		ClassKey: "fighter",
	})
	require.NoError(t, err)

	// Find the weapon + shield choice
	var weaponShieldOption *character.ChoiceOption
	for _, choice := range choices.EquipmentChoices {
		for _, opt := range choice.Options {
			if strings.Contains(opt.Name, "shield") && strings.Contains(opt.Name, "weapon") {
				weaponShieldOption = &opt
				break
			}
		}
	}

	// Verify shield is in bundle items
	if weaponShieldOption != nil {
		assert.Contains(t, weaponShieldOption.BundleItems, "shield", "Shield should be in bundle items")
	}

	// 4. Select specific martial weapon and shield (simulating what handler would do)
	_, err = service.UpdateDraftCharacter(ctx, draftChar.ID, &character.UpdateDraftInput{
		Equipment: []string{"warhammer", "shield"},
	})
	require.NoError(t, err)

	// Finalize the character
	finalChar, err := service.FinalizeDraftCharacter(ctx, draftChar.ID)
	require.NoError(t, err)
	require.NotNil(t, finalChar)

	// Verify the character has all expected equipment
	// Check weapons
	weapons, hasWeapons := finalChar.Inventory[entities.EquipmentTypeWeapon]
	assert.True(t, hasWeapons, "Character should have weapons")
	assert.Len(t, weapons, 1, "Should have 1 weapon")
	if len(weapons) > 0 {
		assert.Equal(t, "warhammer", weapons[0].GetKey())
	}

	// Check armor (includes shields)
	armor, hasArmor := finalChar.Inventory[entities.EquipmentTypeArmor]
	assert.True(t, hasArmor, "Character should have armor")
	assert.Len(t, armor, 2, "Should have 2 armor items (chain mail + shield)")

	// Verify we have both chain mail and shield
	hasChainMail := false
	hasShield := false
	for _, item := range armor {
		if item.GetKey() == "chain-mail" {
			hasChainMail = true
		}
		if item.GetKey() == "shield" {
			hasShield = true
			// Verify shield is properly categorized
			if armorItem, ok := item.(*entities.Armor); ok {
				assert.Equal(t, entities.ArmorCategoryShield, armorItem.ArmorCategory,
					"Shield should be categorized as shield")
			}
		}
	}
	assert.True(t, hasChainMail, "Should have chain mail")
	assert.True(t, hasShield, "Should have shield")
}

// Helper to create Fighter class with equipment choices
func createFighterClassWithEquipmentChoices() *entities.Class {
	return &entities.Class{
		Key:    "fighter",
		Name:   "Fighter",
		HitDie: 10,
		StartingEquipmentChoices: []*entities.Choice{
			{
				Name:  "(a) chain mail or (b) leather armor",
				Count: 1,
				Type:  entities.ChoiceTypeEquipment,
			},
			{
				Name:  "(a) a martial weapon and a shield or (b) two martial weapons",
				Count: 1,
				Type:  entities.ChoiceTypeEquipment,
			},
		},
	}
}

// Mock D&D client for testing
type mockDndClient struct {
	equipment map[string]entities.Equipment
	classes   map[string]*entities.Class
	races     map[string]*entities.Race
}

func (m *mockDndClient) ListClasses() ([]*entities.Class, error) {
	var result []*entities.Class
	for _, c := range m.classes {
		result = append(result, c)
	}
	return result, nil
}

func (m *mockDndClient) GetClass(key string) (*entities.Class, error) {
	if c, ok := m.classes[key]; ok {
		return c, nil
	}
	return nil, errors.New("not found")
}

func (m *mockDndClient) GetClassLevel(classKey string, level int) ([]*entities.Feature, error) {
	return []*entities.Feature{}, nil
}

func (m *mockDndClient) ListRaces() ([]*entities.Race, error) {
	var result []*entities.Race
	for _, r := range m.races {
		result = append(result, r)
	}
	return result, nil
}

func (m *mockDndClient) GetRace(key string) (*entities.Race, error) {
	if r, ok := m.races[key]; ok {
		return r, nil
	}
	return nil, errors.New("not found")
}

func (m *mockDndClient) GetEquipment(key string) (entities.Equipment, error) {
	if e, ok := m.equipment[key]; ok {
		return e, nil
	}
	return nil, errors.New("not found")
}

func (m *mockDndClient) GetProficiency(key string) (*entities.Proficiency, error) {
	return &entities.Proficiency{
		Key:  key,
		Name: key,
		Type: entities.ProficiencyTypeSkill,
	}, nil
}

func (m *mockDndClient) GetMonster(key string) (*entities.MonsterTemplate, error) {
	return nil, errors.New("not found")
}

func (m *mockDndClient) GetEquipmentByCategory(category string) ([]entities.Equipment, error) {
	return nil, nil
}

func (m *mockDndClient) ListEquipment() ([]entities.Equipment, error) {
	var result []entities.Equipment
	for _, e := range m.equipment {
		result = append(result, e)
	}
	return result, nil
}

func (m *mockDndClient) ListClassFeatures(classKey string, level int) ([]*entities.Feature, error) {
	return nil, nil
}

func (m *mockDndClient) ListMonstersByCR(minCR, maxCR float32) ([]*entities.MonsterTemplate, error) {
	return nil, nil
}

func (m *mockDndClient) GetClassFeatures(classKey string, level int) ([]*entities.CharacterFeature, error) {
	return nil, nil
}

func strPtr(s string) *string {
	return &s
}
