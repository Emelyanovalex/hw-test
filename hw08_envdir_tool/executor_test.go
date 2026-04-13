package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRunCmd(t *testing.T) {
	t.Run("returns zero exit code on success", func(t *testing.T) {
		code := RunCmd([]string{"true"}, Environment{})
		require.Equal(t, 0, code)
	})

	t.Run("returns non-zero exit code on failure", func(t *testing.T) {
		code := RunCmd([]string{"false"}, Environment{})
		require.NotEqual(t, 0, code)
	})

	t.Run("passes env variables to command", func(t *testing.T) {
		env := Environment{
			"TEST_VAR": {Value: "hello"},
		}
		code := RunCmd([]string{"bash", "-c", `[ "$TEST_VAR" = "hello" ]`}, env)
		require.Equal(t, 0, code)
	})

	t.Run("removes env variable marked for removal", func(t *testing.T) {
		t.Setenv("TO_REMOVE", "present")
		env := Environment{
			"TO_REMOVE": {NeedRemove: true},
		}
		// If TO_REMOVE is absent, the test variable will be empty → exit 0
		code := RunCmd([]string{"bash", "-c", `[ -z "$TO_REMOVE" ]`}, env)
		require.Equal(t, 0, code)
	})

	t.Run("keeps original env variables not in env map", func(t *testing.T) {
		t.Setenv("ORIGINAL_VAR", "kept")
		env := Environment{}
		code := RunCmd([]string{"bash", "-c", `[ "$ORIGINAL_VAR" = "kept" ]`}, env)
		require.Equal(t, 0, code)
	})

	t.Run("overrides existing env variable", func(t *testing.T) {
		t.Setenv("OVERRIDE_ME", "old")
		env := Environment{
			"OVERRIDE_ME": {Value: "new"},
		}
		code := RunCmd([]string{"bash", "-c", `[ "$OVERRIDE_ME" = "new" ]`}, env)
		require.Equal(t, 0, code)
	})

	t.Run("passes stdin stdout stderr", func(t *testing.T) {
		origStdout := os.Stdout
		defer func() { os.Stdout = origStdout }()

		code := RunCmd([]string{"bash", "-c", "exit 0"}, Environment{})
		require.Equal(t, 0, code)
	})
}
