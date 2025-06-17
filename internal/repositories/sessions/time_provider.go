package sessions

import "time"

//go:generate mockgen -destination=mocks/mock_time_provider.go -package=mocks github.com/KirkDiggler/dnd-bot-discord/internal/repositories/sessions TimeProvider

type TimeProvider interface {
	Now() time.Time
}

type RealTimeProvider struct{}

func (r *RealTimeProvider) Now() time.Time {
	return time.Now()
}
