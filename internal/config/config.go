// Package config loads and validates exporter configuration.
package config

import (
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env/v2"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/posflag"
	"github.com/knadh/koanf/v2"
	"github.com/spf13/pflag"
	"go.uber.org/zap/zapcore"
)

const redacted = "REDACTED"

var (
	errInvalidAPIURL           = errors.New("mihomo api url must use http or https")
	errInvalidPositiveDuration = errors.New("duration must be greater than zero")
	errInvalidConcurrency      = errors.New("probe concurrency must be at least one")
	errInvalidExcludeNames     = errors.New("probe exclude names must be a valid regular expression")
	errSlowInterval            = errors.New("probe slow interval must be at least probe test timeout")
	errStaleness               = errors.New("probe staleness must exceed probe slow interval")
)

// Config contains all exporter configuration.
type Config struct {
	Web     WebConfig
	Mihomo  MihomoConfig
	Probe   ProbeConfig
	Metrics MetricsConfig
	Log     LogConfig
}

// WebConfig controls the exporter HTTP server.
type WebConfig struct {
	ListenAddress   string
	TelemetryPath   string
	ShutdownTimeout time.Duration
}

// MihomoConfig controls the Mihomo API client.
type MihomoConfig struct {
	APIURL  string
	Secret  Secret
	Timeout time.Duration
}

// ProbeConfig controls asynchronous Mihomo probing.
type ProbeConfig struct {
	TestURL      string
	TestTimeout  time.Duration
	FastInterval time.Duration
	SlowInterval time.Duration
	Concurrency  int
	Staleness    time.Duration
	ExcludeTypes string
	ExcludeNames string
}

// MetricsConfig controls optional metric groups.
type MetricsConfig struct {
	EnableConnections bool
}

// LogConfig controls the application logger.
type LogConfig struct {
	Level  string
	Format string
	Emoji  string
	Color  string
}

// Secret holds a value that must never appear in logs.
type Secret string

// String returns a redacted representation of Secret.
func (Secret) String() string {
	return redacted
}

// LogValue returns a redacted structured-log value for Secret.
func (Secret) LogValue() slog.Value {
	return slog.StringValue(redacted)
}

// MarshalLogObject writes a redacted representation of Secret to a zap encoder.
func (Secret) MarshalLogObject(encoder zapcore.ObjectEncoder) error {
	encoder.AddString("value", redacted)

	return nil
}

// MarshalLogObject writes Config without exposing its Mihomo API secret.
func (c Config) MarshalLogObject(encoder zapcore.ObjectEncoder) error {
	encoder.AddString("web_listen_address", c.Web.ListenAddress)
	encoder.AddString("web_telemetry_path", c.Web.TelemetryPath)
	encoder.AddDuration("web_shutdown_timeout", c.Web.ShutdownTimeout)
	encoder.AddString("mihomo_api_url", c.Mihomo.APIURL)
	if err := encoder.AddObject("mihomo_secret", c.Mihomo.Secret); err != nil {
		return fmt.Errorf("encode mihomo secret: %w", err)
	}
	encoder.AddDuration("mihomo_timeout", c.Mihomo.Timeout)
	encoder.AddString("probe_test_url", c.Probe.TestURL)
	encoder.AddDuration("probe_test_timeout", c.Probe.TestTimeout)
	encoder.AddDuration("probe_fast_interval", c.Probe.FastInterval)
	encoder.AddDuration("probe_slow_interval", c.Probe.SlowInterval)
	encoder.AddInt("probe_concurrency", c.Probe.Concurrency)
	encoder.AddDuration("probe_staleness", c.Probe.Staleness)
	encoder.AddString("probe_exclude_types", c.Probe.ExcludeTypes)
	encoder.AddString("probe_exclude_names", c.Probe.ExcludeNames)
	encoder.AddBool("metrics_enable_connections", c.Metrics.EnableConnections)
	encoder.AddString("log_level", c.Log.Level)
	encoder.AddString("log_format", c.Log.Format)
	encoder.AddString("log_emoji", c.Log.Emoji)
	encoder.AddString("log_color", c.Log.Color)

	return nil
}

