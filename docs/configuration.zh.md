# 配置参考

[English](configuration.md)

配置优先级为命令行 flag > 环境变量 > YAML 文件 > 内置默认值。`--config` 用于选择文件，且优先于 `MIHOMO_EXPORTER_CONFIG`；它不是 YAML key。两者都未设置时，会按顺序加载第一个存在的路径：`./mihomo-exporter.yaml`、`/etc/mihomo-exporter/config.yaml`、`/etc/mihomo-exporter.yaml`。配置文件不是必需的。显式选择但不存在的文件会导致启动失败。

请使用 [config.example.yaml](../config.example.yaml) 作为完整 YAML 参考。Duration 使用 Go duration 语法，例如 `5s`、`1m` 和 `5m`。

| Flag | 环境变量 | YAML key | 默认值 |
| --- | --- | --- | --- |
| `--web.listen-address` | `MIHOMO_EXPORTER_WEB_LISTEN_ADDRESS` | `web.listen_address` | `:9878` |
| `--web.telemetry-path` | `MIHOMO_EXPORTER_WEB_TELEMETRY_PATH` | `web.telemetry_path` | `/metrics` |
| `--web.shutdown-timeout` | `MIHOMO_EXPORTER_WEB_SHUTDOWN_TIMEOUT` | `web.shutdown_timeout` | `10s` |
| `--mihomo.api-url` | `MIHOMO_EXPORTER_MIHOMO_API_URL` | `mihomo.api_url` | `http://127.0.0.1:9090` |
| `--mihomo.secret` | `MIHOMO_EXPORTER_MIHOMO_SECRET` | `mihomo.secret` | 空 |
| `--mihomo.timeout` | `MIHOMO_EXPORTER_MIHOMO_TIMEOUT` | `mihomo.timeout` | `10s` |
| `--probe.test-url` | `MIHOMO_EXPORTER_PROBE_TEST_URL` | `probe.test_url` | `https://www.gstatic.com/generate_204` |
| `--probe.test-timeout` | `MIHOMO_EXPORTER_PROBE_TEST_TIMEOUT` | `probe.test_timeout` | `5s` |
| `--probe.fast-interval` | `MIHOMO_EXPORTER_PROBE_FAST_INTERVAL` | `probe.fast_interval` | `15s` |
| `--probe.slow-interval` | `MIHOMO_EXPORTER_PROBE_SLOW_INTERVAL` | `probe.slow_interval` | `1m` |
| `--probe.concurrency` | `MIHOMO_EXPORTER_PROBE_CONCURRENCY` | `probe.concurrency` | `8` |
| `--probe.staleness` | `MIHOMO_EXPORTER_PROBE_STALENESS` | `probe.staleness` | `5m` |
| `--probe.exclude-types` | `MIHOMO_EXPORTER_PROBE_EXCLUDE_TYPES` | `probe.exclude_types` | 空 |
| `--probe.exclude-names` | `MIHOMO_EXPORTER_PROBE_EXCLUDE_NAMES` | `probe.exclude_names` | 空 |
| `--metrics.enable-connections` | `MIHOMO_EXPORTER_METRICS_ENABLE_CONNECTIONS` | `metrics.enable_connections` | `false` |
| `--log.level` | `MIHOMO_EXPORTER_LOG_LEVEL` | `log.level` | `info` |
| `--log.format` | `MIHOMO_EXPORTER_LOG_FORMAT` | `log.format` | `auto` |
| `--log.emoji` | `MIHOMO_EXPORTER_LOG_EMOJI` | `log.emoji` | `auto` |
| `--log.color` | `MIHOMO_EXPORTER_LOG_COLOR` | `log.color` | `auto` |
| `--config` | `MIHOMO_EXPORTER_CONFIG` | — | 空 |
| `--version` | — | — | `false` |

`--metrics.enable-connections` 可以接受并保存，但目前是预留项：不会采集或暴露连接指标。

## Validation

启动会拒绝以下配置：没有 `http` 或 `https` scheme 及 host 的 API URL；非正数的 shutdown、API、probe、interval 或 staleness duration；`probe.slow_interval < probe.test_timeout`；`probe.staleness <= probe.slow_interval`；`probe.concurrency < 1`；以及无效的 `probe.exclude_names` 正则表达式。HTTP server 还要求 `web.telemetry_path` 以 `/` 开头。

## Secrets

将 `mihomo.secret` 放在受保护的文件或 `MIHOMO_EXPORTER_MIHOMO_SECRET` 中，不要放入仓库、镜像层、命令历史或 dashboard。配置类型在结构化日志或格式化时会将该值脱敏为 `REDACTED`；脱敏不能使已泄露的文件变得安全。请限制配置文件权限，例如设为 `0600`。

## Logging

`log.level` 接受 Zap level，例如 `debug`、`info`、`warn`、`error`、`dpanic`、`panic` 和 `fatal`。

- `log.format`：`json` 为机器可读格式；`text` 为控制台输出；`auto` 在 TTY 上选择 `text`，否则选择 `json`。
- `log.emoji`：`on`、`off` 或 `auto`。Emoji 仅用于 text 输出。`auto` 需要 TTY 和受支持的装饰；在 JSON 输出时会忽略 `on`。
- `log.color`：`on`、`off` 或 `auto`。Color 仅用于 text 输出。`auto` 需要 TTY 和受支持的装饰；`on` 会强制为 text 输出着色。
- 自动 emoji 和 color 在 `TERM` 为空或为 `dumb` 时禁用；`NO_COLOR` 为非空值时也会禁用。请设置 `NO_COLOR=1` 以便跨环境抑制装饰。

日志采集器应使用 `log.format=json`、`log.emoji=off` 和 `log.color=off`。在 systemd 和典型容器 stderr 中，`auto` 已会因非 TTY 而解析为 JSON。
