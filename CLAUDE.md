# Context for Claude - D&D Discord Bot

## Important Working Principles

### Clarity Over Agreement
- **Don't say "You're right!" when still confused** - This creates false confidence and wastes time
- **Ask clarifying questions when understanding is cloudy** - Better to pause and clarify than charge ahead with wrong assumptions
- **Be honest about uncertainty** - Say "I'm not entirely clear on..." instead of agreeing while misunderstanding
- **Gain confidence in assumptions** - Test understanding by asking specific scenario questions like:
  - "Can you walk me through the exact scenario?"
  - "What specific message should be updating?"
  - "Is this about X or Y?"

## Session Summary - June 17, 2025

### Recent Major Fixes
1. **Character Creation JSON Unmarshaling Bug** - Fixed polymorphic option types from D&D API by implementing custom UnmarshalJSON methods
2. **Character List UI Issues** - Added show buttons, progress indicators, and filtering for empty drafts
3. **Critical Mutex Copying Bug** - Added Character.Clone() method to prevent copying sync.Mutex
4. **GitHub Actions CI/CD** - Set up comprehensive testing pipeline with Go 1.24 and golangci-lint v2.1.6
5. **Character Disappearing Bug (June 18, 2025)** - Fixed Equipment interface JSON marshaling in Redis repository by implementing custom marshaling for equipment data

### Code Standards & Patterns

#### Testing Strategy
- **Write tests BEFORE fixing bugs** - This ensures the bug is understood and doesn't regress
- Use table-driven tests for Go
- Integration tests use real Redis when available
- Mock external dependencies (D&D API, Discord)

#### Code Quality
- Using golangci-lint v2.1.6 with Google-style base configuration
- **Clean up files as we touch them** - Don't fix all lint issues at once
- Run `make lint` before committing
- Essential linters: errcheck, govet, staticcheck, gosec

#### Error Handling
- Always check errors from Discord API calls (InteractionRespond, etc.)
- Use custom error types from internal/errors package
- Log errors that can't be returned to user

### Project Structure

```
/cmd/bot          - Main application entry point
/internal
  /entities       - Domain models (Character, Race, Class, etc.)
  /handlers       - Discord interaction handlers
  /services       - Business logic layer
  /repositories   - Data persistence layer
  /clients        - External API clients (D&D 5e API)
  /testutils      - Test fixtures and utilities
```

### Character Creation Flow
1. Race → 2. Class → 3. Abilities → 4. Proficiencies → 5. Equipment → 6. Features (TODO) → 7. Name

### Current State

#### What's Working
- Character creation with full equipment/proficiency selection
- Character list with filtering and quick actions
- Continue button for resuming draft characters
- Redis persistence for characters and sessions
- Comprehensive GitHub Actions CI
- Dungeon service with state machine for room progression
- Dynamic monster selection from D&D 5e API with CR filtering
- Loot service with treasure generation
- Session metadata persistence for dungeon state

#### Recent Improvements (June 18, 2025)
- **Dungeon System Architecture**:
  - Created clean service layer for dungeon management
  - Implemented state machine (awaiting_party → room_ready → in_progress → room_cleared → complete/failed)
  - Moved all business logic from handlers to services
  - Fixed session metadata persistence with new SaveSession method
- **D&D 5e API Integration**:
  - Implemented stub methods in dnd5e client:
    - ListEquipment() - fetches all equipment from API
    - ListClassFeatures() - uses GetClassLevel to avoid N+1 queries
    - ListMonstersByCR() - temporarily hardcoded until API supports filtering
  - Fixed AbilityScore.AddBonus() to correctly apply racial bonuses to score
  - Monster service now fetches from API with fallback to hardcoded
  - Loot service integrates with equipment API
- **Service Provider Pattern**:
  - Added MonsterService and LootService to provider
  - Services properly inject dependencies

