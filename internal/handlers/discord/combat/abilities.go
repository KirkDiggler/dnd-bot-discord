package combat

import (
	"context"
	"fmt"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/game/combat"
	"strconv"
	"strings"

	"github.com/KirkDiggler/dnd-bot-discord/internal/services/ability"
	"github.com/bwmarrin/discordgo"
)

// handleShowAbilities displays available abilities for the player
func (h *Handler) handleShowAbilities(s *discordgo.Session, i *discordgo.InteractionCreate, encounterID string) error {
	// Get encounter
	enc, err := h.encounterService.GetEncounter(context.Background(), encounterID)
	if err != nil {
		return respondError(s, i, "Failed to get encounter", err)
	}

	// Find the player's combatant
	var playerCombatant *combat.Combatant
	for _, c := range enc.Combatants {
		if c.PlayerID == i.Member.User.ID && c.IsActive {
			playerCombatant = c
			break
		}
	}

	if playerCombatant == nil {
		return respondError(s, i, "You are not in this combat!", nil)
	}

	// Get available abilities
	abilities, err := h.abilityService.GetAvailableAbilities(context.Background(), playerCombatant.CharacterID)
	if err != nil {
		return respondError(s, i, "Failed to get abilities", err)
	}

	// Build ability list
	var abilityList []string
	var buttons []discordgo.MessageComponent

	for _, avail := range abilities {
		ab := avail.Ability

		// Build status string
		status := "✅"
		statusText := "Ready"
		if !avail.Available {
			status = "❌"
			statusText = avail.Reason
		}

		// Add to list
		abilityList = append(abilityList, fmt.Sprintf("%s **%s** - %s\n   *%s* | %s",
			status, ab.Name, ab.Description, formatActionType(ab.ActionType), statusText))

		// Add button if available
		if avail.Available && len(buttons) < 5 {
			buttons = append(buttons, discordgo.Button{
				Label:    ab.Name,
				Style:    discordgo.PrimaryButton,
				CustomID: fmt.Sprintf("combat:use_ability:%s:%s", encounterID, ab.Key),
				Emoji:    &discordgo.ComponentEmoji{Name: getAbilityEmoji(ab.Key)},
			})
		}
	}

	// Build embed
	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("✨ %s's Abilities", playerCombatant.Name),
		Color:       0x9b59b6, // Purple
		Description: strings.Join(abilityList, "\n\n"),
	}

	if len(abilityList) == 0 {
		embed.Description = "You have no special abilities available."
	}

	// Add footer
	embed.Footer = &discordgo.MessageEmbedFooter{
		Text: "Select an ability to use it",
	}

	// Build components
	components := []discordgo.MessageComponent{}
	if len(buttons) > 0 {
		components = append(components, discordgo.ActionsRow{
			Components: buttons,
		})
	}

	// Always add back button
	components = append(components, discordgo.ActionsRow{
		Components: []discordgo.MessageComponent{
			discordgo.Button{
				Label:    "Back to Actions",
				Style:    discordgo.SecondaryButton,
				CustomID: fmt.Sprintf("combat:my_actions:%s", encounterID),
				Emoji:    &discordgo.ComponentEmoji{Name: "🎯"},
			},
		},
	})

	// Update the message if from ephemeral, otherwise create new ephemeral
	if isEphemeralInteraction(i) {
		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseUpdateMessage,
			Data: &discordgo.InteractionResponseData{
				Embeds:     []*discordgo.MessageEmbed{embed},
				Components: components,
			},
		})
	}

	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds:     []*discordgo.MessageEmbed{embed},
			Components: components,
			Flags:      discordgo.MessageFlagsEphemeral,
		},
	})
}

