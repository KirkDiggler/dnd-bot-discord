# gRPC Streaming Implementation

## Why Streaming?

Using gRPC streaming simplifies the architecture significantly:

1. **No separate WebSocket server** - gRPC handles real-time updates
2. **Type-safe streaming** - Proto definitions ensure consistency
3. **Built-in reconnection** - gRPC handles connection management
4. **Unified protocol** - Same interface for Discord bot and web UI

## Streaming Patterns

### 1. Bidirectional Streaming (Combat)

Perfect for combat where both client and server send messages continuously:

```go
// Server implementation
func (s *CombatStreamingServer) CombatStream(stream pb.CombatStreamingService_CombatStreamServer) error {
    // Create a context for this stream
    ctx := stream.Context()
    participantID := uuid.New().String()
    
    // Channel for outgoing events
    events := make(chan *pb.CombatEvent, 100)
    
    // Register this stream
    s.registerStream(participantID, events)
    defer s.unregisterStream(participantID)
    
    // Error channel
    errCh := make(chan error, 2)
    
    // Goroutine to receive requests
    go func() {
        for {
            req, err := stream.Recv()
            if err != nil {
                errCh <- err
                return
            }
            
            // Process request based on type
            if err := s.processRequest(ctx, participantID, req); err != nil {
                // Send error event
                events <- &pb.CombatEvent{
                    Event: &pb.CombatEvent_Error{
                        Error: &pb.ErrorOccurred{
                            Code:          "PROCESSING_ERROR",
                            Message:       err.Error(),
                            ParticipantId: participantID,
                        },
                    },
                }
            }
        }
    }()
    
    // Goroutine to send events
    go func() {
        for {
            select {
            case event := <-events:
                if err := stream.Send(event); err != nil {
                    errCh <- err
                    return
                }
            case <-ctx.Done():
                return
            }
        }
    }()
    
    // Wait for error or context cancellation
    select {
    case err := <-errCh:
        return err
    case <-ctx.Done():
        return ctx.Err()
    }
}

func (s *CombatStreamingServer) processRequest(ctx context.Context, participantID string, req *pb.CombatRequest) error {
    switch r := req.Request.(type) {
    case *pb.CombatRequest_Initiate:
        return s.handleInitiate(ctx, participantID, r.Initiate)
    case *pb.CombatRequest_Join:
        return s.handleJoin(ctx, participantID, r.Join)
    case *pb.CombatRequest_ExecuteAction:
        return s.handleAction(ctx, participantID, r.ExecuteAction)
    // ... other cases
    }
    return fmt.Errorf("unknown request type")
}
```

### 2. Server Streaming (Observers)

For web UI or spectators who only receive updates:

```go
func (s *CombatStreamingServer) ObserveCombat(req *pb.ObserveCombatRequest, stream pb.CombatStreamingService_ObserveCombatServer) error {
    ctx := stream.Context()
    
    // Get combat session
    session, err := s.combatService.GetSession(ctx, req.SessionId)
    if err != nil {
        return status.Errorf(codes.NotFound, "combat session not found")
    }
    
    // Send initial state
    if req.IncludeHistory {
        // Send full combat state
        if err := stream.Send(&pb.CombatEvent{
            SessionId: session.ID,
            Event: &pb.CombatEvent_StateChange{
                StateChange: s.buildFullState(session),
            },
        }); err != nil {
            return err
        }
    }
    
    // Subscribe to updates
    events := make(chan *pb.CombatEvent, 100)
    s.subscribeObserver(req.SessionId, req.ObserverId, events)
    defer s.unsubscribeObserver(req.SessionId, req.ObserverId)
    
    // Stream events
    for {
        select {
        case event := <-events:
            if err := stream.Send(event); err != nil {
                return err
            }
        case <-ctx.Done():
            return ctx.Err()
        }
    }
}
```

### 3. Client Streaming (Batch Actions)

For complex multi-step actions:

```go
func (s *CombatStreamingServer) ExecuteBatchActions(stream pb.CombatStreamingService_ExecuteBatchActionsServer) error {
    var actions []*pb.CombatAction
    var sessionID string
    
    // Collect all actions
    for {
        action, err := stream.Recv()
        if err == io.EOF {
            break
        }
        if err != nil {
            return err
        }
        
        actions = append(actions, action)
        if sessionID == "" && action.SessionId != "" {
            sessionID = action.SessionId
        }
    }
    
    // Execute all actions in a transaction
    results, finalState, err := s.combatService.ExecuteBatch(stream.Context(), sessionID, actions)
    if err != nil {
        return err
    }
    
    // Send results
    return stream.SendAndClose(&pb.BatchActionResult{
        Results:    results,
        FinalState: finalState,
    })
}
```

## Discord Bot Integration

