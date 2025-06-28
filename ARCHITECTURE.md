# D&D Discord Bot Architecture

This document provides a comprehensive overview of the codebase architecture, making it easier to find and understand different components.

## Table of Contents
- [Overall Architecture](#overall-architecture)
- [Directory Structure](#directory-structure)
- [Core Entities](#core-entities)
- [Service Layer](#service-layer)
- [Handler Layer](#handler-layer)
- [Repository Layer](#repository-layer)
- [Key Workflows](#key-workflows)
- [Configuration](#configuration)
- [Testing Strategy](#testing-strategy)

## Overall Architecture

The bot follows a **layered architecture** pattern:

```
Discord API ↔ Handlers ↔ Services ↔ Repositories ↔ Redis/External APIs
```

### Key Principles
- **Domain-Driven Design**: Core game concepts are modeled as entities
- **Dependency Injection**: Services are injected via `ServiceProvider`
- **Interface Segregation**: Clean interfaces between layers
- **Command Pattern**: Discord interactions are handled as commands

## Directory Structure

```
/cmd/                           # Application entry points
  bot/                          # Main Discord bot application
  debug-*/                      # Debug utilities and scripts

/internal/                      # Private application code
  ├── clients/                  # External API clients
  │   └── dnd5e/               # D&D 5e API client
  ├── config/                   # Configuration management
  ├── dice/                     # Dice rolling engine
  ├── entities/                 # Core domain models
  │   ├── attack/              # Attack system
  │   ├── damage/              # Damage calculations
  │   └── features/            # Class/race features
  ├── handlers/                 # Discord interaction handlers
  │   └── discord/             # Discord-specific handlers
  ├── repositories/             # Data persistence layer
  ├── services/                 # Business logic layer
  ├── testutils/               # Test utilities and fixtures
  └── uuid/                    # UUID generation utilities
```

## Core Entities

### Character System (`/internal/entities/`)

#### Primary Entities
- **`Character`** (`character.go`): Main character entity with stats, inventory, equipment
- **`Race`** (`race.go`): Character races (Human, Elf, etc.)
- **`Class`** (`class.go`): Character classes (Fighter, Wizard, etc.)
- **`Background`** (`background.go`): Character backgrounds

#### Equipment System
- **`Equipment`** (`equipment.go`): Base equipment interface
- **`Weapon`** (`weapon.go`): Weapons with attack calculations
- **`Armor`** (`armor.go`): Armor with AC calculations
- **Slots**: `SlotMainHand`, `SlotOffHand`, `SlotBody`, `SlotTwoHanded`

#### Game Mechanics
- **`CharacterResources`** (`character_resources.go`): HP, spell slots, abilities
- **`ActiveAbility`** (`active_ability.go`): Temporary effects (Rage, etc.)
- **`Proficiency`** (`proficiency.go`): Skills, weapons, armor proficiencies

### Combat System

#### Attack & Damage (`/internal/entities/attack/`, `/internal/entities/damage/`)
- **`attack.Result`**: Attack roll results with hit/miss
- **`damage.Damage`**: Damage calculations (dice + modifiers)
- **Critical hits**: Double damage dice on natural 20

#### Combat Flow
1. **Target Selection**: Choose target from encounter participants
2. **Attack Calculation**: Weapon proficiency + ability modifier
3. **Damage Application**: Weapon damage + ability modifier + effects
4. **Status Effects**: Rage damage bonus, advantage/disadvantage

### Session & Encounter System

#### Sessions (`/internal/entities/session.go`)
- **Session Types**: `dungeon`, `combat`, `roleplay`, `oneshot`
- **Metadata**: Type-safe key-value storage
- **Participants**: Players and their characters

#### Encounters (`/internal/entities/encounter.go`)
- **Combat Encounters**: Turn-based combat with initiative
- **Participants**: Mix of player characters and monsters
- **State Management**: Active turn, round tracking

## Service Layer

### Service Provider Pattern (`/internal/services/`)

All services are injected via `ServiceProvider`:

```go
type Provider struct {
    CharacterService   character.Service
    SessionService     session.Service
    EncounterService   encounter.Service
    DungeonService     dungeon.Service
    // ... etc
}
```

### Key Services

#### Character Service (`/internal/services/character/`)
- **Character Creation**: Multi-step wizard (race → class → abilities → etc.)
- **Equipment Management**: Weapon/armor equipping, inventory
- **Persistence**: Save/load characters from Redis

#### Encounter Service (`/internal/services/encounter/`)
- **Combat Management**: Initiative, turn order, damage application
- **Monster Integration**: Add monsters from D&D 5e API
- **Status Tracking**: HP, conditions, active effects

#### Dungeon Service (`/internal/services/dungeon/`)
- **State Machine**: Room progression, party management
- **Encounter Generation**: Dynamic monster encounters
- **Loot System**: Treasure and equipment rewards

#### Session Service (`/internal/services/session/`)
- **Session Lifecycle**: Create, join, start, end
- **Metadata Management**: Type-safe session data storage
- **Participant Tracking**: Player and character associations

## Handler Layer

### Discord Handlers (`/internal/handlers/discord/`)

#### Main Handler (`handler.go`)
- **Command Registration**: Slash command definitions
- **Request Routing**: Route interactions to specific handlers
- **Error Handling**: Centralized error responses

#### Character Handlers (`/internal/handlers/discord/dnd/character/`)
- **`CreateHandler`**: Character creation wizard
- **`EquipmentHandler`**: Weapon/armor equipping
- **`ShowHandler`**: Character sheet display
- **`ListHandler`**: Character selection and management

#### Combat Handlers (`/internal/handlers/discord/combat/`)
- **`Handler`**: Main combat coordinator
- **Attack Flow**: Target selection → attack execution → damage application
- **UI Components**: Action buttons, status displays

#### Session Handlers (`/internal/handlers/discord/dnd/session/`)
- **`CreateHandler`**: Session creation
- **`JoinHandler`**: Player joining sessions
- **`StartHandler`**: Begin session activities

## Repository Layer

### Data Persistence (`/internal/repositories/`)

#### Redis-Based Storage
- **Characters** (`characters/redis.go`): Character data with equipment
- **Sessions** (`sessions/`): Session metadata and participants
- **Encounters** (`encounters/redis.go`): Combat state persistence

#### External APIs
- **D&D 5e Client** (`/internal/clients/dnd5e/`): Equipment, spells, monsters
- **Mock Client**: Test doubles for offline development

## Key Workflows

### Character Creation Flow
1. **Start**: `/dnd character create` command
2. **Race Selection**: Choose from available races
3. **Class Selection**: Choose from available classes  
4. **Ability Scores**: Roll or assign ability scores
5. **Proficiencies**: Select skill proficiencies
6. **Equipment**: Choose starting equipment
7. **Finalize**: Save character and make active

### Combat Flow
1. **Encounter Start**: Add players and monsters
2. **Initiative**: Roll initiative for turn order
3. **Turn Loop**: 
   - Display current turn
   - Player selects action (attack, etc.)
   - Execute action and apply results
   - Advance to next turn
4. **Combat End**: When all enemies defeated

### Equipment Flow
1. **Inventory Display**: `/dnd character inventory`
2. **Item Selection**: Choose weapon/armor from inventory
3. **Equip Action**: `/dnd character equip <item>`
4. **Slot Management**: Handle slot conflicts and replacements
5. **Persistence**: Save equipment changes to database

## Configuration

### Environment Variables (`/internal/config/`)
- **`DISCORD_TOKEN`**: Bot authentication token
- **`REDIS_URL`**: Redis connection string
- **`DND5E_API_URL`**: D&D 5e API endpoint

### Feature Flags
- **Equipment System**: Weapons and shields fully implemented
- **Spell System**: Partial implementation (structure exists)
- **Class Features**: Fighter, Rogue, Barbarian, Monk complete

## Testing Strategy

### Test Organization
- **Unit Tests**: `*_test.go` files alongside source code
- **Integration Tests**: `*_integration_test.go` with Redis
- **Mock Generation**: Automated mocks via `gomock`

### Test Utilities (`/internal/testutils/`)
- **Character Fixtures**: Pre-built test characters
- **Redis Fixtures**: Test data setup/teardown
- **Mock Services**: Service doubles for unit testing

### Key Test Patterns
- **Table-Driven Tests**: Multiple test cases in loops
- **Test-Driven Development**: Write tests before implementation
- **Integration Testing**: Real Redis for repository tests

## Finding Things in the Codebase

### "I want to add a new class feature"
1. **Define Feature**: Add to `/internal/entities/features/features.go`
2. **Feature Logic**: Implement in `/internal/entities/features/`
3. **Character Integration**: Update character methods
4. **Tests**: Add to `/internal/entities/features/*_test.go`

### "I want to add a new Discord command"
1. **Handler**: Create in `/internal/handlers/discord/dnd/`
2. **Service Logic**: Implement in `/internal/services/`
3. **Command Registration**: Add to `/internal/handlers/discord/handler.go`
4. **Tests**: Add handler and service tests

### "I want to modify combat mechanics"
1. **Entities**: Update `/internal/entities/attack/` or `/internal/entities/damage/`
2. **Combat Handler**: Modify `/internal/handlers/discord/combat/`
3. **Service Logic**: Update `/internal/services/encounter/`
4. **Tests**: Comprehensive combat scenario tests

### "I want to add new equipment"
1. **Equipment Data**: Usually comes from D&D 5e API
2. **Equipment Logic**: Modify `/internal/entities/weapon.go` or `/internal/entities/armor.go`
3. **Equipment Handler**: Update `/internal/handlers/discord/dnd/character/equipment.go`
4. **Tests**: Add equipment-specific tests

### "I want to modify character creation"
1. **Character Handlers**: Update `/internal/handlers/discord/dnd/character/`
2. **Character Service**: Modify `/internal/services/character/`
3. **Entities**: Update core character models
4. **UI Flow**: Modify interaction sequences in handlers

## Current Architecture Strengths
- **Clear Separation**: Each layer has distinct responsibilities
- **Testability**: Services are easily mockable
- **Extensibility**: Easy to add new commands and features
- **Type Safety**: Strong typing throughout the codebase
- **Error Handling**: Consistent error patterns

## Areas for Improvement
- **Documentation**: More inline code documentation needed
- **Caching**: Could benefit from better caching strategies
- **Event System**: Could use pub/sub for cross-cutting concerns
- **Validation**: More comprehensive input validation needed
- **Metrics**: Observability and monitoring capabilities

## Common Patterns

### Service Injection
```go
func NewHandler(provider *services.Provider) *Handler {
    return &Handler{
        characterService: provider.CharacterService,
    }
}
```

### Error Handling
```go
if err != nil {
    log.Printf("Error: %v", err)
    return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
        Type: discordgo.InteractionResponseChannelMessageWithSource,
        Data: &discordgo.InteractionResponseData{
            Content: "❌ Something went wrong!",
            Flags:   discordgo.MessageFlagsEphemeral,
        },
    })
}
```

### Repository Pattern
```go
type Repository interface {
    GetByID(id string) (*Entity, error)
    Save(entity *Entity) error
    Delete(id string) error
}
```

This architecture document should be updated as the codebase evolves and new patterns emerge.