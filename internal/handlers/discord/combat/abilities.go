package combat

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/game/combat"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/shared"
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

	// Get the ability details to check targeting requirements
	abilities, err := h.abilityService.GetAvailableAbilities(context.Background(), playerCombatant.CharacterID)
	if err != nil {
		return respondError(s, i, "Failed to get abilities", err)
	}

	var selectedAbility *ability.AvailableAbility
	for _, ab := range abilities {
		if ab.Ability.Key == abilityKey {
			selectedAbility = ab
			break
		}
	}

	if selectedAbility == nil {
		return respondError(s, i, "Ability not found", nil)
	}

	// Check if ability needs target selection
	if selectedAbility.Ability.Targeting != nil &&
		selectedAbility.Ability.Targeting.TargetType != shared.TargetTypeSelf &&
		selectedAbility.Ability.Targeting.TargetType != shared.TargetTypeNone &&
		selectedAbility.Ability.Targeting.TargetType != "" {
		// Show target selection UI
		return h.showAbilityTargetSelection(s, i, encounterID, abilityKey)
	}

	// For self-targeted or no-target abilities, proceed directly
	targetID := ""
	if selectedAbility.Ability.Targeting != nil && selectedAbility.Ability.Targeting.TargetType == shared.TargetTypeSelf {
		targetID = playerCombatant.CharacterID
	}

	// Special handling for abilities that need value selection
	if abilityKey == shared.AbilityKeyLayOnHands {
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

// showAbilityTargetSelection shows target selection UI for abilities
func (h *Handler) showAbilityTargetSelection(s *discordgo.Session, i *discordgo.InteractionCreate, encounterID, abilityKey string) error {
	// Get encounter to build target list
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

	// Get ability details to determine valid targets
	abilities, err := h.abilityService.GetAvailableAbilities(context.Background(), playerCombatant.CharacterID)
	if err != nil {
		return respondError(s, i, "Failed to get abilities", err)
	}

	var selectedAbility *ability.AvailableAbility
	for _, ab := range abilities {
		if ab.Ability.Key == abilityKey {
			selectedAbility = ab
			break
		}
	}

	if selectedAbility == nil || selectedAbility.Ability.Targeting == nil {
		return respondError(s, i, "Ability not found or has no targeting info", nil)
	}

	// Build target buttons based on targeting rules
	var buttons []discordgo.MessageComponent
	targeting := selectedAbility.Ability.Targeting

	for _, target := range enc.Combatants {
		// Skip dead targets
		if !target.IsActive || target.CurrentHP <= 0 {
			continue
		}

		// Apply targeting rules
		isValidTarget := false
		switch targeting.TargetType {
		case shared.TargetTypeSingleEnemy:
			// Can only target enemies (different type from caster)
			isValidTarget = target.Type != playerCombatant.Type
		case shared.TargetTypeSingleAlly:
			// Can only target allies (same type as caster)
			isValidTarget = target.Type == playerCombatant.Type
		case shared.TargetTypeSingleAny:
			// Can target any creature
			isValidTarget = true
		default:
			// Other target types not supported for selection
			continue
		}

		if !isValidTarget {
			continue
		}

		// Create button for this target
		emoji := "👤"
		if target.Type == combat.CombatantTypeMonster {
			emoji = "👹"
		}

		label := fmt.Sprintf("%s (HP: %d/%d)", target.Name, target.CurrentHP, target.MaxHP)
		if len(label) > 80 {
			label = fmt.Sprintf("%s (%d/%d)", target.Name, target.CurrentHP, target.MaxHP)
		}

		buttons = append(buttons, discordgo.Button{
			Label:    label,
			Style:    discordgo.PrimaryButton,
			CustomID: fmt.Sprintf("combat:abt:%s:%s:%s", encounterID, abilityKey, target.ID),
			Emoji:    &discordgo.ComponentEmoji{Name: emoji},
		})
	}

	// Build embed
	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("🎯 Select Target for %s", selectedAbility.Ability.Name),
		Color:       0x3498db, // Blue
		Description: selectedAbility.Ability.Description,
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Choose your target",
		},
	}

	// Add range info if available
	if targeting.Range > 0 {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "Range",
			Value:  fmt.Sprintf("%d feet", targeting.Range),
			Inline: true,
		})
	}

	// Build components
	components := []discordgo.MessageComponent{}

	// Add buttons in rows of 5
	for i := 0; i < len(buttons); i += 5 {
		end := i + 5
		if end > len(buttons) {
			end = len(buttons)
		}
		components = append(components, discordgo.ActionsRow{
			Components: buttons[i:end],
		})
	}

	// Add cancel button
	components = append(components, discordgo.ActionsRow{
		Components: []discordgo.MessageComponent{
			discordgo.Button{
				Label:    "Cancel",
				Style:    discordgo.SecondaryButton,
				CustomID: fmt.Sprintf("combat:abilities:%s", encounterID),
				Emoji:    &discordgo.ComponentEmoji{Name: "❌"},
			},
		},
	})

	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Embeds:     []*discordgo.MessageEmbed{embed},
			Components: components,
		},
	})
}

