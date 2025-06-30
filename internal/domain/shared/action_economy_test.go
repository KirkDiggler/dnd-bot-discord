package shared

import (
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/equipment"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/rulebook"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestActionEconomy_Reset(t *testing.T) {
	ae := &ActionEconomy{
		ActionUsed:      true,
		BonusActionUsed: true,
		ReactionUsed:    true,
		MovementUsed:    30,
		ActionsThisTurn: []ActionRecord{
			{Type: "attack", Subtype: "weapon"},
		},
		AvailableBonusActions: []BonusActionOption{
			{Key: "test"},
		},
	}

	ae.Reset()

	assert.False(t, ae.ActionUsed, "Action should be reset")
	assert.False(t, ae.BonusActionUsed, "Bonus action should be reset")
	assert.True(t, ae.ReactionUsed, "Reaction should NOT be reset")
	assert.Equal(t, 0, ae.MovementUsed, "Movement should be reset")
	assert.Empty(t, ae.ActionsThisTurn, "Actions history should be cleared")
	assert.Empty(t, ae.AvailableBonusActions, "Available bonus actions should be cleared")
}

func TestActionEconomy_RecordAction(t *testing.T) {
	tests := []struct {
		name       string
		actionType string
		expectUsed bool
	}{
		{"attack uses action", "attack", true},
		{"spell uses action", "spell", true},
		{"dash uses action", "dash", true},
		{"dodge uses action", "dodge", true},
		{"help uses action", "help", true},
		{"ready uses action", "ready", true},
		{"bonus_action doesn't use main action", "bonus_action", false},
		{"free action doesn't use main action", "free", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ae := &ActionEconomy{}
			ae.RecordAction(tt.actionType, "test", "test-weapon")

			assert.Equal(t, tt.expectUsed, ae.ActionUsed)
			assert.Len(t, ae.ActionsThisTurn, 1)
			assert.Equal(t, tt.actionType, ae.ActionsThisTurn[0].Type)
		})
	}
}

func TestActionEconomy_HasTakenAction(t *testing.T) {
	ae := &ActionEconomy{}

	assert.False(t, ae.HasTakenAction("attack"))

	ae.RecordAction("attack", "weapon", "longsword")
	assert.True(t, ae.HasTakenAction("attack"))
	assert.False(t, ae.HasTakenAction("spell"))

	ae.RecordAction("spell", "cantrip", "")
	assert.True(t, ae.HasTakenAction("spell"))
}

func TestActionEconomy_GetActionsByType(t *testing.T) {
	ae := &ActionEconomy{}

	// Record multiple actions
	ae.RecordAction("attack", "weapon", "longsword")
	time.Sleep(time.Millisecond) // Ensure different timestamps
	ae.RecordAction("bonus_action", "unarmed_strike", "")
	ae.RecordAction("attack", "weapon", "shortsword")

	attacks := ae.GetActionsByType("attack")
	assert.Len(t, attacks, 2)
	assert.Equal(t, "longsword", attacks[0].WeaponKey)
	assert.Equal(t, "shortsword", attacks[1].WeaponKey)

	bonusActions := ae.GetActionsByType("bonus_action")
	assert.Len(t, bonusActions, 1)
	assert.Equal(t, "unarmed_strike", bonusActions[0].Subtype)
}

func TestCharacter_StartNewTurn(t *testing.T) {
	char := &character.Character{
		Name:  "Test Character",
		Level: 1,
		Resources: &character.CharacterResources{
			ActionEconomy: ActionEconomy{
				ActionUsed:      true,
				BonusActionUsed: true,
				ReactionUsed:    true,
				MovementUsed:    25,
			},
			SneakAttackUsedThisTurn: true,
		},
	}

	char.StartNewTurn()

	assert.False(t, char.Resources.ActionEconomy.ActionUsed)
	assert.False(t, char.Resources.ActionEconomy.BonusActionUsed)
	assert.False(t, char.Resources.ActionEconomy.ReactionUsed, "Reaction SHOULD be reset at start of turn")
	assert.Equal(t, 0, char.Resources.ActionEconomy.MovementUsed)
	assert.False(t, char.Resources.SneakAttackUsedThisTurn)
}

func TestCharacter_MartialArtsBonusAction(t *testing.T) {
	// Create a monk with martial arts
	monk := &character.Character{
		Name:  "Test Monk",
		Level: 1,
		Features: []*rulebook.CharacterFeature{
			{Key: "martial-arts", Name: "Martial Arts"},
		},
		Resources: &character.CharacterResources{},
		EquippedSlots: map[Slot]equipment.Equipment{
			SlotMainHand: &equipment.Weapon{
				Base:           equipment.BasicEquipment{Key: equipment.WeaponKeyShortsword},
				WeaponCategory: "Martial",
				WeaponRange:    "Melee",
			},
		},
	}

	// Initially no bonus actions
	monk.StartNewTurn() // This will update available bonus actions
	assert.Empty(t, monk.Resources.ActionEconomy.AvailableBonusActions)

	// Take attack action with monk weapon
	monk.RecordAction("attack", "weapon", equipment.WeaponKeyShortsword)

	// Now martial arts bonus action should be available
	assert.Len(t, monk.Resources.ActionEconomy.AvailableBonusActions, 1)
	bonus := monk.Resources.ActionEconomy.AvailableBonusActions[0]
	assert.Equal(t, "martial_arts_strike", bonus.Key)
	assert.Equal(t, "martial_arts", bonus.Source)
	assert.Equal(t, "unarmed_strike", bonus.ActionType)

	// Use the bonus action
	assert.True(t, monk.UseBonusAction("martial_arts_strike"))
	assert.True(t, monk.Resources.ActionEconomy.BonusActionUsed)

	// Should no longer be available
	assert.Empty(t, monk.Resources.ActionEconomy.AvailableBonusActions)
	assert.False(t, monk.CanTakeBonusAction())
}

