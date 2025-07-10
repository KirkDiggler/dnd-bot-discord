# Proto File Extensions for Hex-Based Board System

## Overview

This document outlines the specific protobuf extensions needed to support the hex-based board system while maintaining compatibility with the existing Discord bot infrastructure.

## File: `proto/common.proto`

### Extensions to Add

```proto
// Add to existing Position message
message Position {
  oneof coordinate_system {
    CartesianCoordinate cartesian = 1;  // Existing system
    HexCoordinate hex = 2;              // New hex system
  }
}

// New coordinate system for hex grids
message HexCoordinate {
  int32 q = 1;  // Hex column (axial coordinate)
  int32 r = 2;  // Hex row (axial coordinate)
  int32 s = 3;  // Derived: s = -q - r (for validation)
}

// For backward compatibility with existing cartesian system
message CartesianCoordinate {
  int32 x = 1;
  int32 y = 2;
}

// Hex-specific dice result for positioning
message HexRange {
  HexCoordinate center = 1;
  int32 radius = 2;
  repeated HexCoordinate excluded_hexes = 3;
}
```

## File: `proto/dungeon.proto` (New File)

### Complete Service Definition

```proto
syntax = "proto3";

package dnd;

option go_package = "github.com/yourusername/dnd-bot-discord/proto/dnd";

import "common.proto";
import "character.proto";

// Service for managing dungeon rooms and hex grids
service DungeonService {
  // Room management
  rpc CreateRoom(CreateRoomRequest) returns (Room);
  rpc GetRoom(GetRoomRequest) returns (Room);
  rpc UpdateRoom(UpdateRoomRequest) returns (Room);
  rpc DeleteRoom(DeleteRoomRequest) returns (Empty);
  rpc ListRooms(ListRoomsRequest) returns (RoomList);
  
  // Hex grid operations
  rpc GetHexGrid(GetHexGridRequest) returns (HexGrid);
  rpc UpdateHexGrid(UpdateHexGridRequest) returns (HexGrid);
  rpc ValidateHexGrid(ValidateHexGridRequest) returns (HexGridValidation);
  
  // Token management
  rpc PlaceToken(PlaceTokenRequest) returns (PlaceTokenResponse);
  rpc MoveToken(MoveTokenRequest) returns (MoveTokenResponse);
  rpc RemoveToken(RemoveTokenRequest) returns (RemoveTokenResponse);
  rpc GetTokensInRoom(GetTokensInRoomRequest) returns (TokenList);
  
  // Real-time updates
  rpc SubscribeToRoom(SubscribeToRoomRequest) returns (stream RoomUpdate);
  rpc SubscribeToRoomEvents(SubscribeToRoomEventsRequest) returns (stream RoomEvent);
  
  // Pathfinding and range calculations
  rpc FindPath(FindPathRequest) returns (FindPathResponse);
  rpc GetHexesInRange(GetHexesInRangeRequest) returns (GetHexesInRangeResponse);
  rpc GetHexesInArea(GetHexesInAreaRequest) returns (GetHexesInAreaResponse);
  rpc CheckLineOfSight(CheckLineOfSightRequest) returns (CheckLineOfSightResponse);
}

// Room representation
message Room {
  string id = 1;
  string name = 2;
  string description = 3;
  string session_id = 4;
  string created_by = 5;
  
  HexGrid hex_grid = 6;
  repeated RoomFeature features = 7;
  RoomVisibility visibility = 8;
  
  string created_at = 9;
  string updated_at = 10;
}

// Hex grid definition
message HexGrid {
  int32 width = 1;        // Number of hex columns
  int32 height = 2;       // Number of hex rows
  HexOrientation orientation = 3;
  double hex_size = 4;    // Radius in pixels for rendering
  
  repeated HexTile tiles = 5;
  GridBounds bounds = 6;
  string background_image = 7;
  double scale = 8;       // Scale factor for background image
}

// Individual hex tile
message HexTile {
  HexCoordinate coordinate = 1;
  TileType type = 2;
  bool walkable = 3;
  bool blocks_sight = 4;
  bool blocks_movement = 5;
  
  string texture_url = 6;
  string overlay_url = 7;
  repeated TileModifier modifiers = 8;
  
  // Lighting and effects
  double light_level = 9;      // 0.0 to 1.0
  string tint_color = 10;      // Hex color code
  repeated TileEffect effects = 11;
}

// Token representation
message Token {
  string id = 1;
  string name = 2;
  TokenType type = 3;
  HexCoordinate position = 4;
  
  string character_id = 5;     // Links to Character service
  string user_id = 6;          // Owner of the token
  string room_id = 7;          // Room containing the token
  
  TokenSize size = 8;
  string sprite_url = 9;
  TokenState state = 10;
  
  // Visual properties
  double rotation = 11;        // Rotation in degrees
  string tint_color = 12;      // Hex color code
  repeated TokenEffect effects = 13;
  
  // Metadata
  map<string, string> metadata = 14;
  bool visible_to_players = 15;
  
  string created_at = 16;
  string updated_at = 17;
}

// Room features (doors, traps, etc.)
message RoomFeature {
  string id = 1;
  string name = 2;
  FeatureType type = 3;
  HexCoordinate position = 4;
  
  string description = 5;
  bool active = 6;
  bool visible_to_players = 7;
  
  map<string, string> properties = 8;
}

// Grid bounds for optimization
message GridBounds {
  HexCoordinate min = 1;
  HexCoordinate max = 2;
  int32 total_tiles = 3;
}

// Room update events
message RoomUpdate {
  string room_id = 1;
  string timestamp = 2;
  string user_id = 3;
  
  oneof update {
    HexGridUpdate grid_update = 4;
    TokenUpdate token_update = 5;
    FeatureUpdate feature_update = 6;
    RoomSettingsUpdate settings_update = 7;
  }
}

message HexGridUpdate {
  repeated HexTile updated_tiles = 1;
  repeated HexCoordinate removed_tiles = 2;
  GridBounds new_bounds = 3;
}

message TokenUpdate {
  TokenUpdateType type = 1;
  Token token = 2;
  HexCoordinate old_position = 3;
}

message FeatureUpdate {
  FeatureUpdateType type = 1;
  RoomFeature feature = 2;
}

message RoomSettingsUpdate {
  optional string name = 1;
  optional string description = 2;
  optional RoomVisibility visibility = 3;
}

// Room events for real-time updates
message RoomEvent {
  string room_id = 1;
  string timestamp = 2;
  string actor_id = 3;
  
  oneof event {
    TokenMoved token_moved = 4;
    TokenPlaced token_placed = 5;
    TokenRemoved token_removed = 6;
    TileUpdated tile_updated = 7;
    UserJoined user_joined = 8;
    UserLeft user_left = 9;
  }
}

// Event details
message TokenMoved {
  string token_id = 1;
  HexCoordinate from = 2;
  HexCoordinate to = 3;
  repeated HexCoordinate path = 4;
}

message TokenPlaced {
  Token token = 1;
}

message TokenRemoved {
  string token_id = 1;
  HexCoordinate position = 2;
}

message TileUpdated {
  HexTile tile = 1;
  HexTile old_tile = 2;
}

message UserJoined {
  string user_id = 1;
  string username = 2;
  UserRole role = 3;
}

message UserLeft {
  string user_id = 1;
  string username = 2;
}

// Enums
enum HexOrientation {
  HEX_ORIENTATION_UNSPECIFIED = 0;
  HEX_ORIENTATION_POINTY_TOP = 1;
  HEX_ORIENTATION_FLAT_TOP = 2;
}

enum TileType {
  TILE_TYPE_UNSPECIFIED = 0;
  TILE_TYPE_FLOOR = 1;
  TILE_TYPE_WALL = 2;
  TILE_TYPE_DOOR = 3;
  TILE_TYPE_WINDOW = 4;
  TILE_TYPE_WATER = 5;
  TILE_TYPE_LAVA = 6;
  TILE_TYPE_PIT = 7;
  TILE_TYPE_STAIRS_UP = 8;
  TILE_TYPE_STAIRS_DOWN = 9;
  TILE_TYPE_DIFFICULT_TERRAIN = 10;
  TILE_TYPE_RUBBLE = 11;
  TILE_TYPE_ICE = 12;
  TILE_TYPE_SAND = 13;
  TILE_TYPE_GRASS = 14;
  TILE_TYPE_STONE = 15;
}

enum TokenType {
  TOKEN_TYPE_UNSPECIFIED = 0;
  TOKEN_TYPE_PLAYER = 1;
  TOKEN_TYPE_MONSTER = 2;
  TOKEN_TYPE_NPC = 3;
  TOKEN_TYPE_OBJECT = 4;
  TOKEN_TYPE_EFFECT = 5;
  TOKEN_TYPE_MARKER = 6;
  TOKEN_TYPE_TRAP = 7;
  TOKEN_TYPE_TREASURE = 8;
}

enum TokenSize {
  TOKEN_SIZE_UNSPECIFIED = 0;
  TOKEN_SIZE_TINY = 1;       // 1/4 hex
  TOKEN_SIZE_SMALL = 2;      // 1/2 hex
  TOKEN_SIZE_MEDIUM = 3;     // 1 hex
  TOKEN_SIZE_LARGE = 4;      // 2 hexes
  TOKEN_SIZE_HUGE = 5;       // 3 hexes
  TOKEN_SIZE_GARGANTUAN = 6; // 4+ hexes
}

enum TokenState {
  TOKEN_STATE_UNSPECIFIED = 0;
  TOKEN_STATE_ACTIVE = 1;
  TOKEN_STATE_INACTIVE = 2;
  TOKEN_STATE_HIDDEN = 3;
  TOKEN_STATE_PRONE = 4;
  TOKEN_STATE_UNCONSCIOUS = 5;
  TOKEN_STATE_DEAD = 6;
}

enum FeatureType {
  FEATURE_TYPE_UNSPECIFIED = 0;
  FEATURE_TYPE_DOOR = 1;
  FEATURE_TYPE_TRAP = 2;
  FEATURE_TYPE_TREASURE = 3;
  FEATURE_TYPE_ALTAR = 4;
  FEATURE_TYPE_FOUNTAIN = 5;
  FEATURE_TYPE_STATUE = 6;
  FEATURE_TYPE_PORTAL = 7;
  FEATURE_TYPE_CAMPFIRE = 8;
  FEATURE_TYPE_CHEST = 9;
  FEATURE_TYPE_LEVER = 10;
  FEATURE_TYPE_BUTTON = 11;
}

enum RoomVisibility {
  ROOM_VISIBILITY_UNSPECIFIED = 0;
  ROOM_VISIBILITY_PUBLIC = 1;
  ROOM_VISIBILITY_PRIVATE = 2;
  ROOM_VISIBILITY_DM_ONLY = 3;
}

enum TokenUpdateType {
  TOKEN_UPDATE_TYPE_UNSPECIFIED = 0;
  TOKEN_UPDATE_TYPE_CREATED = 1;
  TOKEN_UPDATE_TYPE_MOVED = 2;
  TOKEN_UPDATE_TYPE_UPDATED = 3;
  TOKEN_UPDATE_TYPE_REMOVED = 4;
}

enum FeatureUpdateType {
  FEATURE_UPDATE_TYPE_UNSPECIFIED = 0;
  FEATURE_UPDATE_TYPE_CREATED = 1;
  FEATURE_UPDATE_TYPE_UPDATED = 2;
  FEATURE_UPDATE_TYPE_REMOVED = 3;
  FEATURE_UPDATE_TYPE_ACTIVATED = 4;
  FEATURE_UPDATE_TYPE_DEACTIVATED = 5;
}

enum UserRole {
  USER_ROLE_UNSPECIFIED = 0;
  USER_ROLE_PLAYER = 1;
  USER_ROLE_DM = 2;
  USER_ROLE_OBSERVER = 3;
}

// Tile and token effects
message TileModifier {
  string type = 1;        // "difficult_terrain", "slippery", etc.
  string value = 2;       // Modifier value if applicable
  int32 duration = 3;     // Duration in rounds, -1 for permanent
}

message TileEffect {
  string id = 1;
  string name = 2;
  string description = 3;
  string visual_effect = 4;  // Animation or particle effect
  int32 remaining_rounds = 5;
  string source_id = 6;      // What created this effect
}

message TokenEffect {
  string id = 1;
  string name = 2;
  string description = 3;
  string visual_effect = 4;
  int32 remaining_rounds = 5;
  string source_id = 6;
}

// Request/Response messages
message CreateRoomRequest {
  string name = 1;
  string description = 2;
  string session_id = 3;
  string created_by = 4;
  HexGrid initial_grid = 5;
  RoomVisibility visibility = 6;
}

message GetRoomRequest {
  string room_id = 1;
  string user_id = 2;  // For permission checking
}

message UpdateRoomRequest {
  string room_id = 1;
  string user_id = 2;
  optional string name = 3;
  optional string description = 4;
  optional RoomVisibility visibility = 5;
}

message DeleteRoomRequest {
  string room_id = 1;
  string user_id = 2;
}

message ListRoomsRequest {
  string session_id = 1;
  string user_id = 2;
  bool include_private = 3;
}

message RoomList {
  repeated Room rooms = 1;
  int32 total_count = 2;
}

message GetHexGridRequest {
  string room_id = 1;
  string user_id = 2;
  bool include_hidden_tiles = 3;
}

message UpdateHexGridRequest {
  string room_id = 1;
  string user_id = 2;
  repeated HexTile tiles = 3;
  optional GridBounds bounds = 4;
}

message ValidateHexGridRequest {
  HexGrid grid = 1;
}

message HexGridValidation {
  bool valid = 1;
  repeated string errors = 2;
  repeated string warnings = 3;
}

message PlaceTokenRequest {
  string room_id = 1;
  string user_id = 2;
  Token token = 3;
  bool validate_placement = 4;
}

message PlaceTokenResponse {
  bool success = 1;
  Token token = 2;
  string error_message = 3;
  repeated string warnings = 4;
}

message MoveTokenRequest {
  string room_id = 1;
  string user_id = 2;
  string token_id = 3;
  HexCoordinate new_position = 4;
  bool validate_movement = 5;
  bool animate = 6;
}

message MoveTokenResponse {
  bool success = 1;
  Token token = 2;
  repeated HexCoordinate path = 3;
  string error_message = 4;
  int32 movement_cost = 5;
}

message RemoveTokenRequest {
  string room_id = 1;
  string user_id = 2;
  string token_id = 3;
}

message RemoveTokenResponse {
  bool success = 1;
  string error_message = 2;
}

message GetTokensInRoomRequest {
  string room_id = 1;
  string user_id = 2;
  bool include_hidden = 3;
}

message TokenList {
  repeated Token tokens = 1;
  int32 total_count = 2;
}

message SubscribeToRoomRequest {
  string room_id = 1;
  string user_id = 2;
  bool include_initial_state = 3;
}

message SubscribeToRoomEventsRequest {
  string room_id = 1;
  string user_id = 2;
  repeated string event_types = 3;  // Filter for specific event types
}

// Pathfinding and range calculation requests
message FindPathRequest {
  string room_id = 1;
  HexCoordinate start = 2;
  HexCoordinate end = 3;
  PathfindingOptions options = 4;
}

message FindPathResponse {
  bool path_found = 1;
  repeated HexCoordinate path = 2;
  int32 total_cost = 3;
  string error_message = 4;
}

message PathfindingOptions {
  bool ignore_difficult_terrain = 1;
  bool ignore_other_tokens = 2;
  TokenSize token_size = 3;
  int32 max_distance = 4;
}

message GetHexesInRangeRequest {
  string room_id = 1;
  HexCoordinate center = 2;
  int32 range = 3;
  RangeOptions options = 4;
}

message GetHexesInRangeResponse {
  repeated HexCoordinate hexes = 1;
  int32 total_count = 2;
}

message RangeOptions {
  bool require_line_of_sight = 1;
  bool include_blocked_hexes = 2;
  TokenSize token_size = 3;
}

message GetHexesInAreaRequest {
  string room_id = 1;
  HexCoordinate center = 2;
  AreaType area_type = 3;
  int32 size = 4;
  optional HexCoordinate direction = 5;  // For cones and lines
}

message GetHexesInAreaResponse {
  repeated HexCoordinate hexes = 1;
  int32 total_count = 2;
}

enum AreaType {
  AREA_TYPE_UNSPECIFIED = 0;
  AREA_TYPE_CIRCLE = 1;
  AREA_TYPE_SQUARE = 2;
  AREA_TYPE_CONE = 3;
  AREA_TYPE_LINE = 4;
  AREA_TYPE_CUBE = 5;
}

message CheckLineOfSightRequest {
  string room_id = 1;
  HexCoordinate start = 2;
  HexCoordinate end = 3;
  LineOfSightOptions options = 4;
}

message CheckLineOfSightResponse {
  bool has_line_of_sight = 1;
  repeated HexCoordinate blocking_hexes = 2;
  repeated HexCoordinate line_hexes = 3;
}

message LineOfSightOptions {
  bool ignore_tokens = 1;
  TokenSize token_size = 2;
  bool strict_corners = 3;
}
```

