import {
    ChatInputCommandInteraction,
    InteractionContextType,
    PermissionFlagsBits,
    SlashCommandBuilder,
} from "discord.js";
import { mcpClient, mcpTransport } from "../../../core/docs";
import { docsCategories } from "..";
import { formatDate } from "../../../core/time";
import { capitalizeFirstLetter } from "../../../core/strings";

export const data = new SlashCommandBuilder()
    .setName("search")
    .setDescription("Search information through SwiftlyS2 Docs")
    .setContexts(InteractionContextType.Guild)
    .addStringOption((option) =>
        option
            .setChoices(...docsCategories)
            .setName("category")
            .setDescription("The category to search in")
            .setRequired(true),
    )
    .addStringOption((option) =>
        option
            .setName("name")
            .setDescription("The name to search")
            .setRequired(true),
    );

export const command = async (interaction: ChatInputCommandInteraction) => {
    const category = interaction.options.getString("category", true);
    const name = interaction.options.getString("name", true);
    const mcpCommand = `${category}_search`;

    await interaction.deferReply();

    let args: Record<string, string> = { q: name };

    const prettyNameCategory =
        docsCategories.find((c) => c.value == category)?.name ?? "Unknown";

    for (let i = 0; i < 3; i++) {
        try {
            const result = await mcpClient.callTool({
                name: mcpCommand,
                arguments: args,
            });

            const typedResult = result as {
                content: { text: string }[];
                isError?: boolean;
            };

            if (typedResult.isError == true) {
                await interaction.editReply({
                    embeds: [
                        {
                            color: 0xff0000,
                            title: `Search Result - ${prettyNameCategory}`,
                            description: `\`${name}\` could not be found.`,
                        },
                    ],
                });
                return;
            }

            const responseJSON = JSON.parse(typedResult.content[0].text);

            let description = ``;
            description += `Search query: \`${name}\`\n\n`;

            if (category == "schema") {
                const items = responseJSON.classes.slice(0, 25);
                const byModules: Record<string, any[]> = items.reduce(
                    (acc: Record<string, any[]>, item: any) => {
                        if (!acc[item.project]) {
                            acc[item.project] = [];
                        }
                        acc[item.project].push(item);
                        return acc;
                    },
                    {},
                );

                for (const [module, items] of Object.entries(byModules)) {
                    description += `**Module - ${module}**\n`;
                    for (const entry of items) {
                        description += `[${entry.name}](${entry.url}) - ${capitalizeFirstLetter(entry.kind)}\n`;
                    }
                }
            } else if (category == "entity") {
                if (responseJSON.classes.length > 0) {
                    description += `**Classes**\n`;
                    for (const entry of responseJSON.classes) {
                        description += `[${entry.name}](${entry.url})\n`;
                    }
                    description += `\n`;
                }
                if (responseJSON.fields.length > 0) {
                    description += `**Fields**\n`;
                    for (const entry of responseJSON.fields) {
                        description += `[${entry.className}::${entry.fieldName}](${entry.url})\n`;
                    }
                }
            } else if (category == "gameevent") {
                description += `**Results**\n`;
                for (const entry of responseJSON) {
                    description += `[${entry.name}](${entry.url}) - ${entry.files.map((v: string) => `\`${v}\``).join(", ")}\n`;
                }
            } else if (category == "protobuf") {
                description += `**Results**\n`;
                for (const entry of responseJSON) {
                    description += `[${entry.name}](${entry.url}) - ${capitalizeFirstLetter(entry.kind)}\n`;
                }
            } else if (category == "convar") {
                const items = responseJSON.items.slice(0, 25);
                const byModules: Record<string, any[]> = items.reduce(
                    (acc: Record<string, any[]>, item: any) => {
                        if (!acc[item.module]) {
                            acc[item.module] = [];
                        }
                        acc[item.module].push(item);
                        return acc;
                    },
                    {},
                );

                for (const [module, items] of Object.entries(byModules)) {
                    description += `**Module - ${module}**\n`;
                    for (const entry of items) {
                        description += `[${entry.name}](${entry.url}) - ${capitalizeFirstLetter(entry.kind)}\n`;
                    }
                }
            } else if (category == "panorama") {
                description += `**Results**\n`;
                for (const entry of responseJSON) {
                    description += `[${entry.name}](${entry.url}) - ${entry.snippet}\n`;
                }
            }

            await interaction.editReply({
                embeds: [
                    {
                        color: 0x00feed,
                        title: `Search Result - ${prettyNameCategory}`,
                        description,
                        footer: {
                            text: `SwiftlyS2 Docs - requested by ${interaction.user.tag} | ${formatDate(new Date())}`,
                        },
                    },
                ],
            });

            return;
        } catch (err) {
            console.log(err);
        }
    }
};
