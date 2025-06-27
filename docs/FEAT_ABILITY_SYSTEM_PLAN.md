# Feat and Ability System Implementation Plan

## Overview

This document outlines the implementation plan for adding feats, abilities, and spells to support all level 1 D&D 5e classes in the Discord bot.

## Goals

1. Support all level 1 class abilities (Rage, Second Wind, Spellcasting, etc.)
2. Create an interactive combat action controller (character sheet style)
3. Track resource usage (spell slots, ability uses, etc.)
4. Handle temporary effects and durations
5. Prepare for future positioning system integration

## Architecture Overview

### Core Components

```go
// entities/abilities.go
type AbilityType string

const (
    AbilityTypeAction      AbilityType = "action"
    AbilityTypeBonusAction AbilityType = "bonus_action"
    AbilityTypeReaction    AbilityType = "reaction"
    AbilityTypeFree        AbilityType = "free" // No action required
)

type RestType string

const (
    RestTypeShort RestType = "short"
    RestTypeLong  RestType = "long"
)

type ActiveAbility struct {
    Key           string
    Name          string
    Description   string
    FeatureKey    string      // Links to CharacterFeature
    ActionType    AbilityType
    UsesMax       int
    UsesRemaining int
    RestType      RestType
    IsActive      bool        // For toggle abilities like Rage
    Duration      int         // Rounds remaining (-1 for unlimited)
}

type CharacterResources struct {
    HP            HPResource
    SpellSlots    map[int]int // level -> remaining
    Abilities     map[string]*ActiveAbility
    ActiveEffects []ActiveEffect
}
```

### Combat Action System

```go
type CombatActionType string

const (
    ActionTypeWeaponAttack CombatActionType = "weapon_attack"
    ActionTypeAbility      CombatActionType = "ability"
    ActionTypeSpell        CombatActionType = "spell"
    ActionTypeItem         CombatActionType = "item"
)

type CombatAction struct {
    Type        CombatActionType
    Name        string
    Key         string
    ActionCost  AbilityType // action, bonus_action, etc
    Available   bool
    Reason      string // Why unavailable (e.g., "No uses remaining")
}
```

### Effect System

```go
type ActiveEffect struct {
    ID           string
    Name         string
    Source       string // Spell/Feature that created it
    Target       string // Combatant ID
    Duration     int    // Rounds remaining
    DurationType string // "rounds", "minutes", "until_rest"
    Modifiers    []Modifier
}

type Modifier struct {
    Type        string   // "damage_bonus", "resistance", "advantage"
    Value       int
    DamageTypes []string // For resistances/vulnerabilities
    SkillTypes  []string // For advantage/disadvantage
}
```

## Implementation Phases

### Phase 1: Foundation (Issues #130-132)

**Issue #130: Implement Active Ability System Foundation**
- Add ability tracking to Character entity
- Create resource management system
- Add CharacterResources to track HP, abilities, effects
- Implement rest mechanics (short/long rest)

**Issue #131: Add Effect Duration Tracking to Combat**
- Create ActiveEffect system
- Add effect application/removal during combat
- Track duration countdowns each round
- Handle effect expiration

**Issue #132: Enhance Combat Actions Controller**
- Transform "Get My Actions" into full character sheet
- Show available abilities alongside weapon attacks
- Track action economy (action/bonus/reaction used)
- Add ability usage buttons with Discord interactions

### Phase 2: Basic Class Abilities (Issues #133-136)

**Issue #133: Implement Barbarian Rage**
- First toggle ability implementation
- Damage bonus to melee attacks
- Damage resistance (bludgeoning, piercing, slashing)
- Advantage on Strength checks/saves
- Duration: 10 rounds (1 minute)
- Uses: 2/long rest at level 1

**Issue #134: Implement Fighter Second Wind**
- Healing ability (1d10 + level)
- Bonus action
- 1 use per short/long rest
- Test healing mechanics

**Issue #135: Implement Rogue Sneak Attack**
- Conditional damage bonus
- Once per turn limitation
- Advantage or ally adjacency requirement
- Always available (no resource cost)

**Issue #136: Implement Monk Martial Arts**
- Bonus action unarmed strike
- Dexterity for unarmed attacks
- Modified unarmed damage (d4 at level 1)

### Phase 3: Spell System Foundation (Issues #137-139)

**Issue #137: Add Spell Support to D&D Client**
```go
// Add to dnd5e client interface
GetSpell(key string) (*entities.Spell, error)
ListSpells(classKey string, level int) ([]*entities.Spell, error)
ListClassSpells(classKey string) ([]*entities.Spell, error)
```

**Issue #138: Implement Spell Slot System**
- Add spell slots to CharacterResources
- Track slots by level
- Implement slot consumption on cast
- Restore on long rest

