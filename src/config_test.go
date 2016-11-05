package main

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var yamlFixtureGoodVersion1 string = `
foo:
  bar:
    baz: qux
another: value
yet:
  another:
    - value1
    - value2
    - value3
`

var yamlFixtureGoodVersion2 string = `
version: "2"
services:
  foo:
    bar:
      baz: qux
  another: value
  yet:
    another:
      - value1
      - value2
      - value3
`

var yamlFixtureBad string = `
foo: bar
    - baz: qux
`

func TestCreateNewConfig(t *testing.T) {
	os.Unsetenv("DCM_DIR")
	os.Unsetenv("DCM_PROJECT")

	c := NewConfig()
	wd, _ := os.Getwd()

	assert.Equal(t, &Config{
		Dir:     wd,
		Project: "dcm",
		File:    wd + "/dcm.yml",
		Srv:     wd + "/srv/dcm",
	}, c)
}

func TestCreateNewConfigWithEnvVars(t *testing.T) {
	os.Setenv("DCM_DIR", "/test/dcm/dir")
	os.Setenv("DCM_PROJECT", "testproj")

	c := NewConfig()

	assert.Equal(t, &Config{
		Dir:     "/test/dcm/dir",
		Project: "testproj",
		File:    "/test/dcm/dir/testproj.yml",
		Srv:     "/test/dcm/dir/srv/testproj",
	}, c)
}

func helperCreateTestFile(t *testing.T, prefix, fixture string) string {
	tf, err := ioutil.TempFile("", prefix)
	require.Nil(t, err)

	_, err = tf.WriteString(fixture)
	require.Nil(t, err)

	err = tf.Close()
	require.Nil(t, err)

	return tf.Name()
}

func TestCreateNewConfigFile(t *testing.T) {
	var (
		file   string
		err    error
		config *Config
	)

	// Negative case: Bad YAML file name
	os.Setenv("DCM_CONFIG_FILE", "bad_file_name.yml")
	_, err = NewConfigFile()
	assert.Error(t, err)

	// Negative case: Bad YAML config formatting
	file = helperCreateTestFile(t, "bad_yaml", yamlFixtureBad)
	defer os.Remove(file)
	os.Setenv("DCM_CONFIG_FILE", file)
	_, err = NewConfigFile()
	assert.Error(t, err)

	// Positive case: success
	expectedYaml := yamlConfig{
		"foo": yamlConfig{
			"bar": yamlConfig{
				"baz": "qux",
			},
		},
		"another": "value",
		"yet": yamlConfig{
			"another": []interface{}{
				"value1",
				"value2",
				"value3",
			},
		},
	}

	file = helperCreateTestFile(t, "good_yaml_ver1", yamlFixtureGoodVersion1)
	defer os.Remove(file)
	os.Setenv("DCM_CONFIG_FILE", file)
	config, err = NewConfigFile()
	assert.NoError(t, err)
	assert.Equal(t, expectedYaml, config.Config)

	file = helperCreateTestFile(t, "good_yaml_ver2", yamlFixtureGoodVersion2)
	defer os.Remove(file)
	os.Setenv("DCM_CONFIG_FILE", file)
	config, err = NewConfigFile()
	assert.NoError(t, err)
	assert.Equal(t, expectedYaml, config.Config)
}

func TestIsDockerComposeVersion2(t *testing.T) {
	assert.False(t, isDockerComposeVersion2(yamlConfig{
		"foo": "bar",
	}))

	assert.False(t, isDockerComposeVersion2(yamlConfig{
		"version": "2",
		"foo":     "bar",
	}))

	assert.False(t, isDockerComposeVersion2(yamlConfig{
		"services": yamlConfig{
			"foo": "bar",
		},
	}))

	assert.True(t, isDockerComposeVersion2(yamlConfig{
		"version": "2",
		"services": yamlConfig{
			"foo": "bar",
		},
	}))
}
