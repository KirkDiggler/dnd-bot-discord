package combat

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/KirkDiggler/dnd-bot-discord/internal/entities"
	"github.com/KirkDiggler/dnd-bot-discord/internal/services/ability"
	"github.com/KirkDiggler/dnd-bot-discord/internal/services/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/services/encounter"
	"github.com/bwmarrin/discordgo"
)

// Handler handles combat-related Discord interactions
type Handler struct {
	encounterService encounter.Service
	abilityService   ability.Service
	characterService character.Service
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
func NewHandler(encounterService encounter.Service, abilityService ability.Service, characterService character.Service) *Handler {
	return &Handler{
		encounterService: encounterService,
		abilityService:   abilityService,
		characterService: characterService,
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
	case "bonus_action":
		return h.handleBonusAction(s, i, encounterID)
	case "bonus_target":
		return h.handleBonusTarget(s, i, encounterID)
	case "bt":
		// Short form for bonus target to avoid Discord's 100 char limit
		return h.handleBonusTargetShort(s, i, encounterID)
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
		// For ephemeral messages, defer update to keep the same message
		if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredMessageUpdate,
		}); err != nil {
			log.Printf("Failed to defer ephemeral update: %v", err)
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
		// For ephemeral interactions, update the existing ephemeral message
		// with the action controller after the attack
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

		// After attack, automatically return to action controller
		// This provides a smoother flow without needing to click "Back to Actions"
		actionEmbed, actionComponents, buildErr := h.buildActionController(enc, encounterID, i.Member.User.ID)
		if buildErr != nil {
			// Fallback to combat status if we can't build action controller
			ephemeralEmbed := BuildCombatStatusEmbed(enc, result.MonsterAttacks)
			if result.PlayerAttack != nil {
				ephemeralEmbed.Description = attackSummary + "\n\n" + ephemeralEmbed.Description
			}
			appendCombatEndMessage(ephemeralEmbed, result.CombatEnded, result.PlayersWon)
			ephemeralComponents := BuildCombatComponents(encounterID, result)

			_, updateErr := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Embeds:     &[]*discordgo.MessageEmbed{ephemeralEmbed},
				Components: &ephemeralComponents,
			})
			return updateErr
		}

		// Add attack result to the top of the action controller
		if result.PlayerAttack != nil {
			attackResultField := &discordgo.MessageEmbedField{
				Name:   "⚔️ Attack Result",
				Value:  attackSummary,
				Inline: false,
			}
			// Insert at the beginning of fields, after status
			if len(actionEmbed.Fields) > 0 {
				actionEmbed.Fields = append([]*discordgo.MessageEmbedField{actionEmbed.Fields[0], attackResultField}, actionEmbed.Fields[1:]...)
			} else {
				actionEmbed.Fields = append([]*discordgo.MessageEmbedField{attackResultField}, actionEmbed.Fields...)
			}
		}

		// Update the ephemeral message with action controller
		_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Embeds:     &[]*discordgo.MessageEmbed{actionEmbed},
			Components: &actionComponents,
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
	// Defer response appropriately based on message type
	if isEphemeralInteraction(i) {
		// For ephemeral messages, defer with update
		if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredMessageUpdate,
		}); err != nil {
			log.Printf("Failed to defer ephemeral update: %v", err)
		}
	} else {
		// For shared messages, defer with update
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

	// Build components - use shared combat components
	components := BuildCombatComponents(encounterID, &encounter.ExecuteAttackResult{})

	// Handle ephemeral vs shared message updates
	if isEphemeralInteraction(i) {
		// For ephemeral interactions, update the ephemeral message with turn end confirmation
		currentCombatant := enc.GetCurrentCombatant()
		turnMessage := "Your turn has ended."
		if currentCombatant != nil {
			turnMessage = fmt.Sprintf("Your turn has ended. It's now **%s's** turn.", currentCombatant.Name)
		}

		// Create a simple confirmation embed
		confirmEmbed := &discordgo.MessageEmbed{
			Title:       "✅ Turn Complete",
			Description: turnMessage,
			Color:       0x00ff00,
			Footer: &discordgo.MessageEmbedFooter{
				Text: "The combat continues...",
			},
		}

		// Update the ephemeral message
		_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Embeds: &[]*discordgo.MessageEmbed{confirmEmbed},
			Components: &[]discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.Button{
							Label:    "Get My Actions",
							Style:    discordgo.SuccessButton,
							CustomID: fmt.Sprintf("combat:my_actions:%s", encounterID),
							Emoji:    &discordgo.ComponentEmoji{Name: "🎯"},
						},
						discordgo.Button{
							Label:    "View Combat",
							Style:    discordgo.SecondaryButton,
							CustomID: fmt.Sprintf("combat:view:%s", encounterID),
							Emoji:    &discordgo.ComponentEmoji{Name: "📊"},
						},
					},
				},
			},
		})
		if err != nil {
			log.Printf("Failed to update ephemeral message: %v", err)
		}

		// Update the main shared combat message with proper components
		sharedComponents := BuildCombatComponents(encounterID, &encounter.ExecuteAttackResult{})
		if updateErr := updateSharedCombatMessage(s, encounterID, enc.MessageID, enc.ChannelID, embed, sharedComponents); updateErr != nil {
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

	// Check if combat has ended
	combatEnded := enc.Status == entities.EncounterStatusCompleted
	playersWon := false

	// Check if players won by seeing if any players are still active
	if combatEnded {
		for _, c := range enc.Combatants {
			if c.Type == entities.CombatantTypePlayer && c.IsActive {
				playersWon = true
				break
			}
		}
	}

	// Use the standard combat status embed for consistency
	embed := BuildCombatStatusEmbedForPlayer(enc, nil, playerCombatant.Name)

	// Update title and description for personalized view
	embed.Title = fmt.Sprintf("🎯 %s's Action Controller", playerCombatant.Name)
	embed.Description = "Choose your action:"

	// Add combat end message if applicable
	if combatEnded {
		appendCombatEndMessage(embed, combatEnded, playersWon)
	} else if !isMyTurn && enc.Status == entities.EncounterStatusActive {
		if current := enc.GetCurrentCombatant(); current != nil {
			embed.Description = fmt.Sprintf("Waiting for %s's turn...", current.Name)
		}
	}

	// Add player status field showing HP, AC, and active effects
	statusValue := fmt.Sprintf("**HP:** %d/%d | **AC:** %d", playerCombatant.CurrentHP, playerCombatant.MaxHP, playerCombatant.AC)

	// Get character data to check available bonus actions and action economy
	var actionEconomyInfo string
	var availableBonusActions []entities.BonusActionOption
	var char *entities.Character
	if playerCombatant.CharacterID != "" && h.characterService != nil {
		// Get the character
		ch, err := h.characterService.GetByID(playerCombatant.CharacterID)
		if err == nil && ch != nil {
			char = ch
			// Get action economy status
			actionStatus := "✅ Available"
			if char.Resources != nil && char.Resources.ActionEconomy.ActionUsed {
				actionStatus = "❌ Used"
			}

			bonusActionStatus := "✅ Available"
			if char.Resources != nil && char.Resources.ActionEconomy.BonusActionUsed {
				bonusActionStatus = "❌ Used"
			}

			actionEconomyInfo = fmt.Sprintf("\n**Action:** %s | **Bonus Action:** %s", actionStatus, bonusActionStatus)

			// Get available bonus actions
			if char.Resources != nil {
				availableBonusActions = char.GetAvailableBonusActions()
				if len(availableBonusActions) > 0 && !char.Resources.ActionEconomy.BonusActionUsed {
					actionEconomyInfo += "\n**Bonus Actions Available:**"
					for _, ba := range availableBonusActions {
						actionEconomyInfo += fmt.Sprintf("\n• %s", ba.Name)
					}
				}
			}
		}
	}

	// Add status as first field
	statusField := &discordgo.MessageEmbedField{
		Name:   "📊 Your Status",
		Value:  statusValue + actionEconomyInfo,
		Inline: false,
	}
	embed.Fields = append([]*discordgo.MessageEmbedField{statusField}, embed.Fields...)

	var components []discordgo.MessageComponent

	if combatEnded {
		// Show end of combat buttons
		components = BuildCombatComponents(encounterID, &encounter.ExecuteAttackResult{
			CombatEnded: combatEnded,
			PlayersWon:  playersWon,
		})
	} else {
		// Build action buttons - disable if action already used
		attackDisabled := enc.Status != entities.EncounterStatusActive
		if char != nil && char.Resources != nil && char.Resources.ActionEconomy.ActionUsed {
			attackDisabled = true
		}

		components = []discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.Button{
						Label:    "Attack",
						Style:    discordgo.DangerButton,
						CustomID: fmt.Sprintf("combat:attack:%s", encounterID),
						Emoji:    &discordgo.ComponentEmoji{Name: "⚔️"},
						Disabled: attackDisabled,
					},
					discordgo.Button{
						Label:    "Abilities",
						Style:    discordgo.PrimaryButton,
						CustomID: fmt.Sprintf("combat:abilities:%s", encounterID),
						Emoji:    &discordgo.ComponentEmoji{Name: "✨"},
						Disabled: enc.Status != entities.EncounterStatusActive,
					},
					discordgo.Button{
						Label:    "End Turn",
						Style:    discordgo.SecondaryButton,
						CustomID: fmt.Sprintf("combat:next_turn:%s", encounterID),
						Emoji:    &discordgo.ComponentEmoji{Name: "⏭️"},
						Disabled: enc.Status != entities.EncounterStatusActive,
					},
				},
			},
		}
	}

	// Add bonus action buttons if available and combat is still active
	if len(availableBonusActions) > 0 && !combatEnded {
		bonusActionButtons := []discordgo.MessageComponent{}
		for i, ba := range availableBonusActions {
			if i >= 5 { // Discord has a 5-button limit per row
				break
			}

			// Determine emoji based on action type
			emoji := "🎯"
			switch ba.ActionType {
			case "unarmed_strike":
				emoji = "👊"
			case "weapon_attack":
				emoji = "🗡️"
			}

			// Disable bonus action buttons if bonus action already used
			bonusActionDisabled := enc.Status != entities.EncounterStatusActive
			if char != nil && char.Resources != nil && char.Resources.ActionEconomy.BonusActionUsed {
				bonusActionDisabled = true
			}

			bonusActionButtons = append(bonusActionButtons, discordgo.Button{
				Label:    ba.Name,
				Style:    discordgo.SuccessButton,
				CustomID: fmt.Sprintf("combat:bonus_action:%s:%s", encounterID, ba.Key),
				Emoji:    &discordgo.ComponentEmoji{Name: emoji},
				Disabled: bonusActionDisabled,
			})
		}

		if len(bonusActionButtons) > 0 {
			components = append(components, discordgo.ActionsRow{
				Components: bonusActionButtons,
			})
		}
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

// handleBonusAction handles bonus action button interactions
func (h *Handler) handleBonusAction(s *discordgo.Session, i *discordgo.InteractionCreate, encounterID string) error {
	// Parse bonus action key from custom ID: combat:bonus_action:encounterID:bonusActionKey
	parts := parseCustomID(i.MessageComponentData().CustomID)
	if len(parts) < 4 {
		return respondError(s, i, "Invalid bonus action format", nil)
	}
	bonusActionKey := parts[3]

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

	// Verify it's the player's turn
	current := enc.GetCurrentCombatant()
	if current == nil || current.ID != playerCombatant.ID {
		embed := &discordgo.MessageEmbed{
			Title:       "⏳ Not Your Turn",
			Description: fmt.Sprintf("It's currently **%s's** turn.", current.Name),
			Color:       0xf39c12, // Orange
			Footer: &discordgo.MessageEmbedFooter{
				Text: "Wait for your turn to use bonus actions",
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

	// Handle different bonus action types
	switch bonusActionKey {
	case "martial_arts_strike":
		return h.handleMartialArtsStrike(s, i, encounterID, playerCombatant)
	case "two_weapon_attack":
		return h.handleTwoWeaponAttack(s, i, encounterID, playerCombatant)
	default:
		return respondError(s, i, "Unknown bonus action type", nil)
	}
}

// handleMartialArtsStrike shows target selection for unarmed strike bonus action
func (h *Handler) handleMartialArtsStrike(s *discordgo.Session, i *discordgo.InteractionCreate, encounterID string, attacker *entities.Combatant) error {
	// Get encounter for target list
	enc, err := h.encounterService.GetEncounter(context.Background(), encounterID)
	if err != nil {
		return respondError(s, i, "Failed to get encounter", err)
	}

	// Build target buttons for unarmed strike
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

		// Shorten only target ID to fit Discord's 100 char limit
		// Keep full encounter ID for proper routing
		shortTargetID := target.ID
		if len(shortTargetID) > 8 {
			shortTargetID = shortTargetID[:8]
		}

		buttons = append(buttons, discordgo.Button{
			Label:    fmt.Sprintf("%s (HP: %d/%d)", target.Name, target.CurrentHP, target.MaxHP),
			Style:    discordgo.PrimaryButton,
			CustomID: fmt.Sprintf("combat:bt:%s:%s:mas", encounterID, shortTargetID),
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
		Title:       fmt.Sprintf("👊 %s's Martial Arts Strike", attacker.Name),
		Description: "Choose your target for the bonus unarmed strike:",
		Color:       0x00aedb, // Blue for monk actions
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Bonus Action - Martial Arts",
		},
	}

	// Add cancel button
	buttons = append(buttons, discordgo.Button{
		Label:    "Cancel",
		Style:    discordgo.SecondaryButton,
		CustomID: fmt.Sprintf("combat:my_actions:%s", encounterID),
		Emoji:    &discordgo.ComponentEmoji{Name: "❌"},
	})

	// Update the ephemeral message with target selection
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{Components: buttons[:min(5, len(buttons))]},
			},
		},
	})
}

