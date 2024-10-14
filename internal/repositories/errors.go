package repositories

import (
	"github.com/KirkDiggler/dnd-bot-discord/internal"
)

type RepositoryError string

func (e RepositoryError) Error() string {
	return string(e)
}

const (
	ErrRecord RepositoryError = "record error"
)

type RecordError struct {
	internal.ErrorWrapper
}

func NewRecordNotFoundError(id string) error {
	return &RecordError{
		ErrorWrapper: internal.ErrorWrapper{
			Err:     internal.ErrNotFound,
			Message: string(ErrRecord) + ": " + id,
		},
	}
}
