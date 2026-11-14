# go-backend-testing

Test helpers for services built on the `go-backend-*` family. Import as `testkit`. No assertion framework is pulled in; everything works with the standard `testing` package.

```sh
go get github.com/ahmedbaig/go-backend-testing
```

## Isolated config

```go
testkit.ConfigDir(t, map[string]any{
	"my-app":        map[string]any{"port": 0, "environment": "testing"},
	"my-app-logger": `{"logToFile": false}`,
})
```

Writes each entry as `<name>.json` in a temporary directory and points `CONFIG_DIR` at it for the length of the test, so a test can never read real config or secrets from the machine.

## Testing an app

```go
opts, logs := testkit.AppOptions(t, "my-app")
app := httpx.New(opts)
routes.Register(app, deps)

c := testkit.NewClient(t, app.Handler()).WithHeader("Authorization", "Bearer test")

item := testkit.Decode[Item](c.Post("/api/v1/items", Item{Name: "a"}).Expect(201))
c.Post("/api/v1/items", `{}`).ExpectError(400, httpx.CodeValidation)
c.Post("/api/v1/items", `{}`).ExpectField("name", "required")

logs.RequireLogged(t, "item created")
```

`AppOptions` gives an `httpx.Options` with a capturing logger, `testing` environment and probes that don't run on a schedule. Requests go straight to the handler (no network). Captured log lines are printed when a test fails.

## Stubbing downstream services

```go
users := testkit.NewStub(t).
	On("GET /api/v1/users/{id}", 200, User{ID: "7"}).
	On("POST /api/v1/users", 409, `{"errors":[{"code":"E409"}]}`)

svc := NewService(users.URL())
...
calls := users.RequireCalled("GET /api/v1/users/{id}", 1)
```

A real `httptest` server using `ServeMux` patterns. Every request is recorded (method, path, query, headers, body). Unknown routes return 501, so a missing stub fails loudly.

## Fixtures and golden files

```go
order := testkit.Fixture[Order](t, "order-paid")        // testdata/fixtures/order-paid.json
testkit.Golden(t, "invoice.json", renderInvoice(order))   // testdata/golden/invoice.json
```

Run `go test ./... -update` to rewrite golden files.

## Assertions

`NoError`, `ErrorIs`, `Equal` (generic, deep), `True`, `Contains`, `NotContains`, `Eventually(t, timeout, cond)`.

## License

MIT