// handleTwoWeaponAttack shows target selection for off-hand weapon attack
func (h *Handler) handleTwoWeaponAttack(s *discordgo.Session, i *discordgo.InteractionCreate, encounterID string, attacker *entities.Combatant) error {
	// Get encounter for target list
	enc, err := h.encounterService.GetEncounter(context.Background(), encounterID)
	if err != nil {
		return respondError(s, i, "Failed to get encounter", err)
	}

	// Build target buttons for off-hand attack
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

		// Shorten only target ID to fit Discord's 100 char limit
		// Keep full encounter ID for proper routing
		shortTargetID := target.ID
		if len(shortTargetID) > 8 {
			shortTargetID = shortTargetID[:8]
		}

		buttons = append(buttons, discordgo.Button{
			Label:    fmt.Sprintf("%s (HP: %d/%d)", target.Name, target.CurrentHP, target.MaxHP),
			Style:    discordgo.PrimaryButton,
			CustomID: fmt.Sprintf("combat:bt:%s:%s:twa", encounterID, shortTargetID),
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
		Title:       fmt.Sprintf("🗡️ %s's Off-Hand Attack", attacker.Name),
		Description: "Choose your target for the off-hand weapon attack:",
		Color:       0xdc143c, // Crimson for dual wielding
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Bonus Action - Two-Weapon Fighting",
		},
	}

	// Add cancel button
	buttons = append(buttons, discordgo.Button{
		Label:    "Cancel",
		Style:    discordgo.SecondaryButton,
		CustomID: fmt.Sprintf("combat:my_actions:%s", encounterID),
		Emoji:    &discordgo.ComponentEmoji{Name: "❌"},
	})

	// Update the ephemeral message with target selection
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{Components: buttons[:min(5, len(buttons))]},
			},
		},
	})
}

