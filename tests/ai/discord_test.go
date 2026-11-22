package ai_test

import (
	"strings"
	"testing"

	"github.com/swiftly-solution/discord-bot/ai"
)

func TestCleanMentions(t *testing.T) {
	got := ai.CleanMentions("hi <@1> <@!2> <@&3> @everyone @ here", []string{"1"})
	if want := "hi <@1>   everyone here"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestLinksStayClickable(t *testing.T) {
	got := ai.Prepare("see topic #42 and `https://a.b/c`\n```\nhttps://x.y/z\n```", nil)
	want := "see https://forum.swiftlys2.net/t/-/42 and https://a.b/c\nhttps://x.y/z"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestChunkReopensFences(t *testing.T) {
	text := "intro\n\n```cs\n" + strings.Repeat("var x = 1;\n", 50) + "```\nend"
	parts := ai.Chunk(text, 200)
	if len(parts) < 2 {
		t.Fatalf("want several parts, got %d", len(parts))
	}
	for i, p := range parts {
		if len(p) > 200 {
			t.Errorf("part %d is %d bytes", i, len(p))
		}
		if strings.Count(p, "```")%2 != 0 {
			t.Errorf("part %d has an unclosed fence:\n%s", i, p)
		}
		if i > 0 && i < len(parts)-1 && !strings.HasPrefix(p, "```cs\n") {
			t.Errorf("part %d does not reopen the cs fence:\n%s", i, p)
		}
	}
	if joined := strings.Join(parts, ""); !strings.Contains(joined, "end") || !strings.Contains(joined, "intro") {
		t.Errorf("text lost: %q", joined)
	}
}

func TestChunkCutsOnRuneBoundary(t *testing.T) {
	for _, p := range ai.Chunk(strings.Repeat("é", 300), 100) {
		if !strings.HasPrefix(p, "é") || strings.ContainsRune(p, '�') {
			t.Fatalf("broken rune in %q", p)
		}
	}
}

func TestExtractCode(t *testing.T) {
	long := strings.Repeat("x\n", 800)
	text, files := ai.ExtractCode("short\n```py\nprint(1)\n```\nlong\n```python\n" + long + "```")
	if len(files) != 1 || files[0].Name != "snippet-1.py" {
		t.Fatalf("files = %+v", files)
	}
	if !strings.Contains(text, "```py\nprint(1)\n```") || !strings.Contains(text, "**snippet-1.py** (800 lines) is attached below.") {
		t.Errorf("text = %q", text)
	}
}
