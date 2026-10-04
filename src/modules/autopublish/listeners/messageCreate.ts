import {
    ActionRow,
    ActionRowBuilder,
    ButtonBuilder,
    ButtonComponent,
    Message,
    TextChannel,
} from "discord.js";
import { apChannels } from "../functions/initChannels";

export const type = "on";
export const task = async (message: Message) => {
    if (apChannels.includes(message.channelId)) {
        if (!message.channel.isTextBased()) return;

        if (message.crosspostable) await message.crosspost();
    }
};