// handleBonusTargetShort handles the shortened custom ID format for bonus targets
// It resolves the short IDs to full IDs and delegates to executeBonusAction
func (h *Handler) handleBonusTargetShort(s *discordgo.Session, i *discordgo.InteractionCreate, encounterID string) error {
	// Parse custom ID format: combat:bt:shortEncID:shortTargetID:actionType
	parts := parseCustomID(i.MessageComponentData().CustomID)
	if len(parts) < 5 {
		return respondError(s, i, "Invalid bonus target format", nil)
	}

	// Parts: [combat, bt, shortEncID, shortTargetID, actionType]
	shortTargetID := parts[3]
	actionType := parts[4]

	// Get the encounter to resolve the short target ID
	enc, err := h.encounterService.GetEncounter(context.Background(), encounterID)
	if err != nil {
		return respondError(s, i, "Failed to get encounter", err)
	}

	// Find target that starts with shortTargetID
	var fullTargetID string
	for _, combatant := range enc.Combatants {
		if strings.HasPrefix(combatant.ID, shortTargetID) {
			fullTargetID = combatant.ID
			break
		}
	}

	if fullTargetID == "" {
		return respondError(s, i, "Target not found", nil)
	}

	// Map short action type to full bonus action key
	bonusActionKey := ""
	switch actionType {
	case "mas":
		bonusActionKey = "martial_arts_strike"
	case "twa":
		bonusActionKey = "two_weapon_attack"
	default:
		return respondError(s, i, "Unknown bonus action type", nil)
	}

	// Execute the bonus action with resolved IDs
	return h.executeBonusAction(s, i, encounterID, fullTargetID, bonusActionKey)
}

