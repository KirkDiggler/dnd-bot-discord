# D&D Bot Development Roadmap

## Phase 0: Core Combat Systems 🔴 CURRENT PHASE
Essential infrastructure that all other features depend on.

### Goals
- Implement action economy (actions, bonus actions, reactions)
- Create extensible combat system
- Enable interrupt/trigger mechanics

### Issues
- [ ] #187 - Bonus Action System
- [ ] #201 - Reaction System
- [ ] Combat state machine improvements

## Phase 1: Complete Level 1 Features
Implement all Level 1 class features for martial and hybrid classes.

## Core Systems Needed

### 1. Bonus Action System (#187) 🔴 HIGH PRIORITY
- Required for: Monks, Rogues (two-weapon fighting), many spells
- Impacts combat flow significantly

### 2. Basic Spellcasting Framework (#196, #139) 🔴 HIGH PRIORITY
- Spell slots tracking
- Cantrips (at-will spells)
- Spell preparation (for prepared casters)
- Spell component tracking (V, S, M)

### 3. Reaction System 🔴 HIGH PRIORITY
- Required for: Protection fighting style, Shield spell, Counterspell
- Needs interrupt mechanism in combat

## Classes and Their Level 1 Features

### ✅ Barbarian (Mostly Complete)
- [x] Rage (#134 - merged)
- [x] Unarmored Defense (implemented)
- [ ] Rage deactivation conditions (#135)

### ⚠️ Fighter (Partial)
- [x] Fighting Style framework (implemented)
  - [x] Archery (implemented)
  - [x] Dueling (implemented)
  - [x] Great Weapon Fighting (#172 - needs damage reroll)
  - [ ] Defense (#174 - needs AC calculation)
  - [ ] Protection (#176, #179 - needs reaction system)
  - [x] Two-Weapon Fighting (implemented)
- [ ] Second Wind (bonus action heal)

### ⚠️ Ranger (Partial)
- [x] Favored Enemy (PR #154)
- [x] Natural Explorer (PR #154)
- [ ] Tracking bonuses UI (#163)

### ⚠️ Rogue (Partial)
- [x] Sneak Attack (PR #136)
- [ ] Thieves' Cant (flavor/RP feature)
- [ ] Expertise (double proficiency bonus on chosen skills)

### ⚠️ Monk (In Progress)
- [x] Martial Arts - DEX for attacks (PR #184)
- [ ] Martial Arts - 1d4 damage (#185)
- [ ] Martial Arts - Bonus action unarmed strike (#187)
- [ ] Monk weapons (#186)
- [x] Unarmored Defense (implemented)

### ❌ Cleric (Not Started)
- [ ] Spellcasting (wisdom-based, prepared)
- [ ] Divine Domain choice
- [ ] Domain spells
- [ ] Cantrips (3 at level 1)
- [ ] Ritual casting

### ❌ Wizard (Not Started)
- [ ] Spellcasting (intelligence-based, prepared)
- [ ] Spellbook (6 spells at level 1)
- [ ] Cantrips (3 at level 1)
- [ ] Ritual casting
- [ ] Arcane Recovery

### ❌ Warlock (Not Started)
- [ ] Pact Magic (special spell slots)
- [ ] Otherworldly Patron choice
- [ ] Patron features
- [ ] Cantrips (2 at level 1)

### ❌ Sorcerer (Not Started)
- [ ] Spellcasting (charisma-based, known spells)
- [ ] Sorcerous Origin choice
- [ ] Cantrips (4 at level 1)

### ❌ Paladin (Not Started)
- [ ] Divine Sense
- [ ] Lay on Hands (healing pool)

### ❌ Druid (Not Started)
- [ ] Druidic language
- [ ] Spellcasting (wisdom-based, prepared)
- [ ] Ritual casting
- [ ] Cantrips (2 at level 1)

### ❌ Bard (Not Started)
- [ ] Bardic Inspiration (bonus action, dice pool)
- [ ] Spellcasting (charisma-based, known spells)
- [ ] Ritual casting
- [ ] Cantrips (2 at level 1)

## Implementation Priority

### 🔴 Critical Path (Block other features)
1. **Bonus Action System** (#187) - Blocks many class features
2. **Basic Spellcasting** (#196, #139) - Blocks 7/12 classes
3. **Reaction System** - Blocks Protection style, many spells

### 🟡 High Priority (Core gameplay)
1. Complete Fighter features (Second Wind)
2. Complete Monk features (#185, #186)
3. Complete Rogue features (Expertise, Thieves' Cant)
4. Fix combat issues (#172, #174)

### 🟢 Medium Priority (Full experience)
1. Implement one spellcasting class (suggest Cleric - simplest)
2. Complete Barbarian (rage deactivation)
3. UI improvements (#163, #181)

### 🔵 Lower Priority (Can wait)
1. Remaining spellcasting classes
2. Paladin/Ranger spell features (they get spells at level 2)
3. Pure RP features (Druidic, Thieves' Cant)

## Suggested Sprint Plan

### Sprint 1: Core Systems
- [ ] Implement Bonus Action System (#187)
- [ ] Design Reaction System architecture
- [ ] Complete Monk Level 1 (#185, #186)

### Sprint 2: Combat Polish
- [ ] Fix Great Weapon Fighting (#172)
- [ ] Implement Defense fighting style (#174)
- [ ] Implement Protection with reactions (#176, #179)
- [ ] Fighter Second Wind

### Sprint 3: Basic Spellcasting
- [ ] Core spell system (#196, #139)
- [ ] Implement Cleric as first caster
- [ ] Cantrips and 1st level spells

### Sprint 4: Complete Remaining Martials
- [ ] Rogue Expertise
- [ ] Barbarian rage deactivation (#135)
- [ ] UI improvements for features

## Success Criteria for Phase 1
- [ ] All martial classes fully playable at level 1
- [ ] At least one spellcasting class implemented
- [ ] Combat supports actions, bonus actions, and reactions
- [ ] All implemented features have tests
- [ ] Documentation for players on class features