package gamesessions

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/game/session"
	"time"

	"github.com/KirkDiggler/dnd-bot-discord/internal/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	// Key patterns
	sessionKeyPrefix    = "session:"
	inviteCodeKeyPrefix = "invite:"
	realmSessionsKey    = "realm:%s:sessions"
	userSessionsKey     = "user:%s:sessions"

	// TTL for sessions (7 days)
	sessionTTL = 7 * 24 * time.Hour
)

// RedisRepoConfig holds configuration for the Redis repository
type RedisRepoConfig struct {
	Client        redis.UniversalClient
	UUIDGenerator uuid.Generator
	SessionTTL    time.Duration
}

// redisRepository implements Repository using Redis
type redisRepository struct {
	client        redis.UniversalClient
	uuidGenerator uuid.Generator
	sessionTTL    time.Duration
}

// NewRedisRepository creates a new Redis-backed session repository
func NewRedisRepository(cfg *RedisRepoConfig) Repository {
	if cfg.Client == nil {
		panic("redis client is required")
	}

	ttl := cfg.SessionTTL
	if ttl == 0 {
		ttl = sessionTTL
	}

	return &redisRepository{
		client:        cfg.Client,
		uuidGenerator: cfg.UUIDGenerator,
		sessionTTL:    ttl,
	}
}

// Create creates a new session
func (r *redisRepository) Create(ctx context.Context, session *session.Session) error {
	if session == nil {
		return fmt.Errorf("session cannot be nil")
	}
	if session.ID == "" {
		return fmt.Errorf("session ID cannot be empty")
	}

	// Serialize session
	data, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("failed to serialize session: %w", err)
	}

	// Use pipeline for atomic operations
	pipe := r.client.TxPipeline()

	// Check if session already exists
	sessionKey := sessionKeyPrefix + session.ID
	pipe.Exists(ctx, sessionKey)

	// Store session
	pipe.Set(ctx, sessionKey, data, r.sessionTTL)

	// Store invite code mapping if present
	if session.InviteCode != "" {
		inviteKey := inviteCodeKeyPrefix + session.InviteCode
		pipe.Set(ctx, inviteKey, session.ID, r.sessionTTL)
	}

	// Add to realm index
	pipe.SAdd(ctx, fmt.Sprintf(realmSessionsKey, session.RealmID), session.ID)

	// Add to user indexes for all members
	for userID := range session.Members {
		pipe.SAdd(ctx, fmt.Sprintf(userSessionsKey, userID), session.ID)
	}

	// Execute pipeline
	cmds, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	if _, ok := cmds[0].(*redis.IntCmd); !ok {
		return fmt.Errorf("failed to create session: %w", err)
	}

	// Check if session already existed
	if cmdVal, ok := cmds[0].(*redis.IntCmd); ok {
		if cmdVal.Val() > 0 {
			return fmt.Errorf("session with ID %s already exists", session.ID)
		}
	}

	return nil
}

// Get retrieves a session by ID
func (r *redisRepository) Get(ctx context.Context, id string) (*session.Session, error) {
	sessionKey := sessionKeyPrefix + id

	data, err := r.client.Get(ctx, sessionKey).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("session not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	var session session.Session
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, fmt.Errorf("failed to deserialize session: %w", err)
	}

	// Refresh TTL
	r.client.Expire(ctx, sessionKey, r.sessionTTL)

	return &session, nil
}

// GetByInviteCode retrieves a session by its invite code
func (r *redisRepository) GetByInviteCode(ctx context.Context, code string) (*session.Session, error) {
	inviteKey := inviteCodeKeyPrefix + code

	sessionID, err := r.client.Get(ctx, inviteKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("no session found with invite code: %s", code)
		}
		return nil, fmt.Errorf("failed to get session by invite code: %w", err)
	}

	// Refresh invite code TTL
	r.client.Expire(ctx, inviteKey, r.sessionTTL)

	return r.Get(ctx, sessionID)
}

// Update updates an existing session
func (r *redisRepository) Update(ctx context.Context, session *session.Session) error {
	if session == nil {
		return fmt.Errorf("session cannot be nil")
	}
	if session.ID == "" {
		return fmt.Errorf("session ID cannot be empty")
	}

	sessionKey := sessionKeyPrefix + session.ID

	// Get existing session to check for changes
	existing, err := r.Get(ctx, session.ID)
	if err != nil {
		return fmt.Errorf("session not found: %s", session.ID)
	}

	// Serialize updated session
	data, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("failed to serialize session: %w", err)
	}

	// Use pipeline for atomic operations
	pipe := r.client.TxPipeline()

	// Update session
	pipe.Set(ctx, sessionKey, data, r.sessionTTL)

	// Handle invite code changes
	if existing.InviteCode != session.InviteCode {
		// Remove old invite code mapping
		if existing.InviteCode != "" {
			oldInviteKey := inviteCodeKeyPrefix + existing.InviteCode
			pipe.Del(ctx, oldInviteKey)
		}

		// Add new invite code mapping
		if session.InviteCode != "" {
			newInviteKey := inviteCodeKeyPrefix + session.InviteCode
			pipe.Set(ctx, newInviteKey, session.ID, r.sessionTTL)
		}
	}

	// Update user indexes for member changes
	existingMembers := make(map[string]bool)
	for userID := range existing.Members {
		existingMembers[userID] = true
	}

	// Add new members to index
	for userID := range session.Members {
		if !existingMembers[userID] {
			pipe.SAdd(ctx, fmt.Sprintf(userSessionsKey, userID), session.ID)
		}
	}

	// Remove departed members from index
	for userID := range existingMembers {
		if _, exists := session.Members[userID]; !exists {
			pipe.SRem(ctx, fmt.Sprintf(userSessionsKey, userID), session.ID)
		}
	}

	// Execute pipeline
	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}

	return nil
}