// handleUseAbility executes an ability
func (h *Handler) handleUseAbility(s *discordgo.Session, i *discordgo.InteractionCreate, encounterID string) error {
	// Parse ability key from custom ID: combat:use_ability:encounterID:abilityKey
	parts := parseCustomID(i.MessageComponentData().CustomID)
	if len(parts) < 4 {
		return respondError(s, i, "Invalid ability selection", nil)
	}
	abilityKey := parts[3]

	// Get encounter
	enc, err := h.encounterService.GetEncounter(context.Background(), encounterID)
	if err != nil {
		return respondError(s, i, "Failed to get encounter", err)
	}

	// Find the player's combatant
	var playerCombatant *combat.Combatant
	for _, c := range enc.Combatants {
		if c.PlayerID == i.Member.User.ID && c.IsActive {
			playerCombatant = c
			break
		}
	}

	if playerCombatant == nil {
		return respondError(s, i, "You are not in this combat!", nil)
	}

	// Check if it's their turn
	current := enc.GetCurrentCombatant()
	if current == nil || current.ID != playerCombatant.ID {
		// Not their turn - show a friendly message
		var currentName string
		if current == nil {
			currentName = "Unknown"
		} else {
			currentName = current.Name
		}

		embed := &discordgo.MessageEmbed{
			Title:       "⏳ Not Your Turn",
			Description: fmt.Sprintf("It's currently **%s's** turn.", currentName),
			Color:       0xf39c12, // Orange
			Footer: &discordgo.MessageEmbedFooter{
				Text: "Wait for your turn to use abilities",
			},
		}

		components := []discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.Button{
						Label:    "Back to Abilities",
						Style:    discordgo.PrimaryButton,
						CustomID: fmt.Sprintf("combat:abilities:%s", encounterID),
						Emoji:    &discordgo.ComponentEmoji{Name: "✨"},
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

		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseUpdateMessage,
			Data: &discordgo.InteractionResponseData{
				Embeds:     []*discordgo.MessageEmbed{embed},
				Components: components,
			},
		})
	}

	// For targeted abilities, we might need to show target selection
	// For now, we'll handle self-targeted and instant abilities
	targetID := ""
	if needsTarget(abilityKey) {
		// TODO: Show target selection UI
		// For now, default to self-target for healing abilities
		if isHealingAbility(abilityKey) {
			targetID = playerCombatant.CharacterID
		}
	}

	// Special handling for abilities that need value selection
	if abilityKey == "lay-on-hands" {
		// Show heal amount selection UI
		return h.showLayOnHandsAmountSelection(s, i, encounterID, playerCombatant)
	}

	// Use the ability
	result, err := h.abilityService.UseAbility(context.Background(), &ability.UseAbilityInput{
		CharacterID: playerCombatant.CharacterID,
		AbilityKey:  abilityKey,
		TargetID:    targetID,
		EncounterID: encounterID,
		Value:       0, // Default value for abilities that don't need it
	})
	if err != nil {
		return respondError(s, i, "Failed to use ability", err)
	}

	// Build result embed
	var embed *discordgo.MessageEmbed
	if result.Success {
		embed = &discordgo.MessageEmbed{
			Title:       fmt.Sprintf("✨ %s Used!", getAbilityName(abilityKey)),
			Description: result.Message,
			Color:       0x00ff00, // Green
			Fields:      []*discordgo.MessageEmbedField{},
		}

		// Add effect info if applicable
		if result.EffectApplied {
			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
				Name:   "Effect Applied",
				Value:  fmt.Sprintf("%s (Duration: %d rounds)", result.EffectName, result.Duration),
				Inline: true,
			})
		}

		// Add healing info if applicable
		if result.HealingDone > 0 {
			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
				Name:   "Healing",
				Value:  fmt.Sprintf("%d HP restored (New HP: %d)", result.HealingDone, result.TargetNewHP),
				Inline: true,
			})
		}

		// Add uses remaining
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "Uses Remaining",
			Value:  fmt.Sprintf("%d", result.UsesRemaining),
			Inline: true,
		})

		// Add to combat log
		if err := h.encounterService.LogCombatAction(context.Background(), encounterID,
			fmt.Sprintf("**%s** used **%s**: %s", playerCombatant.Name, getAbilityName(abilityKey), result.Message)); err != nil {
			// Log the error but don't fail the ability use
			fmt.Printf("Failed to log combat action: %v\n", err)
		}
	} else {
		embed = &discordgo.MessageEmbed{
			Title:       "❌ Ability Failed",
			Description: result.Message,
			Color:       0xff0000, // Red
		}
	}

	// Add footer
	embed.Footer = &discordgo.MessageEmbedFooter{
		Text: "Ability used successfully",
	}

	// Components for navigation
	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "End Turn",
					Style:    discordgo.SuccessButton,
					CustomID: fmt.Sprintf("combat:next_turn:%s", encounterID),
					Emoji:    &discordgo.ComponentEmoji{Name: "✅"},
					Disabled: !result.Success,
				},
				discordgo.Button{
					Label:    "More Abilities",
					Style:    discordgo.PrimaryButton,
					CustomID: fmt.Sprintf("combat:abilities:%s", encounterID),
					Emoji:    &discordgo.ComponentEmoji{Name: "✨"},
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

	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Embeds:     []*discordgo.MessageEmbed{embed},
			Components: components,
		},
	})
}

