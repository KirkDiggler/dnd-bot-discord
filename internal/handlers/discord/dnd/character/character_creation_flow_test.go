package character_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/KirkDiggler/dnd-bot-discord/internal/entities"
	"github.com/KirkDiggler/dnd-bot-discord/internal/handlers/discord/dnd/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/repositories/characters"
	characterService "github.com/KirkDiggler/dnd-bot-discord/internal/services/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/testutils"
	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockDiscordSession for testing
type MockDiscordSession struct {
	discordgo.Session
	RespondFunc     func(*discordgo.Interaction, *discordgo.InteractionResponse) error
	ResponseEditFunc func(*discordgo.Interaction, *discordgo.WebhookEdit) (*discordgo.Message, error)
}

func (m *MockDiscordSession) InteractionRespond(i *discordgo.Interaction, r *discordgo.InteractionResponse) error {
	if m.RespondFunc != nil {
		return m.RespondFunc(i, r)
	}
	return nil
}

func (m *MockDiscordSession) InteractionResponseEdit(i *discordgo.Interaction, e *discordgo.WebhookEdit) (*discordgo.Message, error) {
	if m.ResponseEditFunc != nil {
		return m.ResponseEditFunc(i, e)
	}
	return &discordgo.Message{}, nil
}

func TestCharacterCreation_AbilityAssignmentFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// Setup services
	redisClient := testutils.CreateTestRedisClient(t, nil)
	repo := characters.NewRedisRepository(&characters.RedisRepoConfig{
		Client: redisClient,
	})
	charService := characterService.NewService(&characterService.ServiceConfig{
		Repository: repo,
	})

	// Create handlers
	assignHandler := character.NewAssignAbilitiesHandler(&character.AssignAbilitiesHandlerConfig{
		CharacterService: charService,
	})

	ctx := context.Background()
	userID := "discord-user-123"
	guildID := "guild-123"

	// Step 1: Create draft with race, class, and ability rolls
	draft, err := charService.GetOrCreateDraftCharacter(ctx, userID, guildID)
	require.NoError(t, err)

	// Set up the draft with race, class, and rolls
	raceKey := "half-orc"
	classKey := "barbarian"
	_, err = charService.UpdateDraftCharacter(ctx, draft.ID, &characterService.UpdateDraftInput{
		RaceKey:  &raceKey,
		ClassKey: &classKey,
		AbilityRolls: []entities.AbilityRoll{
			{ID: "roll_1", Value: 18},
			{ID: "roll_2", Value: 16},
			{ID: "roll_3", Value: 14},
			{ID: "roll_4", Value: 12},
			{ID: "roll_5", Value: 11},
			{ID: "roll_6", Value: 9},
		},
	})
	require.NoError(t, err)

	// Step 2: Simulate the assign abilities interaction
	mockSession := &MockDiscordSession{
		Session: &discordgo.Session{},
	}

	var capturedEmbed *discordgo.MessageEmbed
	mockSession.RespondFunc = func(i *discordgo.Interaction, r *discordgo.InteractionResponse) error {
		assert.Equal(t, discordgo.InteractionResponseUpdateMessage, r.Type)
		return nil
	}
	mockSession.ResponseEditFunc = func(i *discordgo.Interaction, e *discordgo.WebhookEdit) (*discordgo.Message, error) {
		if e.Embeds != nil && len(*e.Embeds) > 0 {
			capturedEmbed = (*e.Embeds)[0]
		}
		return &discordgo.Message{}, nil
	}

	interaction := &discordgo.InteractionCreate{
		Interaction: &discordgo.Interaction{
			Type: discordgo.InteractionMessageComponent,
			Member: &discordgo.Member{
				User: &discordgo.User{ID: userID},
			},
			GuildID: guildID,
			Data: discordgo.MessageComponentInteractionData{
				CustomID: "assign_ability:STR:roll_1",
			},
		},
	}

	// Handle the ability assignment
	req := &character.AssignAbilitiesRequest{
		Session:     &mockSession.Session,
		Interaction: interaction,
		RaceKey:     raceKey,
		ClassKey:    classKey,
	}

	err = assignHandler.Handle(req)
	require.NoError(t, err)

	// Verify the embed was created
	assert.NotNil(t, capturedEmbed)
	assert.Contains(t, capturedEmbed.Title, "Assign Abilities")

	// Step 3: Complete all assignments
	assignments := map[string]string{
		"STR": "roll_1", // 18
		"CON": "roll_2", // 16
		"DEX": "roll_3", // 14
		"WIS": "roll_4", // 12
		"INT": "roll_5", // 11
		"CHA": "roll_6", // 9
	}

	for ability, rollID := range assignments {
		interaction.Data = discordgo.MessageComponentInteractionData{
			CustomID: fmt.Sprintf("assign_ability:%s:%s", ability, rollID),
		}
		err = assignHandler.Handle(req)
		require.NoError(t, err)
	}

	// Step 4: Verify assignments were saved
	updatedDraft, err := charService.GetCharacter(ctx, draft.ID)
	require.NoError(t, err)
	assert.Equal(t, assignments, updatedDraft.AbilityAssignments)

	// Step 5: Finalize and verify conversion
	charName := "Grok the Barbarian"
	_, err = charService.UpdateDraftCharacter(ctx, draft.ID, &characterService.UpdateDraftInput{
		Name: &charName,
	})
	require.NoError(t, err)

	finalChar, err := charService.FinalizeDraftCharacter(ctx, draft.ID)
	require.NoError(t, err)

	// Verify final attributes (half-orc gets +2 STR, +1 CON)
	assert.Equal(t, 20, finalChar.Attributes[entities.AttributeStrength].Score)     // 18 + 2
	assert.Equal(t, 17, finalChar.Attributes[entities.AttributeConstitution].Score) // 16 + 1
	assert.Equal(t, 14, finalChar.Attributes[entities.AttributeDexterity].Score)
	assert.Equal(t, 12, finalChar.Attributes[entities.AttributeWisdom].Score)
	assert.Equal(t, 11, finalChar.Attributes[entities.AttributeIntelligence].Score)
	assert.Equal(t, 9, finalChar.Attributes[entities.AttributeCharisma].Score)

	// Verify modifiers
	assert.Equal(t, 5, finalChar.Attributes[entities.AttributeStrength].Bonus)
	assert.Equal(t, 3, finalChar.Attributes[entities.AttributeConstitution].Bonus)

	// Verify character is complete
	assert.True(t, finalChar.IsComplete())
	assert.Equal(t, entities.CharacterStatusActive, finalChar.Status)
}