// executeBonusAction handles the common logic for executing bonus actions
func (h *Handler) executeBonusAction(s *discordgo.Session, i *discordgo.InteractionCreate, encounterID, targetID, bonusActionKey string) error {

	// Defer update since this comes from ephemeral message
	if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredMessageUpdate,
	}); err != nil {
		log.Printf("Failed to defer message update: %v", err)
		return fmt.Errorf("failed to defer: %w", err)
	}

	// Get encounter
	enc, err := h.encounterService.GetEncounter(context.Background(), encounterID)
	if err != nil {
		return respondEditError(s, i, "Failed to get encounter", err)
	}

	// Find attacker combatant
	var attacker *entities.Combatant
	for _, c := range enc.Combatants {
		if c.PlayerID == i.Member.User.ID && c.IsActive {
			attacker = c
			break
		}
	}

	if attacker == nil {
		return respondEditError(s, i, "No active character found", nil)
	}

	// Verify it's still the attacker's turn
	current := enc.GetCurrentCombatant()
	if current == nil || current.ID != attacker.ID {
		return respondEditError(s, i, "No longer your turn", nil)
	}

	// Get target combatant
	var target *entities.Combatant
	for _, c := range enc.Combatants {
		if c.ID == targetID {
			target = c
			break
		}
	}

	if target == nil || !target.IsActive || target.CurrentHP <= 0 {
		return respondEditError(s, i, "Invalid target", nil)
	}

	// Get character and mark bonus action as used
	var char *entities.Character
	if attacker.CharacterID != "" && h.characterService != nil {
		ch, errGetChar := h.characterService.GetByID(attacker.CharacterID)
		if errGetChar != nil {
			return respondEditError(s, i, "Failed to get character", errGetChar)
		}
		char = ch

		// Mark the bonus action as used
		if !char.UseBonusAction(bonusActionKey) {
			return respondEditError(s, i, "Bonus action no longer available", nil)
		}

		// Save character to persist the bonus action used state
		if errUpdate := h.characterService.UpdateEquipment(char); errUpdate != nil {
			log.Printf("Failed to update character after bonus action: %v", errUpdate)
		}
	}

	// Execute the appropriate attack based on bonus action type
	var result *encounter.AttackResult
	switch bonusActionKey {
	case "martial_arts_strike":
		// Execute unarmed strike with martial arts damage
		result, err = h.executeUnarmedStrike(enc, attacker, target, char, true)
		if err != nil {
			return respondEditError(s, i, "Failed to execute martial arts strike", err)
		}

	case "two_weapon_attack":
		// Execute off-hand weapon attack
		result, err = h.executeOffHandAttack(enc, attacker, target, char)
		if err != nil {
			return respondEditError(s, i, "Failed to execute off-hand attack", err)
		}

	default:
		return respondEditError(s, i, "Unknown bonus action type", nil)
	}

	// Log the combat action
	if errLog := h.encounterService.LogCombatAction(context.Background(), encounterID, result.LogEntry); errLog != nil {
		log.Printf("Failed to log bonus action: %v", errLog)
	}

	// Get updated encounter for display
	enc, err = h.encounterService.GetEncounter(context.Background(), encounterID)
	if err != nil {
		return respondEditError(s, i, "Failed to get updated encounter", err)
	}

	// Build success embed with attack results
	actionName := "Bonus Action"
	switch bonusActionKey {
	case "martial_arts_strike":
		actionName = "Martial Arts Strike"
	case "two_weapon_attack":
		actionName = "Off-Hand Attack"
	}
	attackSummary := fmt.Sprintf("**%s used %s on %s!**\n", attacker.Name, actionName, target.Name)
	if result.Hit {
		if result.Critical {
			attackSummary += fmt.Sprintf("🎆 CRITICAL HIT! %d damage!", result.Damage)
		} else {
			attackSummary += fmt.Sprintf("✅ Hit for %d damage!", result.Damage)
		}
		if result.TargetDefeated {
			attackSummary += " 💀 **DEFEATED!**"
		}
	} else {
		attackSummary += "❌ **MISS!**"
	}

	// Check for combat end
	combatEnded := false
	playersWon := false
	if shouldEnd, won := enc.CheckCombatEnd(); shouldEnd {
		combatEnded = true
		playersWon = won
	}

	// After bonus action, return to action controller for smoother flow
	actionEmbed, actionComponents, err := h.buildActionController(enc, encounterID, i.Member.User.ID)
	if err != nil {
		// Fallback to combat status if we can't build action controller
		embed := BuildCombatStatusEmbed(enc, nil)
		embed.Description = attackSummary + "\n\n" + embed.Description
		appendCombatEndMessage(embed, combatEnded, playersWon)

		components := []discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.Button{
						Label:    "End Turn",
						Style:    discordgo.PrimaryButton,
						CustomID: fmt.Sprintf("combat:next_turn:%s", encounterID),
						Emoji:    &discordgo.ComponentEmoji{Name: "➡️"},
						Disabled: enc.Status != entities.EncounterStatusActive,
					},
					discordgo.Button{
						Label:    "Back to Actions",
						Style:    discordgo.SecondaryButton,
						CustomID: fmt.Sprintf("combat:my_actions:%s", encounterID),
						Emoji:    &discordgo.ComponentEmoji{Name: "🎯"},
					},
				},
			},
		}

		_, updateErr := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Embeds:     &[]*discordgo.MessageEmbed{embed},
			Components: &components,
		})
		return updateErr
	}

	// Add bonus action result to the action controller
	bonusResultField := &discordgo.MessageEmbedField{
		Name:   "🎯 Bonus Action Result",
		Value:  attackSummary,
		Inline: false,
	}
	// Insert after status field
	if len(actionEmbed.Fields) > 0 {
		actionEmbed.Fields = append([]*discordgo.MessageEmbedField{actionEmbed.Fields[0], bonusResultField}, actionEmbed.Fields[1:]...)
	} else {
		actionEmbed.Fields = append([]*discordgo.MessageEmbedField{bonusResultField}, actionEmbed.Fields...)
	}

	// Update the ephemeral message with action controller
	_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Embeds:     &[]*discordgo.MessageEmbed{actionEmbed},
		Components: &actionComponents,
	})
	if err != nil {
		log.Printf("Failed to update ephemeral message with bonus action result: %v", err)
	}

	// Update the main shared combat message
	sharedEmbed := BuildCombatStatusEmbed(enc, nil)
	sharedEmbed.Description = attackSummary + "\n\n" + sharedEmbed.Description
	appendCombatEndMessage(sharedEmbed, combatEnded, playersWon)

	sharedComponents := BuildCombatComponents(encounterID, &encounter.ExecuteAttackResult{
		CombatEnded: combatEnded,
		PlayersWon:  playersWon,
	})

	if updateErr := updateSharedCombatMessage(s, encounterID, enc.MessageID, enc.ChannelID, sharedEmbed, sharedComponents); updateErr != nil {
		log.Printf("Failed to update shared combat message: %v", updateErr)
	}

	return nil
}

