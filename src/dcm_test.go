package main

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
