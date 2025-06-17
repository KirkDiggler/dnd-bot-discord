package characters

import (
	"time"

	"github.com/KirkDiggler/dnd-bot-discord/internal/uuid"
	"github.com/redis/go-redis/v9"
)

// NewRedis creates a new Redis-backed character repository
func NewRedis(client redis.UniversalClient) Repository {
	return NewRedisRepository(&RedisRepoConfig{
		Client:        client,
		UUIDGenerator: uuid.NewGoogleUUIDGenerator(),
		DraftTTL:      24 * time.Hour,
	})
}