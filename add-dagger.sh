#!/bin/bash

# Check if character ID is provided
if [ -z "$1" ]; then
    echo "Usage: ./add-dagger.sh CHARACTER_ID"
    exit 1
fi

CHAR_ID=$1
CHAR_KEY="character:$CHAR_ID"

# Get current character data
CHAR_DATA=$(redis-cli GET "$CHAR_KEY")

if [ -z "$CHAR_DATA" ]; then
    echo "Character not found with ID: $CHAR_ID"
    exit 1
fi

# Create a dagger JSON object
DAGGER='{
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
}'

# Add dagger to weapon inventory
UPDATED_DATA=$(echo "$CHAR_DATA" | jq --argjson dagger "$DAGGER" '.Inventory.weapon += [$dagger]')

# Save back to Redis
echo "$UPDATED_DATA" | redis-cli -x SET "$CHAR_KEY"

echo "Dagger added to character inventory!"