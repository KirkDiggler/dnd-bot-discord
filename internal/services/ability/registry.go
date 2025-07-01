package ability

import (
	"sync"
)

// HandlerRegistry manages ability handlers
type HandlerRegistry struct {
	mu       sync.RWMutex
	handlers map[string]Handler
}

// NewHandlerRegistry creates a new handler registry
func NewHandlerRegistry() *HandlerRegistry {
	return &HandlerRegistry{
		handlers: make(map[string]Handler),
	}
}

// Register adds a handler to the registry
func (r *HandlerRegistry) Register(handler Handler) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.handlers[handler.Key()] = handler
}

// Get retrieves a handler by key
func (r *HandlerRegistry) Get(key string) (Handler, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	handler, exists := r.handlers[key]
	return handler, exists
}

// List returns all registered handler keys
func (r *HandlerRegistry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	keys := make([]string, 0, len(r.handlers))
	for key := range r.handlers {
		keys = append(keys, key)
	}
	return keys
}
