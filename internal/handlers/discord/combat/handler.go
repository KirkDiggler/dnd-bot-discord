package combat

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/KirkDiggler/dnd-bot-discord/internal/entities"
	"github.com/KirkDiggler/dnd-bot-discord/internal/services/ability"
	"github.com/KirkDiggler/dnd-bot-discord/internal/services/encounter"
	"github.com/bwmarrin/discordgo"
)

// Handler handles combat-related Discord interactions
type Handler struct {
	encounterService encounter.Service
	abilityService   ability.Service
}

// appendCombatEndMessage adds combat end information to an embed
func appendCombatEndMessage(embed *discordgo.MessageEmbed, combatEnded, playersWon bool) {
	if !combatEnded {
		return
	}

	var endMessage string
	if playersWon {
		endMessage = "\n\n🎉 **VICTORY!** All enemies have been defeated!\n🪙 *Loot and XP will be distributed...*"
		embed.Color = 0x00ff00 // Green for victory
	} else {
		endMessage = "\n\n💀 **DEFEAT!** The party has fallen...\n⚰️ *Better luck next time...*"
		embed.Color = 0xff0000 // Red for defeat
	}
	embed.Description += endMessage
}

// getCombatEndMessage returns a short combat end message for ephemeral responses
func getCombatEndMessage(combatEnded, playersWon bool) string {
	if !combatEnded {
		return ""
	}

	if playersWon {
		return "\n\n🎉 **VICTORY!** All enemies defeated!"
	}
	return "\n\n💀 **DEFEAT!** Party has fallen..."
}

// NewHandler creates a new combat handler
func NewHandler(encounterService encounter.Service, abilityService ability.Service) *Handler {
	return &Handler{
		encounterService: encounterService,
		abilityService:   abilityService,
	}
}

// HandleButton handles combat button interactions
func (h *Handler) HandleButton(s *discordgo.Session, i *discordgo.InteractionCreate, action, encounterID string) error {
	// Removed verbose button logging - too noisy during combat

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
	case "my_actions":
		return h.handleMyActions(s, i, encounterID)
	case "summary":
		return h.handleSummary(s, i, encounterID)
	case "abilities":
		return h.handleShowAbilities(s, i, encounterID)
	case "use_ability":
		return h.handleUseAbility(s, i, encounterID)
	case "lay_on_hands_amount":
		return h.handleLayOnHandsAmount(s, i, encounterID)
	default:
		return fmt.Errorf("unknown combat action: %s", action)
	}
}

