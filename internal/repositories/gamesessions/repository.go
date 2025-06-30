package gamesessions

//go:generate mockgen -destination=mock/mock_repository.go -package=mockgamesessions -source=repository.go

import (
	"context"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/game/session"
)

// Repository defines the interface for game session storage
type Repository interface {
	// Create creates a new session
	Create(ctx context.Context, session *session.Session) error

	// Get retrieves a session by ID
	Get(ctx context.Context, id string) (*session.Session, error)

	// GetByInviteCode retrieves a session by its invite code
	GetByInviteCode(ctx context.Context, code string) (*session.Session, error)

	// Update updates an existing session
	Update(ctx context.Context, session *session.Session) error

	// Delete removes a session
	Delete(ctx context.Context, id string) error

	// GetByRealm retrieves all sessions for a realm (Discord server)
	GetByRealm(ctx context.Context, realmID string) ([]*session.Session, error)

	// GetByUser retrieves all sessions a user is part of
	GetByUser(ctx context.Context, userID string) ([]*session.Session, error)

	// GetActiveByRealm retrieves all active sessions for a realm
	GetActiveByRealm(ctx context.Context, realmID string) ([]*session.Session, error)

	// GetActiveByUser retrieves all active sessions a user is part of
	GetActiveByUser(ctx context.Context, userID string) ([]*session.Session, error)
}
