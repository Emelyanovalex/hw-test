package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
)

type Environment map[string]EnvValue

// EnvValue helps to distinguish between empty files and files with the first empty line.
type EnvValue struct {
	Value      string
	NeedRemove bool
}

// ReadDir reads a specified directory and returns map of env variables.
// Variables represented as files where filename is name of variable, file first line is a value.
func ReadDir(dir string) (Environment, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	env := make(Environment)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.Contains(name, "=") {
			continue
		}

		data, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return nil, err
		}

		if len(data) == 0 {
			env[name] = EnvValue{NeedRemove: true}
			continue
		}

		// Take only the first line
		if idx := bytes.IndexByte(data, '\n'); idx != -1 {
			data = data[:idx]
		}

		// Replace null bytes with newline
		data = bytes.ReplaceAll(data, []byte{0x00}, []byte("\n"))

		// Trim trailing spaces, tabs and carriage returns
		value := strings.TrimRight(string(data), " \t\r")

		env[name] = EnvValue{Value: value}
	}

	return env, nil
}