// handleAttack shows target selection UI
func (h *Handler) handleAttack(s *discordgo.Session, i *discordgo.InteractionCreate, encounterID string) error {
	// Check if this is from an ephemeral message (like My Actions)
	// If so, we need to send a new ephemeral response instead of updating
	if isEphemeralInteraction(i) {
		// For ephemeral sources, send a new ephemeral response with target selection
		return h.handleAttackFromEphemeral(s, i, encounterID)
	}

	// Defer response for processing (for shared messages)
	if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredMessageUpdate,
	}); err != nil {
		log.Printf("Failed to defer interaction response: %v", err)
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

	// Check if this interaction came from an ephemeral message
	isEphemeral := isEphemeralInteraction(i)

	if isEphemeral {
		// For ephemeral messages, defer with ephemeral flag
		if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags: discordgo.MessageFlagsEphemeral,
			},
		}); err != nil {
			log.Printf("Failed to defer ephemeral interaction response: %v", err)
			return fmt.Errorf("failed to defer ephemeral: %w", err)
		}
	} else {
		// For non-ephemeral messages, defer update
		if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredMessageUpdate,
		}); err != nil {
			log.Printf("Failed to defer interaction response: %v", err)
			return fmt.Errorf("failed to defer: %w", err)
		}
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

	// Use the standard combat embed for consistency
	embed := BuildCombatStatusEmbed(enc, result.MonsterAttacks)

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

	// Add combat end information if applicable
	appendCombatEndMessage(embed, result.CombatEnded, result.PlayersWon)

	// Build components based on state
	components := BuildCombatComponents(encounterID, result)

	if isEphemeral {
		// For ephemeral interactions, we need to:
		// 1. Respond to the ephemeral interaction
		// 2. Update the main shared combat message

		// Show attack result with option to get new action controller
		attackSummary := "Attack executed!"
		if result.PlayerAttack != nil {
			if result.PlayerAttack.Hit {
				if result.PlayerAttack.Critical {
					attackSummary = fmt.Sprintf("🎆 CRITICAL HIT! You dealt %d damage!", result.PlayerAttack.Damage)
				} else {
					attackSummary = fmt.Sprintf("✅ HIT! You dealt %d damage!", result.PlayerAttack.Damage)
				}
				if result.PlayerAttack.TargetDefeated {
					attackSummary += "\n💀 Target defeated!"
				}
			} else {
				attackSummary = "❌ MISS! Your attack missed!"
			}
		}

		// Add combat end information to ephemeral message
		attackSummary += getCombatEndMessage(result.CombatEnded, result.PlayersWon)

		// Show full combat status in ephemeral response, just like the shared message
		ephemeralEmbed := BuildCombatStatusEmbed(enc, result.MonsterAttacks)

		// Prepend the attack result to the description
		if result.PlayerAttack != nil {
			ephemeralEmbed.Description = attackSummary + "\n\n" + ephemeralEmbed.Description
		}

		// Add combat end information if applicable
		appendCombatEndMessage(ephemeralEmbed, result.CombatEnded, result.PlayersWon)

		// Use the same components as the shared message
		// BuildCombatComponents already includes "Get My Actions" button
		ephemeralComponents := BuildCombatComponents(encounterID, result)

		// Update the ephemeral message with the result
		_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Embeds:     &[]*discordgo.MessageEmbed{ephemeralEmbed},
			Components: &ephemeralComponents,
		})
		if err != nil {
			log.Printf("Failed to update ephemeral message with result: %v", err)
			return fmt.Errorf("failed to edit ephemeral: %w", err)
		}

		// Now update the main shared combat message
		if updateErr := updateSharedCombatMessage(s, encounterID, enc.MessageID, enc.ChannelID, embed, components); updateErr != nil {
			log.Printf("Failed to update shared combat message: %v", updateErr)
		}
		return nil
	}

	// For non-ephemeral interactions, update the message directly
	_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Embeds:     &[]*discordgo.MessageEmbed{embed},
		Components: &components,
	})

	// Also update the shared message if this wasn't the shared message itself
	if i.Message == nil || i.Message.ID != enc.MessageID {
		if updateErr := updateSharedCombatMessage(s, encounterID, enc.MessageID, enc.ChannelID, embed, components); updateErr != nil {
			log.Printf("Failed to update shared combat message: %v", updateErr)
		}
	}

	return err
}

