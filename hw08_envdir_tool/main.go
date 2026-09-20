// Package main implements the go-envdir command-line utility.
package main

import (
	"fmt"
	"os"
)

func main() {
	input, err := ParseInputCommand()
	if err != nil {
		fmt.Fprintf(os.Stderr, "can't process input command: %v\n", err)
		os.Exit(1)
	}

	env, err := ReadDir(input.PathToDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "can't read data from %s: %v\n", input.PathToDir, err)
		os.Exit(1)
	}
	code := RunCmd(input, env)
	os.Exit(code)
}
