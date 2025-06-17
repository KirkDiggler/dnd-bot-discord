package sessions

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/KirkDiggler/dnd-bot-discord/internal"
	"github.com/KirkDiggler/dnd-bot-discord/internal/entities"
	"github.com/KirkDiggler/dnd-bot-discord/internal/repositories/sessions/mocks"
	mockUUID "github.com/KirkDiggler/dnd-bot-discord/internal/uuid/mocks"
	"github.com/go-redis/redismock/v9"
	"go.uber.org/mock/gomock"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/suite"
)

type RedisRepoTestSuite struct {
	suite.Suite
	mockClient    *redis.Client
	mock          redismock.ClientMock
	repo          *redisRepo
	mockCtrl      *gomock.Controller
	timeProvider  *mocks.MockTimeProvider
	uuidGenerator *mockUUID.MockGenerator
}

func (s *RedisRepoTestSuite) SetupTest() {
	s.mockClient, s.mock = redismock.NewClientMock()
	s.mockCtrl = gomock.NewController(s.T())
	s.timeProvider = mocks.NewMockTimeProvider(s.mockCtrl)
	s.uuidGenerator = mockUUID.NewMockGenerator(s.mockCtrl)
	s.repo = &redisRepo{
		client:        s.mockClient,
		timeProvider:  s.timeProvider,
		uuidGenerator: s.uuidGenerator,
	}
}

func (s *RedisRepoTestSuite) TearDownTest() {
	s.mockCtrl.Finish()
	s.NoError(s.mock.ExpectationsWereMet())
}

func TestRedisRepoTestSuite(t *testing.T) {
	suite.Run(t, new(RedisRepoTestSuite))
}

func (s *RedisRepoTestSuite) TestGet_HappyPath() {
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

	s.mock.ExpectGet("session:test-id").SetVal(string(jsonData))

	session, err := s.repo.Get(ctx, sessionID)
	s.NoError(err)
	s.Equal(sessionID, session.ID)
}

func (s *RedisRepoTestSuite) TestGet_InputValidation() {
	ctx := context.Background()

	_, err := s.repo.Get(ctx, "")
	s.Error(err)
	s.True(errors.Is(err, internal.ErrMissingParam))
	s.EqualError(err, "sessions.Get missing parameter: id")
}

func (s *RedisRepoTestSuite) TestGet_DependencyError() {
	ctx := context.Background()
	sessionID := "test-id"

	s.mock.ExpectGet("session:test-id").SetErr(errors.New("redis error"))

	_, err := s.repo.Get(ctx, sessionID)
	s.Error(err)
	s.EqualError(err, "failed to get session from Redis: redis error")
}

func (s *RedisRepoTestSuite) TestGet_SessionNotFound() {
	ctx := context.Background()
	sessionID := "test-id"

	s.mock.ExpectGet("session:test-id").RedisNil()

	_, err := s.repo.Get(ctx, sessionID)
	s.Error(err)
	s.True(errors.Is(err, internal.ErrNotFound))
	s.Equal("sessions.Get not found: record error: test-id", err.Error())
}

func (s *RedisRepoTestSuite) TestSet_HappyPath() {
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

	s.mock.ExpectSet("session:test-id", string(expectedData), 0).SetVal("OK")
	s.mock.ExpectSAdd("user:user-id:sessions", "test-id").SetVal(1)

	err = s.repo.Set(ctx, session)
	s.NoError(err)
}

func (s *RedisRepoTestSuite) TestSet_DependencyError() {
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

	s.mock.ExpectSet("session:test-id", string(expectedData), 0).SetErr(errors.New("redis error"))

	err = s.repo.Set(ctx, session)
	s.Error(err)
}