// handleNextTurn advances the turn
func (h *Handler) handleNextTurn(s *discordgo.Session, i *discordgo.InteractionCreate, encounterID string) error {
	// Check if this is from an ephemeral message
	if !isEphemeralInteraction(i) {
		// Defer response for processing (for shared messages)
		if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredMessageUpdate,
		}); err != nil {
			log.Printf("Failed to defer interaction response: %v", err)
		}
	}

	// Advance turn
	if err := h.encounterService.NextTurn(context.Background(), encounterID, i.Member.User.ID); err != nil {
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
		monsterResults, err = h.encounterService.ProcessAllMonsterTurns(context.Background(), encounterID)
		if err != nil {
			log.Printf("Error processing monster turns: %v", err)
		}

		// Re-get encounter after monster turns
		updatedEnc, getErr := h.encounterService.GetEncounter(context.Background(), encounterID)
		if getErr != nil {
			log.Printf("Error getting encounter after monster turns: %v", getErr)
			// Continue with the existing encounter state rather than risk a nil pointer
		} else {
			enc = updatedEnc
		}
	}

	// Build combat status embed with clearer display
	embed := BuildCombatStatusEmbed(enc, monsterResults)

	// Add round complete indicator if needed
	if len(monsterResults) > 0 {
		var roundActions strings.Builder
		roundActions.WriteString("🔄 **Monster Actions This Turn:**\n")
		for _, ma := range monsterResults {
			if ma.Hit {
				if ma.TargetDefeated {
					roundActions.WriteString(fmt.Sprintf("• ⚔️ **%s** → **%s** | HIT 🩸 **%d** 💀\n", ma.AttackerName, ma.TargetName, ma.Damage))
				} else {
					roundActions.WriteString(fmt.Sprintf("• ⚔️ **%s** → **%s** | HIT 🩸 **%d**\n", ma.AttackerName, ma.TargetName, ma.Damage))
				}
			} else {
				roundActions.WriteString(fmt.Sprintf("• ❌ **%s** → **%s** | MISS\n", ma.AttackerName, ma.TargetName))
			}
		}
		embed.Description = roundActions.String() + "\n" + embed.Description
	}

	// Check whose turn it is now
	// No longer needed since Attack button removed from shared messages

	// Build components - no Attack button on shared messages
	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "Next Turn",
					Style:    discordgo.PrimaryButton,
					CustomID: fmt.Sprintf("combat:next_turn:%s", encounterID),
					Emoji:    &discordgo.ComponentEmoji{Name: "➡️"},
				},
				discordgo.Button{
					Label:    "Get My Actions",
					Style:    discordgo.SuccessButton,
					CustomID: fmt.Sprintf("combat:my_actions:%s", encounterID),
					Emoji:    &discordgo.ComponentEmoji{Name: "🎯"},
				},
				discordgo.Button{
					Label:    "View Status",
					Style:    discordgo.SecondaryButton,
					CustomID: fmt.Sprintf("combat:view:%s", encounterID),
					Emoji:    &discordgo.ComponentEmoji{Name: "📊"},
				},
			},
		},
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "History",
					Style:    discordgo.SecondaryButton,
					CustomID: fmt.Sprintf("combat:history:%s", encounterID),
					Emoji:    &discordgo.ComponentEmoji{Name: "📜"},
				},
			},
		},
	}

	// Handle ephemeral vs shared message updates
	if isEphemeralInteraction(i) {
		// For ephemeral interactions, acknowledge and update main combat message
		err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Turn skipped! Check the combat message for updates.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		if err != nil {
			log.Printf("Failed to respond to ephemeral interaction: %v", err)
		}

		// Update the main shared combat message
		if updateErr := updateSharedCombatMessage(s, encounterID, enc.MessageID, enc.ChannelID, embed, components); updateErr != nil {
			log.Printf("Failed to update shared combat message: %v", updateErr)
		}
		return nil
	}

	// For non-ephemeral, update the original message
	_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Embeds:     &[]*discordgo.MessageEmbed{embed},
		Components: &components,
	})

	// Also update the shared message if this wasn't the shared message itself
	if i.Message == nil || i.Message.ID != enc.MessageID {
		if updateErr := updateSharedCombatMessage(s, encounterID, enc.MessageID, enc.ChannelID, embed, components); updateErr != nil {
			log.Printf("Failed to update shared combat message: %v", updateErr)
		}
	}

	return err
}

