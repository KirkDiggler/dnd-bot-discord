package feats

import (
	"context"
	"log"

	"github.com/KirkDiggler/dnd-bot-discord/internal/adapters/rpgtoolkit"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	rpgevents "github.com/KirkDiggler/rpg-toolkit/events"
)

// LuckyFeat implements the Lucky feat
// - 3 luck points per long rest
// - Can spend a luck point to roll an additional d20 for attack, ability check, or saving throw
// - Can spend a luck point when attacked to roll a d20 and choose which roll the attacker uses
type LuckyFeat struct {
	BaseFeat
	luckPoints map[string]int // Track luck points per character ID
}

// NewLuckyFeat creates a new Lucky feat
func NewLuckyFeat() Feat {
	return &LuckyFeat{
		BaseFeat: BaseFeat{
			key:  "lucky",
			name: "Lucky",
			description: "You have inexplicable luck that seems to kick in at just the right moment. " +
				"You have 3 luck points that can be used to reroll attack rolls, ability checks, or saving throws.",
			prerequisites: []Prerequisite{
				// No prerequisites for Lucky
			},
		},
		luckPoints: make(map[string]int),
	}
}

// Apply applies the Lucky feat benefits
func (f *LuckyFeat) Apply(char *character.Character) error {
	// First apply the base feat (adds to character features)
	if err := f.BaseFeat.Apply(char); err != nil {
		return err
	}

	// Initialize luck points for this character
	f.luckPoints[char.ID] = 3

	// Add luck points to character metadata
	for i, feature := range char.Features {
		if feature.Key == "lucky" && feature.Type == "feat" {
			if feature.Metadata == nil {
				feature.Metadata = make(map[string]any)
			}
			feature.Metadata["luck_points"] = 3
			char.Features[i] = feature
			break
		}
	}

	return nil
}

// RegisterHandlers registers event handlers for Lucky
func (f *LuckyFeat) RegisterHandlers(bus *rpgevents.Bus, char *character.Character) {
	if bus == nil {
		return
	}

	// Helper function to check and consume luck points
	useLuckPoint := func(charID string) bool {
		points, exists := f.luckPoints[charID]
		if !exists || points <= 0 {
			return false
		}
		f.luckPoints[charID]--
		return true
	}

	// Subscribe to various roll events where luck can be used
	// Attack rolls
	bus.SubscribeFunc(rpgevents.EventAfterAttackRoll, 70, func(ctx context.Context, event rpgevents.Event) error {
		// Check if this is our character attacking
		actor, ok := rpgtoolkit.ExtractCharacter(event.Source())
		if !ok || actor == nil || actor.ID != char.ID {
			return nil
		}

		// Check if character has Lucky feat
		hasLucky := false
		for _, feature := range actor.Features {
			if feature.Key == "lucky" && feature.Type == "feat" {
				hasLucky = true
				break
			}
		}

		if !hasLucky {
			return nil
		}

		// Check if the attack missed (would need to implement miss detection)
		// For now, we'll assume the player would want to use lucky on low rolls
		attackRoll, _ := rpgtoolkit.GetIntContext(event, "attack_roll")
		if attackRoll <= 10 && useLuckPoint(actor.ID) {
			// In a real implementation, we'd:
			// 1. Roll another d20
			// 2. Let the player choose which roll to use
			// 3. Update the attack result

			log.Printf("[LUCKY] %s uses a luck point on attack roll (original: %d, %d points remaining)",
				actor.Name, attackRoll, f.luckPoints[actor.ID])

			// Emit event for UI to handle the reroll
			luckyEvent := rpgevents.NewGameEvent(
				"dndbot.lucky_reroll",
				rpgtoolkit.WrapCharacter(actor),
				nil,
			)
			luckyEvent.Context().Set("roll_type", "attack")
			luckyEvent.Context().Set("original_roll", attackRoll)
			luckyEvent.Context().Set("luck_points_remaining", f.luckPoints[actor.ID])

			if err := bus.Publish(ctx, luckyEvent); err != nil {
				log.Printf("[LUCKY] Failed to publish reroll event: %v", err)
			}
		}

		return nil
	})

	// Saving throws
	bus.SubscribeFunc(rpgevents.EventAfterSavingThrow, 70, func(ctx context.Context, event rpgevents.Event) error {
		// Check if this is our character making a save
		actor, ok := rpgtoolkit.ExtractCharacter(event.Source())
		if !ok || actor == nil || actor.ID != char.ID {
			return nil
		}

		// Check if character has Lucky feat
		hasLucky := false
		for _, feature := range actor.Features {
			if feature.Key == "lucky" && feature.Type == "feat" {
				hasLucky = true
				break
			}
		}

		if !hasLucky {
			return nil
		}

		// Check if the save failed
		saveRoll, _ := rpgtoolkit.GetIntContext(event, "save_roll")
		dc, _ := rpgtoolkit.GetIntContext(event, "dc")

		if saveRoll < dc && useLuckPoint(actor.ID) {
			log.Printf("[LUCKY] %s uses a luck point on saving throw (original: %d vs DC %d, %d points remaining)",
				actor.Name, saveRoll, dc, f.luckPoints[actor.ID])

			// Emit event for UI to handle the reroll
			luckyEvent := rpgevents.NewGameEvent(
				"dndbot.lucky_reroll",
				rpgtoolkit.WrapCharacter(actor),
				nil,
			)
			luckyEvent.Context().Set("roll_type", "save")
			luckyEvent.Context().Set("original_roll", saveRoll)
			luckyEvent.Context().Set("dc", dc)
			luckyEvent.Context().Set("luck_points_remaining", f.luckPoints[actor.ID])

			if err := bus.Publish(ctx, luckyEvent); err != nil {
				log.Printf("[LUCKY] Failed to publish reroll event: %v", err)
			}
		}

		return nil
	})

	// Long rest to restore luck points
	bus.SubscribeFunc(rpgevents.EventOnLongRest, 100, func(ctx context.Context, event rpgevents.Event) error {
		// Check if this is our character resting
		actor, ok := rpgtoolkit.ExtractCharacter(event.Source())
		if !ok || actor == nil || actor.ID != char.ID {
			return nil
		}

		// Restore luck points
		f.luckPoints[actor.ID] = 3

		// Update character feature metadata
		for i, feature := range actor.Features {
			if feature.Key == "lucky" && feature.Type == "feat" {
				if feature.Metadata == nil {
					feature.Metadata = make(map[string]any)
				}
				feature.Metadata["luck_points"] = 3
				actor.Features[i] = feature
				break
			}
		}

		log.Printf("[LUCKY] %s restores all luck points on long rest", actor.Name)

		return nil
	})
}
