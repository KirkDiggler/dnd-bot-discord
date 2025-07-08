package middleware

import (
	"log"
	"time"

	"github.com/KirkDiggler/dnd-bot-discord/internal/discord/v2/core"
)

// DeferConfig configures the defer middleware
type DeferConfig struct {
	// AlwaysDefer forces deferred response for all interactions
	AlwaysDefer bool

	// EphemeralByDefault makes deferred responses ephemeral by default
	EphemeralByDefault bool

	// DeferAfter defers if handler doesn't respond within this duration
	// Set to 0 to disable auto-defer
	DeferAfter time.Duration

	// SkipDeferFor allows skipping defer for specific domains/actions
	SkipDeferFor []DeferSkipRule
}

// DeferSkipRule defines when to skip deferring
type DeferSkipRule struct {
	Domain string
	Action string
}

// DefaultDeferConfig returns a sensible default configuration
func DefaultDeferConfig() *DeferConfig {
	return &DeferConfig{
		AlwaysDefer:        false,
		EphemeralByDefault: false,
		DeferAfter:         2 * time.Second, // Discord requires response within 3s
		SkipDeferFor:       []DeferSkipRule{},
	}
}

// DeferMiddleware automatically handles Discord's 3-second response requirement
func DeferMiddleware(config *DeferConfig) core.Middleware {
	if config == nil {
		config = DefaultDeferConfig()
	}

	return func(next core.Handler) core.Handler {
		return core.HandlerFunc(func(ctx *core.InteractionContext) (*core.HandlerResult, error) {
			// Get responder from context
			responderVal := ctx.Value("responder")
			responder, ok := responderVal.(*core.DiscordResponder)
			if !ok {
				// If no responder, just pass through
				return next.Handle(ctx)
			}

			// Check if we should skip defer
			if shouldSkipDefer(ctx, config) {
				return next.Handle(ctx)
			}

			// If always defer is enabled, defer immediately
			if config.AlwaysDefer {
				if err := responder.Defer(config.EphemeralByDefault); err != nil {
					log.Printf("Failed to defer interaction: %v", err)
				}
				result, err := next.Handle(ctx)
				if result != nil {
					result.Deferred = true
				}
				return result, err
			}

			// If DeferAfter is set, use a timer
			if config.DeferAfter > 0 {
				// Create channels for result and completion
				type handlerResponse struct {
					result *core.HandlerResult
					err    error
				}
				responseChan := make(chan handlerResponse, 1)

				// Run handler in goroutine
				go func() {
					result, err := next.Handle(ctx)
					responseChan <- handlerResponse{result, err}
				}()

				// Create timer for defer
				timer := time.NewTimer(config.DeferAfter)
				defer timer.Stop()

				select {
				case resp := <-responseChan:
					// Handler completed before timeout
					return resp.result, resp.err

				case <-timer.C:
					// Timer expired, send defer
					ephemeral := config.EphemeralByDefault

					// Check if handler suggests ephemeral
					if ctx.IsComponent() {
						// Components often want ephemeral responses
						ephemeral = true
					}

					if err := responder.Defer(ephemeral); err != nil {
						log.Printf("Failed to defer interaction after timeout: %v", err)
					}

					// Wait for handler to complete
					resp := <-responseChan
					if resp.result != nil {
						resp.result.Deferred = true
					}
					return resp.result, resp.err
				}
			}

			// No defer configuration, just pass through
			return next.Handle(ctx)
		})
	}
}

// shouldSkipDefer checks if defer should be skipped for this interaction
func shouldSkipDefer(ctx *core.InteractionContext, config *DeferConfig) bool {
	// Parse custom ID for component interactions
	if ctx.IsComponent() {
		customID, err := core.ParseCustomID(ctx.GetCustomID())
		if err == nil {
			for _, rule := range config.SkipDeferFor {
				if rule.Domain == customID.Domain &&
					(rule.Action == "*" || rule.Action == customID.Action) {
					return true
				}
			}
		}
	}

	// Check command name for slash commands
	if ctx.IsCommand() {
		commandName := ctx.GetCommandName()
		for _, rule := range config.SkipDeferFor {
			if rule.Domain == commandName {
				return true
			}
		}
	}

	return false
}

// AlwaysDeferMiddleware is a simple middleware that always defers
func AlwaysDeferMiddleware() core.Middleware {
	return DeferMiddleware(&DeferConfig{
		AlwaysDefer: true,
	})
}

// SmartDeferMiddleware defers after 2 seconds if handler hasn't responded
func SmartDeferMiddleware() core.Middleware {
	return DeferMiddleware(DefaultDeferConfig())
}