// handleView shows current encounter status
func (h *Handler) handleView(s *discordgo.Session, i *discordgo.InteractionCreate, encounterID string) error {
	// Defer response for processing
	if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredMessageUpdate,
	}); err != nil {
		log.Printf("Failed to defer interaction response: %v", err)
	}

	enc, err := h.encounterService.GetEncounter(context.Background(), encounterID)
	if err != nil {
		return respondEditError(s, i, "Failed to get encounter", err)
	}

	embed := buildDetailedCombatEmbed(enc)

	// Build the combat status embed

	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "Next Turn",
					Style:    discordgo.PrimaryButton,
					CustomID: fmt.Sprintf("combat:next_turn:%s", encounterID),
					Emoji:    &discordgo.ComponentEmoji{Name: "➡️"},
					Disabled: enc.Status != entities.EncounterStatusActive,
				},
				discordgo.Button{
					Label:    "Get My Actions",
					Style:    discordgo.SuccessButton,
					CustomID: fmt.Sprintf("combat:my_actions:%s", encounterID),
					Emoji:    &discordgo.ComponentEmoji{Name: "🎯"},
				},
				discordgo.Button{
					Label:    "View Status",
					Style:    discordgo.SecondaryButton,
					CustomID: fmt.Sprintf("combat:view:%s", encounterID),
					Emoji:    &discordgo.ComponentEmoji{Name: "📊"},
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
		log.Printf("Failed to defer interaction response: %v", err)
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
		monsterResults, err = h.encounterService.ProcessAllMonsterTurns(context.Background(), encounterID)
		if err != nil {
			log.Printf("Error processing monster turns in continue round: %v", err)
		}

		// Re-get encounter
		enc, err = h.encounterService.GetEncounter(context.Background(), encounterID)
		if err != nil {
			log.Printf("Error getting encounter in continue round: %v", err)
			return respondEditError(s, i, "Failed to get updated encounter", err)
		}
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

	// Build the combat status embed

	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "Next Turn",
					Style:    discordgo.PrimaryButton,
					CustomID: fmt.Sprintf("combat:next_turn:%s", encounterID),
					Emoji:    &discordgo.ComponentEmoji{Name: "➡️"},
				},
				discordgo.Button{
					Label:    "Get My Actions",
					Style:    discordgo.SuccessButton,
					CustomID: fmt.Sprintf("combat:my_actions:%s", encounterID),
					Emoji:    &discordgo.ComponentEmoji{Name: "🎯"},
				},
				discordgo.Button{
					Label:    "View Status",
					Style:    discordgo.SecondaryButton,
					CustomID: fmt.Sprintf("combat:view:%s", encounterID),
					Emoji:    &discordgo.ComponentEmoji{Name: "📊"},
				},
			},
		},
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
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

// handleMyActions shows personalized actions for the player
func (h *Handler) handleMyActions(s *discordgo.Session, i *discordgo.InteractionCreate, encounterID string) error {
	// Get encounter
	enc, err := h.encounterService.GetEncounter(context.Background(), encounterID)
	if err != nil {
		return respondError(s, i, "Failed to get encounter", err)
	}

	// Find the player's combatant
	var playerCombatant *entities.Combatant
	for _, c := range enc.Combatants {
		if c.PlayerID == i.Member.User.ID && c.IsActive {
			playerCombatant = c
			break
		}
	}

	if playerCombatant == nil {
		return respondError(s, i, "You are not in this combat!", nil)
	}

	// Check if it's the player's turn
	isMyTurn := false
	if current := enc.GetCurrentCombatant(); current != nil {
		isMyTurn = current.ID == playerCombatant.ID
	}

	// Use the standard combat status embed for consistency
	embed := BuildCombatStatusEmbedForPlayer(enc, nil, playerCombatant.Name)

	// Update title and description for personalized view
	embed.Title = fmt.Sprintf("🎯 %s's Action Controller", playerCombatant.Name)
	embed.Description = "Choose your action:"
	if !isMyTurn && enc.Status == entities.EncounterStatusActive {
		if current := enc.GetCurrentCombatant(); current != nil {
			embed.Description = fmt.Sprintf("Waiting for %s's turn...", current.Name)
		}
	}

	// Add player status field showing HP, AC, and active effects
	statusValue := fmt.Sprintf("**HP:** %d/%d | **AC:** %d", playerCombatant.CurrentHP, playerCombatant.MaxHP, playerCombatant.AC)

	// TODO: Add active effects display when character service is accessible through encounter service
	// For now, we'll need to enhance the encounter service to expose character data for effects

	// Add status as first field
	statusField := &discordgo.MessageEmbedField{
		Name:   "📊 Your Status",
		Value:  statusValue,
		Inline: false,
	}
	embed.Fields = append([]*discordgo.MessageEmbedField{statusField}, embed.Fields...)

	// Build action buttons - always enabled unless combat is not active
	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "Attack",
					Style:    discordgo.DangerButton,
					CustomID: fmt.Sprintf("combat:attack:%s", encounterID),
					Emoji:    &discordgo.ComponentEmoji{Name: "⚔️"},
					Disabled: enc.Status != entities.EncounterStatusActive,
				},
				discordgo.Button{
					Label:    "Abilities",
					Style:    discordgo.PrimaryButton,
					CustomID: fmt.Sprintf("combat:abilities:%s", encounterID),
					Emoji:    &discordgo.ComponentEmoji{Name: "✨"},
					Disabled: enc.Status != entities.EncounterStatusActive,
				},
				discordgo.Button{
					Label:    "Skip Turn",
					Style:    discordgo.SecondaryButton,
					CustomID: fmt.Sprintf("combat:next_turn:%s", encounterID),
					Emoji:    &discordgo.ComponentEmoji{Name: "⏭️"},
					Disabled: enc.Status != entities.EncounterStatusActive,
				},
			},
		},
	}

	// TODO: Add more action types in the future:
	// - Use Item (potions, scrolls)
	// - Cast Spell
	// - Special Abilities
	// - Defensive Actions (dodge, dash, disengage)

	// Add footer with helpful info
	embed.Footer = &discordgo.MessageEmbedFooter{
		Text: "Action controller - Only visible to you",
	}

	// Check if this is being called from an ephemeral message (like "Back to Actions")
	if isEphemeralInteraction(i) {
		// Update the existing ephemeral message
		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseUpdateMessage,
			Data: &discordgo.InteractionResponseData{
				Embeds:     []*discordgo.MessageEmbed{embed},
				Components: components,
			},
		})
	}

	// Create new ephemeral message (when called from shared message buttons)
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds:     []*discordgo.MessageEmbed{embed},
			Components: components,
			Flags:      discordgo.MessageFlagsEphemeral,
		},
	})
}

