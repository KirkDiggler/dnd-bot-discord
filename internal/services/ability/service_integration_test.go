//go:build integration
// +build integration

package ability

import (
	"context"
	character2 "github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/damage"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/rulebook"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/shared"
	"testing"

	"github.com/KirkDiggler/dnd-bot-discord/internal/dice"
	"github.com/KirkDiggler/dnd-bot-discord/internal/repositories/characters"
	"github.com/KirkDiggler/dnd-bot-discord/internal/services/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/testutils"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type AbilityServiceIntegrationSuite struct {
	suite.Suite
	ctx            context.Context
	redisClient    redis.UniversalClient
	charRepo       characters.Repository
	charService    character.Service
	abilityService Service
}

func (s *AbilityServiceIntegrationSuite) SetupSuite() {
	s.ctx = context.Background()

	// Setup Redis
	s.redisClient = testutils.CreateTestRedisClientOrSkip(s.T())

	// Create repositories
	s.charRepo = characters.NewRedis(s.redisClient)

	// Create services
	s.charService = character.NewService(&character.ServiceConfig{
		Repository: s.charRepo,
	})

	s.abilityService = NewService(&ServiceConfig{
		CharacterService: s.charService,
		DiceRoller:       dice.NewRandomRoller(),
	})
}

func (s *AbilityServiceIntegrationSuite) SetupTest() {
	// Clean Redis before each test
	err := s.redisClient.FlushDB(s.ctx).Err()
	require.NoError(s.T(), err)
}

func (s *AbilityServiceIntegrationSuite) TestRageAbilityFullCycle() {
	// Create a level 1 barbarian
	char := &character2.Character{
		ID:               "barb_123",
		OwnerID:          "player_123",
		RealmID:          "realm_123",
		Name:             "Grunk",
		Class:            &rulebook.Class{Key: "barbarian", Name: "Barbarian"},
		Level:            1,
		Status:           character2.CharacterStatusActive,
		CurrentHitPoints: 12,
		MaxHitPoints:     12,
		Attributes: map[character2.Attribute]*character2.AbilityScore{
			character2.AttributeStrength:     {Score: 16},
			character2.AttributeDexterity:    {Score: 14},
			character2.AttributeConstitution: {Score: 15},
			character2.AttributeIntelligence: {Score: 8},
			character2.AttributeWisdom:       {Score: 12},
			character2.AttributeCharisma:     {Score: 10},
		},
		Resources: &shared.CharacterResources{
			Abilities: map[string]*character2.ActiveAbility{
				"rage": {
					Name:          "Rage",
					Key:           "rage",
					Description:   "Enter a battle fury",
					ActionType:    character2.AbilityTypeBonusAction,
					UsesMax:       2,
					UsesRemaining: 2,
					RestType:      character2.RestTypeLong,
					IsActive:      false,
					Duration:      0,
				},
			},
			ActiveEffects: []*shared.ActiveEffect{},
		},
	}

	// Save character to Redis
	err := s.charRepo.Create(s.ctx, char)
	require.NoError(s.T(), err)

	// Test 1: Activate rage
	result, err := s.abilityService.UseAbility(s.ctx, &UseAbilityInput{
		CharacterID: char.ID,
		AbilityKey:  "rage",
	})

	require.NoError(s.T(), err)
	assert.True(s.T(), result.Success)
	assert.Equal(s.T(), 1, result.UsesRemaining)
	assert.True(s.T(), result.EffectApplied)
	assert.Equal(s.T(), "Rage", result.EffectName)
	assert.Equal(s.T(), 10, result.Duration)

	// Verify character state was persisted
	loadedChar, err := s.charRepo.Get(s.ctx, char.ID)
	require.NoError(s.T(), err)

	// Check rage is active
	rage := loadedChar.Resources.Abilities["rage"]
	assert.True(s.T(), rage.IsActive)
	assert.Equal(s.T(), 1, rage.UsesRemaining)
	assert.Equal(s.T(), 10, rage.Duration)

	// Check active effect exists
	assert.Len(s.T(), loadedChar.Resources.ActiveEffects, 1)
	effect := loadedChar.Resources.ActiveEffects[0]
	assert.Equal(s.T(), "Rage", effect.Name)

	// Check damage bonus modifier
	assert.Equal(s.T(), 2, effect.GetDamageBonus(""))

	// Check resistance modifier
	assert.True(s.T(), effect.HasResistance(string(damage.TypeBludgeoning)))

	// Test 2: Try to activate rage again (should toggle off)
	result2, err := s.abilityService.UseAbility(s.ctx, &UseAbilityInput{
		CharacterID: char.ID,
		AbilityKey:  "rage",
	})

	require.NoError(s.T(), err)
	assert.True(s.T(), result2.Success)
	assert.Contains(s.T(), result2.Message, "no longer raging")

	// Verify rage is deactivated
	loadedChar2, err := s.charRepo.Get(s.ctx, char.ID)
	require.NoError(s.T(), err)

	rage2 := loadedChar2.Resources.Abilities["rage"]
	assert.False(s.T(), rage2.IsActive)
	assert.Empty(s.T(), loadedChar2.Resources.ActiveEffects)
}

func (s *AbilityServiceIntegrationSuite) TestSecondWindHealing() {
	// Create a level 3 fighter
	char := &character2.Character{
		ID:               "fighter_123",
		OwnerID:          "player_123",
		RealmID:          "realm_123",
		Name:             "Aldric",
		Class:            &rulebook.Class{Key: "fighter", Name: "Fighter"},
		Level:            3,
		Status:           character2.CharacterStatusActive,
		CurrentHitPoints: 15, // Damaged
		MaxHitPoints:     28,
		Attributes: map[character2.Attribute]*character2.AbilityScore{
			character2.AttributeStrength:     {Score: 16},
			character2.AttributeDexterity:    {Score: 13},
			character2.AttributeConstitution: {Score: 14},
			character2.AttributeIntelligence: {Score: 10},
			character2.AttributeWisdom:       {Score: 12},
			character2.AttributeCharisma:     {Score: 8},
		},
		Resources: &shared.CharacterResources{
			Abilities: map[string]*character2.ActiveAbility{
				"second_wind": {
					Name:          "Second Wind",
					Key:           "second_wind",
					Description:   "Regain hit points",
					ActionType:    character2.AbilityTypeBonusAction,
					UsesMax:       1,
					UsesRemaining: 1,
					RestType:      character2.RestTypeShort,
				},
			},
		},
	}

	// Save character
	err := s.charRepo.Create(s.ctx, char)
	require.NoError(s.T(), err)

	// Use Second Wind
	result, err := s.abilityService.UseAbility(s.ctx, &UseAbilityInput{
		CharacterID: char.ID,
		AbilityKey:  "second_wind",
	})

	require.NoError(s.T(), err)
	assert.True(s.T(), result.Success)
	assert.Equal(s.T(), 0, result.UsesRemaining)
	assert.Contains(s.T(), result.Message, "regain")
	assert.Contains(s.T(), result.Message, "hit points")

	// Verify HP was increased and persisted
	loadedChar, err := s.charRepo.Get(s.ctx, char.ID)
	require.NoError(s.T(), err)
	assert.Greater(s.T(), loadedChar.CurrentHitPoints, 15)
	assert.LessOrEqual(s.T(), loadedChar.CurrentHitPoints, loadedChar.MaxHitPoints)

	// Verify ability was consumed
	assert.Equal(s.T(), 0, loadedChar.Resources.Abilities["second_wind"].UsesRemaining)
}

func (s *AbilityServiceIntegrationSuite) TestMultipleRageUses() {
	// Test using rage multiple times
	char := &character2.Character{
		ID:               "multi_123",
		OwnerID:          "player_123",
		RealmID:          "realm_123",
		Name:             "RageTest",
		Class:            &rulebook.Class{Key: "barbarian", Name: "Barbarian"},
		Level:            1,
		Status:           character2.CharacterStatusActive,
		CurrentHitPoints: 12,
		MaxHitPoints:     12,
		Attributes: map[character2.Attribute]*character2.AbilityScore{
			character2.AttributeStrength:     {Score: 16},
			character2.AttributeDexterity:    {Score: 14},
			character2.AttributeConstitution: {Score: 15},
			character2.AttributeIntelligence: {Score: 8},
			character2.AttributeWisdom:       {Score: 12},
			character2.AttributeCharisma:     {Score: 10},
		},
		Resources: &shared.CharacterResources{
			Abilities: map[string]*character2.ActiveAbility{
				"rage": {
					Name:          "Rage",
					Key:           "rage",
					ActionType:    character2.AbilityTypeBonusAction,
					UsesMax:       2,
					UsesRemaining: 2,
					RestType:      character2.RestTypeLong,
				},
			},
		},
	}

	err := s.charRepo.Create(s.ctx, char)
	require.NoError(s.T(), err)

	// Use rage first time
	result1, err := s.abilityService.UseAbility(s.ctx, &UseAbilityInput{
		CharacterID: char.ID,
		AbilityKey:  "rage",
	})
	require.NoError(s.T(), err)
	assert.True(s.T(), result1.Success)
	assert.Equal(s.T(), 1, result1.UsesRemaining)

	// Deactivate rage
	result2, err := s.abilityService.UseAbility(s.ctx, &UseAbilityInput{
		CharacterID: char.ID,
		AbilityKey:  "rage",
	})
	require.NoError(s.T(), err)
	assert.True(s.T(), result2.Success)

	// Use rage second time
	result3, err := s.abilityService.UseAbility(s.ctx, &UseAbilityInput{
		CharacterID: char.ID,
		AbilityKey:  "rage",
	})
	require.NoError(s.T(), err)
	assert.True(s.T(), result3.Success)
	assert.Equal(s.T(), 0, result3.UsesRemaining)

	// Try to use rage when out of uses (should fail)
	// First deactivate current rage
	_, err = s.abilityService.UseAbility(s.ctx, &UseAbilityInput{
		CharacterID: char.ID,
		AbilityKey:  "rage",
	})
	require.NoError(s.T(), err)

	// Now try to activate again
	result4, err := s.abilityService.UseAbility(s.ctx, &UseAbilityInput{
		CharacterID: char.ID,
		AbilityKey:  "rage",
	})
	require.NoError(s.T(), err)
	assert.False(s.T(), result4.Success)
	assert.Contains(s.T(), result4.Message, "no uses remaining")
}

func (s *AbilityServiceIntegrationSuite) TestConcurrentAbilityUse() {
	// Test that concurrent ability uses don't cause race conditions
	char := &character2.Character{
		ID:               "concurrent_123",
		OwnerID:          "player_123",
		RealmID:          "realm_123",
		Name:             "Concurrent",
		Class:            &rulebook.Class{Key: "barbarian", Name: "Barbarian"},
		Level:            1,
		Status:           character2.CharacterStatusActive,
		CurrentHitPoints: 12,
		MaxHitPoints:     12,
		Resources: &shared.CharacterResources{
			Abilities: map[string]*character2.ActiveAbility{
				"rage": {
					Name:          "Rage",
					Key:           "rage",
					ActionType:    character2.AbilityTypeBonusAction,
					UsesMax:       2,
					UsesRemaining: 2,
					RestType:      character2.RestTypeLong,
				},
			},
		},
	}

	err := s.charRepo.Create(s.ctx, char)
	require.NoError(s.T(), err)

	// Run multiple concurrent ability uses
	done := make(chan bool, 3)
	errors := make(chan error, 3)

	for i := 0; i < 3; i++ {
		go func() {
			_, abilityErr := s.abilityService.UseAbility(s.ctx, &UseAbilityInput{
				CharacterID: char.ID,
				AbilityKey:  "rage",
			})
			if abilityErr != nil {
				errors <- abilityErr
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 3; i++ {
		<-done
	}
	close(errors)

	// Check for errors
	for err := range errors {
		assert.NoError(s.T(), err)
	}

	// Verify final state is consistent
	loadedChar, err := s.charRepo.Get(s.ctx, char.ID)
	require.NoError(s.T(), err)

	// Should have used at least one rage
	assert.Less(s.T(), loadedChar.Resources.Abilities["rage"].UsesRemaining, 2)
}

func TestAbilityServiceIntegration(t *testing.T) {
	suite.Run(t, new(AbilityServiceIntegrationSuite))
}