func TestCharacterCreation_AutoAssign(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// Setup services
	redisClient := testutils.CreateTestRedisClient(t, nil)
	repo := characters.NewRedisRepository(&characters.RedisRepoConfig{
		Client: redisClient,
	})
	charService := characterService.NewService(&characterService.ServiceConfig{
		Repository: repo,
	})

	// Create handlers
	assignHandler := character.NewAssignAbilitiesHandler(&character.AssignAbilitiesHandlerConfig{
		CharacterService: charService,
	})

	ctx := context.Background()
	userID := "discord-user-456"
	guildID := "guild-456"

	// Create draft with rolls
	draft, err := charService.GetOrCreateDraftCharacter(ctx, userID, guildID)
	require.NoError(t, err)

	raceKey := "dwarf"
	classKey := "cleric"
	_, err = charService.UpdateDraftCharacter(ctx, draft.ID, &characterService.UpdateDraftInput{
		RaceKey:  &raceKey,
		ClassKey: &classKey,
		AbilityRolls: []entities.AbilityRoll{
			{ID: "roll_1", Value: 17},
			{ID: "roll_2", Value: 15},
			{ID: "roll_3", Value: 13},
			{ID: "roll_4", Value: 12},
			{ID: "roll_5", Value: 10},
			{ID: "roll_6", Value: 8},
		},
	})
	require.NoError(t, err)

	// Test auto-assign
	mockSession := &MockDiscordSession{Session: &discordgo.Session{}}
	interaction := &discordgo.InteractionCreate{
		Interaction: &discordgo.Interaction{
			Type: discordgo.InteractionMessageComponent,
			Member: &discordgo.Member{
				User: &discordgo.User{ID: userID},
			},
			GuildID: guildID,
			Data: discordgo.MessageComponentInteractionData{
				CustomID: "auto_assign",
			},
		},
	}

	req := &character.AssignAbilitiesRequest{
		Session:     &mockSession.Session,
		Interaction: interaction,
		RaceKey:     raceKey,
		ClassKey:    classKey,
	}

	err = assignHandler.Handle(req)
	require.NoError(t, err)

	// Verify auto-assign put highest rolls in WIS and CON for cleric
	updatedDraft, err := charService.GetCharacter(ctx, draft.ID)
	require.NoError(t, err)
	
	// For cleric, the priority should be WIS, CON, STR, CHA, DEX, INT
	assert.NotEmpty(t, updatedDraft.AbilityAssignments)
	
	// Check that WIS got the highest roll
	wisRollID := updatedDraft.AbilityAssignments["WIS"]
	var wisRoll *entities.AbilityRoll
	for _, roll := range updatedDraft.AbilityRolls {
		if roll.ID == wisRollID {
			wisRoll = &roll
			break
		}
	}
	assert.NotNil(t, wisRoll)
	assert.Equal(t, 17, wisRoll.Value) // Highest roll should go to WIS
}