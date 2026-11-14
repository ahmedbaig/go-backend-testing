package testkit

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ahmedbaig/go-backend-http"
)

type Client struct {
	t       testing.TB
	handler http.Handler
	headers http.Header
}

func NewClient(t testing.TB, h http.Handler) *Client {
	return &Client{t: t, handler: h, headers: http.Header{}}
}

func (c *Client) WithHeader(k, v string) *Client {
	n := &Client{t: c.t, handler: c.handler, headers: c.headers.Clone()}
	n.headers.Set(k, v)
	return n
}

type Response struct {
	t      testing.TB
	Status int
	Header http.Header
	Body   []byte
}

func (c *Client) Do(method, target string, body any) *Response {
	c.t.Helper()
	var rdr io.Reader
	switch b := body.(type) {
	case nil:
	case string:
		rdr = strings.NewReader(b)
	case []byte:
		rdr = bytes.NewReader(b)
	default:
		buf, err := json.Marshal(b)
		if err != nil {
			c.t.Fatalf("encode body: %v", err)
		}
		rdr = bytes.NewReader(buf)
	}
	req := httptest.NewRequest(method, target, rdr)
	if rdr != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, vs := range c.headers {
		req.Header[k] = vs
	}
	rec := httptest.NewRecorder()
	c.handler.ServeHTTP(rec, req)
	return &Response{t: c.t, Status: rec.Code, Header: rec.Header(), Body: rec.Body.Bytes()}
}

func (c *Client) Get(target string) *Response { c.t.Helper(); return c.Do(http.MethodGet, target, nil) }

func (c *Client) Post(target string, body any) *Response {
	c.t.Helper()
	return c.Do(http.MethodPost, target, body)
}

func (c *Client) Put(target string, body any) *Response {
	c.t.Helper()
	return c.Do(http.MethodPut, target, body)
}

func (c *Client) Patch(target string, body any) *Response {
	c.t.Helper()
	return c.Do(http.MethodPatch, target, body)
}

func (c *Client) Delete(target string) *Response {
	c.t.Helper()
	return c.Do(http.MethodDelete, target, nil)
}

func (r *Response) Expect(status int) *Response {
	r.t.Helper()
	if r.Status != status {
		r.t.Fatalf("status %d, want %d\n%s", r.Status, status, r.Body)
	}
	return r
}

func (r *Response) JSON(v any) *Response {
	r.t.Helper()
	if err := json.Unmarshal(r.Body, v); err != nil {
		r.t.Fatalf("decode response: %v\n%s", err, r.Body)
	}
	return r
}

func Decode[T any](r *Response) T {
	r.t.Helper()
	var v T
	r.JSON(&v)
	return v
}

func (r *Response) Envelope() httpx.Envelope {
	r.t.Helper()
	var e httpx.Envelope
	r.JSON(&e)
	if len(e.Errors) == 0 {
		r.t.Fatalf("response is not an error envelope: %s", r.Body)
	}
	return e
}

func (r *Response) ExpectError(status int, code httpx.Code) httpx.Envelope {
	r.t.Helper()
	r.Expect(status)
	e := r.Envelope()
	if e.Errors[0].Code != code {
		r.t.Fatalf("error code %s, want %s\n%s", e.Errors[0].Code, code, r.Body)
	}
	return e
}

func (r *Response) ExpectField(field, rule string) *Response {
	r.t.Helper()
	for _, f := range r.Envelope().Errors[0].Fields {
		if f.Field == field && (rule == "" || f.Rule == rule) {
			return r
		}
	}
	r.t.Fatalf("no field error %s/%s in %s", field, rule, r.Body)
	return r
}
