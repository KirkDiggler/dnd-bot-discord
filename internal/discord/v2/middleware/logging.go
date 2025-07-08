package middleware

import (
	"log"
	"time"

	"github.com/KirkDiggler/dnd-bot-discord/internal/discord/v2/core"
)

// LogConfig configures logging behavior
type LogConfig struct {
	// LogRequests logs incoming interactions
	LogRequests bool

	// LogResponses logs outgoing responses
	LogResponses bool

	// LogDuration logs handler execution time
	LogDuration bool

	// LogErrors logs errors (if not using ErrorMiddleware)
	LogErrors bool

	// Logger allows custom logging implementation
	Logger Logger

	// RequestFilter filters which requests to log
	RequestFilter func(*core.InteractionContext) bool
}

// Logger is a custom logging interface
type Logger interface {
	LogRequest(ctx *core.InteractionContext)
	LogResponse(ctx *core.InteractionContext, result *core.HandlerResult, duration time.Duration)
	LogError(ctx *core.InteractionContext, err error)
}

// DefaultLogConfig returns sensible defaults
func DefaultLogConfig() *LogConfig {
	return &LogConfig{
		LogRequests:  true,
		LogResponses: false,
		LogDuration:  true,
		LogErrors:    true,
		Logger:       &defaultLogger{},
	}
}

// LoggingMiddleware provides request/response logging
func LoggingMiddleware(config *LogConfig) core.Middleware {
	if config == nil {
		config = DefaultLogConfig()
	}

	return func(next core.Handler) core.Handler {
		return core.HandlerFunc(func(ctx *core.InteractionContext) (*core.HandlerResult, error) {
			// Check filter
			if config.RequestFilter != nil && !config.RequestFilter(ctx) {
				return next.Handle(ctx)
			}

			// Log request
			if config.LogRequests && config.Logger != nil {
				config.Logger.LogRequest(ctx)
			}

			// Track start time
			start := time.Now()

			// Call next handler
			result, err := next.Handle(ctx)

			// Calculate duration
			duration := time.Since(start)

			// Log error
			if err != nil && config.LogErrors && config.Logger != nil {
				config.Logger.LogError(ctx, err)
			}

			// Log response
			if config.LogResponses && config.Logger != nil {
				config.Logger.LogResponse(ctx, result, duration)
			} else if config.LogDuration {
				// Just log duration
				logDuration(ctx, duration)
			}

			return result, err
		})
	}
}

// MetricsMiddleware tracks handler metrics
func MetricsMiddleware(collector MetricsCollector) core.Middleware {
	return func(next core.Handler) core.Handler {
		return core.HandlerFunc(func(ctx *core.InteractionContext) (*core.HandlerResult, error) {
			// Extract labels
			labels := extractLabels(ctx)

			// Track request
			collector.IncrementCounter("discord_interactions_total", labels)

			// Track duration
			start := time.Now()
			result, err := next.Handle(ctx)
			duration := time.Since(start)

			// Track duration
			collector.ObserveHistogram("discord_interaction_duration_seconds", duration.Seconds(), labels)

			// Track errors
			if err != nil {
				errorLabels := make(map[string]string)
				for k, v := range labels {
					errorLabels[k] = v
				}

				// Add error type
				if handlerErr, ok := err.(*core.HandlerError); ok {
					errorLabels["error_code"] = string(rune(handlerErr.Code))
				} else {
					errorLabels["error_code"] = "500"
				}

				collector.IncrementCounter("discord_interactions_errors_total", errorLabels)
			}

			return result, err
		})
	}
}

// MetricsCollector collects metrics
type MetricsCollector interface {
	IncrementCounter(name string, labels map[string]string)
	ObserveHistogram(name string, value float64, labels map[string]string)
}

// defaultLogger provides basic stdout logging
type defaultLogger struct{}

