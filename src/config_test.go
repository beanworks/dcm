package main

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var yamlFixture string = `
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

func TestCreateNewConfig(t *testing.T) {
	os.Unsetenv("DCM_DIR")
	os.Unsetenv("DCM_PROJECT")

	c := NewConfig()
	wd, _ := os.Getwd()

	assert.Equal(t, &Config{
		Dir:     wd,
		Project: "bean",
		File:    wd + "/bean.yml",
		Srv:     wd + "/srv/bean",
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

func TestCreateNewConfigFile(t *testing.T) {
	tf, err := ioutil.TempFile("", "testproj.yml")
	require.Nil(t, err)
	defer os.Remove(tf.Name())

	_, err = tf.WriteString(yamlFixture)
	require.Nil(t, err)
	err = tf.Close()
	require.Nil(t, err)

	os.Setenv("DCM_CONFIG_FILE", tf.Name())

	c := NewConfigFile()

	assert.Equal(t, yamlConfig{
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
	}, c.Config)
}
