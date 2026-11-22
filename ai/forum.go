package ai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const forumBase = "https://forum.swiftlys2.net"

var (
	forumHTTP  = &http.Client{Timeout: 30 * time.Second}
	topicPath  = regexp.MustCompile(`^/t/(?:[^/]*[^/\d][^/]*/)?(\d+)(?:/(\d+))?/?$`)
	htmlTag    = regexp.MustCompile(`<[^>]+>`)
	htmlEntity = strings.NewReplacer("&nbsp;", " ", "&lt;", "<", "&gt;", ">", "&quot;", `"`, "&#39;", "'", "&amp;", "&")
	errURL     = errors.New("url must be a forum.swiftlys2.net topic or post url")
)

// Scrub hides the Cloudflare bypass key in text shown to the model or written to logs.
func Scrub(text string) string {
	if cfg.BypassKey == "" {
		return text
	}
	text = strings.ReplaceAll(text, cfg.BypassKey, "***")
	return strings.ReplaceAll(text, url.QueryEscape(cfg.BypassKey), "***")
}

func forumTool(name, description, prop string) ToolDef {
	return ToolDef{Type: "function", Function: ToolFunc{
		Name:        "forum__" + name,
		Description: description,
		Parameters: map[string]any{
			"type":       "object",
			"properties": map[string]any{prop: map[string]any{"type": "string", "description": prop}},
			"required":   []string{prop},
		},
	}}
}

var forumTools = []ToolDef{
	forumTool("search", "Search the SwiftlyS2 forum (Discourse). Returns matching topics and post snippets.", "query"),
	forumTool("get_topic", "Read a forum topic (title and the first posts). Pass the topic or post url from a search result.", "url"),
	forumTool("get_post", "Read one forum post (raw markdown). Pass the post url from a search result or topic.", "url"),
}

func forumGet(ctx context.Context, path string, params url.Values, out any) error {
	qs := params.Encode()
	if qs != "" {
		qs += "&"
	}
	qs += "bypass_key=" + url.QueryEscape(cfg.BypassKey) // always last
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, forumBase+path+"?"+qs, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	res, err := forumHTTP.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode/100 != 2 {
		return fmt.Errorf("Discourse answered %d for %s", res.StatusCode, path)
	}
	return json.NewDecoder(res.Body).Decode(out)
}

func link(slug string, topicID, postNumber int) string {
	if slug == "" {
		slug = "-"
	}
	s := fmt.Sprintf("%s/t/%s/%d", forumBase, slug, topicID)
	if postNumber > 0 {
		s += "/" + strconv.Itoa(postNumber)
	}
	return s
}

// ParseForumURL returns the topic id and post number (0 when absent) of a forum url.
func ParseForumURL(value string) (topicID, postNumber int, err error) {
	u, err := url.Parse(value)
	if err != nil || u.Scheme+"://"+u.Host != forumBase {
		return 0, 0, errURL
	}
	m := topicPath.FindStringSubmatch(u.EscapedPath())
	if m == nil {
		return 0, 0, errURL
	}
	topicID, _ = strconv.Atoi(m[1])
	postNumber, _ = strconv.Atoi(m[2])
	return topicID, postNumber, nil
}

// callForum runs a forum tool; errors are returned as the result text, like an MCP isError result.
func callForum(ctx context.Context, name string, args map[string]any) string {
	result, err := forum(ctx, name, args)
	if err != nil {
		return Scrub(err.Error())
	}
	b, _ := json.Marshal(result)
	return string(b)
}

func forum(ctx context.Context, name string, args map[string]any) (any, error) {
	arg := func(k string) string {
		if v, ok := args[k]; ok && v != nil {
			return fmt.Sprint(v)
		}
		return ""
	}
	switch name {
	case "search":
		var data struct {
			Topics []struct {
				ID         int    `json:"id"`
				Title      string `json:"title"`
				Slug       string `json:"slug"`
				PostsCount int    `json:"posts_count"`
			} `json:"topics"`
			Posts []struct {
				Username   string `json:"username"`
				Blurb      string `json:"blurb"`
				TopicID    int    `json:"topic_id"`
				PostNumber int    `json:"post_number"`
			} `json:"posts"`
		}
		if err := forumGet(ctx, "/search.json", url.Values{"q": {arg("query")}}, &data); err != nil {
			return nil, err
		}
		slugs := map[int]string{}
		type topic struct {
			Title string `json:"title"`
			Posts int    `json:"posts"`
			URL   string `json:"url"`
		}
		type post struct {
			User    string `json:"user"`
			Snippet string `json:"snippet"`
			URL     string `json:"url"`
		}
		result := struct {
			Topics []topic `json:"topics"`
			Posts  []post  `json:"posts"`
		}{Topics: []topic{}, Posts: []post{}}
		for _, t := range data.Topics {
			slugs[t.ID] = t.Slug
		}
		for _, t := range data.Topics[:min(10, len(data.Topics))] {
			result.Topics = append(result.Topics, topic{t.Title, t.PostsCount, link(t.Slug, t.ID, 0)})
		}
		for _, p := range data.Posts[:min(10, len(data.Posts))] {
			result.Posts = append(result.Posts, post{p.Username, p.Blurb, link(slugs[p.TopicID], p.TopicID, p.PostNumber)})
		}
		return result, nil

	case "get_topic":
		topicID, _, err := ParseForumURL(arg("url"))
		if err != nil {
			return nil, err
		}
		var data struct {
			Title      string `json:"title"`
			Slug       string `json:"slug"`
			PostStream struct {
				Posts []struct {
					PostNumber int    `json:"post_number"`
					Username   string `json:"username"`
					Cooked     string `json:"cooked"`
				} `json:"posts"`
			} `json:"post_stream"`
		}
		if err := forumGet(ctx, fmt.Sprintf("/t/%d.json", topicID), nil, &data); err != nil {
			return nil, err
		}
		type post struct {
			Number int    `json:"number"`
			User   string `json:"user"`
			Text   string `json:"text"`
			URL    string `json:"url"`
		}
		posts := []post{}
		for _, p := range data.PostStream.Posts {
			posts = append(posts, post{p.PostNumber, p.Username, htmlEntity.Replace(htmlTag.ReplaceAllString(p.Cooked, "")), link(data.Slug, topicID, p.PostNumber)})
		}
		return struct {
			Title string `json:"title"`
			URL   string `json:"url"`
			Posts []post `json:"posts"`
		}{data.Title, link(data.Slug, topicID, 0), posts}, nil

	case "get_post":
		topicID, postNumber, err := ParseForumURL(arg("url"))
		if err != nil {
			return nil, err
		}
		var data struct {
			Username   string `json:"username"`
			TopicSlug  string `json:"topic_slug"`
			PostNumber int    `json:"post_number"`
			Raw        string `json:"raw"`
		}
		if err := forumGet(ctx, fmt.Sprintf("/posts/by_number/%d/%d.json", topicID, max(postNumber, 1)), nil, &data); err != nil {
			return nil, err
		}
		return struct {
			User string `json:"user"`
			URL  string `json:"url"`
			Raw  string `json:"raw"`
		}{data.Username, link(data.TopicSlug, topicID, data.PostNumber), data.Raw}, nil
	}
	return nil, fmt.Errorf("Unknown tool %s", name)
}
