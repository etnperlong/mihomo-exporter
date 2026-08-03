# Metrics and Scraping

[简体中文](metrics.zh.md)

The exporter serves only its custom collector at `/metrics` by default. Scraping reads the latest in-memory snapshot; it does not make a Mihomo API request.

## Metric Reference

| Metric | Type | Labels | Meaning |
| --- | --- | --- | --- |
| `mihomo_proxy_delay_milliseconds` | gauge | `name`, `type`, `provider` | Latest successful, fresh proxy delay in milliseconds. |
| `mihomo_proxy_available` | gauge | `name`, `type`, `provider` | `1` when a fresh probe found the proxy available; `0` when the fresh probe found it unavailable. |
| `mihomo_proxy_last_probe_timestamp_seconds` | gauge | `name`, `type`, `provider` | Unix timestamp of the latest probe result. |
| `mihomo_provider_info` | gauge | `provider`, `vehicle_type`, `type`, `test_url` | Provider metadata; value is always `1`. |
| `mihomo_provider_proxy_count` | gauge | `provider` | Number of proxies reported by the provider. |
| `mihomo_provider_proxy_available_count` | gauge | `provider` | Number of fresh, available proxy snapshots belonging to the provider. |
| `mihomo_provider_updated_timestamp_seconds` | gauge | `provider` | Provider update Unix timestamp; absent when Mihomo omits it. |
| `mihomo_provider_healthcheck_duration_seconds` | gauge | `provider` | Duration of the latest recorded provider health check. |
| `mihomo_provider_healthcheck_success` | gauge | `provider` | `1` if the latest recorded provider health check succeeded, otherwise `0`. |
| `mihomo_up` | gauge | — | `1` only when the latest fast probe completed `version`, `proxies`, and `providers`; otherwise `0`. |
| `mihomo_version_info` | gauge | `version`, `meta` | Mihomo version metadata; value is `1`. |
| `mihomo_exporter_build_info` | gauge | `version`, `revision`, `goversion` | Exporter build metadata; value is `1`. |
| `mihomo_exporter_scrape_errors_total` | counter | `collector` | Failed API or probe operations, by collector name. |
| `mihomo_exporter_probe_duration_seconds` | gauge | `loop` | Duration of the latest `fast` or `slow` probe loop. |
| `mihomo_exporter_last_success_timestamp_seconds` | gauge | `loop` | Unix timestamp of the latest successful `fast` or `slow` loop. |
| `mihomo_exporter_probe_fallback_total` | counter | — | Proxies measured through individual fallback delay probes. |
| `mihomo_exporter_unknown_proxy_type_total` | counter | `type` | Encountered outbound proxy types represented as `Unknown`. |

`metrics.enable_connections` does not currently add any metrics. No connection or traffic metric is emitted.

## Cardinality and Semantics

The proxy metrics create one series per configured proxy and `type`/`provider` combination. `name` and provider metadata therefore scale with the configured proxy set; keep that set bounded and use `probe.exclude_names` or `probe.exclude_types` for entries that should not be monitored. The exporter deliberately does not expose per-connection labels such as source host or destination.

An unavailable fresh proxy emits `mihomo_proxy_available 0` but not `mihomo_proxy_delay_milliseconds`; failed latency is never represented by a sentinel delay. `mihomo_proxy_last_probe_timestamp_seconds` can remain available to show the age of the last result.

When a proxy result is older than `probe.staleness`, the exporter stops emitting its availability and delay series. Prometheus then marks the series stale after its normal staleness handling. A provider may still expose metadata while its proxy availability series are absent. On startup, before a slow probe completes, there may be no proxy metrics at all.

Counters retain their values only while the exporter process runs; a restart resets them. `mihomo_up` represents the fast topology probe, not whether every slow provider health check or fallback delay probe succeeded.

## Prometheus

```yaml
scrape_configs:
  - job_name: mihomo-exporter
    scrape_interval: 15s
    static_configs:
      - targets:
          - 127.0.0.1:9878
```

Set `metrics_path` only when `web.telemetry_path` differs from `/metrics`. Use reachable addresses instead of `127.0.0.1:9878` for remote exporters.

## Grafana Alloy

This Alloy configuration scrapes a local exporter and forwards metrics through an existing `prometheus.remote_write.default` component:

```river
prometheus.scrape "mihomo_exporter" {
  targets = [
    { "__address__" = "127.0.0.1:9878" },
  ]
  forward_to = [prometheus.remote_write.default.receiver]
}
```

If `web.telemetry_path` is not `/metrics`, add `metrics_path = "/your-path"`. Configure the remote-write destination independently according to the target service.

For a complete Prometheus or Alloy setup that sends these metrics to Grafana Cloud, see [Grafana Cloud Integration](grafana-cloud.md).
