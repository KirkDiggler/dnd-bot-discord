package middleware

import (
	"errors"
	"fmt"
	"log"

	"github.com/KirkDiggler/dnd-bot-discord/internal/discord/v2/core"
)

// ErrorConfig configures error handling behavior
type ErrorConfig struct {
	// LogErrors controls whether errors are logged
	LogErrors bool

	// IncludeStackTrace includes stack trace in logs (not user messages)
	IncludeStackTrace bool

	// DefaultUserMessage is shown when no user-friendly message exists
	DefaultUserMessage string

	// ErrorFormatter allows custom error formatting
	ErrorFormatter ErrorFormatter

	// ErrorLogger allows custom logging
	ErrorLogger ErrorLogger
}

// ErrorFormatter formats errors for user display
type ErrorFormatter func(err error) string

// ErrorLogger logs errors
type ErrorLogger func(ctx *core.InteractionContext, err error)

// DefaultErrorConfig returns sensible defaults
func DefaultErrorConfig() *ErrorConfig {
	return &ErrorConfig{
		LogErrors:          true,
		IncludeStackTrace:  false,
		DefaultUserMessage: "An error occurred while processing your request.",
		ErrorFormatter:     defaultErrorFormatter,
		ErrorLogger:        defaultErrorLogger,
	}
}

// ErrorMiddleware handles errors from handlers
func ErrorMiddleware(config *ErrorConfig) core.Middleware {
	if config == nil {
		config = DefaultErrorConfig()
	}

	return func(next core.Handler) core.Handler {
		return core.HandlerFunc(func(ctx *core.InteractionContext) (*core.HandlerResult, error) {
			// Call the next handler
			result, err := next.Handle(ctx)

			// If no error, pass through
			if err == nil {
				return result, nil
			}

			// Log the error if configured
			if config.LogErrors && config.ErrorLogger != nil {
				config.ErrorLogger(ctx, err)
			}

			// Create error response
			response := createErrorResponse(err, config)

			// Return error result (don't propagate error up)
			return &core.HandlerResult{
				Response: response,
				Context: map[string]interface{}{
					"error": err,
				},
			}, nil
		})
	}
}

// RecoveryMiddleware recovers from panics
func RecoveryMiddleware() core.Middleware {
	return func(next core.Handler) core.Handler {
		return core.HandlerFunc(func(ctx *core.InteractionContext) (result *core.HandlerResult, err error) {
			// Recover from panics
			defer func() {
				if r := recover(); r != nil {
					// Log the panic
					log.Printf("Panic recovered in handler: %v", r)

					// Convert panic to error
					switch v := r.(type) {
					case error:
						err = v
					case string:
						err = errors.New(v)
					default:
						err = fmt.Errorf("panic: %v", r)
					}

					// Create error response
					result = &core.HandlerResult{
						Response: core.NewEphemeralResponse("An unexpected error occurred. Please try again later."),
					}
				}
			}()

			// Call next handler
			return next.Handle(ctx)
		})
	}
}

// createErrorResponse creates a user-friendly error response
func createErrorResponse(err error, config *ErrorConfig) *core.Response {
	var message string

	// Check if it's a HandlerError with a user message
	var handlerErr *core.HandlerError
	if errors.As(err, &handlerErr) && handlerErr.ShowToUser {
		message = handlerErr.UserMessage
	} else if config.ErrorFormatter != nil {
		// Use custom formatter
		message = config.ErrorFormatter(err)
	} else {
		// Use default message
		message = config.DefaultUserMessage
	}

	// Always make error responses ephemeral
	return core.NewEphemeralResponse(message)
}

// defaultErrorFormatter provides basic error formatting
func defaultErrorFormatter(err error) string {
	// Try to extract user-friendly message from known error types
	var handlerErr *core.HandlerError
	if errors.As(err, &handlerErr) && handlerErr.ShowToUser {
		return handlerErr.UserMessage
	}

	// Generic message for unknown errors
	return "An error occurred while processing your request."
}

// defaultErrorLogger provides basic error logging
func defaultErrorLogger(ctx *core.InteractionContext, err error) {
	// Build log context
	logCtx := map[string]interface{}{
		"user_id":    ctx.UserID,
		"guild_id":   ctx.GuildID,
		"channel_id": ctx.ChannelID,
	}

	// Add interaction-specific context
	if ctx.IsCommand() {
		logCtx["command"] = ctx.GetCommandName()
		logCtx["subcommand"] = ctx.GetSubcommand()
	} else if ctx.IsComponent() {
		if customID, parseErr := core.ParseCustomID(ctx.GetCustomID()); parseErr == nil {
			logCtx["domain"] = customID.Domain
			logCtx["action"] = customID.Action
		}
	} else if ctx.IsModal() {
		logCtx["modal_id"] = ctx.GetCustomID()
	}

	// Log the error
	log.Printf("Handler error: %v, context: %+v", err, logCtx)
}

// ValidationErrorMiddleware checks for validation errors and provides helpful messages
func ValidationErrorMiddleware() core.Middleware {
	return func(next core.Handler) core.Handler {
		return core.HandlerFunc(func(ctx *core.InteractionContext) (*core.HandlerResult, error) {
			result, err := next.Handle(ctx)

			// Check for validation errors
			if err != nil {
				var validationErr *core.HandlerError
				if errors.As(err, &validationErr) && validationErr.Code == core.ErrorCodeBadRequest {
					// Enhance validation error messages
					message := fmt.Sprintf("❌ **Validation Error**\n%s", validationErr.UserMessage)
					return &core.HandlerResult{
						Response: core.NewEphemeralResponse(message),
					}, nil
				}
			}

			return result, err
		})
	}
}
