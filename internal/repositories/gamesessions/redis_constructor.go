package gamesessions

import (
	"github.com/KirkDiggler/dnd-bot-discord/internal/uuid"
	"github.com/redis/go-redis/v9"
)

// NewRedis creates a new Redis-backed session repository with default configuration
func NewRedis(client redis.UniversalClient) Repository {
	return NewRedisRepository(&RedisRepoConfig{
		Client:        client,
		UUIDGenerator: uuid.NewGoogleUUIDGenerator(),
		SessionTTL:    sessionTTL,
	})
}
