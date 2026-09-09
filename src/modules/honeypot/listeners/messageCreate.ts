import {
    ActionRow,
    ActionRowBuilder,
    ButtonBuilder,
    ButtonComponent,
    Message,
    TextChannel,
} from "discord.js";
import { hpChannelMessage, hpChannels } from "../functions/initChannels";

export const type = "on";
export const task = async (message: Message) => {
    if (hpChannels.includes(message.channelId)) {
        if (message.author.bot) return;
        if (!message.channel.isTextBased()) return;
        const textChannel = message.channel as TextChannel;

        await message.delete();
        await message.member?.timeout(1000 * 60 * 60 * 24 * 7, "Honeypot");

        const honeypotMessageId = hpChannelMessage.get(message.channelId);
        if (!honeypotMessageId) return;
        const honeypotMessage =
            await textChannel.messages.fetch(honeypotMessageId);
        if (!honeypotMessage) return;

        const actionRow = honeypotMessage
            .components[0] as ActionRow<ButtonComponent>;
        const button = actionRow.components[0] as ButtonComponent;
        const currentTimeouts = Number.parseInt(
            button.label!.split(": ")[1],
            10,
        );

        await honeypotMessage.edit({
            components: [
                new ActionRowBuilder<ButtonBuilder>().addComponents(
                    ButtonBuilder.from(button).setLabel(
                        `Timeouts: ${currentTimeouts + 1}`,
                    ),
                ),
            ],
        });
    }
};
