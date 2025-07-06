//go:build ignore
// +build ignore

package spells

import (
	"context"
	"testing"

	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/events"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockDiceRoller for testing
type mockDiceRoller struct {
	result int
}

func (m *mockDiceRoller) Roll(numDice, sides, modifier int) (struct{ Total int }, error) {
	// For 1d4+1, return consistent value
	return struct{ Total int }{Total: m.result}, nil
}

// mockCharacterService for testing
type mockCharacterService struct {
	characters map[string]*character.Character
}

func (m *mockCharacterService) UpdateEquipment(char *character.Character) error {
	return nil
}

func (m *mockCharacterService) GetByID(id string) (*character.Character, error) {
	if char, ok := m.characters[id]; ok {
		return char, nil
	}
	return nil, nil
}

func TestMagicMissileHandler_Execute(t *testing.T) {
	tests := []struct {
		name           string
		caster         *character.Character
		input          *SpellInput
		diceResult     int
		expectSuccess  bool
		expectDamage   int
		expectMissiles int
	}{
		{
			name: "level 1 magic missile single target",
			caster: &character.Character{
				ID:    "wizard-123",
				Name:  "Gandalf",
				Level: 5,
			},
			input: &SpellInput{
				SpellLevel: 1,
				TargetIDs:  []string{"goblin-1"},
			},
			diceResult:     3, // 1d4+1 = 3
			expectSuccess:  true,
			expectDamage:   9, // 3 missiles * 3 damage
			expectMissiles: 3,
		},
		{
			name: "level 3 magic missile two targets",
			caster: &character.Character{
				ID:    "wizard-123",
				Name:  "Gandalf",
				Level: 5,
			},
			input: &SpellInput{
				SpellLevel: 3,
				TargetIDs:  []string{"goblin-1", "goblin-2"},
			},
			diceResult:     4, // 1d4+1 = 4
			expectSuccess:  true,
			expectDamage:   20, // 5 missiles * 4 damage (3 + 2 split)
			expectMissiles: 5,  // 3 + 2 extra for level 3
		},
		{
			name: "invalid spell level",
			caster: &character.Character{
				ID:   "wizard-123",
				Name: "Gandalf",
			},
			input: &SpellInput{
				SpellLevel: 10,
				TargetIDs:  []string{"goblin-1"},
			},
			expectSuccess: false,
		},
		{
			name: "no targets",
			caster: &character.Character{
				ID:   "wizard-123",
				Name: "Gandalf",
			},
			input: &SpellInput{
				SpellLevel: 1,
				TargetIDs:  []string{},
			},
			expectSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create event bus
			eventBus := events.NewEventBus()

			// Track emitted events
			var spellCastEvent *events.GameEvent
			var spellDamageEvents []*events.GameEvent

			// Subscribe to events
			eventBus.Subscribe(events.OnSpellCast, &testEventListener{
				handleFunc: func(event *events.GameEvent) error {
					spellCastEvent = event
					return nil
				},
			})
			eventBus.Subscribe(events.OnSpellDamage, &testEventListener{
				handleFunc: func(event *events.GameEvent) error {
					spellDamageEvents = append(spellDamageEvents, event)
					return nil
				},
			})

			// Create handler
			handler := NewMagicMissileHandler(eventBus)
			handler.SetDiceRoller(&mockDiceRoller{result: tt.diceResult})
			handler.SetCharacterService(&mockCharacterService{
				characters: map[string]*character.Character{
					"goblin-1": {ID: "goblin-1", Name: "Goblin 1"},
					"goblin-2": {ID: "goblin-2", Name: "Goblin 2"},
				},
			})

			// Execute
			result, err := handler.Execute(context.Background(), tt.caster, tt.input)

			// Verify
			require.NoError(t, err)
			assert.Equal(t, tt.expectSuccess, result.Success)

			if tt.expectSuccess {
				assert.Equal(t, tt.expectDamage, result.TotalDamage)
				assert.Equal(t, tt.input.SpellLevel, result.SpellSlotUsed)

				// Check spell cast event
				assert.NotNil(t, spellCastEvent)
				level, _ := spellCastEvent.GetIntContext(events.ContextSpellLevel)
				assert.Equal(t, tt.input.SpellLevel, level)

				// Check damage events
				assert.Len(t, spellDamageEvents, len(tt.input.TargetIDs))
			}
		})
	}
}

// testEventListener is a simple event listener for testing
type testEventListener struct {
	handleFunc func(event *events.GameEvent) error
}

func (t *testEventListener) HandleEvent(event *events.GameEvent) error {
	return t.handleFunc(event)
}

func (t *testEventListener) Priority() int {
	return 0
}
