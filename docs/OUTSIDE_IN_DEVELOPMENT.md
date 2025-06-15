# Outside-In Development Approach

## Philosophy

Start with Discord handlers (what users actually interact with) and work inward. Each layer defines the interface it needs from the layer below, leading to minimal, focused interfaces.

## Layer Architecture

```
┌─────────────────────────────────────────────┐
│         Discord Handlers (Outer)            │ ← Start here
├─────────────────────────────────────────────┤
│            Use Cases / Services             │ ← Define based on handler needs
├─────────────────────────────────────────────┤
│              Domain Models                  │ ← Emerge from use cases
├─────────────────────────────────────────────┤
│           Repository Interfaces             │ ← Define based on service needs
├─────────────────────────────────────────────┤
│        Repository Implementations           │ ← Implement last (Redis/Mongo/Couch)
└─────────────────────────────────────────────┘
```

## Development Flow Example: Character View Command

### Step 1: Start with Discord Handler (Outermost)

```go
// internal/handlers/discord/character_handler_test.go
func TestCharacterViewCommand(t *testing.T) {
    // Define what we WANT to happen
    mockCharService := &MockCharacterService{}
    handler := NewCharacterHandler(mockCharService)
    
    // User wants to view their character
    interaction := createTestInteraction("/character view", "user123")
    
    expectedCharacter := &Character{
        Name:  "Thorin",
        Level: 5,
        HP:    CurrentMax{Current: 44, Max: 44},
    }
    
    mockCharService.On("GetActiveCharacter", "user123").Return(expectedCharacter, nil)
    
    // Execute
    response := handler.HandleCharacterView(interaction)
    
    // Verify Discord response
    assert.Equal(t, "ephemeral", response.Type)
    assert.Contains(t, response.Embed.Title, "Thorin - Level 5")
    assert.Contains(t, response.Embed.Fields, "HP: 44/44")
}
```

### Step 2: Implement Handler, Discover Service Interface

```go
// internal/handlers/discord/character_handler.go
type CharacterHandler struct {
    charService CharacterService // Interface emerges from our needs
}

// This interface is EXACTLY what we need, nothing more
type CharacterService interface {
    GetActiveCharacter(userID string) (*Character, error)
    GetCharacterByName(userID, name string) (*Character, error)
}

func (h *CharacterHandler) HandleCharacterView(i *Interaction) Response {
    userID := i.UserID
    
    character, err := h.charService.GetActiveCharacter(userID)
    if err != nil {
        return ErrorResponse("No active character found")
    }
    
    return EphemeralResponse(
        BuildCharacterEmbed(character),
        BuildCharacterActions(character),
    )
}
```

### Step 3: Implement Service, Discover Repository Interface

```go
// internal/services/character_service_test.go
func TestGetActiveCharacter(t *testing.T) {
    mockRepo := &MockCharacterRepository{}
    mockCache := &MockCache{}
    service := NewCharacterService(mockRepo, mockCache)
    
    // Service needs to get user's characters and find active one
    characters := []*Character{
        {ID: "1", Status: "archived"},
        {ID: "2", Status: "active", Name: "Thorin"},
    }
    
    mockCache.On("Get", "active_char:user123").Return(nil, redis.Nil)
    mockRepo.On("FindByUser", "user123").Return(characters, nil)
    mockCache.On("Set", "active_char:user123", "2", 5*time.Minute).Return(nil)
    
    result, err := service.GetActiveCharacter("user123")
    
    assert.NoError(t, err)
    assert.Equal(t, "Thorin", result.Name)
}
```

```go
// internal/services/character_service.go
type CharacterService struct {
    repo  CharacterRepository  // Interface emerges from needs
    cache CacheRepository
}

// Repository interface is EXACTLY what service needs
type CharacterRepository interface {
    FindByUser(userID string) ([]*Character, error)
    FindByID(id string) (*Character, error)
    Save(character *Character) error
}

type CacheRepository interface {
    Get(key string) ([]byte, error)
    Set(key string, value interface{}, ttl time.Duration) error
}

func (s *CharacterService) GetActiveCharacter(userID string) (*Character, error) {
    // Check cache first
    cacheKey := fmt.Sprintf("active_char:%s", userID)
    if charID, err := s.cache.Get(cacheKey); err == nil {
        return s.repo.FindByID(string(charID))
    }
    
    // Get all user's characters
    characters, err := s.repo.FindByUser(userID)
    if err != nil {
        return nil, err
    }
    
    // Find active one
    for _, char := range characters {
        if char.Status == "active" {
            // Cache for next time
            s.cache.Set(cacheKey, char.ID, 5*time.Minute)
            return char, nil
        }
    }
    
    return nil, ErrNoActiveCharacter
}
```

