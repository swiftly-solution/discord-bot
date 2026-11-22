# SwiftlyS2 Discord Bot

A Go Discord bot ([discordgo](https://github.com/bwmarrin/discordgo)) for moderating SwiftlyS2.

## Prerequisites

- Go (see `go.mod` for the version)
- A PostgreSQL database
- A Discord application with a bot token

The bot needs permissions to manage messages, moderate members, manage channel
permissions, and send/pin messages. Enable the privileged **Message Content
Intent** for the bot in the Discord Developer Portal.

## Installation

1. Create `.env` from the example and fill in the values:

    ```sh
    cp .env.example .env
    ```

    | Variable              | Description                                                    |
    | --------------------- | -------------------------------------------------------------- |
    | `DISCORD_TOKEN`       | Bot token.                                                     |
    | `DATABASE_URL`        | `postgresql://username:password@host:5432/database_name`       |
    | `CF_BYPASS_KEY`       | Cloudflare bypass key for swiftlys2.net (docs and forum).      |
    | `DEEPSEEK_AI_API_KEY` | DeepSeek API key for the AI forum.                             |
    | `PERSONALITY_PATH`    | Markdown file with the AI personality.                         |
    | `GIT_TOKEN`           | GitHub token for the AI's GitHub tools (optional).             |

2. Create the database schema (skip this for a database the TypeScript bot already used):

    ```sh
    psql "$DATABASE_URL" -f schema.sql
    ```

## Usage

```sh
go run .
```

or build a binary with `go build -o discord-bot .`. Prebuilt binaries are
attached to every [release](../../releases).

After connecting, the bot registers its slash commands in every guild it can
access and sets its activity to `Watching https://swiftlys2.net`.

## Layout

| Path             | Description                                                        |
| ---------------- | ------------------------------------------------------------------ |
| `main.go`        | Configuration, database, Discord session and module startup.       |
| `bot/`           | Slash command registration and dispatch, shared helpers.           |
| `store/`         | PostgreSQL access to `channel_roles`.                              |
| `mcpconn/`       | Reconnecting MCP client (SwiftlyS2 docs, GitHub).                  |
| `administrative/`| `/channel`: honeypot, autopublish, AI forum, lock and unlock.      |
| `ai/`            | Swiftly-chan, the DeepSeek assistant for AI forums.                |
| `autopublish/`   | Publishes messages in announcement channels.                       |
| `docs/`          | `/lookup` and `/search` over the SwiftlyS2 docs.                   |
| `donate/`        | `/donate`.                                                         |
| `honeypot/`      | Times out spam bots that write in honeypot channels.               |
| `tests/`         | Tests, one folder per package.                                     |

## Tests

```sh
go vet ./...
go test ./...
```

## License

SwiftlyS2 Discord Bot is licensed under the [GNU General Public License v3.0 or
later](LICENSE).
