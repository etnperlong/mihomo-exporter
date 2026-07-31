# Installation and Deployment

[简体中文](installation.zh.md)

## Release Binary

Release archives are published at [GitHub Releases](https://github.com/etnperlong/mihomo-exporter/releases). Each archive contains `mihomo-exporter` and `LICENSE`; verify it against the accompanying `checksums.txt` before installation.

```bash
VERSION=vX.Y.Z
ARCH=amd64
curl -LO "https://github.com/etnperlong/mihomo-exporter/releases/download/${VERSION}/mihomo-exporter_${VERSION}_linux_${ARCH}.tar.gz"
curl -LO "https://github.com/etnperlong/mihomo-exporter/releases/download/${VERSION}/checksums.txt"
sha256sum -c checksums.txt --ignore-missing
tar -xzf "mihomo-exporter_${VERSION}_linux_${ARCH}.tar.gz"
./mihomo-exporter --version
```

Binary archives target `linux` and `darwin` on `amd64` and `arm64`. Linux additionally has `mipsle_softfloat` and `mipsle_hardfloat` archives; choose the ABI matching the router firmware. There is no Darwin `mipsle` archive.

## Docker

Release tags publish `ghcr.io/etnperlong/mihomo-exporter:<version>` and `ghcr.io/etnperlong/mihomo-exporter:latest` for `linux/amd64` and `linux/arm64`. The runtime image is non-root. Use a configuration file and networking that lets the container reach Mihomo:

```bash
docker run --rm --network host \
  -v "$PWD/config.yaml:/etc/mihomo-exporter/config.yaml:ro" \
  ghcr.io/etnperlong/mihomo-exporter:latest
```

`--network host` is a Linux example for a Mihomo controller listening on `127.0.0.1`. With other networking, set `mihomo.api_url` to an address reachable from the container. A plain `docker build .` produces a local image using the Dockerfile's `local` stage.

### Docker Compose

[`docker-compose.yml`](../docker-compose.yml) defines both the published GHCR image and a local multi-stage build. Its default `pull_policy: always` makes `docker compose up` pull `ghcr.io/etnperlong/mihomo-exporter:${MIHOMO_EXPORTER_TAG:-latest}` before starting, so deployment prefers the released image:

```bash
cp config.example.yaml config.yaml
docker compose up -d
```

Set `MIHOMO_EXPORTER_TAG` to pin a release, for example `MIHOMO_EXPORTER_TAG=v1.2.3 docker compose up -d`. To build from the checked-out source instead, explicitly build and suppress the pull for that start:

```bash
docker compose build
docker compose up -d --pull never
```

The Compose service uses host networking and therefore targets Linux hosts where Mihomo listens on `127.0.0.1`. Override `MIHOMO_EXPORTER_CONFIG` when the local configuration has a different path.

## OpenWRT

Install the appropriate Linux release binary at `/usr/bin/mihomo-exporter`, then install the supplied `procd` script and configuration:

```sh
install -m 0755 mihomo-exporter /usr/bin/mihomo-exporter
install -m 0755 contrib/openwrt/mihomo-exporter.init /etc/init.d/mihomo-exporter
install -d /etc/mihomo-exporter
install -m 0644 config.example.yaml /etc/mihomo-exporter/config.yaml
/etc/init.d/mihomo-exporter enable
/etc/init.d/mihomo-exporter start
```

The init script runs as `nobody`, discovers `/etc/mihomo-exporter/config.yaml`, and sends stdout and stderr to `procd`. The shown mode makes the file readable to that account. Do not use it for a real `mihomo.secret`: first adapt the service to run as a dedicated account with restricted file access. On `mipsle`, select either the `softfloat` or `hardfloat` archive for the firmware ABI.

## systemd

Install the binary, configuration, and supplied unit, then reload systemd:

```bash
sudo install -m 0755 mihomo-exporter /usr/local/bin/mihomo-exporter
sudo install -d /etc/mihomo-exporter
sudo install -m 0644 config.example.yaml /etc/mihomo-exporter/config.yaml
sudo install -m 0644 contrib/systemd/mihomo-exporter.service /etc/systemd/system/mihomo-exporter.service
sudo systemctl daemon-reload
sudo systemctl enable --now mihomo-exporter
```

The unit starts `/usr/local/bin/mihomo-exporter --config /etc/mihomo-exporter/config.yaml` with `DynamicUser=true` and filesystem hardening. The shown mode makes the file readable to its dynamic service user. Do not put a real `mihomo.secret` in this configuration unchanged: adapt the unit to use a dedicated service account and restricted file permissions first.