**Issue #139: Create Spell Casting UI**
- Add spell list to combat actions
- Group by spell level
- Show available slots
- Handle spell selection and targeting

### Phase 4: Basic Spells (Issues #140-143)

**Issue #140: Implement Cure Wounds**
- First healing spell
- 1d8 + spellcasting modifier
- Touch range
- Tests targeting allies

**Issue #141: Implement Healing Word**
- Bonus action healing
- 1d4 + spellcasting modifier
- Range: 60 feet
- Tests ranged healing

**Issue #142: Implement Sacred Flame**
- Cantrip (unlimited uses)
- Dexterity save
- 1d8 radiant damage
- Tests saving throws

**Issue #143: Implement Shield of Faith**
- Bonus action
- +2 AC bonus
- Concentration (10 minutes)
- Tests buff spells

### Phase 5: Advanced Class Features (Issues #144-147)

**Issue #144: Implement Paladin Divine Sense**
- Detect celestials, fiends, undead
- Limited uses (1 + Charisma modifier)
- Action to use
- Range: 60 feet

**Issue #145: Implement Ranger Favored Enemy**
- Advantage on Survival checks
- Damage bonus vs chosen enemy type
- Passive ability

**Issue #146: Implement Warlock Hex**
- Bonus action curse
- Extra damage on hits
- Disadvantage on ability checks
- Concentration

**Issue #147: Implement Sorcerer Font of Magic**
- Sorcery points system
- Convert spell slots to points
- Flexible resource management

## Combat UI Vision

### Enhanced Combat Actions Display
```
🎯 Your Combat Actions - Ragnor (Barbarian 1)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

⚔️ Weapon Attacks
├─ Greataxe (Action) +5 to hit, 1d12+3 slashing
└─ Handaxe (Action/Thrown) +5 to hit, 1d6+3 slashing

💪 Class Abilities
├─ 🔥 Rage (Bonus Action) [2/2 uses] - Gain resistance and +2 damage
└─ 🛡️ Unarmored Defense (Passive) - AC = 10 + DEX + CON

🧪 Items
├─ Healing Potion (Action) [2 potions] - Heal 2d4+2
└─ Torch (Free) - Provide light

📊 Action Economy         🔄 Resources
├─ Action: ✅             ├─ HP: 12/12
├─ Bonus Action: ✅       ├─ Rage: 2/2
├─ Reaction: ✅           └─ Hit Dice: 1d12
└─ Movement: 30 ft

🎮 Quick Actions
[⚔️ Attack] [🔥 Rage] [🏃 Dash] [🛡️ Dodge] [💬 More...]
```

### Spell Caster Display
```
🎯 Your Combat Actions - Elara (Cleric 1)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

⚔️ Weapon Attacks
└─ Mace (Action) +2 to hit, 1d6 bludgeoning

✨ Spells Prepared (4/4)
├─ Cantrips (∞)
│  ├─ 🔥 Sacred Flame - DEX save or 1d8 radiant
│  └─ 💡 Light - Touch object glows 20ft
└─ 1st Level (2 slots)
   ├─ ❤️ Cure Wounds - Touch heal 1d8+3
   ├─ 💬 Healing Word - 60ft heal 1d4+3 (Bonus)
   └─ 🛡️ Shield of Faith - +2 AC for 10 min (Bonus)

📊 Action Economy         🔄 Resources
├─ Action: ✅             ├─ HP: 8/8
├─ Bonus Action: ✅       ├─ Spell Slots
├─ Reaction: ✅           │  └─ 1st: ⬤⬤
└─ Movement: 25 ft        └─ Channel Divinity: 0

🎮 Quick Spells
[🔥 Sacred Flame] [❤️ Cure Wounds] [💬 Healing Word] [📖 All Spells...]
```

## Testing Strategy

1. **Unit Tests**: Each ability handler tested in isolation
2. **Integration Tests**: Full combat scenarios with abilities
3. **Discord Tests**: Interaction flow testing
4. **Balance Tests**: Ensure abilities match D&D 5e rules

## Future Considerations

### Positioning System Integration
- Range calculations for spells/abilities
- Area of effect handling
- Line of sight requirements
- Movement abilities (Dash, Disengage)

### Higher Level Features
- Multi-classing support
- Feat selection at level 4
- Subclass features
- Higher level spells

### Advanced Effect Stacking
- Concentration limits
- Conflicting effects
- Buff/debuff priorities
- Condition immunities

## Success Metrics

1. All level 1 classes fully playable
2. Combat remains fast and intuitive
3. Resource tracking is accurate
4. Effects apply correctly
5. UI is clear and responsive

## Notes

- Start simple, iterate based on feedback
- Prioritize core class features over edge cases
- Keep combat flow smooth
- Maintain D&D 5e rule accuracy
- Design for future positioning system