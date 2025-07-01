package character

import (
	"fmt"
)

type AbilityScore struct {
	Score int
	Bonus int
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