// Load returns configuration with command-line flags taking precedence over
// environment variables, configuration files, and built-in defaults.
func Load(args []string) (Config, error) {
	flags := newFlagSet()
	if err := flags.Parse(args); err != nil {
		return Config{}, fmt.Errorf("parse flags: %w", err)
	}

	ko := koanf.New(".")
	if err := loadFile(ko, configPath(flags)); err != nil {
		return Config{}, err
	}
	if err := ko.Load(env.Provider(".", env.Opt{
		Prefix: "MIHOMO_EXPORTER_",
		TransformFunc: func(key, value string) (string, any) {
			return environmentKey(strings.TrimPrefix(key, "MIHOMO_EXPORTER_")), value
		},
	}), nil); err != nil {
		return Config{}, fmt.Errorf("load environment: %w", err)
	}
	if err := ko.Load(posflag.ProviderWithFlag(flags, ".", ko, func(flag *pflag.Flag) (string, any) {
		return strings.ReplaceAll(flag.Name, "-", "_"), flag.Value.String()
	}), nil); err != nil {
		return Config{}, fmt.Errorf("load flags: %w", err)
	}

	config := Config{
		Web: WebConfig{
			ListenAddress:   ko.String("web.listen_address"),
			TelemetryPath:   ko.String("web.telemetry_path"),
			ShutdownTimeout: ko.Duration("web.shutdown_timeout"),
		},
		Mihomo: MihomoConfig{
			APIURL:  ko.String("mihomo.api_url"),
			Secret:  Secret(ko.String("mihomo.secret")),
			Timeout: ko.Duration("mihomo.timeout"),
		},
		Probe: ProbeConfig{
			TestURL:      ko.String("probe.test_url"),
			TestTimeout:  ko.Duration("probe.test_timeout"),
			FastInterval: ko.Duration("probe.fast_interval"),
			SlowInterval: ko.Duration("probe.slow_interval"),
			Concurrency:  ko.Int("probe.concurrency"),
			Staleness:    ko.Duration("probe.staleness"),
			ExcludeTypes: ko.String("probe.exclude_types"),
			ExcludeNames: ko.String("probe.exclude_names"),
		},
		Metrics: MetricsConfig{
			EnableConnections: ko.Bool("metrics.enable_connections"),
		},
		Log: LogConfig{
			Level:  ko.String("log.level"),
			Format: ko.String("log.format"),
			Emoji:  ko.String("log.emoji"),
			Color:  ko.String("log.color"),
		},
	}
	if err := config.Validate(); err != nil {
		return Config{}, err
	}

	return config, nil
}

// Validate checks that Config is safe to use before the exporter starts.
func (c Config) Validate() error {
	parsedURL, err := url.Parse(c.Mihomo.APIURL)
	if err != nil {
		return fmt.Errorf("parse mihomo api url: %w", err)
	}
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" || parsedURL.Host == "" {
		return errInvalidAPIURL
	}

	for name, duration := range map[string]time.Duration{
		"web shutdown timeout": c.Web.ShutdownTimeout,
		"mihomo timeout":       c.Mihomo.Timeout,
		"probe test timeout":   c.Probe.TestTimeout,
		"probe fast interval":  c.Probe.FastInterval,
		"probe slow interval":  c.Probe.SlowInterval,
		"probe staleness":      c.Probe.Staleness,
	} {
		if duration <= 0 {
			return fmt.Errorf("%s: %w", name, errInvalidPositiveDuration)
		}
	}
	if c.Probe.SlowInterval < c.Probe.TestTimeout {
		return errSlowInterval
	}
	if c.Probe.Staleness <= c.Probe.SlowInterval {
		return errStaleness
	}
	if c.Probe.Concurrency < 1 {
		return errInvalidConcurrency
	}
	if _, err := regexp.Compile(c.Probe.ExcludeNames); err != nil {
		return fmt.Errorf("%w: %w", errInvalidExcludeNames, err)
	}

	return nil
}

