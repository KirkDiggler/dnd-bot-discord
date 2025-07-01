package features

import (
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/rulebook"
)

// KeenSensesHandler implements the Elf "Keen Senses" feature
type KeenSensesHandler struct{}

func NewKeenSensesHandler() *KeenSensesHandler {
	return &KeenSensesHandler{}
}

func (h *KeenSensesHandler) GetKey() string {
	return "keen_senses"
}

func (h *KeenSensesHandler) ApplyPassiveEffects(char *character.Character) error {
	// Grant Perception proficiency
	perceptionProf := &rulebook.Proficiency{
		Key:  "skill-perception",
		Name: "Perception",
		Type: rulebook.ProficiencyTypeSkill,
	}

	char.AddProficiency(perceptionProf)
	return nil
}

func (h *KeenSensesHandler) ModifySkillCheck(char *character.Character, skillKey string, baseResult int) (int, bool) {
	// Keen Senses doesn't modify skill checks beyond granting proficiency
	return baseResult, false
}

func (h *KeenSensesHandler) GetPassiveDisplayInfo(char *character.Character) (string, bool) {
	// Check if character has Perception proficiency (might be from this feature or others)
	if char.HasSkillProficiency("skill-perception") {
		return "👁️ Keen Senses: Proficient in Perception", true
	}
	return "", false
}

// DarkvisionHandler implements the Darkvision racial feature
type DarkvisionHandler struct{}

func NewDarkvisionHandler() *DarkvisionHandler {
	return &DarkvisionHandler{}
}

func (h *DarkvisionHandler) GetKey() string {
	return "darkvision"
}

func (h *DarkvisionHandler) ApplyPassiveEffects(char *character.Character) error {
	// Darkvision is a passive ability that doesn't modify character stats
	// The mechanical effect would be handled in specific game situations
	return nil
}

func (h *DarkvisionHandler) ModifySkillCheck(char *character.Character, skillKey string, baseResult int) (int, bool) {
	// Darkvision doesn't directly modify skill checks
	// In actual play, it would affect Perception checks in darkness, but that requires situational context
	return baseResult, false
}

func (h *DarkvisionHandler) GetPassiveDisplayInfo(char *character.Character) (string, bool) {
	return "🌙 Darkvision: See in dim light within 60 feet", true
}

// StonecunningHandler implements the Dwarf "Stonecunning" feature
type StonecunningHandler struct{}

func NewStonecunningHandler() *StonecunningHandler {
	return &StonecunningHandler{}
}

func (h *StonecunningHandler) GetKey() string {
	return "stonecunning"
}

func (h *StonecunningHandler) ApplyPassiveEffects(char *character.Character) error {
	// Stonecunning doesn't grant general proficiency, it's situational
	// The mechanical effect is applied during specific History checks
	return nil
}

func (h *StonecunningHandler) ModifySkillCheck(char *character.Character, skillKey string, baseResult int) (int, bool) {
	// This would need context about whether the History check is stonework-related
	// For now, we just document that this feature exists
	// TODO: Implement situational skill check modifications
	return baseResult, false
}

func (h *StonecunningHandler) GetPassiveDisplayInfo(char *character.Character) (string, bool) {
	return "🪨 Stonecunning: Double proficiency bonus on stonework History checks", true
}

// BraveHandler implements the Halfling "Brave" feature
type BraveHandler struct{}

func NewBraveHandler() *BraveHandler {
	return &BraveHandler{}
}

func (h *BraveHandler) GetKey() string {
	return "brave"
}

func (h *BraveHandler) ApplyPassiveEffects(char *character.Character) error {
	// Brave is a situational feature that affects fear saves
	// No permanent character modifications needed
	return nil
}

func (h *BraveHandler) ModifySkillCheck(char *character.Character, skillKey string, baseResult int) (int, bool) {
	// Brave affects saving throws, not skill checks
	return baseResult, false
}

func (h *BraveHandler) GetPassiveDisplayInfo(char *character.Character) (string, bool) {
	return "🛡️ Brave: Advantage on saving throws against being frightened", true
}

// DwarvenResilienceHandler implements the Dwarf "Dwarven Resilience" feature
type DwarvenResilienceHandler struct{}

func NewDwarvenResilienceHandler() *DwarvenResilienceHandler {
	return &DwarvenResilienceHandler{}
}

func (h *DwarvenResilienceHandler) GetKey() string {
	return "dwarven_resilience"
}

func (h *DwarvenResilienceHandler) ApplyPassiveEffects(char *character.Character) error {
	// Dwarven Resilience provides situational bonuses
	// No permanent character modifications needed
	return nil
}

func (h *DwarvenResilienceHandler) ModifySkillCheck(char *character.Character, skillKey string, baseResult int) (int, bool) {
	// Affects saves, not skill checks
	return baseResult, false
}

func (h *DwarvenResilienceHandler) GetPassiveDisplayInfo(char *character.Character) (string, bool) {
	return "💚 Dwarven Resilience: Advantage on poison saves, resistance to poison damage", true
}

// FeyAncestryHandler implements the Elf "Fey Ancestry" feature
type FeyAncestryHandler struct{}

func NewFeyAncestryHandler() *FeyAncestryHandler {
	return &FeyAncestryHandler{}
}

func (h *FeyAncestryHandler) GetKey() string {
	return "fey_ancestry"
}

func (h *FeyAncestryHandler) ApplyPassiveEffects(char *character.Character) error {
	// Fey Ancestry provides situational bonuses
	// No permanent character modifications needed
	return nil
}

func (h *FeyAncestryHandler) ModifySkillCheck(char *character.Character, skillKey string, baseResult int) (int, bool) {
	// Affects saves, not skill checks
	return baseResult, false
}

func (h *FeyAncestryHandler) GetPassiveDisplayInfo(char *character.Character) (string, bool) {
	return "🧚 Fey Ancestry: Advantage on charm saves, immune to magical sleep", true
}