func (s *RedisRepoTestSuite) TestSet_InputValidation() {
	ctx := context.Background()

	err := s.repo.Set(ctx, nil)
	s.Error(err)
	s.True(errors.Is(err, internal.ErrMissingParam))
	s.EqualError(err, "sessions.Set missing parameter: session")

	err = s.repo.Set(ctx, &entities.Session{})
	s.Error(err)
	s.True(errors.Is(err, internal.ErrMissingParam))
	s.EqualError(err, "sessions.Set missing parameter: session.ID")

	err = s.repo.Set(ctx, &entities.Session{ID: "test-id"})
	s.Error(err)
	s.True(errors.Is(err, internal.ErrMissingParam))
	s.EqualError(err, "sessions.Set missing parameter: session.UserID")
}

func (s *RedisRepoTestSuite) TestCreate_HappyPath() {
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Millisecond)
	s.timeProvider.EXPECT().Now().Return(now)

	expectedUUID := "test-id"
	s.uuidGenerator.EXPECT().New().Return(expectedUUID)

	session := &entities.Session{
		UserID:    "user-id",
		DraftID:   "draft-id",
		LastToken: "last-token",
	}

	expectedData := Data{
		ID:        expectedUUID,
		UserID:    session.UserID,
		DraftID:   session.DraftID,
		LastToken: session.LastToken,
		CreatedAt: now,
		UpdatedAt: now,
	}
	jsonData, err := json.Marshal(expectedData)
	s.Require().NoError(err)

	s.mock.ExpectSet("session:test-id", string(jsonData), 0).SetVal("OK")
	s.mock.ExpectSAdd("user:user-id:sessions", expectedUUID).SetVal(1)

	err = s.repo.Create(ctx, session)
	s.NoError(err)
	s.Equal(expectedUUID, session.ID)
}

func (s *RedisRepoTestSuite) TestCreate_DependencyError() {
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Millisecond)
	s.timeProvider.EXPECT().Now().Return(now)
	s.uuidGenerator.EXPECT().New().Return("test-id")

	session := &entities.Session{
		UserID:    "user-id",
		DraftID:   "draft-id",
		LastToken: "last-token",
	}

	expectedData := Data{
		ID:        "test-id",
		UserID:    session.UserID,
		DraftID:   session.DraftID,
		LastToken: session.LastToken,
		CreatedAt: now,
		UpdatedAt: now,
	}
	jsonData, err := json.Marshal(expectedData)
	s.Require().NoError(err)

	s.mock.ExpectSet("session:test-id", string(jsonData), 0).SetErr(errors.New("redis error"))

	err = s.repo.Create(ctx, session)
	s.Error(err)
}

func (s *RedisRepoTestSuite) TestCreate_InputValidation() {
	ctx := context.Background()

	err := s.repo.Create(ctx, nil)
	s.Error(err)
	s.EqualError(err, "session cannot be nil")

	err = s.repo.Create(ctx, &entities.Session{ID: "test-id"})
	s.Error(err)
	s.True(errors.Is(err, internal.ErrInvalidParam))
	s.EqualError(err, "sessions.Create invalid parameter: ID cannot be set")

	err = s.repo.Create(ctx, &entities.Session{})
	s.Error(err)
	s.True(errors.Is(err, internal.ErrMissingParam))
	s.EqualError(err, "sessions.Create missing parameter: UserID")
}

func (s *RedisRepoTestSuite) TestUpdate_HappyPath() {
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

	s.mock.ExpectGet("session:test-id").SetVal(string(jsonData))

	s.mock.ExpectSet("session:test-id", string(jsonData), 0).SetVal("OK")
	s.mock.ExpectSAdd("user:user-id:sessions", "test-id").SetVal(1)

	actual, err := s.repo.Update(ctx, session)

	s.Equal(expectedData.ID, actual.ID)
	s.NoError(err)
}

func (s *RedisRepoTestSuite) TestUpdate_DependencyError() {
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

	s.mock.ExpectGet("session:test-id").SetVal(string(jsonData))
	s.mock.ExpectSet("session:test-id", string(jsonData), 0).SetErr(errors.New("redis error"))

	_, err = s.repo.Update(ctx, session)
	s.Error(err)
}

