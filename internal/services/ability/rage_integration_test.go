package ability_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/rulebook/dnd5e"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/rulebook/dnd5e/abilities"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/shared"
	"github.com/KirkDiggler/dnd-bot-discord/internal/services/ability"
	mockcharacter "github.com/KirkDiggler/dnd-bot-discord/internal/services/character/mock"
	rpgevents "github.com/KirkDiggler/rpg-toolkit/events"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestRageAbilityIntegration(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup
	mockCharService := mockcharacter.NewMockService(ctrl)
	eventBus := rpgevents.NewBus()

	// Create ability service
	svc := ability.NewService(&ability.ServiceConfig{
		CharacterService: mockCharService,
		EventBus:         eventBus,
	})

	// Register rage handler
	rageHandler := abilities.NewRageHandler(eventBus, mockCharService)
	svc.RegisterHandler(abilities.NewServiceHandlerAdapter(rageHandler))

	// Create a barbarian character
	barbarian := &character.Character{
		ID:    "barb-1",
		Name:  "Grunk",
		Level: 5, // Level 5 for testing
		Class: &rulebook.Class{
			Key:  "barbarian",
			Name: "Barbarian",
		},
		MaxHitPoints:     45,
		CurrentHitPoints: 45,
		Resources: &character.CharacterResources{
			Abilities: map[string]*shared.ActiveAbility{
				shared.AbilityKeyRage: {
					Key:           shared.AbilityKeyRage,
					Name:          "Rage",
					ActionType:    "bonus_action",
					UsesMax:       3,
					UsesRemaining: 3,
					Duration:      0,
					IsActive:      false,
				},
			},
			ActiveEffects: []*shared.ActiveEffect{},
		},
	}

	// Mock expectations
	mockCharService.EXPECT().GetByID("barb-1").Return(barbarian, nil).Times(1)
	mockCharService.EXPECT().UpdateEquipment(gomock.Any()).Return(nil).Times(2) // Called by service and handler

	// Test 1: Activate rage
	ctx := context.Background()
	result, err := svc.UseAbility(ctx, &ability.UseAbilityInput{
		CharacterID: "barb-1",
		AbilityKey:  shared.AbilityKeyRage,
	})
	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Contains(t, result.Message, "enters a rage")
	assert.Contains(t, result.Message, "+2 melee damage") // Level 5 = +2 bonus
	assert.Equal(t, 2, result.UsesRemaining)              // Used 1 of 3

	// Verify rage is active
	assert.True(t, barbarian.Resources.Abilities[shared.AbilityKeyRage].IsActive)
	assert.Equal(t, 10, barbarian.Resources.Abilities[shared.AbilityKeyRage].Duration)

	// Verify rage effect was added
	assert.Len(t, barbarian.Resources.ActiveEffects, 1)
	rageEffect := barbarian.Resources.ActiveEffects[0]
	assert.Equal(t, "Rage", rageEffect.Name)
	assert.Equal(t, shared.DurationTypeRounds, rageEffect.DurationType)
	assert.Equal(t, 10, rageEffect.Duration)

	// Test 2: Deactivate rage
	mockCharService.EXPECT().GetByID("barb-1").Return(barbarian, nil).Times(1)
	mockCharService.EXPECT().UpdateEquipment(gomock.Any()).Return(nil).Times(2) // Called by service and handler

	result2, err := svc.UseAbility(ctx, &ability.UseAbilityInput{
		CharacterID: "barb-1",
		AbilityKey:  shared.AbilityKeyRage,
	})
	require.NoError(t, err)
	assert.True(t, result2.Success)
	assert.Contains(t, result2.Message, "no longer raging")
	assert.Equal(t, 1, result2.UsesRemaining) // Deactivating still uses the ability service's charge

	// Verify rage is inactive
	assert.False(t, barbarian.Resources.Abilities[shared.AbilityKeyRage].IsActive)
	assert.Equal(t, 0, barbarian.Resources.Abilities[shared.AbilityKeyRage].Duration)

	// Verify rage effect was removed
	assert.Len(t, barbarian.Resources.ActiveEffects, 0)
}

func TestRageDamageBonusIntegration(t *testing.T) {
	tests := []struct {
		name     string
		level    int
		expected int
	}{
		{"level 1", 1, 2},
		{"level 8", 8, 2},
		{"level 9", 9, 3},
		{"level 15", 15, 3},
		{"level 16", 16, 4},
		{"level 20", 20, 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockCharService := mockcharacter.NewMockService(ctrl)
			eventBus := rpgevents.NewBus()

			svc := ability.NewService(&ability.ServiceConfig{
				CharacterService: mockCharService,
				EventBus:         eventBus,
			})

			rageHandler := abilities.NewRageHandler(eventBus, mockCharService)
			svc.RegisterHandler(abilities.NewServiceHandlerAdapter(rageHandler))

			barbarian := &character.Character{
				ID:    "barb-1",
				Name:  "TestBarb",
				Level: tt.level,
				Resources: &character.CharacterResources{
					Abilities: map[string]*shared.ActiveAbility{
						shared.AbilityKeyRage: {
							Key:           shared.AbilityKeyRage,
							Name:          "Rage",
							UsesRemaining: 3,
							IsActive:      false,
						},
					},
					ActiveEffects: []*shared.ActiveEffect{},
				},
			}

			mockCharService.EXPECT().GetByID("barb-1").Return(barbarian, nil)
			mockCharService.EXPECT().UpdateEquipment(gomock.Any()).Return(nil).Times(2)

			result, err := svc.UseAbility(context.Background(), &ability.UseAbilityInput{
				CharacterID: "barb-1",
				AbilityKey:  shared.AbilityKeyRage,
			})
			require.NoError(t, err)
			assert.True(t, result.Success)
			assert.Contains(t, result.Message, fmt.Sprintf("+%d melee damage", tt.expected))
		})
	}
}
