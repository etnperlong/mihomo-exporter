# Mihomo Exporter Grafana Dashboard

Import `mihomo-exporter.json` in Grafana, then select the Prometheus data source
in the `datasource` variable. Use `instance`, `provider`, and `proxy_type` to
scope every panel; `instance` supports selecting multiple exporters or all of
them.

Load `alerts.yaml` as a Prometheus rule group or paste its rules into Grafana
Cloud alerting. `MihomoProxyMetricsAbsent` is necessary because a failed proxy
collection can make every `mihomo_proxy_available` series disappear. A normal
`mihomo_proxy_available == 0` alert cannot fire after those series become stale.
