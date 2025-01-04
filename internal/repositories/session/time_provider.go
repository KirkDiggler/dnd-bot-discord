package session

import "time"

//go:generate mockgen -destination=mocks/mock_time_provider.go -package=mocks github.com/KirkDiggler/dnd-bot-discord/internal/repositories/session TimeProvider

type TimeProvider interface {
	Now() time.Time
}
