# D&D Discord Bot - Development Game Plan 🎲

## Overview
This document outlines the development roadmap for creating a fully-featured D&D 5e Discord bot that enables groups to play tabletop RPG sessions directly through Discord.

## ✅ Completed Features

### Character Management
- **Character Creation** (`/dnd character create`)
  - Interactive flow with race, class, ability scores, proficiencies, and equipment selection
  - Full D&D 5e API integration
  - Draft system for incomplete characters
  
- **Character Commands**
  - `/dnd character list` - View all your characters (active, draft, archived)
  - `/dnd character show <id>` - Detailed character sheet with stats
  - Character management buttons (archive, restore, delete)

### Technical Infrastructure
- Redis persistence layer for character storage
- Docker setup for deployment (optimized for Raspberry Pi)
- Clean architecture with bounded contexts
- Comprehensive test coverage

## 🚧 In Progress

### Session Management (Current Focus)
The foundation for multiplayer D&D gameplay.

#### Commands to Implement:
```
/dnd session create "Campaign Name" - DM creates a new session
/dnd session invite @player - DM invites players
/dnd session join <code> - Players join with invite code
/dnd session start - Begin the session (locks character selection)
/dnd session end - End and archive the session
/dnd session list - Show your active sessions
/dnd session info - Display current session details
```

#### Session Features:
- DM role assignment
- Character selection per session
- Invite system with codes or direct mentions
- Session state management (planning → active → ended)
- Party composition display

## 📋 Upcoming Features

### 1. Initiative Tracker
Combat management system for DMs.
```
/dnd initiative start - Begin combat encounter
/dnd initiative add <name> <roll> - Add combatant
/dnd initiative roll - Auto-roll for party
/dnd initiative next - Move to next turn
/dnd initiative remove <name> - Remove from combat
/dnd initiative end - Clear initiative order
```

### 2. Dice Rolling System
Core dice mechanics with context awareness.
```
/roll 1d20+5 - Basic dice rolling
/roll 2d6+3 fire damage - Typed damage rolls
/dnd attack - Use equipped weapon
/dnd check <ability> - Ability checks
/dnd save <ability> - Saving throws
```

### 3. Character State Management
Quick updates during gameplay.
```
/dnd hp current - Show current HP
/dnd hp set <value> - Set HP directly
/dnd hp heal <amount/dice> - Heal damage
/dnd hp damage <amount> - Take damage
/dnd hp temp <amount> - Add temporary HP
```

### 4. Rest System
Short and long rest mechanics.
```
/dnd rest short - Short rest (roll hit dice)
/dnd rest long - Long rest (full recovery)
```

### 5. Spell Management
Spell slots and casting.
```
/dnd spells list - Show known/prepared spells
/dnd cast <spell> <level> - Cast a spell
/dnd slots - Show remaining spell slots
```

### 6. Inventory Management
Equipment and item tracking.
```
/dnd inventory add <item> - Add item
/dnd inventory remove <item> - Remove item
/dnd equip <item> - Equip weapon/armor
/dnd unequip <slot> - Unequip item
```

## 🎯 Long-term Goals

### Advanced Features
- **Conditions/Effects Tracking** - Poisoned, frightened, blessed, etc.
- **NPC Generator** - Quick NPC creation for DMs
- **Encounter Builder** - CR-based encounter generation
- **Campaign Notes** - Shared notes system
- **Character Sheets Export** - PDF generation
- **Rule References** - Quick rule lookups
- **Homebrew Support** - Custom races, classes, items

### Quality of Life
- **Auto Character Backup** - Version history
- **Character Templates** - Quick character creation
- **Macro System** - Custom command shortcuts
- **Multi-language Support** - Internationalization
- **Voice Integration** - Audio cues for turns

## 🏗️ Architecture Decisions

### Session Storage
- Sessions stored in Redis with expiration
- Session-character associations
- Real-time state updates

### Permission Model
- DM permissions for session management
- Player permissions for own characters
- Spectator mode for observers

### Scalability
- Redis pub/sub for real-time updates
- Stateless bot design
- Horizontal scaling ready

## 📝 Development Priorities

1. **Session Management** (Current)
2. **Initiative Tracker** (Next)
3. **Dice Rolling System**
4. **Character State Management**
5. **Rest System**

Each feature should be:
- Fully tested
- Documented
- Intuitive to use
- Integrated with existing systems

## 🎮 User Experience Goals

- **For Players**: Quick access to character actions, easy state management
- **For DMs**: Powerful tools that don't slow down gameplay
- **For Everyone**: Intuitive commands that feel natural

## 🚀 Deployment Strategy

- Development on local Docker
- Production on Raspberry Pi
- Redis persistence with backups
- Zero-downtime deployments
- Monitoring and logging

---

This is a living document and will be updated as development progresses!