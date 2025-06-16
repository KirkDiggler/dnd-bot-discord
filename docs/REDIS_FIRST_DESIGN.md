# Redis-First Repository Design

## Why Redis for Everything?

Since you have experience with Redis repositories and we're using outside-in development, let's start with Redis for ALL data. It's simpler than you might think and we can migrate later if needed.

## Key Design Patterns

### 1. Character Storage

```go
// Key patterns
char:{charID}                    -> HASH (character data)
char:user:{userID}              -> SET (character IDs)
char:user:{userID}:active       -> STRING (active character ID)
char:name:{userID}:{name}       -> STRING (character ID for name lookup)

// Example implementation
type RedisCharacterRepo struct {
    client *redis.Client
}

func (r *RedisCharacterRepo) Save(char *Character) error {
    ctx := context.Background()
    pipe := r.client.Pipeline()
    
    charKey := fmt.Sprintf("char:%s", char.ID)
    
    // Store character as hash
    pipe.HSet(ctx, charKey, map[string]interface{}{
        "id":         char.ID,
        "user_id":    char.UserID,
        "name":       char.Name,
        "level":      char.Level,
        "race":       char.Race.ToJSON(),      // JSON for nested objects
        "class":      char.Class.ToJSON(),
        "attributes": char.Attributes.ToJSON(),
        "hp":         char.HP.ToJSON(),
        "equipment":  char.Equipment.ToJSON(),
        "created_at": char.CreatedAt.Unix(),
        "updated_at": time.Now().Unix(),
    })
    
    // Add to user's character set
    userCharsKey := fmt.Sprintf("char:user:%s", char.UserID)
    pipe.SAdd(ctx, userCharsKey, char.ID)
    
    // Set name lookup
    nameKey := fmt.Sprintf("char:name:%s:%s", char.UserID, strings.ToLower(char.Name))
    pipe.Set(ctx, nameKey, char.ID, 0)
    
    // Update active character if this is the only one
    if count := r.client.SCard(ctx, userCharsKey).Val(); count == 0 {
        activeKey := fmt.Sprintf("char:user:%s:active", char.UserID)
        pipe.Set(ctx, activeKey, char.ID, 0)
    }
    
    _, err := pipe.Exec(ctx)
    return err
}

func (r *RedisCharacterRepo) FindByUser(userID string) ([]*Character, error) {
    ctx := context.Background()
    
    // Get all character IDs for user
    userCharsKey := fmt.Sprintf("char:user:%s", userID)
    charIDs, err := r.client.SMembers(ctx, userCharsKey).Result()
    if err != nil {
        return nil, err
    }
    
    // Get each character
    var characters []*Character
    for _, id := range charIDs {
        char, err := r.FindByID(id)
        if err == nil {
            characters = append(characters, char)
        }
    }
    
    // Sort by updated_at
    sort.Slice(characters, func(i, j int) bool {
        return characters[i].UpdatedAt.After(characters[j].UpdatedAt)
    })
    
    return characters, nil
}

func (r *RedisCharacterRepo) FindByID(id string) (*Character, error) {
    ctx := context.Background()
    charKey := fmt.Sprintf("char:%s", id)
    
    data, err := r.client.HGetAll(ctx, charKey).Result()
    if err != nil {
        return nil, err
    }
    
    if len(data) == 0 {
        return nil, ErrNotFound
    }
    
    // Reconstruct character
    char := &Character{
        ID:      data["id"],
        UserID:  data["user_id"],
        Name:    data["name"],
        Level:   parseInt(data["level"]),
    }
    
    // Unmarshal JSON fields
    char.Race.FromJSON(data["race"])
    char.Class.FromJSON(data["class"])
    char.Attributes.FromJSON(data["attributes"])
    char.HP.FromJSON(data["hp"])
    char.Equipment.FromJSON(data["equipment"])
    
    return char, nil
}
```

### 2. Combat Session Storage

```go
// Key patterns
combat:session:{sessionID}       -> STRING (JSON of entire session)
combat:channel:{channelID}       -> STRING (active session ID)
combat:participant:{userID}      -> STRING (session ID user is in)

type RedisCombatRepo struct {
    client *redis.Client
    ttl    time.Duration
}

func (r *RedisCombatRepo) CreateSession(session *CombatSession) error {
    ctx := context.Background()
    pipe := r.client.Pipeline()
    
    sessionKey := fmt.Sprintf("combat:session:%s", session.ID)
    channelKey := fmt.Sprintf("combat:channel:%s", session.ChannelID)
    
    sessionJSON, _ := json.Marshal(session)
    
    // Store session
    pipe.Set(ctx, sessionKey, sessionJSON, r.ttl)
    
    // Map channel to session
    pipe.Set(ctx, channelKey, session.ID, r.ttl)
    
    _, err := pipe.Exec(ctx)
    return err
}

func (r *RedisCombatRepo) UpdateSession(session *CombatSession) error {
    ctx := context.Background()
    sessionKey := fmt.Sprintf("combat:session:%s", session.ID)
    
    // Get current TTL to preserve it
    ttl, err := r.client.TTL(ctx, sessionKey).Result()
    if err != nil || ttl < 0 {
        ttl = r.ttl
    }
    
    sessionJSON, _ := json.Marshal(session)
    return r.client.Set(ctx, sessionKey, sessionJSON, ttl).Err()
}

func (r *RedisCombatRepo) JoinCombat(sessionID, userID string) error {
    ctx := context.Background()
    participantKey := fmt.Sprintf("combat:participant:%s", userID)
    
    // Track which combat user is in
    return r.client.Set(ctx, participantKey, sessionID, r.ttl).Err()
}
```

