package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "Usage: go-envdir /path/to/env/dir command [args...]")
		os.Exit(1)
	}

	dir := os.Args[1]
	cmd := os.Args[2:]

	env, err := ReadDir(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading env dir: %v\n", err)
		os.Exit(1)
	}

	os.Exit(RunCmd(cmd, env))
}
