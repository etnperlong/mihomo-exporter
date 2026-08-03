# 指标与抓取

[English](metrics.md)

exporter 默认仅在 `/metrics` 暴露自定义 collector。抓取仅读取最新内存快照，不会发起 Mihomo API 请求。

## 指标参考

| Metric | 类型 | Labels | 含义 |
| --- | --- | --- | --- |
| `mihomo_proxy_delay_milliseconds` | gauge | `name`, `type`, `provider` | 最新成功且未过期的代理延迟，单位为毫秒。 |
| `mihomo_proxy_available` | gauge | `name`, `type`, `provider` | 新鲜探测判定代理可用时为 `1`；新鲜探测判定不可用时为 `0`。 |
| `mihomo_proxy_last_probe_timestamp_seconds` | gauge | `name`, `type`, `provider` | 最近探测结果的 Unix 时间戳。 |
| `mihomo_provider_info` | gauge | `provider`, `vehicle_type`, `type`, `test_url` | Provider 元数据；值恒为 `1`。 |
| `mihomo_provider_proxy_count` | gauge | `provider` | Provider 报告的代理数量。 |
| `mihomo_provider_proxy_available_count` | gauge | `provider` | 属于该 Provider 的新鲜且可用代理快照数量。 |
| `mihomo_provider_updated_timestamp_seconds` | gauge | `provider` | Provider 更新的 Unix 时间戳；Mihomo 未提供时不存在。 |
| `mihomo_provider_healthcheck_duration_seconds` | gauge | `provider` | 最近一次记录的 Provider health check 耗时。 |
| `mihomo_provider_healthcheck_success` | gauge | `provider` | 最近一次记录的 Provider health check 成功时为 `1`，否则为 `0`。 |
| `mihomo_up` | gauge | — | 仅当最近一次快速探测完成 `version`、`proxies` 和 `providers` 时为 `1`；否则为 `0`。 |
| `mihomo_version_info` | gauge | `version`, `meta` | Mihomo 版本元数据；值为 `1`。 |
| `mihomo_exporter_build_info` | gauge | `version`, `revision`, `goversion` | Exporter 构建元数据；值为 `1`。 |
| `mihomo_exporter_scrape_errors_total` | counter | `collector` | 按 collector 名称区分的失败 API 或 probe 操作。 |
| `mihomo_exporter_probe_duration_seconds` | gauge | `loop` | 最近 `fast` 或 `slow` probe loop 的耗时。 |
| `mihomo_exporter_last_success_timestamp_seconds` | gauge | `loop` | 最近成功 `fast` 或 `slow` loop 的 Unix 时间戳。 |
| `mihomo_exporter_probe_fallback_total` | counter | — | 通过单个 fallback delay probe 测量的代理数。 |
| `mihomo_exporter_unknown_proxy_type_total` | counter | `type` | 遇到并表示为 `Unknown` 的出站代理类型。 |

`metrics.enable_connections` 目前不会添加任何指标。不会暴露连接或流量指标。

## 基数与语义

代理指标会为每个已配置代理以及其 `type`/`provider` 组合创建一个 series。因此 `name` 和 Provider 元数据会随已配置代理集增长；请控制该集合规模，并使用 `probe.exclude_names` 或 `probe.exclude_types` 排除无需监控的条目。exporter 特意不暴露带有 source host 或 destination 等单连接 label 的指标。

不可用但仍新鲜的代理会暴露 `mihomo_proxy_available 0`，但不会暴露 `mihomo_proxy_delay_milliseconds`；失败延迟绝不会用 sentinel delay 表示。`mihomo_proxy_last_probe_timestamp_seconds` 可以继续存在，用于展示最后结果的年龄。

代理结果早于 `probe.staleness` 时，exporter 会停止暴露其 availability 和 delay series。之后 Prometheus 按正常 staleness 处理将该 series 标记为 stale。即使代理 availability series 缺失，Provider 仍可能暴露元数据。启动后、第一次 slow probe 完成前，可能完全没有代理指标。

Counter 仅在 exporter 进程运行期间保留数值；重启会重置它们。`mihomo_up` 表示快速拓扑 probe，而不是每个慢速 Provider health check 或 fallback delay probe 是否成功。

## Prometheus 抓取配置

```yaml
scrape_configs:
  - job_name: mihomo-exporter
    scrape_interval: 15s
    static_configs:
      - targets:
          - 127.0.0.1:9878
```

仅当 `web.telemetry_path` 不同于 `/metrics` 时才设置 `metrics_path`。远程 exporter 请使用可访问的地址而非 `127.0.0.1:9878`。

## Grafana Alloy 抓取配置

以下 Alloy 配置抓取本地 exporter，并通过现有的 `prometheus.remote_write.default` component 转发指标：

```river
prometheus.scrape "mihomo_exporter" {
  targets = [
    { "__address__" = "127.0.0.1:9878" },
  ]
  forward_to = [prometheus.remote_write.default.receiver]
}
```

如果 `web.telemetry_path` 不是 `/metrics`，请添加 `metrics_path = "/your-path"`。请根据目标服务独立配置 remote-write destination。

如需配置 Prometheus 或 Alloy 并将指标发送到 Grafana Cloud，请参阅 [Grafana Cloud 集成](grafana-cloud.zh.md)。
