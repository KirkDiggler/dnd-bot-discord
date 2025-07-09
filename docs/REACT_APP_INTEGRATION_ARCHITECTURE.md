# React App Integration Architecture

## Overview

This document outlines the architecture for integrating a React web application with the existing Discord bot infrastructure to provide a visual hex-based board interface for tactical positioning and room layout.

## Goals

- **Hex-based Board System**: Provide Gloomhaven-style tactical positioning
- **Real-time Updates**: Synchronize state between Discord bot and React app
- **Unified Backend**: Leverage existing gRPC services and bot infrastructure
- **Room Layout Management**: Enable DM tools for creating and managing battle maps

## Architecture Diagram

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Discord Bot   │    │   React App     │    │   Mobile App    │
│                 │    │                 │    │   (Future)      │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         │                       │                       │
         ▼                       ▼                       ▼
┌───────────────────────────────────────────────────────────────┐
│                    gRPC Services Layer                        │
├─────────────────┬─────────────────┬─────────────────┬─────────┤
│  GameService    │ DungeonService  │ CharacterService│ Combat  │
│                 │                 │                 │ Service │
└─────────────────┴─────────────────┴─────────────────┴─────────┘
         │                       │                       │
         ▼                       ▼                       ▼
┌───────────────────────────────────────────────────────────────┐
│                    Service Layer (Go)                         │
├─────────────────┬─────────────────┬─────────────────┬─────────┤
│  Game Services  │ Dungeon Service │ Character Svc   │ Combat  │
│                 │                 │                 │ Service │
└─────────────────┴─────────────────┴─────────────────┴─────────┘
         │                       │                       │
         ▼                       ▼                       ▼
┌───────────────────────────────────────────────────────────────┐
│                    Repository Layer                           │
├─────────────────┬─────────────────┬─────────────────┬─────────┤
│     Redis       │     D&D 5e      │     MongoDB     │   File  │
│   (Primary)     │      API        │   (Optional)    │ Storage │
└─────────────────┴─────────────────┴─────────────────┴─────────┘
```

## Key Components

### 1. gRPC Services Extension

#### New DungeonService
```proto
service DungeonService {
  // Room management
  rpc CreateRoom(CreateRoomRequest) returns (Room);
  rpc GetRoom(GetRoomRequest) returns (Room);
  rpc UpdateRoomLayout(UpdateRoomLayoutRequest) returns (Room);
  rpc ListRooms(ListRoomsRequest) returns (RoomList);
  
  // Real-time room updates
  rpc SubscribeToRoom(SubscribeToRoomRequest) returns (stream RoomUpdate);
  
  // Hex grid operations
  rpc GetHexGrid(GetHexGridRequest) returns (HexGrid);
  rpc PlaceToken(PlaceTokenRequest) returns (PlaceTokenResponse);
  rpc MoveToken(MoveTokenRequest) returns (MoveTokenResponse);
}
```

#### Extended Position System
```proto
message Position {
  oneof coordinate_system {
    CartesianCoordinate cartesian = 1;  // For existing system
    HexCoordinate hex = 2;              // For new hex system
  }
}

message HexCoordinate {
  int32 q = 1;  // Hex column (axial coordinates)
  int32 r = 2;  // Hex row (axial coordinates)
  int32 s = 3;  // Derived: s = -q - r
}