### Step 4: Implement Repository (Innermost) - Defer Decision

```go
// internal/repositories/memory/character_repo.go
// Start with in-memory for rapid development
type InMemoryCharacterRepo struct {
    characters map[string]*Character
    mu         sync.RWMutex
}

func (r *InMemoryCharacterRepo) FindByUser(userID string) ([]*Character, error) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    
    var result []*Character
    for _, char := range r.characters {
        if char.UserID == userID {
            result = append(result, char)
        }
    }
    return result, nil
}
```

Later, implement real repositories:

```go
// internal/repositories/mongodb/character_repo.go
type MongoCharacterRepo struct {
    coll *mongo.Collection
}

func (r *MongoCharacterRepo) FindByUser(userID string) ([]*Character, error) {
    cursor, err := r.coll.Find(ctx, bson.M{"user_id": userID})
    // ... implementation
}

// internal/repositories/redis/character_repo.go  
type RedisCharacterRepo struct {
    client *redis.Client
}

func (r *RedisCharacterRepo) FindByUser(userID string) ([]*Character, error) {
    pattern := fmt.Sprintf("char:user:%s:*", userID)
    // ... scan and unmarshal
}
```

## Benefits of This Approach

1. **No Wasted Code**: Every interface method exists because it's actually used
2. **Easy Testing**: Mock exactly what you need at each layer
3. **Flexible Storage**: Swap databases without changing business logic
4. **Clear Dependencies**: Each layer only knows about the layer directly below
5. **Fast Development**: Use in-memory repos initially, add real DB later

## Testing Strategy

```go
// Each layer has its own test doubles
type Mocks struct {
    CharacterService *MockCharacterService // For handler tests
    CharacterRepo    *MockCharacterRepo    // For service tests
    Cache            *MockCache            // For service tests
}

// Integration tests use real implementations
func TestCharacterFlowIntegration(t *testing.T) {
    // Use in-memory repo for fast tests
    repo := memory.NewCharacterRepo()
    cache := memory.NewCache()
    service := services.NewCharacterService(repo, cache)
    handler := discord.NewCharacterHandler(service)
    
    // Test full flow
}
```

## Development Order

### Phase 1: Core Commands (All with Mocks)
1. `/character view` → Handler → Service interface
2. `/character create` → Handler → Service interface  
3. `/combat start` → Handler → Service interface
4. `/roll` → Handler → Dice service interface

### Phase 2: Services (With Mock Repos)
1. CharacterService → Repository interfaces
2. CombatService → Repository interfaces
3. DiceService (no repo needed)

### Phase 3: Repositories (Choose Implementation)
1. Try Redis for everything initially (you mentioned experience)
2. Move character data to MongoDB/CouchDB if needed
3. Keep combat in Redis (perfect for sessions)

## Repository Decision Framework

Start with Redis for EVERYTHING:
```go
// Redis can handle all our data with proper key design
// characters -> HASH    char:user123:char456 
// combat     -> STRING  combat:session:abc123 (JSON)
// indexes    -> SET     char:user:user123 -> [char456, char789]
```

Later, if needed:
- **Need complex queries?** → MongoDB or CouchDB
- **Need ACID transactions?** → PostgreSQL with JSONB
- **Happy with Redis?** → Keep it!

## Example: Combat Start Command

```go
// Step 1: Handler test defines what we need
func TestCombatStartCommand(t *testing.T) {
    mockCombatService := &MockCombatService{}
    
    expectedSession := &CombatSession{
        ID:        "session123",
        ChannelID: "channel456",
        State:     "initiating",
    }
    
    mockCombatService.On("StartCombat", "channel456", "user123").
        Return(expectedSession, nil)
    
    // ... test handler behavior
}

// Step 2: This drives the service interface
type CombatService interface {
    StartCombat(channelID, initiatorID string) (*CombatSession, error)
    JoinCombat(sessionID, userID string) error
    GetSession(sessionID string) (*CombatSession, error)
}

// Step 3: Service test drives repository interface  
type CombatRepository interface {
    CreateSession(session *CombatSession) error
    GetSession(id string) (*CombatSession, error)
    GetSessionByChannel(channelID string) (*CombatSession, error)
    UpdateSession(session *CombatSession) error
}
```

This approach ensures we build exactly what we need, when we need it!