package mcpconn

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const Timeout = 30 * time.Second

// Conn is a lazily (re)connected MCP session to a streamable HTTP server.
type Conn struct {
	name     string
	endpoint string
	header   http.Header

	mu      sync.Mutex
	session *mcp.ClientSession
}

func New(name, endpoint string, header http.Header) *Conn {
	return &Conn{name: name, endpoint: endpoint, header: header}
}

func (c *Conn) Session(ctx context.Context) (*mcp.ClientSession, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.session != nil {
		return c.session, nil
	}
	client := mcp.NewClient(&mcp.Implementation{Name: c.name, Version: "1.0.0"}, nil)
	ctx, cancel := context.WithTimeout(ctx, Timeout)
	defer cancel()
	session, err := client.Connect(ctx, &mcp.StreamableClientTransport{
		Endpoint:   c.endpoint,
		HTTPClient: &http.Client{Transport: headers{c.header}},
	}, nil)
	if err != nil {
		return nil, c.redact(err)
	}
	c.session = session
	return session, nil
}

// redact hides the endpoint's query string (it can hold a key) in errors that get logged.
func (c *Conn) redact(err error) error {
	if base, query, ok := strings.Cut(c.endpoint, "?"); ok && query != "" {
		return errors.New(strings.ReplaceAll(err.Error(), c.endpoint, base+"?***"))
	}
	return err
}

func (c *Conn) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.session != nil {
		c.session.Close()
		c.session = nil
	}
}

func (c *Conn) ListTools(ctx context.Context) ([]*mcp.Tool, error) {
	session, err := c.Session(ctx)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, Timeout)
	defer cancel()
	res, err := session.ListTools(ctx, nil)
	if err != nil {
		c.Reset()
		return nil, c.redact(err)
	}
	return res.Tools, nil
}

// CallTool returns the text content of the result; a failed call resets the session.
func (c *Conn) CallTool(ctx context.Context, name string, args any) (text string, isError bool, err error) {
	session, err := c.Session(ctx)
	if err != nil {
		return "", false, err
	}
	ctx, cancel := context.WithTimeout(ctx, Timeout)
	defer cancel()
	res, err := session.CallTool(ctx, &mcp.CallToolParams{Name: name, Arguments: args})
	if err != nil {
		c.Reset()
		return "", false, c.redact(err)
	}
	var texts []string
	for _, content := range res.Content {
		if t, ok := content.(*mcp.TextContent); ok {
			texts = append(texts, t.Text)
		}
	}
	return strings.Join(texts, "\n"), res.IsError, nil
}

type headers struct{ h http.Header }

func (t headers) RoundTrip(req *http.Request) (*http.Response, error) {
	if len(t.h) > 0 {
		req = req.Clone(req.Context())
		for k, v := range t.h {
			req.Header[k] = v
		}
	}
	return http.DefaultTransport.RoundTrip(req)
}
