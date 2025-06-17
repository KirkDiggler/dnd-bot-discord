# D&D Discord Bot 🎲 - A Clean Architecture Example in Go

A fully-featured Discord bot for playing Dungeons & Dragons 5th Edition online. Create characters, manage sessions, track combat, and roll dice - all within Discord! Built with Go using clean architecture principles, this project serves as an excellent example of building production-ready Discord bots with proper testing, error handling, and external API integration.

[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org)
[![Discord.js](https://img.shields.io/badge/DiscordGo-Latest-7289da.svg)](https://github.com/bwmarrin/discordgo)
[![D&D 5e API](https://img.shields.io/badge/D&D%205e-API-red.svg)](https://www.dnd5eapi.co/)

## 🌟 Features

### ✅ Implemented
- **Complete Character Creation Wizard**: Multi-step Discord interaction flow
- **D&D 5e API Integration**: Real-time data from the official D&D 5e API  
- **Smart Ability Assignment**: Auto-assign abilities based on class optimization
- **Complex Choice Resolution**: Handles nested equipment and proficiency choices
- **Redis Persistence**: Full character, session, and encounter storage
- **Character Management**: List, view, archive, and delete characters
- **Session/Party System**: Create, join, and manage game sessions
- **Combat Encounters**: Add monsters, roll initiative, track turns
- **Dungeon Mode**: Cooperative play with bot as DM
- **Class Features**: Proper AC calculation (Monk unarmored defense, etc)
- **Help System**: Built-in help command with all available commands
- **Docker Deployment**: Ready for Raspberry Pi deployment
- **Comprehensive Test Coverage**: Unit and integration tests
- **Clean Architecture**: Separation of concerns with interfaces

### 🚧 In Development
- Bot-controlled monster turns in combat
- Dungeon room mechanics (puzzles, traps, treasure)
- Spell system integration
- Character leveling system

### 📋 Planned Features
- Advanced dice rolling expressions
- Character conditions and status effects
- Inventory management system
- Campaign persistence and world state
- DM tools for custom content
- See [GAMEPLAN.md](GAMEPLAN.md) for the complete development roadmap

## 🚀 Quick Start

### Prerequisites
- Go 1.21+
- Docker & Docker Compose
- Discord Bot Token

### Setup

1. **Clone the repository**
   ```bash
   git clone https://github.com/yourusername/dnd-bot-discord.git
   cd dnd-bot-discord
   ```

2. **Create Discord Application**
   - Visit [Discord Developer Portal](https://discord.com/developers/applications)
   - Create new application and bot
   - Enable "Message Content Intent" under Privileged Gateway Intents
   - Copy bot token and application ID

3. **Configure Environment**
   ```bash
   cp .env.example .env
   # Edit .env with your bot token and application ID
   # Add Redis URL: REDIS_URL=redis://localhost:6379
   ```

4. **Install Dependencies**
   ```bash
   make deps
   ```

5. **Run the Bot**
   ```bash
   # Start Redis (required for persistence)
   docker compose up -d redis
   
   # Run the bot
   make run
   # or
   go run cmd/bot/main.go
   ```

6. **Invite Bot to Server**
   - Go to OAuth2 > URL Generator in Discord Developer Portal
   - Select scopes: `bot`, `applications.commands`
   - Select permissions: `Send Messages`, `Use Slash Commands`, `Embed Links`, `Read Message History`
   - Use generated URL to invite bot

## 📖 Usage

### Available Commands

#### Character Management
```
/dnd character create      # Start character creation wizard
/dnd character list        # View all your characters
/dnd character show <id>   # Display detailed character sheet
/dnd character delete <id> # Delete a character
```

#### Session Management
```
/dnd session create <name> # Create a new game session
/dnd session list          # View active sessions
/dnd session join <code>   # Join a session with invite code
/dnd session info          # View current session details
/dnd session start         # Start the game session (DM only)
/dnd session end           # End the session (DM only)
```

#### Combat & Encounters
```
/dnd encounter add <monster> # Add a monster to encounter (DM only)
/dnd test combat [monster]   # Quick test combat with bot as DM
```

#### Dungeon Mode (Cooperative Play)
```
/dnd dungeon [difficulty]    # Start a dungeon delve (easy/medium/hard)
# Bot acts as DM, all players cooperate
# Features room exploration, combat encounters, and treasure
```

#### Help
```
/dnd help                    # Show all available commands
```

#### Character Actions (via buttons)
- **Archive**: Move character to archived status
- **Restore**: Restore archived character to active
- **Delete**: Permanently delete character (with confirmation)

### Character Creation Flow
1. **Race Selection**: Choose from dropdown of all D&D 5e races
2. **Class Selection**: Pick class with hit die and proficiency info
3. **Ability Scores**: Roll 4d6 drop lowest, with auto-assign option
4. **Proficiencies**: Select skills based on class and race options
5. **Equipment**: Choose starting equipment with smart nested selections
6. **Finalize**: Name your character and save

## 🏗️ Architecture

This project demonstrates clean architecture principles in Go:

```
internal/
├── handlers/          # Discord interaction handlers
│   └── discord/       # Discord-specific implementations
├── services/          # Business logic layer
│   └── character/     # Character management service
├── repositories/      # Data persistence layer
│   └── character/     # Character repository interface
├── entities/          # Domain models
│   ├── character.go   # Core character entity
│   └── ...           # Other domain entities
└── clients/          # External service clients
    └── dnd5e/        # D&D 5e API client
```

### Key Design Patterns
- **Repository Pattern**: Abstract data storage behind interfaces
- **Service Layer**: Business logic separated from infrastructure
- **Dependency Injection**: Constructor-based DI for testability
- **Interface Segregation**: Small, focused interfaces
- **Error Handling**: Consistent error types and handling

## 🧪 Testing

### Run Tests
```bash
# All tests
make test

# With coverage
make test-coverage

# Specific package
go test ./internal/services/character -v

# Specific test suite
go test ./internal/services/character -run "TestEquipmentChoiceResolverSuite"
```

### Test Structure
- **Unit Tests**: Mock external dependencies
- **Table-Driven Tests**: Comprehensive test cases
- **Test Suites**: Organized by feature
- **Edge Cases**: Extensive error condition testing

## 🔧 Development

### Code Generation
```bash
# Generate mocks
make generate-mocks

# Format code
go fmt ./...

# Lint
golangci-lint run
```

### Project Structure
```
.
├── cmd/bot/           # Application entrypoint
├── internal/          # Private application code
├── docs/              # Documentation
├── proto/             # Protocol buffer definitions (future)
├── docker-compose.yml # Local development services
├── Makefile          # Common development tasks
└── go.mod            # Go module definition
```

### Adding New Features
1. Define interfaces in appropriate package
2. Implement with tests
3. Wire up in service provider
4. Add Discord handler if needed

## 📊 Project Status

See [PROGRESS.md](PROGRESS.md) for detailed progress tracking and [GAMEPLAN.md](GAMEPLAN.md) for the development roadmap.

### Recent Achievements
- ✅ Complete character creation flow with all D&D 5e content
- ✅ Redis persistence with full test coverage
- ✅ Character management commands (list, show, delete)
- ✅ AC calculation with proper armor stacking
- ✅ Docker deployment setup for Raspberry Pi
- ✅ 80%+ test coverage on core services

### Currently Working On
- 🚧 Session management system
- 🚧 Party formation and invites
- 🚧 Initiative tracker

## 🐳 Deployment

### Docker Deployment (Recommended)
```bash
# Build and run with Docker Compose
docker compose up -d

# View logs
docker compose logs -f bot

# Stop services
docker compose down
```

### Raspberry Pi Deployment
The project is optimized for Raspberry Pi deployment:
- Memory-efficient Redis configuration (256MB limit)
- ARM-compatible Docker images
- Resource-conscious design

See `docker-compose.yml` for production configuration.

## 🤝 Contributing

This project is an excellent example for learning:
- Discord bot development in Go
- Clean architecture principles
- Test-driven development
- External API integration
- Complex domain modeling
- Redis integration patterns

Feel free to:
- Report bugs
- Suggest features
- Submit pull requests
- Use as reference for your own projects

## 📄 License

MIT License - feel free to use this code as reference or starting point for your own projects.

## 🙏 Acknowledgments

- [D&D 5e API](http://www.dnd5eapi.co/) for providing comprehensive D&D data
- [discordgo](https://github.com/bwmarrin/discordgo) for the excellent Discord library
- The D&D community for inspiration

---

**Note**: This bot is not affiliated with Wizards of the Coast or D&D Beyond. It uses the open-source D&D 5e API for game data.