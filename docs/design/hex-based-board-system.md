# Hex-Based Board System Design

## Overview

This document details the design of a hexagonal grid system for tactical positioning in D&D combat, inspired by Gloomhaven's board mechanics. The system provides precise positioning while maintaining the flexibility needed for various room layouts and combat scenarios.

## Why Hexagonal Grids?

### Advantages over Square Grids
- **Consistent Distance**: All adjacent tiles are equidistant
- **Natural Movement**: 6 directions feel more organic than 4 or 8
- **Better Diagonals**: No awkward diagonal movement rules
- **Visual Appeal**: More aesthetically pleasing than square grids
- **Gloomhaven Familiarity**: Players already know this system

### Disadvantages
- **Complexity**: More complex coordinate math
- **Rendering**: SVG/Canvas rendering more involved
- **Pathfinding**: Algorithms need hex-specific implementations

## Coordinate System

### Axial Coordinates
We use the axial coordinate system (q, r, s) where:
- `q`: Column (x-axis equivalent)
- `r`: Row (y-axis equivalent)  
- `s`: Derived value (s = -q - r)

```
     / \ / \ / \
    | 0,1 | 1,0 | 2,0 |
   / \ / \ / \ / \ / \
  |-1,1 | 0,0 | 1,0 | 2,0 |
 / \ / \ / \ / \ / \ / \
|-1,2 | 0,1 | 1,1 | 2,1 |
 \ / \ / \ / \ / \ / \ /
  \ / \ / \ / \ / \ /
```

### Coordinate Conversion

#### Axial to Pixel (Pointy-Top)
```javascript
function axialToPixel(q, r, size) {
  const x = size * (Math.sqrt(3) * q + Math.sqrt(3)/2 * r);
  const y = size * (3/2 * r);
  return { x, y };
}
```

#### Pixel to Axial (Pointy-Top)
```javascript
function pixelToAxial(x, y, size) {
  const q = (Math.sqrt(3)/3 * x - 1/3 * y) / size;
  const r = (2/3 * y) / size;
  return hexRound(q, r);
}
```

#### Hex Rounding
```javascript
function hexRound(q, r) {
  const s = -q - r;
  let rq = Math.round(q);
  let rr = Math.round(r);
  let rs = Math.round(s);
  
  const q_diff = Math.abs(rq - q);
  const r_diff = Math.abs(rr - r);
  const s_diff = Math.abs(rs - s);
  
  if (q_diff > r_diff && q_diff > s_diff) {
    rq = -rr - rs;
  } else if (r_diff > s_diff) {
    rr = -rq - rs;
  } else {
    rs = -rq - rr;
  }
  
  return { q: rq, r: rr, s: rs };
}
```

## Hex Grid Components

### HexTile Properties
```proto
message HexTile {
  HexCoordinate coordinate = 1;
  TileType type = 2;
  bool walkable = 3;
  bool blocks_sight = 4;
  string texture = 5;
  repeated string modifiers = 6;
  optional string overlay = 7;
}

enum TileType {
  TILE_TYPE_UNSPECIFIED = 0;
  TILE_TYPE_FLOOR = 1;
  TILE_TYPE_WALL = 2;
  TILE_TYPE_DOOR = 3;
  TILE_TYPE_WATER = 4;
  TILE_TYPE_DIFFICULT_TERRAIN = 5;
  TILE_TYPE_PIT = 6;
  TILE_TYPE_STAIRS = 7;
  TILE_TYPE_RUBBLE = 8;
}
```

