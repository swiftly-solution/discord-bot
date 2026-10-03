import { ChannelType, Client, TextChannel } from "discord.js";
import { db } from "../../../core/db/drizzle";
import { channelRoles } from "../../../core/db/schema/schema";
import { eq } from "drizzle-orm";

export const apChannels: string[] = [];

export default async (client: Client) => {
    const autopublishChannels = await db
        .select()
        .from(channelRoles)
        .where(eq(channelRoles.role, "autopublish"));

    for (const autopublishChannel of autopublishChannels) {
        try {
            const channel = await client.channels.fetch(
                autopublishChannel.channelId,
            );

            if (!channel) continue;
            if (channel.type != ChannelType.GuildAnnouncement) continue;

            apChannels.push(autopublishChannel.channelId);
        } catch (err) {}
    }
};
