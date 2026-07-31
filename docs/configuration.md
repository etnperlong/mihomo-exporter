# Configuration Reference

[简体中文](configuration.zh.md)

Configuration precedence is command-line flag > environment variable > YAML file > built-in default. `--config` selects the file and takes precedence over `MIHOMO_EXPORTER_CONFIG`; it is not a YAML key. Without either, the first existing path is loaded: `./mihomo-exporter.yaml`, `/etc/mihomo-exporter/config.yaml`, then `/etc/mihomo-exporter.yaml`. No file is required. An explicitly selected missing file stops startup.

Use [config.example.yaml](../config.example.yaml) as a complete YAML reference. Durations use Go duration syntax such as `5s`, `1m`, and `5m`.

| Flag | Environment variable | YAML key | Default |
| --- | --- | --- | --- |
| `--web.listen-address` | `MIHOMO_EXPORTER_WEB_LISTEN_ADDRESS` | `web.listen_address` | `:9878` |
| `--web.telemetry-path` | `MIHOMO_EXPORTER_WEB_TELEMETRY_PATH` | `web.telemetry_path` | `/metrics` |
| `--web.shutdown-timeout` | `MIHOMO_EXPORTER_WEB_SHUTDOWN_TIMEOUT` | `web.shutdown_timeout` | `10s` |
| `--mihomo.api-url` | `MIHOMO_EXPORTER_MIHOMO_API_URL` | `mihomo.api_url` | `http://127.0.0.1:9090` |
| `--mihomo.secret` | `MIHOMO_EXPORTER_MIHOMO_SECRET` | `mihomo.secret` | empty |
| `--mihomo.timeout` | `MIHOMO_EXPORTER_MIHOMO_TIMEOUT` | `mihomo.timeout` | `10s` |
| `--probe.test-url` | `MIHOMO_EXPORTER_PROBE_TEST_URL` | `probe.test_url` | `https://www.gstatic.com/generate_204` |
| `--probe.test-timeout` | `MIHOMO_EXPORTER_PROBE_TEST_TIMEOUT` | `probe.test_timeout` | `5s` |
| `--probe.fast-interval` | `MIHOMO_EXPORTER_PROBE_FAST_INTERVAL` | `probe.fast_interval` | `15s` |
| `--probe.slow-interval` | `MIHOMO_EXPORTER_PROBE_SLOW_INTERVAL` | `probe.slow_interval` | `1m` |
| `--probe.concurrency` | `MIHOMO_EXPORTER_PROBE_CONCURRENCY` | `probe.concurrency` | `8` |
| `--probe.staleness` | `MIHOMO_EXPORTER_PROBE_STALENESS` | `probe.staleness` | `5m` |
| `--probe.exclude-types` | `MIHOMO_EXPORTER_PROBE_EXCLUDE_TYPES` | `probe.exclude_types` | empty |
| `--probe.exclude-names` | `MIHOMO_EXPORTER_PROBE_EXCLUDE_NAMES` | `probe.exclude_names` | empty |
| `--metrics.enable-connections` | `MIHOMO_EXPORTER_METRICS_ENABLE_CONNECTIONS` | `metrics.enable_connections` | `false` |
| `--log.level` | `MIHOMO_EXPORTER_LOG_LEVEL` | `log.level` | `info` |
| `--log.format` | `MIHOMO_EXPORTER_LOG_FORMAT` | `log.format` | `auto` |
| `--log.emoji` | `MIHOMO_EXPORTER_LOG_EMOJI` | `log.emoji` | `auto` |
| `--log.color` | `MIHOMO_EXPORTER_LOG_COLOR` | `log.color` | `auto` |
| `--config` | `MIHOMO_EXPORTER_CONFIG` | — | empty |
| `--version` | — | — | `false` |

`--metrics.enable-connections` is accepted and stored but is currently reserved: no connection metrics are collected or emitted.

## Validation

Startup rejects an API URL without an `http` or `https` scheme and host; non-positive shutdown, API, probe, interval, or staleness durations; `probe.slow_interval < probe.test_timeout`; `probe.staleness <= probe.slow_interval`; `probe.concurrency < 1`; and an invalid `probe.exclude_names` regular expression. The HTTP server also requires `web.telemetry_path` to begin with `/`.

## Secrets

Put `mihomo.secret` in a protected file or `MIHOMO_EXPORTER_MIHOMO_SECRET`, not in a repository, image layer, command history, or dashboard. The configuration type redacts this value as `REDACTED` when structured logging or formatting receives it; redaction does not make an exposed file safe. Restrict configuration files, for example with mode `0600`.

## Logging

`log.level` accepts Zap levels such as `debug`, `info`, `warn`, `error`, `dpanic`, `panic`, and `fatal`.

- `log.format`: `json` is machine-readable; `text` is console output; `auto` chooses `text` for a TTY and `json` otherwise.
- `log.emoji`: `on`, `off`, or `auto`. Emoji apply only to text output. `auto` requires a TTY and supported decorations; `on` is ignored for JSON output.
- `log.color`: `on`, `off`, or `auto`. Color applies only to text output. `auto` requires a TTY and supported decorations; `on` forces color for text output.
- For automatic emoji and color, decorations are disabled when `TERM` is empty or `dumb`, or when `NO_COLOR` has a non-empty value. Set `NO_COLOR=1` for portable suppression.

Use `log.format=json`, `log.emoji=off`, and `log.color=off` for log collectors. `auto` already resolves to JSON under systemd and typical container stderr because it is not a TTY.
