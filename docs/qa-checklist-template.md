# QA Checklist Template

## Manual Verification Checklist

Copy this checklist to your PR description and check off items as you verify them:

### Combat Action Economy
- [ ] **Martial Arts Bonus Action**
  - [ ] Appears only after attacking with monk weapon or unarmed strike
  - [ ] Disappears after being used
  - [ ] Shows correct damage dice (d4 at level 1-4, d6 at 5-10, etc.)
  - [ ] Uses DEX modifier for attack and damage rolls
  - [ ] Combat log shows "(Martial Arts)" label

- [ ] **Two-Weapon Fighting Bonus Action**
  - [ ] Appears only after attacking with light melee weapon in main hand
  - [ ] Requires light weapon in off-hand
  - [ ] Shows correct weapon name in button
  - [ ] Damage doesn't add ability modifier (unless Two-Weapon Fighting style)
  - [ ] Combat log shows off-hand weapon name

- [ ] **Action Economy Display**
  - [ ] Action status shows ✅/❌ correctly
  - [ ] Bonus action status shows ✅/❌ correctly
  - [ ] Available bonus actions list updates dynamically
  - [ ] Turn reset clears all action states

### General Combat Flow
- [ ] **Turn Order**
  - [ ] Can only act on your turn
  - [ ] "Not Your Turn" message appears with current player name
  - [ ] Next Turn button advances correctly
  - [ ] New round message appears when cycling back

- [ ] **Attack Flow**
  - [ ] Target selection shows valid targets only
  - [ ] Attack results show hit/miss/critical correctly
  - [ ] Damage is applied to target HP
  - [ ] Defeated enemies are marked appropriately
  - [ ] Combat ends when all enemies defeated

### Abilities
- [ ] **Rage (Barbarian)**
  - [ ] Uses bonus action
  - [ ] Shows duration remaining
  - [ ] Adds +2 damage to melee attacks
  - [ ] Persists across bot restarts

- [ ] **Second Wind (Fighter)**
  - [ ] Uses action (not bonus action)
  - [ ] Heals 1d10 + fighter level
  - [ ] Shows healing amount in log
  - [ ] Once per short rest

### Edge Cases
- [ ] **Multiple Players**
  - [ ] Each player gets own "Get My Actions" button
  - [ ] Ephemeral messages are private
  - [ ] Shared combat message updates for all

- [ ] **Bot Restart**
  - [ ] Combat state persists
  - [ ] Character resources persist
  - [ ] Active effects persist

### Known Issues to Verify Fixed
- [ ] Dead monsters don't attack
- [ ] Rage marks bonus action as used
- [ ] Equipment persistence works

## How to Use This Checklist

1. Copy the entire checklist to your PR description
2. Test each item locally or in dev environment
3. Check off items that work correctly
4. Leave unchecked items that need fixing
5. Add notes for any unexpected behavior

Example PR description:
```markdown
## Description
Implements monk martial arts bonus action and two-weapon fighting

## QA Checklist
- [x] **Martial Arts Bonus Action**
  - [x] Appears only after attacking with monk weapon or unarmed strike
  - [x] Disappears after being used
  - [ ] Shows correct damage dice (needs fix - always showing d4)
  ...
```

## Automated vs Manual Testing

**Automated (Unit Tests)**
- Damage calculations
- Ability requirements
- State transitions

**Manual (This Checklist)**
- Discord UI interactions
- Message formatting
- Button states
- Multi-player scenarios
- Visual feedback

## Adding New Items

When implementing new features, add corresponding QA items:
1. What should happen (happy path)
2. What shouldn't happen (validations)
3. Edge cases specific to the feature
4. Integration with existing features