// Delete removes a session
func (r *redisRepository) Delete(ctx context.Context, id string) error {
	// Get session to clean up indexes
	session, err := r.Get(ctx, id)
	if err != nil {
		return fmt.Errorf("session not found: %s", id)
	}

	// Use pipeline for atomic operations
	pipe := r.client.TxPipeline()

	// Delete session
	sessionKey := sessionKeyPrefix + id
	pipe.Del(ctx, sessionKey)

	// Delete invite code mapping
	if session.InviteCode != "" {
		inviteKey := inviteCodeKeyPrefix + session.InviteCode
		pipe.Del(ctx, inviteKey)
	}

	// Remove from realm index
	pipe.SRem(ctx, fmt.Sprintf(realmSessionsKey, session.RealmID), id)

	// Remove from user indexes
	for userID := range session.Members {
		pipe.SRem(ctx, fmt.Sprintf(userSessionsKey, userID), id)
	}

	// Execute pipeline
	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	return nil
}

// GetByRealm retrieves all sessions for a realm
func (r *redisRepository) GetByRealm(ctx context.Context, realmID string) ([]*session.Session, error) {
	// Get session IDs for realm
	sessionIDs, err := r.client.SMembers(ctx, fmt.Sprintf(realmSessionsKey, realmID)).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get sessions for realm: %w", err)
	}

	return r.getMultipleSessions(ctx, sessionIDs)
}

// GetByUser retrieves all sessions a user is part of
func (r *redisRepository) GetByUser(ctx context.Context, userID string) ([]*session.Session, error) {
	// Get session IDs for user
	sessionIDs, err := r.client.SMembers(ctx, fmt.Sprintf(userSessionsKey, userID)).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get sessions for user: %w", err)
	}

	return r.getMultipleSessions(ctx, sessionIDs)
}

// GetActiveByRealm retrieves all active sessions for a realm
func (r *redisRepository) GetActiveByRealm(ctx context.Context, realmID string) ([]*session.Session, error) {
	sessions, err := r.GetByRealm(ctx, realmID)
	if err != nil {
		return nil, err
	}

	// Filter active sessions
	var activeSessions []*session.Session
	for _, sessionValue := range sessions {
		if sessionValue.Status == session.SessionStatusPlanning ||
			sessionValue.Status == session.SessionStatusActive ||
			sessionValue.Status == session.SessionStatusPaused {
			activeSessions = append(activeSessions, sessionValue)
		}
	}

	return activeSessions, nil
}

// GetActiveByUser retrieves all active sessions a user is part of
func (r *redisRepository) GetActiveByUser(ctx context.Context, userID string) ([]*session.Session, error) {
	sessions, err := r.GetByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Filter active sessions
	var activeSessions []*session.Session
	for _, sessionValue := range sessions {
		if sessionValue.Status == session.SessionStatusPlanning ||
			sessionValue.Status == session.SessionStatusActive ||
			sessionValue.Status == session.SessionStatusPaused {
			activeSessions = append(activeSessions, sessionValue)
		}
	}

	return activeSessions, nil
}

// getMultipleSessions retrieves multiple sessions by their IDs
func (r *redisRepository) getMultipleSessions(ctx context.Context, sessionIDs []string) ([]*session.Session, error) {
	if len(sessionIDs) == 0 {
		return []*session.Session{}, nil
	}

	// Build keys
	keys := make([]string, len(sessionIDs))
	for i, id := range sessionIDs {
		keys[i] = sessionKeyPrefix + id
	}

	// Get all sessions
	values, err := r.client.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get multiple sessions: %w", err)
	}

	// Deserialize sessions
	sessions := make([]*session.Session, 0, len(sessionIDs))
	for i, val := range values {
		if val == nil {
			// Session was deleted, remove from index
			// This is handled lazily during reads
			continue
		}

		data, ok := val.(string)
		if !ok {
			continue
		}

		var session session.Session
		if err := json.Unmarshal([]byte(data), &session); err != nil {
			// Log error but continue with other sessions
			continue
		}

		sessions = append(sessions, &session)

		// Refresh TTL
		r.client.Expire(ctx, keys[i], r.sessionTTL)
	}

	return sessions, nil
}
