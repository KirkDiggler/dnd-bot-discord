package v2

import (
	"context"
	"log"

	"github.com/KirkDiggler/dnd-bot-discord/internal/discord/v2/core"
	"github.com/KirkDiggler/dnd-bot-discord/internal/discord/v2/middleware"
	"github.com/KirkDiggler/dnd-bot-discord/internal/discord/v2/routers"
	"github.com/KirkDiggler/dnd-bot-discord/internal/services"
	"github.com/bwmarrin/discordgo"
)

// SetupV2Handlers sets up the v2 handler system
func SetupV2Handlers(dg *discordgo.Session, provider *services.Provider) *core.Pipeline {
	// Create pipeline
	pipeline := core.NewPipeline()

	// Apply global middleware
	pipeline.Use(
		middleware.DeferMiddleware(nil),   // Auto-defer interactions
		middleware.ErrorMiddleware(nil),   // Error handling
		middleware.LoggingMiddleware(nil), // Request logging
		// middleware.GuildRequiredMiddleware(),          // TODO: Implement when needed
		middleware.UserRateLimitMiddleware(60, 60), // 60 requests per minute per user
	)

	// Create routers
	characterRouter := routers.NewCharacterRouter(pipeline, provider)
	_ = characterRouter // Router self-registers

	// Add more routers as they're implemented
	// combatRouter := routers.NewCombatRouter(pipeline, provider)
	// dungeonRouter := routers.NewDungeonRouter(pipeline, provider)

	return pipeline
}

// IntegrateWithMainHandler integrates v2 with the existing handler
// This allows gradual migration of commands
func IntegrateWithMainHandler(dg *discordgo.Session, provider *services.Provider) {
	// Create v2 pipeline
	v2Pipeline := SetupV2Handlers(dg, provider)

	// Create a hybrid handler that routes to v2 for character creation
	hybridHandler := func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		// Check if this is a character creation command
		if i.Type == discordgo.InteractionApplicationCommand {
			data := i.ApplicationCommandData()

			// Route character create to v2
			if data.Name == "dnd" && len(data.Options) > 0 {
				if data.Options[0].Name == "character" && len(data.Options[0].Options) > 0 {
					if data.Options[0].Options[0].Name == "create" {
						// Use v2 handler
						if err := v2Pipeline.Execute(context.Background(), s, i); err != nil {
							log.Printf("V2 handler error: %v", err)
							// Fallback could go here
						}
						return
					}
				}
			}
		}

		// Check if this is a character creation component interaction
		if i.Type == discordgo.InteractionMessageComponent {
			customID := i.MessageComponentData().CustomID

			// Route creation components to v2
			if len(customID) > 9 && customID[:9] == "creation:" {
				// Use v2 handler
				if err := v2Pipeline.Execute(context.Background(), s, i); err != nil {
					log.Printf("V2 handler error: %v", err)
				}
				return
			}
		}

		// Otherwise, let the main handler process it
		// (This would be the existing handler logic)
	}

	// Register the hybrid handler
	dg.AddHandler(hybridHandler)
}

// Example of how to use in main.go:
/*
func main() {
	// ... existing setup ...

	// Option 1: Use v2 for character creation only
	v2.IntegrateWithMainHandler(dg, serviceProvider)

	// Option 2: Use v2 for everything (when ready)
	// pipeline := v2.SetupV2Handlers(dg, serviceProvider)
	// dg.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
	//     if err := pipeline.Execute(context.Background(), s, i); err != nil {
	//         log.Printf("Handler error: %v", err)
	//     }
	// })

	// ... rest of main ...
}
*/
