import { Awaitable, Client } from "discord.js";
import { existsSync, readdirSync } from "node:fs";
import initChannels from "./functions/initChannels";

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

    await initChannels();
    setInterval(
        () => initChannels().catch((err) => console.error(err)),
        1000 * 60 * 2,
    );
};
