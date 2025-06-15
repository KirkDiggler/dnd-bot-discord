# Go Implementation Examples

## Handler Layer Examples

### Discord Command Handler
```go
package discord

import (
    "context"
    "fmt"
    
    "github.com/bwmarrin/discordgo"
    "github.com/yourusername/dnd-bot-discord/internal/services"
    "github.com/yourusername/dnd-bot-discord/internal/models"
)

type CommandHandler struct {
    characterSvc *services.CharacterService
    combatSvc    *services.CombatService
    messageSvc   *services.MessageService
}

func NewCommandHandler(cs *services.CharacterService, cbs *services.CombatService, ms *services.MessageService) *CommandHandler {
    return &CommandHandler{
        characterSvc: cs,
        combatSvc:    cbs,
        messageSvc:   ms,
    }
}

// Register all slash commands
func (h *CommandHandler) RegisterCommands(s *discordgo.Session) error {
    commands := []*discordgo.ApplicationCommand{
        {
            Name:        "character",
            Description: "Character management commands",
            Options: []*discordgo.ApplicationCommandOption{
                {
                    Type:        discordgo.ApplicationCommandOptionSubCommand,
                    Name:        "create",
                    Description: "Create a new character",
                },
                {
                    Type:        discordgo.ApplicationCommandOptionSubCommand,
                    Name:        "view",
                    Description: "View your character sheet",
                    Options: []*discordgo.ApplicationCommandOption{
                        {
                            Type:        discordgo.ApplicationCommandOptionString,
                            Name:        "name",
                            Description: "Character name",
                            Required:    false,
                        },
                    },
                },
            },
        },
        {
            Name:        "combat",
            Description: "Combat commands",
            Options: []*discordgo.ApplicationCommandOption{
                {
                    Type:        discordgo.ApplicationCommandOptionSubCommand,
                    Name:        "start",
                    Description: "Start a combat encounter",
                },
                {
                    Type:        discordgo.ApplicationCommandOptionSubCommand,
                    Name:        "join",
                    Description: "Join the active combat",
                },
            },
        },
        {
            Name:        "roll",
            Description: "Roll dice",
            Options: []*discordgo.ApplicationCommandOption{
                {
                    Type:        discordgo.ApplicationCommandOptionString,
                    Name:        "dice",
                    Description: "Dice expression (e.g., 1d20+5)",
                    Required:    true,
                },
            },
        },
    }
    
    for _, cmd := range commands {
        _, err := s.ApplicationCommandCreate(s.State.User.ID, "", cmd)
        if err != nil {
            return fmt.Errorf("failed to create command %s: %w", cmd.Name, err)
        }
    }
    
    return nil
}

// Handle character view command
func (h *CommandHandler) HandleCharacterView(s *discordgo.Session, i *discordgo.InteractionCreate) {
    // Defer response for processing
    s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
        Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
        Data: &discordgo.InteractionResponseData{
            Flags: discordgo.MessageFlagsEphemeral,
        },
    })
    
    ctx := context.Background()
    userID := i.Member.User.ID
    
    // Get character name from options if provided
    var characterName string
    if len(i.ApplicationCommandData().Options[0].Options) > 0 {
        characterName = i.ApplicationCommandData().Options[0].Options[0].StringValue()
    }
    
    // Fetch character
    var character *models.Character
    var err error
    
    if characterName != "" {
        character, err = h.characterSvc.GetCharacterByName(ctx, userID, characterName)
    } else {
        character, err = h.characterSvc.GetActiveCharacter(ctx, userID)
    }
    
    if err != nil {
        h.respondError(s, i, "Character not found")
        return
    }
    
    // Build character sheet embed
    embed := h.messageSvc.BuildCharacterSheet(character)
    components := h.messageSvc.BuildCharacterActions(character)
    
    // Send ephemeral response
    s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
        Embeds:     &[]*discordgo.MessageEmbed{embed},
        Components: &components,
    })
}

// Handle combat start command
func (h *CommandHandler) HandleCombatStart(s *discordgo.Session, i *discordgo.InteractionCreate) {
    ctx := context.Background()
    
    // Check if combat already exists in channel
    existingCombat, _ := h.combatSvc.GetActiveByChannel(ctx, i.ChannelID)
    if existingCombat != nil {
        h.respondError(s, i, "Combat already active in this channel")
        return
    }
    
    // Defer response
    s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
        Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
    })
    
    // Create combat session
    combat, err := h.combatSvc.InitiateCombat(ctx, &services.InitiateCombatRequest{
        ChannelID:   i.ChannelID,
        InitiatorID: i.Member.User.ID,
    })
    
    if err != nil {
        h.respondError(s, i, fmt.Sprintf("Failed to start combat: %v", err))
        return
    }
    
    // Build combat message
    embed := h.messageSvc.BuildCombatStatus(combat)
    components := h.messageSvc.BuildCombatControls(combat, false) // Not in turn yet
    
    // Send response
    msg, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
        Embeds:     &[]*discordgo.MessageEmbed{embed},
        Components: &components,
    })
    
    if err == nil && msg != nil {
        // Store message ID for updates
        combat.MessageID = msg.ID
        h.combatSvc.UpdateCombat(ctx, combat)
    }
}

func (h *CommandHandler) respondError(s *discordgo.Session, i *discordgo.InteractionCreate, message string) {
    s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
        Content: &message,
    })
}
```

