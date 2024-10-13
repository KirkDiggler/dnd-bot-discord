package session

import (
	"context"

	"github.com/KirkDiggler/dnd-bot-discord/internal/entities"
)

// Repository defines the interface for session storage operations
type Repository interface {
	Set(ctx context.Context, session *entities.Session) error
	Create(ctx context.Context, session *entities.Session) error
	Get(ctx context.Context, id string) (*entities.Session, error)
	Update(ctx context.Context, session *entities.Session) error
	Delete(ctx context.Context, id string) error
	ListByUser(ctx context.Context, userID string) ([]*entities.Session, error)
}
