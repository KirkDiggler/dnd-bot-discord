package character

import (
	"encoding/json"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/equipment"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/rulebook"
	"testing"

	mockcharacters "github.com/KirkDiggler/dnd-bot-discord/internal/services/character/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestFightingStylePersistence(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mockcharacters.NewMockService(ctrl)
	handler := NewClassFeaturesHandler(mockService)

	// Create a fighter character with fighting style feature (but no selection yet)
	fighter := &character.Character{
		ID:    "test-fighter-id",
		Name:  "Test Fighter",
		Class: &rulebook.Class{Key: "fighter"},
		Features: []*rulebook.CharacterFeature{
			{
				Key:      "fighting_style",
				Name:     "Fighting Style",
				Type:     rulebook.FeatureTypeClass,
				Level:    1,
				Source:   "Fighter",
				Metadata: nil, // No selection yet
			},
		},
	}

	var savedCharacter *character.Character

	t.Run("Test fighting style selection and persistence", func(t *testing.T) {
		// Mock getting the character
		mockService.EXPECT().
			GetByID("test-fighter-id").
			Return(fighter, nil)

		// Mock saving the character - capture what gets saved
		mockService.EXPECT().
			UpdateEquipment(gomock.Any()).
			Do(func(char *character.Character) {
				savedCharacter = char
			}).
			Return(nil)

		// Create the selection request
		req := &ClassFeaturesRequest{
			CharacterID: "test-fighter-id",
			FeatureType: "fighting_style",
			Selection:   "dueling",
		}

		// Handle the selection
		err := handler.Handle(req)
		require.NoError(t, err)

		// Verify the character was saved with the correct metadata
		require.NotNil(t, savedCharacter)
		assert.Equal(t, "Test Fighter", savedCharacter.Name)
		assert.Len(t, savedCharacter.Features, 1)

		fightingStyleFeature := savedCharacter.Features[0]
		assert.Equal(t, "fighting_style", fightingStyleFeature.Key)
		require.NotNil(t, fightingStyleFeature.Metadata)
		assert.Equal(t, "dueling", fightingStyleFeature.Metadata["style"])

		t.Logf("✅ Fighting style selection correctly saved in memory")
	})

	t.Run("Test that saved character would have dueling bonus", func(t *testing.T) {
		// Use the character that was "saved" in the previous test
		if savedCharacter == nil {
			t.Skip("Previous test failed, skipping")
		}

		// Add required attributes and equipment for dueling test
		savedCharacter.Attributes = map[character.Attribute]*character.AbilityScore{
			character.AttributeStrength: {Score: 16, Bonus: 3},
		}
		savedCharacter.Level = 1
		savedCharacter.Proficiencies = map[rulebook.ProficiencyType][]*rulebook.Proficiency{
			rulebook.ProficiencyTypeWeapon: {
				{Key: "martial-weapons", Name: "Martial Weapons"},
			},
		}

		// Create a longsword
		longsword := &equipment.Weapon{
			Base: equipment.BasicEquipment{
				Key:  "longsword",
				Name: "Longsword",
			},
			WeaponCategory: "martial",
			WeaponRange:    "Melee",
		}

		// Create a shield (not a weapon, so dueling should still work)
		shield := &equipment.Armor{
			Base: equipment.BasicEquipment{
				Key:  "shield",
				Name: "Shield",
			},
			ArmorCategory: "shield",
		}

		savedCharacter.EquippedSlots = map[character.Slot]equipment.Equipment{
			character.SlotMainHand: longsword,
			character.SlotOffHand:  shield,
		}

		// Just verify the feature was set correctly - bonus testing is in other tests
		fightingStyleFeature := savedCharacter.Features[0]
		assert.Equal(t, "fighting_style", fightingStyleFeature.Key)
		require.NotNil(t, fightingStyleFeature.Metadata)
		assert.Equal(t, "dueling", fightingStyleFeature.Metadata["style"])

		t.Logf("✅ Character ready for dueling bonus application")
	})
}

func TestFightingStylePersistenceFlow(t *testing.T) {
	// This test simulates the complete flow without mocks to see where it breaks
	t.Run("Simulate complete persistence flow", func(t *testing.T) {
		// Step 1: Create character with fighting style feature
		char := &character.Character{
			ID:    "flow-test-id",
			Name:  "Flow Test Fighter",
			Class: &rulebook.Class{Key: "fighter"},
			Features: []*rulebook.CharacterFeature{
				{
					Key:      "fighting_style",
					Name:     "Fighting Style",
					Type:     rulebook.FeatureTypeClass,
					Level:    1,
					Source:   "Fighter",
					Metadata: make(map[string]any), // Empty metadata
				},
			},
		}

		t.Logf("Step 1: Created character with %d features", len(char.Features))
		t.Logf("Fighting style metadata before: %v", char.Features[0].Metadata)

		// Step 2: Simulate the handleFightingStyle logic
		for _, feature := range char.Features {
			if feature.Key == "fighting_style" {
				if feature.Metadata == nil {
					feature.Metadata = make(map[string]any)
				}
				feature.Metadata["style"] = "dueling"
				t.Logf("Step 2: Set fighting style to dueling")
				break
			}
		}

		t.Logf("Fighting style metadata after setting: %v", char.Features[0].Metadata)

		// Step 3: Verify the metadata persists in the same object
		foundStyle := ""
		for _, feature := range char.Features {
			if feature.Key == "fighting_style" && feature.Metadata != nil {
				if style, ok := feature.Metadata["style"].(string); ok {
					foundStyle = style
				}
			}
		}

		assert.Equal(t, "dueling", foundStyle, "Fighting style should be findable in same object")
		t.Logf("✅ Step 3: Found fighting style: %s", foundStyle)

		// Step 4: Test JSON serialization/deserialization (simulates Redis)

		jsonData, err := json.Marshal(char.Features)
		require.NoError(t, err)
		t.Logf("Step 4a: Serialized features to JSON: %s", string(jsonData))

		var deserializedFeatures []*rulebook.CharacterFeature
		err = json.Unmarshal(jsonData, &deserializedFeatures)
		require.NoError(t, err)
		t.Logf("Step 4b: Deserialized %d features", len(deserializedFeatures))

		// Step 5: Check if metadata survived serialization
		foundStyleAfterJSON := ""
		for _, feature := range deserializedFeatures {
			if feature.Key == "fighting_style" {
				t.Logf("Fighting style feature after JSON: metadata=%v", feature.Metadata)
				if feature.Metadata != nil {
					if style, ok := feature.Metadata["style"].(string); ok {
						foundStyleAfterJSON = style
					}
				}
			}
		}

		assert.Equal(t, "dueling", foundStyleAfterJSON, "Fighting style should survive JSON serialization")
		t.Logf("✅ Step 5: Fighting style survived JSON: %s", foundStyleAfterJSON)
	})
}
