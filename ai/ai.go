package ai

import (
	"context"
	"net/http"
	"slices"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/swiftly-solution/discord-bot/bot"
	"github.com/swiftly-solution/discord-bot/logger"
	"github.com/swiftly-solution/discord-bot/mcpconn"
	"github.com/swiftly-solution/discord-bot/store"
)

type Config struct {
	APIKey          string // DeepSeek
	PersonalityPath string
	GitToken        string // GitHub MCP; empty disables the GitHub tools
	BypassKey       string // Cloudflare bypass for swiftlys2.net
}

var (
	cfg    Config
	tb     *toolbox
	forums store.Cache[struct{}]
)

func Setup(ctx context.Context, s *discordgo.Session, st *store.Store, docs *mcpconn.Conn, c Config) error {
	cfg = c
	tb = &toolbox{sources: []source{{prefix: "docs", conn: docs}}, defs: map[string][]ToolDef{}}
	if c.GitToken != "" {
		header := http.Header{}
		header.Set("Authorization", "Bearer "+c.GitToken)
		header.Set("X-MCP-Readonly", "true")
		github := mcpconn.New("swiftlychan-mcp-github", "https://api.githubcopilot.com/mcp/", header)
		tb.sources = append(tb.sources, source{prefix: "github", conn: github, allow: githubAllow})
	} else {
		logger.Printf("[AI] GIT_TOKEN is not set, GitHub tools are disabled")
	}

	s.AddHandler(onMessage)
	s.AddHandler(onThread)

	return bot.Every(ctx, 2*time.Minute, "AI", func(ctx context.Context) error {
		ids, err := st.Channels(ctx, store.Forum)
		if err != nil {
			return err
		}
		found := map[string]struct{}{}
		for _, id := range ids {
			found[id] = struct{}{}
		}
		forums.Replace(found)
		return nil
	})
}

func onMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author == nil || m.Author.Bot {
		return
	}
	thread, err := bot.Channel(s, m.ChannelID)
	if err != nil || !thread.IsThread() || m.ID == thread.ID {
		return
	}
	if _, ok := forums.Get(thread.ParentID); !ok {
		return
	}
	if !slices.ContainsFunc(m.Mentions, func(u *discordgo.User) bool { return u.ID == s.State.User.ID }) {
		return
	}
	respond(s, thread, m.ID)
}

func onThread(s *discordgo.Session, t *discordgo.ThreadCreate) {
	if !t.NewlyCreated {
		return
	}
	if _, ok := forums.Get(t.ParentID); !ok {
		return
	}
	var first *discordgo.Message
	for i := 0; i < 5 && first == nil; i++ {
		if first, _ = starter(s, t.Channel); first == nil {
			time.Sleep(time.Second)
		}
	}
	if first == nil || first.Author == nil || first.Author.Bot {
		return
	}
	s.ThreadJoin(t.ID)
	respond(s, t.Channel, "")
}
