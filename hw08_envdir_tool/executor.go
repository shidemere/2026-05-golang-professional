package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
)

// RunCmd runs a command with arguments and environment variables from env.
func RunCmd(input *InputCommand, env Environment) (returnCode int) {
	for name, value := range env {
		processEnv(name, value)
	}

	// The command and its arguments are intentionally supplied by the envdir user.
	cmd := exec.CommandContext(context.Background(), input.ExecutedProgram, input.Args...) //nolint:gosec
	cmd.Env = os.Environ()
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout

	err := cmd.Run()
	if err == nil {
		return 0
	}
	if exitError, ok := errors.AsType[*exec.ExitError](err); ok {
		return exitError.ExitCode()
	}
	return 1
}

func processEnv(name string, value EnvValue) {
	if value.NeedRemove {
		_ = os.Unsetenv(name)
		return
	}

	_ = os.Setenv(name, value.Value)
}

// ParseInputCommand parses the environment directory, command and its arguments from os.Args.
func ParseInputCommand() (*InputCommand, error) {
	args := os.Args[1:]
	if len(args) < 2 {
		return nil, fmt.Errorf("not defined program for start or path to dir with environment")
	}

	return &InputCommand{
		PathToDir:       args[0],
		ExecutedProgram: args[1],
		Args:            args[2:],
	}, nil
}

// InputCommand contains the environment directory and command to execute.
type InputCommand struct {
	PathToDir       string
	ExecutedProgram string
	Args            []string
}