#### Recent Fixes (June 19, 2025)
- **Attack Flow Implementation (Issue #27)**:
  - Implemented automatic weapon selection from character's equipped items
  - Added target selection UI showing all combatants with HP/AC
  - Execute attacks using character's Attack() method
  - Added comprehensive logging throughout attack flow
  - Fixed CustomID parsing for select_target (parts[3] not parts[2])
  - Simplified attack flow to execute immediately after target selection (Discord 100 char limit)
- **Dungeon Encounter Regression Fixes**:
  - Fixed "only the DM can create encounters" error by checking sessionType="dungeon"
  - Fixed SaveSession vs UpdateSession to persist bot as DM
  - Fixed "no active encounter in session" by returning nil,nil instead of error
  - Added comprehensive tests to prevent regression
- **Player Not in Encounter Fix**:
  - Root cause: Player had no CharacterID set in session (hadn't selected character)
  - Fixed by adding validation in enter_room to require character selection
  - Fixed join handler to handle users already in session
  - Added comprehensive logging to track session member states
- **Workflow Enforcement**:
  - Players must select character before entering dungeon rooms
  - Join handler now handles both new joins and character selection for existing members
  - Changed button label from "Join Party" to "Select Character" for clarity
  - Added comprehensive unit tests for workflow validation

#### Architectural Learning Process (Event-Driven Exploration)

**Philosophy**: Instead of jumping straight into a new architecture, we implement features while thinking about how they would work in an event-driven system. This allows us to:
- Learn what we actually need vs what we think we need
- Validate the architecture idea through real implementation experience  
- Make informed decisions based on concrete examples
- Avoid building something that doesn't fit our actual requirements

**Process**:
1. Implement features normally but think about event-driven integration
2. Create issues with `prototype` and `event-driven` labels documenting learnings
3. Review after implementation to see if event-driven would be easier
4. Either continue developing the idea or drop it based on evidence

**Example**: When implementing Monk Martial Arts:
- Current: Direct method calls and state modifications
- Event-Driven: Would emit "AttackComplete" event, Martial Arts feature listens and enables bonus action
- Document the friction points and benefits we discover

## Architecture Insights

### Character Creation Step Registry (ADR-001)
The flow service defines WHAT steps exist, Discord defines HOW to render them:
```go
// Rulebook declares step types
var dnd5e.CreationStepTypes = []character.CreationStepType{
    character.StepTypeRaceSelection,
    character.StepTypeClassSelection,
    // ... etc
}

// Discord validates complete coverage at startup
for _, stepType := range dnd5e.CreationStepTypes {
    if _, ok := h.stepHandlers[stepType]; !ok {
        return nil, fmt.Errorf("no handler for step type: %s", stepType)
    }
}
```

### V2 Handler System
- Domain-based routing: `/dnd create character` → "create" domain → CharacterCreationHandler
- Custom ID format: `domain:action:target:args...`
- Modular handlers registered in routers
- See `/internal/discord/v2/README.md` for details

### RPG Toolkit Vision
This bot is a prototype for a universal rpg-toolkit:
- Clean boundaries enable future extraction
- Discord bot serves as test bed for patterns
- Architecture decisions have higher stakes
- Located at `/internal/adapters/rpgtoolkit`

### Pagination Pattern for Discord's 25-Option Limit
When step options exceed Discord's select menu limit:
```go
// Step declares it needs pagination
UIHints: &character.StepUIHints{
    RequiresPagination: true,
    Actions: []character.StepAction{{
        ID:    "open_spell_selection",
        Label: "Browse Spells",
        Style: "primary",
    }},
}

// Handler implements paginated UI
const itemsPerPage = 10
response := h.buildSpellSelectionPage(char, spells, page, itemsPerPage)
```
### Known Issues (Fix as we touch the code)
- ~10 unchecked errors (errcheck)
- Shadow variable declarations
- Some unused parameters
- Using math/rand instead of crypto/rand in places

#### Next Priorities
1. Continue monitoring logs to discover edge cases in dungeon workflow
2. Add more strategic logging points as issues arise
3. Implement dungeon repository for Redis persistence
4. Test dungeon flow end-to-end with Discord integration
5. Implement features selection step (SelectFeaturesStep)
6. Create end-to-end Discord bot tests
7. Add more Redis integration tests

### Development Commands

```bash
# Run tests
make test
make test-integration  # Uses Redis DB 15 to protect dev data

# Run linter
make lint

# Format code
make fmt

# Pre-commit checks (ALWAYS run before committing!)
make pre-commit

# Build
make build

# Generate mocks (uses /home/kirk/go/bin/mockgen)
make generate-mocks

# Run locally (needs Redis and Discord token)
DISCORD_TOKEN=xxx REDIS_URL=redis://localhost:6379 ./bin/dnd-bot
```

### PR Review Workflow
1. **Check for inline feedback**: `gh pr view --web` or `gh pr checks`
2. **Address all comments** before requesting re-review
3. **Run pre-commit**: `make pre-commit` (never use --no-verify)
4. **Update PR description** if scope changes
5. **Use conventional commits**: feat:, fix:, refactor:, docs:, test:
6. **Add QA Checklist**: Copy from `docs/qa-checklist-template.md` and check off manual verifications

#### Pre-Commit Workflow
**IMPORTANT**: Always run `make pre-commit` before committing code. This will:
1. Format all Go code with `go fmt`
2. Tidy module dependencies with `go mod tidy`
3. Run the linter to catch common issues
4. Run unit tests to ensure nothing is broken

If any step fails, fix the issues before committing.

#### Mock Generation
- mockgen is installed at `/home/kirk/go/bin/mockgen`
- The Makefile automatically adds this to PATH when running `make generate-mocks`
- If you see "mockgen not found", use the full path or run `make generate-mocks`

#### Mock Library Migration (July 2, 2025)
**Issue**: Both github.com/golang/mock and go.uber.org/mock in go.mod causing conflicts
**Solution**: 
1. Use only go.uber.org/mock (Uber's maintained fork)
2. All mock files must import "go.uber.org/mock/gomock" not "github.com/golang/mock/gomock"
3. Run `go mod edit -droprequire github.com/golang/mock` to remove old library
4. Regenerate all mocks with: `make generate-mocks`
5. Update any manual imports in test files
**Root Cause**: The golang/mock library is deprecated, Uber's fork is the maintained version

### Key Decisions Made
- Use Uber's gomock for mocking (go.uber.org/mock)
- Character status flow: Draft → Active (via FinalizeDraftCharacter)
- Proficiency choices come from both race and class
- Equipment choices can be nested (e.g., "choose a martial weapon")
- Empty draft characters (no name/race/class) are hidden from list

### Git Workflow
- Main branch has branch protection
- CI must pass before merge
- Using squash and merge strategy

#### Issue and PR Best Practices
1. **Before Starting Work**:
   - Check if an issue exists, if not create one
   - Add issue to project board: `gh issue edit <number> --add-project "Co Op Dungeon Adventure Start"`
   - Assign yourself: `gh issue edit <number> --assignee @me`

2. **Branch Naming**:
   - Use descriptive branch names: `fix-<issue-number>-<short-description>`
   - Example: `fix-59-dead-monsters-attacking`

3. **Commit Messages**:
   - Reference issues: `Fix dead monsters attacking (#59)`
   - Use conventional commits: `feat:`, `fix:`, `docs:`, `refactor:`, `test:`

4. **Pull Request**:
   - **ALWAYS include `Fixes #<issue-number>`** in PR description
   - This automatically closes the issue when PR is merged
   - Example: `Fixes #59` or `Partially addresses #74`

5. **Quick Commands**:
   ```bash
   # Add issue to project
   gh issue edit <number> --add-project "Co Op Dungeon Adventure Start"
   
   # Create PR with fixes
   gh pr create --title "Fix: Dead monsters continue to attack" --body "Fixes #59"
   ```

#### IMPORTANT: PR Merge Workflow
- **ALWAYS merge PRs before starting new features**
- After creating a PR, ensure it's merged to main before moving to the next feature
- This prevents regression bugs and merge conflicts
- Workflow: Create PR → Merge PR → Update branch → Start next feature
- If multiple PRs exist, merge them in order of creation to avoid conflicts

#### Merge vs Rebase Philosophy
- **Prefer merge over rebase** - "Rewriting history is something I cannot get back. Understanding complex merges is at least possible."
- History preservation is more valuable than a clean linear history
- Complex merges can be understood with tools and patience
- Lost history from rebase cannot be recovered
- Since we squash-and-merge PRs, the main branch stays clean anyway

### Debugging Tips
- Check `go.mod` for dependency versions if something seems off
- The D&D API returns polymorphic JSON - always check the actual response
- Discord interactions must be acknowledged within 3 seconds
- Redis keys follow pattern: `character:{id}`, `session:{id}`, etc.

### Project Organization (June 19, 2025)
- **GitHub Project #6**: Main project board for organizing all issues
  - Uses Kanban board with custom statuses: Ideas → Investigate → Backlog → Ready → In Progress → Review → Done
- **Planning Mode**: Use well-defined issues to help future development
- **Context Management**: Compress context takes longer over time - consider fresh sessions for new features

### Workspace Structure
The project lives in `/home/kirk/personal/` with related projects:
- **dnd-bot-discord**: Main Discord bot (current focus)
- **dnd5e-api**: D&D 5e API integration library
- **ronnied**: Dice game bot project

Proto files already exist in dnd-bot-discord:
- `character.proto`, `combat.proto`, `combat_streaming.proto`, `game.proto`

### Current Work: Enhanced Character Creation with V2 Handler System - IN PROGRESS
- **Branch**: `implement-wizard-spell-selection` (continuing work)
- **Goal**: Rich character creation experience with preview functionality
- **Status**: 
  - ✅ V2 handler system implemented (PR #295 merged)
  - ✅ Race preview functionality working
  - ✅ Rich character sheet display
  - ✅ Concurrent race data fetching from API
  - 🔧 Class preview next

#### What Was Implemented (July 8, 2025):
1. **V2 Handler System**: Modular Discord interaction handlers with domain routing
2. **Race Preview**: Select races to preview before confirming
3. **Rich Character Sheet**: Sections for race, class, abilities, proficiencies
4. **API Fix**: ListRaces() returns references only - now fetch full details concurrently
5. **Documentation**: Created `/internal/discord/v2/README.md` for v2 system

#### Key Learnings:
- Always use integration tests to verify API data flow
- CustomID format must be consistent: `domain:action:target:args...`
- Router domain must match Discord command name
- Component handlers need full pattern with domain prefix

### Previous Work: Weapon Equipping (Issue #37) - COMPLETE ✅
- **Branch**: `implement-weapon-equipping`
- **PR**: #44 - Ready for merge
- **Goal**: Allow players to equip weapons from inventory for proper attack calculations
- **Status**: ✅ UI commands implemented, ✅ Attack calculations enhanced, ✅ Persistence implemented

#### What's Implemented:
1. **Slash Commands**: `/dnd character equip`, `/dnd character unequip`, `/dnd character inventory`
2. **Attack Calculations**: Enhanced weapon attacks with proficiency bonus calculation
3. **Equipment System**: Character.Equip() method handles weapon slot management
4. **Proficiency Check**: HasWeaponProficiency() method checks weapon proficiencies

#### Key Files Modified:
- `internal/handlers/discord/dnd/character/weapon.go` - New weapon management UI handlers
- `internal/handlers/discord/handler.go` - Added weapon commands and routing
- `internal/entities/weapon.go` - Enhanced Attack() method with proficiency bonus
- `internal/entities/character.go` - Added HasWeaponProficiency() method

#### Completed Features:
1. **Equipment Persistence**: ✅ Fixed! Equipment changes now persist to Redis database
   - Added UpdateEquipment method to character service
   - Equip/unequip commands save changes immediately

#### Remaining Enhancements (Future PRs):
1. **Weapon Autocomplete**: `/dnd character equip` requires manually typing weapon keys
   - Need autocomplete from character's weapon inventory
2. **Combat Integration**: Need to show equipped weapon name in attack messages
   - Attack messages should display weapon name and attack bonuses

#### D&D 5e Rules Implemented:
- **Proficiency Bonus**: +2 at level 1-4, +3 at 5-8, +4 at 9-12, etc.
- **Attack Bonus**: Ability modifier + proficiency bonus (if proficient)
- **Damage Bonus**: Only ability modifier applies to damage
- **Weapon Types**: Melee uses STR, Ranged uses DEX
- **Two-Handed Weapons**: Use TwoHandedDamage if available

### Next Session GitHub Issues to Create:

#### High Priority Issues:
1. **Equipment Persistence Service** (Blocks production use)
   - Title: "Add character equipment persistence to database"
   - Description: Equipped weapons currently only persist in memory, reset on bot restart
   - Tasks: Add CharacterService.SaveEquipment() method, update repository layer
   - Acceptance: Equipment changes persist across bot restarts

2. **Combat Message Enhancement** (Improves UX)
   - Title: "Show equipped weapon name in attack messages"
   - Description: Attack messages should display weapon name and attack bonuses
   - Current: "Grunk attacks goblin for 8 damage"
   - Desired: "Grunk attacks goblin with Longsword (+5 to hit) for 8 slashing damage"

#### Medium Priority Issues:
3. **Weapon Autocomplete** (Quality of life)
   - Title: "Add weapon autocomplete to /dnd character equip command"
   - Description: Users should see dropdown of available weapons from inventory
   - Implementation: Discord autocomplete with character weapon inventory

4. **Equipment Quick Actions** (Quality of life)
   - Title: "Add equip/unequip buttons to character sheet"
   - Description: Character sheet should have quick action buttons for equipment
   - Implementation: Add buttons to character show embed

#### Future Enhancement Issues:
5. **Armor Class Calculation** (Game mechanics)
   - Title: "Implement proper armor AC calculation with equipped gear"
   - Description: AC should update based on equipped armor and DEX modifier limits

6. **Two-Weapon Fighting** (Game mechanics)
   - Title: "Implement two-weapon fighting bonus action attacks"
   - Description: Characters with two light weapons should get bonus action attack

### Session Summary (June 19, 2025):
**Weapon Equipping Implementation - SUCCESSFUL** ✅
- **Time**: ~2 hours of development
- **Lines Added**: ~700 lines of new code + tests
- **Features**: 3 new slash commands, enhanced attack calculations, comprehensive test suite
- **Blockers**: Equipment persistence (memory-only currently)
- **Ready for**: User testing in development environment

**Key Learning**: The Character.Attack() method was already well-architected to use equipped weapons, making this implementation smoother than expected. The main missing piece was the UI layer and proper attack bonus calculations.

### Interactive Character Sheet System (Issues #37-#43)

#### Project Overview
**GitHub Project Board**: https://github.com/users/KirkDiggler/projects/6
- Using "Ideas" and "Investigate" statuses for planning
- Phased approach: Discord bot → gRPC API + protos → React visualization

#### Design Vision
The interactive character sheet aims to provide a rich, real-time D&D experience through Discord:

1. **Ephemeral Messages**: Character sheets use ephemeral responses to reduce channel clutter
2. **Equipment Interaction**: Players can equip/unequip items directly from their character sheet
3. **Encounter Actions**: During combat, character sheets show available actions based on equipped weapons and abilities
4. **Real-time Updates**: Sheet updates reflect changes immediately (HP, conditions, equipment)

#### Implementation Issues Created (June 19, 2025)

**Issue #37: Interactive Character Sheet - Main View** ✅
- Ephemeral character sheet with complete stats display
- Equipment section with equip/unequip buttons
- Refresh button for real-time updates

**Issue #38: Character Sheet Actions Menu**
- Action buttons based on encounter context
- Attack, cast spell, use item options
- Dynamic based on equipped items and abilities

**Issue #39: Quick Actions Bar**
- Frequently used actions at top of sheet
- Customizable shortcuts for spells/abilities
- Context-aware (combat vs exploration)

**Issue #40: Equipment Management UI**
- Interactive inventory with drag-and-drop feel
- Quick equip/unequip toggles
- Item details on hover/select

**Issue #41: Spell Management Interface**
- Spell slots tracking
- Prepared spells selection
- Quick cast buttons with targeting

**Issue #42: Character Conditions Display**
- Visual indicators for conditions (poisoned, prone, etc.)
- Temporary HP tracking
- Concentration management

**Issue #43: Character Sheet Customization**
- Player preferences for sheet layout
- Collapsible sections
- Theme/color preferences

#### Technical Architecture Plan

**Phase 1: Discord Bot Enhancement** (Current)
- Implement interactive components using Discord's button/select APIs
- Use ephemeral messages for character sheets
- Store UI state in Redis for performance

**Phase 2: gRPC API Layer**
- Extract character management into gRPC service
- Define protobuf schemas for all entities
- Enable multi-client support (Discord, web, mobile)

**Phase 3: React Visualization**
- Web dashboard for campaign management
- Real-time character sheet updates via WebSocket
- Advanced visualizations (3D dice, battle maps)

### Rogue Sneak Attack Implementation (June 27, 2025)

**Issue #136: Rogue Sneak Attack** ✅
- **Branch**: `implement-rogue-sneak-attack`
- **Status**: Implementation complete using TDD approach
- **Problem Solved**: Rogues needed their signature Sneak Attack ability for proper combat mechanics

#### What Was Implemented:
1. **Character Resources Update**:
   - Added `SneakAttackUsedThisTurn` bool to track once-per-turn limitation
   
2. **Sneak Attack Methods** (`character_sneak_attack.go`):
   - `CanSneakAttack()` - Checks weapon eligibility and combat conditions
   - `GetSneakAttackDice()` - Returns d6 count based on level (1d6 at level 1, +1d6 every 2 levels)
   - `ApplySneakAttack()` - Rolls damage and marks as used
   - `StartNewTurn()` - Resets the used flag

3. **Combat Integration**:
   - Modified `AttackInput` to include `HasAdvantage`, `HasDisadvantage`, and `AllyAdjacent` fields
   - Updated `PerformAttack` to check sneak attack eligibility and apply bonus damage
   - Added sneak attack info to combat log entries (shows as "🗡️ 3d6 Sneak Attack: 12")
   - Integrated turn reset in `NextTurn` when a new round starts

4. **D&D 5e Rules Implemented**:
   - Requires finesse or ranged weapon
   - Needs advantage OR ally adjacent to target
   - Once per turn (not per round)
   - Damage scales: 1d6 at level 1-2, 2d6 at 3-4, etc.
   - Critical hits double sneak attack dice
   - Cannot sneak attack if advantage and disadvantage cancel out

#### Key Design Decisions:
- **No Position Tracking**: Since the combat system doesn't have a grid, advantage and adjacency are passed as parameters
- **Persistence**: The `SneakAttackUsedThisTurn` flag is saved to character data via `UpdateEquipment` method
- **Turn Management**: All characters get `StartNewTurn()` called when a new round begins

#### Testing:
- Comprehensive unit tests for all sneak attack scenarios
- Tests for eligibility, damage scaling, critical hits, and turn resets
- All tests passing with proper TDD approach

### Session Summary - November 16, 2024

**Issue #98: Get My Actions Button Implementation** ✅
- **Branch**: `fix-98-get-my-actions`
- **Status**: Implementation complete, ready for PR
- **Problem Solved**: Multiplayer combat had shared button states causing "not your turn" errors
- **Solution**: Added ephemeral "Get My Actions" button for personalized combat interfaces

#### What Was Implemented:
1. **New handleMyActions method** in combat handler
   - Sends private ephemeral response to each player
   - Shows personalized action buttons based on turn state
   - Displays player's HP, AC, and current turn status

2. **UI Updates**:
   - Added "Get My Actions" button to all combat displays
   - Button uses green success style with 🎯 emoji
   - Integrated into enter_room, combat handler, and embed builders

3. **Code Quality**:
   - Fixed 4 errcheck linting issues
   - Fixed parameter shadowing (max → maxHP)
   - Added proper error logging for monster turn processing

#### Next Steps:
1. Create PR for issue #98
2. Test multiplayer combat with new personalized actions
3. Monitor for the duplicate combat UI issue (added TODO comment)
4. Consider adding more action types (potions, spells, abilities)

#### Known Issues to Monitor:
- **Duplicate Combat UI**: First player sometimes gets old style message, needs to enter room again
- **Turn Order Sync**: May need additional testing with 3+ players
- **Button Limits**: Discord's 25-button limit may affect future action additions

### Rage Ability Implementation (PR #134) - June 28, 2025

**Branch**: `implement-rage-damage-bonus`
**Status**: Implementation complete, persistence fixed

#### What Was Implemented:
1. **Rage Ability System**:
   - Toggle ability with 10-round duration
   - Limited uses per long rest (based on barbarian level)
   - Automatically deactivates after duration expires
   - Activated as a bonus action

2. **Rage Effects**:
   - +2 damage bonus to all melee weapon attacks
   - Resistance to physical damage (bludgeoning, piercing, slashing)
   - Effects tracked in `CharacterResources.ActiveEffects`
   - Proper integration with attack calculations

3. **Persistence Fix**:
   - Added `Resources` field to `CharacterData` struct in redis.go
   - Updated `toCharacterData()` method to include Resources (line 488)
   - Updated `fromCharacterData()` method to restore Resources (line 541)
   - Rage state now persists across bot restarts

4. **Combat Integration**:
   - Attack messages show "(includes effects)" when rage bonus applies
   - Damage calculations properly include rage bonus for all melee attacks
   - Works with main hand, off-hand, two-handed, and improvised weapons

#### Key Files Modified:
- `internal/repositories/characters/redis.go` - Fixed persistence layer
- `internal/entities/character.go` - Attack() method uses active effects
- `internal/entities/active_ability.go` - Rage ability definition
- `internal/services/ability_service.go` - Rage activation logic

#### Testing:
- Rage properly activates and adds damage bonus
- State persists to Redis and survives bot restarts
- Effects expire after 10 rounds as expected
- Build passes all tests with `make build`

### Session Summary - June 28, 2025 (Ranger Implementation)

**PR #154: Ranger Class Features and Weapon Category Proficiency** ✅ MERGED
- **Branch**: `implement-ranger-features`
- **Issues Fixed**: #148 (Ranger features), #149 (weapon category proficiency)
- **Status**: Successfully merged after resolving conflicts

#### What Was Implemented:
1. **Ranger Class Features**:
   - Favored Enemy feature (advantage on survival/intelligence checks)
   - Natural Explorer feature (benefits in favored terrain)
   - Comprehensive tests for Ranger character creation

2. **Weapon Category Proficiency Fix**:
   - Fixed `HasWeaponProficiency()` to check weapon categories (simple/martial)
   - Added `hasWeaponCategoryProficiency()` method with case-insensitive comparison
   - Rangers can now use all martial and simple weapons properly

3. **Critical Bug Fixes**:
   - **Root Cause Found**: `FinalizeDraftCharacter` wasn't adding class proficiencies
   - Fixed by adding proficiency loading during character finalization
   - Normalized weapon categories to lowercase in API serializer ("Martial" → "martial")
   - Fixed natural 1 attack rolls to show modifiers (1+6=7 instead of 1=1)

4. **Testing & Infrastructure**:
   - Updated Makefile to use Redis DB 15 for integration tests (protects dev data)
   - Added comprehensive Ranger weapon proficiency tests
   - Skipped one integration test temporarily (needs update for new proficiency loading)
   - Created test for natural 1 fix (awaiting dice roller interface #116)

#### Key Code Changes:
```go
// character.go - Added weapon category checking
func (c *Character) hasWeaponCategoryProficiency(weaponCategory string) bool {
    categoryMap := map[string]string{
        "simple":  "simple-weapons",
        "martial": "martial-weapons",
    }
    lowerCategory := strings.ToLower(weaponCategory)
    // ... check proficiencies
}

// service.go - Fixed proficiency loading
if !hasClassProficiencies {
    for _, prof := range char.Class.Proficiencies {
        proficiency, err := s.dndClient.GetProficiency(prof.Key)
        if err == nil && proficiency != nil {
            char.AddProficiency(proficiency)
        }
    }
}

// attack/result.go - Fixed natural 1 display
// Always add attack bonus to the roll
attackRoll += attackBonus
// Natural 1 is still an automatic miss, but we keep the modifiers for display
```

#### Debugging Journey:
- Started with manual testing showing "+2" attack bonus (missing proficiency)
- User redirected to TDD approach: "I am starting to think this adding logs and manually going through the chatr creation is not the most ideal way to solve this problem"
- Discovered proficiencies weren't showing on character sheet
- Root cause: draft finalization wasn't adding automatic proficiencies
- Fixed and verified: "ooooohhhhh muuuy gawd! I seen them!" (proficiencies appeared)

#### Merge Conflict Resolution:
- equipment.go (our branch) vs weapon.go (main branch) conflict
- Kept our more general equipment.go that handles both weapons and shields
- User noted: "i knoew this was going to bite ue. we changed it to equipment but had another branch open"

### Current TODOs:
1. **Create weapon category constants and normalization function** - Avoid hardcoded strings
2. **Fix integration test** - Update to handle automatic proficiency loading
3. **Implement finesse weapons** - Allow DEX for attack/damage rolls

### Known Issues:
- **Issue #155**: Human racial ability bonuses not being applied
- **Issue #116**: Dice roller interface needed for deterministic testing

### Next Priorities:
1. Fix human racial ability bonuses (#155) 
2. Implement finesse weapon property
3. Create weapon category constants
4. Continue with other class implementations

### Issue Management Guidelines

#### Label Usage
**IMPORTANT**: Always use appropriate labels when creating issues. This helps with organization and filtering.

**Common Labels**:
- **Class Labels**: `monk`, `ranger`, `fighter`, `rogue`, `barbarian`, etc.
- **Feature Labels**: `combat`, `enhancement`, `bug`, `documentation`
- **Priority Labels**: `high-priority`, `low-priority`
- **System Labels**: `ui`, `database`, `api`

**When Creating Issues**:
1. Add relevant class label if it's class-specific (e.g., `monk` for Martial Arts)
2. Add feature type label (e.g., `combat` for attack mechanics)
3. Add priority if applicable
4. Consider multiple labels when appropriate (e.g., `monk`, `combat`, `enhancement`)

**Example**:
```bash
gh issue create --title "Implement Ki points system" --label "monk,combat,enhancement"
```

### Rage Implementation and Event-Driven Architecture (PR #235) - July 1, 2025

**Context**: Started with restructure-rage-modifier branch, evolved into comprehensive architecture design.

#### What Was Fixed in PR #235 (MERGED)
1. **Rage Duration Bug**: Fixed counting to properly track rounds (not turns)
   - Rage lasts 10 rounds in D&D 5e, not 10 turns
   - With 3 combatants, 30 turns = 10 rounds
   - Added proper round calculation: `turnsElapsed / numCombatants`

2. **Effect Synchronization**: Fixed dual effect system persistence
   - New event system (`effects.StatusEffect`) 
   - Old UI system (`shared.ActiveEffect`)
   - Added `SyncEffects()` method to ensure both stay in sync

3. **Combat Messages**: Now show "(includes effects)" when rage damage applies

4. **Created Package READMEs** defining boundaries:
   - `/internal/services/ability/README.md` - Ruleset-agnostic service
   - `/internal/domain/rulebook/README.md` - ALL game-specific logic
   - `/internal/domain/events/README.md` - Event bus architecture
   - `/internal/domain/character/README.md` - Universal vs ruleset boundaries

#### Architectural Discovery Session (July 1, 2025)

**Key Insight**: "Things just do things and emit events. Everything else listens and modifies."

This led to documenting a comprehensive event-driven architecture:

1. **Unified Tracking System** (PR #239)
   - Composable trackers: Duration, Pool, Usage, Event-based
   - Separates "what tracks" from "what it does"
   - Example: Rage = DurationTracker(10 rounds) + UsageTracker(2/long rest) + Modifiers

2. **Event-Driven Modifiers**
   - Core entities emit events (OnAttackRoll, OnDamageRoll, etc.)
   - Modifiers listen and apply changes
   - No direct coupling between systems
   - Enables multiple rulesets to coexist

3. **Documentation Created**:
   - `/docs/architecture/unified-tracking-system.md`
   - `/docs/architecture/event-driven-modifiers.md`
   - `/docs/examples/modifier-examples.md` (5 concrete implementations)
   - `/docs/examples/event-flow-examples.md` (5 combat scenarios)

#### Issues Created
- **#237**: Refactor ability service to be ruleset-agnostic (technical debt from rage)
- **#236**: Show rage duration in combat actions UI
- **#238**: Complete README documentation for remaining packages
- **#240**: Implement unified tracking and modifier system (based on architecture docs)

#### Current Architecture State
- Ability service has hardcoded switch statement for abilities (technical debt)
- Event system partially implemented (rage uses it, other abilities don't)
- Good foundation but needs the refactor described in #237/#240

### Contact
- GitHub Issues: https://github.com/KirkDiggler/dnd-bot-discord/issues
- GitHub Project: https://github.com/users/KirkDiggler/projects/6
- This is Kirk's personal project for D&D sessions