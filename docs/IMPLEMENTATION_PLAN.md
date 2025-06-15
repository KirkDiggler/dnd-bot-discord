# Implementation Plan

## Phase 1: Foundation (Week 1-2)

### 1.1 Project Setup
- [ ] Initialize TypeScript project with proper config
- [ ] Set up Docker Compose for PostgreSQL and Redis
- [ ] Configure Prisma with initial schema
- [ ] Set up proto files and code generation
- [ ] Configure ESLint and Prettier
- [ ] Set up Git hooks (Husky) for code quality

### 1.2 Core Services Structure
- [ ] Create gRPC game service skeleton
- [ ] Create Discord bot service skeleton
- [ ] Implement proto compilation pipeline
- [ ] Set up service communication
- [ ] Create shared types package

### 1.3 Repository Layer
- [ ] Implement base repository interface
- [ ] Create PostgreSQL repositories (Prisma)
- [ ] Create Redis repositories
- [ ] Add repository factory pattern
- [ ] Write repository unit tests

## Phase 2: Character System (Week 3-4)

### 2.1 Character Creation
- [ ] Port character entity from Go code
- [ ] Implement character creation flow
- [ ] Add D&D 5e data (races, classes, backgrounds)
- [ ] Create character repository
- [ ] Add character validation

### 2.2 Discord Commands
- [ ] `/character create` - Start creation wizard
- [ ] `/character list` - View your characters
- [ ] `/character view [name]` - View character sheet
- [ ] `/character select [name]` - Set active character
- [ ] Character sheet as ephemeral message

### 2.3 Character Management Service
- [ ] gRPC character service implementation
- [ ] Character state management
- [ ] Equipment system
- [ ] Proficiency calculations
- [ ] Ability score modifiers

## Phase 3: Basic Combat (Week 5-6)

### 3.1 Combat Engine
- [ ] Port combat mechanics from Go
- [ ] Initiative system
- [ ] Turn order management
- [ ] Basic attack actions
- [ ] Damage calculation
- [ ] HP tracking

### 3.2 Combat Discord Integration
- [ ] `/combat start` - Initiate encounter
- [ ] `/combat join` - Join active combat
- [ ] Combat state in shared message
- [ ] Action buttons (Attack, Move, End Turn)
- [ ] Private roll results (ephemeral)

### 3.3 State Management
- [ ] Redis combat session storage
- [ ] Combat state machine
- [ ] Action validation
- [ ] Combat history/log
- [ ] Auto-cleanup expired sessions

## Phase 4: Web Interface (Week 7-8)

### 4.1 Basic Web App
- [ ] React app setup
- [ ] gRPC-Web client configuration
- [ ] WebSocket connection to game service
- [ ] Authentication with Discord OAuth
- [ ] Basic routing

### 4.2 Battle Map
- [ ] Grid-based map component
- [ ] Character token rendering
- [ ] Drag-and-drop movement
- [ ] Initiative tracker sidebar
- [ ] HP/status indicators

### 4.3 Real-time Integration
- [ ] WebSocket event handling
- [ ] State synchronization
- [ ] Optimistic UI updates
- [ ] Error recovery
- [ ] Connection status indicator

## Phase 5: Enhanced Combat (Week 9-10)

### 5.1 Advanced Actions
- [ ] Movement system with speed limits
- [ ] Ranged attacks with distance
- [ ] Area of effect indicators
- [ ] Advantage/disadvantage
- [ ] Conditions (stunned, prone, etc.)

### 5.2 Spellcasting
- [ ] Spell list management
- [ ] Spell slot tracking
- [ ] Spell effects implementation
- [ ] Concentration mechanics
- [ ] Spell components

### 5.3 Monster Integration
- [ ] Monster templates
- [ ] Monster AI (basic)
- [ ] Multi-enemy encounters
- [ ] Challenge rating balance
- [ ] Loot generation

## Phase 6: Polish & AI Integration (Week 11-12)

### 6.1 AI Dungeon Master Prep
- [ ] Design AI service architecture
- [ ] Create narrative generation prompts
- [ ] Implement context management
- [ ] Add story state tracking
- [ ] Design AI-friendly interfaces

### 6.2 Quality of Life
- [ ] Combat recap/summary
- [ ] Character backup/restore
- [ ] Battle replay system
- [ ] Performance optimizations
- [ ] Error handling improvements

### 6.3 Testing & Documentation
- [ ] Integration test suite
- [ ] Load testing for concurrent combats
- [ ] API documentation
- [ ] User guide
- [ ] DM guide

## Technical Milestones

### Milestone 1: Hello World
- Bot responds to `/ping` command
- gRPC service returns health check
- Services communicate successfully

### Milestone 2: First Character
- Create a character through Discord
- View character sheet
- Character persisted to PostgreSQL

### Milestone 3: First Attack
- Start combat with one player
- Make an attack roll
- See damage applied

### Milestone 4: First Battle Map
- Web interface shows grid
- Character token visible
- Movement updates in real-time

### Milestone 5: Complete Encounter
- Multiple participants
- Full combat with winner
- Battle history available

## Development Workflow

### Daily Development
1. Pull latest code
2. Run `docker-compose up -d` for services
3. Run `npm run dev` for hot reload
4. Run `npm test:watch` for tests

### Code Generation
```bash
# Generate types from protos
npm run proto:generate

# Generate Prisma client
npm run prisma:generate

# Run database migrations
npm run prisma:migrate
```

### Testing Commands
```bash
# Unit tests
npm test

# Integration tests
npm run test:integration

# E2E tests
npm run test:e2e

# Load tests
npm run test:load
```

## Risk Mitigation

### Technical Risks
1. **Discord Rate Limits**
   - Solution: Implement exponential backoff
   - Cache frequently accessed data
   - Batch operations where possible

2. **WebSocket Scalability**
   - Solution: Implement Socket.io adapter for Redis
   - Horizontal scaling with sticky sessions
   - Connection pooling

3. **Complex State Management**
   - Solution: Event sourcing for combat
   - Comprehensive logging
   - State snapshot capability

### Architecture Decisions to Revisit
1. **Monorepo vs Separate Repos**
   - Start with monorepo
   - Split if services grow large

2. **GraphQL vs gRPC for Web**
   - Start with gRPC-Web
   - Add GraphQL gateway if needed

3. **TypeORM vs Prisma**
   - Start with Prisma for simplicity
   - Migrate if need more control