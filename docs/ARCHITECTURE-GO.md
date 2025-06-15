# D&D Discord Bot Architecture (Go + NoSQL)

## Overview

A Discord bot for D&D 5e gameplay built with Go, using MongoDB for document storage and Redis for session management. Follows clean architecture patterns similar to the Ronnied bot.

## Tech Stack

### Core Technologies
- **Language**: Go 1.21+
- **Discord Integration**: discordgo
- **IPC/Services**: gRPC with Protocol Buffers
- **Database**: MongoDB (document store) + Redis (session/combat state)
- **Web Interface**: React + WebSockets (or htmx for simplicity)
- **Container**: Docker & Docker Compose

### Key Libraries
- **github.com/bwmarrin/discordgo**: Discord bot framework
- **google.golang.org/grpc**: gRPC server/client
- **go.mongodb.org/mongo-driver**: MongoDB driver
- **github.com/redis/go-redis/v9**: Redis client
- **github.com/gorilla/websocket**: WebSocket support
- **github.com/google/uuid**: UUID generation
- **github.com/spf13/viper**: Configuration management

## Architecture Patterns

### 1. Clean Architecture (Similar to Ronnied)

```
┌─────────────────────────────────────────────────┐
│                  Handlers Layer                  │
│          (Discord Commands & Interactions)       │
└─────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────┐
│                 Services Layer                   │
│         (Business Logic & Game Rules)            │
└─────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────┐
│               Repositories Layer                 │
│          (MongoDB & Redis Persistence)           │
└─────────────────────────────────────────────────┘
```

### 2. Service Communication

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│  Discord Bot    │────▶│  Game Service   │────▶│  Web Server     │
│   (Gateway)     │gRPC │   (Core Logic)  │gRPC │   (htmx/React)  │
└─────────────────┘     └─────────────────┘     └─────────────────┘
        │                        │                        │
        └────────────────────────┴────────────────────────┘
                                 │
                    ┌────────────┴────────────┐
                    │                         │
              ┌─────▼─────┐           ┌──────▼──────┐
              │  MongoDB  │           │    Redis    │
              │(Documents)│           │ (Sessions)  │
              └───────────┘           └─────────────┘
```

## Project Structure

```
/dnd-bot-discord
├── /cmd
│   ├── /bot                   # Discord bot entry point
│   │   └── main.go
│   ├── /game                  # Game service entry point
│   │   └── main.go
│   └── /web                   # Web server entry point
│       └── main.go
├── /internal
│   ├── /handlers
│   │   ├── /discord          # Discord interaction handlers
│   │   │   ├── commands.go
│   │   │   ├── interactions.go
│   │   │   └── messages.go
│   │   └── /grpc             # gRPC handlers
│   │       ├── character.go
│   │       └── combat.go
│   ├── /services
│   │   ├── /character        # Character management
│   │   ├── /combat           # Combat engine
│   │   ├── /dice             # Dice rolling (from Ronnied)
│   │   └── /rules            # D&D 5e rules engine
│   ├── /repositories
│   │   ├── /mongodb          # MongoDB implementations
│   │   │   ├── character.go
│   │   │   └── campaign.go
│   │   └── /redis            # Redis implementations
│   │       ├── combat.go
│   │       └── session.go
│   ├── /models               # Domain models
│   │   ├── character.go
│   │   ├── combat.go
│   │   ├── equipment.go
│   │   └── spells.go
│   ├── /common               # Shared utilities
│   │   ├── errors.go
│   │   └── logger.go
│   └── /config               # Configuration
│       └── config.go
├── /proto                     # Protocol Buffer definitions
├── /web                       # Web interface (if React)
├── /migrations                # MongoDB migrations
├── /scripts                   # Build/deploy scripts
├── docker-compose.yml
├── go.mod
└── go.sum
```

## MongoDB Schema Design

### Character Collection
```go
type Character struct {
    ID               primitive.ObjectID `bson:"_id,omitempty"`
    UserID           string            `bson:"user_id"`
    Name             string            `bson:"name"`
    Level            int               `bson:"level"`
    Experience       int               `bson:"experience"`
    
    // Embedded documents for denormalization
    Race             Race              `bson:"race"`
    Class            Class             `bson:"class"`
    Background       Background        `bson:"background"`
    
    // Attributes as subdocument
    Attributes       Attributes        `bson:"attributes"`
    
    // Combat stats
    HP               HitPoints         `bson:"hp"`
    AC               int               `bson:"ac"`
    Initiative       int               `bson:"initiative"`
    Speed            int               `bson:"speed"`
    
    // Arrays of subdocuments
    Equipment        []Equipment       `bson:"equipment"`
    Inventory        []Item            `bson:"inventory"`
    Proficiencies    []Proficiency     `bson:"proficiencies"`
    Spells           []Spell           `bson:"spells"`
    
    Status          CharacterStatus   `bson:"status"`
    CreatedAt       time.Time         `bson:"created_at"`
    UpdatedAt       time.Time         `bson:"updated_at"`
}

