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
		if len(c.args) == 3 && c.args[0] == "rev-parse" &&
			c.dir == "/test/dcm/git/rev-parse/error" {
			return errors.New("exit status 1")
		}
		if len(c.args) == 2 && c.args[0] == "checkout" {
			switch c.args[1] {
			case "master", "test-dcm-update-error":
				return errors.New("exit status 1")
			}
		}
		if len(c.args) == 1 && c.args[0] == "pull" &&
			c.dir == "/test/dcm/git/pull/error" {
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
	case "docker":
		if len(c.args) == 4 && c.args[0] == "exec" &&
			c.args[2] == "dcmtest_failed_to_run_docker_exec_1" {
			return errors.New("exit status 1")
		}
		if len(c.args) == 2 && c.args[0] == "rmi" &&
			c.args[1] == "dcmtest_bad" {
			return errors.New("exit status 1")
		}
		if len(c.args) == 2 && c.args[0] == "kill" &&
			c.args[1] == "dcmtest_docker_kill_error_1" {
			return errors.New("exit status 1")
		}
		if len(c.args) == 3 && c.args[0] == "rm" &&
			c.args[2] == "dcmtest_docker_rm_error_1" {
			return errors.New("exit status 1")
		}
	}
	return nil
}

func (c *CmdMock) Out() ([]byte, error) {
	switch c.name {
	case "docker":
		if len(c.args) == 4 && c.args[0] == "ps" {
			switch c.args[3] {
			case "name=dcmtest_empty_container_id_":
				return []byte(""), nil
			case "name=dcmtest_ok_":
				return []byte("dcmtest_ok_1"), nil
			case "name=dcmtest_failed_to_run_docker_exec_":
				return []byte("dcmtest_failed_to_run_docker_exec_1"), nil
			case "name=dcmtest_docker_kill_error_":
				return []byte("dcmtest_docker_kill_error_1"), nil
			case "name=dcmtest_docker_rm_error_":
				return []byte("dcmtest_docker_rm_error_1"), nil
			default:
				return []byte("error"), errors.New("exit status 1")
			}
		}
		if len(c.args) == 1 && c.args[0] == "images" {
			switch c.dir {
			case "/test/docker/images/error":
				return []byte("error"), errors.New("exit status 1")
			case "/test/docker/images/remove/error":
				return []byte("dcmtest_bad foobar bazqux"), nil
			default:
				return []byte("dcmtest_ok foobar bazqux"), nil
			}
		}
	}
	return []byte(""), nil
}

// ========== Here starts the real tests for Dcm ==========

func TestCommand(t *testing.T) {
	var (
		code int
		err  error
	)

	dir, err := ioutil.TempDir("", "dcm")
	require.Nil(t, err)
	defer os.Remove(dir)

	dcm := NewDcm(NewConfig(), []string{})
	dcm.Config.Project = "dcmtest"
	dcm.Config.Dir = dir
	dcm.Cmd = &CmdMock{}
	dcm.Cmd.Setdir("/test/dcm/git/rev-parse/ok")

	tests := []struct {
		name string
		args []string
		code int
	}{
		{
			name: "No args passed, print usage, and return code 1",
			args: []string{},
			code: 1,
		},
		{
			name: "test command `dcm help`",
			args: []string{"help"},
			code: 0,
		},
		{
			name: "test command `dcm setup`",
			args: []string{"setup"},
			code: 0,
		},
		{
			name: "test command `dcm run`",
			args: []string{"run"},
			code: 0,
		},
		{
			name: "test command `dcm build`",
			args: []string{"build"},
			code: 0,
		},
		{
			name: "test command `dcm dir`",
			args: []string{"dir"},
			code: 0,
		},
		{
			name: "test command `dcm shell`",
			args: []string{"shell", "ok"},
			code: 0,
		},
		{
			name: "test command `dcm branch`",
			args: []string{"branch", "dcm"},
			code: 0,
		},
		{
			name: "test command `dcm update`",
			args: []string{"update"},
			code: 0,
		},
		{
			name: "dcm command `dcm purge`",
			args: []string{"purge"},
			code: 0,
		},
		{
			name: "dcm command `dcm list`",
			args: []string{"list"},
			code: 0,
		},
		{
			name: "Invalid args passed, print usage, and return code 127",
			args: []string{"invalid"},
			code: 127,
		},
	}

	for n, test := range tests {
		dcm.Args = test.args
		code, err = dcm.Command()
		assert.Equal(t, code, test.code, "[%d: %s] Incorrect error code returned", n, test.name)
		assert.Nil(t, err, "[%d: %s] Non-nil error returned", n, test.name)
	}
}

