# mihomo-exporter

`mihomo-exporter` is a Prometheus exporter that runs alongside a Mihomo
instance. It reports outbound proxy latency and availability for collection by
Prometheus or Grafana Alloy and forwarding to Grafana Cloud.

## Development

```bash
make check
make vuln
```

The project requires Go 1.26 and Make. Full configuration, deployment, metric,
dashboard, and alerting documentation will be added with the implementation.

## License

MIT
