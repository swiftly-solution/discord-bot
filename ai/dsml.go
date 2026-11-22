package ai

import (
	"encoding/json"
	"regexp"
	"strings"
)

// DeepSeek sometimes writes its tool calls as DSML markup in the text instead of tool_calls.

type ParsedCall struct {
	Name string
	Args map[string]any
}

var (
	dsmlMark    = regexp.MustCompile(`<[^<>\n]*DSML[^<>\n]*>`)
	dsmlInvoke  = regexp.MustCompile(`invoke\s+name="`)
	dsmlParam   = regexp.MustCompile(`parameter\s+name="([^"]+)"([^>]*)>([\s\S]*?)<[^<>]*/[^<>]*parameter>`)
	dsmlNoQuote = regexp.MustCompile(`string="false"`)
)

func HasDsml(text string) bool { return strings.Contains(text, "DSML") }

func ParseDsml(text string) []ParsedCall {
	var calls []ParsedCall
	for _, part := range dsmlInvoke.Split(text, -1)[1:] {
		name, _, _ := strings.Cut(part, `"`)
		if name == "" {
			continue
		}
		args := map[string]any{}
		for _, m := range dsmlParam.FindAllStringSubmatch(part, -1) {
			raw := strings.TrimSpace(m[3])
			var v any
			if dsmlNoQuote.MatchString(m[2]) && json.Unmarshal([]byte(raw), &v) == nil {
				args[m[1]] = v
				continue
			}
			args[m[1]] = raw
		}
		calls = append(calls, ParsedCall{Name: name, Args: args})
	}
	return calls
}

func StripDsml(text string) string {
	if loc := dsmlMark.FindStringIndex(text); loc != nil {
		text = text[:loc[0]]
	}
	return strings.TrimSpace(text)
}
