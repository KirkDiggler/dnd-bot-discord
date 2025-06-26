# Combat Display Enhancement Demo

## New Initiative Field Display

The new display separates players and monsters into distinct embed fields:

### 🛡️ Party Members
▶ `18` 💚 **Grunk**
├─ 🟢🟢🟢🟢🟢🟢🟢🟢🟡⬜ HP: 45/50 | AC: 16
└─ *Barbarian*

  `10` 💚 **Thorin**
├─ 🟢🟢🟢🟢🟢🟢🟢🟢🟢🟢 HP: 40/40 | AC: 18
└─ *Fighter*

### ⚔️ Enemies
  `15` 🧡 **Goblin**
├─ 🔴🔴🔴⬜⬜⬜⬜⬜⬜⬜ HP: 3/12 | AC: 13


## Features Demonstrated

1. **Visual HP Indicators**:
   - 💚 = Healthy (>75% HP)
   - 💛 = Good (50-75% HP)
   - 🧡 = Hurt (25-50% HP)
   - ❤️ = Critical (<25% HP)
   - 💀 = Dead (0 HP)

2. **HP Bars**: Visual representation using emoji squares showing exact health percentage

3. **Turn Indicator**: ▶ shows whose turn it is currently

4. **Name Truncation**: Long names are truncated to 15 characters with "..."

5. **Class Display**: Player characters show their class in italics below their stats

## Alternative Compact Display

Also implemented a CSS-styled compact view:

```css
[Initiative Order]
─────────────────────────────────────────
▶ [18] 👤 Grunk        HP[▰▰▰▰▰▰▰▱] AC:16
  [15] 👹 Goblin       HP[▰▰▱▱▱▱▱▱] AC:13
  [10] 👤 Thorin       HP[▰▰▰▰▰▰▰▰] AC:18
```

## Detailed Combatant View

For individual combatant details:

**💚 Grunk**
📊 **Stats**
**HP:** 45/50
**AC:** 16
**Initiative:** 18

⚔️ **Class**
Barbarian

💚 **Health**
🟩🟩🟩🟩🟩🟩🟩🟩🟩⬜