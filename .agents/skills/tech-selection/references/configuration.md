# Configuration and Secrets

Versions verified against the Go module proxy on 2026-07-31.
Re-resolve with `go list -m -versions <module>` before adding one.

## Principles

1. **Environment variables are the default source.** They work identically in
   local shells, containers, and CI.
2. **Load configuration once, at startup, in `main`.** Parse into a typed
   struct and pass it down. Never read `os.Getenv` from deep inside business
   logic — it makes code untestable and hides dependencies.
3. **Fail fast.** Validate all required configuration during startup and exit
   with a clear message naming the missing key. Never start half-configured.
4. **Never commit secrets.** No credentials in `go` files, YAML, or the repo.
5. **No implicit `.env` loading in production.** A `.env` file is a local
   developer convenience only.

## Approaches

| Approach | Version | When to choose it |
| --- | --- | --- |
| `os.Getenv` + a typed struct | stdlib | **Default.** A dozen settings need no library. |
| `github.com/caarlos0/env/v11` | v11.4.1 | Struct-tag environment parsing with defaults and required fields. Small and focused. |
| `github.com/knadh/koanf/v2` | v2.3.5 | Need layered config (file, env, flags) without Viper's dependency weight. |
| `github.com/spf13/viper` | v1.21.0 | Already using Cobra and need config files, live reload, and remote providers. Heavy. |
| `github.com/joho/godotenv` | v1.5.1 | Local `.env` loading for development only. |

Prefer the standard library. Reach for `caarlos0/env` when the struct is large
enough that hand-written parsing becomes repetitive. Adopt Viper only when you
genuinely need multi-source configuration; read `golang-spf13-viper` first.

## Recommended shape

```go
type Config struct {
    Addr        string
    DatabaseURL string
    LogLevel    slog.Level
    Timeout     time.Duration
}

func Load() (Config, error) {
    cfg := Config{
        Addr:     envOr("ADDR", ":8080"),
        LogLevel: slog.LevelInfo,
        Timeout:  15 * time.Second,
    }

    cfg.DatabaseURL = os.Getenv("DATABASE_URL")
    if cfg.DatabaseURL == "" {
        return Config{}, errors.New("DATABASE_URL is required")
    }

    return cfg, nil
}
```

Return the error; do not `log.Fatal` inside `Load`.

## Secrets

- Read secrets from the environment or the platform's secret manager.
- Keep secrets out of logs. Never log a whole config struct that contains a
  password, token, or connection string. Implement a redacting `String()` or
  `LogValue()` method on types that hold secrets.
- Prefer a short-lived credential over a long-lived one where the platform
  supports it.
- Provide a documented `.env.example` with placeholder values, and ensure the
  real `.env` is gitignored.

## Rejected by default

- Reading configuration through a global singleton.
- Scattering `os.Getenv` calls across packages.
- Committing any real credential, even to a private repository.
- Loading `.env` automatically when the process runs in production.
