package character

import "github.com/KirkDiggler/dnd-bot-discord/internal/domain/rulebook/dnd5e"

// SpellList represents a character's known/prepared spells
type SpellList struct {
	KnownSpells    []string `json:"known_spells"`    // Spell keys for known spells
	PreparedSpells []string `json:"prepared_spells"` // Spell keys for prepared spells (wizards, clerics, etc.)
	Cantrips       []string `json:"cantrips"`        // Cantrip keys
}

// AddKnownSpell adds a spell to the known spells list
func (c *Character) AddKnownSpell(spellKey string) {
	if c.Spells == nil {
		c.Spells = &SpellList{}
	}

	// Check if already known
	for _, known := range c.Spells.KnownSpells {
		if known == spellKey {
			return
		}
	}

	c.Spells.KnownSpells = append(c.Spells.KnownSpells, spellKey)
}

// AddCantrip adds a cantrip to the cantrips list
func (c *Character) AddCantrip(cantripKey string) {
	if c.Spells == nil {
		c.Spells = &SpellList{}
	}

	// Check if already known
	for _, known := range c.Spells.Cantrips {
		if known == cantripKey {
			return
		}
	}

	c.Spells.Cantrips = append(c.Spells.Cantrips, cantripKey)
}

// GetSpellSlotsForLevel returns how many spells a character knows at a given level
func GetSpellSlotsForLevel(class *rulebook.Class, level int) (cantrips, spellsKnown int) {
	// This would be populated from class data
	// For now, hardcode some basics
	switch class.Key {
	case "bard":
		// Bards know spells
		cantrips = 2 // 2 cantrips at level 1
		if level >= 4 {
			cantrips = 3
		}
		if level >= 10 {
			cantrips = 4
		}

		// Spells known starts at 4 and increases
		spellsKnown = 4
		if level >= 2 {
			spellsKnown = 5
		}
		if level >= 3 {
			spellsKnown = 6
		}
		// ... etc

	case "wizard":
		// Wizards prepare spells from spellbook
		cantrips = 3 // 3 cantrips at level 1
		if level >= 4 {
			cantrips = 4
		}
		if level >= 10 {
			cantrips = 5
		}

		// Wizards start with 6 spells in spellbook
		spellsKnown = 6
		// +2 per level
		spellsKnown += (level - 1) * 2

		// ... other classes
	}

	return cantrips, spellsKnown
}
