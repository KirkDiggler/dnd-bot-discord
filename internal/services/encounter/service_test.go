package encounter_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/KirkDiggler/dnd-bot-discord/internal/entities"
	mockencrepo "github.com/KirkDiggler/dnd-bot-discord/internal/repositories/encounters/mock"
	mockcharacter "github.com/KirkDiggler/dnd-bot-discord/internal/services/character/mock"
	"github.com/KirkDiggler/dnd-bot-discord/internal/services/encounter"
	mocksession "github.com/KirkDiggler/dnd-bot-discord/internal/services/session/mock"
	mockuuid "github.com/KirkDiggler/dnd-bot-discord/internal/uuid/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestCreateEncounter_DungeonSession(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSessionService := mocksession.NewMockService(ctrl)
	mockCharService := mockcharacter.NewMockService(ctrl)
	mockRepo := mockencrepo.NewMockRepository(ctrl)
	mockUUID := mockuuid.NewMockGenerator(ctrl)

	svc := encounter.NewService(&encounter.ServiceConfig{
		Repository:       mockRepo,
		SessionService:   mockSessionService,
		CharacterService: mockCharService,
		UUIDGenerator:    mockUUID,
	})

	// Test case: Bot creates encounter for dungeon session
	t.Run("Bot can create encounter for dungeon session", func(t *testing.T) {
		botUserID := "bot-user-123"
		sessionID := "session-123"
		channelID := "channel-123"

		// Create a session with sessionType=dungeon in metadata
		dungeonSession := &entities.Session{
			ID:      sessionID,
			Name:    "Test Dungeon",
			Members: map[string]*entities.SessionMember{
				// Bot is not a member or DM
			},
			Metadata: map[string]interface{}{
				"sessionType": "dungeon",
			},
		}

		// Mock session service to return dungeon session
		mockSessionService.EXPECT().
			GetSession(gomock.Any(), sessionID).
			Return(dungeonSession, nil)

		// Mock UUID generation
		mockUUID.EXPECT().
			New().
			Return("enc-1")

		// Mock repository - expect GetActiveBySession to check for existing encounters
		mockRepo.EXPECT().
			GetActiveBySession(gomock.Any(), sessionID).
			Return(nil, nil) // No active encounter

		// Mock repository - expect Create to be called
		mockRepo.EXPECT().
			Create(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, enc *entities.Encounter) error {
				// Validate the encounter being created
				assert.Equal(t, "enc-1", enc.ID)
				assert.Equal(t, sessionID, enc.SessionID)
				assert.Equal(t, channelID, enc.ChannelID)
				assert.Equal(t, "Dungeon Room 1", enc.Name)
				assert.Equal(t, botUserID, enc.CreatedBy)
				return nil
			})

		// Create encResult as bot
		encResult, err := svc.CreateEncounter(context.Background(), &encounter.CreateEncounterInput{
			SessionID:   sessionID,
			ChannelID:   channelID,
			Name:        "Dungeon Room 1",
			Description: "A dark room",
			UserID:      botUserID,
		})

		// Should succeed for dungeon sessions
		require.NoError(t, err)
		assert.NotNil(t, encResult)
		assert.Equal(t, "Dungeon Room 1", encResult.Name)
		assert.Equal(t, sessionID, encResult.SessionID)
		assert.Equal(t, botUserID, encResult.CreatedBy)
	})

	// Test case: Non-DM cannot create encounter for regular session
	t.Run("Non-DM cannot create encounter for regular session", func(t *testing.T) {
		playerUserID := "player-123"
		sessionID := "session-456"
		channelID := "channel-456"

		// Create a regular session without dungeon metadata
		regularSession := &entities.Session{
			ID:   sessionID,
			Name: "Regular Campaign",
			Members: map[string]*entities.SessionMember{
				playerUserID: {
					UserID: playerUserID,
					Role:   entities.SessionRolePlayer,
				},
			},
			Metadata: map[string]interface{}{},
		}

		// Mock session service to return regular session
		mockSessionService.EXPECT().
			GetSession(gomock.Any(), sessionID).
			Return(regularSession, nil)

		// Try to create encounter as player
		_, err := svc.CreateEncounter(context.Background(), &encounter.CreateEncounterInput{
			SessionID:   sessionID,
			ChannelID:   channelID,
			Name:        "Test Encounter",
			Description: "Should fail",
			UserID:      playerUserID,
		})

		// Should fail with permission denied
		require.Error(t, err)
		assert.Contains(t, err.Error(), "only the DM can create encounters")
	})

	// Test case: DM can create encounter for regular session
	t.Run("DM can create encounter for regular session", func(t *testing.T) {
		dmUserID := "dm-123"
		sessionID := "session-789"
		channelID := "channel-789"

		// Create a regular session with DM
		regularSession := &entities.Session{
			ID:   sessionID,
			Name: "DM Campaign",
			Members: map[string]*entities.SessionMember{
				dmUserID: {
					UserID: dmUserID,
					Role:   entities.SessionRoleDM,
				},
			},
			Metadata: map[string]interface{}{},
		}

		// Mock session service to return session with DM
		mockSessionService.EXPECT().
			GetSession(gomock.Any(), sessionID).
			Return(regularSession, nil)

		// Mock UUID generation for DM test
		mockUUID.EXPECT().
			New().
			Return("enc-2")

		// Mock repository - expect GetActiveBySession to check for existing encounters
		mockRepo.EXPECT().
			GetActiveBySession(gomock.Any(), sessionID).
			Return(nil, nil) // No active encounter

		// Mock repository - expect Create to be called
		mockRepo.EXPECT().
			Create(gomock.Any(), gomock.Any()).
			Return(nil)

		// Create encounter as DM
		encResult, err := svc.CreateEncounter(context.Background(), &encounter.CreateEncounterInput{
			SessionID:   sessionID,
			ChannelID:   channelID,
			Name:        "DM Encounter",
			Description: "Created by DM",
			UserID:      dmUserID,
		})

		// Should succeed for DM
		require.NoError(t, err)
		assert.NotNil(t, encResult)
		assert.Equal(t, "DM Encounter", encResult.Name)
		assert.Equal(t, dmUserID, encResult.CreatedBy)
	})
}

