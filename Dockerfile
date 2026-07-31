FROM golang:1.26 AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -trimpath -tags 'netgo osusergo' -o /out/mihomo-exporter ./cmd/mihomo-exporter

# GoReleaser targets this stage after supplying its already-built binary.
FROM gcr.io/distroless/static-debian12:nonroot AS release

ARG TARGETPLATFORM
COPY $TARGETPLATFORM/mihomo-exporter /mihomo-exporter
USER nonroot:nonroot
ENTRYPOINT ["/mihomo-exporter"]

# The final stage makes a plain `docker build .` compile and run the application.
FROM gcr.io/distroless/static-debian12:nonroot AS local

COPY --from=build /out/mihomo-exporter /mihomo-exporter
USER nonroot:nonroot
ENTRYPOINT ["/mihomo-exporter"]
