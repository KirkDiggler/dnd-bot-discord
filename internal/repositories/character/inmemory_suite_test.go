package character_test

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/KirkDiggler/dnd-bot-discord/internal/entities"
	dnderr "github.com/KirkDiggler/dnd-bot-discord/internal/errors"
	"github.com/KirkDiggler/dnd-bot-discord/internal/repositories/character"
	"github.com/stretchr/testify/suite"
)

// InMemoryRepositoryTestSuite defines the test suite for in-memory repository
type InMemoryRepositoryTestSuite struct {
	suite.Suite
	repo character.Repository
	ctx  context.Context
}

// SetupTest runs before each test
func (s *InMemoryRepositoryTestSuite) SetupTest() {
	s.repo = character.NewInMemoryRepository()
	s.ctx = context.Background()
}

// Test suite runner
func TestInMemoryRepositorySuite(t *testing.T) {
	suite.Run(t, new(InMemoryRepositoryTestSuite))
}

// Create Tests

func (s *InMemoryRepositoryTestSuite) TestCreate_Success() {
	// Setup
	char := &entities.Character{
		ID:      "char_123",
		OwnerID: "user_456",
		RealmID: "realm_789",
		Name:    "Thorin",
	}

	// Execute
	err := s.repo.Create(s.ctx, char)

	// Assert
	s.NoError(err)
	
	// Verify by getting the character
	gotChar, err := s.repo.Get(s.ctx, "char_123")
	s.NoError(err)
	s.Equal("Thorin", gotChar.Name)
}

func (s *InMemoryRepositoryTestSuite) TestCreate_DuplicateID() {
	// Setup
	char := &entities.Character{
		ID:      "char_123",
		OwnerID: "user_456",
		RealmID: "realm_789",
		Name:    "Thorin",
	}

	// Create first time
	err := s.repo.Create(s.ctx, char)
	s.NoError(err)

	// Try to create again with same ID
	err = s.repo.Create(s.ctx, char)

	// Assert
	s.Error(err)
	s.True(dnderr.IsAlreadyExists(err))
}

func (s *InMemoryRepositoryTestSuite) TestCreate_IsolatesData() {
	// Setup
	char := &entities.Character{
		ID:      "char_123",
		OwnerID: "user_456",
		RealmID: "realm_789",
		Name:    "Thorin",
	}

	// Execute
	err := s.repo.Create(s.ctx, char)
	s.NoError(err)

	// Modify original
	char.Name = "Modified"

	// Get from repo
	gotChar, err := s.repo.Get(s.ctx, "char_123")

	// Assert - should not be affected by external modification
	s.NoError(err)
	s.Equal("Thorin", gotChar.Name)
}

// Get Tests

func (s *InMemoryRepositoryTestSuite) TestGet_Success() {
	// Setup
	char := &entities.Character{
		ID:      "char_123",
		OwnerID: "user_456",
		RealmID: "realm_789",
		Name:    "Thorin",
	}
	s.repo.Create(s.ctx, char)

	// Execute
	gotChar, err := s.repo.Get(s.ctx, "char_123")

	// Assert
	s.NoError(err)
	s.NotNil(gotChar)
	s.Equal("Thorin", gotChar.Name)
	s.Equal("user_456", gotChar.OwnerID)
	s.Equal("realm_789", gotChar.RealmID)
}

func (s *InMemoryRepositoryTestSuite) TestGet_NotFound() {
	// Execute
	gotChar, err := s.repo.Get(s.ctx, "nonexistent")

	// Assert
	s.Error(err)
	s.Nil(gotChar)
	s.True(dnderr.IsNotFound(err))
}

func (s *InMemoryRepositoryTestSuite) TestGet_ReturnsCopy() {
	// Setup
	char := &entities.Character{
		ID:      "char_123",
		OwnerID: "user_456",
		RealmID: "realm_789",
		Name:    "Thorin",
	}
	s.repo.Create(s.ctx, char)

	// Get character twice
	gotChar1, err := s.repo.Get(s.ctx, "char_123")
	s.NoError(err)
	
	gotChar2, err := s.repo.Get(s.ctx, "char_123")
	s.NoError(err)

	// Modify first copy
	gotChar1.Name = "Modified"

	// Assert - second copy should not be affected
	s.Equal("Thorin", gotChar2.Name)
}

// GetByOwner Tests

func (s *InMemoryRepositoryTestSuite) TestGetByOwner_Success() {
	// Setup - create multiple characters
	chars := []*entities.Character{
		{
			ID:      "char_1",
			OwnerID: "user_123",
			RealmID: "realm_456",
			Name:    "Thorin",
		},
		{
			ID:      "char_2",
			OwnerID: "user_123",
			RealmID: "realm_789",
			Name:    "Gandalf",
		},
		{
			ID:      "char_3",
			OwnerID: "user_999",
			RealmID: "realm_456",
			Name:    "Legolas",
		},
	}

	for _, char := range chars {
		s.repo.Create(s.ctx, char)
	}

	// Execute
	results, err := s.repo.GetByOwner(s.ctx, "user_123")

	// Assert
	s.NoError(err)
	s.Len(results, 2)
	
	// Check names
	names := []string{results[0].Name, results[1].Name}
	s.Contains(names, "Thorin")
	s.Contains(names, "Gandalf")
}

