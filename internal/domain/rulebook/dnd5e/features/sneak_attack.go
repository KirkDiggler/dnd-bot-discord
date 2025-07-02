package features

import (
	"fmt"
	"log"

	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/events"
)

// SneakAttackModifier implements the rogue sneak attack feature as an event modifier
type SneakAttackModifier struct {
	id           string
	characterID  string
	level        int
	diceCount    int
	usedThisTurn bool
	diceRoller   interface {
		Roll(numDice, sides, modifier int) (struct{ Total int }, error)
	}
}

// NewSneakAttackModifier creates a new sneak attack modifier for a specific character
func NewSneakAttackModifier(characterID string, level int, diceRoller interface{}) *SneakAttackModifier {
	// Calculate sneak attack dice: 1d6 per 2 levels (rounded up)
	diceCount := (level + 1) / 2

	// Type assert the dice roller - safe to ignore error as we handle nil
	roller, ok := diceRoller.(interface {
		Roll(numDice, sides, modifier int) (struct{ Total int }, error)
	})
	if !ok {
		roller = nil
	}

	return &SneakAttackModifier{
		id:           fmt.Sprintf("sneak_attack_%s", characterID),
		characterID:  characterID,
		level:        level,
		diceCount:    diceCount,
		usedThisTurn: false,
		diceRoller:   roller,
	}
}

// ID returns the unique identifier for this modifier
func (s *SneakAttackModifier) ID() string {
	return s.id
}

// Source returns information about where this modifier comes from
func (s *SneakAttackModifier) Source() events.ModifierSource {
	return events.ModifierSource{
		Type:        "feature",
		Name:        "Sneak Attack",
		Description: "Rogue class feature",
	}
}

// Priority returns the execution priority (90 = slightly lower than rage)
func (s *SneakAttackModifier) Priority() int {
	return 90
}

// Condition checks if this modifier should apply to the event
func (s *SneakAttackModifier) Condition(event *events.GameEvent) bool {
	// Handle turn start events to reset usage
	if event.Type == events.OnTurnStart {
		// Check if this is the rogue's turn starting
		if event.Actor != nil && event.Actor.ID == s.characterID {
			return true
		}
		return false
	}

	// Only apply to damage rolls
	if event.Type != events.OnDamageRoll {
		return false
	}

	// Must be this character's attack
	if event.Actor == nil || event.Actor.ID != s.characterID {
		return false
	}

	// Check if already used this turn
	if s.usedThisTurn {
		return false
	}

	// Check weapon eligibility
	weaponKey, hasWeapon := event.GetStringContext("weapon_key")
	if !hasWeapon {
		return false
	}

	// Check for finesse or ranged weapon
	weaponType, _ := event.GetStringContext("weapon_type")
	isRanged := weaponType == "ranged"

	// Check for finesse property
	hasFinesse, _ := event.GetBoolContext("weapon_has_finesse")

	if !isRanged && !hasFinesse {
		log.Printf("Sneak attack: weapon %s is not eligible (ranged: %v, finesse: %v)", weaponKey, isRanged, hasFinesse)
		return false
	}

	// Check combat conditions - need advantage OR ally adjacent
	hasAdvantage, _ := event.GetBoolContext("has_advantage")
	hasDisadvantage, _ := event.GetBoolContext("has_disadvantage")
	allyAdjacent, _ := event.GetBoolContext("ally_adjacent")

	// Can't sneak attack if advantage and disadvantage cancel out
	if hasAdvantage && hasDisadvantage {
		return false
	}

	// Must have advantage OR an ally adjacent to target
	if !hasAdvantage && !allyAdjacent {
		log.Printf("Sneak attack: conditions not met (advantage: %v, ally adjacent: %v)", hasAdvantage, allyAdjacent)
		return false
	}

	return true
}

// Apply modifies the event based on sneak attack effects
func (s *SneakAttackModifier) Apply(event *events.GameEvent) error {
	// Handle turn start - reset usage
	if event.Type == events.OnTurnStart {
		s.usedThisTurn = false
		log.Printf("Sneak attack: reset for new turn (character: %s)", s.characterID)
		return nil
	}

	// Handle damage roll - add sneak attack damage
	if event.Type == events.OnDamageRoll {
		// Get current damage
		currentDamage, exists := event.GetIntContext("damage")
		if !exists {
			return fmt.Errorf("no damage value in event context")
		}

		// Check if this is a critical hit
		isCritical, _ := event.GetBoolContext("is_critical")

		// Calculate dice to roll
		diceToRoll := s.diceCount
		if isCritical {
			diceToRoll *= 2
		}

		// Roll sneak attack damage
		result, err := s.diceRoller.Roll(diceToRoll, 6, 0)
		if err != nil {
			return fmt.Errorf("failed to roll sneak attack damage: %w", err)
		}

		// Apply the damage
		event.WithContext("damage", currentDamage+result.Total)
		event.WithContext("sneak_attack_damage", result.Total)
		event.WithContext("sneak_attack_dice", fmt.Sprintf("%dd6", diceToRoll))
		event.WithContext("damage_bonus_source", fmt.Sprintf("Sneak Attack (%dd6: %d)", diceToRoll, result.Total))

		// Mark as used this turn
		s.usedThisTurn = true

		log.Printf("Sneak attack applied: %dd6 = %d damage (critical: %v)", diceToRoll, result.Total, isCritical)
	}

	return nil
}

// Duration returns how long this modifier lasts (permanent for sneak attack)
func (s *SneakAttackModifier) Duration() events.ModifierDuration {
	// Sneak attack is a permanent feature, not a temporary effect
	return &events.PermanentDuration{}
}

// SneakAttackListener wraps a sneak attack modifier to work with the event bus
type SneakAttackListener struct {
	modifier *SneakAttackModifier
}

// NewSneakAttackListener creates a new sneak attack listener
func NewSneakAttackListener(characterID string, level, startTurn int, diceRoller interface{}) *SneakAttackListener {
	return &SneakAttackListener{
		modifier: NewSneakAttackModifier(characterID, level, diceRoller),
	}
}

// HandleEvent processes events for sneak attack
func (sl *SneakAttackListener) HandleEvent(event *events.GameEvent) error {
	// Check if modifier condition is met
	if !sl.modifier.Condition(event) {
		return nil
	}

	// Apply the modifier
	return sl.modifier.Apply(event)
}

// Priority returns the listener priority
func (sl *SneakAttackListener) Priority() int {
	return sl.modifier.Priority()
}

// ID returns the modifier ID for tracking
func (sl *SneakAttackListener) ID() string {
	return sl.modifier.ID()
}

// Duration returns the modifier's duration
func (sl *SneakAttackListener) Duration() events.ModifierDuration {
	return sl.modifier.Duration()
}
