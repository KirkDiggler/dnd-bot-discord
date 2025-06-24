package dungeon

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/KirkDiggler/dnd-bot-discord/internal/dice"
	"github.com/KirkDiggler/dnd-bot-discord/internal/entities"
	"github.com/KirkDiggler/dnd-bot-discord/internal/entities/attack"
	"github.com/KirkDiggler/dnd-bot-discord/internal/entities/damage"
	"github.com/KirkDiggler/dnd-bot-discord/internal/services"
	"github.com/bwmarrin/discordgo"
)

// CombatController manages the combat flow and interactions
type CombatController struct {
	services         *services.Provider
	combatLogUpdater func(ctx context.Context, encounterID string) error
}

// NewCombatController creates a new combat controller
func NewCombatController(services *services.Provider, combatLogUpdater func(ctx context.Context, encounterID string) error) *CombatController {
	return &CombatController{
		services:         services,
		combatLogUpdater: combatLogUpdater,
	}
}

// HandleMyTurn shows the ephemeral action controller for a player
func (cc *CombatController) HandleMyTurn(s *discordgo.Session, i *discordgo.InteractionCreate, encounterID string) error {
	log.Printf("CombatController.HandleMyTurn - User %s, Encounter %s", i.Member.User.ID, encounterID)
	
	// Get the encounter first to check if it exists
	encounter, err := cc.services.EncounterService.GetEncounter(context.Background(), encounterID)
	if err != nil {
		// Handle stale encounter IDs gracefully
		log.Printf("Failed to get encounter %s: %v", encounterID, err)
		content := "❌ This encounter no longer exists. It may have ended or the bot was restarted.\n\nPlease start a new encounter to continue playing."
		
		// Respond with error message
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags: discordgo.MessageFlagsEphemeral,
				Content: content,
			},
		})
		if err != nil {
			log.Printf("Failed to send error response: %v", err)
		}
		return nil // Don't propagate error since we handled it
	}
	
	// Now respond with the action controller
	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags: discordgo.MessageFlagsEphemeral,
			Content: "⏳ Loading your actions...",
		},
	})
	if err != nil {
		log.Printf("Failed to send initial response: %v", err)
		return err
	}
	
	// Verify it's the player's turn
	current := encounter.GetCurrentCombatant()
	if current == nil || current.PlayerID != i.Member.User.ID {
		content := "❌ It's not your turn!"
		_, editErr := s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: content,
		})
		if editErr != nil {
			log.Printf("Failed to send not your turn followup: %v", editErr)
		}
		return nil
	}
	
	// Build action controller
	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("🎮 %s's Action Controller", current.Name),
		Description: "It's your turn! Choose an action:",
		Color:       0x2ecc71,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "📊 Your Status",
				Value:  fmt.Sprintf("HP: %d/%d | AC: %d", current.CurrentHP, current.MaxHP, current.AC),
				Inline: false,
			},
		},
	}
	
	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "Attack",
					Style:    discordgo.DangerButton,
					CustomID: fmt.Sprintf("encounter:attack:%s", encounterID),
					Emoji:    &discordgo.ComponentEmoji{Name: "⚔️"},
				},
				discordgo.Button{
					Label:    "End Turn",
					Style:    discordgo.SecondaryButton,
					CustomID: fmt.Sprintf("encounter:end_turn:%s", encounterID),
					Emoji:    &discordgo.ComponentEmoji{Name: "➡️"},
				},
			},
		},
	}
	
	// Edit the initial message with the action controller
	_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content:    &[]string{""}[0], // Clear loading message
		Embeds:     &[]*discordgo.MessageEmbed{embed},
		Components: &components,
	})
	if err != nil {
		log.Printf("Failed to show action controller: %v", err)
		return err
	}
	
	return nil
}

