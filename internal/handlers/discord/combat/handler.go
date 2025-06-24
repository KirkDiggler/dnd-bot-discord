package combat

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/KirkDiggler/dnd-bot-discord/internal/entities"
	"github.com/KirkDiggler/dnd-bot-discord/internal/services/encounter"
	"github.com/bwmarrin/discordgo"
)

// Handler handles combat-related Discord interactions
type Handler struct {
	encounterService encounter.Service
}

// NewHandler creates a new combat handler
func NewHandler(encounterService encounter.Service) *Handler {
	return &Handler{
		encounterService: encounterService,
	}
}

// HandleButton handles combat button interactions
func (h *Handler) HandleButton(s *discordgo.Session, i *discordgo.InteractionCreate, action string, encounterID string) error {
	log.Printf("Combat button: action=%s, encounter=%s, user=%s", action, encounterID, i.Member.User.ID)

	switch action {
	case "attack":
		return h.handleAttack(s, i, encounterID)
	case "select_target":
		return h.handleSelectTarget(s, i, encounterID)
	case "next_turn":
		return h.handleNextTurn(s, i, encounterID)
	case "view":
		return h.handleView(s, i, encounterID)
	case "continue_round":
		return h.handleContinueRound(s, i, encounterID)
	case "history":
		return h.handleHistory(s, i, encounterID)
	default:
		return fmt.Errorf("unknown combat action: %s", action)
	}
}

// handleAttack shows target selection UI
func (h *Handler) handleAttack(s *discordgo.Session, i *discordgo.InteractionCreate, encounterID string) error {
	// Defer response for processing
	if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredMessageUpdate,
	}); err != nil {
		log.Printf("Failed to defer response: %v", err)
	}

	// Get encounter to build target list
	enc, err := h.encounterService.GetEncounter(context.Background(), encounterID)
	if err != nil {
		return respondEditError(s, i, "Failed to get encounter", err)
	}

	// Find attacker - player who clicked or current turn for DM
	var attacker *entities.Combatant
	for _, c := range enc.Combatants {
		if c.PlayerID == i.Member.User.ID && c.IsActive {
			attacker = c
			break
		}
	}

	// If not found and user is DM, use current turn
	if attacker == nil && enc.CreatedBy == i.Member.User.ID {
		attacker = enc.GetCurrentCombatant()
	}

	if attacker == nil || !attacker.IsActive {
		return respondEditError(s, i, "No active character found", nil)
	}

	// Build target buttons
	var buttons []discordgo.MessageComponent
	for _, target := range enc.Combatants {
		if target.ID == attacker.ID || !target.IsActive || target.CurrentHP <= 0 {
			continue
		}
		
		// Players cannot attack other players
		if attacker.Type == entities.CombatantTypePlayer && target.Type == entities.CombatantTypePlayer {
			continue
		}

		emoji := "🧑"
		if target.Type == entities.CombatantTypeMonster {
			emoji = "👹"
		}

		buttons = append(buttons, discordgo.Button{
			Label:    fmt.Sprintf("%s (HP: %d/%d)", target.Name, target.CurrentHP, target.MaxHP),
			Style:    discordgo.PrimaryButton,
			CustomID: fmt.Sprintf("combat:select_target:%s:%s", encounterID, target.ID),
			Emoji:    &discordgo.ComponentEmoji{Name: emoji},
		})

		if len(buttons) >= 5 {
			break // Discord limit
		}
	}

	if len(buttons) == 0 {
		return respondEditError(s, i, "No valid targets available", nil)
	}

	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("⚔️ %s's Attack", attacker.Name),
		Description: "Select your target:",
		Color:       0xe74c3c,
	}

	_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Embeds: &[]*discordgo.MessageEmbed{embed},
		Components: &[]discordgo.MessageComponent{
			discordgo.ActionsRow{Components: buttons},
		},
	})
	return err
}

