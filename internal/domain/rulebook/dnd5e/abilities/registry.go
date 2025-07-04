package abilities

import (
	"github.com/KirkDiggler/dnd-bot-discord/internal/dice"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/events"
	abilityService "github.com/KirkDiggler/dnd-bot-discord/internal/services/ability"
	rpgevents "github.com/KirkDiggler/rpg-toolkit/events"
)

// RegistryConfig contains dependencies needed for ability handlers
type RegistryConfig struct {
	EventBus         events.Bus     // Now using the interface
	RPGEventBus      *rpgevents.Bus // RPG toolkit event bus for future migration
	DiceRoller       dice.Roller
	EncounterService interface{} // Should have GetEncounter method
	CharacterService interface{} // Should have UpdateEquipment method
}

// RegisterAll registers all D&D 5e ability handlers with the provided registry
func RegisterAll(registry interface {
	RegisterHandler(handler abilityService.Handler)
}, cfg *RegistryConfig) {
	// Register rage
	rageHandler := NewRageHandler(cfg.EventBus)
	if cfg.EncounterService != nil {
		rageHandler.SetEncounterService(cfg.EncounterService)
	}
	if cfg.CharacterService != nil {
		rageHandler.SetCharacterService(cfg.CharacterService)
	}
	registry.RegisterHandler(NewServiceHandlerAdapter(rageHandler))

	// Register second wind
	secondWindHandler := NewSecondWindHandler(cfg.DiceRoller)
	registry.RegisterHandler(NewServiceHandlerAdapter(secondWindHandler))

	// Register bardic inspiration
	bardicHandler := NewBardicInspirationHandler()
	registry.RegisterHandler(NewServiceHandlerAdapter(bardicHandler))

	// Register lay on hands
	layOnHandsHandler := NewLayOnHandsHandler()
	registry.RegisterHandler(NewServiceHandlerAdapter(layOnHandsHandler))

	// Register divine sense
	divineSenseHandler := NewDivineSenseHandler()
	registry.RegisterHandler(NewServiceHandlerAdapter(divineSenseHandler))

	// Register vicious mockery (bard cantrip)
	RegisterViciousMockeryHandler(registry, cfg)
}
