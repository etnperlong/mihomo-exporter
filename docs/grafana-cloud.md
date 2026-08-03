# Grafana Cloud Integration

[简体中文](grafana-cloud.zh.md)

Mihomo Exporter does not push metrics directly. Prometheus or Grafana Alloy scrapes the exporter and sends the samples to Grafana Cloud Metrics with `remote_write`:

```text
Mihomo Exporter <- scrape - Prometheus or Grafana Alloy - remote_write -> Grafana Cloud
```

## Prerequisites

Start the exporter and confirm that its metrics endpoint is reachable from the process that will scrape it:

```bash
curl -fsS http://127.0.0.1:9878/metrics
```

The response should include `mihomo_up`. Use a reachable address instead of `127.0.0.1:9878` when the scraper runs on another host. See [Configuration](configuration.md) when changing the listen address or metrics path.

In the Grafana Cloud Portal, open the Prometheus card and select **Details**. Record these values:

- Metrics instance remote-write endpoint, normally ending in `/api/prom/push`.
- Metrics instance ID, used as the basic-auth username.
- Cloud Access Policy token with the `metrics:write` scope, used as the basic-auth password.

Do not commit the token to source control. The examples below use a restricted token file for Prometheus and environment variables for Alloy.

## Prometheus

Store the Access Policy token in a file readable only by the Prometheus service account, for example `/etc/prometheus/grafana-cloud-token`, then add the following configuration to `prometheus.yml`:

```yaml
scrape_configs:
  - job_name: mihomo-exporter
    scrape_interval: 15s
    static_configs:
      - targets:
          - 127.0.0.1:9878
        labels:
          instance: router-a

remote_write:
  - url: https://prometheus-xxx.grafana.net/api/prom/push
    basic_auth:
      username: "123456"
      password_file: /etc/prometheus/grafana-cloud-token
```

Replace the endpoint, instance ID, and `router-a` with values for your deployment. Keep the `job_name` as `mihomo-exporter` to use the bundled dashboard without changes. Set `metrics_path` in the scrape job only when `web.telemetry_path` is not `/metrics`.

Use a stable, unique `instance` label when separate Prometheus processes scrape exporters through the same local address. Without it, multiple nodes using `127.0.0.1:9878` produce identical series and can cause out-of-order samples in Grafana Cloud. A central Prometheus that scrapes distinct network addresses can normally use the automatically generated `instance` labels instead.

Validate and reload Prometheus using the commands appropriate for your installation. A typical systemd installation uses:

```bash
promtool check config /etc/prometheus/prometheus.yml
sudo systemctl reload prometheus
```

## Grafana Alloy

Set the Grafana Cloud values in the environment of the Alloy service:

```text
GRAFANA_CLOUD_METRICS_URL=https://prometheus-xxx.grafana.net/api/prom/push
GRAFANA_CLOUD_METRICS_USERNAME=123456
GRAFANA_CLOUD_ACCESS_POLICY_TOKEN=replace-with-access-policy-token
```

For a packaged Linux installation, the service environment file is normally `/etc/default/alloy` on Debian or Ubuntu and `/etc/sysconfig/alloy` on RHEL, Fedora, or SUSE. Restrict access to the file because it contains the token.

Add the following components to the Alloy configuration:

```alloy
prometheus.remote_write "grafana_cloud" {
  endpoint {
    url = sys.env("GRAFANA_CLOUD_METRICS_URL")

    basic_auth {
      username = sys.env("GRAFANA_CLOUD_METRICS_USERNAME")
      password = sys.env("GRAFANA_CLOUD_ACCESS_POLICY_TOKEN")
    }
  }
}

prometheus.scrape "mihomo_exporter" {
  targets = [
    {
      "__address__" = "127.0.0.1:9878",
      "instance"    = "router-a",
    },
  ]
  job_name        = "mihomo-exporter"
  scrape_interval = "15s"
  forward_to      = [prometheus.remote_write.grafana_cloud.receiver]
}
```

Replace `router-a` with a stable name that is unique among Alloy instances. If `web.telemetry_path` is not `/metrics`, add `metrics_path = "/your-path"` to `prometheus.scrape`.

Restart Alloy after changing its service environment so it receives the new variables:

```bash
sudo systemctl restart alloy
```

When only the Alloy configuration changes, `sudo systemctl reload alloy` is sufficient for a packaged Linux installation.

## Verify the Integration

In Grafana, open **Explore**, select the Grafana Cloud Metrics data source, and run:

```promql
mihomo_up{job="mihomo-exporter"}
```

A returned series confirms that scraping and remote write are working. A value of `1` means the latest fast topology probe reached Mihomo; a value of `0` means the metric arrived but that probe failed. After data appears, follow [Grafana Dashboard and Alerts](dashboards.md) to import the bundled dashboard.

## Troubleshooting

- No local `mihomo_up`: check the exporter process, Mihomo connection, listen address, and `web.telemetry_path`.
- Target down in Prometheus or Alloy: verify that the scraper can reach the configured target and path.
- HTTP `401` or `403` from remote write: verify the endpoint, Metrics instance ID, token, and `metrics:write` scope.
- No data in Explore: expand the time range, query `mihomo_up` without selectors, and verify that the scrape job uses `mihomo-exporter`.
- Out-of-order errors: ensure that only the intended scraper collects each target and that each independently deployed scraper uses a unique `instance` label.

For current service-specific details, see the official Grafana documentation for [sending Prometheus metrics](https://grafana.com/docs/grafana-cloud/send-data/metrics/metrics-prometheus/) and the Alloy [`prometheus.remote_write`](https://grafana.com/docs/alloy/latest/reference/components/prometheus/prometheus.remote_write/) component.
