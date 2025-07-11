package ability

import (
	"context"
	"testing"

	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/shared"

	mockdice "github.com/KirkDiggler/dnd-bot-discord/internal/dice/mock"
	"github.com/KirkDiggler/dnd-bot-discord/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	mockchar "github.com/KirkDiggler/dnd-bot-discord/internal/services/character/mock"
)

func TestAbilityService_UseAbility(t *testing.T) {
	tests := []struct {
		name         string
		setupMocks   func(*mockchar.MockService, *mockdice.ManualMockRoller)
		input        *UseAbilityInput
		character    *character.Character
		wantResult   *UseAbilityResult
		wantErr      bool
		validateChar func(*testing.T, *character.Character)
	}{
		{
			name: "barbarian rage success",
			setupMocks: func(charSvc *mockchar.MockService, roller *mockdice.ManualMockRoller) {
				// No dice rolls needed for rage
			},
			input: &UseAbilityInput{
				CharacterID: "char_123",
				AbilityKey:  "rage",
			},
			character: createTestCharacter("barbarian", 1, map[string]int{"STR": 16}),
			wantResult: &UseAbilityResult{
				Success:       true,
				UsesRemaining: 1,           // Started with 2, used 1
				Message:       "Used Rage", // Generic message when no handler registered
				EffectApplied: false,
				EffectName:    "",
				Duration:      0,
			},
			validateChar: func(t *testing.T, char *character.Character) {
				// Check that the ability was used
				rage := char.Resources.Abilities["rage"]
				assert.Equal(t, 1, rage.UsesRemaining)
				// No effects should be added without a handler
			},
		},
		{
			name: "fighter second wind healing",
			setupMocks: func(charSvc *mockchar.MockService, roller *mockdice.ManualMockRoller) {
				// Mock healing roll: 1d10+1 = 7
				roller.SetRolls([]int{6})
			},
			input: &UseAbilityInput{
				CharacterID: "char_123",
				AbilityKey:  "second-wind",
			},
			character: func() *character.Character {
				char := createTestCharacter("fighter", 1, map[string]int{"CON": 14})
				char.Resources.HP.Current = 5 // Damaged
				char.CurrentHitPoints = 5     // Also set this to match
				return char
			}(),
			wantResult: &UseAbilityResult{
				Success:       true,
				UsesRemaining: 0,                  // Used the only use
				Message:       "Used Second Wind", // Generic message when no handler registered
				HealingDone:   0,
				TargetNewHP:   0,
			},
			validateChar: func(t *testing.T, char *character.Character) {
				// Without handler, no healing occurs
				assert.Equal(t, 5, char.Resources.HP.Current)
				assert.Equal(t, 5, char.CurrentHitPoints)
			},
		},
		{
			name: "bard bardic inspiration",
			setupMocks: func(charSvc *mockchar.MockService, roller *mockdice.ManualMockRoller) {
				// No dice rolls for giving inspiration
			},
			input: &UseAbilityInput{
				CharacterID: "char_123",
				AbilityKey:  "bardic-inspiration",
				TargetID:    "ally_456",
			},
			character: createTestCharacter("bard", 1, map[string]int{"CHA": 16}),
			wantResult: &UseAbilityResult{
				Success:       true,
				UsesRemaining: 2,                         // Started with 3 (CHA mod), used 1
				Message:       "Used Bardic Inspiration", // Generic message when no handler registered
				EffectApplied: false,
				EffectName:    "",
				Duration:      0,
			},
		},
		{
			name: "paladin lay on hands self heal",
			setupMocks: func(charSvc *mockchar.MockService, roller *mockdice.ManualMockRoller) {
				// No dice rolls for lay on hands
			},
			input: &UseAbilityInput{
				CharacterID: "char_123",
				AbilityKey:  "lay-on-hands",
				Value:       3, // Heal 3 HP
			},
			character: func() *character.Character {
				char := createTestCharacter("paladin", 1, map[string]int{"CHA": 14})
				char.Resources.HP.Current = 7 // Damaged
				return char
			}(),
			wantResult: &UseAbilityResult{
				Success:       true,
				UsesRemaining: 4,                   // Started with 5, but it still uses 1 even without handler (different from pool usage)
				Message:       "Used Lay on Hands", // Generic message when no handler registered
				HealingDone:   0,
				TargetNewHP:   0,
			},
			validateChar: func(t *testing.T, char *character.Character) {
				// Without handler, no healing occurs
				assert.Equal(t, 7, char.Resources.HP.Current)
				// The ability still uses 1 resource even without handler
				assert.Equal(t, 4, char.Resources.Abilities["lay-on-hands"].UsesRemaining)
			},
		},
		{
			name: "paladin divine sense",
			setupMocks: func(charSvc *mockchar.MockService, roller *mockdice.ManualMockRoller) {
				// No dice rolls for divine sense
			},
			input: &UseAbilityInput{
				CharacterID: "char_123",
				AbilityKey:  "divine-sense",
			},
			character: createTestCharacter("paladin", 1, map[string]int{"CHA": 14}),
			wantResult: &UseAbilityResult{
				Success:       true,
				UsesRemaining: 2,                   // Started with 3 (1 + CHA mod), used 1
				Message:       "Used Divine Sense", // Generic message when no handler registered
				EffectApplied: false,
				EffectName:    "",
				Duration:      0,
			},
		},
		{
			name:       "no uses remaining",
			setupMocks: func(charSvc *mockchar.MockService, roller *mockdice.ManualMockRoller) {},
			input: &UseAbilityInput{
				CharacterID: "char_123",
				AbilityKey:  "second-wind",
			},
			character: func() *character.Character {
				char := createTestCharacter("fighter", 1, nil)
				char.Resources.Abilities["second-wind"].UsesRemaining = 0
				return char
			}(),
			wantResult: &UseAbilityResult{
				Success:       false,
				Message:       "No uses remaining",
				UsesRemaining: 0,
			},
		},
		{
			name:       "ability not found",
			setupMocks: func(charSvc *mockchar.MockService, roller *mockdice.ManualMockRoller) {},
			input: &UseAbilityInput{
				CharacterID: "char_123",
				AbilityKey:  "fireball", // Not a level 1 ability
			},
			character: createTestCharacter("fighter", 1, nil),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockCharSvc := mockchar.NewMockService(ctrl)
			mockRoller := mockdice.NewManualMockRoller()

			// Setup character service to return our test character
			mockCharSvc.EXPECT().
				GetByID(tt.input.CharacterID).
				Return(tt.character, nil)

			// Allow UpdateEquipment to be called
			mockCharSvc.EXPECT().
				UpdateEquipment(gomock.Any()).
				Return(nil).
				AnyTimes()

			tt.setupMocks(mockCharSvc, mockRoller)

			svc := NewService(&ServiceConfig{
				CharacterService: mockCharSvc,
				DiceRoller:       mockRoller,
			})

			result, err := svc.UseAbility(context.Background(), tt.input)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)

			assert.Equal(t, tt.wantResult.Success, result.Success)
			assert.Equal(t, tt.wantResult.Message, result.Message)
			assert.Equal(t, tt.wantResult.UsesRemaining, result.UsesRemaining)
			assert.Equal(t, tt.wantResult.HealingDone, result.HealingDone)
			assert.Equal(t, tt.wantResult.TargetNewHP, result.TargetNewHP)
			assert.Equal(t, tt.wantResult.EffectApplied, result.EffectApplied)
			assert.Equal(t, tt.wantResult.Duration, result.Duration)

			if tt.validateChar != nil {
				tt.validateChar(t, tt.character)
			}
		})
	}
}

