import chalk from "chalk";
import { ActivityType, Client } from "discord.js";
import "dotenv/config";
import { readdirSync } from "node:fs";
import { RegisterCommands } from "./core/commands";
import { mcpClient, mcpTransport } from "./core/docs";

const botClient = new Client({
    intents: [
        "GuildMessages",
        "GuildWebhooks",
        "MessageContent",
        "Guilds",
        "GuildMessageReactions",
        "GuildBans",
        "DirectMessages",
        "GuildInvites",
        "GuildModeration",
    ],
});

setTimeout(async () => {
    try {
        await botClient.login(String(process.env.DISCORD_TOKEN));

        for (const file of readdirSync("./src/modules")) {
            const { default: func } = await import(
                `./modules/${file}/index.ts`
            );
            await func(botClient, file);
            console.log(
                `[${chalk.green(file.toUpperCase())}] Module has been succesfully loaded!`,
            );
        }

        RegisterCommands(botClient);
        botClient.user?.setActivity({
            type: ActivityType.Watching,
            name: "https://swiftlys2.net",
        });

        await mcpClient.connect(mcpTransport);
    } catch (err) {
        console.error(`[${chalk.red("Core")}] Failed to start bot!`);
        console.error(err);
    }
}, 0);