### Button Interaction Handler
```go
package discord

import (
    "context"
    "strings"
    
    "github.com/bwmarrin/discordgo"
)

func (h *CommandHandler) HandleButtonInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
    customID := i.MessageComponentData().CustomID
    parts := strings.Split(customID, ":")
    
    if len(parts) < 2 {
        return
    }
    
    action := parts[0]
    
    switch action {
    case "combat_join":
        h.handleCombatJoin(s, i, parts[1])
    case "combat_attack":
        h.handleCombatAttack(s, i, parts[1])
    case "combat_move":
        h.handleCombatMove(s, i, parts[1])
    case "combat_endturn":
        h.handleCombatEndTurn(s, i, parts[1])
    case "character_rest":
        h.handleCharacterRest(s, i, parts[1])
    }
}

func (h *CommandHandler) handleCombatJoin(s *discordgo.Session, i *discordgo.InteractionCreate, sessionID string) {
    ctx := context.Background()
    userID := i.Member.User.ID
    
    // Defer ephemeral response
    s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
        Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
        Data: &discordgo.InteractionResponseData{
            Flags: discordgo.MessageFlagsEphemeral,
        },
    })
    
    // Get user's active character
    character, err := h.characterSvc.GetActiveCharacter(ctx, userID)
    if err != nil {
        h.respondError(s, i, "You need an active character to join combat")
        return
    }
    
    // Join combat
    combat, err := h.combatSvc.JoinCombat(ctx, sessionID, character)
    if err != nil {
        h.respondError(s, i, fmt.Sprintf("Failed to join combat: %v", err))
        return
    }
    
    // Update shared combat message
    h.updateCombatMessage(s, combat)
    
    // Send ephemeral response
    s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
        Content: &"You've joined the combat! Roll initiative when prompted.",
    })
}

func (h *CommandHandler) handleCombatAttack(s *discordgo.Session, i *discordgo.InteractionCreate, sessionID string) {
    ctx := context.Background()
    userID := i.Member.User.ID
    
    // Get combat state
    combat, err := h.combatSvc.GetCombat(ctx, sessionID)
    if err != nil {
        return
    }
    
    // Verify it's player's turn
    currentParticipant := combat.GetCurrentParticipant()
    if currentParticipant.UserID != userID {
        s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
            Type: discordgo.InteractionResponseChannelMessageWithSource,
            Data: &discordgo.InteractionResponseData{
                Content: "It's not your turn!",
                Flags:   discordgo.MessageFlagsEphemeral,
            },
        })
        return
    }
    
    // Show target selection
    options := h.messageSvc.BuildTargetSelectMenu(combat, currentParticipant.ID)
    
    s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
        Type: discordgo.InteractionResponseChannelMessageWithSource,
        Data: &discordgo.InteractionResponseData{
            Content:    "Select your target:",
            Components: []discordgo.MessageComponent{options},
            Flags:      discordgo.MessageFlagsEphemeral,
        },
    })
}

func (h *CommandHandler) updateCombatMessage(s *discordgo.Session, combat *models.CombatSession) {
    embed := h.messageSvc.BuildCombatStatus(combat)
    isCurrentTurn := false // Determine based on context
    components := h.messageSvc.BuildCombatControls(combat, isCurrentTurn)
    
    s.ChannelMessageEditComplex(&discordgo.MessageEdit{
        Channel:    combat.ChannelID,
        ID:         combat.MessageID,
        Embeds:     &[]*discordgo.MessageEmbed{embed},
        Components: &components,
    })
}
```

## Service Layer Examples

