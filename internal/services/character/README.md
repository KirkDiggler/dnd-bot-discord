# Character Service Layer

This package implements the business logic for D&D character management, following the Input/Output pattern as requested.

## Architecture

### Service Interface
- Uses Input/Output DTOs instead of Request/Response (which feels too handler-like)
- All methods accept a context for cancellation and tracing
- Returns domain entities and errors
- Input validation with `Validate()` methods on all inputs
- Comprehensive error handling with structured errors

### Key Components

1. **Service** (`service.go`)
   - Main business logic orchestrator
   - Validates inputs
   - Coordinates with other components
   - Will eventually persist to repository

2. **Choice Resolver** (`choice_resolver.go`)
   - Handles complex D&D 5e API choice structures
   - Converts nested choices to simplified UI-friendly format
   - Special handling for cases like Monk's tool choices

3. **Validation** (`validation.go`)
   - Input validation with `Validate()` methods
   - Defensive programming against nil inputs
   - Clear error messages for invalid data

4. **Errors** (`/internal/errors/errors.go`)
   - Structured error types with codes and metadata
   - Error wrapping for context as errors bubble up
   - Type-safe error checking functions

## Domain Language

This service layer uses domain-appropriate terminology:
- `RealmID` instead of Discord's `GuildID` - represents the world/realm where characters exist
- `OwnerID` - the user who owns the character
- Platform-specific translations happen at the handler layer

## Usage Example

```go
// Create service
svc := character.NewService(&character.ServiceConfig{
    DNDClient:  dndClient,
    Repository: repository,
})

// Resolve available choices for a race/class combo
choices, err := svc.ResolveChoices(ctx, &character.ResolveChoicesInput{
    RaceKey:  "human",
    ClassKey: "fighter",
})

// Create a character
output, err := svc.CreateCharacter(ctx, &character.CreateCharacterInput{
    UserID:        "user123",
    RealmID:       "realm456",  // Platform handlers translate to this
    Name:          "Thorin",
    RaceKey:       "dwarf",
    ClassKey:      "fighter",
    AbilityScores: map[string]int{...},
    Proficiencies: []string{...},
    Equipment:     []string{...},
})

// Handle errors with context
if err != nil {
    if dnderr.IsNotFound(err) {
        // Handle not found
    } else if dnderr.IsInvalidArgument(err) {
        // Handle validation error
    }
    
    // Get error metadata
    meta := dnderr.GetMeta(err)
    log.Printf("Error: %v, metadata: %+v", err, meta)
}
```

## Error Handling

The service uses structured errors that:
- Preserve error codes through wrapping
- Accumulate metadata as errors bubble up
- Provide clear context for debugging
- Enable type-safe error checking

Example error chain:
```
failed to save character: database unavailable: connection timeout
Metadata: {
    character_name: "Thorin",
    character_id: "char_123", 
    owner_id: "user_456",
    retry_after: 30,
    connection_pool: "exhausted"
}
```

## Testing

The service layer is fully tested with:
- Test suites using testify/suite
- Mocks generated with uber/mock
- Comprehensive input validation tests
- Error handling and propagation tests

```bash
go test ./internal/services/character/... -v
```

Tests cover:
- Character creation flow
- Input validation
- Choice resolution
- Error handling and propagation
- Defensive programming against nil/invalid inputs

## Next Steps

1. Add Repository layer for persistence ✓
2. Implement ID generation (currently using timestamp)
3. Add more validation for proficiency selections
4. Implement equipment validation
5. Add character update/delete operations
6. Wire up to Discord handlers with realm translation