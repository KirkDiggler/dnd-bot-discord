package abilities

import (
	"context"
	"fmt"
	"testing"

	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/shared"
	mockcharacter "github.com/KirkDiggler/dnd-bot-discord/internal/services/character/mock"
	rpgevents "github.com/KirkDiggler/rpg-toolkit/events"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestRageHandler_Execute(t *testing.T) {
	tests := []struct {
		name           string
		setupChar      func() *character.Character
		setupAbility   func() *shared.ActiveAbility
		input          *Input
		mockSetup      func(*mockcharacter.MockService)
		expectedResult func(*testing.T, *Result)
		expectedChar   func(*testing.T, *character.Character)
	}{
		{
			name: "activate rage successfully",
			setupChar: func() *character.Character {
				char := &character.Character{
					ID:    "test-char",
					Name:  "Grunk",
					Level: 3,
					Resources: &character.CharacterResources{
						Abilities: map[string]*shared.ActiveAbility{
							shared.AbilityKeyRage: {
								Key:           shared.AbilityKeyRage,
								Name:          "Rage",
								UsesRemaining: 2,
								IsActive:      false,
							},
						},
						ActiveEffects: []*shared.ActiveEffect{},
					},
				}
				return char
			},
			setupAbility: func() *shared.ActiveAbility {
				// Return nil - we'll use the one from character.Resources
				return nil
			},
			input: &Input{
				CharacterID: "test-char",
				AbilityKey:  shared.AbilityKeyRage,
			},
			mockSetup: func(m *mockcharacter.MockService) {
				m.EXPECT().UpdateEquipment(gomock.Any()).Return(nil)
			},
			expectedResult: func(t *testing.T, r *Result) {
				assert.True(t, r.Success)
				assert.Contains(t, r.Message, "enters a rage")
				assert.Contains(t, r.Message, "+2 melee damage") // Level 3 = +2 bonus
				assert.Equal(t, 2, r.UsesRemaining)
				assert.True(t, r.EffectApplied)
				assert.Equal(t, "Rage", r.EffectName)
				assert.Equal(t, 10, r.Duration)
				assert.Equal(t, 2, r.DamageBonus)
			},
			expectedChar: func(t *testing.T, c *character.Character) {
				// The ability passed to Execute is the one that gets modified
				// We need to check the character's active effects instead
				assert.True(t, c.Resources.Abilities[shared.AbilityKeyRage].IsActive)
				assert.Equal(t, 10, c.Resources.Abilities[shared.AbilityKeyRage].Duration)

				// Check rage effect was added
				assert.Len(t, c.Resources.ActiveEffects, 1)
				effect := c.Resources.ActiveEffects[0]
				assert.Equal(t, "Rage", effect.Name)
				assert.Equal(t, shared.DurationTypeRounds, effect.DurationType)
				assert.Equal(t, 10, effect.Duration)
			},
		},
		{
			name: "deactivate rage",
			setupChar: func() *character.Character {
				return &character.Character{
					ID:    "test-char",
					Name:  "Grunk",
					Level: 3,
					Resources: &character.CharacterResources{
						Abilities: map[string]*shared.ActiveAbility{
							shared.AbilityKeyRage: {
								Key:           shared.AbilityKeyRage,
								Name:          "Rage",
								UsesRemaining: 1,
								IsActive:      true,
								Duration:      5,
							},
						},
						ActiveEffects: []*shared.ActiveEffect{
							{
								Name: "Rage",
							},
						},
					},
				}
			},
			setupAbility: func() *shared.ActiveAbility {
				// Return nil - we'll use the one from character.Resources
				return nil
			},
			input: &Input{
				CharacterID: "test-char",
				AbilityKey:  shared.AbilityKeyRage,
			},
			mockSetup: func(m *mockcharacter.MockService) {
				m.EXPECT().UpdateEquipment(gomock.Any()).Return(nil)
			},
			expectedResult: func(t *testing.T, r *Result) {
				assert.True(t, r.Success)
				assert.Contains(t, r.Message, "no longer raging")
				assert.Equal(t, 1, r.UsesRemaining)
				assert.False(t, r.EffectApplied)
			},
			expectedChar: func(t *testing.T, c *character.Character) {
				ability := c.Resources.Abilities[shared.AbilityKeyRage]
				assert.False(t, ability.IsActive)
				assert.Equal(t, 0, ability.Duration)

				// Check rage effect was removed
				assert.Len(t, c.Resources.ActiveEffects, 0)
			},
		},
		{
			name: "no uses remaining",
			setupChar: func() *character.Character {
				return &character.Character{
					ID:    "test-char",
					Name:  "Grunk",
					Level: 3,
					Resources: &character.CharacterResources{
						Abilities: map[string]*shared.ActiveAbility{
							shared.AbilityKeyRage: {
								Key:           shared.AbilityKeyRage,
								Name:          "Rage",
								UsesRemaining: 0,
								IsActive:      false,
							},
						},
					},
				}
			},
			setupAbility: func() *shared.ActiveAbility {
				// Return nil - we'll use the one from character.Resources
				return nil
			},
			input: &Input{
				CharacterID: "test-char",
				AbilityKey:  shared.AbilityKeyRage,
			},
			mockSetup: func(m *mockcharacter.MockService) {
				// No UpdateEquipment call expected
			},
			expectedResult: func(t *testing.T, r *Result) {
				assert.False(t, r.Success)
				assert.Contains(t, r.Message, "No rage uses remaining")
				assert.Equal(t, 0, r.UsesRemaining)
			},
			expectedChar: func(t *testing.T, c *character.Character) {
				ability := c.Resources.Abilities[shared.AbilityKeyRage]
				assert.False(t, ability.IsActive)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockCharService := mockcharacter.NewMockService(ctrl)
			eventBus := rpgevents.NewBus()

			handler := NewRageHandler(eventBus, mockCharService)
			char := tt.setupChar()
			ability := char.Resources.Abilities[shared.AbilityKeyRage]

			tt.mockSetup(mockCharService)

			result, err := handler.Execute(context.Background(), char, ability, tt.input)
			require.NoError(t, err)

			tt.expectedResult(t, result)
			tt.expectedChar(t, char)
		})
	}
}

func TestRageHandler_DamageBonus(t *testing.T) {
	tests := []struct {
		level    int
		expected int
	}{
		{1, 2},
		{5, 2},
		{8, 2},
		{9, 3},
		{15, 3},
		{16, 4},
		{20, 4},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("level_%d", tt.level), func(t *testing.T) {
			handler := &RageHandler{}
			bonus := handler.getRageDamageBonus(tt.level)
			assert.Equal(t, tt.expected, bonus)
		})
	}
}
