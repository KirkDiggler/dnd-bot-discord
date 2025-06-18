package dungeons

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/KirkDiggler/dnd-bot-discord/internal/entities"
	"github.com/redis/go-redis/v9"
)

// RedisRepoConfig holds configuration for the Redis repository
type RedisRepoConfig struct {
	Client *redis.Client
}

// redisRepository implements Repository using Redis
type redisRepository struct {
	client *redis.Client
}

// NewRedisRepository creates a new Redis-backed repository
func NewRedisRepository(cfg *RedisRepoConfig) Repository {
	if cfg == nil || cfg.Client == nil {
		panic("RedisRepoConfig and Client are required")
	}
	
	return &redisRepository{
		client: cfg.Client,
	}
}

// Create creates a new dungeon
func (r *redisRepository) Create(ctx context.Context, dungeon *entities.Dungeon) error {
	key := fmt.Sprintf("dungeon:%s", dungeon.ID)
	data, err := json.Marshal(dungeon)
	if err != nil {
		return err
	}
	
	if err := r.client.Set(ctx, key, data, 0).Err(); err != nil {
		return err
	}
	
	// Also add to session index
	sessionKey := fmt.Sprintf("session:%s:dungeons", dungeon.SessionID)
	if err := r.client.SAdd(ctx, sessionKey, dungeon.ID).Err(); err != nil {
		return err
	}
	
	// Add to active index if active
	if dungeon.IsActive() {
		if err := r.client.SAdd(ctx, "dungeons:active", dungeon.ID).Err(); err != nil {
			return err
		}
	}
	
	return nil
}

// Get retrieves a dungeon by ID
func (r *redisRepository) Get(ctx context.Context, id string) (*entities.Dungeon, error) {
	key := fmt.Sprintf("dungeon:%s", id)
	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}
	
	var dungeon entities.Dungeon
	if err := json.Unmarshal(data, &dungeon); err != nil {
		return nil, err
	}
	
	return &dungeon, nil
}

// Update updates an existing dungeon
func (r *redisRepository) Update(ctx context.Context, dungeon *entities.Dungeon) error {
	key := fmt.Sprintf("dungeon:%s", dungeon.ID)
	data, err := json.Marshal(dungeon)
	if err != nil {
		return err
	}
	
	if err := r.client.Set(ctx, key, data, 0).Err(); err != nil {
		return err
	}
	
	// Update active index
	if dungeon.IsActive() {
		if err := r.client.SAdd(ctx, "dungeons:active", dungeon.ID).Err(); err != nil {
			return err
		}
	} else {
		if err := r.client.SRem(ctx, "dungeons:active", dungeon.ID).Err(); err != nil {
			return err
		}
	}
	
	return nil
}

// Delete removes a dungeon
func (r *redisRepository) Delete(ctx context.Context, id string) error {
	// Get dungeon first to find session ID
	dungeon, err := r.Get(ctx, id)
	if err != nil {
		return err
	}
	if dungeon == nil {
		return nil
	}
	
	// Remove from session index
	sessionKey := fmt.Sprintf("session:%s:dungeons", dungeon.SessionID)
	if err := r.client.SRem(ctx, sessionKey, id).Err(); err != nil {
		return err
	}
	
	// Remove from active index
	if err := r.client.SRem(ctx, "dungeons:active", id).Err(); err != nil {
		return err
	}
	
	// Delete the dungeon
	key := fmt.Sprintf("dungeon:%s", id)
	return r.client.Del(ctx, key).Err()
}

// GetBySession retrieves a dungeon by session ID
func (r *redisRepository) GetBySession(ctx context.Context, sessionID string) (*entities.Dungeon, error) {
	sessionKey := fmt.Sprintf("session:%s:dungeons", sessionID)
	ids, err := r.client.SMembers(ctx, sessionKey).Result()
	if err != nil {
		return nil, err
	}
	
	// Find the active dungeon for this session
	for _, id := range ids {
		dungeon, err := r.Get(ctx, id)
		if err != nil {
			continue
		}
		if dungeon != nil && dungeon.IsActive() {
			return dungeon, nil
		}
	}
	
	return nil, nil
}

// ListActive retrieves all active dungeons
func (r *redisRepository) ListActive(ctx context.Context) ([]*entities.Dungeon, error) {
	ids, err := r.client.SMembers(ctx, "dungeons:active").Result()
	if err != nil {
		return nil, err
	}
	
	var dungeons []*entities.Dungeon
	for _, id := range ids {
		dungeon, err := r.Get(ctx, id)
		if err != nil {
			continue
		}
		if dungeon != nil {
			dungeons = append(dungeons, dungeon)
		}
	}
	
	return dungeons, nil
}