// handleSelectTarget executes the attack
func (h *Handler) handleSelectTarget(s *discordgo.Session, i *discordgo.InteractionCreate, encounterID string) error {
	// Parse target ID from custom ID: combat:select_target:encounterID:targetID
	parts := parseCustomID(i.MessageComponentData().CustomID)
	if len(parts) < 4 {
		return respondError(s, i, "Invalid target selection", nil)
	}
	targetID := parts[3]

	// Defer response for long operation
	if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredMessageUpdate,
	}); err != nil {
		log.Printf("Failed to defer response: %v", err)
	}

	// Execute attack with service
	result, err := h.encounterService.ExecuteAttackWithTarget(context.Background(), &encounter.ExecuteAttackInput{
		EncounterID: encounterID,
		TargetID:    targetID,
		UserID:      i.Member.User.ID,
	})
	if err != nil {
		return respondEditError(s, i, "Failed to execute attack", err)
	}

	// Get updated encounter for detailed view
	enc, err := h.encounterService.GetEncounter(context.Background(), encounterID)
	if err != nil {
		return respondEditError(s, i, "Failed to get updated encounter", err)
	}

	// Build detailed combat embed (like view status)
	embed := buildDetailedCombatEmbed(enc)

	// Add attack result summary at the top
	if result.PlayerAttack != nil {
		attack := result.PlayerAttack
		attackSummary := fmt.Sprintf("**%s attacked %s!**\n", attack.AttackerName, attack.TargetName)
		if attack.Hit {
			if attack.Critical {
				attackSummary += fmt.Sprintf("🎆 CRITICAL HIT! %d damage!", attack.Damage)
			} else {
				attackSummary += fmt.Sprintf("✅ Hit for %d damage!", attack.Damage)
			}
			if attack.TargetDefeated {
				attackSummary += " 💀 **DEFEATED!**"
			}
		} else {
			attackSummary += "❌ **MISS!**"
		}
		embed.Description = attackSummary + "\n\n" + embed.Description
	}

	// Build components based on state
	components := buildCombatComponents(encounterID, result)

	_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Embeds:     &[]*discordgo.MessageEmbed{embed},
		Components: &components,
	})
	return err
}

// handleNextTurn advances the turn
func (h *Handler) handleNextTurn(s *discordgo.Session, i *discordgo.InteractionCreate, encounterID string) error {
	log.Printf("handleNextTurn: encounterID=%s, userID=%s", encounterID, i.Member.User.ID)
	
	// Defer response for processing
	if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredMessageUpdate,
	}); err != nil {
		log.Printf("Failed to defer response: %v", err)
	}

	// Advance turn
	if err := h.encounterService.NextTurn(context.Background(), encounterID, i.Member.User.ID); err != nil {
		log.Printf("NextTurn failed: %v", err)
		return respondEditError(s, i, "Failed to advance turn", err)
	}

	// Get updated encounter
	enc, err := h.encounterService.GetEncounter(context.Background(), encounterID)
	if err != nil {
		return respondEditError(s, i, "Failed to get encounter", err)
	}

	// Check if round is complete
	if enc.IsRoundComplete() {
		return h.showRoundComplete(s, i, enc)
	}

	// Process monster turns if any
	var monsterResults []*encounter.AttackResult
	if current := enc.GetCurrentCombatant(); current != nil && current.Type == entities.CombatantTypeMonster {
		monsterResults, _ = h.encounterService.ProcessAllMonsterTurns(context.Background(), encounterID)

		// Re-get encounter after monster turns
		enc, _ = h.encounterService.GetEncounter(context.Background(), encounterID)
	}

	// Build detailed combat embed (like view status)
	embed := buildDetailedCombatEmbed(enc)

	// Add monster actions summary if any
	if len(monsterResults) > 0 {
		var monsterSummary strings.Builder
		for _, ma := range monsterResults {
			if ma.Hit {
				monsterSummary.WriteString(fmt.Sprintf("👹 **%s** attacked %s for %d damage!\n", ma.AttackerName, ma.TargetName, ma.Damage))
			} else {
				monsterSummary.WriteString(fmt.Sprintf("👹 **%s** missed %s!\n", ma.AttackerName, ma.TargetName))
			}
		}
		embed.Description = monsterSummary.String() + "\n" + embed.Description
	}

	// Check whose turn it is now
	isPlayerTurn := false
	if current := enc.GetCurrentCombatant(); current != nil {
		isPlayerTurn = current.PlayerID == i.Member.User.ID
	}

	// Build components
	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "Attack",
					Style:    discordgo.DangerButton,
					CustomID: fmt.Sprintf("combat:attack:%s", encounterID),
					Emoji:    &discordgo.ComponentEmoji{Name: "⚔️"},
					Disabled: !isPlayerTurn,
				},
				discordgo.Button{
					Label:    "Next Turn",
					Style:    discordgo.PrimaryButton,
					CustomID: fmt.Sprintf("combat:next_turn:%s", encounterID),
					Emoji:    &discordgo.ComponentEmoji{Name: "➡️"},
				},
				discordgo.Button{
					Label:    "View Status",
					Style:    discordgo.SecondaryButton,
					CustomID: fmt.Sprintf("combat:view:%s", encounterID),
					Emoji:    &discordgo.ComponentEmoji{Name: "📊"},
				},
				discordgo.Button{
					Label:    "History",
					Style:    discordgo.SecondaryButton,
					CustomID: fmt.Sprintf("combat:history:%s", encounterID),
					Emoji:    &discordgo.ComponentEmoji{Name: "📜"},
				},
			},
		},
	}

	// Update the original message
	_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Embeds:     &[]*discordgo.MessageEmbed{embed},
		Components: &components,
	})
	return err
}