### Character Service
```go
package services

import (
    "context"
    "fmt"
    "time"
    
    "github.com/yourusername/dnd-bot-discord/internal/models"
    "github.com/yourusername/dnd-bot-discord/internal/repositories"
    "go.mongodb.org/mongo-driver/bson/primitive"
)

type CharacterService struct {
    repo      repositories.CharacterRepository
    cache     repositories.CacheRepository
    rulesRepo repositories.GameDataRepository
}

func NewCharacterService(repo repositories.CharacterRepository, cache repositories.CacheRepository, rules repositories.GameDataRepository) *CharacterService {
    return &CharacterService{
        repo:      repo,
        cache:     cache,
        rulesRepo: rules,
    }
}

func (s *CharacterService) CreateCharacter(ctx context.Context, req *CreateCharacterRequest) (*models.Character, error) {
    // Validate race, class, and background exist
    race, err := s.rulesRepo.GetRace(ctx, req.RaceID)
    if err != nil {
        return nil, fmt.Errorf("invalid race: %w", err)
    }
    
    class, err := s.rulesRepo.GetClass(ctx, req.ClassID)
    if err != nil {
        return nil, fmt.Errorf("invalid class: %w", err)
    }
    
    background, err := s.rulesRepo.GetBackground(ctx, req.BackgroundID)
    if err != nil {
        return nil, fmt.Errorf("invalid background: %w", err)
    }
    
    // Calculate derived stats
    character := &models.Character{
        ID:         primitive.NewObjectID(),
        UserID:     req.UserID,
        Name:       req.Name,
        Level:      1,
        Experience: 0,
        Race:       race,
        Class:      class,
        Background: background,
        Attributes: req.Attributes,
        Status:     models.CharacterStatusActive,
        CreatedAt:  time.Now(),
        UpdatedAt:  time.Now(),
    }
    
    // Apply racial bonuses
    character.Attributes.Strength += race.AbilityBonuses.Strength
    character.Attributes.Dexterity += race.AbilityBonuses.Dexterity
    character.Attributes.Constitution += race.AbilityBonuses.Constitution
    character.Attributes.Intelligence += race.AbilityBonuses.Intelligence
    character.Attributes.Wisdom += race.AbilityBonuses.Wisdom
    character.Attributes.Charisma += race.AbilityBonuses.Charisma
    
    // Calculate HP
    character.HP = models.HitPoints{
        Max:     class.HitDie + character.GetModifier("con"),
        Current: class.HitDie + character.GetModifier("con"),
        Temp:    0,
    }
    
    // Calculate combat stats
    character.CombatStats = models.CombatStats{
        AC:               10 + character.GetModifier("dex"), // Base AC
        InitiativeBonus:  character.GetModifier("dex"),
        Speed:           race.Speed,
        ProficiencyBonus: 2, // Level 1
    }
    
    // Set proficiencies from class and background
    character.SetInitialProficiencies()
    
    // Save to database
    if err := s.repo.Create(ctx, character); err != nil {
        return nil, fmt.Errorf("failed to create character: %w", err)
    }
    
    // Cache the character
    s.cache.SetCharacter(ctx, character)
    
    return character, nil
}

func (s *CharacterService) GetActiveCharacter(ctx context.Context, userID string) (*models.Character, error) {
    // Check cache first
    if char, err := s.cache.GetActiveCharacter(ctx, userID); err == nil {
        return char, nil
    }
    
    // Get from database
    characters, err := s.repo.FindByUserID(ctx, userID)
    if err != nil {
        return nil, err
    }
    
    // Find active character
    for _, char := range characters {
        if char.Status == models.CharacterStatusActive {
            // Cache and return
            s.cache.SetCharacter(ctx, char)
            s.cache.SetActiveCharacter(ctx, userID, char.ID.Hex())
            return char, nil
        }
    }
    
    return nil, fmt.Errorf("no active character found")
}

func (s *CharacterService) LevelUp(ctx context.Context, characterID string) (*models.Character, error) {
    character, err := s.repo.FindByID(ctx, characterID)
    if err != nil {
        return nil, err
    }
    
    // Check if character has enough XP
    requiredXP := models.GetRequiredXP(character.Level + 1)
    if character.Experience < requiredXP {
        return nil, fmt.Errorf("insufficient experience: need %d, have %d", requiredXP, character.Experience)
    }
    
    // Level up
    character.Level++
    
    // Roll for HP increase
    hpRoll := models.RollDice(fmt.Sprintf("1d%d", character.Class.HitDie))
    conMod := character.GetModifier("con")
    hpIncrease := hpRoll.Total + conMod
    if hpIncrease < 1 {
        hpIncrease = 1
    }
    
    character.HP.Max += hpIncrease
    character.HP.Current += hpIncrease
    
    // Update proficiency bonus
    character.CombatStats.ProficiencyBonus = models.GetProficiencyBonus(character.Level)
    
    // Check for class features at new level
    newFeatures := s.rulesRepo.GetClassFeatures(ctx, character.Class.ID, character.Level)
    character.Features = append(character.Features, newFeatures...)
    
    // Save changes
    character.UpdatedAt = time.Now()
    if err := s.repo.Update(ctx, character); err != nil {
        return nil, err
    }
    
    // Invalidate cache
    s.cache.InvalidateCharacter(ctx, characterID)
    
    return character, nil
}
```

