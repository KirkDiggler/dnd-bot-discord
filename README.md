# D&D Discord Bot - A Clean Architecture Example in Go

A fully-featured Discord bot for D&D 5e character management, built with Go using clean architecture principles. This project serves as an excellent example of building production-ready Discord bots with proper testing, error handling, and external API integration.

## 🌟 Features

### Implemented
- **Complete Character Creation Wizard**: Multi-step Discord interaction flow
- **D&D 5e API Integration**: Real-time data from the official D&D 5e API  
- **Smart Ability Assignment**: Auto-assign abilities based on class optimization
- **Complex Choice Resolution**: Handles nested equipment and proficiency choices
- **Comprehensive Test Coverage**: Unit tests for all components
- **Clean Architecture**: Separation of concerns with interfaces

### Coming Soon
- Redis persistence for character storage
- Character management commands (list, view, edit, delete)
- Dice rolling with character modifiers
- Combat and initiative tracking

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
   ```

4. **Install Dependencies**
   ```bash
   make deps
   ```

5. **Run the Bot**
   ```bash
   # With Redis (recommended)
   make run-with-redis

   # Without Redis (in-memory only)
   make run
   ```

6. **Invite Bot to Server**
   - Go to OAuth2 > URL Generator in Discord Developer Portal
   - Select scopes: `bot`, `applications.commands`
   - Select permissions: `Send Messages`, `Use Slash Commands`, `Embed Links`, `Read Message History`
   - Use generated URL to invite bot

## 📖 Usage

### Available Commands

#### Character Creation
```
/dnd character create
```
Starts an interactive character creation wizard with:
- Race selection (all official D&D 5e races)
- Class selection with specialization info
- Ability score generation and assignment
- Skill and proficiency selection
- Starting equipment choices
- Character naming

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

See [PROGRESS.md](PROGRESS.md) for detailed progress tracking and roadmap.

### Highlights
- ✅ Complete character creation flow
- ✅ 80%+ test coverage on core services  
- ✅ Production-ready error handling
- ✅ Clean architecture implementation
- 🚧 Redis persistence layer
- 📋 Character management commands

## 🤝 Contributing

This project is an excellent example for learning:
- Discord bot development in Go
- Clean architecture principles
- Test-driven development
- External API integration
- Complex domain modeling

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