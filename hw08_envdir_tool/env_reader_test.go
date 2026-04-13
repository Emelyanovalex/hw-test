package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReadDir(t *testing.T) {
	t.Run("testdata env directory", func(t *testing.T) {
		env, err := ReadDir("testdata/env")
		require.NoError(t, err)

		// FOO: "   foo\x00with new line" → null replaced with \n, no trailing trim needed
		require.Equal(t, EnvValue{Value: "   foo\nwith new line"}, env["FOO"])

		// BAR: first line "bar\r\n", second line ignored; trim \r → "bar"
		require.Equal(t, EnvValue{Value: "bar"}, env["BAR"])

		// HELLO: '"hello"' no newline
		require.Equal(t, EnvValue{Value: `"hello"`}, env["HELLO"])

		// UNSET: empty file → NeedRemove
		require.Equal(t, EnvValue{NeedRemove: true}, env["UNSET"])

		// EMPTY: " \r\n" → first line " \r", trim → "" (not empty file, so not NeedRemove)
		require.Equal(t, EnvValue{Value: ""}, env["EMPTY"])
	})

	t.Run("skips files with = in name", func(t *testing.T) {
		dir := t.TempDir()
		err := os.WriteFile(filepath.Join(dir, "KEY=BAD"), []byte("value"), 0o600)
		require.NoError(t, err)
		err = os.WriteFile(filepath.Join(dir, "GOOD"), []byte("ok"), 0o600)
		require.NoError(t, err)

		env, err := ReadDir(dir)
		require.NoError(t, err)
		require.Contains(t, env, "GOOD")
		require.NotContains(t, env, "KEY=BAD")
	})

	t.Run("skips subdirectories", func(t *testing.T) {
		dir := t.TempDir()
		err := os.Mkdir(filepath.Join(dir, "SUBDIR"), 0o700)
		require.NoError(t, err)
		err = os.WriteFile(filepath.Join(dir, "VAR"), []byte("val"), 0o600)
		require.NoError(t, err)

		env, err := ReadDir(dir)
		require.NoError(t, err)
		require.Contains(t, env, "VAR")
		require.NotContains(t, env, "SUBDIR")
	})

	t.Run("returns error on missing directory", func(t *testing.T) {
		_, err := ReadDir("/no/such/directory")
		require.Error(t, err)
	})
}