### Combat Service
```go
package services

import (
    "context"
    "fmt"
    "sort"
    "time"
    
    "github.com/google/uuid"
    "github.com/yourusername/dnd-bot-discord/internal/models"
    "github.com/yourusername/dnd-bot-discord/internal/repositories"
)

type CombatService struct {
    combatRepo repositories.CombatRepository
    charRepo   repositories.CharacterRepository
    diceService *DiceService
    events     chan CombatEvent
}

func NewCombatService(cr repositories.CombatRepository, chr repositories.CharacterRepository, ds *DiceService) *CombatService {
    return &CombatService{
        combatRepo:  cr,
        charRepo:    chr,
        diceService: ds,
        events:      make(chan CombatEvent, 100),
    }
}

func (s *CombatService) InitiateCombat(ctx context.Context, req *InitiateCombatRequest) (*models.CombatSession, error) {
    combat := &models.CombatSession{
        ID:           uuid.New().String(),
        ChannelID:    req.ChannelID,
        State:        models.CombatStateInitiating,
        Participants: make(map[string]*models.CombatParticipant),
        Round:        0,
        CreatedAt:    time.Now(),
    }
    
    // Initialize map if provided
    if req.MapWidth > 0 && req.MapHeight > 0 {
        combat.Map = &models.BattleMap{
            Width:    req.MapWidth,
            Height:   req.MapHeight,
            GridSize: 5, // Standard 5-foot squares
        }
    }
    
    // Save to Redis
    if err := s.combatRepo.CreateSession(ctx, combat); err != nil {
        return nil, err
    }
    
    // Set channel mapping
    s.combatRepo.SetChannelSession(ctx, req.ChannelID, combat.ID)
    
    // Emit event
    s.emitEvent(CombatEvent{
        Type:      EventCombatStarted,
        SessionID: combat.ID,
        Combat:    combat,
    })
    
    return combat, nil
}

func (s *CombatService) JoinCombat(ctx context.Context, sessionID string, character *models.Character) (*models.CombatSession, error) {
    combat, err := s.combatRepo.GetSession(ctx, sessionID)
    if err != nil {
        return nil, err
    }
    
    if combat.State != models.CombatStateInitiating && combat.State != models.CombatStateRollingInitiative {
        return nil, fmt.Errorf("combat has already started")
    }
    
    // Create participant
    participant := &models.CombatParticipant{
        ID:          uuid.New().String(),
        CharacterID: character.ID.Hex(),
        UserID:      character.UserID,
        Name:        character.Name,
        CurrentHP:   character.HP.Current,
        MaxHP:       character.HP.Max,
        AC:          character.CombatStats.AC,
        Initiative:  -1, // Not rolled yet
        Position:    s.findStartingPosition(combat),
    }
    
    combat.Participants[participant.ID] = participant
    combat.State = models.CombatStateRollingInitiative
    
    // Save changes
    if err := s.combatRepo.UpdateSession(ctx, combat); err != nil {
        return nil, err
    }
    
    // Emit event
    s.emitEvent(CombatEvent{
        Type:        EventParticipantJoined,
        SessionID:   combat.ID,
        Participant: participant,
    })
    
    return combat, nil
}

func (s *CombatService) RollInitiative(ctx context.Context, sessionID, participantID string) (*models.InitiativeResult, error) {
    combat, err := s.combatRepo.GetSession(ctx, sessionID)
    if err != nil {
        return nil, err
    }
    
    participant, exists := combat.Participants[participantID]
    if !exists {
        return nil, fmt.Errorf("participant not found")
    }
    
    if participant.Initiative >= 0 {
        return nil, fmt.Errorf("initiative already rolled")
    }
    
    // Get character for initiative bonus
    character, err := s.charRepo.FindByID(ctx, participant.CharacterID)
    if err != nil {
        return nil, err
    }
    
    // Roll initiative
    roll := s.diceService.Roll("1d20")
    total := roll.Total + character.CombatStats.InitiativeBonus
    
    participant.Initiative = total
    
    // Check if all participants have rolled
    allRolled := true
    for _, p := range combat.Participants {
        if p.Initiative < 0 {
            allRolled = false
            break
        }
    }
    
    if allRolled {
        // Sort participants by initiative
        combat.InitiativeOrder = s.sortByInitiative(combat.Participants)
        combat.State = models.CombatStateActive
        combat.CurrentTurn = 0
        combat.Round = 1
    }
    
    // Save changes
    if err := s.combatRepo.UpdateSession(ctx, combat); err != nil {
        return nil, err
    }
    
    result := &models.InitiativeResult{
        ParticipantID: participantID,
        Roll:          roll,
        Total:         total,
    }
    
    // Emit event
    s.emitEvent(CombatEvent{
        Type:      EventInitiativeRolled,
        SessionID: combat.ID,
        Initiative: result,
    })
    
    if allRolled {
        s.emitEvent(CombatEvent{
            Type:      EventCombatStarted,
            SessionID: combat.ID,
            Combat:    combat,
        })
    }
    
    return result, nil
}

func (s *CombatService) ExecuteAttack(ctx context.Context, req *AttackRequest) (*models.AttackResult, error) {
    combat, err := s.combatRepo.GetSession(ctx, req.SessionID)
    if err != nil {
        return nil, err
    }
    
    // Validate turn
    if combat.InitiativeOrder[combat.CurrentTurn] != req.AttackerID {
        return nil, fmt.Errorf("not your turn")
    }
    
    attacker := combat.Participants[req.AttackerID]
    target := combat.Participants[req.TargetID]
    
    if target == nil {
        return nil, fmt.Errorf("invalid target")
    }
    
    // Get attacker character for bonuses
    character, err := s.charRepo.FindByID(ctx, attacker.CharacterID)
    if err != nil {
        return nil, err
    }
    
    // Get weapon if specified
    var weapon *models.Equipment
    if req.WeaponID != "" {
        weapon = character.GetEquipment(req.WeaponID)
    }
    
    // Calculate attack bonus
    attackBonus := character.GetAttackBonus(weapon)
    
    // Roll attack
    attackExpr := "1d20"
    if req.Advantage {
        attackExpr = "2d20kh1" // Keep highest
    } else if req.Disadvantage {
        attackExpr = "2d20kl1" // Keep lowest
    }
    
    attackRoll := s.diceService.Roll(attackExpr)
    attackTotal := attackRoll.Total + attackBonus
    
    result := &models.AttackResult{
        AttackRoll:  attackRoll,
        AttackTotal: attackTotal,
        Hit:         attackTotal >= target.AC,
        Critical:    attackRoll.Rolls[0] == 20,
        CriticalMiss: attackRoll.Rolls[0] == 1,
    }
    
    // If hit, roll damage
    if result.Hit && !result.CriticalMiss {
        damageDice := character.GetDamageDice(weapon)
        if result.Critical {
            // Double dice on crit
            damageDice = models.DoubleDice(damageDice)
        }
        
        damageRoll := s.diceService.Roll(damageDice)
        damageBonus := character.GetDamageBonus(weapon)
        totalDamage := damageRoll.Total + damageBonus
        
        result.DamageRoll = damageRoll
        result.DamageTotal = totalDamage
        
        // Apply damage
        target.CurrentHP -= totalDamage
        if target.CurrentHP < 0 {
            target.CurrentHP = 0
            target.Conditions = append(target.Conditions, models.ConditionUnconscious)
        }
    }
    
    // Add to combat history
    combat.History = append(combat.History, models.CombatAction{
        Timestamp:   time.Now(),
        ActorID:     req.AttackerID,
        ActionType:  models.ActionAttack,
        Description: s.formatAttackDescription(attacker.Name, target.Name, result),
        Details:     result,
    })
    
    // Save changes
    if err := s.combatRepo.UpdateSession(ctx, combat); err != nil {
        return nil, err
    }
    
    // Emit event
    s.emitEvent(CombatEvent{
        Type:      EventActionExecuted,
        SessionID: combat.ID,
        Action:    combat.History[len(combat.History)-1],
    })
    
    return result, nil
}

func (s *CombatService) sortByInitiative(participants map[string]*models.CombatParticipant) []string {
    type initiativeEntry struct {
        ID         string
        Initiative int
    }
    
    entries := make([]initiativeEntry, 0, len(participants))
    for id, p := range participants {
        entries = append(entries, initiativeEntry{ID: id, Initiative: p.Initiative})
    }
    
    sort.Slice(entries, func(i, j int) bool {
        return entries[i].Initiative > entries[j].Initiative
    })
    
    result := make([]string, len(entries))
    for i, entry := range entries {
        result[i] = entry.ID
    }
    
    return result
}

func (s *CombatService) findStartingPosition(combat *models.CombatSession) models.Position {
    if combat.Map == nil {
        return models.Position{X: 0, Y: 0}
    }
    
    // Find an unoccupied position
    occupied := make(map[string]bool)
    for _, p := range combat.Participants {
        key := fmt.Sprintf("%d,%d", p.Position.X, p.Position.Y)
        occupied[key] = true
    }
    
    // Start from bottom of map
    y := combat.Map.Height - 1
    for x := 0; x < combat.Map.Width; x++ {
        key := fmt.Sprintf("%d,%d", x, y)
        if !occupied[key] {
            return models.Position{X: x, Y: y}
        }
    }
    
    return models.Position{X: 0, Y: 0}
}

func (s *CombatService) emitEvent(event CombatEvent) {
    select {
    case s.events <- event:
    default:
        // Channel full, log warning
    }
}
```

