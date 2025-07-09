# React App Implementation Plan

## Overview

This document outlines the step-by-step implementation plan for integrating a React web application with hex-based board system into the existing Discord bot infrastructure.

## Phase 1: gRPC Service Extension (Estimated: 1-2 weeks)

### 1.1 Proto File Extensions

**Files to modify:**
- `proto/common.proto`
- `proto/game.proto`
- `proto/combat.proto`

**Tasks:**
- [ ] Add `HexCoordinate` message to `common.proto`
- [ ] Extend `Position` message with oneof for hex coordinates
- [ ] Create `dungeon.proto` with DungeonService definition
- [ ] Add hex-related messages (HexGrid, HexTile, Token)
- [ ] Update `combat.proto` to include hex positioning
- [ ] Generate Go and TypeScript bindings

**Acceptance Criteria:**
- Proto files compile without errors
- Generated Go code builds successfully
- TypeScript types are generated correctly

### 1.2 Go Service Implementation

**Files to create:**
- `internal/services/dungeon/service.go`
- `internal/services/dungeon/hex_utils.go`
- `internal/repositories/rooms/interface.go`
- `internal/repositories/rooms/redis.go`

**Tasks:**
- [ ] Implement `DungeonService` gRPC handlers
- [ ] Create hex coordinate utility functions
- [ ] Implement room repository with Redis storage
- [ ] Add hex grid serialization/deserialization
- [ ] Create token management functions
- [ ] Add room layout validation

**Acceptance Criteria:**
- All gRPC methods implemented and tested
- Redis storage works for room data
- Hex coordinate math functions pass unit tests
- Integration tests pass

### 1.3 Database Schema Updates

**Tasks:**
- [ ] Design Redis key patterns for rooms and hex grids
- [ ] Create room data structures
- [ ] Implement room persistence layer
- [ ] Add migration scripts if needed

**Redis Key Patterns:**
```
room:{room_id}                    # Room metadata and hex grid
room:{room_id}:tokens            # Token positions
session:{session_id}:room        # Session -> Room mapping
```

**Acceptance Criteria:**
- Room data persists correctly in Redis
- Token positions are stored and retrieved accurately
- Migration scripts run without errors

## Phase 2: React App Foundation (Estimated: 2-3 weeks)

### 2.1 Project Setup

**Tasks:**
- [ ] Create React app with TypeScript template
- [ ] Set up gRPC-web client dependencies
- [ ] Configure build tools (Webpack, Vite, etc.)
- [ ] Set up testing framework (Jest, React Testing Library)
- [ ] Configure linting and formatting (ESLint, Prettier)
- [ ] Set up development server with hot reload

**Dependencies:**
```json
{
  "dependencies": {
    "@grpc/grpc-js": "^1.9.0",
    "grpc-web": "^1.4.2",
    "react": "^18.2.0",
    "react-dom": "^18.2.0",
    "typescript": "^5.0.0"
  }
}
```

**Acceptance Criteria:**
- React app runs in development mode
- gRPC-web client can connect to Go server
- Build process works without errors
- Tests can be run successfully

### 2.2 Core Components

**Files to create:**
- `src/components/HexBoard/HexGrid.tsx`
- `src/components/HexBoard/HexTile.tsx`
- `src/components/HexBoard/HexToken.tsx`
- `src/services/hex/coordinates.ts`
- `src/services/hex/utils.ts`

**Tasks:**
- [ ] Implement `HexGrid` component with SVG rendering
- [ ] Create `HexTile` component with click handling
- [ ] Implement `HexToken` component with drag support
- [ ] Build hex coordinate utility functions
- [ ] Add pathfinding algorithms
- [ ] Implement range and area calculations

**Acceptance Criteria:**
- Hex grid renders correctly in browser
- Tiles can be clicked and highlighted
- Tokens can be placed and moved
- Hex math functions pass unit tests

### 2.3 gRPC Client Integration

**Files to create:**
- `src/services/grpc/client.ts`
- `src/services/grpc/dungeon.service.ts`
- `src/services/grpc/game.service.ts`
- `src/hooks/useGrpcStream.ts`

