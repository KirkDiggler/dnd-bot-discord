# Class-Specific Character Creation Steps

## Current Status

### ✅ Implemented
- **Cleric**: Divine Domain selection, Knowledge domain skills/languages
- **Fighter**: Fighting Style selection
- **Ranger**: Favored Enemy, Natural Explorer
- **Wizard**: Cantrip selection (3), Spellbook selection (6 spells)

### 🔧 Needs Implementation

#### Barbarian
- No special choices at level 1
- Path selection at level 3

#### Bard
- **Cantrips**: Choose 2 cantrips from bard spell list
- **Spells Known**: Choose 4 1st-level spells
- **Expertise**: Choose 2 skills for expertise (double proficiency bonus)

#### Druid
- **Cantrips**: Choose 2 cantrips from druid spell list
- **Spells**: Prepare Wisdom modifier + level spells (auto-prepared, no selection needed)
- Druid Circle at level 2

#### Monk
- No special choices at level 1
- Monastic Tradition at level 3

#### Paladin
- No special choices at level 1
- Fighting Style at level 2
- Sacred Oath at level 3

#### Rogue
- **Expertise**: Choose 2 skills for expertise
- **Thieves' Cant**: Automatic language (no choice)
- Roguish Archetype at level 3

#### Sorcerer
- **Cantrips**: Choose 4 cantrips
- **Spells Known**: Choose 2 1st-level spells
- **Sorcerous Origin**: Choose subclass at level 1

#### Warlock
- **Cantrips**: Choose 2 cantrips
- **Spells Known**: Choose 2 1st-level spells
- **Otherworldly Patron**: Choose patron at level 1
- **Pact Magic**: Different from regular spellcasting

## Implementation Priority

1. **Spellcasters First** (most complex):
   - Bard (cantrips + spells + expertise)
   - Sorcerer (cantrips + spells + origin)
   - Warlock (cantrips + spells + patron)
   - Druid (cantrips only, spells are prepared)

2. **Skill-based Classes**:
   - Rogue (expertise)

3. **Simple Classes** (no level 1 choices):
   - Barbarian
   - Monk
   - Paladin

## Common Patterns

### Cantrip Selection
- Used by: Bard (2), Druid (2), Sorcerer (4), Warlock (2), Wizard (3)
- Should create reusable cantrip selection step

### Spell Selection
- Used by: Bard (4), Sorcerer (2), Warlock (2), Wizard (6)
- Different from prepared casters (Cleric, Druid, Paladin)

### Expertise
- Used by: Bard, Rogue
- Choose skills to double proficiency bonus

### Subclass at Level 1
- Cleric (Divine Domain)
- Sorcerer (Sorcerous Origin)  
- Warlock (Otherworldly Patron)