# Grafana Dashboard 与告警

[English](dashboards.md)

## Import

在 Grafana 中导入 [../dashboards/mihomo-exporter.json](../dashboards/mihomo-exporter.json)，为 `datasource` 选择 Prometheus data source，然后保存 dashboard。默认时间范围为六小时，每分钟刷新一次。

可见变量如下：

| 变量 | 用途 |
| --- | --- |
| `datasource` | 所有 panel 使用的 Prometheus data source。 |
| `instance` | 从 `mihomo_up` 发现的一个、多个或全部 exporter instance。 |
| `provider` | 来自 `mihomo_proxy_available` 的一个、多个或全部 Provider。 |
| `proxy_type` | 来自 `mihomo_proxy_available` 的一个、多个或全部代理类型。 |

内部变量 `job` 默认值为 `mihomo-exporter`，用于限定 exporter-online panel。当 Prometheus scrape job 使用其他名称时，请修改它。

## Alerts

将 [../dashboards/alerts.yaml](../dashboards/alerts.yaml) 作为 Prometheus rule groups 加载，或在 Grafana Alerting 中重建这些规则。文件包含：

| 告警 | 条件 |
| --- | --- |
| `MihomoExporterDown` | Prometheus 连续五分钟无法抓取 `mihomo-exporter` job。 |
| `MihomoUnreachable` | 最近快速 probe 连续五分钟无法连接 Mihomo。 |
| `MihomoProxyDown` | 新鲜代理 availability sample 连续十分钟为 `0`。 |
| `MihomoProviderAllDown` | 非空 Provider 连续十分钟没有可用代理。 |
| `MihomoProxyMetricsAbsent` | 连续十分钟没有 `mihomo_proxy_available` sample，之后仍持续缺失五分钟。 |
| `MihomoProbeStalled` | slow loop 最近一次成功已超过五分钟，并持续五分钟。 |

当 scrape job 使用其他名称时，请调整 `up{job="mihomo-exporter"}` selector，并根据本地 notification policy 路由告警 label 和 annotation。

## Why Metrics Can Be Absent

`mihomo_proxy_available == 0` 用于检测已 probe 且判定不可用的代理，无法检测缺失的 series。第一次 slow probe 前、结果超过 `probe.staleness` 后，或当前 snapshot 没有符合条件的代理时，不会暴露 availability 和 delay series。因此，长期 Mihomo 或 slow-probe 故障可能导致没有新鲜 availability series。

`MihomoProxyMetricsAbsent` 覆盖该 no-data 条件。它有意使用全局范围：没有 `instance` selector，因此无法识别哪个 target 丢失了指标；当所有 exporter 都被有意配置为没有符合条件的代理时，它也可能触发。请结合 `MihomoExporterDown`、`MihomoUnreachable` 及 dashboard 中的 `Stale Proxy Data` panel 判断原因。需要按 target 告警时，请使用该规则的 scoped 副本。
