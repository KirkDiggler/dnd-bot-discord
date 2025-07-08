package core

import (
	"fmt"
	"strings"
)

// Router manages handlers for a specific domain
type Router struct {
	// Domain name (e.g., "character", "combat")
	domain string

	// Handlers organized by pattern
	handlers map[string]Handler

	// Middleware specific to this router
	middleware []Middleware

	// CustomID builder for this domain
	customIDBuilder *CustomIDBuilder

	// Parent pipeline to register with
	pipeline *Pipeline
}

// NewRouter creates a new domain router
func NewRouter(domain string, pipeline *Pipeline) *Router {
	return &Router{
		domain:          domain,
		handlers:        make(map[string]Handler),
		middleware:      make([]Middleware, 0),
		customIDBuilder: NewCustomIDBuilder(domain),
		pipeline:        pipeline,
	}
}

// Use adds middleware to this router
func (r *Router) Use(middleware ...Middleware) *Router {
	r.middleware = append(r.middleware, middleware...)
	return r
}

// Handle registers a handler for a specific action pattern
func (r *Router) Handle(pattern string, handler Handler) *Router {
	// Apply router middleware to handler
	wrapped := handler
	for i := len(r.middleware) - 1; i >= 0; i-- {
		wrapped = r.middleware[i](wrapped)
	}

	r.handlers[pattern] = wrapped
	return r
}

// HandleFunc registers a handler function
func (r *Router) HandleFunc(pattern string, fn func(*InteractionContext) (*HandlerResult, error)) *Router {
	return r.Handle(pattern, HandlerFunc(fn))
}

// Command registers a slash command handler
func (r *Router) Command(name string, handler Handler) *Router {
	pattern := fmt.Sprintf("cmd:%s", name)
	return r.Handle(pattern, handler)
}

// CommandFunc registers a slash command handler function
func (r *Router) CommandFunc(name string, fn func(*InteractionContext) (*HandlerResult, error)) *Router {
	return r.Command(name, HandlerFunc(fn))
}

// Subcommand registers a subcommand handler
func (r *Router) Subcommand(parent, sub string, handler Handler) *Router {
	pattern := fmt.Sprintf("cmd:%s:%s", parent, sub)
	return r.Handle(pattern, handler)
}

// SubcommandFunc registers a subcommand handler function
func (r *Router) SubcommandFunc(parent, sub string, fn func(*InteractionContext) (*HandlerResult, error)) *Router {
	return r.Subcommand(parent, sub, HandlerFunc(fn))
}

// Component registers a component interaction handler
func (r *Router) Component(action string, handler Handler) *Router {
	pattern := fmt.Sprintf("component:%s", action)
	return r.Handle(pattern, handler)
}

// ComponentFunc registers a component interaction handler function
func (r *Router) ComponentFunc(action string, fn func(*InteractionContext) (*HandlerResult, error)) *Router {
	return r.Component(action, HandlerFunc(fn))
}

// Modal registers a modal submit handler
func (r *Router) Modal(action string, handler Handler) *Router {
	pattern := fmt.Sprintf("modal:%s", action)
	return r.Handle(pattern, handler)
}

// ModalFunc registers a modal submit handler function
func (r *Router) ModalFunc(action string, fn func(*InteractionContext) (*HandlerResult, error)) *Router {
	return r.Modal(action, HandlerFunc(fn))
}

// Build creates a single handler from all registered routes
func (r *Router) Build() Handler {
	return &routerHandler{
		domain:   r.domain,
		handlers: r.handlers,
	}
}

// Register registers this router with the pipeline
func (r *Router) Register() {
	if r.pipeline != nil {
		r.pipeline.Register(r.Build())
	}
}

// GetCustomIDBuilder returns the CustomID builder for this router
func (r *Router) GetCustomIDBuilder() *CustomIDBuilder {
	return r.customIDBuilder
}

// routerHandler implements Handler for a router
type routerHandler struct {
	domain   string
	handlers map[string]Handler
}

// CanHandle checks if this router can handle the interaction
func (h *routerHandler) CanHandle(ctx *InteractionContext) bool {
	pattern := h.extractPattern(ctx)
	if pattern == "" {
		return false
	}

	// Check exact match
	if _, ok := h.handlers[pattern]; ok {
		return true
	}

	// Check wildcard patterns
	parts := strings.Split(pattern, ":")
	for i := len(parts); i > 0; i-- {
		wildcardPattern := strings.Join(parts[:i], ":") + ":*"
		if _, ok := h.handlers[wildcardPattern]; ok {
			return true
		}
	}

	return false
}

// Handle processes the interaction
func (h *routerHandler) Handle(ctx *InteractionContext) (*HandlerResult, error) {
	pattern := h.extractPattern(ctx)
	if pattern == "" {
		return nil, NewNotFoundError("handler")
	}

	// Try exact match first
	if handler, ok := h.handlers[pattern]; ok {
		return handler.Handle(ctx)
	}

	// Try wildcard patterns
	parts := strings.Split(pattern, ":")
	for i := len(parts); i > 0; i-- {
		wildcardPattern := strings.Join(parts[:i], ":") + ":*"
		if handler, ok := h.handlers[wildcardPattern]; ok {
			return handler.Handle(ctx)
		}
	}

	return nil, NewNotFoundError("handler")
}

// extractPattern extracts the routing pattern from the interaction
func (h *routerHandler) extractPattern(ctx *InteractionContext) string {
	if ctx.IsCommand() {
		// Check if it's our domain command
		if ctx.GetCommandName() != h.domain {
			return ""
		}

		// Build pattern from subcommand
		sub := ctx.GetSubcommand()
		if sub != "" {
			return fmt.Sprintf("cmd:%s:%s", h.domain, sub)
		}
		return fmt.Sprintf("cmd:%s", h.domain)
	}

	if ctx.IsComponent() {
		// Parse custom ID
		customID, err := ParseCustomID(ctx.GetCustomID())
		if err != nil || customID.Domain != h.domain {
			return ""
		}
		return fmt.Sprintf("component:%s", customID.Action)
	}

	if ctx.IsModal() {
		// Parse custom ID
		customID, err := ParseCustomID(ctx.GetCustomID())
		if err != nil || customID.Domain != h.domain {
			return ""
		}
		return fmt.Sprintf("modal:%s", customID.Action)
	}

	return ""
}

// RouteBuilder provides a fluent API for building routes
type RouteBuilder struct {
	router  *Router
	pattern string
}

// Route creates a new route builder
func (r *Router) Route(pattern string) *RouteBuilder {
	return &RouteBuilder{
		router:  r,
		pattern: pattern,
	}
}

// Handler sets the handler for this route
func (b *RouteBuilder) Handler(handler Handler) *Router {
	return b.router.Handle(b.pattern, handler)
}

// HandlerFunc sets the handler function for this route
func (b *RouteBuilder) HandlerFunc(fn func(*InteractionContext) (*HandlerResult, error)) *Router {
	return b.Handler(HandlerFunc(fn))
}