func newFlagSet() *pflag.FlagSet {
	flags := pflag.NewFlagSet("mihomo-exporter", pflag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	flags.String("web.listen-address", ":9878", "address to listen on")
	flags.String("web.telemetry-path", "/metrics", "path to expose metrics")
	flags.Duration("web.shutdown-timeout", 10*time.Second, "graceful shutdown timeout")
	flags.String("mihomo.api-url", "http://127.0.0.1:9090", "Mihomo API URL")
	flags.String("mihomo.secret", "", "Mihomo API secret")
	flags.Duration("mihomo.timeout", 10*time.Second, "Mihomo API request timeout")
	flags.String("probe.test-url", "https://www.gstatic.com/generate_204", "fallback probe URL")
	flags.Duration("probe.test-timeout", 5*time.Second, "fallback probe timeout")
	flags.Duration("probe.fast-interval", 15*time.Second, "fast probe interval")
	flags.Duration("probe.slow-interval", time.Minute, "slow probe interval")
	flags.Int("probe.concurrency", 8, "maximum concurrent probes")
	flags.Duration("probe.staleness", 5*time.Minute, "maximum probe data age")
	flags.String("probe.exclude-types", "", "comma-separated proxy types to exclude")
	flags.String("probe.exclude-names", "", "regular expression for proxy names to exclude")
	flags.Bool("metrics.enable-connections", false, "enable connection metrics")
	flags.String("log.level", "info", "log level")
	flags.String("log.format", "auto", "log format")
	flags.String("log.emoji", "auto", "emoji output mode")
	flags.String("log.color", "auto", "color output mode")
	flags.String("config", "", "configuration file path")
	flags.Bool("version", false, "print version and exit")

	return flags
}

func configPath(flags *pflag.FlagSet) string {
	if flags.Changed("config") {
		return flags.Lookup("config").Value.String()
	}

	return os.Getenv("MIHOMO_EXPORTER_CONFIG")
}

func environmentKey(key string) string {
	keys := map[string]string{
		"WEB_LISTEN_ADDRESS":         "web.listen_address",
		"WEB_TELEMETRY_PATH":         "web.telemetry_path",
		"WEB_SHUTDOWN_TIMEOUT":       "web.shutdown_timeout",
		"MIHOMO_API_URL":             "mihomo.api_url",
		"MIHOMO_SECRET":              "mihomo.secret",
		"MIHOMO_TIMEOUT":             "mihomo.timeout",
		"PROBE_TEST_URL":             "probe.test_url",
		"PROBE_TEST_TIMEOUT":         "probe.test_timeout",
		"PROBE_FAST_INTERVAL":        "probe.fast_interval",
		"PROBE_SLOW_INTERVAL":        "probe.slow_interval",
		"PROBE_CONCURRENCY":          "probe.concurrency",
		"PROBE_STALENESS":            "probe.staleness",
		"PROBE_EXCLUDE_TYPES":        "probe.exclude_types",
		"PROBE_EXCLUDE_NAMES":        "probe.exclude_names",
		"METRICS_ENABLE_CONNECTIONS": "metrics.enable_connections",
		"LOG_LEVEL":                  "log.level",
		"LOG_FORMAT":                 "log.format",
		"LOG_EMOJI":                  "log.emoji",
		"LOG_COLOR":                  "log.color",
	}
	if configKey, ok := keys[key]; ok {
		return configKey
	}

	return strings.ToLower(strings.ReplaceAll(key, "_", "."))
}

func loadFile(ko *koanf.Koanf, path string) error {
	if path != "" {
		if err := ko.Load(file.Provider(path), yaml.Parser()); err != nil {
			return fmt.Errorf("load configuration file %q: %w", path, err)
		}

		return nil
	}

	for _, candidate := range []string{
		"./mihomo-exporter.yaml",
		"/etc/mihomo-exporter/config.yaml",
		"/etc/mihomo-exporter.yaml",
	} {
		if _, statErr := os.Stat(candidate); statErr == nil {
			if loadErr := ko.Load(file.Provider(candidate), yaml.Parser()); loadErr != nil {
				return fmt.Errorf("load configuration file %q: %w", candidate, loadErr)
			}

			return nil
		} else if !errors.Is(statErr, os.ErrNotExist) {
			return fmt.Errorf("stat configuration file %q: %w", candidate, statErr)
		}
	}

	return nil
}
