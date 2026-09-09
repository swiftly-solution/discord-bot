import {
    ActionRowBuilder,
    ButtonBuilder,
    ButtonStyle,
    Client,
    TextChannel,
} from "discord.js";
import { db } from "../../../core/db/drizzle";
import { channelRoles } from "../../../core/db/schema/schema";
import { eq } from "drizzle-orm";

export const hpChannels: string[] = [];
export const hpChannelMessage: Map<string, string> = new Map();

export default async (client: Client) => {
    const honeypotChannels = await db
        .select()
        .from(channelRoles)
        .where(eq(channelRoles.role, "honeypot"));

    for (const honeypotChannel of honeypotChannels) {
        try {
            const channel = await client.channels.fetch(
                honeypotChannel.channelId,
            );

            const textChannel = channel?.isTextBased()
                ? (channel as TextChannel)
                : null;

            if (!textChannel) continue;
            hpChannels.push(honeypotChannel.channelId);

            let found = false;
            const messages = await textChannel.messages.fetchPins({
                limit: 50,
            });
            for (const message of messages.items) {
                if (message.message.author.id != client.user?.id) continue;
                if (message.message.embeds.length == 0) continue;

                if (
                    message.message.embeds[0].title ==
                    "DO NOT SEND MESSAGES IN THIS CHANNEL"
                ) {
                    found = true;
                    hpChannelMessage.set(
                        honeypotChannel.channelId,
                        message.message.id,
                    );
                    break;
                }
            }

            if (found == true) continue;

            const embed = await textChannel.send({
                embeds: [
                    {
                        title: "DO NOT SEND MESSAGES IN THIS CHANNEL",
                        description:
                            "This channel is used to catch spam bots. Any messages sent here will result in a **timeout for 1 week.**",
                    },
                ],
                components: [
                    new ActionRowBuilder<ButtonBuilder>().addComponents(
                        new ButtonBuilder()
                            .setCustomId("honeypot-timeouts")
                            .setEmoji("🎂")
                            .setLabel("Timeouts: 0")
                            .setStyle(ButtonStyle.Secondary)
                            .setDisabled(true),
                    ),
                ],
            });
            await embed.pin();
        } catch (err) {}
    }
};