### Token System
```proto
message Token {
  string id = 1;
  string name = 2;
  TokenType type = 3;
  HexCoordinate position = 4;
  string character_id = 5;
  string sprite_url = 6;
  TokenState state = 7;
  TokenSize size = 8;
  optional string overlay_effect = 9;
}

enum TokenType {
  TOKEN_TYPE_UNSPECIFIED = 0;
  TOKEN_TYPE_PLAYER = 1;
  TOKEN_TYPE_MONSTER = 2;
  TOKEN_TYPE_NPC = 3;
  TOKEN_TYPE_OBJECT = 4;
  TOKEN_TYPE_EFFECT = 5;
}

enum TokenSize {
  TOKEN_SIZE_UNSPECIFIED = 0;
  TOKEN_SIZE_SMALL = 1;      // 1 hex
  TOKEN_SIZE_MEDIUM = 2;     // 1 hex
  TOKEN_SIZE_LARGE = 3;      // 2 hexes
  TOKEN_SIZE_HUGE = 4;       // 3 hexes
  TOKEN_SIZE_GARGANTUAN = 5; // 4+ hexes
}
```

## React Implementation

### Core Components

#### HexGrid Component
```typescript
interface HexGridProps {
  grid: HexGrid;
  tokens: Token[];
  selectedTile?: HexCoordinate;
  highlightedTiles?: HexCoordinate[];
  onTileClick: (coordinate: HexCoordinate) => void;
  onTokenMove: (tokenId: string, newPosition: HexCoordinate) => void;
  onTokenSelect: (tokenId: string) => void;
}

export const HexGrid: React.FC<HexGridProps> = ({
  grid,
  tokens,
  selectedTile,
  highlightedTiles = [],
  onTileClick,
  onTokenMove,
  onTokenSelect
}) => {
  // Calculate viewport dimensions
  const viewBox = useMemo(() => {
    const bounds = calculateGridBounds(grid);
    return `${bounds.minX} ${bounds.minY} ${bounds.width} ${bounds.height}`;
  }, [grid]);

  return (
    <svg viewBox={viewBox} className="hex-grid">
      {/* Background tiles */}
      {grid.tiles.map(tile => (
        <HexTile
          key={`${tile.coordinate.q},${tile.coordinate.r}`}
          tile={tile}
          hexSize={grid.hex_size}
          selected={selectedTile && coordinatesEqual(tile.coordinate, selectedTile)}
          highlighted={highlightedTiles.some(coord => coordinatesEqual(coord, tile.coordinate))}
          onClick={() => onTileClick(tile.coordinate)}
        />
      ))}
      
      {/* Token layer */}
      {tokens.map(token => (
        <HexToken
          key={token.id}
          token={token}
          hexSize={grid.hex_size}
          orientation={grid.orientation}
          onMove={(newPosition) => onTokenMove(token.id, newPosition)}
          onSelect={() => onTokenSelect(token.id)}
        />
      ))}
    </svg>
  );
};
```

#### HexTile Component
```typescript
interface HexTileProps {
  tile: HexTile;
  hexSize: number;
  selected?: boolean;
  highlighted?: boolean;
  onClick: () => void;
}

export const HexTile: React.FC<HexTileProps> = ({
  tile,
  hexSize,
  selected,
  highlighted,
  onClick
}) => {
  const position = HexUtils.axialToPixel(tile.coordinate, hexSize);
  const hexPath = HexUtils.generateHexPath(hexSize);
  
  const className = clsx('hex-tile', {
    'hex-tile--selected': selected,
    'hex-tile--highlighted': highlighted,
    'hex-tile--walkable': tile.walkable,
    'hex-tile--blocks-sight': tile.blocks_sight,
    [`hex-tile--${tile.type.toLowerCase()}`]: true,
  });

  return (
    <g 
      className={className}
      transform={`translate(${position.x}, ${position.y})`}
      onClick={onClick}
    >
      <path
        d={hexPath}
        fill={getTileColor(tile.type)}
        stroke={getTileStroke(tile.type)}
        strokeWidth="1"
      />
      
      {tile.texture && (
        <image
          href={tile.texture}
          x={-hexSize}
          y={-hexSize}
          width={hexSize * 2}
          height={hexSize * 2}
          opacity="0.8"
        />
      )}
      
      {tile.modifiers?.map(modifier => (
        <TileModifier
          key={modifier}
          modifier={modifier}
          hexSize={hexSize}
        />
      ))}
    </g>
  );
};
```

