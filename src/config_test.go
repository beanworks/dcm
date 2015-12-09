package main

import (
	"io/ioutil"
	"os"
	"strings"
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
	c := NewConfig()
	wd, _ := os.Getwd()
	dir := strings.TrimSuffix(wd, "/src")

	assert.Equal(t, &Config{
		Dir:     dir,
		Project: "bean",
		File:    dir + "/bean.yml",
		Srv:     dir + "/srv/bean",
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