// Helper functions

func formatActionType(actionType character.AbilityType) string {
	switch actionType {
	case character.AbilityTypeAction:
		return "Action"
	case character.AbilityTypeBonusAction:
		return "Bonus Action"
	case character.AbilityTypeReaction:
		return "Reaction"
	case character.AbilityTypeFree:
		return "Free Action"
	default:
		return string(actionType)
	}
}

func getAbilityEmoji(abilityKey string) string {
	emojiMap := map[string]string{
		"rage":               "😡",
		"second-wind":        "💨",
		"bardic-inspiration": "🎵",
		"lay-on-hands":       "🙏",
		"divine-sense":       "👁️",
	}

	if emoji, exists := emojiMap[abilityKey]; exists {
		return emoji
	}
	return "✨"
}

func getAbilityName(abilityKey string) string {
	nameMap := map[string]string{
		"rage":               "Rage",
		"second-wind":        "Second Wind",
		"bardic-inspiration": "Bardic Inspiration",
		"lay-on-hands":       "Lay on Hands",
		"divine-sense":       "Divine Sense",
	}

	if name, exists := nameMap[abilityKey]; exists {
		return name
	}
	return abilityKey
}

func needsTarget(abilityKey string) bool {
	// Abilities that need target selection
	targetedAbilities := []string{
		"bardic-inspiration",
		"lay-on-hands",
	}

	for _, ab := range targetedAbilities {
		if ab == abilityKey {
			return true
		}
	}
	return false
}

func isHealingAbility(abilityKey string) bool {
	healingAbilities := []string{
		"lay-on-hands",
		"second-wind",
	}

	for _, ab := range healingAbilities {
		if ab == abilityKey {
			return true
		}
	}
	return false
}

// showLayOnHandsAmountSelection shows UI for selecting heal amount
func (h *Handler) showLayOnHandsAmountSelection(s *discordgo.Session, i *discordgo.InteractionCreate, encounterID string, playerCombatant *combat.Combatant) error {
	// Get character to check available healing pool
	char, err := h.abilityService.GetAvailableAbilities(context.Background(), playerCombatant.CharacterID)
	if err != nil {
		return respondError(s, i, "Failed to get character abilities", err)
	}

	// Find lay on hands ability to check remaining pool
	var layOnHands *character.ActiveAbility
	for _, avail := range char {
		if avail.Ability.Key == "lay-on-hands" {
			layOnHands = avail.Ability
			break
		}
	}

	if layOnHands == nil {
		return respondError(s, i, "Lay on Hands ability not found", nil)
	}

	// Create buttons for common heal amounts
	var buttons []discordgo.MessageComponent
	healAmounts := []int{1, 2, 3, 5}

	// Add the full remaining pool if it's not already in the list
	if layOnHands.UsesRemaining > 5 {
		healAmounts = append(healAmounts, layOnHands.UsesRemaining)
	}

	// Use a map to track which amounts we've already added to avoid duplicates
	addedAmounts := make(map[int]bool)

	for _, amount := range healAmounts {
		if amount > 0 && amount <= layOnHands.UsesRemaining && !addedAmounts[amount] {
			addedAmounts[amount] = true
			buttons = append(buttons, discordgo.Button{
				Label:    fmt.Sprintf("%d HP", amount),
				Style:    discordgo.PrimaryButton,
				CustomID: fmt.Sprintf("combat:lay_on_hands_amount:%s:%d", encounterID, amount),
			})
		}
	}

	embed := &discordgo.MessageEmbed{
		Title:       "🙏 Lay on Hands",
		Description: fmt.Sprintf("Select how many hit points to heal (Pool: %d/%d)", layOnHands.UsesRemaining, layOnHands.UsesMax),
		Color:       0x3498db, // Blue
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Choose the amount of healing",
		},
	}

	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: buttons,
		},
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "Cancel",
					Style:    discordgo.SecondaryButton,
					CustomID: fmt.Sprintf("combat:abilities:%s", encounterID),
					Emoji:    &discordgo.ComponentEmoji{Name: "❌"},
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

