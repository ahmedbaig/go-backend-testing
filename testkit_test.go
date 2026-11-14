package testkit

import (
	"context"
	"errors"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ahmedbaig/go-backend-config"
	"github.com/ahmedbaig/go-backend-http"
)

type item struct {
	Name string `json:"name" validate:"required"`
}

func TestClientAgainstApp(t *testing.T) {
	opts, logs := AppOptions(t, "svc")
	app := httpx.New(opts)
	app.Group("/api").Post("/items", httpx.Handle(func(ctx context.Context, in item) (item, error) {
		httpx.LoggerFrom(ctx).InfoContext(ctx, "creating", "name", in.Name)
		return in, nil
	}))
	c := NewClient(t, app.Handler()).WithHeader(httpx.HeaderCorrelationID, "corr-1")

	got := Decode[item](c.Post("/api/items", item{Name: "a"}).Expect(200))
	Equal(t, got, item{Name: "a"})

	c.Post("/api/items", `{}`).ExpectError(400, httpx.CodeValidation)
	c.Post("/api/items", `{}`).ExpectField("name", "required")
	c.Get("/missing").ExpectError(404, httpx.CodeRouteNotFound)

	e := logs.RequireLogged(t, "creating")
	Equal(t, e.Attrs["correlationId"], any("corr-1"))
	Equal(t, e.Attrs["name"], any("a"))
	True(t, len(logs.AtLevel(slog.LevelInfo)) > 0)
}

func TestStub(t *testing.T) {
	s := NewStub(t).On("GET /users/{id}", 200, map[string]string{"id": "7"})
	res, err := http.Get(s.URL() + "/users/7?x=1")
	NoError(t, err)
	res.Body.Close()
	Equal(t, res.StatusCode, 200)
	calls := s.RequireCalled("GET /users/{id}", 1)
	Equal(t, calls[0].Query, "x=1")

	res, _ = http.Get(s.URL() + "/other")
	res.Body.Close()
	Equal(t, res.StatusCode, http.StatusNotImplemented)
	s.RequireCalled("GET /other", 1)
}

func TestConfigDir(t *testing.T) {
	dir := ConfigDir(t, map[string]any{"svc": map[string]int{"port": 9}, "svc-raw": `{"port": 10}`})
	Equal(t, os.Getenv("CONFIG_DIR"), dir)
	type cfg struct {
		Port int `json:"port"`
	}
	v, err := config.Load("svc", cfg{}, config.WithLookup(func(string) (string, bool) { return "", false }))
	NoError(t, err)
	Equal(t, v.Port, 9)
	_, err = os.Stat(filepath.Join(dir, "svc-raw.json"))
	NoError(t, err)
}

func TestFixtureAndGolden(t *testing.T) {
	type f struct {
		Name string `json:"name"`
	}
	Equal(t, Fixture[f](t, "sample"), f{Name: "sample"})
	Golden(t, "sample.json", f{Name: "sample"})
}

func TestAssertHelpers(t *testing.T) {
	ErrorIs(t, errors.Join(fs.ErrNotExist), fs.ErrNotExist)
	Contains(t, "hello world", "lo w")
	NotContains(t, "hello", "bye")
	start := time.Now()
	Eventually(t, time.Second, func() bool { return time.Since(start) > 20*time.Millisecond })
}

func TestStubHandlerSeesBody(t *testing.T) {
	var got string
	s := NewStub(t).Handle("POST /echo", func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		got = string(b)
	})
	res, err := http.Post(s.URL()+"/echo", "application/json", strings.NewReader(`{"a":1}`))
	NoError(t, err)
	res.Body.Close()
	Equal(t, got, `{"a":1}`)
	Equal(t, string(s.Calls("POST /echo")[0].Body), `{"a":1}`)
}
