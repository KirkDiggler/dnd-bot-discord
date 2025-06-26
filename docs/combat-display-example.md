# Combat Display Example

## Combat Status Embed

**⚔️ Combat - Round 3**

⚔️ **Round 3** | 🛡️ Players: 2 | 🐉 Monsters: 2
🎯 **Grunk's turn**

### 🎯 Initiative Order
```
Init │ Name              │ HP          │ AC
─────┼───────────────────┼─────────────┼────
▶ 18  │ 🪓 Grunk          │ ▰▰▰▰▰▰▰▱  45/50 │ 16
  15  │ 🐉 Goblin         │ ▰▰▱▱▱▱▱▱   3/12 │ 13 ❗
  12  │ 🐉 Orc            │ ▰▰▰▰▰▰▱▱  15/20 │ 14
  10  │ ⚔️ Thorin         │ ▰▰▰▰▰▰▰▰  40/40 │ 18
```

### 📜 Recent Actions
• ⚔️ **Grunk** → **Goblin** | d20:19+7=26 vs AC:13 | HIT 🩸 **12** [10]+2
• ⚔️ **Goblin** → **Grunk** | d20:8+4=12 vs AC:16 | ❌ MISS
• ⚔️ **Thorin** → **Orc** | d20:15+5=20 vs AC:14 | HIT 🩸 **8** [6]+2
• ⚔️ **Orc** → **Thorin** | d20:12+4=16 vs AC:18 | ❌ MISS
• ⚔️ **Grunk** → **Goblin** | d20:**20**+7=27 vs AC:13 | 💥 CRIT! 🩸 **24** [10,8]+4

## What Each Part Means

### Attack Roll Format
`⚔️ **Attacker** → **Target** | d20:roll+bonus=total vs AC:armor | result`

- **d20:19+7=26**: Shows the d20 roll (19), attack bonus (+7), and total (26)
- **vs AC:13**: Shows what AC the attack is trying to beat
- **HIT/MISS**: Clear result indicator
- **🩸 12 [10]+2**: Total damage (12) with dice rolls [10] and bonus (+2)

### Critical Hits
- **d20:20**: Natural 20 shown in bold
- **💥 CRIT!**: Clear critical indicator
- Damage shows doubled dice: [10,8]+4

### Initiative Table Features
- **▶**: Current turn indicator
- **Class Icons**: 🪓 Barbarian, ⚔️ Fighter, 🧙 Wizard, etc.
- **HP Bars**: Visual health representation with 8 segments
- **HP Numbers**: Current/Max for exact values
- **❗**: Low health warning (<25% HP)
- **💀**: Defeated indicator (0 HP)

## Benefits
1. **Dice Transparency**: Everyone can see all rolls immediately
2. **Quick Status Check**: Table format shows all combatant status at a glance
3. **Combat Flow**: Recent actions show the flow of battle
4. **Visual Indicators**: Icons and bars make status instantly recognizable
5. **Compact Display**: All information fits in a single embed