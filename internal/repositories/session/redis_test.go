package session

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/KirkDiggler/dnd-bot-discord/internal/entities"
	"github.com/KirkDiggler/dnd-bot-discord/internal/repositories/session/mocks"
	"github.com/go-redis/redismock/v9"
	"go.uber.org/mock/gomock"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/suite"
)

type RedisRepoTestSuite struct {
	suite.Suite
	mockClient    *redis.Client
	mock          redismock.ClientMock
	repo          Repository
	mockCtrl      *gomock.Controller
	timeProvider  *mocks.MockTimeProvider
}

func (s *RedisRepoTestSuite) SetupTest() {
	s.mockClient, s.mock = redismock.NewClientMock()
	s.mockCtrl = gomock.NewController(s.T())
	s.timeProvider = mocks.NewMockTimeProvider(s.mockCtrl)
	s.repo = NewRedis(s.mockClient, s.timeProvider)
}

func (s *RedisRepoTestSuite) TearDownTest() {
	s.mockCtrl.Finish()
	s.NoError(s.mock.ExpectationsWereMet())
}

func TestRedisRepoTestSuite(t *testing.T) {
	suite.Run(t, new(RedisRepoTestSuite))
}

func (s *RedisRepoTestSuite) TestSet() {
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Millisecond)
	session := &entities.Session{
		ID:        "test-id",
		UserID:    "user-id",
		DraftID:   "draft-id",
		LastToken: "last-token",
		CreatedAt: now,
		UpdatedAt: now,
	}

	expectedData, err := json.Marshal(Data{
		ID:        session.ID,
		UserID:    session.UserID,
		DraftID:   session.DraftID,
		LastToken: session.LastToken,
		CreatedAt: session.CreatedAt,
		UpdatedAt: session.UpdatedAt,
	})
	s.Require().NoError(err)

	// Happy path
	s.mock.ExpectSet("session:test-id", string(expectedData), 0).SetVal("OK")
	s.mock.ExpectSAdd("user:user-id:sessions", "test-id").SetVal(1)

	err = s.repo.Set(ctx, session)
	s.NoError(err)
	s.mock.ExpectationsWereMet()

	// Dependency error
	s.mock.ExpectSet("session:test-id", string(expectedData), 0).SetErr(errors.New("redis error"))

	err = s.repo.Set(ctx, session)
	s.Error(err)
	s.mock.ExpectationsWereMet()

	// Input validation
	err = s.repo.Set(ctx, nil)
	s.Error(err)
}

func (s *RedisRepoTestSuite) TestCreate() {
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Millisecond)
	s.timeProvider.EXPECT().Now().Return(now)

	session := &entities.Session{
		ID:        "test-id",
		UserID:    "user-id",
		DraftID:   "draft-id",
		LastToken: "last-token",
	}

	expectedData := Data{
		ID:        session.ID,
		UserID:    session.UserID,
		DraftID:   session.DraftID,
		LastToken: session.LastToken,
		CreatedAt: now,
		UpdatedAt: now,
	}
	jsonData, err := json.Marshal(expectedData)
	s.Require().NoError(err)

	s.mock.ExpectSet("session:test-id", string(jsonData), 0).SetVal("OK")
	s.mock.ExpectSAdd("user:user-id:sessions", "test-id").SetVal(1)

	err = s.repo.Create(ctx, session)
	s.NoError(err)
}

func (s *RedisRepoTestSuite) TestGet() {
	ctx := context.Background()
	sessionID := "test-id"
	now := time.Now().UTC().Truncate(time.Millisecond)
	sessionData := Data{
		ID:        sessionID,
		UserID:    "user-id",
		DraftID:   "draft-id",
		LastToken: "last-token",
		CreatedAt: now,
		UpdatedAt: now,
	}
	jsonData, err := json.Marshal(sessionData)
	s.Require().NoError(err)

	// Happy path
	s.mock.ExpectGet("session:test-id").SetVal(string(jsonData))

	session, err := s.repo.Get(ctx, sessionID)
	s.NoError(err)
	s.Equal(sessionID, session.ID)

	// Dependency error
	s.mock.ExpectGet("session:test-id").SetErr(errors.New("redis error"))

	_, err = s.repo.Get(ctx, sessionID)
	s.Error(err)

	// Input validation
	_, err = s.repo.Get(ctx, "")
	s.Error(err)
}