// handleBonusTarget executes the bonus action attack after target selection
func (h *Handler) handleBonusTarget(s *discordgo.Session, i *discordgo.InteractionCreate, encounterID string) error {
	// Parse custom ID format: combat:bonus_target:encounterID:targetID:bonusActionKey
	parts := parseCustomID(i.MessageComponentData().CustomID)
	if len(parts) < 5 {
		return respondError(s, i, "Invalid bonus target format", nil)
	}
	targetID := parts[3]
	bonusActionKey := parts[4]

	// Delegate to the shared executeBonusAction method
	return h.executeBonusAction(s, i, encounterID, targetID, bonusActionKey)
}

// buildActionController builds the action controller embed and components
func (h *Handler) buildActionController(enc *entities.Encounter, encounterID, userID string) (*discordgo.MessageEmbed, []discordgo.MessageComponent, error) {
	// Find the player's combatant
	var playerCombatant *entities.Combatant
	for _, c := range enc.Combatants {
		if c.PlayerID == userID && c.IsActive {
			playerCombatant = c
			break
		}
	}

	if playerCombatant == nil {
		return nil, nil, fmt.Errorf("player not in combat")
	}

	// Check if it's the player's turn
	isMyTurn := false
	if current := enc.GetCurrentCombatant(); current != nil {
		isMyTurn = current.ID == playerCombatant.ID
	}

	// Check if combat has ended
	combatEnded := enc.Status == entities.EncounterStatusCompleted
	playersWon := false

	// Check if players won by seeing if any players are still active
	if combatEnded {
		for _, c := range enc.Combatants {
			if c.Type == entities.CombatantTypePlayer && c.IsActive {
				playersWon = true
				break
			}
		}
	}

	// Use the standard combat status embed for consistency
	embed := BuildCombatStatusEmbedForPlayer(enc, nil, playerCombatant.Name)

	// Update title and description for personalized view
	embed.Title = fmt.Sprintf("🎯 %s's Action Controller", playerCombatant.Name)
	embed.Description = "Choose your action:"

	// Add combat end message if applicable
	if combatEnded {
		appendCombatEndMessage(embed, combatEnded, playersWon)
	} else if !isMyTurn && enc.Status == entities.EncounterStatusActive {
		if current := enc.GetCurrentCombatant(); current != nil {
			embed.Description = fmt.Sprintf("Waiting for %s's turn...", current.Name)
		}
	}

	// Add player status field showing HP, AC, and active effects
	statusValue := fmt.Sprintf("**HP:** %d/%d | **AC:** %d", playerCombatant.CurrentHP, playerCombatant.MaxHP, playerCombatant.AC)

	// Get character data to check available bonus actions and action economy
	var actionEconomyInfo string
	var availableBonusActions []entities.BonusActionOption
	var char *entities.Character
	if playerCombatant.CharacterID != "" && h.characterService != nil {
		// Get the character
		ch, err := h.characterService.GetByID(playerCombatant.CharacterID)
		if err == nil && ch != nil {
			char = ch
			// Get action economy status
			actionStatus := "✅ Available"
			if char.Resources != nil && char.Resources.ActionEconomy.ActionUsed {
				actionStatus = "❌ Used"
			}

			bonusActionStatus := "✅ Available"
			if char.Resources != nil && char.Resources.ActionEconomy.BonusActionUsed {
				bonusActionStatus = "❌ Used"
			}

			actionEconomyInfo = fmt.Sprintf("\n**Action:** %s | **Bonus Action:** %s", actionStatus, bonusActionStatus)

			// Get available bonus actions
			if char.Resources != nil {
				availableBonusActions = char.GetAvailableBonusActions()
				if len(availableBonusActions) > 0 && !char.Resources.ActionEconomy.BonusActionUsed {
					actionEconomyInfo += "\n**Bonus Actions Available:**"
					for _, ba := range availableBonusActions {
						actionEconomyInfo += fmt.Sprintf("\n• %s", ba.Name)
					}
				}
			}
		}
	}

	// Add status as first field
	statusField := &discordgo.MessageEmbedField{
		Name:   "📊 Your Status",
		Value:  statusValue + actionEconomyInfo,
		Inline: false,
	}
	embed.Fields = append([]*discordgo.MessageEmbedField{statusField}, embed.Fields...)

	var components []discordgo.MessageComponent

	if combatEnded {
		// Show end of combat buttons
		components = BuildCombatComponents(encounterID, &encounter.ExecuteAttackResult{
			CombatEnded: combatEnded,
			PlayersWon:  playersWon,
		})
	} else {
		// Build action buttons - disable if action already used
		attackDisabled := enc.Status != entities.EncounterStatusActive || !isMyTurn
		if char != nil && char.Resources != nil && char.Resources.ActionEconomy.ActionUsed {
			attackDisabled = true
		}

		components = []discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.Button{
						Label:    "Attack",
						Style:    discordgo.DangerButton,
						CustomID: fmt.Sprintf("combat:attack:%s", encounterID),
						Emoji:    &discordgo.ComponentEmoji{Name: "⚔️"},
						Disabled: attackDisabled,
					},
					discordgo.Button{
						Label:    "Abilities",
						Style:    discordgo.PrimaryButton,
						CustomID: fmt.Sprintf("combat:abilities:%s", encounterID),
						Emoji:    &discordgo.ComponentEmoji{Name: "✨"},
						Disabled: enc.Status != entities.EncounterStatusActive || !isMyTurn,
					},
					discordgo.Button{
						Label:    "End Turn",
						Style:    discordgo.SecondaryButton,
						CustomID: fmt.Sprintf("combat:next_turn:%s", encounterID),
						Emoji:    &discordgo.ComponentEmoji{Name: "⏭️"},
						Disabled: enc.Status != entities.EncounterStatusActive || !isMyTurn,
					},
				},
			},
		}
	}

	// Add bonus action buttons if available and combat is still active
	if len(availableBonusActions) > 0 && isMyTurn && !combatEnded {
		bonusActionButtons := []discordgo.MessageComponent{}
		for i, ba := range availableBonusActions {
			if i >= 5 { // Discord has a 5-button limit per row
				break
			}

			// Determine emoji based on action type
			emoji := "🎯"
			switch ba.ActionType {
			case "unarmed_strike":
				emoji = "👊"
			case "weapon_attack":
				emoji = "🗡️"
			}

			// Disable bonus action buttons if bonus action already used
			bonusActionDisabled := enc.Status != entities.EncounterStatusActive || !isMyTurn
			if char != nil && char.Resources != nil && char.Resources.ActionEconomy.BonusActionUsed {
				bonusActionDisabled = true
			}

			bonusActionButtons = append(bonusActionButtons, discordgo.Button{
				Label:    ba.Name,
				Style:    discordgo.SuccessButton,
				CustomID: fmt.Sprintf("combat:bonus_action:%s:%s", encounterID, ba.Key),
				Emoji:    &discordgo.ComponentEmoji{Name: emoji},
				Disabled: bonusActionDisabled,
			})
		}

		if len(bonusActionButtons) > 0 {
			components = append(components, discordgo.ActionsRow{
				Components: bonusActionButtons,
			})
		}
	}

	// Add footer with helpful info
	embed.Footer = &discordgo.MessageEmbedFooter{
		Text: "Action controller - Only visible to you",
	}

	return embed, components, nil
}