// handleView shows current encounter status
func (h *Handler) handleView(s *discordgo.Session, i *discordgo.InteractionCreate, encounterID string) error {
	// Defer response for processing
	if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredMessageUpdate,
	}); err != nil {
		log.Printf("Failed to defer response: %v", err)
	}

	enc, err := h.encounterService.GetEncounter(context.Background(), encounterID)
	if err != nil {
		return respondEditError(s, i, "Failed to get encounter", err)
	}

	embed := buildDetailedCombatEmbed(enc)

	// Check whose turn it is
	isPlayerTurn := false
	if current := enc.GetCurrentCombatant(); current != nil {
		isPlayerTurn = current.PlayerID == i.Member.User.ID
		log.Printf("handleView: current turn=%s (playerID=%s), checking user=%s, isPlayerTurn=%v",
			current.Name, current.PlayerID, i.Member.User.ID, isPlayerTurn)
	}

	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "Attack",
					Style:    discordgo.DangerButton,
					CustomID: fmt.Sprintf("combat:attack:%s", encounterID),
					Emoji:    &discordgo.ComponentEmoji{Name: "⚔️"},
					Disabled: !isPlayerTurn || enc.Status != entities.EncounterStatusActive,
				},
				discordgo.Button{
					Label:    "Next Turn",
					Style:    discordgo.PrimaryButton,
					CustomID: fmt.Sprintf("combat:next_turn:%s", encounterID),
					Emoji:    &discordgo.ComponentEmoji{Name: "➡️"},
					Disabled: enc.Status != entities.EncounterStatusActive,
				},
			},
		},
	}

	_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Embeds:     &[]*discordgo.MessageEmbed{embed},
		Components: &components,
	})
	return err
}

// handleContinueRound starts the next round
func (h *Handler) handleContinueRound(s *discordgo.Session, i *discordgo.InteractionCreate, encounterID string) error {
	// Defer response for processing
	if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredMessageUpdate,
	}); err != nil {
		log.Printf("Failed to defer response: %v", err)
	}

	// Start next round
	enc, err := h.encounterService.GetEncounter(context.Background(), encounterID)
	if err != nil {
		return respondEditError(s, i, "Failed to get encounter", err)
	}

	// Reset round
	enc.NextRound()

	// Process any monster turns at start of round
	var monsterResults []*encounter.AttackResult
	if current := enc.GetCurrentCombatant(); current != nil && current.Type == entities.CombatantTypeMonster {
		monsterResults, _ = h.encounterService.ProcessAllMonsterTurns(context.Background(), encounterID)

		// Re-get encounter
		enc, _ = h.encounterService.GetEncounter(context.Background(), encounterID)
	}

	// Build detailed combat embed
	embed := buildDetailedCombatEmbed(enc)

	// Add round start and monster actions if any
	roundSummary := fmt.Sprintf("🔄 **Round %d Begins!**\n\n", enc.Round)
	if len(monsterResults) > 0 {
		for _, ma := range monsterResults {
			if ma.Hit {
				roundSummary += fmt.Sprintf("👹 **%s** attacked %s for %d damage!\n", ma.AttackerName, ma.TargetName, ma.Damage)
			} else {
				roundSummary += fmt.Sprintf("👹 **%s** missed %s!\n", ma.AttackerName, ma.TargetName)
			}
		}
	}
	embed.Description = roundSummary + "\n" + embed.Description

	// Check whose turn it is
	isPlayerTurn := false
	if current := enc.GetCurrentCombatant(); current != nil {
		isPlayerTurn = current.PlayerID == i.Member.User.ID
	}

	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "Attack",
					Style:    discordgo.DangerButton,
					CustomID: fmt.Sprintf("combat:attack:%s", encounterID),
					Emoji:    &discordgo.ComponentEmoji{Name: "⚔️"},
					Disabled: !isPlayerTurn,
				},
				discordgo.Button{
					Label:    "Next Turn",
					Style:    discordgo.PrimaryButton,
					CustomID: fmt.Sprintf("combat:next_turn:%s", encounterID),
					Emoji:    &discordgo.ComponentEmoji{Name: "➡️"},
				},
				discordgo.Button{
					Label:    "View Status",
					Style:    discordgo.SecondaryButton,
					CustomID: fmt.Sprintf("combat:view:%s", encounterID),
					Emoji:    &discordgo.ComponentEmoji{Name: "📊"},
				},
				discordgo.Button{
					Label:    "History",
					Style:    discordgo.SecondaryButton,
					CustomID: fmt.Sprintf("combat:history:%s", encounterID),
					Emoji:    &discordgo.ComponentEmoji{Name: "📜"},
				},
			},
		},
	}

	// Update the message (we're already deferred from handleNextTurn)
	_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Embeds:     &[]*discordgo.MessageEmbed{embed},
		Components: &components,
	})
	return err
}

