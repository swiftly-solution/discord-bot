import {
    AnySelectMenuInteraction,
    Awaitable,
    ButtonInteraction,
    Client,
    Collection,
    CommandInteraction,
    ModalSubmitInteraction,
} from "discord.js";
import { existsSync, readdirSync } from "node:fs";
import commands from "../../core/commands";

export const docsCategories = [
    { name: "Schema", value: "schema" },
    { name: "Entities", value: "entity" },
    { name: "Game Events", value: "gameevent" },
    { name: "NetMessages", value: "protobuf" },
    { name: "ConVars & ConCommands", value: "convar" },
];

const buttons = new Collection<
    string,
    (interaction: ButtonInteraction) => Awaitable<void>
>();
const localCommands = new Collection<
    string,
    (interaction: CommandInteraction) => Awaitable<void>
>();
const modals = new Collection<
    string,
    (interaction: ModalSubmitInteraction) => Awaitable<void>
>();
const anymenus = new Collection<
    string,
    (interaction: AnySelectMenuInteraction) => Awaitable<void>
>();

export default async (client: Client, module_name: string) => {
    if (existsSync(`./src/modules/${module_name}/listeners`)) {
        for (const listener of readdirSync(
            `./src/modules/${module_name}/listeners`,
        )) {
            const {
                type,
                task,
            }: { type: string; task: (...args: any[]) => Awaitable<void> } =
                await import(`./listeners/${listener}`);
            if (type == "once") client.once(listener.split(".")[0], task);
            else client.on(listener.split(".")[0], task);
        }
    }

    if (existsSync(`./src/modules/${module_name}/commands`)) {
        for (const name of readdirSync(
            `./src/modules/${module_name}/commands`,
        )) {
            const { data, command } = await import(`./commands/${name}`);
            commands.push({ data, command });
            localCommands.set(name.split(".")[0], command);
        }
    }

    if (existsSync(`./src/modules/${module_name}/buttons`)) {
        for (const name of readdirSync(
            `./src/modules/${module_name}/buttons`,
        )) {
            buttons.set(
                name.split(".")[0],
                (await import(`./buttons/${name}`)).default,
            );
        }
    }

    if (existsSync(`./src/modules/${module_name}/modals`)) {
        for (const name of readdirSync(`./src/modules/${module_name}/modals`)) {
            const { command } = await import(`./modals/${name}`);
            modals.set(name.split(".")[0], command);
        }
    }

    if (existsSync(`./src/modules/${module_name}/menus`)) {
        for (const name of readdirSync(`./src/modules/${module_name}/menus`)) {
            anymenus.set(
                name.split(".")[0],
                (await import(`./menus/${name}`)).default,
            );
        }
    }
};

export { buttons, localCommands, modals, anymenus };
