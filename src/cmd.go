package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

type Executable interface {
	Exec(string, ...string) Executable
	Setcmd(*exec.Cmd) Executable
	Getcmd() *exec.Cmd
	SetStdin(io.Reader) Executable
	SetStdout(io.Writer) Executable
	SetStderr(io.Writer) Executable
	Setdir(string) Executable
	Getdir() string
	Setenv([]string) Executable
	Getenv() []string
	Run() error
	Out() ([]byte, error)
	FormatOutput([]byte) string
	FormatError(error, []byte) error
}

type Cmd struct {
	name           string
	args           []string
	cmd            *exec.Cmd
	stdin          io.Reader
	stdout, stderr io.Writer
}

func NewCmd() Executable {
	return &Cmd{
		stdin:  os.Stdin,
		stdout: os.Stdout,
		stderr: os.Stderr,
	}
}

func (c *Cmd) Exec(name string, args ...string) Executable {
	c.name = name
	c.args = args
	c.cmd = exec.Command(name, args...)
	return c
}

func (c *Cmd) Setcmd(cmd *exec.Cmd) Executable {
	c.cmd = cmd
	return c
}

func (c *Cmd) Getcmd() *exec.Cmd {
	return c.cmd
}

func (c *Cmd) SetStdin(stdin io.Reader) Executable {
	c.stdin = stdin
	return c
}

func (c *Cmd) SetStdout(stdout io.Writer) Executable {
	c.stdout = stdout
	return c
}

func (c *Cmd) SetStderr(stderr io.Writer) Executable {
	c.stderr = stderr
	return c
}

func (c *Cmd) Setdir(dir string) Executable {
	if c.cmd != nil {
		c.cmd.Dir = dir
	}
	return c
}

func (c *Cmd) Getdir() string {
	if c.cmd != nil {
		return c.cmd.Dir
	}
	return ""
}

func (c *Cmd) Setenv(env []string) Executable {
	if c.cmd != nil {
		c.cmd.Env = env
	}
	return c
}

func (c *Cmd) Getenv() []string {
	if c.cmd != nil {
		return c.cmd.Env
	}
	return []string{}
}

func (c *Cmd) Run() error {
	c.cmd.Stdin = c.stdin
	c.cmd.Stdout = c.stdout
	c.cmd.Stderr = c.stderr
	return c.cmd.Run()
}

func (c *Cmd) Out() ([]byte, error) {
	return c.cmd.CombinedOutput()
}

func (c *Cmd) FormatOutput(out []byte) string {
	return strings.TrimSpace(string(out))
}

func (c *Cmd) FormatError(err error, out []byte) error {
	return fmt.Errorf("%v: %s", err, c.FormatOutput(out))
}