// HandleAttack shows target selection
func (cc *CombatController) HandleAttack(s *discordgo.Session, i *discordgo.InteractionCreate, encounterID string) error {
	log.Printf("CombatController.HandleAttack - User %s, Encounter %s", i.Member.User.ID, encounterID)
	
	// Get encounter first to check if it exists
	encounter, err := cc.services.EncounterService.GetEncounter(context.Background(), encounterID)
	if err != nil {
		// Handle stale encounter IDs gracefully
		log.Printf("Failed to get encounter %s: %v", encounterID, err)
		content := "❌ This encounter no longer exists. It may have ended or the bot was restarted.\n\nPlease start a new encounter to continue playing."
		
		// Update the ephemeral message
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseUpdateMessage,
			Data: &discordgo.InteractionResponseData{
				Content: content,
				Components: []discordgo.MessageComponent{},
			},
		})
		if err != nil {
			log.Printf("Failed to send error response: %v", err)
		}
		return nil // Don't propagate error since we handled it
	}
	
	// This is from an ephemeral message, so update it
	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredMessageUpdate,
	})
	if err != nil {
		log.Printf("Failed to defer update: %v", err)
	}
	
	// Find the attacker
	var current *entities.Combatant
	for _, combatant := range encounter.Combatants {
		if combatant.PlayerID == i.Member.User.ID {
			current = combatant
			break
		}
	}
	
	if current == nil {
		content := "❌ You don't have a character in this encounter!"
		cc.updateEphemeralMessage(s, i, content, nil, nil)
		return nil
	}
	
	// Build target buttons
	var targetButtons []discordgo.MessageComponent
	for id, combatant := range encounter.Combatants {
		if combatant.ID == current.ID || !combatant.IsActive || combatant.CurrentHP <= 0 {
			continue
		}
		
		emoji := "🧑"
		if combatant.Type == entities.CombatantTypeMonster {
			emoji = "👹"
		}
		
		targetButtons = append(targetButtons, discordgo.Button{
			Label:    fmt.Sprintf("%s (HP: %d/%d)", combatant.Name, combatant.CurrentHP, combatant.MaxHP),
			Style:    discordgo.PrimaryButton,
			CustomID: fmt.Sprintf("encounter:select_target:%s:%s", encounterID, id),
			Emoji:    &discordgo.ComponentEmoji{Name: emoji},
		})
		
		if len(targetButtons) >= 5 {
			break
		}
	}
	
	if len(targetButtons) == 0 {
		content := "❌ No valid targets available!"
		cc.updateEphemeralMessage(s, i, content, nil, nil)
		return nil
	}
	
	// Add back button
	backButton := discordgo.Button{
		Label:    "Back",
		Style:    discordgo.SecondaryButton,
		CustomID: fmt.Sprintf("encounter:my_turn:%s", encounterID),
		Emoji:    &discordgo.ComponentEmoji{Name: "⬅️"},
	}
	
	if len(targetButtons) < 5 {
		targetButtons = append(targetButtons, backButton)
	}
	
	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("🎮 %s's Action Controller", current.Name),
		Description: "**Select your target:**",
		Color:       0xe74c3c,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "📊 Your Status",
				Value:  fmt.Sprintf("HP: %d/%d | AC: %d", current.CurrentHP, current.MaxHP, current.AC),
				Inline: false,
			},
		},
	}
	
	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: targetButtons,
		},
	}
	
	cc.updateEphemeralMessage(s, i, "", []*discordgo.MessageEmbed{embed}, components)
	return nil
}

