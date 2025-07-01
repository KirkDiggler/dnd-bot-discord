package character

import (
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/shared"
)

// GetResources returns the character's resources, initializing if needed
func (c *Character) GetResources() *CharacterResources {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.Resources == nil {
		c.initializeResourcesInternal()
	}
	return c.Resources
}

// InitializeResources sets up the character's resources based on class and level
func (c *Character) InitializeResources() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.initializeResourcesInternal()
}

// initializeResourcesInternal is the internal resource initialization (caller must hold lock)
func (c *Character) initializeResourcesInternal() {
	if c.Resources == nil {
		c.Resources = &CharacterResources{}
	}

	// Initialize basic resources
	if c.Class != nil {
		c.Resources.Initialize(c.Class, c.Level)
	}

	// Set HP based on character's max HP (includes CON bonus)
	c.Resources.HP = shared.HPResource{
		Current: c.CurrentHitPoints,
		Max:     c.MaxHitPoints,
	}

	// Initialize class-specific abilities at level 1
	c.initializeClassAbilities()
}
