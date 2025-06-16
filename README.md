# D&D Discord Bot

A Discord bot for playing D&D 5e, featuring character creation, dice rolling, and combat management.

## Setup

1. Create a Discord application and bot:
   - Go to https://discord.com/developers/applications
   - Create a new application
   - Go to the "Bot" section
   - Create a bot and copy the token
   - Under "Privileged Gateway Intents", enable "Message Content Intent"

2. Copy `.env.example` to `.env` and fill in your values:
   ```bash
   cp .env.example .env
   ```

3. Edit `.env` with your bot token and application ID:
   ```
   DISCORD_TOKEN=your_bot_token_here
   DISCORD_APP_ID=your_application_id_here
   ```

4. Invite the bot to your server:
   - In the Discord Developer Portal, go to OAuth2 > URL Generator
   - Select scopes: `bot`, `applications.commands`
   - Select bot permissions: `Send Messages`, `Use Slash Commands`, `Embed Links`
   - Use the generated URL to invite the bot

5. Run the bot:
   ```bash
   make run
   ```

## Commands

- `/dnd character create` - Start the character creation wizard

## Development

Run tests:
```bash
make test
```

Build binary:
```bash
make build
```