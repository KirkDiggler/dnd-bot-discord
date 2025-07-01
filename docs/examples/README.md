# Architecture Examples

This directory contains concrete examples of how the proposed event-driven modifier system would work in practice.

## Files

### [Modifier Examples](./modifier-examples.md)
Detailed implementations of different types of modifiers:

- **Duration-Based**: Bless spell, Rage with conditional cancellation
- **Resource-Based**: Ki points, Spell slots with consumption tracking  
- **Movement-Triggered**: Bleed effects, Difficult terrain
- **Complex Conditional**: Sneak attack, Magic weapon enchantments

Shows what can be modified on each event type and the rules for proper modification patterns.

### [Event Flow Examples](./event-flow-examples.md)
Step-by-step walkthroughs of complete scenarios:

- **Barbarian Rage Attack**: Multiple modifiers affecting single attack
- **Rogue Sneak Attack**: Conditional damage with magic weapon
- **Complex Spell Combat**: Haste + Action Surge + Bless interactions
- **Environmental Effects**: Movement through difficult terrain while bleeding
- **Multi-Character Round**: Full combat round with overlapping effects

Demonstrates how modifiers compose automatically to create rich emergent behavior.

## Key Insights

### Things Just Do Things and Emit Events
- Core entities (Character, Monster, Encounter) only handle mechanical actions
- All game rules live in modifiers that listen to events
- No direct coupling between entities and rulesets

### Composition Creates Complexity
- Simple modifiers combine to create rich interactions
- Rage + Great Weapon Fighting + Magic Weapon all work together automatically
- Each modifier only cares about its specific piece

### Easy to Extend
- New abilities are just new event listeners
- Different rulesets can coexist by registering different modifiers  
- No code changes needed to core systems

### Perfect for RPG Design
Whether implementing D&D 5e abilities or creating custom game mechanics, the pattern is the same:
1. Define what events your effect listens to
2. Specify what it modifies when triggered
3. Set up any tracking it needs (duration, uses, resources)
4. Register the listener with the event bus

The system handles the rest automatically!