func TestAbilityService_GetAvailableAbilities(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCharSvc := mockchar.NewMockService(ctrl)

	char := createTestCharacter("fighter", 1, nil)
	// Set second wind to no uses
	char.Resources.Abilities["second-wind"].UsesRemaining = 0

	mockCharSvc.EXPECT().
		GetByID("char_123").
		Return(char, nil)

	svc := NewService(&ServiceConfig{
		CharacterService: mockCharSvc,
	})

	abilities, err := svc.GetAvailableAbilities(context.Background(), "char_123")
	require.NoError(t, err)
	require.Len(t, abilities, 1) // Fighter only has second wind at level 1

	secondWind := abilities[0]
	assert.Equal(t, "second-wind", secondWind.Ability.Key)
	assert.False(t, secondWind.Available)
	assert.Equal(t, "No uses remaining", secondWind.Reason)
}

// Helper function to create test characters
func createTestCharacter(classKey string, level int, attributes map[string]int) *character.Character {
	char := &character.Character{
		ID:               "char_123",
		Name:             "Test Character",
		Level:            level,
		Class:            testutils.CreateTestClass(classKey, classKey, 10),
		MaxHitPoints:     10 + 2*level, // Simplified HP
		CurrentHitPoints: 10 + 2*level,
	}

	// Set attributes if provided
	if attributes != nil {
		char.Attributes = make(map[shared.Attribute]*character.AbilityScore)
		for attr, score := range attributes {
			var attribute shared.Attribute
			switch attr {
			case "STR":
				attribute = shared.AttributeStrength
			case "DEX":
				attribute = shared.AttributeDexterity
			case "CON":
				attribute = shared.AttributeConstitution
			case "INT":
				attribute = shared.AttributeIntelligence
			case "WIS":
				attribute = shared.AttributeWisdom
			case "CHA":
				attribute = shared.AttributeCharisma
			}
			char.AddAttribute(attribute, score)
		}
	}

	// Initialize resources to get abilities
	char.InitializeResources()

	return char
}
