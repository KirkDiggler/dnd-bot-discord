# Event-Driven Feature Architecture

## Vision
Transform the current tightly-coupled feature system into a modular, event-driven architecture where features (feats, talents, abilities) are standalone plugins that can modify game mechanics through event interception.

## Current State
- Features are deeply integrated into character state
- Direct mutation of properties (AC, attack bonuses, etc.)
- Tight coupling between features and core systems
- Difficult to add new features without modifying core code

## Proposed Architecture

### Core Concepts

1. **Event System**
   - All game actions emit events (OnAttack, OnDefend, OnDamage, etc.)
   - Features register listeners/interceptors for relevant events
   - Events carry mutable context that features can modify
   - Event processing creates a stack of modifications

2. **Feature Plugins**
   - Each feature is a standalone module (rage.go, sneak_attack.go, etc.)
   - Features only know about the event interface, not the core system
   - Can be loaded/unloaded dynamically
   - Pure functions that transform event context

3. **Event Repository Pattern**
   - Core emits events without knowing about persistence
   - Event handlers can store events for event sourcing
   - Storage adapters handle actual persistence
   - Complete separation of business logic and storage

## Implementation Options

### Option 1: Middleware Chain Pattern
```go
type AttackContext struct {
    Attacker    Character
    Target      Character
    Weapon      *Weapon
    AttackRoll  int
    DamageRoll  int
    Modifiers   []Modifier
}

type AttackMiddleware func(ctx *AttackContext, next AttackHandler) error
type AttackHandler func(ctx *AttackContext) error

// Features register middleware
rage.RegisterAttackMiddleware(func(ctx *AttackContext, next AttackHandler) error {
    if ctx.Attacker.HasEffect("rage") && ctx.Weapon.IsMelee() {
        ctx.Modifiers = append(ctx.Modifiers, Modifier{
            Type: "rage",
            Value: 2,
            Target: "damage",
        })
    }
    return next(ctx)
})
```

### Option 2: Event Emitter Pattern
```go
type EventBus interface {
    Emit(event Event) error
    On(eventType string, handler EventHandler)
    OnBefore(eventType string, interceptor EventInterceptor)
}

type AttackEvent struct {
    Type     string
    Attacker Character
    Target   Character
    Weapon   *Weapon
    Results  *AttackResults
}

// Features subscribe to events
rage.OnBefore("attack", func(e Event) error {
    if attack, ok := e.(*AttackEvent); ok {
        if attack.Attacker.HasEffect("rage") {
            attack.Results.DamageBonus += 2
        }
    }
    return nil
})
```

### Option 3: Hook System
```go
type HookPoint string

const (
    BeforeAttack    HookPoint = "before_attack"
    CalculateHit    HookPoint = "calculate_hit"
    CalculateDamage HookPoint = "calculate_damage"
    AfterAttack     HookPoint = "after_attack"
)

type HookContext map[string]interface{}

type Hook func(ctx HookContext) error

// Features register hooks
features.RegisterHook(CalculateDamage, "rage", func(ctx HookContext) error {
    if attacker := ctx["attacker"].(Character); attacker.HasEffect("rage") {
        if weapon := ctx["weapon"].(*Weapon); weapon.IsMelee() {
            ctx["damage_bonus"] = ctx["damage_bonus"].(int) + 2
        }
    }
    return nil
})
```

## Benefits

1. **Modularity**: Features can be developed and tested in isolation
2. **Extensibility**: New features don't require core changes
3. **Reusability**: rpg-toolkit becomes game-agnostic
4. **Testability**: Each feature can be unit tested independently
5. **Composition**: Complex behaviors emerge from simple feature combinations

## Challenges

1. **Event Ordering**: Need clear rules for event processing order
2. **Performance**: Event dispatch overhead
3. **Debugging**: Stack traces become more complex
4. **Type Safety**: Maintaining type safety with dynamic events
5. **Migration**: Moving existing features to new system

## Migration Strategy

1. **Phase 1**: Implement event bus alongside existing system
2. **Phase 2**: Migrate one feature (e.g., Rage) as proof of concept
3. **Phase 3**: Gradually migrate other features
4. **Phase 4**: Extract to rpg-toolkit as standalone library

## Example: Rage Feature

```go
// internal/features/rage/rage.go
package rage

import "github.com/KirkDiggler/rpg-toolkit/events"

func init() {
    events.RegisterFeature(&RageFeature{})
}

type RageFeature struct{}

func (r *RageFeature) ID() string { return "rage" }

func (r *RageFeature) OnAttack(ctx *events.AttackContext) error {
    if !ctx.Attacker.HasEffect("rage") {
        return nil
    }
    
    if ctx.Weapon.IsMelee() {
        ctx.AddModifier(events.Modifier{
            Source: "rage",
            Type:   events.ModifierDamage,
            Value:  r.calculateRageBonus(ctx.Attacker.Level),
        })
    }
    
    return nil
}

func (r *RageFeature) OnDefend(ctx *events.DefendContext) error {
    if !ctx.Defender.HasEffect("rage") {
        return nil
    }
    
    // Add physical damage resistance
    if ctx.DamageType.IsPhysical() {
        ctx.AddModifier(events.Modifier{
            Source: "rage",
            Type:   events.ModifierResistance,
            Value:  0.5, // Half damage
        })
    }
    
    return nil
}
```

## Discussion Points

1. Which pattern (middleware, events, hooks) best fits our needs?
2. How do we handle feature dependencies and conflicts?
3. What's the right level of granularity for events?
4. How do we maintain performance with many features?
5. Should features be able to cancel/abort events?
6. How do we handle feature configuration and state?

## Next Steps

1. Create GitHub Discussion for community input
2. Build proof of concept with one feature
3. Define event taxonomy and interfaces
4. Design feature registration system
5. Plan migration roadmap