// handleLayOnHandsAmount processes the selected heal amount
func (h *Handler) handleLayOnHandsAmount(s *discordgo.Session, i *discordgo.InteractionCreate, encounterID string) error {
	// Parse amount from custom ID: combat:lay_on_hands_amount:encounterID:amount
	parts := parseCustomID(i.MessageComponentData().CustomID)
	if len(parts) < 4 {
		return respondError(s, i, "Invalid heal amount selection", nil)
	}

	amount, err := strconv.Atoi(parts[3])
	if err != nil {
		return respondError(s, i, "Invalid heal amount", err)
	}

	// Get encounter
	enc, err := h.encounterService.GetEncounter(context.Background(), encounterID)
	if err != nil {
		return respondError(s, i, "Failed to get encounter", err)
	}

	// Find the player's combatant
	var playerCombatant *combat.Combatant
	for _, c := range enc.Combatants {
		if c.PlayerID == i.Member.User.ID && c.IsActive {
			playerCombatant = c
			break
		}
	}

	if playerCombatant == nil {
		return respondError(s, i, "You are not in this combat!", nil)
	}

	// Use lay on hands with the selected amount
	result, err := h.abilityService.UseAbility(context.Background(), &ability.UseAbilityInput{
		CharacterID: playerCombatant.CharacterID,
		AbilityKey:  "lay-on-hands",
		TargetID:    playerCombatant.CharacterID, // Self-target for now
		EncounterID: encounterID,
		Value:       amount,
	})
	if err != nil {
		return respondError(s, i, "Failed to use Lay on Hands", err)
	}

	// Build result embed (reuse existing logic)
	var embed *discordgo.MessageEmbed
	if result.Success {
		embed = &discordgo.MessageEmbed{
			Title:       fmt.Sprintf("✨ %s Used!", "Lay on Hands"),
			Description: result.Message,
			Color:       0x00ff00, // Green
			Fields:      []*discordgo.MessageEmbedField{},
		}

		if result.HealingDone > 0 {
			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
				Name:   "Healing",
				Value:  fmt.Sprintf("%d HP restored (New HP: %d)", result.HealingDone, result.TargetNewHP),
				Inline: true,
			})
		}

		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "Pool Remaining",
			Value:  fmt.Sprintf("%d HP", result.UsesRemaining),
			Inline: true,
		})

		// Add to combat log
		if err := h.encounterService.LogCombatAction(context.Background(), encounterID,
			fmt.Sprintf("**%s** used **Lay on Hands**: %s", playerCombatant.Name, result.Message)); err != nil {
			// Log the error but don't fail the ability use
			fmt.Printf("Failed to log combat action: %v\n", err)
		}
	} else {
		embed = &discordgo.MessageEmbed{
			Title:       "❌ Lay on Hands Failed",
			Description: result.Message,
			Color:       0xff0000, // Red
		}
	}

	// Components for navigation
	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "End Turn",
					Style:    discordgo.SuccessButton,
					CustomID: fmt.Sprintf("combat:next_turn:%s", encounterID),
					Emoji:    &discordgo.ComponentEmoji{Name: "✅"},
					Disabled: !result.Success,
				},
				discordgo.Button{
					Label:    "More Abilities",
					Style:    discordgo.PrimaryButton,
					CustomID: fmt.Sprintf("combat:abilities:%s", encounterID),
					Emoji:    &discordgo.ComponentEmoji{Name: "✨"},
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

	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Embeds:     []*discordgo.MessageEmbed{embed},
			Components: components,
		},
	})
}