// HandleContinueRound advances to the next round
func (cc *CombatController) HandleContinueRound(s *discordgo.Session, i *discordgo.InteractionCreate, encounterID string) error {
	log.Printf("CombatController.HandleContinueRound - User %s, Encounter %s", i.Member.User.ID, encounterID)
	
	// Check if encounter exists first
	_, err := cc.services.EncounterService.GetEncounter(context.Background(), encounterID)
	if err != nil {
		// Handle stale encounter IDs gracefully
		log.Printf("Failed to get encounter %s: %v", encounterID, err)
		content := "❌ This encounter no longer exists. It may have ended or the bot was restarted.\n\nPlease start a new encounter to continue playing."
		
		// Update the message to show error
		_, editErr := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: &content,
			Embeds: &[]*discordgo.MessageEmbed{},
			Components: &[]discordgo.MessageComponent{},
		})
		if editErr != nil {
			// If edit fails, try responding normally
			_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseUpdateMessage,
				Data: &discordgo.InteractionResponseData{
					Content: content,
					Components: []discordgo.MessageComponent{},
				},
			})
		}
		return nil // Don't propagate error since we handled it
	}
	
	// This updates the public message, so defer update
	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredMessageUpdate,
	})
	if err != nil {
		log.Printf("Failed to defer update: %v", err)
	}
	
	// Continue the round
	err = cc.services.EncounterService.ContinueRound(context.Background(), encounterID, i.Member.User.ID)
	if err != nil {
		log.Printf("Failed to continue round: %v", err)
		return err
	}
	
	// Get updated encounter
	encounter, err := cc.services.EncounterService.GetEncounter(context.Background(), encounterID)
	if err != nil {
		log.Printf("Failed to get encounter after continue: %v", err)
		return err
	}
	
	// Process monster turns at the start of the new round
	botID := s.State.User.ID
	for encounter.Status == entities.EncounterStatusActive && !encounter.RoundPending {
		current := encounter.GetCurrentCombatant()
		if current == nil || current.Type != entities.CombatantTypeMonster {
			break // Stop when we reach a player's turn
		}
		
		log.Printf("Processing monster turn for %s in round %d", current.Name, encounter.Round)
		
		// Find a target (first active player)
		var target *entities.Combatant
		for _, combatant := range encounter.Combatants {
			if combatant.Type == entities.CombatantTypePlayer && combatant.IsActive {
				target = combatant
				break
			}
		}
		
		if target != nil && len(current.Actions) > 0 {
			// Use first available action
			action := current.Actions[0]
			
			// Roll attack
			attackResult, _ := dice.Roll(1, 20, 0)
			attackRoll := attackResult.Total
			totalAttack := attackRoll + action.AttackBonus
			
			// Check if hit
			hit := totalAttack >= target.AC
			if hit && len(action.Damage) > 0 {
				totalDamage := 0
				var damageDetails []string
				for _, dmg := range action.Damage {
					diceCount := dmg.DiceCount
					if attackRoll == 20 { // Critical hit doubles dice
						diceCount *= 2
					}
					rollResult, _ := dice.Roll(diceCount, dmg.DiceSize, dmg.Bonus)
					totalDamage += rollResult.Total
					
					// Build damage notation string
					damageStr := fmt.Sprintf("%dd%d", diceCount, dmg.DiceSize)
					if dmg.Bonus != 0 {
						damageStr += fmt.Sprintf("%+d", dmg.Bonus)
					}
					rollsStr := fmt.Sprintf("%v", rollResult.Rolls)
					damageStr += fmt.Sprintf("=%s", rollsStr)
					if dmg.DamageType != "" {
						damageStr += fmt.Sprintf(" %s", dmg.DamageType)
					}
					damageDetails = append(damageDetails, damageStr)
				}
				
				// Apply damage
				err = cc.services.EncounterService.ApplyDamage(context.Background(), encounter.ID, target.ID, botID, totalDamage)
				if err != nil {
					log.Printf("Error applying monster damage: %v", err)
				}
				
				// Log the action
				attackMsg := fmt.Sprintf("%s attacks %s with %s: Attack [%d]+%d=%d vs AC %d (HIT!) | Damage %s = %d total",
					current.Name, target.Name, action.Name,
					attackRoll, action.AttackBonus, totalAttack, target.AC,
					strings.Join(damageDetails, " + "), totalDamage)
				
				_ = cc.services.EncounterService.LogCombatAction(context.Background(), encounter.ID, attackMsg)
			} else {
				// Log miss
				missMsg := fmt.Sprintf("%s attacks %s with %s: Attack [%d]+%d=%d vs AC %d (MISS!)",
					current.Name, target.Name, action.Name, attackRoll, action.AttackBonus, totalAttack, target.AC)
				_ = cc.services.EncounterService.LogCombatAction(context.Background(), encounter.ID, missMsg)
			}
		}
		
		// Advance turn
		err = cc.services.EncounterService.NextTurn(context.Background(), encounter.ID, botID)
		if err != nil {
			log.Printf("Error advancing monster turn: %v", err)
			break
		}
		
		// Re-get encounter for next iteration
		encounter, err = cc.services.EncounterService.GetEncounter(context.Background(), encounter.ID)
		if err != nil {
			log.Printf("Error getting encounter after monster turn: %v", err)
			break
		}
	}
	
	// Update the public combat log after all monster turns
	if cc.combatLogUpdater != nil {
		if updateErr := cc.combatLogUpdater(context.Background(), encounterID); updateErr != nil {
			log.Printf("Failed to update combat log after round continue: %v", updateErr)
		}
	}
	
	return nil
}

