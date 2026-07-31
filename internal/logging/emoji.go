package logging

import (
	"fmt"

	"go.uber.org/zap/buffer"
	"go.uber.org/zap/zapcore"
)

var bufPool = buffer.NewPool()

type emojiEncoder struct {
	zapcore.Encoder
}

func (e emojiEncoder) Clone() zapcore.Encoder {
	return emojiEncoder{Encoder: e.Encoder.Clone()}
}

func (e emojiEncoder) EncodeEntry(entry zapcore.Entry, fields []zapcore.Field) (*buffer.Buffer, error) {
	encoded, err := e.Encoder.EncodeEntry(entry, fields)
	if err != nil {
		return nil, fmt.Errorf("encode log entry: %w", err)
	}

	output := bufPool.Get()
	output.AppendString(emojiFor(entry.Level))
	output.AppendString(" ")
	output.AppendString(encoded.String())
	encoded.Free()

	return output, nil
}

func emojiFor(level zapcore.Level) string {
	switch {
	case level <= zapcore.DebugLevel:
		return "🔍"
	case level == zapcore.InfoLevel:
		return "ℹ️"
	case level == zapcore.WarnLevel:
		return "⚠️"
	case level == zapcore.ErrorLevel:
		return "❌"
	default:
		return "🔥"
	}
}