## Repository Layer Examples

### MongoDB Character Repository
```go
package mongodb

import (
    "context"
    "fmt"
    "time"
    
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
    
    "github.com/yourusername/dnd-bot-discord/internal/models"
    "github.com/yourusername/dnd-bot-discord/internal/repositories"
)

type characterRepository struct {
    db         *mongo.Database
    collection *mongo.Collection
}

func NewCharacterRepository(db *mongo.Database) repositories.CharacterRepository {
    return &characterRepository{
        db:         db,
        collection: db.Collection("characters"),
    }
}

func (r *characterRepository) Create(ctx context.Context, character *models.Character) error {
    character.ID = primitive.NewObjectID()
    character.CreatedAt = time.Now()
    character.UpdatedAt = time.Now()
    
    _, err := r.collection.InsertOne(ctx, character)
    if err != nil {
        return fmt.Errorf("failed to insert character: %w", err)
    }
    
    return nil
}

func (r *characterRepository) FindByID(ctx context.Context, id string) (*models.Character, error) {
    objectID, err := primitive.ObjectIDFromHex(id)
    if err != nil {
        return nil, fmt.Errorf("invalid character ID: %w", err)
    }
    
    var character models.Character
    err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&character)
    if err != nil {
        if err == mongo.ErrNoDocuments {
            return nil, repositories.ErrNotFound
        }
        return nil, fmt.Errorf("failed to find character: %w", err)
    }
    
    return &character, nil
}

func (r *characterRepository) FindByUserID(ctx context.Context, userID string) ([]*models.Character, error) {
    filter := bson.M{
        "user_id": userID,
        "status":  bson.M{"$ne": models.CharacterStatusArchived},
    }
    
    cursor, err := r.collection.Find(ctx, filter, options.Find().SetSort(bson.M{"updated_at": -1}))
    if err != nil {
        return nil, fmt.Errorf("failed to find characters: %w", err)
    }
    defer cursor.Close(ctx)
    
    var characters []*models.Character
    if err := cursor.All(ctx, &characters); err != nil {
        return nil, fmt.Errorf("failed to decode characters: %w", err)
    }
    
    return characters, nil
}

func (r *characterRepository) Update(ctx context.Context, character *models.Character) error {
    character.UpdatedAt = time.Now()
    
    filter := bson.M{"_id": character.ID}
    update := bson.M{"$set": character}
    
    result, err := r.collection.UpdateOne(ctx, filter, update)
    if err != nil {
        return fmt.Errorf("failed to update character: %w", err)
    }
    
    if result.MatchedCount == 0 {
        return repositories.ErrNotFound
    }
    
    return nil
}

func (r *characterRepository) Delete(ctx context.Context, id string) error {
    objectID, err := primitive.ObjectIDFromHex(id)
    if err != nil {
        return fmt.Errorf("invalid character ID: %w", err)
    }
    
    // Soft delete - just archive
    filter := bson.M{"_id": objectID}
    update := bson.M{
        "$set": bson.M{
            "status":     models.CharacterStatusArchived,
            "updated_at": time.Now(),
        },
    }
    
    result, err := r.collection.UpdateOne(ctx, filter, update)
    if err != nil {
        return fmt.Errorf("failed to archive character: %w", err)
    }
    
    if result.MatchedCount == 0 {
        return repositories.ErrNotFound
    }
    
    return nil
}

// Additional useful methods
func (r *characterRepository) FindByName(ctx context.Context, userID, name string) (*models.Character, error) {
    filter := bson.M{
        "user_id": userID,
        "name":    bson.M{"$regex": fmt.Sprintf("^%s$", name), "$options": "i"}, // Case insensitive
        "status":  bson.M{"$ne": models.CharacterStatusArchived},
    }
    
    var character models.Character
    err := r.collection.FindOne(ctx, filter).Decode(&character)
    if err != nil {
        if err == mongo.ErrNoDocuments {
            return nil, repositories.ErrNotFound
        }
        return nil, err
    }
    
    return &character, nil
}
```