### 3. Game Data Storage (Races, Classes, etc.)

```go
// Preload game data into Redis on startup
// Key patterns
game:race:{raceID}              -> HASH
game:class:{classID}            -> HASH
game:races                      -> SET (all race IDs)
game:classes                    -> SET (all class IDs)

type RedisGameDataRepo struct {
    client *redis.Client
}

func (r *RedisGameDataRepo) LoadGameData(data *GameData) error {
    ctx := context.Background()
    pipe := r.client.Pipeline()
    
    // Load races
    for _, race := range data.Races {
        raceKey := fmt.Sprintf("game:race:%s", race.ID)
        pipe.HSet(ctx, raceKey, map[string]interface{}{
            "id":     race.ID,
            "name":   race.Name,
            "size":   race.Size,
            "speed":  race.Speed,
            "traits": race.Traits.ToJSON(),
            "bonuses": race.AbilityBonuses.ToJSON(),
        })
        pipe.SAdd(ctx, "game:races", race.ID)
    }
    
    // Load classes
    for _, class := range data.Classes {
        classKey := fmt.Sprintf("game:class:%s", class.ID)
        pipe.HSet(ctx, classKey, map[string]interface{}{
            "id":      class.ID,
            "name":    class.Name,
            "hit_die": class.HitDie,
            "saves":   strings.Join(class.Saves, ","),
        })
        pipe.SAdd(ctx, "game:classes", class.ID)
    }
    
    _, err := pipe.Exec(ctx)
    return err
}
```

### 4. Efficient Queries with Secondary Indexes

```go
// Create secondary indexes for common queries
func (r *RedisCharacterRepo) CreateIndexes(char *Character) error {
    ctx := context.Background()
    pipe := r.client.Pipeline()
    
    // Index by level for leaderboards
    levelKey := fmt.Sprintf("idx:char:level:%d", char.Level)
    pipe.ZAdd(ctx, levelKey, redis.Z{
        Score:  float64(char.Experience),
        Member: char.ID,
    })
    
    // Index by class
    classKey := fmt.Sprintf("idx:char:class:%s", char.Class.ID)
    pipe.SAdd(ctx, classKey, char.ID)
    
    // Index by race
    raceKey := fmt.Sprintf("idx:char:race:%s", char.Race.ID)
    pipe.SAdd(ctx, raceKey, char.ID)
    
    _, err := pipe.Exec(ctx)
    return err
}

// Query characters by class
func (r *RedisCharacterRepo) FindByClass(classID string) ([]*Character, error) {
    ctx := context.Background()
    classKey := fmt.Sprintf("idx:char:class:%s", classID)
    
    charIDs, err := r.client.SMembers(ctx, classKey).Result()
    if err != nil {
        return nil, err
    }
    
    var characters []*Character
    for _, id := range charIDs {
        if char, err := r.FindByID(id); err == nil {
            characters = append(characters, char)
        }
    }
    
    return characters, nil
}
```

## Advantages of Redis-First

1. **Single Technology**: One database to manage, backup, monitor
2. **Performance**: Everything in memory, sub-millisecond responses
3. **Simple Operations**: No complex queries, just key lookups
4. **Atomic Operations**: Redis transactions for consistency
5. **TTL Support**: Built-in expiration for sessions
6. **Pub/Sub**: Free real-time updates for combat events

## When to Consider Migration

Monitor these metrics:
- **Memory Usage**: If > 1GB, consider moving cold data
- **Query Complexity**: If you need complex aggregations
- **Relationship Queries**: If you need graph traversals
- **Full-Text Search**: If you need character/spell search

## Migration Path (If Needed)

```go
// Repository interface stays the same!
type CharacterRepository interface {
    Save(char *Character) error
    FindByID(id string) (*Character, error)
    FindByUser(userID string) ([]*Character, error)
}

// Just swap implementation
container.Register(func() CharacterRepository {
    // return &RedisCharacterRepo{client}     // Start here
    // return &MongoCharacterRepo{db}        // Or here later
    // return &CouchCharacterRepo{bucket}    // Or here
})
```

## Redis Persistence Configuration

```yaml
# redis.conf for DND bot
save 900 1      # Save after 900 sec if at least 1 key changed
save 300 10     # Save after 300 sec if at least 10 keys changed  
save 60 10000   # Save after 60 sec if at least 10000 keys changed

appendonly yes  # Enable AOF for durability
appendfsync everysec  # Sync to disk every second

maxmemory 2gb   # Set memory limit
maxmemory-policy allkeys-lru  # Evict least recently used keys
```

## Example: Complete Character Flow

```go
// 1. Handler calls service
char, err := s.charService.GetActiveCharacter(userID)

// 2. Service checks cache (already in Redis!)
activeKey := fmt.Sprintf("char:user:%s:active", userID)
charID, err := s.repo.client.Get(ctx, activeKey).Result()

// 3. Service gets character
char, err := s.repo.FindByID(charID)

// 4. Redis repo does simple key lookup
charKey := fmt.Sprintf("char:%s", charID)
data := s.client.HGetAll(ctx, charKey).Result()

// Total operations: 2 Redis commands
// Total time: ~1ms
```

Start simple, stay simple until you can't!