// HandleSelectTarget processes the attack when a target is selected
func (cc *CombatController) HandleSelectTarget(s *discordgo.Session, i *discordgo.InteractionCreate, encounterID, targetID string) error {
	log.Printf("CombatController.HandleSelectTarget - User %s, Encounter %s, Target %s", i.Member.User.ID, encounterID, targetID)
	
	// Get encounter first to check if it exists
	encounter, err := cc.services.EncounterService.GetEncounter(context.Background(), encounterID)
	if err != nil {
		// Handle stale encounter IDs gracefully
		log.Printf("Failed to get encounter %s: %v", encounterID, err)
		content := "❌ This encounter no longer exists. It may have ended or the bot was restarted.\n\nPlease start a new encounter to continue playing."
		
		// Update the ephemeral message
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseUpdateMessage,
			Data: &discordgo.InteractionResponseData{
				Content: content,
				Components: []discordgo.MessageComponent{},
			},
		})
		if err != nil {
			log.Printf("Failed to send error response: %v", err)
		}
		return nil // Don't propagate error since we handled it
	}
	
	// This is from an ephemeral message, so defer update
	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredMessageUpdate,
	})
	if err != nil {
		log.Printf("Failed to defer update: %v", err)
	}
	
	// Find the attacker
	var attacker *entities.Combatant
	for _, combatant := range encounter.Combatants {
		if combatant.PlayerID == i.Member.User.ID {
			attacker = combatant
			break
		}
	}
	
	if attacker == nil {
		content := "❌ You don't have a character in this encounter!"
		cc.updateEphemeralMessage(s, i, content, nil, nil)
		return nil
	}
	
	// Get target
	target, exists := encounter.Combatants[targetID]
	if !exists {
		content := "❌ Target not found!"
		cc.updateEphemeralMessage(s, i, content, nil, nil)
		return nil
	}
	
	// Execute the attack
	log.Printf("Executing attack from %s to %s", attacker.Name, target.Name)
	
	// Perform the attack
	var attackResults []*attack.Result
	var attackName string
	
	if attacker.Type == entities.CombatantTypePlayer && attacker.CharacterID != "" {
		// Get character for player attacks
		char, err := cc.services.CharacterService.GetByID(attacker.CharacterID)
		if err == nil && char != nil {
			attackResults, _ = char.Attack()
			// Get weapon name
			if char.EquippedSlots[entities.SlotMainHand] != nil {
				attackName = char.EquippedSlots[entities.SlotMainHand].GetName()
			} else if char.EquippedSlots[entities.SlotTwoHanded] != nil {
				attackName = char.EquippedSlots[entities.SlotTwoHanded].GetName()
			} else {
				attackName = "Unarmed Strike"
			}
		}
	} else if attacker.Type == entities.CombatantTypeMonster && len(attacker.Actions) > 0 {
		// Monster attacks
		action := attacker.Actions[0]
		attackName = action.Name
		attackRoll, _ := dice.Roll(1, 20, 0)
		totalAttack := attackRoll.Total + action.AttackBonus
		
		totalDamage := 0
		damageType := damage.TypeBludgeoning
		var damageResult *dice.RollResult
		
		// For monsters, we'll track the first damage roll for display
		// In D&D, most monsters have a single damage type per action
		for i, dmg := range action.Damage {
			dmgRoll, _ := dice.Roll(dmg.DiceCount, dmg.DiceSize, dmg.Bonus)
			totalDamage += dmgRoll.Total
			
			// Keep the first damage result for display
			if i == 0 {
				damageResult = dmgRoll
			}
			
			if dmg.DamageType != "" {
				damageType = dmg.DamageType
			}
		}
		
		attackResults = []*attack.Result{{
			AttackRoll:   totalAttack,
			DamageRoll:   totalDamage,
			AttackType:   damageType,
			AttackResult: attackRoll,
			DamageResult: damageResult,
		}}
	}
	
	// Default to unarmed strike if no attack results
	if len(attackResults) == 0 {
		attackName = "Unarmed Strike"
		attackRoll, _ := dice.Roll(1, 20, 0)
		damageRoll, _ := dice.Roll(1, 4, 0)
		attackResults = []*attack.Result{{
			AttackRoll:   attackRoll.Total,
			DamageRoll:   damageRoll.Total,
			AttackType:   damage.TypeBludgeoning,
			AttackResult: attackRoll,
			DamageResult: damageRoll,
		}}
	}
	
	// Build result embed
	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("🎮 %s's Action Controller", attacker.Name),
		Description: fmt.Sprintf("**Attack Result:** %s vs %s", attackName, target.Name),
		Color:       0x2ecc71,
		Fields:      []*discordgo.MessageEmbedField{},
	}
	
	// Add status field
	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
		Name:   "📊 Your Status",
		Value:  fmt.Sprintf("HP: %d/%d | AC: %d", attacker.CurrentHP, attacker.MaxHP, attacker.AC),
		Inline: false,
	})
	
	// Process attack results
	totalDamageDealt := 0
	for _, result := range attackResults {
		hit := result.AttackRoll >= target.AC
		var hitText string
		
		if result.AttackResult != nil && result.AttackResult.Total == 20 {
			hitText = "🎆 **CRITICAL HIT!**"
			hit = true
		} else if result.AttackResult != nil && result.AttackResult.Total == 1 {
			hitText = "⚠️ **CRITICAL MISS!**"
			hit = false
		} else if hit {
			hitText = "✅ **HIT!**"
		} else {
			hitText = "❌ **MISS!**"
		}
		
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "🎲 Attack Roll",
			Value:  fmt.Sprintf("%d vs AC %d\n%s", result.AttackRoll, target.AC, hitText),
			Inline: true,
		})
		
		if hit {
			// Build damage display with dice rolls if available
			damageDisplay := fmt.Sprintf("%d %s damage", result.DamageRoll, result.AttackType)
			if result.DamageResult != nil && len(result.DamageResult.Rolls) > 0 {
				// Show the actual dice rolls
				rollsStr := fmt.Sprintf("%v", result.DamageResult.Rolls)
				if result.DamageResult.Bonus != 0 {
					if result.DamageResult.Bonus > 0 {
						rollsStr += fmt.Sprintf("+%d", result.DamageResult.Bonus)
					} else {
						rollsStr += fmt.Sprintf("%d", result.DamageResult.Bonus)
					}
				}
				damageDisplay = fmt.Sprintf("%s = **%d** %s damage", rollsStr, result.DamageRoll, result.AttackType)
			}
			
			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
				Name:   "💥 Damage",
				Value:  damageDisplay,
				Inline: true,
			})
			totalDamageDealt += result.DamageRoll
		}
	}
	
	// Apply damage if any
	if totalDamageDealt > 0 {
		err = cc.services.EncounterService.ApplyDamage(context.Background(), encounterID, targetID, i.Member.User.ID, totalDamageDealt)
		if err != nil {
			log.Printf("Error applying damage: %v", err)
		}
		
		// Add target status
		target.CurrentHP -= totalDamageDealt
		if target.CurrentHP < 0 {
			target.CurrentHP = 0
		}
		
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "🩸 Target Status",
			Value:  fmt.Sprintf("%s now has **%d/%d HP**", target.Name, target.CurrentHP, target.MaxHP),
			Inline: false,
		})
		
		if target.CurrentHP == 0 {
			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
				Name:   "💀 Defeated!",
				Value:  fmt.Sprintf("%s has been defeated!", target.Name),
				Inline: false,
			})
		}
	}
	
	// Log the action with dice details
	for _, result := range attackResults {
		var logEntry string
		if result.AttackRoll >= target.AC {
			// Build damage notation
			damageNotation := fmt.Sprintf("%d", result.DamageRoll)
			if result.DamageResult != nil && len(result.DamageResult.Rolls) > 0 {
				// Show actual rolls
				rollsStr := fmt.Sprintf("%v", result.DamageResult.Rolls)
				if result.DamageResult.Bonus != 0 {
					if result.DamageResult.Bonus > 0 {
						rollsStr += fmt.Sprintf("+%d", result.DamageResult.Bonus)
					} else {
						rollsStr += fmt.Sprintf("%d", result.DamageResult.Bonus)
					}
				}
				damageNotation = fmt.Sprintf("%s=%d", rollsStr, result.DamageRoll)
			}
			
			// Include attack roll details
			attackNotation := ""
			if result.AttackResult != nil {
				bonus := result.AttackRoll - result.AttackResult.Total
				if bonus > 0 {
					attackNotation = fmt.Sprintf(" (Attack: [%d]+%d=%d vs AC %d)", 
						result.AttackResult.Total, bonus, result.AttackRoll, target.AC)
				} else {
					attackNotation = fmt.Sprintf(" (Attack: %d vs AC %d)", result.AttackRoll, target.AC)
				}
			}
			
			logEntry = fmt.Sprintf("⚔️ **%s** hits **%s** with %s for %s %s damage!%s",
				attacker.Name, target.Name, attackName, damageNotation, result.AttackType, attackNotation)
		} else {
			// Include miss details
			attackNotation := ""
			if result.AttackResult != nil {
				bonus := result.AttackRoll - result.AttackResult.Total
				if bonus > 0 {
					attackNotation = fmt.Sprintf(" (Attack: [%d]+%d=%d vs AC %d)", 
						result.AttackResult.Total, bonus, result.AttackRoll, target.AC)
				} else {
					attackNotation = fmt.Sprintf(" (Attack: %d vs AC %d)", result.AttackRoll, target.AC)
				}
			}
			
			logEntry = fmt.Sprintf("⚔️ **%s** misses **%s** with %s!%s",
				attacker.Name, target.Name, attackName, attackNotation)
		}
		_ = cc.services.EncounterService.LogCombatAction(context.Background(), encounterID, logEntry)
	}
	
	// Auto-advance turn
	err = cc.services.EncounterService.NextTurn(context.Background(), encounterID, i.Member.User.ID)
	if err != nil {
		log.Printf("Error advancing turn: %v", err)
	}
	
	// Update the public combat log after turn advancement
	if cc.combatLogUpdater != nil {
		if updateErr := cc.combatLogUpdater(context.Background(), encounterID); updateErr != nil {
			log.Printf("Failed to update combat log after turn advance: %v", updateErr)
		}
	}
	
	// Check if round is pending after turn advance
	encounter, _ = cc.services.EncounterService.GetEncounter(context.Background(), encounterID)
	
	// Build action buttons
	var components []discordgo.MessageComponent
	if encounter != nil && encounter.RoundPending {
		// Show continue button
		components = []discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.Button{
						Label:    fmt.Sprintf("Continue to Round %d", encounter.Round+1),
						Style:    discordgo.SuccessButton,
						CustomID: fmt.Sprintf("encounter:continue_round:%s", encounterID),
						Emoji:    &discordgo.ComponentEmoji{Name: "➡️"},
					},
				},
			},
		}
		
		// Add round complete message
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "🏁 Round Complete!",
			Value:  fmt.Sprintf("**Round %d has ended.** Click 'Continue' to proceed.", encounter.Round),
			Inline: false,
		})
	} else {
		// Normal action buttons
		components = []discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.Button{
						Label:    "Back to Actions",
						Style:    discordgo.PrimaryButton,
						CustomID: fmt.Sprintf("encounter:my_turn:%s", encounterID),
						Emoji:    &discordgo.ComponentEmoji{Name: "🎮"},
					},
					discordgo.Button{
						Label:    "End Turn",
						Style:    discordgo.SecondaryButton,
						CustomID: fmt.Sprintf("encounter:end_turn:%s", encounterID),
						Emoji:    &discordgo.ComponentEmoji{Name: "➡️"},
					},
				},
			},
		}
	}
	
	cc.updateEphemeralMessage(s, i, "", []*discordgo.MessageEmbed{embed}, components)
	return nil
}

