# mihomo-exporter

[English](README.md) | [简体中文](README.zh.md)

`mihomo-exporter` 将 Mihomo 代理延迟、可用性、Provider 健康状态和 exporter 自身健康状态暴露为 Prometheus 指标。它与 Mihomo 同机运行，并为 Prometheus 或 Grafana Alloy 提供只读指标端点。

## 快速开始

下载适用于当前平台的 release 二进制文件，然后连接 Mihomo 的 external controller：

```bash
./mihomo-exporter --mihomo.api-url http://127.0.0.1:9090
```

指标位于 `http://127.0.0.1:9878/metrics`。如果 Mihomo 配置了 `secret`，请通过 `--mihomo.secret` 传入，或在版本控制之外配置。

## 功能特性

- 异步执行快速拓扑探测和慢速延迟探测；抓取仅读取缓存快照，不会调用 Mihomo。
- 提供每个代理的可用性和成功延迟、Provider 健康状态、新鲜度及 exporter 自监控指标。
- 支持 YAML、环境变量和命令行的分层配置，并会脱敏 secret。
- 提供 release 二进制、非 root 容器镜像、OpenWRT `procd`、systemd、Grafana dashboard 和告警规则。

## 文档

- [安装与部署](docs/installation.zh.md)
- [配置参考](docs/configuration.zh.md)
- [指标与抓取](docs/metrics.zh.md)
- [Grafana dashboard 与告警](docs/dashboards.zh.md)
- [配置示例](config.example.yaml)

## 许可证

MIT。参见 [LICENSE](LICENSE)。
