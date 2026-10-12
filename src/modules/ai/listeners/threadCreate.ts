import { ThreadChannel } from "discord.js";
import { respond } from "../functions/chat";
import { aiChannels } from "../functions/initChannels";

export const type = "on";
export const task = async (thread: ThreadChannel, newlyCreated: boolean) => {
    if (
        !newlyCreated ||
        !thread.parentId ||
        !aiChannels.includes(thread.parentId)
    )
        return;

    let starter = null;
    for (let i = 0; i < 5 && !starter; i++) {
        starter = await thread.fetchStarterMessage().catch(() => null);
        if (!starter) await new Promise((r) => setTimeout(r, 1000));
    }
    if (!starter || starter.author.bot) return;

    if (thread.joinable) await thread.join().catch(() => {});
    await respond(thread);
};
