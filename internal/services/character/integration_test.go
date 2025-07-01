package character_test

import (
	"context"
	character2 "github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/equipment"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/rulebook/dnd5e"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/shared"
	"testing"

	mockdnd5e "github.com/KirkDiggler/dnd-bot-discord/internal/clients/dnd5e/mock"
	mockcharrepo "github.com/KirkDiggler/dnd-bot-discord/internal/repositories/characters/mock"
	"github.com/KirkDiggler/dnd-bot-discord/internal/services/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestCharacterCreationFlow_Integration(t *testing.T) {
	t.Skip("TODO: Fix test to handle automatic class proficiency loading during finalization")
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup mocks
	mockRepo := mockcharrepo.NewMockRepository(ctrl)
	mockClient := mockdnd5e.NewMockClient(ctrl)

	// Create service
	svc := character.NewService(&character.ServiceConfig{
		DNDClient:  mockClient,
		Repository: mockRepo,
	})

	ctx := context.Background()
	userID := "user123"
	realmID := "realm456"

	t.Run("full character creation flow", func(t *testing.T) {
		// Step 1: Get or create draft character
		draftChar := &character2.Character{
			ID:            "draft123",
			OwnerID:       userID,
			RealmID:       realmID,
			Name:          "Draft Character",
			Status:        shared.CharacterStatusDraft,
			Level:         1,
			Attributes:    make(map[shared.Attribute]*character2.AbilityScore),
			Proficiencies: make(map[rulebook.ProficiencyType][]*rulebook.Proficiency),
			Inventory:     make(map[equipment.EquipmentType][]equipment.Equipment),
			EquippedSlots: make(map[shared.Slot]equipment.Equipment),
		}

		mockRepo.EXPECT().
			GetByOwnerAndRealm(ctx, userID, realmID).
			Return([]*character2.Character{}, nil)

		mockRepo.EXPECT().
			Create(ctx, gomock.Any()).
			DoAndReturn(func(ctx context.Context, char *character2.Character) error {
				assert.Equal(t, userID, char.OwnerID)
				assert.Equal(t, realmID, char.RealmID)
				assert.Equal(t, shared.CharacterStatusDraft, char.Status)
				// Copy fields without mutex
				draftChar.ID = char.ID
				draftChar.OwnerID = char.OwnerID
				draftChar.RealmID = char.RealmID
				draftChar.Status = char.Status
				return nil
			})

		draft, err := svc.GetOrCreateDraftCharacter(ctx, userID, realmID)
		require.NoError(t, err)
		assert.NotNil(t, draft)

		// Step 2: Update with race and class
		race := testutils.CreateTestRace("human", "Human")
		class := testutils.CreateTestClass("fighter", "Fighter", 10)

		mockClient.EXPECT().GetRace("human").Return(race, nil)
		mockClient.EXPECT().GetClass("fighter").Return(class, nil)

		// Mock the repository Get and Update calls
		mockRepo.EXPECT().
			Get(ctx, draft.ID).
			Return(draftChar, nil).
			Times(2) // Called twice: once for race/class update, once for ability scores

		mockRepo.EXPECT().
			Update(ctx, gomock.Any()).
			DoAndReturn(func(ctx context.Context, char *character2.Character) error {
				// Verify race and class are set
				assert.NotNil(t, char.Race)
				assert.NotNil(t, char.Class)
				assert.Equal(t, "human", char.Race.Key)
				assert.Equal(t, "fighter", char.Class.Key)
				// Copy fields without mutex
				draftChar.ID = char.ID
				draftChar.OwnerID = char.OwnerID
				draftChar.RealmID = char.RealmID
				draftChar.Status = char.Status
				return nil
			}).
			Times(2) // Called twice: once for race/class, once for abilities

		// Update race and class
		raceKey := "human"
		classKey := "fighter"
		_, err = svc.UpdateDraftCharacter(ctx, draft.ID, &character.UpdateDraftInput{
			RaceKey:  &raceKey,
			ClassKey: &classKey,
		})
		require.NoError(t, err)

		// Step 3: Set ability scores
		abilityScores := map[string]int{
			"STR": 15,
			"DEX": 14,
			"CON": 13,
			"INT": 12,
			"WIS": 10,
			"CHA": 8,
		}

		_, err = svc.UpdateDraftCharacter(ctx, draft.ID, &character.UpdateDraftInput{
			AbilityScores: abilityScores,
		})
		require.NoError(t, err)

		// Step 4: Add proficiencies
		mockRepo.EXPECT().
			Get(ctx, draft.ID).
			Return(draftChar, nil)

		mockClient.EXPECT().
			GetProficiency("skill-athletics").
			Return(&rulebook.Proficiency{
				Key:  "skill-athletics",
				Name: "Athletics",
				Type: rulebook.ProficiencyTypeSkill,
			}, nil)

		mockClient.EXPECT().
			GetProficiency("skill-intimidation").
			Return(&rulebook.Proficiency{
				Key:  "skill-intimidation",
				Name: "Intimidation",
				Type: rulebook.ProficiencyTypeSkill,
			}, nil)

		mockRepo.EXPECT().
			Update(ctx, gomock.Any()).
			DoAndReturn(func(ctx context.Context, char *character2.Character) error {
				// Verify proficiencies are added
				assert.NotEmpty(t, char.Proficiencies)
				// Copy fields without mutex
				draftChar.ID = char.ID
				draftChar.OwnerID = char.OwnerID
				draftChar.RealmID = char.RealmID
				draftChar.Status = char.Status
				return nil
			})

		_, err = svc.UpdateDraftCharacter(ctx, draft.ID, &character.UpdateDraftInput{
			Proficiencies: []string{"skill-athletics", "skill-intimidation"},
		})
		require.NoError(t, err)

		// Step 5: Add equipment
		mockRepo.EXPECT().
			Get(ctx, draft.ID).
			Return(draftChar, nil)

		mockClient.EXPECT().
			GetEquipment("longsword").
			Return(&equipment.Weapon{
				Base: equipment.BasicEquipment{
					Key:  "longsword",
					Name: "Longsword",
				},
				WeaponCategory: "Martial",
				WeaponRange:    "Melee",
			}, nil).
			Times(2) // Once for UpdateDraftCharacter, once for FinalizeDraftCharacter (starting equipment)

		mockClient.EXPECT().
			GetEquipment("chain-mail").
			Return(&equipment.Armor{
				Base: equipment.BasicEquipment{
					Key:  "chain-mail",
					Name: "Chain Mail",
				},
				ArmorCategory: "Heavy",
				ArmorClass: &equipment.ArmorClass{
					Base: 16,
				},
			}, nil)

		mockRepo.EXPECT().
			Update(ctx, gomock.Any()).
			DoAndReturn(func(ctx context.Context, char *character2.Character) error {
				// Verify equipment is added
				assert.NotEmpty(t, char.Inventory)
				// Copy fields without mutex
				draftChar.ID = char.ID
				draftChar.OwnerID = char.OwnerID
				draftChar.RealmID = char.RealmID
				draftChar.Status = char.Status
				return nil
			})

		_, err = svc.UpdateDraftCharacter(ctx, draft.ID, &character.UpdateDraftInput{
			Equipment: []string{"longsword", "chain-mail"},
		})
		require.NoError(t, err)

		// Step 6: Set name and finalize
		charName := "Thorin Ironforge"
		mockRepo.EXPECT().
			Get(ctx, draft.ID).
			Return(draftChar, nil).
			Times(2) // Once for name update, once for finalize

		mockRepo.EXPECT().
			Update(ctx, gomock.Any()).
			DoAndReturn(func(ctx context.Context, char *character2.Character) error {
				if char.Name == charName {
					// Name update
					assert.Equal(t, charName, char.Name)
				} else if char.Status == shared.CharacterStatusActive {
					// Finalize update
					assert.Equal(t, shared.CharacterStatusActive, char.Status)
				}
				// Copy fields without mutex
				draftChar.ID = char.ID
				draftChar.OwnerID = char.OwnerID
				draftChar.RealmID = char.RealmID
				draftChar.Status = char.Status
				return nil
			}).
			Times(2) // Once for name, once for status

		_, err = svc.UpdateDraftCharacter(ctx, draft.ID, &character.UpdateDraftInput{
			Name: &charName,
		})
		require.NoError(t, err)

		// Finalize the character
		finalChar, err := svc.FinalizeDraftCharacter(ctx, draft.ID)
		require.NoError(t, err)
		assert.NotNil(t, finalChar)
		assert.Equal(t, shared.CharacterStatusActive, finalChar.Status)
		assert.Equal(t, charName, finalChar.Name)
		assert.NotNil(t, finalChar.Race)
		assert.NotNil(t, finalChar.Class)
		assert.NotEmpty(t, finalChar.Attributes)
	})
}

