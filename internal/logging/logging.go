package logging

import (
	"fmt"
	"io"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/term"
)

// New constructs a zap logger from opts.
func New(opts Options) (*zap.Logger, error) {
	if err := ValidateOptions(opts); err != nil {
		return nil, err
	}
	if opts.Output == nil {
		opts.Output = os.Stderr
	}
	if opts.IsTTY == nil {
		opts.IsTTY = isTerminal
	}

	isTTY := opts.IsTTY(opts.Output)
	format := resolveFormat(opts.Format, isTTY)
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = rfc3339MillisTimeEncoder
	encoderConfig.EncodeLevel = levelEncoder(textColorEnabled(opts.Color, format, isTTY))

	var encoder zapcore.Encoder
	if format == FormatJSON {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
		if emojiEnabled(opts.Emoji, isTTY) {
			encoder = emojiEncoder{Encoder: encoder}
		}
	}

	return zap.New(zapcore.NewCore(encoder, zapcore.AddSync(opts.Output), opts.Level)), nil
}

func resolveFormat(format Format, isTTY bool) Format {
	if format == FormatAuto {
		if isTTY {
			return FormatText
		}

		return FormatJSON
	}

	return format
}

func textColorEnabled(toggle Toggle, format Format, isTTY bool) bool {
	if format != FormatText {
		return false
	}
	if toggle == ToggleOn {
		return true
	}

	return toggle == ToggleAuto && isTTY && decorationsSupported()
}

func emojiEnabled(toggle Toggle, isTTY bool) bool {
	if toggle == ToggleOn {
		return true
	}

	return toggle == ToggleAuto && isTTY && decorationsSupported()
}

func decorationsSupported() bool {
	if os.Getenv("NO_COLOR") != "" || os.Getenv("TERM") == "dumb" {
		return false
	}

	return os.Getenv("TERM") != ""
}

func isTerminal(output io.Writer) bool {
	file, ok := output.(interface{ Fd() uintptr })
	return ok && term.IsTerminal(int(file.Fd()))
}

func levelEncoder(color bool) zapcore.LevelEncoder {
	if color {
		return zapcore.CapitalColorLevelEncoder
	}

	return zapcore.CapitalLevelEncoder
}

func rfc3339MillisTimeEncoder(value time.Time, encoder zapcore.PrimitiveArrayEncoder) {
	encoder.AppendString(value.Format("2006-01-02T15:04:05.000Z07:00"))
}

// ValidateOptions checks options that are invalid regardless of the output environment.
func ValidateOptions(opts Options) error {
	if opts.Format > FormatJSON {
		return fmt.Errorf("%w: %d", errInvalidFormat, opts.Format)
	}
	if opts.Emoji > ToggleOff {
		return fmt.Errorf("%w: %d", errInvalidToggle, opts.Emoji)
	}
	if opts.Color > ToggleOff {
		return fmt.Errorf("%w: %d", errInvalidToggle, opts.Color)
	}

	return nil
}
