package feats

import (
	"context"
	"log"

	"github.com/KirkDiggler/dnd-bot-discord/internal/adapters/rpgtoolkit"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	rpgevents "github.com/KirkDiggler/rpg-toolkit/events"
)

// AlertFeat implements the Alert feat
// - +5 bonus to initiative
// - Can't be surprised while conscious
// - Other creatures don't gain advantage on attack rolls against you as a result of being unseen by you
type AlertFeat struct {
	BaseFeat
}

// NewAlertFeat creates a new Alert feat
func NewAlertFeat() Feat {
	return &AlertFeat{
		BaseFeat: BaseFeat{
			key:           "alert",
			name:          "Alert",
			description:   "Always on the lookout for danger, you gain +5 to initiative and can't be surprised while conscious.",
			prerequisites: []Prerequisite{
				// No prerequisites for Alert
			},
		},
	}
}

// Apply applies the Alert feat benefits
func (f *AlertFeat) Apply(char *character.Character) error {
	// First apply the base feat (adds to character features)
	if err := f.BaseFeat.Apply(char); err != nil {
		return err
	}

	// Alert provides a flat +5 to initiative
	// This will be handled by the event handler

	return nil
}

// RegisterHandlers registers event handlers for Alert
func (f *AlertFeat) RegisterHandlers(bus *rpgevents.Bus, char *character.Character) {
	if bus == nil {
		return
	}

	// Subscribe to initiative roll events
	bus.SubscribeFunc("dndbot.on_initiative_roll", 25, func(ctx context.Context, event rpgevents.Event) error {
		// Check if this is our character rolling initiative
		actor, ok := rpgtoolkit.ExtractCharacter(event.Source())
		if !ok || actor == nil || actor.ID != char.ID {
			return nil
		}

		// Check if character has Alert feat
		hasAlert := false
		for _, feature := range actor.Features {
			if feature.Key == "alert" && feature.Type == "feat" {
				hasAlert = true
				break
			}
		}

		if !hasAlert {
			return nil
		}

		// Add +5 to initiative
		currentInit, _ := rpgtoolkit.GetIntContext(event, "initiative")
		event.Context().Set("initiative", currentInit+5)
		event.Context().Set("initiative_bonus_source", "Alert feat")

		log.Printf("[ALERT] %s gains +5 to initiative from Alert feat", actor.Name)

		return nil
	})

	// Subscribe to surprise check events
	bus.SubscribeFunc(rpgevents.EventOnConditionApplied, 10, func(ctx context.Context, event rpgevents.Event) error {
		// Check if this is a surprise condition being applied to our character
		target, ok := rpgtoolkit.ExtractCharacter(event.Target())
		if !ok || target == nil || target.ID != char.ID {
			return nil
		}

		conditionType, _ := rpgtoolkit.GetStringContext(event, "condition_type")
		if conditionType != "surprised" {
			return nil
		}

		// Check if character has Alert feat
		hasAlert := false
		for _, feature := range target.Features {
			if feature.Key == "alert" && feature.Type == "feat" {
				hasAlert = true
				break
			}
		}

		if hasAlert {
			// Cancel the surprise condition
			event.Cancel()
			log.Printf("[ALERT] %s cannot be surprised due to Alert feat", target.Name)
		}

		return nil
	})
}
