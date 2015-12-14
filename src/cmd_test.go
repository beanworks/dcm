package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCmdDir(t *testing.T) {
	c := NewCmd()
	c.Cmd = helperCommand(t, "echo")

	assert.Equal(t, "", c.Cmd.Dir)
	c.Dir("/test/dir")
	assert.Equal(t, "/test/dir", c.Cmd.Dir)
}

func TestCmdEnv(t *testing.T) {
	c := NewCmd()
	c.Cmd = helperCommand(t, "echo")

	assert.Equal(t, []string{"GO_WANT_HELPER_PROCESS=1"}, c.Cmd.Env)
	c.Env([]string{"foo=bar", "baz=qux"})
	assert.Equal(t, []string{"GO_WANT_HELPER_PROCESS=1", "foo=bar", "baz=qux"}, c.Cmd.Env)
}

func TestCmdRun(t *testing.T) {
	var out bytes.Buffer
	c := NewCmd()
	c.Stdout = &out
	c.Cmd = helperCommand(t, "echo", "foo", "bar")
	c.Run()

	assert.Equal(t, "foo bar\n", out.String())
}

func TestCmdOut(t *testing.T) {
	c := NewCmd()
	c.Cmd = helperCommand(t, "echo", "baz", "qux")
	out, _ := c.Out()

	assert.Equal(t, "baz qux\n", string(out))
}

func TestCmdString(t *testing.T) {
	c := NewCmd()
	fixture := []byte("foobar\n")

	assert.Equal(t, "foobar", c.String(fixture))
}

func TestCmdError(t *testing.T) {
	c := NewCmd()
	err := errors.New("foobar")
	out := []byte("bazqux")

	assert.Equal(t, errors.New("foobar: bazqux"), c.Error(err, out))
}

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