func (s *RedisRepoTestSuite) TestUpdate_SessionNotFound() {
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Millisecond)

	session := &entities.Session{
		ID:        "test-id",
		UserID:    "user-id",
		DraftID:   "draft-id",
		LastToken: "last-token",
		CreatedAt: now.Add(-1 * time.Hour), // Assume created an hour ago
	}

	s.mock.ExpectGet("session:test-id").RedisNil()

	_, err := s.repo.Update(ctx, session)
	s.Error(err)

	s.True(errors.Is(err, internal.ErrNotFound))

	s.EqualError(err, "sessions.Get not found: record error: test-id")
}

func (s *RedisRepoTestSuite) TestUpdate_InputValidation() {
	ctx := context.Background()

	_, err := s.repo.Update(ctx, nil)
	s.Error(err)
	s.True(errors.Is(err, internal.ErrMissingParam))
	s.EqualError(err, "sessions.Update missing parameter: session")

	_, err = s.repo.Update(ctx, &entities.Session{})
	s.Error(err)
	s.True(errors.Is(err, internal.ErrMissingParam))
	s.EqualError(err, "sessions.Update missing parameter: ID")

	_, err = s.repo.Update(ctx, &entities.Session{ID: "test-id"})
	s.Error(err)
	s.True(errors.Is(err, internal.ErrMissingParam))
	s.EqualError(err, "sessions.Update missing parameter: UserID")
}

func (s *RedisRepoTestSuite) TestDelete_HappyPath() {
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

	s.mock.ExpectGet("session:test-id").SetVal(string(jsonData))
	s.mock.ExpectDel("session:test-id").SetVal(1)
	s.mock.ExpectSRem("user:user-id:sessions", "test-id").SetVal(1)

	actualErr := s.repo.Delete(ctx, sessionID)
	s.NoError(actualErr)
}

func (s *RedisRepoTestSuite) TestDelete_DependencyError() {
	ctx := context.Background()
	sessionID := "test-id"

	s.mock.ExpectGet("session:test-id").SetErr(errors.New("redis error"))

	actualErr := s.repo.Delete(ctx, sessionID)
	s.Error(actualErr)
}

func (s *RedisRepoTestSuite) TestDelete_InputValidation() {
	ctx := context.Background()

	err := s.repo.Delete(ctx, "")
	s.Error(err)
	s.True(errors.Is(err, internal.ErrMissingParam))
	s.EqualError(err, "sessions.Delete missing parameter: id")
}

func (s *RedisRepoTestSuite) TestListByUser_HappyPath() {
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

	s.mock.ExpectSMembers("user:user-id:sessions").SetVal(sessionIDs)
	s.mock.ExpectGet("session:session-2").SetVal(string(jsonData2))
	s.mock.ExpectGet("session:session-1").SetVal(string(jsonData1))

	sessions, err := s.repo.ListByUser(ctx, userID)
	s.NoError(err)
	s.Len(sessions, 2)
	s.Equal("session-1", sessions[0].ID)
	s.Equal("session-2", sessions[1].ID)
}

func (s *RedisRepoTestSuite) TestListByUser_DependencyError() {
	ctx := context.Background()
	userID := "user-id"

	s.mock.ExpectSMembers("user:user-id:sessions").SetErr(errors.New("redis error"))

	_, err := s.repo.ListByUser(ctx, userID)
	s.Error(err)
}

func (s *RedisRepoTestSuite) TestListByUser_InputValidation() {
	ctx := context.Background()

	_, err := s.repo.ListByUser(ctx, "")
	s.Error(err)
	s.True(errors.Is(err, internal.ErrMissingParam))
	s.EqualError(err, "sessions.ListByUser missing parameter: userID")
}

func (s *RedisRepoTestSuite) TestListByUser_UserNotFound() {
	ctx := context.Background()
	userID := "user-id"

	s.mock.ExpectSMembers("user:user-id:sessions").RedisNil()

	sessions, err := s.repo.ListByUser(ctx, userID)
	s.NoError(err)

	s.Len(sessions, 0)
}