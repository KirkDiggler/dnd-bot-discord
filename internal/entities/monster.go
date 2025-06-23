package entities

import "github.com/KirkDiggler/dnd-bot-discord/internal/entities/damage"

type Monster struct {
	ID          string           `json:"id"`
	Template    *MonsterTemplate `json:"template"`
	CharacterID string           `json:"character_id"`
	CurrentHP   int              `json:"current_hp"`
	Key         string           `json:"key"`
}

type MonsterTemplate struct {
	Key             string           `json:"key"`
	Name            string           `json:"name"`
	Type            string           `json:"type"`
	ArmorClass      int              `json:"armor_class"`
	HitPoints       int              `json:"hit_points"`
	HitDice         string           `json:"hit_dice"`
	Actions         []*MonsterAction `json:"actions"`
	XP              int              `json:"xp"`
	ChallengeRating float32          `json:"challenge_rating"`
}

type MonsterAction struct {
	Name        string           `json:"name"`
	AttackBonus int              `json:"attack_bonus"`
	Description string           `json:"desc"`
	Damage      []*damage.Damage `json:"damage"`
	
	// For abilities that require saving throws
	SaveDC        int       `json:"save_dc,omitempty"`        // DC for the saving throw
	SaveAttribute Attribute `json:"save_attribute,omitempty"` // Which attribute to save against (STR, DEX, etc.)
}
