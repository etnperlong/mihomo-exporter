# mihomo-exporter

[English](README.md) | [简体中文](README.zh.md)

`mihomo-exporter` exposes Mihomo proxy latency, availability, provider health, and exporter health as Prometheus metrics. It runs beside Mihomo and serves a read-only metrics endpoint for Prometheus or Grafana Alloy.

## Quick Start

Download a release binary for your platform, then run it against Mihomo's external controller:

```bash
./mihomo-exporter --mihomo.api-url http://127.0.0.1:9090
```

Metrics are available at `http://127.0.0.1:9878/metrics`. When Mihomo uses `secret`, pass it with `--mihomo.secret` or configure it outside source control.

## Features

- Asynchronous fast topology and slow latency probing; scrapes read cached snapshots and do not call Mihomo.
- Per-proxy availability and successful latency, provider health, freshness, and exporter self-monitoring metrics.
- Layered YAML, environment, and command-line configuration with secret redaction.
- Release binaries, a non-root container image, OpenWRT `procd`, systemd, Grafana dashboard, and alert rules.

## Documentation

- [Installation and deployment](docs/installation.md)
- [Configuration reference](docs/configuration.md)
- [Metrics and scraping](docs/metrics.md)
- [Grafana dashboard and alerts](docs/dashboards.md)
- [Example configuration](config.example.yaml)

## License

MIT. See [LICENSE](LICENSE).