// handleUseAbilityWithTarget executes an ability with a selected target
func (h *Handler) handleUseAbilityWithTarget(s *discordgo.Session, i *discordgo.InteractionCreate, encounterID string) error {
	// Parse custom ID: combat:abt:encounterID:abilityKey:targetID
	parts := parseCustomID(i.MessageComponentData().CustomID)
	if len(parts) < 5 {
		return respondError(s, i, "Invalid target selection", nil)
	}
	abilityKey := parts[3]
	targetID := parts[4]

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

	// Find target combatant to get character ID
	var targetCombatant *combat.Combatant
	for _, c := range enc.Combatants {
		if c.ID == targetID {
			targetCombatant = c
			break
		}
	}

	if targetCombatant == nil {
		return respondError(s, i, "Target not found", nil)
	}

	// Use character ID for the ability
	actualTargetID := targetCombatant.CharacterID
	if actualTargetID == "" && targetCombatant.Type == combat.CombatantTypeMonster {
		// For monsters, use combatant ID
		actualTargetID = targetCombatant.ID
	}

	// Use the ability
	result, err := h.abilityService.UseAbility(context.Background(), &ability.UseAbilityInput{
		CharacterID: playerCombatant.CharacterID,
		AbilityKey:  abilityKey,
		TargetID:    actualTargetID,
		EncounterID: encounterID,
		Value:       0,
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

		// Add target info
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "Target",
			Value:  targetCombatant.Name,
			Inline: true,
		})

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

		// Add uses remaining if not unlimited
		if result.UsesRemaining >= 0 {
			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
				Name:   "Uses Remaining",
				Value:  fmt.Sprintf("%d", result.UsesRemaining),
				Inline: true,
			})
		}

		// Add to combat log
		if err := h.encounterService.LogCombatAction(context.Background(), encounterID,
			fmt.Sprintf("**%s** used **%s** on **%s**: %s", playerCombatant.Name, getAbilityName(abilityKey), targetCombatant.Name, result.Message)); err != nil {
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

func formatActionType(actionType shared.AbilityType) string {
	switch actionType {
	case shared.AbilityTypeAction:
		return "Action"
	case shared.AbilityTypeBonusAction:
		return "Bonus Action"
	case shared.AbilityTypeReaction:
		return "Reaction"
	case shared.AbilityTypeFree:
		return "Free Action"
	default:
		return string(actionType)
	}
}

func getAbilityEmoji(abilityKey string) string {
	emojiMap := map[string]string{
		shared.AbilityKeyRage:              "😡",
		shared.AbilityKeySecondWind:        "💨",
		shared.AbilityKeyBardicInspiration: "🎵",
		shared.AbilityKeyLayOnHands:        "🙏",
		shared.AbilityKeyDivineSense:       "👁️",
		shared.AbilityKeyViciousMockery:    "🎭",
	}

	if emoji, exists := emojiMap[abilityKey]; exists {
		return emoji
	}
	return "✨"
}

func getAbilityName(abilityKey string) string {
	nameMap := map[string]string{
		shared.AbilityKeyRage:              "Rage",
		shared.AbilityKeySecondWind:        "Second Wind",
		shared.AbilityKeyBardicInspiration: "Bardic Inspiration",
		shared.AbilityKeyLayOnHands:        "Lay on Hands",
		shared.AbilityKeyDivineSense:       "Divine Sense",
		shared.AbilityKeyViciousMockery:    "Vicious Mockery",
	}

	if name, exists := nameMap[abilityKey]; exists {
		return name
	}
	return abilityKey
}

// showLayOnHandsAmountSelection shows UI for selecting heal amount
func (h *Handler) showLayOnHandsAmountSelection(s *discordgo.Session, i *discordgo.InteractionCreate, encounterID string, playerCombatant *combat.Combatant) error {
	// Get character to check available healing pool
	char, err := h.abilityService.GetAvailableAbilities(context.Background(), playerCombatant.CharacterID)
	if err != nil {
		return respondError(s, i, "Failed to get character abilities", err)
	}

	// Find lay on hands ability to check remaining pool
	var layOnHands *shared.ActiveAbility
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
