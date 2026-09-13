import { ChatInputCommandInteraction, SlashCommandBuilder } from "discord.js";

export const data = new SlashCommandBuilder()
    .setName("donate")
    .setDescription("See informations about donating to Swiftly Group.");

export const command = async (interaction: ChatInputCommandInteraction) => {
    await interaction.deferReply();

    await interaction.editReply({
        embeds: [
            {
                title: "Donate to Swiftly Group",
                description: `First we want to thank you for considering donating to Swiftly Group! Your support helps us continue to develop and maintain our projects, including SwiftlyS2 and other open-source initiatives such as CS2Browser. You'll receive <@&1420834529370832978> role upon request.`,
                color: 0x00feed,
            },
            {
                title: "Running Costs",
                description: `Running servers and maintaining our infrastructure comes with the following costs:`,
                color: 0x00feed,
                fields: [
                    {
                        name: "Domain Registration (cs2browser.net)",
                        value: "€2 / month",
                        inline: true,
                    },
                    {
                        name: "Domain Registration (swiftlys2.net)",
                        value: "€2 / month",
                        inline: true,
                    },
                    {
                        name: "Cloudflare Pro Plan (cs2browser.net)",
                        value: "€25 / month",
                        inline: true,
                    },
                    {
                        name: "Server Hosting (CS2Browser & SwiftlyS2)",
                        value: "€30 / month",
                        inline: true,
                    },
                ],
            },
            {
                title: "Donation Methods",
                description: `There are several ways you can support our projects:\n\n- [GitHub Sponsors](http://github.com/sponsors/swiftly-solution) - Credit Card\n- [Ko-fi](http://ko-fi.com/swiftlygroup) - Credit Card / PayPal\n- [Steam Skins](https://steamcommunity.com/tradeoffer/new/?partner=1139088750&token=9defPx3R)\n\n**Crypto**`,
                color: 0x00feed,
                fields: [
                    {
                        name: "BTC (Bitcoin network)",
                        value: "`3KJd3rndVkMpEiSa42oBRehTMBKCtwgDAb`",
                    },
                    {
                        name: "ETH (Ethereum network)",
                        value: "`0x5f13c44488359ae0DB72379003c19A97c7aD36C5`",
                    },
                    {
                        name: "USDC (Ethereum network)",
                        value: "`0x05FC1E124E3a7d3FA16c55dF08B8cCaBf30b4363`",
                    },
                ],
            },
        ],
    });
};