## File: `proto/combat.proto`

### Extensions to Add

```proto
// Add to existing CombatParticipant message
message CombatParticipant {
  // ... existing fields ...
  
  // New hex positioning fields
  HexCoordinate hex_position = 20;
  int32 movement_remaining = 21;
  repeated HexCoordinate movement_path = 22;
  
  // Range and area indicators
  repeated HexCoordinate attack_range = 23;
  repeated HexCoordinate spell_range = 24;
  repeated HexCoordinate movement_range = 25;
}

// Add to existing CombatSession message
message CombatSession {
  // ... existing fields ...
  
  // New hex-specific fields
  string room_id = 20;              // Link to room with hex grid
  bool use_hex_positioning = 21;    // Enable hex positioning for this combat
  HexGrid combat_grid = 22;         // Snapshot of grid at combat start
  
  // Movement tracking
  map<string, int32> movement_used = 23;    // participant_id -> movement used
  map<string, HexCoordinate> starting_positions = 24;
}

// New combat actions for hex positioning
message HexMoveAction {
  string participant_id = 1;
  HexCoordinate from = 2;
  HexCoordinate to = 3;
  repeated HexCoordinate path = 4;
  int32 movement_cost = 5;
  bool dash = 6;
}

message HexAttackAction {
  string attacker_id = 1;
  string target_id = 2;
  HexCoordinate attacker_position = 3;
  HexCoordinate target_position = 4;
  string weapon_id = 5;
  int32 range = 6;
  bool has_line_of_sight = 7;
}

message HexSpellAction {
  string caster_id = 1;
  string spell_id = 2;
  HexCoordinate caster_position = 3;
  HexCoordinate target_position = 4;
  repeated HexCoordinate area_of_effect = 5;
  int32 spell_level = 6;
  repeated string target_ids = 7;
}

// Add to existing CombatActionRequest
message CombatActionRequest {
  // ... existing fields ...
  
  // New hex-specific actions
  oneof hex_action {
    HexMoveAction hex_move = 10;
    HexAttackAction hex_attack = 11;
    HexSpellAction hex_spell = 12;
  }
}

// Add to existing CombatLogEntry
message CombatLogEntry {
  // ... existing fields ...
  
  // New hex-specific details
  oneof hex_details {
    HexMovementDetails hex_movement = 10;
    HexAttackDetails hex_attack = 11;
    HexSpellDetails hex_spell = 12;
  }
}

message HexMovementDetails {
  HexCoordinate from = 1;
  HexCoordinate to = 2;
  repeated HexCoordinate path = 3;
  int32 distance = 4;
  int32 movement_cost = 5;
  bool used_dash = 6;
}

message HexAttackDetails {
  HexCoordinate attacker_position = 1;
  HexCoordinate target_position = 2;
  int32 distance = 3;
  bool had_line_of_sight = 4;
  string weapon_used = 5;
  // ... existing attack details ...
}

message HexSpellDetails {
  HexCoordinate caster_position = 1;
  HexCoordinate target_position = 2;
  repeated HexCoordinate area_of_effect = 3;
  int32 spell_level = 4;
  string spell_name = 5;
  repeated string affected_participants = 6;
}
```