```go
// Discord handler maintains a gRPC stream per active combat
type DiscordCombatHandler struct {
    client       pb.CombatStreamingServiceClient
    activeStreams map[string]pb.CombatStreamingService_CombatStreamClient
    mu           sync.RWMutex
}

func (h *DiscordCombatHandler) HandleCombatStart(s *discordgo.Session, i *discordgo.InteractionCreate) {
    // Create stream for this combat
    stream, err := h.client.CombatStream(context.Background())
    if err != nil {
        h.respondError(s, i, "Failed to start combat")
        return
    }
    
    // Send initiate request
    err = stream.Send(&pb.CombatRequest{
        Request: &pb.CombatRequest_Initiate{
            Initiate: &pb.InitiateCombatReq{
                ChannelId:   i.ChannelID,
                InitiatorId: i.Member.User.ID,
            },
        },
    })
    
    if err != nil {
        h.respondError(s, i, "Failed to initiate combat")
        return
    }
    
    // Start goroutine to handle events
    go h.handleCombatEvents(stream, i.ChannelID)
    
    // Store stream
    h.mu.Lock()
    h.activeStreams[i.ChannelID] = stream
    h.mu.Unlock()
}

func (h *DiscordCombatHandler) handleCombatEvents(stream pb.CombatStreamingService_CombatStreamClient, channelID string) {
    for {
        event, err := stream.Recv()
        if err != nil {
            log.Printf("Stream error for channel %s: %v", channelID, err)
            h.cleanupStream(channelID)
            return
        }
        
        // Process event
        switch e := event.Event.(type) {
        case *pb.CombatEvent_Initiated:
            h.createCombatMessage(channelID, e.Initiated)
        case *pb.CombatEvent_Action:
            h.updateCombatMessage(channelID, e.Action)
        case *pb.CombatEvent_TurnStarted:
            h.notifyTurnStart(channelID, e.TurnStarted)
        case *pb.CombatEvent_Error:
            h.handleError(channelID, e.Error)
        }
    }
}
```

## Web Interface Integration

```typescript
// React hook for combat observation
export function useCombatObserver(sessionId: string) {
    const [combatState, setCombatState] = useState<CombatState | null>(null);
    const [events, setEvents] = useState<CombatEvent[]>([]);
    
    useEffect(() => {
        if (!sessionId) return;
        
        const client = new CombatStreamingServiceClient(GRPC_ENDPOINT);
        const stream = client.observeCombat({
            sessionId,
            observerId: getUserId(),
            includeHistory: true
        });
        
        stream.on('data', (event: CombatEvent) => {
            // Update state based on event type
            if (event.hasStateChange()) {
                setCombatState(prev => applyStateChange(prev, event.getStateChange()));
            }
            
            setEvents(prev => [...prev, event]);
        });
        
        stream.on('error', (err) => {
            console.error('Combat stream error:', err);
        });
        
        return () => {
            stream.cancel();
        };
    }, [sessionId]);
    
    return { combatState, events };
}
```

## Benefits Over REST + WebSocket

1. **Single Connection**: One gRPC stream replaces multiple REST calls + WebSocket
2. **Ordered Delivery**: gRPC ensures messages arrive in order
3. **Backpressure**: Built-in flow control prevents overwhelming clients
4. **Error Handling**: Structured error propagation through the stream
5. **Type Safety**: Generated code ensures type consistency

## Stream Management

```go
type StreamManager struct {
    streams map[string]map[string]chan *pb.CombatEvent // sessionID -> participantID -> channel
    mu      sync.RWMutex
}

func (m *StreamManager) BroadcastToSession(sessionID string, event *pb.CombatEvent) {
    m.mu.RLock()
    defer m.mu.RUnlock()
    
    if participants, ok := m.streams[sessionID]; ok {
        for _, ch := range participants {
            select {
            case ch <- event:
            default:
                // Channel full, log and skip
            }
        }
    }
}

func (m *StreamManager) SendToParticipant(sessionID, participantID string, event *pb.CombatEvent) {
    m.mu.RLock()
    defer m.mu.RUnlock()
    
    if participants, ok := m.streams[sessionID]; ok {
        if ch, ok := participants[participantID]; ok {
            select {
            case ch <- event:
            default:
                // Channel full
            }
        }
    }
}
```

## Database Choice: Alternative to MongoDB

Given your concerns about MongoDB, here are better alternatives for this use case:

### 1. **DynamoDB** (Recommended)
```go
// Simple key-value with complex attributes
type Character struct {
    PK string `dynamodbav:"PK"` // USER#<user_id>
    SK string `dynamodbav:"SK"` // CHARACTER#<character_id>
    
    // All character data as nested attributes
    Name       string     `dynamodbav:"name"`
    Level      int        `dynamodbav:"level"`
    Race       Race       `dynamodbav:"race"`
    Class      Class      `dynamodbav:"class"`
    Attributes Attributes `dynamodbav:"attributes"`
    // ... etc
}

// Single-table design with GSIs for queries
// GSI1: character name lookups
// GSI2: active characters by user
```

### 2. **PostgreSQL with JSONB**
```sql
-- Best of both worlds: relational + document
CREATE TABLE characters (
    id UUID PRIMARY KEY,
    user_id TEXT NOT NULL,
    name TEXT NOT NULL,
    data JSONB NOT NULL, -- All character data
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Indexes on JSONB fields
CREATE INDEX idx_character_user ON characters(user_id);
CREATE INDEX idx_character_name ON characters(name);
CREATE INDEX idx_character_level ON characters((data->>'level')::int);
```

### 3. **BadgerDB** (Embedded)
```go
// For small deployments, embedded key-value store
db, _ := badger.Open(badger.DefaultOptions("./data"))

// Simple key patterns
key := fmt.Sprintf("character:%s:%s", userID, characterID)
```

My recommendation: **DynamoDB** for production or **PostgreSQL with JSONB** if you want SQL backup. Both handle scaling better than MongoDB while keeping document flexibility.