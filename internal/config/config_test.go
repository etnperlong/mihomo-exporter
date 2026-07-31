package config_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/etnperlong/mihomo-exporter/internal/config"
)

const validConfig = `
web:
  listen_address: ":1111"
  telemetry_path: "/file-metrics"
  shutdown_timeout: 12s
mihomo:
  api_url: "http://file.example:9090"
  secret: "file-secret"
  timeout: 11s
probe:
  test_url: "https://file.example/generate_204"
  test_timeout: 4s
  fast_interval: 14s
  slow_interval: 61s
  concurrency: 7
  staleness: 6m
  exclude_types: "File"
  exclude_names: "^file$"
metrics:
  enable_connections: true
log:
  level: warn
  format: text
  emoji: off
  color: on
`

func TestLoadPrecedence(t *testing.T) {
	path := writeConfig(t, validConfig)

	tests := map[string]struct {
		env  map[string]string
		args []string
		want config.Config
	}{
		"file overrides defaults": {
			args: []string{"--config", path},
			want: config.Config{
				Web: config.WebConfig{
					ListenAddress:   ":1111",
					TelemetryPath:   "/file-metrics",
					ShutdownTimeout: 12 * time.Second,
				},
				Mihomo: config.MihomoConfig{
					APIURL:  "http://file.example:9090",
					Secret:  config.Secret("file-secret"),
					Timeout: 11 * time.Second,
				},
				Probe: config.ProbeConfig{
					TestURL:      "https://file.example/generate_204",
					TestTimeout:  4 * time.Second,
					FastInterval: 14 * time.Second,
					SlowInterval: 61 * time.Second,
					Concurrency:  7,
					Staleness:    6 * time.Minute,
					ExcludeTypes: "File",
					ExcludeNames: "^file$",
				},
				Metrics: config.MetricsConfig{EnableConnections: true},
				Log:     config.LogConfig{Level: "warn", Format: "text", Emoji: "off", Color: "on"},
			},
		},
		"environment overrides file": {
			env: map[string]string{
				"MIHOMO_EXPORTER_MIHOMO_API_URL":    "https://env.example",
				"MIHOMO_EXPORTER_PROBE_CONCURRENCY": "3",
			},
			args: []string{"--config", path},
			want: config.Config{
				Web: config.WebConfig{
					ListenAddress:   ":1111",
					TelemetryPath:   "/file-metrics",
					ShutdownTimeout: 12 * time.Second,
				},
				Mihomo: config.MihomoConfig{
					APIURL:  "https://env.example",
					Secret:  config.Secret("file-secret"),
					Timeout: 11 * time.Second,
				},
				Probe: config.ProbeConfig{
					TestURL:      "https://file.example/generate_204",
					TestTimeout:  4 * time.Second,
					FastInterval: 14 * time.Second,
					SlowInterval: 61 * time.Second,
					Concurrency:  3,
					Staleness:    6 * time.Minute,
					ExcludeTypes: "File",
					ExcludeNames: "^file$",
				},
				Metrics: config.MetricsConfig{EnableConnections: true},
				Log:     config.LogConfig{Level: "warn", Format: "text", Emoji: "off", Color: "on"},
			},
		},
		"flags override environment": {
			env: map[string]string{
				"MIHOMO_EXPORTER_MIHOMO_API_URL": "https://env.example",
			},
			args: []string{"--config", path, "--mihomo.api-url", "https://flag.example"},
			want: config.Config{
				Web: config.WebConfig{
					ListenAddress:   ":1111",
					TelemetryPath:   "/file-metrics",
					ShutdownTimeout: 12 * time.Second,
				},
				Mihomo: config.MihomoConfig{
					APIURL:  "https://flag.example",
					Secret:  config.Secret("file-secret"),
					Timeout: 11 * time.Second,
				},
				Probe: config.ProbeConfig{
					TestURL:      "https://file.example/generate_204",
					TestTimeout:  4 * time.Second,
					FastInterval: 14 * time.Second,
					SlowInterval: 61 * time.Second,
					Concurrency:  7,
					Staleness:    6 * time.Minute,
					ExcludeTypes: "File",
					ExcludeNames: "^file$",
				},
				Metrics: config.MetricsConfig{EnableConnections: true},
				Log:     config.LogConfig{Level: "warn", Format: "text", Emoji: "off", Color: "on"},
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			for key, value := range tt.env {
				t.Setenv(key, value)
			}

			got, err := config.Load(tt.args)
			if err != nil {
				t.Fatalf("Load(%q) returned unexpected error: %v", tt.args, err)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("Load(%q) mismatch (-want +got):\n%s", tt.args, diff)
			}
		})
	}
}

func TestLoadFileSemantics(t *testing.T) {
	tests := map[string]struct {
		args    []string
		wantErr bool
	}{
		"missing explicit path returns error": {
			args:    []string{"--config", filepath.Join(t.TempDir(), "missing.yaml")},
			wantErr: true,
		},
		"no configuration file uses defaults": {},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := config.Load(tt.args)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("Load(%q) = nil error, want error", tt.args)
				}

				return
			}
			if err != nil {
				t.Fatalf("Load(%q) returned unexpected error: %v", tt.args, err)
			}
			if got.Web.ListenAddress != ":9878" {
				t.Errorf("Load(%q) listen address = %q, want %q", tt.args, got.Web.ListenAddress, ":9878")
			}
		})
	}
}

func TestConfigValidate(t *testing.T) {
	valid, err := config.Load(nil)
	if err != nil {
		t.Fatalf("Load(nil) returned unexpected error: %v", err)
	}

	tests := map[string]struct {
		mutate func(*config.Config)
	}{
		"invalid API URL": {
			mutate: func(cfg *config.Config) { cfg.Mihomo.APIURL = "ftp://example.com" },
		},
		"zero timeout": {
			mutate: func(cfg *config.Config) { cfg.Mihomo.Timeout = 0 },
		},
		"slow interval below test timeout": {
			mutate: func(cfg *config.Config) { cfg.Probe.SlowInterval = cfg.Probe.TestTimeout - time.Second },
		},
		"staleness not above slow interval": {
			mutate: func(cfg *config.Config) { cfg.Probe.Staleness = cfg.Probe.SlowInterval },
		},
		"zero concurrency": {
			mutate: func(cfg *config.Config) { cfg.Probe.Concurrency = 0 },
		},
		"invalid excluded name expression": {
			mutate: func(cfg *config.Config) { cfg.Probe.ExcludeNames = "[" },
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := valid
			tt.mutate(&got)
			if err := got.Validate(); err == nil {
				t.Errorf("Validate() with %s = nil error, want error", name)
			}
		})
	}
}

func TestSecretRedaction(t *testing.T) {
	t.Parallel()

	const secret = "super-secret-token"
	var output bytes.Buffer
	logger := zap.New(zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		zapcore.AddSync(&output),
		zapcore.InfoLevel,
	))
	logger.Info("configuration", zap.Object("config", config.Config{
		Mihomo: config.MihomoConfig{Secret: config.Secret(secret)},
	}))

	got := output.String()
	if strings.Contains(got, secret) {
		t.Errorf("logger output %q contains secret %q", got, secret)
	}
	if !strings.Contains(got, "REDACTED") {
		t.Errorf("logger output %q does not contain %q", got, "REDACTED")
	}
}

func writeConfig(t *testing.T, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("WriteFile(%q) returned unexpected error: %v", path, err)
	}

	return path
}
