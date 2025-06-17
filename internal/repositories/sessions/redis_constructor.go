package sessions

import (
	"github.com/KirkDiggler/dnd-bot-discord/internal/uuid"
	"github.com/redis/go-redis/v9"
)

// NewRedis creates a new Redis-backed session repository
func NewRedis(client redis.UniversalClient) Repository {
	repo, err := NewRedisRepository(&RedisConfig{
		Client:        client,
		UUIDGenerator: uuid.NewGoogleUUIDGenerator(),
		TimeProvider:  &RealTimeProvider{},
	})
	if err != nil {
		// This should never happen with valid configuration
		panic(err)
	}
	return repo
}