func TestSetup(t *testing.T) {
	fixtures := []struct {
		name   string
		config yamlConfig
		code   int
		err    error
	}{
		{
			name: "Negative case: failed to read git repository config",
			config: yamlConfig{
				"service": yamlConfig{"build": "./build/dir"},
			},
			code: 1,
			err:  errors.New("Error reading git repository config for service [service]"),
		},
		{
			name: "Negative case: failed to clone git repository",
			config: yamlConfig{
				"service": yamlConfig{
					"labels": yamlConfig{"dcm.repository": "test-dcm-setup-error"},
				},
			},
			code: 1,
			err:  errors.New("Error cloning git repository for service [service]: exit status 1"),
		},
		{
			name: "Negative case: failed to switch to pre-configured git branch",
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
			name: "Positive case: success with docker hub image",
			config: yamlConfig{
				"service": yamlConfig{
					"image": "docker-hub-image",
				},
			},
			code: 0,
			err:  nil,
		},
		{
			name: "Positive case: success with local build",
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

	// Negative case: silently fail when encountering bad config
	dcm.Config.Config = fixtureBad
	doSrv = func(service string, configs yamlConfig) (int, error) {
		return 0, nil
	}
	code, err = dcm.doForEachService(doSrv)
	assert.Equal(t, 1, code)
	assert.Error(t, err)

	// Negative case: fail with error
	dcm.Config.Config = fixtureGood
	doSrv = func(service string, configs yamlConfig) (int, error) {
		return 1, errors.New("Error")
	}
	code, err = dcm.doForEachService(doSrv)
	assert.Equal(t, 1, code)
	assert.Error(t, err)

	// Positive case: success
	dcm.Config.Config = fixtureGood
	doSrv = func(service string, configs yamlConfig) (int, error) {
		return 0, nil
	}
	code, err = dcm.doForEachService(doSrv)
	assert.Equal(t, 0, code)
	assert.NoError(t, err)
}

func TestRun(t *testing.T) {
	var (
		code int
		err  error
	)

	dcm := NewDcm(NewConfig(), []string{})
	dcm.Cmd = &CmdMock{}

	tests := []struct {
		name string
		args []string
		code int
	}{
		{
			name: "No args passed, run `dcm run up` as default option",
			args: []string{},
			code: 0,
		},
		{
			name: "dcm command `dcm run execute`",
			args: []string{"execute"},
			code: 0,
		},
		{
			name: "dcm command `dcm run init`",
			args: []string{"init"},
			code: 0,
		},
		{
			name: "dcm command `dcm run build`",
			args: []string{"build"},
			code: 0,
		},
		{
			name: "dcm command `dcm run start`",
			args: []string{"start"},
			code: 0,
		},
		{
			name: "dcm command `dcm run stop`",
			args: []string{"stop"},
			code: 0,
		},
		{
			name: "dcm command `dcm run restart`",
			args: []string{"restart"},
			code: 0,
		},
		{
			name: "dcm command `dcm run up`",
			args: []string{"up"},
			code: 0,
		},
	}

	for n, test := range tests {
		code, err = dcm.Run(test.args...)
		assert.Equal(t, code, test.code, "[%d: %s] Incorrect error code returned", n, test.name)
		assert.Nil(t, err, "[%d: %s] Non-nil error returned", n, test.name)
	}
}

func TestRunExecute(t *testing.T) {
	fixtures := []struct {
		name, dir string
		code      int
	}{
		{
			name: "Negative case: failed to run docker-compose command",
			dir:  "/test/dcm/run/execute/error",
			code: 1,
		},
		{
			name: "Positive case: success",
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
			name: "Negative case: config has no init script",
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
			name: "Negative case: failed to exuecute init script",
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
			name: "Positive case: success",
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

	// Test Dir() with args, not exists, fall back to dcm.Config.Dir
	out = helperTestOsStdout(t, func() {
		dcm.Config.Srv = dir
		dcm.Dir("not_exists")
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
	var (
		code int
		err  error
	)

	dcm := NewDcm(NewConfig(), []string{})
	dcm.Cmd = &CmdMock{}
	dcm.Config.Project = "dcmtest"

	// Negative case: failed when there is no arg passed
	code, err = dcm.Shell()
	assert.Equal(t, 1, code)
	assert.EqualError(t, err, "Error: no service name specified.")

	// Negative case: failed to get docker container id
	code, err = dcm.Shell("failed_to_get_container_id")
	assert.Equal(t, 1, code)
	assert.EqualError(t, err, "exit status 1: error")

	// Negative case: failed to run docker exec command
	code, err = dcm.Shell("failed_to_run_docker_exec")
	assert.Equal(t, 1, code)
	assert.EqualError(t, err, "exit status 1")

	// Positive case: success
	code, err = dcm.Shell("ok")
	assert.Equal(t, 0, code)
	assert.NoError(t, err)
}

func TestGetContainerId(t *testing.T) {
	fixtures := []struct {
		name, service, cid string
		err                error
	}{
		{
			name:    "Negative case: failed to get docker container id",
			service: "docker_ps_error",
			cid:     "",
			err:     errors.New("exit status 1: error"),
		},
		{
			name:    "Negative case: got an empty container id",
			service: "empty_container_id",
			cid:     "",
			err:     errors.New("Error: no running container name starts with dcmtest_empty_container_id_"),
		},
		{
			name:    "Positive case: success",
			service: "ok",
			cid:     "dcmtest_ok_1",
			err:     nil,
		},
	}

	dcm := NewDcm(NewConfig(), []string{})
	dcm.Cmd = &CmdMock{}
	dcm.Config.Project = "dcmtest"

	for n, test := range fixtures {
		cid, err := dcm.getContainerId(test.service)
		assert.Equal(t, test.cid, cid, "[%d: %s] Incorrect docker container ID returned", n, test.name)
		if test.err != nil {
			assert.EqualError(t, err, test.err.Error(), "[%d: %s] Incorrect error returned", n, test.name)
		} else {
			assert.NoError(t, err, "[%d: %s] Non-nil error returned", n, test.name)
		}
	}
}

func TestGetImageRepository(t *testing.T) {
	fixtures := []struct {
		name, dir, service, repo string
		err                      error
	}{
		{
			name:    "Negative case: failed to execute docker images",
			dir:     "/test/docker/images/error",
			service: "empty_image_repo",
			repo:    "",
			err:     errors.New("exit status 1: error"),
		},
		{
			name:    "Negative case: service image repo name is not in docker images list",
			dir:     "/test/docker/images/ok",
			service: "empty_image_repo",
			repo:    "",
			err:     nil,
		},
		{
			name:    "Positive case: success",
			dir:     "/test/docker/images/ok",
			service: "ok",
			repo:    "dcmtest_ok",
			err:     nil,
		},
	}

	dcm := NewDcm(NewConfig(), []string{})
	dcm.Cmd = &CmdMock{}
	dcm.Config.Project = "dcmtest"

	for n, test := range fixtures {
		dcm.Cmd.Setdir(test.dir)
		repo, err := dcm.getImageRepository(test.service)
		assert.Equal(t, test.repo, repo, "[%d: %s] Incorrect docker image repository name returned", n, test.name)
		if test.err != nil {
			assert.EqualError(t, err, test.err.Error(), "[%d: %s] Incorrect error returned", n, test.name)
		} else {
			assert.NoError(t, err, "[%d: %s] Non-nil error returned", n, test.name)
		}
	}
}

func TestBranchForOne(t *testing.T) {
	var (
		code int
		err  error
	)

	dir, err := ioutil.TempDir("", "dcm")
	require.Nil(t, err)
	srv, err := ioutil.TempDir(dir, "service")
	require.Nil(t, err)
	defer os.RemoveAll(dir)

	dcm := NewDcm(NewConfig(), []string{})
	dcm.Cmd = &CmdMock{}

	// Negative case: get dcm branch failed at os.Chdir()
	dcm.Config.Dir = "/fake/dcm/dir"
	code, err = dcm.branchForOne("dcm")
	assert.Equal(t, 0, code)
	assert.EqualError(t, err, "chdir /fake/dcm/dir: no such file or directory")

	// Negative case: git failed to get dcm branch
	dcm.Config.Dir = dir
	dcm.Cmd.Setdir("/test/dcm/git/rev-parse/error")
	code, err = dcm.branchForOne("dcm")
	assert.Equal(t, 0, code)
	assert.EqualError(t, err, "exit status 1")

	// Negative case: service not exists
	dcm.Config.Srv = "/fake/dcm/srv"
	code, err = dcm.branchForOne("invalid")
	assert.Equal(t, 0, code)
	assert.EqualError(t, err, "Service not exists.")

	// Negative case: get service branch failed at os.Chdir()
	dcm.Config.Srv = "/fake/dcm/srv"
	dcm.Config.Config = yamlConfig{"service": yamlConfig{}}
	code, err = dcm.branchForOne("service")
	assert.Equal(t, 0, code)
	assert.EqualError(t, err, "chdir /fake/dcm/srv/service: no such file or directory")

	// Negative case: git failed to get service branch
	dcm.Config.Srv = dir
	dcm.Config.Config = yamlConfig{path.Base(srv): yamlConfig{}}
	dcm.Cmd.Setdir("/test/dcm/git/rev-parse/error")
	code, err = dcm.branchForOne(path.Base(srv))
	assert.Equal(t, 0, code)
	assert.EqualError(t, err, "exit status 1")

	// Positive case: success with a service using docker hub image
	dcm.Config.Config = yamlConfig{"service": yamlConfig{"image": "docker-hub-image"}}
	code, err = dcm.branchForOne("service")
	assert.Equal(t, 0, code)
	assert.NoError(t, err)

	// Positive case: success with dcm branch
	dcm.Config.Dir = dir
	dcm.Cmd.Setdir("/test/dcm/git/rev-parse/ok")
	code, err = dcm.branchForOne("dcm")
	assert.Equal(t, 0, code)
	assert.NoError(t, err)
}

func TestUpdateForOne(t *testing.T) {
	dir, err := ioutil.TempDir("", "dcm")
	require.Nil(t, err)
	srv, err := ioutil.TempDir(dir, "service")
	require.Nil(t, err)
	defer os.RemoveAll(dir)

	service := path.Base(srv)

	dcm := NewDcm(NewConfig(), []string{})
	dcm.Cmd = &CmdMock{}

	fixtures := []struct {
		name, srv, dir string
		config         yamlConfig
		service        string
		code           int
		err            error
	}{
		{
			name:    "Negative case: service not exists",
			dir:     "",
			srv:     "",
			config:  yamlConfig{},
			service: "invalid",
			code:    0,
			err:     errors.New("Service not exists."),
		},
		{
			name: "Negative case: service not updateable",
			dir:  "",
			srv:  "",
			config: yamlConfig{
				"service": yamlConfig{
					"labels": yamlConfig{
						"dcm.updateable": "false",
					},
				},
			},
			service: "service",
			code:    0,
			err:     errors.New("Service not updateable. Skipping the update."),
		},
		{
			name: "Negative case: failed to os.Chdir()",
			dir:  "/test/dcm/update",
			srv:  "/test/dcm/dir/srv/testproj",
			config: yamlConfig{
				"invalid": yamlConfig{
					"folder": "test",
				},
			},
			service: "invalid",
			code:    0,
			err:     errors.New("chdir /test/dcm/dir/srv/testproj/invalid: no such file or directory"),
		},
		{
			name: "Negative case: cannot read default branch config, use master instead, and got `git checkout` error",
			dir:  "/test/dcm/update",
			srv:  dir,
			config: yamlConfig{
				service: yamlConfig{
					"labels": yamlConfig{
						"dcm.some.other": "label",
					},
				},
			},
			service: service,
			code:    0,
			err:     errors.New("exit status 1"),
		},
		{
			name: "Negative case: failed to execute `git checkout`",
			dir:  "/test/dcm/update",
			srv:  dir,
			config: yamlConfig{
				service: yamlConfig{
					"labels": yamlConfig{
						"dcm.branch": "test-dcm-update-error",
					},
				},
			},
			service: service,
			code:    0,
			err:     errors.New("exit status 1"),
		},
		{
			name: "Negative case: failed to execute `git pull`",
			dir:  "/test/dcm/git/pull/error",
			srv:  dir,
			config: yamlConfig{
				service: yamlConfig{
					"labels": yamlConfig{
						"dcm.branch": "test-dcm-update-ok",
					},
				},
			},
			service: service,
			code:    0,
			err:     errors.New("exit status 1"),
		},
		{
			name: "Positive case: success with docker hub image",
			dir:  "",
			srv:  "",
			config: yamlConfig{
				"service": yamlConfig{
					"image": "docker-hub-image",
				},
			},
			service: "service",
			code:    0,
			err:     nil,
		},
		{
			name: "Positive case: success with local build",
			dir:  "/test/dcm/update",
			srv:  dir,
			config: yamlConfig{
				service: yamlConfig{
					"labels": yamlConfig{
						"dcm.branch": "test-dcm-update-ok",
					},
				},
			},
			service: service,
			code:    0,
			err:     nil,
		},
	}

	for n, test := range fixtures {
		dcm.Cmd.Setdir(test.dir)
		dcm.Config.Srv = test.srv
		dcm.Config.Config = test.config
		code, err := dcm.updateForOne(test.service)
		assert.Equal(t, test.code, code, "[%d: %s] Incorrect error code returned", n, test.name)
		if test.err != nil {
			assert.EqualError(t, err, test.err.Error(), "[%d: %s] Incorrect error returned", n, test.name)
		} else {
			assert.NoError(t, err, "[%d: %s] Non-nil error returned", n, test.name)
		}
	}
}

func TestPurge(t *testing.T) {
	var (
		code int
		err  error
	)

	dcm := NewDcm(NewConfig(), []string{})
	dcm.Cmd = &CmdMock{}

	tests := []struct {
		name string
		args []string
		code int
	}{
		{
			name: "No args passed, run `dcm purge containers` as default option",
			args: []string{},
			code: 0,
		},
		{
			name: "dcm command `dcm purge images`",
			args: []string{"images"},
			code: 0,
		},
		{
			name: "dcm command `dcm run containers`",
			args: []string{"containers"},
			code: 0,
		},
		{
			name: "dcm command `dcm run all`",
			args: []string{"all"},
			code: 0,
		},
	}

	for n, test := range tests {
		code, err = dcm.Purge(test.args...)
		assert.Equal(t, code, test.code, "[%d: %s] Incorrect error code returned", n, test.name)
		assert.Nil(t, err, "[%d: %s] Non-nil error returned", n, test.name)
	}
}

func TestPurgeImages(t *testing.T) {
	fixtures := []struct {
		name, dir string
		config    yamlConfig
		code      int
		err       error
	}{
		{
			name: "Negative case: failed to get image repo name",
			dir:  "/test/docker/images/error",
			config: yamlConfig{
				"service": yamlConfig{
					"test": "purgeImages",
				},
			},
			code: 0,
			err:  nil,
		},
		{
			name: "Negative case: failed to execute `docker rmi`",
			dir:  "/test/docker/images/remove/error",
			config: yamlConfig{
				"bad": yamlConfig{
					"test": "purgeImages",
				},
			},
			code: 0,
			err:  nil,
		},
		{
			name: "Positive case: success",
			dir:  "/test/docker/images/ok",
			config: yamlConfig{
				"ok": yamlConfig{
					"test": "purgeImages",
				},
			},
			code: 0,
			err:  nil,
		},
	}

	dcm := NewDcm(NewConfig(), []string{})
	dcm.Cmd = &CmdMock{}
	dcm.Config.Project = "dcmtest"

	for n, test := range fixtures {
		dcm.Cmd.Setdir(test.dir)
		dcm.Config.Config = test.config
		code, err := dcm.purgeImages()
		assert.Equal(t, test.code, code, "[%d: %s] Incorrect error code returned", n, test.name)
		if test.err != nil {
			assert.EqualError(t, err, test.err.Error(), "[%d: %s] Incorrect error returned", n, test.name)
		} else {
			assert.NoError(t, err, "[%d: %s] Non-nil error returned", n, test.name)
		}
	}
}

func TestPurgeContainers(t *testing.T) {
	fixtures := []struct {
		name   string
		config yamlConfig
		code   int
		err    error
	}{
		{
			name: "Negative case: failed to get container id",
			config: yamlConfig{
				"docker_ps_error": yamlConfig{
					"test": "purgeContainers",
				},
			},
			code: 0,
			err:  nil,
		},
		{
			name: "Negative case: failed to execute `docker kill`",
			config: yamlConfig{
				"docker_kill_error": yamlConfig{
					"test": "purgeContainers",
				},
			},
			code: 0,
			err:  nil,
		},
		{
			name: "Negative case: failed to execute `docker rm`",
			config: yamlConfig{
				"docker_rm_error": yamlConfig{
					"test": "purgeContainers",
				},
			},
			code: 0,
			err:  nil,
		},
		{
			name: "Positive case: success",
			config: yamlConfig{
				"ok": yamlConfig{
					"test": "purgeContainers",
				},
			},
			code: 0,
			err:  nil,
		},
	}

	dcm := NewDcm(NewConfig(), []string{})
	dcm.Cmd = &CmdMock{}
	dcm.Config.Project = "dcmtest"

	for n, test := range fixtures {
		dcm.Config.Config = test.config
		code, err := dcm.purgeContainers()
		assert.Equal(t, test.code, code, "[%d: %s] Incorrect error code returned", n, test.name)
		if test.err != nil {
			assert.EqualError(t, err, test.err.Error(), "[%d: %s] Incorrect error returned", n, test.name)
		} else {
			assert.NoError(t, err, "[%d: %s] Non-nil error returned", n, test.name)
		}
	}
}

func TestList(t *testing.T) {
	out := helperTestOsStdout(t, func() {
		dcm := NewDcm(NewConfig(), []string{})
		dcm.Config.Config = yamlConfig{
			"service": yamlConfig{},
		}
		dcm.List()
	})
	assert.Equal(t, "service\n", out)
}

func TestUsage(t *testing.T) {
	out := helperTestOsStdout(t, func() {
		dcm := NewDcm(NewConfig(), []string{})
		dcm.Usage()
	})

	assert.Contains(t, out, "Docker Compose Manager\n")
	assert.Contains(t, out, "Usage:\n")
	assert.Contains(t, out, "Example:\n")
}
