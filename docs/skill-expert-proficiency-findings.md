# Skill-Expert Class Proficiency Testing Results

## Summary
All skill-expert classes (Rogue, Bard, and Ranger) are correctly loading their proficiencies from the D&D 5e API during character finalization.

## Key Findings

### Rogue
- **Weapon Proficiencies**: Simple weapons + longswords, rapiers, shortswords, hand crossbows
- **Armor Proficiencies**: Light armor only
- **Tool Proficiencies**: Thieves' tools (categorized as "Other" proficiency type)
- **Saving Throws**: DEX and INT
- **Skills**: 4 skill choices from class list
- **Special Features**: Expertise feature at level 1 (doubles proficiency bonus on 2 skills)

### Bard
- **Weapon Proficiencies**: Simple weapons + longswords, rapiers, shortswords, hand crossbows (same as Rogue)
- **Armor Proficiencies**: Light armor only
- **Tool Proficiencies**: None automatic, but gets to choose 3 musical instruments
- **Saving Throws**: DEX and CHA
- **Skills**: 3 skill choices from "any three"
- **Special**: Has 2 proficiency choice groups - one for skills, one for instruments

### Ranger
- **Weapon Proficiencies**: Simple weapons + martial weapons (full martial proficiency)
- **Armor Proficiencies**: Light armor, medium armor, and shields
- **Tool Proficiencies**: None
- **Saving Throws**: STR and DEX
- **Skills**: 3 skill choices from class list

## Implementation Notes

1. **Thieves' Tools**: The D&D 5e API categorizes thieves' tools as "Other" proficiency type, not "Tool" type. Our tests handle this correctly.

2. **Instrument Proficiencies**: Bards get a separate proficiency choice for 3 musical instruments. This appears in the API data but may need special handling during character creation to allow instrument selection.

3. **Weapon Categories**: All classes correctly receive their weapon proficiencies, both category-based (simple/martial) and specific weapons.

4. **Expertise**: The Rogue's expertise feature is correctly loaded as a character feature, not a proficiency.

## Next Steps
- Consider implementing instrument selection UI for Bards during character creation
- Verify that expertise properly doubles proficiency bonus in skill checks
- Test multiclassing scenarios with these classes