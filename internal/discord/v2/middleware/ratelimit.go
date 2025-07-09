package middleware

import (
	"fmt"
	"sync"
	"time"

	"github.com/KirkDiggler/dnd-bot-discord/internal/discord/v2/core"
)

// RateLimitConfig configures rate limiting behavior
type RateLimitConfig struct {
	// MaxRequests is the maximum number of requests allowed
	MaxRequests int

	// Window is the time window for rate limiting
	Window time.Duration

	// KeyFunc extracts the rate limit key from context
	KeyFunc func(*core.InteractionContext) string

	// Message shown when rate limited
	Message string

	// Store for tracking rate limits (if nil, uses in-memory)
	Store RateLimitStore
}

// RateLimitStore tracks rate limit data
type RateLimitStore interface {
	// Increment increments the counter for a key and returns the new count
	Increment(key string, window time.Duration) (int, error)

	// Reset resets the counter for a key
	Reset(key string) error
}

// defaultKeyFunc uses user ID as the rate limit key
func defaultKeyFunc(ctx *core.InteractionContext) string {
	return ctx.UserID
}

// RateLimitMiddleware applies rate limiting
func RateLimitMiddleware(config *RateLimitConfig) core.Middleware {
	// Set defaults
	if config.KeyFunc == nil {
		config.KeyFunc = defaultKeyFunc
	}
	if config.Message == "" {
		config.Message = fmt.Sprintf("You're doing that too fast! Please wait %v before trying again.", config.Window)
	}
	if config.Store == nil {
		config.Store = NewMemoryRateLimitStore()
	}

	return func(next core.Handler) core.Handler {
		return core.HandlerFunc(func(ctx *core.InteractionContext) (*core.HandlerResult, error) {
			// Get rate limit key
			key := config.KeyFunc(ctx)
			if key == "" {
				// No key, skip rate limiting
				return next.Handle(ctx)
			}

			// Check rate limit
			count, err := config.Store.Increment(key, config.Window)
			if err != nil {
				// Log error but don't block request
				return next.Handle(ctx)
			}

			// Check if over limit
			if count > config.MaxRequests {
				return &core.HandlerResult{
					Response: core.NewEphemeralResponse("⏱️ " + config.Message),
				}, nil
			}

			// Continue with request
			return next.Handle(ctx)
		})
	}
}

// UserRateLimitMiddleware applies per-user rate limiting
func UserRateLimitMiddleware(maxRequests int, window time.Duration) core.Middleware {
	return RateLimitMiddleware(&RateLimitConfig{
		MaxRequests: maxRequests,
		Window:      window,
		KeyFunc:     defaultKeyFunc,
	})
}

// GuildRateLimitMiddleware applies per-guild rate limiting
func GuildRateLimitMiddleware(maxRequests int, window time.Duration) core.Middleware {
	return RateLimitMiddleware(&RateLimitConfig{
		MaxRequests: maxRequests,
		Window:      window,
		KeyFunc: func(ctx *core.InteractionContext) string {
			return ctx.GuildID
		},
	})
}

// CommandRateLimitMiddleware applies rate limiting per command
func CommandRateLimitMiddleware(maxRequests int, window time.Duration) core.Middleware {
	return RateLimitMiddleware(&RateLimitConfig{
		MaxRequests: maxRequests,
		Window:      window,
		KeyFunc: func(ctx *core.InteractionContext) string {
			if ctx.IsCommand() {
				return fmt.Sprintf("%s:%s", ctx.UserID, ctx.GetCommandName())
			}
			return ctx.UserID
		},
	})
}

// MemoryRateLimitStore is an in-memory rate limit store
type MemoryRateLimitStore struct {
	mu      sync.RWMutex
	buckets map[string]*bucket
}

type bucket struct {
	count   int
	resetAt time.Time
}

// NewMemoryRateLimitStore creates a new in-memory store
func NewMemoryRateLimitStore() *MemoryRateLimitStore {
	store := &MemoryRateLimitStore{
		buckets: make(map[string]*bucket),
	}

	// Start cleanup goroutine
	go store.cleanup()

	return store
}

// Increment increments the counter for a key
func (s *MemoryRateLimitStore) Increment(key string, window time.Duration) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()

	// Get or create bucket
	b, exists := s.buckets[key]
	if !exists || now.After(b.resetAt) {
		// Create new bucket
		b = &bucket{
			count:   0,
			resetAt: now.Add(window),
		}
		s.buckets[key] = b
	}

	// Increment counter
	b.count++

	return b.count, nil
}

// Reset resets the counter for a key
func (s *MemoryRateLimitStore) Reset(key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.buckets, key)
	return nil
}

// cleanup periodically removes expired buckets
func (s *MemoryRateLimitStore) cleanup() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		s.mu.Lock()
		now := time.Now()

		// Remove expired buckets
		for key, b := range s.buckets {
			if now.After(b.resetAt) {
				delete(s.buckets, key)
			}
		}

		s.mu.Unlock()
	}
}