// HandleEndTurn advances to the next turn
func (cc *CombatController) HandleEndTurn(s *discordgo.Session, i *discordgo.InteractionCreate, encounterID string) error {
	log.Printf("CombatController.HandleEndTurn - User %s, Encounter %s", i.Member.User.ID, encounterID)
	
	// Check if encounter exists first
	_, err := cc.services.EncounterService.GetEncounter(context.Background(), encounterID)
	if err != nil {
		// Handle stale encounter IDs gracefully
		log.Printf("Failed to get encounter %s: %v", encounterID, err)
		content := "❌ This encounter no longer exists. It may have ended or the bot was restarted.\n\nPlease start a new encounter to continue playing."
		
		// Update the ephemeral message
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseUpdateMessage,
			Data: &discordgo.InteractionResponseData{
				Content: content,
				Components: []discordgo.MessageComponent{},
			},
		})
		if err != nil {
			log.Printf("Failed to send error response: %v", err)
		}
		return nil // Don't propagate error since we handled it
	}
	
	// This is from an ephemeral message, so defer update
	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredMessageUpdate,
	})
	if err != nil {
		log.Printf("Failed to defer update: %v", err)
	}
	
	// Advance turn
	err = cc.services.EncounterService.NextTurn(context.Background(), encounterID, i.Member.User.ID)
	if err != nil {
		content := fmt.Sprintf("❌ Failed to advance turn: %v", err)
		cc.updateEphemeralMessage(s, i, content, nil, nil)
		return err
	}
	
	// Get updated encounter
	encounter, err := cc.services.EncounterService.GetEncounter(context.Background(), encounterID)
	if err != nil {
		content := fmt.Sprintf("❌ Failed to get encounter: %v", err)
		cc.updateEphemeralMessage(s, i, content, nil, nil)
		return err
	}
	
	// Check if round is pending
	if encounter.RoundPending {
		// Show round complete message
		embed := &discordgo.MessageEmbed{
			Title:       "🏁 Round Complete!",
			Description: fmt.Sprintf("**Round %d has ended.**", encounter.Round),
			Color:       0x3498db,
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:   "⏭️ Next Round",
					Value:  "Click the button below to continue to the next round.",
					Inline: false,
				},
			},
		}
		
		components := []discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.Button{
						Label:    fmt.Sprintf("Continue to Round %d", encounter.Round+1),
						Style:    discordgo.SuccessButton,
						CustomID: fmt.Sprintf("encounter:continue_round:%s", encounterID),
						Emoji:    &discordgo.ComponentEmoji{Name: "➡️"},
					},
				},
			},
		}
		
		cc.updateEphemeralMessage(s, i, "", []*discordgo.MessageEmbed{embed}, components)
	} else {
		// Turn advanced normally
		content := "✅ Turn ended!"
		cc.updateEphemeralMessage(s, i, content, nil, nil)
	}
	
	return nil
}

// Helper to update ephemeral messages
func (cc *CombatController) updateEphemeralMessage(s *discordgo.Session, i *discordgo.InteractionCreate, content string, embeds []*discordgo.MessageEmbed, components []discordgo.MessageComponent) {
	edit := &discordgo.WebhookEdit{}
	
	if content != "" {
		edit.Content = &content
	}
	if embeds != nil {
		edit.Embeds = &embeds
	}
	if components != nil {
		edit.Components = &components
	}
	
	_, err := s.InteractionResponseEdit(i.Interaction, edit)
	if err != nil {
		log.Printf("Failed to update ephemeral message: %v", err)
	}
}