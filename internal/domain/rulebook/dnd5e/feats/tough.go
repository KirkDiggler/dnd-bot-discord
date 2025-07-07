package feats

import (
	"log"

	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	rpgevents "github.com/KirkDiggler/rpg-toolkit/events"
)

// ToughFeat implements the Tough feat
// - Your hit point maximum increases by an amount equal to twice your level when you gain this feat
// - Whenever you gain a level thereafter, your hit point maximum increases by an additional 2 hit points
type ToughFeat struct {
	BaseFeat
}

// NewToughFeat creates a new Tough feat
func NewToughFeat() Feat {
	return &ToughFeat{
		BaseFeat: BaseFeat{
			key:  "tough",
			name: "Tough",
			description: "Your hit point maximum increases by an amount equal to twice your level. " +
				"Whenever you gain a level, your hit point maximum increases by an additional 2 hit points.",
			prerequisites: []Prerequisite{
				// No prerequisites for Tough
			},
		},
	}
}

// Apply applies the Tough feat benefits
func (f *ToughFeat) Apply(char *character.Character) error {
	// First apply the base feat (adds to character features)
	if err := f.BaseFeat.Apply(char); err != nil {
		return err
	}

	// Increase hit points by 2 per level
	hpIncrease := 2 * char.Level
	char.MaxHitPoints += hpIncrease

	// Also increase current HP if character is at full health
	if char.CurrentHitPoints == char.MaxHitPoints-hpIncrease {
		char.CurrentHitPoints = char.MaxHitPoints
	}

	log.Printf("[TOUGH] %s gains %d HP from Tough feat (new max: %d)",
		char.Name, hpIncrease, char.MaxHitPoints)

	// Add feat metadata to track HP bonus
	for i, feature := range char.Features {
		if feature.Key == "tough" && feature.Type == "feat" {
			if feature.Metadata == nil {
				feature.Metadata = make(map[string]any)
			}
			feature.Metadata["hp_bonus"] = hpIncrease
			char.Features[i] = feature
			break
		}
	}

	return nil
}

// RegisterHandlers registers event handlers for Tough
func (f *ToughFeat) RegisterHandlers(bus *rpgevents.Bus, char *character.Character) {
	// Tough feat doesn't need event handlers - it applies passive bonuses
	// The HP increase would be recalculated on level up by the character advancement system
}
