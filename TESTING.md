# Testing Guide

This document outlines the testing strategy and practices for the D&D Discord Bot.

## Testing Philosophy

We follow a comprehensive testing approach to ensure code stability and prevent regressions:

1. **Unit Tests** - Test individual components in isolation
2. **Integration Tests** - Test component interactions with real dependencies
3. **End-to-End Tests** - Test complete user workflows

## Running Tests

### Prerequisites

- Go 1.21+
- Redis (for integration tests)
- Make (optional, for convenience)

### Quick Start

```bash
# Run all unit tests
make test

# Run integration tests (requires Redis)
make test-integration

# Run all tests with coverage
make coverage

# Run specific test
go test ./internal/services/character -v -run TestCharacterCreationFlow
```

### Test Categories

#### Unit Tests
Fast, isolated tests that mock all dependencies:

```bash
go test ./... -short
```

#### Integration Tests
Tests that use real Redis and other dependencies:

```bash
# Start Redis first
docker run -d --name test-redis -p 6379:6379 redis:alpine

# Run integration tests
go test ./... -tags=integration

# Clean up
docker stop test-redis && docker rm test-redis
```

## Test Structure

### Test Files

- `*_test.go` - Standard unit tests
- `*_integration_test.go` - Integration tests (build tag: `integration`)
- `testutils/` - Shared test utilities and fixtures

### Test Utilities

#### Fixtures (`testutils/fixtures.go`)
Pre-configured test data:

```go
// Create a complete test character
char := testutils.CreateTestCharacter("id", "owner", "realm", "name")

// Create test race/class
race := testutils.CreateTestRace("human", "Human")
class := testutils.CreateTestClass("fighter", "Fighter", 10)
```

#### Redis Test Helper (`testutils/redis.go`)
Utilities for Redis-based tests:

```go
// Get test Redis client (skips test if Redis unavailable)
client := testutils.CreateTestRedisClientOrSkip(t)
```

## Writing Tests

### Unit Test Example

```go
func TestCharacter_IsComplete(t *testing.T) {
    tests := []struct {
        name     string
        setup    func() *entities.Character
        expected bool
    }{
        {
            name: "complete character",
            setup: func() *entities.Character {
                return testutils.CreateTestCharacter("1", "user", "realm", "Test")
            },
            expected: true,
        },
        {
            name: "missing race",
            setup: func() *entities.Character {
                char := testutils.CreateTestCharacter("2", "user", "realm", "Test")
                char.Race = nil
                return char
            },
            expected: false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            char := tt.setup()
            assert.Equal(t, tt.expected, char.IsComplete())
        })
    }
}
```

### Integration Test Example

```go
// +build integration

func TestRedisRepository_Integration(t *testing.T) {
    client := testutils.CreateTestRedisClientOrSkip(t)
    repo := characters.NewRedisRepository(&characters.RedisRepoConfig{
        Client: client,
    })
    
    t.Run("create and retrieve", func(t *testing.T) {
        char := testutils.CreateTestCharacter("1", "user", "realm", "Test")
        
        err := repo.Create(context.Background(), char)
        require.NoError(t, err)
        
        retrieved, err := repo.Get(context.Background(), char.ID)
        require.NoError(t, err)
        assert.Equal(t, char.Name, retrieved.Name)
    })
}
```

### Mock Usage

We use Uber's gomock for mocking:

```go
func TestService_WithMocks(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()
    
    mockRepo := mockcharrepo.NewMockRepository(ctrl)
    mockClient := mockdnd5e.NewMockClient(ctrl)
    
    // Set expectations
    mockRepo.EXPECT().
        Get(gomock.Any(), "char-id").
        Return(testChar, nil)
    
    // Test the service
    svc := character.NewService(&character.ServiceConfig{
        Repository: mockRepo,
        DNDClient:  mockClient,
    })
    
    result, err := svc.GetCharacter(context.Background(), "char-id")
    require.NoError(t, err)
    assert.NotNil(t, result)
}
```

## Test Coverage

Current coverage targets:
- Overall: 70%+
- Critical paths: 90%+
- New code: 80%+

Check coverage:
```bash
make coverage
# Opens coverage.html in browser
```

## Common Test Patterns

### Testing JSON Unmarshaling

```go
func TestChoice_UnmarshalJSON(t *testing.T) {
    json := `{"name": "Test", "type": "proficiency", "options": []}`
    
    var choice entities.Choice
    err := json.Unmarshal([]byte(json), &choice)
    
    require.NoError(t, err)
    assert.Equal(t, "Test", choice.Name)
}
```

### Testing Error Cases

```go
func TestService_ErrorHandling(t *testing.T) {
    // Test missing required fields
    _, err := svc.CreateCharacter(ctx, &CreateCharacterInput{})
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "required")
}
```

### Testing Concurrent Access

```go
func TestRepository_ConcurrentAccess(t *testing.T) {
    var wg sync.WaitGroup
    
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            
            char := testutils.CreateTestCharacter(
                fmt.Sprintf("char-%d", id),
                "user", "realm", "Test",
            )
            err := repo.Create(context.Background(), char)
            assert.NoError(t, err)
        }(i)
    }
    
    wg.Wait()
}
```

## Continuous Integration

Tests run automatically on:
- Every push to main/develop
- Every pull request

See `.github/workflows/ci.yml` for configuration.

## Troubleshooting

### Redis Connection Failed
```
Redis not available for testing: dial tcp 127.0.0.1:6379: connect: connection refused
```
**Solution**: Start Redis or tests will be skipped automatically

### Mock Generation Failed
```
mockgen: command not found
```
**Solution**: Install mockgen
```bash
go install go.uber.org/mock/mockgen@latest
```

### Test Timeout
```
panic: test timed out after 10m0s
```
**Solution**: Use shorter timeouts for unit tests
```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
```

## Best Practices

1. **Keep tests fast** - Unit tests should complete in milliseconds
2. **Use table-driven tests** - Easier to add test cases
3. **Test edge cases** - Empty values, nil pointers, concurrent access
4. **Mock external dependencies** - Don't call real APIs in unit tests
5. **Clean up after tests** - Use t.Cleanup() for deferred cleanup
6. **Use meaningful test names** - Describe what's being tested and expected outcome
7. **Avoid test interdependence** - Each test should be independent
8. **Use fixtures for complex data** - Centralize test data creation

## Future Improvements

- [ ] Add Discord bot integration tests
- [ ] Implement load testing for concurrent users
- [ ] Add mutation testing
- [ ] Create test data generators
- [ ] Add performance benchmarks
- [ ] Implement contract testing for D&D API