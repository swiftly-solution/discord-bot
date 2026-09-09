# SwiftlyS2 Discord Bot

A TypeScript Discord bot for moderating SwiftlyS2.

## Prerequisites

- A current Node.js LTS release and [pnpm](https://pnpm.io/)
- A PostgreSQL database
- A Discord application with a bot token

The bot needs permissions to manage messages, moderate members, manage channel
permissions, and send/pin messages. Enable the privileged **Message Content
Intent** for the bot in the Discord Developer Portal.

## Installation

1. Install dependencies:

    ```sh
    pnpm install
    ```

2. Create `.env` from the example and configure the required values:

    ```sh
    cp .env.example .env
    ```

    ```dotenv
    DISCORD_TOKEN=your_discord_bot_token
    DATABASE_URL=postgresql://username:password@host:5432/database_name
    ```

3. Create the database schema:

    ```sh
    pnpm db:push
    ```

## Usage

Start the bot:

```sh
pnpm start
```

For development, use:

```sh
pnpm dev
```

After connecting, the bot registers its slash commands in every guild it can
access and sets its activity to `Watching https://swiftlys2.net`.

## Scripts

| Script             | Description                               |
| ------------------ | ----------------------------------------- |
| `pnpm start`       | Runs the bot with `tsx`.                  |
| `pnpm build`       | Type-checks and compiles TypeScript.      |
| `pnpm db:generate` | Generates Drizzle migrations.             |
| `pnpm db:push`     | Generates and applies Drizzle migrations. |

## License

SwiftlyS2 Discord Bot is licensed under the [GNU General Public License v3.0 or
later](LICENSE).
