# Grafana Dashboard and Alerts

[简体中文](dashboards.zh.md)

## Import

In Grafana, import [../dashboards/mihomo-exporter.json](../dashboards/mihomo-exporter.json), select a Prometheus data source for `datasource`, and save the dashboard. It defaults to a six-hour range and refreshes every minute.

The visible variables are:

| Variable | Purpose |
| --- | --- |
| `datasource` | Prometheus data source used by all panels. |
| `instance` | One, many, or all exporter instances discovered from `mihomo_up`. |
| `provider` | One, many, or all providers from `mihomo_proxy_available`. |
| `proxy_type` | One, many, or all proxy types from `mihomo_proxy_available`. |

The internal `job` variable defaults to `mihomo-exporter` and scopes the exporter-online panel. Change it when the Prometheus scrape job has a different name.

## Alerts

Load [../dashboards/alerts.yaml](../dashboards/alerts.yaml) as Prometheus rule groups, or recreate the rules in Grafana Alerting. The file includes:

| Alert | Condition |
| --- | --- |
| `MihomoExporterDown` | Prometheus cannot scrape the `mihomo-exporter` job for five minutes. |
| `MihomoUnreachable` | The latest fast probe cannot reach Mihomo for five minutes. |
| `MihomoProxyDown` | A fresh proxy availability sample remains `0` for ten minutes. |
| `MihomoProviderAllDown` | A non-empty provider has zero available proxies for ten minutes. |
| `MihomoProxyMetricsAbsent` | No `mihomo_proxy_available` sample exists for ten minutes, then remains absent for five minutes. |
| `MihomoProbeStalled` | The slow loop's last success is more than five minutes old for five minutes. |

Adjust the `up{job="mihomo-exporter"}` selector if the scrape job has another name, and route alert labels and annotations for the local notification policy.

## Why Metrics Can Be Absent

`mihomo_proxy_available == 0` detects a proxy that was probed and found unavailable. It cannot detect a missing series. Availability and delay series are not emitted before the first slow probe, after a result exceeds `probe.staleness`, or when the current snapshot has no eligible proxy. A prolonged Mihomo or slow-probe failure can therefore leave no fresh availability series.

`MihomoProxyMetricsAbsent` covers that no-data condition. It is intentionally global: it has no `instance` selector, so it cannot identify which target lost metrics and it can also fire when every exporter is intentionally configured with no eligible proxies. Pair it with `MihomoExporterDown`, `MihomoUnreachable`, and the dashboard's `Stale Proxy Data` panel to identify the cause. Use a scoped copy of the rule when per-target absence alerts are required.
