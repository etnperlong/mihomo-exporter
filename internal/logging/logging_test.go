package logging_test

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/etnperlong/mihomo-exporter/internal/config"
	"github.com/etnperlong/mihomo-exporter/internal/logging"
)

func TestNewOutput(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		format logging.Format
		emoji  logging.Toggle
		isTTY  bool
		level  zapcore.Level
		write  func(*zap.Logger)
		check  func(t *testing.T, output string)
	}{
		"JSON is machine readable without emoji": {
			format: logging.FormatJSON,
			emoji:  logging.ToggleOn,
			level:  zapcore.DebugLevel,
			write:  func(logger *zap.Logger) { logger.Info("hello") },
			check: func(t *testing.T, output string) {
				t.Helper()
				var entry map[string]any
				if err := json.Unmarshal([]byte(output), &entry); err != nil {
					t.Fatalf("JSON output %q cannot be parsed: %v", output, err)
				}
				for _, key := range []string{"level", "ts", "msg"} {
					if _, ok := entry[key]; !ok {
						t.Errorf("JSON output %q does not contain key %q", output, key)
					}
				}
				if strings.ContainsAny(output, "🔍ℹ⚠❌🔥") {
					t.Errorf("JSON output %q contains an emoji", output)
				}
			},
		},
		"text emoji remains after With": {
			format: logging.FormatText,
			emoji:  logging.ToggleOn,
			level:  zapcore.DebugLevel,
			write: func(logger *zap.Logger) {
				logger.With(zap.String("component", "test")).Named("child").Info("hello")
			},
			check: func(t *testing.T, output string) {
				t.Helper()
				if !strings.HasPrefix(output, "ℹ️ ") {
					t.Errorf("text output %q does not start with the info emoji", output)
				}
			},
		},
		"auto format uses text on a terminal": {
			format: logging.FormatAuto,
			emoji:  logging.ToggleOff,
			isTTY:  true,
			level:  zapcore.DebugLevel,
			write:  func(logger *zap.Logger) { logger.Info("hello") },
			check: func(t *testing.T, output string) {
				t.Helper()
				if strings.HasPrefix(output, "{") {
					t.Errorf("terminal auto output %q is JSON, want text", output)
				}
			},
		},
		"auto format uses JSON off a terminal": {
			format: logging.FormatAuto,
			emoji:  logging.ToggleAuto,
			isTTY:  false,
			level:  zapcore.DebugLevel,
			write:  func(logger *zap.Logger) { logger.Info("hello") },
			check: func(t *testing.T, output string) {
				t.Helper()
				if !strings.HasPrefix(output, "{") {
					t.Errorf("non-terminal auto output %q is not JSON", output)
				}
			},
		},
		"warn level filters info": {
			format: logging.FormatJSON,
			emoji:  logging.ToggleOff,
			level:  zapcore.WarnLevel,
			write: func(logger *zap.Logger) {
				logger.Info("discarded")
				logger.Warn("included")
			},
			check: func(t *testing.T, output string) {
				t.Helper()
				if strings.Contains(output, "discarded") || !strings.Contains(output, "included") {
					t.Errorf("warn-level output = %q, want only warning", output)
				}
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var output bytes.Buffer
			logger, err := logging.New(logging.Options{
				Level:  tt.level,
				Format: tt.format,
				Emoji:  tt.emoji,
				Color:  logging.ToggleOff,
				Output: &output,
				IsTTY:  func(io.Writer) bool { return tt.isTTY },
			})
			if err != nil {
				t.Fatalf("New() returned unexpected error: %v", err)
			}
			tt.write(logger)
			tt.check(t, output.String())
		})
	}
}

func TestNewEmojiAutoEnvironment(t *testing.T) {
	tests := map[string]struct {
		term    string
		noColor string
		want    bool
	}{
		"supported terminal": {term: "xterm-256color", want: true},
		"dumb terminal":      {term: "dumb", want: false},
		"NO_COLOR":           {term: "xterm", noColor: "1", want: false},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Setenv("TERM", tt.term)
			t.Setenv("NO_COLOR", tt.noColor)

			var output bytes.Buffer
			logger, err := logging.New(logging.Options{
				Level:  zapcore.DebugLevel,
				Format: logging.FormatText,
				Emoji:  logging.ToggleAuto,
				Color:  logging.ToggleOff,
				Output: &output,
				IsTTY:  func(io.Writer) bool { return true },
			})
			if err != nil {
				t.Fatalf("New() returned unexpected error: %v", err)
			}
			logger.Info("hello")

			got := strings.HasPrefix(output.String(), "ℹ️ ")
			if got != tt.want {
				t.Errorf("emoji enabled for TERM=%q NO_COLOR=%q = %t, want %t", tt.term, tt.noColor, got, tt.want)
			}
		})
	}
}

func TestNewRedactsSecret(t *testing.T) {
	t.Parallel()

	const secret = "sensitive-value"
	var output bytes.Buffer
	logger, err := logging.New(logging.Options{
		Level:  zapcore.DebugLevel,
		Format: logging.FormatJSON,
		Emoji:  logging.ToggleOff,
		Color:  logging.ToggleOff,
		Output: &output,
		IsTTY:  func(io.Writer) bool { return false },
	})
	if err != nil {
		t.Fatalf("New() returned unexpected error: %v", err)
	}
	logger.Info("configuration", zap.Object("config", config.Config{
		Mihomo: config.MihomoConfig{Secret: config.Secret(secret)},
	}))

	if strings.Contains(output.String(), secret) {
		t.Errorf("logger output %q contains secret %q", output.String(), secret)
	}
}