// executeUnarmedStrike executes an unarmed strike attack
func (h *Handler) executeUnarmedStrike(enc *entities.Encounter, attacker, target *entities.Combatant, char *entities.Character, isMartialArts bool) (*encounter.AttackResult, error) {
	result := &encounter.AttackResult{
		AttackerName: attacker.Name,
		TargetName:   target.Name,
		TargetAC:     target.AC,
		WeaponName:   "Unarmed Strike",
	}

	// Determine ability bonus - monks can use DEX instead of STR
	abilityBonus := 0
	if char != nil && char.Attributes != nil {
		strBonus := 0
		dexBonus := 0
		if char.Attributes[entities.AttributeStrength] != nil {
			strBonus = char.Attributes[entities.AttributeStrength].Bonus
		}
		if char.Attributes[entities.AttributeDexterity] != nil {
			dexBonus = char.Attributes[entities.AttributeDexterity].Bonus
		}

		// Monks with martial arts can use DEX instead of STR
		if isMartialArts && dexBonus > strBonus {
			abilityBonus = dexBonus
		} else {
			abilityBonus = strBonus
		}
	}

	// Calculate proficiency bonus (everyone is proficient with unarmed strikes)
	proficiencyBonus := 0
	if char != nil {
		proficiencyBonus = 2 + ((char.Level - 1) / 4)
	}

	// Roll attack
	attackBonus := abilityBonus + proficiencyBonus
	// For now, we'll use a simple attack simulation since we don't have access to dice roller here
	// In a real implementation, this should be handled by the encounter service
	attackRoll := 10 + attackBonus // Simulated average roll
	result.AttackRoll = 10
	result.AttackBonus = attackBonus
	result.TotalAttack = attackRoll
	result.Critical = false // Simplified for now
	result.Hit = result.TotalAttack >= target.AC

	// Determine damage dice size based on monk level
	diceSize := 4 // Default
	if isMartialArts && char != nil {
		switch {
		case char.Level >= 17:
			diceSize = 10
		case char.Level >= 11:
			diceSize = 8
		case char.Level >= 5:
			diceSize = 6
		}
	}

	if result.Hit {
		// Roll damage - using average for simulation
		baseDamage := (diceSize / 2) + 1 // Average roll

		result.Damage = baseDamage + abilityBonus
		result.DamageType = "bludgeoning"

		// Apply damage to target
		target.CurrentHP -= result.Damage
		if target.CurrentHP <= 0 {
			target.CurrentHP = 0
			target.IsActive = false
			result.TargetDefeated = true
		}

		// Build log entry
		// Calculate the actual damage roll (using average for now)
		damageRoll := (diceSize / 2) + 1

		if result.Critical {
			result.LogEntry = fmt.Sprintf("👊 **%s** → **%s** | 💥 CRIT! 🩸 **%d** ||d20:**%d**%+d=%d vs AC:%d, dmg:1d%d: [%d]+%d||",
				result.AttackerName, result.TargetName, result.Damage,
				result.AttackRoll, result.AttackBonus, result.TotalAttack, result.TargetAC,
				diceSize, damageRoll, abilityBonus)
		} else {
			result.LogEntry = fmt.Sprintf("👊 **%s** → **%s** | HIT 🩸 **%d** ||d20:%d%+d=%d vs AC:%d, dmg:1d%d: [%d]+%d||",
				result.AttackerName, result.TargetName, result.Damage,
				result.AttackRoll, result.AttackBonus, result.TotalAttack, result.TargetAC,
				diceSize, damageRoll, abilityBonus)
		}

		if result.TargetDefeated {
			result.LogEntry += " 💀"
		}
	} else {
		result.LogEntry = fmt.Sprintf("👊 **%s** → **%s** | ❌ MISS ||d20:%d%+d=%d vs AC:%d||",
			result.AttackerName, result.TargetName,
			result.AttackRoll, result.AttackBonus, result.TotalAttack, result.TargetAC)
	}

	return result, nil
}

