# D&D Discord Bot - Project Roadmap

## Current Status
- ✅ Basic character creation
- ✅ Combat system with attacks
- ✅ Dungeon exploration
- ⚠️ Partial class features (Fighter, Barbarian, Ranger, Rogue)
- ❌ No spellcasting
- ❌ No complete action economy

## Phase 0: Core Combat Systems 🔴 CURRENT
**Goal**: Build the foundation for D&D 5e action economy

### Milestone Issues
1. **Enhanced Effects System** (#202) 🎯 FORCE MULTIPLIER
   - Foundation for spells, conditions, items, features
   - Advantage/disadvantage, rerolls, immunities
   - Save DCs, concentration, triggers
   - Makes everything else easier!

2. **Bonus Action System** (#187)
   - Required for: Monk attacks, Second Wind, many spells
   - Design: Action tracking per turn
   - UI: Separate bonus action buttons

3. **Reaction System** (#201)
   - Required for: Shield spell, Protection style, Opportunity attacks
   - Design: Interrupt/trigger mechanism
   - UI: Ephemeral prompts with timeout

4. **Combat State Machine** (#87)
   - Better turn/round tracking
   - Action economy enforcement
   - Combat event system

### Success Criteria
- [ ] Characters can take 1 action, 1 bonus action, 1 reaction per round
- [ ] UI clearly shows available actions
- [ ] Reactions can interrupt other turns
- [ ] Extensible system for adding new actions/reactions

## Phase 1: Complete Level 1 Features
**Goal**: All classes playable at Level 1 (except full casters)

### Milestone Issues
1. **Complete Martial Classes**
   - [ ] Monk: Martial Arts damage (#185), monk weapons (#186)
   - [ ] Fighter: Second Wind (#199)
   - [ ] Rogue: Expertise (#200)
   - [ ] Barbarian: Rage deactivation (#135)

2. **Complete Fighting Styles**
   - [ ] Defense AC bonus (#174)
   - [ ] Protection with reactions (#176, #179)
   - [ ] Great Weapon damage rerolls (#172)

3. **Basic Half-Caster Features**
   - [ ] Ranger spell preparation (at level 2)
   - [ ] Paladin Divine Sense, Lay on Hands

### Success Criteria
- [ ] All non-caster classes fully playable at level 1
- [ ] All fighting styles working correctly
- [ ] Clear documentation for players

## Why Effects System First?

The effects system turns complex features into data:

```go
// Instead of hard-coding spell logic:
func CastBless(caster *Character, targets []*Character) {
    // Complex logic for adding bonuses...
}

// We define spells as effects:
var BlessSpell = Effect{
    Name: "Bless",
    Modifiers: []Modifier{
        {Target: TargetAttackRoll, Value: "1d4"},
        {Target: TargetSavingThrow, Value: "1d4"},
    },
}
```

This means:
- **90% of spells** become effect definitions
- **Conditions** (poisoned, stunned) are just effects
- **Magic items** (+1 sword) are equipment with effects
- **Class features** (rage, bardic inspiration) are effects
- **Environmental** (difficult terrain) are area effects

## Phase 1.5: Basic Spellcasting 🧙‍♂️
**Goal**: Implement core spell system with one full caster

### Approach
1. Start with Cleric (simplest prepared caster)
2. Implement only cantrips + 1st level spells
3. Use effects system for spell effects
4. Focus on combat spells first

### Core Systems
- [ ] Spell slot tracking
- [ ] Spell preparation mechanics
- [ ] Spell components (V, S, M)
- [ ] Spell targeting UI
- [ ] Concentration tracking

### Success Criteria
- [ ] Cleric fully playable with spells
- [ ] At least 5 cantrips, 10 1st level spells
- [ ] Spell effects use status effect system
- [ ] Clear spell UI in combat

## Phase 2: Loot & XP System
**Goal**: Character progression and rewards

### Milestone Issues
- [ ] #197 - Experience points and leveling
- [ ] #198 - Loot system with item drops
- [ ] Character level up flow
- [ ] Item identification mechanics

## Phase 3: Levels 2-3 & Subclasses
**Goal**: Deeper character customization

### Milestone Issues
- [ ] All Level 2 features (Ki, Action Surge, etc.)
- [ ] All Level 3 features 
- [ ] Subclass selection system
- [ ] #188-#195 - Monk Ki and subclasses

## Future Phases

### Phase 4: Full Spellcasting
- Implement all caster classes
- Higher level spells (2nd-3rd)
- Ritual casting
- Spell scrolls/wands

### Phase 5: Advanced Combat
- Conditions and status effects
- Environmental hazards
- Cover and terrain
- Mounted combat

### Phase 6: Social & Exploration
- Skill challenges
- Social encounters
- Downtime activities
- Crafting system

## Development Principles

### 1. Iterative Delivery
- Each phase should deliver playable features
- Don't build everything at once
- Get feedback early and often

### 2. Foundation First
- Phase 0 blocks everything else
- Don't skip architectural work
- Test infrastructure thoroughly

### 3. Simplify Where Possible
- Start with PHB content only
- Implement core rules first
- Add complexity gradually

### 4. Leverage Existing Systems
- Use effects system for spells
- Reuse UI patterns
- Build on proven architecture

## Risk Mitigation

### Technical Risks
1. **Spellcasting Complexity**
   - Mitigation: Start with one class, limited spells
   - Prototype with cantrips first

2. **Reaction System Performance**
   - Mitigation: Timeout on decisions
   - Async handling of interrupts

3. **State Management**
   - Mitigation: Clear state machines
   - Comprehensive testing
   - Redis persistence

### Design Risks
1. **UI Complexity**
   - Mitigation: Progressive disclosure
   - Consistent patterns
   - User testing

2. **Rules Accuracy**
   - Mitigation: Focus on "good enough"
   - Document deviations
   - Community feedback

## Success Metrics
- **Phase 0**: 100% test coverage on combat systems
- **Phase 1**: All martial classes playable
- **Phase 1.5**: One full caster implemented
- **Phase 2**: Complete level 1-3 progression
- **Phase 3**: 3+ subclasses per class

## Timeline Estimates
- **Phase 0**: 2-3 weeks (critical path)
- **Phase 1**: 3-4 weeks  
- **Phase 1.5**: 4-6 weeks (spellcasting is complex)
- **Phase 2**: 2-3 weeks
- **Phase 3**: 4-6 weeks

Total to "feature complete" for levels 1-3: ~4 months