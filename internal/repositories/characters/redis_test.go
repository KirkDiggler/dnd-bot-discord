package characters

import (
	"context"
	"encoding/json"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/rulebook"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/shared"
	"testing"
	"time"

	dnderr "github.com/KirkDiggler/dnd-bot-discord/internal/errors"
	mockredis "github.com/KirkDiggler/dnd-bot-discord/internal/mocks/redis"
	mockUUID "github.com/KirkDiggler/dnd-bot-discord/internal/uuid/mocks"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

type RedisMockTestSuite struct {
	suite.Suite
	mockClient    *mockredis.MockUniversalClient
	repo          *redisRepo
	mockCtrl      *gomock.Controller
	uuidGenerator *mockUUID.MockGenerator
}

func (s *RedisMockTestSuite) SetupTest() {
	s.mockCtrl = gomock.NewController(s.T())
	s.mockClient = mockredis.NewMockUniversalClient(s.mockCtrl)
	s.uuidGenerator = mockUUID.NewMockGenerator(s.mockCtrl)
	s.repo = &redisRepo{
		client:        s.mockClient,
		uuidGenerator: s.uuidGenerator,
		ttl:           24 * time.Hour,
	}
}

func (s *RedisMockTestSuite) TearDownTest() {
	s.mockCtrl.Finish()
}

func TestRedisMockTestSuite(t *testing.T) {
	suite.Run(t, new(RedisMockTestSuite))
}

// Helper function to create test character
func (s *RedisMockTestSuite) createTestCharacter() *character.Character {
	return &character.Character{
		ID:               "test-id",
		OwnerID:          "owner-id",
		RealmID:          "realm-id",
		Name:             "Test Character",
		Level:            1,
		Race:             &rulebook.Race{Key: "human", Name: "Human"},
		Class:            &rulebook.Class{Key: "fighter", Name: "Fighter"},
		HitDie:           10,
		MaxHitPoints:     10,
		CurrentHitPoints: 10,
		Attributes: map[shared.Attribute]*character.AbilityScore{
			shared.AttributeStrength:     {Score: 16, Bonus: 3},
			shared.AttributeDexterity:    {Score: 14, Bonus: 2},
			shared.AttributeConstitution: {Score: 15, Bonus: 2},
			shared.AttributeIntelligence: {Score: 10, Bonus: 0},
			shared.AttributeWisdom:       {Score: 12, Bonus: 1},
			shared.AttributeCharisma:     {Score: 8, Bonus: -1},
		},
		Status: shared.CharacterStatusActive,
	}
}

// Test Create with pipeline support
func (s *RedisMockTestSuite) TestCreate_HappyPath() {
	ctx := context.Background()
	character := s.createTestCharacter()

	// We're using the generated mock, so isTestMode will return false (not a *redis.Client)

	// Expect existence check
	existsCmd := redis.NewIntCmd(ctx, "exists", "character:test-id")
	existsCmd.SetVal(0)
	s.mockClient.EXPECT().Exists(ctx, "character:test-id").Return(existsCmd)

	// Create and expect pipeline
	pipeline := mockredis.NewMockPipeliner(s.mockCtrl)
	s.mockClient.EXPECT().Pipeline().Return(pipeline)

	// Set up pipeline expectations
	// The pipeline will receive: 1 SET + 3 SADD commands
	setCmd := redis.NewStatusCmd(ctx, "set")
	setCmd.SetVal("OK")
	pipeline.EXPECT().Set(ctx, "character:test-id", gomock.Any(), time.Duration(0)).Return(setCmd)

	sAddCmd1 := redis.NewIntCmd(ctx, "sadd")
	sAddCmd1.SetVal(1)
	pipeline.EXPECT().SAdd(ctx, "owner:owner-id:characters", "test-id").Return(sAddCmd1)

	sAddCmd2 := redis.NewIntCmd(ctx, "sadd")
	sAddCmd2.SetVal(1)
	pipeline.EXPECT().SAdd(ctx, "realm:realm-id:characters", "test-id").Return(sAddCmd2)

	sAddCmd3 := redis.NewIntCmd(ctx, "sadd")
	sAddCmd3.SetVal(1)
	pipeline.EXPECT().SAdd(ctx, "owner:owner-id:realm:realm-id:characters", "test-id").Return(sAddCmd3)

	// Expect pipeline exec
	cmds := []redis.Cmder{setCmd, sAddCmd1, sAddCmd2, sAddCmd3}
	pipeline.EXPECT().Exec(ctx).Return(cmds, nil)

	err := s.repo.Create(ctx, character)
	s.NoError(err)
}

// Test Create - already exists
func (s *RedisMockTestSuite) TestCreate_AlreadyExists() {
	ctx := context.Background()
	character := s.createTestCharacter()

	// Expect existence check returns 1 (exists)
	existsCmd := redis.NewIntCmd(ctx, "exists", "character:test-id")
	existsCmd.SetVal(1)
	s.mockClient.EXPECT().Exists(ctx, "character:test-id").Return(existsCmd)

	err := s.repo.Create(ctx, character)
	s.Error(err)

	var alreadyExists *dnderr.Error
	s.ErrorAs(err, &alreadyExists)
	s.Equal(dnderr.CodeAlreadyExists, alreadyExists.Code)
}

// Test Get
func (s *RedisMockTestSuite) TestGet_HappyPath() {
	ctx := context.Background()
	character := s.createTestCharacter()
	data, err := s.repo.toCharacterData(character)
	s.Require().NoError(err)
	data.CreatedAt = time.Now().UTC()
	data.UpdatedAt = data.CreatedAt

	jsonData, err := json.Marshal(data)
	s.Require().NoError(err)

	getCmd := redis.NewStringCmd(ctx, "get", "character:test-id")
	getCmd.SetVal(string(jsonData))
	s.mockClient.EXPECT().Get(ctx, "character:test-id").Return(getCmd)

	result, err := s.repo.Get(ctx, "test-id")
	s.NoError(err)
	s.Equal(character.ID, result.ID)
	s.Equal(character.Name, result.Name)
}

// Test Update with owner/realm change using pipeline
func (s *RedisMockTestSuite) TestUpdate_WithOwnerRealmChange() {
	ctx := context.Background()
	existingChar := s.createTestCharacter()
	existingData, err := s.repo.toCharacterData(existingChar)
	s.Require().NoError(err)
	existingData.CreatedAt = time.Now().Add(-24 * time.Hour).UTC()
	existingData.UpdatedAt = existingData.CreatedAt

	jsonData, err := json.Marshal(existingData)
	s.Require().NoError(err)

	// Updated character with new owner
	updatedChar := existingChar.Clone()
	updatedChar.OwnerID = "new-owner-id"

	// We're using the generated mock, so isTestMode will return false (not a *redis.Client)

	// Expect get existing
	getCmd := redis.NewStringCmd(ctx, "get", "character:test-id")
	getCmd.SetVal(string(jsonData))
	s.mockClient.EXPECT().Get(ctx, "character:test-id").Return(getCmd)

	// Expect set updated
	setCmd := redis.NewStatusCmd(ctx, "set", "character:test-id", gomock.Any(), time.Duration(0))
	setCmd.SetVal("OK")
	s.mockClient.EXPECT().Set(ctx, "character:test-id", gomock.Any(), time.Duration(0)).Return(setCmd)

	// Create and expect pipeline for index updates
	pipeline := mockredis.NewMockPipeliner(s.mockCtrl)
	s.mockClient.EXPECT().Pipeline().Return(pipeline)

	// Expect 3 SREM commands for removing from old indexes
	sRemCmd1 := redis.NewIntCmd(ctx, "srem")
	sRemCmd1.SetVal(1)
	pipeline.EXPECT().SRem(ctx, "owner:owner-id:characters", "test-id").Return(sRemCmd1)

	sRemCmd2 := redis.NewIntCmd(ctx, "srem")
	sRemCmd2.SetVal(1)
	pipeline.EXPECT().SRem(ctx, "realm:realm-id:characters", "test-id").Return(sRemCmd2)

	sRemCmd3 := redis.NewIntCmd(ctx, "srem")
	sRemCmd3.SetVal(1)
	pipeline.EXPECT().SRem(ctx, "owner:owner-id:realm:realm-id:characters", "test-id").Return(sRemCmd3)

	// Expect 3 SADD commands for adding to new indexes
	sAddCmd1 := redis.NewIntCmd(ctx, "sadd")
	sAddCmd1.SetVal(1)
	pipeline.EXPECT().SAdd(ctx, "owner:new-owner-id:characters", "test-id").Return(sAddCmd1)

	sAddCmd2 := redis.NewIntCmd(ctx, "sadd")
	sAddCmd2.SetVal(1)
	pipeline.EXPECT().SAdd(ctx, "realm:realm-id:characters", "test-id").Return(sAddCmd2)

	sAddCmd3 := redis.NewIntCmd(ctx, "sadd")
	sAddCmd3.SetVal(1)
	pipeline.EXPECT().SAdd(ctx, "owner:new-owner-id:realm:realm-id:characters", "test-id").Return(sAddCmd3)

	// Expect pipeline exec
	cmds := []redis.Cmder{sRemCmd1, sRemCmd2, sRemCmd3, sAddCmd1, sAddCmd2, sAddCmd3}
	pipeline.EXPECT().Exec(ctx).Return(cmds, nil)

	err = s.repo.Update(ctx, updatedChar)
	s.NoError(err)
}

// Test Delete with pipeline
func (s *RedisMockTestSuite) TestDelete_HappyPath() {
	ctx := context.Background()
	character := s.createTestCharacter()
	data, err := s.repo.toCharacterData(character)
	s.Require().NoError(err)
	data.CreatedAt = time.Now().UTC()
	data.UpdatedAt = data.CreatedAt

	jsonData, err := json.Marshal(data)
	s.Require().NoError(err)

	// We're using the generated mock, so isTestMode will return false (not a *redis.Client)

	// Expect get to find owner/realm
	getCmd := redis.NewStringCmd(ctx, "get", "character:test-id")
	getCmd.SetVal(string(jsonData))
	s.mockClient.EXPECT().Get(ctx, "character:test-id").Return(getCmd)

	// Create and expect pipeline
	pipeline := mockredis.NewMockPipeliner(s.mockCtrl)
	s.mockClient.EXPECT().Pipeline().Return(pipeline)

	// Expect DEL command
	delCmd := redis.NewIntCmd(ctx, "del")
	delCmd.SetVal(1)
	pipeline.EXPECT().Del(ctx, "character:test-id").Return(delCmd)

	// Expect 3 SREM commands
	sRemCmd1 := redis.NewIntCmd(ctx, "srem")
	sRemCmd1.SetVal(1)
	pipeline.EXPECT().SRem(ctx, "owner:owner-id:characters", "test-id").Return(sRemCmd1)

	sRemCmd2 := redis.NewIntCmd(ctx, "srem")
	sRemCmd2.SetVal(1)
	pipeline.EXPECT().SRem(ctx, "realm:realm-id:characters", "test-id").Return(sRemCmd2)

	sRemCmd3 := redis.NewIntCmd(ctx, "srem")
	sRemCmd3.SetVal(1)
	pipeline.EXPECT().SRem(ctx, "owner:owner-id:realm:realm-id:characters", "test-id").Return(sRemCmd3)

	// Expect pipeline exec
	cmds := []redis.Cmder{delCmd, sRemCmd1, sRemCmd2, sRemCmd3}
	pipeline.EXPECT().Exec(ctx).Return(cmds, nil)

	err = s.repo.Delete(ctx, "test-id")
	s.NoError(err)
}

// Test GetByOwner
func (s *RedisMockTestSuite) TestGetByOwner_HappyPath() {
	ctx := context.Background()

	// Create test characters
	char1 := s.createTestCharacter()
	char1.ID = "char-1"
	char1.Name = "Character 1"
	data1, err := s.repo.toCharacterData(char1)
	s.Require().NoError(err)
	data1.CreatedAt = time.Now().UTC()
	data1.UpdatedAt = data1.CreatedAt

	char2 := s.createTestCharacter()
	char2.ID = "char-2"
	char2.Name = "Character 2"
	data2, err := s.repo.toCharacterData(char2)
	s.Require().NoError(err)
	data2.CreatedAt = time.Now().UTC()
	data2.UpdatedAt = data2.CreatedAt

	jsonData1, err := json.Marshal(data1)
	s.Require().NoError(err)
	jsonData2, err := json.Marshal(data2)
	s.Require().NoError(err)

	// Expect list IDs
	sMembersCmd := redis.NewStringSliceCmd(ctx, "smembers", "owner:owner-id:characters")
	sMembersCmd.SetVal([]string{"char-1", "char-2"})
	s.mockClient.EXPECT().SMembers(ctx, "owner:owner-id:characters").Return(sMembersCmd)

	// Expect get each character
	getCmd1 := redis.NewStringCmd(ctx, "get", "character:char-1")
	getCmd1.SetVal(string(jsonData1))
	s.mockClient.EXPECT().Get(ctx, "character:char-1").Return(getCmd1)

	getCmd2 := redis.NewStringCmd(ctx, "get", "character:char-2")
	getCmd2.SetVal(string(jsonData2))
	s.mockClient.EXPECT().Get(ctx, "character:char-2").Return(getCmd2)

	result, err := s.repo.GetByOwner(ctx, "owner-id")
	s.NoError(err)
	s.Len(result, 2)
}
