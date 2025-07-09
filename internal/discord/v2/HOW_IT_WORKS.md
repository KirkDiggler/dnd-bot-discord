# How the V2 Handler System Works

## Overview

The V2 handler system is a routing framework that connects Discord interactions to handler functions. It uses a pattern-based routing system similar to web frameworks.

## Core Concepts

### 1. **Custom IDs**
Discord components (buttons, select menus, etc.) have a `customID` field that identifies them. The V2 system encodes structured data into these IDs:

```
domain:action:target:arg1:arg2...
```

Example: `dnd:preview_race:char_123`
- `dnd` = domain (which router handles this)
- `preview_race` = action (what to do)
- `char_123` = target (what to do it on)

### 2. **Pipeline**
The Pipeline is the main entry point that receives all Discord interactions and routes them to the appropriate handlers.

### 3. **Router**
Each Router handles a specific domain (e.g., "dnd" for D&D commands). The router:
- Registers handlers for specific patterns
- Extracts routing information from interactions
- Matches patterns to find the right handler

### 4. **Handlers**
Handlers are functions that process specific interactions and return responses.

## Flow Walkthrough

### Step 1: Bot Startup & Registration

```go
// 1. Create pipeline
pipeline := core.NewPipeline()

// 2. Create router for "dnd" domain
router := core.NewRouter("dnd", pipeline)

// 3. Register handlers
router.ComponentFunc("preview_race", handleRacePreview)

// 4. Register router with pipeline
router.Register()
```

This creates a handler mapping:
- Pattern: `component:preview_race`
- Handler: `handleRacePreview`

### Step 2: Building UI Components

When creating a select menu:
```go
components.SelectMenuWithTarget(
    "Select a race...",
    "preview_race",    // action
    char.ID,           // target
    options,
)
```

This creates a select menu with customID: `dnd:preview_race:char_123`

### Step 3: User Interaction

When a user selects from the menu, Discord sends:
```json
{
  "type": 3,  // MESSAGE_COMPONENT
  "custom_id": "dnd:preview_race:char_123",
  "values": ["elf"]
}
```

### Step 4: Pipeline Processing

```go
// Pipeline receives interaction
pipeline.Execute(ctx, session, interaction)

// 1. Pipeline checks all registered routers
// 2. Each router's CanHandle method is called
```

### Step 5: Router Pattern Extraction

The router's `extractPattern` method:
```go
func (h *routerHandler) extractPattern(ctx *InteractionContext) string {
    if ctx.IsComponent() {
        // Parse: "dnd:preview_race:char_123"
        customID, err := ParseCustomID(ctx.GetCustomID())
        if err != nil || customID.Domain != h.domain {
            return ""  // Not for this router
        }
        // Returns: "component:preview_race"
        return fmt.Sprintf("component:%s", customID.Action)
    }
}
```

### Step 6: Handler Lookup & Execution

```go
// Router looks up pattern in its handlers map
handler := h.handlers["component:preview_race"]

// Execute the handler
result, err := handler.Handle(ctx)
```

### Step 7: Handler Processing

```go
func handleRacePreview(ctx *InteractionContext) (*HandlerResult, error) {
    // 1. Parse custom ID to get character ID
    customID, _ := ParseCustomID(ctx.GetCustomID())
    characterID := customID.Target  // "char_123"
    
    // 2. Get selected value
    data := ctx.Interaction.MessageComponentData()
    selectedRace := data.Values[0]  // "elf"
    
    // 3. Process and return response
    return &HandlerResult{
        Response: response,
    }, nil
}
```

## Pattern Types

### Commands
- Pattern: `cmd:dnd` or `cmd:dnd:character create`
- Triggered by: Slash commands

### Components
- Pattern: `component:preview_race`
- Triggered by: Buttons, select menus

### Modals
- Pattern: `modal:character_name`
- Triggered by: Modal submissions

## Domain Architecture

### The Domain Constraint

Discord slash commands create a constraint: the base command (e.g., `/dnd`) determines the router domain. This means:

1. **Slash Command Routers**: Must use the base command as their domain
   - `/dnd character create` → Router domain must be "dnd"
   - `/admin user ban` → Router domain must be "admin"

2. **Component Custom IDs**: Can use any domain
   - You could have: `creation:preview_race:char_123`
   - Or: `combat:attack:monster_456`

### Design Options

1. **Single Router per Command** (Current approach)
   - One "dnd" router handles all `/dnd` subcommands and related components
   - Simple but can become large

2. **Multiple Specialized Routers**
   ```go
   // Slash command router
   dndRouter := NewRouter("dnd", pipeline)
   dndRouter.SubcommandFunc("dnd", "character create", forwardToCreation)
   
   // Specialized component router  
   creationRouter := NewRouter("creation", pipeline)
   creationRouter.ComponentFunc("preview_race", handlePreview)
   ```

3. **Subdomain Components**
   - Keep one router but use component action names as subdomains
   - `creation_preview_race`, `combat_attack`, etc.

## Key Design Principles

1. **Domain Isolation**: Each router handles only its domain
2. **Type Safety**: Custom IDs are parsed into structured data
3. **Middleware Support**: Cross-cutting concerns (logging, auth)
4. **Testability**: Each component can be tested in isolation

## Common Pitfalls

1. **Domain Mismatch**: Router domain must match custom ID domain
2. **Pattern Registration**: Component handlers need just the action, not the full pattern
3. **Custom ID Length**: Discord limits to 100 characters

## Example: Complete Flow

1. User types `/dnd character create`
2. Pipeline routes to "dnd" router's `cmd:dnd:character create` handler
3. Handler shows race selection with custom ID `dnd:preview_race:char_123`
4. User selects "Elf"
5. Pipeline routes to `component:preview_race` handler
6. Handler shows preview with "Confirm" button
7. User clicks confirm
8. Process continues to next step