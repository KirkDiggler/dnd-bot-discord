# Discord Handler Middleware

This package provides common middleware for Discord interaction handlers.

## Available Middleware

### Defer Middleware
Handles Discord's 3-second response requirement automatically.

```go
// Always defer immediately
pipeline.Use(AlwaysDeferMiddleware())

// Smart defer - waits 2s then defers if no response
pipeline.Use(SmartDeferMiddleware())

// Custom configuration
pipeline.Use(DeferMiddleware(&DeferConfig{
    DeferAfter: 1 * time.Second,
    EphemeralByDefault: true,
    SkipDeferFor: []DeferSkipRule{
        {Domain: "ping", Action: "*"},
    },
}))
```

### Error Middleware
Handles errors gracefully with user-friendly messages.

```go
// Basic error handling
pipeline.Use(ErrorMiddleware(nil))

// Custom error handling
pipeline.Use(ErrorMiddleware(&ErrorConfig{
    LogErrors: true,
    ErrorFormatter: func(err error) string {
        // Custom formatting
        return "Oops! Something went wrong."
    },
}))

// Recovery from panics
pipeline.Use(RecoveryMiddleware())

// Validation error enhancement
pipeline.Use(ValidationErrorMiddleware())
```

### Logging Middleware
Provides request/response logging and metrics.

```go
// Basic logging
pipeline.Use(LoggingMiddleware(nil))

// Custom logging
pipeline.Use(LoggingMiddleware(&LogConfig{
    LogRequests: true,
    LogResponses: true,
    LogDuration: true,
    RequestFilter: func(ctx *InteractionContext) bool {
        // Only log commands
        return ctx.IsCommand()
    },
}))

// Request ID tracking
pipeline.Use(RequestIDMiddleware())

// Metrics collection
pipeline.Use(MetricsMiddleware(metricsCollector))
```

### Authorization Middleware
Controls access to handlers based on roles, permissions, etc.

```go
// Require specific role
pipeline.Use(RoleRequiredMiddleware("role_id_123"))

// Require permissions
pipeline.Use(PermissionRequiredMiddleware(
    discordgo.PermissionManageGuild | 
    discordgo.PermissionManageChannels,
))

// Owner only
pipeline.Use(OwnerOnlyMiddleware("owner_user_id"))

// Complex authorization
pipeline.Use(AuthorizationMiddleware(&AuthConfig{
    RequireGuildMember: true,
    RequiredRoles: []string{"admin_role", "mod_role"},
    UserWhitelist: []string{"special_user_id"},
    CustomChecker: func(ctx *InteractionContext) (bool, string) {
        // Custom logic
        return true, ""
    },
}))

// Domain-specific auth
pipeline.Use(DomainAuthMiddleware(map[string]*AuthConfig{
    "admin": {RequiredRoles: []string{"admin_role"}},
    "mod": {RequiredPermissions: discordgo.PermissionKickMembers},
}))
```

### Rate Limiting Middleware
Prevents abuse by limiting request frequency.

```go
// Per-user rate limiting
pipeline.Use(UserRateLimitMiddleware(10, 1*time.Minute))

// Per-guild rate limiting
pipeline.Use(GuildRateLimitMiddleware(100, 1*time.Minute))

// Per-command rate limiting
pipeline.Use(CommandRateLimitMiddleware(5, 30*time.Second))

// Custom rate limiting
pipeline.Use(RateLimitMiddleware(&RateLimitConfig{
    MaxRequests: 10,
    Window: 5 * time.Minute,
    KeyFunc: func(ctx *InteractionContext) string {
        // Custom key generation
        return fmt.Sprintf("%s:%s", ctx.GuildID, ctx.UserID)
    },
    Store: redisRateLimitStore, // Use Redis for distributed rate limiting
}))
```

## Middleware Order

The order of middleware matters. Here's a recommended order:

```go
pipeline := core.NewPipeline()

// 1. Recovery - catch panics first
pipeline.Use(RecoveryMiddleware())

// 2. Request ID - for tracking
pipeline.Use(RequestIDMiddleware())

// 3. Logging - log all requests
pipeline.Use(LoggingMiddleware(nil))

// 4. Rate limiting - prevent abuse
pipeline.Use(UserRateLimitMiddleware(100, 1*time.Hour))

// 5. Authorization - check permissions
pipeline.Use(AuthorizationMiddleware(authConfig))

// 6. Defer - handle Discord timeout
pipeline.Use(SmartDeferMiddleware())

// 7. Error handling - catch handler errors
pipeline.Use(ErrorMiddleware(nil))

// 8. Metrics - track performance
pipeline.Use(MetricsMiddleware(collector))
```

## Creating Custom Middleware

```go
func MyCustomMiddleware(config *MyConfig) core.Middleware {
    return func(next core.Handler) core.Handler {
        return core.HandlerFunc(func(ctx *core.InteractionContext) (*core.HandlerResult, error) {
            // Before handler
            log.Printf("Before: %s", ctx.GetCommandName())
            
            // Call next handler
            result, err := next.Handle(ctx)
            
            // After handler
            log.Printf("After: %s", ctx.GetCommandName())
            
            // Modify result if needed
            if result != nil && result.Response != nil {
                result.Response.Content += "\n\n_Processed by custom middleware_"
            }
            
            return result, err
        })
    }
}
```

## Testing Middleware

The middleware are designed to be testable:

```go
func TestMyMiddleware(t *testing.T) {
    // Create mock handler
    called := false
    handler := core.HandlerFunc(func(ctx *core.InteractionContext) (*core.HandlerResult, error) {
        called = true
        return &core.HandlerResult{
            Response: core.NewResponse("test"),
        }, nil
    })
    
    // Apply middleware
    wrapped := MyCustomMiddleware(nil)(handler)
    
    // Create test context
    ctx := createTestContext()
    
    // Execute
    result, err := wrapped.Handle(ctx)
    
    // Assert
    assert.True(t, called)
    assert.NoError(t, err)
    assert.NotNil(t, result)
}
```