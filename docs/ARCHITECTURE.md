# D&D Discord Bot Architecture

## Overview

A Discord bot for D&D 5e gameplay featuring character management, combat encounters, and real-time web interface for battle maps. Built with TypeScript, gRPC, and modern architectural patterns.

## Tech Stack

### Core Technologies
- **Language**: TypeScript (Node.js)
- **Discord Integration**: Discord.js v14
- **IPC/Services**: gRPC with Protocol Buffers
- **Database**: PostgreSQL (relational data) + Redis (session/combat state)
- **Web Interface**: React + WebSockets
- **Container**: Docker & Docker Compose

### Key Libraries
- **discord.js**: Discord bot framework
- **@grpc/grpc-js**: gRPC client/server
- **@grpc/proto-loader**: Dynamic proto loading
- **prisma**: Type-safe ORM for PostgreSQL
- **ioredis**: Redis client with TypeScript support
- **socket.io**: Real-time web communication
- **zod**: Runtime type validation

## Architecture Patterns

### 1. Microservice Architecture with gRPC

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│  Discord Bot    │────▶│  Game Service   │────▶│  Web Interface  │
│   (Gateway)     │gRPC │   (Core Logic)  │gRPC │   (React App)   │
└─────────────────┘     └─────────────────┘     └─────────────────┘
        │                        │                        │
        └────────────────────────┴────────────────────────┘
                                 │
                    ┌────────────┴────────────┐
                    │                         │
              ┌─────▼─────┐           ┌──────▼──────┐
              │PostgreSQL │           │    Redis    │
              │(Permanent)│           │ (Ephemeral) │
              └───────────┘           └─────────────┘
```

### 2. Repository Pattern

Each data entity has its own repository interface:

```typescript
interface CharacterRepository {
  create(character: CharacterCreateInput): Promise<Character>
  findById(id: string): Promise<Character | null>
  findByUserId(userId: string): Promise<Character[]>
  update(id: string, data: CharacterUpdateInput): Promise<Character>
  delete(id: string): Promise<void>
}
```

### 3. Event-Driven Combat State Machine

```
┌─────────┐      ┌──────────┐      ┌────────┐      ┌──────────┐
│  IDLE   │─────▶│INITIATIVE│─────▶│ ACTIVE │─────▶│ RESOLVED │
└─────────┘      └──────────┘      └────────┘      └──────────┘
                                        │ ▲
                                        ▼ │
                                   ┌────────┐
                                   │  TURN  │
                                   └────────┘
```

### 4. Command Pattern for Discord Interactions

```typescript
interface Command {
  data: SlashCommandBuilder
  execute(interaction: CommandInteraction): Promise<void>
  autocomplete?(interaction: AutocompleteInteraction): Promise<void>
}
```

## Project Structure

```
/dnd-bot-discord
├── /proto                     # Protocol Buffer definitions
│   ├── character.proto
│   ├── combat.proto
│   ├── game.proto
│   └── common.proto
├── /src
│   ├── /bot                   # Discord bot service
│   │   ├── /commands          # Slash commands
│   │   ├── /handlers          # Button/select handlers
│   │   ├── /services          # Discord-specific services
│   │   └── index.ts
│   ├── /game                  # Core game service
│   │   ├── /grpc              # gRPC server implementation
│   │   ├── /services          # Business logic
│   │   ├── /repositories      # Data access layer
│   │   ├── /entities          # Domain models
│   │   └── index.ts
│   ├── /web                   # Web interface service
│   │   ├── /components
│   │   ├── /hooks
│   │   ├── /services
│   │   └── index.tsx
│   └── /shared                # Shared utilities
│       ├── /types             # Generated from protos
│       ├── /constants
│       └── /utils
├── /prisma                    # Database schema
│   └── schema.prisma
├── /scripts                   # Build/dev scripts
├── docker-compose.yml
├── package.json
└── tsconfig.json
```

## Core Components

### 1. Discord Bot Service

Handles all Discord interactions:
- Slash commands for character creation, combat actions
- Ephemeral responses for private information
- Persistent character sheet messages
- Combat state updates in shared channel

### 2. Game Service (gRPC Server)

Core game logic separated from Discord:
- Character management
- Combat engine
- D&D 5e rules engine
- Session/party management
- AI DM capabilities (future)

### 3. Web Interface

Real-time battle visualization:
- Grid-based battle map
- Character token positioning
- Initiative tracker
- Health/status indicators
- DM controls (for human DM)

## Data Models

### Character Model
```typescript
interface Character {
  id: string
  userId: string
  name: string
  race: Race
  class: Class
  level: number
  attributes: Attributes
  hp: { current: number; max: number }
  ac: number
  equipment: Equipment[]
  inventory: Item[]
  proficiencies: Proficiency[]
  spells?: Spell[]
}
```

### Combat Session Model
```typescript
interface CombatSession {
  id: string
  channelId: string
  participants: CombatParticipant[]
  currentTurn: number
  round: number
  state: CombatState
  map: BattleMap
}
```

## Message Architecture

### Shared Messages (Public Channel)
- Combat state overview
- Initiative order
- Public actions (attacks, movement)
- Environmental changes

### Ephemeral Messages (Private)
- Character sheet with action buttons
- Perception check results
- Private DM information
- Inventory management

## gRPC Service Definitions

### Character Service
```proto
service CharacterService {
  rpc CreateCharacter(CreateCharacterRequest) returns (Character);
  rpc GetCharacter(GetCharacterRequest) returns (Character);
  rpc UpdateCharacter(UpdateCharacterRequest) returns (Character);
  rpc ListCharacters(ListCharactersRequest) returns (CharacterList);
}
```

### Combat Service
```proto
service CombatService {
  rpc InitiateCombat(InitiateCombatRequest) returns (CombatSession);
  rpc JoinCombat(JoinCombatRequest) returns (CombatSession);
  rpc ExecuteAction(CombatActionRequest) returns (CombatActionResult);
  rpc GetCombatState(GetCombatStateRequest) returns (CombatSession);
  rpc EndCombat(EndCombatRequest) returns (CombatSummary);
}
```

## Security & Performance

### Security
- Discord OAuth2 for authentication
- JWT tokens for web interface
- Rate limiting on all endpoints
- Input validation with Zod schemas

### Performance
- Redis for fast combat state access
- PostgreSQL connection pooling
- gRPC for efficient service communication
- WebSocket connection pooling

## Deployment Strategy

### Local Development
```bash
docker-compose up -d postgres redis
npm run dev:bot
npm run dev:game
npm run dev:web
```

### Production
- Kubernetes deployment with separate pods per service
- Horizontal scaling for bot and game services
- CloudFlare for web interface CDN
- Managed PostgreSQL and Redis instances

## Future Enhancements

1. **AI Dungeon Master**
   - GPT integration for narrative generation
   - Automated NPC dialogue
   - Dynamic quest generation

2. **Advanced Combat**
   - Spell effects and areas
   - Environmental hazards
   - Multi-enemy encounters

3. **Campaign Management**
   - Persistent world state
   - Quest tracking
   - Party inventory

4. **Enhanced Web Interface**
   - 3D battle maps
   - Fog of war
   - Character portraits
   - Sound effects