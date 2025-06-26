package combat

import (
	"testing"

	"github.com/KirkDiggler/dnd-bot-discord/internal/services/encounter"
	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildCombatComponents_AlwaysHasCustomID(t *testing.T) {
	tests := []struct {
		name   string
		result *encounter.ExecuteAttackResult
	}{
		{
			name: "Combat ended with victory",
			result: &encounter.ExecuteAttackResult{
				CombatEnded: true,
				PlayersWon:  true,
			},
		},
		{
			name: "Combat ended with defeat",
			result: &encounter.ExecuteAttackResult{
				CombatEnded: true,
				PlayersWon:  false,
			},
		},
		{
			name: "Combat continues",
			result: &encounter.ExecuteAttackResult{
				CombatEnded: false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			components := buildCombatComponents("test-encounter-id", tt.result)
			require.NotEmpty(t, components, "Should have at least one component row")

			// Check all buttons have custom IDs
			for _, row := range components {
				if actionRow, ok := row.(discordgo.ActionsRow); ok {
					for _, component := range actionRow.Components {
						if button, ok := component.(discordgo.Button); ok {
							assert.NotEmpty(t, button.CustomID,
								"Button '%s' must have a CustomID even if disabled", button.Label)
							assert.Contains(t, button.CustomID, "combat:",
								"CustomID should start with 'combat:' prefix")
						}
					}
				}
			}
		})
	}
}
