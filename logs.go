package testkit

import (
	"context"
	"log/slog"
	"strings"
	"sync"
	"testing"

	"github.com/ahmedbaig/go-backend-logging"
)

type Entry struct {
	Level   slog.Level
	Message string
	Attrs   map[string]any
}

type LogCapture struct {
	mu      sync.Mutex
	entries []Entry
}

type captureHandler struct {
	c     *LogCapture
	attrs []slog.Attr
	group string
}

func (h *captureHandler) Enabled(context.Context, slog.Level) bool { return true }

func (h *captureHandler) Handle(_ context.Context, r slog.Record) error {
	e := Entry{Level: r.Level, Message: r.Message, Attrs: map[string]any{}}
	for _, a := range h.attrs {
		flatten(e.Attrs, "", a)
	}
	r.Attrs(func(a slog.Attr) bool {
		flatten(e.Attrs, h.group, a)
		return true
	})
	h.c.mu.Lock()
	h.c.entries = append(h.c.entries, e)
	h.c.mu.Unlock()
	return nil
}

func flatten(m map[string]any, prefix string, a slog.Attr) {
	a.Value = a.Value.Resolve()
	key := a.Key
	if prefix != "" {
		key = prefix + "." + key
	}
	if a.Value.Kind() == slog.KindGroup {
		for _, g := range a.Value.Group() {
			flatten(m, key, g)
		}
		return
	}
	m[key] = a.Value.Any()
}

func (h *captureHandler) WithAttrs(as []slog.Attr) slog.Handler {
	c := *h
	c.attrs = append([]slog.Attr{}, h.attrs...)
	for _, a := range as {
		if h.group != "" {
			a.Key = h.group + "." + a.Key
		}
		c.attrs = append(c.attrs, a)
	}
	return &c
}

func (h *captureHandler) WithGroup(name string) slog.Handler {
	c := *h
	if c.group != "" {
		c.group += "." + name
	} else {
		c.group = name
	}
	return &c
}

func Logger(t testing.TB) (*slog.Logger, *LogCapture) {
	t.Helper()
	c := &LogCapture{}
	t.Cleanup(func() {
		if t.Failed() {
			for _, e := range c.Entries() {
				t.Logf("[%s] %s %v", logging.LevelName(e.Level), e.Message, e.Attrs)
			}
		}
	})
	return slog.New(logging.NewContextHandler(&captureHandler{c: c})), c
}

func (c *LogCapture) Entries() []Entry {
	c.mu.Lock()
	defer c.mu.Unlock()
	return append([]Entry(nil), c.entries...)
}

func (c *LogCapture) Reset() {
	c.mu.Lock()
	c.entries = nil
	c.mu.Unlock()
}

func (c *LogCapture) Find(sub string) (Entry, bool) {
	for _, e := range c.Entries() {
		if strings.Contains(e.Message, sub) {
			return e, true
		}
	}
	return Entry{}, false
}

func (c *LogCapture) AtLevel(l slog.Level) []Entry {
	var out []Entry
	for _, e := range c.Entries() {
		if e.Level == l {
			out = append(out, e)
		}
	}
	return out
}

func (c *LogCapture) RequireLogged(t testing.TB, sub string) Entry {
	t.Helper()
	e, ok := c.Find(sub)
	if !ok {
		t.Fatalf("no log entry containing %q", sub)
	}
	return e
}