### Redis Combat Repository
```go
package redis

import (
    "context"
    "encoding/json"
    "fmt"
    "time"
    
    "github.com/redis/go-redis/v9"
    "github.com/yourusername/dnd-bot-discord/internal/models"
    "github.com/yourusername/dnd-bot-discord/internal/repositories"
)

type combatRepository struct {
    client *redis.Client
    ttl    time.Duration
}

func NewCombatRepository(client *redis.Client) repositories.CombatRepository {
    return &combatRepository{
        client: client,
        ttl:    4 * time.Hour, // Combat sessions expire after 4 hours
    }
}

func (r *combatRepository) CreateSession(ctx context.Context, session *models.CombatSession) error {
    data, err := json.Marshal(session)
    if err != nil {
        return fmt.Errorf("failed to marshal session: %w", err)
    }
    
    key := r.sessionKey(session.ID)
    if err := r.client.Set(ctx, key, data, r.ttl).Err(); err != nil {
        return fmt.Errorf("failed to save session: %w", err)
    }
    
    return nil
}

func (r *combatRepository) GetSession(ctx context.Context, sessionID string) (*models.CombatSession, error) {
    key := r.sessionKey(sessionID)
    
    data, err := r.client.Get(ctx, key).Result()
    if err != nil {
        if err == redis.Nil {
            return nil, repositories.ErrNotFound
        }
        return nil, fmt.Errorf("failed to get session: %w", err)
    }
    
    var session models.CombatSession
    if err := json.Unmarshal([]byte(data), &session); err != nil {
        return nil, fmt.Errorf("failed to unmarshal session: %w", err)
    }
    
    return &session, nil
}

func (r *combatRepository) UpdateSession(ctx context.Context, session *models.CombatSession) error {
    data, err := json.Marshal(session)
    if err != nil {
        return fmt.Errorf("failed to marshal session: %w", err)
    }
    
    key := r.sessionKey(session.ID)
    
    // Use SET with XX to ensure it exists
    result := r.client.Set(ctx, key, data, r.ttl)
    if err := result.Err(); err != nil {
        return fmt.Errorf("failed to update session: %w", err)
    }
    
    return nil
}

func (r *combatRepository) DeleteSession(ctx context.Context, sessionID string) error {
    key := r.sessionKey(sessionID)
    
    if err := r.client.Del(ctx, key).Err(); err != nil {
        return fmt.Errorf("failed to delete session: %w", err)
    }
    
    // Also remove channel mapping if exists
    // We'll scan for it since we don't have reverse lookup
    pattern := "combat:channel:*"
    iter := r.client.Scan(ctx, 0, pattern, 0).Iterator()
    for iter.Next(ctx) {
        val, err := r.client.Get(ctx, iter.Val()).Result()
        if err == nil && val == sessionID {
            r.client.Del(ctx, iter.Val())
            break
        }
    }
    
    return nil
}

func (r *combatRepository) GetActiveSessionByChannel(ctx context.Context, channelID string) (*models.CombatSession, error) {
    key := r.channelKey(channelID)
    
    sessionID, err := r.client.Get(ctx, key).Result()
    if err != nil {
        if err == redis.Nil {
            return nil, repositories.ErrNotFound
        }
        return nil, err
    }
    
    return r.GetSession(ctx, sessionID)
}

func (r *combatRepository) SetChannelSession(ctx context.Context, channelID, sessionID string) error {
    key := r.channelKey(channelID)
    return r.client.Set(ctx, key, sessionID, r.ttl).Err()
}

func (r *combatRepository) sessionKey(sessionID string) string {
    return fmt.Sprintf("combat:session:%s", sessionID)
}

func (r *combatRepository) channelKey(channelID string) string {
    return fmt.Sprintf("combat:channel:%s", channelID)
}

// Additional methods for participant management
func (r *combatRepository) AddParticipant(ctx context.Context, sessionID string, participant *models.CombatParticipant) error {
    session, err := r.GetSession(ctx, sessionID)
    if err != nil {
        return err
    }
    
    if session.Participants == nil {
        session.Participants = make(map[string]*models.CombatParticipant)
    }
    
    session.Participants[participant.ID] = participant
    
    return r.UpdateSession(ctx, session)
}

func (r *combatRepository) UpdateParticipant(ctx context.Context, sessionID string, participant *models.CombatParticipant) error {
    session, err := r.GetSession(ctx, sessionID)
    if err != nil {
        return err
    }
    
    if _, exists := session.Participants[participant.ID]; !exists {
        return fmt.Errorf("participant not found")
    }
    
    session.Participants[participant.ID] = participant
    
    return r.UpdateSession(ctx, session)
}
```

