package donate

import (
	"github.com/bwmarrin/discordgo"

	"github.com/swiftly-solution/discord-bot/bot"
)

var Command = bot.Command{
	Def: &discordgo.ApplicationCommand{
		Name:        "donate",
		Description: "See informations about donating to Swiftly Group.",
	},
	Run: func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
		if err := bot.Defer(s, i); err != nil {
			return err
		}
		return bot.Reply(s, i, embeds...)
	},
}

var embeds = []*discordgo.MessageEmbed{
	{
		Title:       "Donate to Swiftly Group",
		Description: "First we want to thank you for considering donating to Swiftly Group! Your support helps us continue to develop and maintain our projects, including SwiftlyS2 and other open-source initiatives such as CS2Browser. You'll receive <@&1420834529370832978> role upon request.",
		Color:       bot.ColorOK,
	},
	{
		Title:       "Running Costs",
		Description: "Running servers and maintaining our infrastructure comes with the following costs:",
		Color:       bot.ColorOK,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Domain Registration (cs2browser.net)", Value: "€2 / month", Inline: true},
			{Name: "Domain Registration (swiftlys2.net)", Value: "€2 / month", Inline: true},
			{Name: "Cloudflare Pro Plan (cs2browser.net)", Value: "€25 / month", Inline: true},
			{Name: "Server Hosting (CS2Browser & SwiftlyS2)", Value: "€30 / month", Inline: true},
		},
	},
	{
		Title:       "Donation Methods",
		Description: "There are several ways you can support our projects:\n\n- [GitHub Sponsors](http://github.com/sponsors/swiftly-solution) - Credit Card\n- [Ko-fi](http://ko-fi.com/swiftlygroup) - Credit Card / PayPal\n- [Steam Skins](https://steamcommunity.com/tradeoffer/new/?partner=1139088750&token=9defPx3R)\n\n**Crypto**",
		Color:       bot.ColorOK,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "BTC (Bitcoin network)", Value: "`3KJd3rndVkMpEiSa42oBRehTMBKCtwgDAb`"},
			{Name: "ETH (Ethereum network)", Value: "`0x5f13c44488359ae0DB72379003c19A97c7aD36C5`"},
			{Name: "USDC (Ethereum network)", Value: "`0x05FC1E124E3a7d3FA16c55dF08B8cCaBf30b4363`"},
		},
	},
}