func TestAddPlayer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSessionService := mocksession.NewMockService(ctrl)
	mockCharService := mockcharacter.NewMockService(ctrl)
	mockRepo := mockencrepo.NewMockRepository(ctrl)
	mockUUID := mockuuid.NewMockGenerator(ctrl)

	svc := encounter.NewService(&encounter.ServiceConfig{
		Repository:       mockRepo,
		SessionService:   mockSessionService,
		CharacterService: mockCharService,
		UUIDGenerator:    mockUUID,
	})

	// Create a test encounter
	encounterID := "enc-123"
	sessionID := "session-123"
	testEncounter := entities.NewEncounter(encounterID, sessionID, "channel-123", "Test Encounter", "dm-123")
	testEncounter.Status = entities.EncounterStatusSetup

	// Mock repository expectations for Get and Update calls
	mockRepo.EXPECT().
		Get(gomock.Any(), encounterID).
		Return(testEncounter, nil).
		AnyTimes()

	mockRepo.EXPECT().
		Update(gomock.Any(), gomock.Any()).
		Return(nil).
		AnyTimes()

	// Default mock for session service - can be overridden in specific tests
	mockSessionService.EXPECT().
		GetSession(gomock.Any(), sessionID).
		Return(&entities.Session{
			ID: sessionID,
			Metadata: map[string]interface{}{
				"sessionType": "combat", // Not a dungeon by default
			},
		}, nil).
		AnyTimes()

	t.Run("Successfully adds player with correct character data", func(t *testing.T) {
		playerID := "player-123"
		characterID := "char-123"

		// Mock UUID for combatant ID
		mockUUID.EXPECT().
			New().
			Return("combatant-1").
			Times(1)

		// Create test character
		testChar := &entities.Character{
			ID:               characterID,
			Name:             "Legolas",
			OwnerID:          playerID,
			CurrentHitPoints: 45,
			MaxHitPoints:     45,
			AC:               16,
			Attributes: map[entities.Attribute]*entities.AbilityScore{
				entities.AttributeDexterity: {Score: 18, Bonus: 4},
			},
		}

		// Mock character service to return test character
		mockCharService.EXPECT().
			GetByID(characterID).
			Return(testChar, nil)

		// Expect the character to be saved after action economy reset
		mockCharService.EXPECT().
			UpdateEquipment(gomock.Any()).
			Return(nil)

		// Using default session mock from parent test setup

		// Add player
		combatant, err := svc.AddPlayer(context.Background(), encounterID, playerID, characterID)

		// Verify success
		require.NoError(t, err)
		require.NotNil(t, combatant)

		// Verify combatant has correct data from character
		assert.Equal(t, "Legolas", combatant.Name)
		assert.Equal(t, entities.CombatantTypePlayer, combatant.Type)
		assert.Equal(t, playerID, combatant.PlayerID)
		assert.Equal(t, characterID, combatant.CharacterID)
		assert.Equal(t, 45, combatant.CurrentHP)
		assert.Equal(t, 45, combatant.MaxHP)
		assert.Equal(t, 16, combatant.AC)
		assert.Equal(t, 4, combatant.InitiativeBonus) // From DEX bonus
		assert.True(t, combatant.IsActive)

		// Verify combatant was added to encounter
		updatedEnc, err := mockRepo.Get(context.Background(), encounterID)
		require.NoError(t, err)
		assert.Len(t, updatedEnc.Combatants, 1)
		assert.NotNil(t, updatedEnc.Combatants[combatant.ID])
	})

	t.Run("Player name different from monster names", func(t *testing.T) {
		// Mock UUID for second combatant
		mockUUID.EXPECT().
			New().
			Return("combatant-2").
			Times(1)

		// Reset encounter
		testEncounter.Combatants = make(map[string]*entities.Combatant)

		// Add an Orc monster first
		orcCombatant := &entities.Combatant{
			ID:        "orc-1",
			Name:      "Orc",
			Type:      entities.CombatantTypeMonster,
			CurrentHP: 15,
			MaxHP:     15,
			AC:        13,
			IsActive:  true,
		}
		testEncounter.AddCombatant(orcCombatant)

		// Now add a player
		playerID := "player-456"
		characterID := "char-456"

		testChar := &entities.Character{
			ID:               characterID,
			Name:             "Aragorn", // Not "Orc"
			OwnerID:          playerID,
			CurrentHitPoints: 50,
			MaxHitPoints:     50,
			AC:               17,
			Attributes: map[entities.Attribute]*entities.AbilityScore{
				entities.AttributeDexterity: {Score: 14, Bonus: 2},
			},
		}

		mockCharService.EXPECT().
			GetByID(characterID).
			Return(testChar, nil)

		// Expect the character to be saved after action economy reset
		mockCharService.EXPECT().
			UpdateEquipment(gomock.Any()).
			Return(nil)

		// Add player
		_, err := svc.AddPlayer(context.Background(), encounterID, playerID, characterID)
		require.NoError(t, err)

		// Verify both combatants exist with correct names
		updatedEnc, err := mockRepo.Get(context.Background(), encounterID)
		require.NoError(t, err)
		assert.Len(t, updatedEnc.Combatants, 2)

		// Find each combatant and verify
		var foundOrc, foundPlayer bool
		for _, c := range updatedEnc.Combatants {
			if c.Type == entities.CombatantTypeMonster && c.Name == "Orc" {
				foundOrc = true
			}
			if c.Type == entities.CombatantTypePlayer && c.Name == "Aragorn" {
				foundPlayer = true
				// Verify it's our player
				assert.Equal(t, playerID, c.PlayerID)
			}
		}
		assert.True(t, foundOrc, "Should have Orc monster")
		assert.True(t, foundPlayer, "Should have Aragorn player")
	})

	t.Run("Handles player with same name as monster", func(t *testing.T) {
		// Mock UUID for third combatant
		mockUUID.EXPECT().
			New().
			Return("combatant-3").
			Times(1)

		// Reset encounter
		testEncounter.Combatants = make(map[string]*entities.Combatant)

		// Add a Goblin monster
		goblinMonster := &entities.Combatant{
			ID:        "goblin-1",
			Name:      "Goblin",
			Type:      entities.CombatantTypeMonster,
			CurrentHP: 7,
			MaxHP:     7,
			AC:        15,
			IsActive:  true,
		}
		testEncounter.AddCombatant(goblinMonster)

		// Add a player named "Goblin" (edge case)
		playerID := "player-789"
		characterID := "char-789"

		testChar := &entities.Character{
			ID:               characterID,
			Name:             "Goblin", // Same name as monster!
			OwnerID:          playerID,
			CurrentHitPoints: 25,
			MaxHitPoints:     25,
			AC:               14,
			Attributes: map[entities.Attribute]*entities.AbilityScore{
				entities.AttributeDexterity: {Score: 16, Bonus: 3},
			},
		}

		mockCharService.EXPECT().
			GetByID(characterID).
			Return(testChar, nil)

		// Expect the character to be saved after action economy reset
		mockCharService.EXPECT().
			UpdateEquipment(gomock.Any()).
			Return(nil)

		// Add player
		_, err := svc.AddPlayer(context.Background(), encounterID, playerID, characterID)
		require.NoError(t, err)

		// Verify both exist and can be distinguished by Type
		updatedEnc, getErr := mockRepo.Get(context.Background(), encounterID)
		require.NoError(t, getErr)
		assert.Len(t, updatedEnc.Combatants, 2)

		// Count each type
		var monsterGoblins, playerGoblins int
		for _, c := range updatedEnc.Combatants {
			if c.Name == "Goblin" {
				switch c.Type {
				case entities.CombatantTypeMonster:
					monsterGoblins++
					assert.Empty(t, c.PlayerID) // Monsters don't have PlayerID
				case entities.CombatantTypePlayer:
					playerGoblins++
					assert.Equal(t, playerID, c.PlayerID) // Players have PlayerID
					assert.Equal(t, characterID, c.CharacterID)
				}
			}
		}
		assert.Equal(t, 1, monsterGoblins, "Should have 1 monster Goblin")
		assert.Equal(t, 1, playerGoblins, "Should have 1 player Goblin")
	})

	t.Run("Fails when character doesn't belong to player", func(t *testing.T) {
		playerID := "player-999"
		characterID := "char-999"

		testChar := &entities.Character{
			ID:      characterID,
			Name:    "Gimli",
			OwnerID: "different-player", // Not the requesting player!
		}

		mockCharService.EXPECT().
			GetByID(characterID).
			Return(testChar, nil)

		// Using default session mock from parent test setup

		// Expect the character to be saved after action economy reset
		// (happens before ownership check)
		mockCharService.EXPECT().
			UpdateEquipment(gomock.Any()).
			Return(nil)

		// Try to add player with someone else's character
		combatant, err := svc.AddPlayer(context.Background(), encounterID, playerID, characterID)

		// Should fail with permission error
		require.Error(t, err)
		assert.Contains(t, err.Error(), "character does not belong to player")
		assert.Nil(t, combatant)
	})

	t.Run("Fails when player already in encounter", func(t *testing.T) {
		// Reset encounter and add a player
		testEncounter.Combatants = make(map[string]*entities.Combatant)
		existingCombatant := &entities.Combatant{
			ID:       "existing-1",
			Name:     "Existing Player",
			Type:     entities.CombatantTypePlayer,
			PlayerID: "player-123", // This player is already in
		}
		testEncounter.AddCombatant(existingCombatant)

		// Try to add same player again
		characterID := "char-new"
		testChar := &entities.Character{
			ID:      characterID,
			Name:    "New Character",
			OwnerID: "player-123", // Same player!
		}

		mockCharService.EXPECT().
			GetByID(characterID).
			Return(testChar, nil)

		// Using default session mock from parent test setup

		// Expect the character to be saved after action economy reset
		// (happens before duplicate player check)
		mockCharService.EXPECT().
			UpdateEquipment(gomock.Any()).
			Return(nil)

		combatant, err := svc.AddPlayer(context.Background(), encounterID, "player-123", characterID)

		// Should fail
		require.Error(t, err)
		assert.Contains(t, err.Error(), "player is already in the encounter")
		assert.Nil(t, combatant)
	})
}

