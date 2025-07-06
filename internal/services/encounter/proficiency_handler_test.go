package encounter_test

// TODO: Migrate this test to use rpg-toolkit event system
// Temporarily commented out during migration from internal/domain/events to rpg-toolkit

/*
import (
	"fmt"
	"testing"

	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/equipment"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/events"
	rulebook "github.com/KirkDiggler/dnd-bot-discord/internal/domain/rulebook/dnd5e"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/shared"
	"github.com/KirkDiggler/dnd-bot-discord/internal/services/encounter"
	mockencounter "github.com/KirkDiggler/dnd-bot-discord/internal/services/encounter/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestProficiencyHandler_HandleEvent(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mock service
	mockService := mockencounter.NewMockService(ctrl)

	// Create handler
	handler := encounter.NewProficiencyHandler(mockService)

	t.Run("logs proficiency status for weapon attack", func(t *testing.T) {
		// Create a level 5 fighter with longsword proficiency
		fighter := &character.Character{
			ID:    "fighter-1",
			Name:  "Grendel",
			Level: 5, // Proficiency bonus = +3
			Class: &rulebook.Class{
				Name: "Fighter",
			},
			Proficiencies: make(map[rulebook.ProficiencyType][]*rulebook.Proficiency),
			EquippedSlots: make(map[shared.Slot]equipment.Equipment),
		}

		// Add martial weapon proficiency
		fighter.Proficiencies[rulebook.ProficiencyTypeWeapon] = []*rulebook.Proficiency{
			{Key: "martial-weapons", Name: "Martial Weapons"},
		}

		// Equip a longsword
		longsword := &equipment.Weapon{
			Base: equipment.BasicEquipment{
				Key:  "longsword",
				Name: "Longsword",
			},
			WeaponRange:    "Melee",
			WeaponCategory: "Martial",
		}
		fighter.EquippedSlots[shared.SlotMainHand] = longsword

		// Create attack roll event (values already include proficiency from character.Attack())
		event := events.NewGameEvent(events.OnAttackRoll).
			WithActor(fighter).
			WithContext("weapon", "Longsword").
			WithContext("attack_bonus", 7).  // STR(4) + Prof(3) already included
			WithContext("total_attack", 17) // d20(10) + STR(4) + Prof(3)

		// Handle the event
		err := handler.HandleEvent(event)
		require.NoError(t, err)

		// Values should NOT be modified (proficiency already included)
		attackBonus, ok := event.GetIntContext("attack_bonus")
		assert.True(t, ok)
		assert.Equal(t, 7, attackBonus) // Should remain unchanged

		totalAttack, ok := event.GetIntContext("total_attack")
		assert.True(t, ok)
		assert.Equal(t, 17, totalAttack) // Should remain unchanged
	})

	t.Run("logs non-proficient weapon status", func(t *testing.T) {
		// Create a wizard without martial weapon proficiency
		wizard := &character.Character{
			ID:    "wizard-1",
			Name:  "Merlin",
			Level: 5,
			Class: &rulebook.Class{
				Name: "Wizard",
			},
			Proficiencies: make(map[rulebook.ProficiencyType][]*rulebook.Proficiency),
			EquippedSlots: make(map[shared.Slot]equipment.Equipment),
		}

		// Add simple weapon proficiency
		wizard.Proficiencies[rulebook.ProficiencyTypeWeapon] = []*rulebook.Proficiency{
			{Key: "simple-weapons", Name: "Simple Weapons"},
		}

		// Equip a longsword (martial weapon)
		longsword := &equipment.Weapon{
			Base: equipment.BasicEquipment{
				Key:  "longsword",
				Name: "Longsword",
			},
			WeaponRange:    "Melee",
			WeaponCategory: "Martial",
		}
		wizard.EquippedSlots[shared.SlotMainHand] = longsword

		// Create attack roll event (no proficiency included)
		event := events.NewGameEvent(events.OnAttackRoll).
			WithActor(wizard).
			WithContext("weapon", "Longsword").
			WithContext("attack_bonus", 0). // STR modifier only
			WithContext("total_attack", 10) // d20(10) + STR(0)

		// Handle the event
		err := handler.HandleEvent(event)
		require.NoError(t, err)

		// Values should NOT be modified
		attackBonus, ok := event.GetIntContext("attack_bonus")
		assert.True(t, ok)
		assert.Equal(t, 0, attackBonus) // Should remain unchanged

		totalAttack, ok := event.GetIntContext("total_attack")
		assert.True(t, ok)
		assert.Equal(t, 10, totalAttack) // Should remain unchanged
	})

	t.Run("skips unarmed strikes", func(t *testing.T) {
		// Create a character
		char := &character.Character{
			ID:    "char-1",
			Name:  "Bob",
			Level: 5,
		}

		// Create attack roll event for unarmed strike
		event := events.NewGameEvent(events.OnAttackRoll).
			WithActor(char).
			WithContext("weapon", "Unarmed Strike").
			WithContext("attack_bonus", 2).
			WithContext("total_attack", 12)

		// Handle the event
		err := handler.HandleEvent(event)
		require.NoError(t, err)

		// Check that nothing changed
		attackBonus, ok := event.GetIntContext("attack_bonus")
		assert.True(t, ok)
		assert.Equal(t, 2, attackBonus)
	})

	t.Run("future saving throw implementation", func(t *testing.T) {
		// Create a level 9 fighter with STR and CON save proficiency
		fighter := &character.Character{
			ID:    "fighter-1",
			Name:  "Tank",
			Level: 9, // Proficiency bonus = +4
			Class: &rulebook.Class{
				Name: "Fighter",
			},
			Proficiencies: make(map[rulebook.ProficiencyType][]*rulebook.Proficiency),
		}

		// Add saving throw proficiencies
		fighter.Proficiencies[rulebook.ProficiencyTypeSavingThrow] = []*rulebook.Proficiency{
			{Key: "saving-throw-str", Name: "Strength Saving Throws"},
			{Key: "saving-throw-con", Name: "Constitution Saving Throws"},
		}

		// Create saving throw event
		event := events.NewGameEvent(events.OnSavingThrow).
			WithTarget(fighter).
			WithContext("save_type", "str").
			WithContext("save_bonus", 3). // STR modifier
			WithContext("total_save", 15) // d20(12) + STR(3)

		// Handle the event
		err := handler.HandleEvent(event)
		require.NoError(t, err)

		// In future implementation, this would add proficiency
		// For now, it just logs and prepares the architecture
	})
}

func TestGetProficiencyBonus(t *testing.T) {
	tests := []struct {
		level    int
		expected int
	}{
		{1, 2},
		{4, 2},
		{5, 3},
		{8, 3},
		{9, 4},
		{12, 4},
		{13, 5},
		{16, 5},
		{17, 6},
		{20, 6},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("level %d", tt.level), func(t *testing.T) {
			// Verify the proficiency bonus formula
			assert.Equal(t, tt.expected, 2+((tt.level-1)/4))
		})
	}
}
*/
