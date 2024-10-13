package repositories

import (
	"fmt"
	"github.com/KirkDiggler/dnd-bot-discord/internal"
)

type RecordError struct {
	Err error
	ID  string
}

func (e *RecordError) Error() string {
	return fmt.Sprintf("record %s: %v", e.ID, e.Err)
}

func (e *RecordError) Unwrap() error {
	return e.Err
}

func NewRecordNotFoundError(id string) error {
	return &RecordError{
		Err: internal.ErrNotFound,
		ID:  id,
	}
}
