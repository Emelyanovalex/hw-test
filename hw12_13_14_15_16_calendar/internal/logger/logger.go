package logger

import (
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger is a thin wrapper around zap.Logger with a simple string-based API.
type Logger struct {
	z *zap.Logger
}

// New constructs a Logger that writes JSON entries to STDOUT at the given level.
// Allowed levels: "debug", "info", "warn", "error". Unknown values fall back to info.
func New(level string) *Logger {
	cfg := zap.NewProductionEncoderConfig()
	cfg.TimeKey = "ts"
	cfg.MessageKey = "msg"
	cfg.LevelKey = "level"
	cfg.EncodeTime = zapcore.ISO8601TimeEncoder
	cfg.EncodeLevel = zapcore.CapitalLevelEncoder

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(cfg),
		zapcore.AddSync(os.Stdout),
		parseLevel(level),
	)
	return &Logger{z: zap.New(core)}
}

func parseLevel(level string) zapcore.Level {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		return zapcore.DebugLevel
	case "warn", "warning":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}

func (l *Logger) Debug(msg string) { l.z.Debug(msg) }
func (l *Logger) Info(msg string)  { l.z.Info(msg) }
func (l *Logger) Warn(msg string)  { l.z.Warn(msg) }
func (l *Logger) Error(msg string) { l.z.Error(msg) }

// Sync flushes any buffered log entries.
func (l *Logger) Sync() error { return l.z.Sync() }
