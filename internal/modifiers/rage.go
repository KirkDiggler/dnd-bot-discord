package modifiers

import (
	"fmt"
	"log"

	"github.com/KirkDiggler/dnd-bot-discord/internal/entities"
	"github.com/KirkDiggler/dnd-bot-discord/internal/entities/damage"
	"github.com/KirkDiggler/dnd-bot-discord/internal/events"
)

// RageModifier implements the barbarian rage feature
type RageModifier struct {
	BaseModifier
	characterID string
	level       int
}

func NewRageModifier(characterID string, level int) *RageModifier {
	return &RageModifier{
		BaseModifier: NewBaseModifier(
			fmt.Sprintf("rage_%s", characterID),
			Source{
				Type: SourceTypeClassFeature,
				Name: "Rage",
				ID:   "barbarian_rage",
			},
			events.PriorityFeatures,
			NewRoundsDuration(10), // 10 rounds = 1 minute
		),
		characterID: characterID,
		level:       level,
	}
}

// IsActive checks if this modifier applies to the character
func (m *RageModifier) IsActive(character *entities.Character) bool {
	if character.ID != m.characterID {
		return false
	}
	return m.BaseModifier.IsActive(character)
}

// GetRageBonus returns the rage damage bonus based on level
func (m *RageModifier) GetRageBonus() int {
	if m.level >= 16 {
		return 4
	} else if m.level >= 9 {
		return 3
	}
	return 2
}

// VisitBeforeAttackRollEvent applies advantage on Strength checks/saves
func (m *RageModifier) VisitBeforeAttackRollEvent(e *events.BeforeAttackRollEvent) {
	// Rage doesn't affect attack rolls directly
}

// VisitOnDamageRollEvent adds rage damage bonus to melee attacks
func (m *RageModifier) VisitOnDamageRollEvent(e *events.OnDamageRollEvent) {
	if e.Actor == nil || e.Actor.ID != m.characterID {
		return
	}

	// Only applies to melee weapon attacks
	if e.Weapon == nil {
		return
	}

	weapon, ok := e.Weapon.(*entities.Weapon)
	if !ok || !weapon.IsMelee() {
		return
	}

	bonus := m.GetRageBonus()
	e.DamageBonus += bonus
	e.TotalDamage += bonus

	log.Printf("Rage: Adding +%d damage bonus to melee attack", bonus)
}

// VisitBeforeTakeDamageEvent applies resistance to physical damage
func (m *RageModifier) VisitBeforeTakeDamageEvent(e *events.BeforeTakeDamageEvent) {
	if e.Target == nil || e.Target.ID != m.characterID {
		return
	}

	// Check if damage type is physical
	switch e.DamageType {
	case damage.TypeBludgeoning, damage.TypePiercing, damage.TypeSlashing:
		// Add resistance if not already present
		hasResistance := false
		for _, r := range e.Resistances {
			if r == e.DamageType {
				hasResistance = true
				break
			}
		}

		if !hasResistance {
			e.Resistances = append(e.Resistances, e.DamageType)
			log.Printf("Rage: Adding resistance to %s damage", e.DamageType)
		}
	}
}

// VisitBeforeAbilityCheckEvent applies advantage to Strength checks
func (m *RageModifier) VisitBeforeAbilityCheckEvent(e *events.BeforeAbilityCheckEvent) {
	if e.Actor == nil || e.Actor.ID != m.characterID {
		return
	}

	if e.Ability == entities.AttributeStrength && !e.Disadvantage {
		e.Advantage = true
		log.Printf("Rage: Granting advantage on Strength check")
	}
}

// VisitBeforeSavingThrowEvent applies advantage to Strength saves
func (m *RageModifier) VisitBeforeSavingThrowEvent(e *events.BeforeSavingThrowEvent) {
	if e.Actor == nil || e.Actor.ID != m.characterID {
		return
	}

	if e.Ability == entities.AttributeStrength && !e.Disadvantage {
		e.Advantage = true
		log.Printf("Rage: Granting advantage on Strength saving throw")
	}
}
