package main

import (
	"errors"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ========== Mocked CmdTest struct as test helper for Dcm ==========

type CmdTest struct {
	Cmd

	name, dir string
	args, env []string
}

func (c *CmdTest) Exec(name string, args ...string) Executable {
	c.name = name
	c.args = args
	return c
}

func (c *CmdTest) Setdir(dir string) Executable {
	c.dir = dir
	return c
}

func (c *CmdTest) Setenv(env []string) Executable {
	c.env = env
	return c
}

func (c *CmdTest) Getenv() []string {
	return c.env
}

func (c *CmdTest) Run() error {
	switch c.name {
	case "git":
		if len(c.args) == 3 && c.args[0] == "clone" &&
			c.args[1] == "test-dcm-setup-error" {
			return errors.New("exit status 1")
		}
		if len(c.args) == 2 && c.args[0] == "checkout" &&
			c.args[1] == "test-dcm-setup-error" {
			return errors.New("exit status 1")
		}
	}
	return nil
}

func (c *CmdTest) Out() ([]byte, error) {
	return []byte(""), nil
}

// ========== Here starts the real tests for Dcm ==========

func TestSetup(t *testing.T) {
	td, err := ioutil.TempDir("", "dcm")
	require.Nil(t, err)
	defer os.Remove(td)

	d := NewDcm(NewConfig(), []string{})
	d.Cmd = &CmdTest{}
	d.Config.Srv = td

	// Negative test case for failing reading git repository config
	d.Config.Config = yamlConfig{
		"service": yamlConfig{"build": "./build/dir"},
	}
	code, err := d.Setup()
	assert.Equal(t, 1, code)
	assert.EqualError(t, err, "Error reading git repository config for service [service]")

	// Negative test case for failing cloning git repository
	d.Config.Config = yamlConfig{
		"service": yamlConfig{
			"labels": yamlConfig{"dcm.repository": "test-dcm-setup-error"},
		},
	}
	code, err = d.Setup()
	assert.Equal(t, 1, code)
	assert.EqualError(t, err, "Error cloning git repository for service [service]: exit status 1")

	// Negative test case for failing switching to pre-configured git branch
	d.Config.Config = yamlConfig{
		"service": yamlConfig{
			"labels": yamlConfig{
				"dcm.repository": "test-dcm-setup-ok",
				"dcm.branch":     "test-dcm-setup-error",
			},
		},
	}
	code, err = d.Setup()
	assert.Equal(t, 1, code)
	assert.EqualError(t, err, "exit status 1")

	// Positive test case, success
	d.Config.Config = yamlConfig{
		"service": yamlConfig{
			"labels": yamlConfig{
				"dcm.repository": "test-dcm-setup-ok",
				"dcm.branch":     "test-dcm-setup-ok",
			},
		},
	}
	code, err = d.Setup()
	assert.Equal(t, 0, code)
	assert.NoError(t, err)
}

func TestDoForEachServiceFailedWithPanic(t *testing.T) {
	c := NewConfig()
	c.Config = yamlConfig{
		"srv1": "config1",
		"srv2": "config2",
	}
	dcm := NewDcm(c, []string{})
	doSrv := func(service string, configs yamlConfig) (int, error) {
		return 0, nil
	}

	assert.Panics(t, func() { dcm.doForEachService(doSrv) })
}

func TestDoForEachServiceFailedWithError(t *testing.T) {
	c := NewConfig()
	c.Config = yamlConfig{
		"srv1": yamlConfig{
			"config": "value",
		},
		"srv2": yamlConfig{
			"config": "value",
		},
	}
	dcm := NewDcm(c, []string{})
	doSrv := func(service string, configs yamlConfig) (int, error) {
		return 1, errors.New("Error")
	}

	assert.NotPanics(t, func() { dcm.doForEachService(doSrv) })

	code, err := dcm.doForEachService(doSrv)

	assert.Equal(t, 1, code)
	assert.Error(t, err)
}

func TestDoForEachServiceSuccess(t *testing.T) {
	c := NewConfig()
	c.Config = yamlConfig{
		"srv1": yamlConfig{
			"config": "value",
		},
		"srv2": yamlConfig{
			"config": "value",
		},
	}
	dcm := NewDcm(c, []string{})
	doSrv := func(service string, configs yamlConfig) (int, error) {
		return 0, nil
	}

	assert.NotPanics(t, func() { dcm.doForEachService(doSrv) })

	code, err := dcm.doForEachService(doSrv)

	assert.Equal(t, 0, code)
	assert.NoError(t, err)
}
