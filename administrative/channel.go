package administrative

import (
	"context"
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"

	"github.com/swiftly-solution/discord-bot/bot"
	"github.com/swiftly-solution/discord-bot/store"
)

type roleText struct{ title, noun, already, missing string }

var roles = map[string]roleText{
	store.Honeypot:    {"Honeypot", "a honeypot", "is already set as a honeypot", "is not set as a honeypot"},
	store.Autopublish: {"Autopublish", "an autopublish", "is already set as an autopublish", "is not set as an autopublish"},
	store.Forum:       {"AI Forum", "an AI forum", "is already an AI forum", "is not an AI forum"},
}

func Command(st *store.Store) bot.Command {
	admin := int64(discordgo.PermissionAdministrator)
	return bot.Command{
		Def: &discordgo.ApplicationCommand{
			Name:                     "channel",
			Description:              "Channel commands",
			Contexts:                 &[]discordgo.InteractionContextType{discordgo.InteractionContextGuild},
			DefaultMemberPermissions: &admin,
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:         discordgo.ApplicationCommandOptionChannel,
					Name:         "channel",
					Description:  "The channel to use",
					Required:     true,
					ChannelTypes: []discordgo.ChannelType{discordgo.ChannelTypeGuildText, discordgo.ChannelTypeGuildNews, discordgo.ChannelTypeGuildForum},
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "action",
					Description: "The action to perform",
					Required:    true,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{Name: "Set Honeypot", Value: "set-honeypot"},
						{Name: "Remove Honeypot", Value: "remove-honeypot"},
						{Name: "Set Autopublish", Value: "set-autopublish"},
						{Name: "Remove Autopublish", Value: "remove-autopublish"},
						{Name: "Set AI Forum", Value: "set-forum"},
						{Name: "Remove AI Forum", Value: "remove-forum"},
						{Name: "Lock", Value: "lock"},
						{Name: "Unlock", Value: "unlock"},
					},
				},
			},
		},
		Run: func(s *discordgo.Session, i *discordgo.InteractionCreate) error { return run(st, s, i) },
	}
}

func run(st *store.Store, s *discordgo.Session, i *discordgo.InteractionCreate) error {
	var channelID, action string
	for _, o := range i.ApplicationCommandData().Options {
		switch o.Name {
		case "channel":
			channelID = o.ChannelValue(nil).ID
		case "action":
			action = o.StringValue()
		}
	}
	mention := "<#" + channelID + ">"

	if err := bot.Defer(s, i); err != nil {
		return err
	}

	if action == "lock" || action == "unlock" {
		if err := setSendMessages(s, channelID, i.GuildID, action == "unlock"); err != nil {
			return err
		}
		title := map[string]string{"lock": "Channel Locked", "unlock": "Channel Unlocked"}[action]
		return bot.Reply(s, i, &discordgo.MessageEmbed{
			Title:       title,
			Description: fmt.Sprintf("The channel %s has been %sed.", mention, action),
			Color:       bot.ColorOK,
		})
	}

	verb, role, _ := strings.Cut(action, "-")
	text, ok := roles[role]
	if !ok {
		return fmt.Errorf("unknown action %q", action)
	}
	adding := verb == "set"
	ctx := context.Background()
	var changed bool
	var err error
	if adding {
		changed, err = st.Add(ctx, i.GuildID, channelID, role)
	} else {
		changed, err = st.Remove(ctx, i.GuildID, channelID, role)
	}
	if err != nil {
		return err
	}

	switch {
	case !changed && adding:
		return bot.Reply(s, i, &discordgo.MessageEmbed{Title: text.title + " Already Set", Description: fmt.Sprintf("The channel %s %s.", mention, text.already), Color: bot.ColorError})
	case !changed:
		return bot.Reply(s, i, &discordgo.MessageEmbed{Title: text.title + " Not Found", Description: fmt.Sprintf("The channel %s %s.", mention, text.missing), Color: bot.ColorError})
	case adding:
		return bot.Reply(s, i, &discordgo.MessageEmbed{Title: text.title + " Set", Description: fmt.Sprintf("The channel %s has been set as %s.", mention, text.noun), Color: bot.ColorOK})
	default:
		return bot.Reply(s, i, &discordgo.MessageEmbed{Title: text.title + " Removed", Description: fmt.Sprintf("The channel %s has been removed as %s.", mention, text.noun), Color: bot.ColorOK})
	}
}

// setSendMessages flips only the Send Messages bit of the @everyone overwrite, keeping its other bits.
func setSendMessages(s *discordgo.Session, channelID, guildID string, allow bool) error {
	c, err := s.Channel(channelID)
	if err != nil {
		return err
	}
	var a, d int64
	for _, o := range c.PermissionOverwrites {
		if o.ID == guildID {
			a, d = o.Allow, o.Deny
		}
	}
	if allow {
		a, d = a|discordgo.PermissionSendMessages, d&^discordgo.PermissionSendMessages
	} else {
		a, d = a&^discordgo.PermissionSendMessages, d|discordgo.PermissionSendMessages
	}
	return s.ChannelPermissionSet(channelID, guildID, discordgo.PermissionOverwriteTypeRole, a, d)
}
