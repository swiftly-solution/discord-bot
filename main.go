package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"

	"github.com/swiftly-solution/discord-bot/administrative"
	"github.com/swiftly-solution/discord-bot/ai"
	"github.com/swiftly-solution/discord-bot/autopublish"
	"github.com/swiftly-solution/discord-bot/bot"
	"github.com/swiftly-solution/discord-bot/docs"
	"github.com/swiftly-solution/discord-bot/donate"
	"github.com/swiftly-solution/discord-bot/honeypot"
	"github.com/swiftly-solution/discord-bot/logger"
	"github.com/swiftly-solution/discord-bot/store"
)

func main() {
	_ = godotenv.Load()

	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		logger.Fatalf("DISCORD_TOKEN is not set")
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db, err := pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		logger.Fatalf("Database: %v", err)
	}
	defer db.Close()
	if err := db.Ping(ctx); err != nil {
		logger.Fatalf("Database: %v", err)
	}
	st := store.New(db)

	s, err := discordgo.New("Bot " + token)
	if err != nil {
		logger.Fatalf("Discord: %v", err)
	}
	s.Identify.Intents = discordgo.IntentsGuilds | discordgo.IntentsGuildMessages | discordgo.IntentsGuildWebhooks |
		discordgo.IntentsMessageContent | discordgo.IntentsGuildMessageReactions | discordgo.IntentGuildModeration |
		discordgo.IntentsDirectMessages | discordgo.IntentsGuildInvites
	if err := s.Open(); err != nil {
		logger.Fatalf("Discord: %v", err)
	}
	defer s.Close()

	bypassKey := os.Getenv("CF_BYPASS_KEY")
	docsConn := docs.New(bypassKey)

	modules := []struct {
		name  string
		setup func() error
	}{
		{"HONEYPOT", func() error { return honeypot.Setup(ctx, s, st) }},
		{"AUTOPUBLISH", func() error { return autopublish.Setup(ctx, s, st) }},
		{"AI", func() error {
			return ai.Setup(ctx, s, st, docsConn, ai.Config{
				APIKey:          os.Getenv("DEEPSEEK_AI_API_KEY"),
				PersonalityPath: os.Getenv("PERSONALITY_PATH"),
				GitToken:        os.Getenv("GIT_TOKEN"),
				BypassKey:       bypassKey,
			})
		}},
	}
	for _, m := range modules {
		if err := m.setup(); err != nil {
			logger.Fatalf("[%s] Failed to load module: %v", m.name, err)
		}
		logger.Printf("[%s] Module has been successfully loaded!", m.name)
	}

	commands := append([]bot.Command{administrative.Command(st), donate.Command}, docs.Commands(docsConn)...)
	if err := bot.Register(s, commands); err != nil {
		logger.Fatalf("[Commands] %v", err)
	}
	if err := s.UpdateWatchStatus(0, "https://swiftlys2.net"); err != nil {
		logger.Printf("Cannot set activity: %v", err)
	}
	if _, err := docsConn.Session(ctx); err != nil {
		logger.Printf("[Docs] MCP unavailable, retrying on first use: %v", err)
	}

	logger.Printf("Bot is running.")
	<-ctx.Done()
	logger.Printf("Shutting down.")
}
