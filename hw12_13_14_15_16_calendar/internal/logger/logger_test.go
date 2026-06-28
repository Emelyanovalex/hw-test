package logger

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
)

func TestParseLevel(t *testing.T) {
	cases := map[string]zapcore.Level{
		"debug":   zapcore.DebugLevel,
		"DEBUG":   zapcore.DebugLevel,
		"info":    zapcore.InfoLevel,
		" INFO ":  zapcore.InfoLevel, //nolint:gocritic // intentional: testing space-trimming
		"warn":    zapcore.WarnLevel,
		"warning": zapcore.WarnLevel,
		"error":   zapcore.ErrorLevel,
		"":        zapcore.InfoLevel,
		"unknown": zapcore.InfoLevel,
	}
	for in, expected := range cases {
		require.Equal(t, expected, parseLevel(in), "level %q", in)
	}
}

func TestNewDoesNotPanic(t *testing.T) {
	require.NotPanics(t, func() {
		l := New("info")
		l.Info("hello")
		l.Debug("ignored")
		l.Warn("careful")
		l.Error("oops")
	})
}