func TestEncounterCombatantFiltering(t *testing.T) {
	// Test that we can properly filter combatants by type
	encounterResult := entities.NewEncounter("enc-1", "session-1", "channel-1", "Mixed Combat", "dm-1")

	// Add various combatants
	player1 := &entities.Combatant{
		ID:       "p1",
		Name:     "Aragorn",
		Type:     entities.CombatantTypePlayer,
		PlayerID: "player-1",
		IsActive: true,
	}

	player2 := &entities.Combatant{
		ID:       "p2",
		Name:     "Orc", // Player named Orc!
		Type:     entities.CombatantTypePlayer,
		PlayerID: "player-2",
		IsActive: true,
	}

	monster1 := &entities.Combatant{
		ID:       "m1",
		Name:     "Orc",
		Type:     entities.CombatantTypeMonster,
		IsActive: true,
	}

	monster2 := &entities.Combatant{
		ID:       "m2",
		Name:     "Goblin",
		Type:     entities.CombatantTypeMonster,
		IsActive: true,
	}

	encounterResult.AddCombatant(player1)
	encounterResult.AddCombatant(player2)
	encounterResult.AddCombatant(monster1)
	encounterResult.AddCombatant(monster2)

	t.Run("Can filter for monsters only", func(t *testing.T) {
		var monsters []*entities.Combatant
		for _, c := range encounterResult.Combatants {
			if c.Type == entities.CombatantTypeMonster {
				monsters = append(monsters, c)
			}
		}

		assert.Len(t, monsters, 2)
		// Both should be monsters
		for _, m := range monsters {
			assert.Equal(t, entities.CombatantTypeMonster, m.Type)
			assert.Empty(t, m.PlayerID)
		}
	})

	t.Run("Can filter for players only", func(t *testing.T) {
		var players []*entities.Combatant
		for _, c := range encounterResult.Combatants {
			if c.Type == entities.CombatantTypePlayer {
				players = append(players, c)
			}
		}

		assert.Len(t, players, 2)
		// Both should be players
		for _, p := range players {
			assert.Equal(t, entities.CombatantTypePlayer, p.Type)
			assert.NotEmpty(t, p.PlayerID)
		}
	})

	t.Run("Can distinguish same-named player and monster", func(t *testing.T) {
		var orcs []*entities.Combatant
		for _, c := range encounterResult.Combatants {
			if c.Name == "Orc" {
				orcs = append(orcs, c)
			}
		}

		assert.Len(t, orcs, 2, "Should find 2 combatants named Orc")

		// One should be player, one should be monster
		var foundPlayer, foundMonster bool
		for _, orc := range orcs {
			if orc.Type == entities.CombatantTypePlayer {
				foundPlayer = true
				assert.Equal(t, "player-2", orc.PlayerID)
			}
			if orc.Type == entities.CombatantTypeMonster {
				foundMonster = true
				assert.Empty(t, orc.PlayerID)
			}
		}
		assert.True(t, foundPlayer, "Should find player Orc")
		assert.True(t, foundMonster, "Should find monster Orc")
	})
}

