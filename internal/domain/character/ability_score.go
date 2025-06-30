package character

import (
	"fmt"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/shared"
)

type AbilityScore struct {
	Score int
	Bonus int
}

type AbilityBonus struct {
	Attribute shared.Attribute
	Bonus     int
}

func (a *AbilityScore) AddBonus(bonus int) *AbilityScore {
	// Add the bonus to the score
	a.Score += bonus

	// Recalculate the modifier based on the new score
	a.Bonus = (a.Score - 10) / 2

	return a
}

func (a *AbilityScore) String() string {
	return fmt.Sprintf("%d (%+d)", a.Score, a.Bonus)
}

func (a shared.Attribute) Short() string {
	return string(a)
}