// handleAttackFromEphemeral handles attack button clicks from ephemeral messages
func (h *Handler) handleAttackFromEphemeral(s *discordgo.Session, i *discordgo.InteractionCreate, encounterID string) error {
	// Get encounter to build target list
	enc, err := h.encounterService.GetEncounter(context.Background(), encounterID)
	if err != nil {
		return respondError(s, i, "Failed to get encounter", err)
	}

	// Find attacker - player who clicked
	var attacker *entities.Combatant
	for _, c := range enc.Combatants {
		if c.PlayerID == i.Member.User.ID && c.IsActive {
			attacker = c
			break
		}
	}

	if attacker == nil || !attacker.IsActive {
		return respondError(s, i, "No active character found", nil)
	}

	// Check if it's actually this player's turn
	current := enc.GetCurrentCombatant()
	if current == nil || current.ID != attacker.ID {
		// Not their turn - show a friendly message with action controller button
		embed := &discordgo.MessageEmbed{
			Title:       "⏳ Not Your Turn",
			Description: fmt.Sprintf("It's currently **%s's** turn.", current.Name),
			Color:       0xf39c12, // Orange
			Footer: &discordgo.MessageEmbedFooter{
				Text: "Wait for your turn to attack",
			},
		}

		components := []discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.Button{
						Label:    "Back to Actions",
						Style:    discordgo.PrimaryButton,
						CustomID: fmt.Sprintf("combat:my_actions:%s", encounterID),
						Emoji:    &discordgo.ComponentEmoji{Name: "🎯"},
					},
				},
			},
		}

		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseUpdateMessage,
			Data: &discordgo.InteractionResponseData{
				Embeds:     []*discordgo.MessageEmbed{embed},
				Components: components,
			},
		})
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
		return respondError(s, i, "No valid targets available", nil)
	}

	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("⚔️ %s's Target Selection", attacker.Name),
		Description: "Choose your target:",
		Color:       0xe74c3c,
	}

	// Update the existing ephemeral message with target selection
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{Components: buttons},
			},
		},
	})
}
