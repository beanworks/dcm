package main

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path"
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
	case "docker-compose":
		if c.dir == "/test/dcm/run/execute/error" {
			return errors.New("exit status 1")
		}
	case "/bin/bash":
		if len(c.args) == 1 &&
			c.args[0] == "test/dcm/run/init/error" {
			return errors.New("exit status 1")
		}
	}
	return nil
}

func (c *CmdMock) Out() ([]byte, error) {
	switch c.name {
	case "docker":
		if len(c.args) == 4 && c.args[0] == "ps" &&
			c.args[3] == "name=dcmtest_error_" {
			return []byte("error"), errors.New("exit status 1")
		}
		if len(c.args) == 4 && c.args[0] == "ps" &&
			c.args[3] == "name=dcmtest_ok_" {
			return []byte("dcmtest_ok_1"), nil
		}
	}
	return []byte(""), nil
}

// ========== Here starts the real tests for Dcm ==========

func TestSetup(t *testing.T) {
	fixtures := []struct {
		name   string
		config yamlConfig
		code   int
		err    error
	}{
		{
			name: "Negative case for failing reading git repository config",
			config: yamlConfig{
				"service": yamlConfig{"build": "./build/dir"},
			},
			code: 1,
			err:  errors.New("Error reading git repository config for service [service]"),
		},
		{
			name: "Negative case for failing cloning git repository",
			config: yamlConfig{
				"service": yamlConfig{
					"labels": yamlConfig{"dcm.repository": "test-dcm-setup-error"},
				},
			},
			code: 1,
			err:  errors.New("Error cloning git repository for service [service]: exit status 1"),
		},
		{
			name: "Negative case for failing switching to pre-configured git branch",
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
			name: "Positive case, success",
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

	for n, test := range fixtures {
		dcm.Config.Config = test.config
		code, err := dcm.Setup()
		assert.Equal(t, test.code, code, "[%d: %s] Incorrect error code returned", n, test.name)
		if test.err != nil {
			assert.EqualError(t, err, test.err.Error(), "[%d: %s] Incorrect error returned", n, test.name)
		} else {
			assert.NoError(t, err, "[%d: %s] Non-nil error returned", n, test.name)
		}
	}
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

func TestRunExecute(t *testing.T) {
	fixtures := []struct {
		name, dir string
		code      int
	}{
		{
			name: "Negative case for failing to run docker-compose command",
			dir:  "/test/dcm/run/execute/error",
			code: 1,
		},
		{
			name: "Positive case, success",
			dir:  "/test/dcm/run/execute/ok",
			code: 0,
		},
	}

	dcm := NewDcm(NewConfig(), []string{})
	dcm.Cmd = &CmdMock{}

	for n, test := range fixtures {
		dcm.Config.Dir = test.dir
		code, err := dcm.runExecute()
		assert.Equal(t, test.code, code, "[%d: %s] Incorrect error code returned", n, test.name)
		if test.code == 1 {
			assert.Error(t, err, "[%d: %s] Incorrect error returned", n, test.name)
		} else {
			assert.NoError(t, err, "[%d: %s] Non-nil error returned", n, test.name)
		}
	}
}

func TestRunInit(t *testing.T) {
	fixtures := []struct {
		name   string
		config yamlConfig
		code   int
		err    error
	}{
		{
			name: "Negative case for config has no init script",
			config: yamlConfig{
				"service": yamlConfig{
					"labels": yamlConfig{
						"dcm.test": "test",
					},
				},
			},
			code: 0,
			err:  nil,
		},
		{
			name: "Negative case for failing to exuecute init script",
			config: yamlConfig{
				"service": yamlConfig{
					"labels": yamlConfig{
						"dcm.initscript": "test/dcm/run/init/error",
					},
				},
			},
			code: 1,
			err:  errors.New("Error executing init script [test/dcm/run/init/error] for service [service]: exit status 1"),
		},
		{
			name: "Positive case, success",
			config: yamlConfig{
				"service": yamlConfig{
					"labels": yamlConfig{
						"dcm.initscript": "test/dcm/run/init/ok",
					},
				},
			},
			code: 0,
			err:  nil,
		},
	}

	dcm := NewDcm(NewConfig(), []string{})
	dcm.Cmd = &CmdMock{}

	for n, test := range fixtures {
		dcm.Config.Config = test.config
		code, err := dcm.runInit()
		assert.Equal(t, test.code, code, "[%d: %s] Incorrect error code returned", n, test.name)
		if test.err != nil {
			assert.EqualError(t, err, test.err.Error(), "[%d: %s] Incorrect error returned", n, test.name)
		} else {
			assert.NoError(t, err, "[%d: %s] Non-nil error returned", n, test.name)
		}
	}
}

func TestDir(t *testing.T) {
	dir, err := ioutil.TempDir("", "dcm")
	require.Nil(t, err)
	srv, err := ioutil.TempDir(dir, "service")
	require.Nil(t, err)
	defer os.RemoveAll(dir)

	dcm := NewDcm(NewConfig(), []string{})
	var out string

	// Test Dir() without args
	out = helperTestOsStdout(t, func() {
		dcm.Config.Dir = dir
		dcm.Dir()
	})
	assert.Equal(t, dir, out)

	// Test Dir() with args
	out = helperTestOsStdout(t, func() {
		dcm.Config.Srv = dir
		dcm.Dir(path.Base(srv))
	})
	assert.Equal(t, srv, out)
}

func helperTestOsStdout(t *testing.T, fn func()) (out string) {
	// Capture stdout
	stdout := os.Stdout
	r, w, err := os.Pipe()
	require.Nil(t, err)
	os.Stdout = w
	outC := make(chan string)

	go func() {
		var buf bytes.Buffer
		_, err := io.Copy(&buf, r)
		r.Close()
		require.Nil(t, err)
		outC <- buf.String()
	}()

	fn()

	w.Close()
	os.Stdout = stdout
	out = <-outC

	return
}

func TestShell(t *testing.T) {
}

func TestGetContainerId(t *testing.T) {
	var (
		cid string
		err error
	)

	dcm := NewDcm(NewConfig(), []string{})
	dcm.Cmd = &CmdMock{}
	dcm.Config.Project = "dcmtest"

	cid, err = dcm.getContainerId("error")
	assert.Equal(t, "", cid)
	assert.EqualError(t, err, "exit status 1: error")

	cid, err = dcm.getContainerId("empty")
	assert.Equal(t, "", cid)
	assert.EqualError(t, err, "Error: no running container name starts with dcmtest_empty_")

	cid, err = dcm.getContainerId("ok")
	assert.Equal(t, "dcmtest_ok_1", cid)
	assert.NoError(t, err)
}

func TestGetImageRepository(t *testing.T) {
}

func TestBranchForOne(t *testing.T) {
}

func TestUpdate(t *testing.T) {
}

func TestPurgeImages(t *testing.T) {
}

func TestPurgeContainers(t *testing.T) {
}
