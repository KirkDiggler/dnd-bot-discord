# D&D Bot Implementation Documentation

## Overview

This directory contains comprehensive documentation of the D&D Discord bot's implementation. These documents describe the current architecture, patterns, and flows as implemented in the codebase.

## Documentation Structure

### Core Systems

1. **[Character Actions](./character-actions.md)**
   - Complete flow from Discord interaction to game mechanics
   - UI integration patterns
   - Action economy management
   - Central reference for understanding player actions

2. **[Rulebook System](./rulebook-system.md)**
   - Abilities, Features, and Feats implementation
   - Handler patterns and registration
   - Distinction between active and passive mechanics
   - Guidelines for adding new game content

3. **[Combat Flow](./combat-flow.md)**
   - Detailed attack and damage calculations
   - Order of operations for combat resolution
   - Resistance and vulnerability application
   - Critical hit mechanics

### Technical Architecture

4. **[Effect Systems](./effect-systems.md)**
   - Dual effect system (ActiveEffect vs StatusEffect)
   - Synchronization between legacy and new systems
   - Effect lifecycle and persistence
   - Migration strategy and known issues

5. **[Event Architecture](./event-architecture.md)**
   - rpg-toolkit event bus integration
   - Event types and flow
   - Priority system and best practices
   - Cross-cutting concerns implementation

## Key Architectural Decisions

### Layered Architecture
```
Discord UI Layer
    ↓
Handler Layer (Discord interactions)
    ↓
Service Layer (Business logic)
    ↓
Domain Layer (Game rules)
    ↓
Infrastructure Layer (Database, Events)
```

### Event-Driven Design
- Loose coupling between systems
- Extensible mechanics through event subscription
- Priority-based execution order
- Synchronous event processing

### Dual Effect System
- Maintaining backward compatibility
- Gradual migration strategy
- Persistence through character Resources
- Event integration for reactive effects

## Implementation Patterns

### Consistency Guidelines

1. **Handler Pattern**: All abilities/features implement specific interfaces
2. **Registry Pattern**: Central registration for dynamic lookup
3. **Service Layer**: Business logic separated from UI and domain
4. **Event Integration**: Cross-cutting concerns via event bus
5. **Effect Application**: Always through effect system, never direct

### Common Workflows

**Adding a New Ability**:
1. Implement handler interface
2. Register in ability registry
3. Add to character initialization
4. Create UI integration

**Adding a New Feature**:
1. Implement feature handler
2. Register in feature registry
3. Add to race/class definitions
4. Apply during character creation

**Modifying Combat Mechanics**:
1. Subscribe to appropriate events
2. Set correct priority
3. Modify event context
4. Test event interactions

## Current State Assessment

### What's Working Well
- Clean separation of concerns
- Extensible event system
- Comprehensive D&D 5e rule implementation
- Robust combat calculations
- Good test coverage for core mechanics

### Areas for Improvement
- Complete migration to StatusEffect system
- Unified handler interfaces
- More comprehensive condition system
- Better action economy tracking
- Performance optimizations for large combats

### Technical Debt
- Dual effect system complexity
- Some hardcoded ability checks in services
- Incomplete modifier conversion
- Mixed UI update patterns

## Future Enhancements

### High Priority
1. Complete StatusEffect migration
2. Implement reaction system
3. Add comprehensive conditions
4. Improve action economy UI

### Medium Priority
1. Grid-based movement
2. Spell slot management
3. Concentration tracking
4. Area effect targeting

### Long Term
1. Multi-ruleset support
2. Advanced combat options
3. Environmental interactions
4. Campaign persistence

## Development Guidelines

### Code Organization
- Domain logic in `/internal/domain/`
- Service orchestration in `/internal/services/`
- UI handlers in `/internal/handlers/discord/`
- Shared types in `/internal/domain/shared/`

### Testing Strategy
- Unit tests for domain logic
- Integration tests for services
- Mock Discord interactions
- Event system testing

### Documentation Standards
- Update docs when changing mechanics
- Include examples in documentation
- Document special cases and edge conditions
- Keep architecture diagrams current

## Getting Started

For developers new to the codebase:

1. Read [Character Actions](./character-actions.md) for the overall flow
2. Understand the [Rulebook System](./rulebook-system.md) for game mechanics
3. Review [Event Architecture](./event-architecture.md) for extensibility
4. Study [Combat Flow](./combat-flow.md) for combat details
5. Learn about [Effect Systems](./effect-systems.md) for status management

## Contributing

When making changes:
1. Follow existing patterns for consistency
2. Update relevant documentation
3. Add tests for new functionality
4. Consider event system integration
5. Maintain backward compatibility

This implementation documentation provides a comprehensive view of the D&D bot's architecture and serves as the authoritative reference for understanding and extending the system.