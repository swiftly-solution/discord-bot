import {
    ChannelType,
    ChatInputCommandInteraction,
    InteractionContextType,
    PermissionsBitField,
    SlashCommandBuilder,
    TextChannel,
} from "discord.js";
import { db } from "../../../core/db/drizzle";
import { channelRoles } from "../../../core/db/schema/schema";
import { and, eq } from "drizzle-orm";

export const data = new SlashCommandBuilder()
    .setName("channel")
    .setDescription("Channel commands")
    .setContexts(InteractionContextType.Guild)
    .addChannelOption((option) =>
        option
            .setName("channel")
            .setDescription("The channel to use")
            .setRequired(true)
            .addChannelTypes(
                ChannelType.GuildText,
                ChannelType.GuildAnnouncement,
                ChannelType.GuildForum,
            ),
    )
    .addStringOption((option) =>
        option
            .setChoices(
                { name: "Set Honeypot", value: "set-honeypot" },
                { name: "Remove Honeypot", value: "remove-honeypot" },
                { name: "Set Autopublish", value: "set-autopublish" },
                { name: "Remove Autopublish", value: "remove-autopublish" },
                { name: "Set AI Forum", value: "set-forum" },
                { name: "Remove AI Forum", value: "remove-forum" },
                { name: "Lock", value: "lock" },
                { name: "Unlock", value: "unlock" },
            )
            .setName("action")
            .setDescription("The action to perform")
            .setRequired(true),
    )
    .setDefaultMemberPermissions(PermissionsBitField.Flags.Administrator);