type Attributes struct {
    Strength     int `bson:"str"`
    Dexterity    int `bson:"dex"`
    Constitution int `bson:"con"`
    Intelligence int `bson:"int"`
    Wisdom       int `bson:"wis"`
    Charisma     int `bson:"cha"`
}

type HitPoints struct {
    Current int `bson:"current"`
    Max     int `bson:"max"`
    Temp    int `bson:"temp"`
}
```

### Combat Session (Redis)
```go
type CombatSession struct {
    ID              string                    `json:"id"`
    ChannelID       string                    `json:"channel_id"`
    State           CombatState               `json:"state"`
    Participants    map[string]*Participant   `json:"participants"`
    InitiativeOrder []string                  `json:"initiative_order"`
    CurrentTurn     int                       `json:"current_turn"`
    Round           int                       `json:"round"`
    Map             *BattleMap                `json:"map"`
    History         []CombatAction            `json:"history"`
}
```

## Repository Pattern Implementation

```go
// Repository interfaces
type CharacterRepository interface {
    Create(ctx context.Context, character *models.Character) error
    FindByID(ctx context.Context, id string) (*models.Character, error)
    FindByUserID(ctx context.Context, userID string) ([]*models.Character, error)
    Update(ctx context.Context, character *models.Character) error
    Delete(ctx context.Context, id string) error
}

type CombatRepository interface {
    CreateSession(ctx context.Context, session *models.CombatSession) error
    GetSession(ctx context.Context, sessionID string) (*models.CombatSession, error)
    UpdateSession(ctx context.Context, session *models.CombatSession) error
    DeleteSession(ctx context.Context, sessionID string) error
    GetActiveSessionByChannel(ctx context.Context, channelID string) (*models.CombatSession, error)
}

// MongoDB implementation
type mongoCharacterRepo struct {
    db         *mongo.Database
    collection *mongo.Collection
}

func (r *mongoCharacterRepo) Create(ctx context.Context, character *models.Character) error {
    character.CreatedAt = time.Now()
    character.UpdatedAt = time.Now()
    
    result, err := r.collection.InsertOne(ctx, character)
    if err != nil {
        return fmt.Errorf("failed to create character: %w", err)
    }
    
    character.ID = result.InsertedID.(primitive.ObjectID)
    return nil
}

// Redis implementation  
type redisCombatRepo struct {
    client *redis.Client
    ttl    time.Duration
}

func (r *redisCombatRepo) CreateSession(ctx context.Context, session *models.CombatSession) error {
    data, err := json.Marshal(session)
    if err != nil {
        return err
    }
    
    key := fmt.Sprintf("combat:%s", session.ID)
    return r.client.Set(ctx, key, data, r.ttl).Err()
}
```

## Service Layer Example

```go
type CombatService struct {
    combatRepo    repositories.CombatRepository
    characterRepo repositories.CharacterRepository
    diceService   *dice.Service
    rulesEngine   *rules.Engine
}

