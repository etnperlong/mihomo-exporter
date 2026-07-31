// Package logging constructs the exporter logger.
package logging

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"go.uber.org/zap/zapcore"
)

var (
	errInvalidFormat = errors.New("invalid log format")
	errInvalidToggle = errors.New("invalid log toggle")
)

// Format controls the logger output encoding.
type Format uint8

const (
	// FormatAuto selects text for terminals and JSON otherwise.
	FormatAuto Format = iota
	// FormatText selects a human-readable console encoding.
	FormatText
	// FormatJSON selects a machine-readable JSON encoding.
	FormatJSON
)

// Toggle controls an optional logger feature.
type Toggle uint8

const (
	// ToggleAuto enables a feature only in suitable environments.
	ToggleAuto Toggle = iota
	// ToggleOn enables a feature unconditionally when applicable.
	ToggleOn
	// ToggleOff disables a feature.
	ToggleOff
)

// Options describes the logger configuration.
type Options struct {
	Level  zapcore.Level
	Format Format
	Emoji  Toggle
	Color  Toggle
	Output io.Writer
	IsTTY  func(io.Writer) bool
}

// ParseFormat parses a configured log format.
func ParseFormat(value string) (Format, error) {
	switch strings.ToLower(value) {
	case "auto":
		return FormatAuto, nil
	case "text":
		return FormatText, nil
	case "json":
		return FormatJSON, nil
	default:
		return FormatAuto, fmt.Errorf("%w: %q", errInvalidFormat, value)
	}
}

// ParseToggle parses a configured on, off, or auto toggle.
func ParseToggle(value string) (Toggle, error) {
	switch strings.ToLower(value) {
	case "auto":
		return ToggleAuto, nil
	case "on":
		return ToggleOn, nil
	case "off":
		return ToggleOff, nil
	default:
		return ToggleAuto, fmt.Errorf("%w: %q", errInvalidToggle, value)
	}
}
