package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

type Cmd struct {
	Name   string
	Args   []string
	Cmd    *exec.Cmd
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}

func NewCmd() *Cmd {
	return &Cmd{
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}
}

func (c *Cmd) Exec(name string, args ...string) *Cmd {
	c.Name = name
	c.Args = args
	c.Cmd = exec.Command(name, args...)
	return c
}

func (c *Cmd) Dir(dir string) *Cmd {
	if c.Cmd != nil {
		c.Cmd.Dir = dir
	}
	return c
}

func (c *Cmd) Env(env []string) *Cmd {
	if c.Cmd != nil {
		c.Cmd.Env = append(c.Cmd.Env, env...)
	}
	return c
}

func (c *Cmd) Run() error {
	c.Cmd.Stdin = c.Stdin
	c.Cmd.Stdout = c.Stdout
	c.Cmd.Stderr = c.Stderr
	return c.Cmd.Run()
}

func (c *Cmd) Out() ([]byte, error) {
	return c.Cmd.CombinedOutput()
}

func (c *Cmd) String(out []byte) string {
	return strings.TrimSpace(string(out))
}

func (c *Cmd) Error(err error, out []byte) error {
	return fmt.Errorf("%v: %s", err, c.String(out))
}
