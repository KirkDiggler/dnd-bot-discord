package services

import (
	"github.com/KirkDiggler/dnd-bot-discord/internal/clients/dnd5e"
	"github.com/KirkDiggler/dnd-bot-discord/internal/repositories/characters"
	"github.com/KirkDiggler/dnd-bot-discord/internal/repositories/encounters"
	"github.com/KirkDiggler/dnd-bot-discord/internal/repositories/gamesessions"
	characterService "github.com/KirkDiggler/dnd-bot-discord/internal/services/character"
	encounterService "github.com/KirkDiggler/dnd-bot-discord/internal/services/encounter"
	sessionService "github.com/KirkDiggler/dnd-bot-discord/internal/services/session"
)

// Provider holds all service instances
type Provider struct {
	CharacterService characterService.Service
	SessionService   sessionService.Service
	EncounterService encounterService.Service
}

// ProviderConfig holds configuration for creating services
type ProviderConfig struct {
	DNDClient            dnd5e.Client
	CharacterRepository  characters.Repository
	SessionRepository    gamesessions.Repository
	EncounterRepository  encounters.Repository
}

// NewProvider creates a new service provider with all services initialized
func NewProvider(cfg *ProviderConfig) *Provider {
	// Use in-memory repository if none provided
	charRepo := cfg.CharacterRepository
	if charRepo == nil {
		charRepo = characters.NewInMemoryRepository()
	}
	
	sessionRepo := cfg.SessionRepository
	if sessionRepo == nil {
		sessionRepo = gamesessions.NewInMemoryRepository()
	}
	
	encounterRepo := cfg.EncounterRepository
	if encounterRepo == nil {
		encounterRepo = encounters.NewInMemoryRepository()
	}

	// Create character service
	charService := characterService.NewService(&characterService.ServiceConfig{
		DNDClient:  cfg.DNDClient,
		Repository: charRepo,
	})
	
	// Create session service
	sessService := sessionService.NewService(&sessionService.ServiceConfig{
		Repository:       sessionRepo,
		CharacterService: charService,
	})
	
	// Create encounter service
	encService := encounterService.NewService(&encounterService.ServiceConfig{
		Repository:       encounterRepo,
		SessionService:   sessService,
		CharacterService: charService,
	})

	return &Provider{
		CharacterService: charService,
		SessionService:   sessService,
		EncounterService: encService,
	}
}