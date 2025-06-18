package discord_test

import (
	"context"
	"testing"

	mockdnd5e "github.com/KirkDiggler/dnd-bot-discord/internal/clients/dnd5e/mock"
	"github.com/KirkDiggler/dnd-bot-discord/internal/entities"
	"github.com/KirkDiggler/dnd-bot-discord/internal/handlers/discord"
	mockcharrepo "github.com/KirkDiggler/dnd-bot-discord/internal/repositories/characters/mock"
	"github.com/KirkDiggler/dnd-bot-discord/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"github.com/bwmarrin/discordgo"
)

func TestCharacterCreationBug_ReproduceRealWorldFailure(t *testing.T) {
	// This test reproduces the exact bug: characters end up with 0 attributes
	// even though they go through the complete creation flow
	
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mockdnd5e.NewMockClient(ctrl)
	mockRepo := mockcharrepo.NewMockRepository(ctrl)
	
	// Set up service provider
	provider := services.NewProvider(&services.ProviderConfig{
		DNDClient:           mockClient,
		CharacterRepository: mockRepo,
	})
	
	handler := discord.NewHandler(&discord.HandlerConfig{
		ServiceProvider: provider,
	})

	// Mock Discord session and interaction
	session := &discordgo.Session{}
	user := &discordgo.User{ID: "test_user"}
	interaction := &discordgo.InteractionCreate{
		Interaction: &discordgo.Interaction{
			Type: discordgo.InteractionMessageComponent,
			Data: discordgo.ApplicationCommandInteractionData{},
			Member: &discordgo.Member{User: user},
			GuildID: "test_guild",
		},
	}

	ctx := context.Background()

	// Step 1: Character creation starts - user clicks "Create Character"
	// This should create a draft character
	draftChar := &entities.Character{
		ID:      "draft_123",
		OwnerID: user.ID,
		RealmID: "test_guild",
		Name:    "Draft Character",
		Status:  entities.CharacterStatusDraft,
		Level:   1,
		Attributes:    make(map[entities.Attribute]*entities.AbilityScore),
		Proficiencies: make(map[entities.ProficiencyType][]*entities.Proficiency),
		Inventory:     make(map[entities.EquipmentType][]entities.Equipment),
		EquippedSlots: make(map[entities.Slot]entities.Equipment),
	}

	// Mock: GetOrCreateDraftCharacter
	mockRepo.EXPECT().GetByOwnerAndRealm(ctx, user.ID, "test_guild").Return([]*entities.Character{}, nil)
	mockRepo.EXPECT().Create(ctx, gomock.Any()).Return(nil)

	// Step 2-5: User goes through race/class/abilities/proficiencies/equipment selection
	// Each step calls UpdateDraftCharacter with new data
	
	// Mock race selection
	mockClient.EXPECT().GetRace("elf").Return(&entities.Race{
		Key:  "elf",
		Name: "Elf",
		AbilityBonuses: []*entities.AbilityBonus{
			{Attribute: entities.AttributeDexterity, Bonus: 2},
		},
	}, nil).AnyTimes()

	// Mock class selection  
	mockClient.EXPECT().GetClass("monk").Return(&entities.Class{
		Key:    "monk",
		Name:   "Monk", 
		HitDie: 8,
	}, nil).AnyTimes()

	// This is the key: character after ability assignment but before finalization
	charWithAbilities := &entities.Character{
		ID:      "draft_123",
		OwnerID: user.ID,
		RealmID: "test_guild",
		Name:    "Draft Character",
		Status:  entities.CharacterStatusDraft,
		Level:   1,
		Race: &entities.Race{
			Key:  "elf",
			Name: "Elf",
			AbilityBonuses: []*entities.AbilityBonus{
				{Attribute: entities.AttributeDexterity, Bonus: 2},
			},
		},
		Class: &entities.Class{
			Key:    "monk",
			Name:   "Monk",
			HitDie: 8,
		},
		// This is what should happen: AbilityAssignments and AbilityRolls are set
		AbilityRolls: []entities.AbilityRoll{
			{ID: "roll_1", Value: 15},
			{ID: "roll_2", Value: 14},
			{ID: "roll_3", Value: 13},
			{ID: "roll_4", Value: 12},
			{ID: "roll_5", Value: 11},
			{ID: "roll_6", Value: 10},
		},
		AbilityAssignments: map[string]string{
			"STR": "roll_3",
			"DEX": "roll_2", 
			"CON": "roll_4",
			"INT": "roll_1",
			"WIS": "roll_5",
			"CHA": "roll_6",
		},
		// But Attributes should be empty at this point (conversion happens in UpdateDraftCharacter)
		Attributes:    make(map[entities.Attribute]*entities.AbilityScore),
		Proficiencies: make(map[entities.ProficiencyType][]*entities.Proficiency),
		Inventory:     make(map[entities.EquipmentType][]entities.Equipment),
		EquippedSlots: make(map[entities.Slot]entities.Equipment),
	}

	// Mock: Multiple UpdateDraftCharacter calls during creation process
	mockRepo.EXPECT().Get(ctx, "draft_123").Return(draftChar, nil).AnyTimes()
	mockRepo.EXPECT().Update(ctx, gomock.Any()).Return(nil).AnyTimes()

	// Step 6: Final step - user enters character name and character is finalized
	// This simulates the modal submit that calls FinalizeCharacterWithName
	
	// The bug reproduction: Get character for finalization
	mockRepo.EXPECT().Get(ctx, "draft_123").Return(charWithAbilities, nil)
	
	// Mock: FinalizeDraftCharacter should convert abilities and mark as active
	finalizedChar := &entities.Character{
		ID:      "draft_123",
		OwnerID: user.ID,
		RealmID: "test_guild",
		Name:    "TestMonk", // Name added
		Status:  entities.CharacterStatusActive, // Status changed
		Level:   1,
		Race:    charWithAbilities.Race,
		Class:   charWithAbilities.Class,
		// THE BUG: In real world, this ends up empty despite having AbilityAssignments
		AbilityRolls:       charWithAbilities.AbilityRolls,
		AbilityAssignments: charWithAbilities.AbilityAssignments,
		Attributes:         make(map[entities.Attribute]*entities.AbilityScore), // EMPTY!
		Proficiencies:      make(map[entities.ProficiencyType][]*entities.Proficiency),
		Inventory:          make(map[entities.EquipmentType][]entities.Equipment),
		EquippedSlots:      make(map[entities.Slot]entities.Equipment),
	}

	// This should save the broken character (reproducing the bug)
	mockRepo.EXPECT().Update(ctx, gomock.Any()).DoAndReturn(func(_ context.Context, char *entities.Character) error {
		// Verify this is the broken state we see in production
		assert.Equal(t, entities.CharacterStatusActive, char.Status, "Character should be finalized")
		assert.Empty(t, char.Attributes, "BUG: Character has empty attributes")
		assert.NotEmpty(t, char.AbilityAssignments, "Character should still have ability assignments")
		assert.False(t, char.IsComplete(), "Character should be incomplete due to missing attributes")
		return nil
	})

	// Call the actual service method that the handler uses
	result, err := provider.CharacterService.FinalizeCharacterWithName(
		ctx, 
		"draft_123", 
		"TestMonk", 
		"elf", 
		"monk",
	)
	
	// This test should FAIL initially, proving we can reproduce the bug
	require.NoError(t, err)
	assert.NotEmpty(t, result.Attributes, "Character should have attributes after finalization")
	assert.True(t, result.IsComplete(), "Character should be complete")
}