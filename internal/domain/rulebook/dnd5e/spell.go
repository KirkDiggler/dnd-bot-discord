package rulebook

import "github.com/KirkDiggler/dnd-bot-discord/internal/domain/shared"

// SpellReference is a lightweight reference to a spell
type SpellReference struct {
	Key  string `json:"key"`
	Name string `json:"name"`
	URL  string `json:"url"`
}

// Spell represents a D&D 5e spell
type Spell struct {
	Key           string                   `json:"key"`
	Name          string                   `json:"name"`
	Level         int                      `json:"level"` // 0 for cantrips
	School        string                   `json:"school"`
	CastingTime   string                   `json:"casting_time"`
	Range         string                   `json:"range"`
	Components    []string                 `json:"components"`
	Duration      string                   `json:"duration"`
	Concentration bool                     `json:"concentration"`
	Ritual        bool                     `json:"ritual"`
	Description   string                   `json:"description"`
	HigherLevel   string                   `json:"higher_level,omitempty"`
	Classes       []string                 `json:"classes"`
	Damage        *SpellDamage             `json:"damage,omitempty"`
	DC            *SpellDC                 `json:"dc,omitempty"`
	AreaOfEffect  *SpellAreaOfEffect       `json:"area_of_effect,omitempty"`
	Targeting     *shared.AbilityTargeting `json:"targeting,omitempty"` // Parsed targeting rules
}

// SpellDamage represents spell damage information
type SpellDamage struct {
	DamageType    string         `json:"damage_type"`
	DamageAtLevel map[int]string `json:"damage_at_level"` // Key: spell slot level, Value: damage dice
}

// SpellDC represents spell save DC information
type SpellDC struct {
	Type    shared.Attribute `json:"type"`    // Which save
	Success string           `json:"success"` // What happens on success
}

// SpellAreaOfEffect represents area of effect information
type SpellAreaOfEffect struct {
	Type string `json:"type"` // cone, sphere, cube, line, etc.
	Size int    `json:"size"` // In feet
}
