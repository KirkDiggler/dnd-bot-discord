package feats

import (
	"fmt"
	"log"
	"sync"

	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	rpgevents "github.com/KirkDiggler/rpg-toolkit/events"
)

// Registry manages all available feats
type Registry struct {
	mu    sync.RWMutex
	feats map[string]Feat
}

// GlobalRegistry is the singleton feat registry
var GlobalRegistry = &Registry{
	feats: make(map[string]Feat),
}

// Register adds a feat to the registry
func (r *Registry) Register(feat Feat) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if feat == nil {
		return fmt.Errorf("feat cannot be nil")
	}

	key := feat.Key()
	if key == "" {
		return fmt.Errorf("feat key cannot be empty")
	}

	if _, exists := r.feats[key]; exists {
		return fmt.Errorf("feat %s already registered", key)
	}

	r.feats[key] = feat
	return nil
}

// Get retrieves a feat by key
func (r *Registry) Get(key string) (Feat, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	feat, exists := r.feats[key]
	return feat, exists
}

// List returns all registered feats
func (r *Registry) List() []Feat {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]Feat, 0, len(r.feats))
	for _, feat := range r.feats {
		result = append(result, feat)
	}
	return result
}

// AvailableFeats returns feats that a character can take
func (r *Registry) AvailableFeats(char *character.Character) []Feat {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := []Feat{}
	for _, feat := range r.feats {
		// Check if character already has this feat
		hasFeat := false
		for _, feature := range char.Features {
			if feature.Key == feat.Key() && feature.Type == "feat" {
				hasFeat = true
				break
			}
		}

		// Check if character meets prerequisites
		if !hasFeat && feat.CanTake(char) {
			result = append(result, feat)
		}
	}
	return result
}

// ApplyFeat applies a feat to a character and registers its handlers
func (r *Registry) ApplyFeat(key string, char *character.Character, eventBus *rpgevents.Bus) error {
	feat, exists := r.Get(key)
	if !exists {
		return fmt.Errorf("feat %s not found", key)
	}

	// Check if character already has this feat
	for _, feature := range char.Features {
		if feature.Key == key && feature.Type == "feat" {
			return fmt.Errorf("character already has feat %s", key)
		}
	}

	// Check prerequisites
	if !feat.CanTake(char) {
		return fmt.Errorf("character does not meet prerequisites for feat %s", key)
	}

	// Apply the feat
	if err := feat.Apply(char); err != nil {
		return fmt.Errorf("failed to apply feat %s: %w", key, err)
	}

	// Register event handlers if event bus is provided
	if eventBus != nil {
		feat.RegisterHandlers(eventBus, char)
	}

	return nil
}

// RegisterAllHandlers registers event handlers for all of a character's feats
func (r *Registry) RegisterAllHandlers(char *character.Character, eventBus *rpgevents.Bus) {
	if eventBus == nil {
		return
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, feature := range char.Features {
		if feature.Type == "feat" {
			if feat, exists := r.feats[feature.Key]; exists {
				feat.RegisterHandlers(eventBus, char)
			}
		}
	}
}

// RegisterAll registers all standard D&D 5e feats
func RegisterAll() {
	// Combat feats
	if err := GlobalRegistry.Register(NewAlertFeat()); err != nil {
		log.Printf("Failed to register Alert feat: %v", err)
	}
	if err := GlobalRegistry.Register(NewGreatWeaponMasterFeat()); err != nil {
		log.Printf("Failed to register Great Weapon Master feat: %v", err)
	}
	if err := GlobalRegistry.Register(NewLuckyFeat()); err != nil {
		log.Printf("Failed to register Lucky feat: %v", err)
	}
	if err := GlobalRegistry.Register(NewSharpshooterFeat()); err != nil {
		log.Printf("Failed to register Sharpshooter feat: %v", err)
	}
	if err := GlobalRegistry.Register(NewToughFeat()); err != nil {
		log.Printf("Failed to register Tough feat: %v", err)
	}
	if err := GlobalRegistry.Register(NewWarCasterFeat()); err != nil {
		log.Printf("Failed to register War Caster feat: %v", err)
	}

	// TODO: Add more feats
	// - Sentinel
	// - Polearm Master
	// - Crossbow Expert
	// - Mobile
	// - Resilient
	// - Observant
	// - etc.
}