message HexGrid {
  int32 width = 1;
  int32 height = 2;
  HexOrientation orientation = 3;
  double hex_size = 4;
  repeated HexTile tiles = 5;
}
```

### 2. React Application Architecture

```
react-dnd-client/
├── src/
│   ├── components/
│   │   ├── HexBoard/           # Hex grid rendering
│   │   ├── CharacterSheet/     # Character management
│   │   ├── Combat/             # Combat UI
│   │   └── RoomLayout/         # Room editing tools
│   ├── services/
│   │   ├── grpc/               # gRPC client services
│   │   └── hex/                # Hex coordinate utilities
│   ├── hooks/
│   │   ├── useGrpcStream.ts    # Real-time updates
│   │   └── useHexGrid.ts       # Hex grid state
│   └── types/
│       └── generated/          # Proto-generated types
```

### 3. Real-time Communication

#### gRPC Streaming
- **Server Streaming**: Room updates push to all connected clients
- **Bidirectional Streaming**: Combat actions with immediate feedback
- **Client Streaming**: Batch operations for complex room edits

#### Event Flow
```
Discord Bot Action → gRPC Service → Room Update Stream → React App
React App Action → gRPC Service → Room Update Stream → Discord Bot
```

## Integration Points

### With Existing Systems

#### Character Service
- React app consumes existing Character proto messages
- No changes needed to character management
- Character sheets display in both Discord and React

#### Combat Service
- Existing CombatService extended with hex positioning
- Initiative tracking works across both interfaces
- Combat log synchronized between Discord and React

#### Session Service
- Room layouts tied to existing session management
- DM permissions use existing authorization
- Session state includes hex grid data

### New Components

#### Room Repository
```go
type RoomRepository interface {
    CreateRoom(ctx context.Context, room *Room) error
    GetRoom(ctx context.Context, id string) (*Room, error)
    UpdateRoomLayout(ctx context.Context, id string, layout *HexGrid) error
    ListRooms(ctx context.Context, sessionID string) ([]*Room, error)
}
```

#### Hex Coordinate System
```go
type HexCoordinate struct {
    Q int32 `json:"q"`
    R int32 `json:"r"`
    S int32 `json:"s"`
}

type HexUtils struct{}

func (HexUtils) Distance(a, b HexCoordinate) int32
func (HexUtils) Neighbors(coord HexCoordinate) []HexCoordinate
func (HexUtils) AxialToPixel(coord HexCoordinate, size float64) PixelCoordinate
func (HexUtils) PixelToAxial(pixel PixelCoordinate, size float64) HexCoordinate
```

## Data Flow

### Room Creation
1. DM creates room via Discord bot or React app
2. Room stored in Redis with hex grid layout
3. Room ID shared with players
4. Players join via Discord or React app

### Token Movement
1. Player moves token in React app
2. gRPC MoveToken call to DungeonService
3. Service validates move and updates state
4. Update streamed to all connected clients
5. Discord bot receives update and can announce

### Combat Integration
1. Combat initiated via existing Discord bot flow
2. Combat state includes hex positions
3. React app displays tactical view
4. Actions taken in either interface
5. Results synchronized across both

## Technical Considerations

### Performance
- **Streaming Efficiency**: Only send position deltas, not full state
- **Client-side Caching**: Cache hex grid layouts locally
- **Batched Updates**: Group multiple position changes

### Security
- **Authentication**: Extend existing Discord OAuth
- **Authorization**: Use existing DM/player permissions
- **Input Validation**: Validate hex coordinates server-side

### Scalability
- **Stateless Services**: Keep gRPC services stateless
- **Redis Scaling**: Use Redis clustering for large sessions
- **Client Management**: Efficient stream management

## Migration Strategy

### Phase 1: Proto Extensions
- Add hex coordinate messages to existing proto files
- Extend Position message with hex coordinate option
- Add DungeonService proto definition

### Phase 2: Service Implementation
- Implement DungeonService in Go
- Add hex coordinate utilities
- Create room repository layer

### Phase 3: React Foundation
- Set up React app with gRPC-web
- Implement basic hex grid rendering
- Add token placement/movement

### Phase 4: Integration
- Connect React app to existing services
- Implement real-time streaming
- Add combat synchronization

## Benefits

### For Players
- **Visual Positioning**: See exact tactical positions
- **Better Planning**: Visualize movement and range
- **Rich Interface**: Character sheets and combat in one view

### For DMs
- **Room Designer**: Create custom battle maps
- **Real-time Control**: Manage combat from visual interface
- **Dual Interface**: Use Discord for narrative, React for tactics

### For Development
- **Unified Backend**: Single source of truth
- **Extensible**: Easy to add mobile apps later
- **Maintainable**: Leverages existing architecture patterns

## Next Steps

1. **Review and Approval**: Validate architecture with stakeholders
2. **Proto Definition**: Create detailed proto file extensions
3. **Service Implementation**: Begin DungeonService implementation
4. **React Setup**: Initialize React app with gRPC-web
5. **Integration Testing**: Test real-time synchronization