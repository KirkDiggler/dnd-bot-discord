package session

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/KirkDiggler/dnd-bot-discord/internal"
	"github.com/KirkDiggler/dnd-bot-discord/internal/entities"
	"github.com/KirkDiggler/dnd-bot-discord/internal/repositories"
	"github.com/KirkDiggler/dnd-bot-discord/internal/uuid"
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
	client        redis.UniversalClient
	timeProvider  TimeProvider
	uuidGenerator uuid.Generator
}

type RedisConfig struct {
	Client        redis.UniversalClient
	TimeProvider  TimeProvider
	UUIDGenerator uuid.Generator
}

func NewRedis(cfg *RedisConfig) (*redisRepo, error) {
	if cfg == nil {
		return nil, fmt.Errorf("session.NewRedis %w", internal.NewMissingParamError("cfg"))
	}

	if cfg.Client == nil {
		return nil, fmt.Errorf("session.NewRedis %w", internal.NewMissingParamError("client"))
	}

	if cfg.TimeProvider == nil {
		return nil, fmt.Errorf("session.NewRedis %w", internal.NewMissingParamError("timeProvider"))
	}

	if cfg.UUIDGenerator == nil {
		return nil, fmt.Errorf("session.NewRedis %w", internal.NewMissingParamError("uuidGenerator"))
	}

	return &redisRepo{
		client:        cfg.Client,
		timeProvider:  cfg.TimeProvider,
		uuidGenerator: cfg.UUIDGenerator,
	}, nil
}

func (r *redisRepo) Set(ctx context.Context, session *entities.Session) error {
	if session == nil {
		return fmt.Errorf("session.Set %w", internal.NewMissingParamError("session"))
	}

	if session.ID == "" {
		return fmt.Errorf("session.Set %w", internal.NewMissingParamError("session.ID"))
	}

	if session.UserID == "" {
		return fmt.Errorf("session.Set %w", internal.NewMissingParamError("session.UserID"))
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

	if session.ID != "" {
		return fmt.Errorf("session.Create %w", internal.NewInvalidParamError("ID cannot be set"))
	}

	if session.UserID == "" {
		return fmt.Errorf("session.Create %w", internal.NewMissingParamError("UserID"))
	}

	now := r.timeProvider.Now()
	session.ID = r.uuidGenerator.New()
	session.CreatedAt = now
	session.UpdatedAt = now

	return r.Set(ctx, session)
}

func (r *redisRepo) get(ctx context.Context, id string) (*entities.Session, error) {
	jsonData, err := r.client.Get(ctx, fmt.Sprintf("session:%s", id)).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, repositories.NewRecordNotFoundError(id)
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

func (r *redisRepo) Get(ctx context.Context, id string) (*entities.Session, error) {
	if id == "" {
		return nil, fmt.Errorf("session.Get %w", internal.NewMissingParamError("id"))
	}

	session, err := r.get(ctx, id)
	if err != nil {
		if errors.Is(err, internal.ErrNotFound) {
			return nil, fmt.Errorf("session.Get %w", NewSessionNotFoundError(id))
		}
		return nil, err
	}

	return session, nil
}

func (r *redisRepo) Update(ctx context.Context, session *entities.Session) (*entities.Session, error) {
	if session == nil {
		return nil, fmt.Errorf("session.Update %w", internal.NewMissingParamError("session"))
	}

	if session.ID == "" {
		return nil, fmt.Errorf("session.Update %w", internal.NewMissingParamError("ID"))
	}

	if session.UserID == "" {
		return nil, fmt.Errorf("session.Update %w", internal.NewMissingParamError("UserID"))
	}

	// Check if the session exists before updating
	existing, err := r.get(ctx, session.ID)
	if err != nil {
		if errors.Is(err, internal.ErrNotFound) {
			return nil, fmt.Errorf("session.Get %w", NewSessionNotFoundError(session.ID))
		}

		return nil, err
	}

	session.CreatedAt = existing.CreatedAt
	session.UpdatedAt = r.timeProvider.Now()

	setErr := r.Set(ctx, session)
	if setErr != nil {
		return nil, setErr
	}

	return session, nil
}

func (r *redisRepo) Delete(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("session.Delete %w", internal.NewMissingParamError("id"))
	}

	session, err := r.get(ctx, id)
	if err != nil {
		if errors.Is(err, internal.ErrNotFound) {
			return NewSessionNotFoundError(id)
		}

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
	if userID == "" {
		return nil, fmt.Errorf("session.ListByUser %w", internal.NewMissingParamError("userID"))
	}

	sessionIDs, err := r.client.SMembers(ctx, fmt.Sprintf("user:%s:sessions", userID)).Result()
	if err != nil {
		if err == redis.Nil {
			return []*entities.Session{}, nil
		}

		return nil, fmt.Errorf("failed to get user sessions from Redis: %w", err)
	}

	if len(sessionIDs) == 0 {
		return []*entities.Session{}, nil
	}

	sessions := make([]*entities.Session, len(sessionIDs))

	g, ctx := errgroup.WithContext(ctx)
	for i, id := range sessionIDs {
		g.Go(func() error {
			session, err := r.get(ctx, id)
			if err != nil {
				if errors.Is(err, internal.ErrNotFound) {
					// If a session is not found, we'll just skip it
					return nil
				}
				return fmt.Errorf("failed to get session %s: %w", id, err)
			}
			sessions[i] = session
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	// Remove any nil sessions (those that were not found)
	validSessions := []*entities.Session{}
	for _, s := range sessions {
		if s != nil {
			validSessions = append(validSessions, s)
		}
	}

	return validSessions, nil
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