export const command = async (interaction: ChatInputCommandInteraction) => {
    const channel = interaction.options.getChannel("channel", true);
    const textChannel = channel as TextChannel;
    const action = interaction.options.getString("action", true);

    await interaction.deferReply();

    if (action == "set-honeypot") {
        const channelId = textChannel.id;
        const guildId = textChannel.guildId;

        const channelsFound = await db
            .select()
            .from(channelRoles)
            .where(
                and(
                    eq(channelRoles.channelId, channelId),
                    eq(channelRoles.guildId, guildId),
                    eq(channelRoles.role, "honeypot"),
                ),
            );

        if (channelsFound.length > 0) {
            await interaction.editReply({
                embeds: [
                    {
                        title: "Honeypot Already Set",
                        description: `The channel ${textChannel} is already set as a honeypot.`,
                        color: 0xff0000,
                    },
                ],
            });
            return;
        }

        await db.insert(channelRoles).values({
            channelId: channelId,
            guildId: guildId,
            createdAt: new Date(),
            role: "honeypot",
            id: crypto.randomUUID(),
        });

        await interaction.editReply({
            embeds: [
                {
                    title: "Honeypot Set",
                    description: `The channel ${textChannel} has been set as a honeypot.`,
                    color: 0x00feed,
                },
            ],
        });
    } else if (action == "remove-honeypot") {
        const channelId = textChannel.id;
        const guildId = textChannel.guildId;

        const channelsFound = await db
            .select()
            .from(channelRoles)
            .where(
                and(
                    eq(channelRoles.channelId, channelId),
                    eq(channelRoles.guildId, guildId),
                    eq(channelRoles.role, "honeypot"),
                ),
            );

        if (channelsFound.length == 0) {
            await interaction.editReply({
                embeds: [
                    {
                        title: "Honeypot Not Found",
                        description: `The channel ${textChannel} is not set as a honeypot.`,
                        color: 0xff0000,
                    },
                ],
            });
            return;
        }

        await db
            .delete(channelRoles)
            .where(
                and(
                    eq(channelRoles.channelId, channelId),
                    eq(channelRoles.guildId, guildId),
                    eq(channelRoles.role, "honeypot"),
                ),
            );

        await interaction.editReply({
            embeds: [
                {
                    title: "Honeypot Removed",
                    description: `The channel ${textChannel} has been removed as a honeypot.`,
                    color: 0x00feed,
                },
            ],
        });
    } else if (action == "set-autopublish") {
        const channelId = textChannel.id;
        const guildId = textChannel.guildId;

        const channelsFound = await db
            .select()
            .from(channelRoles)
            .where(
                and(
                    eq(channelRoles.channelId, channelId),
                    eq(channelRoles.guildId, guildId),
                    eq(channelRoles.role, "autopublish"),
                ),
            );

        if (channelsFound.length > 0) {
            await interaction.editReply({
                embeds: [
                    {
                        title: "Autopublish Already Set",
                        description: `The channel ${textChannel} is already set as an autopublish.`,
                        color: 0xff0000,
                    },
                ],
            });
            return;
        }

        await db.insert(channelRoles).values({
            channelId: channelId,
            guildId: guildId,
            createdAt: new Date(),
            role: "autopublish",
            id: crypto.randomUUID(),
        });

        await interaction.editReply({
            embeds: [
                {
                    title: "Autopublish Set",
                    description: `The channel ${textChannel} has been set as an autopublish.`,
                    color: 0x00feed,
                },
            ],
        });
    } else if (action == "remove-autopublish") {
        const channelId = textChannel.id;
        const guildId = textChannel.guildId;

        const channelsFound = await db
            .select()
            .from(channelRoles)
            .where(
                and(
                    eq(channelRoles.channelId, channelId),
                    eq(channelRoles.guildId, guildId),
                    eq(channelRoles.role, "autopublish"),
                ),
            );

        if (channelsFound.length == 0) {
            await interaction.editReply({
                embeds: [
                    {
                        title: "Autopublish Not Found",
                        description: `The channel ${textChannel} is not set as an autopublish.`,
                        color: 0xff0000,
                    },
                ],
            });
            return;
        }

        await db
            .delete(channelRoles)
            .where(
                and(
                    eq(channelRoles.channelId, channelId),
                    eq(channelRoles.guildId, guildId),
                    eq(channelRoles.role, "autopublish"),
                ),
            );

        await interaction.editReply({
            embeds: [
                {
                    title: "Autopublish Removed",
                    description: `The channel ${textChannel} has been removed as an autopublish.`,
                    color: 0x00feed,
                },
            ],
        });
    } else if (action == "set-forum" || action == "remove-forum") {
        const channelId = textChannel.id;
        const guildId = textChannel.guildId;
        const adding = action == "set-forum";
        const where = and(
            eq(channelRoles.channelId, channelId),
            eq(channelRoles.guildId, guildId),
            eq(channelRoles.role, "forum"),
        );

        const channelsFound = await db.select().from(channelRoles).where(where);
        const exists = channelsFound.length > 0;

        if (adding == exists) {
            await interaction.editReply({
                embeds: [
                    {
                        title: adding
                            ? "AI Forum Already Set"
                            : "AI Forum Not Found",
                        description: adding
                            ? `The channel ${textChannel} is already an AI forum.`
                            : `The channel ${textChannel} is not an AI forum.`,
                        color: 0xff0000,
                    },
                ],
            });
            return;
        }

        if (adding) {
            await db.insert(channelRoles).values({
                channelId: channelId,
                guildId: guildId,
                createdAt: new Date(),
                role: "forum",
                id: crypto.randomUUID(),
            });
        } else {
            await db.delete(channelRoles).where(where);
        }

        await interaction.editReply({
            embeds: [
                {
                    title: adding ? "AI Forum Set" : "AI Forum Removed",
                    description: adding
                        ? `The channel ${textChannel} has been set as an AI forum.`
                        : `The channel ${textChannel} has been removed as an AI forum.`,
                    color: 0x00feed,
                },
            ],
        });
    } else if (action == "lock") {
        textChannel.permissionOverwrites.edit(textChannel.guildId, {
            SendMessages: false,
        });
        await interaction.editReply({
            embeds: [
                {
                    title: "Channel Locked",
                    description: `The channel ${textChannel} has been locked.`,
                    color: 0x00feed,
                },
            ],
        });
    } else if (action == "unlock") {
        textChannel.permissionOverwrites.edit(textChannel.guildId, {
            SendMessages: true,
        });
        await interaction.editReply({
            embeds: [
                {
                    title: "Channel Unlocked",
                    description: `The channel ${textChannel} has been unlocked.`,
                    color: 0x00feed,
                },
            ],
        });
    }
};
