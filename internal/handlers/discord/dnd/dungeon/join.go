package dungeon

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/KirkDiggler/dnd-bot-discord/internal/entities"
	"github.com/KirkDiggler/dnd-bot-discord/internal/services"
	"github.com/bwmarrin/discordgo"
)

// Helper function to edit deferred responses
func editDeferredResponse(s *discordgo.Session, i *discordgo.InteractionCreate, content string, embeds ...*discordgo.MessageEmbed) error {
	edit := &discordgo.WebhookEdit{
		Content: &content,
	}
	if len(embeds) > 0 {
		edit.Embeds = &embeds
		emptyContent := ""
		edit.Content = &emptyContent
	}
	_, err := s.InteractionResponseEdit(i.Interaction, edit)
	return err
}

type JoinPartyHandler struct {
	services *services.Provider
}

func NewJoinPartyHandler(serviceProvider *services.Provider) *JoinPartyHandler {
	return &JoinPartyHandler{
		services: serviceProvider,
	}
}

func (h *JoinPartyHandler) HandleButton(s *discordgo.Session, i *discordgo.InteractionCreate, sessionID string) error {
	log.Printf("JoinParty - User %s attempting to join session %s", i.Member.User.ID, sessionID)

	// Defer the response immediately to avoid timeout
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags: discordgo.MessageFlagsEphemeral,
		},
	})
	if err != nil {
		log.Printf("Failed to defer interaction: %v", err)
		return err
	}

	// Get user's active character
	chars, err := h.services.CharacterService.ListByOwner(i.Member.User.ID)
	if err != nil {
		log.Printf("JoinParty - Error listing characters for user %s: %v", i.Member.User.ID, err)
		return editDeferredResponse(s, i, "❌ Failed to list characters. Please try again.")
	}

	if len(chars) == 0 {
		log.Printf("JoinParty - User %s has no characters", i.Member.User.ID)
		return editDeferredResponse(s, i, "❌ You need a character to join! Use `/dnd character create`")
	}

	log.Printf("JoinParty - User %s has %d characters", i.Member.User.ID, len(chars))

	// Find all active characters
	var activeChars []*entities.Character
	for _, char := range chars {
		log.Printf("JoinParty - Character: ID=%s, Name=%s, Status=%s", char.ID, char.Name, char.Status)
		if char.Status == entities.CharacterStatusActive {
			activeChars = append(activeChars, char)
		}
	}

	log.Printf("JoinParty - Found %d active characters", len(activeChars))

	if len(activeChars) == 0 {
		log.Printf("JoinParty - No active character found for user %s", i.Member.User.ID)
		return editDeferredResponse(s, i, "❌ No active character found! Activate a character first.")
	}

	// If user has multiple active characters, show selection menu
	if len(activeChars) > 1 {
		components := buildCharacterSelectMenu(activeChars, sessionID)
		edit := &discordgo.WebhookEdit{
			Content:    &[]string{"🎭 Select your character for this dungeon:"}[0],
			Components: &components,
		}
		_, err := s.InteractionResponseEdit(i.Interaction, edit)
		return err
	}

	// If only one active character, use it
	playerChar := activeChars[0]

	// Check if character is complete
	if !playerChar.IsComplete() {
		missingInfo := []string{}
		if playerChar.Name == "" {
			missingInfo = append(missingInfo, "name")
		}
		if playerChar.Race == nil {
			missingInfo = append(missingInfo, "race")
		}
		if playerChar.Class == nil {
			missingInfo = append(missingInfo, "class")
		}
		if len(playerChar.Attributes) == 0 {
			missingInfo = append(missingInfo, "ability scores")
		}

		log.Printf("Character %s (ID: %s) is incomplete. Missing: %v",
			playerChar.Name, playerChar.ID, missingInfo)

		return editDeferredResponse(s, i, fmt.Sprintf("❌ Your character is incomplete! Missing: %s\n\nPlease create a new character or contact an admin if this is an error.",
			strings.Join(missingInfo, ", ")))
	}

	// Check if user is already in the session
	sess, err := h.services.SessionService.GetSession(context.Background(), sessionID)
	if err != nil {
		return editDeferredResponse(s, i, fmt.Sprintf("❌ Failed to get session: %v", err))
	}

	// If not in session, join it
	if !sess.IsUserInSession(i.Member.User.ID) {
		log.Printf("User %s not in session, joining...", i.Member.User.ID)
		_, err = h.services.SessionService.JoinSession(context.Background(), sessionID, i.Member.User.ID)
		if err != nil {
			return editDeferredResponse(s, i, fmt.Sprintf("❌ Failed to join party: %v", err))
		}
	} else {
		log.Printf("User %s already in session, updating character selection...", i.Member.User.ID)
	}

	// Select character
	log.Printf("Selecting character %s (ID: %s) for user %s in session %s", playerChar.Name, playerChar.ID, i.Member.User.ID, sessionID)
	err = h.services.SessionService.SelectCharacter(context.Background(), sessionID, i.Member.User.ID, playerChar.ID)
	if err != nil {
		log.Printf("Failed to select character: %v", err)
		return editDeferredResponse(s, i, fmt.Sprintf("❌ Failed to select character: %v", err))
	}
	log.Printf("Successfully selected character %s for user %s", playerChar.Name, i.Member.User.ID)

	// Verify the character was set
	sess, getErr := h.services.SessionService.GetSession(context.Background(), sessionID)
	if getErr == nil && sess != nil {
		if member, exists := sess.Members[i.Member.User.ID]; exists {
			log.Printf("JoinParty - Verified session member - UserID: %s, Role: %s, CharacterID: %s",
				member.UserID, member.Role, member.CharacterID)
		} else {
			log.Printf("JoinParty - WARNING: User %s not found in session members after join", i.Member.User.ID)
		}

		// Log all session members for debugging
		log.Printf("JoinParty - Current session members:")
		for uid, m := range sess.Members {
			log.Printf("  - UserID: %s, Role: %s, CharacterID: %s", uid, m.Role, m.CharacterID)
		}
	} else {
		log.Printf("JoinParty - Could not verify session state: %v", getErr)
	}

	// Build character info
	charInfo := fmt.Sprintf("%s (Level %d)", playerChar.GetDisplayInfo(), playerChar.Level)

	// Success response
	embed := &discordgo.MessageEmbed{
		Title:       "🎉 Joined the Party!",
		Description: fmt.Sprintf("**%s** has joined the dungeon delve!", playerChar.Name),
		Color:       0x2ecc71, // Green
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Character",
				Value:  charInfo,
				Inline: true,
			},
			{
				Name:   "HP",
				Value:  fmt.Sprintf("%d/%d", playerChar.CurrentHitPoints, playerChar.MaxHitPoints),
				Inline: true,
			},
			{
				Name:   "AC",
				Value:  fmt.Sprintf("%d", playerChar.AC),
				Inline: true,
			},
		},
	}

	return editDeferredResponse(s, i, "", embed)
}

// buildCharacterSelectMenu creates a dropdown menu for character selection
func buildCharacterSelectMenu(characters []*entities.Character, sessionID string) []discordgo.MessageComponent {
	options := make([]discordgo.SelectMenuOption, 0, len(characters))
	for _, char := range characters {
		options = append(options, discordgo.SelectMenuOption{
			Label:       fmt.Sprintf("%s - %s", char.Name, char.GetDisplayInfo()),
			Description: fmt.Sprintf("Level %d | HP: %d/%d | AC: %d", char.Level, char.CurrentHitPoints, char.MaxHitPoints, char.AC),
			Value:       char.ID,
		})
	}

	return []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.SelectMenu{
					CustomID:    fmt.Sprintf("dungeon:select_character:%s", sessionID),
					Placeholder: "Choose your character...",
					Options:     options,
				},
			},
		},
	}
}
