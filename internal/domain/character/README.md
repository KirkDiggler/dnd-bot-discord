# Character Domain Package

## Purpose
The Character package defines the core character entity and universal behaviors that apply across all game systems. It provides the foundation that rulesets build upon.

## Core Responsibilities
- **Character Entity**: Core attributes like name, ID, owner
- **Universal Mechanics**: HP, equipment slots, inventory
- **Combat Basics**: Attack() method that uses equipped weapons
- **State Management**: Thread-safe character modifications
- **Effect System**: Managing active status effects
- **Resource Tracking**: HP, ability uses, spell slots (generic)

## Universal vs Ruleset-Specific

### Universal (Lives Here)
```go
type Character struct {
    ID               string
    Name             string
    CurrentHitPoints int
    MaxHitPoints     int
    EquippedSlots    map[Slot]Equipment
    Inventory        map[EquipmentType][]Equipment
    Resources        *CharacterResources
    // ...
}
```

### Ruleset-Specific (Lives in /rulebook)
- How AC is calculated (D&D: 10 + DEX vs other systems)
- What attributes exist (STR/DEX/CON vs other systems)
- Specific proficiencies and skills
- Class/race implementations
- Level progression rules

## Key Methods

### Universal Combat
```go
// Attack uses whatever weapon is equipped
func (c *Character) Attack() ([]*AttackResult, error)

// ApplyDamage handles HP reduction
func (c *Character) ApplyDamage(amount int)

// Heal increases current HP up to max
func (c *Character) Heal(amount int) int
```

### Equipment Management
```go
// Equip handles slot management generically
func (c *Character) Equip(item Equipment, slot Slot) error

// Unequip removes from slot and returns to inventory  
func (c *Character) Unequip(slot Slot) error
```

### Effect Management
```go
// AddStatusEffect adds any temporary effect
func (c *Character) AddStatusEffect(effect *StatusEffect) error

// GetActiveStatusEffects returns current effects
func (c *Character) GetActiveStatusEffects() []*StatusEffect
```

## Design Principles
1. **Game Agnostic**: No D&D-specific logic in the core entity
2. **Thread Safe**: All mutations use mutex protection
3. **Event Integration**: Key actions emit events for ruleset handling
4. **Extensible**: Rulesets can add behavior through composition

## What Doesn't Belong Here
- Specific ability implementations (rage, sneak attack)
- Class progression tables
- Skill check calculations  
- Saving throw mechanics
- Spell implementations
- Combat maneuvers

## Extension Pattern
Rulesets extend characters through:
1. Features system (character.Features slice)
2. Event listeners (react to character actions)
3. Resource definitions (define what resources exist)
4. Metadata fields (store ruleset-specific data)

## Future Considerations
- Generic advancement system that rulesets can hook into
- Pluggable calculation system for derived stats
- More flexible resource system for different game types
- Better separation of combat from character entity