func (s *InMemoryRepositoryTestSuite) TestGetByOwner_Empty() {
	// Execute - no characters for this owner
	results, err := s.repo.GetByOwner(s.ctx, "user_000")

	// Assert
	s.NoError(err)
	s.Empty(results)
}

// GetByOwnerAndRealm Tests

func (s *InMemoryRepositoryTestSuite) TestGetByOwnerAndRealm_Success() {
	// Setup
	chars := []*entities.Character{
		{
			ID:      "char_1",
			OwnerID: "user_123",
			RealmID: "realm_456",
			Name:    "Thorin",
		},
		{
			ID:      "char_2",
			OwnerID: "user_123",
			RealmID: "realm_789",
			Name:    "Gandalf",
		},
		{
			ID:      "char_3",
			OwnerID: "user_123",
			RealmID: "realm_456",
			Name:    "Bilbo",
		},
	}

	for _, char := range chars {
		s.repo.Create(s.ctx, char)
	}

	// Execute
	results, err := s.repo.GetByOwnerAndRealm(s.ctx, "user_123", "realm_456")

	// Assert
	s.NoError(err)
	s.Len(results, 2)
	
	// Check names
	names := []string{results[0].Name, results[1].Name}
	s.Contains(names, "Thorin")
	s.Contains(names, "Bilbo")
}

// Update Tests

func (s *InMemoryRepositoryTestSuite) TestUpdate_Success() {
	// Setup
	char := &entities.Character{
		ID:      "char_123",
		OwnerID: "user_456",
		RealmID: "realm_789",
		Name:    "Thorin",
	}
	s.repo.Create(s.ctx, char)

	// Update character
	char.Name = "Thorin Oakenshield"
	err := s.repo.Update(s.ctx, char)

	// Assert
	s.NoError(err)
	
	// Verify update
	gotChar, err := s.repo.Get(s.ctx, "char_123")
	s.NoError(err)
	s.Equal("Thorin Oakenshield", gotChar.Name)
}

func (s *InMemoryRepositoryTestSuite) TestUpdate_NotFound() {
	// Setup
	char := &entities.Character{
		ID:      "nonexistent",
		OwnerID: "user_456",
		RealmID: "realm_789",
		Name:    "Thorin",
	}

	// Execute
	err := s.repo.Update(s.ctx, char)

	// Assert
	s.Error(err)
	s.True(dnderr.IsNotFound(err))
}

// Delete Tests

func (s *InMemoryRepositoryTestSuite) TestDelete_Success() {
	// Setup
	char := &entities.Character{
		ID:      "char_123",
		OwnerID: "user_456",
		RealmID: "realm_789",
		Name:    "Thorin",
	}
	s.repo.Create(s.ctx, char)

	// Execute
	err := s.repo.Delete(s.ctx, "char_123")

	// Assert
	s.NoError(err)
	
	// Verify deletion
	_, err = s.repo.Get(s.ctx, "char_123")
	s.Error(err)
	s.True(dnderr.IsNotFound(err))
}

func (s *InMemoryRepositoryTestSuite) TestDelete_NotFound() {
	// Execute
	err := s.repo.Delete(s.ctx, "nonexistent")

	// Assert
	s.Error(err)
	s.True(dnderr.IsNotFound(err))
}

// Concurrency Tests

func (s *InMemoryRepositoryTestSuite) TestConcurrentCreates() {
	// Setup
	var wg sync.WaitGroup
	numGoroutines := 10

	// Execute - create characters concurrently
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			char := &entities.Character{
				ID:      fmt.Sprintf("char_%d", id),
				OwnerID: "user_123",
				RealmID: "realm_456",
				Name:    fmt.Sprintf("Character %d", id),
			}
			err := s.repo.Create(s.ctx, char)
			s.NoError(err)
		}(i)
	}

	wg.Wait()

	// Assert - verify all characters were created
	results, err := s.repo.GetByOwner(s.ctx, "user_123")
	s.NoError(err)
	s.Len(results, numGoroutines)
}

func (s *InMemoryRepositoryTestSuite) TestConcurrentReadsAndWrites() {
	// Setup - create initial character
	char := &entities.Character{
		ID:      "char_123",
		OwnerID: "user_456",
		RealmID: "realm_789",
		Name:    "Thorin",
		Level:   1,
	}
	s.repo.Create(s.ctx, char)

	var wg sync.WaitGroup
	numReaders := 5
	numWriters := 3

	// Start readers
	for i := 0; i < numReaders; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				gotChar, err := s.repo.Get(s.ctx, "char_123")
				s.NoError(err)
				s.NotNil(gotChar)
			}
		}()
	}

	// Start writers
	for i := 0; i < numWriters; i++ {
		wg.Add(1)
		go func(writerID int) {
			defer wg.Done()
			for j := 0; j < 5; j++ {
				char := &entities.Character{
					ID:      "char_123",
					OwnerID: "user_456",
					RealmID: "realm_789",
					Name:    fmt.Sprintf("Thorin v%d", writerID*10+j),
					Level:   writerID*10 + j,
				}
				err := s.repo.Update(s.ctx, char)
				s.NoError(err)
			}
		}(i)
	}

	wg.Wait()

	// Assert - character should still exist and be readable
	finalChar, err := s.repo.Get(s.ctx, "char_123")
	s.NoError(err)
	s.NotNil(finalChar)
	s.Contains(finalChar.Name, "Thorin")
}