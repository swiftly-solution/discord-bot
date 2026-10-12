import { Message } from "discord.js";
import { respond } from "../functions/chat";
import { aiChannels } from "../functions/initChannels";

export const type = "on";
export const task = async (message: Message) => {
    if (message.author.bot || !message.channel.isThread()) return;
    const thread = message.channel;
    if (message.id == thread.id) return;
    if (!thread.parentId || !aiChannels.includes(thread.parentId)) return;
    if (
        !message.mentions.has(message.client.user, {
            ignoreEveryone: true,
            ignoreRoles: true,
        })
    )
        return;

    await respond(thread, message);
};
