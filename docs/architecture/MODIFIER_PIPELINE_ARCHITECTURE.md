# Universal Modifier Pipeline Architecture

## Overview

This document outlines a unified event-driven modifier system that replaces the current approach of checking individual conditions (like `isRage()`) throughout the codebase. This system will handle ALL modifiers in D&D 5e including status effects, proficiencies, feats, equipment bonuses, class features, racial traits, and temporary effects.

## Problem Statement

Current implementation has several issues:
- Monolithic methods (Attack() is 270+ lines) with embedded condition checks
- Dual effect systems (ActiveEffect and StatusEffect) running in parallel
- Tight coupling between combat mechanics and effect application
- Difficult to test individual modifiers in isolation
- Hard to add new effects without modifying core combat code
- No clear order of operations for stacking modifiers

## Proposed Solution

An event-driven pipeline where all game mechanics emit events, and modifiers subscribe to relevant events to apply their effects.

## Core Architecture

### Event System

```go
type EventType int
const (
    // Combat Events
    BeforeAttackRoll EventType = iota
    OnAttackRoll
    AfterAttackRoll
    BeforeHit
    OnHit
    BeforeDamageRoll
    OnDamageRoll
    AfterDamageRoll
    BeforeTakeDamage
    OnTakeDamage
    AfterTakeDamage
    
    // Ability Check Events
    BeforeAbilityCheck
    OnAbilityCheck
    AfterAbilityCheck
    
    // Saving Throw Events
    BeforeSavingThrow
    OnSavingThrow
    AfterSavingThrow
    
    // Spell Events
    BeforeSpellCast
    OnSpellCast
    AfterSpellCast
    
    // Movement Events
    BeforeMove
    OnMove
    AfterMove
    
    // Resource Events
    OnShortRest
    OnLongRest
    OnTurnStart
    OnTurnEnd
)

type GameEvent struct {
    Type        EventType
    Actor       *Character
    Target      *Character
    Context     map[string]interface{} // Flexible context data
    Modifiers   []Modifier             // Collected modifiers
    Cancelled   bool                   // Events can be cancelled
}
```

### Modifier Interface

```go
type Modifier interface {
    // Unique identifier for debugging/logging
    ID() string
    
    // Source of the modifier (feat, spell, item, etc)
    Source() ModifierSource
    
    // Priority determines order of application (lower = earlier)
    Priority() int
    
    // Condition determines if this modifier applies to the event
    Condition(event *GameEvent) bool
    
    // Apply the modifier to the event
    Apply(event *GameEvent) error
    
    // Duration/expiration logic
    Duration() ModifierDuration
}

type ModifierSource struct {
    Type   SourceType // Feat, Spell, Item, ClassFeature, etc
    Name   string
    ID     string
}

type ModifierDuration interface {
    IsExpired() bool
    OnEventOccurred(event *GameEvent)
}
```

### Event Bus

```go
type EventBus struct {
    listeners map[EventType][]EventListener
    mu        sync.RWMutex
}

type EventListener interface {
    HandleEvent(event *GameEvent) error
    Priority() int
}

func (eb *EventBus) Emit(event *GameEvent) error {
    listeners := eb.getListeners(event.Type)
    
    // Sort by priority
    sort.Slice(listeners, func(i, j int) bool {
        return listeners[i].Priority() < listeners[j].Priority()
    })
    
    // Execute listeners
    for _, listener := range listeners {
        if err := listener.HandleEvent(event); err != nil {
            return err
        }
        if event.Cancelled {
            break
        }
    }
    
    return nil
}
```

### Character Integration

