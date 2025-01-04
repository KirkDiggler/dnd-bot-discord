package entities

import "fmt"

type Attribute string

var Attributes = []Attribute{AttributeStrength, AttributeDexterity, AttributeConstitution, AttributeIntelligence, AttributeWisdom, AttributeCharisma}

const (
	AttributeNone         Attribute = ""
	AttributeStrength     Attribute = "Str"
	AttributeDexterity    Attribute = "Dex"
	AttributeConstitution Attribute = "Con"
	AttributeIntelligence Attribute = "Int"
	AttributeWisdom       Attribute = "Wis"
	AttributeCharisma     Attribute = "Cha"
)

type AbilityScore struct {
	Score int
	Bonus int
}

type AbilityBonus struct {
	Attribute Attribute
	Bonus     int
}

func (a *AbilityScore) AddBonus(bonus int) *AbilityScore {
	a.Bonus += bonus

	return a
}

func (a *AbilityScore) String() string {
	return fmt.Sprintf("%d (%+d)", a.Score, a.Bonus)
}
