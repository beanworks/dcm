package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

// ========== Mocked command executer as test helpers for Cmd ==========

// Test helper functions and command mock
func helperCommand(t *testing.T, s ...string) *exec.Cmd {
	cs := []string{"-test.run=TestHelperProcess", "--"}
	cs = append(cs, s...)
	cmd := exec.Command(os.Args[0], cs...)
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
	return cmd
}

// TestHelperProcess isn't a real test. It's used as a helper process
// for TestParameterRun.
func TestHelperProcess(*testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	defer os.Exit(0)

	args := os.Args
	for len(args) > 0 {
		if args[0] == "--" {
			args = args[1:]
			break
		}
		args = args[1:]
	}
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "No command\n")
		os.Exit(2)
	}

	cmd, args := args[0], args[1:]
	switch cmd {
	case "echo":
		iargs := []interface{}{}
		for _, s := range args {
			iargs = append(iargs, s)
		}
		fmt.Println(iargs...)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command %q\n", cmd)
		os.Exit(2)
	}
}

// ========== Here starts the real tests for Cmd ==========

func TestCmdExec(t *testing.T) {
	name := os.Args[0]
	args := []string{"-test.run=TestHelperProcess", "--", "echo"}

	c := &Cmd{}
	c.Exec(name, args...)

	assert.Equal(t, name, c.name)
	assert.Equal(t, args, c.args)
	assert.Equal(t, "*exec.Cmd", reflect.TypeOf(c.cmd).String())
}

func TestCmdSetStdin(t *testing.T) {
	c := &Cmd{}
	c.SetStdin(os.Stdin)

	assert.IsType(t, os.Stdin, c.stdin)
}

func TestCmdSetStderr(t *testing.T) {
	c := &Cmd{}
	c.SetStderr(os.Stderr)

	assert.IsType(t, os.Stderr, c.stderr)
}

func TestCmdSetdir(t *testing.T) {
	c := &Cmd{}
	c.cmd = helperCommand(t, "echo")

	assert.Equal(t, "", c.cmd.Dir)
	c.Setdir("/test/dir")
	assert.Equal(t, "/test/dir", c.cmd.Dir)
}

func TestCmdSetenvAndGetenv(t *testing.T) {
	c := &Cmd{}

	assert.Equal(t, []string{}, c.Getenv())

	c.cmd = helperCommand(t, "echo")

	assert.Equal(t, []string{"GO_WANT_HELPER_PROCESS=1"}, c.cmd.Env)
	assert.Equal(t, []string{"GO_WANT_HELPER_PROCESS=1"}, c.Getenv())

	c.Setenv([]string{"GO_WANT_HELPER_PROCESS=1", "foo=bar", "baz=qux"})

	assert.Equal(t, []string{"GO_WANT_HELPER_PROCESS=1", "foo=bar", "baz=qux"}, c.cmd.Env)
	assert.Equal(t, []string{"GO_WANT_HELPER_PROCESS=1", "foo=bar", "baz=qux"}, c.Getenv())
}

func TestCmdRun(t *testing.T) {
	var out bytes.Buffer

	NewCmd().
		SetStdout(&out).
		Setcmd(helperCommand(t, "echo", "foo", "bar")).
		Run()

	assert.Equal(t, "foo bar\n", out.String())
}

func TestCmdOut(t *testing.T) {
	out, _ := NewCmd().
		Setcmd(helperCommand(t, "echo", "baz", "qux")).
		Out()

	assert.Equal(t, "baz qux\n", string(out))
}

func TestCmdFormatOutput(t *testing.T) {
	c := NewCmd()
	fixture := []byte("foobar\n")

	assert.Equal(t, "foobar", c.FormatOutput(fixture))
}

func TestCmdFormatError(t *testing.T) {
	c := NewCmd()
	err := errors.New("foobar")
	out := []byte("bazqux")

	assert.Equal(t, errors.New("foobar: bazqux"), c.FormatError(err, out))
}
