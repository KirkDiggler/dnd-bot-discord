package character

import (
	"fmt"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/rulebook"
	"strings"

	"github.com/KirkDiggler/dnd-bot-discord/internal/dice"
)

func (c *Character) AddProficiency(p *rulebook.Proficiency) {
	if c.Proficiencies == nil {
		c.Proficiencies = make(map[rulebook.ProficiencyType][]*rulebook.Proficiency)
	}
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.Proficiencies[p.Type] == nil {
		c.Proficiencies[p.Type] = make([]*rulebook.Proficiency, 0)
	}

	// Check for duplicates
	for _, existing := range c.Proficiencies[p.Type] {
		if existing.Key == p.Key {
			return // Already have this proficiency
		}
	}

	c.Proficiencies[p.Type] = append(c.Proficiencies[p.Type], p)
}

// SetProficiencies replaces all proficiencies of a given type
// This is used when selecting proficiencies during character creation
func (c *Character) SetProficiencies(profType rulebook.ProficiencyType, proficiencies []*rulebook.Proficiency) {
	if c.Proficiencies == nil {
		c.Proficiencies = make(map[rulebook.ProficiencyType][]*rulebook.Proficiency)
	}
	c.mu.Lock()
	defer c.mu.Unlock()

	// Replace all proficiencies of this type
	c.Proficiencies[profType] = proficiencies
}

// GetProficiencyBonus returns the character's proficiency bonus based on level
func (c *Character) GetProficiencyBonus() int {
	if c.Level == 0 {
		return 2 // Default for level 1
	}
	return 2 + ((c.Level - 1) / 4)
}

// HasSavingThrowProficiency checks if the character is proficient in a saving throw
func (c *Character) HasSavingThrowProficiency(attribute Attribute) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.Proficiencies == nil || c.Proficiencies[rulebook.ProficiencyTypeSavingThrow] == nil {
		return false
	}

	// Convert attribute to the expected saving throw key format
	savingThrowKey := fmt.Sprintf("saving-throw-%s", strings.ToLower(string(attribute)))

	for _, prof := range c.Proficiencies[rulebook.ProficiencyTypeSavingThrow] {
		if strings.EqualFold(prof.Key, savingThrowKey) {
			return true
		}
	}

	return false
}

// GetSavingThrowBonus calculates the total saving throw bonus for an attribute
func (c *Character) GetSavingThrowBonus(attribute Attribute) int {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Get ability modifier
	modifier := 0
	if abilityScore, exists := c.Attributes[attribute]; exists && abilityScore != nil {
		modifier = abilityScore.Bonus
	}

	// Check proficiency without locking again
	isProficient := false
	if c.Proficiencies != nil && c.Proficiencies[rulebook.ProficiencyTypeSavingThrow] != nil {
		savingThrowKey := fmt.Sprintf("saving-throw-%s", strings.ToLower(string(attribute)))
		for _, prof := range c.Proficiencies[rulebook.ProficiencyTypeSavingThrow] {
			if strings.EqualFold(prof.Key, savingThrowKey) {
				isProficient = true
				break
			}
		}
	}

	// Add proficiency bonus if proficient
	if isProficient {
		modifier += c.GetProficiencyBonus()
	}

	return modifier
}

// RollSavingThrow rolls a saving throw for the given attribute
func (c *Character) RollSavingThrow(attribute Attribute) (*dice.RollResult, int, error) {
	bonus := c.GetSavingThrowBonus(attribute)

	// Roll 1d20
	result, err := c.getDiceRoller().Roll(1, 20, bonus)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to roll saving throw: %w", err)
	}

	return result, result.Total, nil
}

// HasSkillProficiency checks if the character is proficient in a skill
func (c *Character) HasSkillProficiency(skillKey string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.Proficiencies == nil || c.Proficiencies[rulebook.ProficiencyTypeSkill] == nil {
		return false
	}

	for _, prof := range c.Proficiencies[rulebook.ProficiencyTypeSkill] {
		if strings.EqualFold(prof.Key, skillKey) {
			return true
		}
	}

	return false
}
