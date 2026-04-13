package main

import (
	"os"
	"os/exec"
	"strings"
)

// RunCmd runs a command + arguments (cmd) with environment variables from env.
func RunCmd(cmd []string, env Environment) (returnCode int) {
	command := exec.Command(cmd[0], cmd[1:]...) //nolint:gosec

	// Build new env: start from current process env, remove overridden keys, then add new ones
	newEnv := make([]string, 0, len(os.Environ()))
	for _, e := range os.Environ() {
		key := e
		if idx := strings.Index(e, "="); idx != -1 {
			key = e[:idx]
		}
		if _, exists := env[key]; exists {
			continue
		}
		newEnv = append(newEnv, e)
	}

	for name, val := range env {
		if !val.NeedRemove {
			newEnv = append(newEnv, name+"="+val.Value)
		}
	}

	command.Env = newEnv
	command.Stdin = os.Stdin
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr

	if err := command.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return exitErr.ExitCode()
		}
		return 1
	}

	return 0
}
