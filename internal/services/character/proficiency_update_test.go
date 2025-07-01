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

// TestProficiencyUpdateHandling tests the fix for issue #73
func TestProficiencyUpdateHandling(t *testing.T) {
	ctx := context.Background()

	t.Run("Updating proficiencies should replace skill proficiencies only", func(t *testing.T) {
		// Setup
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockDNDClient := mockdnd5e.NewMockClient(ctrl)
		mockRepo := mockcharrepo.NewMockRepository(ctrl)

		service := character.NewService(&character.ServiceConfig{
			DNDClient:  mockDNDClient,
			Repository: mockRepo,
		})

		// Create a character with base proficiencies
		char := &character2.Character{
			ID:      "test-char-1",
			OwnerID: "test-user",
			RealmID: "test-realm",
			Status:  shared.CharacterStatusDraft,
			Proficiencies: map[rulebook.ProficiencyType][]*rulebook.Proficiency{
				// Base proficiencies from class
				rulebook.ProficiencyTypeArmor: {
					{Key: "armor-light", Name: "Light Armor", Type: rulebook.ProficiencyTypeArmor},
				},
				rulebook.ProficiencyTypeWeapon: {
					{Key: "weapon-simple", Name: "Simple Weapons", Type: rulebook.ProficiencyTypeWeapon},
				},
				rulebook.ProficiencyTypeSavingThrow: {
					{Key: "saving-throw-dex", Name: "Dexterity", Type: rulebook.ProficiencyTypeSavingThrow},
					{Key: "saving-throw-int", Name: "Intelligence", Type: rulebook.ProficiencyTypeSavingThrow},
				},
				// Previously chosen skills
				rulebook.ProficiencyTypeSkill: {
					{Key: "skill-acrobatics", Name: "Acrobatics", Type: rulebook.ProficiencyTypeSkill},
					{Key: "skill-stealth", Name: "Stealth", Type: rulebook.ProficiencyTypeSkill},
				},
			},
		}

		// Mock repository calls
		mockRepo.EXPECT().Get(ctx, "test-char-1").Return(char, nil)
		mockRepo.EXPECT().Update(ctx, gomock.Any()).Return(nil)

		// Mock proficiency lookups for new selections
		mockDNDClient.EXPECT().GetProficiency("skill-perception").Return(
			&rulebook.Proficiency{Key: "skill-perception", Name: "Perception", Type: rulebook.ProficiencyTypeSkill}, nil)
		mockDNDClient.EXPECT().GetProficiency("skill-investigation").Return(
			&rulebook.Proficiency{Key: "skill-investigation", Name: "Investigation", Type: rulebook.ProficiencyTypeSkill}, nil)

		// Update with new skill selections
		updates := &character.UpdateDraftInput{
			Proficiencies: []string{"skill-perception", "skill-investigation"},
		}

		updatedChar, err := service.UpdateDraftCharacter(ctx, "test-char-1", updates)
		assert.NoError(t, err)
		assert.NotNil(t, updatedChar)

		// Verify base proficiencies are preserved
		assert.Len(t, updatedChar.Proficiencies[rulebook.ProficiencyTypeArmor], 1)
		assert.Len(t, updatedChar.Proficiencies[rulebook.ProficiencyTypeWeapon], 1)
		assert.Len(t, updatedChar.Proficiencies[rulebook.ProficiencyTypeSavingThrow], 2)

		// Verify skill proficiencies were replaced (not accumulated)
		assert.Len(t, updatedChar.Proficiencies[rulebook.ProficiencyTypeSkill], 2)

		// Check the actual skills
		skillKeys := make([]string, 0)
		for _, prof := range updatedChar.Proficiencies[rulebook.ProficiencyTypeSkill] {
			skillKeys = append(skillKeys, prof.Key)
		}
		assert.Contains(t, skillKeys, "skill-perception")
		assert.Contains(t, skillKeys, "skill-investigation")
		assert.NotContains(t, skillKeys, "skill-acrobatics") // Old skill should be gone
		assert.NotContains(t, skillKeys, "skill-stealth")    // Old skill should be gone
	})
}
