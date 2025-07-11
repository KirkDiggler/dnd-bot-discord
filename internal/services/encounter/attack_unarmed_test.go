package encounter_test

import (
	"context"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/game/combat"
	session2 "github.com/KirkDiggler/dnd-bot-discord/internal/domain/game/session"
	"testing"
	"time"

	"github.com/KirkDiggler/dnd-bot-discord/internal/dice/mock"
	"github.com/KirkDiggler/dnd-bot-discord/internal/repositories/character_draft"
	"github.com/KirkDiggler/dnd-bot-discord/internal/repositories/characters"
	"github.com/KirkDiggler/dnd-bot-discord/internal/repositories/encounters"
	"github.com/KirkDiggler/dnd-bot-discord/internal/repositories/gamesessions"
	"github.com/KirkDiggler/dnd-bot-discord/internal/services/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/services/encounter"
	"github.com/KirkDiggler/dnd-bot-discord/internal/services/session"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPerformAttack_UnarmedStrike_AllScenarios(t *testing.T) {
	tests := []struct {
		name           string
		setupAttacker  func(ctx context.Context, enc *combat.Encounter, service encounter.Service) (*combat.Combatant, error)
		actionIndex    int
		expectedWeapon string
		expectedDamage int
		expectedBonus  int
	}{
		{
			name: "Monster with no actions uses unarmed strike",
			setupAttacker: func(ctx context.Context, enc *combat.Encounter, service encounter.Service) (*combat.Combatant, error) {
				return service.AddMonster(ctx, enc.ID, "dm-1", &encounter.AddMonsterInput{
					Name:  "Commoner",
					MaxHP: 4,
					AC:    10,
					// No actions defined
				})
			},
			actionIndex:    -1,
			expectedWeapon: "Unarmed Strike",
			expectedDamage: 3, // 1d4 roll result
			expectedBonus:  0,
		},
		{
			name: "Monster with empty actions array uses unarmed strike",
			setupAttacker: func(ctx context.Context, enc *combat.Encounter, service encounter.Service) (*combat.Combatant, error) {
				return service.AddMonster(ctx, enc.ID, "dm-1", &encounter.AddMonsterInput{
					Name:    "Peasant",
					MaxHP:   4,
					AC:      10,
					Actions: []*combat.MonsterAction{}, // Empty actions
				})
			},
			actionIndex:    0,
			expectedWeapon: "Unarmed Strike",
			expectedDamage: 3,
			expectedBonus:  0,
		},
		{
			name: "Monster with invalid action index uses unarmed strike",
			setupAttacker: func(ctx context.Context, enc *combat.Encounter, service encounter.Service) (*combat.Combatant, error) {
				return service.AddMonster(ctx, enc.ID, "dm-1", &encounter.AddMonsterInput{
					Name:  "Guard",
					MaxHP: 11,
					AC:    16,
					Actions: []*combat.MonsterAction{
						{
							Name:        "Spear",
							AttackBonus: 3,
						},
					},
				})
			},
			actionIndex:    99, // Invalid index
			expectedWeapon: "Unarmed Strike",
			expectedDamage: 3,
			expectedBonus:  0,
		},
		{
			name: "Non-combatant type uses unarmed strike",
			setupAttacker: func(ctx context.Context, enc *combat.Encounter, service encounter.Service) (*combat.Combatant, error) {
				// Manually create an NPC combatant
				npc := &combat.Combatant{
					ID:        "npc-1",
					Name:      "Villager",
					Type:      combat.CombatantTypeNPC,
					CurrentHP: 4,
					MaxHP:     4,
					AC:        10,
					IsActive:  true,
				}
				enc.AddCombatant(npc)
				return npc, nil
			},
			actionIndex:    0,
			expectedWeapon: "Unarmed Strike",
			expectedDamage: 3,
			expectedBonus:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			mockDice := mockdice.NewManualMockRoller()

			// Set up deterministic rolls
			mockDice.SetRolls([]int{
				10, // Attack roll
				3,  // Damage roll (1d4)
			})

			// Create services
			charRepo := characters.NewInMemoryRepository()
			draftRepo := character_draft.NewInMemoryRepository()
			charService := character.NewService(&character.ServiceConfig{
				Repository:      charRepo,
				DraftRepository: draftRepo,
			})

			sessionRepo := gamesessions.NewInMemoryRepository()
			sessionService := session.NewService(&session.ServiceConfig{
				Repository:       sessionRepo,
				CharacterService: charService,
			})

			encounterService := encounter.NewService(&encounter.ServiceConfig{
				Repository:       encounters.NewInMemoryRepository(),
				SessionService:   sessionService,
				CharacterService: charService,
				DiceRoller:       mockDice,
			})

			// Create a test session
			sess := &session2.Session{
				ID:        "test-session",
				Name:      "Test Session",
				ChannelID: "channel-1",
				CreatorID: "dm-1",
				DMID:      "dm-1",
				Members: map[string]*session2.SessionMember{
					"dm-1": {UserID: "dm-1", Role: session2.SessionRoleDM},
				},
				Status:     session2.SessionStatusActive,
				CreatedAt:  time.Now(),
				LastActive: time.Now(),
			}
			err := sessionRepo.Create(ctx, sess)
			require.NoError(t, err)

			// Create encounter
			enc, err := encounterService.CreateEncounter(ctx, &encounter.CreateEncounterInput{
				SessionID: "test-session",
				ChannelID: "channel-1",
				Name:      "Test Combat",
				UserID:    "dm-1",
			})
			require.NoError(t, err)

			// Setup attacker
			attacker, err := tt.setupAttacker(ctx, enc, encounterService)
			require.NoError(t, err)

			// Add target
			target, err := encounterService.AddMonster(ctx, enc.ID, "dm-1", &encounter.AddMonsterInput{
				Name:  "Target Dummy",
				MaxHP: 10,
				AC:    10,
			})
			require.NoError(t, err)

			// Get updated encounter and set it up for combat
			enc, err = encounterService.GetEncounter(ctx, enc.ID)
			require.NoError(t, err)
			enc.Status = combat.EncounterStatusActive
			enc.Turn = 0
			enc.TurnOrder = []string{attacker.ID, target.ID}

			// Perform attack
			result, err := encounterService.PerformAttack(ctx, &encounter.AttackInput{
				EncounterID: enc.ID,
				AttackerID:  attacker.ID,
				TargetID:    target.ID,
				UserID:      "dm-1",
				ActionIndex: tt.actionIndex,
			})

			require.NoError(t, err)
			assert.NotNil(t, result)

			// Verify unarmed strike details
			assert.Equal(t, tt.expectedWeapon, result.WeaponName)
			assert.Equal(t, 10, result.AttackRoll)
			assert.Equal(t, tt.expectedBonus, result.AttackBonus)
			assert.Equal(t, 10, result.TotalAttack)
			assert.True(t, result.Hit)
			assert.Equal(t, tt.expectedDamage, result.Damage)
			assert.Equal(t, "bludgeoning", result.DamageType)
		})
	}
}
