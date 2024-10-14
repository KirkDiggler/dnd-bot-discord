package damage

import "github.com/KirkDiggler/dnd-bot-discord/internal/dice"

type Type string

const (
	TypeAcid        Type = "acid"
	TypeCold        Type = "cold"
	TypeFire        Type = "fire"
	TypeForce       Type = "force"
	TypeLightning   Type = "lightning"
	TypeNecrotic    Type = "necrotic"
	TypePoison      Type = "poison"
	TypePsychic     Type = "psychic"
	TypeRadiant     Type = "radiant"
	TypeThunder     Type = "thunder"
	TypeBludgeoning Type = "bludgeoning"
	TypePiercing    Type = "piercing"
	TypeSlashing    Type = "slashing"
	TypeNone        Type = "none"
)

type Damage struct {
	DiceCount  int
	DiceSize   int
	Bonus      int
	DamageType Type
}

func (d *Damage) Deal() int {
	base, _ := dice.Roll(d.DiceCount, d.DiceSize, 0)
	return base.Total + d.Bonus
}