## File: `proto/game.proto`

### Extensions to Add

```proto
// Add to existing GameService
service GameService {
  // ... existing methods ...
  
  // New room management methods
  rpc CreateRoomForSession(CreateRoomForSessionRequest) returns (Room);
  rpc GetSessionRoom(GetSessionRoomRequest) returns (Room);
  rpc SetSessionRoom(SetSessionRoomRequest) returns (SessionRoomResponse);
  
  // Integration methods
  rpc StartCombatInRoom(StartCombatInRoomRequest) returns (CombatSession);
  rpc LinkCharacterToToken(LinkCharacterToTokenRequest) returns (LinkCharacterToTokenResponse);
}

// New request/response messages
message CreateRoomForSessionRequest {
  string session_id = 1;
  string user_id = 2;
  string room_name = 3;
  HexGrid initial_grid = 4;
}

message GetSessionRoomRequest {
  string session_id = 1;
  string user_id = 2;
}

message SetSessionRoomRequest {
  string session_id = 1;
  string room_id = 2;
  string user_id = 3;
}

message SessionRoomResponse {
  bool success = 1;
  string error_message = 2;
  Room room = 3;
}

message StartCombatInRoomRequest {
  string session_id = 1;
  string room_id = 2;
  string initiator_id = 3;
  repeated string monster_ids = 4;
  bool use_hex_positioning = 5;
}

message LinkCharacterToTokenRequest {
  string character_id = 1;
  string token_id = 2;
  string user_id = 3;
}

message LinkCharacterToTokenResponse {
  bool success = 1;
  string error_message = 2;
  Token token = 3;
}
```