func (s *RedisRepoTestSuite) TestUpdate() {
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Millisecond)
	s.timeProvider.EXPECT().Now().Return(now)

	session := &entities.Session{
		ID:        "test-id",
		UserID:    "user-id",
		DraftID:   "draft-id",
		LastToken: "last-token",
		CreatedAt: now.Add(-1 * time.Hour), // Assume created an hour ago
	}

	expectedData := Data{
		ID:        session.ID,
		UserID:    session.UserID,
		DraftID:   session.DraftID,
		LastToken: session.LastToken,
		CreatedAt: session.CreatedAt,
		UpdatedAt: now,
	}
	jsonData, err := json.Marshal(expectedData)
	s.Require().NoError(err)

	s.mock.ExpectSet("session:test-id", string(jsonData), 0).SetVal("OK")
	s.mock.ExpectSAdd("user:user-id:sessions", "test-id").SetVal(1)

	err = s.repo.Update(ctx, session)
	s.NoError(err)
}

func (s *RedisRepoTestSuite) TestDelete() {
	ctx := context.Background()
	sessionID := "test-id"
	now := time.Now().UTC().Truncate(time.Millisecond)
	sessionData := Data{
		ID:        sessionID,
		UserID:    "user-id",
		DraftID:   "draft-id",
		LastToken: "last-token",
		CreatedAt: now,
		UpdatedAt: now,
	}
	jsonData, err := json.Marshal(sessionData)
	s.Require().NoError(err)

	// Happy path
	s.mock.ExpectGet("session:test-id").SetVal(string(jsonData))
	s.mock.ExpectDel("session:test-id").SetVal(1)
	s.mock.ExpectSRem("user:user-id:sessions", "test-id").SetVal(1)

	err = s.repo.Delete(ctx, sessionID)
	s.NoError(err)
	s.mock.ExpectationsWereMet()

	// Dependency error
	s.mock.ExpectGet("session:test-id").SetErr(errors.New("redis error"))

	err = s.repo.Delete(ctx, sessionID)
	s.Error(err)
	s.mock.ExpectationsWereMet()

	// Input validation
	err = s.repo.Delete(ctx, "")
	s.Error(err)
}

func (s *RedisRepoTestSuite) TestListByUser() {
	ctx := context.Background()
	userID := "user-id"
	sessionIDs := []string{"session-1", "session-2"}
	now := time.Now().UTC().Truncate(time.Millisecond)

	sessionData1 := Data{
		ID:        "session-1",
		UserID:    userID,
		DraftID:   "draft-id-1",
		LastToken: "last-token-1",
		CreatedAt: now,
		UpdatedAt: now,
	}
	jsonData1, err := json.Marshal(sessionData1)
	s.Require().NoError(err)

	sessionData2 := Data{
		ID:        "session-2",
		UserID:    userID,
		DraftID:   "draft-id-2",
		LastToken: "last-token-2",
		CreatedAt: now,
		UpdatedAt: now,
	}
	jsonData2, err := json.Marshal(sessionData2)
	s.Require().NoError(err)

	// Happy path
	s.mock.ExpectSMembers("user:user-id:sessions").SetVal(sessionIDs)
	s.mock.ExpectGet("session:session-2").SetVal(string(jsonData2))
	s.mock.ExpectGet("session:session-1").SetVal(string(jsonData1))

	sessions, err := s.repo.ListByUser(ctx, userID)
	s.NoError(err)
	s.Len(sessions, 2)
	s.Equal("session-1", sessions[0].ID)
	s.Equal("session-2", sessions[1].ID)

	// Dependency error
	s.mock.ExpectSMembers("user:user-id:sessions").SetErr(errors.New("redis error"))

	_, err = s.repo.ListByUser(ctx, userID)
	s.Error(err)

	// Input validation
	_, err = s.repo.ListByUser(ctx, "")
	s.Error(err)
}