**Tasks:**
- [ ] Set up gRPC client with proper authentication
- [ ] Implement service wrapper classes
- [ ] Create React hooks for gRPC operations
- [ ] Add error handling and retry logic
- [ ] Implement streaming connections
- [ ] Add connection status monitoring

**Acceptance Criteria:**
- gRPC client can call all service methods
- Streaming connections work reliably
- Error handling is robust
- Connection status is displayed to users

## Phase 3: Real-time Integration (Estimated: 1-2 weeks)

### 3.1 Streaming Implementation

**Files to create:**
- `src/hooks/useRoomStream.ts`
- `src/hooks/useCombatStream.ts`
- `src/context/RoomContext.tsx`

**Tasks:**
- [ ] Implement room update streaming
- [ ] Create combat event streaming
- [ ] Add real-time token position updates
- [ ] Implement optimistic updates
- [ ] Add conflict resolution
- [ ] Create connection recovery logic

**Acceptance Criteria:**
- Multiple clients see updates in real-time
- Token movements are synchronized
- Combat events are properly streamed
- Connection drops are handled gracefully

### 3.2 State Management

**Files to create:**
- `src/store/roomStore.ts`
- `src/store/combatStore.ts`
- `src/store/characterStore.ts`

**Tasks:**
- [ ] Implement Redux or Zustand store
- [ ] Create state slices for rooms, combat, characters
- [ ] Add optimistic update patterns
- [ ] Implement state persistence
- [ ] Add undo/redo functionality

**Acceptance Criteria:**
- State is managed consistently
- Optimistic updates work smoothly
- State persists across page reloads
- Undo/redo works for token movements

### 3.3 Discord Bot Integration

**Files to modify:**
- `internal/handlers/discord/handler.go`
- `internal/services/encounter/service.go`

**Tasks:**
- [ ] Add hex coordinate support to existing combat
- [ ] Create Discord commands for room management
- [ ] Implement position synchronization
- [ ] Add web app notifications in Discord
- [ ] Create DM controls for room visibility

**Acceptance Criteria:**
- Discord bot can create and manage rooms
- Position changes sync between Discord and React
- Combat works with hex positioning
- DM can control room access from Discord

## Phase 4: Advanced Features (Estimated: 2-3 weeks)

### 4.1 Combat Integration

**Files to create:**
- `src/components/Combat/InitiativeTracker.tsx`
- `src/components/Combat/ActionPanel.tsx`
- `src/components/Combat/CombatLog.tsx`

**Tasks:**
- [ ] Implement initiative tracking UI
- [ ] Create action selection interface
- [ ] Add combat log display
- [ ] Implement turn highlighting
- [ ] Add movement and attack range visualization
- [ ] Create area of effect indicators

**Acceptance Criteria:**
- Initiative tracker shows correct turn order
- Players can select actions from UI
- Combat log displays all actions
- Turn indicators are clear and visible

### 4.2 Character Sheet Integration

**Files to create:**
- `src/components/Character/CharacterSheet.tsx`
- `src/components/Character/Equipment.tsx`
- `src/components/Character/Abilities.tsx`

**Tasks:**
- [ ] Display character information in React
- [ ] Show equipment and abilities
- [ ] Add character action buttons
- [ ] Implement resource tracking (HP, spell slots)
- [ ] Create character selection interface

**Acceptance Criteria:**
- Character sheets display correctly
- Equipment and abilities are shown
- Resource tracking is accurate
- Character selection works smoothly

### 4.3 Room Editor

**Files to create:**
- `src/components/RoomEditor/RoomEditor.tsx`
- `src/components/RoomEditor/TileSelector.tsx`
- `src/components/RoomEditor/TokenPalette.tsx`

**Tasks:**
- [ ] Create room editing interface
- [ ] Implement tile painting tools
- [ ] Add token placement controls
- [ ] Create room template system
- [ ] Add import/export functionality
- [ ] Implement room sharing

**Acceptance Criteria:**
- DMs can create and edit rooms
- Tile painting works intuitively
- Token placement is easy to use
- Room templates can be saved and loaded

## Phase 5: Polish and Performance (Estimated: 1-2 weeks)

### 5.1 Performance Optimization