func (s *CombatService) ExecuteAttack(ctx context.Context, sessionID, attackerID, targetID string) (*AttackResult, error) {
    session, err := s.combatRepo.GetSession(ctx, sessionID)
    if err != nil {
        return nil, err
    }
    
    attacker := session.Participants[attackerID]
    target := session.Participants[targetID]
    
    // Validate turn
    if session.InitiativeOrder[session.CurrentTurn] != attackerID {
        return nil, ErrNotYourTurn
    }
    
    // Roll attack
    attackRoll := s.diceService.Roll("1d20")
    attackTotal := attackRoll.Total + attacker.AttackBonus
    
    // Check hit
    if attackTotal >= target.AC {
        damageRoll := s.diceService.Roll(attacker.DamageDice)
        target.CurrentHP -= damageRoll.Total
        
        // Update session
        session.History = append(session.History, CombatAction{
            Type:      ActionAttack,
            ActorID:   attackerID,
            TargetID:  targetID,
            Result:    fmt.Sprintf("Hit for %d damage", damageRoll.Total),
            Timestamp: time.Now(),
        })
        
        return &AttackResult{
            Hit:        true,
            Damage:     damageRoll.Total,
            AttackRoll: attackRoll,
            DamageRoll: damageRoll,
        }, s.combatRepo.UpdateSession(ctx, session)
    }
    
    return &AttackResult{Hit: false}, nil
}
```

## Discord Handler Pattern

```go
type DiscordHandler struct {
    session         *discordgo.Session
    characterSvc    *services.CharacterService
    combatSvc       *services.CombatService
    messageBuilder  *MessageBuilder
}

func (h *DiscordHandler) HandleCombatStart(s *discordgo.Session, i *discordgo.InteractionCreate) {
    // Defer response for processing
    s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
        Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
    })
    
    // Create combat session
    session, err := h.combatSvc.InitiateCombat(context.Background(), i.ChannelID, i.Member.User.ID)
    if err != nil {
        h.respondError(s, i, err)
        return
    }
    
    // Build shared combat message
    embed := h.messageBuilder.BuildCombatEmbed(session)
    components := h.messageBuilder.BuildCombatControls(session)
    
    // Send response
    s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
        Embeds:     &[]*discordgo.MessageEmbed{embed},
        Components: &components,
    })
}
```

## WebSocket Integration

```go
type WebSocketHub struct {
    clients    map[string]*Client
    combat     chan CombatUpdate
    register   chan *Client
    unregister chan *Client
    mu         sync.RWMutex
}

func (h *WebSocketHub) RunCombatUpdates(combatService *services.CombatService) {
    combatService.OnStateChange(func(update CombatUpdate) {
        h.mu.RLock()
        defer h.mu.RUnlock()
        
        for _, client := range h.clients {
            if client.SessionID == update.SessionID {
                select {
                case client.send <- update:
                default:
                    close(client.send)
                    delete(h.clients, client.ID)
                }
            }
        }
    })
}
```

## Configuration

```yaml
# config.yaml
bot:
  token: ${DISCORD_TOKEN}
  application_id: ${DISCORD_APP_ID}

mongodb:
  uri: mongodb://localhost:27017
  database: dnd_bot

redis:
  addr: localhost:6379
  password: ""
  db: 0
  
grpc:
  game_service: localhost:50051
  web_service: localhost:50052

combat:
  session_ttl: 4h
  max_participants: 8
  
logging:
  level: info
  format: json
```

## Benefits of Go + NoSQL Approach

1. **Performance**: Go's concurrency model perfect for real-time combat
2. **Type Safety**: Still get compile-time checks with Go's type system
3. **MongoDB Flexibility**: Easy to evolve schema as features grow
4. **Document Model**: Natural fit for D&D entities (characters, spells, items)
5. **Proven Architecture**: Leverage patterns from successful Ronnied bot
6. **Simple Deployment**: Single binary per service
7. **Great Libraries**: discordgo is mature and well-maintained