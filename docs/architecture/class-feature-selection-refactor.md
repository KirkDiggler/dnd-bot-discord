# Class Feature Selection Architecture Refactor

## Current Problems
1. Feature options (fighting styles, domains, etc.) are hardcoded in Discord handlers
2. No connection to D&D 5e API data
3. Cleric Divine Domain selection is completely missing
4. Violates separation of concerns - game logic in UI layer

## Proposed Architecture

### Data Flow
```
D&D 5e API → Domain Models → Character Service → Discord Handler
(source)      (structure)     (business logic)   (rendering only)
```

### Layer Responsibilities

#### 1. D&D 5e API Client
- Fetch class level data including feature choices
- Example: `/api/classes/fighter/levels/1` returns fighting style options

#### 2. Rulebook Domain (`/internal/domain/rulebook/dnd5e`)
All D&D 5e specific types belong here:

```go
// class.go - Extend existing Class struct
type Class struct {
    // ... existing fields ...
    LevelFeatures map[int][]*ClassLevelFeature `json:"level_features"`
}

// class_features.go - New file for feature-specific types
type ClassLevelFeature struct {
    Level       int                `json:"level"`
    Feature     *CharacterFeature  `json:"feature"`
    Choices     []*FeatureChoice   `json:"choices"`
}

type FeatureChoice struct {
    Type        string             `json:"type"` // "fighting_style", "divine_domain", etc.
    Choose      int                `json:"choose"`
    Options     []*FeatureOption   `json:"options"`
}

type FeatureOption struct {
    Key         string             `json:"key"`
    Name        string             `json:"name"`
    Description string             `json:"description"`
    // For domains, could include domain spells
    Metadata    map[string]any     `json:"metadata"`
}

// fighting_styles.go - Fighting style definitions
type FightingStyle struct {
    Key         string `json:"key"`
    Name        string `json:"name"`
    Description string `json:"description"`
    Classes     []string `json:"classes"` // ["fighter", "ranger", "paladin"]
}

// divine_domains.go - Divine domain definitions  
type DivineDomain struct {
    Key          string   `json:"key"`
    Name         string   `json:"name"`
    Description  string   `json:"description"`
    DomainSpells []string `json:"domain_spells"` // Spell keys by level
}
```

**Key Principle**: The rulebook is the single source of truth for all D&D 5e game mechanics

#### 3. Character Service
- Query domain for required feature selections
- Validate selections
- Store selections in character features with metadata

#### 4. Discord Handler
- Only responsible for rendering
- Gets options from service/domain
- Presents them in Discord UI format
- No hardcoded game data

## Implementation Progress

### ✅ Completed
1. **Domain Models** - Created proper types in rulebook domain:
   - `feature_choices.go` - Core types for feature selection system
   - `fighting_styles.go` - All fighting styles with class restrictions
   - `divine_domains.go` - All PHB divine domains with domain spells
   - `ranger_features.go` - Favored enemy and natural explorer choices
   - `class_feature_registry.go` - Registry to get choices by class/level

### 🚧 Next Steps
1. **Update Character Service**
   - Use `GetPendingFeatureChoices` to check what needs selection
   - Store selections in character feature metadata
   
2. **Refactor Discord Handler**
   - Remove hardcoded options
   - Use rulebook domain for all feature data
   - Handler only formats for Discord UI

3. **Testing**
   - Integration test for Fighter fighting style
   - Verify Cleric divine domain works
   - Test all existing classes still function

## Implementation Plan

### Phase 1: Fighter (Proof of Concept)
1. **API Client Enhancement**
   - Add method to fetch class level features
   - Parse fighting style options from API

2. **Domain Model Updates**
   - Extend Class struct with level features
   - Create FeatureChoice models

3. **Service Layer**
   - Add method to get pending feature choices
   - Validate and store selections

4. **Handler Refactor**
   - Remove hardcoded fighting styles
   - Use domain data for options

5. **Testing**
   - Integration test full flow
   - Verify Fighter creation works end-to-end

### Phase 2: Apply Pattern
Once Fighter works:
1. **Ranger** - Two selections (favored enemy, natural explorer)
2. **Cleric** - Divine Domain (includes domain spells)
3. **Future** - Sorcerer origin, Warlock patron (when implemented)

### Phase 3: Advanced Features
- Domain spells automatically granted
- Feature prerequisites/restrictions
- Multi-level features (e.g., Wizard school at 2)

## Benefits
1. **Maintainability** - Changes to D&D rules only require API updates
2. **Extensibility** - Easy to add new classes/features
3. **Correctness** - Single source of truth (API)
4. **Testing** - Each layer can be tested independently

## Risks & Mitigation
1. **API Availability** - Cache responses, have fallback data
2. **Breaking Changes** - Version API calls
3. **Performance** - Cache class data after first fetch
4. **Complexity** - Start simple with Fighter, iterate

## Success Criteria
1. Fighter can select fighting style from API data
2. Selection stored properly in character
3. No hardcoded options in handler
4. All existing tests pass
5. Can extend pattern to other classes

## Next Steps
1. Create GitHub issue for tracking
2. Implement Fighter vertical slice
3. Get PR reviewed and merged
4. Apply to remaining classes