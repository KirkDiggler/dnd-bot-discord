# GitHub Discussion: Event-Driven Feature Architecture for RPG Toolkit

## 🎯 The Vision

We're exploring a major architectural shift to make our RPG mechanics more modular and reusable. The goal is to transform our tightly-coupled feature system into an event-driven plugin architecture.

## 🤔 The Problem

Currently, our features (Rage, Sneak Attack, Martial Arts, etc.) are deeply embedded in the codebase:
- Direct property mutation scattered across files
- Adding new features requires modifying core code
- Hard to test features in isolation
- Can't easily reuse mechanics in other projects

## 💡 The Proposal

Create an event-driven system where:
1. **Core mechanics emit events** (OnAttack, OnDamage, OnDefend)
2. **Features are plugins** that listen to events and modify outcomes
3. **Event context flows through a stack** of feature modifications
4. **Pure libraries** with no persistence responsibility

### Example: How Rage Would Work
```go
// Instead of this (current):
if char.HasRage() {
    damage += 2  // Hardcoded in attack logic
}

// We'd have this (proposed):
// rage.go - standalone feature
func (r *RageFeature) OnAttack(ctx *AttackContext) error {
    if ctx.Attacker.HasEffect("rage") && ctx.IsMeleeAttack() {
        ctx.AddDamageModifier("rage", 2)
    }
    return ctx.Next() // Pass to next feature in stack
}
```

## 🏗️ Architecture Options

### 1. **Middleware Chain** (Express.js style)
- Features as middleware functions
- Context passes through chain
- Each feature can modify or abort

### 2. **Event Bus** (Pub/Sub)
- Features subscribe to event types
- Loose coupling via event names
- Async possibilities

### 3. **Hook System** (WordPress style)
- Predefined hook points
- Features register callbacks
- Priority ordering

## 🎮 Real Game Example

When a Monk with Rage attacks:
1. **AttackEvent** fired with initial context
2. **RageFeature** adds +2 damage (if raging)
3. **MartialArtsFeature** allows DEX for attack
4. **EquipmentFeature** applies weapon properties
5. Final modified context used for resolution

## 💭 Questions for Discussion

1. **Pattern Preference**: Which approach (middleware/events/hooks) resonates with you?

2. **Event Granularity**: How fine-grained should events be?
   - One big `OnAttack` event?
   - Separate `BeforeAttack`, `CalculateHit`, `CalculateDamage`, `AfterAttack`?

3. **Feature Conflicts**: How should we handle conflicting modifications?
   - Priority system?
   - First wins? Last wins?
   - Explicit conflict resolution?

4. **Performance**: Are you concerned about event dispatch overhead?
   - Is a 10-20% performance hit worth the flexibility?

5. **Type Safety**: How important is compile-time type checking?
   - Fully typed event contexts?
   - Interface{} with runtime assertions?

6. **Feature State**: Where should feature configuration live?
   - In the feature itself?
   - In character data?
   - In a separate config system?

## 🚀 Benefits We'd Gain

- **Modularity**: Develop features in isolation
- **Extensibility**: Add features without core changes  
- **Reusability**: Extract to game-agnostic toolkit
- **Testability**: Unit test individual features
- **Moddability**: Community can add custom features

## 🤝 Get Involved

This is a big architectural decision that will shape the project's future. We want your input!

- Share your experience with similar systems
- Propose alternative approaches
- Point out potential pitfalls
- Suggest implementation details

Let's build something awesome together! 🎲

---

**Tagged**: architecture, enhancement, discussion, help-wanted, rpg-toolkit