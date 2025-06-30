# RPG Toolkit Integration Strategy

## Overview
Transform `dnd-bot-discord` from a monolithic Discord bot into a consumer of the modular `rpg-toolkit` library. This enables code reuse across platforms while maintaining clean architecture.

## Current State
```
dnd-bot-discord (Go)
├── Everything tightly coupled
├── Discord-specific code mixed with game logic  
├── Features hardcoded into character/combat systems
└── Storage tied to Redis
```

## Target State
```
rpg-toolkit (TypeScript)          dnd-bot-discord (Go)
├── Pure game mechanics     →     ├── Discord handlers
├── Event-driven features   →     ├── rpg-toolkit Go client
├── Storage interfaces      →     ├── Redis adapter
└── No platform deps        →     └── Discord-specific UI
```

## Integration Approach

### Option 1: Go Port of RPG Toolkit
- Create `rpg-toolkit-go` as a Go implementation
- Maintain feature parity with TypeScript version
- Share design patterns, not code

### Option 2: gRPC Service
- Run rpg-toolkit as a microservice
- Discord bot communicates via gRPC
- Enables polyglot architecture

### Option 3: WebAssembly Bridge
- Compile TypeScript to WASM
- Run in Go via wasmer/wasmtime
- Single codebase, multiple runtimes

## Event System Design

### Core Events (Language Agnostic)
```yaml
# Event Schema
BeforeAttack:
  attacker: EntityRef
  target: EntityRef
  weapon: ItemRef
  context: AttackContext

CalculateDamage:
  source: EntityRef
  target: EntityRef
  baseDamage: number
  damageType: string
  modifiers: Modifier[]

StatusEffectApplied:
  entity: EntityRef
  effect: StatusEffect
  source: EntityRef
```

### Go Implementation
```go
// internal/rpgtoolkit/events/types.go
type Event interface {
    Type() string
    Context() Context
}

type EventHandler interface {
    Handle(event Event) error
    Priority() int
}

type Feature interface {
    ID() string
    RegisterHandlers(bus EventBus)
}
```

### TypeScript Implementation
```typescript
// rpg-toolkit/src/events/types.ts
export interface Event<T = unknown> {
  type: string;
  context: T;
}

export interface EventHandler<T = unknown> {
  handle(event: Event<T>): Promise<void>;
  priority: number;
}

export interface Feature {
  id: string;
  registerHandlers(bus: EventBus): void;
}
```

## Migration Roadmap

### Phase 1: Event System in dnd-bot-discord (Current)
1. ✅ Implement action economy
2. 🔄 Add event bus alongside existing code
3. ⏳ Migrate one feature (Rage) to events
4. ⏳ Validate approach with community

### Phase 2: Extract Core Interfaces
1. Define language-agnostic event schemas
2. Create `rpg-toolkit-go` package
3. Move event system to toolkit
4. Implement storage interfaces

### Phase 3: Feature Migration
1. Port features one by one
2. Each feature becomes a plugin
3. Remove hardcoded logic from bot
4. Add feature registry system

### Phase 4: Cross-Platform Validation
1. Build simple CLI game using toolkit
2. Create web API using toolkit
3. Ensure Discord bot still works
4. Document patterns for consumers

## Example: Rage Feature Migration

### Before (Hardcoded)
```go
// internal/entities/character_combat.go
if c.HasStatusEffect("rage") && weapon.IsMelee() {
    damageBonus += 2
}
```

### After (Event-Driven Plugin)
```go
// rpg-toolkit-go/features/rage/rage.go
package rage

type RageFeature struct {
    config RageConfig
}

func (f *RageFeature) RegisterHandlers(bus events.EventBus) {
    bus.On("calculate_damage", f.handleDamage)
    bus.On("take_damage", f.handleDefense)  
    bus.On("end_turn", f.handleTurnEnd)
}

func (f *RageFeature) handleDamage(e events.Event) error {
    ctx := e.Context().(DamageContext)
    
    if !f.isRaging(ctx.Attacker) {
        return nil
    }
    
    if ctx.IsMeleeAttack() {
        ctx.AddModifier(Modifier{
            Source: "rage",
            Value:  f.config.DamageBonus(ctx.Attacker.Level),
            Type:   "damage",
        })
    }
    
    return nil
}
```

### In Discord Bot
```go
// cmd/bot/main.go
toolkit := rpgtoolkit.New()
toolkit.RegisterFeature(rage.New())
toolkit.RegisterFeature(sneakattack.New())

// Discord handler just triggers events
func handleAttack(s *discordgo.Session, i *discordgo.InteractionCreate) {
    result := toolkit.PerformAttack(attacker, target, weapon)
    // Display result in Discord
}
```

## Benefits

1. **Code Reuse**: Same features work in Discord, web, Unity
2. **Clean Architecture**: Clear separation of concerns
3. **Testability**: Test features without Discord
4. **Community**: Others can use our mechanics
5. **Modularity**: Pick only needed features

## Challenges

1. **Language Differences**: Go vs TypeScript idioms
2. **Performance**: Event overhead vs direct calls
3. **Type Safety**: Maintaining across languages
4. **Backwards Compatibility**: Don't break existing bot
5. **Complexity**: More moving parts

## Next Steps

1. **Create GitHub Discussion** with this proposal
2. **Build proof of concept** with one feature
3. **Get community feedback** on approach
4. **Choose integration strategy** (Port/gRPC/WASM)
5. **Start incremental migration**

## Questions for Discussion

1. Should rpg-toolkit be polyglot or TypeScript-only?
2. Is event-driven the right approach for all features?
3. How do we handle cross-language type safety?
4. What's the minimum viable toolkit?
5. Who else would use this?