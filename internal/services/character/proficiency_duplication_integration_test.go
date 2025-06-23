package character_test

import (
	"context"
	"testing"

	mockdnd5e "github.com/KirkDiggler/dnd-bot-discord/internal/clients/dnd5e/mock"
	"github.com/KirkDiggler/dnd-bot-discord/internal/entities"
	mockcharrepo "github.com/KirkDiggler/dnd-bot-discord/internal/repositories/characters/mock"
	"github.com/KirkDiggler/dnd-bot-discord/internal/services/character"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

// TestProficiencyDuplicationBugIntegration tests the full flow that causes issue #73
func TestProficiencyDuplicationBugIntegration(t *testing.T) {
	ctx := context.Background()

	t.Run("Multiple updates with same skills should not accumulate", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockDNDClient := mockdnd5e.NewMockClient(ctrl)
		mockRepo := mockcharrepo.NewMockRepository(ctrl)

		service := character.NewService(&character.ServiceConfig{
			DNDClient:  mockDNDClient,
			Repository: mockRepo,
		})

		// Create a rogue character with base proficiencies
		char := &entities.Character{
			ID:      "test-rogue",
			OwnerID: "test-user",
			RealmID: "test-realm",
			Status:  entities.CharacterStatusDraft,
			Race: &entities.Race{
				Key:  "human",
				Name: "Human",
			},
			Class: &entities.Class{
				Key:  "rogue",
				Name: "Rogue",
			},
			Proficiencies: map[entities.ProficiencyType][]*entities.Proficiency{
				// Base proficiencies from rogue class
				entities.ProficiencyTypeArmor: {
					{Key: "armor-light", Name: "Light Armor", Type: entities.ProficiencyTypeArmor},
				},
				entities.ProficiencyTypeWeapon: {
					{Key: "weapon-simple", Name: "Simple Weapons", Type: entities.ProficiencyTypeWeapon},
					{Key: "weapon-shortsword", Name: "Shortsword", Type: entities.ProficiencyTypeWeapon},
				},
				entities.ProficiencyTypeSavingThrow: {
					{Key: "saving-throw-dex", Name: "Dexterity", Type: entities.ProficiencyTypeSavingThrow},
					{Key: "saving-throw-int", Name: "Intelligence", Type: entities.ProficiencyTypeSavingThrow},
				},
			},
		}

		// First skill selection
		mockRepo.EXPECT().Get(ctx, "test-rogue").Return(char, nil)
		mockRepo.EXPECT().Update(ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, c *entities.Character) error {
			// Verify skills were set correctly
			assert.Len(t, c.Proficiencies[entities.ProficiencyTypeSkill], 4)
			return nil
		})

		// Mock proficiency lookups
		mockDNDClient.EXPECT().GetProficiency("skill-acrobatics").Return(
			&entities.Proficiency{Key: "skill-acrobatics", Name: "Acrobatics", Type: entities.ProficiencyTypeSkill}, nil)
		mockDNDClient.EXPECT().GetProficiency("skill-stealth").Return(
			&entities.Proficiency{Key: "skill-stealth", Name: "Stealth", Type: entities.ProficiencyTypeSkill}, nil)
		mockDNDClient.EXPECT().GetProficiency("skill-perception").Return(
			&entities.Proficiency{Key: "skill-perception", Name: "Perception", Type: entities.ProficiencyTypeSkill}, nil)
		mockDNDClient.EXPECT().GetProficiency("skill-investigation").Return(
			&entities.Proficiency{Key: "skill-investigation", Name: "Investigation", Type: entities.ProficiencyTypeSkill}, nil)

		// First update - selecting 4 skills
		updatedChar, err := service.UpdateDraftCharacter(ctx, "test-rogue", &character.UpdateDraftInput{
			Proficiencies: []string{
				"skill-acrobatics",
				"skill-stealth",
				"skill-perception",
				"skill-investigation",
			},
		})
		assert.NoError(t, err)
		assert.Len(t, updatedChar.Proficiencies[entities.ProficiencyTypeSkill], 4)

		// Second skill selection - simulating user changing their mind
		mockRepo.EXPECT().Get(ctx, "test-rogue").Return(updatedChar, nil)
		mockRepo.EXPECT().Update(ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, c *entities.Character) error {
			// Verify skills were replaced, not accumulated
			assert.Len(t, c.Proficiencies[entities.ProficiencyTypeSkill], 4, "Skills should be replaced, not accumulated")

			// Verify base proficiencies are preserved
			assert.Len(t, c.Proficiencies[entities.ProficiencyTypeArmor], 1)
			assert.Len(t, c.Proficiencies[entities.ProficiencyTypeWeapon], 2)
			assert.Len(t, c.Proficiencies[entities.ProficiencyTypeSavingThrow], 2)
			return nil
		})

		// Mock new proficiency lookups
		mockDNDClient.EXPECT().GetProficiency("skill-sleight-of-hand").Return(
			&entities.Proficiency{Key: "skill-sleight-of-hand", Name: "Sleight of Hand", Type: entities.ProficiencyTypeSkill}, nil)
		mockDNDClient.EXPECT().GetProficiency("skill-deception").Return(
			&entities.Proficiency{Key: "skill-deception", Name: "Deception", Type: entities.ProficiencyTypeSkill}, nil)
		mockDNDClient.EXPECT().GetProficiency("skill-insight").Return(
			&entities.Proficiency{Key: "skill-insight", Name: "Insight", Type: entities.ProficiencyTypeSkill}, nil)
		mockDNDClient.EXPECT().GetProficiency("skill-intimidation").Return(
			&entities.Proficiency{Key: "skill-intimidation", Name: "Intimidation", Type: entities.ProficiencyTypeSkill}, nil)

		// Second update - changing to different skills
		finalChar, err := service.UpdateDraftCharacter(ctx, "test-rogue", &character.UpdateDraftInput{
			Proficiencies: []string{
				"skill-sleight-of-hand",
				"skill-deception",
				"skill-insight",
				"skill-intimidation",
			},
		})
		assert.NoError(t, err)

		// Final verification
		assert.Len(t, finalChar.Proficiencies[entities.ProficiencyTypeSkill], 4, "Should have exactly 4 skills")

		// Check the actual skills
		skillKeys := make(map[string]bool)
		for _, prof := range finalChar.Proficiencies[entities.ProficiencyTypeSkill] {
			skillKeys[prof.Key] = true
		}
		assert.True(t, skillKeys["skill-sleight-of-hand"])
		assert.True(t, skillKeys["skill-deception"])
		assert.True(t, skillKeys["skill-insight"])
		assert.True(t, skillKeys["skill-intimidation"])

		// Old skills should not be present
		assert.False(t, skillKeys["skill-acrobatics"])
		assert.False(t, skillKeys["skill-stealth"])
		assert.False(t, skillKeys["skill-perception"])
		assert.False(t, skillKeys["skill-investigation"])
	})
}
