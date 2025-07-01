package sessions

import (
	"context"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/game/session"
)

// Repository defines the interface for session storage operations
type Repository interface {
	Set(ctx context.Context, session *session.Session) error
	Create(ctx context.Context, session *session.Session) error
	Get(ctx context.Context, id string) (*session.Session, error)
	Update(ctx context.Context, session *session.Session) (*session.Session, error)
	Delete(ctx context.Context, id string) error
	ListByUser(ctx context.Context, userID string) ([]*session.Session, error)
}
