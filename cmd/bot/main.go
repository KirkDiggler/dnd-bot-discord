package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"

	"github.com/KirkDiggler/dnd-bot-discord/internal/clients/dnd5e"
	"github.com/KirkDiggler/dnd-bot-discord/internal/config"
	"github.com/KirkDiggler/dnd-bot-discord/internal/discord/v2/core"
	"github.com/KirkDiggler/dnd-bot-discord/internal/discord/v2/middleware"
	"github.com/KirkDiggler/dnd-bot-discord/internal/discord/v2/routers"
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

	// Setup v2 handlers for character creation
	v2Pipeline := setupV2Handlers(dg, serviceProvider)

	// Create hybrid handler that routes character creation to v2
	hybridHandler := createHybridHandler(handler, v2Pipeline)

	// Register the hybrid handler instead of the direct handler
	dg.AddHandler(hybridHandler)

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

// setupV2Handlers sets up the v2 handler system for character creation
func setupV2Handlers(dg *discordgo.Session, provider *services.Provider) *core.Pipeline {
	// Create pipeline
	pipeline := core.NewPipeline()

	// Apply global middleware
	pipeline.Use(
		middleware.DeferMiddleware(nil),            // Auto-defer interactions
		middleware.ErrorMiddleware(nil),            // Error handling
		middleware.LoggingMiddleware(nil),          // Request logging
		middleware.UserRateLimitMiddleware(60, 60), // 60 requests per minute per user
	)

	// Create routers (self-register with pipeline)
	// TODO: Implement character router for /dnd character list, show, etc.
	// _, err := routers.NewCharacterRouter(&routers.CharacterRouterConfig{
	// 	Pipeline: pipeline,
	// 	Provider: provider,
	// })
	// if err != nil {
	// 	panic("Failed to create character router: " + err.Error())
	// }

	// Create router - handles /dnd create character (and future create commands)
	_, err := routers.NewCreateRouter(&routers.CreateRouterConfig{
		Pipeline: pipeline,
		Provider: provider,
	})
	if err != nil {
		panic("Failed to create create router: " + err.Error())
	}

	return pipeline
}

// createHybridHandler creates a handler that routes character creation to v2
func createHybridHandler(mainHandler *discord.Handler, v2Pipeline *core.Pipeline) func(s *discordgo.Session, i *discordgo.InteractionCreate) {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		// Route character creation commands to v2
		if shouldUseV2(i) {
			log.Printf("[V2] Executing V2 pipeline...")
			if err := v2Pipeline.Execute(context.Background(), s, i); err != nil {
				log.Printf("[V2] Pipeline error: %v", err)
				// Don't fall back - let's see the actual error
			} else {
				log.Printf("[V2] Pipeline executed successfully")
			}
			return
		}

		// Use main handler for everything else
		mainHandler.HandleInteraction(s, i)
	}
}

// shouldUseV2 determines if an interaction should be handled by v2
func shouldUseV2(i *discordgo.InteractionCreate) bool {
	// Check if this is a character creation command
	if i.Type == discordgo.InteractionApplicationCommand {
		data := i.ApplicationCommandData()

		// Log for debugging
		log.Printf("[V2 Router] Command: %s", data.Name)
		if len(data.Options) > 0 {
			log.Printf("[V2 Router] First option: %s (type: %v)", data.Options[0].Name, data.Options[0].Type)
			if len(data.Options[0].Options) > 0 {
				log.Printf("[V2 Router] Sub option: %s (type: %v)", data.Options[0].Options[0].Name, data.Options[0].Options[0].Type)
			}
		}

		// Route /dnd create commands to v2
		if data.Name == "dnd" && len(data.Options) > 0 {
			if data.Options[0].Name == "create" {
				log.Printf("[V2 Router] Routing to V2: /dnd create")
				return true
			}
		}
	}

	// Check if this is a character creation component interaction
	if i.Type == discordgo.InteractionMessageComponent {
		customID := i.MessageComponentData().CustomID
		log.Printf("[V2 Router] Component interaction with customID: %s", customID)

		// Route create domain components to v2
		if strings.HasPrefix(customID, "create:") || strings.Contains(customID, ":select_") {
			log.Printf("[V2 Router] Routing component to V2: %s", customID)
			return true
		}
	}

	// Check if this is a modal submission
	if i.Type == discordgo.InteractionModalSubmit {
		customID := i.ModalSubmitData().CustomID
		log.Printf("[V2 Router] Modal submission with customID: %s", customID)

		// Route create domain modals to v2
		if strings.HasPrefix(customID, "create:") {
			log.Printf("[V2 Router] Routing modal to V2: %s", customID)
			return true
		}
	}

	return false
}
