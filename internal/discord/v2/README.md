# Discord V2 Handler System

## Overview

The V2 handler system is a modular, type-safe Discord interaction handler that provides:
- Domain-based routing (e.g., "dnd", "combat")
- Type-safe custom ID encoding/decoding
- Middleware support
- Clean separation of concerns

## Architecture

### Core Components

1. **Pipeline** (`core/pipeline.go`): Main execution engine that processes Discord interactions
2. **Router** (`core/router.go`): Domain-specific handler registration and routing
3. **CustomID** (`core/custom_id.go`): Type-safe encoding/decoding of Discord component custom IDs
4. **InteractionContext** (`core/context.go`): Wraps Discord interaction with helper methods

### Custom ID Format

Custom IDs are encoded as: `domain:action:target:arg1:arg2...`

Example: `creation:preview_race:char_123` 
- Domain: `creation`
- Action: `preview_race` 
- Target: `char_123`

### Handler Registration

Handlers are registered with the router using patterns:

```go
// Command handlers
router.CommandFunc("dnd", handler)                    // /dnd
router.SubcommandFunc("dnd", "character create", handler)  // /dnd character create

// Component handlers (buttons, select menus)
router.ComponentFunc("creation:select", handler)      // Handles creation:select:*
router.ComponentFunc("creation:preview_race", handler) // Handles creation:preview_race:*

// Modal handlers
router.ModalFunc("creation:name_input", handler)
```

### Important Notes

1. **Domain Consistency**: The router domain MUST match the Discord command name
   - Router created with `core.NewRouter("dnd", pipeline)` handles `/dnd` commands
   
2. **Component Registration**: Component handlers need the full pattern including domain
   - ✅ `router.ComponentFunc("creation:preview_race", handler)`
   - ❌ `router.ComponentFunc("preview_race", handler)`

3. **CustomID Builder Usage**:
   ```go
   // For select menus and modals
   customID := h.customIDBuilder.Build("preview_race", characterID)
   
   // For buttons (via ComponentBuilder)
   components.PrimaryButton("Confirm", "confirm_race", characterID)
   ```

## Character Creation Example

The character creation flow demonstrates the v2 system:

1. **Router Setup** (`routers/character.go`):
   ```go
   router := core.NewRouter("dnd", pipeline)
   
   // Register subcommands
   router.SubcommandFunc("dnd", "character create", handler.StartCreation)
   
   // Register component handlers
   router.ComponentFunc("creation:preview_race", handler.HandleRacePreview)
   router.ComponentFunc("creation:confirm_race", handler.HandleConfirmRace)
   ```

2. **Handler Implementation** (`handlers/character_creation.go`):
   - `StartCreation`: Initial command handler
   - `HandleStepSelection`: Generic step processor
   - `HandleRacePreview`: Shows race preview without committing
   - `HandleConfirmRace`: Commits the race selection

3. **Enhanced UI** (`handlers/character_creation_enhanced.go`):
   - Rich character sheet display with sections
   - Dynamic progress tracker
   - Preview functionality for selections

## Common Pitfalls

1. **Handler Not Found**: Usually means:
   - Wrong domain in router registration
   - Missing domain prefix in component registration
   - CustomID not built with proper builder

2. **Router Domain Mismatch**: 
   - Router domain must match Discord command name
   - Check: `core.NewRouter("dnd", ...)` for `/dnd` commands

3. **Component CustomID Format**:
   - Always use the customIDBuilder for consistency
   - Don't manually format: `fmt.Sprintf("action_%s", id)`
   - Do use: `customIDBuilder.Build("action", id)`

## Testing

Integration tests should verify:
- Handler registration and routing
- CustomID encoding/decoding
- Full interaction flows

See `race_preview_integration_test.go` for example testing approach.