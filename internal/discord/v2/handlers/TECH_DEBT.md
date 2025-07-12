# Technical Debt - V2 Character Creation

## Issues to Address

### 1. UI Hints Pattern Violation (High Priority)
**Issue**: Cantrips and spell selection steps use UIHints, violating ADR-002
- Cantrips/spells use a two-stage interaction (show step → open selection UI)
- This causes Discord to create new windows instead of updating existing ones
- Other steps (race, class, abilities) correctly render UI directly

**Solution**: 
- Remove UIHints from wizard spell/cantrip steps
- Implement spell selection directly in `buildEnhancedStepResponse`
- Make it work like ability score selection (single-stage interaction)

**Files affected**:
- `internal/services/character/flow_builder_wizard.go`
- `internal/discord/v2/handlers/character_creation_enhanced.go`

### 2. Equipment Pack Not Displaying
**Issue**: User reports wizard equipment pack choice not showing
- API returns 3 choices correctly (weapon, focus, pack)
- Need to verify Discord UI displays all 3 steps

**Investigation needed**:
- Add logging to equipment step creation
- Verify Discord UI limits aren't hiding steps

### 3. Consistent Interaction Patterns
**Issue**: Mixed patterns for step interactions
- Some steps update in place (race, class, abilities)
- Others create sub-flows (spells, potentially equipment)

**Solution**: Standardize all steps to update in place

## ADR-002 Reference
The flow service should define WHAT steps exist, not HOW to render them.
Discord handlers own the UI implementation details.

## Discord Interaction Patterns

### Modal Submissions
Modal submissions cannot update the original message that triggered them.
- Use ephemeral response for modal submissions
- Regular button interactions can update messages
- This is a Discord API limitation, not a bug