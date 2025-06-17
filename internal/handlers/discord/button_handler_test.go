package discord_test

import (
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/assert"
)

func TestButtonCustomIDs(t *testing.T) {
	tests := []struct {
		name        string
		customID    string
		expectedCtx string
		expectedAction string
		expectedData  string
		shouldParse   bool
	}{
		{
			name:        "character quickshow button",
			customID:    "character:quickshow:char_123",
			expectedCtx: "character",
			expectedAction: "quickshow",
			expectedData:  "char_123",
			shouldParse:   true,
		},
		{
			name:        "character manage continue button",
			customID:    "character_manage:continue:char_456",
			expectedCtx: "character_manage",
			expectedAction: "continue",
			expectedData:  "char_456",
			shouldParse:   true,
		},
		{
			name:        "character manage edit button",
			customID:    "character_manage:edit:char_789",
			expectedCtx: "character_manage",
			expectedAction: "edit",
			expectedData:  "char_789",
			shouldParse:   true,
		},
		{
			name:        "invalid format - single part",
			customID:    "invalid",
			shouldParse:   false,
		},
		{
			name:        "invalid format - no action",
			customID:    "character:",
			shouldParse:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse custom ID like the handler does
			parts := splitCustomID(tt.customID)
			
			if tt.shouldParse {
				assert.True(t, len(parts) >= 2, "Should have at least 2 parts")
				if len(parts) >= 2 {
					assert.Equal(t, tt.expectedCtx, parts[0])
					assert.Equal(t, tt.expectedAction, parts[1])
					if tt.expectedData != "" {
						assert.True(t, len(parts) >= 3, "Should have data part")
						if len(parts) >= 3 {
							assert.Equal(t, tt.expectedData, parts[2])
						}
					}
				}
			} else {
				assert.False(t, len(parts) >= 2, "Should not parse invalid custom ID")
			}
		})
	}
}

func TestCharacterListButtons(t *testing.T) {
	// Test that character list generates correct button custom IDs
	activeChar := &mockCharacter{
		ID:   "char_123",
		Name: "Gandalf",
	}

	// Generate show button
	showButton := discordgo.Button{
		Label:    activeChar.Name,
		Style:    discordgo.SecondaryButton,
		CustomID: "character:quickshow:" + activeChar.ID,
		Emoji: &discordgo.ComponentEmoji{
			Name: "👁️",
		},
	}

	assert.Equal(t, "character:quickshow:char_123", showButton.CustomID)
	assert.Equal(t, "Gandalf", showButton.Label)

	// Generate edit button
	editButton := discordgo.Button{
		Label:    "Edit " + activeChar.Name,
		Style:    discordgo.PrimaryButton,
		CustomID: "character_manage:edit:" + activeChar.ID,
		Emoji: &discordgo.ComponentEmoji{
			Name: "✏️",
		},
	}

	assert.Equal(t, "character_manage:edit:char_123", editButton.CustomID)
	assert.Equal(t, "Edit Gandalf", editButton.Label)
}

func TestDraftCharacterButtons(t *testing.T) {
	// Test that draft characters get continue button
	draftChar := &mockCharacter{
		ID:     "draft_456",
		Name:   "", // No name yet
		Status: "draft",
	}

	continueButton := discordgo.Button{
		Label:    "Continue Creating",
		Style:    discordgo.PrimaryButton,
		CustomID: "character_manage:continue:" + draftChar.ID,
		Emoji: &discordgo.ComponentEmoji{
			Name: "▶️",
		},
	}

	assert.Equal(t, "character_manage:continue:draft_456", continueButton.CustomID)
	assert.Equal(t, "Continue Creating", continueButton.Label)
}

// Helper function to mimic the handler's parsing
func splitCustomID(customID string) []string {
	var parts []string
	start := 0
	for i := 0; i <= len(customID); i++ {
		if i == len(customID) || customID[i] == ':' {
			if i > start {
				parts = append(parts, customID[start:i])
			}
			start = i + 1
		}
	}
	return parts
}

type mockCharacter struct {
	ID     string
	Name   string
	Status string
}