func TestUpdateMessageID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockSessionSvc := mocksession.NewMockService(ctrl)
	mockCharSvc := mockcharacter.NewMockService(ctrl)
	mockRepo := mockencrepo.NewMockRepository(ctrl)
	mockUUID := mockuuid.NewMockGenerator(ctrl)

	svc := encounter.NewService(&encounter.ServiceConfig{
		Repository:       mockRepo,
		SessionService:   mockSessionSvc,
		CharacterService: mockCharSvc,
		UUIDGenerator:    mockUUID,
	})

	// Create test session
	session := &entities.Session{
		ID: "session-1",
		Members: map[string]*entities.SessionMember{
			"dm-1": {Role: entities.SessionRoleDM},
		},
	}

	mockSessionSvc.EXPECT().GetSession(ctx, "session-1").Return(session, nil).AnyTimes()

	// Mock UUID generation
	mockUUID.EXPECT().New().Return("enc-1")

	// Mock repository for create
	mockRepo.EXPECT().
		GetActiveBySession(gomock.Any(), "session-1").
		Return(nil, nil) // No active encounter

	mockRepo.EXPECT().
		Create(gomock.Any(), gomock.Any()).
		Return(nil)

	// Create encounter
	input := &encounter.CreateEncounterInput{
		SessionID:   "session-1",
		ChannelID:   "channel-1",
		Name:        "Test Encounter",
		Description: "Test Description",
		UserID:      "dm-1",
	}

	enc, err := svc.CreateEncounter(ctx, input)
	require.NoError(t, err)
	assert.Equal(t, "enc-1", enc.ID)
	assert.Empty(t, enc.MessageID) // Initially empty

	// Test successful update
	t.Run("Success", func(t *testing.T) {
		// Mock Get to return the encounter
		mockRepo.EXPECT().
			Get(gomock.Any(), enc.ID).
			Return(enc, nil)

		// Mock Update to save changes
		mockRepo.EXPECT().
			Update(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, e *entities.Encounter) error {
				// Update the local encounter object to simulate repository behavior
				enc.MessageID = e.MessageID
				enc.ChannelID = e.ChannelID
				return nil
			})

		err := svc.UpdateMessageID(ctx, enc.ID, "msg-123", "channel-456")
		require.NoError(t, err)

		// Verify the update
		assert.Equal(t, "msg-123", enc.MessageID)
		assert.Equal(t, "channel-456", enc.ChannelID)
	})

	// Test validation errors
	t.Run("ValidationErrors", func(t *testing.T) {
		// Empty encounter ID
		err := svc.UpdateMessageID(ctx, "", "msg-123", "channel-456")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "encounter ID is required")

		// Empty message ID
		err = svc.UpdateMessageID(ctx, enc.ID, "", "channel-456")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "message ID is required")

		// Empty channel ID
		err = svc.UpdateMessageID(ctx, enc.ID, "msg-123", "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "channel ID is required")
	})

	// Test non-existent encounter
	t.Run("NonExistentEncounter", func(t *testing.T) {
		// Mock Get to return not found
		mockRepo.EXPECT().
			Get(gomock.Any(), "non-existent").
			Return(nil, fmt.Errorf("encounter not found"))

		err := svc.UpdateMessageID(ctx, "non-existent", "msg-123", "channel-456")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "encounter not found")
	})
}
