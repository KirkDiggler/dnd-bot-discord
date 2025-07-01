# Ability Service

## Purpose
The Ability Service manages the usage and tracking of character abilities in a **ruleset-agnostic** way. It handles the mechanical aspects of ability usage (uses remaining, action economy, cooldowns) without knowing what specific abilities do.

## Core Responsibilities
- **Usage Tracking**: Track remaining uses, cooldowns, and recharge conditions
- **Action Economy**: Enforce action, bonus action, and reaction usage rules  
- **Availability Validation**: Determine if an ability can be used
- **Event Orchestration**: Emit events when abilities are used
- **Resource Management**: Handle resource consumption (spell slots, ki points, etc.)
- **State Persistence**: Save ability state changes through character service

## NOT Responsible For
- **Ability Effects**: What happens when "rage" is used (damage bonus, resistance, etc.)
- **Ruleset-Specific Logic**: D&D 5e, Pathfinder, or any game system rules
- **Combat Calculations**: Attack rolls, damage rolls, saving throws
- **Target Selection**: UI concerns about picking targets
- **Ability Definitions**: What abilities exist or their properties

## Current State (Technical Debt)
Currently, the service has hardcoded handlers for specific D&D 5e abilities:
```go
switch input.AbilityKey {
case "rage":
    result = s.handleRage(...)
case "second-wind":
    result = s.handleSecondWind(...)
// etc...
}
```

This violates our architecture and will be removed in the refactor (Issue #237).

## Future State
After refactoring:
1. Service emits `OnAbilityUsed` event with ability context
2. Ruleset-specific listeners handle ability effects
3. Service only tracks usage and enforces action economy

Example flow:
```
User uses "rage" → AbilityService validates usage → Emits OnAbilityUsed → 
D&D5e RageListener handles effects → Updates character state
```

## Dependencies
- **CharacterService**: For loading/saving character state
- **EncounterService**: For combat context (current turn, etc.)
- **EventBus**: For emitting ability usage events
- **DiceRoller**: For any generic dice rolls (if needed)

## Events
### Emits
- `OnAbilityUsed`: When any ability is successfully used
  - Context: ability_key, character_id, target_ids, encounter_id

### Listens To
- `OnTurnStart`: Update cooldowns and refresh per-turn abilities
- `OnShortRest`: Refresh short rest abilities
- `OnLongRest`: Refresh all abilities

## API Methods
```go
// UseAbility validates and executes ability usage
UseAbility(ctx, input *UseAbilityInput) (*UseAbilityResult, error)

// GetAvailableAbilities returns all abilities a character can currently use
GetAvailableAbilities(ctx, characterID string) ([]*AvailableAbility, error)
```

## Design Principles
1. **Ruleset Agnostic**: No knowledge of specific game systems
2. **Event-Driven**: Communicate through events, not direct calls
3. **Single Responsibility**: Only manage ability usage mechanics
4. **Extensible**: New rulesets can listen to events without service changes