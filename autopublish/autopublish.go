package autopublish

import (
	"context"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/swiftly-solution/discord-bot/bot"
	"github.com/swiftly-solution/discord-bot/logger"
	"github.com/swiftly-solution/discord-bot/store"
)

var channels store.Cache[struct{}]

func Setup(ctx context.Context, s *discordgo.Session, st *store.Store) error {
	s.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if _, ok := channels.Get(m.ChannelID); !ok || m.Flags&discordgo.MessageFlagsCrossPosted != 0 {
			return
		}
		if m.Type != discordgo.MessageTypeDefault && m.Type != discordgo.MessageTypeReply {
			return
		}
		if _, err := s.ChannelMessageCrosspost(m.ChannelID, m.ID); err != nil {
			logger.Printf("[Autopublish] Cannot publish %s in %s: %v", m.ID, m.ChannelID, err)
		}
	})

	return bot.Every(ctx, 2*time.Minute, "Autopublish", func(ctx context.Context) error {
		ids, err := st.Channels(ctx, store.Autopublish)
		if err != nil {
			return err
		}
		found := map[string]struct{}{}
		for _, id := range ids {
			if c, err := bot.Channel(s, id); err == nil && c.Type == discordgo.ChannelTypeGuildNews {
				found[id] = struct{}{}
			}
		}
		channels.Replace(found)
		return nil
	})
}
