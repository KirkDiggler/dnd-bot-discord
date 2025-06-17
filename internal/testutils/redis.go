package testutils

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

// TestRedisConfig holds configuration for test Redis instances
type TestRedisConfig struct {
	Addr     string
	Password string
	DB       int
}

// DefaultTestRedisConfig returns the default test Redis configuration
func DefaultTestRedisConfig() *TestRedisConfig {
	return &TestRedisConfig{
		Addr:     "localhost:6379",
		Password: "",
		DB:       15, // Use DB 15 for tests to avoid conflicts
	}
}

// CreateTestRedisClient creates a Redis client for testing
func CreateTestRedisClient(t *testing.T, cfg *TestRedisConfig) redis.UniversalClient {
	if cfg == nil {
		cfg = DefaultTestRedisConfig()
	}

	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := client.Ping(ctx).Err()
	if err != nil {
		t.Skipf("Redis not available for testing: %v", err)
	}

	// Clear the test database
	err = client.FlushDB(ctx).Err()
	require.NoError(t, err, "Failed to flush test Redis database")

	// Register cleanup
	t.Cleanup(func() {
		// Clear the database after test
		_ = client.FlushDB(context.Background()).Err()
		_ = client.Close()
	})

	return client
}

// CreateTestRedisClientOrSkip creates a Redis client or skips the test if Redis is not available
func CreateTestRedisClientOrSkip(t *testing.T) redis.UniversalClient {
	t.Helper()
	return CreateTestRedisClient(t, nil)
}

// WaitForRedis waits for Redis to be ready or times out
func WaitForRedis(addr string, timeout time.Duration) error {
	client := redis.NewClient(&redis.Options{
		Addr: addr,
		DB:   15,
	})
	defer client.Close()

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		err := client.Ping(ctx).Err()
		cancel()

		if err == nil {
			return nil
		}

		time.Sleep(100 * time.Millisecond)
	}

	return fmt.Errorf("redis not ready after %v", timeout)
}
