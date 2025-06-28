package combat

import (
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/assert"
)

func TestIsEphemeralInteraction(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() *discordgo.InteractionCreate
		expected bool
	}{
		{
			name: "ephemeral message",
			setup: func() *discordgo.InteractionCreate {
				return &discordgo.InteractionCreate{
					Interaction: &discordgo.Interaction{
						Message: &discordgo.Message{
							Flags: discordgo.MessageFlagsEphemeral,
						},
					},
				}
			},
			expected: true,
		},
		{
			name: "non-ephemeral message",
			setup: func() *discordgo.InteractionCreate {
				return &discordgo.InteractionCreate{
					Interaction: &discordgo.Interaction{
						Message: &discordgo.Message{
							Flags: 0,
						},
					},
				}
			},
			expected: false,
		},
		{
			name: "no message",
			setup: func() *discordgo.InteractionCreate {
				return &discordgo.InteractionCreate{}
			},
			expected: false,
		},
		{
			name: "mixed flags with ephemeral",
			setup: func() *discordgo.InteractionCreate {
				return &discordgo.InteractionCreate{
					Interaction: &discordgo.Interaction{
						Message: &discordgo.Message{
							Flags: discordgo.MessageFlagsEphemeral,
						},
					},
				}
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := tt.setup()
			result := isEphemeralInteraction(i)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseCustomID(t *testing.T) {
	tests := []struct {
		name     string
		customID string
		expected []string
	}{
		{
			name:     "combat select target",
			customID: "combat:select_target:encounter123:target456",
			expected: []string{"combat", "select_target", "encounter123", "target456"},
		},
		{
			name:     "combat attack",
			customID: "combat:attack:encounter123",
			expected: []string{"combat", "attack", "encounter123"},
		},
		{
			name:     "empty string",
			customID: "",
			expected: []string{""},
		},
		{
			name:     "single part",
			customID: "combat",
			expected: []string{"combat"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseCustomID(tt.customID)
			assert.Equal(t, tt.expected, result)
		})
	}
}
