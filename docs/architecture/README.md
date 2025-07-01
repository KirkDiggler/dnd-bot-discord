# Architecture Documentation

This directory contains architectural decisions and design patterns for the D&D Discord Bot.

## Core Architecture Documents

### [Unified Tracking System](./unified-tracking-system.md)
Describes the proposed system for tracking all game effects (abilities, conditions, resources) using composable Tracker + Modifier patterns.

### [Event-Driven Modifiers](./event-driven-modifiers.md) 
Explains how game entities emit events and modifiers listen to apply rule changes, enabling loose coupling between core systems and rulesets.

## Examples

### [Modifier Examples](../examples/modifier-examples.md)
Concrete implementations showing how different types of modifiers would work in practice.

### [Event Flow Examples](../examples/event-flow-examples.md)
Step-by-step examples of how events flow through the system for common scenarios.

## Design Principles

1. **Entities Stay Pure**: Core game objects (Character, Monster, Encounter) only handle mechanical actions
2. **Events Enable Integration**: All modifications happen through event listeners
3. **Composition Over Inheritance**: Complex effects are built by combining simple trackers and modifiers
4. **Ruleset Agnostic**: Core services have zero knowledge of specific game rules
5. **Emergent Complexity**: Rich interactions arise from simple, focused components

## Benefits

- **Easy to Add Features**: New abilities/spells/conditions are just data composition
- **Multiple Rulesets**: Different game systems can coexist by registering different listeners
- **Testable**: Each component can be tested in isolation
- **Debuggable**: Full audit trail of what modified what and why
- **Extensible**: New event types and modifier patterns can be added without breaking existing code