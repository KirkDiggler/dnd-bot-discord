// These tests are for the future sneak attack implementation
// They follow TDD principles - write tests first, see them fail, then implement

package entities

import (
	character2 "github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/damage"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/equipment"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/game/combat/attack"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/rulebook"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/shared"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCharacter_CanSneakAttack(t *testing.T) {
	tests := []struct {
		name            string
		character       *character2.Character
		weapon          *equipment.Weapon
		hasAdvantage    bool
		allyAdjacent    bool
		hasDisadvantage bool
		wantEligible    bool
	}{
		{
			name: "finesse weapon with advantage",
			character: &character2.Character{
				Class: &rulebook.Class{Key: "rogue", Name: "Rogue"},
				Level: 1,
			},
			weapon: &equipment.Weapon{
				Base: equipment.BasicEquipment{
					Key:  "dagger",
					Name: "Dagger",
				},
				Properties: []*shared.ReferenceItem{{Key: "finesse", Name: "Finesse"}},
			},
			hasAdvantage: true,
			wantEligible: true,
		},
		{
			name: "ranged weapon with adjacent ally",
			character: &character2.Character{
				Class: &rulebook.Class{Key: "rogue", Name: "Rogue"},
				Level: 1,
			},
			weapon: &equipment.Weapon{
				Base: equipment.BasicEquipment{
					Key:  "shortbow",
					Name: "Shortbow",
				},
				WeaponRange: "Ranged",
			},
			allyAdjacent: true,
			wantEligible: true,
		},
		{
			name: "non-finesse melee weapon",
			character: &character2.Character{
				Class: &rulebook.Class{Key: "rogue", Name: "Rogue"},
				Level: 1,
			},
			weapon: &equipment.Weapon{
				Base: equipment.BasicEquipment{
					Key:  "longsword",
					Name: "Longsword",
				},
				WeaponRange: "Melee",
			},
			hasAdvantage: true,
			wantEligible: false, // Not finesse
		},
		{
			name: "finesse weapon but no advantage or ally",
			character: &character2.Character{
				Class: &rulebook.Class{Key: "rogue", Name: "Rogue"},
				Level: 1,
			},
			weapon: &equipment.Weapon{
				Base: equipment.BasicEquipment{
					Key:  "rapier",
					Name: "Rapier",
				},
				Properties: []*shared.ReferenceItem{{Key: "finesse", Name: "Finesse"}},
			},
			hasAdvantage: false,
			allyAdjacent: false,
			wantEligible: false,
		},
		{
			name: "has advantage but also disadvantage",
			character: &character2.Character{
				Class: &rulebook.Class{Key: "rogue", Name: "Rogue"},
				Level: 1,
			},
			weapon: &equipment.Weapon{
				Base: equipment.BasicEquipment{
					Key:  "dagger",
					Name: "Dagger",
				},
				Properties: []*shared.ReferenceItem{{Key: "finesse", Name: "Finesse"}},
			},
			hasAdvantage:    true,
			hasDisadvantage: true,
			allyAdjacent:    false,
			wantEligible:    false, // Advantage and disadvantage cancel
		},
		{
			name: "non-rogue with finesse weapon",
			character: &character2.Character{
				Class: &rulebook.Class{Key: "fighter", Name: "Fighter"},
				Level: 1,
			},
			weapon: &equipment.Weapon{
				Base: equipment.BasicEquipment{
					Key:  "dagger",
					Name: "Dagger",
				},
				Properties: []*shared.ReferenceItem{{Key: "finesse", Name: "Finesse"}},
			},
			hasAdvantage: true,
			wantEligible: false, // Not a rogue
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eligible := tt.character.CanSneakAttack(tt.weapon, tt.hasAdvantage, tt.allyAdjacent, tt.hasDisadvantage)
			assert.Equal(t, tt.wantEligible, eligible)
		})
	}
}

func TestCharacter_GetSneakAttackDice(t *testing.T) {
	tests := []struct {
		name     string
		level    int
		wantDice int
	}{
		{"level 1", 1, 1},
		{"level 3", 3, 2},
		{"level 5", 5, 3},
		{"level 7", 7, 4},
		{"level 9", 9, 5},
		{"level 11", 11, 6},
		{"level 13", 13, 7},
		{"level 15", 15, 8},
		{"level 17", 17, 9},
		{"level 19", 19, 10},
		{"level 20", 20, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rogue := &character2.Character{
				Class: &rulebook.Class{Key: "rogue", Name: "Rogue"},
				Level: tt.level,
			}

			dice := rogue.GetSneakAttackDice()
			assert.Equal(t, tt.wantDice, dice)
		})
	}
}

func TestCharacter_ApplySneakAttackDamage(t *testing.T) {
	rogue := &character2.Character{
		ID:    "rogue_123",
		Name:  "Shadowblade",
		Class: &rulebook.Class{Key: "rogue", Name: "Rogue"},
		Level: 5, // 3d6 sneak attack
		Resources: &shared.CharacterResources{
			SneakAttackUsedThisTurn: false,
		},
	}

	// Create a combat context
	ctx := &character2.CombatContext{
		AttackResult: &attack.Result{
			AttackRoll: 15,
			DamageRoll: 8,
			AttackType: damage.TypePiercing,
		},
		IsCritical: false,
	}

	// Apply sneak attack
	sneakDamage := rogue.ApplySneakAttack(ctx)

	// Should add sneak attack damage
	assert.Greater(t, sneakDamage, 0)
	assert.LessOrEqual(t, sneakDamage, 18) // Max 3d6
	assert.True(t, rogue.Resources.SneakAttackUsedThisTurn)

	// Try to apply again this turn - should return 0
	sneakDamage2 := rogue.ApplySneakAttack(ctx)
	assert.Equal(t, 0, sneakDamage2)
}

func TestCharacter_SneakAttackCritical(t *testing.T) {
	rogue := &character2.Character{
		ID:    "rogue_123",
		Name:  "Shadowblade",
		Class: &rulebook.Class{Key: "rogue", Name: "Rogue"},
		Level: 1, // 1d6 sneak attack
		Resources: &shared.CharacterResources{
			SneakAttackUsedThisTurn: false,
		},
	}

	// Critical hit doubles sneak attack dice
	ctx := &character2.CombatContext{
		AttackResult: &attack.Result{
			AttackRoll: 20,
			DamageRoll: 10,
			AttackType: damage.TypePiercing,
		},
		IsCritical: true,
	}

	sneakDamage := rogue.ApplySneakAttack(ctx)

	// On a crit, 1d6 becomes 2d6, so damage should be 2-12
	assert.GreaterOrEqual(t, sneakDamage, 2)
	assert.LessOrEqual(t, sneakDamage, 12)
}

func TestCharacter_ResetSneakAttackOnNewTurn(t *testing.T) {
	rogue := &character2.Character{
		ID:    "rogue_123",
		Class: &rulebook.Class{Key: "rogue", Name: "Rogue"},
		Level: 1,
		Resources: &shared.CharacterResources{
			SneakAttackUsedThisTurn: true,
		},
	}

	// Reset for new turn
	rogue.StartNewTurn()

	assert.False(t, rogue.Resources.SneakAttackUsedThisTurn)
}

// CombatContext is now defined in character_sneak_attack.go
