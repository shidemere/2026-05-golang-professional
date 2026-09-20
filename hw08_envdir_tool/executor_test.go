package main

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestRunCmd(t *testing.T) {
	t.Setenv("ENVDIR_REPLACE", "old")
	t.Setenv("ENVDIR_REMOVE", "old")
	oldAdded, addedExisted := os.LookupEnv("ENVDIR_ADD")
	t.Cleanup(func() {
		if addedExisted {
			_ = os.Setenv("ENVDIR_ADD", oldAdded)
		} else {
			_ = os.Unsetenv("ENVDIR_ADD")
		}
	})

	env := Environment{
		"ENVDIR_REPLACE": {Value: "new"},
		"ENVDIR_ADD":     {Value: "added"},
		"ENVDIR_REMOVE":  {NeedRemove: true},
	}
	checkEnvironment := `[ "$ENVDIR_REPLACE" = new ] && ` +
		`[ "$ENVDIR_ADD" = added ] && ` +
		`[ -z "${ENVDIR_REMOVE+x}" ] && ` +
		`[ "$1" = "argument value" ]`
	input := &InputCommand{
		ExecutedProgram: "/bin/sh",
		Args:            []string{"-c", checkEnvironment, "sh", "argument value"},
	}

	if code := RunCmd(input, env); code != 0 {
		t.Fatalf("RunCmd returned %d, want 0", code)
	}

	input.Args = []string{"-c", "exit 7"}
	if code := RunCmd(input, nil); code != 7 {
		t.Fatalf("RunCmd returned %d, want 7", code)
	}

	input.ExecutedProgram = filepath.Join(t.TempDir(), "missing-command")
	input.Args = nil
	if code := RunCmd(input, nil); code == 0 {
		t.Fatal("RunCmd returned 0 for a missing command")
	}
}

func TestParseInputCommand(t *testing.T) {
	originalArgs := os.Args
	t.Cleanup(func() { os.Args = originalArgs })
	os.Args = []string{"go-envdir", "/env", "command", "arg1", "arg2"}

	got, err := ParseInputCommand()
	if err != nil {
		t.Fatalf("ParseInputCommand returned an error: %v", err)
	}
	want := &InputCommand{
		PathToDir:       "/env",
		ExecutedProgram: "command",
		Args:            []string{"arg1", "arg2"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ParseInputCommand result mismatch: got %#v, want %#v", got, want)
	}
}
