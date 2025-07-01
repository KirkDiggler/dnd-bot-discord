package character_test

import (
	"context"
	character2 "github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/rulebook/dnd5e"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/shared"
	"testing"

	mockdnd5e "github.com/KirkDiggler/dnd-bot-discord/internal/clients/dnd5e/mock"
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
		char := &character2.Character{
			ID:      "test-rogue",
			OwnerID: "test-user",
			RealmID: "test-realm",
			Status:  shared.CharacterStatusDraft,
			Race: &rulebook.Race{
				Key:  "human",
				Name: "Human",
			},
			Class: &rulebook.Class{
				Key:  "rogue",
				Name: "Rogue",
			},
			Proficiencies: map[rulebook.ProficiencyType][]*rulebook.Proficiency{
				// Base proficiencies from rogue class
				rulebook.ProficiencyTypeArmor: {
					{Key: "armor-light", Name: "Light Armor", Type: rulebook.ProficiencyTypeArmor},
				},
				rulebook.ProficiencyTypeWeapon: {
					{Key: "weapon-simple", Name: "Simple Weapons", Type: rulebook.ProficiencyTypeWeapon},
					{Key: "weapon-shortsword", Name: "Shortsword", Type: rulebook.ProficiencyTypeWeapon},
				},
				rulebook.ProficiencyTypeSavingThrow: {
					{Key: "saving-throw-dex", Name: "Dexterity", Type: rulebook.ProficiencyTypeSavingThrow},
					{Key: "saving-throw-int", Name: "Intelligence", Type: rulebook.ProficiencyTypeSavingThrow},
				},
			},
		}

		// First skill selection
		mockRepo.EXPECT().Get(ctx, "test-rogue").Return(char, nil)
		mockRepo.EXPECT().Update(ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, c *character2.Character) error {
			// Verify skills were set correctly
			assert.Len(t, c.Proficiencies[rulebook.ProficiencyTypeSkill], 4)
			return nil
		})

		// Mock proficiency lookups
		mockDNDClient.EXPECT().GetProficiency("skill-acrobatics").Return(
			&rulebook.Proficiency{Key: "skill-acrobatics", Name: "Acrobatics", Type: rulebook.ProficiencyTypeSkill}, nil)
		mockDNDClient.EXPECT().GetProficiency("skill-stealth").Return(
			&rulebook.Proficiency{Key: "skill-stealth", Name: "Stealth", Type: rulebook.ProficiencyTypeSkill}, nil)
		mockDNDClient.EXPECT().GetProficiency("skill-perception").Return(
			&rulebook.Proficiency{Key: "skill-perception", Name: "Perception", Type: rulebook.ProficiencyTypeSkill}, nil)
		mockDNDClient.EXPECT().GetProficiency("skill-investigation").Return(
			&rulebook.Proficiency{Key: "skill-investigation", Name: "Investigation", Type: rulebook.ProficiencyTypeSkill}, nil)

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
		assert.Len(t, updatedChar.Proficiencies[rulebook.ProficiencyTypeSkill], 4)

		// Second skill selection - simulating user changing their mind
		mockRepo.EXPECT().Get(ctx, "test-rogue").Return(updatedChar, nil)
		mockRepo.EXPECT().Update(ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, c *character2.Character) error {
			// Verify skills were replaced, not accumulated
			assert.Len(t, c.Proficiencies[rulebook.ProficiencyTypeSkill], 4, "Skills should be replaced, not accumulated")

			// Verify base proficiencies are preserved
			assert.Len(t, c.Proficiencies[rulebook.ProficiencyTypeArmor], 1)
			assert.Len(t, c.Proficiencies[rulebook.ProficiencyTypeWeapon], 2)
			assert.Len(t, c.Proficiencies[rulebook.ProficiencyTypeSavingThrow], 2)
			return nil
		})

		// Mock new proficiency lookups
		mockDNDClient.EXPECT().GetProficiency("skill-sleight-of-hand").Return(
			&rulebook.Proficiency{Key: "skill-sleight-of-hand", Name: "Sleight of Hand", Type: rulebook.ProficiencyTypeSkill}, nil)
		mockDNDClient.EXPECT().GetProficiency("skill-deception").Return(
			&rulebook.Proficiency{Key: "skill-deception", Name: "Deception", Type: rulebook.ProficiencyTypeSkill}, nil)
		mockDNDClient.EXPECT().GetProficiency("skill-insight").Return(
			&rulebook.Proficiency{Key: "skill-insight", Name: "Insight", Type: rulebook.ProficiencyTypeSkill}, nil)
		mockDNDClient.EXPECT().GetProficiency("skill-intimidation").Return(
			&rulebook.Proficiency{Key: "skill-intimidation", Name: "Intimidation", Type: rulebook.ProficiencyTypeSkill}, nil)

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
		assert.Len(t, finalChar.Proficiencies[rulebook.ProficiencyTypeSkill], 4, "Should have exactly 4 skills")

		// Check the actual skills
		skillKeys := make(map[string]bool)
		for _, prof := range finalChar.Proficiencies[rulebook.ProficiencyTypeSkill] {
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
