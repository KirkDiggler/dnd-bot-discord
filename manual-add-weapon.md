# Manual Redis Commands to Add a Shortsword

1. First, find your character:
```bash
redis-cli KEYS "character:*"
```

2. Get your character data and save to a file:
```bash
redis-cli GET "character:YOUR_CHARACTER_ID" > char.json
```

3. Open char.json in a text editor and find the "Inventory" section

4. Add this to the "weapon" array in inventory:
```json
{
  "Type": "weapon",
  "Data": {
    "base": {
      "key": "shortsword",
      "name": "Shortsword",
      "equipment_category": "weapon",
      "weight": 2,
      "cost": {
        "quantity": 10,
        "unit": "gp"
      }
    },
    "damage": {
      "damage_dice": "1d6",
      "damage_type": "piercing",
      "dice_count": 1,
      "dice_size": 6
    },
    "range": 5,
    "weapon_category": "Martial",
    "weapon_range": "Melee",
    "category_range": "martial-melee",
    "properties": [
      {
        "key": "finesse",
        "name": "Finesse",
        "url": "/api/weapon-properties/finesse"
      },
      {
        "key": "light",
        "name": "Light",
        "url": "/api/weapon-properties/light"
      },
      {
        "key": "monk",
        "name": "Monk",
        "url": "/api/weapon-properties/monk"
      }
    ]
  }
}
```

5. Save the file and load it back to Redis:
```bash
redis-cli SET "character:YOUR_CHARACTER_ID" "$(cat char.json)"
```

## Adding a Dagger (for off-hand)

Use the same process but with this JSON:
```json
{
  "Type": "weapon",
  "Data": {
    "base": {
      "key": "dagger",
      "name": "Dagger",
      "equipment_category": "weapon",
      "weight": 1,
      "cost": {
        "quantity": 2,
        "unit": "gp"
      }
    },
    "damage": {
      "damage_dice": "1d4",
      "damage_type": "piercing",
      "dice_count": 1,
      "dice_size": 4
    },
    "range": 20,
    "weapon_category": "Simple",
    "weapon_range": "Melee",
    "category_range": "simple-melee",
    "properties": [
      {
        "key": "finesse",
        "name": "Finesse",
        "url": "/api/weapon-properties/finesse"
      },
      {
        "key": "light",
        "name": "Light",
        "url": "/api/weapon-properties/light"
      },
      {
        "key": "thrown",
        "name": "Thrown",
        "url": "/api/weapon-properties/thrown"
      },
      {
        "key": "monk",
        "name": "Monk",
        "url": "/api/weapon-properties/monk"
      }
    ]
  }
}
```