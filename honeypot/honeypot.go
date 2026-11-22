package honeypot

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/swiftly-solution/discord-bot/bot"
	"github.com/swiftly-solution/discord-bot/logger"
	"github.com/swiftly-solution/discord-bot/store"
)

const (
	title   = "DO NOT SEND MESSAGES IN THIS CHANNEL"
	timeout = 7 * 24 * time.Hour
)

// channel id -> id of the pinned warning message ("" when it could not be posted)
var channels store.Cache[string]

func Setup(ctx context.Context, s *discordgo.Session, st *store.Store) error {
	s.AddHandler(onMessage)

	return bot.Every(ctx, 2*time.Minute, "Honeypot", func(ctx context.Context) error {
		ids, err := st.Channels(ctx, store.Honeypot)
		if err != nil {
			return err
		}
		found := map[string]string{}
		for _, id := range ids {
			c, err := bot.Channel(s, id)
			if err != nil || (c.Type != discordgo.ChannelTypeGuildText && c.Type != discordgo.ChannelTypeGuildNews) {
				continue
			}
			pin, err := warning(s, id)
			if err != nil {
				logger.Printf("[Honeypot] Cannot set up the warning in %s: %v", id, err)
			}
			found[id] = pin
		}
		channels.Replace(found)
		return nil
	})
}

func warning(s *discordgo.Session, channelID string) (string, error) {
	pins, err := s.ChannelMessagesPinned(channelID)
	if err != nil {
		return "", err
	}
	for _, m := range pins {
		if m.Author != nil && m.Author.ID == s.State.User.ID && len(m.Embeds) > 0 && m.Embeds[0].Title == title {
			return m.ID, nil
		}
	}
	m, err := s.ChannelMessageSendComplex(channelID, &discordgo.MessageSend{
		Embeds: []*discordgo.MessageEmbed{{
			Title:       title,
			Description: "This channel is used to catch spam bots. Any messages sent here will result in a **timeout for 1 week.**",
		}},
		Components: []discordgo.MessageComponent{counter(0)},
	})
	if err != nil {
		return "", err
	}
	return m.ID, s.ChannelMessagePin(channelID, m.ID)
}

func counter(n int) discordgo.ActionsRow {
	return discordgo.ActionsRow{Components: []discordgo.MessageComponent{discordgo.Button{
		CustomID: "honeypot-timeouts",
		Emoji:    &discordgo.ComponentEmoji{Name: "🎂"},
		Label:    fmt.Sprintf("Timeouts: %d", n),
		Style:    discordgo.SecondaryButton,
		Disabled: true,
	}}}
}

func onMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	pin, ok := channels.Get(m.ChannelID)
	if !ok || m.Author == nil || m.Author.Bot {
		return
	}
	if err := s.ChannelMessageDelete(m.ChannelID, m.ID); err != nil {
		logger.Printf("[Honeypot] Cannot delete %s: %v", m.ID, err)
	}
	until := time.Now().Add(timeout)
	if err := s.GuildMemberTimeout(m.GuildID, m.Author.ID, &until, discordgo.WithAuditLogReason("Honeypot")); err != nil {
		logger.Printf("[Honeypot] Cannot time out %s: %v", m.Author.ID, err)
		return
	}
	if pin == "" {
		return
	}

	msg, err := s.ChannelMessage(m.ChannelID, pin)
	if err != nil {
		return
	}
	n := 0
	if len(msg.Components) > 0 {
		if row, ok := msg.Components[0].(*discordgo.ActionsRow); ok && len(row.Components) > 0 {
			if b, ok := row.Components[0].(*discordgo.Button); ok {
				_, count, _ := strings.Cut(b.Label, ": ")
				n, _ = strconv.Atoi(count)
			}
		}
	}
	// shortcut: read-modify-write, two spammers in the same instant count once
	components := []discordgo.MessageComponent{counter(n + 1)}
	if _, err := s.ChannelMessageEditComplex(&discordgo.MessageEdit{ID: pin, Channel: m.ChannelID, Components: &components}); err != nil {
		logger.Printf("[Honeypot] Cannot update the counter: %v", err)
	}
}