**Tasks:**
- [ ] Implement viewport culling for large grids
- [ ] Add object pooling for hex tiles
- [ ] Optimize React rendering with memo/useMemo
- [ ] Implement lazy loading for room data
- [ ] Add caching for calculated values
- [ ] Optimize network requests

**Acceptance Criteria:**
- Large grids render smoothly
- Memory usage is reasonable
- Network requests are minimized
- User interactions feel responsive

### 5.2 User Experience Polish

**Tasks:**
- [ ] Add animations for token movement
- [ ] Implement drag and drop improvements
- [ ] Add keyboard shortcuts
- [ ] Create loading states and error messages
- [ ] Add tooltips and help text
- [ ] Implement responsive design

**Acceptance Criteria:**
- Animations enhance user experience
- Drag and drop feels natural
- Keyboard shortcuts work intuitively
- Error messages are helpful

### 5.3 Testing and Quality

**Tasks:**
- [ ] Write unit tests for hex math functions
- [ ] Create integration tests for gRPC services
- [ ] Add end-to-end tests for user flows
- [ ] Implement visual regression testing
- [ ] Add performance benchmarks
- [ ] Create accessibility tests

**Acceptance Criteria:**
- Test coverage is above 80%
- All critical user flows are tested
- Performance benchmarks are met
- Accessibility standards are followed

## Deployment and Infrastructure

### 6.1 Build and Deployment

**Tasks:**
- [ ] Set up production build process
- [ ] Configure Docker containers
- [ ] Set up CI/CD pipeline
- [ ] Configure environment variables
- [ ] Set up monitoring and logging
- [ ] Create deployment scripts

**Acceptance Criteria:**
- Production builds are optimized
- Deployment process is automated
- Monitoring is in place
- Logging is comprehensive

### 6.2 Security

**Tasks:**
- [ ] Implement authentication (Discord OAuth)
- [ ] Add authorization for room access
- [ ] Validate all gRPC inputs
- [ ] Implement rate limiting
- [ ] Add HTTPS termination
- [ ] Configure CORS properly

**Acceptance Criteria:**
- Authentication works correctly
- Authorization is enforced
- Input validation is comprehensive
- Rate limiting prevents abuse

## Testing Strategy

### Unit Tests
- Hex coordinate math functions
- Component rendering
- gRPC service methods
- State management logic

### Integration Tests
- gRPC client-server communication
- Real-time streaming
- Database operations
- Discord bot integration

### End-to-End Tests
- Complete user workflows
- Multi-user scenarios
- Combat sequences
- Room editing flows

### Performance Tests
- Large grid rendering
- Many concurrent users
- Real-time update latency
- Memory usage monitoring

## Risk Mitigation

### Technical Risks

**gRPC-web Limitations**
- Risk: Limited browser support or features
- Mitigation: Thorough testing, fallback options

**Real-time Synchronization**
- Risk: Race conditions or inconsistent state
- Mitigation: Optimistic updates, conflict resolution

**Performance with Large Grids**
- Risk: Slow rendering or high memory usage
- Mitigation: Viewport culling, object pooling

### Project Risks

**Scope Creep**
- Risk: Adding too many features
- Mitigation: Strict phase boundaries, MVP focus

**Integration Complexity**
- Risk: Discord bot integration issues
- Mitigation: Incremental integration, thorough testing

**User Adoption**
- Risk: Users prefer Discord-only interface
- Mitigation: User testing, feedback collection

## Success Metrics

### Technical Metrics
- Page load time < 3 seconds
- Token movement latency < 100ms
- Memory usage < 200MB for typical session
- 99.9% uptime

### User Metrics
- User session length
- Feature usage statistics
- Error rate < 1%
- User satisfaction scores

### Business Metrics
- Monthly active users
- Session frequency
- Feature adoption rates
- User retention

## Conclusion

This implementation plan provides a structured approach to building the React app integration. Each phase builds on the previous one, allowing for incremental delivery and testing. The plan includes comprehensive testing, performance optimization, and risk mitigation strategies.

The phased approach allows for early feedback and course correction, ensuring the final product meets user needs while maintaining technical quality. Regular checkpoints and clear acceptance criteria help track progress and maintain momentum throughout the development process.