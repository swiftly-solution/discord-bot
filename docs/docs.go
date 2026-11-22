package docs

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/bwmarrin/discordgo"

	"github.com/swiftly-solution/discord-bot/bot"
	"github.com/swiftly-solution/discord-bot/mcpconn"
)

var categories = []*discordgo.ApplicationCommandOptionChoice{
	{Name: "Schema", Value: "schema"},
	{Name: "Entities", Value: "entity"},
	{Name: "Game Events", Value: "gameevent"},
	{Name: "NetMessages", Value: "protobuf"},
	{Name: "ConVars & ConCommands", Value: "convar"},
	{Name: "Panorama", Value: "panorama"},
}

// New connects to the SwiftlyS2 docs MCP server; it is shared with the AI tools.
func New(bypassKey string) *mcpconn.Conn {
	return mcpconn.New("swiftlychan-mcp-docs", "https://swiftlys2.net/api/mcp?bypass_key="+url.QueryEscape(bypassKey), nil)
}

func Commands(conn *mcpconn.Conn) []bot.Command {
	return []bot.Command{
		command(conn, "lookup", "Look up exact information through SwiftlyS2 Docs", "Lookup", describeLookup),
		command(conn, "search", "Search information through SwiftlyS2 Docs", "Search", describeSearch),
	}
}

type describer func(category, name string, result any) string

func command(conn *mcpconn.Conn, verb, description, title string, describe describer) bot.Command {
	return bot.Command{
		Def: &discordgo.ApplicationCommand{
			Name:        verb,
			Description: description,
			Contexts:    &[]discordgo.InteractionContextType{discordgo.InteractionContextGuild},
			Options: []*discordgo.ApplicationCommandOption{
				{Type: discordgo.ApplicationCommandOptionString, Name: "category", Description: "The category to search in", Required: true, Choices: categories},
				{Type: discordgo.ApplicationCommandOptionString, Name: "name", Description: "The name to " + verb, Required: true},
			},
		},
		Run: func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
			var category, name string
			for _, o := range i.ApplicationCommandData().Options {
				switch o.Name {
				case "category":
					category = o.StringValue()
				case "name":
					name = o.StringValue()
				}
			}
			pretty := "Unknown"
			for _, c := range categories {
				if c.Value == category {
					pretty = c.Name
				}
			}
			args := map[string]any{"q": name}
			if verb == "lookup" {
				args = map[string]any{"name": name}
				if category == "entity" {
					args = map[string]any{"className": name}
				}
			}

			if err := bot.Defer(s, i); err != nil {
				return err
			}

			var err error
			for range 3 {
				var text string
				var isError bool
				text, isError, err = conn.CallTool(context.Background(), category+"_"+verb, args)
				if err != nil {
					continue
				}
				if isError {
					return bot.Reply(s, i, &discordgo.MessageEmbed{
						Title:       title + " Result - " + pretty,
						Description: "`" + name + "` could not be found.",
						Color:       bot.ColorError,
					})
				}
				var result any
				if err = json.Unmarshal([]byte(text), &result); err != nil {
					continue
				}
				return bot.Reply(s, i, &discordgo.MessageEmbed{
					Title:       title + " Result - " + pretty,
					Description: describe(category, name, result),
					Color:       bot.ColorOK,
					Footer: &discordgo.MessageEmbedFooter{
						Text: fmt.Sprintf("SwiftlyS2 Docs - requested by %s | %s", bot.User(i), time.Now().Format("02/01/2006 15:04:05")),
					},
				})
			}
			return err
		},
	}
}

// get walks nested JSON objects; a missing key gives nil.
func get(v any, path ...string) any {
	for _, k := range path {
		m, ok := v.(map[string]any)
		if !ok {
			return nil
		}
		v = m[k]
	}
	return v
}

// str prints a JSON value like a JS template literal would, with def for null/missing.
func str(v any, def string) string {
	switch x := v.(type) {
	case nil:
		return def
	case string:
		return x
	case float64:
		return strconv.FormatFloat(x, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(x)
	default:
		b, _ := json.Marshal(x)
		return string(b)
	}
}

func list(v any) []any {
	l, _ := v.([]any)
	return l
}

func first(l []any, n int) []any {
	return l[:min(n, len(l))]
}

func ticked(l []any, sep string) string {
	parts := make([]string, len(l))
	for i, v := range l {
		parts[i] = "`" + str(v, "") + "`"
	}
	return strings.Join(parts, sep)
}

func capitalize(s string) string {
	if s == "" {
		return s
	}
	r, size := utf8.DecodeRuneInString(s)
	return string(unicode.ToUpper(r)) + s[size:]
}
