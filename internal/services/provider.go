package services

import (
	"github.com/KirkDiggler/dnd-bot-discord/internal/clients/dnd5e"
	"github.com/KirkDiggler/dnd-bot-discord/internal/repositories/characters"
	characterService "github.com/KirkDiggler/dnd-bot-discord/internal/services/character"
)

// Provider holds all service instances
type Provider struct {
	CharacterService characterService.Service
}

// ProviderConfig holds configuration for creating services
type ProviderConfig struct {
	DNDClient           dnd5e.Client
	CharacterRepository characters.Repository
}

// NewProvider creates a new service provider with all services initialized
func NewProvider(cfg *ProviderConfig) *Provider {
	// Use in-memory repository if none provided
	charRepo := cfg.CharacterRepository
	if charRepo == nil {
		charRepo = characters.NewInMemoryRepository()
	}

	// Create character service
	charService := characterService.NewService(&characterService.ServiceConfig{
		DNDClient:  cfg.DNDClient,
		Repository: charRepo,
	})

	return &Provider{
		CharacterService: charService,
	}
}