// executeOffHandAttack executes an off-hand weapon attack
func (h *Handler) executeOffHandAttack(enc *entities.Encounter, attacker, target *entities.Combatant, char *entities.Character) (*encounter.AttackResult, error) {
	result := &encounter.AttackResult{
		AttackerName: attacker.Name,
		TargetName:   target.Name,
		TargetAC:     target.AC,
	}

	// Get off-hand weapon
	var offHandWeapon *entities.Weapon
	if char != nil && char.EquippedSlots != nil {
		if weapon, ok := char.EquippedSlots[entities.SlotOffHand].(*entities.Weapon); ok {
			offHandWeapon = weapon
			result.WeaponName = offHandWeapon.GetName()
		}
	}

	if offHandWeapon == nil {
		return nil, fmt.Errorf("no weapon equipped in off-hand")
	}

	// Calculate ability bonus (no ability modifier to damage for off-hand unless Two-Weapon Fighting style)
	abilityBonus := 0
	damageAbilityBonus := 0
	if char != nil && char.Attributes != nil {
		if char.Attributes[entities.AttributeDexterity] != nil && offHandWeapon.IsFinesse() {
			abilityBonus = char.Attributes[entities.AttributeDexterity].Bonus
		} else if char.Attributes[entities.AttributeStrength] != nil {
			abilityBonus = char.Attributes[entities.AttributeStrength].Bonus
		}

		// Check for Two-Weapon Fighting style
		// Note: getFightingStyle is not exported, so we'll check for the feature
		hasTwoWeaponFighting := false
		for _, feature := range char.Features {
			if feature != nil && feature.Key == "fighting-style-two-weapon-fighting" {
				hasTwoWeaponFighting = true
				break
			}
		}
		if hasTwoWeaponFighting {
			damageAbilityBonus = abilityBonus
		}
	}

	// Calculate proficiency bonus
	proficiencyBonus := 0
	if char != nil && char.HasWeaponProficiency(offHandWeapon.GetKey()) {
		proficiencyBonus = 2 + ((char.Level - 1) / 4)
	}

	// Roll attack
	attackBonus := abilityBonus + proficiencyBonus
	// For now, we'll use a simple attack simulation
	attackRoll := 10 + attackBonus // Simulated average roll
	result.AttackRoll = 10
	result.AttackBonus = attackBonus
	result.TotalAttack = attackRoll
	result.Critical = false // Simplified for now
	result.Hit = result.TotalAttack >= target.AC

	if result.Hit {
		// Roll damage
		damage := offHandWeapon.Damage
		if damage == nil {
			return nil, fmt.Errorf("weapon has no damage dice")
		}

		// Roll damage - using average for simulation
		baseDamage := 0
		for i := 0; i < damage.DiceCount; i++ {
			baseDamage += (damage.DiceSize / 2) + 1 // Average roll per die
		}

		result.Damage = baseDamage + damageAbilityBonus
		result.DamageType = string(damage.DamageType)

		// Apply damage to target
		target.CurrentHP -= result.Damage
		if target.CurrentHP <= 0 {
			target.CurrentHP = 0
			target.IsActive = false
			result.TargetDefeated = true
		}

		// Build log entry
		// Calculate the actual damage roll (using average for now)
		avgRoll := (damage.DiceSize / 2) + 1
		totalRoll := avgRoll * damage.DiceCount

		// Build damage string similar to main attacks
		damageStr := fmt.Sprintf("%dd%d: [%d]", damage.DiceCount, damage.DiceSize, totalRoll)
		if damageAbilityBonus != 0 {
			damageStr += fmt.Sprintf("%+d", damageAbilityBonus)
		}

		if result.Critical {
			result.LogEntry = fmt.Sprintf("🗡️ **%s** → **%s** | 💥 CRIT! 🩸 **%d** ||d20:**%d**%+d=%d vs AC:%d, dmg:%s||",
				result.AttackerName, result.TargetName, result.Damage,
				result.AttackRoll, result.AttackBonus, result.TotalAttack, result.TargetAC,
				damageStr)
		} else {
			result.LogEntry = fmt.Sprintf("🗡️ **%s** → **%s** | HIT 🩸 **%d** ||d20:%d%+d=%d vs AC:%d, dmg:%s||",
				result.AttackerName, result.TargetName, result.Damage,
				result.AttackRoll, result.AttackBonus, result.TotalAttack, result.TargetAC,
				damageStr)
		}

		if result.TargetDefeated {
			result.LogEntry += " 💀"
		}
	} else {
		result.LogEntry = fmt.Sprintf("🗡️ **%s** → **%s** | ❌ MISS ||d20:%d%+d=%d vs AC:%d||",
			result.AttackerName, result.TargetName,
			result.AttackRoll, result.AttackBonus, result.TotalAttack, result.TargetAC)
	}

	return result, nil
}
