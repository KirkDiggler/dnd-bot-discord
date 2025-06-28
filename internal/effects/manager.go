package effects

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// Manager manages status effects for a character
type Manager struct {
	effects map[string]*StatusEffect
	mu      sync.RWMutex
}

// NewManager creates a new effect manager
func NewManager() *Manager {
	return &Manager{
		effects: make(map[string]*StatusEffect),
	}
}

// AddEffect adds a new status effect
func (m *Manager) AddEffect(effect *StatusEffect) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if effect.ID == "" {
		return fmt.Errorf("effect must have an ID")
	}

	// Handle stacking rules with existing effects of the same name
	for id, existing := range m.effects {
		if existing.Name == effect.Name && existing.Source == effect.Source {
			switch effect.StackingRule {
			case StackingReplace:
				// Remove the old effect
				delete(m.effects, id)
			case StackingTakeHighest:
				// Compare effect values (simplified for now)
				// In a real implementation, we'd parse and compare modifier values
				return nil // Keep existing
			case StackingTakeLowest:
				// Compare effect values (simplified for now)
				return nil // Keep existing
			case StackingStack:
				// Allow multiple instances
			}
		}
	}

	// Set expiration time based on duration
	switch effect.Duration.Type {
	case DurationRounds:
		// For now, assume 1 round = 6 seconds (typical combat round)
		expireTime := time.Now().Add(time.Duration(effect.Duration.Rounds*6) * time.Second)
		effect.ExpiresAt = &expireTime
	case DurationInstant:
		// Instant effects expire immediately after application
		expireTime := time.Now()
		effect.ExpiresAt = &expireTime
	case DurationPermanent, DurationWhileEquipped:
		// No expiration
		effect.ExpiresAt = nil
	}

	effect.CreatedAt = time.Now()
	effect.Active = true
	m.effects[effect.ID] = effect

	return nil
}

// RemoveEffect removes a status effect by ID
func (m *Manager) RemoveEffect(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.effects, id)
}

// RemoveEffectsBySource removes all effects from a specific source
func (m *Manager) RemoveEffectsBySource(source EffectSource, sourceID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	toRemove := []string{}
	for id, effect := range m.effects {
		if effect.Source == source && effect.SourceID == sourceID {
			toRemove = append(toRemove, id)
		}
	}

	for _, id := range toRemove {
		delete(m.effects, id)
	}
}

// GetActiveEffects returns all active, non-expired effects
func (m *Manager) GetActiveEffects() []*StatusEffect {
	m.mu.RLock()
	defer m.mu.RUnlock()

	active := []*StatusEffect{}
	for _, effect := range m.effects {
		if effect.Active && !effect.IsExpired() {
			active = append(active, effect)
		}
	}
	return active
}

// GetModifiers returns all modifiers for a specific target
func (m *Manager) GetModifiers(target ModifierTarget, conditions map[string]string) []Modifier {
	m.mu.RLock()
	defer m.mu.RUnlock()

	modifiers := []Modifier{}

	for _, effect := range m.effects {
		if !effect.Active || effect.IsExpired() {
			continue
		}

		// Check if effect conditions are met
		conditionsMet := true
		for _, cond := range effect.Conditions {
			if val, exists := conditions[cond.Type]; !exists || val != cond.Value {
				conditionsMet = false
				break
			}
		}

		if !conditionsMet {
			continue
		}

		// Add applicable modifiers
		for _, mod := range effect.Modifiers {
			if mod.Target == target {
				// Check modifier-specific conditions
				if mod.Condition != "" {
					if !checkModifierCondition(mod.Condition, conditions) {
						continue
					}
				}
				modifiers = append(modifiers, mod)
			}
		}
	}

	return modifiers
}

// checkModifierCondition checks if a modifier condition is met
func checkModifierCondition(condition string, conditions map[string]string) bool {
	// Parse condition like "melee_only" or "vs_enemy_type:orc"
	if condition == "melee_only" {
		return conditions["attack_type"] == "melee"
	}

	if strings.HasPrefix(condition, "vs_enemy_type:") {
		enemyType := strings.TrimPrefix(condition, "vs_enemy_type:")
		return conditions["enemy_type"] == enemyType
	}

	if strings.HasPrefix(condition, "with_weapon:") {
		weapon := strings.TrimPrefix(condition, "with_weapon:")
		return conditions["weapon"] == weapon
	}

	// Add more condition types as needed
	return true
}

// ProcessRoundEnd handles effects that expire at end of round
func (m *Manager) ProcessRoundEnd() {
	m.mu.Lock()
	defer m.mu.Unlock()

	toRemove := []string{}
	for id, effect := range m.effects {
		if effect.Duration.Type == DurationRounds && effect.ExpiresAt != nil {
			// Decrement rounds (in a real implementation, we'd track rounds properly)
			if effect.IsExpired() {
				toRemove = append(toRemove, id)
			}
		}
	}

	for _, id := range toRemove {
		delete(m.effects, id)
	}
}

// ClearConcentration removes all concentration effects
func (m *Manager) ClearConcentration() {
	m.mu.Lock()
	defer m.mu.Unlock()

	toRemove := []string{}
	for id, effect := range m.effects {
		if effect.Duration.Concentration {
			toRemove = append(toRemove, id)
		}
	}

	for _, id := range toRemove {
		delete(m.effects, id)
	}
}

// GetEffectBySourceAndName finds an effect by its source and name
func (m *Manager) GetEffectBySourceAndName(source EffectSource, name string) *StatusEffect {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, effect := range m.effects {
		if effect.Source == source && effect.Name == name && effect.Active && !effect.IsExpired() {
			return effect
		}
	}
	return nil
}
