import chalk from "chalk";
import { Client, REST, Routes } from "discord.js";

const commands: any[] = [];

const RegisterCommands = async (botClient: Client) => {
    const rest = new REST().setToken(String(process.env.DISCORD_TOKEN));

    console.log(`[${chalk.green("Commands")}] Loading ${commands.length} application (/) commands.`);

    const commands_data: any[] = [];
    for (const command of commands) commands_data.push(command.data.toJSON());

    const guilds = await botClient.guilds.fetch()
    const application = botClient.application;

    if(application == null) {
        console.error(`[${chalk.red("Commands")}] Failed to fetch application!`);
        return;
    }

    var cmdCount = 0
    for (const guild of guilds) {
        const data = await rest.put(
            Routes.applicationGuildCommands(application.id, guild[0]),
            { body: commands_data }
        ) as any[];
        if (cmdCount == 0) cmdCount = data.length
    }

    console.log(`[${chalk.green("Commands")}] Updated ${cmdCount} application (/) commands.`);
}

export default commands;
export { RegisterCommands };