package character

import (
	"log"

	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/shared"
	"github.com/KirkDiggler/dnd-bot-discord/internal/effects"
)

// AddStatusEffect adds a new status effect to the character
func (c *Character) AddStatusEffect(effect *effects.StatusEffect) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.addStatusEffectInternal(effect)
}

// addStatusEffectInternal adds a status effect without locking (caller must hold lock)
func (c *Character) addStatusEffectInternal(effect *effects.StatusEffect) error {
	err := c.getEffectManagerInternal().AddEffect(effect)
	if err != nil {
		return err
	}

	// Also sync to the persisted ActiveEffects for backward compatibility
	c.syncEffectManagerToResources()
	return nil
}

// syncEffectManagerToResources syncs EffectManager to the persisted ActiveEffects
func (c *Character) syncEffectManagerToResources() {
	if c.EffectManager == nil || c.Resources == nil {
		return
	}

	// Get active effects from manager
	activeEffects := c.EffectManager.GetActiveEffects()

	// Clear old active effects
	c.Resources.ActiveEffects = []*shared.ActiveEffect{}

	// Convert new status effects to old ActiveEffect format for persistence
	for _, effect := range activeEffects {
		if effect == nil {
			continue
		}

		// Convert to old format (simplified for now)
		oldEffect := &shared.ActiveEffect{
			ID:           effect.ID,
			Name:         effect.Name,
			Description:  effect.Description,
			Source:       string(effect.Source),
			SourceID:     effect.SourceID,
			Duration:     effect.Duration.Rounds,
			DurationType: shared.DurationTypeRounds, // Simplified
			Modifiers:    []shared.Modifier{},       // TODO: Convert modifiers if needed
		}

		// Debug logging for rage sync
		if effect.Name == "Rage" {
			log.Printf("=== SYNCING RAGE EFFECT TO RESOURCES ===")
			log.Printf("Effect Duration Type: %s, Rounds: %d", effect.Duration.Type, effect.Duration.Rounds)
			log.Printf("Old Effect Duration: %d", oldEffect.Duration)
		}

		// Set duration type based on new system
		switch effect.Duration.Type {
		case effects.DurationPermanent:
			oldEffect.DurationType = shared.DurationTypePermanent
		case effects.DurationRounds:
			oldEffect.DurationType = shared.DurationTypeRounds
		case effects.DurationInstant:
			oldEffect.DurationType = shared.DurationTypeRounds // Instant effects map to rounds
		case effects.DurationWhileEquipped:
			oldEffect.DurationType = shared.DurationTypePermanent
		case effects.DurationUntilRest:
			oldEffect.DurationType = shared.DurationTypeUntilRest
		}

		c.Resources.ActiveEffects = append(c.Resources.ActiveEffects, oldEffect)
	}
}

// RemoveStatusEffect removes a status effect by ID
func (c *Character) RemoveStatusEffect(effectID string) {
	c.GetEffectManager().RemoveEffect(effectID)
}

// GetActiveStatusEffects returns all active status effects
func (c *Character) GetActiveStatusEffects() []*effects.StatusEffect {
	return c.GetEffectManager().GetActiveEffects()
}

// SyncEffects ensures effect manager state is synced to persisted resources
func (c *Character) SyncEffects() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.syncEffectManagerToResources()
}

// GetEffectManager returns the character's effect manager, initializing it if needed
func (c *Character) GetEffectManager() *effects.Manager {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.getEffectManagerInternal()
}

// getEffectManagerInternal returns the effect manager without locking (caller must hold lock)
func (c *Character) getEffectManagerInternal() *effects.Manager {
	if c.EffectManager == nil {
		c.EffectManager = effects.NewManager()
		// Restore effects from persisted data
		c.syncResourcestoEffectManager()
	}
	return c.EffectManager
}

// syncResourcestoEffectManager restores EffectManager from persisted ActiveEffects
func (c *Character) syncResourcestoEffectManager() {
	if c.Resources == nil || c.Resources.ActiveEffects == nil {
		return
	}

	// Convert old ActiveEffect format back to new StatusEffect format
	for _, oldEffect := range c.Resources.ActiveEffects {
		if oldEffect == nil {
			continue
		}

		// Create new status effect
		newEffect := &effects.StatusEffect{
			ID:          oldEffect.ID,
			Name:        oldEffect.Name,
			Description: oldEffect.Description,
			Source:      effects.EffectSource(oldEffect.Source),
			SourceID:    oldEffect.SourceID,
			Duration: effects.Duration{
				Rounds: oldEffect.Duration,
			},
			Modifiers:  []effects.Modifier{}, // TODO: Convert modifiers if needed
			Conditions: []effects.Condition{},
			Active:     true,
		}

		// Set duration type
		switch oldEffect.DurationType {
		case shared.DurationTypePermanent:
			newEffect.Duration.Type = effects.DurationPermanent
		case shared.DurationTypeRounds:
			newEffect.Duration.Type = effects.DurationRounds
		case shared.DurationTypeUntilRest:
			newEffect.Duration.Type = effects.DurationUntilRest
		default:
			newEffect.Duration.Type = effects.DurationPermanent
		}

		// For well-known effects, rebuild them properly
		if oldEffect.Name == "Rage" && oldEffect.Source == string(effects.SourceAbility) {
			// Rebuild rage effect with proper modifiers
			rageEffect := effects.BuildRageEffect(c.Level)
			rageEffect.ID = oldEffect.ID // Keep the same ID
			if err := c.EffectManager.AddEffect(rageEffect); err != nil {
				log.Printf("Failed to add rage effect: %v", err)
			}
		} else if oldEffect.Name == "Favored Enemy" && oldEffect.Source == string(effects.SourceFeature) {
			// Rebuild favored enemy effect with the selected enemy type
			enemyType := "humanoids" // default fallback
			for _, feature := range c.Features {
				if feature.Key == "favored_enemy" && feature.Metadata != nil {
					if et, ok := feature.Metadata["enemy_type"].(string); ok {
						enemyType = et
						break
					}
				}
			}
			favoredEnemyEffect := effects.BuildFavoredEnemyEffect(enemyType)
			favoredEnemyEffect.ID = oldEffect.ID
			if err := c.EffectManager.AddEffect(favoredEnemyEffect); err != nil {
				log.Printf("Failed to add favored enemy effect: %v", err)
			}
		} else {
			// Add generic effect
			if err := c.EffectManager.AddEffect(newEffect); err != nil {
				log.Printf("Failed to add effect %s: %v", newEffect.Name, err)
			}
		}
	}
}
