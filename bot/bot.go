package bot

import (
	"context"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/swiftly-solution/discord-bot/logger"
)

const (
	ColorOK    = 0x00feed
	ColorError = 0xff0000
)

type Command struct {
	Def *discordgo.ApplicationCommand
	Run func(s *discordgo.Session, i *discordgo.InteractionCreate) error
}

func Register(s *discordgo.Session, commands []Command) error {
	byName := map[string]Command{}
	defs := make([]*discordgo.ApplicationCommand, 0, len(commands))
	for _, c := range commands {
		byName[c.Def.Name] = c
		defs = append(defs, c.Def)
	}

	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.Type != discordgo.InteractionApplicationCommand {
			return
		}
		c, ok := byName[i.ApplicationCommandData().Name]
		if !ok {
			return
		}
		if err := c.Run(s, i); err != nil {
			logger.Printf("[Commands] /%s failed: %v", c.Def.Name, err)
			fail(s, i)
		}
	})

	logger.Printf("[Commands] Loading %d application (/) commands.", len(defs))
	// shortcut: first 200 guilds only, paginate if the bot ever joins more
	guilds, err := s.UserGuilds(200, "", "", false)
	if err != nil {
		return err
	}
	for _, g := range guilds {
		if _, err := s.ApplicationCommandBulkOverwrite(s.State.User.ID, g.ID, defs); err != nil {
			return err
		}
	}
	logger.Printf("[Commands] Updated %d application (/) commands in %d guilds.", len(defs), len(guilds))
	return nil
}

func fail(s *discordgo.Session, i *discordgo.InteractionCreate) {
	const text = "An error has occurred, please try again."
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{Content: text, Flags: discordgo.MessageFlagsEphemeral},
	})
	if err != nil { // already deferred
		content := text
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: &content})
	}
}

// Defer shows "thinking..." so the handler has 15 minutes to reply with Reply.
func Defer(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
}

func Reply(s *discordgo.Session, i *discordgo.InteractionCreate, embeds ...*discordgo.MessageEmbed) error {
	_, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Embeds: &embeds})
	return err
}

func User(i *discordgo.InteractionCreate) *discordgo.User {
	if i.Member != nil {
		return i.Member.User
	}
	return i.User
}

func Channel(s *discordgo.Session, id string) (*discordgo.Channel, error) {
	if c, err := s.State.Channel(id); err == nil {
		return c, nil
	}
	return s.Channel(id)
}

// Every runs fn now, then every d until ctx is done; only the first error is returned, later ones are logged.
func Every(ctx context.Context, d time.Duration, name string, fn func(context.Context) error) error {
	if err := fn(ctx); err != nil {
		return err
	}
	go func() {
		t := time.NewTicker(d)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				if err := fn(ctx); err != nil {
					logger.Printf("[%s] Refresh failed: %v", name, err)
				}
			}
		}
	}()
	return nil
}
