package encounter_test

import (
	"context"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/game/combat"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/game/session"
	"strings"
	"testing"

	mockencrepo "github.com/KirkDiggler/dnd-bot-discord/internal/repositories/encounters/mock"
	mockcharacter "github.com/KirkDiggler/dnd-bot-discord/internal/services/character/mock"
	"github.com/KirkDiggler/dnd-bot-discord/internal/services/encounter"
	mocksession "github.com/KirkDiggler/dnd-bot-discord/internal/services/session/mock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

// TestInitiativeRollDisplay tests the fix for issue #60
func TestInitiativeRollDisplay(t *testing.T) {
	t.Run("Initiative rolls should show combatant names and individual rolls", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mockencrepo.NewMockRepository(ctrl)
		mockSessionService := mocksession.NewMockService(ctrl)
		mockCharacterService := mockcharacter.NewMockService(ctrl)

		service := encounter.NewService(&encounter.ServiceConfig{
			Repository:       mockRepo,
			SessionService:   mockSessionService,
			CharacterService: mockCharacterService,
		})

		ctx := context.Background()
		encounterID := "test-encounter-1"
		userID := "dm-user"

		// Create encounter with combatants
		enc := &combat.Encounter{
			ID:        encounterID,
			CreatedBy: userID,
			SessionID: "test-session",
			Status:    combat.EncounterStatusSetup,
			Combatants: map[string]*combat.Combatant{
				"goblin1": {
					ID:              "goblin1",
					Name:            "Goblin",
					Type:            combat.CombatantTypeMonster,
					InitiativeBonus: 2,
					MaxHP:           7,
					CurrentHP:       7,
					AC:              15,
					IsActive:        true,
				},
				"player1": {
					ID:              "player1",
					Name:            "Aragorn",
					Type:            combat.CombatantTypePlayer,
					InitiativeBonus: 3,
					MaxHP:           20,
					CurrentHP:       20,
					AC:              16,
					IsActive:        true,
				},
				"goblin2": {
					ID:              "goblin2",
					Name:            "Goblin Archer",
					Type:            combat.CombatantTypeMonster,
					InitiativeBonus: 2,
					MaxHP:           7,
					CurrentHP:       7,
					AC:              15,
					IsActive:        true,
				},
			},
			TurnOrder: []string{},
			CombatLog: []string{},
		}

		// Mock repository calls
		mockRepo.EXPECT().Get(ctx, encounterID).Return(enc, nil)
		mockRepo.EXPECT().Update(ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, e *combat.Encounter) error {
			// Verify combat log contains expected entries
			assert.NotEmpty(t, e.CombatLog)
			assert.Equal(t, "🎲 **Rolling Initiative**", e.CombatLog[0])

			// Check that each combatant has an initiative roll entry
			combatantNames := []string{"Goblin", "Aragorn", "Goblin Archer"}
			foundCombatants := make(map[string]bool)

			for i := 1; i < len(e.CombatLog); i++ {
				entry := e.CombatLog[i]
				// Each entry should contain combatant name and roll details
				for _, name := range combatantNames {
					if strings.Contains(entry, name) && strings.Contains(entry, "rolls initiative:") {
						foundCombatants[name] = true
						// Verify format: "**Name** rolls initiative: [roll] + bonus = **total**"
						assert.Contains(t, entry, "rolls initiative:")
						assert.Contains(t, entry, "+")
						assert.Contains(t, entry, "=")
					}
				}
			}

			// Verify all combatants were found
			for _, name := range combatantNames {
				assert.True(t, foundCombatants[name], "Initiative roll for %s not found in combat log", name)
			}

			return nil
		})

		// Roll initiative
		err := service.RollInitiative(ctx, encounterID, userID)
		assert.NoError(t, err)
	})

	t.Run("Initiative display should work for dungeon encounters", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mockencrepo.NewMockRepository(ctrl)
		mockSessionService := mocksession.NewMockService(ctrl)
		mockCharacterService := mockcharacter.NewMockService(ctrl)

		service := encounter.NewService(&encounter.ServiceConfig{
			Repository:       mockRepo,
			SessionService:   mockSessionService,
			CharacterService: mockCharacterService,
		})

		ctx := context.Background()
		encounterID := "dungeon-encounter-1"
		botID := "system"
		dmID := "dm-user"

		// Create dungeon encounter (created by a different user to trigger permission check)
		enc := &combat.Encounter{
			ID:        encounterID,
			CreatedBy: dmID,
			SessionID: "dungeon-session",
			Status:    combat.EncounterStatusSetup,
			Combatants: map[string]*combat.Combatant{
				"skeleton1": {
					ID:              "skeleton1",
					Name:            "Skeleton",
					Type:            combat.CombatantTypeMonster,
					InitiativeBonus: 2,
					MaxHP:           13,
					CurrentHP:       13,
					AC:              13,
					IsActive:        true,
				},
				"player1": {
					ID:              "player1",
					Name:            "Legolas",
					Type:            combat.CombatantTypePlayer,
					InitiativeBonus: 4,
					MaxHP:           18,
					CurrentHP:       18,
					AC:              15,
					IsActive:        true,
				},
			},
			TurnOrder: []string{},
			CombatLog: []string{},
		}

		// Mock session service for dungeon check (called during permission check)
		session := &session.Session{
			ID: "dungeon-session",
			Metadata: map[string]interface{}{
				"sessionType": "dungeon",
			},
		}
		mockSessionService.EXPECT().GetSession(ctx, "dungeon-session").Return(session, nil).Times(1)

		// Mock repository calls
		mockRepo.EXPECT().Get(ctx, encounterID).Return(enc, nil)
		mockRepo.EXPECT().Update(ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, e *combat.Encounter) error {
			// Verify combat log for dungeon encounter
			assert.NotEmpty(t, e.CombatLog)

			// Check for skeleton and player initiative rolls
			foundSkeleton := false
			foundPlayer := false

			for _, entry := range e.CombatLog {
				if strings.Contains(entry, "Skeleton") && strings.Contains(entry, "rolls initiative:") {
					foundSkeleton = true
				}
				if strings.Contains(entry, "Legolas") && strings.Contains(entry, "rolls initiative:") {
					foundPlayer = true
				}
			}

			assert.True(t, foundSkeleton, "Skeleton initiative roll not found")
			assert.True(t, foundPlayer, "Player initiative roll not found")

			return nil
		})

		// Roll initiative as bot
		err := service.RollInitiative(ctx, encounterID, botID)
		assert.NoError(t, err)
	})
}
