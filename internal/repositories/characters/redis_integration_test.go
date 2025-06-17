//go:build integration
// +build integration

package characters

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/KirkDiggler/dnd-bot-discord/internal/entities"
	"github.com/KirkDiggler/dnd-bot-discord/internal/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type RedisIntegrationTestSuite struct {
	suite.Suite
	redisContainer testcontainers.Container
	redisClient    *redis.Client
	repo           Repository
	ctx            context.Context
}

func (s *RedisIntegrationTestSuite) SetupSuite() {
	s.ctx = context.Background()

	// Start Redis container
	req := testcontainers.ContainerRequest{
		Image:        "redis:7-alpine",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForLog("Ready to accept connections"),
	}

	container, err := testcontainers.GenericContainer(s.ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	s.Require().NoError(err)
	s.redisContainer = container

	// Get Redis connection details
	host, err := container.Host(s.ctx)
	s.Require().NoError(err)

	port, err := container.MappedPort(s.ctx, "6379")
	s.Require().NoError(err)

	// Create Redis client
	s.redisClient = redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", host, port.Port()),
	})

	// Wait for Redis to be ready
	err = s.redisClient.Ping(s.ctx).Err()
	s.Require().NoError(err)

	// Create repository
	s.repo = NewRedisRepository(&RedisRepoConfig{
		Client:        s.redisClient,
		UUIDGenerator: uuid.NewGoogleUUIDGenerator(),
		DraftTTL:      24 * time.Hour,
	})
}

func (s *RedisIntegrationTestSuite) TearDownSuite() {
	if s.redisClient != nil {
		s.redisClient.Close()
	}
	if s.redisContainer != nil {
		s.redisContainer.Terminate(s.ctx)
	}
}

func (s *RedisIntegrationTestSuite) SetupTest() {
	// Clear Redis before each test
	err := s.redisClient.FlushDB(s.ctx).Err()
	s.Require().NoError(err)
}

func (s *RedisIntegrationTestSuite) createTestCharacter(id string) *entities.Character {
	return &entities.Character{
		ID:       id,
		OwnerID:  "owner-" + id,
		RealmID:  "realm-1",
		Name:     "Test Character " + id,
		Level:    1,
		Race:     &entities.Race{Key: "human", Name: "Human"},
		Class:    &entities.Class{Key: "fighter", Name: "Fighter"},
		HitDie:   10,
		MaxHitPoints: 10,
		CurrentHitPoints: 10,
		Attributes: map[entities.Attribute]*entities.AbilityScore{
			entities.AttributeStrength:     {Score: 16, Bonus: 3},
			entities.AttributeDexterity:    {Score: 14, Bonus: 2},
			entities.AttributeConstitution: {Score: 15, Bonus: 2},
			entities.AttributeIntelligence: {Score: 10, Bonus: 0},
			entities.AttributeWisdom:       {Score: 12, Bonus: 1},
			entities.AttributeCharisma:     {Score: 8, Bonus: -1},
		},
		Status: entities.CharacterStatusActive,
	}
}

func (s *RedisIntegrationTestSuite) TestCreateAndGet() {
	// Create character
	char := s.createTestCharacter("test-1")
	err := s.repo.Create(s.ctx, char)
	s.NoError(err)

	// Get character
	retrieved, err := s.repo.Get(s.ctx, "test-1")
	s.NoError(err)
	s.Equal(char.Name, retrieved.Name)
	s.Equal(char.OwnerID, retrieved.OwnerID)
}

func (s *RedisIntegrationTestSuite) TestGetByOwner() {
	// Create multiple characters for same owner
	owner := "owner-123"
	for i := 1; i <= 3; i++ {
		char := s.createTestCharacter(fmt.Sprintf("char-%d", i))
		char.OwnerID = owner
		err := s.repo.Create(s.ctx, char)
		s.NoError(err)
	}

	// Get by owner
	chars, err := s.repo.GetByOwner(s.ctx, owner)
	s.NoError(err)
	s.Len(chars, 3)
}

func (s *RedisIntegrationTestSuite) TestUpdate() {
	// Create character
	char := s.createTestCharacter("update-test")
	err := s.repo.Create(s.ctx, char)
	s.NoError(err)

	// Update character
	char.Name = "Updated Name"
	char.Level = 5
	err = s.repo.Update(s.ctx, char)
	s.NoError(err)

	// Verify update
	retrieved, err := s.repo.Get(s.ctx, "update-test")
	s.NoError(err)
	s.Equal("Updated Name", retrieved.Name)
	s.Equal(5, retrieved.Level)
}

func (s *RedisIntegrationTestSuite) TestDelete() {
	// Create character
	char := s.createTestCharacter("delete-test")
	err := s.repo.Create(s.ctx, char)
	s.NoError(err)

	// Delete character
	err = s.repo.Delete(s.ctx, "delete-test")
	s.NoError(err)

	// Verify deletion
	_, err = s.repo.Get(s.ctx, "delete-test")
	s.Error(err)
}

func (s *RedisIntegrationTestSuite) TestConcurrentOperations() {
	// Test concurrent creates
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(id int) {
			char := s.createTestCharacter(fmt.Sprintf("concurrent-%d", id))
			err := s.repo.Create(s.ctx, char)
			s.NoError(err)
			done <- true
		}(i)
	}

	// Wait for all creates
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify all were created
	for i := 0; i < 10; i++ {
		_, err := s.repo.Get(s.ctx, fmt.Sprintf("concurrent-%d", i))
		s.NoError(err)
	}
}

func (s *RedisIntegrationTestSuite) TestPipelineAtomicity() {
	// Create a character
	char := s.createTestCharacter("pipeline-test")
	err := s.repo.Create(s.ctx, char)
	s.NoError(err)

	// Verify all indexes were created atomically
	// Check character data exists
	exists := s.redisClient.Exists(s.ctx, "character:pipeline-test").Val()
	s.Equal(int64(1), exists)

	// Check all indexes
	isMember := s.redisClient.SIsMember(s.ctx, "owner:owner-pipeline-test:characters", "pipeline-test").Val()
	s.True(isMember)

	isMember = s.redisClient.SIsMember(s.ctx, "realm:realm-1:characters", "pipeline-test").Val()
	s.True(isMember)

	isMember = s.redisClient.SIsMember(s.ctx, "owner:owner-pipeline-test:realm:realm-1:characters", "pipeline-test").Val()
	s.True(isMember)
}

func TestRedisIntegrationTestSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	suite.Run(t, new(RedisIntegrationTestSuite))
}