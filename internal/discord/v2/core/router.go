package core

import (
	"fmt"
	"log"
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

// Action registers an action handler for the router's domain
// Example: router.Action("show", handler) on a "character" router handles /dnd character show
func (r *Router) Action(action string, handler Handler) *Router {
	pattern := fmt.Sprintf("cmd:dnd:%s:%s", r.domain, action)
	return r.Handle(pattern, handler)
}

// ActionFunc registers an action handler function
func (r *Router) ActionFunc(action string, fn func(*InteractionContext) (*HandlerResult, error)) *Router {
	return r.Action(action, HandlerFunc(fn))
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

	// Log for debugging
	log.Printf("[Router %s] Pattern extracted: %s, Available handlers: %v", h.domain, pattern, h.getHandlerKeys())

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

// getHandlerKeys returns all registered handler keys for debugging
func (h *routerHandler) getHandlerKeys() []string {
	keys := make([]string, 0, len(h.handlers))
	for k := range h.handlers {
		keys = append(keys, k)
	}
	return keys
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
		// For /dnd commands, the domain is the first part of the subcommand
		// Example: /dnd character show -> domain="character", action="show"
		if ctx.GetCommandName() != "dnd" {
			// We only handle /dnd commands
			return ""
		}

		sub := ctx.GetSubcommand()
		if sub == "" {
			// Just /dnd with no subcommand
			return ""
		}

		// Split subcommand to extract domain and action
		// "character show" -> ["character", "show"]
		parts := strings.SplitN(sub, " ", 2)
		domain := parts[0]

		// Check if this router handles this domain
		if domain != h.domain {
			log.Printf("[Router %s] Command domain mismatch: got %s", h.domain, domain)
			return ""
		}

		// Build pattern based on what we have
		if len(parts) > 1 {
			// Has action: /dnd character show
			action := parts[1]
			pattern := fmt.Sprintf("cmd:%s:%s:%s", "dnd", domain, action)
			log.Printf("[Router %s] Command pattern: %s", h.domain, pattern)
			return pattern
		}

		// No action: /dnd character (would be like an index)
		pattern := fmt.Sprintf("cmd:%s:%s", "dnd", domain)
		log.Printf("[Router %s] Command pattern (no action): %s", h.domain, pattern)
		return pattern
	}

	if ctx.IsComponent() {
		// Parse custom ID
		rawCustomID := ctx.GetCustomID()
		log.Printf("[Router %s] Parsing component customID: %s", h.domain, rawCustomID)

		customID, err := ParseCustomID(rawCustomID)
		if err != nil {
			log.Printf("[Router %s] Failed to parse customID: %v", h.domain, err)
			return ""
		}

		log.Printf("[Router %s] Parsed - Domain: %s, Action: %s, Target: %s",
			h.domain, customID.Domain, customID.Action, customID.Target)

		if customID.Domain != h.domain {
			log.Printf("[Router %s] Domain mismatch (want %s, got %s)", h.domain, h.domain, customID.Domain)
			return ""
		}

		pattern := fmt.Sprintf("component:%s", customID.Action)
		log.Printf("[Router %s] Returning pattern: %s", h.domain, pattern)
		return pattern
	}

	if ctx.IsModal() {
		// Parse custom ID
		rawCustomID := ctx.GetCustomID()
		log.Printf("[Router %s] Parsing modal customID: %s", h.domain, rawCustomID)

		customID, err := ParseCustomID(rawCustomID)
		if err != nil {
			log.Printf("[Router %s] Failed to parse modal customID: %v", h.domain, err)
			return ""
		}

		log.Printf("[Router %s] Modal parsed - Domain: %s, Action: %s", h.domain, customID.Domain, customID.Action)

		if customID.Domain != h.domain {
			log.Printf("[Router %s] Modal domain mismatch (want %s, got %s)", h.domain, h.domain, customID.Domain)
			return ""
		}

		pattern := fmt.Sprintf("modal:%s", customID.Action)
		log.Printf("[Router %s] Returning modal pattern: %s", h.domain, pattern)
		return pattern
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
