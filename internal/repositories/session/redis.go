package session

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/KirkDiggler/dnd-bot-discord/internal/entities"
	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/errgroup"
)

type Data struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	DraftID   string    `json:"draft_id"`
	LastToken string    `json:"last_token"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type redisRepo struct {
	client       *redis.Client
	timeProvider TimeProvider
}

func NewRedis(redisClient *redis.Client, timeProvider TimeProvider) Repository {
	return &redisRepo{
		client:       redisClient,
		timeProvider: timeProvider,
	}
}

func (r *redisRepo) Set(ctx context.Context, session *entities.Session) error {
	if session == nil {
		return errors.New("session cannot be nil")
	}

	data := Data{
		ID:        session.ID,
		UserID:    session.UserID,
		DraftID:   session.DraftID,
		LastToken: session.LastToken,
		CreatedAt: session.CreatedAt,
		UpdatedAt: session.UpdatedAt,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal session data: %w", err)
	}

	pipe := r.client.Pipeline()
	pipe.Set(ctx, fmt.Sprintf("session:%s", session.ID), string(jsonData), 0)
	pipe.SAdd(ctx, fmt.Sprintf("user:%s:sessions", session.UserID), session.ID)
	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to set session in Redis: %w", err)
	}

	return nil
}

func (r *redisRepo) Create(ctx context.Context, session *entities.Session) error {
	if session == nil {
		return errors.New("session cannot be nil")
	}

	now := r.timeProvider.Now()
	session.CreatedAt = now
	session.UpdatedAt = now

	return r.Set(ctx, session)
}

func (r *redisRepo) Get(ctx context.Context, id string) (*entities.Session, error) {
	jsonData, err := r.client.Get(ctx, fmt.Sprintf("session:%s", id)).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("session not found")
		}
		return nil, fmt.Errorf("failed to get session from Redis: %w", err)
	}

	var data Data
	err = json.Unmarshal(jsonData, &data)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal session data: %w", err)
	}

	return toSession(&data), nil
}

func (r *redisRepo) Update(ctx context.Context, session *entities.Session) error {
	if session == nil {
		return errors.New("session cannot be nil")
	}

	session.UpdatedAt = r.timeProvider.Now()

	return r.Set(ctx, session)
}

func (r *redisRepo) Delete(ctx context.Context, id string) error {
	session, err := r.Get(ctx, id)
	if err != nil {
		return err
	}

	pipe := r.client.Pipeline()
	pipe.Del(ctx, fmt.Sprintf("session:%s", id))
	pipe.SRem(ctx, fmt.Sprintf("user:%s:sessions", session.UserID), id)
	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete session from Redis: %w", err)
	}

	return nil
}

func (r *redisRepo) ListByUser(ctx context.Context, userID string) ([]*entities.Session, error) {
	sessionIDs, err := r.client.SMembers(ctx, fmt.Sprintf("user:%s:sessions", userID)).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get user sessions from Redis: %w", err)
	}

	sessions := make([]*entities.Session, len(sessionIDs))

	g, ctx := errgroup.WithContext(ctx)
	for i, id := range sessionIDs {
		g.Go(func() error {
			session, err := r.Get(ctx, id)
			if err != nil {
				return fmt.Errorf("failed to get session %s: %w", id, err)
			}
			sessions[i] = session
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return sessions, nil
}

func toSessionData(session *entities.Session) *Data {
	if session == nil {
		return nil
	}

	return &Data{
		ID:        session.ID,
		UserID:    session.UserID,
		DraftID:   session.DraftID,
		LastToken: session.LastToken,
		CreatedAt: session.CreatedAt,
		UpdatedAt: session.UpdatedAt,
	}
}

func toSession(data *Data) *entities.Session {
	if data == nil {
		return nil
	}

	return &entities.Session{
		ID:        data.ID,
		UserID:    data.UserID,
		DraftID:   data.DraftID,
		LastToken: data.LastToken,
		CreatedAt: data.CreatedAt,
		UpdatedAt: data.UpdatedAt,
		Draft:     &entities.CharacterDraft{}, // Initialize with empty draft
	}
}
