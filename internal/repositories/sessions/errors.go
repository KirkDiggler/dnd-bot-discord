package sessions

import (
	"github.com/KirkDiggler/dnd-bot-discord/internal/repositories"
)

type SessionError string

func (e SessionError) Error() string {
	return string(e)
}

const (
	ErrSessionRepo SessionError = "session repository error"
)

type SessionRepositoryError struct {
	repositories.RecordError
}

func NewSessionNotFoundError(id string) error {
	return &SessionRepositoryError{
		RecordError: repositories.RecordError{
			ErrorWrapper: repositories.NewRecordNotFoundError(id).ErrorWrapper,
		},
	}
}
