# Discord Handler Core Package

This package provides the foundational types and interfaces for building modular Discord interaction handlers.

## Core Types

### InteractionContext
Wraps Discord interactions with convenient helpers:
- Automatic parameter extraction from commands, components, and modals
- Type-safe parameter access methods
- Context propagation for cancellation and values
- Convenient helper methods for common checks

### Handler Interface
Simple interface that all handlers implement:
```go
type Handler interface {
    CanHandle(ctx *InteractionContext) bool
    Handle(ctx *InteractionContext) (*HandlerResult, error)
}
```

### HandlerResult
Standardized response from handlers:
- `Response`: Discord-agnostic response structure
- `Deferred`: Whether the response was already deferred
- `StopPropagation`: Stop processing additional handlers
- `Context`: Pass data to middleware

### Response
Discord-agnostic response builder with fluent API:
```go
response := NewResponse("Hello!").
    WithEmbeds(embed).
    WithComponents(button).
    AsEphemeral()
```

### InteractionResponder
Abstraction over Discord's interaction response API:
- Handles defer/respond/edit/followup flow
- Prevents common mistakes (double responding, editing before responding)
- Testable interface for unit tests

### HandlerError
Structured errors with user-friendly messages:
- Separates internal errors from user-facing messages
- HTTP-like status codes for categorization
- Unwrappable for error inspection

## Usage Example

```go
type MyHandler struct {
    service MyService
}

func (h *MyHandler) CanHandle(ctx *InteractionContext) bool {
    return ctx.GetCommandName() == "mycommand"
}

func (h *MyHandler) Handle(ctx *InteractionContext) (*HandlerResult, error) {
    // Extract parameters
    name := ctx.GetStringParam("name")
    
    // Call service
    result, err := h.service.DoSomething(ctx.Context, name)
    if err != nil {
        return nil, NewInternalError(err)
    }
    
    // Build response
    response := NewResponse(fmt.Sprintf("Hello %s!", result.Name)).
        AsEphemeral()
    
    return &HandlerResult{
        Response: response,
    }, nil
}
```

## Design Principles

1. **Discord Abstraction**: Core types hide Discord implementation details where possible
2. **Type Safety**: Structured types instead of maps and interfaces
3. **Testability**: Interfaces allow for easy mocking
4. **Error Handling**: Consistent error structure with user-friendly messages
5. **Composability**: Types designed to work well with middleware and pipelines