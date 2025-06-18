package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"

	"github.com/KirkDiggler/dnd-bot-discord/internal/clients/dnd5e"
	"github.com/KirkDiggler/dnd-bot-discord/internal/config"
	"github.com/KirkDiggler/dnd-bot-discord/internal/handlers/discord"
	"github.com/KirkDiggler/dnd-bot-discord/internal/repositories/characters"
	"github.com/KirkDiggler/dnd-bot-discord/internal/repositories/gamesessions"
	"github.com/KirkDiggler/dnd-bot-discord/internal/services"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	} else {
		log.Println("Loaded .env file")
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	log.Printf("Bot Token: %s...%s", cfg.Discord.Token[:8], cfg.Discord.Token[len(cfg.Discord.Token)-4:])
	log.Printf("Application ID: %s", cfg.Discord.AppID)
	if cfg.Discord.GuildID != "" {
		log.Printf("Guild ID: %s", cfg.Discord.GuildID)
	}

	// Create Discord session
	dg, err := discordgo.New("Bot " + cfg.Discord.Token)
	if err != nil {
		log.Fatalf("Failed to create Discord session: %v", err)
	}

	// Create D&D 5e API client
	dndClient, err := dnd5e.New(&dnd5e.Config{
		HttpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	})
	if err != nil {
		log.Fatalf("Failed to create D&D 5e client: %v", err)
	}

	// Create service provider config
	providerConfig := &services.ProviderConfig{
		DNDClient: dndClient,
	}

	// Keep Redis client for cleanup
	var redisClient *redis.Client

	// Try to connect to Redis if URL is provided
	if redisURL := os.Getenv("REDIS_URL"); redisURL != "" {
		log.Printf("Connecting to Redis at: %s", redisURL)

		opts, parseErr := redis.ParseURL(redisURL)
		if parseErr != nil {
			log.Printf("Failed to parse Redis URL: %v", parseErr)
			log.Println("Falling back to in-memory repositories")
		} else {
			redisClient = redis.NewClient(opts)

			// Test connection
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

			if pingErr := redisClient.Ping(ctx).Err(); pingErr != nil {
				cancel()
				log.Printf("Failed to connect to Redis: %v", pingErr)
				log.Println("Falling back to in-memory repositories")
			} else {
				defer cancel()
				log.Println("Successfully connected to Redis")

				// Create Redis repositories using bounded context constructors
				providerConfig.CharacterRepository = characters.NewRedis(redisClient)
				providerConfig.SessionRepository = gamesessions.NewRedis(redisClient)

				log.Println("Using Redis for persistence")
			}
		}
	} else {
		log.Println("No REDIS_URL found, using in-memory repositories")
	}

	// Create service provider
	serviceProvider := services.NewProvider(providerConfig)

	// Create Discord handler
	handler := discord.NewHandler(&discord.HandlerConfig{
		ServiceProvider: serviceProvider,
	})

	// Register interaction handler
	dg.AddHandler(handler.HandleInteraction)

	// Open connection to Discord
	err = dg.Open()
	if err != nil {
		log.Printf("Failed to open Discord connection: %v", err)
		return
	}
	defer func() {
		clientErr := dg.Close()
		if clientErr != nil {
			log.Printf("Failed to close Discord connection: %v", clientErr)
		}
	}()

	// Register commands
	// Use empty string for global commands, or set a specific guild ID for testing
	if err := handler.RegisterCommands(dg, cfg.Discord.GuildID); err != nil {
		log.Printf("Failed to register commands: %v", err)
		return
	}

	if cfg.Discord.GuildID != "" {
		log.Printf("Registered commands for guild: %s", cfg.Discord.GuildID)
	} else {
		log.Println("Registered global commands (may take up to 1 hour to propagate)")
	}

	fmt.Println("Bot is now running. Press CTRL-C to exit.")

	// Wait for interrupt signal
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	fmt.Println("Shutting down...")

	// Clean up Redis connection if we have one
	if redisClient != nil {
		if err := redisClient.Close(); err != nil {
			log.Printf("Error closing Redis connection: %v", err)
		} else {
			log.Println("Closed Redis connection")
		}
	}
}
