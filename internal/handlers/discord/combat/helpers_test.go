package combat

import (
	"testing"

	"github.com/bwmarrin/discordgo"
)

// TestUpdateSharedCombatMessage tests the updateSharedCombatMessage helper
func TestUpdateSharedCombatMessage(t *testing.T) {
	// Create a mock session that tracks calls
	callCount := 0
	var capturedEdit *discordgo.MessageEdit

	mockSession := &mockSession{
		channelMessageEditComplex: func(edit *discordgo.MessageEdit) (*discordgo.Message, error) {
			callCount++
			capturedEdit = edit
			return &discordgo.Message{}, nil
		},
	}

	embed := &discordgo.MessageEmbed{Title: "Test Combat"}
	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{Label: "Test"},
			},
		},
	}

	// Test with valid MessageID and ChannelID
	t.Run("ValidIDs", func(t *testing.T) {
		callCount = 0
		err := updateSharedCombatMessage(mockSession, "enc-123", "msg-456", "channel-789", embed, components)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if callCount != 1 {
			t.Errorf("Expected 1 call, got %d", callCount)
		}
		if capturedEdit.ID != "msg-456" {
			t.Errorf("Expected message ID msg-456, got %s", capturedEdit.ID)
		}
		if capturedEdit.Channel != "channel-789" {
			t.Errorf("Expected channel ID channel-789, got %s", capturedEdit.Channel)
		}
	})

	// Test with empty MessageID
	t.Run("EmptyMessageID", func(t *testing.T) {
		callCount = 0
		err := updateSharedCombatMessage(mockSession, "enc-123", "", "channel-789", embed, components)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if callCount != 0 {
			t.Errorf("Expected no calls with empty MessageID, got %d", callCount)
		}
	})

	// Test with empty ChannelID
	t.Run("EmptyChannelID", func(t *testing.T) {
		callCount = 0
		err := updateSharedCombatMessage(mockSession, "enc-123", "msg-456", "", embed, components)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if callCount != 0 {
			t.Errorf("Expected no calls with empty ChannelID, got %d", callCount)
		}
	})
}

// mockSession is a minimal mock for testing
type mockSession struct {
	channelMessageEditComplex func(*discordgo.MessageEdit) (*discordgo.Message, error)
}

func (m *mockSession) ChannelMessageEditComplex(edit *discordgo.MessageEdit, options ...discordgo.RequestOption) (*discordgo.Message, error) {
	if m.channelMessageEditComplex != nil {
		return m.channelMessageEditComplex(edit)
	}
	return nil, nil
}
