package testkit

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

type Call struct {
	Method string
	Path   string
	Query  string
	Header http.Header
	Body   []byte
}

type Stub struct {
	t     testing.TB
	srv   *httptest.Server
	mux   *http.ServeMux
	mu    sync.Mutex
	calls map[string][]Call
}

func NewStub(t testing.TB) *Stub {
	t.Helper()
	s := &Stub{t: t, mux: http.NewServeMux(), calls: map[string][]Call{}}
	s.srv = httptest.NewServer(http.HandlerFunc(s.serve))
	t.Cleanup(s.srv.Close)
	return s
}

func (s *Stub) URL() string { return s.srv.URL }

func (s *Stub) serve(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	r.Body = io.NopCloser(bytes.NewReader(body))
	h, pattern := s.mux.Handler(r)
	key := pattern
	if key == "" {
		key = r.Method + " " + r.URL.Path
	}
	s.mu.Lock()
	s.calls[key] = append(s.calls[key], Call{Method: r.Method, Path: r.URL.Path, Query: r.URL.RawQuery, Header: r.Header.Clone(), Body: body})
	s.mu.Unlock()
	if pattern == "" {
		http.Error(w, `{"stub":"no route for `+r.Method+" "+r.URL.Path+`"}`, http.StatusNotImplemented)
		return
	}
	h.ServeHTTP(w, r)
}

func (s *Stub) On(pattern string, status int, body any) *Stub {
	return s.Handle(pattern, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		switch b := body.(type) {
		case nil:
		case string:
			_, _ = io.WriteString(w, b)
		case []byte:
			_, _ = w.Write(b)
		default:
			_ = json.NewEncoder(w).Encode(b)
		}
	})
}

func (s *Stub) Handle(pattern string, fn http.HandlerFunc) *Stub {
	s.mux.Handle(pattern, fn)
	return s
}

func (s *Stub) Calls(pattern string) []Call {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]Call(nil), s.calls[pattern]...)
}

func (s *Stub) RequireCalled(pattern string, times int) []Call {
	s.t.Helper()
	c := s.Calls(pattern)
	if len(c) != times {
		s.t.Fatalf("%s called %d times, want %d", pattern, len(c), times)
	}
	return c
}

func (c Call) JSON(t testing.TB, v any) {
	t.Helper()
	if err := json.Unmarshal(c.Body, v); err != nil {
		t.Fatalf("decode call body: %v\n%s", err, c.Body)
	}
}
