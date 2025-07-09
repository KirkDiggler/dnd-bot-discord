package core

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/bwmarrin/discordgo"
)

// Pipeline manages handler registration and execution
type Pipeline struct {
	// Handlers registered in the pipeline
	handlers []Handler

	// Middleware to apply to all handlers
	middleware []Middleware

	// Error handler for uncaught errors
	errorHandler ErrorHandler

	// Whether to stop on first handler that can handle
	stopOnFirst bool

	// Mutex for thread-safe handler registration
	mu sync.RWMutex
}

// Middleware is a function that wraps a handler
type Middleware func(Handler) Handler

// ErrorHandler handles errors that occur during pipeline execution
type ErrorHandler func(ctx *InteractionContext, err error) *HandlerResult

// NewPipeline creates a new handler pipeline
func NewPipeline() *Pipeline {
	return &Pipeline{
		handlers:     make([]Handler, 0),
		middleware:   make([]Middleware, 0),
		errorHandler: defaultErrorHandler,
		stopOnFirst:  true,
	}
}

// Register adds handlers to the pipeline
func (p *Pipeline) Register(handlers ...Handler) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, h := range handlers {
		// Apply all middleware to the handler
		wrapped := h
		for i := len(p.middleware) - 1; i >= 0; i-- {
			wrapped = p.middleware[i](wrapped)
		}
		p.handlers = append(p.handlers, wrapped)
	}
}

// Use adds middleware to the pipeline
func (p *Pipeline) Use(middleware ...Middleware) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.middleware = append(p.middleware, middleware...)
}

// SetErrorHandler sets a custom error handler
func (p *Pipeline) SetErrorHandler(handler ErrorHandler) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.errorHandler = handler
}

// SetStopOnFirst configures whether to stop after the first handler that can handle
func (p *Pipeline) SetStopOnFirst(stop bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.stopOnFirst = stop
}

// Execute runs the pipeline for an interaction
func (p *Pipeline) Execute(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) error {
	// Create interaction context
	interactionCtx := NewInteractionContext(ctx, s, i)

	log.Printf("[Pipeline] Executing for command: %s, subcommand: %s",
		interactionCtx.GetCommandName(),
		interactionCtx.GetSubcommand())

	// Create responder
	responder := NewDiscordResponder(s, i)
	interactionCtx.WithValue("responder", responder)

	p.mu.RLock()
	handlers := make([]Handler, len(p.handlers))
	copy(handlers, p.handlers)
	stopOnFirst := p.stopOnFirst
	errorHandler := p.errorHandler
	p.mu.RUnlock()

	log.Printf("[Pipeline] Found %d handlers", len(handlers))

	// Track if any handler processed the interaction
	handled := false

	// Execute handlers
	for idx, handler := range handlers {
		if !handler.CanHandle(interactionCtx) {
			log.Printf("[Pipeline] Handler %d cannot handle this interaction", idx)
			continue
		}
		log.Printf("[Pipeline] Handler %d can handle, executing...", idx)
		result, err := handler.Handle(interactionCtx)

		if err != nil {
			// Use error handler
			result = errorHandler(interactionCtx, err)
		}

		// Send response if we have one
		if result != nil && result.Response != nil {
			if err := p.sendResponse(responder, result); err != nil {
				return fmt.Errorf("failed to send response: %w", err)
			}
		}

		handled = true

		// Stop if configured to or if handler requested
		if stopOnFirst || (result != nil && result.StopPropagation) {
			break
		}
	}

	// If no handler processed the interaction, send a default response
	if !handled && !responder.HasResponded() {
		result := &HandlerResult{
			Response: NewEphemeralResponse("I don't know how to handle that command."),
		}
		return p.sendResponse(responder, result)
	}

	return nil
}

// sendResponse sends a response using the responder
func (p *Pipeline) sendResponse(responder *DiscordResponder, result *HandlerResult) error {
	// If already deferred, use edit
	if result.Deferred || responder.IsDeferred() {
		return responder.Edit(result.Response)
	}

	// Otherwise, send initial response
	return responder.Respond(result.Response)
}

// Clear removes all handlers from the pipeline
func (p *Pipeline) Clear() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.handlers = make([]Handler, 0)
}

// HandlerCount returns the number of registered handlers
func (p *Pipeline) HandlerCount() int {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return len(p.handlers)
}

// defaultErrorHandler is the default error handler
func defaultErrorHandler(ctx *InteractionContext, err error) *HandlerResult {
	// Check if it's a handler error with a user message
	if handlerErr, ok := err.(*HandlerError); ok && handlerErr.ShowToUser {
		return &HandlerResult{
			Response: NewEphemeralResponse(handlerErr.UserMessage),
		}
	}

	// Generic error response
	return &HandlerResult{
		Response: NewEphemeralResponse("An error occurred while processing your request."),
	}
}

// MiddlewareChain creates a single middleware from multiple middleware
func MiddlewareChain(middleware ...Middleware) Middleware {
	return func(next Handler) Handler {
		// Apply middleware in reverse order
		for i := len(middleware) - 1; i >= 0; i-- {
			next = middleware[i](next)
		}
		return next
	}
}
