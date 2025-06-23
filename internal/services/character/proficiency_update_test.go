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
		char := &entities.Character{
			ID:      "test-char-1",
			OwnerID: "test-user",
			RealmID: "test-realm",
			Status:  entities.CharacterStatusDraft,
			Proficiencies: map[entities.ProficiencyType][]*entities.Proficiency{
				// Base proficiencies from class
				entities.ProficiencyTypeArmor: {
					{Key: "armor-light", Name: "Light Armor", Type: entities.ProficiencyTypeArmor},
				},
				entities.ProficiencyTypeWeapon: {
					{Key: "weapon-simple", Name: "Simple Weapons", Type: entities.ProficiencyTypeWeapon},
				},
				entities.ProficiencyTypeSavingThrow: {
					{Key: "saving-throw-dex", Name: "Dexterity", Type: entities.ProficiencyTypeSavingThrow},
					{Key: "saving-throw-int", Name: "Intelligence", Type: entities.ProficiencyTypeSavingThrow},
				},
				// Previously chosen skills
				entities.ProficiencyTypeSkill: {
					{Key: "skill-acrobatics", Name: "Acrobatics", Type: entities.ProficiencyTypeSkill},
					{Key: "skill-stealth", Name: "Stealth", Type: entities.ProficiencyTypeSkill},
				},
			},
		}

		// Mock repository calls
		mockRepo.EXPECT().Get(ctx, "test-char-1").Return(char, nil)
		mockRepo.EXPECT().Update(ctx, gomock.Any()).Return(nil)

		// Mock proficiency lookups for new selections
		mockDNDClient.EXPECT().GetProficiency("skill-perception").Return(
			&entities.Proficiency{Key: "skill-perception", Name: "Perception", Type: entities.ProficiencyTypeSkill}, nil)
		mockDNDClient.EXPECT().GetProficiency("skill-investigation").Return(
			&entities.Proficiency{Key: "skill-investigation", Name: "Investigation", Type: entities.ProficiencyTypeSkill}, nil)

		// Update with new skill selections
		updates := &character.UpdateDraftInput{
			Proficiencies: []string{"skill-perception", "skill-investigation"},
		}

		updatedChar, err := service.UpdateDraftCharacter(ctx, "test-char-1", updates)
		assert.NoError(t, err)
		assert.NotNil(t, updatedChar)

		// Verify base proficiencies are preserved
		assert.Len(t, updatedChar.Proficiencies[entities.ProficiencyTypeArmor], 1)
		assert.Len(t, updatedChar.Proficiencies[entities.ProficiencyTypeWeapon], 1)
		assert.Len(t, updatedChar.Proficiencies[entities.ProficiencyTypeSavingThrow], 2)

		// Verify skill proficiencies were replaced (not accumulated)
		assert.Len(t, updatedChar.Proficiencies[entities.ProficiencyTypeSkill], 2)

		// Check the actual skills
		skillKeys := make([]string, 0)
		for _, prof := range updatedChar.Proficiencies[entities.ProficiencyTypeSkill] {
			skillKeys = append(skillKeys, prof.Key)
		}
		assert.Contains(t, skillKeys, "skill-perception")
		assert.Contains(t, skillKeys, "skill-investigation")
		assert.NotContains(t, skillKeys, "skill-acrobatics") // Old skill should be gone
		assert.NotContains(t, skillKeys, "skill-stealth")    // Old skill should be gone
	})
}