func TestCharacterValidation_Integration(t *testing.T) {
	tests := []struct {
		name      string
		character *character2.Character
		wantValid bool
		missing   []string
	}{
		{
			name:      "complete character",
			character: testutils.CreateTestCharacter("char1", "user1", "realm1", "Test Character"),
			wantValid: true,
			missing:   []string{},
		},
		{
			name: "missing race",
			character: func() *character2.Character {
				char := testutils.CreateTestCharacter("char2", "user1", "realm1", "Test Character")
				char.Race = nil
				return char
			}(),
			wantValid: false,
			missing:   []string{"race"},
		},
		{
			name: "missing class",
			character: func() *character2.Character {
				char := testutils.CreateTestCharacter("char3", "user1", "realm1", "Test Character")
				char.Class = nil
				return char
			}(),
			wantValid: false,
			missing:   []string{"class"},
		},
		{
			name: "missing attributes",
			character: func() *character2.Character {
				char := testutils.CreateTestCharacter("char4", "user1", "realm1", "Test Character")
				char.Attributes = make(map[shared.Attribute]*character2.AbilityScore)
				return char
			}(),
			wantValid: false,
			missing:   []string{"ability scores"},
		},
		{
			name: "missing name",
			character: func() *character2.Character {
				char := testutils.CreateTestCharacter("char5", "user1", "realm1", "")
				return char
			}(),
			wantValid: false,
			missing:   []string{"name"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isComplete := tt.character.IsComplete()
			assert.Equal(t, tt.wantValid, isComplete)

			// Verify the character can provide display info even if incomplete
			displayInfo := tt.character.GetDisplayInfo()
			assert.NotEmpty(t, displayInfo)
		})
	}
}
