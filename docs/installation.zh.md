# 安装与部署

[English](installation.md)

## 发布二进制文件

Release 压缩包发布在 [GitHub Releases](https://github.com/etnperlong/mihomo-exporter/releases)。每个压缩包包含 `mihomo-exporter` 和 `LICENSE`；安装前请使用随附的 `checksums.txt` 验证。

```bash
VERSION=vX.Y.Z
ARCH=amd64
curl -LO "https://github.com/etnperlong/mihomo-exporter/releases/download/${VERSION}/mihomo-exporter_${VERSION}_linux_${ARCH}.tar.gz"
curl -LO "https://github.com/etnperlong/mihomo-exporter/releases/download/${VERSION}/checksums.txt"
sha256sum -c checksums.txt --ignore-missing
tar -xzf "mihomo-exporter_${VERSION}_linux_${ARCH}.tar.gz"
./mihomo-exporter --version
```

二进制压缩包支持 `linux` 和 `darwin` 的 `amd64`、`arm64`。Linux 还提供 `mipsle_softfloat` 和 `mipsle_hardfloat` 压缩包；请选择与路由器固件匹配的 ABI。没有 Darwin `mipsle` 压缩包。

## Docker

Release tag 会发布 `linux/amd64` 和 `linux/arm64` 的 `ghcr.io/etnperlong/mihomo-exporter:<version>` 与 `ghcr.io/etnperlong/mihomo-exporter:latest`。运行时镜像以非 root 用户运行。请使用配置文件和能够让容器访问 Mihomo 的网络：

```bash
docker run --rm --network host \
  -v "$PWD/config.yaml:/etc/mihomo-exporter/config.yaml:ro" \
  ghcr.io/etnperlong/mihomo-exporter:latest
```

`--network host` 是针对 Mihomo controller 监听在 `127.0.0.1` 的 Linux 示例。使用其他网络时，请将 `mihomo.api_url` 设为容器可访问的地址。普通的 `docker build .` 使用 Dockerfile 的 `local` stage 构建本地镜像。

## OpenWRT

将适用于当前 Linux 平台的 release 二进制安装到 `/usr/bin/mihomo-exporter`，然后安装随附的 `procd` 脚本和配置：

```sh
install -m 0755 mihomo-exporter /usr/bin/mihomo-exporter
install -m 0755 contrib/openwrt/mihomo-exporter.init /etc/init.d/mihomo-exporter
install -d /etc/mihomo-exporter
install -m 0644 config.example.yaml /etc/mihomo-exporter/config.yaml
/etc/init.d/mihomo-exporter enable
/etc/init.d/mihomo-exporter start
```

init 脚本以 `nobody` 运行，自动发现 `/etc/mihomo-exporter/config.yaml`，并将 stdout 和 stderr 交给 `procd`。示例权限使该账户可读取文件。不要以此方式使用真实 `mihomo.secret`：应先调整服务，使其以受限文件访问权限的专用账户运行。在 `mipsle` 上，请为固件 ABI 选择 `softfloat` 或 `hardfloat` 压缩包。

## systemd

安装二进制、配置及随附 unit 后，重新加载 systemd：

```bash
sudo install -m 0755 mihomo-exporter /usr/local/bin/mihomo-exporter
sudo install -d /etc/mihomo-exporter
sudo install -m 0644 config.example.yaml /etc/mihomo-exporter/config.yaml
sudo install -m 0644 contrib/systemd/mihomo-exporter.service /etc/systemd/system/mihomo-exporter.service
sudo systemctl daemon-reload
sudo systemctl enable --now mihomo-exporter
```

该 unit 使用 `DynamicUser=true` 和文件系统加固，启动 `/usr/local/bin/mihomo-exporter --config /etc/mihomo-exporter/config.yaml`。示例权限使动态服务用户可读取文件。不要原样在此配置中放入真实 `mihomo.secret`：应先调整 unit，使其使用专用服务账户和受限文件权限。
