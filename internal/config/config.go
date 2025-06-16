package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds all configuration for the application
type Config struct {
	Discord DiscordConfig
	Redis   RedisConfig
	DND5E   DND5EConfig
}

// DiscordConfig holds Discord-specific configuration
type DiscordConfig struct {
	Token   string
	AppID   string
	GuildID string // Optional: for guild-specific commands
}

// RedisConfig holds Redis-specific configuration
type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

// DND5EConfig holds D&D 5e API configuration
type DND5EConfig struct {
	BaseURL string
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{
		Discord: DiscordConfig{
			Token:   os.Getenv("DISCORD_TOKEN"),
			AppID:   os.Getenv("DISCORD_APP_ID"),
			GuildID: os.Getenv("DISCORD_GUILD_ID"),
		},
		Redis: RedisConfig{
			Addr:     getEnvOrDefault("REDIS_ADDR", "localhost:6379"),
			Password: os.Getenv("REDIS_PASSWORD"),
			DB:       getEnvAsIntOrDefault("REDIS_DB", 0),
		},
		DND5E: DND5EConfig{
			BaseURL: getEnvOrDefault("DND5E_API_URL", "https://www.dnd5eapi.co/api"),
		},
	}

	// Validate required fields
	if cfg.Discord.Token == "" {
		return nil, fmt.Errorf("DISCORD_TOKEN is required")
	}
	if cfg.Discord.AppID == "" {
		return nil, fmt.Errorf("DISCORD_APP_ID is required")
	}

	return cfg, nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}