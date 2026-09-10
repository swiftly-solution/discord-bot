import {
    ChatInputCommandInteraction,
    InteractionContextType,
    SlashCommandBuilder,
} from "discord.js";
import { mcpClient, mcpTransport } from "../../../core/docs";
import { docsCategories } from "..";
import { formatDate } from "../../../core/time";

export const data = new SlashCommandBuilder()
    .setName("lookup")
    .setDescription("Look up exact information through SwiftlyS2 Docs")
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
            .setDescription("The name to lookup")
            .setRequired(true),
    );

export const command = async (interaction: ChatInputCommandInteraction) => {
    const category = interaction.options.getString("category", true);
    const name = interaction.options.getString("name", true);
    const mcpCommand = `${category}_lookup`;

    await interaction.deferReply();

    let args: Record<string, string> = {};
    if (category == "entity") args = { className: name };
    else args = { name };

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
                            title: `Lookup Result - ${prettyNameCategory}`,
                            description: `\`${name}\` could not be found.`,
                        },
                    ],
                });
                return;
            }

            const responseJSON = JSON.parse(typedResult.content[0].text);

            let description = ``;
            description += `[SwiftlyS2 Documentation](${responseJSON.url})\n\n`;

            if (category == "schema") {
                if (responseJSON.kind == "class") {
                    description += `\`\`\`yaml\n`;
                    description += `${name}:\n`;
                    description += `  project: ${responseJSON.project ?? "server"}\n`;
                    description += `  kind: ${responseJSON.kind ?? "class"}\n`;
                    description += `  base_classes: ${responseJSON.entry.base_classes_count ?? "None"}\n`;
                    description += `  fields: ${responseJSON.entry.fields_count ?? "None"}\n`;
                    description += `  size: ${responseJSON.entry.size ?? "0"}\n`;
                    description += `  is_struct: ${responseJSON.entry.is_struct ?? "false"}\n`;
                    description += `  has_chainer: ${responseJSON.entry.has_chainer ?? "false"}\n`;
                    description += `\`\`\``;
                    if (responseJSON.entry.base_classes_count > 0) {
                        description += `\n**Base Classes**\n`;
                        description += responseJSON.entry.base_classes
                            .map((v: string) => `\`${v}\``)
                            .join(",");
                    }
                } else {
                    description += `\`\`\`yaml\n`;
                    description += `${name}:\n`;
                    description += `  project: ${responseJSON.project ?? "server"}\n`;
                    description += `  kind: ${responseJSON.kind ?? "enum"}\n`;
                    description += `  fields: ${responseJSON.entry.fields_count ?? "None"}\n`;
                    description += `  size: ${responseJSON.entry.size ?? "0"}\n`;
                    description += `\`\`\``;
                }
            } else if (category == "entity") {
                description += `\`\`\`yaml\n`;
                description += `${name}:\n`;
                description += `  designer_name: ${responseJSON.entityClass.designer_name ?? "None"}\n`;
                description += `  datamap:\n`;
                description += `    data_class_name: ${responseJSON.datamap.data_class_name}\n`;
                description += `    fields:\n`;
                description += `      inputs: ${(responseJSON.datamap.fields.inputs ?? []).length}\n`;
                description += `      outputs: ${(responseJSON.datamap.fields.outputs ?? []).length}\n`;
                description += `      members: ${(responseJSON.datamap.fields.members ?? []).length}\n`;
                description += `    think_functions: ${(responseJSON.datamap.think_functions ?? []).length}\n`;
                description += `\`\`\``;
                if (responseJSON.entityClass.flags.length > 0) {
                    description += `\n**Flags**\n`;
                    description += responseJSON.entityClass.flags
                        .map((v: string) => `\`${v}\``)
                        .join(",");
                }
            } else if (category == "gameevent") {
                description += `\`\`\`yaml\n`;
                if (responseJSON.entry.comment != "")
                    description += `# ${responseJSON.entry.comment ?? "No comment"}\n`;
                description += `${name}:\n`;
                description += `  raw_name: ${responseJSON.entry.name ?? "None"}\n`;
                description += `  files:\n`;
                for (const file of responseJSON.entry.files) {
                    description += `    - ${file}\n`;
                }
                description += `  fields: ${responseJSON.entry.fields.length}\n`;
                description += `\`\`\``;
            } else if (category == "protobuf") {
                if (responseJSON.kind == "message") {
                    description += `\`\`\`yaml\n`;
                    description += `${name}:\n`;
                    description += `  file: ${responseJSON.file ?? "unknown"}\n`;
                    description += `  kind: ${responseJSON.kind ?? "message"}\n`;
                    description += `  fields: ${responseJSON.entry.fields.length}\n`;
                    description += `\`\`\``;
                } else {
                    description += `\`\`\`yaml\n`;
                    description += `${name}:\n`;
                    description += `  file: ${responseJSON.file ?? "unknown"}\n`;
                    description += `  kind: ${responseJSON.kind ?? "enum"}\n`;
                    description += `  values: ${responseJSON.entry.values.length}\n`;
                    description += `\`\`\``;
                }
            } else if (category == "convar") {
                if (responseJSON.kind == "convar") {
                    description += `\`\`\`yaml\n`;
                    description += `# ${responseJSON.entry.description ?? "No description"}\n`;
                    description += `${name}:\n`;
                    description += `  module: ${responseJSON.module ?? "unknown"}\n`;
                    description += `  kind: ${responseJSON.kind ?? "convar"}\n`;
                    description += `  attributes:\n`;
                    description += `    has_callback: ${responseJSON.entry.attributes.has_callback ?? "false"}\n`;
                    description += `    has_min: ${responseJSON.entry.attributes.has_min ?? "false"}\n`;
                    description += `    has_max: ${responseJSON.entry.attributes.has_max ?? "false"}\n`;
                    description += `    has_default: ${responseJSON.entry.attributes.has_default ?? "false"}\n`;
                    description += `  default: ${responseJSON.entry.default ?? "none"}\n`;
                    description += `\`\`\``;
                    if (responseJSON.entry.flags.length > 0) {
                        description += `\n**Flags**\n`;
                        description += responseJSON.entry.flags
                            .map((v: string) => `\`${v}\``)
                            .join(",");
                    }
                } else {
                    description += `\`\`\`yaml\n`;
                    description += `# ${responseJSON.entry.description ?? "No description"}\n`;
                    description += `${name}:\n`;
                    description += `  module: ${responseJSON.module ?? "unknown"}\n`;
                    description += `  kind: ${responseJSON.kind ?? "concommandd"}\n`;
                    description += `  attributes:\n`;
                    description += `    has_callback: ${responseJSON.entry.attributes.has_callback ?? "false"}\n`;
                    description += `    has_completion_callback: ${responseJSON.entry.attributes.has_completion_callback ?? "false"}\n`;
                    description += `\`\`\``;
                    if (responseJSON.entry.flags.length > 0) {
                        description += `\n**Flags**\n`;
                        description += responseJSON.entry.flags
                            .map((v: string) => `\`${v}\``)
                            .join(",");
                    }
                }
            }

            await interaction.editReply({
                embeds: [
                    {
                        color: 0x00feed,
                        title: `Lookup Result - ${prettyNameCategory}`,
                        description,
                        footer: {
                            text: `SwiftlyS2 Docs - requested by ${interaction.user.tag} | ${formatDate(new Date())}`,
                        },
                    },
                ],
            });

            return;
        } catch (err) {
            await mcpClient.close();
            await mcpClient.connect(mcpTransport);
        }
    }
};
