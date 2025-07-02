package shared

// TargetType defines what kinds of targets an ability can affect
type TargetType string

const (
	TargetTypeSelf        TargetType = "self"         // Self only
	TargetTypeSingleAlly  TargetType = "single_ally"  // One ally
	TargetTypeSingleEnemy TargetType = "single_enemy" // One enemy
	TargetTypeSingleAny   TargetType = "single_any"   // Any one target
	TargetTypeAllAllies   TargetType = "all_allies"   // All allies
	TargetTypeAllEnemies  TargetType = "all_enemies"  // All enemies
	TargetTypeArea        TargetType = "area"         // Area effect
	TargetTypeNone        TargetType = "none"         // No target needed
)

// RangeType defines how range is calculated
type RangeType string

const (
	RangeTypeSelf      RangeType = "self"
	RangeTypeTouch     RangeType = "touch"
	RangeTypeRanged    RangeType = "ranged"
	RangeTypeSight     RangeType = "sight"
	RangeTypeUnlimited RangeType = "unlimited"
)

// ComponentType defines spell/ability components
type ComponentType string

const (
	ComponentVerbal   ComponentType = "V"
	ComponentSomatic  ComponentType = "S"
	ComponentMaterial ComponentType = "M"
)

// AbilityTargeting defines targeting rules for an ability
type AbilityTargeting struct {
	TargetType    TargetType      `json:"target_type"`
	RangeType     RangeType       `json:"range_type"`
	Range         int             `json:"range"` // In feet, 0 for self/touch
	Components    []ComponentType `json:"components,omitempty"`
	SaveType      Attribute       `json:"save_type,omitempty"` // Which attribute save
	SaveDC        int             `json:"save_dc,omitempty"`   // 0 means calculate from caster
	Concentration bool            `json:"concentration"`
}
