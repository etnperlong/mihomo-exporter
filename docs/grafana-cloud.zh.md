# Grafana Cloud 集成

[English](grafana-cloud.md)

Mihomo Exporter 不会直接推送指标。Prometheus 或 Grafana Alloy 抓取 exporter，再通过 `remote_write` 将 sample 发送到 Grafana Cloud Metrics：

```text
Mihomo Exporter <- scrape - Prometheus 或 Grafana Alloy - remote_write -> Grafana Cloud
```

## 前置条件

启动 exporter，并确认负责抓取指标的进程能够访问其指标端点：

```bash
curl -fsS http://127.0.0.1:9878/metrics
```

响应中应包含 `mihomo_up`。抓取进程位于其他主机时，请使用可访问的地址，而不是 `127.0.0.1:9878`。如需修改监听地址或指标路径，请参阅[配置参考](configuration.zh.md)。

在 Grafana Cloud Portal 中打开 Prometheus 卡片并选择 **Details**，记录以下信息：

- Metrics instance remote-write endpoint，通常以 `/api/prom/push` 结尾。
- Metrics instance ID，作为 basic-auth username。
- 具有 `metrics:write` scope 的 Cloud Access Policy token，作为 basic-auth password。

不要将 token 提交到版本控制。以下 Prometheus 示例使用受限 token 文件，Alloy 示例使用环境变量。

## Prometheus

将 Access Policy token 保存在仅 Prometheus 服务账户可读的文件中，例如 `/etc/prometheus/grafana-cloud-token`，然后在 `prometheus.yml` 中加入以下配置：

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

请将 endpoint、instance ID 和 `router-a` 替换为实际值。保留 `mihomo-exporter` 作为 `job_name`，即可直接使用仓库提供的 dashboard。仅当 `web.telemetry_path` 不是 `/metrics` 时，才在 scrape job 中设置 `metrics_path`。

当多个独立 Prometheus 进程都通过相同本地地址抓取 exporter 时，请设置稳定且唯一的 `instance` label。否则，多个节点使用 `127.0.0.1:9878` 会产生完全相同的 series，并可能在 Grafana Cloud 中导致 out-of-order sample。由一个中央 Prometheus 抓取不同网络地址时，通常可以直接使用自动生成的 `instance` label。

使用适合当前安装方式的命令验证并重新加载 Prometheus。典型的 systemd 安装可以执行：

```bash
promtool check config /etc/prometheus/prometheus.yml
sudo systemctl reload prometheus
```

## Grafana Alloy

在 Alloy 服务环境中设置 Grafana Cloud 参数：

```text
GRAFANA_CLOUD_METRICS_URL=https://prometheus-xxx.grafana.net/api/prom/push
GRAFANA_CLOUD_METRICS_USERNAME=123456
GRAFANA_CLOUD_ACCESS_POLICY_TOKEN=replace-with-access-policy-token
```

使用 Linux 软件包安装时，服务环境文件在 Debian 或 Ubuntu 上通常为 `/etc/default/alloy`，在 RHEL、Fedora 或 SUSE 上通常为 `/etc/sysconfig/alloy`。该文件包含 token，请限制其访问权限。

在 Alloy 配置中加入以下 component：

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

请将 `router-a` 替换为在所有 Alloy instance 中唯一且稳定的名称。如果 `web.telemetry_path` 不是 `/metrics`，请在 `prometheus.scrape` 中添加 `metrics_path = "/your-path"`。

修改服务环境后，需要重启 Alloy 以读取新的环境变量：

```bash
sudo systemctl restart alloy
```

如果只修改 Alloy 配置，使用 Linux 软件包安装时执行 `sudo systemctl reload alloy` 即可。

## 验证集成

在 Grafana 中打开 **Explore**，选择 Grafana Cloud Metrics data source，然后执行：

```promql
mihomo_up{job="mihomo-exporter"}
```

查询返回 series 表示抓取和 remote write 已正常工作。值为 `1` 表示最近一次快速拓扑 probe 成功连接 Mihomo；值为 `0` 表示指标已经到达，但该 probe 失败。确认数据出现后，请按照 [Grafana dashboard 与告警](dashboards.zh.md) 导入仓库提供的 dashboard。

## 故障排查

- 本地没有 `mihomo_up`：检查 exporter 进程、Mihomo 连接、监听地址和 `web.telemetry_path`。
- Prometheus 或 Alloy target down：确认抓取进程能够访问配置的 target 和路径。
- Remote write 返回 HTTP `401` 或 `403`：检查 endpoint、Metrics instance ID、token 和 `metrics:write` scope。
- Explore 中没有数据：扩大时间范围，去掉 selector 查询 `mihomo_up`，并确认 scrape job 使用 `mihomo-exporter`。
- 出现 out-of-order 错误：确认每个 target 仅由预期的抓取进程采集，并为每个独立部署的抓取进程设置唯一的 `instance` label。

服务配置的最新细节请参阅 Grafana 官方文档：[发送 Prometheus 指标](https://grafana.com/docs/grafana-cloud/send-data/metrics/metrics-prometheus/) 和 Alloy [`prometheus.remote_write`](https://grafana.com/docs/alloy/latest/reference/components/prometheus/prometheus.remote_write/) component。
