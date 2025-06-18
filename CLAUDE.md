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

#### Known Issues (Fix as we touch the code)
- ~10 unchecked errors (errcheck)
- Shadow variable declarations
- Some unused parameters
- Using math/rand instead of crypto/rand in places

#### Next Priorities
1. Implement dungeon repository for Redis persistence
2. Test dungeon flow end-to-end with Discord integration
3. Implement features selection step (SelectFeaturesStep)
4. Create end-to-end Discord bot tests
5. Add more Redis integration tests

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

### Debugging Tips
- Check `go.mod` for dependency versions if something seems off
- The D&D API returns polymorphic JSON - always check the actual response
- Discord interactions must be acknowledged within 3 seconds
- Redis keys follow pattern: `character:{id}`, `session:{id}`, etc.

### Contact
- GitHub Issues: https://github.com/KirkDiggler/dnd-bot-discord/issues
- This is Kirk's personal project for D&D sessions