# Design Decisions

## Why TypeScript + gRPC?

### TypeScript Benefits
- **Type Safety**: Catch errors at compile time, especially important for complex D&D rules
- **Better IDE Support**: Autocomplete for Discord.js and game entities
- **Proto Integration**: Generate TypeScript types from .proto files
- **Ecosystem**: Vast NPM ecosystem for D&D tools and Discord bots

### gRPC Benefits
- **Service Separation**: Clean boundaries between Discord bot, game logic, and web interface
- **Type Safety**: Protocol Buffers ensure consistent data structures across services
- **Performance**: Binary protocol more efficient than REST/JSON
- **Streaming**: Perfect for real-time combat updates to web interface
- **Language Agnostic**: Could add Python AI service or Go performance-critical service later

## Repository Pattern Rationale

Using the repository pattern provides several advantages:

1. **Testability**: Mock repositories for unit tests
2. **Flexibility**: Switch between PostgreSQL/MongoDB/in-memory without changing business logic
3. **Caching**: Add Redis caching layer transparently
4. **Consistency**: Same interface pattern across all entities

Example implementation:
```typescript
class CharacterRepositoryImpl implements CharacterRepository {
  constructor(
    private prisma: PrismaClient,
    private redis: Redis
  ) {}

  async findById(id: string): Promise<Character | null> {
    // Check cache first
    const cached = await this.redis.get(`character:${id}`)
    if (cached) return JSON.parse(cached)
    
    // Fetch from database
    const character = await this.prisma.character.findUnique({
      where: { id },
      include: { race: true, class: true, equipment: true }
    })
    
    // Cache for next time
    if (character) {
      await this.redis.setex(`character:${id}`, 300, JSON.stringify(character))
    }
    
    return character
  }
}
```

## State Management Strategy

### Discord Bot State
- **Stateless**: Each interaction is independent
- **Context from Discord**: User/channel/guild IDs provide context
- **Ephemeral Storage**: Redis for active sessions

### Combat State (Redis)
```typescript
interface CombatState {
  sessionId: string
  channelId: string
  participants: Map<string, CombatParticipant>
  initiativeOrder: string[]
  currentTurnIndex: number
  round: number
  history: CombatAction[]
}
```

### Character State (PostgreSQL)
- Permanent character data
- Equipment and inventory
- Experience and progression
- Character history/logs

## Message Architecture Details

### Shared Combat Message
```typescript
const combatEmbed = new EmbedBuilder()
  .setTitle("⚔️ Combat Encounter")
  .addFields([
    { name: "Round", value: combat.round.toString(), inline: true },
    { name: "Turn", value: currentParticipant.name, inline: true },
  ])
  .setDescription(generateCombatMap(combat))
  
const actionRow = new ActionRowBuilder()
  .addComponents(
    new ButtonBuilder()
      .setCustomId('end_turn')
      .setLabel('End Turn')
      .setStyle(ButtonStyle.Primary)
      .setDisabled(!isCurrentTurn)
  )
```

### Ephemeral Character Sheet
```typescript
const characterSheet = new EmbedBuilder()
  .setTitle(`${character.name} - Level ${character.level} ${character.class.name}`)
  .addFields([
    { name: "HP", value: `${character.hp.current}/${character.hp.max}`, inline: true },
    { name: "AC", value: character.ac.toString(), inline: true },
    { name: "Initiative", value: `+${character.initiativeBonus}`, inline: true },
  ])

const actionRows = [
  createAttackActionRow(character),
  createMovementActionRow(character),
  createSpellActionRow(character),
  createItemActionRow(character),
]
```

## Web Interface Architecture

### Real-time Updates via WebSocket
```typescript
// Server-side
combat.on('stateChange', (newState) => {
  io.to(`combat:${combat.sessionId}`).emit('combatUpdate', newState)
})

// Client-side React
const useCombatState = (sessionId: string) => {
  const [combat, setCombat] = useState<CombatState>()
  
  useEffect(() => {
    socket.on('combatUpdate', setCombat)
    socket.emit('joinCombat', sessionId)
    
    return () => {
      socket.emit('leaveCombat', sessionId)
      socket.off('combatUpdate')
    }
  }, [sessionId])
  
  return combat
}
```

### Battle Map Rendering
- **Canvas/WebGL**: For performance with many tokens
- **Grid System**: Standard D&D 5-foot squares
- **Layers**: Background, grid, tokens, effects, fog of war

## Database Schema Design

### PostgreSQL Schema (via Prisma)
```prisma
model Character {
  id        String   @id @default(cuid())
  userId    String
  name      String
  level     Int      @default(1)
  experience Int     @default(0)
  
  raceId    String
  race      Race     @relation(fields: [raceId], references: [id])
  
  classId   String
  class     Class    @relation(fields: [classId], references: [id])
  
  attributes Json    // STR, DEX, CON, INT, WIS, CHA
  
  maxHp     Int
  currentHp Int
  tempHp    Int      @default(0)
  
  equipment Equipment[]
  inventory Item[]
  spells    CharacterSpell[]
  
  createdAt DateTime @default(now())
  updatedAt DateTime @updatedAt
  
  @@index([userId])
}
```

### Redis Schema
```typescript
// Combat session
combat:${sessionId} = {
  channelId: string
  participants: Participant[]
  state: 'initiative' | 'active' | 'complete'
  currentTurn: number
  round: number
}

// Character cache
character:${characterId} = Character // 5 min TTL

// Active character sheets (ephemeral message IDs)
sheet:${userId}:${channelId} = messageId
```

## Error Handling Strategy

### Discord Interactions
```typescript
try {
  await interaction.deferReply({ ephemeral: true })
  const result = await gameService.executeAction(request)
  await interaction.editReply({ embeds: [formatResult(result)] })
} catch (error) {
  if (error instanceof GameError) {
    await interaction.editReply({ 
      content: `❌ ${error.message}`,
      ephemeral: true 
    })
  } else {
    logger.error('Unexpected error', error)
    await interaction.editReply({ 
      content: '❌ An unexpected error occurred',
      ephemeral: true 
    })
  }
}
```

### gRPC Error Codes
- `INVALID_ARGUMENT`: Bad input (wrong spell slot, invalid target)
- `NOT_FOUND`: Character/session doesn't exist
- `PERMISSION_DENIED`: Not your turn, not your character
- `FAILED_PRECONDITION`: Can't act while unconscious
- `RESOURCE_EXHAUSTED`: No spell slots remaining

## Testing Strategy

### Unit Tests
- Repository implementations with in-memory database
- Combat engine rules
- Dice rolling statistics

### Integration Tests
- gRPC service endpoints
- Discord command handlers
- WebSocket events

### E2E Tests
- Full combat encounter flow
- Character creation process
- Multi-user combat scenarios