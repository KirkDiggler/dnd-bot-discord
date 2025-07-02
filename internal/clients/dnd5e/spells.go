package dnd5e

import (
	"fmt"
	"strings"

	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/rulebook/dnd5e"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/shared"
	apiDnd5e "github.com/fadedpez/dnd5e-api/clients/dnd5e"
	"github.com/fadedpez/dnd5e-api/entities"
)

// GetSpell retrieves a spell by key
func (c *client) GetSpell(key string) (*rulebook.Spell, error) {
	apiSpell, err := c.client.GetSpell(key)
	if err != nil {
		return nil, fmt.Errorf("failed to get spell %s: %w", key, err)
	}

	return c.convertSpell(apiSpell), nil
}

// ListSpellsByClass lists all spells available to a class
func (c *client) ListSpellsByClass(classKey string) ([]*rulebook.SpellReference, error) {
	spells, err := c.client.ListSpells(&apiDnd5e.ListSpellsInput{
		Class: classKey,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list spells for class %s: %w", classKey, err)
	}

	return c.convertSpellReferences(spells), nil
}

// ListSpellsByClassAndLevel lists spells available to a class at a specific level
func (c *client) ListSpellsByClassAndLevel(classKey string, level int) ([]*rulebook.SpellReference, error) {
	spells, err := c.client.ListSpells(&apiDnd5e.ListSpellsInput{
		Class: classKey,
		Level: &level,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list level %d spells for class %s: %w", level, classKey, err)
	}

	return c.convertSpellReferences(spells), nil
}

// convertSpellReferences converts API spell references to domain model
func (c *client) convertSpellReferences(refs []*entities.ReferenceItem) []*rulebook.SpellReference {
	result := make([]*rulebook.SpellReference, len(refs))
	for i, ref := range refs {
		result[i] = &rulebook.SpellReference{
			Key:  ref.Key,
			Name: ref.Name,
			URL:  "", // ReferenceItem doesn't have URL
		}
	}
	return result
}

// convertSpell converts API spell to domain model
func (c *client) convertSpell(apiSpell *entities.Spell) *rulebook.Spell {
	spell := &rulebook.Spell{
		Key:           apiSpell.Key,
		Name:          apiSpell.Name,
		Level:         apiSpell.SpellLevel,
		CastingTime:   apiSpell.CastingTime,
		Range:         apiSpell.Range,
		Duration:      apiSpell.Duration,
		Concentration: apiSpell.Concentration,
		Ritual:        apiSpell.Ritual,
		Components:    []string{}, // Will parse from description
		Classes:       c.extractClassNames(apiSpell.SpellClasses),
	}

	// Convert school
	if apiSpell.SpellSchool != nil {
		spell.School = apiSpell.SpellSchool.Name
	}

	// Convert damage
	if apiSpell.SpellDamage != nil {
		spell.Damage = c.convertSpellDamage(apiSpell.SpellDamage)
	}

	// Convert DC
	if apiSpell.DC != nil {
		spell.DC = c.convertSpellDC(apiSpell.DC)
	}

	// Convert area of effect
	if apiSpell.AreaOfEffect != nil {
		spell.AreaOfEffect = &rulebook.SpellAreaOfEffect{
			Type: apiSpell.AreaOfEffect.Type,
			Size: apiSpell.AreaOfEffect.Size,
		}
	}

	// Parse targeting from spell data
	spell.Targeting = c.parseSpellTargeting(apiSpell)

	return spell
}

// convertSpellDamage converts API spell damage to domain model
func (c *client) convertSpellDamage(apiDamage *entities.SpellDamage) *rulebook.SpellDamage {
	damage := &rulebook.SpellDamage{
		DamageAtLevel: make(map[int]string),
	}

	if apiDamage.SpellDamageType != nil {
		damage.DamageType = apiDamage.SpellDamageType.Name
	}

	// Convert damage at each level
	if apiDamage.SpellDamageAtSlotLevel != nil {
		if apiDamage.SpellDamageAtSlotLevel.FirstLevel != "" {
			damage.DamageAtLevel[1] = apiDamage.SpellDamageAtSlotLevel.FirstLevel
		}
		if apiDamage.SpellDamageAtSlotLevel.SecondLevel != "" {
			damage.DamageAtLevel[2] = apiDamage.SpellDamageAtSlotLevel.SecondLevel
		}
		if apiDamage.SpellDamageAtSlotLevel.ThirdLevel != "" {
			damage.DamageAtLevel[3] = apiDamage.SpellDamageAtSlotLevel.ThirdLevel
		}
		if apiDamage.SpellDamageAtSlotLevel.FourthLevel != "" {
			damage.DamageAtLevel[4] = apiDamage.SpellDamageAtSlotLevel.FourthLevel
		}
		if apiDamage.SpellDamageAtSlotLevel.FifthLevel != "" {
			damage.DamageAtLevel[5] = apiDamage.SpellDamageAtSlotLevel.FifthLevel
		}
		if apiDamage.SpellDamageAtSlotLevel.SixthLevel != "" {
			damage.DamageAtLevel[6] = apiDamage.SpellDamageAtSlotLevel.SixthLevel
		}
		if apiDamage.SpellDamageAtSlotLevel.SeventhLevel != "" {
			damage.DamageAtLevel[7] = apiDamage.SpellDamageAtSlotLevel.SeventhLevel
		}
		if apiDamage.SpellDamageAtSlotLevel.EighthLevel != "" {
			damage.DamageAtLevel[8] = apiDamage.SpellDamageAtSlotLevel.EighthLevel
		}
		if apiDamage.SpellDamageAtSlotLevel.NinthLevel != "" {
			damage.DamageAtLevel[9] = apiDamage.SpellDamageAtSlotLevel.NinthLevel
		}
	}

	return damage
}

// convertSpellDC converts API spell DC to domain model
func (c *client) convertSpellDC(apiDC *entities.DC) *rulebook.SpellDC {
	dc := &rulebook.SpellDC{
		Success: apiDC.DCSuccess,
	}

	// Map DC type to attribute
	if apiDC.DCType != nil {
		switch strings.ToLower(apiDC.DCType.Name) {
		case "str", "strength":
			dc.Type = shared.AttributeStrength
		case "dex", "dexterity":
			dc.Type = shared.AttributeDexterity
		case "con", "constitution":
			dc.Type = shared.AttributeConstitution
		case "int", "intelligence":
			dc.Type = shared.AttributeIntelligence
		case "wis", "wisdom":
			dc.Type = shared.AttributeWisdom
		case "cha", "charisma":
			dc.Type = shared.AttributeCharisma
		}
	}

	return dc
}

// parseSpellTargeting parses targeting rules from spell data
func (c *client) parseSpellTargeting(spell *entities.Spell) *shared.AbilityTargeting {
	targeting := &shared.AbilityTargeting{
		Components: c.parseComponents(spell),
	}

	// Parse range
	rangeLower := strings.ToLower(spell.Range)
	switch {
	case rangeLower == "self":
		targeting.RangeType = shared.RangeTypeSelf
		targeting.TargetType = shared.TargetTypeSelf
	case rangeLower == "touch":
		targeting.RangeType = shared.RangeTypeTouch
		targeting.Range = 5
	case strings.Contains(rangeLower, "feet"):
		targeting.RangeType = shared.RangeTypeRanged
		// Extract number from "X feet"
		var feet int
		_, _ = fmt.Sscanf(spell.Range, "%d feet", &feet) //nolint:errcheck
		targeting.Range = feet
	case rangeLower == "sight":
		targeting.RangeType = shared.RangeTypeSight
	case rangeLower == "unlimited":
		targeting.RangeType = shared.RangeTypeUnlimited
	}

	// Parse target type based on area of effect or spell patterns
	if spell.AreaOfEffect != nil {
		targeting.TargetType = shared.TargetTypeArea
	} else if strings.Contains(rangeLower, "self") {
		targeting.TargetType = shared.TargetTypeSelf
	} else {
		// Default to single target for ranged spells
		targeting.TargetType = shared.TargetTypeSingleAny
	}

	// Set concentration
	targeting.Concentration = spell.Concentration

	// Set save DC if applicable
	if spell.DC != nil {
		targeting.SaveType = c.convertSpellDC(spell.DC).Type
	}

	return targeting
}

// parseComponents extracts components from spell (would need actual parsing of description)
func (c *client) parseComponents(spell *entities.Spell) []shared.ComponentType {
	// TODO: Parse from spell description or add to API
	// For now, return common defaults
	components := []shared.ComponentType{}

	// Most spells have verbal components
	components = append(components, shared.ComponentVerbal)

	// Many spells have somatic components
	if spell.Range != "Self" {
		components = append(components, shared.ComponentSomatic)
	}

	return components
}

// extractClassNames extracts class names from reference items
func (c *client) extractClassNames(refs []*entities.ReferenceItem) []string {
	names := make([]string, len(refs))
	for i, ref := range refs {
		names[i] = ref.Key
	}
	return names
}
