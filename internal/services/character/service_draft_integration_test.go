package character_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	mockdnd "github.com/KirkDiggler/dnd-bot-discord/internal/clients/dnd5e/mock"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/rulebook/dnd5e"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/shared"
	inmemoryDraft "github.com/KirkDiggler/dnd-bot-discord/internal/repositories/character_draft"
	inmemoryChar "github.com/KirkDiggler/dnd-bot-discord/internal/repositories/characters"
	characterService "github.com/KirkDiggler/dnd-bot-discord/internal/services/character"
)

func TestService_DraftIntegration(t *testing.T) {
	setup := func(t *testing.T) (
		characterService.Service,
		*mockdnd.MockClient,
		context.Context,
	) {
		ctrl := gomock.NewController(t)
		t.Cleanup(ctrl.Finish)

		mockDNDClient := mockdnd.NewMockClient(ctrl)
		charRepo := inmemoryChar.NewInMemoryRepository()
		draftRepo := inmemoryDraft.NewInMemoryRepository()

		cfg := &characterService.ServiceConfig{
			Repository:      charRepo,
			DraftRepository: draftRepo,
			DNDClient:       mockDNDClient,
		}

		svc := characterService.NewService(cfg)
		ctx := context.Background()

		return svc, mockDNDClient, ctx
	}

	t.Run("GetOrCreateDraftCharacter creates draft with flow state", func(t *testing.T) {
		svc, _, ctx := setup(t)
		userID := "user-123"
		realmID := "realm-456"

		// Create draft
		char, err := svc.GetOrCreateDraftCharacter(ctx, userID, realmID)
		require.NoError(t, err)
		assert.NotNil(t, char)
		assert.Equal(t, userID, char.OwnerID)
		assert.Equal(t, realmID, char.RealmID)
		assert.Equal(t, shared.CharacterStatusDraft, char.Status)

		// Get it again should return same character
		char2, err := svc.GetOrCreateDraftCharacter(ctx, userID, realmID)
		require.NoError(t, err)
		assert.Equal(t, char.ID, char2.ID)
	})

	t.Run("UpdateDraftCharacter tracks flow state", func(t *testing.T) {
		svc, mockDND, ctx := setup(t)
		userID := "user-123"
		realmID := "realm-456"

		// Create draft
		char, err := svc.GetOrCreateDraftCharacter(ctx, userID, realmID)
		require.NoError(t, err)

		// Mock race
		mockDND.EXPECT().
			GetRace("human").
			Return(&rulebook.Race{
				Key:   "human",
				Name:  "Human",
				Speed: 30,
			}, nil)

		// Update with race
		raceKey := "human"
		updated, err := svc.UpdateDraftCharacter(ctx, char.ID, &characterService.UpdateDraftInput{
			RaceKey: &raceKey,
		})
		require.NoError(t, err)
		assert.Equal(t, "human", updated.Race.Key)
		assert.Equal(t, 30, updated.Speed)

		// Mock class
		mockDND.EXPECT().
			GetClass("wizard").
			Return(&rulebook.Class{
				Key:    "wizard",
				Name:   "Wizard",
				HitDie: 6,
			}, nil)

		// Update with class
		classKey := "wizard"
		updated, err = svc.UpdateDraftCharacter(ctx, char.ID, &characterService.UpdateDraftInput{
			ClassKey: &classKey,
		})
		require.NoError(t, err)
		assert.Equal(t, "wizard", updated.Class.Key)
		assert.Equal(t, 6, updated.HitDie)
	})

	t.Run("FinalizeDraftCharacter removes draft", func(t *testing.T) {
		svc, mockDND, ctx := setup(t)
		userID := "user-123"
		realmID := "realm-456"

		// Create draft
		char, err := svc.GetOrCreateDraftCharacter(ctx, userID, realmID)
		require.NoError(t, err)

		// Set required fields
		name := "Gandalf"
		raceKey := "human"
		classKey := "wizard"

		// Mock race
		mockDND.EXPECT().
			GetRace("human").
			Return(&rulebook.Race{
				Key:   "human",
				Name:  "Human",
				Speed: 30,
			}, nil)

		// Mock class
		mockDND.EXPECT().
			GetClass("wizard").
			Return(&rulebook.Class{
				Key:    "wizard",
				Name:   "Wizard",
				HitDie: 6,
			}, nil)

		// Update with required data
		_, err = svc.UpdateDraftCharacter(ctx, char.ID, &characterService.UpdateDraftInput{
			Name:     &name,
			RaceKey:  &raceKey,
			ClassKey: &classKey,
			AbilityAssignments: map[string]string{
				"STR": "roll-1",
				"DEX": "roll-2",
				"CON": "roll-3",
				"INT": "roll-4",
				"WIS": "roll-5",
				"CHA": "roll-6",
			},
			AbilityRolls: []character.AbilityRoll{
				{ID: "roll-1", Value: 10},
				{ID: "roll-2", Value: 12},
				{ID: "roll-3", Value: 14},
				{ID: "roll-4", Value: 16},
				{ID: "roll-5", Value: 13},
				{ID: "roll-6", Value: 11},
			},
		})
		require.NoError(t, err)

		// Finalize
		finalized, err := svc.FinalizeDraftCharacter(ctx, char.ID)
		require.NoError(t, err)
		assert.Equal(t, shared.CharacterStatusActive, finalized.Status)
		assert.Equal(t, "Gandalf", finalized.Name)

		// Creating new draft should give new character (old draft deleted)
		newChar, err := svc.GetOrCreateDraftCharacter(ctx, userID, realmID)
		require.NoError(t, err)
		assert.NotEqual(t, char.ID, newChar.ID, "should be a new character since draft was deleted")
	})
}
