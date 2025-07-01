package features

import (
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
)

// FeatureHandler defines the interface for implementing feature mechanics
type FeatureHandler interface {
	// GetKey returns the unique identifier for this feature
	GetKey() string

	// ApplyPassiveEffects applies any passive effects to the character
	// This is called during character finalization and when features are updated
	ApplyPassiveEffects(char *character.Character) error

	// ModifySkillCheck allows features to modify skill checks
	// Returns the modified result and whether the feature applied
	ModifySkillCheck(char *character.Character, skillKey string, baseResult int) (int, bool)

	// GetPassiveDisplayInfo returns information to display on character sheets
	// Returns display text and whether this feature should be prominently shown
	GetPassiveDisplayInfo(char *character.Character) (string, bool)
}

// FeatureRegistry manages all implemented feature handlers
type FeatureRegistry struct {
	handlers map[string]FeatureHandler
}

// NewFeatureRegistry creates a new feature registry
func NewFeatureRegistry() *FeatureRegistry {
	registry := &FeatureRegistry{
		handlers: make(map[string]FeatureHandler),
	}

	// Register all implemented feature handlers
	registry.registerDefaultHandlers()

	return registry
}

// RegisterHandler registers a feature handler
func (r *FeatureRegistry) RegisterHandler(handler FeatureHandler) {
	r.handlers[handler.GetKey()] = handler
}

// GetHandler retrieves a handler by feature key
func (r *FeatureRegistry) GetHandler(key string) (FeatureHandler, bool) {
	handler, exists := r.handlers[key]
	return handler, exists
}

// ApplyAllPassiveEffects applies passive effects from all character features
func (r *FeatureRegistry) ApplyAllPassiveEffects(char *character.Character) error {
	if char.Features == nil {
		return nil
	}

	for _, feature := range char.Features {
		if feature == nil {
			continue
		}

		if handler, exists := r.GetHandler(feature.Key); exists {
			if err := handler.ApplyPassiveEffects(char); err != nil {
				// Log error but continue with other features
				// TODO: Add proper logging
				continue
			}
		}
	}

	return nil
}

// GetAllPassiveDisplayInfo gets display info for all character features
func (r *FeatureRegistry) GetAllPassiveDisplayInfo(char *character.Character) []string {
	var displayInfo []string

	if char.Features == nil {
		return displayInfo
	}

	for _, feature := range char.Features {
		if feature == nil {
			continue
		}

		if handler, exists := r.GetHandler(feature.Key); exists {
			if info, shouldDisplay := handler.GetPassiveDisplayInfo(char); shouldDisplay && info != "" {
				displayInfo = append(displayInfo, info)
			}
		}
	}

	return displayInfo
}

// registerDefaultHandlers registers all implemented feature handlers
func (r *FeatureRegistry) registerDefaultHandlers() {
	// Register racial feature handlers
	r.RegisterHandler(NewKeenSensesHandler())
	r.RegisterHandler(NewDarkvisionHandler())
	r.RegisterHandler(NewStonecunningHandler())
	r.RegisterHandler(NewBraveHandler())
	r.RegisterHandler(NewDwarvenResilienceHandler())
	r.RegisterHandler(NewFeyAncestryHandler())
}

// DefaultRegistry provides a default feature registry instance
var DefaultRegistry = NewFeatureRegistry()