```go
type Character struct {
    // ... existing fields ...
    
    modifierManager *ModifierManager
    eventBus        *EventBus
}

type ModifierManager struct {
    activeModifiers []Modifier
    eventBus        *EventBus
    mu              sync.RWMutex
}

func (mm *ModifierManager) AddModifier(mod Modifier) {
    mm.mu.Lock()
    defer mm.mu.Unlock()
    
    mm.activeModifiers = append(mm.activeModifiers, mod)
    mm.registerModifierListeners(mod)
}

func (mm *ModifierManager) CollectModifiers(event *GameEvent) []Modifier {
    mm.mu.RLock()
    defer mm.mu.RUnlock()
    
    applicable := []Modifier{}
    for _, mod := range mm.activeModifiers {
        if !mod.Duration().IsExpired() && mod.Condition(event) {
            applicable = append(applicable, mod)
        }
    }
    
    // Sort by priority
    sort.Slice(applicable, func(i, j int) bool {
        return applicable[i].Priority() < applicable[j].Priority()
    })
    
    return applicable
}
```

## Implementation Examples

### Rage Implementation

```go
type RageEffect struct {
    characterID string
    level       int
    startTurn   int
}

func (r *RageEffect) ID() string { return fmt.Sprintf("rage_%s", r.characterID) }

func (r *RageEffect) Source() ModifierSource {
    return ModifierSource{Type: ClassFeature, Name: "Rage", ID: "barbarian_rage"}
}

func (r *RageEffect) Priority() int { return 100 } // Applied early

func (r *RageEffect) Condition(event *GameEvent) bool {
    switch event.Type {
    case OnDamageRoll:
        // Only melee weapon attacks
        ctx := event.Context
        return ctx["weaponType"] == "melee" && event.Actor.ID == r.characterID
    case BeforeTakeDamage:
        // Resistance to physical damage
        ctx := event.Context
        damageType := ctx["damageType"].(DamageType)
        return event.Target.ID == r.characterID && 
               (damageType == Bludgeoning || damageType == Piercing || damageType == Slashing)
    case BeforeAbilityCheck, BeforeSavingThrow:
        // Advantage on Strength
        return event.Actor.ID == r.characterID && event.Context["ability"] == "STR"
    }
    return false
}

func (r *RageEffect) Apply(event *GameEvent) error {
    switch event.Type {
    case OnDamageRoll:
        // Add rage damage bonus
        bonus := r.calculateRageBonus()
        event.Context["damageBonus"] = event.Context["damageBonus"].(int) + bonus
        
    case BeforeTakeDamage:
        // Halve physical damage
        event.Context["damageAmount"] = event.Context["damageAmount"].(int) / 2
        
    case BeforeAbilityCheck, BeforeSavingThrow:
        // Grant advantage
        event.Context["advantage"] = true
    }
    return nil
}

func (r *RageEffect) Duration() ModifierDuration {
    return &TurnDuration{turns: 10, extendOnCombat: true}
}
```

### Proficiency Implementation

```go
type ProficiencyModifier struct {
    characterID string
    profType    ProficiencyType
    profBonus   int
}

func (p *ProficiencyModifier) Condition(event *GameEvent) bool {
    switch event.Type {
    case BeforeAttackRoll:
        weapon := event.Context["weapon"].(Weapon)
        return p.characterID == event.Actor.ID && p.isProficientWith(weapon)
    case BeforeAbilityCheck:
        skill := event.Context["skill"].(Skill)
        return p.characterID == event.Actor.ID && p.isProficientIn(skill)
    }
    return false
}

func (p *ProficiencyModifier) Apply(event *GameEvent) error {
    event.Context["proficiencyBonus"] = p.profBonus
    return nil
}
```

### Fighting Style Implementation

```go
type FightingStyleArchery struct {
    characterID string
}

func (f *FightingStyleArchery) Condition(event *GameEvent) bool {
    if event.Type != BeforeAttackRoll || event.Actor.ID != f.characterID {
        return false
    }
    weapon := event.Context["weapon"].(Weapon)
    return weapon.Properties.Contains(Ranged)
}

func (f *FightingStyleArchery) Apply(event *GameEvent) error {
    event.Context["attackBonus"] = event.Context["attackBonus"].(int) + 2
    return nil
}
```

## Refactored Attack Method