// showRoundComplete shows the round complete UI (updates the message)
func (h *Handler) showRoundComplete(s *discordgo.Session, i *discordgo.InteractionCreate, enc *entities.Encounter) error {
	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("🔄 Round %d Complete!", enc.Round),
		Description: "All combatants have acted this round.",
		Color:       0x3498db,
		Fields:      []*discordgo.MessageEmbedField{},
	}

	// Show current status
	var statusList string
	for _, c := range enc.Combatants {
		if c.IsActive {
			emoji := "💀"
			if c.CurrentHP > c.MaxHP/2 {
				emoji = "💚"
			} else if c.CurrentHP > 0 {
				emoji = "💛"
			}
			statusList += fmt.Sprintf("%s %s: %d/%d HP\n", emoji, c.Name, c.CurrentHP, c.MaxHP)
		}
	}
	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
		Name:   "Current Status",
		Value:  statusList,
		Inline: false,
	})

	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "Continue to Next Round",
					Style:    discordgo.SuccessButton,
					CustomID: fmt.Sprintf("combat:continue_round:%s", enc.ID),
					Emoji:    &discordgo.ComponentEmoji{Name: "▶️"},
				},
				discordgo.Button{
					Label:    "View Status",
					Style:    discordgo.SecondaryButton,
					CustomID: fmt.Sprintf("combat:view:%s", enc.ID),
					Emoji:    &discordgo.ComponentEmoji{Name: "📊"},
				},
			},
		},
	}

	// Update the message (we're already deferred from handleNextTurn)
	_, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Embeds:     &[]*discordgo.MessageEmbed{embed},
		Components: &components,
	})
	return err
}

// handleHistory shows the full combat log
func (h *Handler) handleHistory(s *discordgo.Session, i *discordgo.InteractionCreate, encounterID string) error {
	enc, err := h.encounterService.GetEncounter(context.Background(), encounterID)
	if err != nil {
		return respondError(s, i, "Failed to get encounter", err)
	}

	// Build history embed
	embed := &discordgo.MessageEmbed{
		Title:       "📜 Combat History",
		Description: fmt.Sprintf("**%s** - Round %d", enc.Name, enc.Round),
		Color:       0x9b59b6, // Purple
		Fields:      []*discordgo.MessageEmbedField{},
	}

	// Show all combat log entries
	if len(enc.CombatLog) > 0 {
		// Discord has a 1024 character limit per field, so we may need multiple fields
		const maxFieldLength = 1024
		var currentField strings.Builder
		fieldNum := 1

		for _, entry := range enc.CombatLog {
			line := fmt.Sprintf("• %s\n", entry)
			if currentField.Len()+len(line) > maxFieldLength {
				// Add current field and start a new one
				embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
					Name:   fmt.Sprintf("Page %d", fieldNum),
					Value:  currentField.String(),
					Inline: false,
				})
				currentField.Reset()
				fieldNum++
			}
			currentField.WriteString(line)
		}

		// Add the last field
		if currentField.Len() > 0 {
			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
				Name:   fmt.Sprintf("Page %d", fieldNum),
				Value:  currentField.String(),
				Inline: false,
			})
		}
	} else {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "No History",
			Value:  "No combat actions have been recorded yet.",
			Inline: false,
		})
	}

	// Add footer with return button info
	embed.Footer = &discordgo.MessageEmbedFooter{
		Text: "Use the View Status button to return to combat",
	}

	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
			Flags:  discordgo.MessageFlagsEphemeral,
		},
	})
}
