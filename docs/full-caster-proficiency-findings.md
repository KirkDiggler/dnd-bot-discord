# Full Caster Class Proficiency Testing Results

## Summary
All full caster classes (Wizard, Sorcerer, Warlock, Cleric, and Druid) are correctly loading their proficiencies from the D&D 5e API during character finalization.

## Key Findings

### Wizard
- **Weapon Proficiencies**: Specific weapons only - daggers, darts, slings, quarterstaffs, light crossbows
- **Armor Proficiencies**: None
- **Tool Proficiencies**: None
- **Saving Throws**: INT and WIS
- **Skills**: 2 choices from Arcana, History, Insight, Investigation, Medicine, Religion

### Sorcerer
- **Weapon Proficiencies**: Same as Wizard - daggers, darts, slings, quarterstaffs, light crossbows
- **Armor Proficiencies**: None
- **Tool Proficiencies**: None
- **Saving Throws**: CON and CHA
- **Skills**: 2 choices from Arcana, Deception, Insight, Intimidation, Persuasion, Religion

### Warlock
- **Weapon Proficiencies**: Simple weapons (category proficiency)
- **Armor Proficiencies**: Light armor
- **Tool Proficiencies**: None
- **Saving Throws**: WIS and CHA
- **Skills**: 2 choices from their skill list

### Cleric
- **Weapon Proficiencies**: Simple weapons (category proficiency)
- **Armor Proficiencies**: Light armor, medium armor, and shields
- **Tool Proficiencies**: None
- **Saving Throws**: WIS and CHA
- **Skills**: 2 choices from History, Insight, Medicine, Persuasion, Religion

### Druid
- **Weapon Proficiencies**: Specific weapons - clubs, daggers, darts, javelins, maces, quarterstaffs, scimitars, sickles, slings, spears
- **Armor Proficiencies**: Light armor, medium armor, shields (with non-metal restriction)
- **Tool Proficiencies**: Herbalism kit (categorized as "Other" proficiency type)
- **Saving Throws**: INT and WIS
- **Skills**: 2 choices from Arcana, Animal Handling, Insight, Medicine, Nature, Perception, Religion, Survival

## Implementation Notes

1. **Weapon Proficiencies**: 
   - Wizard, Sorcerer, and Druid get specific weapon proficiencies, not weapon categories
   - Warlock and Cleric get the "simple-weapons" category proficiency
   - This distinction is important for weapon proficiency checks

2. **Armor Restrictions**:
   - Wizard and Sorcerer have no armor proficiencies at all
   - Druids have a roleplay restriction against metal armor/shields (not enforced in code)

3. **Tool Proficiencies**:
   - Only Druid gets a tool proficiency (herbalism kit)
   - Like thieves' tools for Rogues, herbalism kit is categorized as "Other" proficiency type

4. **Skill Choices**:
   - All full casters choose 2 skills
   - Wizards have the most restricted list (6 INT-based skills)
   - Druids have the most diverse list (8 skills)

## Testing Results
All classes passed their proficiency tests:
- ✅ Weapon proficiencies load correctly
- ✅ Armor proficiencies load correctly (where applicable)
- ✅ Saving throw proficiencies load correctly
- ✅ Skill selections work properly
- ✅ Tool proficiencies load correctly (Druid herbalism kit)

## Future Considerations
- Implement material type checking for Druid armor restriction
- Consider domain-specific proficiencies for Clerics (some domains grant additional proficiencies)
- Warlock pact weapons might grant additional proficiencies at higher levels