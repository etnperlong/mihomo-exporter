# HTTP Servers and APIs

Versions below were verified against the Go module proxy on 2026-07-31.
Always re-resolve with `go list -m -versions <module>` before adding one.

## Start with the standard library

Since Go 1.22, `net/http` routes methods and path parameters natively. For a
service with a modest number of endpoints, this is the correct default and adds
no dependency.

```go
mux := http.NewServeMux()
mux.HandleFunc("GET /users/{id}", handleGetUser)
mux.HandleFunc("POST /users", handleCreateUser)

srv := &http.Server{
    Addr:              ":8080",
    Handler:           mux,
    ReadHeaderTimeout: 5 * time.Second,
    ReadTimeout:       15 * time.Second,
    WriteTimeout:      15 * time.Second,
    IdleTimeout:       60 * time.Second,
}
```

Always set timeouts. A `http.Server` with no timeouts is a denial-of-service
risk. Always implement graceful shutdown driven by `signal.NotifyContext` plus
`srv.Shutdown(ctx)`.

## Routers

| Library | Version | When to choose it |
| --- | --- | --- |
| `net/http` | stdlib | **Default.** Method-based routing and path params are built in. |
| `github.com/go-chi/chi/v5` | v5.3.1 | Many routes, nested route groups, and reusable middleware chains. Compatible with `http.Handler`, so it is not a lock-in. |
| `github.com/gin-gonic/gin` | v1.12.0 | Only when the team already knows Gin and wants its binding/rendering helpers. Uses its own context type. |
| `github.com/labstack/echo/v4` | v4.15.4 | Comparable to Gin; pick one, never both. |

Prefer `net/http` or chi. Both keep handlers as plain `http.Handler`, which
keeps middleware and tests portable. Gin and Echo introduce a framework-specific
handler signature that spreads through the codebase.

## Middleware

Write middleware as `func(http.Handler) http.Handler`. Standard concerns:
request ID, structured request logging, panic recovery, timeout, CORS if a
browser calls the API, and authentication.

`chi/middleware` provides tested implementations of most of these. Do not
hand-roll a panic recovery middleware that swallows the stack trace.

## Request validation

| Library | Version | Notes |
| --- | --- | --- |
| hand-written validation | stdlib | Preferred for a small number of fields. Explicit and dependency-free. |
| `github.com/go-playground/validator/v10` | v10.30.3 | Struct-tag validation. Justified when many DTOs need consistent rules. |

Validate at the boundary, convert into a domain type, and let inner layers
assume valid input.

## Clients

Never use `http.DefaultClient` for outbound calls. Construct a client with an
explicit timeout and pass `context.Context` into every request via
`http.NewRequestWithContext`. Always close and drain the response body.

## RPC

| Library | Version | When |
| --- | --- | --- |
| `google.golang.org/grpc` | v1.83.0 | Internal service-to-service RPC with existing gRPC infrastructure. |
| `google.golang.org/protobuf` | v1.36.11 | Required alongside gRPC. |
| `connectrpc.com/connect` | v1.20.0 | Want gRPC-compatible RPC that also works over plain HTTP/1.1 and is browser-friendly. |

Do not introduce gRPC for a public browser-facing API without a gateway.

## Auth

| Library | Version | Notes |
| --- | --- | --- |
| `github.com/golang-jwt/jwt/v5` | v5.3.1 | JWT verification. Always validate the signing method explicitly; never accept `alg: none`. |
| `golang.org/x/crypto/bcrypt` | see `x/crypto` | Password hashing. Never store passwords with a plain hash. |

Read the installed `golang-security` skill before implementing an auth flow.

## Rejected by default

- A full MVC framework. Go services rarely benefit from one.
- An HTTP router that does not interoperate with `http.Handler`.
- Homemade JWT parsing or homemade password hashing.
