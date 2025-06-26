package combat

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRoundSummary(t *testing.T) {
	t.Run("RecordAttack tracks damage correctly", func(t *testing.T) {
		rs := NewRoundSummary(1)

		// Record a hit
		rs.RecordAttack(AttackInfo{
			AttackerName: "Grunk",
			TargetName:   "Goblin",
			Damage:       8,
			Hit:          true,
			WeaponName:   "Longsword",
		})

		// Record a miss
		rs.RecordAttack(AttackInfo{
			AttackerName: "Goblin",
			TargetName:   "Grunk",
			Hit:          false,
			WeaponName:   "Shortsword",
		})

		// Record another hit against player
		rs.RecordAttack(AttackInfo{
			AttackerName: "Orc",
			TargetName:   "Grunk",
			Damage:       6,
			Hit:          true,
			WeaponName:   "Greataxe",
		})

		// Check player stats
		grunkInfo := rs.PlayerActions["Grunk"]
		assert.NotNil(t, grunkInfo)
		assert.Equal(t, 8, grunkInfo.DamageDealt)
		assert.Equal(t, 6, grunkInfo.DamageTaken)
		assert.Len(t, grunkInfo.AttacksMade, 1)
		assert.Len(t, grunkInfo.AttacksRecvd, 2)
	})

	t.Run("GetPlayerSummary formats correctly", func(t *testing.T) {
		rs := NewRoundSummary(2)

		// Record some attacks
		rs.RecordAttack(AttackInfo{
			AttackerName: "Thorin",
			TargetName:   "Goblin",
			Damage:       10,
			Hit:          true,
			Critical:     true,
			WeaponName:   "Warhammer",
		})

		rs.RecordAttack(AttackInfo{
			AttackerName: "Goblin",
			TargetName:   "Thorin",
			Damage:       4,
			Hit:          true,
			WeaponName:   "Scimitar",
		})

		summary := rs.GetPlayerSummary("Thorin")
		assert.Contains(t, summary, "Your Attacks:")
		assert.Contains(t, summary, "💥 Warhammer → **Goblin** 🩸 **10** damage")
		assert.Contains(t, summary, "Attacks Against You:")
		assert.Contains(t, summary, "🩸 **Goblin** → You **4** damage")
		assert.Contains(t, summary, "Damage Taken: **4**")
		assert.Contains(t, summary, "Damage Dealt: **10**")
	})

	t.Run("GetRoundOverview shows all players", func(t *testing.T) {
		rs := NewRoundSummary(3)

		// Multiple players attacking
		rs.RecordAttack(AttackInfo{
			AttackerName: "Grunk",
			TargetName:   "Dragon",
			Damage:       12,
			Hit:          true,
			WeaponName:   "Greatsword",
		})

		rs.RecordAttack(AttackInfo{
			AttackerName: "Thorin",
			TargetName:   "Dragon",
			Damage:       8,
			Hit:          true,
			WeaponName:   "Warhammer",
		})

		rs.RecordAttack(AttackInfo{
			AttackerName: "Dragon",
			TargetName:   "Grunk",
			Damage:       15,
			Hit:          true,
			WeaponName:   "Bite",
		})

		overview := rs.GetRoundOverview()
		assert.Contains(t, overview, "Round 3 Overview:")
		assert.Contains(t, overview, "**Grunk**: dealt **12** | took **15**")
		assert.Contains(t, overview, "**Thorin**: dealt **8**")
	})
}
