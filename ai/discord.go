package ai

import (
	"fmt"
	"regexp"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/bwmarrin/discordgo"
)

const fence = "```"

var (
	roleMention = regexp.MustCompile(`<@&\d+>`)
	userMention = regexp.MustCompile(`<@!?(\d+)>`)
	massMention = regexp.MustCompile(`(?i)@\s*(everyone|here)\b`)
	topicRef    = regexp.MustCompile("(?i)`?\\btopic\\s+#?(\\d+)\\b`?")
	inlineLink  = regexp.MustCompile("`(https?://[^\\s`]+)`")
	fencedLinks = regexp.MustCompile("`{3}[\\w+#.-]*\\n((?:[ \\t]*https?://\\S+[ \\t]*\\n?)+)`{3}")
	fenceLine   = regexp.MustCompile("^\\s*`{3}(.*)$")
	fenceInfo   = regexp.MustCompile(`^[\w+#.-]{1,20}$`)
	codeBlock   = regexp.MustCompile("`{3}([\\w+#.-]*)\\n([\\s\\S]*?)\\n?`{3}")
)

// replace is ReplaceAllStringFunc with access to the submatches.
func replace(re *regexp.Regexp, s string, fn func(m []string) string) string {
	var b strings.Builder
	last := 0
	for _, loc := range re.FindAllStringSubmatchIndex(s, -1) {
		m := make([]string, len(loc)/2)
		for i := range m {
			if loc[2*i] >= 0 {
				m[i] = s[loc[2*i]:loc[2*i+1]]
			}
		}
		b.WriteString(s[last:loc[0]])
		b.WriteString(fn(m))
		last = loc[1]
	}
	b.WriteString(s[last:])
	return b.String()
}

func CleanMentions(text string, users []string) string {
	text = roleMention.ReplaceAllString(text, "")
	text = replace(userMention, text, func(m []string) string {
		if slices.Contains(users, m[1]) {
			return "<@" + m[1] + ">"
		}
		return ""
	})
	return massMention.ReplaceAllString(text, "${1}")
}

func LinkTopics(text string) string {
	return topicRef.ReplaceAllString(text, "https://forum.swiftlys2.net/t/-/${1}")
}

// UnwrapLinks takes URLs out of inline code and out of code blocks that only hold URLs, so they stay clickable.
func UnwrapLinks(text string) string {
	text = inlineLink.ReplaceAllString(text, "${1}")
	return replace(fencedLinks, text, func(m []string) string { return strings.TrimSpace(m[1]) })
}

// lastIndex is JS lastIndexOf(sep, from): the last sep that starts at or before from.
func lastIndex(s, sep string, from int) int {
	return strings.LastIndex(s[:min(len(s), from+len(sep))], sep)
}

// Chunk splits text into messages of at most max bytes, closing and reopening code fences across the cut.
func Chunk(text string, max int) []string {
	var out []string
	rest := strings.TrimSpace(text)
	open := ""
	for rest != "" {
		prefix := ""
		if open != "" {
			prefix = open + "\n"
		}
		budget := max - len(prefix) - len(fence) - 1

		var part string
		if len(rest) <= budget {
			part, rest = rest, ""
		} else {
			cut := lastIndex(rest, "\n\n", budget)
			if cut < budget/2 {
				cut = lastIndex(rest, "\n", budget)
			}
			if cut < budget/2 {
				cut = lastIndex(rest, " ", budget)
			}
			if cut <= 0 {
				cut = budget
				for cut > 0 && !utf8.RuneStart(rest[cut]) {
					cut--
				}
			}
			part, rest = rest[:cut], rest[cut:]
			if strings.HasPrefix(rest, "\n") {
				rest = strings.TrimLeft(rest, "\n")
			} else {
				rest = strings.TrimPrefix(rest, " ")
			}
		}

		for _, line := range strings.Split(part, "\n") {
			m := fenceLine.FindStringSubmatch(line)
			if m == nil {
				continue
			}
			info := strings.TrimSpace(m[1])
			if strings.Contains(info, fence) {
				continue
			}
			if open != "" {
				open = ""
			} else {
				open = fence
				if fenceInfo.MatchString(info) {
					open += info
				}
			}
		}
		if open != "" {
			part += "\n" + fence
		}
		out = append(out, prefix+part)
	}
	return out
}

func Prepare(text string, pingable []string) string {
	return UnwrapLinks(LinkTopics(CleanMentions(text, pingable)))
}

const (
	codeLimit = 1500
	maxFiles  = 10
)

var extensions = map[string]string{
	"cs": "cs", "csharp": "cs", "csproj": "csproj",
	"ts": "ts", "typescript": "ts", "js": "js", "javascript": "js",
	"py": "py", "python": "py", "cpp": "cpp", "c++": "cpp", "h": "h", "lua": "lua",
	"json": "json", "yml": "yml", "yaml": "yaml", "xml": "xml", "toml": "toml", "ini": "ini", "cfg": "cfg",
	"md": "md", "markdown": "md", "log": "log",
}

type File struct{ Name, Content string }

func ExtractCode(text string) (string, []File) {
	var files []File
	out := replace(codeBlock, text, func(m []string) string {
		code := m[2]
		if utf8.RuneCountInString(code) <= codeLimit || len(files) >= maxFiles {
			return m[0]
		}
		ext, ok := extensions[strings.ToLower(m[1])]
		if !ok {
			ext = "txt"
		}
		name := fmt.Sprintf("snippet-%d.%s", len(files)+1, ext)
		files = append(files, File{name, code})
		return fmt.Sprintf("📎 **%s** (%d lines) is attached below.", name, strings.Count(code, "\n")+1)
	})
	return out, files
}

func sendChunks(s *discordgo.Session, threadID, text, replyTo string, pingable []string) error {
	body, files := ExtractCode(Prepare(text, pingable))
	parts := Chunk(body, 1900)
	if len(parts) == 0 {
		parts = []string{"I have nothing to say about that. 😶"}
	}
	for i, content := range parts {
		msg := &discordgo.MessageSend{
			Content:         content,
			AllowedMentions: &discordgo.MessageAllowedMentions{Parse: []discordgo.AllowedMentionType{}, Users: pingable},
		}
		if i == 0 && replyTo != "" {
			msg.Reference = &discordgo.MessageReference{MessageID: replyTo, ChannelID: threadID}
		}
		if i == len(parts)-1 {
			for _, f := range files {
				msg.Files = append(msg.Files, &discordgo.File{Name: f.Name, ContentType: "text/plain", Reader: strings.NewReader(f.Content)})
			}
		}
		if _, err := s.ChannelMessageSendComplex(threadID, msg); err != nil {
			return err
		}
	}
	return nil
}
