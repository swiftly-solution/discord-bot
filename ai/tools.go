package ai

import (
	"context"
	"encoding/json"
	"strings"
	"sync"
	"time"

	"github.com/swiftly-solution/discord-bot/logger"
	"github.com/swiftly-solution/discord-bot/mcpconn"
)

type ToolDef struct {
	Type     string   `json:"type"`
	Function ToolFunc `json:"function"`
}

type ToolFunc struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Parameters  any    `json:"parameters"`
}

const (
	maxResult = 20000
	retryWait = time.Minute
)

var githubAllow = map[string]bool{
	"get_file_contents": true, "search_code": true, "search_repositories": true,
	"list_commits": true, "get_commit": true, "list_branches": true, "list_releases": true,
	"list_issues": true, "issue_read": true, "search_issues": true,
	"list_pull_requests": true, "pull_request_read": true, "search_pull_requests": true,
}

type source struct {
	prefix string
	conn   *mcpconn.Conn
	allow  map[string]bool // nil allows every tool
}

// toolbox exposes the MCP sources (plus the built-in forum tools) to the model as "<prefix>__<tool>".
type toolbox struct {
	sources []source

	mu      sync.Mutex
	defs    map[string][]ToolDef // prefix -> tools; a missing prefix is not connected
	lastTry time.Time
}

// tools lists every tool, first reconnecting missing sources (at most once a minute).
func (t *toolbox) tools(ctx context.Context) []ToolDef {
	t.mu.Lock()
	defer t.mu.Unlock()
	if len(t.defs) < len(t.sources) && time.Since(t.lastTry) >= retryWait {
		t.lastTry = time.Now()
		loaded := make([][]ToolDef, len(t.sources))
		var wg sync.WaitGroup
		for i, src := range t.sources {
			if _, ok := t.defs[src.prefix]; ok {
				continue
			}
			wg.Go(func() { loaded[i] = load(ctx, src) })
		}
		wg.Wait()
		for i, src := range t.sources {
			if loaded[i] != nil {
				t.defs[src.prefix] = loaded[i]
			}
		}
	}

	var all []ToolDef
	for _, src := range t.sources {
		all = append(all, t.defs[src.prefix]...)
	}
	return append(all, forumTools...)
}

func load(ctx context.Context, src source) []ToolDef {
	tools, err := src.conn.ListTools(ctx)
	if err != nil {
		logger.Printf("[AI] MCP %q unavailable: %s", src.prefix, Scrub(err.Error()))
		return nil
	}
	defs := []ToolDef{}
	for _, tool := range tools {
		if src.allow != nil && !src.allow[tool.Name] {
			continue
		}
		defs = append(defs, ToolDef{Type: "function", Function: ToolFunc{
			Name:        src.prefix + "__" + tool.Name,
			Description: tool.Description,
			Parameters:  tool.InputSchema,
		}})
	}
	return defs
}

func (t *toolbox) call(ctx context.Context, name, arguments string) string {
	var args map[string]any
	if arguments == "" || json.Unmarshal([]byte(arguments), &args) != nil || args == nil {
		args = map[string]any{}
	}
	prefix, tool, _ := strings.Cut(name, "__")
	if prefix == "forum" {
		return callForum(ctx, tool, args)
	}

	src, ok := t.find(prefix, name)
	if !ok {
		return "Unknown tool " + name
	}
	text, _, err := src.conn.CallTool(ctx, tool, args)
	if err != nil {
		t.mu.Lock()
		delete(t.defs, prefix) // session may have died; reconnect on the next answer
		t.lastTry = time.Time{}
		t.mu.Unlock()
		return Scrub("Tool error: " + err.Error())
	}
	if r := []rune(Scrub(text)); len(r) > maxResult {
		text = string(r[:maxResult])
	} else {
		text = string(r)
	}
	if text == "" {
		return "(empty result)"
	}
	return text
}

// find returns the source of a listed tool, so the model cannot call tools outside the allow list.
func (t *toolbox) find(prefix, name string) (source, bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	for _, src := range t.sources {
		if src.prefix != prefix {
			continue
		}
		for _, d := range t.defs[prefix] {
			if d.Function.Name == name {
				return src, true
			}
		}
	}
	return source{}, false
}
