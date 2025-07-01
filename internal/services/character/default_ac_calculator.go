package character

import (
	charDomain "github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	features2 "github.com/KirkDiggler/dnd-bot-discord/internal/domain/rulebook/dnd5e/features"
)

// defaultACCalculator is a temporary wrapper around the existing CalculateAC function
// TODO: Remove this once all rulebook calculators are properly implemented
type defaultACCalculator struct{}

// Calculate implements the ACCalculator interface
func (d *defaultACCalculator) Calculate(char *charDomain.Character) int {
	return features2.CalculateAC(char)
}