func TestCharacter_TwoWeaponFightingBonusAction(t *testing.T) {
	// Create a character with two light weapons
	char := &character.Character{
		Name:      "Dual Wielder",
		Level:     1,
		Resources: &character.CharacterResources{},
		EquippedSlots: map[Slot]equipment.Equipment{
			SlotMainHand: &equipment.Weapon{
				Base:           equipment.BasicEquipment{Key: "shortsword"},
				WeaponCategory: "Martial",
				WeaponRange:    "Melee",
				Properties: []*ReferenceItem{
					{Key: "light"},
					{Key: "finesse"},
				},
			},
			SlotOffHand: &equipment.Weapon{
				Base:           equipment.BasicEquipment{Key: "dagger"},
				WeaponCategory: "Simple",
				WeaponRange:    "Melee",
				Properties: []*ReferenceItem{
					{Key: "light"},
					{Key: "finesse"},
					{Key: "thrown"},
				},
			},
		},
	}

	// Initially no bonus actions
	char.updateAvailableBonusActionsInternal()
	assert.Empty(t, char.Resources.ActionEconomy.AvailableBonusActions)

	// Attack with main hand light weapon
	char.RecordAction("attack", "weapon", "shortsword")

	// Two-weapon fighting bonus should be available
	assert.Len(t, char.Resources.ActionEconomy.AvailableBonusActions, 1)
	bonus := char.Resources.ActionEconomy.AvailableBonusActions[0]
	assert.Equal(t, "two_weapon_attack", bonus.Key)
	assert.Equal(t, "two_weapon_fighting", bonus.Source)
	assert.Equal(t, "weapon_attack", bonus.ActionType)
}

func TestCharacter_NoTwoWeaponWithoutLight(t *testing.T) {
	// Character with non-light weapon in main hand
	char := &character.Character{
		Name:      "Fighter",
		Level:     1,
		Resources: &character.CharacterResources{},
		EquippedSlots: map[Slot]equipment.Equipment{
			SlotMainHand: &equipment.Weapon{
				Base:           equipment.BasicEquipment{Key: "longsword"},
				WeaponCategory: "Martial",
				WeaponRange:    "Melee",
				Properties: []*ReferenceItem{
					{Key: "versatile"},
				},
			},
			SlotOffHand: &equipment.Weapon{
				Base:           equipment.BasicEquipment{Key: "shortsword"},
				WeaponCategory: "Martial",
				WeaponRange:    "Melee",
				Properties: []*ReferenceItem{
					{Key: "light"},
					{Key: "finesse"},
				},
			},
		},
	}

	// Attack with non-light weapon
	char.RecordAction("attack", "weapon", "longsword")

	// No two-weapon fighting bonus
	assert.Empty(t, char.Resources.ActionEconomy.AvailableBonusActions)
}

func TestCharacter_ActionAvailability(t *testing.T) {
	char := &character.Character{
		Name:      "Test",
		Level:     1,
		Resources: &character.CharacterResources{},
	}

	// Initially has action available
	assert.True(t, char.HasActionAvailable())

	// Use action
	char.RecordAction("attack", "weapon", "longsword")
	assert.False(t, char.HasActionAvailable())

	// Start new turn
	char.StartNewTurn()
	assert.True(t, char.HasActionAvailable())
}

func TestCharacter_GetActionsTaken(t *testing.T) {
	char := &character.Character{
		Name:      "Test",
		Level:     1,
		Resources: &character.CharacterResources{},
	}

	// Take some actions
	char.RecordAction("attack", "weapon", "longsword")
	char.RecordAction("bonus_action", "dash", "")

	actions := char.GetActionsTaken()
	require.Len(t, actions, 2)
	assert.Equal(t, "attack", actions[0].Type)
	assert.Equal(t, "longsword", actions[0].WeaponKey)
	assert.Equal(t, "bonus_action", actions[1].Type)
	assert.Equal(t, "dash", actions[1].Subtype)
}

func TestCharacter_NonMonkNoMartialArts(t *testing.T) {
	// Create a fighter with a shortsword
	fighter := &character.Character{
		Name:  "Test Fighter",
		Level: 1,
		Features: []*rulebook.CharacterFeature{
			{Key: "second_wind", Name: "Second Wind"},
		},
		Resources: &character.CharacterResources{},
		EquippedSlots: map[Slot]equipment.Equipment{
			SlotMainHand: &equipment.Weapon{
				Base:           equipment.BasicEquipment{Key: equipment.WeaponKeyShortsword},
				WeaponCategory: "Martial",
				WeaponRange:    "Melee",
			},
		},
	}

	fighter.StartNewTurn()

	// Attack with shortsword
	fighter.RecordAction("attack", "weapon", equipment.WeaponKeyShortsword)

	// No martial arts bonus action available
	assert.Empty(t, fighter.Resources.ActionEconomy.AvailableBonusActions)
}
