package events

// SourceType represents the type of source for a modifier
type SourceType string

const (
	SourceTypeFeat         SourceType = "feat"
	SourceTypeSpell        SourceType = "spell"
	SourceTypeItem         SourceType = "item"
	SourceTypeClassFeature SourceType = "class_feature"
	SourceTypeRacialTrait  SourceType = "racial_trait"
	SourceTypeCondition    SourceType = "condition"
)

// ModifierSource represents the source of a modifier
type ModifierSource struct {
	Type SourceType
	Name string
	ID   string
}

// Priority ranges for consistent modifier ordering
const (
	PriorityPreCalculation  = 0   // 0-99: Pre-calculation modifiers (set base values)
	PriorityFeatures        = 100 // 100-199: Feature modifiers (class features, racial traits)
	PriorityStatusEffects   = 200 // 200-299: Status effects (conditions, spells)
	PriorityEquipment       = 300 // 300-399: Equipment modifiers
	PriorityTemporaryBonus  = 400 // 400-499: Temporary bonuses (inspiration, guidance)
	PriorityPostCalculation = 500 // 500+: Post-calculation modifiers (caps, limits)
)