## Migration Strategy

### Step 1: Update Existing Files
1. Add `HexCoordinate` and `CartesianCoordinate` to `common.proto`
2. Update `Position` message to use oneof for coordinate systems
3. Add hex-specific fields to `combat.proto` messages

### Step 2: Add New File
1. Create `dungeon.proto` with complete service definition
2. Generate Go and TypeScript bindings
3. Validate proto compilation

### Step 3: Implement Services
1. Implement `DungeonService` in Go
2. Add hex coordinate utilities
3. Create room repository layer
4. Add integration with existing services

### Step 4: Update Clients
1. Update Discord bot to handle hex coordinates
2. Implement React client with new proto messages
3. Add backward compatibility where needed

## Backward Compatibility

### Existing Position Messages
- Use `cartesian` field in Position oneof for existing functionality
- Default to cartesian coordinates when hex not specified
- Maintain existing API behavior

### Combat System
- Add `use_hex_positioning` flag to enable hex features
- Fall back to existing positioning when flag is false
- Maintain existing combat flow for Discord-only usage

### Database Schema
- Add new Redis keys for hex data
- Keep existing character and session data unchanged
- Use migration scripts for data format updates

## Testing Strategy

### Proto Validation
- Compile all proto files without errors
- Generate bindings for Go and TypeScript
- Validate message serialization/deserialization

### Integration Testing
- Test backward compatibility with existing Discord bot
- Verify hex coordinate conversion functions
- Test streaming with multiple clients

### Performance Testing
- Validate large hex grid performance
- Test real-time updates with many tokens
- Benchmark pathfinding algorithms

## Code Generation Commands

```bash
# Generate Go bindings
protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    proto/*.proto

# Generate TypeScript bindings
protoc --plugin=protoc-gen-ts_proto=./node_modules/.bin/protoc-gen-ts_proto \
    --ts_proto_out=src/types/generated \
    --ts_proto_opt=esModuleInterop=true \
    proto/*.proto
```

## File Structure After Extension

```
proto/
├── common.proto           # Updated with hex coordinates
├── character.proto        # Unchanged
├── combat.proto          # Extended with hex positioning
├── combat_streaming.proto # May need hex event types
├── game.proto            # Extended with room management
└── dungeon.proto         # New file for room/hex management
```

This extension strategy maintains backward compatibility while adding the new hex-based positioning system needed for the React application.