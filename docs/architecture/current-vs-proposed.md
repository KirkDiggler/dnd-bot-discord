# Current vs Proposed Architecture

## Attack Flow Comparison

### Current Implementation 
```go
// internal/services/encounter/service.go
func (s *service) PerformAttack(...) {
    // Get character
    char, err := s.characterService.GetByID(attacker.CharacterID)
    
    // Character has built-in attack logic
    attackResults, err := char.Attack()
    
    // Rage bonus is calculated inside character.Attack()
    // Sneak attack is checked after in PerformAttack
    // Fighting styles are hardcoded in character_combat.go
    // Each feature requires modifying core attack flow
}

// internal/entities/character_combat.go
func (c *Character) Attack() {
    // Check for rage effect
    if c.HasStatusEffect("rage") && weapon.IsMelee() {
        damageBonus += 2
    }
    
    // Check for fighting style
    if c.getFightingStyle() == "great_weapon" {
        // Reroll 1s and 2s
    }
    
    // Check for equipment bonuses
    // Check for other features...
    // All tightly coupled!
}
```

### Proposed Event-Driven Implementation
```go
// internal/services/combat/attack.go
func (s *service) PerformAttack(...) {
    // Create attack context
    ctx := events.NewAttackContext(attacker, target, weapon)
    
    // Emit pre-attack event
    s.eventBus.Emit(events.BeforeAttack, ctx)
    
    // Roll attack with modifiers from context
    attackRoll := s.rollAttack(ctx)
    ctx.Set("attack_roll", attackRoll)
    
    // Emit calculate hit event
    s.eventBus.Emit(events.CalculateHit, ctx)
    
    if ctx.Get("hit").(bool) {
        // Emit calculate damage event
        s.eventBus.Emit(events.CalculateDamage, ctx)
        
        // Apply damage with all modifiers
        damage := s.calculateFinalDamage(ctx)
        target.TakeDamage(damage)
        
        // Emit after attack event
        s.eventBus.Emit(events.AfterAttack, ctx)
    }
}

// Features are now isolated modules

// internal/features/rage/rage.go
func (r *RageFeature) HandleEvent(event EventType, ctx Context) {
    switch event {
    case events.CalculateDamage:
        if r.isRaging(ctx.Attacker()) && r.isMelee(ctx.Weapon()) {
            ctx.AddModifier(DamageModifier{
                Source: "rage",
                Value:  r.getRageBonus(ctx.Attacker().Level),
            })
        }
    }
}

// internal/features/sneak_attack/sneak_attack.go
func (s *SneakAttackFeature) HandleEvent(event EventType, ctx Context) {
    switch event {
    case events.CalculateDamage:
        if s.canSneakAttack(ctx) && !s.usedThisTurn(ctx.Attacker()) {
            ctx.AddModifier(DamageModifier{
                Source: "sneak_attack",
                Dice:   s.getSneakAttackDice(ctx.Attacker().Level),
            })
            s.markUsed(ctx.Attacker())
        }
    }
}

// internal/features/great_weapon_fighting/gwf.go
func (g *GreatWeaponFighting) HandleEvent(event EventType, ctx Context) {
    switch event {
    case events.RollDamage:
        if g.hasStyle(ctx.Attacker()) && g.isTwoHanded(ctx.Weapon()) {
            ctx.Set("reroll_low_dice", true)
            ctx.Set("reroll_threshold", 2)
        }
    }
}
```

## Adding a New Feature

### Current Approach (Hex spell example)
1. Modify `character.go` to track hex target
2. Modify `character_combat.go` to add hex damage
3. Modify `ability_service.go` to handle hex casting
4. Modify `encounter_service.go` to check hex conditions
5. Add hex-specific methods throughout codebase

### Proposed Approach (Hex spell example)
```go
// internal/features/hex/hex.go - ONE FILE!
package hex

type HexFeature struct {
    hexTargets map[string]string // charID -> targetID
}

func (h *HexFeature) HandleEvent(event EventType, ctx Context) {
    switch event {
    case events.SpellCast:
        if ctx.Spell().ID == "hex" {
            h.hexTargets[ctx.Caster().ID] = ctx.Target().ID
            ctx.Target().AddEffect(Effect{
                ID: "hexed",
                Source: ctx.Caster().ID,
                Duration: 3600, // 1 hour
            })
        }
        
    case events.CalculateDamage:
        targetID := h.hexTargets[ctx.Attacker().ID]
        if targetID == ctx.Target().ID {
            ctx.AddModifier(DamageModifier{
                Source: "hex",
                Dice:   "1d6",
                Type:   "necrotic",
            })
        }
        
    case events.CharacterDeath:
        // Clean up hex if target dies
        for caster, target := range h.hexTargets {
            if target == ctx.Character().ID {
                delete(h.hexTargets, caster)
            }
        }
    }
}

// That's it! No core code changes needed.
```

## Benefits Summary

| Aspect | Current | Proposed |
|--------|---------|----------|
| Adding Features | Modify 4-5 core files | Add 1 isolated feature file |
| Testing | Must test entire combat flow | Test feature in isolation |
| Dependencies | Circular dependencies common | Features depend only on event API |
| Reusability | Tied to this specific game | Can extract to rpg-toolkit |
| Mod Support | Impossible without forking | Drop in new feature files |
| Debugging | Stack traces through entire system | Clear event flow with feature names |
| Performance | Direct function calls | ~10-20% overhead from event dispatch |

## Migration Example

We could migrate incrementally:

```go
// Phase 1: Add event emissions to existing code
func (c *Character) Attack() {
    // Existing code...
    
    // NEW: Emit event alongside existing logic
    ctx := NewAttackContext(c, target, weapon)
    c.eventBus.Emit(BeforeAttack, ctx)
    
    // Existing rage check still works
    if c.HasStatusEffect("rage") {
        damageBonus += 2
    }
    
    // But rage feature can ALSO listen to event
    // This allows gradual migration
}

// Phase 2: Features start using events
// Phase 3: Remove old hardcoded logic
// Phase 4: Extract to rpg-toolkit
```