## Message Building Examples

### Discord Message Service
```go
package services

import (
    "fmt"
    "strings"
    
    "github.com/bwmarrin/discordgo"
    "github.com/yourusername/dnd-bot-discord/internal/models"
)

type MessageService struct{}

func NewMessageService() *MessageService {
    return &MessageService{}
}

func (s *MessageService) BuildCharacterSheet(character *models.Character) *discordgo.MessageEmbed {
    // Build ability scores field
    abilities := fmt.Sprintf(
        "**STR** %d (%+d) | **DEX** %d (%+d) | **CON** %d (%+d)\n**INT** %d (%+d) | **WIS** %d (%+d) | **CHA** %d (%+d)",
        character.Attributes.Strength, character.GetModifier("str"),
        character.Attributes.Dexterity, character.GetModifier("dex"),
        character.Attributes.Constitution, character.GetModifier("con"),
        character.Attributes.Intelligence, character.GetModifier("int"),
        character.Attributes.Wisdom, character.GetModifier("wis"),
        character.Attributes.Charisma, character.GetModifier("cha"),
    )
    
    embed := &discordgo.MessageEmbed{
        Title: fmt.Sprintf("%s - Level %d %s %s", 
            character.Name, 
            character.Level, 
            character.Race.Name, 
            character.Class.Name,
        ),
        Color: 0x5865F2, // Discord blurple
        Fields: []*discordgo.MessageEmbedField{
            {
                Name:   "Hit Points",
                Value:  fmt.Sprintf("%d / %d", character.HP.Current, character.HP.Max),
                Inline: true,
            },
            {
                Name:   "Armor Class",
                Value:  fmt.Sprintf("%d", character.CombatStats.AC),
                Inline: true,
            },
            {
                Name:   "Speed",
                Value:  fmt.Sprintf("%d ft", character.CombatStats.Speed),
                Inline: true,
            },
            {
                Name:   "Proficiency Bonus",
                Value:  fmt.Sprintf("+%d", character.CombatStats.ProficiencyBonus),
                Inline: true,
            },
            {
                Name:   "Initiative",
                Value:  fmt.Sprintf("%+d", character.CombatStats.InitiativeBonus),
                Inline: true,
            },
            {
                Name:   "Experience",
                Value:  fmt.Sprintf("%d XP", character.Experience),
                Inline: true,
            },
            {
                Name:  "Ability Scores",
                Value: abilities,
                Inline: false,
            },
        },
        Footer: &discordgo.MessageEmbedFooter{
            Text: fmt.Sprintf("Created %s", character.CreatedAt.Format("Jan 2, 2006")),
        },
    }
    
    // Add equipped items
    var equipped []string
    for _, item := range character.Equipment {
        if item.Equipped {
            equipped = append(equipped, item.Name)
        }
    }
    if len(equipped) > 0 {
        embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
            Name:  "Equipped",
            Value: strings.Join(equipped, ", "),
            Inline: false,
        })
    }
    
    return embed
}

func (s *MessageService) BuildCharacterActions(character *models.Character) []discordgo.MessageComponent {
    return []discordgo.MessageComponent{
        discordgo.ActionsRow{
            Components: []discordgo.MessageComponent{
                discordgo.Button{
                    Label:    "Short Rest",
                    Style:    discordgo.SuccessButton,
                    CustomID: fmt.Sprintf("character_rest:short:%s", character.ID.Hex()),
                    Emoji: &discordgo.ComponentEmoji{
                        Name: "☕",
                    },
                },
                discordgo.Button{
                    Label:    "Long Rest",
                    Style:    discordgo.SuccessButton,
                    CustomID: fmt.Sprintf("character_rest:long:%s", character.ID.Hex()),
                    Emoji: &discordgo.ComponentEmoji{
                        Name: "🛏️",
                    },
                },
                discordgo.Button{
                    Label:    "Inventory",
                    Style:    discordgo.SecondaryButton,
                    CustomID: fmt.Sprintf("character_inventory:%s", character.ID.Hex()),
                    Emoji: &discordgo.ComponentEmoji{
                        Name: "🎒",
                    },
                },
            },
        },
    }
}

func (s *MessageService) BuildCombatStatus(combat *models.CombatSession) *discordgo.MessageEmbed {
    embed := &discordgo.MessageEmbed{
        Title: "⚔️ Combat Encounter",
        Color: 0xED4245, // Red
        Fields: []*discordgo.MessageEmbedField{
            {
                Name:   "Round",
                Value:  fmt.Sprintf("%d", combat.Round),
                Inline: true,
            },
            {
                Name:   "State",
                Value:  string(combat.State),
                Inline: true,
            },
        },
    }
    
    // Add participants
    if len(combat.Participants) > 0 {
        var participants []string
        
        // If we have initiative order, use that
        if len(combat.InitiativeOrder) > 0 {
            for i, id := range combat.InitiativeOrder {
                p := combat.Participants[id]
                marker := ""
                if i == combat.CurrentTurn && combat.State == models.CombatStateActive {
                    marker = "➤ "
                }
                
                hp := fmt.Sprintf("%d/%d", p.CurrentHP, p.MaxHP)
                if p.CurrentHP == 0 {
                    hp = "**UNCONSCIOUS**"
                }
                
                participants = append(participants, fmt.Sprintf(
                    "%s**%s** (Init: %d) - HP: %s, AC: %d",
                    marker, p.Name, p.Initiative, hp, p.AC,
                ))
            }
        } else {
            // Just list participants
            for _, p := range combat.Participants {
                participants = append(participants, fmt.Sprintf(
                    "**%s** - HP: %d/%d, AC: %d",
                    p.Name, p.CurrentHP, p.MaxHP, p.AC,
                ))
            }
        }
        
        embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
            Name:  "Participants",
            Value: strings.Join(participants, "\n"),
            Inline: false,
        })
    }
    
    // Add recent actions
    if len(combat.History) > 0 {
        var recentActions []string
        start := len(combat.History) - 3
        if start < 0 {
            start = 0
        }
        
        for _, action := range combat.History[start:] {
            recentActions = append(recentActions, fmt.Sprintf(
                "• %s",
                action.Description,
            ))
        }
        
        embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
            Name:  "Recent Actions",
            Value: strings.Join(recentActions, "\n"),
            Inline: false,
        })
    }
    
    return embed
}

func (s *MessageService) BuildCombatControls(combat *models.CombatSession, isYourTurn bool) []discordgo.MessageComponent {
    components := []discordgo.MessageComponent{}
    
    // Join button if still initiating
    if combat.State == models.CombatStateInitiating || combat.State == models.CombatStateRollingInitiative {
        components = append(components, discordgo.ActionsRow{
            Components: []discordgo.MessageComponent{
                discordgo.Button{
                    Label:    "Join Combat",
                    Style:    discordgo.PrimaryButton,
                    CustomID: fmt.Sprintf("combat_join:%s", combat.ID),
                    Emoji: &discordgo.ComponentEmoji{
                        Name: "⚔️",
                    },
                },
            },
        })
    }
    
    // Combat actions if it's your turn
    if isYourTurn && combat.State == models.CombatStateActive {
        components = append(components, discordgo.ActionsRow{
            Components: []discordgo.MessageComponent{
                discordgo.Button{
                    Label:    "Attack",
                    Style:    discordgo.DangerButton,
                    CustomID: fmt.Sprintf("combat_attack:%s", combat.ID),
                    Emoji: &discordgo.ComponentEmoji{
                        Name: "🗡️",
                    },
                },
                discordgo.Button{
                    Label:    "Move",
                    Style:    discordgo.PrimaryButton,
                    CustomID: fmt.Sprintf("combat_move:%s", combat.ID),
                    Emoji: &discordgo.ComponentEmoji{
                        Name: "🏃",
                    },
                },
                discordgo.Button{
                    Label:    "End Turn",
                    Style:    discordgo.SecondaryButton,
                    CustomID: fmt.Sprintf("combat_endturn:%s", combat.ID),
                },
            },
        })
    }
    
    return components
}

func (s *MessageService) BuildTargetSelectMenu(combat *models.CombatSession, attackerID string) discordgo.MessageComponent {
    options := []discordgo.SelectMenuOption{}
    
    for _, p := range combat.Participants {
        if p.ID == attackerID || p.CurrentHP <= 0 {
            continue // Can't target self or unconscious
        }
        
        options = append(options, discordgo.SelectMenuOption{
            Label:       p.Name,
            Value:       p.ID,
            Description: fmt.Sprintf("HP: %d/%d, AC: %d", p.CurrentHP, p.MaxHP, p.AC),
            Emoji: &discordgo.ComponentEmoji{
                Name: "🎯",
            },
        })
    }
    
    return discordgo.ActionsRow{
        Components: []discordgo.MessageComponent{
            discordgo.SelectMenu{
                CustomID:    fmt.Sprintf("combat_target:%s", combat.ID),
                Placeholder: "Select a target",
                Options:     options,
            },
        },
    }
}
```