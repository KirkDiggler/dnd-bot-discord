# Context for Claude - D&D Discord Bot

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

#### Known Issues (Fix as we touch the code)
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
make test-integration

# Run linter
make lint

# Build
make build

# Generate mocks (uses /home/kirk/go/bin/mockgen)
make generate-mocks

# Run locally (needs Redis and Discord token)
DISCORD_TOKEN=xxx REDIS_URL=redis://localhost:6379 ./bin/dnd-bot
```

#### Mock Generation
- mockgen is installed at `/home/kirk/go/bin/mockgen`
- The Makefile automatically adds this to PATH when running `make generate-mocks`
- If you see "mockgen not found", use the full path or run `make generate-mocks`

### Key Decisions Made
- Use Uber's gomock for mocking (go.uber.org/mock)
- Character status flow: Draft → Active (via FinalizeDraftCharacter)
- Proficiency choices come from both race and class
- Equipment choices can be nested (e.g., "choose a martial weapon")
- Empty draft characters (no name/race/class) are hidden from list

### Git Workflow
- PR branch: `comprehensive-tests` 
- Main branch has branch protection
- CI must pass before merge
- Using squash and merge strategy

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
- **Planning Mode**: Use well-defined issues to help future development
- **Context Management**: Compress context takes longer over time - consider fresh sessions for new features

### Current Work: Weapon Equipping (Issue #37) - MOSTLY COMPLETE
- **Branch**: `implement-weapon-equipping`
- **Goal**: Allow players to equip weapons from inventory for proper attack calculations
- **Status**: ✅ UI commands implemented, ✅ Attack calculations enhanced, ❌ Persistence missing

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

#### Known Issues (GitHub Issues needed):
1. **Equipment Persistence**: Equipped weapons don't persist to database (changes are memory-only)
   - Need to add character equipment save service method
   - Current limitation: Equipment changes work in session but reset on bot restart
2. **Weapon Autocomplete**: `/dnd character equip` requires manually typing weapon keys
   - Need autocomplete from character's weapon inventory
3. **Combat Integration**: Need to show equipped weapon name in attack messages

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

### Contact
- GitHub Issues: https://github.com/KirkDiggler/dnd-bot-discord/issues
- GitHub Project: https://github.com/users/KirkDiggler/projects/6
- This is Kirk's personal project for D&D sessions