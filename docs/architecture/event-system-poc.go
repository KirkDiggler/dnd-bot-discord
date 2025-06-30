// Package events - Proof of Concept for Event-Driven Features
package events

import (
	"fmt"
	"sort"
)

// EventType represents different game events
type EventType string

const (
	EventAttack       EventType = "attack"
	EventDefend       EventType = "defend"
	EventDamage       EventType = "damage"
	EventHeal         EventType = "heal"
	EventStatusChange EventType = "status_change"
)

// Context is the mutable data passed through event handlers
type Context interface {
	Get(key string) interface{}
	Set(key string, value interface{})
	AddModifier(mod Modifier)
	GetModifiers() []Modifier
}

// AttackContext specific context for attack events
type AttackContext struct {
	data      map[string]interface{}
	modifiers []Modifier
}

func NewAttackContext(attacker, target, weapon interface{}) *AttackContext {
	return &AttackContext{
		data: map[string]interface{}{
			"attacker":    attacker,
			"target":      target,
			"weapon":      weapon,
			"base_damage": 0,
			"attack_roll": 0,
		},
		modifiers: []Modifier{},
	}
}

func (c *AttackContext) Get(key string) interface{} {
	return c.data[key]
}

func (c *AttackContext) Set(key string, value interface{}) {
	c.data[key] = value
}

func (c *AttackContext) AddModifier(mod Modifier) {
	c.modifiers = append(c.modifiers, mod)
}

func (c *AttackContext) GetModifiers() []Modifier {
	return c.modifiers
}

// Modifier represents a modification to game mechanics
type Modifier struct {
	Source   string      // Which feature added this
	Type     string      // damage, attack_bonus, etc
	Value    interface{} // The modification value
	Priority int         // Order of application
}

// Feature interface that all game features must implement
type Feature interface {
	ID() string
	Priority() int
	HandleEvent(eventType EventType, ctx Context) error
}

// EventBus manages feature registration and event dispatch
type EventBus struct {
	features []Feature
}

func NewEventBus() *EventBus {
	return &EventBus{
		features: []Feature{},
	}
}

func (eb *EventBus) RegisterFeature(f Feature) {
	eb.features = append(eb.features, f)
	// Sort by priority
	sort.Slice(eb.features, func(i, j int) bool {
		return eb.features[i].Priority() < eb.features[j].Priority()
	})
}

func (eb *EventBus) Emit(eventType EventType, ctx Context) error {
	for _, feature := range eb.features {
		if err := feature.HandleEvent(eventType, ctx); err != nil {
			return fmt.Errorf("feature %s error: %w", feature.ID(), err)
		}
	}
	return nil
}

// Example Features

// RageFeature implements barbarian rage
type RageFeature struct{}

func (r *RageFeature) ID() string    { return "rage" }
func (r *RageFeature) Priority() int { return 10 }

func (r *RageFeature) HandleEvent(eventType EventType, ctx Context) error {
	switch eventType {
	case EventAttack:
		// Check if attacker has rage
		if attacker := ctx.Get("attacker"); r.hasRage(attacker) {
			if r.isMeleeWeapon(ctx.Get("weapon")) {
				ctx.AddModifier(Modifier{
					Source:   "rage",
					Type:     "damage_bonus",
					Value:    2, // Would calculate based on level
					Priority: 10,
				})
			}
		}
	case EventDefend:
		// Add damage resistance
		if defender := ctx.Get("defender"); r.hasRage(defender) {
			ctx.AddModifier(Modifier{
				Source:   "rage",
				Type:     "damage_resistance",
				Value:    0.5, // Half physical damage
				Priority: 20,
			})
		}
	}
	return nil
}

func (r *RageFeature) hasRage(char interface{}) bool {
	// Check if character has rage effect
	return true // Simplified
}

func (r *RageFeature) isMeleeWeapon(weapon interface{}) bool {
	// Check if weapon is melee
	return true // Simplified
}

// SneakAttackFeature implements rogue sneak attack
type SneakAttackFeature struct{}

func (s *SneakAttackFeature) ID() string    { return "sneak_attack" }
func (s *SneakAttackFeature) Priority() int { return 15 }

func (s *SneakAttackFeature) HandleEvent(eventType EventType, ctx Context) error {
	if eventType != EventAttack {
		return nil
	}

	if s.canSneakAttack(ctx) {
		level := 5 // Would get from character
		dice := (level + 1) / 2
		ctx.AddModifier(Modifier{
			Source:   "sneak_attack",
			Type:     "damage_dice",
			Value:    fmt.Sprintf("%dd6", dice),
			Priority: 15,
		})
	}
	return nil
}

func (s *SneakAttackFeature) canSneakAttack(ctx Context) bool {
	// Check conditions for sneak attack
	return true // Simplified
}

// Usage Example
func Example() {
	// Create event bus
	bus := NewEventBus()

	// Register features
	bus.RegisterFeature(&RageFeature{})
	bus.RegisterFeature(&SneakAttackFeature{})

	// Create attack context
	ctx := NewAttackContext("Barbarian/Rogue", "Goblin", "Shortsword")
	ctx.Set("base_damage", 6)

	// Emit attack event
	if err := bus.Emit(EventAttack, ctx); err != nil {
		fmt.Printf("Error emitting event: %v\n", err)
	}

	// Process modifiers
	fmt.Println("Attack modifiers:")
	for _, mod := range ctx.GetModifiers() {
		fmt.Printf("- %s: %s = %v\n", mod.Source, mod.Type, mod.Value)
	}
	// Output:
	// - rage: damage_bonus = 2
	// - sneak_attack: damage_dice = 3d6
}
