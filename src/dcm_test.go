package main

import (
	"errors"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ========== Mocked CmdMock struct as test helper for Dcm ==========

type CmdMock struct {
	// CmdMock extends Cmd
	Cmd

	// Fields from CmdMock
	name, dir string
	args, env []string
}

func (c *CmdMock) Exec(name string, args ...string) Executable {
	c.name = name
	c.args = args
	return c
}

func (c *CmdMock) Setdir(dir string) Executable {
	c.dir = dir
	return c
}

func (c *CmdMock) Setenv(env []string) Executable {
	c.env = env
	return c
}

func (c *CmdMock) Getenv() []string {
	return c.env
}

func (c *CmdMock) Run() error {
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

func (c *CmdMock) Out() ([]byte, error) {
	return []byte(""), nil
}

// ========== Test helper for Dcm ==========

type helperTestFixture struct {
	name   string
	config yamlConfig
	code   int
	err    error
}

type helperTestFunc func() (int, error)

func helperTestDcm(t *testing.T, dcm *Dcm, fn helperTestFunc, fix []helperTestFixture) {
	for n, test := range fix {
		dcm.Config.Config = test.config
		code, err := fn()
		assert.Equal(t, test.code, code,
			"[%d: %s] Incorrect error code returned", n, test.name)
		if test.err != nil {
			assert.EqualError(t, err, test.err.Error(),
				"[%d: %s] Incorrect error returned", n, test.name)
		} else {
			assert.NoError(t, err,
				"[%d: %s] Non-nil error returned", n, test.name)
		}
	}
}

// ========== Here starts the real tests for Dcm ==========

func TestSetup(t *testing.T) {
	fixtures := []helperTestFixture{
		{
			name: "Negative test case for failing reading git repository config",
			config: yamlConfig{
				"service": yamlConfig{"build": "./build/dir"},
			},
			code: 1,
			err:  errors.New("Error reading git repository config for service [service]"),
		},
		{
			name: "Negative test case for failing cloning git repository",
			config: yamlConfig{
				"service": yamlConfig{
					"labels": yamlConfig{"dcm.repository": "test-dcm-setup-error"},
				},
			},
			code: 1,
			err:  errors.New("Error cloning git repository for service [service]: exit status 1"),
		},
		{
			name: "Negative test case for failing switching to pre-configured git branch",
			config: yamlConfig{
				"service": yamlConfig{
					"labels": yamlConfig{
						"dcm.repository": "test-dcm-setup-ok",
						"dcm.branch":     "test-dcm-setup-error",
					},
				},
			},
			code: 1,
			err:  errors.New("exit status 1"),
		},
		{
			name: "Positive test case, success",
			config: yamlConfig{
				"service": yamlConfig{
					"labels": yamlConfig{
						"dcm.repository": "test-dcm-setup-ok",
						"dcm.branch":     "test-dcm-setup-ok",
					},
				},
			},
			code: 0,
			err:  nil,
		},
	}

	td, err := ioutil.TempDir("", "dcm")
	require.Nil(t, err)
	defer os.Remove(td)

	dcm := NewDcm(NewConfig(), []string{})
	dcm.Cmd = &CmdMock{}
	dcm.Config.Srv = td

	helperTestDcm(t, dcm, func() (int, error) { return dcm.Setup() }, fixtures)
}

func TestDoForEachService(t *testing.T) {
	var (
		doSrv doForService
		code  int
		err   error
	)

	dcm := NewDcm(NewConfig(), []string{})

	// Bad fixture data
	fixtureBad := yamlConfig{
		"srv1": "config1",
		"srv2": "config2",
	}
	// Good fixture data
	fixtureGood := yamlConfig{
		"srv1": yamlConfig{
			"config": "value",
		},
		"srv2": yamlConfig{
			"config": "value",
		},
	}

	// Negative test case for failing with panic
	dcm.Config.Config = fixtureBad
	doSrv = func(service string, configs yamlConfig) (int, error) {
		return 0, nil
	}
	assert.Panics(t, func() { dcm.doForEachService(doSrv) })

	// Negative test case for failing with error
	dcm.Config.Config = fixtureGood
	doSrv = func(service string, configs yamlConfig) (int, error) {
		return 1, errors.New("Error")
	}
	assert.NotPanics(t, func() { dcm.doForEachService(doSrv) })
	code, err = dcm.doForEachService(doSrv)
	assert.Equal(t, 1, code)
	assert.Error(t, err)

	// Positive test case, success
	dcm.Config.Config = fixtureGood
	doSrv = func(service string, configs yamlConfig) (int, error) {
		return 0, nil
	}
	assert.NotPanics(t, func() { dcm.doForEachService(doSrv) })
	code, err = dcm.doForEachService(doSrv)
	assert.Equal(t, 0, code)
	assert.NoError(t, err)
}