```go
func (c *Character) Attack(target *Character, options AttackOptions) *AttackResult {
    // Create attack event
    attackEvent := &GameEvent{
        Type:   BeforeAttackRoll,
        Actor:  c,
        Target: target,
        Context: map[string]interface{}{
            "weapon":      options.Weapon,
            "attackType":  options.Type,
            "attackBonus": 0,
            "advantage":   false,
            "disadvantage": false,
        },
    }
    
    // Emit event to collect modifiers
    c.eventBus.Emit(attackEvent)
    
    // Roll attack with modifiers
    attackRoll := c.rollD20(attackEvent.Context["advantage"].(bool), 
                           attackEvent.Context["disadvantage"].(bool))
    
    totalAttack := attackRoll + 
                  c.getAbilityModifier(options.Weapon.AbilityScore) +
                  attackEvent.Context["attackBonus"].(int)
    
    // Check if hit
    hitEvent := &GameEvent{
        Type:   OnHit,
        Actor:  c,
        Target: target,
        Context: map[string]interface{}{
            "attackRoll": totalAttack,
            "ac":         target.GetAC(),
        },
    }
    
    target.eventBus.Emit(hitEvent) // Target can modify AC
    
    if totalAttack >= hitEvent.Context["ac"].(int) {
        // Roll damage
        damageEvent := &GameEvent{
            Type:   OnDamageRoll,
            Actor:  c,
            Target: target,
            Context: map[string]interface{}{
                "weapon":       options.Weapon,
                "damageBonus":  0,
                "damageType":   options.Weapon.DamageType,
                "criticalHit":  attackRoll == 20,
            },
        }
        
        c.eventBus.Emit(damageEvent)
        
        baseDamage := c.rollDamage(options.Weapon.Damage, 
                                  damageEvent.Context["criticalHit"].(bool))
        totalDamage := baseDamage + 
                      c.getAbilityModifier(options.Weapon.AbilityScore) +
                      damageEvent.Context["damageBonus"].(int)
        
        // Apply damage
        return c.dealDamage(target, totalDamage, options.Weapon.DamageType)
    }
    
    return &AttackResult{Hit: false}
}
```

## Priority Guidelines

Modifiers should use consistent priority values:

- 0-99: Pre-calculation modifiers (set base values)
- 100-199: Feature modifiers (class features, racial traits)
- 200-299: Status effects (conditions, spells)
- 300-399: Equipment modifiers
- 400-499: Temporary bonuses (inspiration, guidance)
- 500+: Post-calculation modifiers (caps, limits)

## Migration Strategy

### Phase 1: Core Infrastructure
1. Implement EventBus and base interfaces
2. Create ModifierManager
3. Add to Character without removing existing systems

### Phase 2: Combat Migration
1. Refactor Attack() to use events
2. Migrate rage to new system
3. Migrate fighting styles
4. Remove old condition checks

### Phase 3: Full Integration
1. Migrate all status effects
2. Implement proficiency modifiers
3. Add equipment modifiers
4. Deprecate old effect systems

### Phase 4: UI Integration
1. Update combat UI to show available abilities
2. Add modifier visualization
3. Implement action economy properly

## Benefits

1. **Extensibility**: New modifiers don't require core code changes
2. **Testability**: Each modifier can be unit tested in isolation
3. **Debugging**: Event log shows exact order of modifier application
4. **Performance**: Modifiers only checked when relevant events fire
5. **Clarity**: Clear separation between game mechanics and modifiers
6. **Reusability**: Same modifier system works for all game mechanics

## Future Considerations

### Discord Activities / React App
- Event system can be serialized to JSON for client/server communication
- Modifiers can be calculated server-side and results sent to client
- Same interfaces work regardless of platform

### RPG Toolkit Extraction
- This architecture is rule-system agnostic
- Events and modifiers can be defined per game system
- Core pipeline remains the same

### Performance Optimizations
- Cache calculated modifiers per turn
- Lazy load modifier conditions
- Event filtering at registration time

## Conclusion

This universal modifier pipeline provides a clean, extensible architecture for handling the complex interactions in D&D 5e. By treating all game mechanics as events and all bonuses/penalties as modifiers, we create a system that's both powerful and maintainable.