### Hex Utility Functions

#### Core Hex Math
```typescript
export class HexUtils {
  static readonly SQRT3 = Math.sqrt(3);
  
  static distance(a: HexCoordinate, b: HexCoordinate): number {
    return (Math.abs(a.q - b.q) + Math.abs(a.q + a.r - b.q - b.r) + Math.abs(a.r - b.r)) / 2;
  }

  static neighbors(coord: HexCoordinate): HexCoordinate[] {
    const directions = [
      { q: 1, r: 0, s: -1 },   // East
      { q: 1, r: -1, s: 0 },   // Northeast
      { q: 0, r: -1, s: 1 },   // Northwest
      { q: -1, r: 0, s: 1 },   // West
      { q: -1, r: 1, s: 0 },   // Southwest
      { q: 0, r: 1, s: -1 }    // Southeast
    ];
    
    return directions.map(dir => ({
      q: coord.q + dir.q,
      r: coord.r + dir.r,
      s: coord.s + dir.s
    }));
  }

  static pathfind(start: HexCoordinate, end: HexCoordinate, grid: HexGrid): HexCoordinate[] {
    // A* pathfinding implementation for hex grids
    // Returns array of coordinates from start to end
    return aStarHex(start, end, grid);
  }

  static generateHexPath(size: number): string {
    const points = [];
    for (let i = 0; i < 6; i++) {
      const angle = (Math.PI / 3) * i;
      const x = size * Math.cos(angle);
      const y = size * Math.sin(angle);
      points.push(`${x},${y}`);
    }
    return `M${points.join('L')}Z`;
  }
}
```

#### Range and Area Calculations
```typescript
export function getHexesInRange(center: HexCoordinate, range: number): HexCoordinate[] {
  const results: HexCoordinate[] = [];
  
  for (let q = -range; q <= range; q++) {
    const r1 = Math.max(-range, -q - range);
    const r2 = Math.min(range, -q + range);
    
    for (let r = r1; r <= r2; r++) {
      const s = -q - r;
      results.push({
        q: center.q + q,
        r: center.r + r,
        s: center.s + s
      });
    }
  }
  
  return results;
}

export function getHexesInLine(start: HexCoordinate, end: HexCoordinate): HexCoordinate[] {
  const distance = HexUtils.distance(start, end);
  const results: HexCoordinate[] = [];
  
  for (let i = 0; i <= distance; i++) {
    const t = i / distance;
    const lerp = hexLerp(start, end, t);
    results.push(hexRound(lerp.q, lerp.r));
  }
  
  return results;
}
```

## Movement and Range

### Movement Rules
1. **Basic Movement**: 1 hex = 5 feet (standard D&D)
2. **Difficult Terrain**: 2 movement per hex
3. **Blocked Hexes**: Cannot move through walls/obstacles
4. **Token Size**: Large creatures occupy multiple hexes

### Range Calculations
```typescript
function getAttackRange(token: Token, weapon: Weapon): HexCoordinate[] {
  const range = weapon.range || 1; // Default melee range
  const center = token.position;
  
  if (weapon.type === 'melee') {
    return getHexesInRange(center, range);
  } else {
    // Ranged weapons need line of sight
    return getHexesInRangeWithLineOfSight(center, range, grid);
  }
}
```

### Area of Effect
```typescript
function getSpellAOE(center: HexCoordinate, aoeType: string, size: number): HexCoordinate[] {
  switch (aoeType) {
    case 'circle':
      return getHexesInRange(center, size);
    case 'cone':
      return getHexesInCone(center, size, direction);
    case 'line':
      return getHexesInLine(center, endpoint);
    default:
      return [center];
  }
}
```

## Performance Considerations

### Rendering Optimization
- **Viewport Culling**: Only render visible hexes
- **Object Pooling**: Reuse hex tile components
- **Efficient Updates**: Use React.memo and useMemo
- **SVG Optimization**: Group similar elements

