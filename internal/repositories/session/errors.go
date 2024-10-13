package session

import (
	"fmt"
	"github.com/KirkDiggler/dnd-bot-discord/internal/repositories"
)

type SessionError struct {
	Err error
}

func (e *SessionError) Error() string {
	return fmt.Sprintf("session %v", e.Err)
}

func (e *SessionError) Unwrap() error {
	return e.Err
}

func NewSessionNotFoundError(id string) error {
	return &SessionError{
		Err: repositories.NewRecordNotFoundError(id),
	}
}
