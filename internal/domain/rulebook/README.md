# Rulebook Package

## Purpose
The Rulebook package contains ALL game-system-specific rules, mechanics, and logic. This is where different rulesets (D&D 5e, Pathfinder, etc.) implement their unique game mechanics.

## Structure
```
rulebook/
├── README.md
├── common.go          # Shared interfaces/types across all rulesets
├── dnd5e/            # D&D 5th Edition implementation
│   ├── features/     # Class features, racial traits
│   │   ├── rage.go   # Barbarian rage implementation
│   │   ├── sneak_attack.go
│   │   └── ...
│   ├── spells/       # Spell implementations
│   ├── conditions/   # Status conditions (poisoned, grappled, etc.)
│   └── rules.go      # Core D&D 5e rules
├── pathfinder/       # Pathfinder implementation (future)
└── ...
```

## Core Responsibilities
- **Ability Implementations**: What happens when abilities are used
- **Combat Rules**: Attack calculations, damage modifiers, critical hits
- **Character Progression**: Level-up benefits, feat selections
- **Conditions & Effects**: Status effect implementations
- **Game-Specific Constants**: AC calculations, skill DCs, etc.
- **Event Listeners**: React to game events with ruleset logic

## NOT Responsible For
- **Service Orchestration**: That's the service layer's job
- **Data Persistence**: Repositories handle storage
- **UI/Discord Integration**: Handlers manage user interaction
- **Generic Game Flow**: Turn order, action economy tracking

## Event Integration
Rulebook implementations listen to events and apply game rules:

```go
// Example: D&D 5e Rage Implementation
type RageListener struct {}

func (r *RageListener) HandleEvent(event *GameEvent) error {
    if event.Type == OnAbilityUsed && event.Context["ability_key"] == "rage" {
        // Apply rage effects
    }
    if event.Type == OnDamageRoll && characterHasRage(event.Actor) {
        // Add rage damage bonus
    }
    if event.Type == BeforeTakeDamage && characterHasRage(event.Target) {
        // Apply physical damage resistance
    }
}
```

## Design Principles
1. **Encapsulation**: All ruleset logic stays within its directory
2. **Event-Driven**: React to events, don't get called directly
3. **Extensible**: New rulesets can be added without touching existing code
4. **Testable**: Each rule/feature can be tested in isolation

## Adding a New Ruleset
1. Create a new directory (e.g., `/pathfinder`)
2. Implement the common interfaces
3. Register event listeners for ruleset-specific logic
4. Add configuration to select active ruleset

## Current Features (D&D 5e)
- **Barbarian**: Rage, Unarmored Defense
- **Fighter**: Second Wind, Fighting Styles
- **Monk**: Martial Arts, Unarmored Defense
- **Ranger**: Favored Enemy, Natural Explorer
- **Rogue**: Sneak Attack

## Future Vision
- Multiple rulesets can be loaded simultaneously
- Server/session can choose which ruleset to use
- Rulesets can be composed (use D&D combat with Pathfinder skills)
- Community rulesets can be added as plugins