### Memory Management
- **Lazy Loading**: Load room data on demand
- **Caching**: Cache calculated hex paths and positions
- **Cleanup**: Remove unused token and tile data

### Network Optimization
- **Delta Updates**: Only send position changes
- **Compression**: Compress hex coordinate arrays
- **Batching**: Group multiple updates together

## Integration with Combat System

### Initiative and Turns
```typescript
interface CombatState {
  currentTurn: string;
  initiative: InitiativeEntry[];
  tokens: Token[];
  grid: HexGrid;
  movementHighlights: HexCoordinate[];
  attackRangeHighlights: HexCoordinate[];
}

function handleTurnStart(participantId: string, combatState: CombatState) {
  const token = combatState.tokens.find(t => t.character_id === participantId);
  if (!token) return;

  // Highlight possible movement
  const movementRange = getMovementRange(token, combatState.grid);
  
  // Highlight attack range
  const attackRange = getAttackRange(token, token.equipped_weapon);
  
  return {
    ...combatState,
    movementHighlights: movementRange,
    attackRangeHighlights: attackRange
  };
}
```

### Action Integration
```typescript
function handleAttackAction(
  attackerId: string,
  targetId: string,
  combatState: CombatState
): CombatActionResult {
  const attacker = combatState.tokens.find(t => t.id === attackerId);
  const target = combatState.tokens.find(t => t.id === targetId);
  
  if (!attacker || !target) {
    return { success: false, error: 'Invalid participants' };
  }

  // Check range
  const distance = HexUtils.distance(attacker.position, target.position);
  const weapon = getEquippedWeapon(attacker);
  
  if (distance > weapon.range) {
    return { success: false, error: 'Target out of range' };
  }

  // Check line of sight
  if (weapon.type === 'ranged' && !hasLineOfSight(attacker.position, target.position, combatState.grid)) {
    return { success: false, error: 'No line of sight' };
  }

  // Execute attack via existing combat service
  return executeCombatAction(attackerId, targetId, 'attack');
}
```

## Room Editor Features

### DM Tools
- **Tile Painting**: Click and drag to paint tile types
- **Token Placement**: Drag tokens from palette to board
- **Room Templates**: Pre-made room layouts
- **Import/Export**: Save and share room designs

### Template System
```typescript
interface RoomTemplate {
  id: string;
  name: string;
  description: string;
  size: { width: number; height: number };
  tiles: HexTile[];
  recommended_tokens: TokenPlacement[];
  tags: string[];
}
```

## Future Enhancements

### Advanced Features
- **Fog of War**: Hide unexplored areas
- **Dynamic Lighting**: Calculate light sources and shadows
- **Animated Tokens**: Token movement animations
- **Sound Integration**: Positional audio based on hex distance

### Mobile Support
- **Touch Controls**: Pinch to zoom, tap to select
- **Responsive Design**: Adapt to different screen sizes
- **Offline Mode**: Cache room data for offline play

### Accessibility
- **Screen Reader Support**: Describe board state
- **High Contrast Mode**: Better visibility options
- **Keyboard Navigation**: Navigate board with keys
- **Voice Commands**: "Move token to hex 3,4"

## Testing Strategy

### Unit Tests
- Hex coordinate math functions
- Range and area calculations
- Token movement validation
- Grid boundary checking

### Integration Tests
- gRPC service integration
- Real-time updates
- Combat action integration
- Room persistence

### Performance Tests
- Large grid rendering
- Many tokens on screen
- Rapid position updates
- Memory usage monitoring

## Conclusion

The hex-based board system provides a robust foundation for tactical D&D combat while maintaining the flexibility needed for various encounter types. The React implementation offers rich interactivity while the gRPC backend ensures real-time synchronization across all clients.

The system is designed to be extensible, with clear separation between the coordinate system, rendering layer, and game logic. This allows for future enhancements while maintaining compatibility with the existing Discord bot infrastructure.