func (l *defaultLogger) LogRequest(ctx *core.InteractionContext) {
	if ctx.IsCommand() {
		log.Printf("[Discord] Command: %s/%s, User: %s, Guild: %s",
			ctx.GetCommandName(),
			ctx.GetSubcommand(),
			ctx.UserID,
			ctx.GuildID,
		)
	} else if ctx.IsComponent() {
		customID := ctx.GetCustomID()
		if parsed, err := core.ParseCustomID(customID); err == nil {
			log.Printf("[Discord] Component: %s:%s, User: %s, Guild: %s",
				parsed.Domain,
				parsed.Action,
				ctx.UserID,
				ctx.GuildID,
			)
		} else {
			log.Printf("[Discord] Component: %s, User: %s, Guild: %s",
				customID,
				ctx.UserID,
				ctx.GuildID,
			)
		}
	} else if ctx.IsModal() {
		log.Printf("[Discord] Modal: %s, User: %s, Guild: %s",
			ctx.GetCustomID(),
			ctx.UserID,
			ctx.GuildID,
		)
	}
}

func (l *defaultLogger) LogResponse(ctx *core.InteractionContext, result *core.HandlerResult, duration time.Duration) {
	status := "success"
	if result == nil {
		status = "no_response"
	} else if result.Response != nil && result.Response.Ephemeral {
		status = "ephemeral"
	}

	log.Printf("[Discord] Response: %s, Duration: %v", status, duration)
}

func (l *defaultLogger) LogError(ctx *core.InteractionContext, err error) {
	interaction := "unknown"
	if ctx.IsCommand() {
		interaction = ctx.GetCommandName()
	} else if ctx.IsComponent() {
		interaction = ctx.GetCustomID()
	} else if ctx.IsModal() {
		interaction = ctx.GetCustomID()
	}

	log.Printf("[Discord] Error in %s: %v", interaction, err)
}

// logDuration logs just the duration
func logDuration(ctx *core.InteractionContext, duration time.Duration) {
	interaction := "unknown"
	if ctx.IsCommand() {
		interaction = ctx.GetCommandName()
		if sub := ctx.GetSubcommand(); sub != "" {
			interaction += "/" + sub
		}
	} else if ctx.IsComponent() {
		if parsed, err := core.ParseCustomID(ctx.GetCustomID()); err == nil {
			interaction = parsed.Domain + ":" + parsed.Action
		}
	}

	log.Printf("[Discord] %s completed in %v", interaction, duration)
}

// extractLabels extracts common labels for metrics
func extractLabels(ctx *core.InteractionContext) map[string]string {
	labels := map[string]string{
		"guild_id": ctx.GuildID,
	}

	if ctx.IsCommand() {
		labels["interaction_type"] = "command"
		labels["command"] = ctx.GetCommandName()
		if sub := ctx.GetSubcommand(); sub != "" {
			labels["subcommand"] = sub
		}
	} else if ctx.IsComponent() {
		labels["interaction_type"] = "component"
		if parsed, err := core.ParseCustomID(ctx.GetCustomID()); err == nil {
			labels["domain"] = parsed.Domain
			labels["action"] = parsed.Action
		}
	} else if ctx.IsModal() {
		labels["interaction_type"] = "modal"
	}

	return labels
}

// RequestIDMiddleware adds a unique request ID to the context
func RequestIDMiddleware() core.Middleware {
	return func(next core.Handler) core.Handler {
		return core.HandlerFunc(func(ctx *core.InteractionContext) (*core.HandlerResult, error) {
			// Generate request ID (in production, use a proper UUID)
			requestID := generateRequestID()

			// Add to context
			ctx.WithValue("request_id", requestID)

			// Log with request ID
			log.Printf("[%s] Starting request", requestID)

			return next.Handle(ctx)
		})
	}
}

// generateRequestID generates a simple request ID
// In production, use a proper UUID library
func generateRequestID() string {
	return time.Now().Format("20060102150405.000")
}
